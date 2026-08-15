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

	target := "./analytics"
	if _, err := os.Stat(target); err != nil {
		target = filepath.Join("examples", "go-python", "analytics")
	}
	analytics, err := hub.Connect(ctx, target)
	if err != nil {
		panic(err)
	}
	defer analytics.Close(ctx)

	result, err := analytics.Call(ctx, "calculate", centralizer.Args{"value": 42})
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
