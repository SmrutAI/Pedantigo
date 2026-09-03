package deserialize

import (
	"fmt"
	"reflect"
	"strings"
)

// StringTransformations holds flags for string transformations to apply during deserialization.
type StringTransformations struct {
	StripWhitespace bool
	ToLower         bool
	ToUpper         bool
}

// applyStringTransformations applies string transformations to a field value.
// Order of operations: strip_whitespace first, then to_lower/to_upper.
func applyStringTransformations(fieldValue reflect.Value, transforms StringTransformations) {
	// Handle pointer to string
	if fieldValue.Kind() == reflect.Pointer {
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
