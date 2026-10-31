package observability

import (
	"time"
)

// RequestCompletion is the canonical internal shape of one finished request.
// It carries only non-sensitive context: the request correlation identifier,
// the matched route template, the outcome, latency, redacted error context,
// and the last observed readiness state.
type RequestCompletion struct {
	RequestID    string
	Method       string
	Route        string
	Status       int
	Latency      time.Duration
	ErrorType    string
	ErrorMessage string
	Readiness    ReadinessState
	StartTime    time.Time
}
