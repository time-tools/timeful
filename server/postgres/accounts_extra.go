package postgres

import (
	"context"
	"time"
)

// Account is the authoritative PostgreSQL identity and profile for every
// account. PlatformIdentityID is the account identifier: the native UUIDv7
// primary key of the platform identity row, which is the value held in the
// sign-in session and in every API payload.
//
// The struct stays handwritten because a GALA struct declaration also emits
// Copy, Equal, Unapply, a NewAccount constructor, and StructMeta_Account on
// this exported type. accounts.gala carries the free functions that read and
// write it.
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

// newAccount constructs the zero account that scanAccount fills, because the
// transpiler refuses a named construction of a struct declared in a
// handwritten sibling.
func newAccount() *Account {
	return &Account{}
}

// createPlatformIdentity exposes CreatePlatformIdentity to accounts.gala. A
// Repository method declared in a handwritten sibling is not resolvable from a
// .gala source once the transpiler builds Repository's method set, so the twin
// reaches the method through this free function.
func createPlatformIdentity(tx *Repository, ctx context.Context) (*PlatformIdentity, error) {
	return tx.CreatePlatformIdentity(ctx)
}

// GetAccountByPlatformIdentityID resolves the account for a sign-in session
// value. A value that is not a canonical platform identity UUID resolves to no
// account because the retired 24-character identifier is never looked up.
func (r *Repository) GetAccountByPlatformIdentityID(ctx context.Context, platformIdentityID string) (*Account, error) {
	return getAccountByPlatformIdentityID(r, ctx, platformIdentityID)
}

// GetAccountByEmail resolves the oldest account for a case-insensitive email.
// Email is not unique by contract, so a deterministic order is required.
func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	return getAccountByEmail(r, ctx, email)
}

// ListAccountsByPlatformIdentityIDs returns the authoritative accounts for the
// supplied platform identity UUIDs, keyed by platform identity, so signup reads
// resolve every account profile in one query. Missing or non-canonical
// identifiers are absent from the map.
//
// This method stays handwritten because its success path returns a non-nil map
// together with rows.Err(); a GALA function declaring a Go result list reports
// a non-nil error as a Failure, which zeroes the returned map.
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
	return createAccount(r, ctx, initial)
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
	return findOrCreateAccountByEmail(r, ctx, email, initial)
}

// UpdateAccountProfile writes the authoritative profile fields. Calendar
// connections, tokens, preferences, and the usage counter are never written
// here; the counter advances only through IncrementAccountEventsCreated.
func (r *Repository) UpdateAccountProfile(ctx context.Context, account *Account) error {
	return updateAccountProfile(r, ctx, account)
}

// IncrementAccountEventsCreated advances the usage counter without touching the
// rest of the profile.
func (r *Repository) IncrementAccountEventsCreated(ctx context.Context, platformIdentityID string) error {
	return incrementAccountEventsCreated(r, ctx, platformIdentityID)
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
	return deleteAccountByPlatformIdentityID(r, ctx, platformIdentityID)
}
