package transport

import (
	"context"
	"sync"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// SharedMemory is an experimental high-performance transport for large
// binary payloads. The v0.1 implementation defines the ownership model
// and rejects use unless explicitly enabled.
//
// Regions carry: id, size, owner, generation, and a closed flag.
// Crash recovery invalidates regions whose owner process has disappeared.
type SharedMemory struct {
	Enabled bool
	mu      sync.Mutex
	regions map[string]*Region
}

// Region is a tracked shared-memory mapping.
type Region struct {
	ID         string
	Size       int
	Owner      string
	Generation uint64
	Closed     bool
}

func (s *SharedMemory) Kind() Kind { return KindSharedMemory }

func (s *SharedMemory) Open(context.Context) error {
	if !s.Enabled {
		return czerr.Wrap(czerr.ErrExperimental, "shared memory", czerr.ErrNotImplemented)
	}
	s.mu.Lock()
	if s.regions == nil {
		s.regions = map[string]*Region{}
	}
	s.mu.Unlock()
	return nil
}

func (s *SharedMemory) Send(context.Context, Frame) error {
	return czerr.Wrap(czerr.ErrExperimental, "shared memory send", czerr.ErrNotImplemented)
}

func (s *SharedMemory) Receive(context.Context) (Frame, error) {
	return Frame{}, czerr.Wrap(czerr.ErrExperimental, "shared memory receive", czerr.ErrNotImplemented)
}

func (s *SharedMemory) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.regions {
		r.Closed = true
	}
	return nil
}

// Alloc records a region. The actual OS mapping is not created in v0.1.
func (s *SharedMemory) Alloc(id, owner string, size int) (*Region, error) {
	if size <= 0 {
		return nil, czerr.New(czerr.ErrInvalidArgument, "invalid region size")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.regions == nil {
		s.regions = map[string]*Region{}
	}
	r := &Region{ID: id, Size: size, Owner: owner, Generation: 1}
	s.regions[id] = r
	return r, nil
}
