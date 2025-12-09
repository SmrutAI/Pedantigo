package pedantigo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SmrutAI/Pedantigo/constraints"
	"github.com/invopop/jsonschema"
)

// ValidatorOptions configures validator behavior
type ValidatorOptions struct {
	// StrictMissingFields controls whether missing fields without defaults are errors
	// When true (default): missing fields without defaults cause validation errors
	// When false: missing fields are left as zero values (user handles with pointers)
	StrictMissingFields bool
}

// DefaultValidatorOptions returns the default validator options
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		StrictMissingFields: true,
	}
}

// missingFieldSentinel is a sentinel value to distinguish missing fields from explicit null
type missingFieldSentinel struct{}

var fieldMissingSentinel = missingFieldSentinel{}

// fieldDeserializer is a closure that deserializes a single field
// inValue is fieldMissingSentinel if field is missing from JSON,
// nil if field is explicitly null, or the actual value if present
type fieldDeserializer func(outPtr *reflect.Value, inValue any) error

// Validator validates structs of type T
type Validator[T any] struct {
	typ                reflect.Type
	options            ValidatorOptions
	fieldDeserializers map[string]fieldDeserializer

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
		fieldDeserializers:    make(map[string]fieldDeserializer),
		fieldCrossConstraints: make(map[string][]constraints.CrossFieldConstraint),
	}

	// Build field deserializers at creation time (fail-fast)
	validator.buildFieldDeserializers(typ)

	// Build cross-field constraints at creation time (fail-fast)
	validator.buildCrossFieldConstraints(typ)

	return validator
}

// buildFieldDeserializers creates field deserializer closures for each struct field
func (v *Validator[T]) buildFieldDeserializers(typ reflect.Type) {
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

		// Get JSON field name
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name
		if jsonTag != "" && jsonTag != "-" {
			// Extract field name from json tag (before comma)
			if name, _, found := strings.Cut(jsonTag, ","); found {
				fieldName = name
			} else {
				fieldName = jsonTag
			}
		}

		// Parse validation constraints
		constraints := parseTag(field.Tag)

		// Safety check: panic if default tags are used when StrictMissingFields is disabled
		if constraints != nil && !v.options.StrictMissingFields {
			if _, hasDefault := constraints["default"]; hasDefault {
				panic(fmt.Sprintf("field %s.%s has 'default=' tag but StrictMissingFields is false. Remove the tag or enable StrictMissingFields.",
					typ.Name(), field.Name))
			}
			if _, hasMethod := constraints["defaultUsingMethod"]; hasMethod {
				panic(fmt.Sprintf("field %s.%s has 'defaultUsingMethod=' tag but StrictMissingFields is false. Remove the tag or enable StrictMissingFields.",
					typ.Name(), field.Name))
			}
		}

		// Get default value and defaultUsingMethod
		var staticDefault *string
		var methodName *string

		if constraints != nil {
			if defVal, hasDefault := constraints["default"]; hasDefault {
				staticDefault = &defVal
			}
			if method, hasMethod := constraints["defaultUsingMethod"]; hasMethod {
				methodName = &method
				// Validate method exists and has correct signature (fail-fast)
				if err := v.validateDefaultMethod(typ, method, field.Type); err != nil {
					panic(fmt.Sprintf("field %s: %v", field.Name, err))
				}
			}
		}

		// Create field deserializer closure
		fieldIndex := i
		fieldType := field.Type
		_, hasRequired := constraints["required"] // Check if key exists, not if value is non-empty

		v.fieldDeserializers[fieldName] = func(outPtr *reflect.Value, inValue any) error {
			fieldValue := outPtr.Field(fieldIndex)

			// Determine if field was present in JSON
			_, fieldMissing := inValue.(missingFieldSentinel)

			if fieldMissing {
				// Field is missing from JSON
				if staticDefault != nil {
					// Apply static default
					v.setDefaultValue(fieldValue, *staticDefault)
					return nil
				}

				if methodName != nil {
					// Call defaultUsingMethod on the pointer receiver
					// outPtr is the struct value, so Addr() gives us the pointer
					ptrValue := outPtr.Addr()
					method := ptrValue.MethodByName(*methodName)
					results := method.Call(nil)
					if len(results) == 2 {
						// Check for error
						if !results[1].IsNil() {
							return results[1].Interface().(error)
						}
						// Set the value
						fieldValue.Set(results[0])
					}
					return nil
				}

				if hasRequired && v.options.StrictMissingFields {
					return fmt.Errorf("is required")
				}

				// Leave as zero value (relaxed mode or not required)
				return nil
			}

			// Field is present in JSON - set the value
			return v.setFieldValue(fieldValue, inValue, fieldType)
		}
	}
}

// validateDefaultMethod checks that a method exists and has the correct signature
func (v *Validator[T]) validateDefaultMethod(structType reflect.Type, methodName string, fieldType reflect.Type) error {
	// Look for the method on the pointer type (methods are typically defined on pointer receivers)
	ptrType := reflect.PointerTo(structType)
	method, found := ptrType.MethodByName(methodName)

	if !found {
		return fmt.Errorf("method %s not found on type %s", methodName, structType.Name())
	}

	methodType := method.Type
	// Method signature should be: func(*T) (FieldType, error)
	// methodType.NumIn() includes the receiver, so we expect 1 (just receiver)
	if methodType.NumIn() != 1 {
		return fmt.Errorf("method %s should take no arguments (only receiver), got %d arguments", methodName, methodType.NumIn()-1)
	}

	if methodType.NumOut() != 2 {
		return fmt.Errorf("method %s should return (value, error), got %d return values", methodName, methodType.NumOut())
	}

	// Check return types
	if methodType.Out(0) != fieldType {
		return fmt.Errorf("method %s should return %s as first value, got %s", methodName, fieldType, methodType.Out(0))
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !methodType.Out(1).Implements(errorInterface) {
		return fmt.Errorf("method %s should return error as second value, got %s", methodName, methodType.Out(1))
	}

	return nil
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
		constraintsMap := parseTag(field.Tag)
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

// setFieldValue sets a field value from a JSON value
func (v *Validator[T]) setFieldValue(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error {
	if !fieldValue.CanSet() {
		return nil
	}

	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		// If inValue is nil, set the pointer field to nil (explicit JSON null)
		if inValue == nil {
			fieldValue.Set(reflect.Zero(fieldType))
			return nil
		}

		// Allocate new pointer of the element type
		elemType := fieldType.Elem()
		newPtr := reflect.New(elemType)

		// Recursively set the value on the dereferenced pointer
		if err := v.setFieldValue(newPtr.Elem(), inValue, elemType); err != nil {
			return err
		}

		// Set the field to the new pointer
		fieldValue.Set(newPtr)
		return nil
	}

	// Handle nil values for slices
	if inValue == nil && fieldType.Kind() == reflect.Slice {
		fieldValue.Set(reflect.Zero(fieldType))
		return nil
	}

	// Handle nil values for maps
	if inValue == nil && fieldType.Kind() == reflect.Map {
		fieldValue.Set(reflect.Zero(fieldType))
		return nil
	}

	// Convert inValue to the correct type
	inVal := reflect.ValueOf(inValue)

	// Handle time.Time special case
	// When unmarshaling to map[string]any, time values remain as strings
	// We need to parse them manually (mimicking what encoding/json does automatically)
	if fieldType == reflect.TypeOf(time.Time{}) {
		if inVal.Kind() == reflect.String {
			// Parse RFC3339 format (same as Go's encoding/json package)
			t, err := time.Parse(time.RFC3339, inVal.String())
			if err != nil {
				return fmt.Errorf("failed to parse time: %v", err)
			}
			fieldValue.Set(reflect.ValueOf(t))
			return nil
		}
	}

	// Handle nested structs: if inValue is map[string]any and target is struct
	if inVal.Kind() == reflect.Map && fieldType.Kind() == reflect.Struct {
		// Re-marshal the map and unmarshal into the struct
		jsonBytes, err := json.Marshal(inValue)
		if err != nil {
			return fmt.Errorf("failed to marshal nested struct: %v", err)
		}

		// Create a new instance of the target type
		newStruct := reflect.New(fieldType)
		if err := json.Unmarshal(jsonBytes, newStruct.Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal nested struct: %v", err)
		}

		fieldValue.Set(newStruct.Elem())
		return nil
	}

	// Handle slices: if inValue is []any and target is slice
	if inVal.Kind() == reflect.Slice && fieldType.Kind() == reflect.Slice {
		elemType := fieldType.Elem()
		newSlice := reflect.MakeSlice(fieldType, inVal.Len(), inVal.Len())

		for i := 0; i < inVal.Len(); i++ {
			elemValue := newSlice.Index(i)
			elemInput := inVal.Index(i).Interface()

			// For structs in slices, manually deserialize fields to track which are present
			if elemType.Kind() == reflect.Struct && reflect.TypeOf(elemInput).Kind() == reflect.Map {
				inputMap, ok := elemInput.(map[string]any)
				if !ok {
					return fmt.Errorf("expected map for struct element")
				}

				// Create new struct instance
				newStruct := reflect.New(elemType).Elem()

				// Iterate through struct fields and set values
				for j := 0; j < elemType.NumField(); j++ {
					field := elemType.Field(j)

					// Skip unexported fields
					if !field.IsExported() {
						continue
					}

					// Get JSON field name
					jsonTag := field.Tag.Get("json")
					jsonFieldName := field.Name
					if jsonTag != "" && jsonTag != "-" {
						if name, _, found := strings.Cut(jsonTag, ","); found {
							jsonFieldName = name
						} else {
							jsonFieldName = jsonTag
						}
					}

					// Check if field exists in JSON
					val, exists := inputMap[jsonFieldName]
					if !exists {
						// Field missing from JSON - leave as zero value
						// Will be checked for 'required' later in validateValue()
						continue
					}

					// Set the field value
					fieldVal := newStruct.Field(j)
					if err := v.setFieldValue(fieldVal, val, field.Type); err != nil {
						return err
					}
				}

				elemValue.Set(newStruct)
			} else {
				if err := v.setFieldValue(elemValue, elemInput, elemType); err != nil {
					return err
				}
			}
		}

		fieldValue.Set(newSlice)
		return nil
	}

	// Handle maps: if inValue is map[string]any and target is map
	if inVal.Kind() == reflect.Map && fieldType.Kind() == reflect.Map {
		keyType := fieldType.Key()
		valueType := fieldType.Elem()

		// Create new map
		newMap := reflect.MakeMap(fieldType)

		// Iterate through map entries
		iter := inVal.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value().Interface()

			// Convert key if needed
			var convertedKey reflect.Value
			if key.Type().AssignableTo(keyType) {
				convertedKey = key
			} else if key.Type().ConvertibleTo(keyType) {
				convertedKey = key.Convert(keyType)
			} else {
				return fmt.Errorf("cannot convert map key %v to %v", key.Type(), keyType)
			}

			// For struct values in maps, manually deserialize fields to track which are present
			if valueType.Kind() == reflect.Struct && reflect.TypeOf(val).Kind() == reflect.Map {
				inputMap, ok := val.(map[string]any)
				if !ok {
					return fmt.Errorf("expected map for struct value")
				}

				// Create new struct instance
				newStruct := reflect.New(valueType).Elem()

				// Iterate through struct fields and set values
				for j := 0; j < valueType.NumField(); j++ {
					field := valueType.Field(j)

					// Skip unexported fields
					if !field.IsExported() {
						continue
					}

					// Get JSON field name
					jsonTag := field.Tag.Get("json")
					jsonFieldName := field.Name
					if jsonTag != "" && jsonTag != "-" {
						if name, _, found := strings.Cut(jsonTag, ","); found {
							jsonFieldName = name
						} else {
							jsonFieldName = jsonTag
						}
					}

					// Check if field exists in JSON
					fieldVal, exists := inputMap[jsonFieldName]
					if !exists {
						// Field missing from JSON - leave as zero value
						// Will be checked for 'required' later in validateValue()
						continue
					}

					// Set the field value
					structFieldVal := newStruct.Field(j)
					if err := v.setFieldValue(structFieldVal, fieldVal, field.Type); err != nil {
						return err
					}
				}

				newMap.SetMapIndex(convertedKey, newStruct)
			} else {
				// For non-struct values, convert normally
				newValue := reflect.New(valueType).Elem()
				if err := v.setFieldValue(newValue, val, valueType); err != nil {
					return err
				}
				newMap.SetMapIndex(convertedKey, newValue)
			}
		}

		fieldValue.Set(newMap)
		return nil
	}

	// Handle type conversion
	if inVal.Type().AssignableTo(fieldType) {
		fieldValue.Set(inVal)
	} else if inVal.Type().ConvertibleTo(fieldType) {
		fieldValue.Set(inVal.Convert(fieldType))
	} else {
		return fmt.Errorf("cannot convert %v to %v", inVal.Type(), fieldType)
	}

	return nil
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

// validateValue recursively validates a reflected value
// NOTE: 'required' constraint is skipped (not built in buildConstraints)
func (v *Validator[T]) validateValue(val reflect.Value, path string) []FieldError {
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
		constraintsMap := parseTag(field.Tag)
		if constraintsMap == nil {
			// No validation tags, but still check nested structs, slices, and maps
			if fieldValue.Kind() == reflect.Struct {
				errors = append(errors, v.validateValue(fieldValue, fieldPath)...)
			} else if fieldValue.Kind() == reflect.Slice {
				// Recursively validate struct elements in slices
				for i := 0; i < fieldValue.Len(); i++ {
					elemValue := fieldValue.Index(i)
					elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
					if elemValue.Kind() == reflect.Struct {
						errors = append(errors, v.validateValue(elemValue, elemPath)...)
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
						errors = append(errors, v.validateValue(mapValue, mapPath)...)
					}
				}
			}
			continue
		}

		// For nested structs (path != ""), check required fields
		// This is needed because nested structs in slices are deserialized via json.Unmarshal
		// which bypasses our custom required field checking in Unmarshal()
		if path != "" && v.options.StrictMissingFields {
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
		validators := constraints.BuildConstraints(constraintsMap, field.Type)

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
					errors = append(errors, v.validateValue(elemValue, elemPath)...)
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
					errors = append(errors, v.validateValue(mapValue, mapPath)...)
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
				errors = append(errors, v.validateValue(fieldValue, fieldPath)...)
			}
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
			inValue = fieldMissingSentinel // Field missing from JSON
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

// setDefaultValue sets a default value on a field
func (v *Validator[T]) setDefaultValue(fieldValue reflect.Value, defaultValue string) {
	if !fieldValue.CanSet() {
		return
	}

	// Handle pointer types
	if fieldValue.Kind() == reflect.Ptr {
		// Create a new value of the element type
		elemType := fieldValue.Type().Elem()
		newPtr := reflect.New(elemType)

		// Recursively set the default on the dereferenced pointer
		v.setDefaultValue(newPtr.Elem(), defaultValue)

		// Set the field to the new pointer
		fieldValue.Set(newPtr)
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
func (v *Validator[T]) Marshal(obj *T) ([]byte, error) {
	// Validate before marshaling
	if err := v.Validate(obj); err != nil {
		return nil, err
	}

	// Marshal to JSON
	return json.Marshal(obj)
}
