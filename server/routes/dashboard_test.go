package routes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// requestArray reads a JSON array response, which the account-contract helper
// cannot decode because it assumes an object envelope.
func (c *accountContractClient) requestArray(method, path string, status int) []map[string]json.RawMessage {
	c.t.Helper()
	req, err := http.NewRequest(method, c.server.URL+path, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	response, err := c.client.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(response.Body)
	if response.StatusCode != status {
		c.t.Fatalf("%s %s: got %d want %d: %s", method, path, response.StatusCode, status, raw)
	}
	var result []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &result); err != nil {
		c.t.Fatal(err)
	}
	return result
}

func dashboardEventField(t *testing.T, row map[string]json.RawMessage, key string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(row[key], &value); err != nil {
		t.Fatalf("decode %s: %v", key, err)
	}
	return value
}

func findDashboardEventByName(t *testing.T, rows []map[string]json.RawMessage, name string) map[string]json.RawMessage {
	t.Helper()
	for _, row := range rows {
		if dashboardEventField(t, row, "name") == name {
			return row
		}
	}
	return nil
}

func createDashboardEvent(t *testing.T, client *accountContractClient, name string) string {
	t.Helper()
	created := client.request(http.MethodPost, "/api/events", canonicalTimedEventPayload(name), http.StatusCreated)
	eventID := decodeAccountString(t, created, "eventId")
	if eventID == "" {
		t.Fatal("signed-in creation did not return an event identifier")
	}
	t.Cleanup(func() {
		_, _ = pgstore.Pool.Exec(context.Background(), `DELETE FROM events WHERE short_id = $1`, eventID)
	})
	return eventID
}

// TestSignedInDashboardListsOwnedAndResponded proves that the dashboard
// returns events the account owns or has responded to, exposes the
// canonical public identifier for both, never reveals an event to an account
// that neither owns nor responded to, and lists an owned+responded event once.
func TestSignedInDashboardListsOwnedAndResponded(t *testing.T) {
	router := signedInEventRouter(t)
	owner, ownerAccount := createSignedInAccount(t, router)
	ownedName := "Dashboard owned event " + models.NewUUID().String()
	eventID := createDashboardEvent(t, owner, ownedName)

	ownedRow := findDashboardEventByName(t, owner.requestArray(http.MethodGet, "/api/user/events", http.StatusOK), ownedName)
	if ownedRow == nil {
		t.Fatal("owner dashboard did not list the owned event")
	}
	if got := dashboardEventField(t, ownedRow, "_id"); got != eventID {
		t.Fatalf("owned _id = %q, want canonical short id %q", got, eventID)
	}
	if got := dashboardEventField(t, ownedRow, "shortId"); got != eventID {
		t.Fatalf("owned shortId = %q, want %q", got, eventID)
	}
	if got := dashboardEventField(t, ownedRow, "ownerId"); got != ownerAccount.PlatformIdentityID {
		t.Fatalf("owned ownerId = %q, want owning account %q", got, ownerAccount.PlatformIdentityID)
	}

	responder, _ := createSignedInAccount(t, router)
	responder.request(http.MethodPost, "/api/events/"+eventID+"/response", map[string]any{
		"createResponse": true,
		"name":           "Dashboard responder",
		"availability":   []string{"2026-01-05T14:00:00Z"},
	}, http.StatusOK)
	respondedRow := findDashboardEventByName(t, responder.requestArray(http.MethodGet, "/api/user/events", http.StatusOK), ownedName)
	if respondedRow == nil {
		t.Fatal("responder dashboard did not list the responded event")
	}
	if got := dashboardEventField(t, respondedRow, "_id"); got != eventID {
		t.Fatalf("responded _id = %q, want canonical short id %q", got, eventID)
	}
	if got := dashboardEventField(t, respondedRow, "ownerId"); got != models.ZeroUUID().String() {
		t.Fatalf("responded ownerId = %q, want anonymous owner", got)
	}

	stranger, _ := createSignedInAccount(t, router)
	if row := findDashboardEventByName(t, stranger.requestArray(http.MethodGet, "/api/user/events", http.StatusOK), ownedName); row != nil {
		t.Fatal("dashboard revealed an event the account neither owns nor responded to")
	}

	// The owner responds too; the event must still appear exactly once.
	owner.request(http.MethodPost, "/api/events/"+eventID+"/response", map[string]any{
		"createResponse": true,
		"name":           "Dashboard owner response",
		"availability":   []string{"2026-01-05T14:00:00Z"},
	}, http.StatusOK)
	occurrences := 0
	for _, row := range owner.requestArray(http.MethodGet, "/api/user/events", http.StatusOK) {
		if dashboardEventField(t, row, "name") == ownedName {
			occurrences++
		}
	}
	if occurrences != 1 {
		t.Fatalf("dashboard listed the owned+responded event %d times, want 1", occurrences)
	}
}

// TestSignedInDashboardExcludesDeletedEvent proves that a deleted
// event stops appearing on the dashboard for its owner.
func TestSignedInDashboardExcludesDeletedEvent(t *testing.T) {
	router := signedInEventRouter(t)
	owner, _ := createSignedInAccount(t, router)
	name := "Deleted dashboard event " + models.NewUUID().String()
	eventID := createDashboardEvent(t, owner, name)

	owner.request(http.MethodDelete, "/api/events/"+eventID, nil, http.StatusOK)

	if row := findDashboardEventByName(t, owner.requestArray(http.MethodGet, "/api/user/events", http.StatusOK), name); row != nil {
		t.Fatal("deleted event still appeared on the dashboard")
	}
}
