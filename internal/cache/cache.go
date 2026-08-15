// Package cache stores discovery results, plans, and generated shims.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// Store is a filesystem cache keyed by fingerprint.
type Store struct {
	Root string
	mu   sync.Mutex
}

// DefaultDir returns the user cache directory for Centralizer.
func DefaultDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "centralizer"), nil
}

// New returns a store rooted at dir, creating it if needed.
func New(dir string) (*Store, error) {
	if dir == "" {
		var err error
		dir, err = DefaultDir()
		if err != nil {
			return nil, czerr.Wrap(czerr.ErrCache, "user cache", err)
		}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, czerr.Wrap(czerr.ErrCache, "mkdir", err)
	}
	return &Store{Root: dir}, nil
}

func (s *Store) path(kind, key string) string {
	return filepath.Join(s.Root, kind, key)
}

// Put writes JSON for key.
func (s *Store) Put(kind, key string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Join(s.Root, kind)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return czerr.Wrap(czerr.ErrCache, "mkdir", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(kind, key), data, 0o600)
}

// Get reads JSON for key.
func (s *Store) Get(kind, key string, dest any) error {
	data, err := os.ReadFile(s.path(kind, key))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// WriteFile stores raw bytes (generated shims).
func (s *Store) WriteFile(kind, key, name string, data []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Join(s.Root, kind, key)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", czerr.Wrap(czerr.ErrCache, "mkdir", err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return "", err
	}
	return p, nil
}

// List returns cache entries under kind, or all kinds if kind is empty.
func (s *Store) List(kind string) ([]string, error) {
	root := s.Root
	if kind != "" {
		root = filepath.Join(s.Root, kind)
	}
	var out []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(s.Root, path)
		out = append(out, rel)
		return nil
	})
	return out, nil
}

// Clear removes the entire cache or one kind.
func (s *Store) Clear(kind string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if kind == "" {
		if err := os.RemoveAll(s.Root); err != nil {
			return err
		}
		return os.MkdirAll(s.Root, 0o700)
	}
	return os.RemoveAll(filepath.Join(s.Root, kind))
}
