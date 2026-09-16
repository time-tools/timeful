package observability

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	collogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type receivedExport struct {
	path       string
	auth       string
	streamName string
	request    *collogpb.ExportLogsServiceRequest
}

func TestStartShipsRecordsToConfiguredOrganization(t *testing.T) {
	exports := make(chan receivedExport, 4)
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read export body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		request := &collogpb.ExportLogsServiceRequest{}
		if err := proto.Unmarshal(body, request); err != nil {
			t.Errorf("decode export body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		exports <- receivedExport{
			path:       r.URL.Path,
			auth:       r.Header.Get("Authorization"),
			streamName: r.Header.Get("stream-name"),
			request:    request,
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()

	config := Config{
		Endpoint:       collector.URL,
		OrganizationID: "org-abc",
		Username:       "ingest@timeful.fun",
		Password:       "ingest-token",
		Environment:    "test",
	}
	recorder, err := Start(context.Background(), config)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !recorder.Enabled() {
		t.Fatal("recorder is disabled after a complete Start()")
	}

	recorder.EmitRequest(context.Background(), RequestCompletion{
		RequestID:    "request-1",
		Method:       "GET",
		Route:        "/api/events/:eventId",
		Status:       500,
		Latency:      25 * time.Millisecond,
		ErrorType:    "event-not-found",
		ErrorMessage: "Authorization: Bearer supersecrettoken",
		Readiness:    ReadinessUnavailable,
		StartTime:    time.Now().Add(-25 * time.Millisecond),
	})

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := recorder.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	select {
	case export := <-exports:
		assertExportRequest(t, export, config)
	case <-time.After(2 * time.Second):
		t.Fatal("no OTLP export reached the collector")
	}
}

func assertExportRequest(t *testing.T, export receivedExport, config Config) {
	t.Helper()

	if want := "/api/org-abc/v1/logs"; export.path != want {
		t.Fatalf("export path = %q, want %q", export.path, want)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("ingest@timeful.fun:ingest-token"))
	if export.auth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", export.auth, wantAuth)
	}
	if export.streamName != ServerLogStream {
		t.Fatalf("stream-name = %q, want %q", export.streamName, ServerLogStream)
	}

	if len(export.request.ResourceLogs) != 1 {
		t.Fatalf("ResourceLogs length = %d, want 1", len(export.request.ResourceLogs))
	}
	resourceLogs := export.request.ResourceLogs[0]
	resource := anyValuesToStrings(resourceLogs.GetResource().GetAttributes())
	if resource["service.name"] != "timeful-server" {
		t.Fatalf("service.name = %q", resource["service.name"])
	}
	if resource["deployment.environment"] != "test" {
		t.Fatalf("deployment.environment = %q", resource["deployment.environment"])
	}

	if len(resourceLogs.ScopeLogs) != 1 || len(resourceLogs.ScopeLogs[0].LogRecords) != 1 {
		t.Fatalf("scope logs = %#v, want one record", resourceLogs.ScopeLogs)
	}
	record := resourceLogs.ScopeLogs[0].LogRecords[0]
	if record.GetEventName() != "http.request.completed" {
		t.Fatalf("event name = %q", record.GetEventName())
	}
	if record.GetBody().GetStringValue() != "request completed" {
		t.Fatalf("body = %q", record.GetBody().GetStringValue())
	}
	if record.GetSeverityText() != "ERROR" {
		t.Fatalf("severity = %q", record.GetSeverityText())
	}
	attributes := anyValuesToStrings(record.GetAttributes())
	want := map[string]string{
		"request.id":                "request-1",
		"http.request.method":       "GET",
		"http.route":                "/api/events/:eventId",
		"http.response.status_code": "500",
		"http.response.outcome":     "server_error",
		"http.response.duration_ms": "25",
		"error.type":                "event-not-found",
		"service.readiness":         "unavailable",
	}
	for key, value := range want {
		if attributes[key] != value {
			t.Fatalf("attribute %s = %q, want %q", key, attributes[key], value)
		}
	}

	for _, leaked := range []string{"ingest-token", "ingest@timeful.fun", "supersecrettoken"} {
		if strings.Contains(prototext.Format(export.request), leaked) {
			t.Fatalf("exported payload leaked %q", leaked)
		}
	}
}

func TestStartDisabledConfigIsNoop(t *testing.T) {
	recorder, err := Start(context.Background(), Config{})
	if err == nil {
		t.Fatal("Start() succeeded with an incomplete contract")
	}
	if recorder != nil {
		t.Fatal("Start() returned a recorder for an incomplete contract")
	}

	recorder.SetReadiness(func() ReadinessState { return ReadinessReady })
	recorder.EmitRequest(context.Background(), RequestCompletion{RequestID: "request-1"})
	if err := recorder.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() on a disabled recorder error = %v", err)
	}
}

func TestStartRejectsInvalidEndpoint(t *testing.T) {
	_, err := Start(context.Background(), Config{
		Endpoint:       "openobserve:5080",
		OrganizationID: "org-abc",
		Username:       "user",
		Password:       "token",
	})
	if err == nil {
		t.Fatal("Start() accepted a non-absolute endpoint")
	}
}

func TestStateErrorTypeMapping(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{status: 400, want: "http_4xx"},
		{status: 404, want: "http_4xx"},
		{status: 500, want: "http_5xx"},
		{status: 503, want: "http_5xx"},
		{status: 200, want: ""},
	}
	for _, test := range tests {
		if got := statusErrorType(test.status); got != test.want {
			t.Fatalf("statusErrorType(%d) = %q, want %q", test.status, got, test.want)
		}
	}
}

func anyValuesToStrings(attributes []*commonpb.KeyValue) map[string]string {
	result := map[string]string{}
	for _, attribute := range attributes {
		result[attribute.GetKey()] = anyValueString(attribute.GetValue())
	}
	return result
}

func anyValueString(value *commonpb.AnyValue) string {
	switch typed := value.GetValue().(type) {
	case *commonpb.AnyValue_StringValue:
		return typed.StringValue
	case *commonpb.AnyValue_IntValue:
		return strconv.FormatInt(typed.IntValue, 10)
	case *commonpb.AnyValue_DoubleValue:
		return strconv.FormatFloat(typed.DoubleValue, 'f', -1, 64)
	case *commonpb.AnyValue_BoolValue:
		return strconv.FormatBool(typed.BoolValue)
	default:
		return fmt.Sprint(value)
	}
}
