// Package node implements the Node.js / JavaScript / TypeScript adapter (Tier 1).
//
// Supported: detect, stdio process bridge, generated shim, CIR call.
// TypeScript is detected via package.json / .ts files and executed by Node
// only when the project already compiles to JS or uses a JS entry.
package node

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/theworker02/centralizer/internal/cache"
	"github.com/theworker02/centralizer/internal/discovery"
	rt "github.com/theworker02/centralizer/internal/runtime"
	"github.com/theworker02/centralizer/internal/shim"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Adapter talks to Node targets through a generated stdio shim.
type Adapter struct {
	Store *cache.Store
}

func (a *Adapter) Name() string { return "node" }

func (a *Adapter) Detect(_ context.Context, target adapter.Target) (adapter.Detection, error) {
	d := adapter.Detection{Adapter: a.Name(), Language: "JavaScript", Runtime: "Node.js"}
	info, err := os.Stat(target.Path)
	if err != nil {
		return d, czerr.ErrUnsupportedTarget
	}
	var evidence []string
	score := 0.0
	if !info.IsDir() {
		ext := strings.ToLower(filepath.Ext(target.Path))
		if ext == ".js" || ext == ".mjs" || ext == ".cjs" {
			score = 0.88
			evidence = append(evidence, filepath.Base(target.Path))
		} else if ext == ".ts" {
			d.Language = "TypeScript"
			score = 0.7
			evidence = append(evidence, filepath.Base(target.Path))
		}
	} else {
		if exists(filepath.Join(target.Path, "package.json")) {
			score += 0.55
			evidence = append(evidence, "package.json")
		}
		if exists(filepath.Join(target.Path, "deno.json")) {
			score += 0.2
			evidence = append(evidence, "deno.json")
		}
		if hasExt(target.Path, ".ts") {
			d.Language = "TypeScript"
			score += 0.15
			evidence = append(evidence, ".ts source files")
		} else if hasExt(target.Path, ".js") {
			score += 0.2
			evidence = append(evidence, ".js source files")
		}
	}
	if score > 0.99 {
		score = 0.99
	}
	if score < 0.2 {
		return d, czerr.ErrUnsupportedTarget
	}
	n := rt.Node()
	d.Confidence = score
	d.Evidence = evidence
	if n.Available {
		d.Version = n.Version
		d.Evidence = append(d.Evidence, "installed interpreter")
	}
	return d, nil
}

func (a *Adapter) Capabilities(context.Context, adapter.Target) ([]capability.Capability, error) {
	return []capability.Capability{
		{Kind: capability.Stdio, Available: true, Confidence: 1, Detail: "generated Node shim"},
		{Kind: capability.TCP, Available: true, Confidence: 0.9, Detail: "localhost length-prefixed frames"},
		{Kind: capability.Process, Available: true, Confidence: 1},
		{Kind: capability.GeneratedShim, Available: true, Confidence: 1},
		{Kind: capability.Streaming, Available: true, Confidence: 0.9, Detail: "iterable/async iterable STREAM_*"},
	}, nil
}

func (a *Adapter) Prepare(context.Context, adapter.Target) error { return nil }

func (a *Adapter) Connect(ctx context.Context, target adapter.Target, plan bridge.Plan) (bridge.Bridge, error) {
	n := rt.Node()
	if !n.Available {
		return nil, czerr.New(czerr.ErrRuntimeUnavailable, "node interpreter not found")
	}
	store := a.Store
	if store == nil {
		var err error
		store, err = cache.New("")
		if err != nil {
			return nil, err
		}
	}
	fp := discovery.Fingerprint(target.Path, nil, adapter.Detection{Adapter: a.Name(), Runtime: "Node.js", Version: n.Version})
	shimPath, err := shim.Materialize(store, "node", fp, shim.NodeShim, "centralizer_shim.js")
	if err != nil {
		return nil, err
	}
	env := append(os.Environ(),
		"CENTRALIZER_TARGET="+target.Path,
		"CENTRALIZER_ENTRY="+target.Entry,
	)
	return shim.Connect(ctx, shim.StdioConfig{
		Argv: []string{n.Command, shimPath},
		Dir:  target.Path,
		Env:  env,
		Plan: plan,
	})
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func hasExt(root, ext string) bool {
	found := false
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && (d.Name() == "node_modules" || d.Name() == ".git") {
				return filepath.SkipDir
			}
			return err
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ext) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
