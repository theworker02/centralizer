// Package explain builds evaluator-oriented reports for centralizer explain.
//
// The report is deliberately honest about detect-vs-call: Invocation is taken
// from the adapter catalog, never inferred from a successful Detect score.
package explain

import (
	"fmt"
	"strings"

	"github.com/theworker02/centralizer/internal/discovery"
	"github.com/theworker02/centralizer/internal/planner"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
)

// Report is the structured output of an evaluation-oriented explain.
type Report struct {
	Target            string            `json:"target"`
	DetectionScore    float64           `json:"detection_score"`
	Language          string            `json:"language"`
	Runtime           string            `json:"runtime,omitempty"`
	Adapter           string            `json:"adapter"`
	Tier              int               `json:"tier"`
	CallImplemented   bool              `json:"call_implemented"`
	AdapterNotes      string            `json:"adapter_notes,omitempty"`
	SelectedStrategy  string            `json:"selected_strategy,omitempty"`
	SelectedTransport string            `json:"selected_transport,omitempty"`
	PlanScore         int               `json:"plan_score,omitempty"`
	Evidence          []string          `json:"evidence,omitempty"`
	Hypotheses        []Hypothesis      `json:"hypotheses,omitempty"`
	NextSteps         []string          `json:"next_steps"`
	Plan              *planner.Result   `json:"plan,omitempty"`
	Analysis          *discovery.Result `json:"analysis,omitempty"`
	PlannerText       string            `json:"-"`
}

// Hypothesis is one scored detection candidate.
type Hypothesis struct {
	Adapter    string  `json:"adapter"`
	Language   string  `json:"language"`
	Confidence float64 `json:"confidence"`
}

// Build combines discovery, planning, and catalog honesty into one report.
func Build(target string, analysis *discovery.Result, plan *planner.Result, catalog []adapter.Info, plannerText string) Report {
	r := Report{
		Target:      target,
		Analysis:    analysis,
		Plan:        plan,
		PlannerText: plannerText,
	}
	if analysis != nil {
		r.DetectionScore = analysis.Primary.Confidence
		r.Language = analysis.Primary.Language
		r.Runtime = analysis.Primary.Runtime
		r.Adapter = analysis.Primary.Adapter
		r.Evidence = append([]string(nil), analysis.Primary.Evidence...)
		for _, d := range analysis.Detections {
			r.Hypotheses = append(r.Hypotheses, Hypothesis{
				Adapter:    d.Adapter,
				Language:   d.Language,
				Confidence: d.Confidence,
			})
		}
	}
	info := lookup(catalog, r.Adapter)
	r.Tier = info.Tier
	r.CallImplemented = info.Invocation
	r.AdapterNotes = info.Notes
	if plan != nil && r.CallImplemented {
		r.SelectedStrategy = string(plan.Selected.Strategy)
		r.SelectedTransport = plan.Selected.Transport
		r.PlanScore = plan.Selected.Scores.Overall
	} else if plan != nil && !r.CallImplemented {
		// Keep raw plan in JSON for diligence, but do not present it as a
		// Call-ready selection in the summary fields.
		r.SelectedStrategy = ""
		r.SelectedTransport = ""
		r.PlanScore = 0
	}
	// Suppress planner detail for detect-only so buyers are not led to Connect.
	if !r.CallImplemented {
		r.PlannerText = ""
	}
	r.NextSteps = nextSteps(r)
	return r
}

func lookup(catalog []adapter.Info, name string) adapter.Info {
	for _, info := range catalog {
		if info.Name == name {
			return info
		}
	}
	// Fall back to static known entry when the registry is empty in tests.
	if info, ok := adapter.Known(name); ok {
		return info
	}
	return adapter.Info{Name: name, Tier: 3, Notes: "unknown adapter"}
}

func nextSteps(r Report) []string {
	var steps []string
	if r.Adapter == "" {
		return []string{
			"No adapter matched this target. Run: centralizer detect <target>",
			"Confirm the path exists and contains language markers (go.mod, pyproject.toml, package.json, …).",
		}
	}
	steps = append(steps, fmt.Sprintf("Detection chose adapter %q (score %.2f).", r.Adapter, r.DetectionScore))
	if r.CallImplemented {
		steps = append(steps,
			"Call is implemented for this adapter. Try: centralizer connect "+quote(r.Target),
			"Then: centralizer call "+quote(r.Target)+" <function> [k=v...]",
			"Or run examples/polyglot-hello to see a Go host call Python and Node in one process.",
		)
		if r.SelectedTransport != "" {
			steps = append(steps, fmt.Sprintf("Planner selected %s via %s (score %d).",
				displayStrategy(bridge.Strategy(r.SelectedStrategy)), r.SelectedTransport, r.PlanScore))
		}
	} else {
		steps = append(steps,
			"Call is NOT implemented for this adapter (detect-only). Do not treat Detect confidence as Call readiness.",
			"Use: centralizer detect "+quote(r.Target)+" and centralizer adapters to review the matrix.",
			"For a working Call path, try examples/go-python, examples/go-node, or examples/polyglot-hello.",
		)
	}
	steps = append(steps,
		"Host readiness: centralizer doctor",
		"Buyer walkthrough: docs/acquisition/BUYER_DEMO.md and docs/EVALUATOR_GUIDE.md",
	)
	return steps
}

func quote(s string) string {
	if s == "" {
		return "<target>"
	}
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}

func displayStrategy(s bridge.Strategy) string {
	switch s {
	case bridge.StrategyUnixSocket:
		return "Unix socket RPC"
	case bridge.StrategyStdio:
		return "stdio supervised process"
	case bridge.StrategyInProcess:
		return "in-process call"
	case bridge.StrategyNamedPipe:
		return "Windows named pipe"
	case bridge.StrategyTCP:
		return "local TCP"
	case bridge.StrategyWASM:
		return "WebAssembly"
	default:
		if s == "" {
			return "(none)"
		}
		return string(s)
	}
}

// Text renders a buyer/evaluator-friendly explain report.
func (r Report) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Centralizer explain — evaluation report\n")
	fmt.Fprintf(&b, "Target: %s\n\n", r.Target)

	b.WriteString("Detection\n")
	if r.Adapter == "" {
		b.WriteString("  (no primary adapter)\n")
	} else {
		fmt.Fprintf(&b, "  score:     %.2f\n", r.DetectionScore)
		fmt.Fprintf(&b, "  language:  %s\n", r.Language)
		if r.Runtime != "" {
			fmt.Fprintf(&b, "  runtime:   %s\n", r.Runtime)
		}
		fmt.Fprintf(&b, "  adapter:   %s\n", r.Adapter)
		if len(r.Evidence) > 0 {
			b.WriteString("  evidence:\n")
			for _, e := range r.Evidence {
				fmt.Fprintf(&b, "    - %s\n", e)
			}
		}
	}
	if len(r.Hypotheses) > 1 {
		b.WriteString("  other hypotheses:\n")
		for _, h := range r.Hypotheses {
			if h.Adapter == r.Adapter {
				continue
			}
			fmt.Fprintf(&b, "    - %s (%s) %.2f\n", h.Adapter, h.Language, h.Confidence)
		}
	}

	b.WriteString("\nAdapter capability (catalog — authoritative)\n")
	fmt.Fprintf(&b, "  tier:             %d\n", r.Tier)
	call := "no (detect-only)"
	if r.CallImplemented {
		call = "yes"
	}
	fmt.Fprintf(&b, "  Call implemented: %s\n", call)
	if r.AdapterNotes != "" {
		fmt.Fprintf(&b, "  notes:            %s\n", r.AdapterNotes)
	}

	b.WriteString("\nBridge plan\n")
	if !r.CallImplemented {
		b.WriteString("  (not applicable — adapter is detect-only; Connect/Call return ErrNotImplemented)\n")
	} else if r.SelectedStrategy == "" {
		b.WriteString("  (no viable plan)\n")
	} else {
		fmt.Fprintf(&b, "  strategy:  %s\n", displayStrategy(bridge.Strategy(r.SelectedStrategy)))
		fmt.Fprintf(&b, "  transport: %s\n", r.SelectedTransport)
		fmt.Fprintf(&b, "  score:     %d / 100\n", r.PlanScore)
	}

	b.WriteString("\nNext steps\n")
	for _, s := range r.NextSteps {
		fmt.Fprintf(&b, "  - %s\n", s)
	}

	if r.PlannerText != "" {
		b.WriteString("\n--- planner detail ---\n")
		b.WriteString(r.PlannerText)
		if !strings.HasSuffix(r.PlannerText, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
