package postgres

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

const eventNameConstraintMigration = "20260913000002_postgres_events_name_length.sql"

// TestPostgresEventsNameLengthConstraint proves the incremental migration adds
// the FR-119 storage guard: event names must be non-empty and at most 100
// Unicode code points. The NOT VALID constraint leaves legacy rows in place
// while new writes are checked, so it exists unvalidated.
func TestPostgresEventsNameLengthConstraint(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)

	var validated bool
	var definition string
	if err := tx.QueryRow(ctx, `SELECT convalidated, pg_get_constraintdef(oid) FROM pg_constraint
WHERE conrelid = 'postgres_events'::regclass AND conname = 'postgres_events_name_length'`).Scan(&validated, &definition); err != nil {
		t.Fatalf("read postgres_events_name_length: %v", err)
	}
	if validated {
		t.Fatal("expected postgres_events_name_length to be NOT VALID")
	}
	if !strings.Contains(definition, "name <> ''") || !strings.Contains(definition, "char_length(name) <= 100") {
		t.Fatalf("unexpected constraint definition: %s", definition)
	}

	accepted := []struct {
		name  string
		value string
	}{
		{name: "100 ascii code points", value: strings.Repeat("a", 100)},
		{name: "100 non-ascii code points", value: strings.Repeat("é", 100)},
	}
	for _, testCase := range accepted {
		t.Run("accepts "+testCase.name, func(t *testing.T) {
			if _, err := tx.Exec(ctx, `INSERT INTO postgres_events (short_id, name, type) VALUES ($1, $2, 'signup')`, signupTestShortID(t), testCase.value); err != nil {
				t.Fatalf("insert %s name: %v", testCase.name, err)
			}
		})
	}

	rejected := []struct {
		name  string
		value string
	}{
		{name: "empty name", value: ""},
		{name: "101 ascii code points", value: strings.Repeat("a", 101)},
		{name: "101 non-ascii code points", value: strings.Repeat("é", 101)},
	}
	for _, testCase := range rejected {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			err := expectSavepointError(t, ctx, tx, func() error {
				_, err := tx.Exec(ctx, `INSERT INTO postgres_events (short_id, name, type) VALUES ($1, $2, 'signup')`, signupTestShortID(t), testCase.value)
				return err
			})
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "23514" || postgresError.ConstraintName != "postgres_events_name_length" {
				t.Fatalf("expected postgres_events_name_length check violation, got %v", err)
			}
		})
	}
}

// TestPostgresEventsNameLengthConstraintDown proves the migration is
// reversible: the Down section drops the guard and allows over-limit names
// again.
func TestPostgresEventsNameLengthConstraintDown(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	applyMigrationDown(t, ctx, tx, eventNameConstraintMigration)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
	SELECT 1 FROM pg_constraint
	WHERE conrelid = 'postgres_events'::regclass AND conname = 'postgres_events_name_length')`).Scan(&exists); err != nil {
		t.Fatalf("read postgres_events_name_length existence: %v", err)
	}
	if exists {
		t.Fatal("expected the Down migration to drop postgres_events_name_length")
	}
	if _, err := tx.Exec(ctx, `INSERT INTO postgres_events (short_id, name, type) VALUES ($1, $2, 'signup')`, signupTestShortID(t), strings.Repeat("a", 101)); err != nil {
		t.Fatalf("insert without the constraint: %v", err)
	}
}
