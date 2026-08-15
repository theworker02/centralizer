package manifest

import (
	"os"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// StarterYAML is the document written by `centralizer init`.
const StarterYAML = `centralizer:
  version: 1
services:
  example:
    source: .
    language: auto
policy:
  recovery: automatic
  isolation: process
  tracing: false
  native_execution: true
  network: localhost_only
  subprocesses: true
  generated_code: true
`

// DefaultPath is the conventional manifest filename.
const DefaultPath = "centralizer.yaml"

// WriteStarter writes a starter manifest. It refuses to overwrite unless force is set.
func WriteStarter(path string, force bool) error {
	if path == "" {
		path = DefaultPath
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return czerr.New(czerr.ErrInvalidArgument, path+" already exists (pass force to overwrite)")
		}
	}
	if err := os.WriteFile(path, []byte(StarterYAML), 0o644); err != nil {
		return czerr.Wrap(czerr.ErrManifestInvalid, "write starter", err)
	}
	return nil
}
