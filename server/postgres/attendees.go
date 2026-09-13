package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Attendee is one email-keyed group invitation. ID is the attendee's PostgreSQL
// UUIDv7 identity. PlatformIdentityID is the resolved account's platform
// identity uuid where an account with Email exists, and is nil when no such
// account exists or after that account is deleted. Declined is nil when unset,
// which is distinct from an explicit false.
type Attendee struct {
	ID                 string
	EventID            string
	Email              string
	PlatformIdentityID *string
	Declined           *bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

const attendeeColumns = `id, event_id, email, platform_identity_id, declined, created_at, updated_at`

// listAttendeesQuery lists one event's email-keyed memberships in write order.
const listAttendeesQuery = `SELECT ` + attendeeColumns + `
FROM event_attendees WHERE event_id = $1 ORDER BY created_at, id`

// nonDeclinedAttendeeEmailExistsQuery answers the group viewer invitee check
// with one EXISTS over event_attendees (event_id, lower(email)) instead of a
// scan of the loaded attendee list. It ignores declined memberships and matches
// the dashboard membership predicate exactly.
const nonDeclinedAttendeeEmailExistsQuery = `SELECT EXISTS (
    SELECT 1 FROM event_attendees
    WHERE event_id = $1 AND declined IS NOT TRUE AND lower(email) = lower($2)
)`

func scanAttendee(row interface{ Scan(...any) error }) (*Attendee, error) {
	attendee := &Attendee{}
	err := row.Scan(&attendee.ID, &attendee.EventID, &attendee.Email, &attendee.PlatformIdentityID, &attendee.Declined, &attendee.CreatedAt, &attendee.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return attendee, nil
}

// AddAttendees inserts one email-keyed membership per distinct email in one
// statement, resolving each email to its oldest PostgreSQL account. Exact input
// duplicates collapse to their first occurrence, matching the unique
// (event_id, email) key, and the conflict update keeps the stored decline state
// while filling in a missing account resolution. Blank emails are skipped.
func (r *Repository) AddAttendees(ctx context.Context, eventID string, emails []string, declined *bool) error {
	if eventID == "" {
		return errors.New("attendee event ID is required")
	}
	input := make([]string, 0, len(emails))
	for _, email := range emails {
		if strings.TrimSpace(email) == "" {
			continue
		}
		input = append(input, email)
	}
	if len(input) == 0 {
		return nil
	}
	_, err := r.db.Exec(ctx, `INSERT INTO event_attendees (event_id, email, platform_identity_id, declined)
SELECT $1, input.email, resolved.platform_identity_id, $3
FROM (
    SELECT DISTINCT ON (email) email, ordinal
    FROM unnest($2::text[]) WITH ORDINALITY AS listed(email, ordinal)
    ORDER BY email, ordinal
) AS input
LEFT JOIN LATERAL (
    SELECT a.platform_identity_id
    FROM accounts a
    WHERE lower(a.email) = lower(input.email)
    ORDER BY a.created_at, a.id
    LIMIT 1
) AS resolved ON TRUE
ORDER BY input.ordinal
ON CONFLICT (event_id, email) DO UPDATE
SET platform_identity_id = COALESCE(event_attendees.platform_identity_id, EXCLUDED.platform_identity_id), updated_at = clock_timestamp()`,
		eventID, input, declined)
	return err
}

// ListAttendees returns every email-keyed membership for a group event in write
// order.
func (r *Repository) ListAttendees(ctx context.Context, eventID string) ([]Attendee, error) {
	if eventID == "" {
		return nil, errors.New("attendee event ID is required")
	}
	rows, err := r.db.Query(ctx, listAttendeesQuery, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	attendees := []Attendee{}
	for rows.Next() {
		attendee, err := scanAttendee(rows)
		if err != nil {
			return nil, err
		}
		attendees = append(attendees, *attendee)
	}
	return attendees, rows.Err()
}

// GetAttendeeByEmail resolves one membership by event and invitation email. A
// missing membership is reported as pgx.ErrNoRows.
func (r *Repository) GetAttendeeByEmail(ctx context.Context, eventID, email string) (*Attendee, error) {
	if eventID == "" || email == "" {
		return nil, errors.New("attendee event ID and email are required")
	}
	return scanAttendee(r.db.QueryRow(ctx, `SELECT `+attendeeColumns+`
FROM event_attendees WHERE event_id = $1 AND email = $2`, eventID, email))
}

// HasNonDeclinedAttendeeEmail reports whether the event has a non-declined
// membership for the email, matching the dashboard membership predicate. It
// backs the group viewer invitee check with one indexed EXISTS instead of
// scanning the loaded attendee list. An empty event or email reports no
// membership.
func (r *Repository) HasNonDeclinedAttendeeEmail(ctx context.Context, eventID, email string) (bool, error) {
	if eventID == "" || email == "" {
		return false, nil
	}
	var exists bool
	err := r.db.QueryRow(ctx, nonDeclinedAttendeeEmailExistsQuery, eventID, email).Scan(&exists)
	return exists, err
}

// SetAttendeeDeclined writes the explicit decline state for one email-keyed
// membership. It serves both decline and undecline. A missing membership is
// reported as pgx.ErrNoRows.
func (r *Repository) SetAttendeeDeclined(ctx context.Context, eventID, email string, declined bool) error {
	if eventID == "" || email == "" {
		return errors.New("attendee event ID and email are required")
	}
	tag, err := r.db.Exec(ctx, `UPDATE event_attendees
SET declined = $3, updated_at = clock_timestamp()
WHERE event_id = $1 AND email = $2`, eventID, email, declined)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// RemoveAttendees deletes the given email-keyed memberships from one event in
// one statement. Missing memberships are ignored.
func (r *Repository) RemoveAttendees(ctx context.Context, eventID string, emails []string) error {
	if eventID == "" {
		return errors.New("attendee event ID is required")
	}
	if len(emails) == 0 {
		return nil
	}
	_, err := r.db.Exec(ctx, `DELETE FROM event_attendees WHERE event_id = $1 AND email = ANY($2::text[])`, eventID, emails)
	return err
}
