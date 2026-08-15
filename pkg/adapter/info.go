package adapter

import (
	"context"
	"sort"
)

// Info is a catalog entry used by `centralizer adapters`.
type Info struct {
	Name         string   `json:"name"`
	Language     string   `json:"language"`
	Tier         int      `json:"tier"`
	Invocation   bool     `json:"invocation"`
	Capabilities []string `json:"capabilities,omitempty"`
	Notes        string   `json:"notes,omitempty"`
}

type staticInfo struct {
	Language   string
	Tier       int
	Invocation bool
	Notes      string
}

var known = map[string]staticInfo{
	"go":     {Language: "Go", Tier: 1, Invocation: true, Notes: "in-process handlers and protocol-speaking binaries"},
	"python": {Language: "Python", Tier: 1, Invocation: true, Notes: "generated stdio/tcp shim"},
	"node":   {Language: "JavaScript", Tier: 1, Invocation: true, Notes: "generated stdio/tcp shim"},
	"rust":   {Language: "Rust", Tier: 1, Invocation: true, Notes: "stdio to protocol-speaking cargo/binaries"},
	"c":      {Language: "C", Tier: 1, Invocation: false, Notes: "detect-only"},
	"cpp":    {Language: "C++", Tier: 1, Invocation: false, Notes: "detect-only"},
	"wasm":   {Language: "WebAssembly", Tier: 1, Invocation: false, Notes: "detect-only"},
	"jvm":    {Language: "Java", Tier: 2, Invocation: false, Notes: "detect-only"},
	"dotnet": {Language: "C#", Tier: 2, Invocation: false, Notes: "detect-only"},
	"ruby":   {Language: "Ruby", Tier: 2, Invocation: false, Notes: "detect-only"},
	"php":    {Language: "PHP", Tier: 2, Invocation: false, Notes: "detect-only"},
	"swift":  {Language: "Swift", Tier: 3, Invocation: false, Notes: "detect-only"},
	"dart":   {Language: "Dart", Tier: 3, Invocation: false, Notes: "detect-only"},
	"lua":    {Language: "Lua", Tier: 3, Invocation: false, Notes: "detect-only"},
	"zig":    {Language: "Zig", Tier: 3, Invocation: false, Notes: "detect-only"},
}

// Catalog lists registered adapters with tier and claimed capabilities.
func Catalog(r *Registry) []Info {
	if r == nil {
		return nil
	}
	var out []Info
	for _, a := range r.All() {
		info := Info{Name: a.Name(), Tier: 3, Notes: "third-party adapter"}
		if s, ok := known[a.Name()]; ok {
			info.Language = s.Language
			info.Tier = s.Tier
			info.Invocation = s.Invocation
			info.Notes = s.Notes
		}
		caps, err := a.Capabilities(context.Background(), Target{})
		if err == nil {
			for _, c := range caps {
				if c.Available {
					info.Capabilities = append(info.Capabilities, string(c.Kind))
				}
			}
			sort.Strings(info.Capabilities)
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return out[i].Tier < out[j].Tier
		}
		return out[i].Name < out[j].Name
	})
	return out
}
