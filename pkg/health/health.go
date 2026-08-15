// Package health defines bridge and service health snapshots.
package health

import "time"

// State is a supervisor lifecycle / health state.
type State string

const (
	Created     State = "created"
	Starting    State = "starting"
	Healthy     State = "healthy"
	Degraded    State = "degraded"
	Recovering  State = "recovering"
	Unhealthy   State = "unhealthy"
	Quarantined State = "quarantined"
	Stopping    State = "stopping"
	Stopped     State = "stopped"
)

// Snapshot is a point-in-time view of a connected service.
type Snapshot struct {
	Service     string        `json:"service"`
	State       State         `json:"state"`
	Transport   string        `json:"transport"`
	Runtime     string        `json:"runtime"`
	Language    string        `json:"language"`
	Latency     time.Duration `json:"latency"`
	SuccessRate float64       `json:"success_rate"`
	Restarts    int           `json:"restarts"`
	Fallbacks   int           `json:"fallbacks"`
	Calls       int64         `json:"calls"`
	Errors      int64         `json:"errors"`
	LastError   string        `json:"last_error,omitempty"`
	LastOK      time.Time     `json:"last_ok,omitempty"`
	PID         int           `json:"pid,omitempty"`
	Breaker     string        `json:"breaker,omitempty"`
	ObservedAt  time.Time     `json:"observed_at"`
}

// Text renders a human-readable health block.
func (s Snapshot) Text() string {
	return formatSnapshot(s)
}
