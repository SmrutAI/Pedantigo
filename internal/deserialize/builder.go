package deserialize

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/SmrutAI/Pedantigo/internal/tags"
)

// MissingFieldSentinel is a sentinel value to distinguish missing fields from explicit null
type MissingFieldSentinel struct{}

// FieldMissingSentinel is the singleton sentinel value
var FieldMissingSentinel = MissingFieldSentinel{}

// FieldDeserializer is a closure that deserializes a single field
// inValue is FieldMissingSentinel if field is missing from JSON,
// nil if field is explicitly null, or the actual value if present
type FieldDeserializer func(outPtr *reflect.Value, inValue any) error

// BuilderOptions configures the deserializer builder
type BuilderOptions struct {
	StrictMissingFields bool
}

// BuildFieldDeserializers creates field deserializer closures for each struct field
func BuildFieldDeserializers(
	typ reflect.Type,
	opts BuilderOptions,
	setFieldValueFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
	setDefaultValueFunc func(fieldValue reflect.Value, defaultValue string),
) map[string]FieldDeserializer {
	deserializers := make(map[string]FieldDeserializer)

	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return deserializers
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
		constraints := tags.ParseTag(field.Tag)

		// Safety check: panic if default tags are used when StrictMissingFields is disabled
		if constraints != nil && !opts.StrictMissingFields {
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
				if err := ValidateDefaultMethod(typ, method, field.Type); err != nil {
					panic(fmt.Sprintf("field %s: %v", field.Name, err))
				}
			}
		}

		// Create field deserializer closure
		fieldIndex := i
		fieldType := field.Type
		_, hasRequired := constraints["required"] // Check if key exists, not if value is non-empty

		deserializers[fieldName] = func(outPtr *reflect.Value, inValue any) error {
			fieldValue := outPtr.Field(fieldIndex)

			// Determine if field was present in JSON
			_, fieldMissing := inValue.(MissingFieldSentinel)

			if fieldMissing {
				// Field is missing from JSON
				if staticDefault != nil {
					// Apply static default
					setDefaultValueFunc(fieldValue, *staticDefault)
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

				if hasRequired && opts.StrictMissingFields {
					return fmt.Errorf("is required")
				}

				// Leave as zero value (relaxed mode or not required)
				return nil
			}

			// Field is present in JSON - set the value
			return setFieldValueFunc(fieldValue, inValue, fieldType)
		}
	}

	return deserializers
}

// ValidateDefaultMethod checks that a method exists and has the correct signature
func ValidateDefaultMethod(structType reflect.Type, methodName string, fieldType reflect.Type) error {
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
