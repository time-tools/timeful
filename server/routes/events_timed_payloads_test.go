package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"timeful/server/models"
	"timeful/server/responses"
)

func timedSlotDateTime(t *testing.T, raw string) models.DateTime {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("parse time %q: %v", raw, err)
	}

	return models.NewDateTimeFromTime(parsed.UTC())
}

func timedEventRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	payload any,
) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

// loadPostgresEventModel reads a PostgreSQL event through the same public API a
// browser uses, so creation normalization is asserted at the HTTP boundary. It
// returns the decoded wire payload and the raw JSON object for omitted-field
// checks.
func loadPostgresEventModel(t *testing.T, router http.Handler, eventID string) (anonymousEventPayload, map[string]any) {
	t.Helper()

	recorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected PostgreSQL event read status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	event := decodeJSONBody[anonymousEventPayload](t, recorder)
	raw := decodeJSONBody[map[string]any](t, recorder)
	return event, raw
}

func assertDateTimesEqual(
	t *testing.T,
	actual []models.DateTime,
	expected []models.DateTime,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d slots, got %d (%v)", len(expected), len(actual), actual)
	}

	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("expected slot %d to be %v, got %v", index, expected[index], actual[index])
		}
	}
}

func TestCreateEventCanonicalTimedPayloadNormalizesAndPersistsCanonicalFields(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := store.newRouter(t)

	payload := map[string]any{
		"name":                 "Canonical timed create",
		"description":          "First line\nSecond line",
		"type":                 string(models.SPECIFIC_DATES),
		"activeSlots":          []string{"2026-01-05T14:30:00Z", "2026-01-05T14:00:00Z", "2026-01-05T14:30:00Z"},
		"eventTimezone":        "America/New_York",
		"slotGeneration":       map[string]any{"startTimeLocal": "09:00:00", "endTimeLocal": "10:00:00", "timeIncrementMinutes": 15},
		"timedRecurrence":      map[string]any{"kind": "specific_dates", "selectedDays": []string{"2026-01-05"}, "selectedDaysOfWeek": []int{}, "startOnMonday": false},
		"daysOnly":             false,
		"collectEmails":        false,
		"notificationsEnabled": false,
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	createResponse := decodeJSONBody[struct {
		EventID string `json:"eventId"`
	}](t, recorder)
	t.Cleanup(func() { store.cleanupEvent(t, createResponse.EventID) })

	storedEvent, storedRaw := loadPostgresEventModel(t, router, createResponse.EventID)
	if storedEvent.Description == nil || *storedEvent.Description != "First line\nSecond line" {
		t.Fatalf("expected stored description to persist, got %#v", storedEvent.Description)
	}
	expectedActiveSlots := []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T14:30:00Z"),
	}

	assertDateTimesEqual(t, storedEvent.ActiveSlots, expectedActiveSlots)
	if storedEvent.EventTimezone == nil || *storedEvent.EventTimezone != "America/New_York" {
		t.Fatalf("expected stored timezone to persist, got %#v", storedEvent.EventTimezone)
	}
	for _, legacyField := range []string{"times", "duration", "timeIncrement", "hasSpecificTimes", "startOnMonday"} {
		if _, exists := storedRaw[legacyField]; exists {
			t.Fatalf("expected stored timed event to omit legacy field %q, got %#v", legacyField, storedRaw[legacyField])
		}
	}
	if storedEvent.SlotGeneration == nil ||
		storedEvent.SlotGeneration.StartTimeLocal != "09:00:00" ||
		storedEvent.SlotGeneration.EndTimeLocal != "10:00:00" ||
		storedEvent.SlotGeneration.TimeIncrementMinutes != 15 {
		t.Fatalf("expected stored slot generation to persist, got %#v", storedEvent.SlotGeneration)
	}
	if storedEvent.TimedRecurrence == nil ||
		storedEvent.TimedRecurrence.Kind != "specific_dates" ||
		len(storedEvent.TimedRecurrence.SelectedDays) != 1 ||
		storedEvent.TimedRecurrence.SelectedDays[0] != "2026-01-05" {
		t.Fatalf("expected stored timed recurrence to persist, got %#v", storedEvent.TimedRecurrence)
	}

	getRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+createResponse.EventID, nil)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", getRecorder.Code, getRecorder.Body.String())
	}

	responseEvent := decodeJSONBody[anonymousEventPayload](t, getRecorder)
	if responseEvent.Description == nil || *responseEvent.Description != "First line\nSecond line" {
		t.Fatalf("expected response description to persist, got %#v", responseEvent.Description)
	}
	responsePayload := decodeJSONBody[map[string]any](t, getRecorder)
	for _, legacyField := range []string{"dates", "times", "duration", "timeIncrement", "hasSpecificTimes", "startOnMonday", "enabledSlots"} {
		if _, exists := responsePayload[legacyField]; exists {
			t.Fatalf("expected timed response to omit legacy field %q", legacyField)
		}
	}
	assertDateTimesEqual(t, responseEvent.ActiveSlots, expectedActiveSlots)
	if responseEvent.EventTimezone == nil || *responseEvent.EventTimezone != "America/New_York" {
		t.Fatalf("expected response timezone to persist, got %#v", responseEvent.EventTimezone)
	}
}

func TestCreateEventIgnoresUnknownEnabledSlotsAndDerivesTheDomain(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := store.newRouter(t)

	// An old frontend still sends enabledSlots; the server ignores the
	// unknown key and derives the domain from the contract, so the stored
	// active subset must be inside the derived domain, not the sent one.
	payload := map[string]any{
		"name":            "Known-domain timed create",
		"type":            string(models.SPECIFIC_DATES),
		"enabledSlots":    []string{"2026-01-05T04:00:00Z"},
		"activeSlots":     []string{"2026-01-05T14:00:00Z", "2026-01-05T14:30:00Z"},
		"eventTimezone":   "America/New_York",
		"slotGeneration":  map[string]any{"startTimeLocal": "09:00:00", "endTimeLocal": "10:00:00", "timeIncrementMinutes": 15},
		"timedRecurrence": map[string]any{"kind": "specific_dates", "selectedDays": []string{"2026-01-05"}, "selectedDaysOfWeek": []int{}, "startOnMonday": false},
		"daysOnly":        false,
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	createResponse := decodeJSONBody[struct {
		EventID string `json:"eventId"`
	}](t, recorder)
	t.Cleanup(func() { store.cleanupEvent(t, createResponse.EventID) })

	storedEvent, _ := loadPostgresEventModel(t, router, createResponse.EventID)
	assertDateTimesEqual(t, storedEvent.ActiveSlots, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T14:30:00Z"),
	})
}

func TestEditEventCanonicalTimedPayloadRoundTripsThroughGet(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := compatibilityOwnerBrowser(store.newRouter(t))
	eventID := createAnonymousCompatibilityEvent(t, router, canonicalTimedEventPayload("Editable timed event"))
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	payload := map[string]any{
		"name":          "Updated weekly timed event",
		"type":          string(models.DOW),
		"activeSlots":   []string{"2026-01-05T17:30:00Z", "2026-01-07T17:00:00Z", "2026-01-05T17:00:00Z"},
		"eventTimezone": "America/Los_Angeles",
		"slotGeneration": map[string]any{
			"startTimeLocal":       "09:00:00",
			"endTimeLocal":         "11:00:00",
			"timeIncrementMinutes": 30,
		},
		"timedRecurrence": map[string]any{
			"kind":               "weekly",
			"selectedDays":       []string{},
			"selectedDaysOfWeek": []int{1, 3},
			"startOnMonday":      true,
		},
		"daysOnly":      false,
		"collectEmails": false,
		"description":   "Canonical weekly update",
	}

	recorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID, payload)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	// The weekly domain derives from the anchor week (week of the earliest
	// active instant, Monday 2026-01-05): Mon + Wed 09:00-11:00 LA.
	expectedActiveSlots := []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T17:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T17:30:00Z"),
		timedSlotDateTime(t, "2026-01-07T17:00:00Z"),
	}

	getRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", getRecorder.Code, getRecorder.Body.String())
	}

	responseEvent := decodeJSONBody[anonymousEventPayload](t, getRecorder)
	responsePayload := decodeJSONBody[map[string]any](t, getRecorder)
	for _, legacyField := range []string{"duration", "timeIncrement", "times", "enabledSlots"} {
		if _, exists := responsePayload[legacyField]; exists {
			t.Fatalf("expected timed response to omit legacy field %q", legacyField)
		}
	}
	assertDateTimesEqual(t, responseEvent.ActiveSlots, expectedActiveSlots)
	if responseEvent.EventTimezone == nil || *responseEvent.EventTimezone != "America/Los_Angeles" {
		t.Fatalf("expected response timezone to persist, got %#v", responseEvent.EventTimezone)
	}
	if responseEvent.TimedRecurrence == nil ||
		responseEvent.TimedRecurrence.Kind != "weekly" ||
		len(responseEvent.TimedRecurrence.SelectedDaysOfWeek) != 2 ||
		responseEvent.TimedRecurrence.SelectedDaysOfWeek[0] != 1 ||
		responseEvent.TimedRecurrence.SelectedDaysOfWeek[1] != 3 ||
		responseEvent.TimedRecurrence.StartOnMonday == nil ||
		!*responseEvent.TimedRecurrence.StartOnMonday {
		t.Fatalf("expected stored weekly recurrence to persist, got %#v", responseEvent.TimedRecurrence)
	}
}

func TestEditDayOnlyEventPersistsTimezone(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := compatibilityOwnerBrowser(store.newRouter(t))
	eventID := createAnonymousCompatibilityEvent(t, router, map[string]any{
		"name":          "Editable day-only event",
		"type":          string(models.SPECIFIC_DATES),
		"daysOnly":      true,
		"dates":         []string{"2026-08-11T00:00:00Z"},
		"eventTimezone": "UTC",
	})
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	payload := map[string]any{
		"name":          "Updated day-only event",
		"type":          string(models.SPECIFIC_DATES),
		"daysOnly":      true,
		"dates":         []string{"2026-08-11T00:00:00Z"},
		"eventTimezone": "America/New_York",
	}

	recorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID, payload)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	getRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID, nil)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", getRecorder.Code, getRecorder.Body.String())
	}
	storedEvent := decodeJSONBody[anonymousEventPayload](t, getRecorder)
	if storedEvent.EventTimezone == nil || *storedEvent.EventTimezone != "America/New_York" {
		t.Fatalf("expected stored timezone to update, got %#v", storedEvent.EventTimezone)
	}
}

func TestUpdateEventResponseCanonicalizesOverlappingTimedSlots(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := store.newRouter(t)
	eventID := createAnonymousCompatibilityEvent(t, router, canonicalTimedEventPayload("Canonical response event"))
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	payload := map[string]any{
		"createResponse": true,
		"guest":          true,
		"name":           "Maya",
		"availability":   []string{"2026-01-05T14:00:00Z", "2026-01-05T14:00:00Z"},
		"ifNeeded":       []string{"2026-01-05T14:00:00Z", "2026-01-05T14:15:00Z", "2026-01-05T14:15:00Z"},
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events/"+eventID+"/response", payload)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	responseID := decodeJSONBody[struct {
		ResponseID string `json:"responseId"`
	}](t, recorder).ResponseID
	if responseID == "" {
		t.Fatal("expected an opaque response id")
	}

	responsesRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID+"/responses?timeMin=2026-01-05T14:00:00Z&timeMax=2026-01-05T14:30:00Z", nil)
	if responsesRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", responsesRecorder.Code, responsesRecorder.Body.String())
	}
	eventResponse, exists := decodeJSONBody[map[string]models.Response](t, responsesRecorder)[responseID]
	if !exists {
		t.Fatalf("expected stored response %q", responseID)
	}

	assertDateTimesEqual(
		t,
		eventResponse.Availability,
		[]models.DateTime{
			timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
		},
	)
	assertDateTimesEqual(
		t,
		eventResponse.IfNeeded,
		[]models.DateTime{
			timedSlotDateTime(t, "2026-01-05T14:15:00Z"),
		},
	)
}

func TestCreateEventRejectsActiveSlotsOutsideDerivedDomain(t *testing.T) {
	initRoutesReadFiltersTestDB(t)
	router := newEventsReadFiltersTestRouter()

	payload := map[string]any{
		"name":            "Invalid canonical timed create",
		"type":            string(models.SPECIFIC_DATES),
		"activeSlots":     []string{"2026-01-06T05:15:00Z"}, // 00:15 Jan 6 NY, outside the picked Jan 5 full civil day
		"eventTimezone":   "America/New_York",
		"slotGeneration":  map[string]any{"startTimeLocal": "09:00", "endTimeLocal": "10:00", "timeIncrementMinutes": 15},
		"timedRecurrence": map[string]any{"kind": "specific_dates", "selectedDays": []string{"2026-01-05"}, "selectedDaysOfWeek": []int{}, "startOnMonday": false},
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}

	errorResponse := decodeJSONBody[responses.Error](t, recorder)
	if errorResponse.Error != errActiveSlotOutsideEnabled.Error() {
		t.Fatalf("expected error %q, got %#v", errActiveSlotOutsideEnabled.Error(), errorResponse.Error)
	}
}

func TestCreateEventRejectsWeeklyActiveSlotsOutsideDerivedDomain(t *testing.T) {
	initRoutesReadFiltersTestDB(t)
	router := newEventsReadFiltersTestRouter()

	payload := map[string]any{
		"name":            "Invalid weekly timed create",
		"type":            string(models.DOW),
		"activeSlots":     []string{"2026-01-04T23:00:00Z"}, // 15:00 Sun Jan 4 LA, not a selected day
		"eventTimezone":   "America/Los_Angeles",
		"slotGeneration":  map[string]any{"startTimeLocal": "09:00", "endTimeLocal": "11:00", "timeIncrementMinutes": 30},
		"timedRecurrence": map[string]any{"kind": "weekly", "selectedDays": []string{}, "selectedDaysOfWeek": []int{1, 3}, "startOnMonday": true},
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}

	errorResponse := decodeJSONBody[responses.Error](t, recorder)
	if errorResponse.Error != errActiveSlotOutsideEnabled.Error() {
		t.Fatalf("expected error %q, got %#v", errActiveSlotOutsideEnabled.Error(), errorResponse.Error)
	}
}

func TestCreateEventAcceptsActiveSlotsInsideFullDayOutsideWindow(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := store.newRouter(t)

	// The enabled domain is the full civil day, not the 09:00-10:00 window:
	// an active at 00:30 New York time is inside the day but outside the
	// stored window and must be accepted and persisted.
	payload := map[string]any{
		"name":            "Out-of-window canonical timed create",
		"type":            string(models.SPECIFIC_DATES),
		"activeSlots":     []string{"2026-01-05T05:30:00Z", "2026-01-05T14:30:00Z"},
		"eventTimezone":   "America/New_York",
		"slotGeneration":  map[string]any{"startTimeLocal": "09:00:00", "endTimeLocal": "10:00:00", "timeIncrementMinutes": 15},
		"timedRecurrence": map[string]any{"kind": "specific_dates", "selectedDays": []string{"2026-01-05"}, "selectedDaysOfWeek": []int{}, "startOnMonday": false},
		"daysOnly":        false,
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	createResponse := decodeJSONBody[struct {
		EventID string `json:"eventId"`
	}](t, recorder)
	t.Cleanup(func() { store.cleanupEvent(t, createResponse.EventID) })

	storedEvent, _ := loadPostgresEventModel(t, router, createResponse.EventID)
	assertDateTimesEqual(t, storedEvent.ActiveSlots, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T05:30:00Z"),
		timedSlotDateTime(t, "2026-01-05T14:30:00Z"),
	})
}

func TestCreateEventPreservesExplicitEmptyActiveSlots(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := store.newRouter(t)

	payload := map[string]any{
		"name":            "Specific times empty active subset",
		"type":            string(models.SPECIFIC_DATES),
		"activeSlots":     []string{},
		"eventTimezone":   "America/New_York",
		"slotGeneration":  map[string]any{"startTimeLocal": "09:00", "endTimeLocal": "10:00", "timeIncrementMinutes": 15},
		"timedRecurrence": map[string]any{"kind": "specific_dates", "selectedDays": []string{"2026-01-05"}, "selectedDaysOfWeek": []int{}, "startOnMonday": false},
		"daysOnly":        false,
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	createResponse := decodeJSONBody[struct {
		EventID string `json:"eventId"`
	}](t, recorder)
	t.Cleanup(func() { store.cleanupEvent(t, createResponse.EventID) })

	storedEvent, _ := loadPostgresEventModel(t, router, createResponse.EventID)
	assertDateTimesEqual(t, storedEvent.ActiveSlots, []models.DateTime{})
}

func TestCreateEventRejectsLegacyTimedFields(t *testing.T) {
	initRoutesReadFiltersTestDB(t)
	router := newEventsReadFiltersTestRouter()

	payload := map[string]any{
		"name":          "Legacy timed create",
		"duration":      1,
		"dates":         []string{"2026-01-05T09:00:00Z"},
		"type":          string(models.SPECIFIC_DATES),
		"timeIncrement": 20,
	}

	recorder := timedEventRequest(t, router, http.MethodPost, "/api/events", payload)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if response := decodeJSONBody[responses.Error](t, recorder); response.Error != "legacy-timed-event-field:duration" {
		t.Fatalf("expected legacy field error, got %#v", response.Error)
	}
}

func TestEditEventDiscardsOutOfDomainActivesWhenActiveSlotsStayEmpty(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := compatibilityOwnerBrowser(store.newRouter(t))
	eventID := createAnonymousCompatibilityEvent(t, router, canonicalTimedEventPayload("Timezone-switch timed event"))
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	payload := canonicalTimedEventPayload("Timezone-switch timed event")
	payload["eventTimezone"] = "Pacific/Auckland"
	payload["activeSlots"] = []string{}

	recorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID, payload)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	storedEvent, _ := loadPostgresEventModel(t, router, eventID)
	assertDateTimesEqual(t, storedEvent.ActiveSlots, []models.DateTime{})
	if storedEvent.EventTimezone == nil || *storedEvent.EventTimezone != "Pacific/Auckland" {
		t.Fatalf("expected stored timezone to update, got %#v", storedEvent.EventTimezone)
	}
	if storedEvent.TimedRecurrence == nil ||
		len(storedEvent.TimedRecurrence.SelectedDays) != 1 ||
		storedEvent.TimedRecurrence.SelectedDays[0] != "2026-01-05" {
		t.Fatalf("expected picked dates to stay stable, got %#v", storedEvent.TimedRecurrence)
	}
}

func TestEditEventDiscardsResponseSlotsOutsideRebuiltActiveDomain(t *testing.T) {
	store := anonymousEventContractStores()[0]
	router := compatibilityOwnerBrowser(store.newRouter(t))
	eventID := createAnonymousCompatibilityEvent(t, router, canonicalTimedEventPayload("Response cleanup timed event"))
	t.Cleanup(func() { store.cleanupEvent(t, eventID) })

	responseRecorder := timedEventRequest(t, router, http.MethodPost, "/api/events/"+eventID+"/response", map[string]any{
		"createResponse": true,
		"guest":          true,
		"name":           "Maya",
		"availability":   []string{"2026-01-05T14:00:00Z", "2026-01-05T14:30:00Z"},
		"ifNeeded":       []string{"2026-01-05T14:15:00Z", "2026-01-05T14:30:00Z"},
	})
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected response status 200, got %d: %s", responseRecorder.Code, responseRecorder.Body.String())
	}
	responseID := decodeJSONBody[struct {
		ResponseID string `json:"responseId"`
	}](t, responseRecorder).ResponseID

	edit := canonicalTimedEventPayload("Response cleanup timed event")
	edit["activeSlots"] = []string{"2026-01-05T14:00:00Z", "2026-01-05T14:15:00Z"}
	editRecorder := timedEventRequest(t, router, http.MethodPut, "/api/events/"+eventID, edit)
	if editRecorder.Code != http.StatusOK {
		t.Fatalf("expected edit status 200, got %d: %s", editRecorder.Code, editRecorder.Body.String())
	}

	responsesRecorder := timedEventRequest(t, router, http.MethodGet, "/api/events/"+eventID+"/responses?timeMin=2026-01-05T00:00:00Z&timeMax=2026-01-06T00:00:00Z", nil)
	if responsesRecorder.Code != http.StatusOK {
		t.Fatalf("expected responses status 200, got %d: %s", responsesRecorder.Code, responsesRecorder.Body.String())
	}
	eventResponse, exists := decodeJSONBody[map[string]models.Response](t, responsesRecorder)[responseID]
	if !exists {
		t.Fatalf("expected stored response %q", responseID)
	}

	assertDateTimesEqual(t, eventResponse.Availability, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
	})
	assertDateTimesEqual(t, eventResponse.IfNeeded, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:15:00Z"),
	})
}
