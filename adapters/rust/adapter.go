// Package rust implements the Rust adapter (Tier 1).
//
// Supported: detect, stdio to a protocol-speaking binary (`cargo run` or
// an existing executable). Native ABI and WASM are detected but not
// connected in v0.1.
package rust

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	rt "github.com/theworker02/centralizer/internal/runtime"
	"github.com/theworker02/centralizer/internal/shim"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Adapter connects Rust targets that speak Centralizer Protocol on stdio.
type Adapter struct{}

func (a *Adapter) Name() string { return "rust" }

func (a *Adapter) Detect(_ context.Context, target adapter.Target) (adapter.Detection, error) {
	d := adapter.Detection{Adapter: a.Name(), Language: "Rust", Runtime: "rustc"}
	info, err := os.Stat(target.Path)
	if err != nil {
		return d, czerr.ErrUnsupportedTarget
	}
	score := 0.0
	var evidence []string
	if !info.IsDir() {
		if strings.EqualFold(filepath.Ext(target.Path), ".rs") {
			score = 0.6
			evidence = append(evidence, filepath.Base(target.Path))
		}
	} else if exists(filepath.Join(target.Path, "Cargo.toml")) {
		score = 0.96
		evidence = append(evidence, "Cargo.toml")
	}
	if score < 0.2 {
		return d, czerr.ErrUnsupportedTarget
	}
	if c := rt.Cargo(); c.Available {
		d.Version = c.Version
		evidence = append(evidence, "installed cargo")
	}
	d.Confidence = score
	d.Evidence = evidence
	return d, nil
}

func (a *Adapter) Capabilities(_ context.Context, target adapter.Target) ([]capability.Capability, error) {
	caps := []capability.Capability{
		{Kind: capability.Stdio, Available: true, Confidence: 1, Detail: "protocol-speaking Rust binary"},
		{Kind: capability.Process, Available: true, Confidence: 1},
		{Kind: capability.CLI, Available: true, Confidence: 0.6},
	}
	if exists(filepath.Join(target.Path, "Cargo.toml")) {
		caps = append(caps, capability.Capability{Kind: capability.WASM, Available: false, Confidence: 0.4, Detail: "wasm32 target planned"})
	}
	return caps, nil
}

func (a *Adapter) Prepare(context.Context, adapter.Target) error { return nil }

func (a *Adapter) Connect(ctx context.Context, target adapter.Target, plan bridge.Plan) (bridge.Bridge, error) {
	info, err := os.Stat(target.Path)
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrTargetNotFound, target.Path, err)
	}
	var argv []string
	dir := target.Path
	if !info.IsDir() {
		argv = []string{target.Path, "--centralizer"}
		dir = filepath.Dir(target.Path)
	} else if cargo := rt.Cargo(); cargo.Available {
		argv = []string{cargo.Command, "run", "--quiet", "--", "--centralizer"}
	} else {
		return nil, czerr.New(czerr.ErrRuntimeUnavailable, "cargo not found")
	}
	b, _, err := shim.ConnectStdio(ctx, shim.StdioConfig{Argv: argv, Dir: dir, Plan: plan})
	return b, err
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
