package postgres

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type recordingSpanExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (e *recordingSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, spans...)
	return nil
}

func (e *recordingSpanExporter) Shutdown(context.Context) error { return nil }

func (e *recordingSpanExporter) snapshot() []sdktrace.ReadOnlySpan {
	e.mu.Lock()
	defer e.mu.Unlock()
	spans := make([]sdktrace.ReadOnlySpan, len(e.spans))
	copy(spans, e.spans)
	return spans
}

func newTestQueryTracer(t *testing.T) (queryTracer, *recordingSpanExporter) {
	t.Helper()
	exporter := &recordingSpanExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})
	return queryTracer{tracer: provider.Tracer(postgresInstrumentationName)}, exporter
}

func TestQueryOperation(t *testing.T) {
	tests := []struct {
		sql  string
		want string
	}{
		{sql: "SELECT * FROM events WHERE id = $1", want: "SELECT"},
		{sql: "  insert into events values ($1)", want: "INSERT"},
		{sql: "UPDATE events SET name = $1", want: "UPDATE"},
		{sql: "WITH recent AS (SELECT 1) SELECT * FROM recent", want: "WITH"},
		{sql: "", want: "unknown"},
		{sql: "   ", want: "unknown"},
	}
	for _, test := range tests {
		if got := queryOperation(test.sql); got != test.want {
			t.Fatalf("queryOperation(%q) = %q, want %q", test.sql, got, test.want)
		}
	}
}

func TestQueryTracerRecordsBoundedSpan(t *testing.T) {
	tracer, exporter := newTestQueryTracer(t)
	ctx := tracer.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{
		SQL: "SELECT * FROM events WHERE token = 'anonymous-edit-token'",
	})
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{})

	spans := exporter.snapshot()
	if len(spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(spans))
	}
	span := spans[0]
	if span.Name() != "SELECT" {
		t.Fatalf("span name = %q, want SELECT", span.Name())
	}
	if span.SpanKind() != trace.SpanKindClient {
		t.Fatalf("span kind = %v, want client", span.SpanKind())
	}
	attributes := spanAttributeMap(span)
	if attributes["db.system.name"] != "postgresql" {
		t.Fatalf("db.system.name = %q, want postgresql", attributes["db.system.name"])
	}
	if attributes["db.operation.name"] != "SELECT" {
		t.Fatalf("db.operation.name = %q, want SELECT", attributes["db.operation.name"])
	}
	for key, value := range attributes {
		if strings.Contains(value, "anonymous-edit-token") || strings.Contains(value, "FROM events") {
			t.Fatalf("span attribute %s leaked SQL or parameters: %q", key, value)
		}
	}
}

func TestQueryTracerRecordsSQLSTATEWithoutMessage(t *testing.T) {
	tracer, exporter := newTestQueryTracer(t)
	ctx := tracer.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{
		SQL: "INSERT INTO events VALUES ($1)",
	})
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		Err: &pgconn.PgError{Code: "23505", Message: "duplicate key value contains supersecrettoken"},
	})

	spans := exporter.snapshot()
	if len(spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(spans))
	}
	span := spans[0]
	attributes := spanAttributeMap(span)
	if attributes["db.response.status_code"] != "23505" {
		t.Fatalf("db.response.status_code = %q, want 23505", attributes["db.response.status_code"])
	}
	if span.Status().Code != codes.Error {
		t.Fatalf("span status = %v, want error", span.Status().Code)
	}
	for key, value := range attributes {
		if strings.Contains(value, "supersecrettoken") || strings.Contains(value, "duplicate key") {
			t.Fatalf("span attribute %s leaked the PostgreSQL error message: %q", key, value)
		}
	}
}

func TestNewQueryTracerWithoutProviderIsNoop(t *testing.T) {
	tracer := newQueryTracer(nil)
	ctx := tracer.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{SQL: "SELECT 1"})
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{})
}

func TestConnectionStatsWithoutPool(t *testing.T) {
	previous := Pool
	Pool = nil
	t.Cleanup(func() { Pool = previous })

	used, idle, maximum := ConnectionStats()
	if used != 0 || idle != 0 || maximum != 0 {
		t.Fatalf("ConnectionStats() = (%d, %d, %d), want zeros", used, idle, maximum)
	}
}

func spanAttributeMap(span sdktrace.ReadOnlySpan) map[string]string {
	attributes := map[string]string{}
	for _, kv := range span.Attributes() {
		attributes[string(kv.Key)] = kv.Value.Emit()
	}
	return attributes
}
