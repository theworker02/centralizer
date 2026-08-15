package security

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/manifest"
)

func TestResolveTargetRejectsURL(t *testing.T) {
	_, err := ResolveTarget("https://example.com/x")
	if err == nil {
		t.Fatal("expected deny")
	}
}

func TestResolveTargetOK(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveTarget(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("not abs: %s", got)
	}
}

func TestPolicyRuntime(t *testing.T) {
	e := Engine{Policy: manifest.Policy{AllowedRuntimes: []string{"python"}}}
	if err := e.AllowRuntime("python"); err != nil {
		t.Fatal(err)
	}
	if err := e.AllowRuntime("node"); err == nil {
		t.Fatal("expected deny")
	}
}

func TestFilterEnv(t *testing.T) {
	in := []string{"PATH=/bin", "LD_PRELOAD=evil", "HOME=" + os.Getenv("HOME")}
	out := FilterEnv(in)
	for _, e := range out {
		if e == "LD_PRELOAD=evil" {
			t.Fatal("ld_preload leaked")
		}
	}
}

func TestWithin(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if !Within(root, child) {
		t.Fatal("expected within")
	}
	if Within(child, root) {
		t.Fatal("parent should not be within child")
	}
}

func TestNativeDenied(t *testing.T) {
	f := false
	e := Engine{Policy: manifest.Policy{NativeExecution: &f}}
	if err := e.AllowNative(); err == nil || !isPolicy(err) {
		t.Fatalf("err=%v", err)
	}
}

func isPolicy(err error) bool {
	return errors.Is(err, czerr.ErrPolicyDenied)
}
