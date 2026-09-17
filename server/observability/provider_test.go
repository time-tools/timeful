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
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type receivedExport struct {
	path       string
	auth       string
	streamName string
	request    *collogpb.ExportLogsServiceRequest
}

type receivedTraceExport struct {
	path       string
	auth       string
	streamName string
	request    *coltracepb.ExportTraceServiceRequest
}

type receivedMetricExport struct {
	path       string
	auth       string
	streamName string
	request    *colmetricpb.ExportMetricsServiceRequest
}

type signalCollector struct {
	logs    chan receivedExport
	traces  chan receivedTraceExport
	metrics chan receivedMetricExport
}

func newSignalCollector(t *testing.T) (*signalCollector, *httptest.Server) {
	t.Helper()
	collector := &signalCollector{
		logs:    make(chan receivedExport, 8),
		traces:  make(chan receivedTraceExport, 8),
		metrics: make(chan receivedMetricExport, 8),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read export body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/logs"):
			request := &collogpb.ExportLogsServiceRequest{}
			if err := proto.Unmarshal(body, request); err != nil {
				t.Errorf("decode logs export: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			collector.logs <- receivedExport{
				path:       r.URL.Path,
				auth:       r.Header.Get("Authorization"),
				streamName: r.Header.Get("stream-name"),
				request:    request,
			}
		case strings.HasSuffix(r.URL.Path, "/v1/traces"):
			request := &coltracepb.ExportTraceServiceRequest{}
			if err := proto.Unmarshal(body, request); err != nil {
				t.Errorf("decode traces export: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			collector.traces <- receivedTraceExport{
				path:       r.URL.Path,
				auth:       r.Header.Get("Authorization"),
				streamName: r.Header.Get("stream-name"),
				request:    request,
			}
		case strings.HasSuffix(r.URL.Path, "/v1/metrics"):
			request := &colmetricpb.ExportMetricsServiceRequest{}
			if err := proto.Unmarshal(body, request); err != nil {
				t.Errorf("decode metrics export: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			collector.metrics <- receivedMetricExport{
				path:       r.URL.Path,
				auth:       r.Header.Get("Authorization"),
				streamName: r.Header.Get("stream-name"),
				request:    request,
			}
		default:
			t.Errorf("unexpected export path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	return collector, server
}

func TestStartShipsRecordsToConfiguredOrganization(t *testing.T) {
	collector, server := newSignalCollector(t)
	defer server.Close()

	config := Config{
		Endpoint:       server.URL,
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
	case export := <-collector.logs:
		assertExportRequest(t, export, config)
	case <-time.After(2 * time.Second):
		t.Fatal("no OTLP logs export reached the collector")
	}
}

func TestStartShipsMetricsAndTracesToConfiguredOrganization(t *testing.T) {
	collector, server := newSignalCollector(t)
	defer server.Close()

	config := Config{
		Endpoint:       server.URL,
		OrganizationID: "org-abc",
		Username:       "ingest@timeful.fun",
		Password:       "ingest-token",
		Environment:    "test",
	}
	recorder, err := Start(context.Background(), config)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	completion := RequestCompletion{
		RequestID:    "request-1",
		Method:       "GET",
		Route:        "/api/events/:eventId",
		Status:       500,
		Latency:      25 * time.Millisecond,
		ErrorType:    "http_5xx",
		ErrorMessage: "Authorization: Bearer supersecrettoken",
		Readiness:    ReadinessUnavailable,
		StartTime:    time.Now().Add(-25 * time.Millisecond),
	}
	_, span := recorder.StartRequestSpan(context.Background(), completion.Method)
	completion.ApplyToSpan(span)
	span.End()
	recorder.RecordRequestMetrics(context.Background(), completion)
	recorder.RecordReadiness(ReadinessUnavailable)
	recorder.SetPoolStats(func() PoolStats { return PoolStats{Used: 3, Idle: 2, Max: 10} })

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := recorder.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("ingest@timeful.fun:ingest-token"))

	select {
	case export := <-collector.traces:
		if want := "/api/org-abc/v1/traces"; export.path != want {
			t.Fatalf("traces path = %q, want %q", export.path, want)
		}
		if export.auth != authHeader {
			t.Fatalf("traces Authorization = %q, want %q", export.auth, authHeader)
		}
		if export.streamName != ServerTraceStream {
			t.Fatalf("traces stream-name = %q, want %q", export.streamName, ServerTraceStream)
		}
		assertTraceExport(t, export.request, completion)
	case <-time.After(2 * time.Second):
		t.Fatal("no OTLP traces export reached the collector")
	}

	select {
	case export := <-collector.metrics:
		if want := "/api/org-abc/v1/metrics"; export.path != want {
			t.Fatalf("metrics path = %q, want %q", export.path, want)
		}
		if export.auth != authHeader {
			t.Fatalf("metrics Authorization = %q, want %q", export.auth, authHeader)
		}
		if export.streamName != "" {
			t.Fatalf("metrics stream-name = %q, want empty: metrics use the default stream naming", export.streamName)
		}
		assertMetricExport(t, export.request)
	case <-time.After(2 * time.Second):
		t.Fatal("no OTLP metrics export reached the collector")
	}
}

func assertTraceExport(t *testing.T, request *coltracepb.ExportTraceServiceRequest, completion RequestCompletion) {
	t.Helper()
	if len(request.ResourceSpans) != 1 {
		t.Fatalf("ResourceSpans length = %d, want 1", len(request.ResourceSpans))
	}
	resourceSpans := request.ResourceSpans[0]
	resource := anyValuesToStrings(resourceSpans.GetResource().GetAttributes())
	if resource["service.name"] != "timeful-server" {
		t.Fatalf("service.name = %q", resource["service.name"])
	}
	if resource["deployment.environment"] != "test" {
		t.Fatalf("deployment.environment = %q", resource["deployment.environment"])
	}
	if len(resourceSpans.ScopeSpans) != 1 || len(resourceSpans.ScopeSpans[0].Spans) != 1 {
		t.Fatalf("scope spans = %#v, want one span", resourceSpans.ScopeSpans)
	}
	span := resourceSpans.ScopeSpans[0].Spans[0]
	if want := "GET /api/events/:eventId"; span.GetName() != want {
		t.Fatalf("span name = %q, want %q", span.GetName(), want)
	}
	if span.GetKind() != tracepb.Span_SPAN_KIND_SERVER {
		t.Fatalf("span kind = %v, want server", span.GetKind())
	}
	if span.GetStatus().GetCode() != tracepb.Status_STATUS_CODE_ERROR {
		t.Fatalf("span status = %v, want error", span.GetStatus().GetCode())
	}
	attributes := anyValuesToStrings(span.GetAttributes())
	want := map[string]string{
		"request.id":                completion.RequestID,
		"http.request.method":       completion.Method,
		"http.route":                completion.Route,
		"http.response.status_code": "500",
		"http.response.outcome":     "server_error",
		"error.type":                "http_5xx",
		"service.readiness":         "unavailable",
	}
	for key, value := range want {
		if attributes[key] != value {
			t.Fatalf("span attribute %s = %q, want %q", key, attributes[key], value)
		}
	}
	if strings.Contains(prototext.Format(request), "supersecrettoken") {
		t.Fatal("traces payload leaked the synthetic bearer token")
	}
}

func assertMetricExport(t *testing.T, request *colmetricpb.ExportMetricsServiceRequest) {
	t.Helper()
	if len(request.ResourceMetrics) == 0 {
		t.Fatal("metrics export carries no resource metrics")
	}
	resource := anyValuesToStrings(request.ResourceMetrics[0].GetResource().GetAttributes())
	if resource["service.name"] != "timeful-server" {
		t.Fatalf("metrics service.name = %q", resource["service.name"])
	}

	if metric := findProtoMetric(request, readinessMetric); metric != nil {
		points := metric.GetGauge().GetDataPoints()
		if len(points) != 1 {
			t.Fatalf("readiness data points = %d, want 1", len(points))
		}
		if got := points[0].GetAsInt(); got != 0 {
			t.Fatalf("readiness value = %d, want 0 for unavailable", got)
		}
	} else {
		t.Fatalf("metric %q not exported", readinessMetric)
	}

	connections := findProtoMetric(request, poolConnectionsMetric)
	if connections == nil {
		t.Fatalf("metric %q not exported", poolConnectionsMetric)
	}
	if sum := connections.GetSum(); sum == nil || sum.GetIsMonotonic() {
		t.Fatalf("pool connections metric must be exported as a non-monotonic sum, got %#v", connections.GetSum())
	}
	states := map[string]int64{}
	for _, point := range connections.GetSum().GetDataPoints() {
		attributes := anyValuesToStrings(point.GetAttributes())
		states[attributes["db.client.connection.state"]] = point.GetAsInt()
	}
	if states["used"] != 3 || states["idle"] != 2 {
		t.Fatalf("pool connection states = %#v, want used=3 idle=2", states)
	}

	maximum := findProtoMetric(request, poolMaxConnectionsMetric)
	if maximum == nil {
		t.Fatalf("metric %q not exported", poolMaxConnectionsMetric)
	}
	if sum := maximum.GetSum(); sum == nil || sum.GetIsMonotonic() {
		t.Fatalf("pool maximum metric must be exported as a non-monotonic sum, got %#v", maximum.GetSum())
	}
	if got := maximum.GetSum().GetDataPoints()[0].GetAsInt(); got != 10 {
		t.Fatalf("pool maximum = %d, want 10", got)
	}

	duration := findProtoMetric(request, requestDurationMetric)
	if duration == nil {
		t.Fatalf("metric %q not exported", requestDurationMetric)
	}
	if got := duration.GetHistogram().GetDataPoints()[0].GetCount(); got != 1 {
		t.Fatalf("request duration count = %d, want 1", got)
	}
}

func findProtoMetric(request *colmetricpb.ExportMetricsServiceRequest, name string) *metricspb.Metric {
	for _, resourceMetrics := range request.ResourceMetrics {
		for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
			for _, metric := range scopeMetrics.Metrics {
				if metric.Name == name {
					return metric
				}
			}
		}
	}
	return nil
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
	if recorder.TracerProvider() != nil {
		t.Fatal("TracerProvider() = non-nil for a disabled recorder")
	}

	recorder.SetReadiness(func() ReadinessState { return ReadinessReady })
	recorder.SetPoolStats(func() PoolStats { return PoolStats{} })
	recorder.RecordReadiness(ReadinessReady)
	recorder.RecordRequestMetrics(context.Background(), RequestCompletion{RequestID: "request-1"})
	recorder.EmitRequest(context.Background(), RequestCompletion{RequestID: "request-1"})
	spanContext, span := recorder.StartRequestSpan(context.Background(), http.MethodGet)
	span.End()
	if spanContext == nil {
		t.Fatal("StartRequestSpan() returned a nil context")
	}
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
