package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// InstrumentDefaultTransport wraps http.DefaultTransport so every outbound
// request made through http.DefaultClient, http.Get, and http.PostForm is
// traced. A nil or disabled recorder leaves the transport untouched.
func InstrumentDefaultTransport(recorder *Recorder) {
	if recorder == nil || recorder.tracer == nil {
		return
	}
	http.DefaultTransport = recorder.Transport(http.DefaultTransport)
}

// Transport returns a RoundTripper that records one bounded client span per
// outbound request. The span carries the method, scheme, host, port, status
// code, and a coarse error class; it never carries the URL path or query,
// headers, bodies, or error text, so token-bearing URLs and credentials cannot
// reach the exported telemetry (QR-004). Any 4xx or 5xx response marks the span
// status as Error, matching otelhttp's client-span behavior.
func (r *Recorder) Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if r == nil || r.tracer == nil {
		return base
	}
	return &tracingTransport{
		base:       base,
		tracer:     r.tracer,
		propagator: propagation.TraceContext{},
	}
}

type tracingTransport struct {
	base       http.RoundTripper
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func (t *tracingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.Start(request.Context(), request.Method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(outboundRequestAttributes(request)...),
	)
	defer span.End()

	if request.Header == nil {
		request.Header = http.Header{}
	}
	t.propagator.Inject(ctx, propagation.HeaderCarrier(request.Header))

	response, err := t.base.RoundTrip(request)
	if err != nil {
		span.SetAttributes(attribute.String("error.type", classifyTransportError(err)))
		span.SetStatus(codes.Error, "")
		return response, err
	}
	span.SetAttributes(attribute.Int("http.response.status_code", response.StatusCode))
	if response.StatusCode >= 400 {
		span.SetStatus(codes.Error, "")
	}
	return response, nil
}

func outboundRequestAttributes(request *http.Request) []attribute.KeyValue {
	attributes := []attribute.KeyValue{
		attribute.String("http.request.method", request.Method),
	}
	if request.URL == nil {
		return attributes
	}
	if request.URL.Scheme != "" {
		attributes = append(attributes, attribute.String("url.scheme", request.URL.Scheme))
	}
	if host := request.URL.Hostname(); host != "" {
		attributes = append(attributes, attribute.String("server.address", host))
	}
	if port := request.URL.Port(); port != "" {
		if parsed, err := strconv.Atoi(port); err == nil {
			attributes = append(attributes, attribute.Int("server.port", parsed))
		}
	}
	return attributes
}

// classifyTransportError maps a transport failure to a coarse class. The error
// text is never used because it can embed the full request URL.
func classifyTransportError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	var netError net.Error
	if errors.As(err, &netError) && netError.Timeout() {
		return "timeout"
	}
	return "transport_error"
}
