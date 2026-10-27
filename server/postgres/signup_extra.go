package postgres

import (
	"context"
	"time"
)

// Handwritten sibling beside the generated signup.go.
//
// SignupBlock and SignupResponse keep their handwritten declarations because a
// GALA struct declaration would add Copy, Equal, Unapply, New<Type>, and
// StructMeta_<Type> to the package's Go-facing surface. The method bodies live
// in signup.gala as free functions that take the repository first; the
// *Repository methods here are one-line wrappers, because a method declared on
// Repository in a .gala file makes the GALA-declared set Repository's complete
// method set, which hides the methods repository_methods.go declares.

// SignupBlock is one ordered, capacity-limited slot on a signup form event. ID
// is a fresh PostgreSQL UUIDv7 block identifier. Capacity is nil when the block
// is unlimited. Position preserves the block array order.
type SignupBlock struct {
	ID        string
	EventID   string
	Name      string
	Capacity  *int
	StartDate *time.Time
	EndDate   *time.Time
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SignupResponse is one respondent's signup on a signup form event. PublicID is
// the opaque identifier exposed to clients; ID and the owning Event Visitor
// Identity stay internal. PlatformIdentityID or CanonicalGuestName identifies
// the respondent, matching the account-uuid-or-canonical-guest-name key.
// BlockIDs holds the claimed event_signup_blocks identities in their stored
// claim order.
type SignupResponse struct {
	ID                     string
	PublicID               string
	EventID                string
	EventVisitorIdentityID string
	RespondentKind         string
	PlatformIdentityID     *string
	CanonicalGuestName     *string
	Name                   string
	Email                  string
	BlockIDs               []string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// readSignupResponseByID reads one response with its claimed blocks aggregated
// in position order.
func (r *Repository) readSignupResponseByID(ctx context.Context, responseID string) (*SignupResponse, error) {
	return readSignupResponseByID(r, ctx, responseID)
}

// writeSignupResponseMemberships replaces a response's claimed blocks with the
// supplied normalized identities in their given order. The delete and the
// set-based insert both target the join table; the capacity check in the same
// transaction has already proven every identity names a block of the event, and
// the unique constraint backstops a repeated claim.
func (r *Repository) writeSignupResponseMemberships(ctx context.Context, responseID, eventID string, blockIDs []string) error {
	return writeSignupResponseMemberships(r, ctx, responseID, eventID, blockIDs)
}

// CreateSignupBlock appends one block to a signup form event. A non-positive
// position appends after the existing blocks. The event row is locked so a
// concurrent replace or signup cannot interleave block ordering.
func (r *Repository) CreateSignupBlock(ctx context.Context, block *SignupBlock) error {
	return createSignupBlock(r, ctx, block)
}

// ReplaceSignupBlocks writes the supplied blocks as the complete ordered block
// set for an event. A block with an ID must belong to the event and keeps its
// identity and response memberships; a block without an ID is inserted.
// Removed blocks are deleted and their memberships cascade away in the same
// transaction, so the block/response relation never dangles. A removal still
// bumps the affected responses' updated_at, which the retired detach path did
// and a cascade alone does not. The event row is locked for the duration.
func (r *Repository) ReplaceSignupBlocks(ctx context.Context, eventID string, blocks []SignupBlock) ([]SignupBlock, error) {
	return replaceSignupBlocks(r, ctx, eventID, blocks)
}

// ListSignupBlocks returns the event's blocks in their stored order.
func (r *Repository) ListSignupBlocks(ctx context.Context, eventID string) ([]SignupBlock, error) {
	return listSignupBlocks(r, ctx, eventID)
}

// CreateSignupResponse inserts one response for an Event Visitor Identity. The
// respondent identity is inferred when RespondentKind is empty. Guest names are
// canonicalized through the shared respondents normalizer; a duplicate canonical
// guest name or account response surfaces as a unique violation. Capacity is
// reserved atomically under the event row lock, so an over-capacity signup is
// rejected before any row is written.
func (r *Repository) CreateSignupResponse(ctx context.Context, response *SignupResponse) error {
	return createSignupResponse(r, ctx, response)
}

// GetSignupResponseByPublicID resolves one response by its opaque public
// identifier within an event. A public ID that is not a canonical UUID resolves
// to no response instead of being forwarded to the uuid column.
func (r *Repository) GetSignupResponseByPublicID(ctx context.Context, eventID, publicID string) (*SignupResponse, error) {
	return getSignupResponseByPublicID(r, ctx, eventID, publicID)
}

// ListSignupResponses returns every signup response for an event in write order.
func (r *Repository) ListSignupResponses(ctx context.Context, eventID string) ([]SignupResponse, error) {
	return listSignupResponses(r, ctx, eventID)
}

// UpdateSignupResponse writes the respondent identity, name, email, and block
// membership for one response selected by internal ID or opaque public ID. The
// existing row is excluded from the capacity count, so re-saving a response
// never consumes an extra seat. The event row is locked, keeping capacity
// reservation atomic.
func (r *Repository) UpdateSignupResponse(ctx context.Context, response *SignupResponse) error {
	return updateSignupResponse(r, ctx, response)
}

// DeleteSignupResponse removes one response selected by opaque public ID. The
// event row is locked so a deletion cannot interleave with a capacity
// reservation. A missing or non-canonical response is reported as pgx.ErrNoRows.
func (r *Repository) DeleteSignupResponse(ctx context.Context, eventID, publicID string) error {
	return deleteSignupResponse(r, ctx, eventID, publicID)
}

// reserveSignupCapacity verifies every requested block belongs to the event and
// that no limited block is already at capacity, excluding the response being
// rewritten. It must run inside the transaction that holds the event row lock.
func (r *Repository) reserveSignupCapacity(ctx context.Context, eventID string, blockIDs []string, excludeResponseID string) error {
	return reserveSignupCapacity(r, ctx, eventID, blockIDs, excludeResponseID)
}
