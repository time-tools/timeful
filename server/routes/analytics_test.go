package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"timeful/server/logger"
	pgstore "timeful/server/postgres"
)

// TestAnalyticsDayEndsUseClientOffsetBoundary proves the day spine runs
// inclusive from startDate through endDate, one entry per day at 23:59:59 in
// the client's fixed offset, in ascending order.
func TestAnalyticsDayEndsUseClientOffsetBoundary(t *testing.T) {
	location := time.FixedZone("UserOffset", -7*60*60)
	ends := analyticsDayEnds(
		time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		location,
	)
	want := []string{
		"2026-09-10T23:59:59-07:00",
		"2026-09-11T23:59:59-07:00",
		"2026-09-12T23:59:59-07:00",
	}
	if len(ends) != len(want) {
		t.Fatalf("day ends = %d, want %d", len(ends), len(want))
	}
	for i, end := range ends {
		if got := end.Format(time.RFC3339); got != want[i] {
			t.Fatalf("day end %d = %s, want %s", i, got, want[i])
		}
	}
}

// TestMonthlyActiveEventCreatorsFailsWholeRequestOnPostgresError proves the
// single-query endpoints report a repository failure as 500 instead of the
// retired loop's short partial 200 array.
func TestMonthlyActiveEventCreatorsFailsWholeRequestOnPostgresError(t *testing.T) {
	initRoutesReadFiltersTestDB(t)
	logger.Init(io.Discard)
	t.Setenv("ANALYTICS_USERNAME", "analytics")
	t.Setenv("ANALYTICS_PASSWORD", "secret")

	previousPool := pgstore.Pool
	pgstore.Pool = closedAccountContractPool(t)
	t.Cleanup(func() { pgstore.Pool = previousPool })

	router := gin.New()
	InitAnalytics(router.Group("/api"))

	paths := []string{
		"/api/analytics/monthly-active-event-creators?startDate=2026-09-01&endDate=2026-09-02&timezoneOffset=0",
		"/api/analytics/monthly-active-event-creators-with-more-than-x-events?startDate=2026-09-01&endDate=2026-09-02&timezoneOffset=0&x=2",
	}
	for _, path := range paths {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.SetBasicAuth("analytics", "secret")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("GET %s: status = %d, want 500: %s", path, recorder.Code, recorder.Body.String())
		}
	}
}

// TestMonthlyActiveEventCreatorsServesOneEntryPerDay proves both endpoints
// return a JSON array with exactly one ascending entry per requested day, even
// when no event falls in any day's window.
func TestMonthlyActiveEventCreatorsServesOneEntryPerDay(t *testing.T) {
	initRoutesReadFiltersTestDB(t)
	if os.Getenv("POSTGRES_APPLICATION_URI") == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required for analytics route contracts")
	}
	routeTestDBOnce.Do(func() { pgstore.Init() })
	t.Setenv("ANALYTICS_USERNAME", "analytics")
	t.Setenv("ANALYTICS_PASSWORD", "secret")

	router := gin.New()
	InitAnalytics(router.Group("/api"))

	paths := []string{
		"/api/analytics/monthly-active-event-creators?startDate=2020-01-01&endDate=2020-01-03&timezoneOffset=0",
		"/api/analytics/monthly-active-event-creators-with-more-than-x-events?startDate=2020-01-01&endDate=2020-01-03&timezoneOffset=0&x=2",
	}
	for _, path := range paths {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.SetBasicAuth("analytics", "secret")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s: status = %d: %s", path, recorder.Code, recorder.Body.String())
		}
		var counts []int64
		if err := json.Unmarshal(recorder.Body.Bytes(), &counts); err != nil {
			t.Fatalf("GET %s: decode body: %v", path, err)
		}
		if len(counts) != 3 {
			t.Fatalf("GET %s: day-spine counts = %v, want three entries", path, counts)
		}
		for i, count := range counts {
			if count != 0 {
				t.Fatalf("GET %s: day %d count = %d, want 0", path, i, count)
			}
		}
	}
}
