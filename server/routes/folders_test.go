package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

func folderEventIDs(t *testing.T, row map[string]json.RawMessage) []string {
	t.Helper()
	var ids []string
	if err := json.Unmarshal(row["eventIds"], &ids); err != nil {
		t.Fatalf("decode eventIds: %v", err)
	}
	return ids
}

func findFolderByName(t *testing.T, rows []map[string]json.RawMessage, name string) map[string]json.RawMessage {
	t.Helper()
	for _, row := range rows {
		var folderName string
		if err := json.Unmarshal(row["name"], &folderName); err != nil {
			continue
		}
		if folderName == name {
			return row
		}
	}
	return nil
}

// TestSignedInFoldersCrudAndIsolation proves that folder create, read, update,
// and delete persist for the owning account and that another account
// can neither see nor mutate the folder.
func TestSignedInFoldersCrudAndIsolation(t *testing.T) {
	router := signedInEventRouter(t)
	owner, _ := createSignedInAccount(t, router)
	stranger, _ := createSignedInAccount(t, router)

	name := "Folder " + models.NewUUID().String()
	created := owner.request(http.MethodPost, "/api/user/folders", map[string]any{"name": name, "color": "#abcdef"}, http.StatusCreated)
	folderID := decodeAccountString(t, created, "id")
	if !validFolderID(folderID) {
		t.Fatalf("created folder id %q is not a canonical UUID", folderID)
	}

	listed := owner.requestArray(http.MethodGet, "/api/user/folders", http.StatusOK)
	row := findFolderByName(t, listed, name)
	if row == nil {
		t.Fatal("owner did not see the created folder")
	}
	if ids := folderEventIDs(t, row); len(ids) != 0 {
		t.Fatalf("new folder had members: %v", ids)
	}

	updatedName := name + " updated"
	owner.request(http.MethodPatch, "/api/user/folders/"+folderID, map[string]any{"name": updatedName, "color": "#123456"}, http.StatusOK)
	row = findFolderByName(t, owner.requestArray(http.MethodGet, "/api/user/folders", http.StatusOK), updatedName)
	if row == nil {
		t.Fatal("owner did not see the updated folder")
	}
	details := owner.request(http.MethodGet, "/api/user/folders/"+folderID, nil, http.StatusOK)
	if got := decodeAccountString(t, details, "_id"); got != folderID {
		t.Fatalf("folder _id = %q, want %q", got, folderID)
	}

	// Another account can neither read nor mutate the folder.
	if findFolderByName(t, stranger.requestArray(http.MethodGet, "/api/user/folders", http.StatusOK), updatedName) != nil {
		t.Fatal("stranger saw another account's folder in the list")
	}
	stranger.request(http.MethodGet, "/api/user/folders/"+folderID, nil, http.StatusNotFound)
	stranger.request(http.MethodPatch, "/api/user/folders/"+folderID, map[string]any{"name": "hijacked"}, http.StatusNotFound)
	stranger.request(http.MethodDelete, "/api/user/folders/"+folderID, nil, http.StatusNotFound)

	owner.request(http.MethodDelete, "/api/user/folders/"+folderID, nil, http.StatusOK)
	owner.request(http.MethodGet, "/api/user/folders/"+folderID, nil, http.StatusNotFound)
	if findFolderByName(t, owner.requestArray(http.MethodGet, "/api/user/folders", http.StatusOK), updatedName) != nil {
		t.Fatal("deleted folder still listed")
	}
}

// TestSignedInFolderMembershipForEvents proves that an event
// can be added to and removed from a folder, that reads expose the canonical
// public identifier, and that a repeated move does not duplicate a member.
func TestSignedInFolderMembershipForEvents(t *testing.T) {
	router := signedInEventRouter(t)
	owner, account := createSignedInAccount(t, router)
	ctx := context.Background()

	folderName := "Membership " + models.NewUUID().String()
	folderID := decodeAccountString(t, owner.request(http.MethodPost, "/api/user/folders", map[string]any{"name": folderName}, http.StatusCreated), "id")

	eventID := createDashboardEvent(t, owner, "Membership event")

	owner.request(http.MethodPost, "/api/user/events/"+eventID+"/set-folder", map[string]any{"folderId": folderID}, http.StatusOK)
	// Re-adding the same event must not create duplicate memberships.
	owner.request(http.MethodPost, "/api/user/events/"+eventID+"/set-folder", map[string]any{"folderId": folderID}, http.StatusOK)

	row := findFolderByName(t, owner.requestArray(http.MethodGet, "/api/user/folders", http.StatusOK), folderName)
	if row == nil {
		t.Fatal("owner did not see the folder after adding members")
	}
	ids := folderEventIDs(t, row)
	if len(ids) != 1 || ids[0] != eventID {
		t.Fatalf("folder members = %v, want exactly the canonical id %q", ids, eventID)
	}

	// The explicit storage reference is recorded once with the event UUID populated.
	var eventRefs int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT count(*) FROM folder_events WHERE platform_identity_id = $1 AND event_id IS NOT NULL`, account.PlatformIdentityID).Scan(&eventRefs); err != nil {
		t.Fatal(err)
	}
	if eventRefs != 1 {
		t.Fatalf("storage references wrong: refs=%d", eventRefs)
	}

	// Another account's folder cannot receive a member.
	stranger, _ := createSignedInAccount(t, router)
	stranger.request(http.MethodPost, "/api/user/events/"+eventID+"/set-folder", map[string]any{"folderId": folderID}, http.StatusNotFound)

	owner.request(http.MethodPost, "/api/user/events/"+eventID+"/set-folder", map[string]any{"folderId": nil}, http.StatusOK)
	row = findFolderByName(t, owner.requestArray(http.MethodGet, "/api/user/folders", http.StatusOK), folderName)
	if row == nil {
		t.Fatal("folder disappeared after removing members")
	}
	if ids := folderEventIDs(t, row); len(ids) != 0 {
		t.Fatalf("members survived removal: %v", ids)
	}
}

// TestSignedInFolderDeleteRemovesMembershipsAndOwnedMembers proves that
// deleting a folder removes its memberships and soft-deletes the account's own member events.
func TestSignedInFolderDeleteRemovesMembershipsAndOwnedMembers(t *testing.T) {
	router := signedInEventRouter(t)
	owner, account := createSignedInAccount(t, router)
	ctx := context.Background()

	folderName := "Delete " + models.NewUUID().String()
	folderID := decodeAccountString(t, owner.request(http.MethodPost, "/api/user/folders", map[string]any{"name": folderName}, http.StatusCreated), "id")

	eventID := createDashboardEvent(t, owner, "Delete member")
	owner.request(http.MethodPost, "/api/user/events/"+eventID+"/set-folder", map[string]any{"folderId": folderID}, http.StatusOK)

	owner.request(http.MethodDelete, "/api/user/folders/"+folderID, nil, http.StatusOK)

	var memberships int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT count(*) FROM folder_events WHERE platform_identity_id = $1`, account.PlatformIdentityID).Scan(&memberships); err != nil {
		t.Fatal(err)
	}
	if memberships != 0 {
		t.Fatalf("folder memberships survived deletion: %d", memberships)
	}
	var deleted bool
	if err := pgstore.Pool.QueryRow(ctx, `SELECT is_deleted FROM events WHERE short_id = $1`, eventID).Scan(&deleted); err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("owned member event was not soft-deleted with the folder")
	}
}
