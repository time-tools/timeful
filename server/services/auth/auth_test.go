package auth

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
	"timeful/server/utils"
)

type refreshTokenRoundTrip func(*http.Request) (*http.Response, error)

func (fn refreshTokenRoundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

// TestRefreshUserTokenIfNecessaryPersistsToPostgres proves that a refreshed
// OAuth2 access token and its expiry are written to the PostgreSQL credential
// row, that the unrelated refresh token and scope survive, and that the
// in-memory user carries the fresh access token.
func TestRefreshUserTokenIfNecessaryPersistsToPostgres(t *testing.T) {
	uri := os.Getenv("POSTGRES_APPLICATION_URI")
	if uri == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required for token refresh persistence tests")
	}
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Database != "timeful-test" && !strings.HasPrefix(config.ConnConfig.Database, "timeful-test-") {
		t.Fatal("token refresh persistence tests require an isolated test database")
	}
	t.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")

	previousPool := pgstore.Pool
	if previousPool == nil {
		closePool := pgstore.Init()
		t.Cleanup(closePool)
	} else {
		t.Cleanup(func() { pgstore.Pool = previousPool })
	}

	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	identity, err := repository.CreatePlatformIdentity(ctx)
	if err != nil {
		t.Fatal(err)
	}
	platformIdentityID := identity.ID
	email := "refresh-" + platformIdentityID + "@example.com"
	calendarKey := utils.GetCalendarAccountKey(email, models.GoogleCalendarType)
	expiredAt := time.Now().Add(-time.Hour)
	stored := &pgstore.CalendarAccount{
		CalendarKey:  calendarKey,
		CalendarType: pgstore.CalendarTypeGoogle,
		Email:        email,
		OAuth2: &pgstore.CalendarOAuth2Credentials{
			AccessToken:          "expired-access-token",
			RefreshToken:         "refresh-token",
			Scope:                "calendar.readonly",
			AccessTokenExpiresAt: &expiredAt,
		},
	}
	if err := repository.UpsertCalendarAccount(ctx, platformIdentityID, stored); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup := context.Background()
		_ = repository.DeleteCalendarAccount(cleanup, platformIdentityID, calendarKey)
		_, _ = pgstore.Pool.Exec(cleanup, `DELETE FROM platform_identities WHERE id = $1`, platformIdentityID)
	})

	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	http.DefaultTransport = refreshTokenRoundTrip(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"access_token":"fresh-access-token","expires_in":3600,"scope":"calendar.readonly","token_type":"Bearer"}`)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})

	user := &models.User{
		Id: models.UUID(platformIdentityID),
		CalendarAccounts: map[string]models.CalendarAccount{
			calendarKey: {
				CalendarType: models.GoogleCalendarType,
				Email:        email,
				OAuth2CalendarAuth: &models.OAuth2CalendarAuth{
					AccessToken:           "expired-access-token",
					RefreshToken:          "refresh-token",
					Scope:                 "calendar.readonly",
					AccessTokenExpireDate: models.NewDateTimeFromTime(expiredAt),
				},
			},
		},
	}

	RefreshUserTokenIfNecessary(user, nil)

	if got := user.CalendarAccounts[calendarKey].OAuth2CalendarAuth.AccessToken; got != "fresh-access-token" {
		t.Fatalf("in-memory access token = %q, want the refreshed token", got)
	}
	reloaded, err := repository.GetCalendarAccountByKey(ctx, platformIdentityID, calendarKey)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.OAuth2 == nil {
		t.Fatal("refreshed credential row is missing")
	}
	if reloaded.OAuth2.AccessToken != "fresh-access-token" {
		t.Fatalf("persisted access token = %q, want the refreshed token", reloaded.OAuth2.AccessToken)
	}
	if reloaded.OAuth2.RefreshToken != "refresh-token" {
		t.Fatalf("refresh token was not preserved: %q", reloaded.OAuth2.RefreshToken)
	}
	if reloaded.OAuth2.Scope != "calendar.readonly" {
		t.Fatalf("scope was not preserved: %q", reloaded.OAuth2.Scope)
	}
	if reloaded.OAuth2.AccessTokenExpiresAt == nil || !reloaded.OAuth2.AccessTokenExpiresAt.After(time.Now()) {
		t.Fatalf("persisted expiry was not advanced: %#v", reloaded.OAuth2.AccessTokenExpiresAt)
	}
}
