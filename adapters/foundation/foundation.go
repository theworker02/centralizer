// Package foundation provides detect-only adapters for languages that are
// not yet a supported invocation path. These adapters exist so discovery
// and the compatibility matrix stay honest: Detect may succeed; Connect
// returns ErrNotImplemented.
package foundation

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Spec describes a detect-only adapter.
type Spec struct {
	AdapterName string
	Language    string
	Runtime     string
	Markers     []string
	Extensions  []string
	Tier        int
}

// Adapter is a detect-only implementation.
type Adapter struct{ Spec Spec }

func (a *Adapter) Name() string { return a.Spec.AdapterName }

func (a *Adapter) Detect(_ context.Context, target adapter.Target) (adapter.Detection, error) {
	d := adapter.Detection{Adapter: a.Name(), Language: a.Spec.Language, Runtime: a.Spec.Runtime}
	info, err := os.Stat(target.Path)
	if err != nil {
		return d, czerr.ErrUnsupportedTarget
	}
	score := 0.0
	var evidence []string
	if info.IsDir() {
		for _, m := range a.Spec.Markers {
			if matchMarker(target.Path, m) {
				score += 0.4
				evidence = append(evidence, m)
			}
		}
		for _, ext := range a.Spec.Extensions {
			if hasExt(target.Path, ext) {
				score += 0.2
				evidence = append(evidence, ext+" source files")
				break
			}
		}
	} else {
		ext := strings.ToLower(filepath.Ext(target.Path))
		for _, want := range a.Spec.Extensions {
			if ext == want || strings.HasSuffix(strings.ToLower(target.Path), strings.ToLower(want)) {
				score = 0.75
				evidence = append(evidence, filepath.Base(target.Path))
			}
		}
	}
	if score > 0.95 {
		score = 0.95
	}
	if score < 0.2 {
		return d, czerr.ErrUnsupportedTarget
	}
	d.Confidence = score
	d.Evidence = evidence
	return d, nil
}

func (a *Adapter) Capabilities(context.Context, adapter.Target) ([]capability.Capability, error) {
	return []capability.Capability{
		{Kind: capability.CLI, Available: false, Confidence: 0.2, Detail: "foundation adapter; invocation not implemented"},
	}, nil
}

func (a *Adapter) Prepare(context.Context, adapter.Target) error { return nil }

func (a *Adapter) Connect(context.Context, adapter.Target, bridge.Plan) (bridge.Bridge, error) {
	return nil, czerr.New(czerr.ErrNotImplemented, a.Spec.Language+" invocation is not implemented in this release")
}

func matchMarker(root, marker string) bool {
	if strings.ContainsAny(marker, "*") {
		matches, _ := filepath.Glob(filepath.Join(root, marker))
		return len(matches) > 0
	}
	_, err := os.Stat(filepath.Join(root, marker))
	return err == nil
}

func hasExt(root, ext string) bool {
	found := false
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "target") {
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

// All returns Tier 1 residual plus Tier 2/3 foundation adapters.
func All() []adapter.Adapter {
	specs := []Spec{
		{AdapterName: "c", Language: "C", Runtime: "native", Markers: []string{"CMakeLists.txt", "Makefile"}, Extensions: []string{".c", ".h"}, Tier: 1},
		{AdapterName: "cpp", Language: "C++", Runtime: "native", Markers: []string{"CMakeLists.txt"}, Extensions: []string{".cpp", ".cc", ".cxx", ".hpp"}, Tier: 1},
		{AdapterName: "wasm", Language: "WebAssembly", Runtime: "wasm", Markers: nil, Extensions: []string{".wasm"}, Tier: 1},
		{AdapterName: "jvm", Language: "Java", Runtime: "JVM", Markers: []string{"pom.xml", "build.gradle", "build.gradle.kts"}, Extensions: []string{".java", ".kt", ".jar"}, Tier: 2},
		{AdapterName: "dotnet", Language: "C#", Runtime: ".NET", Markers: []string{"*.csproj", "*.fsproj"}, Extensions: []string{".cs", ".fs"}, Tier: 2},
		{AdapterName: "ruby", Language: "Ruby", Runtime: "MRI", Markers: []string{"Gemfile"}, Extensions: []string{".rb"}, Tier: 2},
		{AdapterName: "php", Language: "PHP", Runtime: "php", Markers: []string{"composer.json"}, Extensions: []string{".php"}, Tier: 2},
		{AdapterName: "swift", Language: "Swift", Runtime: "swift", Markers: []string{"Package.swift"}, Extensions: []string{".swift"}, Tier: 3},
		{AdapterName: "dart", Language: "Dart", Runtime: "dart", Markers: []string{"pubspec.yaml"}, Extensions: []string{".dart"}, Tier: 3},
		{AdapterName: "lua", Language: "Lua", Runtime: "lua", Markers: nil, Extensions: []string{".lua"}, Tier: 3},
		{AdapterName: "zig", Language: "Zig", Runtime: "zig", Markers: []string{"build.zig"}, Extensions: []string{".zig"}, Tier: 3},
	}
	out := make([]adapter.Adapter, 0, len(specs))
	for _, s := range specs {
		sp := s
		out = append(out, &Adapter{Spec: sp})
	}
	return out
}
