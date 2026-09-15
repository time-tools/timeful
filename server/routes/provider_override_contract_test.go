package routes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"timeful/server/accounts"
	"timeful/server/models"
	"timeful/server/services/providerconfig"
)

func writeProviderOverrideJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

// providerOverrideIDToken builds the unsigned three-part token the Google
// add-account path decodes. Only the email claim is read.
func providerOverrideIDToken() string {
	header, _ := json.Marshal(map[string]string{"typ": "JWT", "alg": "none"})
	claims, _ := json.Marshal(map[string]string{"email": "provider-override@example.invalid"})
	return base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(claims) + ".signature"
}

// TestGetUserCalendarsLoadsGoogleSubCalendarsFromTestProviderOverride proves
// that the test-only provider endpoint overrides route the running server to a
// mock provider: adding a Google account exchanges its code and loads its
// sub-calendars there, and GET /user/calendars returns without any live provider
// network access. It also proves the sub-calendar set is persisted.
func TestGetUserCalendarsLoadsGoogleSubCalendarsFromTestProviderOverride(t *testing.T) {
	var tokenRequests, calendarListRequests atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("/google/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests.Add(1)
		writeProviderOverrideJSON(w, map[string]any{
			"access_token":  "mock-access-token",
			"id_token":      providerOverrideIDToken(),
			"refresh_token": "mock-refresh-token",
			"expires_in":    3600,
			"scope":         "scope",
			"token_type":    "Bearer",
		})
	})
	mux.HandleFunc("/google/calendar/v3/users/me/calendarList", func(w http.ResponseWriter, r *http.Request) {
		calendarListRequests.Add(1)
		writeProviderOverrideJSON(w, map[string]any{
			"items": []map[string]any{
				{"id": "primary", "summary": "Primary", "selected": true},
				{"id": "work", "summary": "Work", "selected": false},
			},
		})
	})
	mux.HandleFunc("/google/calendar/v3/calendars/", func(w http.ResponseWriter, r *http.Request) {
		writeProviderOverrideJSON(w, map[string]any{"items": []any{}})
	})
	mock := httptest.NewServer(mux)
	t.Cleanup(mock.Close)

	t.Setenv(providerconfig.TestGoogleOAuthTokenEndpoint, mock.URL+"/google/token")
	t.Setenv(providerconfig.TestGoogleCalendarAPIBaseURL, mock.URL+"/google/calendar/v3")

	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)

	email := "provider-override-" + models.NewUUID().String() + "@example.com"
	profile := verifyOtpSignIn(t, client, email, "123456")
	externalUserID := decodeAccountString(t, profile, "_id")

	account, err := accounts.Lookup(context.Background(), externalUserID)
	if err != nil {
		t.Fatalf("look up signed-in account: %v", err)
	}
	cleanupOtpAccount(t, account)

	client.request(http.MethodPost, "/api/user/add-google-calendar-account", map[string]any{
		"code": "mock-authorization-code", "scope": "scope",
	}, http.StatusOK)

	timeMin := time.Now().UTC().Format(time.RFC3339)
	timeMax := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	client.request(
		http.MethodGet,
		"/api/user/calendars?timeMin="+url.QueryEscape(timeMin)+"&timeMax="+url.QueryEscape(timeMax),
		nil,
		http.StatusOK,
	)

	if tokenRequests.Load() == 0 {
		t.Fatal("the OAuth token endpoint override was never reached")
	}
	if calendarListRequests.Load() == 0 {
		t.Fatal("the calendar API override was never reached")
	}

	integrations, err := accounts.LoadCalendarIntegrations(context.Background(), externalUserID)
	if err != nil {
		t.Fatalf("load calendar integrations: %v", err)
	}
	var google *models.CalendarAccount
	for _, candidate := range integrations.Accounts {
		if candidate.CalendarType == models.GoogleCalendarType {
			account := candidate
			google = &account
		}
	}
	if google == nil || google.SubCalendars == nil {
		t.Fatalf("Google account with sub-calendars was not persisted: %#v", integrations.Accounts)
	}
	primary, ok := (*google.SubCalendars)["primary"]
	if !ok {
		t.Fatalf("primary sub-calendar missing from %#v", *google.SubCalendars)
	}
	if primary.Name != "Primary" || primary.Enabled == nil || !*primary.Enabled {
		t.Fatalf("primary sub-calendar = %#v, want selected Primary", primary)
	}
	work, ok := (*google.SubCalendars)["work"]
	if !ok {
		t.Fatalf("work sub-calendar missing from %#v", *google.SubCalendars)
	}
	if work.Name != "Work" || work.Enabled == nil || *work.Enabled {
		t.Fatalf("work sub-calendar = %#v, want unselected Work", work)
	}
}
