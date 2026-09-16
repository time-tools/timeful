// Package observability ships structured server diagnostic records to
// OpenObserve over OpenTelemetry OTLP. Export runs on a background batch
// pipeline, so it never blocks or fails request handling, and the existing
// local diagnostic output stays authoritative while OpenObserve is
// unavailable.
package observability

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const (
	endpointEnvironment       = "OPENOBSERVE_ENDPOINT"
	organizationEnvironment   = "OPENOBSERVE_ORGANIZATION_ID"
	ingestUsernameEnvironment = "OPENOBSERVE_INGEST_USERNAME"
	ingestPasswordEnvironment = "OPENOBSERVE_INGEST_PASSWORD"
	appEnvironment            = "APP_ENV"

	// ServerLogStream is the OpenObserve stream that receives the server's
	// structured diagnostic records.
	ServerLogStream = "timeful_server_logs"
)

// Config is the environment-scoped OpenObserve ingest contract established by
// the deployment environment files.
type Config struct {
	Endpoint       string
	OrganizationID string
	Username       string
	Password       string
	Environment    string
}

// ConfigFromEnv reads the server-side OpenObserve contract from the process
// environment. Compose injects it from the selected app env file.
func ConfigFromEnv() Config {
	return Config{
		Endpoint:       strings.TrimSpace(os.Getenv(endpointEnvironment)),
		OrganizationID: strings.TrimSpace(os.Getenv(organizationEnvironment)),
		Username:       strings.TrimSpace(os.Getenv(ingestUsernameEnvironment)),
		Password:       strings.TrimSpace(os.Getenv(ingestPasswordEnvironment)),
		Environment:    strings.TrimSpace(os.Getenv(appEnvironment)),
	}
}

// Enabled reports whether the contract is complete enough to export. Missing
// or blank values disable export instead of failing startup, so local
// diagnostics remain available without OpenObserve credentials.
func (c Config) Enabled() bool {
	return c.Endpoint != "" &&
		c.OrganizationID != "" &&
		c.Username != "" &&
		strings.TrimSpace(c.Password) != ""
}

// LogsURL returns the OTLP/HTTP logs endpoint for the configured
// organization.
func (c Config) LogsURL() (string, error) {
	parsed, err := url.Parse(c.Endpoint)
	if err != nil {
		return "", fmt.Errorf("OPENOBSERVE_ENDPOINT must be a valid URL: %w", err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("OPENOBSERVE_ENDPOINT must be an absolute http(s) URL, got %q", c.Endpoint)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("OPENOBSERVE_ENDPOINT must not carry a query or fragment, got %q", c.Endpoint)
	}
	return parsed.JoinPath("api", c.OrganizationID, "v1", "logs").String(), nil
}

// AuthHeader returns the Basic authorization header OpenObserve expects for
// the environment's ingest service account.
func (c Config) AuthHeader() string {
	credentials := c.Username + ":" + c.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}
