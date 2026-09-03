package deserialize

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	walkerDecoderType   = reflect.TypeOf((*WalkerDecoder)(nil)).Elem()
	jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
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
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type, opts FieldOptions) error,
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
	recursiveSetFunc func(fieldValue reflect.Value, inValue any, fieldType reflect.Type, opts FieldOptions) error,
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

		// Recursively set the value on the dereferenced pointer. recursiveSetFunc is
		// called directly (never bypassed) with opts passed through unchanged
		// (pointer indirection does not add a path segment) — mirrors
		// validateWithCache's explicit path threading in validator.go.
		if err := recursiveSetFunc(newPtr.Elem(), inValue, elemType, opts); err != nil {
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

// RequiredFieldError represents a missing required field error with full path.
type RequiredFieldError struct {
	Field string
	// IsRoot is true when this required field is a direct field of the struct
	// passed straight to Unmarshal/UnmarshalInto (not reached via nested-struct
	// or dive recursion). Root-level required errors use the JSON field name and
	// carry no Code (matching the legacy top-level deserializer behavior); nested
	// required errors use the dotted Go-field path and Code: CodeRequired.
	IsRoot bool
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
