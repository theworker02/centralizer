package explain_test

import (
	"strings"
	"testing"

	"github.com/theworker02/centralizer/internal/discovery"
	"github.com/theworker02/centralizer/internal/planner"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/explain"
)

func TestBuildCallCapable(t *testing.T) {
	analysis := &discovery.Result{
		Primary: adapter.Detection{
			Adapter:    "python",
			Language:   "Python",
			Runtime:    "CPython",
			Confidence: 0.95,
			Evidence:   []string{"pyproject.toml"},
		},
		Detections: []adapter.Detection{
			{Adapter: "python", Language: "Python", Confidence: 0.95},
			{Adapter: "go", Language: "Go", Confidence: 0.1},
		},
	}
	plan := &planner.Result{
		Selected: bridge.Plan{
			Strategy:  bridge.StrategyStdio,
			Transport: "stdio",
			Adapter:   "python",
			Scores:    bridge.Scores{Overall: 82},
		},
	}
	cat := []adapter.Info{{Name: "python", Tier: 1, Invocation: true, Notes: "generated stdio/tcp shim"}}
	rep := explain.Build("./analytics", analysis, plan, cat, "Detected runtime:\nCPython\n")
	if !rep.CallImplemented {
		t.Fatal("expected Call implemented")
	}
	if rep.DetectionScore != 0.95 || rep.Adapter != "python" {
		t.Fatalf("%+v", rep)
	}
	text := rep.Text()
	for _, want := range []string{"Call implemented: yes", "score:     0.95", "Next steps", "centralizer call"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}

func TestBuildDetectOnlyHonesty(t *testing.T) {
	analysis := &discovery.Result{
		Primary: adapter.Detection{
			Adapter:    "lua",
			Language:   "Lua",
			Runtime:    "lua",
			Confidence: 0.75,
			Evidence:   []string{"hello.lua"},
		},
	}
	// Even if a planner selected stdio (host caps), catalog says no Call.
	plan := &planner.Result{
		Selected: bridge.Plan{
			Strategy:  bridge.StrategyStdio,
			Transport: "stdio",
			Adapter:   "lua",
			Scores:    bridge.Scores{Overall: 80},
		},
	}
	cat := []adapter.Info{{Name: "lua", Tier: 3, Invocation: false, Notes: "detect-only"}}
	rep := explain.Build("hello.lua", analysis, plan, cat, "should be suppressed")
	if rep.CallImplemented {
		t.Fatal("lua must not be marked Call-capable")
	}
	if rep.SelectedStrategy != "" || rep.PlannerText != "" {
		t.Fatalf("detect-only must not surface Call plan: %+v", rep)
	}
	text := rep.Text()
	if !strings.Contains(text, "Call implemented: no (detect-only)") {
		t.Fatalf("honesty missing:\n%s", text)
	}
	if !strings.Contains(text, "not applicable") {
		t.Fatalf("bridge plan honesty missing:\n%s", text)
	}
	if !strings.Contains(text, "Call is NOT implemented") {
		t.Fatalf("next steps missing:\n%s", text)
	}
	if strings.Contains(text, "Call is implemented for this adapter") {
		t.Fatalf("must not suggest Call for detect-only:\n%s", text)
	}
}

func TestKnownFallback(t *testing.T) {
	analysis := &discovery.Result{
		Primary: adapter.Detection{Adapter: "rust", Language: "Rust", Confidence: 0.8},
	}
	rep := explain.Build(".", analysis, nil, nil, "")
	if !rep.CallImplemented || rep.Tier != 1 {
		t.Fatalf("%+v", rep)
	}
}
