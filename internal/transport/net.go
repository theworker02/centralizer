package transport

import (
	"context"
	"net"
	"sync"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Net is a length-prefixed JSON connection over TCP, Unix, or a similar
// stream-oriented net.Conn.
type Net struct {
	Network string
	Address string
	kind    Kind
	dial    func(ctx context.Context) (net.Conn, error)

	mu   sync.Mutex
	conn net.Conn
}

// NewTCP dials a TCP address.
func NewTCP(addr string) *Net {
	return &Net{
		Network: "tcp",
		Address: addr,
		kind:    KindTCP,
		dial: func(ctx context.Context) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
	}
}

// NewUnix dials a Unix domain socket. On Windows the planner should not
// select this transport.
func NewUnix(path string) *Net {
	return &Net{
		Network: "unix",
		Address: path,
		kind:    KindUnix,
		dial: func(ctx context.Context) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", path)
		},
	}
}

func (n *Net) Kind() Kind { return n.kind }

func (n *Net) Open(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn != nil {
		return nil
	}
	if n.dial == nil {
		return czerr.New(czerr.ErrTransportFailure, "no dialer")
	}
	c, err := n.dial(ctx)
	if err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, n.Network, err)
	}
	n.conn = c
	return nil
}

func (n *Net) Send(ctx context.Context, frame Frame) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn == nil {
		return czerr.New(czerr.ErrTransportFailure, "not open")
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = n.conn.SetWriteDeadline(deadline)
	}
	return protocol.WriteFrame(n.conn, frame, 0)
}

func (n *Net) Receive(ctx context.Context) (Frame, error) {
	n.mu.Lock()
	c := n.conn
	n.mu.Unlock()
	if c == nil {
		return Frame{}, czerr.New(czerr.ErrTransportFailure, "not open")
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.SetReadDeadline(deadline)
	}
	return protocol.ReadFrame(c, 0)
}

func (n *Net) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn == nil {
		return nil
	}
	err := n.conn.Close()
	n.conn = nil
	return err
}
