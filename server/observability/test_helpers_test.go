package observability

import (
	"context"
	"sync"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// capturingExporter records whatever the batch processor exports so tests can
// assert on the exact structured records without a network.
type capturingExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *capturingExporter) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range records {
		e.records = append(e.records, records[i].Clone())
	}
	return nil
}

func (e *capturingExporter) Shutdown(context.Context) error   { return nil }
func (e *capturingExporter) ForceFlush(context.Context) error { return nil }

func (e *capturingExporter) snapshot() []sdklog.Record {
	e.mu.Lock()
	defer e.mu.Unlock()
	records := make([]sdklog.Record, len(e.records))
	copy(records, e.records)
	return records
}

// capturingSpanExporter records spans and, unlike the SDK in-memory exporter,
// retains them across Shutdown so assertions can run after a flush.
type capturingSpanExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (e *capturingSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, spans...)
	return nil
}

func (e *capturingSpanExporter) Shutdown(context.Context) error { return nil }

func (e *capturingSpanExporter) snapshot() []sdktrace.ReadOnlySpan {
	e.mu.Lock()
	defer e.mu.Unlock()
	spans := make([]sdktrace.ReadOnlySpan, len(e.spans))
	copy(spans, e.spans)
	return spans
}

// testRecorder bundles a Recorder with the in-memory sinks its pipelines
// export to.
type testRecorder struct {
	recorder *Recorder
	logs     *capturingExporter
	spans    *capturingSpanExporter
	metrics  *sdkmetric.ManualReader
}

func newTestRecorder(t *testing.T) *testRecorder {
	t.Helper()
	logs := &capturingExporter{}
	spans := &capturingSpanExporter{}
	metrics := sdkmetric.NewManualReader()
	recorder, err := newRecorder(logs, spans, metrics, "test")
	if err != nil {
		t.Fatalf("newRecorder() error = %v", err)
	}
	return &testRecorder{recorder: recorder, logs: logs, spans: spans, metrics: metrics}
}

func recordAttributes(record sdklog.Record) map[string]string {
	attributes := map[string]string{}
	record.WalkAttributes(func(kv attribute.KeyValue) bool {
		attributes[string(kv.Key)] = kv.Value.Emit()
		return true
	})
	return attributes
}

func apiRecordAttributes(record log.Record) map[string]string {
	attributes := map[string]string{}
	record.WalkAttributes(func(kv attribute.KeyValue) bool {
		attributes[string(kv.Key)] = kv.Value.Emit()
		return true
	})
	return attributes
}

func resourceAttributes(record sdklog.Record) map[string]string {
	attributes := map[string]string{}
	if resource := record.Resource(); resource != nil {
		for _, kv := range resource.Attributes() {
			attributes[string(kv.Key)] = kv.Value.Emit()
		}
	}
	return attributes
}

func keyValueAttributes(values []attribute.KeyValue) map[string]string {
	attributes := map[string]string{}
	for _, kv := range values {
		attributes[string(kv.Key)] = kv.Value.Emit()
	}
	return attributes
}

func findMetric(t *testing.T, metrics metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name == name {
				return metric
			}
		}
	}
	t.Fatalf("metric %q not found in %#v", name, metrics)
	return metricdata.Metrics{}
}
