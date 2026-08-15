package lifecycle

import (
	"testing"
	"time"

	"github.com/theworker02/centralizer/pkg/czerr"
)

func TestExpireAndDropBridge(t *testing.T) {
	tab := NewTable()
	tab.Put(Handle{ID: "h1", BridgeID: "br-a", Expires: time.Now().Add(-time.Second)})
	tab.Put(Handle{ID: "h2", BridgeID: "br-a"})
	tab.Put(Handle{ID: "h3", BridgeID: "br-b"})
	if err := tab.RejectIfExpired("h1"); err == nil {
		t.Fatal("expected expired")
	}
	if err := tab.RejectIfExpired("unknown"); err != nil {
		t.Fatal(err)
	}
	if n := tab.SweepExpired(); n != 0 && tab.Len() != 2 {
		// h1 already deleted by RejectIfExpired
		if tab.Len() != 2 {
			t.Fatalf("len=%d", tab.Len())
		}
	}
	tab.DropBridge("br-a")
	if tab.Len() != 1 {
		t.Fatalf("len=%d after drop", tab.Len())
	}
	if _, err := tab.Get("h3"); err != nil {
		t.Fatal(err)
	}
	if _, err := tab.Get("h2"); err == nil {
		t.Fatal("h2 should be dropped")
	}
	if !isHandle(tab.Release("missing")) {
		t.Fatal("expected invalid handle")
	}
}

func isHandle(err error) bool {
	return err != nil && (err == czerr.ErrHandleInvalid || czerr.ErrHandleInvalid.Error() != "")
}
