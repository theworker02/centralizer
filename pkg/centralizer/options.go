package centralizer

import (
	"log/slog"
	"time"

	"github.com/theworker02/centralizer/internal/security"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/manifest"
)

// Option configures a Hub.
type Option func(*config)

type config struct {
	autoRecover bool
	tracing     bool
	logger      *slog.Logger
	policy      manifest.Policy
	manifest    *manifest.Manifest
	cacheDir    string
	timeout     time.Duration
	entry       string
	prefer      []string
	language    string
	adapters    []adapter.Adapter
	handleTTL   time.Duration
}

func defaultConfig() config {
	return config{
		autoRecover: true,
		policy:      security.DefaultPolicy(),
		timeout:     30 * time.Second,
	}
}

// WithAutoRecovery enables bounded bridge restart after recoverable failures.
func WithAutoRecovery(v bool) Option {
	return func(c *config) { c.autoRecover = v }
}

// WithTracing records an in-process span tree for explain/trace.
func WithTracing(v bool) Option {
	return func(c *config) { c.tracing = v }
}

// WithLogger replaces the structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.logger = l }
}

// WithPolicy installs a connection policy.
func WithPolicy(p manifest.Policy) Option {
	return func(c *config) { c.policy = p }
}

// WithManifest applies a parsed manifest (overrides + policy).
func WithManifest(m *manifest.Manifest) Option {
	return func(c *config) {
		c.manifest = m
		if m != nil {
			c.policy = m.Policy
		}
	}
}

// WithCacheDir overrides the generated-shim cache location.
func WithCacheDir(dir string) Option {
	return func(c *config) { c.cacheDir = dir }
}

// WithTimeout sets the default per-call timeout when the context has none.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// WithEntry selects a module/binary inside the target.
func WithEntry(entry string) Option {
	return func(c *config) { c.entry = entry }
}

// WithPrefer biases planner strategy names (native, stdio, unix_socket, …).
func WithPrefer(strategies ...string) Option {
	return func(c *config) { c.prefer = strategies }
}

// WithLanguage forces an adapter language (default: auto).
func WithLanguage(lang string) Option {
	return func(c *config) { c.language = lang }
}

// WithAdapter registers an additional adapter on this hub only.
func WithAdapter(a adapter.Adapter) Option {
	return func(c *config) { c.adapters = append(c.adapters, a) }
}

// WithHandleTTL expires locally tracked handles after d. Zero means no expiry.
func WithHandleTTL(d time.Duration) Option {
	return func(c *config) { c.handleTTL = d }
}
