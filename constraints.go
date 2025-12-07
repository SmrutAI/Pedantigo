package pedantigo

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// constraint represents a validation constraint
type constraint interface {
	Validate(value any) error
}

// Built-in constraint types
type (
	requiredConstraint struct{}
	minConstraint      struct{ min int }
	maxConstraint      struct{ max int }
	emailConstraint    struct{}
	defaultConstraint  struct{ value string }
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// requiredConstraint validates that a value is not the zero value
func (c requiredConstraint) Validate(value any) error {
	if value == nil {
		return fmt.Errorf("is required")
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() || v.IsZero() {
		return fmt.Errorf("is required")
	}

	return nil
}

// minConstraint validates that a numeric value is >= min
func (c minConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < int64(c.min) {
			return fmt.Errorf("must be at least %d", c.min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() < uint64(c.min) {
			return fmt.Errorf("must be at least %d", c.min)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < float64(c.min) {
			return fmt.Errorf("must be at least %d", c.min)
		}
	case reflect.String:
		if len(v.String()) < c.min {
			return fmt.Errorf("must be at least %d characters", c.min)
		}
	default:
		return fmt.Errorf("min constraint not supported for type %s", v.Kind())
	}

	return nil
}

// maxConstraint validates that a numeric value is <= max
func (c maxConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > int64(c.max) {
			return fmt.Errorf("must be at most %d", c.max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() > uint64(c.max) {
			return fmt.Errorf("must be at most %d", c.max)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > float64(c.max) {
			return fmt.Errorf("must be at most %d", c.max)
		}
	case reflect.String:
		if len(v.String()) > c.max {
			return fmt.Errorf("must be at most %d characters", c.max)
		}
	default:
		return fmt.Errorf("max constraint not supported for type %s", v.Kind())
	}

	return nil
}

// emailConstraint validates that a string is a valid email format
func (c emailConstraint) Validate(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("email constraint requires string value")
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !emailRegex.MatchString(str) {
		return fmt.Errorf("must be a valid email address")
	}

	return nil
}

// defaultConstraint is not a validator - it's handled during unmarshaling
func (c defaultConstraint) Validate(value any) error {
	return nil // No-op for validation
}

// buildConstraints creates constraint instances from parsed tag map
func buildConstraints(constraints map[string]string) []constraint {
	var result []constraint

	for name, value := range constraints {
		switch name {
		case "required":
			result = append(result, requiredConstraint{})
		case "min":
			if min, err := strconv.Atoi(value); err == nil {
				result = append(result, minConstraint{min: min})
			}
		case "max":
			if max, err := strconv.Atoi(value); err == nil {
				result = append(result, maxConstraint{max: max})
			}
		case "email":
			result = append(result, emailConstraint{})
		case "default":
			result = append(result, defaultConstraint{value: value})
		}
	}

	return result
}
