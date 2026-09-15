package accounts

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

func closedExistencePool(t *testing.T) *pgxpool.Pool {
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

func existenceAuthorityTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	uri := os.Getenv("POSTGRES_APPLICATION_URI")
	if uri == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required for the not-found case")
	}
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TestIsNewUserReportsPostgresError proves that a PostgreSQL failure is
// returned instead of reported as account-exists or account-missing.
func TestIsNewUserReportsPostgresError(t *testing.T) {
	previousPool := pgstore.Pool
	t.Cleanup(func() { pgstore.Pool = previousPool })
	pgstore.Pool = closedExistencePool(t)

	email := "error-" + models.NewUUID().String() + "@example.com"
	isNew, err := IsNewUser(email)
	if err == nil {
		t.Fatalf("IsNewUser() error = nil, want a PostgreSQL failure; isNew = %v", isNew)
	}
}

// TestIsNewUserReportsAuthority proves that existence is reported from the
// authoritative account data alone.
func TestIsNewUserReportsAuthority(t *testing.T) {
	pool := existenceAuthorityTestPool(t)
	previousPool := pgstore.Pool
	pgstore.Pool = pool
	t.Cleanup(func() { pgstore.Pool = previousPool })

	repository, err := pgstore.DefaultRepository()
	if err != nil {
		t.Fatal(err)
	}

	freshEmail := "fresh-" + models.NewUUID().String() + "@example.com"
	if isNew, err := IsNewUser(freshEmail); err != nil || !isNew {
		t.Fatalf("IsNewUser(fresh) = %v, %v; want true, nil", isNew, err)
	}

	existingEmail := "existing-" + models.NewUUID().String() + "@example.com"
	if _, _, err := repository.FindOrCreateAccountByEmail(context.Background(), existingEmail, pgstore.Account{Email: existingEmail}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteAccountsByEmail(t, pool, existingEmail) })
	if isNew, err := IsNewUser(existingEmail); err != nil || isNew {
		t.Fatalf("IsNewUser(existing) = %v, %v; want false, nil", isNew, err)
	}
}

// TestIsNewUserEmptyEmail proves that an empty email is reported as new without
// consulting the database.
func TestIsNewUserEmptyEmail(t *testing.T) {
	if isNew, err := IsNewUser("   "); err != nil || !isNew {
		t.Fatalf("IsNewUser(empty) = %v, %v; want true, nil", isNew, err)
	}
}

// deleteAccountsByEmail removes every account and platform identity for an
// email so concurrent contract tests stay rerunnable against a retained
// database.
func deleteAccountsByEmail(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	ctx := context.Background()
	rows, err := pool.Query(ctx, `SELECT a.platform_identity_id FROM accounts a WHERE lower(a.email) = lower($1)`, email)
	if err != nil {
		t.Errorf("list account identities: %v", err)
		return
	}
	var platformIdentityIDs []string
	for rows.Next() {
		var platformIdentityID string
		if err := rows.Scan(&platformIdentityID); err != nil {
			t.Errorf("scan account identity: %v", err)
		}
		platformIdentityIDs = append(platformIdentityIDs, platformIdentityID)
	}
	rows.Close()
	if _, err := pool.Exec(ctx, `DELETE FROM accounts WHERE lower(email) = lower($1)`, email); err != nil {
		t.Errorf("delete accounts by email: %v", err)
	}
	if len(platformIdentityIDs) > 0 {
		if _, err := pool.Exec(ctx, `DELETE FROM platform_identities WHERE id = ANY($1)`, platformIdentityIDs); err != nil {
			t.Errorf("delete platform identities: %v", err)
		}
	}
}

// TestResolveForSignInConcurrentEmailCreatesSingleAccount proves that concurrent
// first-time sign-ins for one email resolve one account and create no
// duplicate account or platform identity.
func TestResolveForSignInConcurrentEmailCreatesSingleAccount(t *testing.T) {
	pool := existenceAuthorityTestPool(t)
	previousPool := pgstore.Pool
	pgstore.Pool = pool
	t.Cleanup(func() { pgstore.Pool = previousPool })

	email := "concurrent-signin-" + models.NewUUID().String() + "@example.com"
	t.Cleanup(func() { deleteAccountsByEmail(t, pool, email) })

	const workers = 8
	ctx := context.Background()
	results := make([]*pgstore.Account, workers)
	created := make([]bool, workers)
	failures := make([]error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], created[i], failures[i] = ResolveForSignIn(ctx, Profile{Email: email, FirstName: "Racer"})
		}(i)
	}
	close(start)
	wg.Wait()

	var winner *pgstore.Account
	creations := 0
	for i := range results {
		if failures[i] != nil {
			t.Fatalf("worker %d failed: %v", i, failures[i])
		}
		if results[i] == nil {
			t.Fatalf("worker %d returned no account", i)
		}
		if created[i] {
			creations++
		}
		if winner == nil {
			winner = results[i]
		} else if results[i].ID != winner.ID {
			t.Fatalf("concurrent sign-ins resolved different accounts: %s vs %s", winner.ID, results[i].ID)
		}
	}
	if creations != 1 {
		t.Fatalf("exactly one sign-in should report creation, got %d", creations)
	}
	var accounts, identities int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM accounts WHERE lower(email) = lower($1)), (SELECT count(*) FROM platform_identities WHERE id = $2)`, email, winner.PlatformIdentityID).Scan(&accounts, &identities); err != nil {
		t.Fatal(err)
	}
	if accounts != 1 || identities != 1 {
		t.Fatalf("concurrent sign-ins created accounts=%d identities=%d", accounts, identities)
	}
}
