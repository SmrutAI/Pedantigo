package serialize

import (
	"reflect"
)

// SerializeOptions internal options for serialization.
type SerializeOptions struct {
	Context  string
	OmitZero bool
}

// ShouldIncludeField determines if a field should be included in output.
func ShouldIncludeField(
	meta FieldMetadata,
	fieldValue reflect.Value,
	opts SerializeOptions,
) bool {
	// 1. Check context-based exclusion
	if opts.Context != "" && meta.ExcludeContexts[opts.Context] {
		return false
	}

	// 2. Check omitzero
	if meta.OmitZero && opts.OmitZero && isZeroValue(fieldValue) {
		return false
	}

	return true
}

// isZeroValue checks if a value is its zero value (including nil pointers).
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v.IsZero()
	}
}

// ToFilteredMap converts a struct to map[string]any with exclusions applied.
func ToFilteredMap(
	val reflect.Value,
	metadata map[string]FieldMetadata,
	opts SerializeOptions,
) map[string]any {
	result := make(map[string]any)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	for jsonName, meta := range metadata {
		fieldValue := val.Field(meta.FieldIndex)

		if !ShouldIncludeField(meta, fieldValue, opts) {
			continue
		}

		// Handle nested structs recursively
		switch {
		case fieldValue.Kind() == reflect.Struct:
			nestedMeta := BuildFieldMetadata(fieldValue.Type())
			result[jsonName] = ToFilteredMap(fieldValue, nestedMeta, opts)
		case fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil():
			elem := fieldValue.Elem()
			if elem.Kind() == reflect.Struct {
				nestedMeta := BuildFieldMetadata(elem.Type())
				result[jsonName] = ToFilteredMap(fieldValue, nestedMeta, opts)
			} else {
				// Dereference pointer to simple type
				result[jsonName] = elem.Interface()
			}
		default:
			result[jsonName] = fieldValue.Interface()
		}
	}

	return result
}
