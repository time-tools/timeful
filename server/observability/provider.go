// Package observability ships structured server diagnostic records, metrics,
// and traces to OpenObserve over OpenTelemetry OTLP. Export runs on background
// pipelines, so it never blocks or fails request handling, and the existing
// local diagnostic output stays authoritative while OpenObserve is
// unavailable.
package observability

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"timeful/server/logger"
)

const (
	serviceName         = "timeful-server"
	instrumentationName = "timeful/server/observability"

	exportInterval     = 5 * time.Second
	exportTimeout      = 5 * time.Second
	maxExportQueueSize = 2048
	exportMaxBatchSize = 512

	// DefaultShutdownTimeout bounds the final flush when the server stops.
	DefaultShutdownTimeout = 10 * time.Second
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

// Start builds the OTLP/HTTP exporters for the configured OpenObserve
// organization and returns a Recorder for logs, metrics, and traces.
// Incomplete configuration returns an error instead of failing startup, so the
// server keeps serving and keeps writing local diagnostics without OpenObserve.
func Start(ctx context.Context, config Config) (*Recorder, error) {
	if !config.Enabled() {
		return nil, errors.New("OpenObserve ingest configuration is incomplete")
	}
	logsURL, err := config.LogsURL()
	if err != nil {
		return nil, err
	}
	metricsURL, err := config.MetricsURL()
	if err != nil {
		return nil, err
	}
	tracesURL, err := config.TracesURL()
	if err != nil {
		return nil, err
	}
	authHeader := config.AuthHeader()

	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(logsURL),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": authHeader,
			"stream-name":   ServerLogStream,
		}),
		otlploghttp.WithTimeout(exportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("build OpenObserve OTLP logs exporter: %w", err)
	}
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(tracesURL),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": authHeader,
			"stream-name":   ServerTraceStream,
		}),
		otlptracehttp.WithTimeout(exportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("build OpenObserve OTLP traces exporter: %w", err)
	}
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(metricsURL),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": authHeader,
		}),
		otlpmetrichttp.WithTimeout(metricExportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("build OpenObserve OTLP metrics exporter: %w", err)
	}

	otel.SetErrorHandler(otel.ErrorHandlerFunc(ReportExportError))
	otel.SetTextMapPropagator(propagation.TraceContext{})

	metricReader := sdkmetric.NewPeriodicReader(metricExporter,
		sdkmetric.WithInterval(metricExportInterval),
		sdkmetric.WithTimeout(metricExportTimeout),
	)
	recorder, err := newRecorder(logExporter, traceExporter, metricReader, config.Environment)
	if err != nil {
		return nil, err
	}
	if err := startRuntimeMetrics(recorder.meterProvider); err != nil && logger.StdErr != nil {
		logger.StdErr.Printf("observability runtime metrics unavailable: %s", Redact(err.Error()))
	}
	return recorder, nil
}

// newRecorder builds a Recorder over SDK exporters and readers. Tests inject
// capturing implementations through it.
func newRecorder(logExporter sdklog.Exporter, spanExporter sdktrace.SpanExporter, metricReader sdkmetric.Reader, environment string) (*Recorder, error) {
	sharedResource := newResource(environment)

	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter,
			sdklog.WithExportInterval(exportInterval),
			sdklog.WithExportTimeout(exportTimeout),
			sdklog.WithMaxQueueSize(maxExportQueueSize),
			sdklog.WithExportMaxBatchSize(exportMaxBatchSize),
		)),
		sdklog.WithResource(sharedResource),
	)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(spanExporter,
			sdktrace.WithBatchTimeout(exportInterval),
			sdktrace.WithExportTimeout(exportTimeout),
			sdktrace.WithMaxQueueSize(maxExportQueueSize),
			sdktrace.WithMaxExportBatchSize(exportMaxBatchSize),
		)),
		sdktrace.WithResource(sharedResource),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricReader),
		sdkmetric.WithResource(sharedResource),
		sdkmetric.WithView(requestDurationView()),
	)

	recorder := &Recorder{
		logProvider:    logProvider,
		logger:         logProvider.Logger(instrumentationName),
		tracerProvider: tracerProvider,
		tracer:         tracerProvider.Tracer(instrumentationName),
		meterProvider:  meterProvider,
	}
	if err := recorder.buildInstruments(); err != nil {
		return nil, err
	}
	return recorder, nil
}

// newResource is the resource shared by every signal, so logs, metrics, and
// traces join on the service name and deployment environment.
func newResource(environment string) *resource.Resource {
	attributes := []attribute.KeyValue{attribute.String("service.name", serviceName)}
	if environment != "" {
		attributes = append(attributes, attribute.String("deployment.environment", environment))
	}
	return resource.NewSchemaless(attributes...)
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

// ReportExportError writes a redacted export failure to the local diagnostic
// output. It never feeds back into the export pipeline.
func ReportExportError(err error) {
	if err == nil || logger.StdErr == nil {
		return
	}
	logger.StdErr.Printf("observability export failed: %s", Redact(err.Error()))
}
