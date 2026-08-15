package schema

import "testing"

func FuzzParseYAML(f *testing.F) {
	f.Add([]byte("service: x\nfunctions:\n  f:\n    args:\n      a: int\n"))
	f.Add([]byte("not: yaml: ["))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseYAML(data)
	})
}
