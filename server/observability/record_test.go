package observability

import (
	"strings"
	"testing"
	"time"
)

func TestRequestCompletionLogRecordCarriesDiagnosticContext(t *testing.T) {
	started := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	completion := RequestCompletion{
		RequestID:    "3f8c1f6e-2f0e-4e2a-9d7a-1c2b3a4d5e6f",
		Method:       "GET",
		Route:        "/api/events/:eventId",
		Status:       503,
		Latency:      1500 * time.Microsecond,
		ErrorType:    "http_5xx",
		ErrorMessage: "postgres unavailable",
		Readiness:    ReadinessUnavailable,
		StartTime:    started,
	}

	record := completion.LogRecord()
	if record.EventName() != "http.request.completed" {
		t.Fatalf("EventName() = %q", record.EventName())
	}
	if record.SeverityText() != "ERROR" {
		t.Fatalf("SeverityText() = %q, want ERROR", record.SeverityText())
	}
	if got := record.Body().AsString(); got != "request completed" {
		t.Fatalf("Body() = %q", got)
	}
	if want := started.Add(1500 * time.Microsecond); !record.Timestamp().Equal(want) {
		t.Fatalf("Timestamp() = %v, want %v", record.Timestamp(), want)
	}

	attributes := apiRecordAttributes(record)
	want := map[string]string{
		"request.id":                "3f8c1f6e-2f0e-4e2a-9d7a-1c2b3a4d5e6f",
		"http.request.method":       "GET",
		"http.route":                "/api/events/:eventId",
		"http.response.status_code": "503",
		"http.response.outcome":     "server_error",
		"http.response.duration_ms": "1.5",
		"error.type":                "http_5xx",
		"error.message":             "postgres unavailable",
		"service.readiness":         "unavailable",
	}
	for key, value := range want {
		if got := attributes[key]; got != value {
			t.Fatalf("attribute %s = %q, want %q", key, got, value)
		}
	}
}

func TestRequestCompletionOutcome(t *testing.T) {
	tests := []struct {
		status      int
		wantOutcome string
		wantSev     string
	}{
		{status: 200, wantOutcome: "success", wantSev: "INFO"},
		{status: 204, wantOutcome: "success", wantSev: "INFO"},
		{status: 404, wantOutcome: "client_error", wantSev: "WARN"},
		{status: 500, wantOutcome: "server_error", wantSev: "ERROR"},
	}
	for _, test := range tests {
		completion := RequestCompletion{Status: test.status}
		record := completion.LogRecord()
		attributes := apiRecordAttributes(record)
		if got := attributes["http.response.outcome"]; got != test.wantOutcome {
			t.Fatalf("status %d outcome = %q, want %q", test.status, got, test.wantOutcome)
		}
		if got := record.SeverityText(); got != test.wantSev {
			t.Fatalf("status %d severity = %q, want %q", test.status, got, test.wantSev)
		}
	}
}

func TestRequestCompletionLogRecordRedactsErrorContext(t *testing.T) {
	completion := RequestCompletion{
		RequestID:    "request-1",
		Status:       500,
		ErrorMessage: "call failed with Authorization: Bearer supersecrettoken",
	}
	attributes := apiRecordAttributes(completion.LogRecord())
	if strings.Contains(attributes["error.message"], "supersecrettoken") {
		t.Fatalf("error.message leaked a credential: %q", attributes["error.message"])
	}
}

func TestRequestCompletionLogRecordDefaultsReadinessToUnknown(t *testing.T) {
	attributes := apiRecordAttributes(RequestCompletion{RequestID: "request-1", Status: 200}.LogRecord())
	if got := attributes["service.readiness"]; got != "unknown" {
		t.Fatalf("service.readiness = %q, want unknown", got)
	}
	if _, exists := attributes["error.type"]; exists {
		t.Fatal("error.type should be absent for a successful request")
	}
}
