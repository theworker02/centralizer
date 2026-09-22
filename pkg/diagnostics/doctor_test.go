package diagnostics

import (
	"strings"
	"testing"

	"github.com/theworker02/centralizer/pkg/adapter"
)

func TestRunIncludesCoreChecks(t *testing.T) {
	rep := Run([]string{"python", "go"})
	if rep.Version == "" || rep.OS == "" {
		t.Fatalf("%+v", rep)
	}
	want := map[string]bool{"cache": false, "Git": false, "protocol": false, "adapters": false}
	for _, c := range rep.Checks {
		if _, ok := want[c.Name]; ok {
			want[c.Name] = true
		}
		if c.Name == "protocol" && !c.OK {
			t.Fatalf("protocol: %+v", c)
		}
		if c.Name == "cache" && !c.OK {
			t.Fatalf("cache: %+v", c)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("missing check %s", name)
		}
	}
	if rep.Text() == "" {
		t.Fatal("empty text")
	}
}

func TestRunCatalogSeparatesCallAndDetect(t *testing.T) {
	cat := []adapter.Info{
		{Name: "python", Invocation: true},
		{Name: "lua", Invocation: false},
	}
	rep := RunCatalog([]string{"python", "lua"}, cat)
	if len(rep.CallCapable) != 1 || rep.CallCapable[0] != "python" {
		t.Fatalf("call=%v", rep.CallCapable)
	}
	if len(rep.DetectOnly) != 1 || rep.DetectOnly[0] != "lua" {
		t.Fatalf("detect=%v", rep.DetectOnly)
	}
	text := rep.Text()
	if !strings.Contains(text, "call-capable") || !strings.Contains(text, "detect-only") {
		t.Fatalf("text:\n%s", text)
	}
}
