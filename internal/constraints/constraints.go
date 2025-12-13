// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// Constraint represents a validation constraint.
type Constraint interface {
	Validate(value any) error
}

// Shared regex patterns used by string constraints.
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// extractNumericValue converts a reflect.Value to a float64 for numeric comparisons.
// Returns (float64, error) where error is non-nil if the value is not numeric.
func extractNumericValue(v reflect.Value) (float64, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type: %s", v.Kind())
	}
}

// derefValue dereferences a pointer value, returning the underlying value or nil if invalid.
// Returns (reflect.Value, bool) where bool is false if the value is nil or invalid.
func derefValue(value any) (reflect.Value, bool) {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return reflect.Value{}, false
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}

	return v, true
}

// extractString extracts a string value from reflect.Value, checking type and dereferencing.
// Returns (string, isValid, error) where isValid is false for nil/invalid values.
func extractString(value any) (str string, isValid bool, err error) {
	v, ok := derefValue(value)
	if !ok {
		return "", false, nil // nil/invalid values should skip validation
	}

	if v.Kind() != reflect.String {
		return "", true, fmt.Errorf("requires string value")
	}

	return v.String(), true, nil
}

// BuildConstraints creates constraint instances from parsed tag map.
func BuildConstraints(constraints map[string]string, fieldType reflect.Type) []Constraint {
	var result []Constraint

	for name, value := range constraints {
		switch name {
		case "required":
			// Skip: 'required' is only checked during Unmarshal (missing JSON keys)
			// It doesn't apply to Validate() on manually created structs
			continue
		case "min":
			if constraint, ok := buildMinConstraint(value, fieldType); ok {
				result = append(result, constraint)
			}
		case "max":
			if constraint, ok := buildMaxConstraint(value, fieldType); ok {
				result = append(result, constraint)
			}
		case "gt":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, gtConstraint{threshold: threshold})
			}
		case "gte":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, geConstraint{threshold: threshold})
			}
		case "lt":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, ltConstraint{threshold: threshold})
			}
		case "lte":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, leConstraint{threshold: threshold})
			}
		case "email":
			result = append(result, emailConstraint{})
		case "url":
			result = append(result, urlConstraint{})
		case "uuid":
			result = append(result, uuidConstraint{})
		case "regexp":
			result = append(result, buildRegexConstraint(value))
		case "ipv4":
			result = append(result, ipv4Constraint{})
		case "ipv6":
			result = append(result, ipv6Constraint{})
		case "oneof":
			result = append(result, buildEnumConstraint(value))
		case "len":
			if constraint, ok := buildLenConstraint(value); ok {
				result = append(result, constraint)
			}
		case "ascii":
			result = append(result, asciiConstraint{})
		case "alpha":
			result = append(result, alphaConstraint{})
		case "alphanum":
			result = append(result, alphanumConstraint{})
		case "contains":
			if constraint, ok := buildContainsConstraint(value); ok {
				result = append(result, constraint)
			}
		case "excludes":
			if constraint, ok := buildExcludesConstraint(value); ok {
				result = append(result, constraint)
			}
		case "startswith":
			if constraint, ok := buildStartswithConstraint(value); ok {
				result = append(result, constraint)
			}
		case "endswith":
			if constraint, ok := buildEndswithConstraint(value); ok {
				result = append(result, constraint)
			}
		case "lowercase":
			result = append(result, lowercaseConstraint{})
		case "uppercase":
			result = append(result, uppercaseConstraint{})
		case "positive":
			result = append(result, positiveConstraint{})
		case "negative":
			result = append(result, negativeConstraint{})
		case "multiple_of":
			if constraint, ok := buildMultipleOfConstraint(value); ok {
				result = append(result, constraint)
			}
		case "max_digits":
			if constraint, ok := buildMaxDigitsConstraint(value); ok {
				result = append(result, constraint)
			}
		case "decimal_places":
			if constraint, ok := buildDecimalPlacesConstraint(value); ok {
				result = append(result, constraint)
			}
		case "disallow_inf_nan":
			result = append(result, disallowInfNanConstraint{})
		case "default":
			result = append(result, defaultConstraint{value: value})
		}
	}

	return result
}
