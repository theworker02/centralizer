// Package discovery inspects targets and returns scored runtime hypotheses.
package discovery

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Result is a complete target analysis.
type Result struct {
	Target      string              `json:"target"`
	Path        string              `json:"path"`
	Detections  []adapter.Detection `json:"detections"`
	Primary     adapter.Detection   `json:"primary"`
	Graph       capability.Graph    `json:"graph"`
	OS          string              `json:"os"`
	Arch        string              `json:"arch"`
	Files       []string            `json:"files,omitempty"`
	Fingerprint string              `json:"fingerprint"`
}

// Engine walks a target and asks registered adapters to detect it.
type Engine struct {
	Registry *adapter.Registry
}

// Analyze inspects path and returns scored detections. Ambiguous results
// keep multiple hypotheses; nothing is reported as certain without evidence.
func (e *Engine) Analyze(ctx context.Context, target adapter.Target) (*Result, error) {
	if target.Path == "" {
		return nil, czerr.New(czerr.ErrInvalidArgument, "empty path")
	}
	if strings.HasPrefix(target.Path, "native:") || strings.HasPrefix(target.Ref, "native:") {
		dets, err := adapter.DetectAll(ctx, e.Registry, target)
		if err != nil {
			return nil, err
		}
		res := &Result{
			Target:     target.Ref,
			Path:       target.Path,
			Detections: dets,
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
		}
		if len(dets) > 0 {
			res.Primary = dets[0]
			res.Primary.Primary = true
		}
		res.Graph = capability.Graph{Target: target.Ref, Language: "Go", Runtime: "gc"}
		res.Graph.Add(capability.Capability{Kind: capability.InProcess, Available: true, Confidence: 1})
		res.Graph.Add(capability.Capability{Kind: capability.Stdio, Available: true, Confidence: 0.2})
		res.Fingerprint = Fingerprint(target.Ref, nil, res.Primary)
		return res, nil
	}
	info, err := os.Stat(target.Path)
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrTargetNotFound, target.Path, err)
	}
	files, err := listMarkers(target.Path, info.IsDir())
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrDiscoveryFailed, "list", err)
	}
	dets, err := adapter.DetectAll(ctx, e.Registry, target)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(dets, func(i, j int) bool {
		if dets[i].Confidence == dets[j].Confidence {
			return dets[i].Adapter < dets[j].Adapter
		}
		return dets[i].Confidence > dets[j].Confidence
	})
	res := &Result{
		Target:     target.Ref,
		Path:       target.Path,
		Detections: dets,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Files:      files,
	}
	if len(dets) > 0 {
		res.Primary = dets[0]
		res.Primary.Primary = true
		dets[0].Primary = true
		res.Detections = dets
	}
	res.Graph = capability.Graph{
		Target:   target.Ref,
		Language: res.Primary.Language,
		Runtime:  res.Primary.Runtime,
	}
	if a, ok := e.Registry.Get(res.Primary.Adapter); ok {
		caps, err := a.Capabilities(ctx, target)
		if err == nil {
			for _, c := range caps {
				res.Graph.Add(c)
			}
		}
	}
	addHostCapabilities(&res.Graph)
	res.Fingerprint = Fingerprint(target.Path, files, res.Primary)
	return res, nil
}

func addHostCapabilities(g *capability.Graph) {
	// Host facts only. Adapter-specific transports must be reported by
	// the adapter; otherwise the planner would select strategies it
	// cannot actually establish.
	g.Add(capability.Capability{Kind: capability.Stdio, Available: true, Confidence: 1, Detail: "host process I/O"})
	g.Add(capability.Capability{Kind: capability.Process, Available: true, Confidence: 1})
}

var markerNames = []string{
	"go.mod", "Cargo.toml", "package.json", "deno.json", "pyproject.toml",
	"requirements.txt", "Pipfile", "pom.xml", "build.gradle", "build.gradle.kts",
	"CMakeLists.txt", "Makefile", "Package.swift", "pubspec.yaml", "Gemfile",
	"composer.json", "build.zig", "centralizer.yaml", "centralizer.yml",
}

func listMarkers(root string, isDir bool) ([]string, error) {
	if !isDir {
		return []string{filepath.Base(root)}, nil
	}
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == "node_modules" || name == "target" || name == "vendor" || name == ".venv" {
				return fs.SkipDir
			}
			return nil
		}
		for _, m := range markerNames {
			if name == m {
				rel, _ := filepath.Rel(root, path)
				out = append(out, rel)
			}
		}
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".csproj", ".fsproj", ".py", ".rs", ".go", ".js", ".ts", ".c", ".cc", ".cpp", ".h", ".hpp", ".wasm", ".so", ".dylib", ".dll":
			rel, _ := filepath.Rel(root, path)
			if !contains(out, rel) {
				out = append(out, rel)
			}
		}
		if len(out) > 64 {
			return fs.SkipAll
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// FormatAnalysis renders the human-readable detection report.
func FormatAnalysis(r *Result) string {
	var b strings.Builder
	b.WriteString("Target Analysis\n")
	for _, d := range r.Detections {
		fmt.Fprintf(&b, "%-14s%.2f\n", d.Language, d.Confidence)
	}
	if r.Primary.Language != "" {
		b.WriteString("Primary runtime:\n")
		b.WriteString(r.Primary.Runtime)
		if r.Primary.Version != "" {
			b.WriteString(" ")
			b.WriteString(r.Primary.Version)
		}
		b.WriteByte('\n')
		if len(r.Primary.Evidence) > 0 {
			b.WriteString("Evidence:\n")
			for _, e := range r.Primary.Evidence {
				b.WriteString("- ")
				b.WriteString(e)
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}
