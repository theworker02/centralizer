package manifest

import "testing"

func TestParse(t *testing.T) {
	raw := []byte(`
centralizer:
  version: 1
services:
  analytics:
    source: ./analytics
    language: auto
  engine:
    source: ./engine
    language: rust
    prefer:
      - native
      - wasm
policy:
  recovery: automatic
  isolation: process
  tracing: true
  native_execution: false
  network: localhost_only
  subprocesses: true
`)
	m, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m.Services["analytics"].Source != "./analytics" {
		t.Fatalf("%+v", m.Services)
	}
	if m.Policy.NativeExecutionAllowed() {
		t.Fatal("native should be denied")
	}
	if !m.Policy.SubprocessesAllowed() {
		t.Fatal("subprocesses should be allowed")
	}
}

func TestParseRejectsBadVersion(t *testing.T) {
	if _, err := Parse([]byte("centralizer:\n  version: 99\n")); err == nil {
		t.Fatal("expected error")
	}
}
