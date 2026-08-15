package telemetry

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics is a process-local counter/histogram set. An OpenTelemetry
// Meter implementation may scrape Snapshot() if attached by the host.
type Metrics struct {
	Calls            atomic.Int64
	Errors           atomic.Int64
	Timeouts         atomic.Int64
	BytesIn          atomic.Int64
	BytesOut         atomic.Int64
	BridgeRestarts   atomic.Int64
	Streams          atomic.Int64
	RuntimeStartups  atomic.Int64
	PlannerDecisions atomic.Int64
	mu               sync.Mutex
	latencyNanos     []int64
}

// DefaultMetrics is the shared in-process meter.
var DefaultMetrics = NewMetrics()

// NewMetrics constructs an empty meter.
func NewMetrics() *Metrics { return &Metrics{} }

// ObserveLatency records one call duration.
func (m *Metrics) ObserveLatency(d time.Duration) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.latencyNanos) > 4096 {
		m.latencyNanos = m.latencyNanos[len(m.latencyNanos)/2:]
	}
	m.latencyNanos = append(m.latencyNanos, d.Nanoseconds())
}

// Snapshot is a point-in-time export.
type Snapshot struct {
	Calls            int64   `json:"calls"`
	Errors           int64   `json:"errors"`
	Timeouts         int64   `json:"timeouts"`
	BytesIn          int64   `json:"bytes_in"`
	BytesOut         int64   `json:"bytes_out"`
	BridgeRestarts   int64   `json:"bridge_restarts"`
	Streams          int64   `json:"streams"`
	RuntimeStartups  int64   `json:"runtime_startups"`
	PlannerDecisions int64   `json:"planner_decisions"`
	LatencyP50MS     float64 `json:"latency_p50_ms"`
	LatencyP99MS     float64 `json:"latency_p99_ms"`
}

// Snap returns current counters.
func (m *Metrics) Snap() Snapshot {
	if m == nil {
		return Snapshot{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Snapshot{
		Calls:            m.Calls.Load(),
		Errors:           m.Errors.Load(),
		Timeouts:         m.Timeouts.Load(),
		BytesIn:          m.BytesIn.Load(),
		BytesOut:         m.BytesOut.Load(),
		BridgeRestarts:   m.BridgeRestarts.Load(),
		Streams:          m.Streams.Load(),
		RuntimeStartups:  m.RuntimeStartups.Load(),
		PlannerDecisions: m.PlannerDecisions.Load(),
	}
	if n := len(m.latencyNanos); n > 0 {
		cp := append([]int64(nil), m.latencyNanos...)
		sortInt64(cp)
		s.LatencyP50MS = float64(cp[n/2]) / 1e6
		idx := int(float64(n) * 0.99)
		if idx >= n {
			idx = n - 1
		}
		s.LatencyP99MS = float64(cp[idx]) / 1e6
	}
	return s
}

func sortInt64(a []int64) {
	for i := 1; i < len(a); i++ {
		j := i
		for j > 0 && a[j-1] > a[j] {
			a[j-1], a[j] = a[j], a[j-1]
			j--
		}
	}
}
