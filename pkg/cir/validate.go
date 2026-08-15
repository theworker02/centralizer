package cir

import (
	"fmt"
	"math"
)

// Type describes an expected CIR shape for schema validation.
type Type struct {
	Kind     Kind
	Elem     *Type
	Fields   map[string]*Type
	Nullable bool
	Name     string
}

// Validate checks that v conforms to t. A nil type accepts any value.
func Validate(v Value, t *Type) error {
	if t == nil {
		return nil
	}
	if v.IsNull() {
		if t.Nullable || t.Kind == KindNull || t.Kind == KindOptional {
			return nil
		}
		return fmt.Errorf("cir: unexpected null for %s", t.Kind)
	}
	if t.Kind == KindOptional {
		if t.Elem == nil {
			return nil
		}
		if v.kind == KindOptional {
			inner, ok := v.OptionalInner()
			if !ok {
				return nil
			}
			return Validate(inner, t.Elem)
		}
		return Validate(v, t.Elem)
	}
	if v.kind != t.Kind && !compatibleNumeric(v.kind, t.Kind) {
		return fmt.Errorf("cir: expected %s, got %s", t.Kind, v.kind)
	}
	switch t.Kind {
	case KindArray, KindTuple:
		if t.Elem == nil {
			return nil
		}
		for i, item := range v.items {
			if err := Validate(item, t.Elem); err != nil {
				return fmt.Errorf("cir: index %d: %w", i, err)
			}
		}
	case KindMap, KindStruct:
		if len(t.Fields) == 0 {
			return nil
		}
		for k, ft := range t.Fields {
			item, ok := v.MapIndex(k)
			if !ok {
				if ft.Nullable {
					continue
				}
				return fmt.Errorf("cir: missing field %q", k)
			}
			if err := Validate(item, ft); err != nil {
				return fmt.Errorf("cir: field %q: %w", k, err)
			}
		}
	case KindFloat:
		if v.kind == KindFloat {
			f := math.Float64frombits(v.num)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				return fmt.Errorf("cir: non-finite float")
			}
		}
	}
	return nil
}

func compatibleNumeric(got, want Kind) bool {
	if got == want {
		return true
	}
	switch want {
	case KindFloat:
		return got == KindInt || got == KindUint
	case KindInt:
		return got == KindUint
	case KindUint:
		return got == KindInt
	default:
		return false
	}
}
