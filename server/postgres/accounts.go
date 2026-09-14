package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Account is the authoritative PostgreSQL identity and profile for every
// account. PlatformIdentityID is the account identifier: the native UUIDv7
// primary key of the platform identity row, which is the value held in the
// sign-in session and in every API payload.
type Account struct {
	ID                 string
	PlatformIdentityID string
	Email              string
	FirstName          string
	LastName           string
	Picture            string
	HasCustomName      *bool
	TimezoneOffset     int
	NumEventsCreated   int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

const accountColumns = `a.id, a.platform_identity_id, a.email, a.first_name, a.last_name, a.picture, a.has_custom_name, a.timezone_offset, a.num_events_created, a.created_at, a.updated_at`

func scanAccount(row interface{ Scan(...any) error }) (*Account, error) {
	account := &Account{}
	err := row.Scan(
		&account.ID,
		&account.PlatformIdentityID,
		&account.Email,
		&account.FirstName,
		&account.LastName,
		&account.Picture,
		&account.HasCustomName,
		&account.TimezoneOffset,
		&account.NumEventsCreated,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *Repository) getAccount(ctx context.Context, predicate string, values ...any) (*Account, error) {
	return scanAccount(r.db.QueryRow(ctx, `SELECT `+accountColumns+` FROM accounts a WHERE `+predicate, values...))
}

// GetAccountByPlatformIdentityID resolves the account for a sign-in session
// value. A value that is not a canonical platform identity UUID resolves to no
// account because the retired 24-character identifier is never looked up.
func (r *Repository) GetAccountByPlatformIdentityID(ctx context.Context, platformIdentityID string) (*Account, error) {
	if !validPlatformIdentityID(platformIdentityID) {
		return nil, pgx.ErrNoRows
	}
	return r.getAccount(ctx, `a.platform_identity_id = $1`, platformIdentityID)
}

// GetAccountByEmail resolves the oldest account for a case-insensitive email.
// Email is not unique by contract, so a deterministic order is required.
func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	if email == "" {
		return nil, errors.New("account email is required")
	}
	return r.getAccount(ctx, `lower(a.email) = lower($1) ORDER BY a.created_at, a.id LIMIT 1`, email)
}

// ListAccountsByPlatformIdentityIDs returns the authoritative accounts for the
// supplied platform identity UUIDs, keyed by platform identity, so signup reads
// resolve every account profile in one query. Missing or non-canonical
// identifiers are absent from the map.
func (r *Repository) ListAccountsByPlatformIdentityIDs(ctx context.Context, platformIdentityIDs []string) (map[string]*Account, error) {
	accounts := make(map[string]*Account, len(platformIdentityIDs))
	if len(platformIdentityIDs) == 0 {
		return accounts, nil
	}
	validIDs := make([]string, 0, len(platformIdentityIDs))
	for _, platformIdentityID := range platformIdentityIDs {
		if validPlatformIdentityID(platformIdentityID) {
			validIDs = append(validIDs, platformIdentityID)
		}
	}
	if len(validIDs) == 0 {
		return accounts, nil
	}
	rows, err := r.db.Query(ctx, `SELECT `+accountColumns+` FROM accounts a WHERE a.platform_identity_id = ANY($1::uuid[])`, validIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts[account.PlatformIdentityID] = account
	}
	return accounts, rows.Err()
}

// CreateAccount mints a platform identity and inserts the account once in one
// transaction, so a crash or cancellation cannot leave a partially applied
// sign-in unit. The identity's uuidv7() default supplies the account
// identifier; no external value is minted or stored. A repository that is
// already transaction-scoped (for example the sign-in path) reuses that
// transaction.
func (r *Repository) CreateAccount(ctx context.Context, initial Account) (*Account, error) {
	var account *Account
	err := r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		platform, err := tx.CreatePlatformIdentity(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.db.Exec(ctx, `INSERT INTO accounts
 (platform_identity_id, email, first_name, last_name, picture, has_custom_name, timezone_offset, num_events_created)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (platform_identity_id) DO NOTHING`,
			platform.ID, initial.Email, initial.FirstName, initial.LastName, initial.Picture, initial.HasCustomName, initial.TimezoneOffset, initial.NumEventsCreated); err != nil {
			return err
		}
		account, err = tx.GetAccountByPlatformIdentityID(ctx, platform.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

// FindOrCreateAccountByEmail resolves the single account for a case-insensitive
// email or creates it when none exists. Concurrent first-time sign-ins for the
// same email are serialized by a transaction-scoped advisory lock, so only one
// account and platform identity can be created; the request that loses the race
// is returned the winner's account instead of inserting a duplicate. Distinct
// accounts whose emails already compare equal are left untouched, because email
// is deliberately not unique. The boolean reports whether this call created the
// account.
func (r *Repository) FindOrCreateAccountByEmail(ctx context.Context, email string, initial Account) (*Account, bool, error) {
	if email == "" {
		return nil, false, errors.New("account email is required")
	}
	var account *Account
	var created bool
	err := r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.db.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended(lower($1), 0))`, email); err != nil {
			return err
		}
		existing, err := tx.GetAccountByEmail(ctx, email)
		if err == nil {
			account = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		account, err = tx.CreateAccount(ctx, initial)
		if err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return account, created, nil
}

// UpdateAccountProfile writes the authoritative profile fields. Calendar
// connections, tokens, preferences, and the usage counter are never written
// here; the counter advances only through IncrementAccountEventsCreated.
func (r *Repository) UpdateAccountProfile(ctx context.Context, account *Account) error {
	if account == nil || account.ID == "" {
		return errors.New("account ID is required")
	}
	return r.db.QueryRow(ctx, `UPDATE accounts
SET email = $2, first_name = $3, last_name = $4, picture = $5, has_custom_name = $6, timezone_offset = $7, updated_at = clock_timestamp()
WHERE id = $1 RETURNING updated_at`,
		account.ID, account.Email, account.FirstName, account.LastName, account.Picture, account.HasCustomName, account.TimezoneOffset).Scan(&account.UpdatedAt)
}

// IncrementAccountEventsCreated advances the usage counter without touching the
// rest of the profile.
func (r *Repository) IncrementAccountEventsCreated(ctx context.Context, platformIdentityID string) error {
	if !validPlatformIdentityID(platformIdentityID) {
		return pgx.ErrNoRows
	}
	_, err := r.db.Exec(ctx, `UPDATE accounts SET num_events_created = num_events_created + 1, updated_at = clock_timestamp()
WHERE platform_identity_id = $1`, platformIdentityID)
	return err
}

// DeleteAccountByPlatformIdentityID permanently removes the account authority
// and everything the deleted visitor owns, then records a tombstone keyed by
// the platform identity uuid so the identity can never be adopted again. It
// runs as one transaction under the same advisory lock that creation paths use.
// The tombstone is an audit record rather than a runtime gate: enforcement is
// structural because deletion removes the platform_identities row that account
// resolution keys on, and a later sign-in mints a fresh uuidv7 identity, so a
// deleted identity is never resolvable again.
//
// Events the account organized survive: their ownership pointers are released
// while the events and every other guest's response stay intact. The account's
// own responses, event visitor identities, credentials, and transfers are
// removed. Repeating the call against an already-deleted account is a no-op.
func (r *Repository) DeleteAccountByPlatformIdentityID(ctx context.Context, platformIdentityID string) error {
	if platformIdentityID == "" {
		return errors.New("account platform identity ID is required")
	}
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.db.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, platformIdentityID); err != nil {
			return err
		}
		var storedIdentityID string
		err := tx.db.QueryRow(ctx, `SELECT id FROM platform_identities WHERE id = $1`, platformIdentityID).Scan(&storedIdentityID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil {
			if err := tx.deleteAccountAuthority(ctx, platformIdentityID); err != nil {
				return err
			}
		}
		_, err = tx.db.Exec(ctx, `INSERT INTO account_deletion_tombstones (platform_identity_id) VALUES ($1)
ON CONFLICT (platform_identity_id) DO NOTHING`, platformIdentityID)
		return err
	})
}

// deleteAccountAuthority removes one platform identity's account and everything
// it owns. Visitor identities tied to the account either directly through the
// platform mapping or through a response's account reference are removed
// together with their credentials and transfers.
func (r *Repository) deleteAccountAuthority(ctx context.Context, platformIdentityID string) error {
	visitorIDs := []string{}
	rows, err := r.db.Query(ctx, `SELECT id FROM event_visitor_identities WHERE platform_identity_id = $1
UNION
SELECT event_visitor_identity_id FROM event_responses WHERE platform_identity_id = $1
UNION
SELECT event_visitor_identity_id FROM event_signup_responses WHERE platform_identity_id = $1`, platformIdentityID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var visitorID string
		if err := rows.Scan(&visitorID); err != nil {
			rows.Close()
			return err
		}
		visitorIDs = append(visitorIDs, visitorID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// Sever transfer references to the account or its visitor credentials
	// before the credentials cascade away with their visitor identities.
	if _, err := r.db.Exec(ctx, `DELETE FROM access_transfers
WHERE platform_identity_id = $1
   OR source_credential_id IN (SELECT id FROM event_visitor_credentials WHERE event_visitor_identity_id = ANY($2))
   OR grant_id IN (SELECT id FROM event_visitor_credentials WHERE event_visitor_identity_id = ANY($2))`,
		platformIdentityID, visitorIDs); err != nil {
		return err
	}

	// Release event ownership while preserving the events themselves and every
	// other guest's response.
	if _, err := r.db.Exec(ctx, `UPDATE events
SET owner_platform_identity_id = NULL, owner_event_visitor_identity_id = NULL, updated_at = clock_timestamp()
WHERE owner_platform_identity_id = $1
   OR owner_event_visitor_identity_id = ANY($2)`, platformIdentityID, visitorIDs); err != nil {
		return err
	}

	// Release the account's group attendee relations; the email-keyed
	// memberships survive so the groups keep their invitee lists.
	if _, err := r.db.Exec(ctx, `UPDATE event_attendees
SET platform_identity_id = NULL, updated_at = clock_timestamp()
WHERE platform_identity_id = $1`, platformIdentityID); err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, `DELETE FROM event_responses
WHERE platform_identity_id = $1 OR event_visitor_identity_id = ANY($2)`, platformIdentityID, visitorIDs); err != nil {
		return err
	}
	// visitorIDs are canonical UUIDs scanned from uuid columns, so the ANY bind
	// infers uuid[] from the column type; no index exists on
	// event_signup_responses.event_visitor_identity_id, so this is type
	// consistency rather than an index restoration.
	if _, err := r.db.Exec(ctx, `DELETE FROM event_signup_responses
WHERE platform_identity_id = $1 OR event_visitor_identity_id = ANY($2)`, platformIdentityID, visitorIDs); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM event_visitor_identities WHERE id = ANY($1)`, visitorIDs); err != nil {
		return err
	}
	// Folder memberships cascade with their account-scoped folders.
	if _, err := r.db.Exec(ctx, `DELETE FROM folders WHERE platform_identity_id = $1`, platformIdentityID); err != nil {
		return err
	}
	// Historical daily logs are reporting-only history. Remove the deleted
	// account's memberships and any log the removal emptied.
	if _, err := r.db.Exec(ctx, `DELETE FROM daily_user_log_members WHERE platform_identity_id = $1`, platformIdentityID); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM daily_user_logs l WHERE NOT EXISTS (SELECT 1 FROM daily_user_log_members m WHERE m.daily_user_log_id = l.id)`); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM accounts WHERE platform_identity_id = $1`, platformIdentityID); err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM platform_identities WHERE id = $1`, platformIdentityID)
	return err
}
