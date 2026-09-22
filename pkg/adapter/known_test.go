package adapter

import "testing"

func TestKnown(t *testing.T) {
	info, ok := Known("python")
	if !ok || !info.Invocation || info.Tier != 1 {
		t.Fatalf("%+v ok=%v", info, ok)
	}
	info, ok = Known("lua")
	if !ok || info.Invocation {
		t.Fatalf("lua must be detect-only: %+v", info)
	}
	if _, ok := Known("nope"); ok {
		t.Fatal("expected miss")
	}
}
