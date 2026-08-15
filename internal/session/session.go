// Package session implements a protocol client over a Transport.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/internal/transport"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Session is a live protocol conversation.
type Session struct {
	id   string
	plan bridge.Plan
	tr   transport.Transport

	mu      sync.Mutex
	pending map[string]chan protocol.Message
	next    atomic.Uint64
	closed  atomic.Bool
	schema  *schema.Schema
	recvErr error
}

// Open performs HELLO and starts the receive loop.
func Open(ctx context.Context, tr transport.Transport, plan bridge.Plan) (*Session, error) {
	if err := tr.Open(ctx); err != nil {
		return nil, err
	}
	s := &Session{
		id:      fmt.Sprintf("br-%d", time.Now().UnixNano()),
		plan:    plan,
		tr:      tr,
		pending: map[string]chan protocol.Message{},
	}
	hello, err := protocol.NewMessage(s.corr(), protocol.TypeHello, protocol.HelloPayload{
		Protocol: protocol.Version,
		Name:     protocol.Name,
		Features: []string{"call", "stream", "handles", "describe"},
	})
	if err != nil {
		_ = tr.Close()
		return nil, err
	}
	if err = tr.Send(ctx, hello); err != nil {
		_ = tr.Close()
		return nil, err
	}
	reply, err := tr.Receive(ctx)
	if err != nil {
		_ = tr.Close()
		return nil, err
	}
	if reply.Type == protocol.TypeError {
		_ = tr.Close()
		return nil, decodeError(reply)
	}
	if reply.Type != protocol.TypeHello && reply.Type != protocol.TypeOK && reply.Type != protocol.TypeCapabilities {
		_ = tr.Close()
		return nil, czerr.New(czerr.ErrProtocolMismatch, "expected HELLO, got "+string(reply.Type))
	}
	if reply.Type == protocol.TypeHello {
		var hp protocol.HelloPayload
		if err := protocol.DecodePayload(reply, &hp); err == nil && hp.Protocol != "" {
			if err := protocol.Compatible(hp.Protocol); err != nil {
				_ = tr.Close()
				return nil, err
			}
		}
	}
	go s.loop()
	return s, nil
}

func (s *Session) loop() {
	for !s.closed.Load() {
		msg, err := s.tr.Receive(context.Background())
		if err != nil {
			s.mu.Lock()
			s.recvErr = err
			for _, ch := range s.pending {
				select {
				case ch <- protocol.Message{Type: protocol.TypeError}:
				default:
				}
				close(ch)
			}
			s.pending = map[string]chan protocol.Message{}
			s.mu.Unlock()
			return
		}
		if msg.Type == protocol.TypeHeartbeat {
			continue
		}
		s.mu.Lock()
		ch, ok := s.pending[msg.ID]
		if !ok && (msg.Type == protocol.TypeStreamData || msg.Type == protocol.TypeStreamClose) {
			var peek struct {
				Stream string `json:"stream"`
			}
			_ = json.Unmarshal(msg.Payload, &peek)
			if peek.Stream != "" {
				ch, ok = s.pending[peek.Stream]
			}
		}
		s.mu.Unlock()
		if ok {
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

func (s *Session) corr() string {
	return fmt.Sprintf("%s-%d", s.id, s.next.Add(1))
}

func (s *Session) roundTrip(ctx context.Context, typ protocol.Type, payload any) (protocol.Message, error) {
	if s.closed.Load() {
		return protocol.Message{}, czerr.ErrClosed
	}
	id := s.corr()
	msg, err := protocol.NewMessage(id, typ, payload)
	if err != nil {
		return protocol.Message{}, err
	}
	ch := make(chan protocol.Message, 1)
	s.mu.Lock()
	if s.recvErr != nil {
		err := s.recvErr
		s.mu.Unlock()
		return protocol.Message{}, czerr.Wrap(czerr.ErrBridgeFailed, "receive loop", err)
	}
	s.pending[id] = ch
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
	}()
	if err := s.tr.Send(ctx, msg); err != nil {
		return protocol.Message{}, err
	}
	select {
	case <-ctx.Done():
		cancel, _ := protocol.NewMessage(id, protocol.TypeCancel, nil)
		_ = s.tr.Send(context.Background(), cancel)
		if ctx.Err() == context.DeadlineExceeded {
			return protocol.Message{}, czerr.Wrap(czerr.ErrTimeout, string(typ), ctx.Err())
		}
		return protocol.Message{}, czerr.Wrap(czerr.ErrCancelled, string(typ), ctx.Err())
	case reply, ok := <-ch:
		if !ok {
			s.mu.Lock()
			err := s.recvErr
			s.mu.Unlock()
			if err == nil {
				err = czerr.ErrBridgeFailed
			}
			return protocol.Message{}, czerr.Wrap(czerr.ErrBridgeFailed, "disconnected", err)
		}
		if reply.Type == protocol.TypeError {
			return protocol.Message{}, decodeError(reply)
		}
		return reply, nil
	}
}

func decodeError(m protocol.Message) error {
	var p protocol.ErrorPayload
	if err := protocol.DecodePayload(m, &p); err != nil {
		return czerr.New(czerr.ErrBridgeFailed, "remote error")
	}
	kind := czerr.ErrBridgeFailed
	switch p.Code {
	case "schema":
		kind = czerr.ErrSchemaMismatch
	case "conversion":
		kind = czerr.ErrConversion
	case "handle":
		kind = czerr.ErrHandleInvalid
	case "timeout":
		kind = czerr.ErrTimeout
	case "cancel":
		kind = czerr.ErrCancelled
	}
	return czerr.New(kind, p.Message)
}

func (s *Session) ID() string { return s.id }

func (s *Session) Plan() bridge.Plan { return s.plan }

func (s *Session) Describe(ctx context.Context) (*schema.Schema, error) {
	if s.schema != nil {
		return s.schema, nil
	}
	reply, err := s.roundTrip(ctx, protocol.TypeDescribe, protocol.DescribePayload{})
	if err != nil {
		return nil, err
	}
	var p protocol.DescribePayload
	if err = protocol.DecodePayload(reply, &p); err != nil {
		return nil, err
	}
	if p.Schema == "" {
		return &schema.Schema{Service: s.plan.Language, Inferred: true}, nil
	}
	sc, err := schema.ParseYAML([]byte(p.Schema))
	if err != nil {
		return nil, err
	}
	s.schema = sc
	return sc, nil
}

func (s *Session) SetSchema(sc *schema.Schema) { s.schema = sc }

func (s *Session) Call(ctx context.Context, fn string, args map[string]cir.Value) (cir.Value, error) {
	return s.Invoke(ctx, bridge.Invocation{Function: fn, Args: args})
}

func (s *Session) Invoke(ctx context.Context, inv bridge.Invocation) (cir.Value, error) {
	reply, err := s.roundTrip(ctx, protocol.TypeCall, protocol.CallPayload{
		Function: inv.Function,
		Handle:   inv.Handle,
		Method:   inv.Method,
		Args:     protocol.EncodeArgs(inv.Args),
	})
	if err != nil {
		return cir.Value{}, err
	}
	return decodeResult(reply)
}

func (s *Session) New(ctx context.Context, typeName string, args map[string]cir.Value) (cir.Value, error) {
	reply, err := s.roundTrip(ctx, protocol.TypeHandleCreate, protocol.HandlePayload{
		TypeName: typeName,
		Args:     protocol.EncodeArgs(args),
	})
	if err != nil {
		return cir.Value{}, err
	}
	return decodeResult(reply)
}

func (s *Session) Get(ctx context.Context, handle, property string) (cir.Value, error) {
	reply, err := s.roundTrip(ctx, protocol.TypeGet, protocol.HandlePayload{
		Handle:   handle,
		Property: property,
	})
	if err != nil {
		return cir.Value{}, err
	}
	return decodeResult(reply)
}

func (s *Session) Set(ctx context.Context, handle, property string, value cir.Value) error {
	w := value.ToWire()
	_, err := s.roundTrip(ctx, protocol.TypeSet, protocol.HandlePayload{
		Handle:   handle,
		Property: property,
		Value:    &w,
	})
	return err
}

func (s *Session) Release(ctx context.Context, handle string) error {
	_, err := s.roundTrip(ctx, protocol.TypeHandleRelease, protocol.HandlePayload{Handle: handle})
	return err
}

func (s *Session) Stream(ctx context.Context, name string, args map[string]cir.Value) (bridge.Stream, error) {
	sid := fmt.Sprintf("st-%d", s.next.Add(1))
	ch := make(chan protocol.Message, 32)
	s.mu.Lock()
	s.pending[sid] = ch
	s.mu.Unlock()
	reply, err := s.roundTrip(ctx, protocol.TypeStreamOpen, protocol.StreamOpenPayload{
		Name:   name,
		Args:   protocol.EncodeArgs(args),
		Stream: sid,
	})
	if err != nil {
		s.mu.Lock()
		delete(s.pending, sid)
		s.mu.Unlock()
		return nil, err
	}
	if reply.Type != protocol.TypeStreamOpen && reply.Type != protocol.TypeOK && reply.Type != protocol.TypeResult {
		s.mu.Lock()
		delete(s.pending, sid)
		s.mu.Unlock()
		return nil, czerr.New(czerr.ErrProtocolMismatch, "unexpected stream reply")
	}
	st := newStream(sid, ctx)
	go s.pumpStreamFrom(st, ch)
	return st, nil
}

func (s *Session) Subscribe(ctx context.Context, event string) (bridge.Stream, error) {
	return s.Stream(ctx, event, nil)
}

func (s *Session) pumpStreamFrom(st *memStream, ch <-chan protocol.Message) {
	defer func() {
		s.mu.Lock()
		delete(s.pending, st.id)
		s.mu.Unlock()
	}()
	for !s.closed.Load() {
		select {
		case <-st.ctx.Done():
			st.closeWith(st.ctx.Err())
			return
		case msg, ok := <-ch:
			if !ok {
				st.closeWith(czerr.ErrStreamClosed)
				return
			}
			switch msg.Type {
			case protocol.TypeStreamData:
				var p protocol.StreamDataPayload
				if err := protocol.DecodePayload(msg, &p); err != nil {
					st.closeWith(err)
					return
				}
				v, err := cir.FromWire(p.Value)
				if err != nil {
					st.closeWith(err)
					return
				}
				st.emit(v)
			case protocol.TypeStreamClose:
				var p protocol.StreamClosePayload
				_ = protocol.DecodePayload(msg, &p)
				if p.Error != "" {
					st.closeWith(czerr.New(czerr.ErrStreamClosed, p.Error))
					return
				}
				st.closeWith(nil)
				return
			case protocol.TypeError:
				st.closeWith(decodeError(msg))
				return
			}
		}
	}
}

func (s *Session) Ping(ctx context.Context) error {
	_, err := s.roundTrip(ctx, protocol.TypeHeartbeat, nil)
	return err
}

func (s *Session) Close(ctx context.Context) error {
	if s.closed.Swap(true) {
		return nil
	}
	msg, err := protocol.NewMessage(s.corr(), protocol.TypeShutdown, nil)
	if err == nil {
		_ = s.tr.Send(ctx, msg)
	}
	return s.tr.Close()
}

func decodeResult(m protocol.Message) (cir.Value, error) {
	if m.Type == protocol.TypeOK {
		return cir.Null(), nil
	}
	var p protocol.ResultPayload
	if err := protocol.DecodePayload(m, &p); err != nil {
		return cir.Value{}, err
	}
	return cir.FromWire(p.Value)
}
