package postgres

import "context"

// DashboardEvent pairs an account-visible event with whether the account owns
// it, whether the account has responded to it, and whether the account is a
// non-declined group member. A responded-but-not-owned event still appears on
// the dashboard, and only owned events carry owner authority. Responded is
// derived from response storage so a group invitee without a response stays in
// the pending state, and Member lets the dashboard derive group responded state
// the same way the attendee lookup does.
//
// The struct stays handwritten because a GALA struct declaration also emits
// Copy, Equal, Unapply, a NewDashboardEvent constructor, and
// StructMeta_DashboardEvent on this exported type. dashboard.gala carries the
// query and the row scanner.
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
func (r *Repository) ListDashboardEvents(ctx context.Context, platformIdentityID, email string) ([]DashboardEvent, error) {
	return listDashboardEvents(r, ctx, platformIdentityID, email)
}
