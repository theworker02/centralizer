package failure

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/internal/supervisor"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/schema"
)

func TestCorruptNDJSON(t *testing.T) {
	_, err := protocol.ReadNDJSON(bufio.NewReader(bytes.NewReader([]byte("{{{notjson\n"))))
	if err == nil {
		t.Fatal("expected frame error")
	}
}

func TestPartialFrame(t *testing.T) {
	_, err := protocol.ReadFrame(bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x10, '{'}), 1024)
	if err == nil {
		t.Fatal("expected short read")
	}
}

func TestHugeFrame(t *testing.T) {
	_, err := protocol.ReadFrame(bytes.NewReader([]byte{0x7f, 0xff, 0xff, 0xff}), 1024)
	if err == nil {
		t.Fatal("expected too large")
	}
}

type crashBridge struct {
	n int
}

func (c *crashBridge) ID() string        { return "crash" }
func (c *crashBridge) Plan() bridge.Plan { return bridge.Plan{Transport: "stdio"} }
func (c *crashBridge) Describe(context.Context) (*schema.Schema, error) {
	return &schema.Schema{Service: "x", Inferred: true}, nil
}
func (c *crashBridge) Call(ctx context.Context, fn string, args map[string]cir.Value) (cir.Value, error) {
	return c.Invoke(ctx, bridge.Invocation{Function: fn, Args: args})
}
func (c *crashBridge) Invoke(context.Context, bridge.Invocation) (cir.Value, error) {
	c.n++
	if c.n == 1 {
		return cir.Value{}, czerr.New(czerr.ErrBridgeFailed, "process crash")
	}
	return cir.Int(1), nil
}
func (c *crashBridge) New(context.Context, string, map[string]cir.Value) (cir.Value, error) {
	return cir.Value{}, czerr.ErrNotImplemented
}
func (c *crashBridge) Get(context.Context, string, string) (cir.Value, error) {
	return cir.Value{}, czerr.ErrNotImplemented
}
func (c *crashBridge) Set(context.Context, string, string, cir.Value) error {
	return czerr.ErrNotImplemented
}
func (c *crashBridge) Release(context.Context, string) error { return nil }
func (c *crashBridge) Stream(context.Context, string, map[string]cir.Value) (bridge.Stream, error) {
	return nil, czerr.ErrNotImplemented
}
func (c *crashBridge) Subscribe(context.Context, string) (bridge.Stream, error) {
	return nil, czerr.ErrNotImplemented
}
func (c *crashBridge) Ping(context.Context) error { return nil }
func (c *crashBridge) Close(context.Context) error {
	return nil
}

func TestSupervisorRecoversFromCrash(t *testing.T) {
	inner := &crashBridge{}
	s := supervisor.New("x", inner, bridge.Plan{Transport: "stdio"}, func(context.Context, bridge.Plan) (bridge.Bridge, error) {
		return &crashBridge{n: 1}, nil
	}, supervisor.Config{AutoRecover: true, MaxRestarts: 3, InitialDelay: time.Millisecond})
	v, err := s.Call(context.Background(), "f", nil)
	if err != nil {
		t.Fatal(err)
	}
	n, err := v.Int()
	if err != nil || n != 1 {
		t.Fatalf("%v %v", n, err)
	}
}

func TestInvalidHandle(t *testing.T) {
	err := czerr.New(czerr.ErrHandleInvalid, "nope")
	if err == nil {
		t.Fatal("expected error")
	}
}

var _ io.Reader = bytes.NewReader(nil)
