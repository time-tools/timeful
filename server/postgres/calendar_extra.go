package postgres

import (
	"context"
	"encoding/json"
	"time"
)

// This file carries the members of the former calendar.go that cannot move to
// calendar.gala without changing the package's Go-facing surface. The exported
// Calendar* structs stay handwritten because a GALA struct declaration would add
// its synthesized Copy, Equal, Unapply, and StructMeta members to each exported
// type and wrap every field. The method bodies live in calendar.gala as free
// functions that take the repository first; the *Repository methods here are
// one-line wrappers, because a method declared on Repository in a .gala file
// becomes the repository's complete GALA method set, which hides the methods the
// package's handwritten files declare. The free functions and provider
// constants live in calendar.gala and its generated calendar.go.

// CalendarAccount is one PostgreSQL-owned calendar connection. ID is a fresh
// identity that replaces the legacy account-map entry. CalendarKey keeps the
// legacy email_CALENDARTYPE map key so runtime key behavior is preserved even
// though the identity is new, and PlatformIdentityID is the owning account's
// platform identity UUID.
// Enabled is nil when the legacy document omitted it, which is distinct from an
// explicit false. Credentials are decrypted on read and are never serialized.
type CalendarAccount struct {
	ID                 string
	PlatformIdentityID string
	CalendarKey        string
	CalendarType       string
	Email              string
	Picture            string
	Enabled            *bool
	OAuth2             *CalendarOAuth2Credentials
	Apple              *CalendarAppleCredentials
	ICS                *CalendarICSCredentials
	SubCalendars       map[string]CalendarSubCalendar
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CalendarOAuth2Credentials holds the OAuth2 provider credentials for one
// connection. AccessToken and RefreshToken are encrypted at rest; Scope and the
// expiry are not secrets and are stored plaintext.
type CalendarOAuth2Credentials struct {
	AccessToken          string
	RefreshToken         string
	AccessTokenExpiresAt *time.Time
	Scope                string
}

// CalendarAppleCredentials holds the Apple app password, encrypted at rest.
type CalendarAppleCredentials struct {
	Password string
}

// CalendarICSCredentials holds the ICS feed URL, encrypted at rest because it
// may embed a private token.
type CalendarICSCredentials struct {
	FeedURL string
}

// CalendarSubCalendar is one provider calendar owned by a connection, keyed by
// the provider calendar id. Enabled is nil when the legacy document omitted it,
// distinct from an explicit false.
type CalendarSubCalendar struct {
	ID                string
	CalendarAccountID string
	SubCalendarID     string
	Name              string
	Enabled           *bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CalendarPreferences holds the per-account calendar preferences. A nil
// PrimaryAccountKey preserves the legacy first-Google-account fallback, a nil
// TokenOrigin means undefined, and a nil CalendarOptions is absent, distinct
// from a present empty object.
type CalendarPreferences struct {
	PlatformIdentityID string
	PrimaryAccountKey  *string
	TokenOrigin        *string
	CalendarOptions    json.RawMessage
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// requirePlatformIdentityID resolves the canonical platform identity UUID that
// owns a calendar record. A non-canonical value and a uuid with no live identity
// are both pgx.ErrNoRows, so a calendar record never creates, merges, or renames
// an account. List and delete paths keep the liveness read because their merged
// statements cannot distinguish a missing identity from an empty result.
func (r *Repository) requirePlatformIdentityID(ctx context.Context, platformIdentityID string) (string, error) {
	return requirePlatformIdentityID(r, ctx, platformIdentityID)
}

// calendarAccountID resolves one connection identity by owner and runtime key.
// The cascade from platform_identities means a deleted identity owns no
// connection, so the row lookup alone preserves pgx.ErrNoRows.
func (r *Repository) calendarAccountID(ctx context.Context, platformIdentityID, calendarKey string) (string, error) {
	return calendarAccountID(r, ctx, platformIdentityID, calendarKey)
}

// CreateCalendarAccount inserts one connection and its credentials. A duplicate
// owner-and-key is a unique violation rather than a silent merge.
func (r *Repository) CreateCalendarAccount(ctx context.Context, platformIdentityID string, account *CalendarAccount) error {
	return createCalendarAccount(r, ctx, platformIdentityID, account)
}

// UpsertCalendarAccount creates or updates one connection by owner and runtime
// key. An incoming nil enabled keeps the stored explicit value, so a partial
// write cannot clear an absent-versus-false distinction; an incoming explicit
// value wins.
func (r *Repository) UpsertCalendarAccount(ctx context.Context, platformIdentityID string, account *CalendarAccount) error {
	return upsertCalendarAccount(r, ctx, platformIdentityID, account)
}

func (r *Repository) writeCalendarAccount(ctx context.Context, platformIdentityID string, account *CalendarAccount, upsert bool) error {
	return writeCalendarAccount(r, ctx, platformIdentityID, account, upsert)
}

// GetCalendarAccountByKey reads one connection and decrypts its credentials. A
// missing connection is pgx.ErrNoRows, which includes a deleted owner because
// the identity cascade removes its connections.
func (r *Repository) GetCalendarAccountByKey(ctx context.Context, platformIdentityID, calendarKey string) (*CalendarAccount, error) {
	return getCalendarAccountByKey(r, ctx, platformIdentityID, calendarKey)
}

// ListCalendarAccountsForUser reads every connection owned by an external
// account identifier and decrypts its credentials. The identity liveness read
// is kept because an empty list is a valid result for a live identity, so the
// merged query could not distinguish it from a non-canonical or deleted owner.
func (r *Repository) ListCalendarAccountsForUser(ctx context.Context, platformIdentityID string) ([]CalendarAccount, error) {
	return listCalendarAccountsForUser(r, ctx, platformIdentityID)
}

// DeleteCalendarAccount removes one connection and, by cascade, its credentials
// and sub-calendars. Deleting a missing key is a no-op, while a missing identity
// stays pgx.ErrNoRows because a merged delete cannot distinguish that from an
// absent connection row.
func (r *Repository) DeleteCalendarAccount(ctx context.Context, platformIdentityID, calendarKey string) error {
	return deleteCalendarAccount(r, ctx, platformIdentityID, calendarKey)
}

// SetCalendarAccountEnabled writes the explicit enabled state for one
// connection. A missing connection is pgx.ErrNoRows, and a deleted owner has no
// connection because of the identity cascade.
func (r *Repository) SetCalendarAccountEnabled(ctx context.Context, platformIdentityID, calendarKey string, enabled bool) error {
	return setCalendarAccountEnabled(r, ctx, platformIdentityID, calendarKey, enabled)
}

// UpdateCalendarOAuthAccessToken replaces only the encrypted OAuth2 access token
// and its plaintext expiry for one connection. It deliberately leaves the
// refresh token, scope, and every non-OAuth credential untouched so a token
// refresh cannot drop provider credentials it did not observe.
func (r *Repository) UpdateCalendarOAuthAccessToken(ctx context.Context, platformIdentityID, calendarKey, accessToken string, expiresAt time.Time) error {
	return updateCalendarOAuthAccessToken(r, ctx, platformIdentityID, calendarKey, accessToken, expiresAt)
}

// SyncCalendarSubCalendars upserts the supplied provider calendars on one
// connection and removes the stored calendars absent from the supplied set in
// one transaction. An empty set removes every stored sub-calendar. Resolving
// the connection first keeps a missing connection as pgx.ErrNoRows even when
// the set is empty.
func (r *Repository) SyncCalendarSubCalendars(ctx context.Context, platformIdentityID, calendarKey string, subCalendars []CalendarSubCalendar) error {
	return syncCalendarSubCalendars(r, ctx, platformIdentityID, calendarKey, subCalendars)
}

// SetCalendarSubCalendarEnabled writes the explicit enabled state for one
// provider calendar. A missing sub-calendar is pgx.ErrNoRows.
func (r *Repository) SetCalendarSubCalendarEnabled(ctx context.Context, platformIdentityID, calendarKey, subCalendarID string, enabled bool) error {
	return setCalendarSubCalendarEnabled(r, ctx, platformIdentityID, calendarKey, subCalendarID, enabled)
}

// GetCalendarPreferences reads the calendar preferences for an external account
// identifier. Missing preferences are pgx.ErrNoRows, and a deleted owner has no
// preferences because of the identity cascade.
func (r *Repository) GetCalendarPreferences(ctx context.Context, platformIdentityID string) (*CalendarPreferences, error) {
	return getCalendarPreferences(r, ctx, platformIdentityID)
}

// UpsertCalendarPreferences creates or replaces the preferences for an external
// account identifier. The caller supplies the complete preference state, so a
// nil field clears it: this is the explicit write path, distinct from the
// read-time absent-versus-present semantics.
func (r *Repository) UpsertCalendarPreferences(ctx context.Context, platformIdentityID string, preferences *CalendarPreferences) error {
	return upsertCalendarPreferences(r, ctx, platformIdentityID, preferences)
}

// DeleteCalendarPreferences removes the preferences for an external account
// identifier. Deleting missing preferences is a no-op, while a missing identity
// stays pgx.ErrNoRows because a merged delete cannot distinguish that from an
// absent preference row.
func (r *Repository) DeleteCalendarPreferences(ctx context.Context, platformIdentityID string) error {
	return deleteCalendarPreferences(r, ctx, platformIdentityID)
}

func (r *Repository) getCalendarAccount(ctx context.Context, predicate string, values ...any) (*CalendarAccount, error) {
	return getCalendarAccount(r, ctx, predicate, values...)
}

func (r *Repository) scanCalendarAccount(row rowScanner) (*CalendarAccount, error) {
	return scanCalendarAccount(row)
}

func (r *Repository) attachCalendarSubCalendars(ctx context.Context, accounts []*CalendarAccount) error {
	return attachCalendarSubCalendars(r, ctx, accounts)
}

// writeCalendarCredentials replaces the single credential row for a connection.
// An absent credential is stored as NULL rather than an encrypted empty secret,
// and a secret requires a usable ENCRYPTION_KEY.
func (r *Repository) writeCalendarCredentials(ctx context.Context, account *CalendarAccount) error {
	return writeCalendarCredentials(r, ctx, account)
}

type encodedCalendarSecrets struct {
	accessToken   *string
	refreshToken  *string
	applePassword *string
	icsFeedURL    *string
}

func (account *CalendarAccount) requiresEncryption() bool {
	if account.OAuth2 != nil && (account.OAuth2.AccessToken != "" || account.OAuth2.RefreshToken != "") {
		return true
	}
	if account.Apple != nil && account.Apple.Password != "" {
		return true
	}
	if account.ICS != nil && account.ICS.FeedURL != "" {
		return true
	}
	return false
}
