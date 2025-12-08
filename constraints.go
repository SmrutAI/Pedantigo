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
	minConstraint       struct{ min int }
	maxConstraint       struct{ max int }
	minLengthConstraint struct{ minLength int }
	maxLengthConstraint struct{ maxLength int }
	gtConstraint        struct{ threshold float64 }
	geConstraint        struct{ threshold float64 }
	ltConstraint        struct{ threshold float64 }
	leConstraint        struct{ threshold float64 }
	emailConstraint     struct{}
	defaultConstraint   struct{ value string }
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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

// minLengthConstraint validates that a string has at least minLength characters
func (c minLengthConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// Ensure we have a string
	if v.Kind() != reflect.String {
		return fmt.Errorf("min_length constraint requires string value")
	}

	str := v.String()
	if len(str) < c.minLength {
		return fmt.Errorf("must be at least %d characters", c.minLength)
	}

	return nil
}

// maxLengthConstraint validates that a string has at most maxLength characters
func (c maxLengthConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// Ensure we have a string
	if v.Kind() != reflect.String {
		return fmt.Errorf("max_length constraint requires string value")
	}

	str := v.String()
	if len(str) > c.maxLength {
		return fmt.Errorf("must be at most %d characters", c.maxLength)
	}

	return nil
}

// gtConstraint validates that a numeric value is > threshold
func (c gtConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	var numValue float64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = v.Float()
	default:
		return fmt.Errorf("gt constraint requires numeric value")
	}

	if numValue <= c.threshold {
		return fmt.Errorf("must be greater than %v", c.threshold)
	}

	return nil
}

// geConstraint validates that a numeric value is >= threshold
func (c geConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	var numValue float64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = v.Float()
	default:
		return fmt.Errorf("ge constraint requires numeric value")
	}

	if numValue < c.threshold {
		return fmt.Errorf("must be at least %v", c.threshold)
	}

	return nil
}

// ltConstraint validates that a numeric value is < threshold
func (c ltConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	var numValue float64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = v.Float()
	default:
		return fmt.Errorf("lt constraint requires numeric value")
	}

	if numValue >= c.threshold {
		return fmt.Errorf("must be less than %v", c.threshold)
	}

	return nil
}

// leConstraint validates that a numeric value is <= threshold
func (c leConstraint) Validate(value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	var numValue float64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = v.Float()
	default:
		return fmt.Errorf("le constraint requires numeric value")
	}

	if numValue > c.threshold {
		return fmt.Errorf("must be at most %v", c.threshold)
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
			// Skip: 'required' is only checked during Unmarshal (missing JSON keys)
			// It doesn't apply to Validate() on manually created structs
			continue
		case "min":
			if min, err := strconv.Atoi(value); err == nil {
				result = append(result, minConstraint{min: min})
			}
		case "max":
			if max, err := strconv.Atoi(value); err == nil {
				result = append(result, maxConstraint{max: max})
			}
		case "min_length":
			if minLength, err := strconv.Atoi(value); err == nil {
				result = append(result, minLengthConstraint{minLength: minLength})
			}
		case "max_length":
			if maxLength, err := strconv.Atoi(value); err == nil {
				result = append(result, maxLengthConstraint{maxLength: maxLength})
			}
		case "gt":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, gtConstraint{threshold: threshold})
			}
		case "ge":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, geConstraint{threshold: threshold})
			}
		case "lt":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, ltConstraint{threshold: threshold})
			}
		case "le":
			if threshold, err := strconv.ParseFloat(value, 64); err == nil {
				result = append(result, leConstraint{threshold: threshold})
			}
		case "email":
			result = append(result, emailConstraint{})
		case "default":
			result = append(result, defaultConstraint{value: value})
		}
	}

	return result
}
