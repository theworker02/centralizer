// Package python implements the CPython adapter (Tier 1).
//
// Supported: detect, stdio process bridge, generated shim, CIR call, handles.
// Not claimed: native embedding, WASM, shared memory.
package python

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

// Adapter talks to CPython targets through a generated stdio shim.
type Adapter struct {
	Store *cache.Store
}

func (a *Adapter) Name() string { return "python" }

func (a *Adapter) Detect(_ context.Context, target adapter.Target) (adapter.Detection, error) {
	d := adapter.Detection{Adapter: a.Name(), Language: "Python"}
	info, err := os.Stat(target.Path)
	if err != nil {
		return d, czerr.ErrUnsupportedTarget
	}
	var evidence []string
	score := 0.0
	if !info.IsDir() {
		if strings.EqualFold(filepath.Ext(target.Path), ".py") {
			score = 0.9
			evidence = append(evidence, filepath.Base(target.Path))
		}
	} else {
		for _, name := range []string{"pyproject.toml", "requirements.txt", "Pipfile", "setup.py"} {
			if exists(filepath.Join(target.Path, name)) {
				score += 0.35
				evidence = append(evidence, name)
			}
		}
		if hasExt(target.Path, ".py") {
			score += 0.25
			evidence = append(evidence, ".py source files")
		}
	}
	if score > 1 {
		score = 0.99
	}
	if score < 0.2 {
		return d, czerr.ErrUnsupportedTarget
	}
	py := rt.Python()
	d.Confidence = score
	d.Evidence = evidence
	d.Runtime = "CPython"
	d.Arch = runtimeArch()
	if py.Available {
		d.Version = py.Version
		d.Evidence = append(d.Evidence, "installed interpreter")
		if d.Confidence < 0.97 {
			d.Confidence += 0.05
			if d.Confidence > 0.99 {
				d.Confidence = 0.99
			}
		}
	}
	return d, nil
}

func (a *Adapter) Capabilities(_ context.Context, _ adapter.Target) ([]capability.Capability, error) {
	caps := []capability.Capability{
		{Kind: capability.Stdio, Available: true, Confidence: 1, Detail: "generated Python shim"},
		{Kind: capability.TCP, Available: true, Confidence: 0.9, Detail: "localhost length-prefixed frames"},
		{Kind: capability.Process, Available: true, Confidence: 1},
		{Kind: capability.GeneratedShim, Available: true, Confidence: 1},
		{Kind: capability.Handles, Available: true, Confidence: 0.8},
		{Kind: capability.Streaming, Available: true, Confidence: 0.9, Detail: "generator/iterable STREAM_*"},
		{Kind: capability.CLI, Available: true, Confidence: 0.5},
	}
	if rt.Python().Available {
		caps = append(caps, capability.Capability{Kind: capability.RuntimeVersion, Available: true, Confidence: 1, Detail: rt.Python().Version})
	}
	return caps, nil
}

func (a *Adapter) Prepare(context.Context, adapter.Target) error { return nil }

func (a *Adapter) Connect(ctx context.Context, target adapter.Target, plan bridge.Plan) (bridge.Bridge, error) {
	py := rt.Python()
	if !py.Available {
		return nil, czerr.New(czerr.ErrRuntimeUnavailable, "python interpreter not found")
	}
	store := a.Store
	if store == nil {
		var err error
		store, err = cache.New("")
		if err != nil {
			return nil, err
		}
	}
	fp := discovery.Fingerprint(target.Path, nil, adapter.Detection{Adapter: a.Name(), Runtime: "CPython", Version: py.Version})
	shimPath, err := shim.Materialize(store, "python", fp, shim.PythonShim, "centralizer_shim.py")
	if err != nil {
		return nil, err
	}
	env := append(os.Environ(),
		"CENTRALIZER_TARGET="+target.Path,
		"CENTRALIZER_ENTRY="+target.Entry,
	)
	return shim.Connect(ctx, shim.StdioConfig{
		Argv: []string{py.Command, "-u", shimPath},
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
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && (d.Name() == ".venv" || d.Name() == "node_modules") {
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

func runtimeArch() string {
	osname, arch := rt.Host()
	return osname + "/" + arch
}
