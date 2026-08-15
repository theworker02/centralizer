// Package capability models what a discovered target can actually do,
// rather than reducing it to a single language label.
package capability

import "strings"

// Kind names a communication or execution capability.
type Kind string

const (
	NativeLibrary   Kind = "native_library"
	SharedLibrary   Kind = "shared_library"
	CLI             Kind = "cli"
	RPC             Kind = "rpc"
	HTTP            Kind = "http"
	WebSocket       Kind = "websocket"
	Stdio           Kind = "stdio"
	UnixSocket      Kind = "unix_socket"
	NamedPipe       Kind = "named_pipe"
	WASM            Kind = "wasm"
	EmbeddedRuntime Kind = "embedded_runtime"
	DynamicLibrary  Kind = "dynamic_library"
	GeneratedShim   Kind = "generated_shim"
	Streaming       Kind = "streaming"
	SharedMemory    Kind = "shared_memory"
	Serialization   Kind = "serialization"
	RuntimeVersion  Kind = "runtime_version"
	Threading       Kind = "threading"
	Process         Kind = "process"
	InProcess       Kind = "in_process"
	TCP             Kind = "tcp"
	SchemaExport    Kind = "schema_export"
	Handles         Kind = "handles"
	Events          Kind = "events"
)

// Capability is one node or edge attribute in the capability graph.
type Capability struct {
	Kind       Kind              `json:"kind"`
	Available  bool              `json:"available"`
	Confidence float64           `json:"confidence"`
	Detail     string            `json:"detail,omitempty"`
	Attrs      map[string]string `json:"attrs,omitempty"`
}

// Graph is a directed capability graph for a single target.
type Graph struct {
	Target       string       `json:"target"`
	Language     string       `json:"language"`
	Runtime      string       `json:"runtime"`
	Capabilities []Capability `json:"capabilities"`
	Edges        []Edge       `json:"edges,omitempty"`
}

// Edge links two capabilities (for example stdio enables RPC).
type Edge struct {
	From   Kind   `json:"from"`
	To     Kind   `json:"to"`
	Reason string `json:"reason,omitempty"`
}

// Has reports whether a capability kind is available.
func (g Graph) Has(k Kind) bool {
	for _, c := range g.Capabilities {
		if c.Kind == k && c.Available {
			return true
		}
	}
	return false
}

// Get returns the first capability of kind k.
func (g Graph) Get(k Kind) (Capability, bool) {
	for _, c := range g.Capabilities {
		if c.Kind == k {
			return c, true
		}
	}
	return Capability{}, false
}

// Add appends a capability if it is not already present.
func (g *Graph) Add(c Capability) {
	for i, existing := range g.Capabilities {
		if existing.Kind == c.Kind {
			if c.Confidence > existing.Confidence {
				g.Capabilities[i] = c
			}
			return
		}
	}
	g.Capabilities = append(g.Capabilities, c)
}

// Summary is a stable, human-readable listing used by explain/inspect.
func (g Graph) Summary() string {
	var b strings.Builder
	b.WriteString("capabilities:")
	for _, c := range g.Capabilities {
		b.WriteString("\n  - ")
		b.WriteString(string(c.Kind))
		if c.Available {
			b.WriteString(" available")
		} else {
			b.WriteString(" unavailable")
		}
		if c.Detail != "" {
			b.WriteString(" (")
			b.WriteString(c.Detail)
			b.WriteString(")")
		}
	}
	return b.String()
}
