package postgres

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
)

func decodeEventPayload(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode payload %s: %v", raw, err)
	}
	return decoded
}

// TestTargetedEventMutations proves the response-counter and lifecycle-flag
// updates change only their target column and leave the stored payload and
// metadata untouched, that a missing event reports pgx.ErrNoRows, and that a
// decrement below zero clamps at zero.
func TestTargetedEventMutations(t *testing.T) {
	ctx, repo, _ := newMigrationTestRepository(t)
	payload := json.RawMessage(`{"name":"Targeted","type":"specific_dates"}`)
	event := &Event{Name: "Targeted", Type: EventTypeSpecificDates, Payload: payload}
	if err := repo.CreateEvent(ctx, event); err != nil {
		t.Fatal(err)
	}

	if err := repo.AdjustEventResponseCount(ctx, event.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := repo.AdjustEventResponseCount(ctx, event.ID, -1); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.NumResponses != 1 {
		t.Fatalf("num_responses = %d, want 1", stored.NumResponses)
	}
	if stored.Name != "Targeted" || !reflect.DeepEqual(decodeEventPayload(t, stored.Payload), decodeEventPayload(t, payload)) {
		t.Fatalf("counter update rewrote the event row: %#v %s", stored, stored.Payload)
	}

	if err := repo.SetEventArchived(ctx, event.ID, true); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.IsArchived || stored.IsDeleted {
		t.Fatalf("archive flag = %t/%t, want true/false", stored.IsArchived, stored.IsDeleted)
	}
	if stored.NumResponses != 1 || !reflect.DeepEqual(decodeEventPayload(t, stored.Payload), decodeEventPayload(t, payload)) {
		t.Fatalf("archive update rewrote the event row: %#v %s", stored, stored.Payload)
	}

	if err := repo.SetEventDeleted(ctx, event.ID, true); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.IsDeleted || !stored.IsArchived || stored.NumResponses != 1 {
		t.Fatalf("delete update changed unrelated state: %#v", stored)
	}
	if !reflect.DeepEqual(decodeEventPayload(t, stored.Payload), decodeEventPayload(t, payload)) {
		t.Fatalf("delete update rewrote the payload: %s", stored.Payload)
	}

	// A decrement below zero is clamped instead of violating the
	// num_responses >= 0 check.
	if err := repo.AdjustEventResponseCount(ctx, event.ID, -1); err != nil {
		t.Fatal(err)
	}
	if err := repo.AdjustEventResponseCount(ctx, event.ID, -1); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.NumResponses != 0 {
		t.Fatalf("counter clamp = %d, want 0", stored.NumResponses)
	}

	// A missing event is pgx.ErrNoRows, matching UpdateEvent.
	missingID := "086f4f9a-1b0e-4b7c-9a3e-6f4c2d1e5a02"
	if err := repo.AdjustEventResponseCount(ctx, missingID, 1); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing event counter error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.SetEventArchived(ctx, missingID, true); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing event archive error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.SetEventDeleted(ctx, missingID, true); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing event delete error = %v, want pgx.ErrNoRows", err)
	}
}
