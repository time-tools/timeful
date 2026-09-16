package observability

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadinessMonitorTracksProbeResult(t *testing.T) {
	var unavailable atomic.Bool
	monitor := NewReadinessMonitor(func(context.Context) error {
		if unavailable.Load() {
			return errors.New("postgres unavailable")
		}
		return nil
	}, 10*time.Millisecond, 50*time.Millisecond)

	monitor.Start(context.Background())
	if got := monitor.State(); got != ReadinessReady {
		t.Fatalf("State() = %q, want ready", got)
	}

	unavailable.Store(true)
	waitForReadiness(t, monitor, ReadinessUnavailable)

	unavailable.Store(false)
	waitForReadiness(t, monitor, ReadinessReady)
}

func TestReadinessMonitorTimesOutBlockedProbe(t *testing.T) {
	monitor := NewReadinessMonitor(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}, time.Hour, 10*time.Millisecond)

	monitor.Start(context.Background())
	if got := monitor.State(); got != ReadinessUnavailable {
		t.Fatalf("State() = %q, want unavailable", got)
	}
}

func TestReadinessMonitorNilSafe(t *testing.T) {
	var monitor *ReadinessMonitor
	monitor.Start(context.Background())
	if got := monitor.State(); got != ReadinessUnknown {
		t.Fatalf("State() = %q, want unknown", got)
	}
}

func waitForReadiness(t *testing.T, monitor *ReadinessMonitor, want ReadinessState) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if monitor.State() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("State() = %q, want %q", monitor.State(), want)
}
