package constraints

import (
	"fmt"
	"reflect"
)

// FieldPath represents a path to a possibly nested struct field.
// Example paths: "Name" (single field), "Inner.Value" (nested), "A.B.C.Field" (multi-level).
type FieldPath struct {
	Raw          string         // Original dotted path (e.g., "Inner.MinValue")
	Parts        []string       // Split path components
	TypeAtLevel  []reflect.Type // Type at each level (for validation)
	IndexAtLevel []int          // Field index at each level (for traversal)
}

// ParseFieldPath parses a dotted field path (e.g., "Inner.MinValue") and validates
// it against the given struct type. Returns a FieldPath that can be used to resolve
// values at runtime.
//
// Panics if:
//   - The path contains an invalid field name
//   - The path references an unexported field
//   - The path goes through a non-struct type (except pointers to structs)
//
// Parameters:
//   - structType: The root struct type to validate against
//   - path: The dotted path string (e.g., "Inner.Value" or just "Value")
//
// Returns: A validated FieldPath ready for use with ResolveValue.
func ParseFieldPath(structType reflect.Type, path string) *FieldPath {
	parts := splitPath(path)

	fp := &FieldPath{
		Raw:          path,
		Parts:        parts,
		TypeAtLevel:  make([]reflect.Type, len(parts)),
		IndexAtLevel: make([]int, len(parts)),
	}

	currentType := structType

	// Traverse the path and validate each part
	for i, part := range parts {
		// Dereference pointers to get to the underlying struct type
		for currentType.Kind() == reflect.Ptr {
			currentType = currentType.Elem()
		}

		// Ensure we're working with a struct
		if currentType.Kind() != reflect.Struct {
			panic("field path traverses through non-struct type at part: " + part)
		}

		// Find the field by name
		field, found := currentType.FieldByName(part)
		if !found {
			panic("field not found: " + part + " in type " + currentType.String())
		}

		// Check if field is exported (first letter uppercase)
		if !field.IsExported() {
			panic("field not exported: " + part + " in type " + currentType.String())
		}

		// Store the field index and type at this level
		fp.IndexAtLevel[i] = field.Index[0] // Use first index for simple fields
		fp.TypeAtLevel[i] = field.Type

		// Move to the next level
		currentType = field.Type
	}

	return fp
}

// splitPath splits a dotted path into parts.
func splitPath(path string) []string {
	parts := []string{}
	current := ""

	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// ResolveValue traverses the struct using the pre-computed indices and returns
// the field value. Handles pointer dereferencing at each level.
//
// Parameters:
//   - structValue: A reflect.Value of the struct instance to traverse
//
// Returns:
//   - The field value as any (interface{})
//   - Error if a nil pointer is encountered in the path
func (fp *FieldPath) ResolveValue(structValue reflect.Value) (any, error) {
	current := structValue

	// Traverse through each part of the path
	for i, fieldIndex := range fp.IndexAtLevel {
		// Dereference pointers until we get to a struct
		for current.Kind() == reflect.Ptr {
			if current.IsNil() {
				// Nil pointer encountered - return error
				return nil, fmt.Errorf("nil pointer encountered in field path %q at part: %s", fp.Raw, fp.Parts[i])
			}
			current = current.Elem()
		}

		// Get the field by index
		current = current.Field(fieldIndex)
	}

	// Return the final field value as interface{}
	return current.Interface(), nil
}

// isNested returns true if this path has multiple levels (contains a dot).
func (fp *FieldPath) isNested() bool {
	return len(fp.Parts) > 1
}

// String returns the original path string.
func (fp *FieldPath) String() string {
	return fp.Raw
}
