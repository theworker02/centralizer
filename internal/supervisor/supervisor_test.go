package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/schema"
)

type fakeBridge struct {
	id     string
	fail   int
	calls  int
	closed bool
	onCall func() (cir.Value, error)
}

func (f *fakeBridge) ID() string        { return f.id }
func (f *fakeBridge) Plan() bridge.Plan { return bridge.Plan{} }
func (f *fakeBridge) Describe(context.Context) (*schema.Schema, error) {
	return &schema.Schema{Service: "t"}, nil
}
func (f *fakeBridge) Call(ctx context.Context, fn string, args map[string]cir.Value) (cir.Value, error) {
	return f.Invoke(ctx, bridge.Invocation{Function: fn, Args: args})
}
func (f *fakeBridge) Invoke(context.Context, bridge.Invocation) (cir.Value, error) {
	f.calls++
	if f.onCall != nil {
		return f.onCall()
	}
	if f.fail > 0 {
		f.fail--
		return cir.Value{}, czerr.New(czerr.ErrBridgeFailed, "boom")
	}
	return cir.Int(1), nil
}
func (f *fakeBridge) New(context.Context, string, map[string]cir.Value) (cir.Value, error) {
	return cir.Handle("h"), nil
}
func (f *fakeBridge) Get(context.Context, string, string) (cir.Value, error) {
	return cir.Null(), nil
}
func (f *fakeBridge) Set(context.Context, string, string, cir.Value) error { return nil }
func (f *fakeBridge) Release(context.Context, string) error                { return nil }
func (f *fakeBridge) Stream(context.Context, string, map[string]cir.Value) (bridge.Stream, error) {
	return nil, czerr.ErrNotImplemented
}
func (f *fakeBridge) Subscribe(context.Context, string) (bridge.Stream, error) {
	return nil, czerr.ErrNotImplemented
}
func (f *fakeBridge) Ping(context.Context) error { return nil }
func (f *fakeBridge) Close(context.Context) error {
	f.closed = true
	return nil
}

func TestRecoverOnce(t *testing.T) {
	inner := &fakeBridge{id: "a", fail: 1}
	rebuilt := &fakeBridge{id: "b"}
	s := New("svc", inner, bridge.Plan{Transport: "stdio"}, func(context.Context, bridge.Plan) (bridge.Bridge, error) {
		return rebuilt, nil
	}, Config{AutoRecover: true, MaxRestarts: 3, InitialDelay: time.Millisecond})
	v, err := s.Call(context.Background(), "f", nil)
	if err != nil {
		t.Fatal(err)
	}
	i, err := v.Int()
	if err != nil || i != 1 {
		t.Fatalf("%v %v", i, err)
	}
	if s.Snapshot().Restarts != 1 {
		t.Fatalf("restarts=%d", s.Snapshot().Restarts)
	}
}

func TestRestartBudget(t *testing.T) {
	inner := &fakeBridge{id: "a", onCall: func() (cir.Value, error) {
		return cir.Value{}, czerr.New(czerr.ErrBridgeFailed, "down")
	}}
	s := New("svc", inner, bridge.Plan{}, func(context.Context, bridge.Plan) (bridge.Bridge, error) {
		return nil, errors.New("still down")
	}, Config{AutoRecover: true, MaxRestarts: 1, InitialDelay: time.Millisecond})
	_, err := s.Call(context.Background(), "f", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = s.Call(context.Background(), "f", nil)
	if !is(err, czerr.ErrQuarantined) && err != czerr.ErrQuarantined {
		// second call should be quarantined after budget
		if s.Snapshot().State != "quarantined" && !is(err, czerr.ErrQuarantined) {
			t.Fatalf("err=%v state=%s", err, s.Snapshot().State)
		}
	}
}

func TestBreakerOpens(t *testing.T) {
	b := NewBreaker()
	b.FailureThreshold = 2
	b.OpenFor = time.Hour
	b.RecordFailure()
	b.RecordFailure()
	if err := b.Allow(); err == nil {
		t.Fatal("expected open")
	}
}
