package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// PoolStats is one sample of the PostgreSQL connection pool.
type PoolStats struct {
	Used int
	Idle int
	Max  int
}

// buildInstruments and registerPoolMetrics stay handwritten beside PoolStats
// because Recorder is declared in provider_recorder.go and its method set is
// shared with transport.go, so the receiver's methods stay in one language.
func (r *Recorder) buildInstruments() error {
	meter := r.meterProvider.Meter(instrumentationName)

	duration, err := meter.Float64Histogram(requestDurationMetric,
		metric.WithUnit("s"),
		metric.WithDescription("Duration of inbound HTTP requests."),
	)
	if err != nil {
		return fmt.Errorf("build request duration histogram: %w", err)
	}
	readiness, err := meter.Int64Gauge(readinessMetric,
		metric.WithDescription("Service readiness: 1 when ready, 0 when unavailable."),
	)
	if err != nil {
		return fmt.Errorf("build readiness gauge: %w", err)
	}
	r.requestDuration = duration
	r.readinessGauge = readiness
	return r.registerPoolMetrics(meter)
}

func (r *Recorder) registerPoolMetrics(meter metric.Meter) error {
	connections, err := meter.Int64ObservableUpDownCounter(poolConnectionsMetric,
		metric.WithDescription("PostgreSQL pool connections by state."),
	)
	if err != nil {
		return fmt.Errorf("build pool connections counter: %w", err)
	}
	maximum, err := meter.Int64ObservableUpDownCounter(poolMaxConnectionsMetric,
		metric.WithDescription("Maximum PostgreSQL pool connections."),
	)
	if err != nil {
		return fmt.Errorf("build pool maximum counter: %w", err)
	}
	_, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		source := r.poolStats.Load()
		if source == nil {
			return nil
		}
		stats := (*source)()
		observer.ObserveInt64(connections, int64(stats.Used),
			metric.WithAttributes(attribute.String("db.client.connection.state", poolUsedState)))
		observer.ObserveInt64(connections, int64(stats.Idle),
			metric.WithAttributes(attribute.String("db.client.connection.state", poolIdleState)))
		observer.ObserveInt64(maximum, int64(stats.Max))
		return nil
	}, connections, maximum)
	if err != nil {
		return fmt.Errorf("register pool metrics callback: %w", err)
	}
	return nil
}
