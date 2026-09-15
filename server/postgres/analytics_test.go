package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// insertAnalyticsEvent inserts one PostgreSQL event row with only the fields the
// creator aggregations read, so tests exercise the query window and creator
// filter directly against authoritative event storage.
func insertAnalyticsEvent(t *testing.T, ctx context.Context, tx pgx.Tx, creatorPosthogID *string, createdAt time.Time, isDeleted bool) {
	t.Helper()
	shortID, err := GenerateShortID()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO events
 (short_id, name, type, is_deleted, creator_posthog_id, created_at, updated_at, payload)
VALUES ($1, 'analytics', 'specific_dates', $2, $3, $4, $4, '{}'::jsonb)`,
		shortID, isDeleted, creatorPosthogID, createdAt); err != nil {
		t.Fatalf("insert analytics event: %v", err)
	}
}

func analyticsCreator(value string) *string { return &value }

// TestCountDistinctMonthlyActiveEventCreatorsByDayReadsPerDayWindows proves the
// day-spine report returns one ascending count per reporting day, counts each
// creator once over that day's half-open [day-30d, day) window, excludes absent
// and empty attribution, keeps soft-deleted events, and returns zero for a day
// whose window holds no events.
func TestCountDistinctMonthlyActiveEventCreatorsByDayReadsPerDayWindows(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	dayOne := time.Date(2026, 1, 10, 23, 59, 59, 0, time.UTC)
	dayGap := time.Date(2026, 3, 1, 23, 59, 59, 0, time.UTC)
	dayTwo := time.Date(2026, 4, 10, 23, 59, 59, 0, time.UTC)
	dayThree := time.Date(2026, 4, 11, 23, 59, 59, 0, time.UTC)

	// creator-a is active only on dayOne: it falls before dayTwo's lower bound.
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-a"), dayOne.AddDate(0, 0, -1), false)
	// creator-b has two events inside dayTwo's and dayThree's windows.
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-b"), dayTwo.AddDate(0, 0, -1), false)
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-b"), dayTwo.AddDate(0, 0, -2), false)
	// Exactly dayThree's lower bound is included; it is also inside dayTwo's window.
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-c"), dayThree.AddDate(0, 0, -30), false)
	// Exactly dayThree's reporting instant is excluded by the half-open bound.
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-d"), dayThree, false)
	// Absent and empty attribution never count.
	insertAnalyticsEvent(t, ctx, tx, nil, dayTwo.AddDate(0, 0, -3), false)
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator(""), dayTwo.AddDate(0, 0, -3), false)
	// Soft-deleted events were counted by the legacy aggregation.
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-e"), dayTwo.AddDate(0, 0, -4), true)

	got, err := repo.CountDistinctMonthlyActiveEventCreatorsByDay(ctx, []time.Time{dayOne, dayGap, dayTwo, dayThree})
	if err != nil {
		t.Fatal(err)
	}
	// dayOne: creator-a. dayGap: none. dayTwo: creator-b, creator-c, creator-e.
	// dayThree: creator-b, creator-c, creator-e.
	want := []int64{1, 0, 3, 3}
	if len(got) != len(want) {
		t.Fatalf("day-spine counts = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("day %d count = %d, want %d (all: %v)", i, got[i], want[i], got)
		}
	}
}

// TestCountDistinctMonthlyActiveEventCreatorsByDayUsesClientOffsetBoundary
// proves the report instant is honored exactly as supplied: an event on the
// previous second counts while events exactly at the 23:59:59 client-offset
// instant and just before the lower bound do not.
func TestCountDistinctMonthlyActiveEventCreatorsByDayUsesClientOffsetBoundary(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	location := time.FixedZone("UserOffset", -7*60*60)
	dayEnd := time.Date(2026, 9, 10, 23, 59, 59, 0, location)
	if dayEnd.UTC().Format(time.RFC3339) != "2026-09-11T06:59:59Z" {
		t.Fatalf("client-offset day end = %s, want 2026-09-11T06:59:59Z", dayEnd.UTC())
	}
	lowerBound := dayEnd.AddDate(0, 0, -monthlyActiveCreatorLookback)

	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("inclusive-second"), dayEnd.Add(-time.Second), false)
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("inclusive-lower"), lowerBound, false)
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("excluded-lower"), lowerBound.Add(-time.Second), false)
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("excluded-upper"), dayEnd, false)

	got, err := repo.CountDistinctMonthlyActiveEventCreatorsByDay(ctx, []time.Time{dayEnd})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("client-offset boundary count = %v, want [2]", got)
	}
}

// TestCountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay proves the
// grouped report applies the same ascending day spine and per-day window with
// an inclusive threshold, keeping a zero entry for each day.
func TestCountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	dayOne := time.Date(2026, 9, 10, 23, 59, 59, 0, time.UTC)
	dayTwo := time.Date(2026, 10, 11, 23, 59, 59, 0, time.UTC)

	// creator-a has three events in dayTwo's window.
	for i := 0; i < 3; i++ {
		insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-a"), dayTwo.AddDate(0, 0, -(i+1)), false)
	}
	// creator-b has exactly two events before dayTwo's lower bound, so only
	// dayOne counts it.
	for i := 0; i < 2; i++ {
		insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-b"), dayOne.AddDate(0, 0, -(i+1)), false)
	}
	// creator-c has one event before dayTwo's lower bound.
	insertAnalyticsEvent(t, ctx, tx, analyticsCreator("creator-c"), dayOne.AddDate(0, 0, -1), false)

	cases := []struct {
		x    int
		want []int64
	}{
		{x: 4, want: []int64{0, 0}},
		{x: 3, want: []int64{0, 1}},
		{x: 2, want: []int64{1, 1}},
		{x: 1, want: []int64{2, 1}},
	}
	for _, tc := range cases {
		got, err := repo.CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(ctx, []time.Time{dayOne, dayTwo}, tc.x)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("x=%d counts = %v, want %v", tc.x, got, tc.want)
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Fatalf("x=%d day %d count = %d, want %d (all: %v)", tc.x, i, got[i], tc.want[i], got)
			}
		}
	}
}

// TestCreatorAnalyticsCountEachPostgresEventOnce proves that a creator with
// multiple events is counted once by the distinct report and that the grouped
// aggregation counts rows rather than distinct creators, so migrated and newly
// created PostgreSQL events each contribute exactly once.
func TestCreatorAnalyticsCountEachPostgresEventOnce(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	dayOne := time.Date(2026, 9, 10, 23, 59, 59, 0, time.UTC)
	dayTwo := time.Date(2026, 9, 11, 23, 59, 59, 0, time.UTC)

	migrated := analyticsCreator("creator-migrated")
	insertAnalyticsEvent(t, ctx, tx, migrated, dayOne.AddDate(0, 0, -20), false)
	insertAnalyticsEvent(t, ctx, tx, migrated, dayOne.AddDate(0, 0, -10), false)
	newEvent := analyticsCreator("creator-new")
	insertAnalyticsEvent(t, ctx, tx, newEvent, dayOne.AddDate(0, 0, -1), false)

	distinct, err := repo.CountDistinctMonthlyActiveEventCreatorsByDay(ctx, []time.Time{dayOne, dayTwo})
	if err != nil {
		t.Fatal(err)
	}
	if len(distinct) != 2 || distinct[0] != 2 || distinct[1] != 2 {
		t.Fatalf("distinct creators = %v, want [2 2]", distinct)
	}

	atLeastTwo, err := repo.CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(ctx, []time.Time{dayOne, dayTwo}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(atLeastTwo) != 2 || atLeastTwo[0] != 1 || atLeastTwo[1] != 1 {
		t.Fatalf("creators with >= 2 events = %v, want [1 1] (the migrated creator)", atLeastTwo)
	}
}
