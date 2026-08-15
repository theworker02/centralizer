package security

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// ResolveTarget validates and cleans a user-supplied target path.
// Remote URLs are rejected in v0.1 unless they are explicitly allowed
// later by policy. Discovered code is not trusted merely because it exists.
func ResolveTarget(ref string) (string, error) {
	if ref == "" {
		return "", czerr.New(czerr.ErrInvalidArgument, "empty target")
	}
	if strings.Contains(ref, "\x00") {
		return "", czerr.New(czerr.ErrSecurity, "null byte in path")
	}
	lower := strings.ToLower(ref)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "file:") {
		return "", czerr.New(czerr.ErrPolicyDenied, "remote targets are not enabled")
	}
	abs, err := filepath.Abs(ref)
	if err != nil {
		return "", czerr.Wrap(czerr.ErrTargetNotFound, ref, err)
	}
	cleaned := filepath.Clean(abs)
	if _, err := os.Stat(cleaned); err != nil {
		if os.IsNotExist(err) {
			return "", czerr.Wrap(czerr.ErrTargetNotFound, cleaned, err)
		}
		return "", czerr.Wrap(czerr.ErrSecurity, "stat", err)
	}
	return cleaned, nil
}

// Within reports whether child is inside parent after cleaning.
func Within(parent, child string) bool {
	p, err := filepath.Abs(parent)
	if err != nil {
		return false
	}
	c, err := filepath.Abs(child)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(p, c)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
