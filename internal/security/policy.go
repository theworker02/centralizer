package security

import (
	"strings"

	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/manifest"
)

// Engine evaluates connection-time policy constraints.
type Engine struct {
	Policy manifest.Policy
}

// DefaultPolicy is the conservative built-in policy.
func DefaultPolicy() manifest.Policy {
	return manifest.Policy{
		Recovery:  "automatic",
		Isolation: "process",
		Network:   "localhost_only",
	}
}

// AllowRuntime checks an adapter/runtime name against the allow-list.
func (e Engine) AllowRuntime(name string) error {
	if len(e.Policy.AllowedRuntimes) == 0 {
		return nil
	}
	for _, a := range e.Policy.AllowedRuntimes {
		if strings.EqualFold(a, name) {
			return nil
		}
	}
	return czerr.New(czerr.ErrPolicyDenied, "runtime "+name+" is not allowed")
}

// AllowTransport checks a transport/strategy name.
func (e Engine) AllowTransport(name string) error {
	if len(e.Policy.AllowedTransports) == 0 {
		return nil
	}
	for _, a := range e.Policy.AllowedTransports {
		if strings.EqualFold(a, name) {
			return nil
		}
	}
	return czerr.New(czerr.ErrPolicyDenied, "transport "+name+" is not allowed")
}

// AllowNative reports whether in-process native execution is permitted.
func (e Engine) AllowNative() error {
	if e.Policy.NativeExecutionAllowed() {
		return nil
	}
	return czerr.New(czerr.ErrPolicyDenied, "native execution disabled")
}

// AllowSubprocess reports whether child processes may be started.
func (e Engine) AllowSubprocess() error {
	if e.Policy.SubprocessesAllowed() {
		return nil
	}
	return czerr.New(czerr.ErrPolicyDenied, "subprocess execution disabled")
}

// AllowGeneratedCode reports whether shim generation is permitted.
func (e Engine) AllowGeneratedCode() error {
	if e.Policy.GeneratedCodeAllowed() {
		return nil
	}
	return czerr.New(czerr.ErrPolicyDenied, "generated code disabled")
}

// NetworkMode returns the configured network constraint.
func (e Engine) NetworkMode() string {
	if e.Policy.Network == "" {
		return "localhost_only"
	}
	return e.Policy.Network
}

// MaxRestarts returns the restart budget, defaulting to 5.
func (e Engine) MaxRestarts() int {
	if e.Policy.MaxRestarts <= 0 {
		return 5
	}
	return e.Policy.MaxRestarts
}

// FilterEnv returns a copy of env with obviously dangerous variables removed.
func FilterEnv(env []string) []string {
	deny := map[string]bool{
		"LD_PRELOAD":            true,
		"LD_LIBRARY_PATH":       true,
		"DYLD_INSERT_LIBRARIES": true,
		"DYLD_LIBRARY_PATH":     true,
		"PYTHONINSPECT":         true,
		"NODE_OPTIONS":          true,
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		key, _, _ := strings.Cut(e, "=")
		if deny[key] {
			continue
		}
		out = append(out, e)
	}
	return out
}
