// Package registry tracks connected services for API and CLI inspection.
package registry

import (
	"sync"
	"time"

	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/health"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Entry is one connected target.
type Entry struct {
	Name         string          `json:"name"`
	Target       string          `json:"target"`
	Language     string          `json:"language"`
	Runtime      string          `json:"runtime"`
	PID          int             `json:"pid,omitempty"`
	Transport    string          `json:"transport"`
	BridgeID     string          `json:"bridge_id"`
	Capabilities []string        `json:"capabilities,omitempty"`
	Schema       *schema.Schema  `json:"schema,omitempty"`
	Plan         bridge.Plan     `json:"plan"`
	Health       health.Snapshot `json:"health"`
	LastSeen     time.Time       `json:"last_seen"`
	Restarts     int             `json:"restarts"`
}

// Registry is concurrency-safe.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{entries: map[string]*Entry{}}
}

// Put inserts or replaces an entry.
func (r *Registry) Put(e *Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.LastSeen = time.Now().UTC()
	r.entries[e.Name] = e
}

// Get returns an entry by service name.
func (r *Registry) Get(name string) (*Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[name]
	return e, ok
}

// Remove deletes an entry.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, name)
}

// List returns a snapshot of all entries.
func (r *Registry) List() []*Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Entry, 0, len(r.entries))
	for _, e := range r.entries {
		cp := *e
		out = append(out, &cp)
	}
	return out
}
