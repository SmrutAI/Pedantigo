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
)

// Validator validates structs of type T.
type Validator[T any] struct {
	typ                reflect.Type
	options            ValidatorOptions
	fieldDeserializers map[string]deserialize.FieldDeserializer

	// Cross-field validation constraints
	fieldCache            *constraints.FieldCache
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
		fieldCache:            constraints.NewFieldCache(),
	}

	// Build field deserializers at creation time (fail-fast)
	validator.fieldDeserializers = deserialize.BuildFieldDeserializers(
		typ,
		deserialize.BuilderOptions{StrictMissingFields: options.StrictMissingFields},
		validator.setFieldValue,
		validator.setDefaultValue,
	)

	// Validate dive/keys/endkeys tag usage at creation time (fail-fast)
	validator.validateDiveTags(typ)

	// Build field constraints at creation time (the key optimization)
	validator.fieldCache = validator.buildFieldConstraints(typ)

	// Build cross-field constraints at creation time (fail-fast)
	validator.buildCrossFieldConstraints(typ)

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

// buildFieldConstraints builds and caches all field constraints at creation time.
func (v *Validator[T]) buildFieldConstraints(typ reflect.Type) *constraints.FieldCache {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	cache := constraints.NewFieldCache()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse tags once
		parsedTag := tags.ParseTagWithDive(field.Tag)

		// Field type info
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		isCollection := fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map
		isMap := fieldType.Kind() == reflect.Map

		cached := constraints.CachedField{
			Name:         field.Name,
			FieldIndex:   i,
			IsCollection: isCollection,
			IsMap:        isMap,
		}

		if parsedTag != nil {
			cached.HasDive = parsedTag.DivePresent

			// Check for required tag
			if _, hasRequired := parsedTag.CollectionConstraints["required"]; hasRequired {
				cached.IsRequired = true
			}

			// Constraints before dive (or regular field constraints)
			if len(parsedTag.CollectionConstraints) > 0 {
				cached.Constraints = constraints.BuildConstraints(parsedTag.CollectionConstraints, field.Type)
			}

			// Element constraints after dive
			if parsedTag.DivePresent && len(parsedTag.ElementConstraints) > 0 {
				cached.ElementConstraints = constraints.BuildConstraints(parsedTag.ElementConstraints, field.Type.Elem())
			}

			// Map key constraints
			if isMap && len(parsedTag.KeyConstraints) > 0 {
				cached.KeyConstraints = constraints.BuildConstraints(parsedTag.KeyConstraints, field.Type.Key())
			}
		}

		// Recurse for nested structs
		switch fieldType.Kind() {
		case reflect.Struct:
			cached.NestedCache = v.buildFieldConstraints(fieldType)
		case reflect.Slice, reflect.Map:
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				cached.NestedCache = v.buildFieldConstraints(elemType)
			}
		}

		cache.Fields = append(cache.Fields, cached)
	}

	return cache
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

// validateValue validates a struct value using cached constraints.
func (v *Validator[T]) validateValue(val reflect.Value, path string) []FieldError {
	return v.validateWithCache(val, path, v.fieldCache)
}

// validateWithCache validates using pre-built cached constraints.
func (v *Validator[T]) validateWithCache(val reflect.Value, path string, cache *constraints.FieldCache) []FieldError {
	if cache == nil {
		return nil
	}

	// Handle pointer indirection
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	var fieldErrors []FieldError

	for i := range cache.Fields {
		cached := &cache.Fields[i]
		fieldVal := val.Field(cached.FieldIndex)

		// Build field path
		fieldPath := cached.Name
		if path != "" {
			fieldPath = path + "." + cached.Name
		}

		// Check required for nested struct fields (path != "")
		if path != "" && v.options.StrictMissingFields && cached.IsRequired {
			if fieldVal.IsZero() {
				fieldErrors = append(fieldErrors, FieldError{
					Field:   fieldPath,
					Code:    constraints.CodeRequired,
					Message: "is required",
					Value:   fieldVal.Interface(),
				})
				continue // Skip further validation for this field
			}
		}

		// Apply field constraints
		for _, c := range cached.Constraints {
			if err := c.Validate(fieldVal.Interface()); err != nil {
				fieldErrors = append(fieldErrors, v.newFieldError(fieldPath, err, fieldVal.Interface()))
			}
		}

		// Handle collections with dive (requires dive to recurse into elements, like playground)
		if cached.IsCollection && cached.HasDive {
			if cached.IsMap {
				fieldErrors = append(fieldErrors, v.validateMapWithCache(fieldVal, fieldPath, cached)...)
			} else {
				fieldErrors = append(fieldErrors, v.validateSliceWithCache(fieldVal, fieldPath, cached)...)
			}
		} else if cached.NestedCache != nil && !cached.IsCollection {
			// Recurse for nested structs (but NOT collection elements without dive)
			fieldErrors = append(fieldErrors, v.validateWithCache(fieldVal, fieldPath, cached.NestedCache)...)
		}
	}

	return fieldErrors
}

// validateSliceWithCache validates slice elements using cached constraints.
func (v *Validator[T]) validateSliceWithCache(val reflect.Value, path string, cached *constraints.CachedField) []FieldError {
	var fieldErrors []FieldError

	for i := 0; i < val.Len(); i++ {
		elemVal := val.Index(i)
		elemPath := fmt.Sprintf("%s[%d]", path, i)

		// Apply element constraints
		for _, c := range cached.ElementConstraints {
			if err := c.Validate(elemVal.Interface()); err != nil {
				fieldErrors = append(fieldErrors, v.newFieldError(elemPath, err, elemVal.Interface()))
			}
		}

		// Recurse for nested structs
		if cached.NestedCache != nil {
			fieldErrors = append(fieldErrors, v.validateWithCache(elemVal, elemPath, cached.NestedCache)...)
		}
	}

	return fieldErrors
}

// validateMapWithCache validates map entries using cached constraints.
func (v *Validator[T]) validateMapWithCache(val reflect.Value, path string, cached *constraints.CachedField) []FieldError {
	var fieldErrors []FieldError

	iter := val.MapRange()
	for iter.Next() {
		mapKey := iter.Key()
		mapVal := iter.Value()
		elemPath := fmt.Sprintf("%s[%v]", path, mapKey.Interface())

		// Apply key constraints
		for _, c := range cached.KeyConstraints {
			if err := c.Validate(mapKey.Interface()); err != nil {
				fieldErrors = append(fieldErrors, v.newFieldError(elemPath, err, mapKey.Interface()))
			}
		}

		// Apply value constraints
		for _, c := range cached.ElementConstraints {
			if err := c.Validate(mapVal.Interface()); err != nil {
				fieldErrors = append(fieldErrors, v.newFieldError(elemPath, err, mapVal.Interface()))
			}
		}

		// Recurse for nested structs
		if cached.NestedCache != nil {
			fieldErrors = append(fieldErrors, v.validateWithCache(mapVal, elemPath, cached.NestedCache)...)
		}
	}

	return fieldErrors
}

// newFieldError creates a FieldError, extracting Code from ConstraintError if available.
func (v *Validator[T]) newFieldError(field string, err error, value any) FieldError {
	fe := FieldError{
		Field:   field,
		Message: err.Error(),
		Value:   value,
	}

	var ce *constraints.ConstraintError
	if errors.As(err, &ce) {
		fe.Code = ce.Code
	}

	return fe
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
