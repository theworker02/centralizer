// Package transport abstracts how protocol frames move between Centralizer
// and a runtime shim.
package transport

import (
	"context"
	"io"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Kind names a transport implementation.
type Kind string

const (
	KindNative       Kind = "native"
	KindSharedMemory Kind = "shared_memory"
	KindStdio        Kind = "stdio"
	KindUnix         Kind = "unix_socket"
	KindPipe         Kind = "named_pipe"
	KindTCP          Kind = "tcp"
	KindHTTP         Kind = "http"
	KindWebSocket    Kind = "websocket"
	KindWASM         Kind = "wasm"
	KindMemory       Kind = "memory"
)

// Frame is a protocol message in transit.
type Frame = protocol.Message

// Transport is a bidirectional framed connection.
type Transport interface {
	Kind() Kind
	Open(ctx context.Context) error
	Send(ctx context.Context, frame Frame) error
	Receive(ctx context.Context) (Frame, error)
	Close() error
}

// Pair is an in-process transport used by tests and the native Go adapter.
func Pair() (Transport, Transport) {
	a, b := io.Pipe()
	c, d := io.Pipe()
	left := &pipeTransport{name: KindMemory, r: a, w: d}
	right := &pipeTransport{name: KindMemory, r: c, w: b}
	return left, right
}

type pipeTransport struct {
	name Kind
	r    io.ReadCloser
	w    io.WriteCloser
}

func (p *pipeTransport) Kind() Kind { return p.name }

func (p *pipeTransport) Open(context.Context) error { return nil }

func (p *pipeTransport) Send(ctx context.Context, frame Frame) error {
	if err := ctx.Err(); err != nil {
		return mapCtx(err)
	}
	return protocol.WriteFrame(p.w, frame, 0)
}

func (p *pipeTransport) Receive(ctx context.Context) (Frame, error) {
	type result struct {
		f   Frame
		err error
	}
	ch := make(chan result, 1)
	go func() {
		f, err := protocol.ReadFrame(p.r, 0)
		ch <- result{f, err}
	}()
	select {
	case <-ctx.Done():
		return Frame{}, mapCtx(ctx.Err())
	case r := <-ch:
		return r.f, r.err
	}
}

func (p *pipeTransport) Close() error {
	_ = p.r.Close()
	return p.w.Close()
}

func mapCtx(err error) error {
	if err == context.Canceled {
		return czerr.Wrap(czerr.ErrCancelled, "transport", err)
	}
	if err == context.DeadlineExceeded {
		return czerr.Wrap(czerr.ErrTimeout, "transport", err)
	}
	return err
}
