package observability

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Recorder emits structured diagnostics, metrics, and traces through bounded,
// asynchronous OTLP pipelines. A nil recorder is a no-op, so instrumentation
// stays safe when the OpenObserve contract is absent.
type Recorder struct {
	logProvider     *sdklog.LoggerProvider
	logger          log.Logger
	tracerProvider  *sdktrace.TracerProvider
	tracer          trace.Tracer
	meterProvider   *sdkmetric.MeterProvider
	requestDuration metric.Float64Histogram
	readinessGauge  metric.Int64Gauge
	readiness       atomic.Pointer[func() ReadinessState]
	poolStats       atomic.Pointer[func() PoolStats]
}

// SetReadiness attaches a readiness source that enriches request records and
// spans. It is a no-op on a nil recorder.
func (r *Recorder) SetReadiness(source func() ReadinessState) {
	if r == nil {
		return
	}
	if source == nil {
		r.readiness.Store(nil)
		return
	}
	r.readiness.Store(&source)
}

// SetPoolStats attaches a PostgreSQL connection-pool source for the connection
// metrics. It is a no-op on a nil recorder.
func (r *Recorder) SetPoolStats(source func() PoolStats) {
	if r == nil {
		return
	}
	if source == nil {
		r.poolStats.Store(nil)
		return
	}
	r.poolStats.Store(&source)
}

func (r *Recorder) readinessState() ReadinessState {
	if r == nil {
		return ReadinessUnknown
	}
	source := r.readiness.Load()
	if source == nil {
		return ReadinessUnknown
	}
	return (*source)()
}

// Enabled reports whether signals are shipped to a collector.
func (r *Recorder) Enabled() bool {
	return r != nil && r.logProvider != nil
}

// TracerProvider exposes the provider that owns request, dependency, and
// outbound spans, so database instrumentation joins the same traces. A nil or
// disabled recorder returns nil.
func (r *Recorder) TracerProvider() trace.TracerProvider {
	if r == nil || r.tracerProvider == nil {
		return nil
	}
	return r.tracerProvider
}

// StartRequestSpan opens the always-on server span for one request. A nil or
// disabled recorder returns the context unchanged with a no-op span.
func (r *Recorder) StartRequestSpan(ctx context.Context, method string) (context.Context, trace.Span) {
	if r == nil || r.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return r.tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindServer))
}

// RecordReadiness reports the latest readiness sample to the metric pipeline.
// A nil or disabled recorder and the unknown state are no-ops.
func (r *Recorder) RecordReadiness(state ReadinessState) {
	if r == nil || r.readinessGauge == nil || state == ReadinessUnknown {
		return
	}
	value := int64(0)
	if state == ReadinessReady {
		value = 1
	}
	r.readinessGauge.Record(context.Background(), value)
}

// RecordRequestMetrics records the request duration histogram. Recording is
// synchronous and in-memory; export happens on the metric reader's background
// pipeline.
func (r *Recorder) RecordRequestMetrics(ctx context.Context, completion RequestCompletion) {
	if r == nil || r.requestDuration == nil {
		return
	}
	r.requestDuration.Record(ctx, completion.Latency.Seconds(), metric.WithAttributes(completion.MetricAttributes()...))
}

// EmitRequest records one finished request. Emitting enqueues to the batch
// processor and returns without waiting for the collector.
func (r *Recorder) EmitRequest(ctx context.Context, completion RequestCompletion) {
	if r == nil || r.logger == nil {
		return
	}
	if completion.Readiness == "" {
		completion.Readiness = r.readinessState()
	}
	r.logger.Emit(ctx, completion.LogRecord())
}

// Shutdown flushes queued telemetry and stops every export pipeline within the
// context budget.
func (r *Recorder) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	type namedShutdown struct {
		name     string
		shutdown func(context.Context) error
	}
	providers := make([]namedShutdown, 0, 3)
	if r.meterProvider != nil {
		providers = append(providers, namedShutdown{"metrics", r.meterProvider.Shutdown})
	}
	if r.tracerProvider != nil {
		providers = append(providers, namedShutdown{"traces", r.tracerProvider.Shutdown})
	}
	if r.logProvider != nil {
		providers = append(providers, namedShutdown{"logs", r.logProvider.Shutdown})
	}

	var (
		waitGroup sync.WaitGroup
		mu        sync.Mutex
		failures  []error
	)
	for _, provider := range providers {
		waitGroup.Add(1)
		go func(provider namedShutdown) {
			defer waitGroup.Done()
			if err := provider.shutdown(ctx); err != nil {
				mu.Lock()
				failures = append(failures, fmt.Errorf("%s: %w", provider.name, err))
				mu.Unlock()
			}
		}(provider)
	}
	waitGroup.Wait()
	return errors.Join(failures...)
}
