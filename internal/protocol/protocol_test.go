package protocol

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/theworker02/centralizer/pkg/cir"
)

func TestNDJSONRoundTrip(t *testing.T) {
	m, err := NewMessage("1", TypeCall, CallPayload{
		Function: "calculate",
		Args:     EncodeArgs(map[string]cir.Value{"value": cir.Int(42)}),
	})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteNDJSON(&buf, m); err != nil {
		t.Fatal(err)
	}
	got, err := ReadNDJSON(bufio.NewReader(&buf))
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeCall || got.ID != "1" {
		t.Fatalf("%+v", got)
	}
	var p CallPayload
	if err := DecodePayload(got, &p); err != nil {
		t.Fatal(err)
	}
	args, err := DecodeArgs(p.Args)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := args["value"]
	if !ok {
		t.Fatal("missing value")
	}
	i, err := v.Int()
	if err != nil || i != 42 {
		t.Fatalf("%v %v", i, err)
	}
}

func TestFrameRoundTrip(t *testing.T) {
	m, err := NewMessage("9", TypeHello, HelloPayload{Protocol: Version, Name: Name})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteFrame(&buf, m, 0); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFrame(&buf, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeHello {
		t.Fatalf("%+v", got)
	}
}

func TestFrameRejectsHuge(t *testing.T) {
	var buf bytes.Buffer
	// 4-byte length claiming 32MiB
	buf.Write([]byte{0x02, 0x00, 0x00, 0x00})
	if _, err := ReadFrame(&buf, 1024); err == nil {
		t.Fatal("expected too large")
	}
}

func TestCompatible(t *testing.T) {
	if err := Compatible("1.2"); err != nil {
		t.Fatal(err)
	}
	if err := Compatible("2.0"); err == nil {
		t.Fatal("expected mismatch")
	}
}
