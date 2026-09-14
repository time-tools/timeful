package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

// TestBaselineVisitorIdentityAndOwnerConstraints proves the baseline schema
// enforces the visitor identity and event owner relations the runtime depends
// on: every response belongs to a visitor identity of its own event, an event
// owner relation is event-scoped, and a base credential can never carry owner
// powers. It also proves the legacy response columns retained for the prior
// release's rollback window keep accepting writes.
func TestBaselineVisitorIdentityAndOwnerConstraints(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	otherEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	// A response must carry a visitor identity scoped to the same event.
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO postgres_event_responses (event_id, event_visitor_identity_id, respondent_kind) VALUES ($1, $2, 'guest')`, otherEventID, visitorID)
		return err
	})
	var platformIdentityID string
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&platformIdentityID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO postgres_event_responses (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, guest_edit_token, payload)
VALUES ($1, $2, 'account', $3, 'compat-token', '{}'::jsonb)`, eventID, visitorID, platformIdentityID); err != nil {
		t.Fatalf("compatibility response insert: %v", err)
	}

	// The owner relation must reference a visitor identity of the same event.
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `UPDATE postgres_events SET owner_event_visitor_identity_id = $2 WHERE id = $1`, otherEventID, visitorID)
		return err
	})
	if _, err := tx.Exec(ctx, `UPDATE postgres_events SET owner_event_visitor_identity_id = $2 WHERE id = $1`, eventID, visitorID); err != nil {
		t.Fatalf("owner association: %v", err)
	}

	// A base credential can never grant owner powers; a granted credential can.
	if _, err := tx.Exec(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash)
VALUES ($1, decode(repeat('ab',32),'hex'))`, visitorID); err != nil {
		t.Fatal(err)
	}
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `UPDATE event_visitor_credentials SET grants_owner = true`)
		return err
	})
	if _, err := tx.Exec(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash, kind, grants_owner)
VALUES ($1, decode(repeat('cd',32),'hex'), 'granted', true)`, visitorID); err != nil {
		t.Fatalf("granted credential with owner powers: %v", err)
	}
}

// TestBaselineConsolidatesAccountIdentity proves the baseline creates the
// consolidated account identity shape directly: every account reference is a
// platform_identity_id uuid and none of the retired legacy columns exist.
func TestBaselineConsolidatesAccountIdentity(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)

	for _, column := range []struct{ table, column string }{
		{"platform_identities", "external_user_id"},
		{"postgres_events", "owner_external_id"},
		{"postgres_event_responses", "account_user_id"},
		{"event_signup_responses", "account_user_id"},
		{"event_attendees", "account_user_id"},
		{"folders", "account_user_id"},
		{"folder_events", "account_user_id"},
		{"access_transfers", "external_user_id"},
		{"daily_user_log_members", "account_user_id"},
		{"account_deletion_tombstones", "external_user_id"},
	} {
		if hasColumn(t, ctx, tx, column.table, column.column) {
			t.Fatalf("%s.%s should not exist in the consolidated baseline", column.table, column.column)
		}
	}

	for _, column := range []struct{ table, column string }{
		{"postgres_events", "owner_platform_identity_id"},
		{"postgres_event_responses", "platform_identity_id"},
		{"event_signup_responses", "platform_identity_id"},
		{"event_attendees", "platform_identity_id"},
		{"folders", "platform_identity_id"},
		{"folder_events", "platform_identity_id"},
		{"access_transfers", "platform_identity_id"},
		{"daily_user_log_members", "platform_identity_id"},
		{"account_deletion_tombstones", "platform_identity_id"},
	} {
		if !hasColumn(t, ctx, tx, column.table, column.column) {
			t.Fatalf("%s.%s should exist in the consolidated baseline", column.table, column.column)
		}
	}
}

// hasColumn reports whether a table in the transaction's temp schema has the
// named column. A missing table fails the test so a negative column assertion
// cannot pass vacuously.
func hasColumn(t *testing.T, ctx context.Context, tx pgx.Tx, table, column string) bool {
	t.Helper()
	var tableExists, columnExists bool
	if err := tx.QueryRow(ctx, `SELECT
        to_regclass($1) IS NOT NULL,
        EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = to_regclass($1) AND attname = $2 AND attnum > 0 AND NOT attisdropped
        )`, table, column).Scan(&tableExists, &columnExists); err != nil {
		t.Fatal(err)
	}
	if !tableExists {
		t.Fatalf("table %s does not exist", table)
	}
	return columnExists
}
