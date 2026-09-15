package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"timeful/server/models"
)

// CreatePlatformIdentity inserts a new platform identity. The identity's
// uuidv7() default supplies the account identifier, so sign-in stores no
// separate external value.
func (r *Repository) CreatePlatformIdentity(ctx context.Context) (*PlatformIdentity, error) {
	value := &PlatformIdentity{}
	err := r.db.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES
 RETURNING id, created_at`).Scan(&value.ID, &value.CreatedAt)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// GetPlatformIdentity resolves the platform identity a sign-in session carries.
// A value that is not a canonical UUID, and a uuid with no live identity such
// as a deleted account's, both report no identity: there is no compatibility
// lookup for the retired 24-character external identifier.
func (r *Repository) GetPlatformIdentity(ctx context.Context, platformIdentityID string) (*PlatformIdentity, error) {
	if !validPlatformIdentityID(platformIdentityID) {
		return nil, pgx.ErrNoRows
	}
	value := &PlatformIdentity{}
	err := r.db.QueryRow(ctx, `SELECT id, created_at FROM platform_identities WHERE id = $1`, platformIdentityID).Scan(&value.ID, &value.CreatedAt)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// validPlatformIdentityID reports whether a value is the canonical wire form of
// a platform identity UUID. Non-canonical session values resolve to no account
// instead of reaching PostgreSQL as an invalid uuid literal.
func validPlatformIdentityID(value string) bool {
	return validUUID(value)
}

// validUUID reports whether a client-supplied identifier is the canonical UUID
// wire form. Opaque identifiers are validated in the repository before they
// reach a uuid bind, so a non-canonical value keeps reporting pgx.ErrNoRows
// instead of surfacing a PostgreSQL 22P02 cast error to routes.
func validUUID(value string) bool {
	_, ok := models.ParseUUID(value)
	return ok
}

func (r *Repository) CreateEventVisitorIdentity(ctx context.Context, eventID string) (*EventVisitorIdentity, error) {
	value := &EventVisitorIdentity{}
	err := r.db.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1)
 RETURNING id, event_id, public_id, platform_identity_id, created_at`, eventID).Scan(&value.ID, &value.EventID, &value.PublicID, &value.PlatformIdentityID, &value.CreatedAt)
	return value, err
}

// GetEventVisitorIdentity resolves one visitor identity by its opaque public
// identifier within an event. A non-canonical identifier resolves to no
// identity instead of reaching the uuid column as an invalid literal.
func (r *Repository) GetEventVisitorIdentity(ctx context.Context, eventID, publicID string) (*EventVisitorIdentity, error) {
	if !validUUID(publicID) {
		return nil, pgx.ErrNoRows
	}
	value := &EventVisitorIdentity{}
	err := r.db.QueryRow(ctx, `SELECT id, event_id, public_id, platform_identity_id, created_at
 FROM event_visitor_identities WHERE event_id = $1 AND public_id = $2`, eventID, publicID).Scan(&value.ID, &value.EventID, &value.PublicID, &value.PlatformIdentityID, &value.CreatedAt)
	return value, err
}

// AssociateEventVisitorIdentity must be called only after proving browser authority.
// An association cannot be silently reassigned to another account.
func (r *Repository) AssociateEventVisitorIdentity(ctx context.Context, visitorID, platformID string) error {
	result, err := r.db.Exec(ctx, `UPDATE event_visitor_identities SET platform_identity_id = $2
 WHERE id = $1 AND (platform_identity_id IS NULL OR platform_identity_id = $2)`, visitorID, platformID)
	if err == nil && result.RowsAffected() != 1 {
		return errors.New("visitor is associated with another platform identity")
	}
	return err
}

func (r *Repository) CreateEventVisitorCredential(ctx context.Context, value *EventVisitorCredential) error {
	if value == nil || len(value.CredentialHash) != 32 {
		return errors.New("credential SHA-256 hash is required")
	}
	if value.Kind == "" {
		value.Kind = CredentialKindBase
	}
	return r.db.QueryRow(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash, kind, grants_owner)
 VALUES ($1, $2, $3, $4) RETURNING id, created_at`, value.EventVisitorIdentityID, value.CredentialHash, value.Kind, value.GrantsOwner).Scan(&value.ID, &value.CreatedAt)
}

// GetEventVisitorCredential resolves one credential by its opaque identifier
// within a visitor identity. A non-canonical identifier resolves to no
// credential instead of reaching the uuid column as an invalid literal.
func (r *Repository) GetEventVisitorCredential(ctx context.Context, visitorID, credentialID string) (*EventVisitorCredential, error) {
	if !validUUID(credentialID) {
		return nil, pgx.ErrNoRows
	}
	value := &EventVisitorCredential{}
	err := r.db.QueryRow(ctx, `SELECT id, event_visitor_identity_id, credential_hash, created_at, revoked_at, kind, grants_owner
 FROM event_visitor_credentials WHERE event_visitor_identity_id = $1 AND id = $2`, visitorID, credentialID).Scan(&value.ID, &value.EventVisitorIdentityID, &value.CredentialHash, &value.CreatedAt, &value.RevokedAt, &value.Kind, &value.GrantsOwner)
	return value, err
}

func (r *Repository) RevokeEventVisitorCredentials(ctx context.Context, visitorID string) error {
	_, err := r.db.Exec(ctx, `UPDATE event_visitor_credentials SET revoked_at = clock_timestamp()
 WHERE event_visitor_identity_id = $1 AND revoked_at IS NULL`, visitorID)
	return err
}

// VisitorBelongsToAccount reports whether one Event Visitor Identity is
// associated with the given platform identity. A non-canonical identifier
// belongs to no account.
func (r *Repository) VisitorBelongsToAccount(ctx context.Context, visitorID, platformIdentityID string) (bool, error) {
	if !validPlatformIdentityID(platformIdentityID) {
		return false, nil
	}
	var authorized bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM event_visitor_identities
 WHERE id = $1 AND platform_identity_id = $2)`, visitorID, platformIdentityID).Scan(&authorized)
	return authorized, err
}

// EventVisitorIdentitiesBelongingToAccount returns the subset of visitorIDs
// associated with the platform identity. It answers batched authorization
// reads with one query instead of one EXISTS per visitor. A non-canonical
// platform identity owns no visitor.
func (r *Repository) EventVisitorIdentitiesBelongingToAccount(ctx context.Context, platformIdentityID string, visitorIDs []string) (map[string]bool, error) {
	owned := make(map[string]bool, len(visitorIDs))
	if !validPlatformIdentityID(platformIdentityID) || len(visitorIDs) == 0 {
		return owned, nil
	}
	validIDs := make([]string, 0, len(visitorIDs))
	for _, visitorID := range visitorIDs {
		if validPlatformIdentityID(visitorID) {
			validIDs = append(validIDs, visitorID)
		}
	}
	if len(validIDs) == 0 {
		return owned, nil
	}
	rows, err := r.db.Query(ctx, `SELECT id FROM event_visitor_identities
WHERE platform_identity_id = $1 AND id = ANY($2::uuid[])`, platformIdentityID, validIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var visitorID string
		if err := rows.Scan(&visitorID); err != nil {
			return nil, err
		}
		owned[visitorID] = true
	}
	return owned, rows.Err()
}

// EventVisitorHasResponse reports whether one Event Visitor Identity owns a
// response on the event. It backs existence reads without loading every
// response row.
func (r *Repository) EventVisitorHasResponse(ctx context.Context, eventID, visitorID string) (bool, error) {
	var hasResponse bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM event_responses
WHERE event_id = $1 AND event_visitor_identity_id = $2)`, eventID, visitorID).Scan(&hasResponse)
	return hasResponse, err
}

// GetResponseByPublicID resolves one response by its opaque public identifier
// within an event. A non-canonical identifier resolves to no response instead
// of reaching the uuid column as an invalid literal.
func (r *Repository) GetResponseByPublicID(ctx context.Context, eventID, publicID string) (*Response, error) {
	if !validUUID(publicID) {
		return nil, pgx.ErrNoRows
	}
	return r.getResponse(ctx, `event_id = $1 AND public_id = $2`, eventID, publicID)
}

// LockEvent serializes response count changes across concurrent requests and
// returns the locked event, so callers need no second read.
func (r *Repository) LockEvent(ctx context.Context, eventID string) (*Event, error) {
	return scanEvent(r.db.QueryRow(ctx, `SELECT `+eventColumns+` FROM events WHERE id = $1 FOR UPDATE`, eventID))
}

// SetEventOwnerToken is used only during event creation; existing EVCCs cannot recover a token.
func (r *Repository) SetEventOwnerToken(ctx context.Context, eventID string, hash []byte) error {
	if len(hash) != 32 {
		return errors.New("owner token SHA-256 hash is required")
	}
	_, err := r.db.Exec(ctx, `UPDATE events SET owner_edit_token_hash = $2 WHERE id = $1 AND owner_edit_token_hash IS NULL`, eventID, hash)
	return err
}

// AssociateEventOwner must run under the event row lock after token proof.
// It deliberately does not reassign any Event Visitor Identity or response.
func (r *Repository) AssociateEventOwner(ctx context.Context, eventID, platformID string) error {
	_, err := r.db.Exec(ctx, `UPDATE events SET owner_platform_identity_id = $2, updated_at = clock_timestamp() WHERE id = $1`, eventID, platformID)
	return err
}

// EventOwnerBelongsToAccount reports whether an event is owned by the given
// platform identity. A non-canonical identifier owns no event.
func (r *Repository) EventOwnerBelongsToAccount(ctx context.Context, eventID, platformIdentityID string) (bool, error) {
	if !validPlatformIdentityID(platformIdentityID) {
		return false, nil
	}
	var authorized bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM events
 WHERE id = $1 AND owner_platform_identity_id = $2)`, eventID, platformIdentityID).Scan(&authorized)
	return authorized, err
}
