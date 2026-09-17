package observability

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestRequestMiddlewareRecordsSuccessfulRequest(t *testing.T) {
	test := newTestRecorder(t)
	router := newMiddlewareRouter(t, test.recorder, false)

	request := httptest.NewRequest(http.MethodGet, "/api/events/ABCD1234?editToken=anonymous-edit-token", nil)
	request.Header.Set("Authorization", "Bearer supersecrettoken")
	request.Header.Set("Cookie", "timeful_owner_ABCD1234=anonymous-edit-token")
	result := performRecordedRequest(t, test, router, request)

	if result.response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.response.Code)
	}
	requestID := result.response.Header().Get(RequestIDHeader)
	if requestID == "" {
		t.Fatal("response is missing the request correlation header")
	}
	if len(result.records) != 1 {
		t.Fatalf("records = %d, want 1", len(result.records))
	}

	attributes := recordAttributes(result.records[0])
	want := map[string]string{
		"request.id":                requestID,
		"http.request.method":       "GET",
		"http.route":                "/api/events/:eventId",
		"http.response.status_code": "200",
		"http.response.outcome":     "success",
	}
	for key, value := range want {
		if attributes[key] != value {
			t.Fatalf("attribute %s = %q, want %q", key, attributes[key], value)
		}
	}
	for _, leaked := range []string{"anonymous-edit-token", "supersecrettoken", "timeful_owner_"} {
		for key, value := range attributes {
			if strings.Contains(value, leaked) {
				t.Fatalf("attribute %s leaked %q: %q", key, leaked, value)
			}
		}
	}

	if len(result.spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(result.spans))
	}
	span := result.spans[0]
	if got := span.Name(); got != "GET /api/events/:eventId" {
		t.Fatalf("span name = %q, want GET /api/events/:eventId", got)
	}
	if span.SpanKind() != trace.SpanKindServer {
		t.Fatalf("span kind = %v, want server", span.SpanKind())
	}
	if result.records[0].TraceID() != span.SpanContext().TraceID() {
		t.Fatalf("log trace id = %s, want span trace id %s", result.records[0].TraceID(), span.SpanContext().TraceID())
	}
	if result.records[0].SpanID() != span.SpanContext().SpanID() {
		t.Fatalf("log span id = %s, want span id %s", result.records[0].SpanID(), span.SpanContext().SpanID())
	}
	for key, value := range want {
		if keyValueAttributes(span.Attributes())[key] != value {
			t.Fatalf("span attribute %s = %q, want %q", key, keyValueAttributes(span.Attributes())[key], value)
		}
	}
	for _, leaked := range []string{"anonymous-edit-token", "supersecrettoken", "timeful_owner_"} {
		for key, value := range keyValueAttributes(span.Attributes()) {
			if strings.Contains(value, leaked) {
				t.Fatalf("span attribute %s leaked %q: %q", key, leaked, value)
			}
		}
	}
}

func TestRequestMiddlewareRecordsErrorCode(t *testing.T) {
	test := newTestRecorder(t)
	router := newMiddlewareRouter(t, test.recorder, false)

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	if result.response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", result.response.Code)
	}

	attributes := recordAttributes(result.records[0])
	if got := attributes["error.type"]; got != "event-not-found" {
		t.Fatalf("error.type = %q, want event-not-found", got)
	}
	if got, exists := attributes["error.message"]; exists {
		t.Fatalf("error.message = %q, want absent for a code-like error", got)
	}
	if got := attributes["http.response.outcome"]; got != "client_error" {
		t.Fatalf("outcome = %q, want client_error", got)
	}
	if status := result.spans[0].Status(); status.Code != codes.Unset {
		t.Fatalf("span status = %v, want unset for a 4xx response", status.Code)
	}

	histogram := findMetric(t, result.metrics, requestDurationMetric)
	data, ok := histogram.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("request duration data = %T, want histogram", histogram.Data)
	}
	for _, point := range data.DataPoints {
		if got, exists := keyValueAttributes(point.Attributes.ToSlice())["error.type"]; exists {
			t.Fatalf("metric error.type = %q, want error context on spans and log records only", got)
		}
	}
}

func TestRequestMiddlewareRecordsRedactedErrorMessage(t *testing.T) {
	test := newTestRecorder(t)
	router := newMiddlewareRouter(t, test.recorder, false)

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/failing?token=anonymous-edit-token", nil))
	if result.response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", result.response.Code)
	}

	attributes := recordAttributes(result.records[0])
	if got := attributes["error.type"]; got != "http_5xx" {
		t.Fatalf("error.type = %q, want http_5xx", got)
	}
	if got := attributes["error.message"]; got != "failed to load account integration data" {
		t.Fatalf("error.message = %q", got)
	}
	if status := result.spans[0].Status(); status.Code != codes.Error {
		t.Fatalf("span status = %v, want error for a 5xx response", status.Code)
	}
}

func TestRequestMiddlewareRecordsReadiness(t *testing.T) {
	test := newTestRecorder(t)
	test.recorder.SetReadiness(func() ReadinessState { return ReadinessUnavailable })
	router := newMiddlewareRouter(t, test.recorder, false)

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/failing", nil))

	attributes := recordAttributes(result.records[0])
	if got := attributes["service.readiness"]; got != "unavailable" {
		t.Fatalf("service.readiness = %q, want unavailable", got)
	}
	if got := keyValueAttributes(result.spans[0].Attributes())["service.readiness"]; got != "unavailable" {
		t.Fatalf("span service.readiness = %q, want unavailable", got)
	}
}

func TestRequestMiddlewareRecordsRecoveredPanic(t *testing.T) {
	test := newTestRecorder(t)
	router := newMiddlewareRouter(t, test.recorder, true)

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/panicking", nil))
	if result.response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", result.response.Code)
	}
	if len(result.records) != 1 {
		t.Fatalf("records = %d, want 1 for a recovered panic", len(result.records))
	}
	attributes := recordAttributes(result.records[0])
	if got := attributes["http.response.status_code"]; got != "500" {
		t.Fatalf("status attribute = %q, want 500", got)
	}
	if len(result.spans) != 1 {
		t.Fatalf("spans = %d, want 1 for a recovered panic", len(result.spans))
	}
	if status := result.spans[0].Status(); status.Code != codes.Error {
		t.Fatalf("span status = %v, want error for a recovered panic", status.Code)
	}
}

func TestRequestMiddlewareRecordsUnmatchedRoute(t *testing.T) {
	test := newTestRecorder(t)
	router := newMiddlewareRouter(t, test.recorder, false)

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/unknown/path?editToken=anonymous-edit-token", nil))

	attributes := recordAttributes(result.records[0])
	if got := attributes["http.route"]; got != "unmatched" {
		t.Fatalf("route = %q, want unmatched", got)
	}
	if got := attributes["error.type"]; got != "route-not-found" {
		t.Fatalf("error.type = %q, want route-not-found", got)
	}
	if got := result.spans[0].Name(); got != "GET unmatched" {
		t.Fatalf("span name = %q, want GET unmatched", got)
	}
}

func TestRequestMiddlewareRecordsDurationMetric(t *testing.T) {
	test := newTestRecorder(t)
	router := newMiddlewareRouter(t, test.recorder, false)

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/events/ABCD1234", nil))
	if len(result.records) != 1 {
		t.Fatalf("records = %d, want 1", len(result.records))
	}

	histogram := findMetric(t, result.metrics, requestDurationMetric)
	data, ok := histogram.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("request duration data = %T, want histogram", histogram.Data)
	}
	if len(data.DataPoints) != 1 {
		t.Fatalf("histogram data points = %d, want 1", len(data.DataPoints))
	}
	point := data.DataPoints[0]
	if point.Count != 1 {
		t.Fatalf("histogram count = %d, want 1", point.Count)
	}
	attributes := keyValueAttributes(point.Attributes.ToSlice())
	want := map[string]string{
		"http.request.method":       "GET",
		"http.route":                "/api/events/:eventId",
		"http.response.status_code": "200",
		"http.response.outcome":     "success",
		"service.readiness":         "unknown",
	}
	for key, value := range want {
		if attributes[key] != value {
			t.Fatalf("metric attribute %s = %q, want %q", key, attributes[key], value)
		}
	}
	if _, exists := attributes["request.id"]; exists {
		t.Fatal("metric attributes must not carry the correlation identifier")
	}
	if got := len(point.Bounds); got != len(requestDurationBoundaries) {
		t.Fatalf("histogram bounds = %d, want %d", got, len(requestDurationBoundaries))
	}
}

func TestRequestMiddlewarePropagatesSpanContext(t *testing.T) {
	test := newTestRecorder(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var handlerTraceID trace.TraceID
	router.Use(RequestMiddleware(test.recorder))
	router.GET("/api/context", func(c *gin.Context) {
		handlerTraceID = trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID()
		c.Status(http.StatusOK)
	})

	result := performRecordedRequest(t, test, router, httptest.NewRequest(http.MethodGet, "/api/context", nil))
	if !handlerTraceID.IsValid() {
		t.Fatal("handler request context carries no span")
	}
	if handlerTraceID != result.records[0].TraceID() {
		t.Fatalf("handler span trace id = %s, want recorded trace id %s", handlerTraceID, result.records[0].TraceID())
	}
}

func TestNewRequestIDIsUUIDShaped(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	first := newRequestID()
	second := newRequestID()
	if !pattern.MatchString(first) {
		t.Fatalf("newRequestID() = %q, want a version-4 UUID", first)
	}
	if first == second {
		t.Fatal("newRequestID() returned the same identifier twice")
	}
}

func newMiddlewareRouter(t *testing.T, recorder *Recorder, withRecovery bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestMiddleware(recorder))
	if withRecovery {
		router.Use(gin.RecoveryWithWriter(io.Discard))
	}
	router.GET("/api/events/:eventId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/missing", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "event-not-found"})
	})
	router.GET("/api/failing", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load account integration data"})
	})
	router.GET("/api/panicking", func(c *gin.Context) {
		panic("boom")
	})
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route-not-found"})
	})
	return router
}

func performRecordedRequest(t *testing.T, test *testRecorder, router http.Handler, request *http.Request) *recordedRequest {
	t.Helper()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	var metrics metricdata.ResourceMetrics
	if err := test.metrics.Collect(context.Background(), &metrics); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := test.recorder.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	return &recordedRequest{
		response: response,
		records:  test.logs.snapshot(),
		spans:    test.spans.snapshot(),
		metrics:  metrics,
	}
}

type recordedRequest struct {
	response *httptest.ResponseRecorder
	records  []sdklog.Record
	spans    []sdktrace.ReadOnlySpan
	metrics  metricdata.ResourceMetrics
}
