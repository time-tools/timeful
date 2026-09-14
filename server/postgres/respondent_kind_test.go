package postgres

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

const respondentKindNotNullMigration = "20260915000000_postgres_event_responses_respondent_kind_not_null.sql"

// TestPostgresEventResponsesRespondentKindNotNull proves the incremental
// migration makes respondent_kind explicitly NOT NULL: every writer sets
// 'account' or 'guest', so the retired NULL coercion has no shape to accept.
func TestPostgresEventResponsesRespondentKindNotNull(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)

	var nullable bool
	if err := tx.QueryRow(ctx, `SELECT NOT attnotnull FROM pg_attribute
WHERE attrelid = 'postgres_event_responses'::regclass AND attname = 'respondent_kind'`).Scan(&nullable); err != nil {
		t.Fatalf("read respondent_kind nullability: %v", err)
	}
	if nullable {
		t.Fatal("expected respondent_kind to be NOT NULL")
	}

	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	nullInsert := expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO postgres_event_responses (event_id, event_visitor_identity_id) VALUES ($1, $2)`, eventID, visitorID)
		return err
	})
	var nullViolation *pgconn.PgError
	if !errors.As(nullInsert, &nullViolation) || nullViolation.Code != "23502" {
		t.Fatalf("expected a not-null violation for a NULL respondent_kind, got %v", nullInsert)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO postgres_event_responses (event_id, event_visitor_identity_id, respondent_kind)
VALUES ($1, $2, 'guest')`, eventID, visitorID); err != nil {
		t.Fatalf("insert guest response: %v", err)
	}

	invalid := expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO postgres_event_responses (event_id, event_visitor_identity_id, respondent_kind)
VALUES ($1, $2, 'spectator')`, eventID, visitorID)
		return err
	})
	var kindViolation *pgconn.PgError
	if !errors.As(invalid, &kindViolation) || kindViolation.Code != "23514" || kindViolation.ConstraintName != "postgres_event_responses_kind" {
		t.Fatalf("expected postgres_event_responses_kind check violation, got %v", invalid)
	}
}

// TestPostgresEventResponsesRespondentKindNotNullDown proves the migration is
// reversible: the Down section restores the nullable column the baseline
// declared.
func TestPostgresEventResponsesRespondentKindNotNullDown(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	applyMigrationDown(t, ctx, tx, respondentKindNotNullMigration)

	var nullable bool
	if err := tx.QueryRow(ctx, `SELECT NOT attnotnull FROM pg_attribute
WHERE attrelid = 'postgres_event_responses'::regclass AND attname = 'respondent_kind'`).Scan(&nullable); err != nil {
		t.Fatalf("read respondent_kind nullability: %v", err)
	}
	if !nullable {
		t.Fatal("expected the Down migration to restore NULL acceptance")
	}

	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	if _, err := tx.Exec(ctx, `INSERT INTO postgres_event_responses (event_id, event_visitor_identity_id) VALUES ($1, $2)`, eventID, visitorID); err != nil {
		t.Fatalf("insert NULL respondent_kind after Down: %v", err)
	}
}

// TestResponseWriterLeavesRetainedCompatibilityColumnsNull proves the active
// writer path no longer touches the legacy guest columns that are retained only
// for the prior release's rollback window.
func TestResponseWriterLeavesRetainedCompatibilityColumnsNull(t *testing.T) {
	ctx, repo, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	response := &Response{
		EventID:                eventID,
		EventVisitorIdentityID: visitorID,
		RespondentKind:         RespondentKindGuest,
		Payload:                json.RawMessage(`{"name":"Ada"}`),
	}
	if err := repo.CreateResponse(ctx, response); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateResponse(ctx, response); err != nil {
		t.Fatal(err)
	}

	var guestID, canonicalName, editPolicy, ownershipMode, editToken *string
	if err := tx.QueryRow(ctx, `SELECT guest_id, canonical_guest_name, guest_edit_policy, guest_ownership_mode, guest_edit_token
FROM postgres_event_responses WHERE id = $1`, response.ID).Scan(&guestID, &canonicalName, &editPolicy, &ownershipMode, &editToken); err != nil {
		t.Fatal(err)
	}
	if guestID != nil || canonicalName != nil || editPolicy != nil || ownershipMode != nil || editToken != nil {
		t.Fatalf("retained compatibility columns were written: guest_id=%v canonical_guest_name=%v guest_edit_policy=%v guest_ownership_mode=%v guest_edit_token=%v",
			guestID, canonicalName, editPolicy, ownershipMode, editToken)
	}
}
