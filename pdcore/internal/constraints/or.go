// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"reflect"
	"strings"
)

// orConstraint wraps multiple constraints and passes if ANY matches.
type orConstraint struct {
	constraints []Constraint
	expression  string // Original expression for error messages (e.g., "hexcolor|rgb|rgba")
}

// Validate checks if the value matches at least one of the constraints.
func (c orConstraint) Validate(value any) error {
	// Skip validation for nil/empty values (let required handle that)
	if value == nil {
		return nil
	}

	// Check if string is empty
	if str, ok := value.(string); ok && str == "" {
		return nil
	}

	// Check if pointer is nil
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}

	// Try each constraint - pass if ANY matches
	for _, constraint := range c.constraints {
		if err := constraint.Validate(value); err == nil {
			return nil // At least one passed!
		}
	}

	// All failed
	return NewConstraintErrorf(CodeOrConstraintFailed,
		"must match one of: %s", c.expression)
}

// buildOrConstraint creates an OR constraint from a pipe-separated expression.
// Example: "hexcolor|rgb|rgba" -> constraint that passes if any matches.
func buildOrConstraint(expression string, fieldType reflect.Type) Constraint {
	parts := strings.Split(expression, "|")
	if len(parts) < 2 {
		return nil // Not a valid OR expression
	}

	var constraints []Constraint
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse individual constraint - handle both bare (email) and key=value (min=5)
		partMap := make(map[string]string)
		if idx := strings.IndexByte(part, '='); idx != -1 {
			partMap[part[:idx]] = part[idx+1:]
		} else {
			partMap[part] = ""
		}

		// Build constraints for this part
		built := BuildConstraints(partMap, fieldType)
		constraints = append(constraints, built...)
	}

	if len(constraints) == 0 {
		return nil
	}

	return orConstraint{
		constraints: constraints,
		expression:  expression,
	}
}
