package deserialize

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SetFieldValue sets a field value from a JSON value
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
					if err := recursiveSetFunc(fieldVal, val, field.Type); err != nil {
						return err
					}
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
					if err := recursiveSetFunc(structFieldVal, fieldVal, field.Type); err != nil {
						return err
					}
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

// SetDefaultValue sets a default value on a field
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
