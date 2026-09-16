package observability

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestRecordReadinessGauge(t *testing.T) {
	test := newTestRecorder(t)

	test.recorder.RecordReadiness(ReadinessReady)
	if value := collectReadiness(t, test); value != 1 {
		t.Fatalf("readiness value = %d, want 1 after ready", value)
	}

	test.recorder.RecordReadiness(ReadinessUnavailable)
	if value := collectReadiness(t, test); value != 0 {
		t.Fatalf("readiness value = %d, want 0 after unavailable", value)
	}

	test.recorder.RecordReadiness(ReadinessUnknown)
	if value := collectReadiness(t, test); value != 0 {
		t.Fatalf("readiness value = %d, want the unknown state to be ignored", value)
	}
}

func TestPoolMetricsObservation(t *testing.T) {
	test := newTestRecorder(t)
	test.recorder.SetPoolStats(func() PoolStats { return PoolStats{Used: 4, Idle: 1, Max: 12} })

	metrics := collectMetrics(t, test)
	connections := findMetric(t, metrics, poolConnectionsMetric)
	data, ok := connections.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("pool connections data = %T, want non-monotonic sum", connections.Data)
	}
	if data.IsMonotonic {
		t.Fatal("pool connections sum must be non-monotonic")
	}
	states := map[string]int64{}
	for _, point := range data.DataPoints {
		attributes := keyValueAttributes(point.Attributes.ToSlice())
		states[attributes["db.client.connection.state"]] = point.Value
	}
	if states[poolUsedState] != 4 || states[poolIdleState] != 1 {
		t.Fatalf("pool connection states = %#v, want used=4 idle=1", states)
	}

	maximum := findMetric(t, metrics, poolMaxConnectionsMetric)
	maxData, ok := maximum.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("pool maximum data = %T, want non-monotonic sum", maximum.Data)
	}
	if maxData.IsMonotonic {
		t.Fatal("pool maximum sum must be non-monotonic")
	}
	if len(maxData.DataPoints) != 1 || maxData.DataPoints[0].Value != 12 {
		t.Fatalf("pool maximum data points = %#v, want 12", maxData.DataPoints)
	}
}

func TestPoolMetricsWithoutSource(t *testing.T) {
	test := newTestRecorder(t)
	metrics := collectMetrics(t, test)
	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name == poolConnectionsMetric || metric.Name == poolMaxConnectionsMetric {
				t.Fatalf("metric %q exported without a pool stats source", metric.Name)
			}
		}
	}
}

func TestRecordRequestMetricsNilRecorderIsNoop(t *testing.T) {
	var recorder *Recorder
	recorder.RecordRequestMetrics(context.Background(), RequestCompletion{RequestID: "request-1"})
	recorder.RecordReadiness(ReadinessReady)
	recorder.SetPoolStats(func() PoolStats { return PoolStats{} })
}

func collectReadiness(t *testing.T, test *testRecorder) int64 {
	t.Helper()
	metrics := collectMetrics(t, test)
	gauge := findMetric(t, metrics, readinessMetric)
	data, ok := gauge.Data.(metricdata.Gauge[int64])
	if !ok {
		t.Fatalf("readiness data = %T, want gauge", gauge.Data)
	}
	if len(data.DataPoints) != 1 {
		t.Fatalf("readiness data points = %d, want 1", len(data.DataPoints))
	}
	return data.DataPoints[0].Value
}

func collectMetrics(t *testing.T, test *testRecorder) metricdata.ResourceMetrics {
	t.Helper()
	var metrics metricdata.ResourceMetrics
	if err := test.metrics.Collect(context.Background(), &metrics); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	return metrics
}
