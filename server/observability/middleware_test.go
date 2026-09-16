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
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func TestRequestMiddlewareRecordsSuccessfulRequest(t *testing.T) {
	exporter := &capturingExporter{}
	recorder := newRecorder(exporter, "test")
	router := newMiddlewareRouter(t, recorder, false)

	request := httptest.NewRequest(http.MethodGet, "/api/events/ABCD1234?editToken=anonymous-edit-token", nil)
	request.Header.Set("Authorization", "Bearer supersecrettoken")
	request.Header.Set("Cookie", "timeful_owner_ABCD1234=anonymous-edit-token")
	result := performRecordedRequest(t, recorder, exporter, router, request)

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
}

func TestRequestMiddlewareRecordsErrorCode(t *testing.T) {
	exporter := &capturingExporter{}
	recorder := newRecorder(exporter, "test")
	router := newMiddlewareRouter(t, recorder, false)

	result := performRecordedRequest(t, recorder, exporter, router, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
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
}

func TestRequestMiddlewareRecordsRedactedErrorMessage(t *testing.T) {
	exporter := &capturingExporter{}
	recorder := newRecorder(exporter, "test")
	router := newMiddlewareRouter(t, recorder, false)

	result := performRecordedRequest(t, recorder, exporter, router, httptest.NewRequest(http.MethodGet, "/api/failing?token=anonymous-edit-token", nil))
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
}

func TestRequestMiddlewareRecordsReadiness(t *testing.T) {
	exporter := &capturingExporter{}
	recorder := newRecorder(exporter, "test")
	recorder.SetReadiness(func() ReadinessState { return ReadinessUnavailable })
	router := newMiddlewareRouter(t, recorder, false)

	result := performRecordedRequest(t, recorder, exporter, router, httptest.NewRequest(http.MethodGet, "/api/failing", nil))

	attributes := recordAttributes(result.records[0])
	if got := attributes["service.readiness"]; got != "unavailable" {
		t.Fatalf("service.readiness = %q, want unavailable", got)
	}
}

func TestRequestMiddlewareRecordsRecoveredPanic(t *testing.T) {
	exporter := &capturingExporter{}
	recorder := newRecorder(exporter, "test")
	router := newMiddlewareRouter(t, recorder, true)

	result := performRecordedRequest(t, recorder, exporter, router, httptest.NewRequest(http.MethodGet, "/api/panicking", nil))
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
}

func TestRequestMiddlewareRecordsUnmatchedRoute(t *testing.T) {
	exporter := &capturingExporter{}
	recorder := newRecorder(exporter, "test")
	router := newMiddlewareRouter(t, recorder, false)

	result := performRecordedRequest(t, recorder, exporter, router, httptest.NewRequest(http.MethodGet, "/api/unknown/path?editToken=anonymous-edit-token", nil))

	attributes := recordAttributes(result.records[0])
	if got := attributes["http.route"]; got != "unmatched" {
		t.Fatalf("route = %q, want unmatched", got)
	}
	if got := attributes["error.type"]; got != "route-not-found" {
		t.Fatalf("error.type = %q, want route-not-found", got)
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

func performRecordedRequest(t *testing.T, recorder *Recorder, exporter *capturingExporter, router http.Handler, request *http.Request) *recordedRequest {
	t.Helper()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := recorder.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	return &recordedRequest{response: response, records: exporter.snapshot()}
}

type recordedRequest struct {
	response *httptest.ResponseRecorder
	records  []sdklog.Record
}
