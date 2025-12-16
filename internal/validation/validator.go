package validation

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/SmrutAI/Pedantigo/internal/constraints"
	"github.com/SmrutAI/Pedantigo/internal/tags"
)

// FieldError represents a validation error for a specific field
// This is a copy of the root package's FieldError to avoid circular imports
// FieldError represents an error condition.
type FieldError struct {
	Field   string
	Code    string // Machine-readable error code (e.g., "INVALID_EMAIL")
	Message string
	Value   any
}

// newFieldError creates a FieldError, extracting Code from ConstraintError if available.
func newFieldError(field string, err error, value any) FieldError {
	fe := FieldError{
		Field:   field,
		Message: err.Error(),
		Value:   value,
	}

	// Extract error code if the error is a ConstraintError
	var ce *constraints.ConstraintError
	if errors.As(err, &ce) {
		fe.Code = ce.Code
	}

	return fe
}

// ConstraintValidator is the interface for validation constraints.
type ConstraintValidator interface {
	Validate(value any) error
}

// TagParser is a function type for parsing struct tags with dive support.
type TagParser func(tag reflect.StructTag) *tags.ParsedTag

// ConstraintBuilder is a function type for building constraint validators.
type ConstraintBuilder func(constraints map[string]string, fieldType reflect.Type) []ConstraintValidator

// ValidateValue recursively validates a reflected value with dive support.
// Uses ParseTagWithDive to handle collection-level and element-level constraints.
// NOTE: 'required' constraint is skipped (not built in BuildConstraints).
func ValidateValue(
	val reflect.Value,
	path string,
	strictMissingFields bool,
	parseTagFunc TagParser,
	buildConstraintsFunc ConstraintBuilder,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var fieldErrs []FieldError

	// Handle pointer indirection
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fieldErrs // nil pointers are handled by required constraint
		}
		val = val.Elem()
	}

	// Only validate structs
	if val.Kind() != reflect.Struct {
		return fieldErrs
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

		// Parse validation tags with dive support
		parsedTag := parseTagFunc(field.Tag)
		if parsedTag == nil {
			// No validation tags, but still check nested structs, slices, and maps
			nestedErrors := validateNestedElements(fieldValue, recursiveValidateFunc, fieldPath)
			fieldErrs = append(fieldErrs, nestedErrors...)
			continue
		}

		// For nested structs (path != ""), check required fields
		if path != "" && strictMissingFields {
			// Check if "required" is in CollectionConstraints (before dive)
			if _, hasRequired := parsedTag.CollectionConstraints["required"]; hasRequired {
				// Check if field is zero value (indicates it was missing from JSON)
				if fieldValue.IsZero() {
					fieldErrs = append(fieldErrs, FieldError{
						Field:   fieldPath,
						Code:    constraints.CodeRequired,
						Message: "is required",
						Value:   fieldValue.Interface(),
					})
					// Skip further validation for this field
					continue
				}
			}
		}

		// Validate based on field kind
		isCollection := fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map
		isMap := fieldValue.Kind() == reflect.Map

		// Panic checks for invalid tag combinations
		if parsedTag.DivePresent && !isCollection {
			panic(fmt.Sprintf("field %s.%s: 'dive' can only be used on slice or map types, got %s",
				typ.Name(), field.Name, fieldValue.Kind()))
		}

		if len(parsedTag.KeyConstraints) > 0 && !isMap {
			panic(fmt.Sprintf("field %s.%s: 'keys' can only be used on map types, got %s",
				typ.Name(), field.Name, fieldValue.Kind()))
		}

		// Validate based on field type
		if isCollection {
			// Always apply collection-level constraints first (constraints before dive)
			if len(parsedTag.CollectionConstraints) > 0 {
				collectionValidators := buildConstraintsFunc(parsedTag.CollectionConstraints, field.Type)
				fieldErrs = append(fieldErrs, validateScalarField(fieldValue, fieldPath, collectionValidators, recursiveValidateFunc)...)
			}

			if parsedTag.DivePresent {
				// Dive into collection: validate elements with ElementConstraints
				elementValidators := buildConstraintsFunc(parsedTag.ElementConstraints, field.Type.Elem())
				if isMap {
					// Map with dive support
					fieldErrs = append(fieldErrs, validateMapElementsWithDive(
						fieldValue, fieldPath, elementValidators, parsedTag.KeyConstraints,
						buildConstraintsFunc, field.Type.Key(), recursiveValidateFunc)...)
				} else {
					// Slice with dive support
					fieldErrs = append(fieldErrs, validateSliceElements(fieldValue, fieldPath, elementValidators, recursiveValidateFunc)...)
				}
			} else {
				// No dive: still check nested structs in the collection
				nestedErrors := validateNestedElements(fieldValue, recursiveValidateFunc, fieldPath)
				fieldErrs = append(fieldErrs, nestedErrors...)
			}
		} else {
			// Non-collection field: apply constraints directly
			validators := buildConstraintsFunc(parsedTag.CollectionConstraints, field.Type)
			fieldErrs = append(fieldErrs, validateScalarField(fieldValue, fieldPath, validators, recursiveValidateFunc)...)
		}
	}

	return fieldErrs
}

func validateNestedElements(fieldValue reflect.Value,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
	fieldPath string,
) []FieldError {
	fieldErrors := make([]FieldError, 0)

	switch fieldValue.Kind() {
	case reflect.Struct:
		fieldErrors = append(fieldErrors, recursiveValidateFunc(fieldValue, fieldPath)...)
	case reflect.Slice:
		// Recursively validate struct elements in slices.
		for i := 0; i < fieldValue.Len(); i++ {
			elemValue := fieldValue.Index(i)
			elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			if elemValue.Kind() == reflect.Struct {
				fieldErrors = append(fieldErrors, recursiveValidateFunc(elemValue, elemPath)...)
			}
		}
	case reflect.Map:
		// Recursively validate struct values in maps.
		iter := fieldValue.MapRange()
		for iter.Next() {
			mapKey := iter.Key()
			mapValue := iter.Value()
			mapPath := fmt.Sprintf("%s[%v]", fieldPath, mapKey.Interface())
			if mapValue.Kind() == reflect.Struct {
				fieldErrors = append(fieldErrors, recursiveValidateFunc(mapValue, mapPath)...)
			}
		}
	}

	return fieldErrors
}

func validateSliceElements(
	fieldValue reflect.Value,
	fieldPath string,
	validators []ConstraintValidator,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var fieldErrs []FieldError
	for i := 0; i < fieldValue.Len(); i++ {
		elemValue := fieldValue.Index(i)
		elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		// Apply constraints to each element
		for _, validator := range validators {
			if err := validator.Validate(elemValue.Interface()); err != nil {
				fieldErrs = append(fieldErrs, newFieldError(elemPath, err, elemValue.Interface()))
			}
		}

		// Recursively validate nested structs in slice
		if elemValue.Kind() == reflect.Struct {
			fieldErrs = append(fieldErrs, recursiveValidateFunc(elemValue, elemPath)...)
		}
	}
	return fieldErrs
}

func validateMapElementsWithDive(
	fieldValue reflect.Value,
	fieldPath string,
	elementValidators []ConstraintValidator,
	keyConstraints map[string]string,
	buildConstraintsFunc ConstraintBuilder,
	keyType reflect.Type,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var fieldErrs []FieldError

	// Build key validators
	keyValidators := buildConstraintsFunc(keyConstraints, keyType)

	iter := fieldValue.MapRange()
	for iter.Next() {
		mapKey := iter.Key()
		mapValue := iter.Value()
		mapPath := fmt.Sprintf("%s[%v]", fieldPath, mapKey.Interface())

		// Validate keys if key constraints exist
		for _, validator := range keyValidators {
			if err := validator.Validate(mapKey.Interface()); err != nil {
				fieldErrs = append(fieldErrs, newFieldError(mapPath, err, mapKey.Interface()))
			}
		}

		// Validate values
		for _, validator := range elementValidators {
			if err := validator.Validate(mapValue.Interface()); err != nil {
				fieldErrs = append(fieldErrs, newFieldError(mapPath, err, mapValue.Interface()))
			}
		}

		// Recursively validate nested structs in map
		if mapValue.Kind() == reflect.Struct {
			fieldErrs = append(fieldErrs, recursiveValidateFunc(mapValue, mapPath)...)
		}
	}
	return fieldErrs
}

func validateScalarField(
	fieldValue reflect.Value,
	fieldPath string,
	validators []ConstraintValidator,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var fieldErrs []FieldError

	// Apply constraints directly
	for _, validator := range validators {
		if err := validator.Validate(fieldValue.Interface()); err != nil {
			fieldErrs = append(fieldErrs, newFieldError(fieldPath, err, fieldValue.Interface()))
		}
	}

	// Recursively validate nested structs
	if fieldValue.Kind() == reflect.Struct {
		fieldErrs = append(fieldErrs, recursiveValidateFunc(fieldValue, fieldPath)...)
	}

	return fieldErrs
}
