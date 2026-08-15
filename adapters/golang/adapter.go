// Package golang implements the Go adapter (Tier 1).
//
// Supported: detect, in-process native handlers, stdio to protocol binaries.
// Planned: WASM compilation of Go targets.
package golang

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/theworker02/centralizer/internal/session"
	"github.com/theworker02/centralizer/internal/shim"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Adapter connects Go targets, including in-process handlers.
type Adapter struct {
	mu       sync.RWMutex
	handlers map[string]*session.Handler
}

func (a *Adapter) Name() string { return "go" }

// Register installs an in-process handler under name.
func (a *Adapter) Register(h *session.Handler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.handlers == nil {
		a.handlers = map[string]*session.Handler{}
	}
	a.handlers[h.Name] = h
}

func (a *Adapter) Detect(_ context.Context, target adapter.Target) (adapter.Detection, error) {
	d := adapter.Detection{Adapter: a.Name(), Language: "Go", Runtime: "gc"}
	if strings.HasPrefix(target.Ref, "native:") || strings.HasPrefix(target.Path, "native:") {
		d.Confidence = 1
		d.Evidence = []string{"in-process handler"}
		return d, nil
	}
	info, err := os.Stat(target.Path)
	if err != nil {
		return d, czerr.ErrUnsupportedTarget
	}
	score := 0.0
	var evidence []string
	if !info.IsDir() {
		if strings.EqualFold(filepath.Ext(target.Path), ".go") {
			score = 0.7
			evidence = append(evidence, filepath.Base(target.Path))
		}
	} else if exists(filepath.Join(target.Path, "go.mod")) {
		score = 0.95
		evidence = append(evidence, "go.mod")
	} else if hasGo(target.Path) {
		score = 0.55
		evidence = append(evidence, ".go source files")
	}
	if score < 0.2 {
		return d, czerr.ErrUnsupportedTarget
	}
	d.Confidence = score
	d.Evidence = evidence
	return d, nil
}

func (a *Adapter) Capabilities(_ context.Context, target adapter.Target) ([]capability.Capability, error) {
	caps := []capability.Capability{
		{Kind: capability.Stdio, Available: true, Confidence: 0.8},
		{Kind: capability.Process, Available: true, Confidence: 0.8},
	}
	if a.lookup(target) != nil {
		caps = append(caps, capability.Capability{Kind: capability.InProcess, Available: true, Confidence: 1})
	}
	return caps, nil
}

func (a *Adapter) Prepare(context.Context, adapter.Target) error { return nil }

func (a *Adapter) Connect(ctx context.Context, target adapter.Target, plan bridge.Plan) (bridge.Bridge, error) {
	if h := a.lookup(target); h != nil {
		return session.NewNative(plan, h), nil
	}
	argv := []string{target.Path}
	dir := target.Path
	info, err := os.Stat(target.Path)
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrTargetNotFound, target.Path, err)
	}
	if info.IsDir() {
		argv = []string{"go", "run", "."}
		if target.Entry != "" {
			argv = []string{"go", "run", target.Entry}
		}
	}
	b, _, err := shim.ConnectStdio(ctx, shim.StdioConfig{Argv: argv, Dir: dir, Plan: plan})
	return b, err
}

func (a *Adapter) lookup(target adapter.Target) *session.Handler {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.handlers == nil {
		return nil
	}
	name := target.Entry
	if name == "" {
		name = strings.TrimPrefix(target.Ref, "native:")
	}
	if name == "" {
		name = filepath.Base(target.Path)
	}
	return a.handlers[name]
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func hasGo(root string) bool {
	found := false
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(d.Name(), ".go") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
