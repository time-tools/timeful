package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"timeful/server/models"
)

// calendarTestEncryptionKey is exactly 32 raw bytes, matching the ENCRYPTION_KEY
// contract the credential codec validates.
const calendarTestEncryptionKey = "0123456789abcdef0123456789abcdef"

// newCalendarTestRepository applies the schema migrations into a
// transaction-scoped set of temporary tables. Temp tables shadow the real
// schema so the isolated tests never mutate test-stack records.
func newCalendarTestRepository(t *testing.T) (context.Context, *Repository, pgx.Tx) {
	t.Helper()
	return newMigrationTestRepository(t)
}

func seedCalendarOwner(t *testing.T, ctx context.Context, repo *Repository) *PlatformIdentity {
	t.Helper()
	identity, err := repo.CreatePlatformIdentity(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

// TestCalendarCredentialCodecRoundTripAndTamperDetection proves the AES-256-GCM
// envelope round-trips a secret, carries the version prefix, rejects tampering,
// rejects an unsupported version, and refuses a wrong-length key.
func TestCalendarCredentialCodecRoundTripAndTamperDetection(t *testing.T) {
	codec, err := newCredentialCodec([]byte(calendarTestEncryptionKey))
	if err != nil {
		t.Fatal(err)
	}
	secret := "ya29.super-secret-refresh-token"
	envelope, err := codec.encrypt(secret)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(envelope, "v1:") {
		t.Fatalf("envelope is missing its version prefix: %q", envelope)
	}
	if strings.Contains(envelope, secret) {
		t.Fatal("envelope leaked the plaintext secret")
	}
	decrypted, err := codec.decrypt(envelope)
	if err != nil || decrypted != secret {
		t.Fatalf("round-trip = %q, %v; want %q", decrypted, err, secret)
	}

	// A second encryption must use a fresh nonce so equal secrets differ.
	again, err := codec.encrypt(secret)
	if err != nil {
		t.Fatal(err)
	}
	if again == envelope {
		t.Fatal("two encryptions reused the nonce")
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(envelope, "v1:"))
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0xff
	tampered := "v1:" + base64.StdEncoding.EncodeToString(raw)
	if _, err := codec.decrypt(tampered); err == nil {
		t.Fatal("tampered ciphertext decrypted successfully")
	}
	if _, err := codec.decrypt("v2:" + strings.TrimPrefix(envelope, "v1:")); err == nil {
		t.Fatal("unsupported envelope version decrypted successfully")
	}
	if _, err := codec.decrypt("not-an-envelope"); err == nil {
		t.Fatal("malformed envelope decrypted successfully")
	}
	if _, err := newCredentialCodec([]byte("short")); !errors.Is(err, ErrEncryptionKeyUnavailable) {
		t.Fatalf("short key error = %v, want ErrEncryptionKeyUnavailable", err)
	}
}

// TestCalendarMigrationSchemaConstraints proves the migration enforces provider
// types, non-empty keys, one connection per owner-and-key, one credential row
// per connection, one sub-calendar per connection-and-id, and a valid token
// origin.
func TestCalendarMigrationSchemaConstraints(t *testing.T) {
	ctx, repo, tx := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)

	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_accounts (platform_identity_id, calendar_key, calendar_type) VALUES ($1, '', 'google')`, identity.ID)
		return err
	})
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_accounts (platform_identity_id, calendar_key, calendar_type) VALUES ($1, 'a_google', 'bogus')`, identity.ID)
		return err
	})
	var accountID string
	if err := tx.QueryRow(ctx, `INSERT INTO calendar_accounts (platform_identity_id, calendar_key, calendar_type) VALUES ($1, 'a_google', 'google') RETURNING id`, identity.ID).Scan(&accountID); err != nil {
		t.Fatal(err)
	}
	duplicate := expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_accounts (platform_identity_id, calendar_key, calendar_type) VALUES ($1, 'a_google', 'google')`, identity.ID)
		return err
	})
	if !IsUniqueViolation(duplicate) {
		t.Fatalf("duplicate calendar key error = %v, want a unique violation", duplicate)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO calendar_account_credentials (calendar_account_id, oauth_access_token_ciphertext) VALUES ($1, 'v1:abc')`, accountID); err != nil {
		t.Fatal(err)
	}
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_account_credentials (calendar_account_id, oauth_scope) VALUES ($1, 'x')`, accountID)
		return err
	})

	if _, err := tx.Exec(ctx, `INSERT INTO calendar_sub_calendars (calendar_account_id, sub_calendar_id, name) VALUES ($1, 'primary', 'Primary')`, accountID); err != nil {
		t.Fatal(err)
	}
	subDuplicate := expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_sub_calendars (calendar_account_id, sub_calendar_id, name) VALUES ($1, 'primary', 'Again')`, accountID)
		return err
	})
	if !IsUniqueViolation(subDuplicate) {
		t.Fatalf("duplicate sub-calendar error = %v, want a unique violation", subDuplicate)
	}
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_sub_calendars (calendar_account_id, sub_calendar_id) VALUES (gen_random_uuid(), 'orphan')`)
		return err
	})

	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_preferences (platform_identity_id, token_origin) VALUES ($1, 'desktop')`, identity.ID)
		return err
	})
}

// TestCalendarAccountRepositoryEncryptsCredentialsAtRest proves every provider
// secret round-trips through the repository, is stored as a versioned GCM
// envelope rather than plaintext, and that non-secret fields stay plaintext.
func TestCalendarAccountRepositoryEncryptsCredentialsAtRest(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", calendarTestEncryptionKey)
	ctx, repo, tx := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)

	expiresAt := time.UnixMilli(1700000000000).UTC()
	oauth := &CalendarAccount{
		CalendarKey:  "ada@example.com_google",
		CalendarType: CalendarTypeGoogle,
		Email:        "ada@example.com",
		OAuth2: &CalendarOAuth2Credentials{
			AccessToken:          "access-token-value",
			RefreshToken:         "refresh-token-value",
			AccessTokenExpiresAt: &expiresAt,
			Scope:                "calendar.readonly",
		},
	}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, oauth); err != nil {
		t.Fatal(err)
	}
	apple := &CalendarAccount{
		CalendarKey:  "ada@example.com_apple",
		CalendarType: CalendarTypeApple,
		Email:        "ada@example.com",
		Apple:        &CalendarAppleCredentials{Password: "app-specific-password"},
	}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, apple); err != nil {
		t.Fatal(err)
	}
	ics := &CalendarAccount{
		CalendarKey:  "Team feed_ics",
		CalendarType: CalendarTypeICS,
		Email:        "Team feed",
		ICS:          &CalendarICSCredentials{FeedURL: "https://example.com/private/feed.ics?token=secret"},
	}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, ics); err != nil {
		t.Fatal(err)
	}

	stored, err := repo.GetCalendarAccountByKey(ctx, identity.ID, oauth.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if stored.OAuth2 == nil || stored.OAuth2.AccessToken != "access-token-value" || stored.OAuth2.RefreshToken != "refresh-token-value" {
		t.Fatalf("oauth round-trip lost credentials: %#v", stored.OAuth2)
	}
	if stored.OAuth2.Scope != "calendar.readonly" || stored.OAuth2.AccessTokenExpiresAt == nil || stored.OAuth2.AccessTokenExpiresAt.UnixMilli() != expiresAt.UnixMilli() {
		t.Fatalf("non-secret oauth fields are unexpected: %#v", stored.OAuth2)
	}
	if stored.CalendarKey != oauth.CalendarKey || stored.CalendarType != CalendarTypeGoogle || stored.Email != "ada@example.com" {
		t.Fatalf("connection identity is unexpected: %#v", stored)
	}

	storedApple, err := repo.GetCalendarAccountByKey(ctx, identity.ID, apple.CalendarKey)
	if err != nil || storedApple.Apple == nil || storedApple.Apple.Password != "app-specific-password" {
		t.Fatalf("apple round-trip = %#v, %v", storedApple.Apple, err)
	}
	storedICS, err := repo.GetCalendarAccountByKey(ctx, identity.ID, ics.CalendarKey)
	if err != nil || storedICS.ICS == nil || storedICS.ICS.FeedURL != "https://example.com/private/feed.ics?token=secret" {
		t.Fatalf("ics round-trip = %#v, %v", storedICS.ICS, err)
	}

	var accessToken, refreshToken string
	if err := tx.QueryRow(ctx, `SELECT oauth_access_token_ciphertext, oauth_refresh_token_ciphertext FROM calendar_account_credentials WHERE calendar_account_id = $1`, oauth.ID).Scan(&accessToken, &refreshToken); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(accessToken, "v1:") || !strings.HasPrefix(refreshToken, "v1:") {
		t.Fatalf("oauth tokens are not versioned envelopes: %q, %q", accessToken, refreshToken)
	}
	if strings.Contains(accessToken, "access-token-value") || strings.Contains(refreshToken, "refresh-token-value") {
		t.Fatal("oauth token ciphertext leaked plaintext")
	}
	// The oauth connection has no Apple or ICS secrets, so those stay NULL.
	var appleEnvelope, feedEnvelope *string
	if err := tx.QueryRow(ctx, `SELECT apple_password_ciphertext, ics_feed_url_ciphertext FROM calendar_account_credentials WHERE calendar_account_id = $1`, oauth.ID).Scan(&appleEnvelope, &feedEnvelope); err != nil {
		t.Fatal(err)
	}
	if appleEnvelope != nil || feedEnvelope != nil {
		t.Fatalf("unrelated credential columns were not left absent: %v, %v", appleEnvelope, feedEnvelope)
	}
	var appleStored, feedStored string
	if err := tx.QueryRow(ctx, `SELECT apple_password_ciphertext FROM calendar_account_credentials WHERE calendar_account_id = $1`, apple.ID).Scan(&appleStored); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(appleStored, "v1:") || strings.Contains(appleStored, "app-specific-password") {
		t.Fatalf("apple password is not encrypted at rest: %q", appleStored)
	}
	if err := tx.QueryRow(ctx, `SELECT ics_feed_url_ciphertext FROM calendar_account_credentials WHERE calendar_account_id = $1`, ics.ID).Scan(&feedStored); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(feedStored, "v1:") || strings.Contains(feedStored, "token=secret") {
		t.Fatalf("ics feed url is not encrypted at rest: %q", feedStored)
	}
	var scope string
	if err := tx.QueryRow(ctx, `SELECT oauth_scope FROM calendar_account_credentials WHERE calendar_account_id = $1`, oauth.ID).Scan(&scope); err != nil || scope != "calendar.readonly" {
		t.Fatalf("scope should stay plaintext: %q, %v", scope, err)
	}
}

// TestCalendarAccountRepositorySurfacesDecryptionFailure proves a corrupted or
// unsupported credential envelope returns an error instead of an empty secret.
func TestCalendarAccountRepositorySurfacesDecryptionFailure(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", calendarTestEncryptionKey)
	ctx, repo, tx := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)
	account := &CalendarAccount{
		CalendarKey:  "ada@example.com_google",
		CalendarType: CalendarTypeGoogle,
		Email:        "ada@example.com",
		OAuth2:       &CalendarOAuth2Credentials{AccessToken: "access-token-value"},
	}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, account); err != nil {
		t.Fatal(err)
	}
	var envelope string
	if err := tx.QueryRow(ctx, `SELECT oauth_access_token_ciphertext FROM calendar_account_credentials WHERE calendar_account_id = $1`, account.ID).Scan(&envelope); err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(envelope, "v1:"))
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0xff
	if _, err := tx.Exec(ctx, `UPDATE calendar_account_credentials SET oauth_access_token_ciphertext = $2 WHERE calendar_account_id = $1`, account.ID, "v1:"+base64.StdEncoding.EncodeToString(raw)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey); err == nil {
		t.Fatal("corrupted credential envelope read without error")
	}
	if _, err := tx.Exec(ctx, `UPDATE calendar_account_credentials SET oauth_access_token_ciphertext = 'v2:still-an-envelope' WHERE calendar_account_id = $1`, account.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey); err == nil {
		t.Fatal("unsupported credential envelope read without error")
	}
}

// TestCalendarAccountEnabledAbsentVersusFalse proves nil enabled stays absent,
// an explicit false is stored, and a later nil upsert cannot clear it.
func TestCalendarAccountEnabledAbsentVersusFalse(t *testing.T) {
	ctx, repo, _ := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)
	account := &CalendarAccount{CalendarKey: "ada@example.com_google", CalendarType: CalendarTypeGoogle, Email: "ada@example.com"}
	if err := repo.UpsertCalendarAccount(ctx, identity.ID, account); err != nil {
		t.Fatal(err)
	}
	if account.Enabled != nil {
		t.Fatalf("absent enabled was not preserved: %#v", account.Enabled)
	}
	account.Enabled = boolPointer(false)
	if err := repo.UpsertCalendarAccount(ctx, identity.ID, account); err != nil {
		t.Fatal(err)
	}
	account.Enabled = nil
	if err := repo.UpsertCalendarAccount(ctx, identity.ID, account); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Enabled == nil || *stored.Enabled {
		t.Fatalf("nil upsert cleared explicit false: %#v", stored.Enabled)
	}
	if err := repo.SetCalendarAccountEnabled(ctx, identity.ID, account.CalendarKey, true); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil || stored.Enabled == nil || !*stored.Enabled {
		t.Fatalf("explicit enable not stored: %#v, %v", stored.Enabled, err)
	}
	if err := repo.SetCalendarAccountEnabled(ctx, identity.ID, "missing_key", true); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing connection toggle error = %v, want pgx.ErrNoRows", err)
	}
}

// TestCalendarSubCalendarLifecycle proves the set-based add, update, remove,
// and enabled toggles preserve absent-versus-false and stay scoped to their
// connection.
func TestCalendarSubCalendarLifecycle(t *testing.T) {
	ctx, repo, _ := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)
	account := &CalendarAccount{CalendarKey: "ada@example.com_google", CalendarType: CalendarTypeGoogle, Email: "ada@example.com"}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, account); err != nil {
		t.Fatal(err)
	}
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, []CalendarSubCalendar{
		{SubCalendarID: "primary", Name: "Primary"},
	}); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if stored.SubCalendars["primary"].Enabled != nil {
		t.Fatalf("absent sub-calendar enabled was not preserved: %#v", stored.SubCalendars["primary"].Enabled)
	}
	if err := repo.SetCalendarSubCalendarEnabled(ctx, identity.ID, account.CalendarKey, "primary", false); err != nil {
		t.Fatal(err)
	}
	// A provider refresh with naming but no enabled choice must not clear false.
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, []CalendarSubCalendar{
		{SubCalendarID: "primary", Name: "Renamed"},
		{SubCalendarID: "work", Name: "Work"},
	}); err != nil {
		t.Fatal(err)
	}

	account, err = repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(account.SubCalendars) != 2 {
		t.Fatalf("unexpected sub-calendar set: %#v", account.SubCalendars)
	}
	renamed := account.SubCalendars["primary"]
	if renamed.Name != "Renamed" || renamed.Enabled == nil || *renamed.Enabled {
		t.Fatalf("sub-calendar update lost name or enabled: %#v", renamed)
	}
	if account.SubCalendars["work"].Enabled != nil {
		t.Fatalf("new sub-calendar should be absent enabled: %#v", account.SubCalendars["work"].Enabled)
	}
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, []CalendarSubCalendar{
		{SubCalendarID: "work", Name: "Work"},
	}); err != nil {
		t.Fatal(err)
	}
	account, err = repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil || len(account.SubCalendars) != 1 {
		t.Fatalf("sub-calendar removal = %#v, %v", account.SubCalendars, err)
	}
	if _, ok := account.SubCalendars["primary"]; ok {
		t.Fatal("removed sub-calendar still present")
	}
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, []CalendarSubCalendar{
		{SubCalendarID: "work", Name: "Work"},
	}); err != nil {
		t.Fatalf("repeated sync should be a no-op: %v", err)
	}
	if err := repo.SetCalendarSubCalendarEnabled(ctx, identity.ID, account.CalendarKey, "missing", true); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing sub-calendar toggle error = %v, want pgx.ErrNoRows", err)
	}
}

// TestSyncCalendarSubCalendarsBatch proves the set-based sync upserts the
// supplied set, collapses duplicate IDs to their first occurrence, removes
// stored calendars absent from it, treats an empty set as delete-all, preserves
// absent-versus-false enabled, and still reports pgx.ErrNoRows for a missing
// connection even when the set is empty.
func TestSyncCalendarSubCalendarsBatch(t *testing.T) {
	ctx, repo, _ := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)
	account := &CalendarAccount{CalendarKey: "batch@example.com_google", CalendarType: CalendarTypeGoogle, Email: "batch@example.com"}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, account); err != nil {
		t.Fatal(err)
	}
	enabled := false
	// The duplicate primary ID must collapse to its first occurrence instead
	// of raising ON CONFLICT ... cannot affect row a second time (21000).
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, []CalendarSubCalendar{
		{SubCalendarID: "primary", Name: "Primary", Enabled: &enabled},
		{SubCalendarID: "primary", Name: "Duplicate"},
		{SubCalendarID: "work", Name: "Work"},
	}); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.SubCalendars) != 2 {
		t.Fatalf("unexpected sub-calendar set: %#v", stored.SubCalendars)
	}
	primary := stored.SubCalendars["primary"]
	if primary.Name != "Primary" || primary.Enabled == nil || *primary.Enabled {
		t.Fatalf("duplicate collapse lost the first occurrence or explicit false: %#v", primary)
	}
	if stored.SubCalendars["work"].Enabled != nil {
		t.Fatalf("absent enabled was not preserved: %#v", stored.SubCalendars["work"])
	}

	// A refresh without an enabled choice keeps false and drops work.
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, []CalendarSubCalendar{
		{SubCalendarID: "primary", Name: "Renamed"},
	}); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.SubCalendars) != 1 {
		t.Fatalf("refresh did not remove the absent sub-calendar: %#v", stored.SubCalendars)
	}
	primary = stored.SubCalendars["primary"]
	if primary.Name != "Renamed" || primary.Enabled == nil || *primary.Enabled {
		t.Fatalf("refresh lost name or explicit false: %#v", primary)
	}

	// An empty incoming set removes every stored sub-calendar.
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, account.CalendarKey, nil); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetCalendarAccountByKey(ctx, identity.ID, account.CalendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.SubCalendars) != 0 {
		t.Fatalf("empty sync did not delete every sub-calendar: %#v", stored.SubCalendars)
	}

	// A missing connection is pgx.ErrNoRows even when the set is empty.
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, "missing@example.com_google", nil); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing connection sync error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.SyncCalendarSubCalendars(ctx, "507f1f77bcf86cd799439011", account.CalendarKey, nil); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("non-canonical owner sync error = %v, want pgx.ErrNoRows", err)
	}
}

// TestCalendarPreferencesAbsentVersusPresent proves an absent preference is
// distinct from a present empty one and that the explicit write path replaces
// the complete preference state.
func TestCalendarPreferencesAbsentVersusPresent(t *testing.T) {
	ctx, repo, _ := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)
	externalUserID := identity.ID

	if _, err := repo.GetCalendarPreferences(ctx, externalUserID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing preferences error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.UpsertCalendarPreferences(ctx, externalUserID, &CalendarPreferences{}); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetCalendarPreferences(ctx, externalUserID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PrimaryAccountKey != nil || stored.TokenOrigin != nil || stored.CalendarOptions != nil {
		t.Fatalf("absent preferences were not preserved: %#v", stored)
	}

	primary := "ada@example.com_google"
	origin := "ios"
	options := []byte(`{"bufferTime":{"enabled":true,"time":15}}`)
	if err := repo.UpsertCalendarPreferences(ctx, externalUserID, &CalendarPreferences{
		PrimaryAccountKey: &primary,
		TokenOrigin:       &origin,
		CalendarOptions:   options,
	}); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetCalendarPreferences(ctx, externalUserID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PrimaryAccountKey == nil || *stored.PrimaryAccountKey != primary {
		t.Fatalf("primary account key not stored: %#v", stored.PrimaryAccountKey)
	}
	if stored.TokenOrigin == nil || *stored.TokenOrigin != origin {
		t.Fatalf("token origin not stored: %#v", stored.TokenOrigin)
	}
	var storedOptions, expectedOptions any
	if err := json.Unmarshal(stored.CalendarOptions, &storedOptions); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(options, &expectedOptions); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(storedOptions, expectedOptions) {
		t.Fatalf("calendar options not stored: %s", stored.CalendarOptions)
	}

	// The explicit write path clears a field the caller omits.
	if err := repo.UpsertCalendarPreferences(ctx, externalUserID, &CalendarPreferences{}); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetCalendarPreferences(ctx, externalUserID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PrimaryAccountKey != nil || stored.TokenOrigin != nil || stored.CalendarOptions != nil {
		t.Fatalf("explicit empty preferences did not clear state: %#v", stored)
	}
	if err := repo.DeleteCalendarPreferences(ctx, externalUserID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetCalendarPreferences(ctx, externalUserID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("deleted preferences error = %v, want pgx.ErrNoRows", err)
	}
}

// TestCalendarRepositoryRequiresExistingOwner proves a calendar record never
// creates an account: an owner with no platform identity is an error.
func TestCalendarRepositoryRequiresExistingOwner(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", calendarTestEncryptionKey)
	ctx, repo, _ := newCalendarTestRepository(t)
	identity := &PlatformIdentity{ID: models.NewUUID().String()}
	account := &CalendarAccount{CalendarKey: "ghost@example.com_google", CalendarType: CalendarTypeGoogle, Email: "ghost@example.com"}
	if err := repo.UpsertCalendarAccount(ctx, identity.ID, account); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner error = %v, want pgx.ErrNoRows", err)
	}
	if _, err := repo.ListCalendarAccountsForUser(ctx, identity.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner list error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, &CalendarAccount{
		CalendarKey:  "ghost@example.com_google",
		CalendarType: CalendarTypeGoogle,
		OAuth2:       &CalendarOAuth2Credentials{AccessToken: "token"},
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner create error = %v, want pgx.ErrNoRows", err)
	}
	// Statements whose identity liveness is merged into the statement keep
	// reporting pgx.ErrNoRows for a missing owner.
	if err := repo.SetCalendarAccountEnabled(ctx, identity.ID, "ghost@example.com_google", true); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner toggle error = %v, want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetCalendarAccountByKey(ctx, identity.ID, "ghost@example.com_google"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner read error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.UpdateCalendarOAuthAccessToken(ctx, identity.ID, "ghost@example.com_google", "token", time.Now()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner token update error = %v, want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetCalendarPreferences(ctx, identity.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner preferences read error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.UpsertCalendarPreferences(ctx, identity.ID, &CalendarPreferences{}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner preferences write error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, "ghost@example.com_google", []CalendarSubCalendar{{SubCalendarID: "primary"}}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner sub-calendar write error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.DeleteCalendarAccount(ctx, identity.ID, "ghost@example.com_google"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner delete error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.DeleteCalendarPreferences(ctx, identity.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing owner preferences delete error = %v, want pgx.ErrNoRows", err)
	}
	// The canonical UUID guard rejects non-canonical identifiers before any
	// statement reaches the uuid columns.
	if err := repo.UpsertCalendarPreferences(ctx, "507f1f77bcf86cd799439011", &CalendarPreferences{}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("non-canonical preferences write error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.SetCalendarAccountEnabled(ctx, "507f1f77bcf86cd799439011", "ghost@example.com_google", true); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("non-canonical toggle error = %v, want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetCalendarAccountByKey(ctx, "507f1f77bcf86cd799439011", "ghost@example.com_google"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("non-canonical read error = %v, want pgx.ErrNoRows", err)
	}
}

// TestCalendarAccountRepositoryListAndDelete proves connections are owned and
// listed independently and that deletion cascades credentials and sub-calendars
// while leaving another connection untouched.
func TestCalendarAccountRepositoryListAndDelete(t *testing.T) {
	ctx, repo, tx := newCalendarTestRepository(t)
	identity := seedCalendarOwner(t, ctx, repo)
	first := &CalendarAccount{CalendarKey: "ada@example.com_google", CalendarType: CalendarTypeGoogle, Email: "ada@example.com"}
	second := &CalendarAccount{CalendarKey: "ada@example.com_outlook", CalendarType: CalendarTypeOutlook, Email: "ada@example.com"}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, first); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateCalendarAccount(ctx, identity.ID, second); err != nil {
		t.Fatal(err)
	}
	if err := repo.SyncCalendarSubCalendars(ctx, identity.ID, first.CalendarKey, []CalendarSubCalendar{{SubCalendarID: "primary"}}); err != nil {
		t.Fatal(err)
	}

	accounts, err := repo.ListCalendarAccountsForUser(ctx, identity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 {
		t.Fatalf("unexpected connection count: %d", len(accounts))
	}
	if len(accounts[0].SubCalendars) != 1 {
		t.Fatalf("listed connection lost sub-calendars: %#v", accounts[0].SubCalendars)
	}

	if err := repo.DeleteCalendarAccount(ctx, identity.ID, first.CalendarKey); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetCalendarAccountByKey(ctx, identity.ID, first.CalendarKey); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("deleted connection lookup error = %v, want pgx.ErrNoRows", err)
	}
	var subs, credentials int
	if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM calendar_sub_calendars WHERE calendar_account_id = $1), (SELECT count(*) FROM calendar_account_credentials WHERE calendar_account_id = $1)`, first.ID).Scan(&subs, &credentials); err != nil {
		t.Fatal(err)
	}
	if subs != 0 || credentials != 0 {
		t.Fatalf("deletion did not cascade: sub-calendars=%d credentials=%d", subs, credentials)
	}
	remaining, err := repo.ListCalendarAccountsForUser(ctx, identity.ID)
	if err != nil || len(remaining) != 1 || remaining[0].CalendarKey != second.CalendarKey {
		t.Fatalf("unexpected remaining connections: %#v, %v", remaining, err)
	}
	// Deleting a missing connection is a no-op.
	if err := repo.DeleteCalendarAccount(ctx, identity.ID, "missing_key"); err != nil {
		t.Fatalf("repeated deletion should be a no-op: %v", err)
	}
}
