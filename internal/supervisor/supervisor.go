// Package supervisor owns bridge lifecycle, health, and bounded recovery.
package supervisor

import (
	"context"
	"sync"
	"time"

	"github.com/theworker02/centralizer/internal/telemetry"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/health"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Config controls recovery behavior.
type Config struct {
	AutoRecover  bool
	MaxRestarts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	PingEvery    time.Duration
}

func (c Config) withDefaults() Config {
	if c.MaxRestarts <= 0 {
		c.MaxRestarts = 5
	}
	if c.InitialDelay <= 0 {
		c.InitialDelay = 200 * time.Millisecond
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = 5 * time.Second
	}
	if c.PingEvery <= 0 {
		c.PingEvery = 10 * time.Second
	}
	return c
}

// Factory rebuilds a bridge after failure.
type Factory func(ctx context.Context, plan bridge.Plan) (bridge.Bridge, error)

// Supervisor wraps a Bridge with health tracking and bounded retries.
type Supervisor struct {
	name    string
	cfg     Config
	factory Factory
	plan    bridge.Plan

	mu        sync.Mutex
	inner     bridge.Bridge
	state     health.State
	restarts  int
	fallbacks int
	calls     int64
	errors    int64
	lastErr   string
	lastOK    time.Time
	latency   time.Duration
	breaker   *Breaker
	closed    bool
	pid       int
}

// New wraps inner. factory may be nil if recovery is disabled.
func New(name string, inner bridge.Bridge, plan bridge.Plan, factory Factory, cfg Config) *Supervisor {
	cfg = cfg.withDefaults()
	return &Supervisor{
		name:    name,
		cfg:     cfg,
		factory: factory,
		plan:    plan,
		inner:   inner,
		state:   health.Healthy,
		breaker: NewBreaker(),
	}
}

func (s *Supervisor) ID() string { return s.inner.ID() }

func (s *Supervisor) Plan() bridge.Plan { return s.plan }

func (s *Supervisor) SetPID(pid int) {
	s.mu.Lock()
	s.pid = pid
	s.mu.Unlock()
}

func (s *Supervisor) Describe(ctx context.Context) (*schema.Schema, error) {
	s.mu.Lock()
	inner := s.inner
	s.mu.Unlock()
	return inner.Describe(ctx)
}

func (s *Supervisor) Call(ctx context.Context, fn string, args map[string]cir.Value) (cir.Value, error) {
	return s.Invoke(ctx, bridge.Invocation{Function: fn, Args: args})
}

func (s *Supervisor) Invoke(ctx context.Context, inv bridge.Invocation) (cir.Value, error) {
	start := time.Now()
	v, err := s.do(ctx, func(b bridge.Bridge) (cir.Value, error) {
		return b.Invoke(ctx, inv)
	})
	s.observe(start, err)
	return v, err
}

func (s *Supervisor) New(ctx context.Context, typeName string, args map[string]cir.Value) (cir.Value, error) {
	return s.do(ctx, func(b bridge.Bridge) (cir.Value, error) {
		return b.New(ctx, typeName, args)
	})
}

func (s *Supervisor) Get(ctx context.Context, handle, property string) (cir.Value, error) {
	return s.do(ctx, func(b bridge.Bridge) (cir.Value, error) {
		return b.Get(ctx, handle, property)
	})
}

func (s *Supervisor) Set(ctx context.Context, handle, property string, value cir.Value) error {
	_, err := s.do(ctx, func(b bridge.Bridge) (cir.Value, error) {
		return cir.Null(), b.Set(ctx, handle, property, value)
	})
	return err
}

func (s *Supervisor) Release(ctx context.Context, handle string) error {
	_, err := s.do(ctx, func(b bridge.Bridge) (cir.Value, error) {
		return cir.Null(), b.Release(ctx, handle)
	})
	return err
}

func (s *Supervisor) Stream(ctx context.Context, name string, args map[string]cir.Value) (bridge.Stream, error) {
	s.mu.Lock()
	inner := s.inner
	s.mu.Unlock()
	st, err := inner.Stream(ctx, name, args)
	if err == nil {
		telemetry.DefaultMetrics.Streams.Add(1)
	}
	return st, err
}

func (s *Supervisor) Subscribe(ctx context.Context, event string) (bridge.Stream, error) {
	s.mu.Lock()
	inner := s.inner
	s.mu.Unlock()
	return inner.Subscribe(ctx, event)
}

func (s *Supervisor) Ping(ctx context.Context) error {
	s.mu.Lock()
	inner := s.inner
	s.mu.Unlock()
	return inner.Ping(ctx)
}

func (s *Supervisor) Close(ctx context.Context) error {
	s.mu.Lock()
	s.closed = true
	s.state = health.Stopping
	inner := s.inner
	s.mu.Unlock()
	err := inner.Close(ctx)
	s.mu.Lock()
	s.state = health.Stopped
	s.mu.Unlock()
	return err
}

func (s *Supervisor) do(ctx context.Context, fn func(bridge.Bridge) (cir.Value, error)) (cir.Value, error) {
	if err := s.breaker.Allow(); err != nil {
		return cir.Value{}, err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return cir.Value{}, czerr.ErrClosed
	}
	if s.state == health.Quarantined {
		s.mu.Unlock()
		return cir.Value{}, czerr.ErrQuarantined
	}
	inner := s.inner
	s.mu.Unlock()

	v, err := fn(inner)
	if err == nil {
		s.breaker.RecordSuccess()
		return v, nil
	}
	if !recoverable(err) || !s.cfg.AutoRecover || s.factory == nil {
		s.breaker.RecordFailure()
		return cir.Value{}, err
	}
	if recErr := s.recover(ctx); recErr != nil {
		s.breaker.RecordFailure()
		return cir.Value{}, recErr
	}
	s.mu.Lock()
	inner = s.inner
	s.mu.Unlock()
	v, err = fn(inner)
	if err != nil {
		s.breaker.RecordFailure()
		return cir.Value{}, err
	}
	s.breaker.RecordSuccess()
	return v, nil
}

func (s *Supervisor) recover(ctx context.Context) error {
	s.mu.Lock()
	if s.restarts >= s.cfg.MaxRestarts {
		s.state = health.Quarantined
		s.mu.Unlock()
		return czerr.New(czerr.ErrQuarantined, "restart budget exhausted")
	}
	s.state = health.Recovering
	plan := s.plan
	attempt := s.restarts
	fallbacks := append([]bridge.Strategy(nil), s.plan.Fallbacks...)
	s.mu.Unlock()

	delay := s.cfg.InitialDelay << attempt
	if delay > s.cfg.MaxDelay {
		delay = s.cfg.MaxDelay
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return czerr.Wrap(czerr.ErrCancelled, "recovery", ctx.Err())
	case <-timer.C:
	}

	b, err := s.factory(ctx, plan)
	if err != nil && len(fallbacks) > 0 {
		fb := plan
		fb.Strategy = fallbacks[0]
		fb.Transport = bridge.TransportName(fallbacks[0])
		b, err = s.factory(ctx, fb)
		if err == nil {
			s.mu.Lock()
			s.fallbacks++
			s.plan = fb
			s.mu.Unlock()
		}
	}
	if err != nil {
		s.mu.Lock()
		s.restarts++
		s.lastErr = err.Error()
		if s.restarts >= s.cfg.MaxRestarts {
			s.state = health.Quarantined
		} else {
			s.state = health.Unhealthy
		}
		s.mu.Unlock()
		return czerr.Wrap(czerr.ErrBridgeFailed, "reconnect", err)
	}
	s.mu.Lock()
	if s.inner != nil {
		_ = s.inner.Close(ctx)
	}
	s.inner = b
	s.restarts++
	s.state = health.Healthy
	s.lastErr = ""
	s.mu.Unlock()
	telemetry.DefaultMetrics.BridgeRestarts.Add(1)
	return nil
}

func (s *Supervisor) observe(start time.Time, err error) {
	d := time.Since(start)
	s.mu.Lock()
	s.calls++
	s.latency = d
	if err != nil {
		s.errors++
		s.lastErr = err.Error()
		if s.state == health.Healthy {
			s.state = health.Degraded
		}
	} else {
		s.lastOK = time.Now().UTC()
		if s.state == health.Degraded {
			s.state = health.Healthy
		}
	}
	s.mu.Unlock()
	telemetry.DefaultMetrics.Calls.Add(1)
	telemetry.DefaultMetrics.ObserveLatency(d)
	if err != nil {
		telemetry.DefaultMetrics.Errors.Add(1)
	}
}

// Snapshot returns machine-readable health.
func (s *Supervisor) Snapshot() health.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	rate := 1.0
	if s.calls > 0 {
		rate = float64(s.calls-s.errors) / float64(s.calls)
	}
	return health.Snapshot{
		Service:     s.name,
		State:       s.state,
		Transport:   s.plan.Transport,
		Runtime:     s.plan.Runtime,
		Language:    s.plan.Language,
		Latency:     s.latency,
		SuccessRate: rate,
		Restarts:    s.restarts,
		Fallbacks:   s.fallbacks,
		Calls:       s.calls,
		Errors:      s.errors,
		LastError:   s.lastErr,
		LastOK:      s.lastOK,
		PID:         s.pid,
		Breaker:     string(s.breaker.State()),
		ObservedAt:  time.Now().UTC(),
	}
}

func recoverable(err error) bool {
	if err == nil {
		return false
	}
	switch {
	case is(err, czerr.ErrBridgeFailed), is(err, czerr.ErrTransportFailure),
		is(err, czerr.ErrTimeout):
		return true
	default:
		return false
	}
}

func is(err, target error) bool {
	type iser interface{ Is(error) bool }
	if u, ok := err.(iser); ok {
		return u.Is(target)
	}
	return err == target
}
