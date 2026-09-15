package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"timeful/server/eventid"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// signedInEventRouter exposes the account-contract sign-in and cleanup
// helpers with the event routes, which always create supported kinds.
func signedInEventRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return newAccountEventContractRouter(t)
}

func createSignedInAccount(t *testing.T, router *gin.Engine) (*accountContractClient, *pgstore.Account) {
	t.Helper()
	client := newAccountContractClient(t, router)
	email := "signed-in-event-" + models.NewUUID().String() + "@example.com"
	verifyOtpSignIn(t, client, email, "123456")
	account, err := repositoryForTest(t).GetAccountByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("signed-in account not created: %v", err)
	}
	cleanupOtpAccount(t, account)
	return client, account
}

func responseMapKeys(t *testing.T, payload map[string]json.RawMessage) []string {
	t.Helper()
	var rows map[string]json.RawMessage
	if err := json.Unmarshal(payload["responses"], &rows); err != nil {
		t.Fatalf("decode responses: %v", err)
	}
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	return keys
}

// TestSignedInEventLifecycle proves that a supported signed-in poll is
// stored with the account as owner, that the usage counter advances,
// that the owner can manage settings, schedule, archive, and delete with the
// session alone, and that another account is rejected.
func TestSignedInEventLifecycle(t *testing.T) {
	router := signedInEventRouter(t)
	owner, account := createSignedInAccount(t, router)
	ctx := context.Background()

	payload := canonicalTimedEventPayload("Signed-in event")
	created := owner.request(http.MethodPost, "/api/events", payload, http.StatusCreated)
	eventID := decodeAccountString(t, created, "eventId")
	if eventID == "" {
		t.Fatal("signed-in creation did not return an event identifier")
	}
	if !eventid.Canonical(eventID) {
		t.Fatalf("expected a canonical event identifier, got %q", eventID)
	}
	t.Cleanup(func() {
		_, _ = pgstore.Pool.Exec(context.Background(), `DELETE FROM events WHERE short_id = $1`, eventID)
	})

	repository := repositoryForTest(t)
	stored, err := repository.GetEventByShortID(ctx, eventID)
	if err != nil {
		t.Fatalf("signed-in event not stored: %v", err)
	}
	if stored.OwnerPlatformIdentityID == nil {
		t.Fatal("expected account ownership association on the event")
	}
	if stored.OwnerPlatformIdentityID == nil || *stored.OwnerPlatformIdentityID != account.PlatformIdentityID {
		t.Fatalf("owner platform identity = %v, want %q", stored.OwnerPlatformIdentityID, account.PlatformIdentityID)
	}
	reloaded, err := repository.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.NumEventsCreated != account.NumEventsCreated+1 {
		t.Fatalf("usage counter = %d, want %d", reloaded.NumEventsCreated, account.NumEventsCreated+1)
	}

	path := "/api/events/" + eventID
	owner.request(http.MethodPut, path, canonicalTimedEventPayload("Edited signed-in event"), http.StatusOK)
	owner.request(http.MethodPut, path+"/schedule", map[string]string{"startDate": "2026-01-05T14:00:00Z", "endDate": "2026-01-05T15:00:00Z"}, http.StatusOK)
	owner.request(http.MethodDelete, path+"/schedule", nil, http.StatusOK)
	owner.request(http.MethodPost, path+"/archive", map[string]bool{"archive": true}, http.StatusOK)
	// Archived events are read-only until unarchived.
	owner.request(http.MethodPut, path, canonicalTimedEventPayload("Blocked while archived"), http.StatusForbidden)
	owner.request(http.MethodPost, path+"/archive", map[string]bool{"archive": false}, http.StatusOK)

	stranger, _ := createSignedInAccount(t, router)
	stranger.request(http.MethodPut, path, canonicalTimedEventPayload("Hijacked"), http.StatusForbidden)
	stranger.request(http.MethodPost, path+"/archive", map[string]bool{"archive": true}, http.StatusForbidden)
	stranger.request(http.MethodDelete, path, nil, http.StatusForbidden)

	owner.request(http.MethodDelete, path, nil, http.StatusOK)
	owner.request(http.MethodGet, path, nil, http.StatusNotFound)
}

// TestSignedInResponseAssociation proves that a signed-in response is
// associated with the account, that the account recovers it from a fresh
// browser, and that blind availability still hides it from non-owners.
func TestSignedInResponseAssociation(t *testing.T) {
	router := signedInEventRouter(t)
	owner, account := createSignedInAccount(t, router)
	ctx := context.Background()

	payload := canonicalTimedEventPayload("Signed-in response association")
	payload["blindAvailabilityEnabled"] = true
	created := owner.request(http.MethodPost, "/api/events", payload, http.StatusCreated)
	eventID := decodeAccountString(t, created, "eventId")
	t.Cleanup(func() {
		_, _ = pgstore.Pool.Exec(context.Background(), `DELETE FROM events WHERE short_id = $1`, eventID)
	})
	path := "/api/events/" + eventID

	createdResponse := owner.request(http.MethodPost, path+"/response", map[string]any{
		"createResponse": true,
		"name":           "Owner Display Name",
		"availability":   []string{"2026-01-05T14:00:00Z"},
	}, http.StatusOK)
	responseID := decodeAccountString(t, createdResponse, "responseId")
	if responseID == "" {
		t.Fatal("signed-in response creation did not return a response identifier")
	}

	stored, err := repositoryForTest(t).GetEventByShortID(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	var associated bool
	if err := pgstore.Pool.QueryRow(ctx, `SELECT EXISTS (
 SELECT 1 FROM event_visitor_identities
 WHERE event_id = $1 AND platform_identity_id = $2)`, stored.ID, account.PlatformIdentityID).Scan(&associated); err != nil {
		t.Fatal(err)
	}
	if !associated {
		t.Fatal("signed-in response visitor not associated with the account")
	}

	// Blind availability hides a non-owner response and its count.
	stranger, _ := createSignedInAccount(t, router)
	strangerEvent := stranger.request(http.MethodGet, path, nil, http.StatusOK)
	if keys := responseMapKeys(t, strangerEvent); len(keys) != 0 {
		t.Fatalf("blind non-owner saw %d responses", len(keys))
	}
	if _, leaked := strangerEvent["numResponses"]; leaked {
		t.Fatal("blind non-owner saw the response count")
	}

	// The account recovers its response from a fresh browser using the session
	// alone: no credential cookie is carried over.
	recovered := newAccountContractClient(t, router)
	recovered.request(http.MethodPost, "/test/account-contract/sign-in/"+account.PlatformIdentityID, nil, http.StatusOK)
	recoveredEvent := recovered.request(http.MethodGet, path, nil, http.StatusOK)
	if keys := responseMapKeys(t, recoveredEvent); len(keys) != 1 || keys[0] != responseID {
		t.Fatalf("account could not recover its response, got %v", keys)
	}
	recovered.request(http.MethodPost, path+"/response", map[string]string{"responseId": responseID, "name": "Recovered Name"}, http.StatusOK)
	recovered.request(http.MethodPost, path+"/rename-user", map[string]string{"responseId": responseID, "newName": "Renamed"}, http.StatusOK)
	recovered.request(http.MethodDelete, path+"/response", map[string]string{"responseId": responseID}, http.StatusOK)

	afterDelete := recovered.request(http.MethodGet, path, nil, http.StatusOK)
	if keys := responseMapKeys(t, afterDelete); len(keys) != 0 {
		t.Fatalf("response survived deletion: %v", keys)
	}
}
