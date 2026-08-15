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
	target := "."
	if _, err := os.Stat("model.py"); err != nil {
		target = filepath.Join("examples", "object-handles")
	}
	svc, err := hub.Connect(ctx, target, centralizer.WithEntry("model.py"))
	if err != nil {
		panic(err)
	}
	defer svc.Close(ctx)

	handle, err := svc.New(ctx, "MachineLearningModel", centralizer.Args{"name": "demo"})
	if err != nil {
		panic(err)
	}
	id, err := handle.HandleID()
	if err != nil {
		panic(err)
	}
	fmt.Println("handle", id)
	if err := svc.Release(ctx, id); err != nil {
		panic(err)
	}
}
