// Package adapter defines the contract implemented by language runtimes.
//
// Adapters are independently testable. They must not duplicate Centralizer
// orchestration, planning, or protocol logic. Register adapters through
// the in-process Registry; do not rely on Go's plugin system.
package adapter

import (
	"context"
	"errors"
	"sync"

	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Target is the discovery input presented to an adapter.
type Target struct {
	Ref        string            `json:"ref"`
	Path       string            `json:"path"`
	Entry      string            `json:"entry,omitempty"`
	Language   string            `json:"language,omitempty"`
	Hints      map[string]string `json:"hints,omitempty"`
	Executable bool              `json:"executable,omitempty"`
}

// Detection is a scored language/runtime hypothesis.
type Detection struct {
	Adapter    string   `json:"adapter"`
	Language   string   `json:"language"`
	Runtime    string   `json:"runtime"`
	Version    string   `json:"version,omitempty"`
	Arch       string   `json:"arch,omitempty"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence,omitempty"`
	Primary    bool     `json:"primary,omitempty"`
}

// Adapter discovers, prepares, and connects a target.
type Adapter interface {
	Name() string
	Detect(ctx context.Context, target Target) (Detection, error)
	Capabilities(ctx context.Context, target Target) ([]capability.Capability, error)
	Prepare(ctx context.Context, target Target) error
	Connect(ctx context.Context, target Target, plan bridge.Plan) (bridge.Bridge, error)
}

// Registry holds adapters keyed by name. It is concurrency-safe.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{adapters: map[string]Adapter{}}
}

// Register adds or replaces an adapter.
func (r *Registry) Register(a Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[a.Name()] = a
}

// Get returns an adapter by name.
func (r *Registry) Get(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[name]
	return a, ok
}

// All returns adapters in unspecified order.
func (r *Registry) All() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Adapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		out = append(out, a)
	}
	return out
}

// Names returns registered adapter names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.adapters))
	for n := range r.adapters {
		out = append(out, n)
	}
	return out
}

// DetectAll runs detection on every adapter and returns results with
// confidence > 0. Adapters that return ErrUnsupportedTarget are skipped.
func DetectAll(ctx context.Context, r *Registry, target Target) ([]Detection, error) {
	var out []Detection
	for _, a := range r.All() {
		d, err := a.Detect(ctx, target)
		if err != nil {
			if skipDetection(err) {
				continue
			}
			return nil, err
		}
		if d.Confidence <= 0 {
			continue
		}
		if d.Adapter == "" {
			d.Adapter = a.Name()
		}
		out = append(out, d)
	}
	return out, nil
}

func skipDetection(err error) bool {
	return errors.Is(err, czerr.ErrUnsupportedTarget) || errors.Is(err, czerr.ErrTargetNotFound)
}
