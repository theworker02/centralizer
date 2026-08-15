package session

import (
	"context"
	"sync"

	"github.com/theworker02/centralizer/pkg/cir"
)

type memStream struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan cir.Value
	mu     sync.Mutex
	err    error
	closed bool
}

func newStream(id string, parent context.Context) *memStream {
	ctx, cancel := context.WithCancel(parent)
	return &memStream{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan cir.Value, 32),
	}
}

func (s *memStream) ID() string { return s.id }

func (s *memStream) Values() <-chan cir.Value { return s.ch }

func (s *memStream) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *memStream) Close() error {
	s.closeWith(nil)
	return nil
}

func (s *memStream) emit(v cir.Value) {
	select {
	case s.ch <- v:
	case <-s.ctx.Done():
	}
}

func (s *memStream) closeWith(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.err = err
	s.cancel()
	close(s.ch)
}
