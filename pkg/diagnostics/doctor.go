// Package diagnostics implements `centralizer doctor`.
package diagnostics

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/theworker02/centralizer/internal/cache"
	rt "github.com/theworker02/centralizer/internal/runtime"
	"github.com/theworker02/centralizer/internal/version"
)

// Check is one doctor finding.
type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

// Report is the full doctor output.
type Report struct {
	Version string  `json:"version"`
	OS      string  `json:"os"`
	Arch    string  `json:"arch"`
	Checks  []Check `json:"checks"`
}

// Run inspects the host.
func Run(adapterNames []string) Report {
	r := Report{Version: version.Version, OS: runtime.GOOS, Arch: runtime.GOARCH}
	r.Checks = append(r.Checks, goCheck())
	r.Checks = append(r.Checks, runtimeCheck("Python", rt.Python(), "Install CPython 3 to use the Python adapter."))
	r.Checks = append(r.Checks, runtimeCheck("Node.js", rt.Node(), "Install Node.js to use the Node adapter."))
	r.Checks = append(r.Checks, runtimeCheck("Cargo", rt.Cargo(), "Install Rust/cargo to build Rust targets."))
	r.Checks = append(r.Checks, Check{Name: "OS sockets", OK: true, Detail: socketDetail()})
	r.Checks = append(r.Checks, Check{
		Name:   "shared memory",
		OK:     false,
		Detail: "experimental; disabled by default",
		Hint:   "Do not enable unless you accept the experimental contract.",
	})
	r.Checks = append(r.Checks, cacheCheck())
	r.Checks = append(r.Checks, gitCheck())
	r.Checks = append(r.Checks, protocolCheck())
	if len(adapterNames) > 0 {
		r.Checks = append(r.Checks, Check{
			Name:   "adapters",
			OK:     true,
			Detail: strings.Join(adapterNames, ", "),
		})
	}
	return r
}

func goCheck() Check {
	info := rt.Go()
	if !info.Available {
		return Check{Name: "Go toolchain", OK: false, Hint: "Install Go 1.23 or later."}
	}
	return Check{Name: "Go toolchain", OK: true, Detail: info.Version}
}

func runtimeCheck(name string, info rt.Info, hint string) Check {
	if !info.Available {
		return Check{Name: name, OK: false, Hint: hint}
	}
	return Check{Name: name, OK: true, Detail: info.Version}
}

func socketDetail() string {
	if runtime.GOOS == "windows" {
		return "TCP loopback available; Unix sockets unavailable; named-pipe client foundation present (server side experimental)"
	}
	return "Unix sockets and TCP loopback available"
}

func cacheCheck() Check {
	dir, err := cache.DefaultDir()
	if err != nil {
		return Check{Name: "cache", OK: false, Detail: err.Error()}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Check{Name: "cache", OK: false, Detail: err.Error(), Hint: "Check local permissions for the user cache directory."}
	}
	probe := filepath.Join(dir, ".doctor-write-probe")
	if err := os.WriteFile(probe, []byte("ok"), 0o600); err != nil {
		return Check{Name: "cache", OK: false, Detail: err.Error(), Hint: "Cache directory is not writable."}
	}
	_ = os.Remove(probe)
	return Check{Name: "cache", OK: true, Detail: dir}
}

func gitCheck() Check {
	path, err := exec.LookPath("git")
	if err != nil {
		return Check{Name: "Git", OK: false, Hint: "Install Git to fingerprint repositories and use VCS metadata."}
	}
	return Check{Name: "Git", OK: true, Detail: path}
}

func protocolCheck() Check {
	var maj, min int
	n, err := fmt.Sscanf(version.Protocol, "%d.%d", &maj, &min)
	if err != nil || n < 1 || maj < 1 {
		return Check{Name: "protocol", OK: false, Detail: version.Protocol, Hint: "This build reports an unparseable protocol version."}
	}
	return Check{
		Name:   "protocol",
		OK:     true,
		Detail: fmt.Sprintf("Centralizer Protocol %s (major %d)", version.Protocol, maj),
	}
}

// Text renders a human report.
func (r Report) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Centralizer %s doctor\n", r.Version)
	fmt.Fprintf(&b, "Host: %s/%s\n\n", r.OS, r.Arch)
	for _, c := range r.Checks {
		mark := "ok"
		if !c.OK {
			mark = "FAIL"
		}
		fmt.Fprintf(&b, "[%s] %s", mark, c.Name)
		if c.Detail != "" {
			fmt.Fprintf(&b, ": %s", c.Detail)
		}
		b.WriteByte('\n')
		if c.Hint != "" && !c.OK {
			fmt.Fprintf(&b, "      %s\n", c.Hint)
		}
	}
	return b.String()
}
