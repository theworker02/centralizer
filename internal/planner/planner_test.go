package planner

import (
	"testing"

	"github.com/theworker02/centralizer/internal/security"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/manifest"
)

func testGraph() capability.Graph {
	g := capability.Graph{Language: "python", Runtime: "CPython"}
	g.Add(capability.Capability{Kind: capability.Stdio, Available: true, Confidence: 1})
	g.Add(capability.Capability{Kind: capability.Process, Available: true, Confidence: 1})
	g.Add(capability.Capability{Kind: capability.UnixSocket, Available: true, Confidence: 1})
	return g
}

func TestDeterministicSelection(t *testing.T) {
	in := Input{
		Language: "python",
		Runtime:  "CPython 3",
		Adapter:  "python",
		Graph:    testGraph(),
		Policy:   security.Engine{Policy: security.DefaultPolicy()},
		HostGOOS: "linux",
	}
	a, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if a.Selected.Strategy != b.Selected.Strategy || a.Selected.Scores.Overall != b.Selected.Scores.Overall {
		t.Fatalf("non-deterministic: %+v vs %+v", a.Selected, b.Selected)
	}
	if a.Selected.Strategy != bridge.StrategyUnixSocket {
		t.Fatalf("expected unix socket, got %s", a.Selected.Strategy)
	}
}

func TestStdioWhenNoUnix(t *testing.T) {
	g := capability.Graph{}
	g.Add(capability.Capability{Kind: capability.Stdio, Available: true, Confidence: 1})
	g.Add(capability.Capability{Kind: capability.Process, Available: true, Confidence: 1})
	in := Input{Graph: g, Policy: security.Engine{Policy: security.DefaultPolicy()}, HostGOOS: "linux"}
	res, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Selected.Strategy != bridge.StrategyStdio {
		t.Fatalf("got %s", res.Selected.Strategy)
	}
}

func TestPolicyBlocksNative(t *testing.T) {
	f := false
	g := capability.Graph{}
	g.Add(capability.Capability{Kind: capability.InProcess, Available: true, Confidence: 1})
	g.Add(capability.Capability{Kind: capability.Stdio, Available: true, Confidence: 1})
	g.Add(capability.Capability{Kind: capability.Process, Available: true, Confidence: 1})
	in := Input{
		Graph:    g,
		Policy:   security.Engine{Policy: manifest.Policy{NativeExecution: &f}},
		HostGOOS: "linux",
	}
	res, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Selected.Strategy == bridge.StrategyInProcess {
		t.Fatal("native should be blocked")
	}
}

func TestWindowsPrefersStdioWithoutPipesCap(t *testing.T) {
	g := capability.Graph{}
	g.Add(capability.Capability{Kind: capability.Stdio, Available: true, Confidence: 1})
	g.Add(capability.Capability{Kind: capability.Process, Available: true, Confidence: 1})
	in := Input{Graph: g, Policy: security.Engine{Policy: security.DefaultPolicy()}, HostGOOS: "windows"}
	res, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Selected.Strategy != bridge.StrategyStdio {
		t.Fatalf("got %s", res.Selected.Strategy)
	}
}
