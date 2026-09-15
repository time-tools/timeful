package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"timeful/server/accounts"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

func newAccountContractRouter(t *testing.T) *gin.Engine {
	t.Helper()
	initRoutesReadFiltersTestDB(t)
	if os.Getenv("POSTGRES_APPLICATION_URI") == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required for account route contracts")
	}
	routeTestDBOnce.Do(func() { pgstore.Init() })
	t.Setenv("LISTMONK_ENABLED", "false")

	router := gin.New()
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("session", store))
	apiRouter := router.Group("/api")
	InitAuth(apiRouter)
	InitUser(apiRouter)
	InitUsers(apiRouter)

	// Test-only session seeding must be registered before the router starts
	// serving so the route table is immutable while requests are handled.
	router.POST("/test/account-contract/sign-in/:id", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("userId", c.Param("id"))
		if err := session.Save(); err != nil {
			t.Error(err)
		}
		c.JSON(http.StatusOK, gin.H{})
	})
	return router
}

type accountContractClient struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
}

func newAccountContractClient(t *testing.T, router *gin.Engine) *accountContractClient {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return &accountContractClient{t: t, server: server, client: &http.Client{Jar: jar}}
}

func (c *accountContractClient) request(method, path string, body any, status int) map[string]json.RawMessage {
	c.t.Helper()
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	req, err := http.NewRequest(method, c.server.URL+path, bytes.NewReader(data))
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(response.Body)
	if response.StatusCode != status {
		c.t.Fatalf("%s %s: got %d want %d: %s", method, path, response.StatusCode, status, raw)
	}
	result := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil {
			c.t.Fatal(err)
		}
	}
	return result
}

func decodeAccountString(t *testing.T, data map[string]json.RawMessage, key string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(data[key], &value); err != nil {
		t.Fatalf("decode %s: %v", key, err)
	}
	return value
}

func decodeAccountInt(t *testing.T, data map[string]json.RawMessage, key string) int {
	t.Helper()
	var value int
	if err := json.Unmarshal(data[key], &value); err != nil {
		t.Fatalf("decode %s: %v", key, err)
	}
	return value
}

func insertOtpCode(t *testing.T, email, code string) {
	t.Helper()
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := repository.CreateOtpChallenge(ctx, email, code, time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = repository.DeleteOtpChallenge(context.Background(), email)
	})
}

// deleteAccountTestFixtures removes every created account and its
// platform identity. The repository deliberately retains platform identities in
// production, so tests that create them must clean them up explicitly to stay
// rerunnable against a retained database.
func deleteAccountTestFixtures(t *testing.T, platformIdentityIDs ...string) {
	t.Helper()
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Errorf("resolve repository for account cleanup: %v", err)
		return
	}
	ctx := context.Background()
	for _, platformIdentityID := range platformIdentityIDs {
		if err := repository.DeleteAccountByPlatformIdentityID(ctx, platformIdentityID); err != nil {
			t.Errorf("delete account %s: %v", platformIdentityID, err)
		}
		if _, err := pgstore.Pool.Exec(ctx, `DELETE FROM platform_identities WHERE id = $1`, platformIdentityID); err != nil {
			t.Errorf("delete platform identity %s: %v", platformIdentityID, err)
		}
		if _, err := pgstore.Pool.Exec(ctx, `DELETE FROM account_deletion_tombstones WHERE platform_identity_id = $1`, platformIdentityID); err != nil {
			t.Errorf("delete account tombstone %s: %v", platformIdentityID, err)
		}
	}
}

func verifyOtpSignIn(t *testing.T, client *accountContractClient, email, code string) map[string]json.RawMessage {
	t.Helper()
	insertOtpCode(t, email, code)
	return client.request(http.MethodPost, "/api/auth/otp/verify", map[string]any{
		"email": email, "code": code, "timezoneOffset": 0,
		"firstName": "Provider", "lastName": "Name",
	}, http.StatusOK)
}

// TestAccountOtpSignInUsesAuthority proves that OTP sign-in resolves a
// account and that profile reads and updates only ever touch the
// authoritative account.
func TestAccountOtpSignInUsesAuthority(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	email := "account-contract-" + models.NewUUID().String() + "@example.com"

	profile := verifyOtpSignIn(t, client, email, "123456")
	if got := decodeAccountString(t, profile, "email"); got != email {
		t.Fatalf("sign-in profile email = %q, want %q", got, email)
	}

	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Fatal(err)
	}
	account, err := repository.GetAccountByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("account not created: %v", err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })

	read := client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if got := decodeAccountString(t, read, "email"); got != email {
		t.Fatalf("profile email = %q, want the stored value", got)
	}
	if got := decodeAccountString(t, read, "firstName"); got != "Provider" {
		t.Fatalf("profile firstName = %q, want the stored value", got)
	}

	client.request(http.MethodPatch, "/api/user/name", map[string]any{"firstName": "Custom", "lastName": "Person"}, http.StatusOK)
	updated, err := repository.GetAccountByPlatformIdentityID(context.Background(), account.PlatformIdentityID)
	if err != nil || updated.FirstName != "Custom" || updated.LastName != "Person" || updated.HasCustomName == nil || !*updated.HasCustomName {
		t.Fatalf("profile update not written: %v %#v", err, updated)
	}

	// Public profiles also resolve the authoritative account.
	public := client.request(http.MethodGet, "/api/users/"+account.PlatformIdentityID, nil, http.StatusOK)
	if got := decodeAccountString(t, public, "firstName"); got != "Custom" {
		t.Fatalf("public profile firstName = %q, want the stored value", got)
	}

	// Repeated sign-in must not create a duplicate account or identity.
	verifyOtpSignIn(t, client, email, "654321")
	accounts, err := repository.GetAccountByEmail(context.Background(), email)
	if err != nil || accounts.ID != account.ID {
		t.Fatalf("repeated sign-in changed the account: %v %#v", err, accounts)
	}
	var accountCount, identityCount int
	if err := pgstore.Pool.QueryRow(context.Background(), `SELECT (SELECT count(*) FROM accounts a JOIN platform_identities p ON p.id = a.platform_identity_id WHERE p.id = $1), (SELECT count(*) FROM platform_identities WHERE id = $1)`, account.PlatformIdentityID).Scan(&accountCount, &identityCount); err != nil {
		t.Fatal(err)
	}
	if accountCount != 1 || identityCount != 1 {
		t.Fatalf("duplicate account=%d identity=%d", accountCount, identityCount)
	}

	// Sign-out clears the session and signed-in access.
	client.request(http.MethodPost, "/api/auth/sign-out", nil, http.StatusOK)
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusUnauthorized)
}

// TestAccountExistingSessionResolvesAuthority proves that a session
// resolves its authoritative account, that calendar state
// is exposed through the boundary, and that a session for an account absent
// from the database is rejected.
func TestAccountExistingSessionResolvesAuthority(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)

	email := "session-" + models.NewUUID().String() + "@example.com"
	primaryKey := email + "_google"
	repository := repositoryForTest(t)
	account, _, err := repository.FindOrCreateAccountByEmail(context.Background(), email, pgstore.Account{
		Email:          email,
		FirstName:      "Existing",
		LastName:       "User",
		TimezoneOffset: 120,
	})
	if err != nil {
		t.Fatal(err)
	}
	externalUserID := account.PlatformIdentityID
	t.Cleanup(func() { deleteAccountTestFixtures(t, externalUserID) })

	client.request(http.MethodPost, "/test/account-contract/sign-in/"+externalUserID, nil, http.StatusOK)
	profile := client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if got := decodeAccountString(t, profile, "firstName"); got != "Existing" {
		t.Fatalf("session profile firstName = %q", got)
	}
	if got := decodeAccountString(t, profile, "email"); got != email {
		t.Fatalf("session profile email = %q", got)
	}
	var calendarAccounts map[string]json.RawMessage
	if err := json.Unmarshal(profile["calendarAccounts"], &calendarAccounts); err != nil || len(calendarAccounts) != 0 {
		t.Fatalf("empty calendar state exposed entries: %v %v", err, calendarAccounts)
	}

	// Calendar state is authoritative and is exposed through the
	// boundary.
	if err := accounts.SaveCalendarAccount(context.Background(), externalUserID, primaryKey, models.CalendarAccount{
		CalendarType: models.GoogleCalendarType,
		Email:        email,
	}); err != nil {
		t.Fatal(err)
	}
	profile = client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if err := json.Unmarshal(profile["calendarAccounts"], &calendarAccounts); err != nil || len(calendarAccounts) != 1 {
		t.Fatalf("calendar integration not exposed through boundary: %v %v", err, calendarAccounts)
	}

	// A repeated signed-in read must not create a second account.
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	var accountCount int
	if err := pgstore.Pool.QueryRow(context.Background(), `SELECT count(*) FROM accounts WHERE platform_identity_id = $1`, externalUserID).Scan(&accountCount); err != nil {
		t.Fatal(err)
	}
	if accountCount != 1 {
		t.Fatalf("session resolution duplicated the account: %d", accountCount)
	}

	// A session whose account has no account row is not adopted from any
	// retained document and is rejected.
	unknownClient := newAccountContractClient(t, router)
	unknownExternalUserID := models.NewUUID().String()
	unknownClient.request(http.MethodPost, "/test/account-contract/sign-in/"+unknownExternalUserID, nil, http.StatusOK)
	unknownClient.request(http.MethodGet, "/api/user/profile", nil, http.StatusUnauthorized)

	// The retired 24-character external identifier has no compatibility lookup:
	// a session carrying one resolves to no account and must re-authenticate.
	legacyClient := newAccountContractClient(t, router)
	legacyClient.request(http.MethodPost, "/test/account-contract/sign-in/507f1f77bcf86cd799439011", nil, http.StatusOK)
	legacyClient.request(http.MethodGet, "/api/user/profile", nil, http.StatusUnauthorized)
}

// TestAccountDuplicateEmailDoesNotMerge proves that two distinct
// accounts with equal emails remain distinct and that sign-in resolves the
// deterministic oldest match without creating a third account.
func TestAccountDuplicateEmailDoesNotMerge(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)

	email := "duplicate-" + models.NewUUID().String() + "@example.com"
	repository := repositoryForTest(t)
	older, err := repository.CreateAccount(context.Background(), pgstore.Account{Email: email, FirstName: "Older"})
	if err != nil {
		t.Fatal(err)
	}
	newer, err := repository.CreateAccount(context.Background(), pgstore.Account{Email: email, FirstName: "Newer"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		deleteAccountTestFixtures(t, older.PlatformIdentityID, newer.PlatformIdentityID)
	})

	profile := verifyOtpSignIn(t, client, email, "123456")
	if got := decodeAccountString(t, profile, "firstName"); got != "Older" {
		t.Fatalf("sign-in firstName = %q, want the deterministic oldest account", got)
	}
	var count int
	if err := pgstore.Pool.QueryRow(context.Background(), `SELECT count(*) FROM accounts WHERE lower(email) = lower($1)`, email).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("equal-email accounts must stay distinct, got %d", count)
	}
}

// TestAccountCalendarRemovalWritesIntegrationOnly proves that removing a
// calendar account removes only the calendar connection and leaves
// the profile unchanged.
func TestAccountCalendarRemovalWritesIntegrationOnly(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)

	email := "calendar-removal-" + models.NewUUID().String() + "@example.com"
	primaryKey := email + "_google"
	repository := repositoryForTest(t)
	account, _, err := repository.FindOrCreateAccountByEmail(context.Background(), email, pgstore.Account{
		Email:     email,
		FirstName: "Calendar",
		LastName:  "Owner",
	})
	if err != nil {
		t.Fatal(err)
	}
	externalUserID := account.PlatformIdentityID
	t.Cleanup(func() { deleteAccountTestFixtures(t, externalUserID) })

	client.request(http.MethodPost, "/test/account-contract/sign-in/"+externalUserID, nil, http.StatusOK)
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	account, err = repository.GetAccountByPlatformIdentityID(context.Background(), externalUserID)
	if err != nil {
		t.Fatalf("session did not resolve an account: %v", err)
	}
	if err := repository.IncrementAccountEventsCreated(context.Background(), account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if err := accounts.SaveCalendarAccount(context.Background(), externalUserID, primaryKey, models.CalendarAccount{
		CalendarType: models.GoogleCalendarType,
		Email:        email,
	}); err != nil {
		t.Fatal(err)
	}

	client.request(http.MethodDelete, "/api/user/remove-calendar-account", map[string]any{
		"email": email, "calendarType": models.GoogleCalendarType,
	}, http.StatusOK)

	// The calendar connection is gone.
	remaining, err := repository.ListCalendarAccountsForUser(context.Background(), externalUserID)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Fatalf("calendar account survived removal: %#v", remaining)
	}

	stored, err := repository.GetAccountByPlatformIdentityID(context.Background(), externalUserID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Email != email || stored.FirstName != "Calendar" || stored.NumEventsCreated != 1 {
		t.Fatalf("calendar removal changed the profile: %#v", stored)
	}
}

// TestAccountProfileCounterIsAuthoritative proves that the profile
// reports the usage counter and that a profile update cannot change
// it.
func TestAccountProfileCounterIsAuthoritative(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	email := "counter-" + models.NewUUID().String() + "@example.com"

	verifyOtpSignIn(t, client, email, "123456")
	repository := repositoryForTest(t)
	account, err := repository.GetAccountByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("account not created: %v", err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })

	for i := 0; i < 2; i++ {
		if err := repository.IncrementAccountEventsCreated(context.Background(), account.PlatformIdentityID); err != nil {
			t.Fatal(err)
		}
	}
	read := client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if got := decodeAccountInt(t, read, "numEventsCreated"); got != 2 {
		t.Fatalf("profile numEventsCreated = %d, want the stored counter 2", got)
	}

	client.request(http.MethodPatch, "/api/user/name", map[string]any{"firstName": "Custom", "lastName": "Person"}, http.StatusOK)
	read = client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if got := decodeAccountInt(t, read, "numEventsCreated"); got != 2 {
		t.Fatalf("profile update changed numEventsCreated to %d", got)
	}
}

// TestAccountProfileReadRecordsDailyUserLog proves the sign-in profile path
// records the account in the authoritative daily log and that
// repeated same-day reads are idempotent.
func TestAccountProfileReadRecordsDailyUserLog(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()
	email := "daily-log-" + models.NewUUID().String() + "@example.com"

	verifyOtpSignIn(t, client, email, "123456")
	repository := repositoryForTest(t)
	account, err := repository.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })

	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)

	var memberships int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT count(*) FROM daily_user_log_members WHERE platform_identity_id = $1`, account.PlatformIdentityID).Scan(&memberships); err != nil {
		t.Fatal(err)
	}
	if memberships != 1 {
		t.Fatalf("same-day profile reads recorded %d memberships, want 1", memberships)
	}
	var logs int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT count(*) FROM daily_user_logs l JOIN daily_user_log_members m ON m.daily_user_log_id = l.id WHERE m.platform_identity_id = $1`, account.PlatformIdentityID).Scan(&logs); err != nil {
		t.Fatal(err)
	}
	if logs != 1 {
		t.Fatalf("account appears in %d daily logs, want 1", logs)
	}
}

func repositoryForTest(t *testing.T) *pgstore.Repository {
	t.Helper()
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

// newSessionAccount mints an account and returns its platform identity UUID so
// the test-only session sign-in route resolves it as an authenticated account.
// The account is removed when the test finishes.
func newSessionAccount(t *testing.T) string {
	t.Helper()
	account, err := repositoryForTest(t).CreateAccount(context.Background(), pgstore.Account{})
	if err != nil {
		t.Fatalf("create session account fixture: %v", err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })
	return account.PlatformIdentityID
}

func accountObjectID(t *testing.T, value string) models.UUID {
	t.Helper()
	objectID, ok := models.ParseUUID(value)
	if !ok {
		t.Fatalf("invalid account identifier %q", value)
	}
	return objectID
}
