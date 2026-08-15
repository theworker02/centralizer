package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/theworker02/centralizer/pkg/centralizer"
)

func main() {
	if _, err := exec.LookPath("python"); err != nil {
		if _, err := exec.LookPath("python3"); err != nil {
			fmt.Fprintln(os.Stderr, "python not installed; streaming example skipped")
			os.Exit(0)
		}
	}
	ctx := context.Background()
	hub := centralizer.New()
	defer hub.Close(ctx)

	target := filepath.Join("examples", "go-python", "analytics")
	if _, err := os.Stat(target); err != nil {
		target = filepath.Join("..", "go-python", "analytics")
	}
	svc, err := hub.Connect(ctx, target)
	if err != nil {
		panic(err)
	}
	defer svc.Close(ctx)

	st, err := svc.Stream(ctx, "count_up", centralizer.Args{"n": 4})
	if err != nil {
		panic(err)
	}
	defer st.Close()
	for v := range st.Values() {
		fmt.Println(v)
	}
	if err := st.Err(); err != nil {
		panic(err)
	}
}
