package providerconfig

import "testing"

func TestEndpointsDefaultToRealProviders(t *testing.T) {
	t.Setenv(TestGoogleOAuthTokenEndpoint, "")
	t.Setenv(TestGoogleCalendarAPIBaseURL, "")
	t.Setenv(TestMicrosoftOAuthTokenEndpoint, "")
	t.Setenv(TestMicrosoftGraphAPIBaseURL, "")

	if got := GoogleOAuthTokenEndpoint(); got != realGoogleOAuthTokenEndpoint {
		t.Fatalf("GoogleOAuthTokenEndpoint() = %q, want %q", got, realGoogleOAuthTokenEndpoint)
	}
	if got := GoogleCalendarAPIBaseURL(); got != realGoogleCalendarAPIBaseURL {
		t.Fatalf("GoogleCalendarAPIBaseURL() = %q, want %q", got, realGoogleCalendarAPIBaseURL)
	}
	if got := MicrosoftOAuthTokenEndpoint(); got != realMicrosoftOAuthTokenEndpoint {
		t.Fatalf("MicrosoftOAuthTokenEndpoint() = %q, want %q", got, realMicrosoftOAuthTokenEndpoint)
	}
	if got := MicrosoftGraphAPIBaseURL(); got != realMicrosoftGraphAPIBaseURL {
		t.Fatalf("MicrosoftGraphAPIBaseURL() = %q, want %q", got, realMicrosoftGraphAPIBaseURL)
	}
}

func TestEndpointsHonorTestOverrides(t *testing.T) {
	t.Setenv(TestGoogleOAuthTokenEndpoint, " http://calendar-mock:8099/google/token ")
	t.Setenv(TestGoogleCalendarAPIBaseURL, "http://calendar-mock:8099/google/calendar/v3/")
	t.Setenv(TestMicrosoftOAuthTokenEndpoint, "http://calendar-mock:8099/microsoft/token")
	t.Setenv(TestMicrosoftGraphAPIBaseURL, "http://calendar-mock:8099/microsoft/graph/v1.0/")

	if got := GoogleOAuthTokenEndpoint(); got != "http://calendar-mock:8099/google/token" {
		t.Fatalf("GoogleOAuthTokenEndpoint() = %q, want the trimmed override", got)
	}
	if got := GoogleCalendarAPIBaseURL(); got != "http://calendar-mock:8099/google/calendar/v3" {
		t.Fatalf("GoogleCalendarAPIBaseURL() = %q, want the override without a trailing slash", got)
	}
	if got := MicrosoftOAuthTokenEndpoint(); got != "http://calendar-mock:8099/microsoft/token" {
		t.Fatalf("MicrosoftOAuthTokenEndpoint() = %q, want the override", got)
	}
	if got := MicrosoftGraphAPIBaseURL(); got != "http://calendar-mock:8099/microsoft/graph/v1.0" {
		t.Fatalf("MicrosoftGraphAPIBaseURL() = %q, want the override without a trailing slash", got)
	}
}

func TestWhitespaceOnlyOverrideFallsBackToRealProvider(t *testing.T) {
	t.Setenv(TestGoogleOAuthTokenEndpoint, "   ")
	t.Setenv(TestMicrosoftGraphAPIBaseURL, "\t")

	if got := GoogleOAuthTokenEndpoint(); got != realGoogleOAuthTokenEndpoint {
		t.Fatalf("GoogleOAuthTokenEndpoint() = %q, want the real default", got)
	}
	if got := MicrosoftGraphAPIBaseURL(); got != realMicrosoftGraphAPIBaseURL {
		t.Fatalf("MicrosoftGraphAPIBaseURL() = %q, want the real default", got)
	}
}
