package observability

import (
	"go.opentelemetry.io/otel/propagation"
	"net/http"
)

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
