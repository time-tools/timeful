package observability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
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
	DefaultShutdownTimeout = 5 * time.Second
)

// Recorder emits structured diagnostics through a bounded, asynchronous OTLP
// log pipeline. A nil recorder is a no-op, so instrumentation stays safe when
// the OpenObserve contract is absent.
type Recorder struct {
	provider  *sdklog.LoggerProvider
	logger    log.Logger
	readiness func() ReadinessState
}

// Start builds the OTLP/HTTP logs exporter for the configured OpenObserve
// organization and returns a Recorder. Incomplete configuration returns an
// error instead of failing startup, so the server keeps serving and keeps
// writing local diagnostics without OpenObserve.
func Start(ctx context.Context, config Config) (*Recorder, error) {
	if !config.Enabled() {
		return nil, errors.New("OpenObserve ingest configuration is incomplete")
	}
	logsURL, err := config.LogsURL()
	if err != nil {
		return nil, err
	}
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(logsURL),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": config.AuthHeader(),
			"stream-name":   ServerLogStream,
		}),
		otlploghttp.WithTimeout(exportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("build OpenObserve OTLP exporter: %w", err)
	}
	otel.SetErrorHandler(otel.ErrorHandlerFunc(ReportExportError))
	return newRecorder(exporter, config.Environment), nil
}

// newRecorder builds a Recorder over an SDK exporter. Tests inject a capturing
// exporter through it.
func newRecorder(exporter sdklog.Exporter, environment string) *Recorder {
	processor := sdklog.NewBatchProcessor(exporter,
		sdklog.WithExportInterval(exportInterval),
		sdklog.WithExportTimeout(exportTimeout),
		sdklog.WithMaxQueueSize(maxExportQueueSize),
		sdklog.WithExportMaxBatchSize(exportMaxBatchSize),
	)
	attributes := []attribute.KeyValue{attribute.String("service.name", serviceName)}
	if environment != "" {
		attributes = append(attributes, attribute.String("deployment.environment", environment))
	}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(resource.NewSchemaless(attributes...)),
	)
	return &Recorder{
		provider: provider,
		logger:   provider.Logger(instrumentationName),
	}
}

// SetReadiness attaches a readiness source that enriches every emitted request
// record. It is a no-op on a nil recorder.
func (r *Recorder) SetReadiness(source func() ReadinessState) {
	if r == nil {
		return
	}
	r.readiness = source
}

// Enabled reports whether records are shipped to a collector.
func (r *Recorder) Enabled() bool {
	return r != nil && r.provider != nil
}

// EmitRequest records one finished request. Emitting enqueues to the batch
// processor and returns without waiting for the collector.
func (r *Recorder) EmitRequest(ctx context.Context, completion RequestCompletion) {
	if r == nil || r.provider == nil {
		return
	}
	if completion.Readiness == "" && r.readiness != nil {
		completion.Readiness = r.readiness()
	}
	r.logger.Emit(ctx, completion.LogRecord())
}

// Shutdown flushes queued records and stops the export pipeline.
func (r *Recorder) Shutdown(ctx context.Context) error {
	if r == nil || r.provider == nil {
		return nil
	}
	return r.provider.Shutdown(ctx)
}

// ReportExportError writes a redacted export failure to the local diagnostic
// output. It never feeds back into the export pipeline.
func ReportExportError(err error) {
	if err == nil || logger.StdErr == nil {
		return
	}
	logger.StdErr.Printf("observability export failed: %s", Redact(err.Error()))
}
