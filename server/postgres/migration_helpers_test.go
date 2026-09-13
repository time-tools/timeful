package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// newMigrationTestTransaction opens a transaction against the isolated test
// database. Tests that drive the migration chain themselves use this directly;
// otherwise prefer newMigrationTestRepository.
func newMigrationTestTransaction(t *testing.T) (context.Context, pgx.Tx) {
	t.Helper()
	uri := os.Getenv("POSTGRES_APPLICATION_URI")
	if uri == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required")
	}
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Database != "timeful-test" && !strings.HasPrefix(config.ConnConfig.Database, "timeful-test-") {
		t.Fatal("requires an isolated test database")
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return ctx, tx
}

// applyMigrationData applies one goose migration's Up section into the
// transaction-scoped temporary tables. Temp tables shadow the real schema so
// the isolated tests never mutate test-stack records.
func applyMigrationData(t *testing.T, ctx context.Context, tx pgx.Tx, name string, data []byte) {
	t.Helper()
	up := strings.Split(string(data), "-- +goose Down")[0]
	up = strings.ReplaceAll(up, "CREATE TABLE ", "CREATE TEMP TABLE ")
	if _, err := tx.Exec(ctx, up); err != nil {
		t.Fatalf("apply %s: %v", name, err)
	}
}

// applyMigration applies one named migration from server/migrations into the
// transaction-scoped temporary tables.
func applyMigration(t *testing.T, ctx context.Context, tx pgx.Tx, name string) {
	t.Helper()
	data, err := os.ReadFile("../migrations/" + name)
	if err != nil {
		t.Fatal(err)
	}
	applyMigrationData(t, ctx, tx, name, data)
}

// applyMigrationDown applies one goose migration's Down section into the
// transaction-scoped temporary tables so rollback behavior is testable.
func applyMigrationDown(t *testing.T, ctx context.Context, tx pgx.Tx, name string) {
	t.Helper()
	data, err := os.ReadFile("../migrations/" + name)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(data), "-- +goose Down", 2)
	if len(parts) != 2 {
		t.Fatalf("migration %s has no Down section", name)
	}
	if _, err := tx.Exec(ctx, parts[1]); err != nil {
		t.Fatalf("apply down %s: %v", name, err)
	}
}

// newMigrationTestRepository applies every goose migration in server/migrations
// in version order into a transaction-scoped set of temporary tables. Temp
// tables shadow the real schema so the isolated tests never mutate test-stack
// records. Reading the migrations directory keeps the harness on the baseline
// schema and automatically covers migrations added after it.
func newMigrationTestRepository(t *testing.T) (context.Context, *Repository, pgx.Tx) {
	t.Helper()
	ctx, tx := newMigrationTestTransaction(t)
	entries, err := os.ReadDir("../migrations")
	if err != nil {
		t.Fatal(err)
	}
	applied := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}
		data, err := os.ReadFile("../migrations/" + name)
		if err != nil {
			t.Fatal(err)
		}
		applyMigrationData(t, ctx, tx, name, data)
		applied++
	}
	if applied == 0 {
		t.Fatal("no migrations found")
	}
	return ctx, &Repository{db: tx}, tx
}
