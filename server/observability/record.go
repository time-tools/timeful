package observability

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

const maxErrorMessageLength = 512

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

// Outcome classifies the response for operators.
func (c RequestCompletion) Outcome() string {
	switch {
	case c.Status >= 500:
		return "server_error"
	case c.Status >= 400:
		return "client_error"
	default:
		return "success"
	}
}

// Attributes returns the bounded, redacted attributes shared by the log record
// and the request span. Error context is redacted here so every shipped signal
// passes through the QR-004 boundary, regardless of the caller.
func (c RequestCompletion) Attributes() []attribute.KeyValue {
	readiness := c.Readiness
	if readiness == "" {
		readiness = ReadinessUnknown
	}
	attributes := []attribute.KeyValue{
		attribute.String("request.id", c.RequestID),
		attribute.String("http.request.method", c.Method),
		attribute.String("http.route", c.Route),
		attribute.Int("http.response.status_code", c.Status),
		attribute.String("http.response.outcome", c.Outcome()),
		attribute.Float64("http.response.duration_ms", float64(c.Latency.Nanoseconds())/1e6),
		attribute.String("service.readiness", string(readiness)),
	}
	if c.ErrorType != "" {
		attributes = append(attributes, attribute.String("error.type", Redact(c.ErrorType)))
	}
	if c.ErrorMessage != "" {
		attributes = append(attributes, attribute.String("error.message", Redact(truncate(c.ErrorMessage, maxErrorMessageLength))))
	}
	return attributes
}

// MetricAttributes returns the bounded labels for the request metrics. The
// correlation identifier is intentionally excluded to keep metric cardinality
// low, and error context stays on spans and log records because handler-authored
// error types do not guarantee a bounded label set.
func (c RequestCompletion) MetricAttributes() []attribute.KeyValue {
	readiness := c.Readiness
	if readiness == "" {
		readiness = ReadinessUnknown
	}
	return []attribute.KeyValue{
		attribute.String("http.request.method", c.Method),
		attribute.String("http.route", c.Route),
		attribute.Int("http.response.status_code", c.Status),
		attribute.String("http.response.outcome", c.Outcome()),
		attribute.String("service.readiness", string(readiness)),
	}
}

// ApplyToSpan sets the request span's final name, attributes, and status.
func (c RequestCompletion) ApplyToSpan(span trace.Span) {
	if span == nil {
		return
	}
	if c.Route != "" {
		span.SetName(c.Method + " " + c.Route)
	}
	span.SetAttributes(c.Attributes()...)
	if c.Status >= 500 {
		span.SetStatus(codes.Error, "")
	}
}

// LogRecord builds the OTLP record for the completion.
func (c RequestCompletion) LogRecord() log.Record {
	var record log.Record
	record.SetEventName("http.request.completed")
	timestamp := c.StartTime
	if !timestamp.IsZero() {
		timestamp = timestamp.Add(c.Latency)
	}
	record.SetTimestamp(timestamp)
	record.SetObservedTimestamp(time.Now())
	record.SetSeverity(c.severity())
	record.SetSeverityText(c.severityText())
	record.SetBody(attribute.StringValue("request completed"))
	record.AddAttributes(c.Attributes()...)
	return record
}

func (c RequestCompletion) severity() log.Severity {
	switch {
	case c.Status >= 500:
		return log.SeverityError
	case c.Status >= 400:
		return log.SeverityWarn
	default:
		return log.SeverityInfo
	}
}

func (c RequestCompletion) severityText() string {
	switch {
	case c.Status >= 500:
		return "ERROR"
	case c.Status >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}
