package routes

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"timeful/server/eventsource"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

var routeTestDBOnce sync.Once

// anonymousEventRouter returns the event API router the anonymous event contract
// tests exercise and initializes the shared route-test database once.
func anonymousEventRouter(t *testing.T) http.Handler {
	t.Helper()
	if os.Getenv("POSTGRES_APPLICATION_URI") == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required for event route contracts")
	}
	routeTestDBOnce.Do(func() { pgstore.Init() })
	return newEventsReadFiltersTestRouter()
}

// cleanupAnonymousEvent removes the event row an anonymous contract test created
// so the suite stays rerunnable against a retained database.
func cleanupAnonymousEvent(t *testing.T, eventID string) {
	t.Helper()
	if pgstore.Pool == nil {
		return
	}
	if _, err := pgstore.Pool.Exec(context.Background(), `DELETE FROM events WHERE short_id = $1`, eventID); err != nil {
		t.Fatalf("delete event: %v", err)
	}
}

// anonymousEventPayload is deliberately an HTTP DTO: event identifiers are UUID
// strings that do not decode into the canonical identifier type.
type anonymousEventPayload struct {
	ID              string                  `json:"_id"`
	Name            string                  `json:"name"`
	Description     *string                 `json:"description"`
	DaysOnly        *bool                   `json:"daysOnly"`
	Dates           []models.DateTime       `json:"dates"`
	ActiveSlots     []models.DateTime       `json:"activeSlots"`
	EventTimezone   *string                 `json:"eventTimezone"`
	SlotGeneration  *models.SlotGeneration  `json:"slotGeneration"`
	TimedRecurrence *models.TimedRecurrence `json:"timedRecurrence"`
	ScheduledEvent  *models.CalendarEvent   `json:"scheduledEvent"`
}

func canonicalTimedEventPayload(name string) map[string]any {
	return map[string]any{
		"name":                 name,
		"type":                 string(models.SPECIFIC_DATES),
		"activeSlots":          []string{"2026-01-05T14:30:00Z", "2026-01-05T14:00:00Z", "2026-01-05T14:30:00Z"},
		"eventTimezone":        "America/New_York",
		"slotGeneration":       map[string]any{"startTimeLocal": "09:00", "endTimeLocal": "10:00", "timeIncrementMinutes": 15},
		"timedRecurrence":      map[string]any{"kind": "specific_dates", "selectedDays": []string{"2026-01-05"}, "selectedDaysOfWeek": []int{}, "startOnMonday": false},
		"daysOnly":             false,
		"collectEmails":        false,
		"notificationsEnabled": false,
	}
}

func createAnonymousEvent(t *testing.T, router http.Handler, payload map[string]any) string {
	t.Helper()

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	return decodeJSONBody[struct {
		EventID string `json:"eventId"`
	}](t, recorder).EventID
}

func assertEventIDsResolve(t *testing.T, router http.Handler, eventID string) string {
	t.Helper()

	longRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID+"/ids", nil)
	if longRecorder.Code != http.StatusOK {
		t.Fatalf("expected long-ID resolution status 200, got %d: %s", longRecorder.Code, longRecorder.Body.String())
	}
	ids := decodeJSONBody[struct {
		ShortID string `json:"shortId"`
		LongID  string `json:"longId"`
	}](t, longRecorder)
	if !eventsource.Canonical(eventID) {
		t.Fatalf("expected a canonical event identifier, got %q", eventID)
	}
	if ids.LongID != eventID || ids.ShortID != eventID {
		t.Fatalf("expected the one canonical public ID in both compatibility fields, got %#v", ids)
	}

	shortRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+ids.ShortID+"/ids", nil)
	if shortRecorder.Code != http.StatusOK {
		t.Fatalf("expected short-ID resolution status 200, got %d: %s", shortRecorder.Code, shortRecorder.Body.String())
	}
	if shortIDs := decodeJSONBody[struct {
		ShortID string `json:"shortId"`
		LongID  string `json:"longId"`
	}](t, shortRecorder); shortIDs != ids {
		t.Fatalf("expected both identifiers to resolve identically, got %#v and %#v", ids, shortIDs)
	}

	return ids.ShortID
}

func TestAnonymousTimedEventContract(t *testing.T) {
	router := ownerCredentialBrowser(anonymousEventRouter(t))
	eventID := createAnonymousEvent(t, router, canonicalTimedEventPayload("Compatibility timed event"))
	t.Cleanup(func() { cleanupAnonymousEvent(t, eventID) })

	shortID := assertEventIDsResolve(t, router, eventID)
	getRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+shortID, nil)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected event read status 200, got %d: %s", getRecorder.Code, getRecorder.Body.String())
	}
	event := decodeJSONBody[anonymousEventPayload](t, getRecorder)
	assertDateTimesEqual(t, event.ActiveSlots, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T14:30:00Z"),
	})

	responseRecorder := timedEventRequest(t, router, http.MethodPost, "/api/events/"+eventID+"/response", map[string]any{
		"guest":          true,
		"createResponse": true,
		"name":           "Ada",
		"availability":   []string{"2026-01-05T14:00:00Z", "2026-01-05T14:00:00Z"},
		"ifNeeded":       []string{"2026-01-05T14:00:00Z", "2026-01-05T14:15:00Z", "2026-01-05T14:15:00Z"},
	})
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected response status 200, got %d: %s", responseRecorder.Code, responseRecorder.Body.String())
	}
	responseKey := decodeJSONBody[struct {
		ResponseID string `json:"responseId"`
	}](t, responseRecorder).ResponseID
	if responseKey == "" {
		t.Fatal("expected an opaque response ID without legacy guest credentials")
	}
	responsesRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID+"/responses?timeMin=2026-01-05T14:00:00Z&timeMax=2026-01-05T14:30:00Z", nil)
	if responsesRecorder.Code != http.StatusOK {
		t.Fatalf("expected response read status 200, got %d: %s", responsesRecorder.Code, responsesRecorder.Body.String())
	}
	response, exists := decodeJSONBody[map[string]models.Response](t, responsesRecorder)[responseKey]
	if !exists {
		t.Fatalf("expected response map key %q", responseKey)
	}
	if !response.UserId.IsZero() {
		t.Fatalf("guest response userId = %q, want the zero identity sentinel", response.UserId)
	}
	assertDateTimesEqual(t, response.Availability, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
	})
	assertDateTimesEqual(t, response.IfNeeded, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:15:00Z"),
	})

	scheduleRecorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID+"/schedule", map[string]any{
		"startDate": "2026-01-05T14:00:00Z",
		"endDate":   "2026-01-05T15:00:00Z",
	})
	if scheduleRecorder.Code != http.StatusOK {
		t.Fatalf("expected schedule save status 200, got %d: %s", scheduleRecorder.Code, scheduleRecorder.Body.String())
	}
	scheduledRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)
	scheduledEvent := decodeJSONBody[anonymousEventPayload](t, scheduledRecorder)
	if scheduledRecorder.Code != http.StatusOK || scheduledEvent.ScheduledEvent == nil ||
		scheduledEvent.ScheduledEvent.Summary != "Compatibility timed event" {
		t.Fatalf("expected saved public schedule snapshot, got status %d and %#v", scheduledRecorder.Code, scheduledEvent.ScheduledEvent)
	}

	clearRecorder := timedEventRequest(t, router, http.MethodDelete, "/api/events/"+eventID+"/schedule", map[string]any{})
	if clearRecorder.Code != http.StatusOK {
		t.Fatalf("expected schedule clear status 200, got %d: %s", clearRecorder.Code, clearRecorder.Body.String())
	}
	clearedEvent := decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil))
	if clearedEvent.ScheduledEvent != nil {
		t.Fatalf("expected public schedule to be cleared, got %#v", clearedEvent.ScheduledEvent)
	}
}

func TestAnonymousDatesOnlyEventContract(t *testing.T) {
	router := anonymousEventRouter(t)
	eventID := createAnonymousEvent(t, router, map[string]any{
		"name":          "Compatibility dates-only event",
		"type":          string(models.SPECIFIC_DATES),
		"daysOnly":      true,
		"eventTimezone": "America/New_York",
		"dates":         []string{"2026-08-11T00:00:00Z", "2026-08-12T00:00:00Z", "2026-08-11T00:00:00Z"},
	})
	t.Cleanup(func() { cleanupAnonymousEvent(t, eventID) })

	assertEventIDsResolve(t, router, eventID)
	getRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected event read status 200, got %d: %s", getRecorder.Code, getRecorder.Body.String())
	}
	event := decodeJSONBody[anonymousEventPayload](t, getRecorder)
	if event.DaysOnly == nil || !*event.DaysOnly {
		t.Fatalf("expected dates-only event, got %#v", event.DaysOnly)
	}
	assertDateTimesEqual(t, event.Dates, []models.DateTime{
		timedSlotDateTime(t, "2026-08-11T00:00:00Z"),
		timedSlotDateTime(t, "2026-08-12T00:00:00Z"),
		timedSlotDateTime(t, "2026-08-11T00:00:00Z"),
	})
}

// These payload checks act as the creating browser, retaining the event's owner
// credential. Authorization rejection has a separate suite.
func ownerCredentialBrowser(router http.Handler) http.Handler {
	jar, _ := cookiejar.New(nil)
	origin, _ := url.Parse("http://example.com/api/")
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, cookie := range jar.Cookies(origin) {
			req.AddCookie(cookie)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		jar.SetCookies(origin, recorder.Result().Cookies())
		for name, values := range recorder.Header() {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(recorder.Code)
		_, _ = w.Write(recorder.Body.Bytes())
	})
}

func TestAnonymousEventEditContract(t *testing.T) {
	router := ownerCredentialBrowser(anonymousEventRouter(t))
	timedID := createAnonymousEvent(t, router, canonicalTimedEventPayload("Original timed event"))
	t.Cleanup(func() { cleanupAnonymousEvent(t, timedID) })

	// Omitted or empty values keep their stored value: the missing
	// description stays "before" and the empty active-slot list leaves
	// the stored slots unchanged.
	timedEdit := canonicalTimedEventPayload("Edited timed event")
	timedEdit["description"] = "before"
	initialEditRecorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+timedID, timedEdit)
	if initialEditRecorder.Code != http.StatusOK {
		t.Fatalf("expected initial timed edit status 200, got %d: %s", initialEditRecorder.Code, initialEditRecorder.Body.String())
	}

	delete(timedEdit, "description")
	timedEdit["activeSlots"] = []string{}
	timedEditRecorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+timedID, timedEdit)
	if timedEditRecorder.Code != http.StatusOK {
		t.Fatalf("expected timed edit status 200, got %d: %s", timedEditRecorder.Code, timedEditRecorder.Body.String())
	}
	timedEvent := decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+timedID, nil))
	if timedEvent.Name != "Edited timed event" || timedEvent.Description == nil || *timedEvent.Description != "before" {
		t.Fatalf("expected omitted description to remain unchanged, got %#v", timedEvent)
	}
	assertDateTimesEqual(t, timedEvent.ActiveSlots, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T14:30:00Z"),
	})

	timedEdit["description"] = ""
	emptyDescriptionRecorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+timedID, timedEdit)
	if emptyDescriptionRecorder.Code != http.StatusOK {
		t.Fatalf("expected empty description edit status 200, got %d: %s", emptyDescriptionRecorder.Code, emptyDescriptionRecorder.Body.String())
	}
	timedEvent = decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+timedID, nil))
	if timedEvent.Description == nil || *timedEvent.Description != "" {
		t.Fatalf("expected explicit empty description to persist, got %#v", timedEvent.Description)
	}

	datesID := createAnonymousEvent(t, router, map[string]any{
		"name":     "Original dates-only event",
		"type":     string(models.SPECIFIC_DATES),
		"daysOnly": true,
		"dates":    []string{"2026-08-11T00:00:00Z"},
	})
	t.Cleanup(func() { cleanupAnonymousEvent(t, datesID) })
	datesEditRecorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+datesID, map[string]any{
		"name":     "Edited dates-only event",
		"type":     string(models.SPECIFIC_DATES),
		"daysOnly": true,
		"dates":    []string{"2026-08-12T00:00:00Z", "2026-08-11T00:00:00Z", "2026-08-12T00:00:00Z"},
	})
	if datesEditRecorder.Code != http.StatusOK {
		t.Fatalf("expected dates-only edit status 200, got %d: %s", datesEditRecorder.Code, datesEditRecorder.Body.String())
	}
	datesEvent := decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+datesID, nil))
	assertDateTimesEqual(t, datesEvent.Dates, []models.DateTime{
		timedSlotDateTime(t, "2026-08-12T00:00:00Z"),
		timedSlotDateTime(t, "2026-08-11T00:00:00Z"),
		timedSlotDateTime(t, "2026-08-12T00:00:00Z"),
	})
}
