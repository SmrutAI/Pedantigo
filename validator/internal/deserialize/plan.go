package deserialize

import (
	"fmt"
	"reflect"
)

// DecoderKind records, once at build time, whether a type customizes its own
// decoding, so the hot path never calls reflect.Implements.
type DecoderKind uint8

const (
	DecoderNone            DecoderKind = iota // plain type; walk normally
	DecoderWalker                             // *T implements WalkerDecoder
	DecoderJSONUnmarshaler                    // *T implements json.Unmarshaler
)

// FieldKind is the decode-shape of a field's (deref'd) Go type, computed once at build.
type FieldKind uint8

const (
	KindScalar FieldKind = iota // leaf set via SetFieldValueWithOptions (primitives, time.Time/Duration, interface{}, conversions)
	KindStruct                  // struct (Nested *TypePlan)
	KindSlice                   // slice (ElemNested/ElemDecoderKind describe the element)
	KindMap                     // map with struct/interface values (ElemNested/ElemDecoderKind describe the value)
)

// FieldPlan is the precomputed decode plan for one struct field. Read by the
// interpreter with no reflection beyond value access and, only when DecoderKind
// != DecoderNone, one flat type assertion.
type FieldPlan struct {
	FieldIndex    int     // parent struct field index for O(1) Field(i)
	GoName        string  // Go struct field name, precomputed (avoids reflect.Type.Field per call)
	JSONName      string  // key to look up in the decoded map[string]any
	Required      bool    // validate:"required" present
	StaticDefault *string // validate:"default=..." value; nil if none
	DefaultMethod string  // validate:"defaultUsingMethod=..." method name; "" if none

	StripWhitespace bool
	ToLower         bool
	ToUpper         bool

	Kind            FieldKind
	IsPointer       bool        // field type is a pointer (deref before Kind applies)
	DecoderKind     DecoderKind // custom decoder for the (deref'd) field type
	ElemDecoderKind DecoderKind // custom decoder for slice/map element (deref'd)

	Nested     *TypePlan // KindStruct: the field type's plan (may be a back-edge / shared node)
	ElemNested *TypePlan // KindSlice/KindMap with struct/interface element: element type's plan (may be a back-edge)
}

// TypePlan is the precomputed plan for one struct type. Fields is a flat slice
// (no per-field closures, no pointer-chains). Nodes are deduplicated per
// reflect.Type in the builder's index and may be referenced by back-edges.
type TypePlan struct {
	Type            reflect.Type
	Fields          []FieldPlan
	ExtraFieldIndex int         // index of the map[string]any `validate:"extra_fields"` field, or -1 if none
	SelfDecoderKind DecoderKind // does THIS type implement WalkerDecoder/json.Unmarshaler (for recurse() dispatch on a struct dst)

	// JSONNameSet is the set of all Fields[].JSONName, precomputed once when
	// ExtraFieldIndex >= 0 (nil otherwise). Used by the ExtraAllow capture path
	// to test key membership without rebuilding the set on every DecodeStruct call.
	JSONNameSet map[string]struct{}
}

// MaxDepthExceededError is returned by both the Unmarshal interpreter and the
// Validate traversal when self-referential recursion exceeds the configured
// limit. It is a security control against deeply-nested/recursive DoS payloads.
type MaxDepthExceededError struct {
	Path  string
	Limit int
}

func (e *MaxDepthExceededError) Error() string {
	return fmt.Sprintf("max recursion depth %d exceeded at %s", e.Limit, e.Path)
}

// FieldDecodeError wraps a non-required decode error (e.g. a scalar type
// conversion failure, or a shape mismatch like "expected object, got string")
// with the field that caused it, so callers can attribute it instead of
// falling back to the struct root. Field is fp.JSONName at the root (matching
// the legacy top-level deserializer) or the dotted Go-field path when nested.
type FieldDecodeError struct {
	Field string
	Err   error
}

func (e *FieldDecodeError) Error() string { return e.Err.Error() }

func (e *FieldDecodeError) Unwrap() error { return e.Err }
