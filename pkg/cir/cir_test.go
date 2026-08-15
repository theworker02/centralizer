package cir

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestScalarRoundTrip(t *testing.T) {
	cases := []Value{
		Null(),
		Bool(true),
		Bool(false),
		Int(-42),
		Uint(42),
		Float(3.5),
		Decimal("1.25"),
		String("hello"),
		Bytes([]byte{1, 2, 3}),
		Timestamp(time.Unix(1_700_000_000, 0).UTC()),
		Duration(1500 * time.Millisecond),
		UUID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
		Handle("h-1"),
		ErrorValue("boom", "failed"),
		Optional(Int(1)),
		None(),
		ResultOK(String("ok")),
		ResultErr(String("no")),
	}
	for _, in := range cases {
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("marshal %s: %v", in.Kind(), err)
		}
		var out Value
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("unmarshal %s: %v (%s)", in.Kind(), err, data)
		}
		if !Equal(in, out) {
			t.Fatalf("round-trip mismatch for %s: %#v vs %#v (%s)", in.Kind(), in, out, data)
		}
	}
}

func TestMapStructArray(t *testing.T) {
	m := MustMap("a", Int(1), "b", String("x"))
	arr := Array(Int(1), Int(2), Int(3))
	st, err := Struct("Point", []string{"x", "y"}, []Value{Float(1), Float(2)})
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []Value{m, arr, st} {
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		var out Value
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if !Equal(in, out) {
			t.Fatalf("mismatch %s", in.Kind())
		}
	}
}

func TestFromNative(t *testing.T) {
	v, err := From(map[string]any{"n": 7, "ok": true})
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind() != KindMap {
		t.Fatalf("got %s", v.Kind())
	}
	n, ok := v.MapIndex("n")
	if !ok {
		t.Fatal("missing n")
	}
	i, err := n.Int()
	if err != nil || i != 7 {
		t.Fatalf("n=%d err=%v", i, err)
	}
}

func TestOverflow(t *testing.T) {
	v := Uint(math.MaxUint64)
	if _, err := v.Int(); err == nil {
		t.Fatal("expected overflow")
	}
	neg := Int(-1)
	if _, err := neg.Uint(); err == nil {
		t.Fatal("expected overflow")
	}
}

func TestConvert(t *testing.T) {
	got, err := Convert(Int(3), KindFloat)
	if err != nil {
		t.Fatal(err)
	}
	f, err := got.Float()
	if err != nil || f != 3 {
		t.Fatalf("got %v %v", f, err)
	}
}

func TestValidate(t *testing.T) {
	typ := &Type{Kind: KindStruct, Fields: map[string]*Type{
		"name": {Kind: KindString},
		"age":  {Kind: KindInt, Nullable: true},
	}}
	v, err := Struct("", []string{"name"}, []Value{String("ada")})
	if err != nil {
		t.Fatal(err)
	}
	if err = Validate(v, typ); err != nil {
		t.Fatal(err)
	}
	bad, err := Struct("", []string{"name"}, []Value{Int(1)})
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(bad, typ); err == nil {
		t.Fatal("expected type error")
	}
}

func TestCloneIndependence(t *testing.T) {
	orig := Array(Bytes([]byte{1, 2}))
	cp := orig.Clone()
	cp.items[0].raw[0] = 9
	if orig.items[0].raw[0] != 1 {
		t.Fatal("clone shared backing array")
	}
}
