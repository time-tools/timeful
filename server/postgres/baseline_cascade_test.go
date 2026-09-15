package postgres

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/jackc/pgx/v5"
)

// tempSchemaRowCount counts rows matching where in the temp-schema table under
// test and fails the test when the table cannot be read.
func tempSchemaRowCount(t *testing.T, ctx context.Context, tx pgx.Tx, table, where string, args ...any) int {
	t.Helper()
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE `+where, args...).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}

// TestBaselineEventHardDeleteCascades proves one event delete removes every
// dependent row class: visitor identities and their credentials, responses,
// signup blocks, signup responses and their block memberships, attendees,
// folder memberships, and access transfers with their requests. The
// account-owned folder itself survives because folders are not event-scoped.
func TestBaselineEventHardDeleteCascades(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	if _, err := tx.Exec(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash)
VALUES ($1, decode(repeat('ab',32),'hex'))`, visitorID); err != nil {
		t.Fatal(err)
	}
	var responseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind)
VALUES ($1, $2, 'guest') RETURNING id`, eventID, visitorID).Scan(&responseID); err != nil {
		t.Fatal(err)
	}
	blockID := seedSignupBlock(t, ctx, tx, eventID, "Block")
	var signupResponseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', 'Ada') RETURNING id`, eventID, visitorID).Scan(&signupResponseID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_signup_response_blocks (response_id, block_id, position)
VALUES ($1, $2, 0)`, signupResponseID, blockID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_attendees (event_id, email) VALUES ($1, 'ada@example.com')`, eventID); err != nil {
		t.Fatal(err)
	}

	var platformIdentityID string
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&platformIdentityID); err != nil {
		t.Fatal(err)
	}
	var folderID string
	if err := tx.QueryRow(ctx, `INSERT INTO folders (platform_identity_id, name) VALUES ($1, 'Cascade') RETURNING id`, platformIdentityID).Scan(&folderID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO folder_events (folder_id, event_id, platform_identity_id)
VALUES ($1, $2, $3)`, folderID, eventID, platformIdentityID); err != nil {
		t.Fatal(err)
	}

	hash := sha256.Sum256([]byte("cascade transfer"))
	var transferID string
	if err := tx.QueryRow(ctx, `INSERT INTO access_transfers (event_id, source_hash, platform_identity_id)
VALUES ($1, $2, $3) RETURNING id`, eventID, hash[:], platformIdentityID).Scan(&transferID); err != nil {
		t.Fatal(err)
	}
	var requestID string
	if err := tx.QueryRow(ctx, `INSERT INTO access_transfer_requests (transfer_id, target_hash, code)
VALUES ($1, $2, 'CASCADE1') RETURNING id`, transferID, hash[:]).Scan(&requestID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE access_transfers SET state = 'approved', approved_request_id = $2 WHERE id = $1`, transferID, requestID); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM events WHERE id = $1`, eventID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	for _, dependent := range []struct {
		table string
		where string
		args  []any
	}{
		{"event_visitor_identities", "id = $1", []any{visitorID}},
		{"event_visitor_credentials", "event_visitor_identity_id = $1", []any{visitorID}},
		{"event_responses", "id = $1", []any{responseID}},
		{"event_signup_blocks", "id = $1", []any{blockID}},
		{"event_signup_responses", "id = $1", []any{signupResponseID}},
		{"event_signup_response_blocks", "response_id = $1", []any{signupResponseID}},
		{"event_attendees", "event_id = $1", []any{eventID}},
		{"folder_events", "event_id = $1", []any{eventID}},
		{"access_transfers", "id = $1", []any{transferID}},
		{"access_transfer_requests", "transfer_id = $1", []any{transferID}},
	} {
		if count := tempSchemaRowCount(t, ctx, tx, dependent.table, dependent.where, dependent.args...); count != 0 {
			t.Fatalf("%s still holds %d dependent rows after the event delete", dependent.table, count)
		}
	}

	if count := tempSchemaRowCount(t, ctx, tx, "folders", "id = $1", folderID); count != 1 {
		t.Fatal("event deletion removed the account-owned folder")
	}
}
