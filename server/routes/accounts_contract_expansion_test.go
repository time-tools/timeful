package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"timeful/server/logger"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
	authservice "timeful/server/services/auth"
)

// accountContractRoundTrip lets a contract test serve every outbound HTTP call
// the sign-in and import paths make without reaching the network.
type accountContractRoundTrip func(*http.Request) (*http.Response, error)

func (fn accountContractRoundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func accountContractJSONResponse(t *testing.T, request *http.Request, body string) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    request,
	}
}

// installMutableGoogleOAuthTransport serves the Google token exchange, ID-token
// verification, and calendar-list calls used by OAuth sign-in. The profile
// pointer is read on every request so a test can change the provider claims
// between sign-ins.
func installMutableGoogleOAuthTransport(t *testing.T, profile *authservice.GoogleIdTokenInfo) {
	t.Helper()
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = accountContractRoundTrip(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.URL.Host == "oauth2.googleapis.com" && request.URL.Path == "/token":
			return accountContractJSONResponse(t, request,
				`{"access_token":"access","id_token":"id-token","expires_in":3600,"refresh_token":"refresh","scope":"scope","token_type":"Bearer"}`), nil
		case request.URL.Host == "oauth2.googleapis.com" && request.URL.Path == "/tokeninfo":
			encoded, err := json.Marshal(*profile)
			if err != nil {
				t.Fatal(err)
			}
			return accountContractJSONResponse(t, request, string(encoded)), nil
		case request.URL.Host == "www.googleapis.com" && strings.Contains(request.URL.Path, "calendarList"):
			return accountContractJSONResponse(t, request, `{"items":[]}`), nil
		default:
			// Requests to the in-process test server must still reach it.
			return previous.RoundTrip(request)
		}
	})
}

func decodeAccountBool(t *testing.T, data map[string]json.RawMessage, key string) bool {
	t.Helper()
	var value bool
	if err := json.Unmarshal(data[key], &value); err != nil {
		t.Fatalf("decode %s: %v", key, err)
	}
	return value
}

// cleanupOtpAccount removes an account created through OTP sign-in, so the
// suite stays rerunnable against a retained database.
func cleanupOtpAccount(t *testing.T, account *pgstore.Account) {
	t.Helper()
	t.Cleanup(func() {
		deleteAccountTestFixtures(t, account.PlatformIdentityID)
	})
}

func closedAccountContractPostgresPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig("postgres://timeful:timeful@127.0.0.1:1/timeful-test?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	pool.Close()
	return pool
}

// TestAccountProviderSignInAppliesNamePrecedence proves that OAuth provider
// sign-in applies the provider name and picture for a new account, and that a
// user's custom name is preserved on later sign-ins while the picture still
// refreshes.
func TestAccountProviderSignInAppliesNamePrecedence(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	t.Setenv("CLIENT_ID", "account-contract-client")

	email := "oauth-name-" + models.NewUUID().String() + "@example.com"
	profile := authservice.GoogleIdTokenInfo{
		Aud:        "account-contract-client",
		Iss:        "https://accounts.google.com",
		Email:      email,
		GivenName:  "Provider",
		FamilyName: "Provided",
		Picture:    "https://provider.example/first.png",
	}
	installMutableGoogleOAuthTransport(t, &profile)
	signInWithGoogle := func() {
		client.request(http.MethodPost, "/api/auth/sign-in", map[string]any{
			"code": "authorization-code", "scope": "scope",
			"calendarType": models.GoogleCalendarType, "timezoneOffset": 0,
		}, http.StatusOK)
	}

	repository := repositoryForTest(t)
	signInWithGoogle()
	account, err := repository.GetAccountByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("provider sign-in did not create an account: %v", err)
	}
	cleanupOtpAccount(t, account)

	if account.FirstName != "Provider" || account.LastName != "Provided" {
		t.Fatalf("provider name not applied: %#v", account)
	}
	if account.Picture != "https://provider.example/first.png" {
		t.Fatalf("provider picture not applied: %q", account.Picture)
	}
	if account.HasCustomName != nil && *account.HasCustomName {
		t.Fatalf("a provider sign-in must not mark the name custom: %#v", account.HasCustomName)
	}

	// A custom name is set by the visitor and must survive a later provider sign-in.
	client.request(http.MethodPatch, "/api/user/name", map[string]any{"firstName": "Custom", "lastName": "Person"}, http.StatusOK)
	profile.GivenName = "Changed"
	profile.FamilyName = "Changed"
	profile.Picture = "https://provider.example/second.png"
	signInWithGoogle()

	account, err = repository.GetAccountByPlatformIdentityID(context.Background(), account.PlatformIdentityID)
	if err != nil {
		t.Fatal(err)
	}
	if account.FirstName != "Custom" || account.LastName != "Person" {
		t.Fatalf("provider sign-in overwrote the custom name: %#v", account)
	}
	if account.HasCustomName == nil || !*account.HasCustomName {
		t.Fatalf("custom name flag was lost: %#v", account.HasCustomName)
	}
	if account.Picture != "https://provider.example/second.png" {
		t.Fatalf("provider picture was not refreshed: %q", account.Picture)
	}
}

// TestAccountExistenceCheckReportsExistenceStates proves that the existence
// check reports a brand-new email as new and an email with a PostgreSQL account
// as existing.
func TestAccountExistenceCheckReportsExistenceStates(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	repository := repositoryForTest(t)
	ctx := context.Background()

	newEmail := "existence-new-" + models.NewUUID().String() + "@example.com"
	if result := client.request(http.MethodPost, "/api/auth/otp/check-email", map[string]any{"email": newEmail}, http.StatusOK); !decodeAccountBool(t, result, "isNewUser") {
		t.Fatalf("a brand-new email must report isNewUser=true: %v", result)
	}

	existingEmail := "existence-existing-" + models.NewUUID().String() + "@example.com"
	existing, _, err := repository.FindOrCreateAccountByEmail(ctx, existingEmail, pgstore.Account{Email: existingEmail})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, existing.PlatformIdentityID) })
	if result := client.request(http.MethodPost, "/api/auth/otp/check-email", map[string]any{"email": existingEmail}, http.StatusOK); decodeAccountBool(t, result, "isNewUser") {
		t.Fatalf("an existing PostgreSQL account must report isNewUser=false: %v", result)
	}
}

// TestAccountExistenceCheckFailsClosedOnPostgresError proves that a PostgreSQL
// lookup failure is reported as a server error, and never inferred as an
// existence result.
func TestAccountExistenceCheckFailsClosedOnPostgresError(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	// The error path logs, so ensure the package logger is initialized.
	logger.Init(io.Discard)

	email := "existence-error-" + models.NewUUID().String() + "@example.com"

	previousPool := pgstore.Pool
	pgstore.Pool = closedAccountContractPostgresPool(t)
	t.Cleanup(func() { pgstore.Pool = previousPool })

	client.request(http.MethodPost, "/api/auth/otp/check-email", map[string]any{"email": email}, http.StatusInternalServerError)
}

// TestAccountIntegrationWritesPreservePostgresProfile proves that calendar add,
// toggle, calendar-options, and remove all write only the PostgreSQL calendar
// store and never change the PostgreSQL profile.
func TestAccountIntegrationWritesPreservePostgresProfile(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()

	email := "integration-" + models.NewUUID().String() + "@example.com"
	verifyOtpSignIn(t, client, email, "123456")
	repository := repositoryForTest(t)
	account, err := repository.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	cleanupOtpAccount(t, account)
	baseline := *account
	label := "Integration-" + models.NewUUID().String()
	calendarKey := label + "_ics"

	assertProfileUnchanged := func(step string) {
		t.Helper()
		stored, err := repository.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID)
		if err != nil {
			t.Fatalf("%s: %v", step, err)
		}
		if *stored != baseline {
			t.Fatalf("%s changed the PostgreSQL profile:\nbefore %#v\nafter  %#v", step, baseline, *stored)
		}
	}

	// Add: the connection and its credential are written to PostgreSQL.
	client.request(http.MethodPost, "/api/user/add-ics-calendar-account", map[string]any{
		"feedUrl": "https://example.com/feed.ics", "label": label,
	}, http.StatusOK)
	stored, err := repository.GetCalendarAccountByKey(ctx, account.PlatformIdentityID, calendarKey)
	if err != nil {
		t.Fatalf("ICS calendar connection was not written to PostgreSQL: %v", err)
	}
	if stored.CalendarType != pgstore.CalendarTypeICS || stored.Email != label {
		t.Fatalf("stored ICS connection = %#v", stored)
	}
	if stored.ICS == nil || stored.ICS.FeedURL != "https://example.com/feed.ics" {
		t.Fatalf("stored ICS feed credential = %#v", stored.ICS)
	}
	assertProfileUnchanged("calendar add")

	// Toggle: only the connection's enabled flag changes.
	client.request(http.MethodPost, "/api/user/toggle-calendar", map[string]any{
		"email": label, "calendarType": models.ICSCalendarType, "enabled": false,
	}, http.StatusOK)
	stored, err = repository.GetCalendarAccountByKey(ctx, account.PlatformIdentityID, calendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Enabled == nil || *stored.Enabled {
		t.Fatalf("toggle did not disable the PostgreSQL connection: %#v", stored.Enabled)
	}
	assertProfileUnchanged("calendar toggle")

	// Calendar options: written to the PostgreSQL preference row only.
	client.request(http.MethodPatch, "/api/user/calendar-options", map[string]any{
		"bufferTime":   map[string]any{"enabled": true, "time": 30},
		"workingHours": map[string]any{"enabled": true, "startTime": 8, "endTime": 18},
	}, http.StatusOK)
	preferences, err := repository.GetCalendarPreferences(ctx, account.PlatformIdentityID)
	if err != nil {
		t.Fatalf("calendar options were not written to PostgreSQL: %v", err)
	}
	if len(preferences.CalendarOptions) == 0 {
		t.Fatal("calendar options preference is empty")
	}
	var options models.CalendarOptions
	if err := json.Unmarshal(preferences.CalendarOptions, &options); err != nil {
		t.Fatal(err)
	}
	if !options.BufferTime.Enabled || options.BufferTime.Time != 30 {
		t.Fatalf("calendar options were not persisted: %#v", options)
	}
	assertProfileUnchanged("calendar options")

	// Remove: the connection and its sub-calendars are deleted from PostgreSQL.
	client.request(http.MethodDelete, "/api/user/remove-calendar-account", map[string]any{
		"email": label, "calendarType": models.ICSCalendarType,
	}, http.StatusOK)
	if _, err := repository.GetCalendarAccountByKey(ctx, account.PlatformIdentityID, calendarKey); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("remove did not delete the PostgreSQL connection: %v", err)
	}
	assertProfileUnchanged("calendar remove")
}

// newAccountEventContractRouter additionally registers the event routes so the
// usage-counter increments on event creation and import can be exercised.
func newAccountEventContractRouter(t *testing.T) *gin.Engine {
	t.Helper()
	initRoutesReadFiltersTestDB(t)
	if os.Getenv("POSTGRES_APPLICATION_URI") == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required for account route contracts")
	}
	anonymousEventPostgresOnce.Do(func() { pgstore.Init() })
	t.Setenv("LISTMONK_ENABLED", "false")

	router := gin.New()
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("session", store))
	apiRouter := router.Group("/api")
	InitAuth(apiRouter)
	InitUser(apiRouter)
	InitUsers(apiRouter)
	InitEvents(apiRouter)
	InitFolders(apiRouter)
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

// installRemoteEventFetchTransport is removed with the legacy import endpoint.

// TestAccountUsageCounterTracksCreatedEvents proves that creating an event
// increments the PostgreSQL usage counter and that the profile reports that
// authoritative counter.
func TestAccountUsageCounterTracksCreatedEvents(t *testing.T) {
	router := newAccountEventContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()

	email := "usage-counter-" + models.NewUUID().String() + "@example.com"
	verifyOtpSignIn(t, client, email, "123456")
	repository := repositoryForTest(t)
	account, err := repository.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	cleanupOtpAccount(t, account)
	t.Cleanup(func() {
		if pgstore.Pool != nil {
			_, _ = pgstore.Pool.Exec(context.Background(), `DELETE FROM events WHERE owner_platform_identity_id = $1`, account.PlatformIdentityID)
		}
	})

	assertCounter := func(step string, want int) {
		t.Helper()
		stored, err := repository.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID)
		if err != nil {
			t.Fatal(err)
		}
		if stored.NumEventsCreated != want {
			t.Fatalf("%s: usage counter = %d, want %d", step, stored.NumEventsCreated, want)
		}
		profile := client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
		if got := decodeAccountInt(t, profile, "numEventsCreated"); got != want {
			t.Fatalf("%s: profile usage counter = %d, want the PostgreSQL counter %d", step, got, want)
		}
	}

	client.request(http.MethodPost, "/api/events", map[string]any{
		"name": "Counter created event", "type": string(models.SPECIFIC_DATES),
		"daysOnly": true, "dates": []string{"2026-08-11T00:00:00Z"},
	}, http.StatusCreated)
	assertCounter("event creation", 1)
}
