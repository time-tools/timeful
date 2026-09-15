package models

import (
	"encoding/json"
	"testing"
)

func assertJSONKeysPresent(t *testing.T, value any, keys ...string) {
	t.Helper()
	payload := marshalJSONString(t, value)
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("decode %T payload: %v", value, err)
	}
	for _, key := range keys {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("%T payload lacks %q: %s", value, key, payload)
		}
	}
}

func TestUserProfileKeepsNullableWireFields(t *testing.T) {
	assertJSONKeysPresent(t, User{},
		"timezoneOffset", "_id", "email", "firstName", "lastName", "picture",
		"hasCustomName", "calendarAccounts", "primaryAccountKey", "calendarOptions", "numEventsCreated")
}

func TestCalendarAccountKeepsNullableWireFields(t *testing.T) {
	assertJSONKeysPresent(t, CalendarAccount{},
		"calendarType", "oAuth2CalendarAuth", "appleCalendarAuth", "icsCalendarAuth",
		"email", "picture", "enabled", "subCalendars")
	assertJSONKeysPresent(t, SubCalendar{}, "name", "enabled")
}

func TestCalendarEventKeepsNullableWireFields(t *testing.T) {
	assertJSONKeysPresent(t, CalendarEvent{},
		"id", "calendarId", "summary", "startDate", "endDate", "free", "allDay")
}

func TestResponseKeepsNullableWireFields(t *testing.T) {
	assertJSONKeysPresent(t, Response{},
		"name", "email", "userId", "user", "availability", "ifNeeded",
		"manualAvailability", "useCalendarAvailability", "enabledCalendars", "calendarOptions")
}

func TestEventPayloadKeepsNullableWireFields(t *testing.T) {
	assertJSONKeysPresent(t, Event{},
		"_id", "shortId", "ownerId", "name", "description", "isArchived", "isDeleted",
		"notificationsEnabled", "sendEmailAfterXResponses", "when2meetHref", "collectEmails",
		"activeSlots", "eventTimezone", "slotGeneration", "timedRecurrence", "type",
		"creatorPosthogId", "isSignUpForm", "signUpBlocks", "signUpResponses",
		"blindAvailabilityEnabled", "daysOnly", "responses", "numResponses", "scheduledEvent",
		"calendarEventId", "remindees", "hasResponded", "attendees")
}
