// Package planner selects a deterministic bridge strategy from a
// capability graph and host policy.
package planner

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/theworker02/centralizer/internal/security"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Weights are fixed so equivalent inputs produce equivalent rankings.
var defaultWeights = map[string]int{
	"performance":   12,
	"reliability":   14,
	"isolation":     14,
	"startup":       6,
	"serialization": 8,
	"compatibility": 16,
	"security":      12,
	"portability":   6,
	"debuggability": 6,
	"availability":  6,
}

// Input is everything the planner needs.
type Input struct {
	Language string
	Runtime  string
	Adapter  string
	Graph    capability.Graph
	Policy   security.Engine
	Prefer   []string
	HostGOOS string
}

// Result is the selected plan plus rejected and ranked candidates.
type Result struct {
	Selected   bridge.Plan        `json:"selected"`
	Candidates []bridge.Candidate `json:"candidates"`
}

// Plan evaluates strategies and returns a deterministic winner.
func Plan(in Input) (*Result, error) {
	if in.HostGOOS == "" {
		in.HostGOOS = runtime.GOOS
	}
	cands := evaluate(in)
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].Plan.Scores.Overall == cands[j].Plan.Scores.Overall {
			return cands[i].Plan.Strategy < cands[j].Plan.Strategy
		}
		return cands[i].Plan.Scores.Overall > cands[j].Plan.Scores.Overall
	})
	var selected *bridge.Candidate
	for i := range cands {
		if cands[i].Viable {
			selected = &cands[i]
			break
		}
	}
	if selected == nil {
		return nil, czerr.New(czerr.ErrPlannerFailed, "no viable bridge strategy")
	}
	var fallbacks []bridge.Strategy
	for _, c := range cands {
		if c.Viable && c.Plan.Strategy != selected.Plan.Strategy {
			fallbacks = append(fallbacks, c.Plan.Strategy)
		}
	}
	selected.Plan.Fallbacks = fallbacks
	return &Result{Selected: selected.Plan, Candidates: cands}, nil
}

func evaluate(in Input) []bridge.Candidate {
	type spec struct {
		strategy  bridge.Strategy
		transport string
		need      []capability.Kind
		score     func(Input) bridge.Scores
		extra     func(Input) (bool, string)
	}
	unixOK := in.HostGOOS != "windows"
	specs := []spec{
		{
			strategy:  bridge.StrategyInProcess,
			transport: "native",
			need:      []capability.Kind{capability.InProcess},
			score:     scoreInProcess,
			extra: func(in Input) (bool, string) {
				if err := in.Policy.AllowNative(); err != nil {
					return false, err.Error()
				}
				return true, ""
			},
		},
		{
			strategy:  bridge.StrategyUnixSocket,
			transport: "unix_socket",
			need:      []capability.Kind{capability.UnixSocket, capability.Process},
			score:     scoreUnix,
			extra: func(in Input) (bool, string) {
				if !unixOK {
					return false, "unix sockets are not available on this host"
				}
				return policyTransport(in, "unix_socket")
			},
		},
		{
			strategy:  bridge.StrategyNamedPipe,
			transport: "named_pipe",
			need:      []capability.Kind{capability.NamedPipe, capability.Process},
			score:     scorePipe,
			extra: func(in Input) (bool, string) {
				if in.HostGOOS != "windows" {
					return false, "named pipes are a Windows transport"
				}
				return policyTransport(in, "named_pipe")
			},
		},
		{
			strategy:  bridge.StrategyStdio,
			transport: "stdio",
			need:      []capability.Kind{capability.Stdio, capability.Process},
			score:     scoreStdio,
			extra: func(in Input) (bool, string) {
				if err := in.Policy.AllowSubprocess(); err != nil {
					return false, err.Error()
				}
				return policyTransport(in, "stdio")
			},
		},
		{
			strategy:  bridge.StrategyTCP,
			transport: "tcp",
			need:      []capability.Kind{capability.TCP, capability.Process},
			score:     scoreTCP,
			extra: func(in Input) (bool, string) {
				if in.Policy.NetworkMode() == "none" {
					return false, "network disabled by policy"
				}
				return policyTransport(in, "tcp")
			},
		},
		{
			strategy:  bridge.StrategyWASM,
			transport: "wasm",
			need:      []capability.Kind{capability.WASM},
			score:     scoreWASM,
			extra: func(in Input) (bool, string) {
				return policyTransport(in, "wasm")
			},
		},
		{
			strategy:  bridge.StrategySharedMemory,
			transport: "shared_memory",
			need:      []capability.Kind{capability.SharedMemory},
			score:     scoreSHM,
			extra: func(Input) (bool, string) {
				return false, "shared memory is experimental and disabled by default"
			},
		},
	}

	out := make([]bridge.Candidate, 0, len(specs))
	for _, sp := range specs {
		c := bridge.Candidate{Plan: bridge.Plan{
			Strategy:  sp.strategy,
			Adapter:   in.Adapter,
			Transport: sp.transport,
			Runtime:   in.Runtime,
			Language:  in.Language,
			Scores:    sp.score(in),
			Attrs:     map[string]string{},
		}}
		ok, why := capabilitiesPresent(in.Graph, sp.need)
		if !ok {
			c.Viable = false
			c.Reject = why
			out = append(out, c)
			continue
		}
		if sp.extra != nil {
			if ok, why = sp.extra(in); !ok {
				c.Viable = false
				c.Reject = why
				out = append(out, c)
				continue
			}
		}
		c.Plan.Reasons = reasons(in, sp.strategy)
		if pref := preferBoost(in.Prefer, sp.strategy, sp.transport); pref > 0 {
			c.Plan.Scores.Overall = clamp(c.Plan.Scores.Overall + pref)
			c.Plan.Reasons = append(c.Plan.Reasons, "manifest prefer list raised this strategy")
		}
		c.Viable = true
		out = append(out, c)
	}
	return out
}

func capabilitiesPresent(g capability.Graph, need []capability.Kind) (bool, string) {
	for _, k := range need {
		if !g.Has(k) {
			return false, "missing capability " + string(k)
		}
	}
	return true, ""
}

func policyTransport(in Input, name string) (bool, string) {
	if err := in.Policy.AllowTransport(name); err != nil {
		return false, err.Error()
	}
	if err := in.Policy.AllowSubprocess(); err != nil && name != "native" && name != "wasm" {
		return false, err.Error()
	}
	return true, ""
}

func preferBoost(prefer []string, strat bridge.Strategy, transport string) int {
	for i, p := range prefer {
		if strings.EqualFold(p, string(strat)) || strings.EqualFold(p, transport) {
			return 8 - i
		}
	}
	return 0
}

func overall(s bridge.Scores) int {
	sum := s.Performance*defaultWeights["performance"] +
		s.Reliability*defaultWeights["reliability"] +
		s.Isolation*defaultWeights["isolation"] +
		s.Startup*defaultWeights["startup"] +
		s.Serialization*defaultWeights["serialization"] +
		s.Compatibility*defaultWeights["compatibility"] +
		s.Security*defaultWeights["security"] +
		s.Portability*defaultWeights["portability"] +
		s.Debuggability*defaultWeights["debuggability"] +
		s.Availability*defaultWeights["availability"]
	return clamp(sum / 100)
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func scoreInProcess(Input) bridge.Scores {
	s := bridge.Scores{Performance: 99, Reliability: 90, Isolation: 20, Startup: 99, Serialization: 99, Compatibility: 80, Security: 40, Portability: 70, Debuggability: 70, Availability: 90}
	s.Overall = overall(s)
	return s
}

func scoreUnix(Input) bridge.Scores {
	s := bridge.Scores{Performance: 88, Reliability: 92, Isolation: 97, Startup: 70, Serialization: 80, Compatibility: 99, Security: 88, Portability: 60, Debuggability: 85, Availability: 90}
	s.Overall = overall(s)
	return s
}

func scorePipe(Input) bridge.Scores {
	s := bridge.Scores{Performance: 86, Reliability: 90, Isolation: 96, Startup: 68, Serialization: 80, Compatibility: 95, Security: 86, Portability: 40, Debuggability: 80, Availability: 88}
	s.Overall = overall(s)
	return s
}

func scoreStdio(Input) bridge.Scores {
	s := bridge.Scores{Performance: 61, Reliability: 88, Isolation: 94, Startup: 75, Serialization: 70, Compatibility: 97, Security: 84, Portability: 99, Debuggability: 90, Availability: 99}
	s.Overall = overall(s)
	return s
}

func scoreTCP(Input) bridge.Scores {
	s := bridge.Scores{Performance: 78, Reliability: 85, Isolation: 90, Startup: 65, Serialization: 78, Compatibility: 92, Security: 70, Portability: 95, Debuggability: 88, Availability: 85}
	s.Overall = overall(s)
	return s
}

func scoreWASM(Input) bridge.Scores {
	s := bridge.Scores{Performance: 82, Reliability: 80, Isolation: 99, Startup: 60, Serialization: 75, Compatibility: 70, Security: 95, Portability: 90, Debuggability: 60, Availability: 50}
	s.Overall = overall(s)
	return s
}

func scoreSHM(Input) bridge.Scores {
	s := bridge.Scores{Performance: 98, Reliability: 70, Isolation: 50, Startup: 55, Serialization: 95, Compatibility: 40, Security: 55, Portability: 40, Debuggability: 40, Availability: 20}
	s.Overall = overall(s)
	return s
}

func reasons(in Input, s bridge.Strategy) []string {
	switch s {
	case bridge.StrategyUnixSocket:
		return []string{
			"target supports persistent process execution",
			"native embedding would reduce crash isolation",
			"Unix sockets are available on this host",
			"expected request frequency favors persistent IPC",
			"serialization overhead remains acceptable",
		}
	case bridge.StrategyStdio:
		return []string{
			"stdio is available on every supported platform",
			"supervised subprocess isolates target crashes",
			"no additional network surface is required",
		}
	case bridge.StrategyInProcess:
		return []string{
			"target can be reached without leaving the host process",
			"serialization cost is avoided",
		}
	case bridge.StrategyTCP:
		return []string{
			"localhost TCP is available when domain sockets are not",
			"policy permits loopback networking",
		}
	case bridge.StrategyNamedPipe:
		return []string{
			"Windows named pipes provide local IPC isolation",
			"target supports persistent process execution",
		}
	case bridge.StrategyWASM:
		return []string{
			"target is a WebAssembly module",
			"WASM provides a strong isolation boundary",
		}
	default:
		return []string{fmt.Sprintf("strategy %s is compatible with %s", s, in.Language)}
	}
}

// Explain renders a human-readable planning report.
func Explain(in Input, res *Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Detected runtime:\n%s\n", displayRuntime(in))
	fmt.Fprintf(&b, "Selected bridge:\n%s\n", displayStrategy(res.Selected.Strategy))
	b.WriteString("Why:\n")
	for _, r := range res.Selected.Reasons {
		fmt.Fprintf(&b, "- %s\n", r)
	}
	if len(res.Selected.Fallbacks) > 0 {
		fmt.Fprintf(&b, "Fallback:\n%s\n", displayStrategy(res.Selected.Fallbacks[0]))
	}
	b.WriteString("\nCandidate Bridges\n")
	for _, c := range res.Candidates {
		if !c.Viable {
			continue
		}
		fmt.Fprintf(&b, "%s\n", displayStrategy(c.Plan.Strategy))
		fmt.Fprintf(&b, "Performance:     %d\n", c.Plan.Scores.Performance)
		fmt.Fprintf(&b, "Isolation:       %d\n", c.Plan.Scores.Isolation)
		fmt.Fprintf(&b, "Compatibility:   %d\n", c.Plan.Scores.Compatibility)
		fmt.Fprintf(&b, "Overall:         %d\n\n", c.Plan.Scores.Overall)
	}
	return b.String()
}

func displayRuntime(in Input) string {
	if in.Runtime != "" {
		return in.Runtime
	}
	return in.Language
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
		return string(s)
	}
}
