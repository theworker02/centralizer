// Package bridge defines connected runtime sessions and planner output.
package bridge

import (
	"context"
	"io"

	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Strategy is a communication approach the planner may select.
type Strategy string

const (
	StrategyInProcess     Strategy = "in_process"
	StrategyNativeABI     Strategy = "native_abi"
	StrategyDynamicLib    Strategy = "dynamic_library"
	StrategyEmbedded      Strategy = "embedded_runtime"
	StrategySharedMemory  Strategy = "shared_memory"
	StrategyUnixSocket    Strategy = "unix_socket"
	StrategyNamedPipe     Strategy = "named_pipe"
	StrategyStdio         Strategy = "stdio"
	StrategyTCP           Strategy = "tcp"
	StrategyRPC           Strategy = "rpc"
	StrategyHTTP          Strategy = "http"
	StrategyWebSocket     Strategy = "websocket"
	StrategyWASM          Strategy = "wasm"
	StrategyGeneratedShim Strategy = "generated_shim"
	StrategySubprocess    Strategy = "supervised_subprocess"
)

// TransportName returns the planner transport label for a strategy.
// Strategy identifiers and transport names are not always identical
// (in_process maps to native). Recovery must use this mapping rather
// than string(strategy), or a fallback reconnect selects the wrong pipe.
func TransportName(s Strategy) string {
	switch s {
	case StrategyInProcess:
		return "native"
	case StrategyUnixSocket:
		return "unix_socket"
	case StrategyNamedPipe:
		return "named_pipe"
	case StrategyStdio:
		return "stdio"
	case StrategyTCP:
		return "tcp"
	case StrategyWASM:
		return "wasm"
	case StrategySharedMemory:
		return "shared_memory"
	default:
		return string(s)
	}
}

// Plan is a scored, deterministic bridge selection.
type Plan struct {
	Strategy    Strategy          `json:"strategy"`
	Adapter     string            `json:"adapter"`
	Transport   string            `json:"transport"`
	Runtime     string            `json:"runtime"`
	Language    string            `json:"language"`
	Scores      Scores            `json:"scores"`
	Reasons     []string          `json:"reasons"`
	Fallbacks   []Strategy        `json:"fallbacks,omitempty"`
	Attrs       map[string]string `json:"attrs,omitempty"`
	Fingerprint string            `json:"fingerprint,omitempty"`
}

// Scores are planner dimensions in 0–100.
type Scores struct {
	Performance   int `json:"performance"`
	Reliability   int `json:"reliability"`
	Isolation     int `json:"isolation"`
	Startup       int `json:"startup"`
	Serialization int `json:"serialization"`
	Compatibility int `json:"compatibility"`
	Security      int `json:"security"`
	Portability   int `json:"portability"`
	Debuggability int `json:"debuggability"`
	Availability  int `json:"availability"`
	Overall       int `json:"overall"`
}

// Candidate is one evaluated plan before selection.
type Candidate struct {
	Plan   Plan
	Viable bool
	Reject string
}

// Bridge is a live session to a target runtime.
type Bridge interface {
	ID() string
	Plan() Plan
	Describe(ctx context.Context) (*schema.Schema, error)
	Call(ctx context.Context, fn string, args map[string]cir.Value) (cir.Value, error)
	Invoke(ctx context.Context, inv Invocation) (cir.Value, error)
	New(ctx context.Context, typeName string, args map[string]cir.Value) (cir.Value, error)
	Get(ctx context.Context, handle, property string) (cir.Value, error)
	Set(ctx context.Context, handle, property string, value cir.Value) error
	Release(ctx context.Context, handle string) error
	Stream(ctx context.Context, name string, args map[string]cir.Value) (Stream, error)
	Subscribe(ctx context.Context, event string) (Stream, error)
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

// Invocation is a fully specified call.
type Invocation struct {
	Function string
	Handle   string
	Method   string
	Args     map[string]cir.Value
}

// Stream is a cancellable value sequence.
type Stream interface {
	ID() string
	Values() <-chan cir.Value
	Err() error
	Close() error
}

// ReadCloser is used by transports that expose raw bytes.
type ReadCloser = io.ReadCloser
