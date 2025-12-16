package pedantigo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/invopop/jsonschema"

	"github.com/SmrutAI/pedantigo/internal/constraints"
	"github.com/SmrutAI/pedantigo/internal/deserialize"
	"github.com/SmrutAI/pedantigo/internal/serialize"
	"github.com/SmrutAI/pedantigo/internal/tags"
	"github.com/SmrutAI/pedantigo/internal/validation"
)

// Validator validates structs of type T.
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

// New creates a new Validator for type T with optional configuration.
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

	// Validate dive/keys/endkeys tag usage at creation time (fail-fast)
	validator.validateDiveTags(typ)

	return validator
}

// buildCrossFieldConstraints builds cross-field constraints for all struct fields.
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

// validateDiveTags validates that dive/keys/endkeys tags are used correctly.
// This is called at creation time to fail fast on invalid tag combinations.
func (v *Validator[T]) validateDiveTags(typ reflect.Type) {
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

		// Parse the tag with dive support
		parsedTag := tags.ParseTagWithDive(field.Tag)
		if parsedTag == nil {
			continue
		}

		// Get the underlying field type (dereference pointers)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		isCollection := fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map
		isMap := fieldType.Kind() == reflect.Map

		// Panic: dive on non-collection field
		if parsedTag.DivePresent && !isCollection {
			panic(fmt.Sprintf("field %s.%s: 'dive' can only be used on slice or map types, got %s",
				typ.Name(), field.Name, fieldType.Kind()))
		}

		// Panic: keys on non-map field
		if len(parsedTag.KeyConstraints) > 0 && !isMap {
			panic(fmt.Sprintf("field %s.%s: 'keys' can only be used on map types, got %s",
				typ.Name(), field.Name, fieldType.Kind()))
		}

		// Panic: unique on non-collection field
		if _, hasUnique := parsedTag.CollectionConstraints["unique"]; hasUnique && !isCollection {
			panic(fmt.Sprintf("field %s.%s: 'unique' can only be used on slice or map types, got %s",
				typ.Name(), field.Name, fieldType.Kind()))
		}

		// Recursively validate nested structs
		switch fieldType.Kind() {
		case reflect.Struct:
			v.validateDiveTags(fieldType)
		case reflect.Slice:
			if fieldType.Elem().Kind() == reflect.Struct {
				v.validateDiveTags(fieldType.Elem())
			}
		case reflect.Map:
			if fieldType.Elem().Kind() == reflect.Struct {
				v.validateDiveTags(fieldType.Elem())
			}
		}
	}
}

// setFieldValue wraps the deserialize package SetFieldValue for use in validator.
func (v *Validator[T]) setFieldValue(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error {
	return deserialize.SetFieldValue(fieldValue, inValue, fieldType, v.setFieldValue)
}

// Validate validates a struct and returns any validation errors
// NOTE: 'required' is NOT checked here - it's only checked during Unmarshal
// Validate checks if the value satisfies the constraint.
func (v *Validator[T]) Validate(obj *T) error {
	if obj == nil {
		return &ValidationError{
			Errors: []FieldError{{Field: "root", Message: "cannot validate nil pointer"}},
		}
	}

	var fieldErrors []FieldError

	// Validate all fields using struct tags (required is skipped via buildConstraints)
	fieldErrors = append(fieldErrors, v.validateValue(reflect.ValueOf(obj).Elem(), "")...)

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
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					fieldErrors = append(fieldErrors, valErr.Errors...)
				} else {
					fieldErrors = append(fieldErrors, FieldError{
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
			var ve *ValidationError
			if errors.As(err, &ve) {
				fieldErrors = append(fieldErrors, ve.Errors...)
			} else {
				// Single error or custom error type
				fieldErrors = append(fieldErrors, FieldError{
					Field:   "root",
					Message: err.Error(),
				})
			}
		}
	}

	if len(fieldErrors) == 0 {
		return nil
	}

	return &ValidationError{Errors: fieldErrors}
}

// validateValue wraps the validation package ValidateValue for use in validator.
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
				Code:    e.Code,
				Message: e.Message,
				Value:   e.Value,
			}
		}
		return validationErrors
	}

	// Call the recursive validate with the wrapper
	return v.validateValueInternal(val, path, recursiveValidate)
}

// validateValueInternal is the actual implementation that uses validation.ValidateValue.
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
		tags.ParseTagWithDive,
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
	fieldErrors := make([]FieldError, len(validationErrors))
	for i, e := range validationErrors {
		fieldErrors[i] = FieldError{
			Field:   e.Field,
			Code:    e.Code,
			Message: e.Message,
			Value:   e.Value,
		}
	}
	return fieldErrors
}

// Unmarshal unmarshals JSON data, applies defaults, and validates.
func (v *Validator[T]) Unmarshal(data []byte) (*T, error) {
	// Fast path: skip 2-step flow if StrictMissingFields is disabled
	if !v.options.StrictMissingFields {
		var obj T

		// Use json.Decoder with DisallowUnknownFields for ExtraForbid
		if v.options.ExtraFields == ExtraForbid {
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&obj); err != nil {
				return &obj, &ValidationError{
					Errors: []FieldError{{
						Field:   "root",
						Message: "JSON decode error: " + ErrMsgUnknownField,
					}},
				}
			}
		} else {
			if err := json.Unmarshal(data, &obj); err != nil {
				return nil, &ValidationError{
					Errors: []FieldError{{
						Field:   "root",
						Message: fmt.Sprintf("JSON decode error: %v", err),
					}},
				}
			}
		}

		// Only run validators (skip required checks and defaults)
		if err := v.Validate(&obj); err != nil {
			return &obj, err
		}
		return &obj, nil
	}

	// Step 0.5: Pre-check for extra fields if ExtraForbid is set (handles nested structs)
	if v.options.ExtraFields == ExtraForbid {
		var obj T
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&obj); err != nil {
			return &obj, &ValidationError{
				Errors: []FieldError{{
					Field:   "root",
					Message: ErrMsgUnknownField,
				}},
			}
		}
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
	var fieldErrors []FieldError
	for fieldName, deserializer := range v.fieldDeserializers {
		var inValue any
		if val, exists := jsonMap[fieldName]; exists {
			inValue = val // Field present in JSON (might be nil for JSON null)
		} else {
			inValue = deserialize.FieldMissingSentinel // Field missing from JSON
		}

		if err := deserializer(&objValue, inValue); err != nil {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   fieldName,
				Message: err.Error(),
			})
		}
	}

	// Return early if deserialization errors
	if len(fieldErrors) > 0 {
		return &obj, &ValidationError{Errors: fieldErrors}
	}

	// Step 4: Run validation constraints (min, max, email, etc.)
	// NOTE: 'required' is already skipped in Validate() via buildConstraints
	if err := v.Validate(&obj); err != nil {
		return &obj, err
	}

	return &obj, nil
}

// setDefaultValue wraps the deserialize package SetDefaultValue for use in validator.
func (v *Validator[T]) setDefaultValue(fieldValue reflect.Value, defaultValue string) {
	deserialize.SetDefaultValue(fieldValue, defaultValue, v.setDefaultValue)
}

// Marshal validates and marshals struct to JSON.
func (v *Validator[T]) Marshal(obj *T) ([]byte, error) {
	// Validate before marshaling
	if err := v.Validate(obj); err != nil {
		return nil, err
	}

	// Marshal to JSON
	return json.Marshal(obj)
}

// MarshalWithOptions validates and marshals struct to JSON with options.
// Options allow context-based field exclusion and omitzero behavior.
func (v *Validator[T]) MarshalWithOptions(obj *T, opts MarshalOptions) ([]byte, error) {
	// Validate before marshaling
	if err := v.Validate(obj); err != nil {
		return nil, err
	}

	// Build field metadata for filtering
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return []byte("null"), nil
		}
		val = val.Elem()
	}

	metadata := serialize.BuildFieldMetadata(val.Type())

	// Convert options
	serializeOpts := serialize.SerializeOptions{
		Context:  opts.Context,
		OmitZero: opts.OmitZero,
	}

	// Convert to filtered map
	filtered := serialize.ToFilteredMap(val, metadata, serializeOpts)

	// Marshal the filtered map
	return json.Marshal(filtered)
}

// Dict converts the object into a dict.
func (v *Validator[T]) Dict(obj *T) (map[string]interface{}, error) {
	data, _ := json.Marshal(obj)
	var dict map[string]interface{}
	if err := json.Unmarshal(data, &dict); err != nil {
		return nil, err
	}
	return dict, nil
}

// NewModel creates a validated instance of T from various input types.
// Accepts: []byte (JSON), T (struct), *T (pointer), or map[string]any (kwargs).
// This is the unified constructor that validates regardless of input source.
func (v *Validator[T]) NewModel(input any) (*T, error) {
	switch val := input.(type) {
	case []byte:
		return v.Unmarshal(val)
	case *T:
		if val == nil {
			return nil, &ValidationError{
				Errors: []FieldError{{Field: "root", Message: "cannot validate nil pointer"}},
			}
		}
		if err := v.Validate(val); err != nil {
			return val, err
		}
		return val, nil
	case map[string]any:
		return v.unmarshalFromMap(val)
	case T:
		if err := v.Validate(&val); err != nil {
			return &val, err
		}
		return &val, nil
	default:
		var zero T
		return nil, &ValidationError{
			Errors: []FieldError{{
				Field:   "root",
				Message: fmt.Sprintf("unsupported input type: %T, expected []byte, %T, *%T, or map[string]any", input, zero, zero),
			}},
		}
	}
}

// unmarshalFromMap creates a validated struct from a map (kwargs pattern).
// Reuses the same deserialization logic as Unmarshal.
func (v *Validator[T]) unmarshalFromMap(jsonMap map[string]any) (*T, error) {
	// Create new struct instance
	var obj T
	objValue := reflect.ValueOf(&obj).Elem()

	// Apply field deserializers (same logic as Unmarshal)
	var fieldErrors []FieldError
	for fieldName, deserializer := range v.fieldDeserializers {
		var inValue any
		if val, exists := jsonMap[fieldName]; exists {
			inValue = val
		} else {
			inValue = deserialize.FieldMissingSentinel
		}

		if err := deserializer(&objValue, inValue); err != nil {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   fieldName,
				Message: err.Error(),
			})
		}
	}

	if len(fieldErrors) > 0 {
		return &obj, &ValidationError{Errors: fieldErrors}
	}

	// Run validation constraints
	if err := v.Validate(&obj); err != nil {
		return &obj, err
	}

	return &obj, nil
}
