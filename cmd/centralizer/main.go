package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/theworker02/centralizer/internal/discovery"
	"github.com/theworker02/centralizer/internal/telemetry"
	"github.com/theworker02/centralizer/internal/version"
	"github.com/theworker02/centralizer/pkg/centralizer"
	"github.com/theworker02/centralizer/pkg/diagnostics"
	"github.com/theworker02/centralizer/pkg/lockfile"
	"github.com/theworker02/centralizer/pkg/manifest"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags, rest := parseGlobal(args)
	telemetry.Configure(flags.JSON && flags.Verbose, telemetry.LevelFromFlags(flags.Verbose, flags.Quiet))
	if len(rest) == 0 {
		usage(os.Stderr)
		return 2
	}
	cmd, rest := rest[0], rest[1:]
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var err error
	switch cmd {
	case "version", "-v", "--version":
		printVersion(flags)
	case "help", "-h", "--help":
		usage(os.Stdout)
	case "detect":
		err = cmdDetect(ctx, flags, rest)
	case "inspect":
		err = cmdInspect(ctx, flags, rest)
	case "describe":
		err = cmdDescribe(ctx, flags, rest)
	case "connect":
		err = cmdConnect(ctx, flags, rest)
	case "call":
		err = cmdCall(ctx, flags, rest)
	case "health":
		err = cmdHealth(ctx, flags, rest)
	case "list":
		err = cmdList(ctx, flags)
	case "graph":
		err = cmdGraph(ctx, flags, rest)
	case "explain":
		err = cmdExplain(ctx, flags, rest)
	case "bench":
		err = cmdBench(ctx, flags, rest)
	case "trace":
		err = cmdTrace(ctx, flags, rest)
	case "doctor":
		err = cmdDoctor(flags)
	case "cache":
		err = cmdCache(flags, rest)
	case "init":
		err = cmdInit(flags, rest)
	case "lock":
		err = cmdLock(ctx, flags, rest)
	case "adapters":
		err = cmdAdapters(flags)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage(os.Stderr)
		return 2
	}
	if err != nil {
		if !flags.Quiet {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		return 1
	}
	return 0
}

type global struct {
	JSON    bool
	Quiet   bool
	Verbose bool
}

func parseGlobal(args []string) (global, []string) {
	var g global
	var rest []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			g.JSON = true
		case "--quiet", "-q":
			g.Quiet = true
		case "--verbose", "-v":
			g.Verbose = true
		default:
			rest = append(rest, args[i])
		}
	}
	return g, rest
}

func usage(w io.Writer) {
	fmt.Fprint(w, `centralizer — One runtime. Every language.

Usage:
  centralizer <command> [target] [flags]

Commands:
  detect     score language/runtime hypotheses
  inspect    emit the resolved plan as text or JSON
  describe   print the inferred or declared schema
  connect    establish a bridge and print health
  call       invoke a function on a connected target
  health     show bridge health
  list       list adapters and (when connected) services
  graph      print the capability graph
  explain    explain bridge selection
  bench      measure available strategies (does not override policy)
  trace      record discovery → plan → connect spans
  doctor     inspect host runtimes and permissions
  cache      list|clear generated artifacts
  init       write a starter centralizer.yaml
  lock       write a resolved-plan centralizer.lock
  adapters   list adapters with tier and capabilities
  version    print version

Global flags:
  --json       machine-readable output
  --quiet, -q  errors only
  --verbose    debug logs
`)
}

func printVersion(g global) {
	if g.JSON {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]string{
			"version":  version.Version,
			"protocol": version.Protocol,
		})
		return
	}
	fmt.Printf("centralizer %s (protocol %s)\n", version.Version, version.Protocol)
}

func requireTarget(rest []string) (string, error) {
	if len(rest) < 1 {
		return "", fmt.Errorf("target required")
	}
	return rest[0], nil
}

func emit(g global, v any, text string) error {
	if g.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	if !g.Quiet {
		fmt.Fprint(os.Stdout, text)
		if !strings.HasSuffix(text, "\n") {
			fmt.Fprintln(os.Stdout)
		}
	}
	return nil
}

func cmdDetect(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New()
	res, err := hub.Analyze(ctx, ref)
	if err != nil {
		return err
	}
	return emit(g, res, discovery.FormatAnalysis(res))
}

func cmdInspect(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New(centralizer.WithTracing(true))
	text, plan, err := hub.Explain(ctx, ref)
	if err != nil {
		return err
	}
	analysis, err := hub.Analyze(ctx, ref)
	if err != nil {
		return err
	}
	out := map[string]any{"analysis": analysis, "plan": plan.Selected, "candidates": plan.Candidates}
	if lf, ok := findLock(ref); ok {
		out["lock"] = lf
		out["lock_matches"] = lf.Matches(plan.Selected)
	}
	return emit(g, out, text)
}

func findLock(target string) (lockfile.File, bool) {
	for _, p := range []string{lockfile.Name, target + ".lock"} {
		if f, err := lockfile.Read(p); err == nil {
			return f, true
		}
	}
	return lockfile.File{}, false
}

func cmdDescribe(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New()
	svc, err := hub.Connect(ctx, ref)
	if err != nil {
		return err
	}
	defer svc.Close(ctx)
	sc, err := svc.Describe(ctx)
	if err != nil {
		return err
	}
	return emit(g, sc, fmt.Sprintf("service: %s\nfunctions: %d\n", sc.Service, len(sc.Functions)))
}

func cmdConnect(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New()
	svc, err := hub.Connect(ctx, ref)
	if err != nil {
		return err
	}
	defer svc.Close(ctx)
	h := svc.Health()
	return emit(g, h, h.Text())
}

func cmdCall(ctx context.Context, g global, rest []string) error {
	if len(rest) < 2 {
		return fmt.Errorf("usage: centralizer call <target> <function> [k=v...]")
	}
	hub := centralizer.New()
	svc, err := hub.Connect(ctx, rest[0])
	if err != nil {
		return err
	}
	defer svc.Close(ctx)
	args := centralizer.Args{}
	for _, kv := range rest[2:] {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return fmt.Errorf("argument %q must be key=value", kv)
		}
		args[k] = coerce(v)
	}
	result, err := svc.Call(ctx, rest[1], args)
	if err != nil {
		return err
	}
	native, err := result.Native()
	if err != nil {
		native = result.String()
	}
	return emit(g, map[string]any{"result": native, "kind": result.Kind().String()}, fmt.Sprint(native))
}

func coerce(s string) any {
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	var i int64
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil && !strings.Contains(s, ".") {
		return i
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
		return f
	}
	return s
}

func cmdHealth(ctx context.Context, g global, rest []string) error {
	if len(rest) == 0 {
		return emit(g, map[string]any{"services": []any{}}, "No connected services in this process.\nUse: centralizer connect <target>\n")
	}
	hub := centralizer.New()
	svc, err := hub.Connect(ctx, rest[0])
	if err != nil {
		return err
	}
	defer svc.Close(ctx)
	h := svc.Health()
	return emit(g, h, h.Text())
}

func cmdList(ctx context.Context, g global) error {
	hub := centralizer.New()
	names := hub.Adapters()
	return emit(g, map[string]any{"adapters": names}, "adapters:\n  "+strings.Join(names, "\n  ")+"\n")
}

func cmdGraph(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New()
	res, err := hub.Analyze(ctx, ref)
	if err != nil {
		return err
	}
	return emit(g, res.Graph, res.Graph.Summary()+"\n")
}

func cmdExplain(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New()
	text, plan, err := hub.Explain(ctx, ref)
	if err != nil {
		return err
	}
	return emit(g, plan, text)
}

func cmdBench(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New()
	_, plan, err := hub.Explain(ctx, ref)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("Testing available bridge strategies...\n")
	b.WriteString("(Planning scores only; this command does not override security policy.)\n\n")
	type row struct {
		Strategy string `json:"strategy"`
		Overall  int    `json:"overall"`
		Viable   bool   `json:"viable"`
	}
	var rows []row
	for _, c := range plan.Candidates {
		if !c.Viable {
			continue
		}
		fmt.Fprintf(&b, "%s:\nscore %d / 100\n\n", c.Plan.Strategy, c.Plan.Scores.Overall)
		rows = append(rows, row{Strategy: string(c.Plan.Strategy), Overall: c.Plan.Scores.Overall, Viable: true})
	}
	fmt.Fprintf(&b, "Selected default:\n%s\n", plan.Selected.Strategy)
	return emit(g, map[string]any{"selected": plan.Selected, "rows": rows}, b.String())
}

func cmdTrace(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	hub := centralizer.New(centralizer.WithTracing(true))
	start := time.Now()
	svc, err := hub.Connect(ctx, ref)
	if err != nil {
		return err
	}
	defer svc.Close(ctx)
	text := fmt.Sprintf("trace %s\n  discovery+plan+startup: %s\n  adapter: %s\n  transport: %s\n",
		ref, time.Since(start).Round(time.Millisecond), svc.Plan().Adapter, svc.Transport())
	return emit(g, map[string]any{
		"target":    ref,
		"elapsed":   time.Since(start).String(),
		"adapter":   svc.Plan().Adapter,
		"transport": svc.Transport(),
		"plan":      svc.Plan(),
	}, text)
}

func cmdDoctor(g global) error {
	hub := centralizer.New()
	rep := diagnostics.Run(hub.Adapters())
	return emit(g, rep, rep.Text())
}

func cmdInit(g global, rest []string) error {
	path := manifest.DefaultPath
	force := false
	for _, a := range rest {
		if a == "--force" {
			force = true
			continue
		}
		path = a
	}
	if err := manifest.WriteStarter(path, force); err != nil {
		return err
	}
	return emit(g, map[string]string{"path": path, "status": "written"}, "wrote "+path+"\n")
}

func cmdLock(ctx context.Context, g global, rest []string) error {
	ref, err := requireTarget(rest)
	if err != nil {
		return err
	}
	outPath := lockfile.Name
	if len(rest) > 1 {
		outPath = rest[1]
	}
	hub := centralizer.New()
	lf, err := hub.LockPlan(ctx, ref)
	if err != nil {
		return err
	}
	if err := lockfile.Write(outPath, lf); err != nil {
		return err
	}
	text := fmt.Sprintf("wrote %s\nadapter=%s transport=%s strategy=%s score=%d\n",
		outPath, lf.Adapter, lf.Transport, lf.Strategy, lf.Scores.Overall)
	return emit(g, lf, text)
}

func cmdAdapters(g global) error {
	hub := centralizer.New()
	cat := hub.AdapterCatalog()
	var b strings.Builder
	b.WriteString("adapter          tier  call  capabilities\n")
	for _, info := range cat {
		call := "no"
		if info.Invocation {
			call = "yes"
		}
		fmt.Fprintf(&b, "%-16s %d     %-3s  %s\n", info.Name, info.Tier, call, strings.Join(info.Capabilities, ","))
	}
	return emit(g, cat, b.String())
}

func cmdCache(g global, rest []string) error {
	hub := centralizer.New()
	store := hub.Cache()
	sub := "list"
	if len(rest) > 0 {
		sub = rest[0]
	}
	switch sub {
	case "list":
		items, err := store.List("")
		if err != nil {
			return err
		}
		return emit(g, items, strings.Join(items, "\n")+"\n")
	case "clear":
		if err := store.Clear(""); err != nil {
			return err
		}
		return emit(g, map[string]string{"status": "cleared"}, "cache cleared\n")
	default:
		return fmt.Errorf("usage: centralizer cache list|clear")
	}
}
