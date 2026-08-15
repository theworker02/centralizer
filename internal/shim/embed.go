package shim

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/theworker02/centralizer/internal/cache"
	"github.com/theworker02/centralizer/pkg/czerr"
)

//go:embed templates/python_shim.py
var PythonShim []byte

//go:embed templates/node_shim.js
var NodeShim []byte

const ShimVersion = "2"

// Materialize writes a shim into the cache if absent or stale.
func Materialize(store *cache.Store, runtime, fingerprint string, body []byte, name string) (string, error) {
	if store == nil {
		return "", czerr.New(czerr.ErrCache, "no cache store")
	}
	sum := sha256.Sum256(body)
	key := fingerprint + "-" + ShimVersion + "-" + hex.EncodeToString(sum[:8])
	path, err := store.WriteFile("shims", key, name, body)
	if err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o700); err != nil {
		// Windows may ignore executable bits; still usable via interpreter.
		_ = err
	}
	return path, nil
}

// EnsureDir creates a restricted directory.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o700)
}

// Join is filepath.Join for adapters.
func Join(elem ...string) string { return filepath.Join(elem...) }
