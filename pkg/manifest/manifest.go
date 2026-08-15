// Package manifest parses optional Centralizer configuration.
package manifest

import (
	"fmt"
	"os"

	"github.com/theworker02/centralizer/pkg/czerr"
	"gopkg.in/yaml.v3"
)

// CurrentVersion is the supported manifest schema version.
const CurrentVersion = 1

// Manifest is an optional project configuration. Zero configuration remains
// valid; the manifest exists for deterministic overrides.
type Manifest struct {
	Centralizer Centralizer        `yaml:"centralizer" json:"centralizer"`
	Services    map[string]Service `yaml:"services" json:"services"`
	Policy      Policy             `yaml:"policy" json:"policy"`
}

// Centralizer identifies the file format.
type Centralizer struct {
	Version int `yaml:"version" json:"version"`
}

// Service describes one target.
type Service struct {
	Source   string   `yaml:"source" json:"source"`
	Language string   `yaml:"language" json:"language"`
	Entry    string   `yaml:"entry,omitempty" json:"entry,omitempty"`
	Prefer   []string `yaml:"prefer,omitempty" json:"prefer,omitempty"`
	Schema   string   `yaml:"schema,omitempty" json:"schema,omitempty"`
}

// Policy constrains automation.
type Policy struct {
	Recovery          string   `yaml:"recovery,omitempty" json:"recovery,omitempty"`
	Isolation         string   `yaml:"isolation,omitempty" json:"isolation,omitempty"`
	Tracing           bool     `yaml:"tracing,omitempty" json:"tracing,omitempty"`
	NativeExecution   *bool    `yaml:"native_execution,omitempty" json:"native_execution,omitempty"`
	Network           string   `yaml:"network,omitempty" json:"network,omitempty"`
	Subprocesses      *bool    `yaml:"subprocesses,omitempty" json:"subprocesses,omitempty"`
	GeneratedCode     *bool    `yaml:"generated_code,omitempty" json:"generated_code,omitempty"`
	MemoryLimitBytes  int64    `yaml:"memory_limit_bytes,omitempty" json:"memory_limit_bytes,omitempty"`
	TimeoutMS         int      `yaml:"timeout_ms,omitempty" json:"timeout_ms,omitempty"`
	MaxRestarts       int      `yaml:"max_restarts,omitempty" json:"max_restarts,omitempty"`
	AllowedRuntimes   []string `yaml:"allowed_runtimes,omitempty" json:"allowed_runtimes,omitempty"`
	AllowedTransports []string `yaml:"allowed_transports,omitempty" json:"allowed_transports,omitempty"`
}

// Parse decodes and validates a manifest document.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, czerr.Wrap(czerr.ErrManifestInvalid, "parse yaml", err)
	}
	if m.Centralizer.Version == 0 {
		m.Centralizer.Version = CurrentVersion
	}
	if m.Centralizer.Version != CurrentVersion {
		return nil, czerr.New(czerr.ErrManifestInvalid, fmt.Sprintf("unsupported manifest version %d", m.Centralizer.Version))
	}
	for name, svc := range m.Services {
		if svc.Source == "" {
			return nil, czerr.New(czerr.ErrManifestInvalid, fmt.Sprintf("service %q missing source", name))
		}
		if svc.Language == "" {
			svc.Language = "auto"
			m.Services[name] = svc
		}
	}
	return &m, nil
}

// Load reads a manifest from disk.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrManifestInvalid, "read", err)
	}
	return Parse(data)
}

// NativeExecutionAllowed reports whether in-process native execution is allowed.
func (p Policy) NativeExecutionAllowed() bool {
	if p.NativeExecution == nil {
		return true
	}
	return *p.NativeExecution
}

// SubprocessesAllowed reports whether child processes may be started.
func (p Policy) SubprocessesAllowed() bool {
	if p.Subprocesses == nil {
		return true
	}
	return *p.Subprocesses
}

// GeneratedCodeAllowed reports whether shim generation is allowed.
func (p Policy) GeneratedCodeAllowed() bool {
	if p.GeneratedCode == nil {
		return true
	}
	return *p.GeneratedCode
}
