package manifest

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte("centralizer:\n  version: 1\nservices:\n  a:\n    source: ./a\n"))
	f.Add([]byte(":::"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = Parse(data)
	})
}
