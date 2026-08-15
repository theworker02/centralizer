package lifecycle

import "testing"

func FuzzHandleIDs(f *testing.F) {
	f.Add("h-1")
	f.Add("")
	f.Add("../../../etc/passwd")
	f.Fuzz(func(t *testing.T, id string) {
		tab := NewTable()
		tab.Put(Handle{ID: id, BridgeID: "b"})
		_, _ = tab.Get(id)
		_ = tab.Release(id)
	})
}
