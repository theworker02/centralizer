package supervisor

import (
	"sync"
	"time"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// BreakerState is the circuit-breaker state machine.
type BreakerState string

const (
	BreakerClosed   BreakerState = "closed"
	BreakerOpen     BreakerState = "open"
	BreakerHalfOpen BreakerState = "half_open"
)

// Breaker trips after consecutive failures and cools down before a probe.
type Breaker struct {
	FailureThreshold int
	SuccessThreshold int
	OpenFor          time.Duration

	mu        sync.Mutex
	state     BreakerState
	failures  int
	successes int
	openedAt  time.Time
}

// NewBreaker returns a closed breaker with conservative defaults.
func NewBreaker() *Breaker {
	return &Breaker{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		OpenFor:          5 * time.Second,
		state:            BreakerClosed,
	}
}

// Allow reports whether a call may proceed.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case BreakerOpen:
		if time.Since(b.openedAt) >= b.OpenFor {
			b.state = BreakerHalfOpen
			b.successes = 0
			return nil
		}
		return czerr.ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess records a successful call.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	if b.state == BreakerHalfOpen {
		b.successes++
		if b.successes >= b.SuccessThreshold {
			b.state = BreakerClosed
			b.successes = 0
		}
	}
}

// RecordFailure records a failed call.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.state == BreakerHalfOpen || b.failures >= b.FailureThreshold {
		b.state = BreakerOpen
		b.openedAt = time.Now()
		b.successes = 0
	}
}

// State returns the current breaker state.
func (b *Breaker) State() BreakerState {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == BreakerOpen && time.Since(b.openedAt) >= b.OpenFor {
		return BreakerHalfOpen
	}
	return b.state
}
