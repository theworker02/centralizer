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

	target := "./engine"
	if _, err := os.Stat(target); err != nil {
		target = filepath.Join("examples", "go-rust", "engine")
	}
	svc, err := hub.Connect(ctx, target)
	if err != nil {
		panic(err)
	}
	defer svc.Close(ctx)

	result, err := svc.Call(ctx, "multiply", centralizer.Args{"value": 21})
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
