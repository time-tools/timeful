package postgres

import (
	"context"
	"errors"
)

// DashboardEvent pairs an account-visible event with whether the account owns
// it, whether the account has responded to it, and whether the account is a
// non-declined group member. A responded-but-not-owned event still appears on
// the dashboard, and only owned events carry owner authority. Responded is
// derived from response storage so a group invitee without a response stays in
// the pending state, and Member lets the dashboard derive group responded state
// the same way the attendee lookup does.
type DashboardEvent struct {
	Event     Event
	Owned     bool
	Responded bool
	Member    bool
}

// ListDashboardEvents returns every non-deleted event the account owns, has
// responded to, or is invited to as a group attendee. Ownership resolves
// through the platform identity uuid that is the account identifier. A
// response counts when it names the platform identity directly or when its
// Event Visitor Identity is associated with it, so a signed-in response is
// recovered from the session alone. Group membership resolves by the account's
// email against non-declined attendees. A group invitee keeps their pending
// state until they respond. Every entry comes from PostgreSQL event storage, so
// callers receive one deduplicated list.
// listDashboardEventsQuery returns every non-deleted event the account owns,
// has responded to, or is invited to as a group attendee. The supporting-index
// forced-plan test runs this statement directly so the membership access path
// cannot drift from production.
const listDashboardEventsQuery = `SELECT e.id, e.short_id, e.owner_edit_token_hash, e.owner_platform_identity_id, e.owner_event_visitor_identity_id, e.name, e.type, e.is_archived, e.is_deleted, e.num_responses, e.schedule_version, e.creator_posthog_id, e.created_at, e.updated_at, e.payload,
       COALESCE(e.owner_platform_identity_id = $1, FALSE) AS owned,
       (EXISTS (
          SELECT 1
          FROM event_responses r
          WHERE r.event_id = e.id
            AND (r.platform_identity_id = $1 OR EXISTS (
              SELECT 1 FROM event_visitor_identities v
              WHERE v.id = r.event_visitor_identity_id AND v.platform_identity_id = $1
            ))
       ) OR EXISTS (
          SELECT 1
          FROM event_signup_responses sr
          WHERE sr.event_id = e.id
            AND (sr.platform_identity_id = $1 OR EXISTS (
              SELECT 1 FROM event_visitor_identities sv
              WHERE sv.id = sr.event_visitor_identity_id AND sv.platform_identity_id = $1
            ))
       )) AS responded,
       ($2 <> '' AND EXISTS (
          SELECT 1
          FROM event_attendees a
          WHERE a.event_id = e.id
            AND a.declined IS NOT TRUE
            AND lower(a.email) = lower($2)
       )) AS member
FROM events e
WHERE e.is_deleted = FALSE
  AND (
    e.owner_platform_identity_id = $1
    OR EXISTS (
      SELECT 1
      FROM event_responses r
      WHERE r.event_id = e.id
        AND (r.platform_identity_id = $1 OR EXISTS (
          SELECT 1 FROM event_visitor_identities v
          WHERE v.id = r.event_visitor_identity_id AND v.platform_identity_id = $1
        ))
    )
    OR EXISTS (
      SELECT 1
      FROM event_signup_responses sr
      WHERE sr.event_id = e.id
        AND (sr.platform_identity_id = $1 OR EXISTS (
          SELECT 1 FROM event_visitor_identities sv
          WHERE sv.id = sr.event_visitor_identity_id AND sv.platform_identity_id = $1
        ))
    )
    OR ($2 <> '' AND EXISTS (
      SELECT 1
      FROM event_attendees a
      WHERE a.event_id = e.id
        AND a.declined IS NOT TRUE
        AND lower(a.email) = lower($2)
    ))
  )
ORDER BY e.created_at DESC, e.id DESC`

func (r *Repository) ListDashboardEvents(ctx context.Context, platformIdentityID, email string) ([]DashboardEvent, error) {
	if platformIdentityID == "" {
		return nil, errors.New("account platform identity ID is required")
	}
	rows, err := r.db.Query(ctx, listDashboardEventsQuery, platformIdentityID, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []DashboardEvent{}
	for rows.Next() {
		var item DashboardEvent
		if err := rows.Scan(
			&item.Event.ID,
			&item.Event.ShortID,
			&item.Event.OwnerEditTokenHash,
			&item.Event.OwnerPlatformIdentityID,
			&item.Event.OwnerEventVisitorIdentityID,
			&item.Event.Name,
			&item.Event.Type,
			&item.Event.IsArchived,
			&item.Event.IsDeleted,
			&item.Event.NumResponses,
			&item.Event.ScheduleVersion,
			&item.Event.CreatorPosthogID,
			&item.Event.CreatedAt,
			&item.Event.UpdatedAt,
			&item.Event.Payload,
			&item.Owned,
			&item.Responded,
			&item.Member,
		); err != nil {
			return nil, err
		}
		item.Event.Payload = decodePayload(item.Event.Payload)
		events = append(events, item)
	}
	return events, rows.Err()
}
