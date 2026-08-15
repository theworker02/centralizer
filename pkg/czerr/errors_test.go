package czerr

import (
	"errors"
	"testing"
)

func TestWrapIs(t *testing.T) {
	err := Wrap(ErrTimeout, "call calculate", errors.New("deadline"))
	if !errors.Is(err, ErrTimeout) {
		t.Fatal("expected ErrTimeout")
	}
	if errors.Is(err, ErrCancelled) {
		t.Fatal("unexpected cancel")
	}
}

func TestDetail(t *testing.T) {
	err := WithDetail(New(ErrTargetNotFound, "missing"), map[string]string{"path": "./x"})
	d := Detail(err)
	if d["path"] != "./x" {
		t.Fatalf("detail=%v", d)
	}
}
