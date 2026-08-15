package transport

import (
	"context"
	"testing"
	"time"

	"github.com/theworker02/centralizer/internal/protocol"
)

func TestMemoryPair(t *testing.T) {
	a, b := Pair()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := a.Open(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()
	defer func() { _ = b.Close() }()

	m, err := protocol.NewMessage("1", protocol.TypeHeartbeat, nil)
	if err != nil {
		t.Fatal(err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Send(ctx, m)
	}()
	got, err := b.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != protocol.TypeHeartbeat {
		t.Fatalf("%+v", got)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}

func TestSharedMemoryDisabled(t *testing.T) {
	s := &SharedMemory{}
	if err := s.Open(context.Background()); err == nil {
		t.Fatal("expected experimental error")
	}
}

func TestTCPLoopbackRoundTrip(t *testing.T) {
	ln, err := ListenLoopback()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	addr := ln.Addr().String()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errCh := make(chan error, 2)
	go func() {
		c, acceptErr := AcceptOne(ctx, ln)
		if acceptErr != nil {
			errCh <- acceptErr
			return
		}
		srv := NewConn(c, KindTCP)
		msg, recvErr := srv.Receive(ctx)
		if recvErr != nil {
			errCh <- recvErr
			return
		}
		reply, _ := protocol.NewMessage(msg.ID, protocol.TypeOK, nil)
		errCh <- srv.Send(ctx, reply)
		_ = srv.Close()
	}()

	cli := NewTCP(addr)
	if err = cli.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cli.Close() }()
	hello, err := protocol.NewMessage("1", protocol.TypeHello, protocol.HelloPayload{Protocol: protocol.Version, Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if err = cli.Send(ctx, hello); err != nil {
		t.Fatal(err)
	}
	got, err := cli.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != protocol.TypeOK {
		t.Fatalf("%+v", got)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}

func TestNamedPipeOpenDoesNotPanic(t *testing.T) {
	p := NewNamedPipe("centralizer-test-missing")
	if p.Kind() != KindPipe {
		t.Fatalf("kind=%s", p.Kind())
	}
	err := p.Open(context.Background())
	if err == nil {
		_ = p.Close()
		t.Fatal("expected missing pipe to fail")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}
