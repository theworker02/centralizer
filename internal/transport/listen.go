package transport

import (
	"context"
	"net"
	"time"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// NewConn wraps an already-accepted or already-dialed connection.
func NewConn(conn net.Conn, kind Kind) *Net {
	network := "tcp"
	if kind == KindUnix {
		network = "unix"
	}
	return &Net{
		Network: network,
		Address: conn.RemoteAddr().String(),
		kind:    kind,
		conn:    conn,
	}
}

// ListenLoopback binds 127.0.0.1:0.
func ListenLoopback() (net.Listener, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrTransportFailure, "listen loopback", err)
	}
	return ln, nil
}

// AcceptOne waits for a single connection or ctx expiry.
func AcceptOne(ctx context.Context, ln net.Listener) (net.Conn, error) {
	type result struct {
		c   net.Conn
		err error
	}
	ch := make(chan result, 1)
	go func() {
		c, err := ln.Accept()
		ch <- result{c, err}
	}()
	select {
	case <-ctx.Done():
		_ = ln.Close()
		return nil, czerr.Wrap(czerr.ErrTimeout, "accept", ctx.Err())
	case r := <-ch:
		if r.err != nil {
			return nil, czerr.Wrap(czerr.ErrTransportFailure, "accept", r.err)
		}
		return r.c, nil
	}
}

// DialLoopback connects to a localhost TCP address.
func DialLoopback(ctx context.Context, addr string) (net.Conn, error) {
	var d net.Dialer
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrTransportFailure, "dial "+addr, err)
	}
	return c, nil
}
