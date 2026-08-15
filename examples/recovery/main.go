package main

import (
	"context"
	"fmt"

	"github.com/theworker02/centralizer/internal/session"
	"github.com/theworker02/centralizer/pkg/centralizer"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
)

func main() {
	ctx := context.Background()
	hub := centralizer.New(centralizer.WithAutoRecovery(true))
	fails := 0
	hub.RegisterNative(&session.Handler{
		Name: "flaky",
		Funcs: map[string]session.Func{
			"once": func(context.Context, map[string]cir.Value) (cir.Value, error) {
				fails++
				if fails == 1 {
					return cir.Value{}, czerr.New(czerr.ErrBridgeFailed, "simulated crash")
				}
				return cir.String("recovered"), nil
			},
		},
	})
	svc, err := hub.Connect(ctx, "native:flaky")
	if err != nil {
		panic(err)
	}
	defer svc.Close(ctx)
	// Native handlers do not tear down a child process; this example
	// shows the Call API and health surface. Process recovery is
	// exercised in tests/failure.
	v, err := svc.Call(ctx, "once", nil)
	fmt.Println(v, err)
	fmt.Println(svc.Health().Text())
}
