// Polyglot hello: one Go host calls Python and Node through a single Hub.
//
// Requires python3 and node on PATH. No Rust/Cargo needed.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/theworker02/centralizer/pkg/centralizer"
)

func main() {
	ctx := context.Background()
	hub := centralizer.New()
	defer hub.Close(ctx)

	root := findRoot()
	pyTarget := filepath.Join(root, "examples", "go-python", "analytics")
	jsTarget := filepath.Join(root, "examples", "go-node", "reporter")

	fmt.Println("=== explain (evaluator path) ===")
	for _, target := range []string{pyTarget, jsTarget} {
		rep, err := hub.ExplainReport(ctx, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "explain %s: %v\n", target, err)
			continue
		}
		call := "no"
		if rep.CallImplemented {
			call = "yes"
		}
		fmt.Printf("%s\n  adapter=%s score=%.2f Call=%s strategy=%s\n",
			target, rep.Adapter, rep.DetectionScore, call, rep.SelectedStrategy)
	}

	fmt.Println("\n=== call ===")
	analytics, err := hub.Connect(ctx, pyTarget)
	if err != nil {
		fatal("python connect", err)
	}
	defer analytics.Close(ctx)
	py, err := analytics.Call(ctx, "calculate", centralizer.Args{"value": 21})
	if err != nil {
		fatal("python call", err)
	}
	fmt.Println("python calculate(21) =", py)

	reporter, err := hub.Connect(ctx, jsTarget)
	if err != nil {
		fatal("node connect", err)
	}
	defer reporter.Close(ctx)
	js, err := reporter.Call(ctx, "report", centralizer.Args{"value": 21})
	if err != nil {
		fatal("node call", err)
	}
	fmt.Println("node report(21) =", js)
}

func fatal(step string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", step, err)
	os.Exit(1)
}

func findRoot() string {
	wd, _ := os.Getwd()
	for dir := wd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
	}
	return wd
}
