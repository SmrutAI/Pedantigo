// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Enum constraint types.
type (
	enumConstraint    struct{ values []string }
	enumCIConstraint  struct{ values []string } // Case-insensitive enum
	eqConstraint      struct{ value string }
	neConstraint      struct{ value string }
	defaultConstraint struct{ value string }
)

// enumConstraint validates that value is one of the allowed values.
func (c enumConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	// Convert value to string for comparison
	var str string
	switch v.Kind() {
	case reflect.String:
		str = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		str = strconv.FormatBool(v.Bool())
	default:
		return fmt.Errorf("enum constraint not supported for type %s", v.Kind())
	}

	// Check if value is in allowed list
	for _, allowed := range c.values {
		if str == allowed {
			return nil
		}
	}

	return fmt.Errorf("must be one of: %s", strings.Join(c.values, ", "))
}

// valueToString converts a reflect.Value to string for comparison.
// Returns (string, error) where error is non-nil if type is unsupported.
func valueToString(v reflect.Value, constraintName string) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	default:
		return "", fmt.Errorf("%s constraint not supported for type %s", constraintName, v.Kind())
	}
}

// eqConstraint validates that value equals a specific value.
func (c eqConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for nil/invalid values
	}

	str, err := valueToString(v, "eq")
	if err != nil {
		return err
	}

	if str != c.value {
		return fmt.Errorf("must be equal to %s", c.value)
	}
	return nil
}

// defaultConstraint is not a validator - it's handled during unmarshaling.
func (c defaultConstraint) Validate(value any) error {
	return nil // No-op for validation
}

// buildEnumConstraint parses space-separated enum values.
func buildEnumConstraint(value string) Constraint {
	values := strings.Fields(value)
	return enumConstraint{values: values}
}

// enumCIConstraint validates that value is one of the allowed values (case-insensitive).
func (c enumCIConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	// Convert value to string for comparison
	var str string
	switch v.Kind() {
	case reflect.String:
		str = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		str = strconv.FormatBool(v.Bool())
	default:
		return fmt.Errorf("oneofci constraint not supported for type %s", v.Kind())
	}

	// Check if value is in allowed list (case-insensitive)
	for _, allowed := range c.values {
		if strings.EqualFold(str, allowed) {
			return nil
		}
	}

	return fmt.Errorf("must be one of (case-insensitive): %s", strings.Join(c.values, ", "))
}

// buildEnumCIConstraint parses space-separated enum values for case-insensitive matching.
func buildEnumCIConstraint(value string) Constraint {
	values := strings.Fields(value)
	return enumCIConstraint{values: values}
}

// buildEqConstraint creates an eq constraint for a specific value.
func buildEqConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false
	}
	return eqConstraint{value: value}, true
}

// neConstraint validates that value does NOT equal a specific value.
func (c neConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for nil/invalid values
	}

	str, err := valueToString(v, "ne")
	if err != nil {
		return err
	}

	if str == c.value {
		return fmt.Errorf("must not be equal to %s", c.value)
	}
	return nil
}

// buildNeConstraint creates a ne constraint for a specific value.
func buildNeConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false
	}
	return neConstraint{value: value}, true
}

// eqIgnoreCaseConstraint validates case-insensitive equality.
type eqIgnoreCaseConstraint struct {
	value string
}

// Validate checks if value equals the constraint value case-insensitively.
func (c eqIgnoreCaseConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for nil/invalid values
	}

	str, err := valueToString(v, "eq_ignore_case")
	if err != nil {
		return err
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !strings.EqualFold(str, c.value) {
		return NewConstraintError(CodeEqIgnoreCase, fmt.Sprintf("must equal %q (case-insensitive)", c.value))
	}

	return nil
}

// buildEqIgnoreCaseConstraint creates an eq_ignore_case constraint.
func buildEqIgnoreCaseConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false
	}
	return eqIgnoreCaseConstraint{value: value}, true
}

// neIgnoreCaseConstraint validates case-insensitive not-equal.
type neIgnoreCaseConstraint struct {
	value string
}

// Validate checks if value does NOT equal the constraint value case-insensitively.
func (c neIgnoreCaseConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for nil/invalid values
	}

	str, err := valueToString(v, "ne_ignore_case")
	if err != nil {
		return err
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if strings.EqualFold(str, c.value) {
		return NewConstraintError(CodeNeIgnoreCase, fmt.Sprintf("must not equal %q (case-insensitive)", c.value))
	}

	return nil
}

// buildNeIgnoreCaseConstraint creates a ne_ignore_case constraint.
func buildNeIgnoreCaseConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false
	}
	return neIgnoreCaseConstraint{value: value}, true
}
