// Package lockfile reads and writes an optional resolved-plan snapshot.
//
// A lock file is not required for Connect. It records the planner's
// selection so inspect/explain can compare a later run against a known plan.
package lockfile

import (
	"encoding/json"
	"os"
	"time"

	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// CurrentVersion is the lock document version.
const CurrentVersion = 1

// Name is the default filename written by `centralizer lock`.
const Name = "centralizer.lock"

// File is a resolved plan snapshot.
type File struct {
	Version     int           `json:"version"`
	Target      string        `json:"target"`
	Language    string        `json:"language,omitempty"`
	Runtime     string        `json:"runtime,omitempty"`
	Adapter     string        `json:"adapter"`
	Transport   string        `json:"transport"`
	Strategy    string        `json:"strategy"`
	Fingerprint string        `json:"fingerprint,omitempty"`
	Scores      bridge.Scores `json:"scores"`
	Reasons     []string      `json:"reasons,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// FromPlan builds a lock document from a selected plan.
func FromPlan(target, language, runtime, fingerprint string, plan bridge.Plan) File {
	return File{
		Version:     CurrentVersion,
		Target:      target,
		Language:    language,
		Runtime:     runtime,
		Adapter:     plan.Adapter,
		Transport:   plan.Transport,
		Strategy:    string(plan.Strategy),
		Fingerprint: firstNonEmpty(fingerprint, plan.Fingerprint),
		Scores:      plan.Scores,
		Reasons:     append([]string(nil), plan.Reasons...),
		CreatedAt:   time.Now().UTC(),
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Write encodes f as indented JSON.
func Write(path string, f File) error {
	if f.Version == 0 {
		f.Version = CurrentVersion
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return czerr.Wrap(czerr.ErrInvalidArgument, "lockfile encode", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return czerr.Wrap(czerr.ErrInvalidArgument, "lockfile write", err)
	}
	return nil
}

// Read loads a lock document.
func Read(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, czerr.Wrap(czerr.ErrTargetNotFound, "lockfile read", err)
	}
	return Parse(data)
}

// Parse decodes a lock document.
func Parse(data []byte) (File, error) {
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, czerr.Wrap(czerr.ErrManifestInvalid, "lockfile parse", err)
	}
	if f.Version == 0 {
		f.Version = CurrentVersion
	}
	if f.Version != CurrentVersion {
		return File{}, czerr.New(czerr.ErrManifestInvalid, "unsupported lockfile version")
	}
	if f.Adapter == "" || f.Transport == "" {
		return File{}, czerr.New(czerr.ErrManifestInvalid, "lockfile missing adapter or transport")
	}
	return f, nil
}

// Matches reports whether a live plan still agrees with the lock on the
// fields that identify a bridge (adapter, transport, strategy).
func (f File) Matches(plan bridge.Plan) bool {
	if f.Adapter != "" && f.Adapter != plan.Adapter {
		return false
	}
	if f.Transport != "" && f.Transport != plan.Transport {
		return false
	}
	if f.Strategy != "" && f.Strategy != string(plan.Strategy) {
		return false
	}
	return true
}
