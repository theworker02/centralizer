// Package runtime locates installed language runtimes on the host.
package runtime

import (
	"os/exec"
	"runtime"
	"strings"
)

// Info describes an installed interpreter or toolchain.
type Info struct {
	Name           string `json:"name"`
	Command        string `json:"command"`
	Version        string `json:"version,omitempty"`
	Available      bool   `json:"available"`
	Implementation string `json:"implementation,omitempty"`
}

// Look returns runtime info for the first available command.
func Look(name string, candidates ...string) Info {
	info := Info{Name: name}
	for _, c := range candidates {
		path, err := exec.LookPath(c)
		if err != nil {
			continue
		}
		if isWindowsStoreStub(path) {
			continue
		}
		info.Available = true
		info.Command = path
		info.Version = versionOf(path)
		return info
	}
	return info
}

func isWindowsStoreStub(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "windowsapps")
}

func versionOf(cmd string) string {
	out, err := exec.Command(cmd, "--version").CombinedOutput()
	if err != nil {
		out, err = exec.Command(cmd, "-version").CombinedOutput()
		if err != nil {
			return ""
		}
	}
	line := strings.TrimSpace(string(out))
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}

// Host reports GOOS/GOARCH.
func Host() (string, string) { return runtime.GOOS, runtime.GOARCH }

// Python locates a Python interpreter.
func Python() Info {
	info := Look("python", "python", "python3", "py")
	if info.Available {
		info.Implementation = "CPython"
	}
	return info
}

// Node locates Node.js.
func Node() Info {
	info := Look("node", "node")
	if info.Available {
		info.Implementation = "Node.js"
	}
	return info
}

// Rustc locates rustc.
func Rustc() Info { return Look("rustc", "rustc") }

// Cargo locates cargo.
func Cargo() Info { return Look("cargo", "cargo") }

// Go locates the Go toolchain.
func Go() Info { return Look("go", "go") }
