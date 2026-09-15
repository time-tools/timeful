package models

import (
	"encoding/json"
	"testing"
	"time"
)

func boolPtrEvent(value bool) *bool { return &value }
func intPtrEvent(value int) *int    { return &value }

func TestEventIDFieldsKeepUUIDWireFormat(t *testing.T) {
	event := Event{Id: "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e", OwnerId: "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4f"}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if decoded["_id"] != "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e" {
		t.Fatalf("_id = %v", decoded["_id"])
	}
	if decoded["ownerId"] != "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4f" {
		t.Fatalf("ownerId = %v", decoded["ownerId"])
	}
}

func TestEventZeroIDsSurfaceGuestSentinel(t *testing.T) {
	payload, err := json.Marshal(Event{})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if decoded["_id"] != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("_id = %v, want the zero UUID sentinel", decoded["_id"])
	}
	if decoded["ownerId"] != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("ownerId = %v, want the zero UUID sentinel", decoded["ownerId"])
	}

	signUp := SignUpResponse{}
	signUpPayload, err := json.Marshal(signUp)
	if err != nil {
		t.Fatalf("marshal signup response: %v", err)
	}
	var signUpDecoded map[string]any
	if err := json.Unmarshal(signUpPayload, &signUpDecoded); err != nil {
		t.Fatalf("decode signup response: %v", err)
	}
	if signUpDecoded["userId"] != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("signup userId = %v, want the zero UUID sentinel", signUpDecoded["userId"])
	}
}

func TestEventMarshalJSONPreservesLegacyScheduleColumns(t *testing.T) {
	dates := []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))}
	times := []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC))}
	event := Event{
		Id:               "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e",
		OwnerId:          "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4f",
		Duration:         float32PtrEvent(1.5),
		Dates:            dates,
		TimeIncrement:    intPtrEvent(15),
		HasSpecificTimes: boolPtrEvent(true),
		Times:            times,
		StartOnMonday:    boolPtrEvent(true),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	var roundTripped Event
	if err := json.Unmarshal(payload, &roundTripped); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if roundTripped.Duration == nil || *roundTripped.Duration != 1.5 {
		t.Fatalf("duration = %v, want 1.5", roundTripped.Duration)
	}
	if len(roundTripped.Dates) != 1 || !roundTripped.Dates[0].Time().Equal(dates[0].Time()) {
		t.Fatalf("dates = %v, want %v", roundTripped.Dates, dates)
	}
	if roundTripped.TimeIncrement == nil || *roundTripped.TimeIncrement != 15 {
		t.Fatalf("timeIncrement = %v, want 15", roundTripped.TimeIncrement)
	}
	if roundTripped.HasSpecificTimes == nil || !*roundTripped.HasSpecificTimes {
		t.Fatalf("hasSpecificTimes = %v, want true", roundTripped.HasSpecificTimes)
	}
	if len(roundTripped.Times) != 1 || !roundTripped.Times[0].Time().Equal(times[0].Time()) {
		t.Fatalf("times = %v, want %v", roundTripped.Times, times)
	}
	if roundTripped.StartOnMonday == nil || !*roundTripped.StartOnMonday {
		t.Fatalf("startOnMonday = %v, want true", roundTripped.StartOnMonday)
	}
}

func TestEventMarshalAPIJSONSuppressesLegacyTimedScheduleColumns(t *testing.T) {
	event := Event{
		Id:               "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e",
		OwnerId:          "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4f",
		Duration:         float32PtrEvent(1.5),
		Dates:            []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))},
		TimeIncrement:    intPtrEvent(15),
		HasSpecificTimes: boolPtrEvent(true),
		Times:            []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC))},
		StartOnMonday:    boolPtrEvent(true),
	}

	payload, err := event.MarshalAPIJSON()
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	for _, legacyField := range []string{"duration", "dates", "timeIncrement", "hasSpecificTimes", "times", "startOnMonday"} {
		if _, exists := decoded[legacyField]; exists {
			t.Fatalf("expected %q to be suppressed for timed events", legacyField)
		}
	}
}

func TestEventMarshalJSONKeepsLegacyAttendeesKeyAsNull(t *testing.T) {
	for name, event := range map[string]Event{
		"timed":    {},
		"daysOnly": {DaysOnly: boolPtrEvent(true)},
	} {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("marshal event: %v", err)
			}
			var decoded map[string]json.RawMessage
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatalf("decode event: %v", err)
			}
			attendees, exists := decoded["attendees"]
			if !exists {
				t.Fatalf("expected attendees key in %s", payload)
			}
			if string(attendees) != "null" {
				t.Fatalf("attendees = %s, want null", attendees)
			}
		})
	}
}

func TestEventMarshalJSONKeepsLegacyColumnsForDaysOnlyEvents(t *testing.T) {
	event := Event{
		Id:            "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e",
		OwnerId:       "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4f",
		DaysOnly:      boolPtrEvent(true),
		Duration:      float32PtrEvent(1.5),
		Dates:         []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))},
		TimeIncrement: intPtrEvent(15),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if _, exists := decoded["duration"]; !exists {
		t.Fatal("expected the day-only event to keep duration")
	}
	if string(decoded["dates"]) != `["2026-01-05T00:00:00Z"]` {
		t.Fatalf("dates = %s", decoded["dates"])
	}
	if _, exists := decoded["timeIncrement"]; !exists {
		t.Fatal("expected the day-only event to keep timeIncrement")
	}
}

func float32PtrEvent(value float32) *float32 { return &value }
