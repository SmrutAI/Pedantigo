package pedantigo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/SmrutAI/Pedantigo/internal/constraints"
	"github.com/SmrutAI/Pedantigo/internal/deserialize"
	"github.com/SmrutAI/Pedantigo/internal/tags"
	"github.com/SmrutAI/Pedantigo/internal/validation"
	"github.com/invopop/jsonschema"
)

// Validator validates structs of type T
type Validator[T any] struct {
	typ                reflect.Type
	options            ValidatorOptions
	fieldDeserializers map[string]deserialize.FieldDeserializer

	// Cross-field validation constraints
	fieldCrossConstraints map[string][]constraints.CrossFieldConstraint

	// Schema caching (lazy initialization with double-checked locking)
	schemaMu          sync.RWMutex
	cachedSchema      *jsonschema.Schema // Schema() result
	cachedSchemaJSON  []byte             // SchemaJSON() result
	cachedOpenAPI     *jsonschema.Schema // SchemaOpenAPI() result
	cachedOpenAPIJSON []byte             // SchemaJSONOpenAPI() result
}

// New creates a new Validator for type T with optional configuration
func New[T any](opts ...ValidatorOptions) *Validator[T] {
	var zero T
	typ := reflect.TypeOf(zero)

	options := DefaultValidatorOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	validator := &Validator[T]{
		typ:                   typ,
		options:               options,
		fieldDeserializers:    make(map[string]deserialize.FieldDeserializer),
		fieldCrossConstraints: make(map[string][]constraints.CrossFieldConstraint),
	}

	// Build field deserializers at creation time (fail-fast)
	validator.fieldDeserializers = deserialize.BuildFieldDeserializers(
		typ,
		deserialize.BuilderOptions{StrictMissingFields: options.StrictMissingFields},
		validator.setFieldValue,
		validator.setDefaultValue,
	)

	// Build cross-field constraints at creation time (fail-fast)
	validator.buildCrossFieldConstraints(typ)

	return validator
}

// buildCrossFieldConstraints builds cross-field constraints for all struct fields
func (v *Validator[T]) buildCrossFieldConstraints(typ reflect.Type) {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse pedantigo validation constraints
		constraintsMap := tags.ParseTag(field.Tag)
		if constraintsMap == nil {
			continue
		}

		// Build cross-field constraints for this field (use struct field name, not JSON name)
		crossConstraints := constraints.BuildCrossFieldConstraintsForField(constraintsMap, typ, i)
		if len(crossConstraints) > 0 {
			v.fieldCrossConstraints[field.Name] = crossConstraints
		}
	}
}

// setFieldValue wraps the deserialize package SetFieldValue for use in validator
func (v *Validator[T]) setFieldValue(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error {
	return deserialize.SetFieldValue(fieldValue, inValue, fieldType, v.setFieldValue)
}

// Validate validates a struct and returns any validation errors
// NOTE: 'required' is NOT checked here - it's only checked during Unmarshal
func (v *Validator[T]) Validate(obj *T) error {
	if obj == nil {
		return &ValidationError{
			Errors: []FieldError{{Field: "root", Message: "cannot validate nil pointer"}},
		}
	}

	var errors []FieldError

	// Validate all fields using struct tags (required is skipped via buildConstraints)
	errors = append(errors, v.validateValue(reflect.ValueOf(obj).Elem(), "")...)

	// Run cross-field validation
	structValue := reflect.ValueOf(obj).Elem()
	for fieldName, crossConstraints := range v.fieldCrossConstraints {
		// Get field value by struct field name
		field := structValue.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}
		fieldValue := field.Interface()

		// Run each cross-field constraint
		for _, constraint := range crossConstraints {
			if err := constraint.ValidateCrossField(fieldValue, structValue, fieldName); err != nil {
				if valErr, ok := err.(*ValidationError); ok {
					errors = append(errors, valErr.Errors...)
				} else {
					errors = append(errors, FieldError{
						Field:   fieldName,
						Message: err.Error(),
					})
				}
			}
		}
	}

	// Then, check if struct implements Validatable for cross-field validation
	if validatable, ok := any(obj).(Validatable); ok {
		if err := validatable.Validate(); err != nil {
			// Check if it's a ValidationError with multiple errors
			if ve, ok := err.(*ValidationError); ok {
				errors = append(errors, ve.Errors...)
			} else {
				// Single error or custom error type
				errors = append(errors, FieldError{
					Field:   "root",
					Message: err.Error(),
				})
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return &ValidationError{Errors: errors}
}

// validateValue wraps the validation package ValidateValue for use in validator
func (v *Validator[T]) validateValue(val reflect.Value, path string) []FieldError {
	// Create a recursive wrapper that converts return types
	var recursiveValidate func(reflect.Value, string) []validation.FieldError
	recursiveValidate = func(val reflect.Value, path string) []validation.FieldError {
		// Call the actual validation logic
		pedantigoErrors := v.validateValueInternal(val, path, recursiveValidate)

		// Convert pedantigo.FieldError to validation.FieldError
		validationErrors := make([]validation.FieldError, len(pedantigoErrors))
		for i, e := range pedantigoErrors {
			validationErrors[i] = validation.FieldError{
				Field:   e.Field,
				Message: e.Message,
				Value:   e.Value,
			}
		}
		return validationErrors
	}

	// Call the recursive validate with the wrapper
	return v.validateValueInternal(val, path, recursiveValidate)
}

// validateValueInternal is the actual implementation that uses validation.ValidateValue
func (v *Validator[T]) validateValueInternal(
	val reflect.Value,
	path string,
	recursiveValidateFunc func(reflect.Value, string) []validation.FieldError,
) []FieldError {
	// Call validation.ValidateValue with all required parameters
	validationErrors := validation.ValidateValue(
		val,
		path,
		v.options.StrictMissingFields,
		tags.ParseTag,
		func(constraintsMap map[string]string, fieldType reflect.Type) []validation.ConstraintValidator {
			// Wrap constraints.Constraint to validation.ConstraintValidator
			constraintList := constraints.BuildConstraints(constraintsMap, fieldType)
			result := make([]validation.ConstraintValidator, len(constraintList))
			for i, c := range constraintList {
				result[i] = c
			}
			return result
		},
		recursiveValidateFunc,
	)

	// Convert validation.FieldError to pedantigo.FieldError
	errors := make([]FieldError, len(validationErrors))
	for i, e := range validationErrors {
		errors[i] = FieldError{
			Field:   e.Field,
			Message: e.Message,
			Value:   e.Value,
		}
	}
	return errors
}

// Unmarshal unmarshals JSON data, applies defaults, and validates
func (v *Validator[T]) Unmarshal(data []byte) (*T, error) {
	// Fast path: skip 2-step flow if StrictMissingFields is disabled
	if !v.options.StrictMissingFields {
		var obj T
		if err := json.Unmarshal(data, &obj); err != nil {
			return nil, &ValidationError{
				Errors: []FieldError{{
					Field:   "root",
					Message: fmt.Sprintf("JSON decode error: %v", err),
				}},
			}
		}

		// Only run validators (skip required checks and defaults)
		if err := v.Validate(&obj); err != nil {
			return &obj, err
		}
		return &obj, nil
	}

	// Step 1: Unmarshal to map[string]any to detect which fields exist
	var jsonMap map[string]any
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, &ValidationError{
			Errors: []FieldError{{
				Field:   "root",
				Message: fmt.Sprintf("JSON decode error: %v", err),
			}},
		}
	}

	// Step 2: Create new struct instance
	var obj T
	objValue := reflect.ValueOf(&obj).Elem()

	// Step 3: Apply field deserializers
	var errors []FieldError
	for fieldName, deserializer := range v.fieldDeserializers {
		var inValue any
		if val, exists := jsonMap[fieldName]; exists {
			inValue = val // Field present in JSON (might be nil for JSON null)
		} else {
			inValue = deserialize.FieldMissingSentinel // Field missing from JSON
		}

		if err := deserializer(&objValue, inValue); err != nil {
			errors = append(errors, FieldError{
				Field:   fieldName,
				Message: err.Error(),
			})
		}
	}

	// Return early if deserialization errors
	if len(errors) > 0 {
		return &obj, &ValidationError{Errors: errors}
	}

	// Step 4: Run validation constraints (min, max, email, etc.)
	// NOTE: 'required' is already skipped in Validate() via buildConstraints
	if err := v.Validate(&obj); err != nil {
		return &obj, err
	}

	return &obj, nil
}

// setDefaultValue wraps the deserialize package SetDefaultValue for use in validator
func (v *Validator[T]) setDefaultValue(fieldValue reflect.Value, defaultValue string) {
	deserialize.SetDefaultValue(fieldValue, defaultValue, v.setDefaultValue)
}

// Marshal validates and marshals struct to JSON
func (v *Validator[T]) Marshal(obj *T) ([]byte, error) {
	// Validate before marshaling
	if err := v.Validate(obj); err != nil {
		return nil, err
	}

	// Marshal to JSON
	return json.Marshal(obj)
}
