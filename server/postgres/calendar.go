package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Provider types stored in calendar_accounts.calendar_type.
const (
	CalendarTypeGoogle  = "google"
	CalendarTypeOutlook = "outlook"
	CalendarTypeApple   = "apple"
	CalendarTypeICS     = "ics"
)

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

const calendarAccountColumns = `a.id, a.platform_identity_id, a.calendar_key, a.calendar_type, a.email, a.picture, a.enabled,
 c.oauth_access_token_ciphertext, c.oauth_refresh_token_ciphertext, c.oauth_access_token_expires_at, c.oauth_scope, c.apple_password_ciphertext, c.ics_feed_url_ciphertext,
 a.created_at, a.updated_at`

const calendarAccountFrom = `FROM calendar_accounts a
 LEFT JOIN calendar_account_credentials c ON c.calendar_account_id = a.id`

// validCalendarType reports whether calendarType is one of the provider types
// the schema admits.
func validCalendarType(calendarType string) bool {
	switch calendarType {
	case CalendarTypeGoogle, CalendarTypeOutlook, CalendarTypeApple, CalendarTypeICS:
		return true
	default:
		return false
	}
}

// requirePlatformIdentityID resolves the canonical platform identity UUID that
// owns a calendar record. A non-canonical value and a uuid with no live identity
// are both pgx.ErrNoRows, so a calendar record never creates, merges, or renames
// an account. List and delete paths keep the liveness read because their merged
// statements cannot distinguish a missing identity from an empty result.
func (r *Repository) requirePlatformIdentityID(ctx context.Context, platformIdentityID string) (string, error) {
	if _, err := r.GetPlatformIdentity(ctx, platformIdentityID); err != nil {
		return "", err
	}
	return platformIdentityID, nil
}

// guardPlatformIdentityID validates the canonical platform identity UUID form
// without a database round trip. Statements that imply identity liveness (an
// insert into an identity-keyed table, or a row-scoped update that maps zero
// rows to pgx.ErrNoRows) merge the liveness requirement themselves, so a
// non-canonical value still reports pgx.ErrNoRows before reaching PostgreSQL.
func guardPlatformIdentityID(platformIdentityID string) error {
	if !validPlatformIdentityID(platformIdentityID) {
		return pgx.ErrNoRows
	}
	return nil
}

// calendarAccountID resolves one connection identity by owner and runtime key.
// The cascade from platform_identities means a deleted identity owns no
// connection, so the row lookup alone preserves pgx.ErrNoRows.
func (r *Repository) calendarAccountID(ctx context.Context, platformIdentityID, calendarKey string) (string, error) {
	if calendarKey == "" {
		return "", errors.New("calendar account key is required")
	}
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return "", err
	}
	var accountID string
	if err := r.db.QueryRow(ctx, `SELECT id FROM calendar_accounts WHERE platform_identity_id = $1 AND calendar_key = $2`, platformIdentityID, calendarKey).Scan(&accountID); err != nil {
		return "", err
	}
	return accountID, nil
}

// CreateCalendarAccount inserts one connection and its credentials. A duplicate
// owner-and-key is a unique violation rather than a silent merge.
func (r *Repository) CreateCalendarAccount(ctx context.Context, platformIdentityID string, account *CalendarAccount) error {
	if account == nil {
		return errors.New("calendar account is nil")
	}
	return r.writeCalendarAccount(ctx, platformIdentityID, account, false)
}

// UpsertCalendarAccount creates or updates one connection by owner and runtime
// key. An incoming nil enabled keeps the stored explicit value, so a partial
// write cannot clear an absent-versus-false distinction; an incoming explicit
// value wins.
func (r *Repository) UpsertCalendarAccount(ctx context.Context, platformIdentityID string, account *CalendarAccount) error {
	if account == nil {
		return errors.New("calendar account is nil")
	}
	return r.writeCalendarAccount(ctx, platformIdentityID, account, true)
}

func (r *Repository) writeCalendarAccount(ctx context.Context, platformIdentityID string, account *CalendarAccount, upsert bool) error {
	if account.CalendarKey == "" {
		return errors.New("calendar account key is required")
	}
	if !validCalendarType(account.CalendarType) {
		return fmt.Errorf("unsupported calendar type %q", account.CalendarType)
	}
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return err
	}
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		account.PlatformIdentityID = platformIdentityID
		// Selecting the owner from platform_identities merges identity liveness
		// into the write, so a deleted identity inserts nothing and the missing
		// RETURNING row surfaces as pgx.ErrNoRows.
		statement := `INSERT INTO calendar_accounts (platform_identity_id, calendar_key, calendar_type, email, picture, enabled)
SELECT $1, $2, $3, $4, $5, $6 FROM platform_identities WHERE id = $1`
		if upsert {
			statement += `
ON CONFLICT (platform_identity_id, calendar_key) DO UPDATE
SET calendar_type = EXCLUDED.calendar_type,
    email = EXCLUDED.email,
    picture = EXCLUDED.picture,
    enabled = COALESCE(EXCLUDED.enabled, calendar_accounts.enabled),
    updated_at = clock_timestamp()`
		}
		statement += `
RETURNING id, enabled, created_at, updated_at`
		if err := tx.db.QueryRow(ctx, statement,
			platformIdentityID, account.CalendarKey, account.CalendarType, account.Email, account.Picture, account.Enabled).
			Scan(&account.ID, &account.Enabled, &account.CreatedAt, &account.UpdatedAt); err != nil {
			return err
		}
		return tx.writeCalendarCredentials(ctx, account)
	})
}

// GetCalendarAccountByKey reads one connection and decrypts its credentials. A
// missing connection is pgx.ErrNoRows, which includes a deleted owner because
// the identity cascade removes its connections.
func (r *Repository) GetCalendarAccountByKey(ctx context.Context, platformIdentityID, calendarKey string) (*CalendarAccount, error) {
	if calendarKey == "" {
		return nil, errors.New("calendar account key is required")
	}
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return nil, err
	}
	return r.getCalendarAccount(ctx, `a.platform_identity_id = $1 AND a.calendar_key = $2`, platformIdentityID, calendarKey)
}

// ListCalendarAccountsForUser reads every connection owned by an external
// account identifier and decrypts its credentials. The identity liveness read
// is kept because an empty list is a valid result for a live identity, so the
// merged query could not distinguish it from a non-canonical or deleted owner.
func (r *Repository) ListCalendarAccountsForUser(ctx context.Context, platformIdentityID string) ([]CalendarAccount, error) {
	platformIdentityID, err := r.requirePlatformIdentityID(ctx, platformIdentityID)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `SELECT `+calendarAccountColumns+` `+calendarAccountFrom+`
 WHERE a.platform_identity_id = $1 ORDER BY a.created_at, a.id`, platformIdentityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accounts := []*CalendarAccount{}
	for rows.Next() {
		account, err := r.scanCalendarAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachCalendarSubCalendars(ctx, accounts); err != nil {
		return nil, err
	}
	result := make([]CalendarAccount, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, *account)
	}
	return result, nil
}

// DeleteCalendarAccount removes one connection and, by cascade, its credentials
// and sub-calendars. Deleting a missing key is a no-op, while a missing identity
// stays pgx.ErrNoRows because a merged delete cannot distinguish that from an
// absent connection row.
func (r *Repository) DeleteCalendarAccount(ctx context.Context, platformIdentityID, calendarKey string) error {
	platformIdentityID, err := r.requirePlatformIdentityID(ctx, platformIdentityID)
	if err != nil {
		return err
	}
	if calendarKey == "" {
		return errors.New("calendar account key is required")
	}
	_, err = r.db.Exec(ctx, `DELETE FROM calendar_accounts WHERE platform_identity_id = $1 AND calendar_key = $2`, platformIdentityID, calendarKey)
	return err
}

// SetCalendarAccountEnabled writes the explicit enabled state for one
// connection. A missing connection is pgx.ErrNoRows, and a deleted owner has no
// connection because of the identity cascade.
func (r *Repository) SetCalendarAccountEnabled(ctx context.Context, platformIdentityID, calendarKey string, enabled bool) error {
	if calendarKey == "" {
		return errors.New("calendar account key is required")
	}
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `UPDATE calendar_accounts SET enabled = $3, updated_at = clock_timestamp()
 WHERE platform_identity_id = $1 AND calendar_key = $2`, platformIdentityID, calendarKey, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpdateCalendarOAuthAccessToken replaces only the encrypted OAuth2 access token
// and its plaintext expiry for one connection. It deliberately leaves the
// refresh token, scope, and every non-OAuth credential untouched so a token
// refresh cannot drop provider credentials it did not observe.
func (r *Repository) UpdateCalendarOAuthAccessToken(ctx context.Context, platformIdentityID, calendarKey, accessToken string, expiresAt time.Time) error {
	if calendarKey == "" {
		return errors.New("calendar account key is required")
	}
	if accessToken == "" {
		return errors.New("oauth access token is required")
	}
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return err
	}
	codec, err := credentialCodecFromEnvironment()
	if err != nil {
		return err
	}
	encrypted, err := codec.encrypt(accessToken)
	if err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `UPDATE calendar_account_credentials c
SET oauth_access_token_ciphertext = $3, oauth_access_token_expires_at = $4, updated_at = clock_timestamp()
FROM calendar_accounts a
WHERE c.calendar_account_id = a.id AND a.platform_identity_id = $1 AND a.calendar_key = $2`,
		platformIdentityID, calendarKey, encrypted, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SyncCalendarSubCalendars upserts the supplied provider calendars on one
// connection and removes the stored calendars absent from the supplied set in
// one transaction. An empty set removes every stored sub-calendar. Resolving
// the connection first keeps a missing connection as pgx.ErrNoRows even when
// the set is empty. An incoming nil enabled keeps the stored explicit value, so
// a provider refresh cannot clear a user's absent-versus-false choice. Exact
// duplicate sub-calendar IDs collapse to their first occurrence, because a
// set-based ON CONFLICT DO UPDATE cannot touch the same conflict row twice.
func (r *Repository) SyncCalendarSubCalendars(ctx context.Context, platformIdentityID, calendarKey string, subCalendars []CalendarSubCalendar) error {
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		accountID, err := tx.calendarAccountID(ctx, platformIdentityID, calendarKey)
		if err != nil {
			return err
		}
		ids := make([]string, 0, len(subCalendars))
		names := make([]string, 0, len(subCalendars))
		enabled := make([]*bool, 0, len(subCalendars))
		for _, sub := range subCalendars {
			ids = append(ids, sub.SubCalendarID)
			names = append(names, sub.Name)
			enabled = append(enabled, sub.Enabled)
		}
		if _, err := tx.db.Exec(ctx, `INSERT INTO calendar_sub_calendars (calendar_account_id, sub_calendar_id, name, enabled)
SELECT $1, listed.sub_calendar_id, listed.name, listed.enabled
FROM (
    SELECT DISTINCT ON (sub_calendar_id) sub_calendar_id, name, enabled
    FROM unnest($2::text[], $3::text[], $4::bool[]) WITH ORDINALITY AS incoming(sub_calendar_id, name, enabled, ordinal)
    ORDER BY sub_calendar_id, ordinal
) AS listed
ON CONFLICT (calendar_account_id, sub_calendar_id) DO UPDATE
SET name = EXCLUDED.name,
    enabled = COALESCE(EXCLUDED.enabled, calendar_sub_calendars.enabled),
    updated_at = clock_timestamp()`, accountID, ids, names, enabled); err != nil {
			return err
		}
		// ids is always non-nil so an empty incoming set binds an empty array,
		// where <> ALL('{}') is true for every stored row and deletes them all.
		_, err = tx.db.Exec(ctx, `DELETE FROM calendar_sub_calendars
WHERE calendar_account_id = $1 AND sub_calendar_id <> ALL($2::text[])`, accountID, ids)
		return err
	})
}

// SetCalendarSubCalendarEnabled writes the explicit enabled state for one
// provider calendar. A missing sub-calendar is pgx.ErrNoRows.
func (r *Repository) SetCalendarSubCalendarEnabled(ctx context.Context, platformIdentityID, calendarKey, subCalendarID string, enabled bool) error {
	if subCalendarID == "" {
		return errors.New("sub-calendar ID is required")
	}
	accountID, err := r.calendarAccountID(ctx, platformIdentityID, calendarKey)
	if err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `UPDATE calendar_sub_calendars SET enabled = $3, updated_at = clock_timestamp()
 WHERE calendar_account_id = $1 AND sub_calendar_id = $2`, accountID, subCalendarID, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetCalendarPreferences reads the calendar preferences for an external account
// identifier. Missing preferences are pgx.ErrNoRows, and a deleted owner has no
// preferences because of the identity cascade.
func (r *Repository) GetCalendarPreferences(ctx context.Context, platformIdentityID string) (*CalendarPreferences, error) {
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return nil, err
	}
	preferences := &CalendarPreferences{}
	var options []byte
	err := r.db.QueryRow(ctx, `SELECT platform_identity_id, primary_account_key, token_origin, calendar_options, created_at, updated_at
 FROM calendar_preferences WHERE platform_identity_id = $1`, platformIdentityID).
		Scan(&preferences.PlatformIdentityID, &preferences.PrimaryAccountKey, &preferences.TokenOrigin, &options, &preferences.CreatedAt, &preferences.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if options != nil {
		preferences.CalendarOptions = decodePayload(options)
	}
	return preferences, nil
}

// UpsertCalendarPreferences creates or replaces the preferences for an external
// account identifier. The caller supplies the complete preference state, so a
// nil field clears it: this is the explicit write path, distinct from the
// read-time absent-versus-present semantics.
func (r *Repository) UpsertCalendarPreferences(ctx context.Context, platformIdentityID string, preferences *CalendarPreferences) error {
	if preferences == nil {
		return errors.New("calendar preferences are nil")
	}
	// An empty-string pointer means absent, not a present empty value.
	if preferences.PrimaryAccountKey != nil && *preferences.PrimaryAccountKey == "" {
		preferences.PrimaryAccountKey = nil
	}
	if preferences.TokenOrigin != nil && *preferences.TokenOrigin == "" {
		preferences.TokenOrigin = nil
	}
	if preferences.TokenOrigin != nil && !validTokenOrigin(*preferences.TokenOrigin) {
		return fmt.Errorf("unsupported token origin %q", *preferences.TokenOrigin)
	}
	var options []byte
	if len(preferences.CalendarOptions) > 0 {
		encoded, err := encodePayload(preferences.CalendarOptions)
		if err != nil {
			return err
		}
		options = encoded
	}
	if err := guardPlatformIdentityID(platformIdentityID); err != nil {
		return err
	}
	return r.db.QueryRow(ctx, `INSERT INTO calendar_preferences (platform_identity_id, primary_account_key, token_origin, calendar_options)
 SELECT $1, $2, $3, $4 FROM platform_identities WHERE id = $1
 ON CONFLICT (platform_identity_id) DO UPDATE
 SET primary_account_key = EXCLUDED.primary_account_key,
     token_origin = EXCLUDED.token_origin,
     calendar_options = EXCLUDED.calendar_options,
     updated_at = clock_timestamp()
 RETURNING created_at, updated_at`,
		platformIdentityID, preferences.PrimaryAccountKey, preferences.TokenOrigin, options).
		Scan(&preferences.CreatedAt, &preferences.UpdatedAt)
}

// DeleteCalendarPreferences removes the preferences for an external account
// identifier. Deleting missing preferences is a no-op, while a missing identity
// stays pgx.ErrNoRows because a merged delete cannot distinguish that from an
// absent preference row.
func (r *Repository) DeleteCalendarPreferences(ctx context.Context, platformIdentityID string) error {
	platformIdentityID, err := r.requirePlatformIdentityID(ctx, platformIdentityID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM calendar_preferences WHERE platform_identity_id = $1`, platformIdentityID)
	return err
}

func validTokenOrigin(tokenOrigin string) bool {
	switch tokenOrigin {
	case "ios", "android", "web":
		return true
	default:
		return false
	}
}

func (r *Repository) getCalendarAccount(ctx context.Context, predicate string, values ...any) (*CalendarAccount, error) {
	account, err := r.scanCalendarAccount(r.db.QueryRow(ctx, `SELECT `+calendarAccountColumns+` `+calendarAccountFrom+` WHERE `+predicate, values...))
	if err != nil {
		return nil, err
	}
	if err := r.attachCalendarSubCalendars(ctx, []*CalendarAccount{account}); err != nil {
		return nil, err
	}
	return account, nil
}

func (r *Repository) scanCalendarAccount(row interface{ Scan(...any) error }) (*CalendarAccount, error) {
	account := &CalendarAccount{}
	var accessToken, refreshToken, scope, applePassword, icsFeedURL *string
	var expiresAt *time.Time
	if err := row.Scan(&account.ID, &account.PlatformIdentityID, &account.CalendarKey, &account.CalendarType, &account.Email, &account.Picture, &account.Enabled,
		&accessToken, &refreshToken, &expiresAt, &scope, &applePassword, &icsFeedURL, &account.CreatedAt, &account.UpdatedAt); err != nil {
		return nil, err
	}
	if err := decodeCalendarCredentials(account, accessToken, refreshToken, expiresAt, scope, applePassword, icsFeedURL); err != nil {
		return nil, err
	}
	return account, nil
}

func (r *Repository) attachCalendarSubCalendars(ctx context.Context, accounts []*CalendarAccount) error {
	for _, account := range accounts {
		account.SubCalendars = map[string]CalendarSubCalendar{}
	}
	if len(accounts) == 0 {
		return nil
	}
	ids := make([]string, 0, len(accounts))
	byID := make(map[string]*CalendarAccount, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
		byID[account.ID] = account
	}
	rows, err := r.db.Query(ctx, `SELECT `+calendarSubCalendarColumns+`
 FROM calendar_sub_calendars WHERE calendar_account_id = ANY($1::uuid[]) ORDER BY created_at, id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		sub, err := scanCalendarSubCalendar(rows)
		if err != nil {
			return err
		}
		if account, ok := byID[sub.CalendarAccountID]; ok {
			account.SubCalendars[sub.SubCalendarID] = *sub
		}
	}
	return rows.Err()
}

const calendarSubCalendarColumns = `id, calendar_account_id, sub_calendar_id, name, enabled, created_at, updated_at`

func scanCalendarSubCalendar(row interface{ Scan(...any) error }) (*CalendarSubCalendar, error) {
	sub := &CalendarSubCalendar{}
	if err := row.Scan(&sub.ID, &sub.CalendarAccountID, &sub.SubCalendarID, &sub.Name, &sub.Enabled, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
		return nil, err
	}
	return sub, nil
}

// writeCalendarCredentials replaces the single credential row for a connection.
// An absent credential is stored as NULL rather than an encrypted empty secret,
// and a secret requires a usable ENCRYPTION_KEY.
func (r *Repository) writeCalendarCredentials(ctx context.Context, account *CalendarAccount) error {
	var accessToken, refreshToken, scope, applePassword, icsFeedURL *string
	var expiresAt *time.Time
	if account.OAuth2 != nil {
		scope = stringPointer(account.OAuth2.Scope)
		expiresAt = account.OAuth2.AccessTokenExpiresAt
	}
	encrypted, err := encodeCalendarSecrets(account)
	if err != nil {
		return err
	}
	accessToken, refreshToken, applePassword, icsFeedURL = encrypted.accessToken, encrypted.refreshToken, encrypted.applePassword, encrypted.icsFeedURL
	_, err = r.db.Exec(ctx, `INSERT INTO calendar_account_credentials
 (calendar_account_id, oauth_access_token_ciphertext, oauth_refresh_token_ciphertext, oauth_access_token_expires_at, oauth_scope, apple_password_ciphertext, ics_feed_url_ciphertext)
 VALUES ($1, $2, $3, $4, $5, $6, $7)
 ON CONFLICT (calendar_account_id) DO UPDATE
 SET oauth_access_token_ciphertext = EXCLUDED.oauth_access_token_ciphertext,
     oauth_refresh_token_ciphertext = EXCLUDED.oauth_refresh_token_ciphertext,
     oauth_access_token_expires_at = EXCLUDED.oauth_access_token_expires_at,
     oauth_scope = EXCLUDED.oauth_scope,
     apple_password_ciphertext = EXCLUDED.apple_password_ciphertext,
     ics_feed_url_ciphertext = EXCLUDED.ics_feed_url_ciphertext,
     updated_at = clock_timestamp()`,
		account.ID, accessToken, refreshToken, expiresAt, scope, applePassword, icsFeedURL)
	return err
}

type encodedCalendarSecrets struct {
	accessToken   *string
	refreshToken  *string
	applePassword *string
	icsFeedURL    *string
}

// encodeCalendarSecrets encrypts only non-empty secrets. An empty secret is
// absent, so it is stored as NULL and needs no key.
func encodeCalendarSecrets(account *CalendarAccount) (encodedCalendarSecrets, error) {
	secrets := encodedCalendarSecrets{}
	if !account.requiresEncryption() {
		return secrets, nil
	}
	codec, err := credentialCodecFromEnvironment()
	if err != nil {
		return secrets, err
	}
	if account.OAuth2 != nil {
		if secrets.accessToken, err = encryptCalendarSecret(codec, account.OAuth2.AccessToken); err != nil {
			return secrets, err
		}
		if secrets.refreshToken, err = encryptCalendarSecret(codec, account.OAuth2.RefreshToken); err != nil {
			return secrets, err
		}
	}
	if account.Apple != nil {
		if secrets.applePassword, err = encryptCalendarSecret(codec, account.Apple.Password); err != nil {
			return secrets, err
		}
	}
	if account.ICS != nil {
		if secrets.icsFeedURL, err = encryptCalendarSecret(codec, account.ICS.FeedURL); err != nil {
			return secrets, err
		}
	}
	return secrets, nil
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

func encryptCalendarSecret(codec *credentialCodec, value string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	encrypted, err := codec.encrypt(value)
	if err != nil {
		return nil, err
	}
	return &encrypted, nil
}

func decodeCalendarCredentials(account *CalendarAccount, accessToken, refreshToken *string, expiresAt *time.Time, scope *string, applePassword, icsFeedURL *string) error {
	needsCodec := accessToken != nil || refreshToken != nil || applePassword != nil || icsFeedURL != nil
	var codec *credentialCodec
	if needsCodec {
		resolved, err := credentialCodecFromEnvironment()
		if err != nil {
			return err
		}
		codec = resolved
	}
	if accessToken != nil || refreshToken != nil || expiresAt != nil || scope != nil {
		credentials := &CalendarOAuth2Credentials{AccessTokenExpiresAt: expiresAt, Scope: dereference(scope)}
		var err error
		if credentials.AccessToken, err = decryptCalendarSecret(codec, accessToken); err != nil {
			return err
		}
		if credentials.RefreshToken, err = decryptCalendarSecret(codec, refreshToken); err != nil {
			return err
		}
		account.OAuth2 = credentials
	}
	if applePassword != nil {
		password, err := decryptCalendarSecret(codec, applePassword)
		if err != nil {
			return err
		}
		account.Apple = &CalendarAppleCredentials{Password: password}
	}
	if icsFeedURL != nil {
		feedURL, err := decryptCalendarSecret(codec, icsFeedURL)
		if err != nil {
			return err
		}
		account.ICS = &CalendarICSCredentials{FeedURL: feedURL}
	}
	return nil
}

// decryptCalendarSecret surfaces a decryption failure as an error and never
// substitutes an empty credential.
func decryptCalendarSecret(codec *credentialCodec, ciphertext *string) (string, error) {
	if ciphertext == nil {
		return "", nil
	}
	if codec == nil {
		return "", ErrEncryptionKeyUnavailable
	}
	return codec.decrypt(*ciphertext)
}

func dereference(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
