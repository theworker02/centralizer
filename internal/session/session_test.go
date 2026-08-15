package session

import (
	"context"
	"testing"
	"time"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/internal/transport"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/cir"
)

func TestSessionRoundTrip(t *testing.T) {
	a, b := transport.Pair()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		if err := b.Open(ctx); err != nil {
			return
		}
		hello, err := b.Receive(ctx)
		if err != nil {
			return
		}
		reply, _ := protocol.NewMessage(hello.ID, protocol.TypeHello, protocol.HelloPayload{
			Protocol: protocol.Version,
			Name:     "test",
		})
		_ = b.Send(ctx, reply)
		for {
			msg, err := b.Receive(ctx)
			if err != nil {
				return
			}
			if msg.Type == protocol.TypeShutdown {
				ok, _ := protocol.NewMessage(msg.ID, protocol.TypeOK, nil)
				_ = b.Send(ctx, ok)
				_ = b.Close()
				return
			}
			if msg.Type == protocol.TypeCall {
				res, _ := protocol.NewMessage(msg.ID, protocol.TypeResult, protocol.ResultPayload{
					Value: cir.Int(7).ToWire(),
				})
				_ = b.Send(ctx, res)
			}
		}
	}()

	sess, err := Open(ctx, a, bridge.Plan{Transport: "memory"})
	if err != nil {
		t.Fatal(err)
	}
	v, err := sess.Call(ctx, "f", nil)
	if err != nil {
		t.Fatal(err)
	}
	n, err := v.Int()
	if err != nil || n != 7 {
		t.Fatalf("%v %v", n, err)
	}
	_ = sess.Close(ctx)
}
