package lockfile

import (
	"path/filepath"
	"testing"

	"github.com/theworker02/centralizer/pkg/bridge"
)

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, Name)
	plan := bridge.Plan{
		Strategy:    bridge.StrategyStdio,
		Adapter:     "python",
		Transport:   "stdio",
		Runtime:     "CPython",
		Language:    "Python",
		Fingerprint: "abc",
		Scores:      bridge.Scores{Overall: 85},
		Reasons:     []string{"stdio is portable"},
	}
	in := FromPlan("./analytics", "Python", "CPython", "abc", plan)
	if err := Write(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.Adapter != "python" || out.Transport != "stdio" || out.Strategy != "stdio" {
		t.Fatalf("%+v", out)
	}
	if !out.Matches(plan) {
		t.Fatal("expected match")
	}
	plan.Transport = "tcp"
	if out.Matches(plan) {
		t.Fatal("expected mismatch after transport change")
	}
}

func TestParseRejectsIncomplete(t *testing.T) {
	if _, err := Parse([]byte(`{"version":1}`)); err == nil {
		t.Fatal("expected error")
	}
}
