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

func TestReadinessMonitorNotifiesObserver(t *testing.T) {
	var unavailable atomic.Bool
	monitor := NewReadinessMonitor(func(context.Context) error {
		if unavailable.Load() {
			return errors.New("postgres unavailable")
		}
		return nil
	}, 5*time.Millisecond, 50*time.Millisecond)

	observed := make(chan ReadinessState, 8)
	monitor.SetObserver(func(state ReadinessState) {
		select {
		case observed <- state:
		default:
		}
	})
	monitor.Start(context.Background())

	waitForObservedState(t, observed, ReadinessReady)
	unavailable.Store(true)
	waitForObservedState(t, observed, ReadinessUnavailable)
}

func TestReadinessMonitorNilSafe(t *testing.T) {
	var monitor *ReadinessMonitor
	monitor.SetObserver(func(ReadinessState) {})
	monitor.Start(context.Background())
	if got := monitor.State(); got != ReadinessUnknown {
		t.Fatalf("State() = %q, want unknown", got)
	}
}

func waitForObservedState(t *testing.T, observed <-chan ReadinessState, want ReadinessState) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case state := <-observed:
			if state == want {
				return
			}
		case <-deadline:
			t.Fatalf("observer never reported %q", want)
		}
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
