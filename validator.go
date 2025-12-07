package pedantigo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Validator validates structs of type T
type Validator[T any] struct {
	typ reflect.Type
}

// New creates a new Validator for type T
func New[T any]() *Validator[T] {
	var zero T
	return &Validator[T]{
		typ: reflect.TypeOf(zero),
	}
}

// Validate validates a struct and returns any validation errors
func (v *Validator[T]) Validate(obj *T) ValidationErrors {
	if obj == nil {
		return ValidationErrors{{Field: "root", Message: "cannot validate nil pointer"}}
	}

	var errors ValidationErrors

	// First, validate all fields using struct tags
	errors = append(errors, v.validateValue(reflect.ValueOf(obj).Elem(), "")...)

	// Then, check if struct implements Validatable for cross-field validation
	if validatable, ok := any(obj).(Validatable); ok {
		if err := validatable.Validate(); err != nil {
			// Check if it's a ValidationError
			if ve, ok := err.(ValidationError); ok {
				errors = append(errors, ve)
			} else {
				errors = append(errors, ValidationError{
					Field:   "root",
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

// validateValue recursively validates a reflected value
func (v *Validator[T]) validateValue(val reflect.Value, path string) ValidationErrors {
	var errors ValidationErrors

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
		constraints := parseTag(field.Tag)
		if constraints == nil {
			// No validation tags, but still check nested structs
			if fieldValue.Kind() == reflect.Struct {
				errors = append(errors, v.validateValue(fieldValue, fieldPath)...)
			}
			continue
		}

		// Build constraint validators
		validators := buildConstraints(constraints)

		// Apply each constraint to the field value
		for _, validator := range validators {
			if err := validator.Validate(fieldValue.Interface()); err != nil {
				errors = append(errors, ValidationError{
					Field:   fieldPath,
					Message: err.Error(),
					Value:   fieldValue.Interface(),
				})
			}
		}

		// Recursively validate nested structs
		if fieldValue.Kind() == reflect.Struct {
			errors = append(errors, v.validateValue(fieldValue, fieldPath)...)
		}
	}

	return errors
}

// Unmarshal unmarshals JSON data, applies defaults, and validates
func (v *Validator[T]) Unmarshal(data []byte) (*T, ValidationErrors) {
	var obj T

	// First, unmarshal the JSON
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, ValidationErrors{{
			Field:   "root",
			Message: fmt.Sprintf("JSON decode error: %v", err),
		}}
	}

	// Apply default values
	v.applyDefaults(reflect.ValueOf(&obj).Elem(), "")

	// Validate the unmarshaled object
	errors := v.Validate(&obj)

	return &obj, errors
}

// applyDefaults recursively applies default values to fields
func (v *Validator[T]) applyDefaults(val reflect.Value, path string) {
	// Handle pointer indirection
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	// Only process structs
	if val.Kind() != reflect.Struct {
		return
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

		// Parse validation tags
		constraints := parseTag(field.Tag)
		if constraints != nil {
			if defaultValue, hasDefault := constraints["default"]; hasDefault {
				// Only apply default if field is zero value
				if fieldValue.IsZero() {
					v.setDefaultValue(fieldValue, defaultValue)
				}
			}
		}

		// Recursively apply defaults to nested structs
		if fieldValue.Kind() == reflect.Struct {
			fieldPath := field.Name
			if path != "" {
				fieldPath = path + "." + field.Name
			}
			v.applyDefaults(fieldValue, fieldPath)
		}
	}
}

// setDefaultValue sets a default value on a field
func (v *Validator[T]) setDefaultValue(fieldValue reflect.Value, defaultValue string) {
	if !fieldValue.CanSet() {
		return
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(defaultValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(defaultValue, 10, 64); err == nil {
			fieldValue.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(defaultValue, 10, 64); err == nil {
			fieldValue.SetUint(u)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(defaultValue, 64); err == nil {
			fieldValue.SetFloat(f)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(defaultValue); err == nil {
			fieldValue.SetBool(b)
		}
	}
}

// Marshal validates and marshals struct to JSON
func (v *Validator[T]) Marshal(obj *T) ([]byte, ValidationErrors) {
	// TODO: implement validate + marshal
	return nil, nil
}
