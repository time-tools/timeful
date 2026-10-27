package postgres

import (
	"context"
	"time"
)

// The former attendees.go's SQL query strings, scanAttendee row scanner, and
// the method bodies live in attendees.gala. Attendee and the *Repository
// wrappers stay here because a method declared on Repository in a .gala file
// makes the GALA-declared set Repository's complete method set, which hides the
// eighteen methods repository_methods.go declares, and because a GALA struct
// declaration emits Copy, Equal, Unapply, and StructMeta helpers that import
// the GALA runtime.

// Attendee is one email-keyed group invitation. ID is the attendee's PostgreSQL
// UUIDv7 identity. PlatformIdentityID is the resolved account's platform
// identity uuid where an account with Email exists, and is nil when no such
// account exists or after that account is deleted. Declined is nil when unset,
// which is distinct from an explicit false.
//
// The type stays handwritten because every GALA struct declaration emits Copy,
// Equal, Unapply, and StructMeta helpers that import the GALA runtime, which
// would change the Go-facing surface of this exported type.
type Attendee struct {
	ID                 string
	EventID            string
	Email              string
	PlatformIdentityID *string
	Declined           *bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// AddAttendees inserts one email-keyed membership per distinct email in one
// statement, resolving each email to its oldest PostgreSQL account. Exact input
// duplicates collapse to their first occurrence, matching the unique
// (event_id, email) key, and the conflict update keeps the stored decline state
// while filling in a missing account resolution. Blank emails are skipped.
func (r *Repository) AddAttendees(ctx context.Context, eventID string, emails []string, declined *bool) error {
	return addAttendees(r, ctx, eventID, emails, declined)
}

// ListAttendees returns every email-keyed membership for a group event in write
// order.
func (r *Repository) ListAttendees(ctx context.Context, eventID string) ([]Attendee, error) {
	return listAttendees(r, ctx, eventID)
}

// GetAttendeeByEmail resolves one membership by event and invitation email. A
// missing membership is reported as pgx.ErrNoRows.
func (r *Repository) GetAttendeeByEmail(ctx context.Context, eventID, email string) (*Attendee, error) {
	return getAttendeeByEmail(r, ctx, eventID, email)
}

// HasNonDeclinedAttendeeEmail reports whether the event has a non-declined
// membership for the email, matching the dashboard membership predicate. It
// backs the group viewer invitee check with one indexed EXISTS instead of
// scanning the loaded attendee list. An empty event or email reports no
// membership.
func (r *Repository) HasNonDeclinedAttendeeEmail(ctx context.Context, eventID, email string) (bool, error) {
	return hasNonDeclinedAttendeeEmail(r, ctx, eventID, email)
}

// SetAttendeeDeclined writes the explicit decline state for one email-keyed
// membership. It serves both decline and undecline. A missing membership is
// reported as pgx.ErrNoRows.
func (r *Repository) SetAttendeeDeclined(ctx context.Context, eventID, email string, declined bool) error {
	return setAttendeeDeclined(r, ctx, eventID, email, declined)
}

// RemoveAttendees deletes the given email-keyed memberships from one event in
// one statement. Missing memberships are ignored.
func (r *Repository) RemoveAttendees(ctx context.Context, eventID string, emails []string) error {
	return removeAttendees(r, ctx, eventID, emails)
}
