package postgres

import "context"

// This file keeps the *Repository wrappers beside the generated identities.go.
// The validators and every method body live in identities.gala; the wrappers
// stay handwritten because a method declared on Repository in a .gala file
// makes the GALA-declared set Repository's complete method set, which hides the
// methods declared in the other handwritten siblings.

// CreatePlatformIdentity inserts a new platform identity. The identity's
// uuidv7() default supplies the account identifier, so sign-in stores no
// separate external value.
func (r *Repository) CreatePlatformIdentity(ctx context.Context) (*PlatformIdentity, error) {
	return insertPlatformIdentity(r, ctx)
}

// GetPlatformIdentity resolves the platform identity a sign-in session carries.
// A value that is not a canonical UUID, and a uuid with no live identity such
// as a deleted account's, both report no identity: there is no compatibility
// lookup for the retired 24-character external identifier.
func (r *Repository) GetPlatformIdentity(ctx context.Context, platformIdentityID string) (*PlatformIdentity, error) {
	return getPlatformIdentity(r, ctx, platformIdentityID)
}

func (r *Repository) CreateEventVisitorIdentity(ctx context.Context, eventID string) (*EventVisitorIdentity, error) {
	return createEventVisitorIdentity(r, ctx, eventID)
}

// GetEventVisitorIdentity resolves one visitor identity by its opaque public
// identifier within an opaque event identity.
func (r *Repository) GetEventVisitorIdentity(ctx context.Context, eventID, publicID string) (*EventVisitorIdentity, error) {
	return getEventVisitorIdentity(r, ctx, eventID, publicID)
}

// AssociateEventVisitorIdentity attaches an event visitor identity to the
// platform identity that owns the visiting browser session.
func (r *Repository) AssociateEventVisitorIdentity(ctx context.Context, visitorID, platformID string) error {
	return associateEventVisitorIdentity(r, ctx, visitorID, platformID)
}

// CreateEventVisitorCredential stores a passwordless credential for an Event
// Visitor Identity.
func (r *Repository) CreateEventVisitorCredential(ctx context.Context, value *EventVisitorCredential) error {
	return createEventVisitorCredential(r, ctx, value)
}

// GetEventVisitorCredential resolves one credential by its opaque identifier.
func (r *Repository) GetEventVisitorCredential(ctx context.Context, visitorID, credentialID string) (*EventVisitorCredential, error) {
	return getEventVisitorCredential(r, ctx, visitorID, credentialID)
}

// RevokeEventVisitorCredentials revokes every live credential of one visitor
// identity.
func (r *Repository) RevokeEventVisitorCredentials(ctx context.Context, visitorID string) error {
	return revokeEventVisitorCredentials(r, ctx, visitorID)
}

// VisitorBelongsToAccount reports whether one Event Visitor Identity is
// associated with the given platform identity.
func (r *Repository) VisitorBelongsToAccount(ctx context.Context, visitorID, platformIdentityID string) (bool, error) {
	return visitorBelongsToAccount(r, ctx, visitorID, platformIdentityID)
}

// EventVisitorIdentitiesBelongingToAccount returns the subset of visitorIDs
// associated with the platform identity.
func (r *Repository) EventVisitorIdentitiesBelongingToAccount(ctx context.Context, platformIdentityID string, visitorIDs []string) (map[string]bool, error) {
	return eventVisitorIdentitiesBelongingToAccount(r, ctx, platformIdentityID, visitorIDs)
}

// EventVisitorHasResponse reports whether one Event Visitor Identity owns a
// response on the event.
func (r *Repository) EventVisitorHasResponse(ctx context.Context, eventID, visitorID string) (bool, error) {
	return eventVisitorHasResponse(r, ctx, eventID, visitorID)
}

// GetResponseByPublicID resolves one response by its opaque public identifier
// within an event.
func (r *Repository) GetResponseByPublicID(ctx context.Context, eventID, publicID string) (*Response, error) {
	return getResponseByPublicID(r, ctx, eventID, publicID)
}

// LockEvent serializes response count changes across concurrent requests and
// returns the locked event, so callers need no second read.
func (r *Repository) LockEvent(ctx context.Context, eventID string) (*Event, error) {
	return lockEvent(r, ctx, eventID)
}

// SetEventOwnerToken stores the owner token hash of a newly created event.
func (r *Repository) SetEventOwnerToken(ctx context.Context, eventID string, hash []byte) error {
	return setEventOwnerToken(r, ctx, eventID, hash)
}

// AssociateEventOwner stores the platform identity that proved ownership.
func (r *Repository) AssociateEventOwner(ctx context.Context, eventID, platformID string) error {
	return associateEventOwner(r, ctx, eventID, platformID)
}

// EventOwnerBelongsToAccount reports whether an event is owned by the given
// platform identity.
func (r *Repository) EventOwnerBelongsToAccount(ctx context.Context, eventID, platformIdentityID string) (bool, error) {
	return eventOwnerBelongsToAccount(r, ctx, eventID, platformIdentityID)
}
