package session

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Handler is an in-process function table used by the Go adapter.
type Handler struct {
	Name    string
	Schema  *schema.Schema
	Funcs   map[string]Func
	Types   map[string]Constructor
	Objects map[string]any
}

// Func is a native callable.
type Func func(ctx context.Context, args map[string]cir.Value) (cir.Value, error)

// Constructor creates a handle-backed object.
type Constructor func(ctx context.Context, args map[string]cir.Value) (any, error)

// NativeBridge calls Go functions without leaving the process.
type NativeBridge struct {
	id      string
	plan    bridge.Plan
	handler *Handler
	mu      sync.Mutex
	handles map[string]any
	next    atomic.Uint64
	closed  bool
}

// NewNative constructs an in-process bridge.
func NewNative(plan bridge.Plan, h *Handler) *NativeBridge {
	if h.Funcs == nil {
		h.Funcs = map[string]Func{}
	}
	if h.Objects == nil {
		h.Objects = map[string]any{}
	}
	return &NativeBridge{
		id:      "native-" + h.Name,
		plan:    plan,
		handler: h,
		handles: map[string]any{},
	}
}

func (n *NativeBridge) ID() string        { return n.id }
func (n *NativeBridge) Plan() bridge.Plan { return n.plan }

// SetSchema installs an explicit schema used by Describe and validation.
func (n *NativeBridge) SetSchema(sc *schema.Schema) {
	if n.handler != nil {
		n.handler.Schema = sc
	}
}

func (n *NativeBridge) Describe(context.Context) (*schema.Schema, error) {
	if n.handler.Schema != nil {
		return n.handler.Schema, nil
	}
	s := &schema.Schema{Service: n.handler.Name, Inferred: true, Functions: map[string]schema.Function{}}
	for name := range n.handler.Funcs {
		s.Functions[name] = schema.Function{}
	}
	return s, nil
}

func (n *NativeBridge) Call(ctx context.Context, fn string, args map[string]cir.Value) (cir.Value, error) {
	return n.Invoke(ctx, bridge.Invocation{Function: fn, Args: args})
}

func (n *NativeBridge) Invoke(ctx context.Context, inv bridge.Invocation) (cir.Value, error) {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return cir.Value{}, czerr.ErrClosed
	}
	n.mu.Unlock()
	if inv.Handle != "" && inv.Method != "" {
		return cir.Value{}, czerr.New(czerr.ErrNotImplemented, "native methods")
	}
	fn, ok := n.handler.Funcs[inv.Function]
	if !ok {
		return cir.Value{}, czerr.New(czerr.ErrSchemaMismatch, "unknown function "+inv.Function)
	}
	return fn(ctx, inv.Args)
}

func (n *NativeBridge) New(ctx context.Context, typeName string, args map[string]cir.Value) (cir.Value, error) {
	ctor, ok := n.handler.Types[typeName]
	if !ok {
		return cir.Value{}, czerr.New(czerr.ErrSchemaMismatch, "unknown type "+typeName)
	}
	obj, err := ctor(ctx, args)
	if err != nil {
		return cir.Value{}, err
	}
	id := fmt.Sprintf("h-%d", n.next.Add(1))
	n.mu.Lock()
	n.handles[id] = obj
	n.mu.Unlock()
	return cir.Handle(id), nil
}

func (n *NativeBridge) Get(context.Context, string, string) (cir.Value, error) {
	return cir.Value{}, czerr.ErrNotImplemented
}

func (n *NativeBridge) Set(context.Context, string, string, cir.Value) error {
	return czerr.ErrNotImplemented
}

func (n *NativeBridge) Release(_ context.Context, handle string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.handles[handle]; !ok {
		return czerr.New(czerr.ErrHandleInvalid, handle)
	}
	delete(n.handles, handle)
	return nil
}

func (n *NativeBridge) Stream(context.Context, string, map[string]cir.Value) (bridge.Stream, error) {
	return nil, czerr.ErrNotImplemented
}

func (n *NativeBridge) Subscribe(context.Context, string) (bridge.Stream, error) {
	return nil, czerr.ErrNotImplemented
}

func (n *NativeBridge) Ping(context.Context) error { return nil }

func (n *NativeBridge) Close(context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	n.handles = map[string]any{}
	return nil
}
