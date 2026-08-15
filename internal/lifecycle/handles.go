// Package lifecycle owns handle tables and graceful shutdown.
package lifecycle

import (
	"context"
	"sync"
	"time"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// Handle is a registry entry for a foreign object. It never stores a
// raw memory pointer from another runtime.
type Handle struct {
	ID        string
	BridgeID  string
	TypeName  string
	Refs      int
	Expires   time.Time
	CreatedAt time.Time
}

// Table tracks opaque handles.
type Table struct {
	mu      sync.Mutex
	handles map[string]*Handle
}

// NewTable returns an empty handle table.
func NewTable() *Table {
	return &Table{handles: map[string]*Handle{}}
}

// Put records a handle with an initial refcount of 1.
func (t *Table) Put(h Handle) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if h.Refs == 0 {
		h.Refs = 1
	}
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now().UTC()
	}
	cp := h
	t.handles[h.ID] = &cp
}

// Get validates a handle id.
func (t *Table) Get(id string) (*Handle, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	h, ok := t.handles[id]
	if !ok {
		return nil, czerr.New(czerr.ErrHandleInvalid, id)
	}
	if !h.Expires.IsZero() && time.Now().After(h.Expires) {
		delete(t.handles, id)
		return nil, czerr.New(czerr.ErrHandleInvalid, "expired "+id)
	}
	return h, nil
}

// Release decrements the refcount and deletes at zero.
func (t *Table) Release(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	h, ok := t.handles[id]
	if !ok {
		return czerr.New(czerr.ErrHandleInvalid, id)
	}
	h.Refs--
	if h.Refs <= 0 {
		delete(t.handles, id)
	}
	return nil
}

// DropBridge invalidates every handle owned by a disconnected bridge.
func (t *Table) DropBridge(bridgeID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, h := range t.handles {
		if h.BridgeID == bridgeID {
			delete(t.handles, id)
		}
	}
}

// RejectIfExpired returns ErrHandleInvalid when id is known and past its
// expiry. Unknown ids return nil so callers can still ask the remote.
func (t *Table) RejectIfExpired(id string) error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	h, ok := t.handles[id]
	if !ok {
		return nil
	}
	if !h.Expires.IsZero() && time.Now().After(h.Expires) {
		delete(t.handles, id)
		return czerr.New(czerr.ErrHandleInvalid, "expired "+id)
	}
	return nil
}

// SweepExpired deletes every handle whose expiry has passed.
func (t *Table) SweepExpired() int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	n := 0
	now := time.Now()
	for id, h := range t.handles {
		if !h.Expires.IsZero() && now.After(h.Expires) {
			delete(t.handles, id)
			n++
		}
	}
	return n
}

// Len returns the number of live handles.
func (t *Table) Len() int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.handles)
}

// ShutdownFunc is a cleanup hook.
type ShutdownFunc func(ctx context.Context) error

// Group runs shutdown hooks in reverse registration order.
type Group struct {
	mu    sync.Mutex
	hooks []ShutdownFunc
}

// Add registers a hook.
func (g *Group) Add(fn ShutdownFunc) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hooks = append(g.hooks, fn)
}

// Shutdown runs hooks. Child processes must be reaped here.
func (g *Group) Shutdown(ctx context.Context) error {
	g.mu.Lock()
	hooks := append([]ShutdownFunc(nil), g.hooks...)
	g.mu.Unlock()
	var first error
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i](ctx); err != nil && first == nil {
			first = err
		}
	}
	return first
}
