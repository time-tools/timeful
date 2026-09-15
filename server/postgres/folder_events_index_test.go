package postgres

import (
	"strings"
	"testing"
)

// TestFolderEventsUniqueIndexDropsVestigialPredicate proves the unique index on
// folder_events no longer carries the partial predicate `WHERE event_id IS NOT
// NULL`, which could never exclude a row because event_id is NOT NULL, and that
// its down migration restores the partial form.
func TestFolderEventsUniqueIndexDropsVestigialPredicate(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)

	definition := indexDefinition(t, ctx, tx, "folder_events_event_unique_idx")
	if definition == "" {
		t.Fatal("folder_events_event_unique_idx is missing")
	}
	if !strings.Contains(definition, "(platform_identity_id, event_id)") {
		t.Fatalf("folder_events_event_unique_idx definition = %q, want columns (platform_identity_id, event_id)", definition)
	}
	if strings.Contains(definition, "WHERE") {
		t.Fatalf("folder_events_event_unique_idx is still partial: %q", definition)
	}

	applyMigrationDown(t, ctx, tx, "20260915000001_folder_events_event_unique_idx_drop_predicate.sql")
	if down := indexDefinition(t, ctx, tx, "folder_events_event_unique_idx"); !strings.Contains(down, "event_id IS NOT NULL") {
		t.Fatalf("down migration did not restore the partial predicate: %q", down)
	}
}
