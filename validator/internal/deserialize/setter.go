package deserialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/SmrutAI/pedantigo/v2/validator/internal/tags"
)

// FieldOptions contains options for field deserialization and required checking.
type FieldOptions struct {
	// StrictMissingFields controls whether missing required fields cause errors.
	StrictMissingFields bool
	// TagName is the struct tag name to parse (default "validate"; e.g., "binding" for Gin).
	TagName string
	// Path is the current field path for error reporting (e.g., "Items[0]").
	Path string
	// FieldName is the Go struct field name for the current field (used to build paths).
	FieldName string
}

// SetFieldValue sets a field value from a JSON value.
// This is the backward-compatible version without options (used by validator.go).
func SetFieldValue(
	fieldValue reflect.Value,
	inValue any,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
) error {
	// Delegate to the options version with empty options (no required checking)
	return SetFieldValueWithOptions(fieldValue, inValue, fieldType, recursiveSetFunc, FieldOptions{})
}

// SetFieldValueWithOptions sets a field value from a JSON value with options.
// FieldOptions enable required field checking during deserialization (for nested structs via dive).
func SetFieldValueWithOptions(
	fieldValue reflect.Value,
	inValue any,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
	opts FieldOptions,
) error {
	if !fieldValue.CanSet() {
		return nil
	}

	// Handle pointer types
	if fieldType.Kind() == reflect.Pointer {
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
		// Convert inValue to map[string]any for deserializeStructFields
		inputMap, ok := inValue.(map[string]any)
		if !ok {
			// Fallback: re-marshal the map and unmarshal into the struct
			jsonBytes, err := json.Marshal(inValue)
			if err != nil {
				return fmt.Errorf("failed to marshal nested struct: %w", err)
			}
			newStruct := reflect.New(fieldType)
			if err := json.Unmarshal(jsonBytes, newStruct.Interface()); err != nil {
				return fmt.Errorf("failed to unmarshal nested struct: %w", err)
			}
			fieldValue.Set(newStruct.Elem())
			return nil
		}

		// Build path for nested struct fields
		nestedPath := opts.Path
		if nestedPath == "" && opts.FieldName != "" {
			nestedPath = opts.FieldName
		}

		// Create nested options with updated path
		nestedOpts := FieldOptions{
			StrictMissingFields: opts.StrictMissingFields,
			TagName:             opts.TagName,
			Path:                nestedPath,
		}

		// Use deserializeStructFields to properly check required fields
		return deserializeStructFields(fieldValue, fieldType, inputMap, recursiveSetFunc, nestedOpts)
	}

	// Handle slices: if inValue is []any and target is slice
	if inVal.Kind() == reflect.Slice && fieldType.Kind() == reflect.Slice {
		return setSliceField(fieldValue, inVal, fieldType, recursiveSetFunc, opts)
	}

	// Handle maps: if inValue is map[string]any and target is map
	if inVal.Kind() == reflect.Map && fieldType.Kind() == reflect.Map {
		return setMapField(fieldValue, inVal, fieldType, recursiveSetFunc, opts)
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
// It handles JSON field name resolution, checks for field presence in the input map,
// and validates required fields during deserialization (to distinguish missing keys from zero values).
// Collects all errors instead of returning early to provide complete error feedback.
func deserializeStructFields(
	structValue reflect.Value,
	structType reflect.Type,
	inputMap map[string]any,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
	opts FieldOptions,
) error {
	var requiredErrors []*RequiredFieldError

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

		// Build full path for error reporting (e.g., "Items[0].City")
		// Use Go struct field name (not JSON name) to match existing error format
		fullPath := field.Name
		if opts.Path != "" {
			fullPath = opts.Path + "." + field.Name
		}

		// Check if field exists in JSON
		val, exists := inputMap[jsonFieldName]
		if !exists {
			// Parse tags to check for defaults and required
			tagName := opts.TagName
			if tagName == "" {
				tagName = tags.DefaultTagName
			}
			parsedTag := tags.ParseTagWithName(field.Tag, tagName)

			// Apply static default if present
			if defVal, hasDefault := parsedTag["default"]; hasDefault {
				fieldVal := structValue.Field(j)
				var setDefault func(reflect.Value, string)
				setDefault = func(fv reflect.Value, dv string) {
					SetDefaultValue(fv, dv, setDefault)
				}
				setDefault(fieldVal, defVal)
				continue
			}

			// Apply defaultUsingMethod if present
			if methodName, hasMethod := parsedTag["defaultUsingMethod"]; hasMethod {
				fieldVal := structValue.Field(j)
				if structValue.CanAddr() {
					ptrValue := structValue.Addr()
					method := ptrValue.MethodByName(methodName)
					if method.IsValid() {
						results := method.Call(nil)
						if len(results) == 2 {
							if !results[1].IsNil() {
								return results[1].Interface().(error)
							}
							fieldVal.Set(results[0])
						}
					}
				}
				continue
			}

			// No default - check if required
			if opts.StrictMissingFields {
				if _, hasRequired := parsedTag["required"]; hasRequired {
					requiredErrors = append(requiredErrors, &RequiredFieldError{Field: fullPath})
				}
			}
			// Leave as zero value
			continue
		}

		// Set the field value
		fieldVal := structValue.Field(j)
		if err := recursiveSetFunc(fieldVal, val, field.Type); err != nil {
			// Check if it's a multi-error from nested struct
			var multiErr *MultiRequiredFieldError
			if errors.As(err, &multiErr) {
				requiredErrors = append(requiredErrors, multiErr.Errors...)
			} else {
				// For other errors, return immediately (type conversion errors, etc.)
				return err
			}
		}
	}

	// Return collected required errors
	if len(requiredErrors) > 0 {
		return &MultiRequiredFieldError{Errors: requiredErrors}
	}

	return nil
}

// RequiredFieldError represents a missing required field error with full path.
type RequiredFieldError struct {
	Field string
}

func (e *RequiredFieldError) Error() string {
	return "is required"
}

// MultiRequiredFieldError collects multiple required field errors.
type MultiRequiredFieldError struct {
	Errors []*RequiredFieldError
}

func (e *MultiRequiredFieldError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d required fields missing", len(e.Errors))
}

// setSliceField handles deserialization of slice types.
// For slices containing structs, it uses deserializeStructFields to track field presence.
// Collects all errors from all elements instead of returning early.
func setSliceField(
	fieldValue reflect.Value,
	inVal reflect.Value,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
	opts FieldOptions,
) error {
	elemType := fieldType.Elem()
	newSlice := reflect.MakeSlice(fieldType, inVal.Len(), inVal.Len())
	var requiredErrors []*RequiredFieldError

	for i := 0; i < inVal.Len(); i++ {
		elemValue := newSlice.Index(i)
		elemInput := inVal.Index(i).Interface()

		// Build element path (e.g., "Items[0]")
		// Use FieldName if Path is empty (top-level collection)
		basePath := opts.Path
		if basePath == "" && opts.FieldName != "" {
			basePath = opts.FieldName
		}
		elemPath := fmt.Sprintf("%s[%d]", basePath, i)

		// For structs in slices, manually deserialize fields to track which are present
		if elemType.Kind() == reflect.Struct && reflect.TypeOf(elemInput).Kind() == reflect.Map {
			inputMap, ok := elemInput.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map for struct element")
			}

			// Create new struct instance
			newStruct := reflect.New(elemType).Elem()

			// Deserialize struct fields using helper (passes options for required check)
			elemOpts := opts
			elemOpts.Path = elemPath
			if err := deserializeStructFields(newStruct, elemType, inputMap, recursiveSetFunc, elemOpts); err != nil {
				// Collect required errors, return immediately on other errors
				var multiErr *MultiRequiredFieldError
				if errors.As(err, &multiErr) {
					requiredErrors = append(requiredErrors, multiErr.Errors...)
				} else {
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

	// Return collected errors
	if len(requiredErrors) > 0 {
		return &MultiRequiredFieldError{Errors: requiredErrors}
	}
	return nil
}

// setMapField handles deserialization of map types.
// For maps with struct values, it uses deserializeStructFields to track field presence.
// Collects all errors from all entries instead of returning early.
func setMapField(
	fieldValue reflect.Value,
	inVal reflect.Value,
	fieldType reflect.Type,
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error,
	opts FieldOptions,
) error {
	keyType := fieldType.Key()
	valueType := fieldType.Elem()

	// Create new map
	newMap := reflect.MakeMap(fieldType)
	var requiredErrors []*RequiredFieldError

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

		// Build element path (e.g., "Offices[branch]")
		// Use FieldName if Path is empty (top-level collection)
		basePath := opts.Path
		if basePath == "" && opts.FieldName != "" {
			basePath = opts.FieldName
		}
		elemPath := fmt.Sprintf("%s[%v]", basePath, key.Interface())

		// For struct values in maps, manually deserialize fields to track which are present
		if valueType.Kind() == reflect.Struct && reflect.TypeOf(val).Kind() == reflect.Map {
			inputMap, ok := val.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map for struct value")
			}

			// Create new struct instance
			newStruct := reflect.New(valueType).Elem()

			// Deserialize struct fields using helper (passes options for required check)
			elemOpts := opts
			elemOpts.Path = elemPath
			if err := deserializeStructFields(newStruct, valueType, inputMap, recursiveSetFunc, elemOpts); err != nil {
				// Collect required errors, return immediately on other errors
				var multiErr *MultiRequiredFieldError
				if errors.As(err, &multiErr) {
					requiredErrors = append(requiredErrors, multiErr.Errors...)
				} else {
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

	// Return collected errors
	if len(requiredErrors) > 0 {
		return &MultiRequiredFieldError{Errors: requiredErrors}
	}
	return nil
}

// SetDefaultValue sets a default value on a field.
func SetDefaultValue(fieldValue reflect.Value, defaultValue string, recursiveSetFunc func(fieldValue reflect.Value, defaultValue string)) {
	if !fieldValue.CanSet() {
		return
	}

	// Handle pointer types
	if fieldValue.Kind() == reflect.Pointer {
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
	case reflect.Slice:
		// Parse space-separated string (e.g., "read write") into the slice type,
		// consistent with oneof constraint syntax.
		parts := strings.Fields(defaultValue)
		if len(parts) == 0 {
			return
		}
		elemType := fieldValue.Type().Elem()
		slice := reflect.MakeSlice(fieldValue.Type(), 0, len(parts))
		for _, part := range parts {
			elem := reflect.New(elemType).Elem()
			switch elemType.Kind() {
			case reflect.String:
				elem.SetString(part)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i, err := strconv.ParseInt(part, 10, 64); err == nil {
					elem.SetInt(i)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if u, err := strconv.ParseUint(part, 10, 64); err == nil {
					elem.SetUint(u)
				}
			case reflect.Float32, reflect.Float64:
				if f, err := strconv.ParseFloat(part, 64); err == nil {
					elem.SetFloat(f)
				}
			case reflect.Bool:
				if b, err := strconv.ParseBool(part); err == nil {
					elem.SetBool(b)
				}
			}
			slice = reflect.Append(slice, elem)
		}
		fieldValue.Set(slice)
	}
}
