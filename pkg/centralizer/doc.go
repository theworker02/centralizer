// Package centralizer is the public Go API for the Centralizer
// interoperability runtime.
//
//	hub := centralizer.New()
//	service, err := hub.Connect(ctx, "./analytics")
//	result, err := service.Call(ctx, "calculate", centralizer.Args{"value": 42})
//
// Centralizer discovers a target, builds a capability graph, selects a
// bridge plan, supervises the connection, and converts values through CIR.
package centralizer
