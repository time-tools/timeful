package routes

import (
	"net/http"
	"testing"

	"timeful/server/models"
)

func TestTimefulScheduleCanBeSavedReplacedAndCleared(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := compatibilityOwnerBrowser(store.newRouter(t))
	eventID := createAnonymousCompatibilityEvent(t, router, map[string]any{
		"name":     "Public planning",
		"type":     string(models.SPECIFIC_DATES),
		"daysOnly": true,
		"dates":    []string{"2026-08-01T00:00:00Z"},
	})
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	firstPayload := map[string]any{
		"startDate": "2026-08-01T09:00:00Z",
		"endDate":   "2026-08-01T10:00:00Z",
	}
	firstRecorder := timedEventRequest(
		t, router, http.MethodPut, "/api/events/"+eventID+"/schedule", firstPayload,
	)
	if firstRecorder.Code != http.StatusOK {
		t.Fatalf("expected save status 200, got %d: %s", firstRecorder.Code, firstRecorder.Body.String())
	}

	saved := decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)).ScheduledEvent
	if saved == nil || saved.Summary != "Public planning" || saved.StartDate != timedSlotDateTime(t, "2026-08-01T09:00:00Z") || saved.EndDate != timedSlotDateTime(t, "2026-08-01T10:00:00Z") {
		t.Fatalf("expected saved Timeful schedule, got %#v", saved)
	}

	replacementPayload := map[string]any{
		"startDate": "2026-08-01T11:00:00Z",
		"endDate":   "2026-08-01T12:30:00Z",
	}
	replacementRecorder := timedEventRequest(
		t, router, http.MethodPut, "/api/events/"+eventID+"/schedule", replacementPayload,
	)
	if replacementRecorder.Code != http.StatusOK {
		t.Fatalf("expected replacement status 200, got %d: %s", replacementRecorder.Code, replacementRecorder.Body.String())
	}

	replaced := decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)).ScheduledEvent
	if replaced == nil || replaced.StartDate != timedSlotDateTime(t, "2026-08-01T11:00:00Z") || replaced.EndDate != timedSlotDateTime(t, "2026-08-01T12:30:00Z") {
		t.Fatalf("expected replacement Timeful schedule, got %#v", replaced)
	}

	clearRecorder := timedEventRequest(
		t, router, http.MethodDelete, "/api/events/"+eventID+"/schedule", map[string]any{},
	)
	if clearRecorder.Code != http.StatusOK {
		t.Fatalf("expected clear status 200, got %d: %s", clearRecorder.Code, clearRecorder.Body.String())
	}
	cleared := decodeJSONBody[anonymousEventPayload](t, timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)).ScheduledEvent
	if cleared != nil {
		t.Fatalf("expected Timeful schedule to be cleared, got %#v", cleared)
	}
}

// The end-before-start validation deliberately runs before owner authorization,
// so a cookie-less request with an invalid range gets 400 rather than 403.
func TestTimefulScheduleRejectsEmptyRangeBeforeOwnerAuthorization(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := store.newRouter(t)
	eventID := createAnonymousCompatibilityEvent(t, router, map[string]any{
		"name":     "Invalid range",
		"type":     string(models.SPECIFIC_DATES),
		"daysOnly": true,
		"dates":    []string{"2026-08-01T00:00:00Z"},
	})
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	recorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID+"/schedule", map[string]any{
		"startDate": "2026-08-01T10:00:00Z",
		"endDate":   "2026-08-01T10:00:00Z",
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid range status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
