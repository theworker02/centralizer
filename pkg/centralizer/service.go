package centralizer

import (
	"context"
	"time"

	"github.com/theworker02/centralizer/internal/discovery"
	"github.com/theworker02/centralizer/internal/lifecycle"
	"github.com/theworker02/centralizer/internal/planner"
	"github.com/theworker02/centralizer/internal/supervisor"
	"github.com/theworker02/centralizer/internal/telemetry"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/health"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Args is a conventional argument map accepted by Call.
type Args map[string]any

// Service is a connected target exposed through the unified API.
type Service struct {
	hub       *Hub
	name      string
	target    adapter.Target
	analysis  *discovery.Result
	planRes   *planner.Result
	sup       *supervisor.Supervisor
	adp       adapter.Adapter
	timeout   time.Duration
	tracer    *telemetry.Tracer
	handles   *lifecycle.Table
	handleTTL time.Duration
	bridgeID  string
}

// Name returns the registry name.
func (s *Service) Name() string { return s.name }

// Language returns the detected language.
func (s *Service) Language() string { return s.analysis.Primary.Language }

// Runtime returns the detected runtime label.
func (s *Service) Runtime() string {
	r := s.analysis.Primary.Runtime
	if s.analysis.Primary.Version != "" {
		return r + " " + s.analysis.Primary.Version
	}
	return r
}

// Capabilities returns the capability graph nodes.
func (s *Service) Capabilities() []capability.Capability {
	return s.analysis.Graph.Capabilities
}

// Health returns the current supervisor snapshot.
func (s *Service) Health() health.Snapshot { return s.sup.Snapshot() }

// Plan returns the selected bridge plan.
func (s *Service) Plan() bridge.Plan { return s.sup.Plan() }

// Transport returns the selected transport name.
func (s *Service) Transport() string { return s.sup.Plan().Transport }

// Schema fetches or returns the remote schema.
func (s *Service) Schema(ctx context.Context) (*schema.Schema, error) {
	return s.sup.Describe(ctx)
}

// Analysis returns the discovery result.
func (s *Service) Analysis() *discovery.Result { return s.analysis }

// Explanation returns the planner report for this connection.
func (s *Service) Explanation() string {
	return planner.Explain(planner.Input{
		Language: s.analysis.Primary.Language,
		Runtime:  s.analysis.Primary.Runtime,
		Graph:    s.analysis.Graph,
	}, s.planRes)
}

// Call invokes a function. Values are converted through CIR.
func (s *Service) Call(ctx context.Context, fn string, args Args) (cir.Value, error) {
	ctx, cancel := withTimeout(ctx, s.timeout)
	defer cancel()
	cirArgs, err := encodeArgs(args)
	if err != nil {
		return cir.Value{}, err
	}
	if sc, serr := s.sup.Describe(ctx); serr == nil {
		if err := sc.ValidateCall(fn, cirArgs); err != nil {
			return cir.Value{}, err
		}
	}
	return s.sup.Call(ctx, fn, cirArgs)
}

// Invoke is the fully specified call form.
func (s *Service) Invoke(ctx context.Context, inv bridge.Invocation) (cir.Value, error) {
	ctx, cancel := withTimeout(ctx, s.timeout)
	defer cancel()
	return s.sup.Invoke(ctx, inv)
}

// Get reads a property from a foreign handle.
func (s *Service) Get(ctx context.Context, handle, property string) (cir.Value, error) {
	if err := s.handles.RejectIfExpired(handle); err != nil {
		return cir.Value{}, err
	}
	return s.sup.Get(ctx, handle, property)
}

// Set writes a property on a foreign handle.
func (s *Service) Set(ctx context.Context, handle, property string, value any) error {
	if err := s.handles.RejectIfExpired(handle); err != nil {
		return err
	}
	v, err := cir.From(value)
	if err != nil {
		return err
	}
	return s.sup.Set(ctx, handle, property, v)
}

// New constructs a foreign object and returns an opaque handle value.
func (s *Service) New(ctx context.Context, typeName string, args Args) (cir.Value, error) {
	cirArgs, err := encodeArgs(args)
	if err != nil {
		return cir.Value{}, err
	}
	v, err := s.sup.New(ctx, typeName, cirArgs)
	if err != nil {
		return cir.Value{}, err
	}
	if id, herr := v.HandleID(); herr == nil && s.handles != nil {
		h := lifecycle.Handle{ID: id, BridgeID: s.bridgeID, TypeName: typeName}
		if s.handleTTL > 0 {
			h.Expires = time.Now().Add(s.handleTTL)
		}
		s.handles.Put(h)
	}
	return v, nil
}

// Release drops a foreign handle.
func (s *Service) Release(ctx context.Context, handle string) error {
	if s.handles != nil {
		_ = s.handles.Release(handle)
	}
	return s.sup.Release(ctx, handle)
}

// Stream opens a named stream.
func (s *Service) Stream(ctx context.Context, name string, args Args) (bridge.Stream, error) {
	cirArgs, err := encodeArgs(args)
	if err != nil {
		return nil, err
	}
	return s.sup.Stream(ctx, name, cirArgs)
}

// Subscribe opens an event stream.
func (s *Service) Subscribe(ctx context.Context, event string) (bridge.Stream, error) {
	return s.sup.Subscribe(ctx, event)
}

// Describe is an alias of Schema.
func (s *Service) Describe(ctx context.Context) (*schema.Schema, error) {
	return s.Schema(ctx)
}

// Close stops the bridge and reaps child processes.
func (s *Service) Close(ctx context.Context) error {
	if s.handles != nil {
		s.handles.SweepExpired()
		s.handles.DropBridge(s.bridgeID)
	}
	if s.hub != nil {
		s.hub.services.Remove(s.name)
	}
	return s.sup.Close(ctx)
}

func encodeArgs(args Args) (map[string]cir.Value, error) {
	if args == nil {
		return nil, nil
	}
	out := make(map[string]cir.Value, len(args))
	for k, v := range args {
		cv, err := cir.From(v)
		if err != nil {
			return nil, czerr.Wrap(czerr.ErrConversion, k, err)
		}
		out[k] = cv
	}
	return out, nil
}

func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(ctx)
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}
