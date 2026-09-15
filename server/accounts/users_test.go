package accounts

import (
	"context"
	"io"
	"testing"

	"timeful/server/logger"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// TestUserLookupsReturnAuthoritativeProfile proves that both account lookups
// return the authoritative profile, including the usage counter, and never any
// calendar integration fields.
func TestUserLookupsReturnAuthoritativeProfile(t *testing.T) {
	pool := existenceAuthorityTestPool(t)
	previousPool := pgstore.Pool
	pgstore.Pool = pool
	t.Cleanup(func() { pgstore.Pool = previousPool })

	ctx := context.Background()
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Fatal(err)
	}

	email := "lookup-" + models.NewUUID().String() + "@example.com"
	t.Cleanup(func() { deleteAccountsByEmail(t, pool, email) })
	account, created, err := repository.FindOrCreateAccountByEmail(ctx, email, pgstore.Account{
		Email:            email,
		FirstName:        "Ada",
		LastName:         "Lovelace",
		Picture:          "https://example.test/picture.png",
		TimezoneOffset:   90,
		NumEventsCreated: 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected the lookup test to create a fresh account")
	}

	assertProfile := func(lookup string, got *models.User) {
		t.Helper()
		if got == nil {
			t.Fatalf("%s returned no user", lookup)
		}
		if got.Id.String() != account.PlatformIdentityID || got.Email != email {
			t.Fatalf("%s returned the wrong account: %#v", lookup, got)
		}
		if got.FirstName != "Ada" || got.LastName != "Lovelace" {
			t.Fatalf("%s returned the wrong name: %#v", lookup, got)
		}
		if got.Picture != "https://example.test/picture.png" {
			t.Fatalf("%s returned the wrong picture: %q", lookup, got.Picture)
		}
		if got.TimezoneOffset != 90 {
			t.Fatalf("%s returned the wrong timezone offset: %d", lookup, got.TimezoneOffset)
		}
		if got.NumEventsCreated != 7 {
			t.Fatalf("%s returned the wrong usage counter: %d", lookup, got.NumEventsCreated)
		}
		if got.CalendarAccounts != nil || got.CalendarOptions != nil || got.PrimaryAccountKey != nil || got.TokenOrigin != "" {
			t.Fatalf("%s returned calendar integration fields: %#v", lookup, got)
		}
	}

	assertProfile("UserByEmail", UserByEmail(email))
	assertProfile("UserByPlatformIdentityID", UserByPlatformIdentityID(account.PlatformIdentityID))
}

// TestUserLookupsPostgresErrorReturnsNil proves that a PostgreSQL lookup failure
// never falls back to an inferred account profile.
func TestUserLookupsPostgresErrorReturnsNil(t *testing.T) {
	logger.Init(io.Discard)
	previousPool := pgstore.Pool
	t.Cleanup(func() { pgstore.Pool = previousPool })
	pgstore.Pool = closedExistencePool(t)

	platformIdentityID := models.NewUUID().String()
	email := "error-" + platformIdentityID + "@example.com"

	if got := UserByPlatformIdentityID(platformIdentityID); got != nil {
		t.Fatalf("PostgreSQL error must yield no account by identifier, got %#v", got)
	}
	if got := UserByEmail(email); got != nil {
		t.Fatalf("PostgreSQL error must yield no account by email, got %#v", got)
	}
}

// TestUserLookupsUnknownAccountReturnsNil proves that a genuine not-found
// returns no account instead of an inferred profile.
func TestUserLookupsUnknownAccountReturnsNil(t *testing.T) {
	pool := existenceAuthorityTestPool(t)
	previousPool := pgstore.Pool
	pgstore.Pool = pool
	t.Cleanup(func() { pgstore.Pool = previousPool })

	platformIdentityID := models.NewUUID().String()
	email := "unknown-" + platformIdentityID + "@example.com"

	if got := UserByPlatformIdentityID(platformIdentityID); got != nil {
		t.Fatalf("genuine not-found must yield no account by identifier, got %#v", got)
	}
	if got := UserByEmail(email); got != nil {
		t.Fatalf("genuine not-found must yield no account by email, got %#v", got)
	}
}

// TestUserFromAccountNil proves that a missing account produces no user.
func TestUserFromAccountNil(t *testing.T) {
	if got := UserFromAccount(nil); got != nil {
		t.Fatalf("UserFromAccount(nil) = %#v, want nil", got)
	}
}
