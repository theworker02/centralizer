package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/theworker02/centralizer/internal/version"
	"github.com/theworker02/centralizer/pkg/adapter"
)

// Fingerprint is a stable identity for cache keys. It incorporates
// source metadata, adapter identity, host architecture, and protocol version.
func Fingerprint(root string, files []string, primary adapter.Detection) string {
	h := sha256.New()
	_, _ = io.WriteString(h, "centralizer-fp-v1\n")
	_, _ = io.WriteString(h, version.Version+"\n")
	_, _ = io.WriteString(h, version.Protocol+"\n")
	_, _ = io.WriteString(h, runtime.GOOS+"/"+runtime.GOARCH+"\n")
	_, _ = io.WriteString(h, primary.Adapter+"|"+primary.Runtime+"|"+primary.Version+"\n")
	_, _ = io.WriteString(h, filepath.Clean(root)+"\n")
	sorted := append([]string(nil), files...)
	sort.Strings(sorted)
	for _, f := range sorted {
		p := filepath.Join(root, f)
		info, err := os.Stat(p)
		if err != nil {
			_, _ = io.WriteString(h, f+"\n")
			continue
		}
		_, _ = fmt.Fprintf(h, "%s\t%d\t%d\n", f, info.Size(), info.ModTime().UnixNano())
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Short returns the first 16 hex characters.
func Short(fp string) string {
	if len(fp) < 16 {
		return fp
	}
	return fp[:16]
}

// JoinKey builds a cache key from stable parts.
func JoinKey(parts ...string) string {
	return strings.Join(parts, ":")
}
