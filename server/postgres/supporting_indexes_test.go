package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// indexDefinition resolves one index in the test transaction's temp schema and
// returns its definition, or an empty string when the index does not exist.
// Resolving through pg_my_temp_schema keeps the assertions on the temp schema
// under test instead of falling back to the migrated real schema.
func indexDefinition(t *testing.T, ctx context.Context, tx pgx.Tx, name string) string {
	t.Helper()
	var definition string
	err := tx.QueryRow(ctx, `SELECT COALESCE(pg_get_indexdef(c.oid), '')
FROM pg_class c
WHERE c.relname = $1 AND c.relnamespace = pg_my_temp_schema()`, name).Scan(&definition)
	if errors.Is(err, pgx.ErrNoRows) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return definition
}

// TestSupportingIndexesSchema proves the supporting-indexes migration created
// each index with the columns its query shape needs, shaped the analytics index
// as a partial index over non-empty creator attribution, and retired the two
// indexes those shapes cannot use.
func TestSupportingIndexesSchema(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)

	for _, expected := range []struct {
		name     string
		fragment string
	}{
		{"postgres_event_responses_event_created_idx", "(event_id, created_at, id)"},
		{"event_attendees_event_created_idx", "(event_id, created_at, id)"},
		{"event_attendees_event_email_lower_idx", "(event_id, lower(email))"},
		{"postgres_events_creator_created_at_idx", "(created_at, creator_posthog_id)"},
		{"daily_user_log_members_log_position_idx", "(daily_user_log_id, first_seen_position, id)"},
	} {
		definition := indexDefinition(t, ctx, tx, expected.name)
		if definition == "" {
			t.Fatalf("%s is missing", expected.name)
		}
		if !strings.Contains(definition, expected.fragment) {
			t.Fatalf("%s definition = %q, want columns %q", expected.name, definition, expected.fragment)
		}
	}

	for _, retired := range []string{
		"postgres_event_responses_event_id_idx",
		"postgres_events_active_creator_posthog_id_idx",
	} {
		if definition := indexDefinition(t, ctx, tx, retired); definition != "" {
			t.Fatalf("retired index %s survived: %q", retired, definition)
		}
	}

	analytics := indexDefinition(t, ctx, tx, "postgres_events_creator_created_at_idx")
	if !strings.Contains(analytics, "WHERE") || !strings.Contains(analytics, "creator_posthog_id IS NOT NULL") || !strings.Contains(analytics, "creator_posthog_id <> ''") {
		t.Fatalf("analytics index is not partial on non-empty creator_posthog_id: %q", analytics)
	}

	// The down migration restores the two retired indexes and removes the new
	// ones, so the temp-table harness can apply either direction.
	applyMigrationDown(t, ctx, tx, "20260913000001_supporting_indexes.sql")
	for _, restored := range []string{
		"postgres_event_responses_event_id_idx",
		"postgres_events_active_creator_posthog_id_idx",
	} {
		if definition := indexDefinition(t, ctx, tx, restored); definition == "" {
			t.Fatalf("down migration did not restore %s", restored)
		}
	}
	for _, removed := range []string{
		"postgres_event_responses_event_created_idx",
		"event_attendees_event_created_idx",
		"event_attendees_event_email_lower_idx",
		"postgres_events_creator_created_at_idx",
		"daily_user_log_members_log_position_idx",
	} {
		if definition := indexDefinition(t, ctx, tx, removed); definition != "" {
			t.Fatalf("down migration left %s: %q", removed, definition)
		}
	}
}

// explainPlan returns the EXPLAIN (ANALYZE, BUFFERS) output for one query
// against the test transaction's temp schema.
func explainPlan(t *testing.T, ctx context.Context, tx pgx.Tx, query string, args ...any) string {
	t.Helper()
	rows, err := tx.Query(ctx, `EXPLAIN (ANALYZE, BUFFERS) `+query, args...)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	defer rows.Close()
	lines := []string{}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return strings.Join(lines, "\n")
}

// assertPlanIndex fails the test with the full plan when the expected index is
// absent, so a planner change is diagnosable instead of a bare mismatch.
func assertPlanIndex(t *testing.T, plan, indexName string) {
	t.Helper()
	if !strings.Contains(plan, indexName) {
		t.Fatalf("plan does not use %s:\n%s", indexName, plan)
	}
}

// TestSupportingIndexesServeQueryShapes seeds representative rows, forces index
// plans with enable_seqscan=off, and proves the supporting indexes are usable
// for the analytics, response-listing, attendee-listing, daily-log, and
// transfer-lock query shapes.
func TestSupportingIndexesServeQueryShapes(t *testing.T) {
	ctx, repo, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	for i := 0; i < 3; i++ {
		response := &Response{EventID: eventID, EventVisitorIdentityID: visitorID, RespondentKind: RespondentKindGuest}
		if err := repo.CreateResponse(ctx, response); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_attendees (event_id, email)
VALUES ($1, 'ada@example.com'), ($1, 'grace@example.com')`, eventID); err != nil {
		t.Fatal(err)
	}

	var platformIdentityID, secondPlatformIdentityID, logID string
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&platformIdentityID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&secondPlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO daily_user_logs (log_date) VALUES (CURRENT_DATE) RETURNING id`).Scan(&logID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO daily_user_log_members (daily_user_log_id, platform_identity_id, first_seen_position)
VALUES ($1, $2, 0), ($1, $3, 1)`, logID, platformIdentityID, secondPlatformIdentityID); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		creator := "creator-" + string(rune('a'+i))
		insertAnalyticsEvent(t, ctx, tx, &creator, now.Add(-time.Duration(i)*time.Hour), false)
	}

	var transferID string
	hash := sha256.Sum256([]byte("transfer index"))
	if err := tx.QueryRow(ctx, `INSERT INTO access_transfers (event_id, source_hash, platform_identity_id)
VALUES ($1, $2, $3) RETURNING id`, eventID, hash[:], platformIdentityID).Scan(&transferID); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.Exec(ctx, `SET LOCAL enable_seqscan = off`); err != nil {
		t.Fatal(err)
	}
	// Bitmap scans can otherwise beat the ordered index on a tiny temp table
	// and hide which index actually serves the ORDER BY.
	if _, err := tx.Exec(ctx, `SET LOCAL enable_bitmapscan = off`); err != nil {
		t.Fatal(err)
	}

	distinctAnalytics := explainPlan(t, ctx, tx, countDistinctMonthlyActiveEventCreatorsByDayQuery, []time.Time{now}, monthlyActiveCreatorLookback)
	assertPlanIndex(t, distinctAnalytics, "postgres_events_creator_created_at_idx")

	groupedAnalytics := explainPlan(t, ctx, tx, countDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDayQuery, []time.Time{now}, monthlyActiveCreatorLookback, 1)
	assertPlanIndex(t, groupedAnalytics, "postgres_events_creator_created_at_idx")

	responseListing := explainPlan(t, ctx, tx, listResponsesQuery, eventID)
	assertPlanIndex(t, responseListing, "postgres_event_responses_event_created_idx")

	attendeeListing := explainPlan(t, ctx, tx, listAttendeesQuery, eventID)
	assertPlanIndex(t, attendeeListing, "event_attendees_event_created_idx")

	dashboardMembership := explainPlan(t, ctx, tx, listDashboardEventsQuery, platformIdentityID, "ada@example.com")
	assertPlanIndex(t, dashboardMembership, "event_attendees_event_email_lower_idx")

	inviteeEmail := explainPlan(t, ctx, tx, nonDeclinedAttendeeEmailExistsQuery, eventID, "ADA@EXAMPLE.COM")
	assertPlanIndex(t, inviteeEmail, "event_attendees_event_email_lower_idx")

	dailyLogAppend := explainPlan(t, ctx, tx, recordDailyUserLogMembershipQuery, logID, platformIdentityID)
	assertPlanIndex(t, dailyLogAppend, "daily_user_log_members_log_position_idx")

	// The full active-user-days report joins members by log. Force the join
	// method so the member access path is the index under test rather than a
	// hash join over the tiny temp tables.
	if _, err := tx.Exec(ctx, `SET LOCAL enable_hashjoin = off`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `SET LOCAL enable_mergejoin = off`); err != nil {
		t.Fatal(err)
	}
	reportDate := now.Format("2006-01-02")
	activeUserDays := explainPlan(t, ctx, tx, listActiveUserDaysQuery, reportDate, reportDate)
	assertPlanIndex(t, activeUserDays, "daily_user_log_members_log_position_idx")

	transferLock := explainPlan(t, ctx, tx, lockAccessTransferQuery, eventID, transferID)
	assertPlanIndex(t, transferLock, "access_transfers_pkey")
}
