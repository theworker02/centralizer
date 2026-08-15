package cir

import (
	"encoding/json"
	"fmt"
	"math"
)

// Wire is the versioned JSON encoding of a CIR value. Adapters and the
// protocol use this shape; it is intentionally explicit rather than
// relying on JSON type guessing.
type Wire struct {
	K    string      `json:"k"`
	B    *bool       `json:"b,omitempty"`
	I    *int64      `json:"i,omitempty"`
	U    *uint64     `json:"u,omitempty"`
	F    *float64    `json:"f,omitempty"`
	S    *string     `json:"s,omitempty"`
	X    []byte      `json:"x,omitempty"`
	A    []Wire      `json:"a,omitempty"`
	M    []WireEntry `json:"m,omitempty"`
	N    *string     `json:"n,omitempty"`
	P    *bool       `json:"p,omitempty"`
	T    *string     `json:"t,omitempty"`
	Meta *WireMeta   `json:"meta,omitempty"`
}

// WireEntry is a map or struct field on the wire.
type WireEntry struct {
	Key string `json:"k"`
	Val Wire   `json:"v"`
}

// WireMeta is optional schema metadata on the wire.
type WireMeta struct {
	TypeName string            `json:"type,omitempty"`
	Lang     string            `json:"lang,omitempty"`
	SchemaID string            `json:"schema,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// MarshalJSON encodes a Value as Wire JSON.
func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ToWire())
}

// UnmarshalJSON decodes Wire JSON into a Value.
func (v *Value) UnmarshalJSON(data []byte) error {
	var w Wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	decoded, err := FromWire(w)
	if err != nil {
		return err
	}
	*v = decoded
	return nil
}

// ToWire converts v into the protocol encoding.
func (v Value) ToWire() Wire {
	w := Wire{K: v.kind.String()}
	if v.meta != nil {
		w.Meta = &WireMeta{
			TypeName: v.meta.TypeName,
			Lang:     v.meta.Lang,
			SchemaID: v.meta.SchemaID,
			Extra:    v.meta.Extra,
		}
	}
	switch v.kind {
	case KindNull:
	case KindBool:
		b := v.num != 0
		w.B = &b
	case KindInt:
		i := int64(v.num)
		w.I = &i
	case KindUint:
		u := v.num
		w.U = &u
	case KindFloat:
		f := math.Float64frombits(v.num)
		w.F = &f
	case KindDecimal, KindString, KindHandle, KindStream, KindEnum, KindOpaque:
		s := v.str
		w.S = &s
		if len(v.raw) > 0 {
			w.X = v.raw
		}
		if v.kind == KindEnum {
			i := int64(v.num)
			w.I = &i
		}
	case KindBytes, KindUUID:
		w.X = v.raw
	case KindTimestamp, KindDuration:
		i := int64(v.num)
		w.I = &i
	case KindError:
		s := v.str
		w.S = &s
		w.X = v.raw
	case KindArray, KindTuple:
		w.A = make([]Wire, len(v.items))
		for i, item := range v.items {
			w.A[i] = item.ToWire()
		}
	case KindMap, KindStruct:
		w.M = make([]WireEntry, len(v.keys))
		for i, k := range v.keys {
			w.M[i] = WireEntry{Key: k, Val: v.items[i].ToWire()}
		}
		if v.kind == KindStruct && v.meta != nil && v.meta.TypeName != "" {
			n := v.meta.TypeName
			w.N = &n
		}
	case KindOptional:
		some := v.flags&flagSome != 0
		w.P = &some
		if some && len(v.items) > 0 {
			w.A = []Wire{v.items[0].ToWire()}
		}
	case KindResult:
		ok := v.flags&flagOK != 0
		w.P = &ok
		if len(v.items) > 0 {
			w.A = []Wire{v.items[0].ToWire()}
		}
	case KindUnion:
		s := v.str
		w.S = &s
		if len(v.items) > 0 {
			w.A = []Wire{v.items[0].ToWire()}
		}
	}
	return w
}

// FromWire decodes a protocol value.
func FromWire(w Wire) (Value, error) {
	kind, err := ParseKind(w.K)
	if err != nil {
		return Value{}, err
	}
	var v Value
	switch kind {
	case KindNull:
		v = Null()
	case KindBool:
		if w.B == nil {
			return Value{}, fmt.Errorf("cir: boolean missing b")
		}
		v = Bool(*w.B)
	case KindInt:
		if w.I == nil {
			return Value{}, fmt.Errorf("cir: int missing i")
		}
		v = Int(*w.I)
	case KindUint:
		if w.U == nil {
			return Value{}, fmt.Errorf("cir: uint missing u")
		}
		v = Uint(*w.U)
	case KindFloat:
		if w.F == nil {
			return Value{}, fmt.Errorf("cir: float missing f")
		}
		v = Float(*w.F)
	case KindDecimal:
		if w.S == nil {
			return Value{}, fmt.Errorf("cir: decimal missing s")
		}
		v = Decimal(*w.S)
	case KindString:
		if w.S == nil {
			return Value{}, fmt.Errorf("cir: string missing s")
		}
		v = String(*w.S)
	case KindBytes:
		v = Bytes(w.X)
	case KindArray:
		items := make([]Value, len(w.A))
		for i, item := range w.A {
			cv, err := FromWire(item)
			if err != nil {
				return Value{}, err
			}
			items[i] = cv
		}
		v = Array(items...)
	case KindTuple:
		items := make([]Value, len(w.A))
		for i, item := range w.A {
			cv, err := FromWire(item)
			if err != nil {
				return Value{}, err
			}
			items[i] = cv
		}
		v = Tuple(items...)
	case KindMap, KindStruct:
		keys := make([]string, len(w.M))
		vals := make([]Value, len(w.M))
		for i, e := range w.M {
			cv, err := FromWire(e.Val)
			if err != nil {
				return Value{}, err
			}
			keys[i] = e.Key
			vals[i] = cv
		}
		if kind == KindMap {
			v, err = Map(keys, vals)
		} else {
			name := ""
			if w.N != nil {
				name = *w.N
			}
			v, err = Struct(name, keys, vals)
		}
		if err != nil {
			return Value{}, err
		}
	case KindEnum:
		name := ""
		if w.S != nil {
			name = *w.S
		}
		var d int64
		if w.I != nil {
			d = *w.I
		}
		v = Enum(name, d)
	case KindUnion:
		tag := ""
		if w.S != nil {
			tag = *w.S
		}
		inner := Null()
		if len(w.A) > 0 {
			inner, err = FromWire(w.A[0])
			if err != nil {
				return Value{}, err
			}
		}
		v = Union(tag, inner)
	case KindOptional:
		if w.P != nil && *w.P && len(w.A) > 0 {
			inner, err := FromWire(w.A[0])
			if err != nil {
				return Value{}, err
			}
			v = Optional(inner)
		} else {
			v = None()
		}
	case KindResult:
		inner := Null()
		if len(w.A) > 0 {
			inner, err = FromWire(w.A[0])
			if err != nil {
				return Value{}, err
			}
		}
		if w.P != nil && *w.P {
			v = ResultOK(inner)
		} else {
			v = ResultErr(inner)
		}
	case KindError:
		code := ""
		if w.S != nil {
			code = *w.S
		}
		v = ErrorValue(code, string(w.X))
	case KindTimestamp:
		if w.I == nil {
			return Value{}, fmt.Errorf("cir: timestamp missing i")
		}
		v = Value{kind: KindTimestamp, num: uint64(*w.I)}
	case KindDuration:
		if w.I == nil {
			return Value{}, fmt.Errorf("cir: duration missing i")
		}
		v = Value{kind: KindDuration, num: uint64(*w.I)}
	case KindUUID:
		if len(w.X) != 16 {
			return Value{}, fmt.Errorf("cir: uuid must be 16 bytes")
		}
		var b [16]byte
		copy(b[:], w.X)
		v = UUID(b)
	case KindStream:
		id := ""
		if w.S != nil {
			id = *w.S
		}
		v = StreamRef(id)
	case KindHandle:
		id := ""
		if w.S != nil {
			id = *w.S
		}
		v = Handle(id)
	case KindOpaque:
		tag := ""
		if w.S != nil {
			tag = *w.S
		}
		v = Opaque(tag, w.X)
	default:
		return Value{}, fmt.Errorf("cir: unsupported wire kind %s", w.K)
	}
	if w.Meta != nil {
		v.meta = &Meta{
			TypeName: w.Meta.TypeName,
			Lang:     w.Meta.Lang,
			SchemaID: w.Meta.SchemaID,
			Extra:    w.Meta.Extra,
		}
	}
	if w.T != nil && v.meta == nil {
		v.meta = &Meta{TypeName: *w.T}
	}
	return v, nil
}
