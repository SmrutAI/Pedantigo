package deserialize

import (
	"fmt"
	"reflect"

	"github.com/SmrutAI/pedantigo/internal/tags"
)

// ExtraFieldInfo holds metadata about a struct's extra_fields field.
type ExtraFieldInfo struct {
	FieldIndex int    // Struct field index for the extra_fields map
	FieldName  string // Go field name (for error messages)
}

// DetectExtraField finds the field tagged with `pedantigo:"extra_fields"`.
// Returns nil if no such field exists.
// Panics if:
//   - Multiple fields have extra_fields tag
//   - Field type is not map[string]any
func DetectExtraField(typ reflect.Type, tagName string) *ExtraFieldInfo {
	// Handle pointer types - dereference to get the actual struct
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Return nil if not a struct
	if typ.Kind() != reflect.Struct {
		return nil
	}

	var foundField *ExtraFieldInfo

	// Iterate through all struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields (private fields)
		if !field.IsExported() {
			continue
		}

		// Check if this field has the extra_fields tag
		tagValue := field.Tag.Get(tagName)
		if tagValue != tags.ExtraFieldsTag {
			continue
		}

		// Found a field with extra_fields tag

		// Check for duplicate extra_fields tags
		if foundField != nil {
			panic("multiple fields tagged with pedantigo:\"extra_fields\" found: only one is allowed")
		}

		// Validate field type is map[string]any or map[string]interface{}
		fieldType := field.Type

		// Check it's not a pointer
		if fieldType.Kind() == reflect.Ptr {
			panic(fmt.Sprintf("field '%s' tagged with pedantigo:\"extra_fields\" must be of type map[string]any", field.Name))
		}

		// Check it's a map
		if fieldType.Kind() != reflect.Map {
			panic(fmt.Sprintf("field '%s' tagged with pedantigo:\"extra_fields\" must be of type map[string]any", field.Name))
		}

		// Check map key type is string
		keyType := fieldType.Key()
		if keyType.Kind() != reflect.String {
			panic(fmt.Sprintf("field '%s' tagged with pedantigo:\"extra_fields\" must be of type map[string]any", field.Name))
		}

		// Check map value type is any/interface{}
		valueType := fieldType.Elem()
		// interface{} is the same as any, both have Kind() == reflect.Interface
		// We need to check if it's the empty interface (no methods)
		if valueType.Kind() != reflect.Interface || valueType.NumMethod() != 0 {
			panic(fmt.Sprintf("field '%s' tagged with pedantigo:\"extra_fields\" must be of type map[string]any", field.Name))
		}

		// All validations passed - store the field info
		foundField = &ExtraFieldInfo{
			FieldIndex: i,
			FieldName:  field.Name,
		}
	}

	return foundField
}
