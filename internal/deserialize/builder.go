package deserialize

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/SmrutAI/pedantigo/internal/tags"
)

// StringTransformations holds flags for string transformations to apply during deserialization.
type StringTransformations struct {
	StripWhitespace bool
	ToLower         bool
	ToUpper         bool
}

// MissingFieldSentinel is a sentinel value to distinguish missing fields from explicit null.
type MissingFieldSentinel struct{}

// FieldMissingSentinel is the singleton sentinel value.
var FieldMissingSentinel = MissingFieldSentinel{}

// FieldDeserializer is a closure that deserializes a single field
// inValue is FieldMissingSentinel if field is missing from JSON,
// nil if field is explicitly null, or the actual value if present
// FieldDeserializer represents the data structure.
type FieldDeserializer func(outPtr *reflect.Value, inValue any) error

// BuilderOptions configures the deserializer builder.
type BuilderOptions struct {
	StrictMissingFields bool
	TagName             string // Custom struct tag name (default: "pedantigo")
}

// BuildFieldDeserializers creates field deserializer closures for each struct field.
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

		// Skip fields with json:"-" (explicitly ignored)
		if jsonTag == "-" {
			continue
		}

		// Skip extra_fields field (handled separately for ExtraAllow mode)
		// Parse validation constraints using the configured tag name
		tagName := opts.TagName
		if tagName == "" {
			tagName = tags.DefaultTagName
		}
		pedantigoTag := field.Tag.Get(tagName)
		if pedantigoTag == tags.ExtraFieldsTag {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			// Extract field name from json tag (before comma)
			if name, _, found := strings.Cut(jsonTag, ","); found {
				fieldName = name
			} else {
				fieldName = jsonTag
			}
		}

		// Parse validation constraints
		constraints := tags.ParseTagWithName(field.Tag, tagName)

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

		// Parse string transformations
		var transformations StringTransformations
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

			// Parse transformation tags
			_, transformations.StripWhitespace = constraints["strip_whitespace"]
			_, transformations.ToLower = constraints["to_lower"]
			_, transformations.ToUpper = constraints["to_upper"]
		}

		// Check if this is a string field (for transformations)
		isStringField := field.Type.Kind() == reflect.String ||
			(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.String)

		// Create field deserializer closure
		fieldIndex := i
		fieldType := field.Type
		_, hasRequired := constraints["required"] // Check if key exists, not if value is non-empty
		fieldTransformations := transformations   // Capture for closure

		deserializers[fieldName] = func(outPtr *reflect.Value, inValue any) error {
			fieldValue := outPtr.Field(fieldIndex)

			// Determine if field was present in JSON
			_, fieldMissing := inValue.(MissingFieldSentinel)

			if fieldMissing {
				// Field is missing from JSON
				if staticDefault != nil {
					// Apply static default
					setDefaultValueFunc(fieldValue, *staticDefault)
					// Apply transformations to default value
					if isStringField {
						applyStringTransformations(fieldValue, fieldTransformations)
					}
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
						// Apply transformations to default value
						if isStringField {
							applyStringTransformations(fieldValue, fieldTransformations)
						}
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
			if err := setFieldValueFunc(fieldValue, inValue, fieldType); err != nil {
				return err
			}

			// Apply string transformations after setting the value
			if isStringField {
				applyStringTransformations(fieldValue, fieldTransformations)
			}

			return nil
		}
	}

	return deserializers
}

// applyStringTransformations applies string transformations to a field value.
// Order of operations: strip_whitespace first, then to_lower/to_upper.
func applyStringTransformations(fieldValue reflect.Value, transforms StringTransformations) {
	// Handle pointer to string
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return
		}
		fieldValue = fieldValue.Elem()
	}

	if fieldValue.Kind() != reflect.String || !fieldValue.CanSet() {
		return
	}

	str := fieldValue.String()

	// Apply strip_whitespace first
	if transforms.StripWhitespace {
		str = strings.TrimSpace(str)
	}

	// Apply case transformations (to_lower takes precedence if both specified)
	if transforms.ToLower {
		str = strings.ToLower(str)
	} else if transforms.ToUpper {
		str = strings.ToUpper(str)
	}

	fieldValue.SetString(str)
}

// ValidateDefaultMethod checks that a method exists and has the correct signature.
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
