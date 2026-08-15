package cir

import "fmt"

// Kind identifies a CIR value type. CIR is the semantic boundary between
// supported languages; adapters convert native values into these kinds.
type Kind uint8

const (
	KindInvalid Kind = iota
	KindNull
	KindBool
	KindInt
	KindUint
	KindFloat
	KindDecimal
	KindString
	KindBytes
	KindArray
	KindMap
	KindTuple
	KindStruct
	KindEnum
	KindUnion
	KindOptional
	KindResult
	KindError
	KindTimestamp
	KindDuration
	KindUUID
	KindStream
	KindHandle
	KindOpaque
)

var kindNames = [...]string{
	KindInvalid:   "invalid",
	KindNull:      "null",
	KindBool:      "boolean",
	KindInt:       "int",
	KindUint:      "uint",
	KindFloat:     "float",
	KindDecimal:   "decimal",
	KindString:    "string",
	KindBytes:     "bytes",
	KindArray:     "array",
	KindMap:       "map",
	KindTuple:     "tuple",
	KindStruct:    "struct",
	KindEnum:      "enum",
	KindUnion:     "union",
	KindOptional:  "optional",
	KindResult:    "result",
	KindError:     "error",
	KindTimestamp: "timestamp",
	KindDuration:  "duration",
	KindUUID:      "uuid",
	KindStream:    "stream",
	KindHandle:    "handle",
	KindOpaque:    "opaque",
}

func (k Kind) String() string {
	if int(k) < len(kindNames) && kindNames[k] != "" {
		return kindNames[k]
	}
	return fmt.Sprintf("kind(%d)", k)
}

// ParseKind maps a wire or schema type name onto a Kind.
func ParseKind(name string) (Kind, error) {
	for i, n := range kindNames {
		if n == name {
			return Kind(i), nil
		}
	}
	switch name {
	case "bool":
		return KindBool, nil
	case "signed", "int64", "i64":
		return KindInt, nil
	case "unsigned", "uint64", "u64":
		return KindUint, nil
	case "float64", "f64", "double":
		return KindFloat, nil
	case "str":
		return KindString, nil
	case "bin", "binary":
		return KindBytes, nil
	case "list":
		return KindArray, nil
	case "object", "record":
		return KindStruct, nil
	case "time", "datetime":
		return KindTimestamp, nil
	}
	return KindInvalid, fmt.Errorf("unknown CIR kind %q", name)
}

// IsPrimitive reports whether k is a scalar kind.
func (k Kind) IsPrimitive() bool {
	switch k {
	case KindNull, KindBool, KindInt, KindUint, KindFloat, KindDecimal,
		KindString, KindBytes, KindTimestamp, KindDuration, KindUUID:
		return true
	default:
		return false
	}
}

// IsComposite reports whether k contains nested values.
func (k Kind) IsComposite() bool {
	switch k {
	case KindArray, KindMap, KindTuple, KindStruct, KindOptional, KindResult, KindUnion:
		return true
	default:
		return false
	}
}
