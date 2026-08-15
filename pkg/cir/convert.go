package cir

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// From converts a native Go value into CIR. Unsupported types return
// ErrConversion. Pointers are dereferenced; nil pointers become null.
func From(v any) (Value, error) {
	if v == nil {
		return Null(), nil
	}
	switch t := v.(type) {
	case Value:
		return t, nil
	case bool:
		return Bool(t), nil
	case int:
		return Int(int64(t)), nil
	case int8:
		return Int(int64(t)), nil
	case int16:
		return Int(int64(t)), nil
	case int32:
		return Int(int64(t)), nil
	case int64:
		return Int(t), nil
	case uint:
		return Uint(uint64(t)), nil
	case uint8:
		return Uint(uint64(t)), nil
	case uint16:
		return Uint(uint64(t)), nil
	case uint32:
		return Uint(uint64(t)), nil
	case uint64:
		return Uint(t), nil
	case float32:
		return Float(float64(t)), nil
	case float64:
		return Float(t), nil
	case string:
		return String(t), nil
	case []byte:
		return Bytes(t), nil
	case time.Time:
		return Timestamp(t), nil
	case time.Duration:
		return Duration(t), nil
	case [16]byte:
		return UUID(t), nil
	case error:
		return ErrorValue("native", t.Error()), nil
	case map[string]any:
		keys := make([]string, 0, len(t))
		vals := make([]Value, 0, len(t))
		for k, val := range t {
			cv, err := From(val)
			if err != nil {
				return Value{}, err
			}
			keys = append(keys, k)
			vals = append(vals, cv)
		}
		return Map(keys, vals)
	case []any:
		items := make([]Value, len(t))
		for i, val := range t {
			cv, err := From(val)
			if err != nil {
				return Value{}, err
			}
			items[i] = cv
		}
		return Array(items...), nil
	}
	return fromReflect(reflect.ValueOf(v))
}

func fromReflect(rv reflect.Value) (Value, error) {
	if !rv.IsValid() {
		return Null(), nil
	}
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return Null(), nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Bool:
		return Bool(rv.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Int(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return Uint(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return Float(rv.Float()), nil
	case reflect.String:
		return String(rv.String()), nil
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return Bytes(rv.Bytes()), nil
		}
		n := rv.Len()
		items := make([]Value, n)
		for i := 0; i < n; i++ {
			cv, err := From(rv.Index(i).Interface())
			if err != nil {
				return Value{}, err
			}
			items[i] = cv
		}
		return Array(items...), nil
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 && rv.Len() == 16 {
			var b [16]byte
			reflect.Copy(reflect.ValueOf(b[:]), rv)
			return UUID(b), nil
		}
		n := rv.Len()
		items := make([]Value, n)
		for i := 0; i < n; i++ {
			cv, err := From(rv.Index(i).Interface())
			if err != nil {
				return Value{}, err
			}
			items[i] = cv
		}
		return Tuple(items...), nil
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return Value{}, czerr.New(czerr.ErrConversion, "map keys must be strings")
		}
		keys := make([]string, 0, rv.Len())
		vals := make([]Value, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			cv, err := From(iter.Value().Interface())
			if err != nil {
				return Value{}, err
			}
			keys = append(keys, iter.Key().String())
			vals = append(vals, cv)
		}
		return Map(keys, vals)
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			ts, ok := rv.Interface().(time.Time)
			if !ok {
				return Value{}, czerr.New(czerr.ErrConversion, "time.Time assertion failed")
			}
			return Timestamp(ts), nil
		}
		t := rv.Type()
		keys := make([]string, 0, t.NumField())
		vals := make([]Value, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := f.Name
			if tag := f.Tag.Get("cir"); tag != "" && tag != "-" {
				name = tag
			} else if tag := f.Tag.Get("json"); tag != "" && tag != "-" {
				name = tag
			}
			cv, err := From(rv.Field(i).Interface())
			if err != nil {
				return Value{}, err
			}
			keys = append(keys, name)
			vals = append(vals, cv)
		}
		return Struct(t.Name(), keys, vals)
	default:
		return Value{}, czerr.New(czerr.ErrConversion, fmt.Sprintf("unsupported Go type %s", rv.Type()))
	}
}

// Native converts a CIR value into a conventional Go representation.
func (v Value) Native() (any, error) {
	switch v.kind {
	case KindNull:
		return nil, nil
	case KindBool:
		return v.num != 0, nil
	case KindInt:
		return int64(v.num), nil
	case KindUint:
		return v.num, nil
	case KindFloat:
		return math.Float64frombits(v.num), nil
	case KindDecimal, KindString:
		return v.str, nil
	case KindBytes:
		return v.Bytes()
	case KindArray, KindTuple:
		out := make([]any, len(v.items))
		for i, item := range v.items {
			n, err := item.Native()
			if err != nil {
				return nil, err
			}
			out[i] = n
		}
		return out, nil
	case KindMap, KindStruct:
		out := make(map[string]any, len(v.keys))
		for i, k := range v.keys {
			n, err := v.items[i].Native()
			if err != nil {
				return nil, err
			}
			out[k] = n
		}
		return out, nil
	case KindTimestamp:
		return v.Timestamp()
	case KindDuration:
		return v.Duration()
	case KindUUID:
		return v.UUIDBytes()
	case KindError:
		return fmt.Errorf("%s: %s", v.str, string(v.raw)), nil
	case KindOptional:
		inner, ok := v.OptionalInner()
		if !ok {
			return nil, nil
		}
		return inner.Native()
	case KindResult:
		inner, ok, err := v.Result()
		if err != nil {
			return nil, err
		}
		if !ok {
			n, nerr := inner.Native()
			if nerr != nil {
				return nil, nerr
			}
			return nil, fmt.Errorf("cir result error: %v", n)
		}
		return inner.Native()
	case KindHandle:
		return HandleRef{ID: v.str}, nil
	case KindStream:
		return StreamID(v.str), nil
	case KindEnum:
		return EnumRef{Name: v.str, Discriminant: int64(v.num)}, nil
	case KindUnion:
		if len(v.items) == 0 {
			return UnionRef{Tag: v.str}, nil
		}
		n, err := v.items[0].Native()
		if err != nil {
			return nil, err
		}
		return UnionRef{Tag: v.str, Value: n}, nil
	case KindOpaque:
		return OpaqueRef{Tag: v.str, Payload: append([]byte(nil), v.raw...)}, nil
	default:
		return nil, czerr.New(czerr.ErrConversion, fmt.Sprintf("cannot convert %s to native Go", v.kind))
	}
}

// HandleRef is the native form of a foreign object handle.
type HandleRef struct{ ID string }

// StreamID identifies a remote stream.
type StreamID string

// EnumRef is the native form of an enum value.
type EnumRef struct {
	Name         string
	Discriminant int64
}

// UnionRef is the native form of a tagged union.
type UnionRef struct {
	Tag   string
	Value any
}

// OpaqueRef is language-specific opaque data.
type OpaqueRef struct {
	Tag     string
	Payload []byte
}

// Convert performs a safe kind conversion with overflow checks.
func Convert(v Value, to Kind) (Value, error) {
	if v.kind == to {
		return v, nil
	}
	if v.kind == KindNull && to == KindOptional {
		return None(), nil
	}
	switch to {
	case KindInt:
		n, err := v.Int()
		if err != nil {
			if v.kind == KindFloat {
				f := math.Float64frombits(v.num)
				if f > float64(math.MaxInt64) || f < float64(math.MinInt64) || math.IsNaN(f) || math.IsInf(f, 0) {
					return Value{}, overflowErr("float", f, "int64")
				}
				return Int(int64(f)), nil
			}
			return Value{}, czerr.Wrap(czerr.ErrConversion, "to int", err)
		}
		return Int(n), nil
	case KindUint:
		n, err := v.Uint()
		if err != nil {
			return Value{}, czerr.Wrap(czerr.ErrConversion, "to uint", err)
		}
		return Uint(n), nil
	case KindFloat:
		n, err := v.Float()
		if err != nil {
			return Value{}, czerr.Wrap(czerr.ErrConversion, "to float", err)
		}
		return Float(n), nil
	case KindString:
		switch v.kind {
		case KindString, KindDecimal:
			return String(v.str), nil
		case KindInt:
			return String(fmt.Sprintf("%d", int64(v.num))), nil
		case KindUint:
			return String(fmt.Sprintf("%d", v.num)), nil
		case KindFloat:
			return String(fmt.Sprintf("%g", math.Float64frombits(v.num))), nil
		case KindBool:
			if v.num != 0 {
				return String("true"), nil
			}
			return String("false"), nil
		default:
			return Value{}, czerr.New(czerr.ErrConversion, "cannot convert "+v.kind.String()+" to string")
		}
	case KindOptional:
		return Optional(v), nil
	default:
		return Value{}, czerr.New(czerr.ErrConversion, fmt.Sprintf("no conversion from %s to %s", v.kind, to))
	}
}
