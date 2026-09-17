package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const postgresInstrumentationName = "timeful/server/postgres"

// queryTracer records one client span per PostgreSQL query. Spans carry the
// operation, database, host, and SQLSTATE only; SQL text, query parameters,
// and error messages never reach the exported telemetry (QR-004).
type queryTracer struct {
	tracer trace.Tracer
}

func newQueryTracer(provider trace.TracerProvider) queryTracer {
	if provider == nil {
		provider = noop.NewTracerProvider()
	}
	return queryTracer{tracer: provider.Tracer(postgresInstrumentationName)}
}

func (t queryTracer) TraceQueryStart(ctx context.Context, connection *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	operation := queryOperation(data.SQL)
	attributes := []attribute.KeyValue{
		attribute.String("db.system.name", "postgresql"),
		attribute.String("db.operation.name", operation),
	}
	if connection != nil {
		config := connection.Config()
		if config.Database != "" {
			attributes = append(attributes, attribute.String("db.namespace", config.Database))
		}
		if config.Host != "" {
			attributes = append(attributes, attribute.String("server.address", config.Host))
		}
	}
	ctx, _ = t.tracer.Start(ctx, operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attributes...),
	)
	return ctx
}

func (t queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	if data.Err != nil {
		var postgresError *pgconn.PgError
		if errors.As(data.Err, &postgresError) && postgresError.Code != "" {
			span.SetAttributes(attribute.String("db.response.status_code", postgresError.Code))
		}
		span.SetStatus(codes.Error, "")
	}
	span.End()
}

func queryOperation(sql string) string {
	fields := strings.Fields(sql)
	if len(fields) == 0 {
		return "unknown"
	}
	return strings.ToUpper(fields[0])
}

// ConnectionStats reports the current PostgreSQL pool usage for the pool
// metrics. It returns zeros before the pool is initialized.
func ConnectionStats() (used int, idle int, max int) {
	if Pool == nil {
		return 0, 0, 0
	}
	stats := Pool.Stat()
	return int(stats.AcquiredConns()), int(stats.IdleConns()), int(stats.MaxConns())
}
