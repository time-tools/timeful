package observability

import (
	"context"
	"sync/atomic"
	"time"
)

// ReadinessState is the last observed service readiness for request records.
type ReadinessState string

const (
	ReadinessUnknown     ReadinessState = "unknown"
	ReadinessReady       ReadinessState = "ready"
	ReadinessUnavailable ReadinessState = "unavailable"
)

const (
	// DefaultReadinessInterval is how often the readiness probe runs.
	DefaultReadinessInterval = 10 * time.Second
	// DefaultReadinessTimeout bounds a single readiness probe.
	DefaultReadinessTimeout = 2 * time.Second
)

// Probe reports whether a dependency is currently reachable.
type Probe func(context.Context) error

// ReadinessMonitor tracks the last observed readiness of a dependency with a
// periodic background probe, so request records read the state without waiting
// on the dependency themselves.
type ReadinessMonitor struct {
	probe    Probe
	interval time.Duration
	timeout  time.Duration
	state    atomic.Value
	observer atomic.Pointer[func(ReadinessState)]
}

// NewReadinessMonitor builds a monitor around the given probe. A nil probe
// keeps the state unknown.
func NewReadinessMonitor(probe Probe, interval, timeout time.Duration) *ReadinessMonitor {
	if interval <= 0 {
		interval = DefaultReadinessInterval
	}
	if timeout <= 0 {
		timeout = DefaultReadinessTimeout
	}
	monitor := &ReadinessMonitor{probe: probe, interval: interval, timeout: timeout}
	monitor.state.Store(ReadinessUnknown)
	return monitor
}

// Start probes once and then keeps probing until ctx is canceled. Calling
// Start on a nil monitor or one without a probe is a no-op.
func (m *ReadinessMonitor) Start(ctx context.Context) {
	if m == nil || m.probe == nil {
		return
	}
	m.check()
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.check()
			}
		}
	}()
}

// SetObserver attaches a callback invoked after every probe with the observed
// state, so readiness can be exported as a metric. A nil observer clears it,
// and a nil monitor is a no-op.
func (m *ReadinessMonitor) SetObserver(observer func(ReadinessState)) {
	if m == nil {
		return
	}
	if observer == nil {
		m.observer.Store(nil)
		return
	}
	m.observer.Store(&observer)
}

// State returns the last observed readiness, or unknown before the first
// probe.
func (m *ReadinessMonitor) State() ReadinessState {
	if m == nil {
		return ReadinessUnknown
	}
	state, _ := m.state.Load().(ReadinessState)
	if state == "" {
		return ReadinessUnknown
	}
	return state
}

func (m *ReadinessMonitor) check() {
	probeContext, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	state := ReadinessReady
	if err := m.probe(probeContext); err != nil {
		state = ReadinessUnavailable
	}
	m.state.Store(state)
	if observer := m.observer.Load(); observer != nil {
		(*observer)(state)
	}
}
