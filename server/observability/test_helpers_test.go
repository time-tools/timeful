package observability

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
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
