package transport

import (
	"context"
	"io"
	"sync"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// NamedPipe is a Windows local IPC transport. Open is implemented per OS.
// On Windows it can dial an existing \\.\pipe\<name> pipe. Creating a
// server pipe and a full bidirectional session remains experimental.
type NamedPipe struct {
	Name string

	mu   sync.Mutex
	conn io.ReadWriteCloser
}

// NewNamedPipe constructs a named-pipe transport. name may be a bare
// identifier or a full \\.\pipe\ path.
func NewNamedPipe(name string) *NamedPipe {
	return &NamedPipe{Name: name}
}

func (p *NamedPipe) Kind() Kind { return KindPipe }

func (p *NamedPipe) Send(ctx context.Context, frame Frame) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		return czerr.New(czerr.ErrTransportFailure, "named pipe not open")
	}
	if err := ctx.Err(); err != nil {
		return mapCtx(err)
	}
	return protocol.WriteFrame(p.conn, frame, 0)
}

func (p *NamedPipe) Receive(ctx context.Context) (Frame, error) {
	p.mu.Lock()
	c := p.conn
	p.mu.Unlock()
	if c == nil {
		return Frame{}, czerr.New(czerr.ErrTransportFailure, "named pipe not open")
	}
	type result struct {
		f   Frame
		err error
	}
	ch := make(chan result, 1)
	go func() {
		f, err := protocol.ReadFrame(c, 0)
		ch <- result{f, err}
	}()
	select {
	case <-ctx.Done():
		return Frame{}, mapCtx(ctx.Err())
	case r := <-ch:
		return r.f, r.err
	}
}

func (p *NamedPipe) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		return nil
	}
	err := p.conn.Close()
	p.conn = nil
	return err
}
