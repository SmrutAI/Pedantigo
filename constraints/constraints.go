// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Constraint represents a validation constraint
type Constraint interface {
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
	urlConstraint       struct{}
	uuidConstraint      struct{}
	regexConstraint     struct {
		pattern string
		regex   *regexp.Regexp
	}
	ipv4Constraint    struct{}
	ipv6Constraint    struct{}
	enumConstraint    struct{ values []string }
	defaultConstraint struct{ value string }
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// minConstraint validates that a numeric value is >= min
func (c minConstraint) Validate(value any) error {
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

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
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

// urlConstraint validates that a string is a valid URL (http or https only)
func (c urlConstraint) Validate(value any) error {
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
		return fmt.Errorf("url constraint requires string value")
	}

	str := v.String()
	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse the URL
	parsedURL, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("must be a valid URL (http or https)")
	}

	// Check scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("must be a valid URL (http or https)")
	}

	// Check host is non-empty
	if parsedURL.Host == "" {
		return fmt.Errorf("must be a valid URL (http or https)")
	}

	return nil
}

// uuidConstraint validates that a string is a valid UUID
func (c uuidConstraint) Validate(value any) error {
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
		return fmt.Errorf("uuid constraint requires string value")
	}

	str := v.String()
	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate UUID format using regex
	if !uuidRegex.MatchString(str) {
		return fmt.Errorf("must be a valid UUID")
	}

	return nil
}

// regexConstraint validates that a string matches a custom regex pattern
func (c regexConstraint) Validate(value any) error {
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
		return fmt.Errorf("regex constraint requires string value")
	}

	str := v.String()
	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate against the compiled regex
	if !c.regex.MatchString(str) {
		return fmt.Errorf("must match pattern '%s'", c.pattern)
	}

	return nil
}

// ipv4Constraint validates that a string is a valid IPv4 address
func (c ipv4Constraint) Validate(value any) error {
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
		return fmt.Errorf("ipv4 constraint requires string value")
	}

	str := v.String()
	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("must be a valid IPv4 address")
	}

	// Check if it's IPv4 (not IPv6)
	// IPv4 addresses return non-nil from To4()
	if ip.To4() == nil {
		return fmt.Errorf("must be a valid IPv4 address")
	}

	return nil
}

// ipv6Constraint validates that a string is a valid IPv6 address
func (c ipv6Constraint) Validate(value any) error {
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
		return fmt.Errorf("ipv6 constraint requires string value")
	}

	str := v.String()
	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("must be a valid IPv6 address")
	}

	// Check if it's IPv6 (not IPv4)
	// IPv6 addresses return nil from To4()
	if ip.To4() != nil {
		return fmt.Errorf("must be a valid IPv6 address")
	}

	return nil
}

// enumConstraint validates that value is one of the allowed values
func (c enumConstraint) Validate(value any) error {
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

// defaultConstraint is not a validator - it's handled during unmarshaling
func (c defaultConstraint) Validate(value any) error {
	return nil // No-op for validation
}

// BuildConstraints creates constraint instances from parsed tag map
func BuildConstraints(constraints map[string]string, fieldType reflect.Type) []Constraint {
	var result []Constraint

	for name, value := range constraints {
		switch name {
		case "required":
			// Skip: 'required' is only checked during Unmarshal (missing JSON keys)
			// It doesn't apply to Validate() on manually created structs
			continue
		case "min":
			// Context-aware: numeric min for numbers, length min for strings/slices
			if min, err := strconv.Atoi(value); err == nil {
				// Handle pointer types - check underlying type
				checkType := fieldType
				if checkType.Kind() == reflect.Ptr {
					checkType = checkType.Elem()
				}
				kind := checkType.Kind()
				if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
					result = append(result, minLengthConstraint{minLength: min})
				} else {
					result = append(result, minConstraint{min: min})
				}
			}
		case "max":
			// Context-aware: numeric max for numbers, length max for strings/slices
			if max, err := strconv.Atoi(value); err == nil {
				// Handle pointer types - check underlying type
				checkType := fieldType
				if checkType.Kind() == reflect.Ptr {
					checkType = checkType.Elem()
				}
				kind := checkType.Kind()
				if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
					result = append(result, maxLengthConstraint{maxLength: max})
				} else {
					result = append(result, maxConstraint{max: max})
				}
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
			// Compile regex pattern (fail-fast on invalid regex)
			compiledRegex, err := regexp.Compile(value)
			if err != nil {
				panic(fmt.Sprintf("invalid regex pattern '%s': %v", value, err))
			}
			result = append(result, regexConstraint{pattern: value, regex: compiledRegex})
		case "ipv4":
			result = append(result, ipv4Constraint{})
		case "ipv6":
			result = append(result, ipv6Constraint{})
		case "oneof":
			// Split oneof values by space (validator compatible)
			values := strings.Fields(value)
			result = append(result, enumConstraint{values: values})
		case "default":
			result = append(result, defaultConstraint{value: value})
		}
	}

	return result
}
