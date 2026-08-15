package cir

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// Meta carries optional schema and language-specific annotations.
// It is never required for a value to be valid.
type Meta struct {
	TypeName string
	Lang     string
	SchemaID string
	Extra    map[string]string
}

// Value is a CIR value. Storage is kind-tagged rather than boxed in any,
// so common scalars avoid an extra heap allocation for the payload itself.
// Composite values store children inline.
type Value struct {
	kind  Kind
	flags uint8
	num   uint64
	str   string
	raw   []byte
	items []Value
	keys  []string
	meta  *Meta
}

const (
	flagNone     uint8 = 0
	flagNegative uint8 = 1 << 0 // unused; sign lives in two's complement bits
	flagOK       uint8 = 1 << 1 // result ok vs error
	flagSome     uint8 = 1 << 2 // optional present
)

// Kind returns the value kind.
func (v Value) Kind() Kind { return v.kind }

// Meta returns optional metadata, which may be nil.
func (v Value) Meta() *Meta { return v.meta }

// WithMeta returns a copy of v with metadata attached.
func (v Value) WithMeta(m *Meta) Value {
	v.meta = m
	return v
}

// IsNull reports whether v is the null value or an empty optional.
func (v Value) IsNull() bool {
	return v.kind == KindNull || (v.kind == KindOptional && v.flags&flagSome == 0)
}

func Null() Value { return Value{kind: KindNull} }

func Bool(b bool) Value {
	var n uint64
	if b {
		n = 1
	}
	return Value{kind: KindBool, num: n}
}

func Int(i int64) Value {
	return Value{kind: KindInt, num: uint64(i)}
}

func Uint(u uint64) Value {
	return Value{kind: KindUint, num: u}
}

func Float(f float64) Value {
	return Value{kind: KindFloat, num: math.Float64bits(f)}
}

func Decimal(s string) Value {
	return Value{kind: KindDecimal, str: s}
}

func String(s string) Value {
	return Value{kind: KindString, str: s}
}

func Bytes(b []byte) Value {
	cp := make([]byte, len(b))
	copy(cp, b)
	return Value{kind: KindBytes, raw: cp}
}

func Array(items ...Value) Value {
	return Value{kind: KindArray, items: items}
}

func Tuple(items ...Value) Value {
	return Value{kind: KindTuple, items: items}
}

func Map(keys []string, values []Value) (Value, error) {
	if len(keys) != len(values) {
		return Value{}, fmt.Errorf("cir: map key/value length mismatch")
	}
	return Value{kind: KindMap, keys: keys, items: values}, nil
}

func MustMap(pairs ...any) Value {
	if len(pairs)%2 != 0 {
		panic("cir.MustMap: odd number of arguments")
	}
	keys := make([]string, 0, len(pairs)/2)
	vals := make([]Value, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		k, ok := pairs[i].(string)
		if !ok {
			panic("cir.MustMap: key must be string")
		}
		v, ok := pairs[i+1].(Value)
		if !ok {
			panic("cir.MustMap: value must be cir.Value")
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}
	return Value{kind: KindMap, keys: keys, items: vals}
}

func Struct(typeName string, keys []string, values []Value) (Value, error) {
	if len(keys) != len(values) {
		return Value{}, fmt.Errorf("cir: struct field length mismatch")
	}
	v := Value{kind: KindStruct, keys: keys, items: values}
	if typeName != "" {
		v.meta = &Meta{TypeName: typeName}
	}
	return v, nil
}

func Enum(name string, discriminant int64) Value {
	return Value{kind: KindEnum, str: name, num: uint64(discriminant)}
}

func Optional(inner Value) Value {
	return Value{kind: KindOptional, flags: flagSome, items: []Value{inner}}
}

func None() Value {
	return Value{kind: KindOptional}
}

func ResultOK(inner Value) Value {
	return Value{kind: KindResult, flags: flagOK, items: []Value{inner}}
}

func ResultErr(inner Value) Value {
	return Value{kind: KindResult, items: []Value{inner}}
}

func ErrorValue(code, message string) Value {
	return Value{kind: KindError, str: code, raw: []byte(message)}
}

func Timestamp(t time.Time) Value {
	return Value{kind: KindTimestamp, num: uint64(t.UTC().UnixNano())}
}

func Duration(d time.Duration) Value {
	return Value{kind: KindDuration, num: uint64(d.Nanoseconds())}
}

func UUID(b [16]byte) Value {
	return Value{kind: KindUUID, raw: b[:]}
}

func Handle(id string) Value {
	return Value{kind: KindHandle, str: id}
}

func Opaque(tag string, payload []byte) Value {
	cp := make([]byte, len(payload))
	copy(cp, payload)
	return Value{kind: KindOpaque, str: tag, raw: cp}
}

func StreamRef(id string) Value {
	return Value{kind: KindStream, str: id}
}

func Union(tag string, inner Value) Value {
	return Value{kind: KindUnion, str: tag, items: []Value{inner}}
}

func (v Value) Bool() (bool, error) {
	if v.kind != KindBool {
		return false, typeErr(KindBool, v.kind)
	}
	return v.num != 0, nil
}

func (v Value) Int() (int64, error) {
	switch v.kind {
	case KindInt:
		return int64(v.num), nil
	case KindUint:
		if v.num > math.MaxInt64 {
			return 0, overflowErr("uint", v.num, "int64")
		}
		return int64(v.num), nil
	default:
		return 0, typeErr(KindInt, v.kind)
	}
}

func (v Value) Uint() (uint64, error) {
	switch v.kind {
	case KindUint:
		return v.num, nil
	case KindInt:
		i := int64(v.num)
		if i < 0 {
			return 0, overflowErr("int", i, "uint64")
		}
		return uint64(i), nil
	default:
		return 0, typeErr(KindUint, v.kind)
	}
}

func (v Value) Float() (float64, error) {
	switch v.kind {
	case KindFloat:
		return math.Float64frombits(v.num), nil
	case KindInt:
		return float64(int64(v.num)), nil
	case KindUint:
		return float64(v.num), nil
	default:
		return 0, typeErr(KindFloat, v.kind)
	}
}

func (v Value) String() string {
	s, err := v.AsString()
	if err != nil {
		return fmt.Sprintf("<%s>", v.kind)
	}
	return s
}

func (v Value) AsString() (string, error) {
	if v.kind != KindString && v.kind != KindDecimal {
		return "", typeErr(KindString, v.kind)
	}
	return v.str, nil
}

func (v Value) Bytes() ([]byte, error) {
	if v.kind != KindBytes {
		return nil, typeErr(KindBytes, v.kind)
	}
	cp := make([]byte, len(v.raw))
	copy(cp, v.raw)
	return cp, nil
}

func (v Value) Items() ([]Value, error) {
	if !v.kind.IsComposite() && v.kind != KindArray && v.kind != KindTuple {
		return nil, typeErr(KindArray, v.kind)
	}
	return v.items, nil
}

func (v Value) Keys() []string { return v.keys }

func (v Value) Len() int {
	switch v.kind {
	case KindArray, KindTuple, KindMap, KindStruct:
		return len(v.items)
	case KindString:
		return len(v.str)
	case KindBytes, KindUUID, KindOpaque:
		return len(v.raw)
	default:
		return 0
	}
}

func (v Value) MapIndex(key string) (Value, bool) {
	if v.kind != KindMap && v.kind != KindStruct {
		return Value{}, false
	}
	for i, k := range v.keys {
		if k == key {
			return v.items[i], true
		}
	}
	return Value{}, false
}

func (v Value) HandleID() (string, error) {
	if v.kind != KindHandle {
		return "", typeErr(KindHandle, v.kind)
	}
	return v.str, nil
}

func (v Value) Timestamp() (time.Time, error) {
	if v.kind != KindTimestamp {
		return time.Time{}, typeErr(KindTimestamp, v.kind)
	}
	return time.Unix(0, int64(v.num)).UTC(), nil
}

func (v Value) Duration() (time.Duration, error) {
	if v.kind != KindDuration {
		return 0, typeErr(KindDuration, v.kind)
	}
	return time.Duration(int64(v.num)), nil
}

func (v Value) UUIDBytes() ([16]byte, error) {
	var out [16]byte
	if v.kind != KindUUID {
		return out, typeErr(KindUUID, v.kind)
	}
	if len(v.raw) != 16 {
		return out, fmt.Errorf("cir: uuid must be 16 bytes, got %d", len(v.raw))
	}
	copy(out[:], v.raw)
	return out, nil
}

func (v Value) ErrorCode() string {
	if v.kind != KindError {
		return ""
	}
	return v.str
}

func (v Value) ErrorMessage() string {
	if v.kind != KindError {
		return ""
	}
	return string(v.raw)
}

func (v Value) OptionalInner() (Value, bool) {
	if v.kind != KindOptional || v.flags&flagSome == 0 || len(v.items) == 0 {
		return Value{}, false
	}
	return v.items[0], true
}

func (v Value) Result() (Value, bool, error) {
	if v.kind != KindResult {
		return Value{}, false, typeErr(KindResult, v.kind)
	}
	if len(v.items) == 0 {
		return Value{}, v.flags&flagOK != 0, nil
	}
	return v.items[0], v.flags&flagOK != 0, nil
}

func typeErr(want, got Kind) error {
	return fmt.Errorf("cir: expected %s, got %s", want, got)
}

func overflowErr(from string, val any, to string) error {
	return fmt.Errorf("cir: numeric overflow converting %s %v to %s", from, val, to)
}

// Equal reports structural equality, ignoring metadata.
func Equal(a, b Value) bool {
	if a.kind != b.kind || a.flags != b.flags || a.num != b.num || a.str != b.str {
		return false
	}
	if len(a.raw) != len(b.raw) || len(a.items) != len(b.items) || len(a.keys) != len(b.keys) {
		return false
	}
	for i := range a.raw {
		if a.raw[i] != b.raw[i] {
			return false
		}
	}
	for i := range a.keys {
		if a.keys[i] != b.keys[i] {
			return false
		}
	}
	for i := range a.items {
		if !Equal(a.items[i], b.items[i]) {
			return false
		}
	}
	return true
}

// Clone returns a deep copy of v.
func (v Value) Clone() Value {
	out := v
	if v.raw != nil {
		out.raw = make([]byte, len(v.raw))
		copy(out.raw, v.raw)
	}
	if v.items != nil {
		out.items = make([]Value, len(v.items))
		for i := range v.items {
			out.items[i] = v.items[i].Clone()
		}
	}
	if v.keys != nil {
		out.keys = append([]string(nil), v.keys...)
	}
	if v.meta != nil {
		m := *v.meta
		if v.meta.Extra != nil {
			m.Extra = make(map[string]string, len(v.meta.Extra))
			for k, val := range v.meta.Extra {
				m.Extra[k] = val
			}
		}
		out.meta = &m
	}
	return out
}

// EncodeUUIDString formats a UUID value as 8-4-4-4-12 hex.
func EncodeUUIDString(b [16]byte) string {
	const hex = "0123456789abcdef"
	var buf [36]byte
	w := 0
	dash := map[int]bool{4: true, 6: true, 8: true, 10: true}
	for i := 0; i < 16; i++ {
		if dash[i] {
			buf[w] = '-'
			w++
		}
		buf[w] = hex[b[i]>>4]
		buf[w+1] = hex[b[i]&0x0f]
		w += 2
	}
	return string(buf[:])
}

// PutUint64BE is exported for protocol helpers that need stable numeric encoding.
func PutUint64BE(b []byte, v uint64) {
	binary.BigEndian.PutUint64(b, v)
}
