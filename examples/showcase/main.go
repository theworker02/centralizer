// Showcase: a Go orchestrator connecting Python analytics, a Rust
// numerical engine, and a Node reporter through one Hub.
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
	hub := centralizer.New(centralizer.WithAutoRecovery(true), centralizer.WithTracing(true))
	defer hub.Close(ctx)

	root := findRoot()
	analytics, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-python", "analytics"))
	if err != nil {
		panic(err)
	}
	engine, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-rust", "engine"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "rust engine unavailable: %v\n", err)
	}
	reporter, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-node", "reporter"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "node reporter unavailable: %v\n", err)
	}

	v, err := analytics.Call(ctx, "calculate", centralizer.Args{"value": 21})
	if err != nil {
		panic(err)
	}
	fmt.Println("python", v)
	if engine != nil {
		ev, err := engine.Call(ctx, "multiply", centralizer.Args{"value": 21})
		if err != nil {
			fmt.Fprintf(os.Stderr, "rust call: %v\n", err)
		} else {
			fmt.Println("rust", ev)
		}
	}
	if reporter != nil {
		rv, err := reporter.Call(ctx, "report", centralizer.Args{"value": 21})
		if err != nil {
			fmt.Fprintf(os.Stderr, "node call: %v\n", err)
		} else {
			fmt.Println("node", rv)
		}
	}
	for _, h := range hub.Health() {
		fmt.Println("---")
		fmt.Print(h.Text())
	}
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
