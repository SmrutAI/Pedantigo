package validation

import (
	"fmt"
	"reflect"
)

// FieldError represents a validation error for a specific field
// This is a copy of the root package's FieldError to avoid circular imports
type FieldError struct {
	Field   string
	Message string
	Value   any
}

// ConstraintValidator is the interface for validation constraints
type ConstraintValidator interface {
	Validate(value any) error
}

// TagParser is a function type for parsing struct tags
type TagParser func(tag reflect.StructTag) map[string]string

// ConstraintBuilder is a function type for building constraint validators
type ConstraintBuilder func(constraints map[string]string, fieldType reflect.Type) []ConstraintValidator

// ValidateValue recursively validates a reflected value
// NOTE: 'required' constraint is skipped (not built in BuildConstraints)
func ValidateValue(
	val reflect.Value,
	path string,
	strictMissingFields bool,
	parseTagFunc TagParser,
	buildConstraintsFunc ConstraintBuilder,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var errors []FieldError

	// Handle pointer indirection
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return errors // nil pointers are handled by required constraint
		}
		val = val.Elem()
	}

	// Only validate structs
	if val.Kind() != reflect.Struct {
		return errors
	}

	typ := val.Type()

	// Iterate through all fields
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Build field path
		fieldPath := field.Name
		if path != "" {
			fieldPath = path + "." + field.Name
		}

		// Parse validation tags
		constraintsMap := parseTagFunc(field.Tag)
		if constraintsMap == nil {
			// No validation tags, but still check nested structs, slices, and maps
			if fieldValue.Kind() == reflect.Struct {
				errors = append(errors, recursiveValidateFunc(fieldValue, fieldPath)...)
			} else if fieldValue.Kind() == reflect.Slice {
				// Recursively validate struct elements in slices
				for i := 0; i < fieldValue.Len(); i++ {
					elemValue := fieldValue.Index(i)
					elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
					if elemValue.Kind() == reflect.Struct {
						errors = append(errors, recursiveValidateFunc(elemValue, elemPath)...)
					}
				}
			} else if fieldValue.Kind() == reflect.Map {
				// Recursively validate struct values in maps
				iter := fieldValue.MapRange()
				for iter.Next() {
					mapKey := iter.Key()
					mapValue := iter.Value()
					mapPath := fmt.Sprintf("%s[%v]", fieldPath, mapKey.Interface())
					if mapValue.Kind() == reflect.Struct {
						errors = append(errors, recursiveValidateFunc(mapValue, mapPath)...)
					}
				}
			}
			continue
		}

		// For nested structs (path != ""), check required fields
		// This is needed because nested structs in slices are deserialized via json.Unmarshal
		// which bypasses our custom required field checking in Unmarshal()
		if path != "" && strictMissingFields {
			if _, hasRequired := constraintsMap["required"]; hasRequired {
				// Check if field is zero value (indicates it was missing from JSON)
				if fieldValue.IsZero() {
					errors = append(errors, FieldError{
						Field:   fieldPath,
						Message: "is required",
						Value:   fieldValue.Interface(),
					})
					// Skip further validation for this field
					continue
				}
			}
		}

		// Build constraint validators (required is already skipped in BuildConstraints)
		validators := buildConstraintsFunc(constraintsMap, field.Type)

		// For slices, validate each element instead of the slice itself
		if fieldValue.Kind() == reflect.Slice {
			for i := 0; i < fieldValue.Len(); i++ {
				elemValue := fieldValue.Index(i)
				elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)

				// Apply constraints to each element
				for _, validator := range validators {
					if err := validator.Validate(elemValue.Interface()); err != nil {
						errors = append(errors, FieldError{
							Field:   elemPath,
							Message: err.Error(),
							Value:   elemValue.Interface(),
						})
					}
				}

				// Recursively validate nested structs in slice
				if elemValue.Kind() == reflect.Struct {
					errors = append(errors, recursiveValidateFunc(elemValue, elemPath)...)
				}
			}
		} else if fieldValue.Kind() == reflect.Map {
			// For maps, validate each value instead of the map itself
			iter := fieldValue.MapRange()
			for iter.Next() {
				mapKey := iter.Key()
				mapValue := iter.Value()
				mapPath := fmt.Sprintf("%s[%v]", fieldPath, mapKey.Interface())

				// Apply constraints to each value
				for _, validator := range validators {
					if err := validator.Validate(mapValue.Interface()); err != nil {
						errors = append(errors, FieldError{
							Field:   mapPath,
							Message: err.Error(),
							Value:   mapValue.Interface(),
						})
					}
				}

				// Recursively validate nested structs in map
				if mapValue.Kind() == reflect.Struct {
					errors = append(errors, recursiveValidateFunc(mapValue, mapPath)...)
				}
			}
		} else {
			// For non-slice/map fields, apply constraints directly
			for _, validator := range validators {
				if err := validator.Validate(fieldValue.Interface()); err != nil {
					errors = append(errors, FieldError{
						Field:   fieldPath,
						Message: err.Error(),
						Value:   fieldValue.Interface(),
					})
				}
			}

			// Recursively validate nested structs
			if fieldValue.Kind() == reflect.Struct {
				errors = append(errors, recursiveValidateFunc(fieldValue, fieldPath)...)
			}
		}
	}

	return errors
}
