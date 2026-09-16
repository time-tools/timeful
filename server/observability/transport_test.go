package observability

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestTransportRecordsBoundedClientSpan(t *testing.T) {
	test := newTestRecorder(t)
	var traceparent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Transport: test.recorder.Transport(http.DefaultTransport)}
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/events/secret?token=anonymous-edit-token", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	request.Header.Set("Authorization", "Bearer supersecrettoken")
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	response.Body.Close()

	span := shutdownAndReadSpan(t, test)
	if got := span.Name(); got != "GET" {
		t.Fatalf("span name = %q, want GET", got)
	}
	if span.SpanKind() != trace.SpanKindClient {
		t.Fatalf("span kind = %v, want client", span.SpanKind())
	}
	attributes := keyValueAttributes(span.Attributes())
	if attributes["url.scheme"] != "http" {
		t.Fatalf("url.scheme = %q, want http", attributes["url.scheme"])
	}
	if attributes["server.address"] != "127.0.0.1" {
		t.Fatalf("server.address = %q, want 127.0.0.1", attributes["server.address"])
	}
	if attributes["http.response.status_code"] != "200" {
		t.Fatalf("status = %q, want 200", attributes["http.response.status_code"])
	}
	if span.Status().Code != codes.Unset {
		t.Fatalf("span status = %v, want unset for a 2xx response", span.Status().Code)
	}
	if _, exists := attributes["url.full"]; exists {
		t.Fatal("outbound span must not carry the full URL")
	}
	if _, exists := attributes["url.path"]; exists {
		t.Fatal("outbound span must not carry the URL path")
	}
	formatted := fmt.Sprint(attributes)
	for _, leaked := range []string{"anonymous-edit-token", "supersecrettoken", "/api/events/secret"} {
		if strings.Contains(formatted, leaked) {
			t.Fatalf("outbound span leaked %q: %s", leaked, formatted)
		}
	}
	if traceparent == "" {
		t.Fatal("traceparent was not injected into the outbound request")
	}
	if !strings.Contains(traceparent, span.SpanContext().TraceID().String()) {
		t.Fatalf("traceparent %q does not carry span trace id %s", traceparent, span.SpanContext().TraceID())
	}
}

func TestTransportSanitizesFailures(t *testing.T) {
	test := newTestRecorder(t)
	transport := test.recorder.Transport(roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("Get %q: dial tcp: connection refused", request.URL.String())
	}))
	client := &http.Client{Transport: transport}
	request, err := http.NewRequest(http.MethodPost, "http://secret.example/api?token=anonymous-edit-token", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if _, err := client.Do(request); err == nil {
		t.Fatal("Do() succeeded, want a transport failure")
	}

	span := shutdownAndReadSpan(t, test)
	attributes := keyValueAttributes(span.Attributes())
	if attributes["error.type"] != "transport_error" {
		t.Fatalf("error.type = %q, want transport_error", attributes["error.type"])
	}
	if span.Status().Code != codes.Error {
		t.Fatalf("span status = %v, want error", span.Status().Code)
	}
	formatted := fmt.Sprint(attributes)
	for _, leaked := range []string{"anonymous-edit-token", "/api", "connection refused"} {
		if strings.Contains(formatted, leaked) {
			t.Fatalf("outbound span leaked %q: %s", leaked, formatted)
		}
	}
}

func TestTransportMarksErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "client error", statusCode: http.StatusNotFound},
		{name: "server error", statusCode: http.StatusBadGateway},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			test := newTestRecorder(t)
			transport := test.recorder.Transport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.statusCode, Body: http.NoBody}, nil
			}))
			request, _ := http.NewRequest(http.MethodGet, "http://example.test/health", nil)
			response, err := transport.RoundTrip(request)
			if err != nil {
				t.Fatalf("RoundTrip() error = %v", err)
			}
			if response.StatusCode != tc.statusCode {
				t.Fatalf("status = %d, want %d", response.StatusCode, tc.statusCode)
			}

			span := shutdownAndReadSpan(t, test)
			attributes := keyValueAttributes(span.Attributes())
			if got := attributes["http.response.status_code"]; got != fmt.Sprint(tc.statusCode) {
				t.Fatalf("status attribute = %q, want %d", got, tc.statusCode)
			}
			if span.Status().Code != codes.Error {
				t.Fatalf("span status = %v, want error for a %d response", span.Status().Code, tc.statusCode)
			}
		})
	}
}

func TestInstrumentDefaultTransportWrapsDefaultClient(t *testing.T) {
	test := newTestRecorder(t)
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })

	InstrumentDefaultTransport(test.recorder)
	wrapped := http.DefaultTransport
	if wrapped == previous {
		t.Fatal("InstrumentDefaultTransport() did not wrap the default transport")
	}
	InstrumentDefaultTransport(nil)
	if http.DefaultTransport != wrapped {
		t.Fatal("InstrumentDefaultTransport() with a nil recorder replaced the transport")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	response, err := http.Get(server.URL + "/api/tx")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	response.Body.Close()

	span := shutdownAndReadSpan(t, test)
	if span.Name() != "GET" {
		t.Fatalf("span name = %q, want GET", span.Name())
	}
}

func TestClassifyTransportError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "deadline", err: context.DeadlineExceeded, want: "timeout"},
		{name: "canceled", err: context.Canceled, want: "canceled"},
		{name: "net timeout", err: timeoutError{}, want: "timeout"},
		{name: "other", err: errors.New("boom"), want: "transport_error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := classifyTransportError(test.err); got != test.want {
				t.Fatalf("classifyTransportError() = %q, want %q", got, test.want)
			}
		})
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

var _ net.Error = timeoutError{}

func shutdownAndReadSpan(t *testing.T, test *testRecorder) sdktrace.ReadOnlySpan {
	t.Helper()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := test.recorder.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	spans := test.spans.snapshot()
	if len(spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(spans))
	}
	return spans[0]
}
