package constraints

import (
	"fmt"
	"reflect"
	"time"
)

// CompareValues compares two values using the specified operator
// op values: "eq", "ne", "gt", "gte", "lt", "lte"
// Returns true if comparison succeeds, false otherwise
// CompareValues compares two values
func CompareValues(op string, left, right any) (bool, error) {
	// Handle nil values (including typed nil pointers)
	leftIsNil := isNilValue(left)
	rightIsNil := isNilValue(right)

	// If both are nil, handle equality/inequality
	if leftIsNil && rightIsNil {
		switch op {
		case "eq":
			return true, nil
		case "ne":
			return false, nil
		case "gt", "gte", "lt", "lte":
			// nil is not greater/less than nil
			return false, nil
		}
	}

	// If only one is nil, they can't be compared for ordering
	if leftIsNil || rightIsNil {
		return false, fmt.Errorf("cannot compare incompatible types: %T vs %T", left, right)
	}

	// Check if both values are time.Time
	if isTime(left) && isTime(right) {
		return compareTime(op, left.(time.Time), right.(time.Time))
	}

	// Try numeric comparison
	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	if isNumeric(leftVal.Kind()) && isNumeric(rightVal.Kind()) {
		return compareNumeric(op, toFloat64(leftVal), toFloat64(rightVal))
	}

	// Try string comparison
	if leftVal.Kind() == reflect.String && rightVal.Kind() == reflect.String {
		return compareString(op, left.(string), right.(string))
	}

	// Unsupported types
	return false, fmt.Errorf("cannot compare incompatible types: %T vs %T", left, right)
}

// isNumeric checks if a reflect.Kind is a numeric type
func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// toFloat64 converts a numeric reflect.Value to float64
func toFloat64(val reflect.Value) float64 {
	kind := val.Kind()

	// Handle signed integers
	switch kind {
	case reflect.Int:
		return float64(val.Int())
	case reflect.Int8:
		return float64(val.Int())
	case reflect.Int16:
		return float64(val.Int())
	case reflect.Int32:
		return float64(val.Int())
	case reflect.Int64:
		return float64(val.Int())
	}

	// Handle unsigned integers
	switch kind {
	case reflect.Uint:
		return float64(val.Uint())
	case reflect.Uint8:
		return float64(val.Uint())
	case reflect.Uint16:
		return float64(val.Uint())
	case reflect.Uint32:
		return float64(val.Uint())
	case reflect.Uint64:
		return float64(val.Uint())
	}

	// Handle floats
	switch kind {
	case reflect.Float32:
		return val.Float()
	case reflect.Float64:
		return val.Float()
	}

	return 0
}

// compareNumeric compares two float64 values
func compareNumeric(op string, left, right float64) (bool, error) {
	switch op {
	case "eq":
		return left == right, nil
	case "ne":
		return left != right, nil
	case "gt":
		return left > right, nil
	case "gte":
		return left >= right, nil
	case "lt":
		return left < right, nil
	case "lte":
		return left <= right, nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %q", op)
	}
}

// compareString compares two strings lexicographically
func compareString(op string, left, right string) (bool, error) {
	switch op {
	case "eq":
		return left == right, nil
	case "ne":
		return left != right, nil
	case "gt":
		return left > right, nil
	case "gte":
		return left >= right, nil
	case "lt":
		return left < right, nil
	case "lte":
		return left <= right, nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %q", op)
	}
}

// isTime checks if a value is time.Time
func isTime(val any) bool {
	_, ok := val.(time.Time)
	return ok
}

// compareTime compares two time.Time values
func compareTime(op string, left, right time.Time) (bool, error) {
	switch op {
	case "eq":
		return left.Equal(right), nil
	case "ne":
		return !left.Equal(right), nil
	case "gt":
		return left.After(right), nil
	case "gte":
		return left.Equal(right) || left.After(right), nil
	case "lt":
		return left.Before(right), nil
	case "lte":
		return left.Equal(right) || left.Before(right), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %q", op)
	}
}

// isNilValue checks if a value is nil, including typed nil pointers/interfaces
func isNilValue(val any) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	}
	return false
}

// IsZeroValue checks if a value is the zero value for its type.
// Returns true for nil, zero integers, empty strings, false booleans, empty slices/maps, etc.
// Returns false for non-zero values.
// IsZeroValue checks the condition
func IsZeroValue(value any) bool {
	v := reflect.ValueOf(value)
	return !v.IsValid() || v.IsZero()
}
