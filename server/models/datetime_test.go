package models

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func TestDateTimeMarshalsRFC3339(t *testing.T) {
	wholeSecond := NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 0, 0, 0, time.UTC))
	if got := marshalJSONString(t, wholeSecond); got != `"2026-01-05T14:00:00Z"` {
		t.Fatalf("whole-second datetime = %s", got)
	}

	withMillis := NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 0, 0, 123000000, time.UTC))
	if got := marshalJSONString(t, withMillis); got != `"2026-01-05T14:00:00.123Z"` {
		t.Fatalf("millisecond datetime = %s", got)
	}
}

func TestDateTimeUnmarshalAcceptsRFC3339AndNull(t *testing.T) {
	var value DateTime
	if err := json.Unmarshal([]byte(`"2026-01-05T14:00:00.123Z"`), &value); err != nil {
		t.Fatalf("unmarshal datetime: %v", err)
	}
	if value != NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 0, 0, 123000000, time.UTC)) {
		t.Fatalf("decoded datetime = %d", int64(value))
	}

	var unset DateTime
	if err := json.Unmarshal([]byte(`null`), &unset); err != nil {
		t.Fatalf("unmarshal null datetime: %v", err)
	}
	if !unset.IsZero() {
		t.Fatalf("null datetime = %d, want zero", int64(unset))
	}

	// The driver leaves a set value unchanged when null is decoded, so keep the
	// same parity contract.
	kept := NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 0, 0, 123000000, time.UTC))
	if err := json.Unmarshal([]byte(`null`), &kept); err != nil {
		t.Fatalf("unmarshal null into set datetime: %v", err)
	}
	if kept != NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 0, 0, 123000000, time.UTC)) {
		t.Fatalf("null changed a set datetime: %d", int64(kept))
	}

	if err := json.Unmarshal([]byte(`"not-a-time"`), &value); err == nil {
		t.Fatal("expected an invalid datetime to be rejected")
	}
}

func TestDateTimeMapKeysEncodeMilliseconds(t *testing.T) {
	key := NewDateTimeFromTime(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))
	payload, err := json.Marshal(map[DateTime]string{key: "available"})
	if err != nil {
		t.Fatalf("marshal millisecond-keyed map: %v", err)
	}

	expectedKey := strconv.FormatInt(int64(key), 10)
	if expected := `{"` + expectedKey + `":"available"}`; string(payload) != expected {
		t.Fatalf("millisecond-keyed map = %s, want %s", payload, expected)
	}
}

func TestResponseManualAvailabilityUsesMillisecondKeysAndRFC3339Values(t *testing.T) {
	day := NewDateTimeFromTime(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))
	slot := NewDateTimeFromTime(time.Date(2026, 1, 5, 9, 30, 0, 0, time.UTC))
	manual := map[DateTime][]DateTime{day: {slot}}
	response := Response{UserId: ZeroUUID(), ManualAvailability: &manual}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded["userId"] != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("userId = %v, want the zero UUID sentinel", decoded["userId"])
	}

	manualDecoded, ok := decoded["manualAvailability"].(map[string]any)
	if !ok {
		t.Fatalf("manualAvailability = %#v", decoded["manualAvailability"])
	}
	key := strconv.FormatInt(int64(day), 10)
	slots, ok := manualDecoded[key].([]any)
	if !ok || len(slots) != 1 {
		t.Fatalf("manualAvailability[%s] = %#v", key, manualDecoded[key])
	}
	if slots[0] != "2026-01-05T09:30:00Z" {
		t.Fatalf("manualAvailability slot = %v", slots[0])
	}
	if _, exists := manualDecoded[time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)]; exists {
		t.Fatal("manualAvailability keys must stay millisecond integers")
	}
}

func TestResponseAvailabilityMarshalsRFC3339(t *testing.T) {
	response := Response{
		UserId:       ZeroUUID(),
		Availability: []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 0, 0, 0, time.UTC))},
		IfNeeded:     []DateTime{NewDateTimeFromTime(time.Date(2026, 1, 5, 14, 15, 0, 0, time.UTC))},
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	availability := decoded["availability"].([]any)
	if len(availability) != 1 || availability[0] != "2026-01-05T14:00:00Z" {
		t.Fatalf("availability = %#v", decoded["availability"])
	}
	ifNeeded := decoded["ifNeeded"].([]any)
	if len(ifNeeded) != 1 || ifNeeded[0] != "2026-01-05T14:15:00Z" {
		t.Fatalf("ifNeeded = %#v", decoded["ifNeeded"])
	}
}
