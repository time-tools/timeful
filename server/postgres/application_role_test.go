package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// TestApplicationRolePrivilegeSeparation proves the isolated stack's
// application role performs DML but cannot run DDL, so an unreplaced CREATE
// TABLE in the migration replay fails loudly instead of mutating the
// goose-applied schema. The row changes happen in the test transaction and are
// rolled back with it.
func TestApplicationRolePrivilegeSeparation(t *testing.T) {
	ctx, tx := newMigrationTestTransaction(t)

	var role string
	var superuser bool
	if err := tx.QueryRow(ctx, `SELECT current_user, rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&role, &superuser); err != nil {
		t.Fatal(err)
	}
	if superuser {
		t.Fatalf("application role %s is a superuser; privilege separation is not asserted", role)
	}
	var canCreate bool
	if err := tx.QueryRow(ctx, `SELECT has_schema_privilege(current_user, 'public', 'CREATE')`).Scan(&canCreate); err != nil {
		t.Fatal(err)
	}
	if canCreate {
		t.Fatalf("application role %s can create objects in the public schema", role)
	}

	var identityID string
	if err := tx.QueryRow(ctx, `INSERT INTO public.platform_identities DEFAULT VALUES RETURNING id`).Scan(&identityID); err != nil {
		t.Fatalf("application role INSERT: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE public.platform_identities SET created_at = created_at WHERE id = $1`, identityID); err != nil {
		t.Fatalf("application role UPDATE: %v", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM public.platform_identities WHERE id = $1`, identityID); err != nil {
		t.Fatalf("application role DELETE: %v", err)
	}

	for _, statement := range []struct {
		name string
		sql  string
	}{
		{"CREATE TABLE", `CREATE TABLE public.schema_integrity_ddl_probe (id integer)`},
		{"ALTER TABLE", `ALTER TABLE public.events ADD COLUMN schema_integrity_ddl_probe integer`},
	} {
		statement := statement
		t.Run(statement.name+" is denied", func(t *testing.T) {
			err := expectSavepointError(t, ctx, tx, func() error {
				_, err := tx.Exec(ctx, statement.sql)
				return err
			})
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
				t.Fatalf("%s = %v, want SQLSTATE 42501 (insufficient_privilege)", statement.name, err)
			}
		})
	}

	var probeExists bool
	if err := tx.QueryRow(ctx, `SELECT to_regclass('public.schema_integrity_ddl_probe') IS NOT NULL`).Scan(&probeExists); err != nil {
		t.Fatal(err)
	}
	if probeExists {
		t.Fatal("DDL probe table exists after the denied CREATE TABLE")
	}
}
