package benchmarks

import (
	"encoding/json"
	"testing"

	"github.com/theworker02/centralizer/pkg/cir"
)

func BenchmarkCIREncode(b *testing.B) {
	v := cir.MustMap("value", cir.Int(42), "ok", cir.Bool(true), "name", cir.String("n"))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCIRFromNative(b *testing.B) {
	in := map[string]any{"value": 42, "ok": true, "name": "n"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := cir.From(in); err != nil {
			b.Fatal(err)
		}
	}
}
