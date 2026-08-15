package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAML(t *testing.T) {
	raw := []byte(`
service: analytics
functions:
  calculate:
    args:
      dataset: bytes
      iterations: uint
    returns:
      type: float
`)
	s, err := ParseYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	fn, ok := s.FunctionOf("calculate")
	if !ok {
		t.Fatal("missing function")
	}
	if fn.Args["dataset"].Type != "bytes" {
		t.Fatalf("args=%v", fn.Args)
	}
}

func TestParseRejectsEmptyService(t *testing.T) {
	if _, err := ParseYAML([]byte("functions: {}")); err == nil {
		t.Fatal("expected error")
	}
}

func TestDiscoverOptionalMissing(t *testing.T) {
	sc, err := Discover(t.TempDir(), "")
	if err != nil || sc != nil {
		t.Fatalf("sc=%v err=%v", sc, err)
	}
}

func TestDiscoverExplicitAndSidecar(t *testing.T) {
	dir := t.TempDir()
	raw := []byte("service: analytics\nfunctions:\n  calculate:\n    args:\n      value: float\n")
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	sc, err := Discover(dir, "")
	if err != nil || sc == nil || sc.Service != "analytics" {
		t.Fatalf("sidecar sc=%v err=%v", sc, err)
	}
	if sc.Inferred {
		t.Fatal("explicit sidecar must not be inferred")
	}
	other := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(other, []byte("service: custom\nfunctions: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sc, err = Discover(dir, "custom.yaml")
	if err != nil || sc == nil || sc.Service != "custom" {
		t.Fatalf("explicit sc=%v err=%v", sc, err)
	}
	if _, err := Discover(dir, "missing.yaml"); err == nil {
		t.Fatal("expected missing explicit schema to fail")
	}
}
