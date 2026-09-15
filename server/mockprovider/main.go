// Command mockprovider is a test-only HTTP server that stands in for the
// external OAuth and calendar providers in the isolated Compose test stack.
//
// It is intentionally registered only in compose.test.yaml, shareable with the
// server-test container over the test network, and never part of production
// Compose. The server reaches it through the TEST_-prefixed provider endpoint
// overrides documented in docs/environments.md.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

const (
	defaultPort    = "8099"
	mockEmail      = "calendar-mock@example.invalid"
	accessTokenTTL = 3600
)

var tokenSequence atomic.Int64

func main() {
	port := strings.TrimSpace(os.Getenv("MOCK_PROVIDER_PORT"))
	if port == "" {
		port = defaultPort
	}

	// The module targets Go 1.20, so this uses classic ServeMux path matching
	// and checks the HTTP method inside each handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", requireMethod(http.MethodGet, writeJSON(map[string]string{"status": "ok"})))
	mux.HandleFunc("/google/token", requireMethod(http.MethodPost, handleToken))
	mux.HandleFunc("/microsoft/token", requireMethod(http.MethodPost, handleToken))
	mux.HandleFunc("/google/calendar/v3/users/me/calendarList", requireMethod(http.MethodGet, handleGoogleCalendarList))
	mux.HandleFunc("/google/calendar/v3/calendars/", requireMethod(http.MethodGet, handleEmptyGoogleEvents))
	mux.HandleFunc("/microsoft/graph/v1.0/me", requireMethod(http.MethodGet, handleMicrosoftUserInfo))
	mux.HandleFunc("/microsoft/graph/v1.0/me/calendars", requireMethod(http.MethodGet, handleMicrosoftCalendarList))
	mux.HandleFunc("/microsoft/graph/v1.0/me/calendars/", requireMethod(http.MethodGet, handleEmptyMicrosoftEvents))

	address := ":" + port
	log.Printf("mock provider listening on %s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("mock provider server failed: %v", err)
	}
}

func requireMethod(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func writeJSON(body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(body); err != nil {
			log.Printf("mock provider encode response: %v", err)
		}
	}
}

// handleToken serves both the Google and Microsoft token exchange. It returns a
// monotonically increasing access token so a refresh persists an observably new
// value, and a valid three-part id_token so the Google add-account path can read
// the provider email without signature verification.
func handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	accessToken := fmt.Sprintf("mock-access-token-%d", tokenSequence.Add(1))
	response := map[string]any{
		"access_token": accessToken,
		"expires_in":   accessTokenTTL,
		"token_type":   "Bearer",
		"scope":        "openid email https://www.googleapis.com/auth/calendar",
	}

	if r.FormValue("grant_type") != "refresh_token" {
		response["refresh_token"] = "mock-refresh-token"
		response["id_token"] = mockIDToken()
	}

	writeJSON(response)(w, r)
}

func handleGoogleCalendarList(w http.ResponseWriter, r *http.Request) {
	writeJSON(map[string]any{
		"items": []map[string]any{
			{"id": "primary", "summary": "Primary", "selected": true},
			{"id": "work", "summary": "Work", "selected": false},
		},
	})(w, r)
}

func handleEmptyGoogleEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(map[string]any{"items": []any{}})(w, r)
}

func handleMicrosoftUserInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(map[string]any{
		"givenName": "Mock",
		"surname":   "Calendar",
		"mail":      mockEmail,
	})(w, r)
}

func handleMicrosoftCalendarList(w http.ResponseWriter, r *http.Request) {
	writeJSON(map[string]any{
		"value": []map[string]any{
			{"id": "primary", "name": "Primary"},
			{"id": "work", "name": "Work"},
		},
	})(w, r)
}

func handleEmptyMicrosoftEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(map[string]any{"value": []any{}})(w, r)
}

// mockIDToken builds an unsigned three-part token accepted by the server's
// sjwt.Parse, carrying only the provider email the add-calendar path reads.
func mockIDToken() string {
	header, _ := json.Marshal(map[string]string{"typ": "JWT", "alg": "none"})
	claims, _ := json.Marshal(map[string]string{"email": mockEmail})
	return fmt.Sprintf(
		"%s.%s.signature",
		base64.RawURLEncoding.EncodeToString(header),
		base64.RawURLEncoding.EncodeToString(claims),
	)
}
