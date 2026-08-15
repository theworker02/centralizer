package schema

import (
	"os"
	"path/filepath"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// LoadFile reads an explicit schema document.
func LoadFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrSchemaInvalid, "read "+path, err)
	}
	return ParseYAML(data)
}

// Discover loads an explicit schema when one exists.
//
// Search order:
//  1. explicit path from a manifest `schema:` field (required if set)
//  2. <target>/schema.yaml
//  3. <target>/schema.yml
//
// A missing optional file is not an error; Discover returns (nil, nil).
func Discover(targetDir, explicit string) (*Schema, error) {
	if explicit != "" {
		path := explicit
		if !filepath.IsAbs(path) {
			if targetDir != "" {
				cand := filepath.Join(targetDir, explicit)
				if _, err := os.Stat(cand); err == nil {
					path = cand
				} else if _, err := os.Stat(explicit); err == nil {
					path = explicit
				} else {
					path = cand
				}
			}
		}
		return LoadFile(path)
	}
	if targetDir == "" {
		return nil, nil
	}
	for _, name := range []string{"schema.yaml", "schema.yml"} {
		p := filepath.Join(targetDir, name)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		return LoadFile(p)
	}
	return nil, nil
}
