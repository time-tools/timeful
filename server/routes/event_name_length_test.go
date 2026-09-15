package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timeful/server/errs"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

func repeatedEventName(unit string, count int) string {
	return strings.Repeat(unit, count)
}

func assertEventNameTooLongRejection(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := decodeJSONBody[struct {
		Error string `json:"error"`
	}](t, recorder)
	if body.Error != errs.EventNameTooLong {
		t.Fatalf("expected %q error token, got %q", errs.EventNameTooLong, body.Error)
	}
}

func countEventsByName(t *testing.T, name string) int {
	t.Helper()

	var count int
	if err := pgstore.Pool.QueryRow(context.Background(), `SELECT count(*) FROM events WHERE name = $1`, name).Scan(&count); err != nil {
		t.Fatalf("count events by name: %v", err)
	}
	return count
}

func TestCreateEventRejectsNamesLongerThan100CodePointsWithoutPersisting(t *testing.T) {
	router := anonymousEventRouter(t)

	cases := []struct {
		name  string
		value string
	}{
		{name: "ascii", value: repeatedEventName("a", models.MaxEventNameLength+1)},
		{name: "non-ascii", value: repeatedEventName("é", models.MaxEventNameLength+1)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", canonicalTimedEventPayload(testCase.value))
			assertEventNameTooLongRejection(t, recorder)
			if count := countEventsByName(t, testCase.value); count != 0 {
				t.Fatalf("expected rejected event name not to persist, found %d stored events", count)
			}
		})
	}
}

func TestCreateEventAcceptsNamesAt100CodePoints(t *testing.T) {
	router := anonymousEventRouter(t)

	cases := []struct {
		name  string
		value string
	}{
		{name: "ascii", value: repeatedEventName("a", models.MaxEventNameLength)},
		{name: "non-ascii", value: repeatedEventName("é", models.MaxEventNameLength)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			eventID := createAnonymousEvent(t, router, canonicalTimedEventPayload(testCase.value))
			t.Cleanup(func() { cleanupAnonymousEvent(t, eventID) })

			stored, _ := loadEventModel(t, router, eventID)
			if stored.Name != testCase.value {
				t.Fatalf("expected stored name %q, got %q", testCase.value, stored.Name)
			}
		})
	}
}

func TestEditEventRejectsNamesLongerThan100CodePointsWithoutPersisting(t *testing.T) {
	router := ownerCredentialBrowser(anonymousEventRouter(t))

	originalName := "Original event name"
	eventID := createAnonymousEvent(t, router, canonicalTimedEventPayload(originalName))
	t.Cleanup(func() { cleanupAnonymousEvent(t, eventID) })

	cases := []struct {
		name  string
		value string
	}{
		{name: "ascii", value: repeatedEventName("a", models.MaxEventNameLength+1)},
		{name: "non-ascii", value: repeatedEventName("é", models.MaxEventNameLength+1)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID, canonicalTimedEventPayload(testCase.value))
			assertEventNameTooLongRejection(t, recorder)

			stored, _ := loadEventModel(t, router, eventID)
			if stored.Name != originalName {
				t.Fatalf("expected rejected edit to leave stored name %q, got %q", originalName, stored.Name)
			}
		})
	}
}

func TestEditEventAcceptsNamesAt100CodePoints(t *testing.T) {
	router := ownerCredentialBrowser(anonymousEventRouter(t))

	eventID := createAnonymousEvent(t, router, canonicalTimedEventPayload("Editable name boundary event"))
	t.Cleanup(func() { cleanupAnonymousEvent(t, eventID) })

	cases := []struct {
		name  string
		value string
	}{
		{name: "ascii", value: repeatedEventName("a", models.MaxEventNameLength)},
		{name: "non-ascii", value: repeatedEventName("é", models.MaxEventNameLength)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID, canonicalTimedEventPayload(testCase.value))
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
			}

			stored, _ := loadEventModel(t, router, eventID)
			if stored.Name != testCase.value {
				t.Fatalf("expected stored name %q, got %q", testCase.value, stored.Name)
			}
		})
	}
}
