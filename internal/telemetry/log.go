// Package telemetry provides structured logs, in-process metrics, and
// lightweight traces. OpenTelemetry exporters can be attached through
// the Tracer and Meter interfaces without forcing a hard dependency.
package telemetry

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

type ctxKey struct{}

// Logger is the process-wide structured logger. Tests may replace it.
var Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

// SetLogger replaces the global logger.
func SetLogger(l *slog.Logger) {
	if l != nil {
		Logger = l
	}
}

// WithLogger stores a logger on the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns a contextual or global logger.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return Logger
}

// LevelFromFlags maps CLI verbosity onto slog levels.
func LevelFromFlags(verbose, quiet bool) slog.Level {
	if quiet {
		return slog.LevelError
	}
	if verbose {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// Configure installs a text or JSON handler on stderr.
func Configure(jsonOut bool, level slog.Level) {
	opts := &slog.HandlerOptions{Level: level}
	if jsonOut {
		Logger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
		return
	}
	Logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// Span is a lightweight in-process trace span.
type Span struct {
	Name     string         `json:"name"`
	Start    time.Time      `json:"start"`
	End      time.Time      `json:"end,omitempty"`
	Attrs    map[string]any `json:"attrs,omitempty"`
	Children []*Span        `json:"children,omitempty"`
	Err      string         `json:"error,omitempty"`
	mu       sync.Mutex
}

// Tracer records a tree of spans for explain/trace commands.
type Tracer struct {
	mu   sync.Mutex
	root *Span
	cur  []*Span
}

// NewTracer starts a root span.
func NewTracer(name string) *Tracer {
	s := &Span{Name: name, Start: time.Now().UTC(), Attrs: map[string]any{}}
	return &Tracer{root: s, cur: []*Span{s}}
}

// Start begins a child span.
func (t *Tracer) Start(name string) *Span {
	if t == nil {
		return &Span{Name: name, Start: time.Now().UTC()}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	s := &Span{Name: name, Start: time.Now().UTC(), Attrs: map[string]any{}}
	parent := t.root
	if len(t.cur) > 0 {
		parent = t.cur[len(t.cur)-1]
	}
	parent.Children = append(parent.Children, s)
	t.cur = append(t.cur, s)
	return s
}

// End finishes the current span.
func (t *Tracer) End(err error) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.cur) == 0 {
		return
	}
	s := t.cur[len(t.cur)-1]
	s.End = time.Now().UTC()
	if err != nil {
		s.Err = err.Error()
	}
	t.cur = t.cur[:len(t.cur)-1]
}

// Root returns the trace tree.
func (t *Tracer) Root() *Span {
	if t == nil {
		return nil
	}
	return t.root
}

// Finish closes the root span.
func (t *Tracer) Finish() {
	if t == nil || t.root == nil {
		return
	}
	t.root.End = time.Now().UTC()
}
