package cir

import (
	"encoding/json"
	"testing"
)

func FuzzDecodeWire(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{"k":"null"}`),
		[]byte(`{"k":"boolean","b":true}`),
		[]byte(`{"k":"int","i":-1}`),
		[]byte(`{"k":"string","s":"hi"}`),
		[]byte(`{"k":"array","a":[{"k":"int","i":1}]}`),
		[]byte(`{"k":"map","m":[{"k":"a","v":{"k":"null"}}]}`),
		[]byte(`{"k":"bytes","x":"AQID"}`),
		[]byte(`not json`),
		[]byte(`{}`),
		[]byte(`{"k":"uuid","x":""}`),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var w Wire
		if err := json.Unmarshal(data, &w); err != nil {
			return
		}
		_, _ = FromWire(w)
	})
}
