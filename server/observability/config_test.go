package observability

import (
	"encoding/base64"
	"testing"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("OPENOBSERVE_ENDPOINT", "http://openobserve:5080/")
	t.Setenv("OPENOBSERVE_ORGANIZATION_ID", "org123")
	t.Setenv("OPENOBSERVE_INGEST_USERNAME", "service-account@timeful.fun")
	t.Setenv("OPENOBSERVE_INGEST_PASSWORD", "ingest-token")
	t.Setenv("APP_ENV", "staging")

	config := ConfigFromEnv()

	if config.Environment != "staging" {
		t.Fatalf("Environment = %q, want staging", config.Environment)
	}
	if !config.Enabled() {
		t.Fatal("Enabled() = false, want true for a complete contract")
	}

	logsURL, err := config.LogsURL()
	if err != nil {
		t.Fatalf("LogsURL() error = %v", err)
	}
	if want := "http://openobserve:5080/api/org123/v1/logs"; logsURL != want {
		t.Fatalf("LogsURL() = %q, want %q", logsURL, want)
	}
	metricsURL, err := config.MetricsURL()
	if err != nil {
		t.Fatalf("MetricsURL() error = %v", err)
	}
	if want := "http://openobserve:5080/api/org123/v1/metrics"; metricsURL != want {
		t.Fatalf("MetricsURL() = %q, want %q", metricsURL, want)
	}
	tracesURL, err := config.TracesURL()
	if err != nil {
		t.Fatalf("TracesURL() error = %v", err)
	}
	if want := "http://openobserve:5080/api/org123/v1/traces"; tracesURL != want {
		t.Fatalf("TracesURL() = %q, want %q", tracesURL, want)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("service-account@timeful.fun:ingest-token"))
	if auth := config.AuthHeader(); auth != wantAuth {
		t.Fatalf("AuthHeader() = %q, want %q", auth, wantAuth)
	}
}

func TestConfigFromEnvTrimsPasswordWhitespace(t *testing.T) {
	t.Setenv("OPENOBSERVE_ENDPOINT", "http://openobserve:5080")
	t.Setenv("OPENOBSERVE_ORGANIZATION_ID", "org123")
	t.Setenv("OPENOBSERVE_INGEST_USERNAME", "user")
	t.Setenv("OPENOBSERVE_INGEST_PASSWORD", "  ingest-token\n")

	config := ConfigFromEnv()
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:ingest-token"))
	if auth := config.AuthHeader(); auth != wantAuth {
		t.Fatalf("AuthHeader() = %q, want %q", auth, wantAuth)
	}
}

func TestConfigLogsURLRejectsQueryAndFragment(t *testing.T) {
	tests := []string{
		"https://observability.example.com?tenant=other",
		"https://observability.example.com#fragment",
	}
	for _, endpoint := range tests {
		t.Run(endpoint, func(t *testing.T) {
			config := Config{Endpoint: endpoint, OrganizationID: "org123"}
			if _, err := config.LogsURL(); err == nil {
				t.Fatalf("LogsURL() accepted %q", endpoint)
			}
		})
	}
}

func TestConfigLogsURLKeepsEndpointBasePath(t *testing.T) {
	config := Config{Endpoint: "https://observability.example.com/prefix", OrganizationID: "org123"}
	logsURL, err := config.LogsURL()
	if err != nil {
		t.Fatalf("LogsURL() error = %v", err)
	}
	if want := "https://observability.example.com/prefix/api/org123/v1/logs"; logsURL != want {
		t.Fatalf("LogsURL() = %q, want %q", logsURL, want)
	}
}

func TestConfigEnabledRequiresEveryValue(t *testing.T) {
	complete := Config{
		Endpoint:       "http://openobserve:5080",
		OrganizationID: "org123",
		Username:       "user",
		Password:       "token",
	}
	if !complete.Enabled() {
		t.Fatal("Enabled() = false for a complete config")
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "missing endpoint", mutate: func(c *Config) { c.Endpoint = "" }},
		{name: "missing organization", mutate: func(c *Config) { c.OrganizationID = "" }},
		{name: "missing username", mutate: func(c *Config) { c.Username = "" }},
		{name: "missing password", mutate: func(c *Config) { c.Password = "" }},
		{name: "blank password", mutate: func(c *Config) { c.Password = "  " }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := complete
			test.mutate(&config)
			if config.Enabled() {
				t.Fatal("Enabled() = true, want false for an incomplete config")
			}
		})
	}
}

func TestConfigLogsURLRejectsNonAbsoluteEndpoints(t *testing.T) {
	tests := []string{"openobserve:5080", "://missing-scheme", "ftp://openobserve:5080", ""}
	for _, endpoint := range tests {
		t.Run(endpoint, func(t *testing.T) {
			config := Config{Endpoint: endpoint, OrganizationID: "org123"}
			if _, err := config.LogsURL(); err == nil {
				t.Fatalf("LogsURL() accepted %q", endpoint)
			}
		})
	}
}

func TestConfigLogsURLEscapesOrganizationIdentifier(t *testing.T) {
	config := Config{Endpoint: "https://observability.example.com", OrganizationID: "org id"}
	logsURL, err := config.LogsURL()
	if err != nil {
		t.Fatalf("LogsURL() error = %v", err)
	}
	if want := "https://observability.example.com/api/org%20id/v1/logs"; logsURL != want {
		t.Fatalf("LogsURL() = %q, want %q", logsURL, want)
	}
}

func TestConfigSignalURLsRejectInvalidEndpoints(t *testing.T) {
	signals := []struct {
		name  string
		build func(Config) (string, error)
	}{
		{name: "logs", build: Config.LogsURL},
		{name: "metrics", build: Config.MetricsURL},
		{name: "traces", build: Config.TracesURL},
	}
	endpoints := []string{
		"openobserve:5080",
		"://missing-scheme",
		"ftp://openobserve:5080",
		"",
		"https://observability.example.com?tenant=other",
		"https://observability.example.com#fragment",
	}
	for _, signal := range signals {
		for _, endpoint := range endpoints {
			t.Run(signal.name+" "+endpoint, func(t *testing.T) {
				config := Config{Endpoint: endpoint, OrganizationID: "org123"}
				if _, err := signal.build(config); err == nil {
					t.Fatalf("%s URL accepted %q", signal.name, endpoint)
				}
			})
		}
	}
}

func TestConfigSignalURLsKeepEndpointBasePath(t *testing.T) {
	config := Config{Endpoint: "https://observability.example.com/prefix", OrganizationID: "org id"}
	signals := []struct {
		name  string
		build func(Config) (string, error)
		want  string
	}{
		{name: "logs", build: Config.LogsURL, want: "https://observability.example.com/prefix/api/org%20id/v1/logs"},
		{name: "metrics", build: Config.MetricsURL, want: "https://observability.example.com/prefix/api/org%20id/v1/metrics"},
		{name: "traces", build: Config.TracesURL, want: "https://observability.example.com/prefix/api/org%20id/v1/traces"},
	}
	for _, signal := range signals {
		t.Run(signal.name, func(t *testing.T) {
			got, err := signal.build(config)
			if err != nil {
				t.Fatalf("%s URL error = %v", signal.name, err)
			}
			if got != signal.want {
				t.Fatalf("%s URL = %q, want %q", signal.name, got, signal.want)
			}
		})
	}
}
