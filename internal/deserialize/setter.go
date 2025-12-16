package deserialize

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SetFieldValue sets a field value from a JSON value.
func SetFieldValue(
	fieldValue reflect.Value,
	inValue any,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
) error {
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
		if err := recursiveSetFunc(newPtr.Elem(), inValue, elemType); err != nil {
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

	// Handle nil/null for other types - set to zero value
	// This handles cases like JSON null for non-pointer string/int fields
	if inValue == nil {
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
				return fmt.Errorf("failed to parse time: %w", err)
			}
			fieldValue.Set(reflect.ValueOf(t))
			return nil
		}
	}

	// Handle time.Duration special case
	// Duration can come as:
	// - String: "1h30m", "500ms", "2h45m30s" (Go duration format)
	// - int64: nanoseconds (Go's internal representation)
	// - float64: seconds (common JSON convention)
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		switch inVal.Kind() {
		case reflect.String:
			// Parse Go duration string: "1h30m", "500ms", "2h45m30s"
			d, err := time.ParseDuration(inVal.String())
			if err != nil {
				return fmt.Errorf("failed to parse duration: %w", err)
			}
			fieldValue.Set(reflect.ValueOf(d))
			return nil
		case reflect.Int, reflect.Int64:
			// Interpret as nanoseconds (Go's internal representation)
			fieldValue.Set(reflect.ValueOf(time.Duration(inVal.Int())))
			return nil
		case reflect.Float64:
			// Interpret as seconds (common JSON convention)
			fieldValue.Set(reflect.ValueOf(time.Duration(inVal.Float() * float64(time.Second))))
			return nil
		default:
			return fmt.Errorf("cannot convert %v to time.Duration", inVal.Kind())
		}
	}

	// Handle nested structs: if inValue is map[string]any and target is struct
	if inVal.Kind() == reflect.Map && fieldType.Kind() == reflect.Struct {
		// Re-marshal the map and unmarshal into the struct
		jsonBytes, err := json.Marshal(inValue)
		if err != nil {
			return fmt.Errorf("failed to marshal nested struct: %w", err)
		}

		// Create a new instance of the target type
		newStruct := reflect.New(fieldType)
		if err := json.Unmarshal(jsonBytes, newStruct.Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal nested struct: %w", err)
		}

		fieldValue.Set(newStruct.Elem())
		return nil
	}

	// Handle slices: if inValue is []any and target is slice
	if inVal.Kind() == reflect.Slice && fieldType.Kind() == reflect.Slice {
		return setSliceField(fieldValue, inVal, fieldType, recursiveSetFunc)
	}

	// Handle maps: if inValue is map[string]any and target is map
	if inVal.Kind() == reflect.Map && fieldType.Kind() == reflect.Map {
		return setMapField(fieldValue, inVal, fieldType, recursiveSetFunc)
	}

	// Handle type conversion
	switch {
	case inVal.Type().AssignableTo(fieldType):
		fieldValue.Set(inVal)
	case inVal.Type().ConvertibleTo(fieldType):
		// Block nonsensical conversions (e.g., int→string which converts to rune)
		// Allow only meaningful conversions between numeric types or within same kind
		if isValidConversion(inVal.Type(), fieldType) {
			fieldValue.Set(inVal.Convert(fieldType))
		} else {
			return fmt.Errorf("cannot convert %v to %v", inVal.Type(), fieldType)
		}
	default:
		return fmt.Errorf("cannot convert %v to %v", inVal.Type(), fieldType)
	}

	return nil
}

// isValidConversion checks if a type conversion is semantically valid for JSON deserialization
// Blocks nonsensical conversions like int→string (which would convert to rune).
func isValidConversion(from, to reflect.Type) bool {
	fromKind := from.Kind()
	toKind := to.Kind()

	// Allow conversions between numeric types
	if isNumericKind(fromKind) && isNumericKind(toKind) {
		return true
	}

	// Block int/uint→string conversions (would convert to rune)
	if isNumericKind(fromKind) && toKind == reflect.String {
		return false
	}

	// Block string→int/uint conversions (ConvertibleTo returns true but panics at runtime)
	if fromKind == reflect.String && isNumericKind(toKind) {
		return false
	}

	// Allow same-kind conversions (e.g., custom string types)
	if fromKind == toKind {
		return true
	}

	return false
}

// isNumericKind checks if a kind is a numeric type.
func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// deserializeStructFields iterates through struct fields and sets their values from a map.
// It handles JSON field name resolution and checks for field presence in the input map.
func deserializeStructFields(
	structValue reflect.Value,
	structType reflect.Type,
	inputMap map[string]any,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
) error {
	// Iterate through struct fields and set values
	for j := 0; j < structType.NumField(); j++ {
		field := structType.Field(j)

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
		fieldVal := structValue.Field(j)
		if err := recursiveSetFunc(fieldVal, val, field.Type); err != nil {
			return err
		}
	}

	return nil
}

// setSliceField handles deserialization of slice types.
// For slices containing structs, it uses deserializeStructFields to track field presence.
func setSliceField(
	fieldValue reflect.Value,
	inVal reflect.Value,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
) error {
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

			// Deserialize struct fields using helper
			if err := deserializeStructFields(newStruct, elemType, inputMap, recursiveSetFunc); err != nil {
				return err
			}

			elemValue.Set(newStruct)
		} else {
			if err := recursiveSetFunc(elemValue, elemInput, elemType); err != nil {
				return err
			}
		}
	}

	fieldValue.Set(newSlice)
	return nil
}

// setMapField handles deserialization of map types.
// For maps with struct values, it uses deserializeStructFields to track field presence.
func setMapField(
	fieldValue reflect.Value,
	inVal reflect.Value,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
) error {
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
		switch {
		case key.Type().AssignableTo(keyType):
			convertedKey = key
		case key.Type().ConvertibleTo(keyType):
			convertedKey = key.Convert(keyType)
		default:
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

			// Deserialize struct fields using helper
			if err := deserializeStructFields(newStruct, valueType, inputMap, recursiveSetFunc); err != nil {
				return err
			}

			newMap.SetMapIndex(convertedKey, newStruct)
		} else {
			// For non-struct values, convert normally
			newValue := reflect.New(valueType).Elem()
			if err := recursiveSetFunc(newValue, val, valueType); err != nil {
				return err
			}
			newMap.SetMapIndex(convertedKey, newValue)
		}
	}

	fieldValue.Set(newMap)
	return nil
}

// SetDefaultValue sets a default value on a field.
func SetDefaultValue(fieldValue reflect.Value, defaultValue string, recursiveSetFunc func(fieldValue reflect.Value, defaultValue string)) {
	if !fieldValue.CanSet() {
		return
	}

	// Handle pointer types
	if fieldValue.Kind() == reflect.Ptr {
		// Create a new value of the element type
		elemType := fieldValue.Type().Elem()
		newPtr := reflect.New(elemType)

		// Recursively set the default on the dereferenced pointer
		recursiveSetFunc(newPtr.Elem(), defaultValue)

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
