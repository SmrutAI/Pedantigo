package serialize

import (
	"reflect"
)

// SerializeOptions internal options for serialization.
type SerializeOptions struct {
	Context  string
	OmitZero bool
	TagName  string // The struct tag name to use (e.g., "pedantigo", "validate", "binding")
}

// ShouldIncludeField determines if a field should be included in output.
func ShouldIncludeField(
	meta FieldMetadata,
	fieldValue reflect.Value,
	opts SerializeOptions,
	hasWhitelistContext bool,
) bool {
	if opts.Context != "" {
		// 1. Check context-based exclusion (blacklist)
		if meta.ExcludeContexts[opts.Context] {
			return false
		}

		// 2. Check context-based inclusion (whitelist)
		// If any field in the struct has include:context, we're in whitelist mode
		if hasWhitelistContext {
			// Field must explicitly have include:context to be included
			if !meta.IncludeContexts[opts.Context] {
				return false
			}
		}
	}

	// 3. Check omitzero
	if meta.OmitZero && opts.OmitZero && isZeroValue(fieldValue) {
		return false
	}

	return true
}

// HasWhitelistContext checks if any field in metadata has include:context.
func HasWhitelistContext(metadata map[string]FieldMetadata, context string) bool {
	if context == "" {
		return false
	}
	for _, meta := range metadata {
		if meta.IncludeContexts[context] {
			return true
		}
	}
	return false
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

	// Check if we're in whitelist mode for this context
	hasWhitelist := HasWhitelistContext(metadata, opts.Context)

	for jsonName, meta := range metadata {
		fieldValue := val.Field(meta.FieldIndex)

		if !ShouldIncludeField(meta, fieldValue, opts, hasWhitelist) {
			continue
		}

		// Handle nested structs recursively
		switch {
		case fieldValue.Kind() == reflect.Struct:
			nestedMeta := BuildFieldMetadata(fieldValue.Type(), opts.TagName)
			result[jsonName] = ToFilteredMap(fieldValue, nestedMeta, opts)
		case fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil():
			elem := fieldValue.Elem()
			if elem.Kind() == reflect.Struct {
				nestedMeta := BuildFieldMetadata(elem.Type(), opts.TagName)
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
