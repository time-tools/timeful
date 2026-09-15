// Package providerconfig resolves the outbound OAuth and calendar provider
// endpoints used by the server.
//
// The real provider endpoints are always the default. Each endpoint can be
// redirected through a TEST_-prefixed environment variable so the isolated test
// stack can reach a mock provider instead of the public internet. The overrides
// are intentionally named after the test stack: production and staging must
// never set them, and leaving them unset keeps the real provider URLs.
package providerconfig

import (
	"os"
	"strings"
)

const (
	// TestGoogleOAuthTokenEndpoint overrides the Google OAuth2 token endpoint.
	TestGoogleOAuthTokenEndpoint = "TEST_GOOGLE_OAUTH_TOKEN_ENDPOINT"
	// TestGoogleCalendarAPIBaseURL overrides the Google Calendar API base URL.
	TestGoogleCalendarAPIBaseURL = "TEST_GOOGLE_CALENDAR_API_BASE_URL"
	// TestMicrosoftOAuthTokenEndpoint overrides the Microsoft OAuth2 token endpoint.
	TestMicrosoftOAuthTokenEndpoint = "TEST_MICROSOFT_OAUTH_TOKEN_ENDPOINT"
	// TestMicrosoftGraphAPIBaseURL overrides the Microsoft Graph API base URL.
	TestMicrosoftGraphAPIBaseURL = "TEST_MICROSOFT_GRAPH_API_BASE_URL"
)

const (
	realGoogleOAuthTokenEndpoint    = "https://oauth2.googleapis.com/token"
	realGoogleCalendarAPIBaseURL    = "https://www.googleapis.com/calendar/v3"
	realMicrosoftOAuthTokenEndpoint = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	realMicrosoftGraphAPIBaseURL    = "https://graph.microsoft.com/v1.0"
)

// overrideOrDefault returns the trimmed environment value when it is non-blank,
// and the real provider endpoint otherwise. Trimming keeps a whitespace-only
// value from being treated as a deliberate override.
func overrideOrDefault(envName string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		return value
	}
	return fallback
}

// GoogleOAuthTokenEndpoint returns the Google OAuth2 token endpoint.
func GoogleOAuthTokenEndpoint() string {
	return overrideOrDefault(TestGoogleOAuthTokenEndpoint, realGoogleOAuthTokenEndpoint)
}

// GoogleCalendarAPIBaseURL returns the Google Calendar API base URL without a
// trailing slash.
func GoogleCalendarAPIBaseURL() string {
	return strings.TrimRight(
		overrideOrDefault(TestGoogleCalendarAPIBaseURL, realGoogleCalendarAPIBaseURL),
		"/",
	)
}

// MicrosoftOAuthTokenEndpoint returns the Microsoft OAuth2 token endpoint.
func MicrosoftOAuthTokenEndpoint() string {
	return overrideOrDefault(TestMicrosoftOAuthTokenEndpoint, realMicrosoftOAuthTokenEndpoint)
}

// MicrosoftGraphAPIBaseURL returns the Microsoft Graph API base URL without a
// trailing slash.
func MicrosoftGraphAPIBaseURL() string {
	return strings.TrimRight(
		overrideOrDefault(TestMicrosoftGraphAPIBaseURL, realMicrosoftGraphAPIBaseURL),
		"/",
	)
}
