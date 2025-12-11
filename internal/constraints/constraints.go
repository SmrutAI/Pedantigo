// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"math"
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
	ipv4Constraint       struct{}
	ipv6Constraint       struct{}
	enumConstraint       struct{ values []string }
	defaultConstraint    struct{ value string }
	lenConstraint        struct{ length int }
	asciiConstraint      struct{}
	alphaConstraint      struct{}
	alphanumConstraint   struct{}
	containsConstraint   struct{ substring string }
	excludesConstraint   struct{ substring string }
	startswithConstraint struct{ prefix string }
	endswithConstraint   struct{ suffix string }
	lowercaseConstraint  struct{}
	uppercaseConstraint  struct{}
	positiveConstraint   struct{}
	negativeConstraint   struct{}
	multipleOfConstraint struct{ factor float64 }
	maxDigitsConstraint  struct{ maxDigits int }
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// minConstraint validates that a numeric value is >= min
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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
// Validate checks if the value satisfies the constraint
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

// Validate for defaultConstraint is not a validator - it's handled during unmarshaling
func (c defaultConstraint) Validate(value any) error {
	return nil // No-op for validation
}

// Validate for lenConstraint validates that a string has exact length
func (c lenConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("len constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Note: len constraint validates empty strings (len=0 is valid)
	// Do NOT skip empty strings like other constraints

	// 6. Validation logic - count runes, not bytes (for Unicode support)
	runeCount := len([]rune(str))
	if runeCount != c.length {
		return fmt.Errorf("must be exactly %d characters", c.length)
	}

	return nil
}

// asciiConstraint validates that a string contains only ASCII characters
// Validate checks if the value satisfies the constraint
func (c asciiConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("ascii constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check all runes are ASCII (0-127)
	for _, r := range str {
		if r > 127 {
			return fmt.Errorf("must contain only ASCII characters")
		}
	}

	return nil
}

// alphaConstraint validates that a string contains only alphabetic characters
// Validate checks if the value satisfies the constraint
func (c alphaConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("alpha constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string matches alphabetic pattern
	if !alphaRegex.MatchString(str) {
		return fmt.Errorf("must contain only alphabetic characters")
	}

	return nil
}

// alphanumConstraint validates that a string contains only alphanumeric characters
// Validate checks if the value satisfies the constraint
func (c alphanumConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("alphanum constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string matches alphanumeric pattern
	if !alphanumRegex.MatchString(str) {
		return fmt.Errorf("must contain only alphanumeric characters")
	}

	return nil
}

// containsConstraint validates that a string contains a specific substring
// Validate checks if the value satisfies the constraint
func (c containsConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("contains constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings only if substring is non-empty
	if str == "" && c.substring != "" {
		return fmt.Errorf("must contain '%s'", c.substring)
	}

	// 6. Validation logic - check if string contains substring
	if !strings.Contains(str, c.substring) {
		return fmt.Errorf("must contain '%s'", c.substring)
	}

	return nil
}

// Validate for excludesConstraint validates that a string does NOT contain a specific substring
func (c excludesConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("excludes constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string does NOT contain substring
	if strings.Contains(str, c.substring) {
		return fmt.Errorf("must not contain '%s'", c.substring)
	}

	return nil
}

// Validate for startswithConstraint validates that a string starts with a specific prefix
func (c startswithConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("startswith constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string starts with prefix
	if !strings.HasPrefix(str, c.prefix) {
		return fmt.Errorf("must start with '%s'", c.prefix)
	}

	return nil
}

// Validate for endswithConstraint validates that a string ends with a specific suffix
func (c endswithConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("endswith constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string ends with suffix
	if !strings.HasSuffix(str, c.suffix) {
		return fmt.Errorf("must end with '%s'", c.suffix)
	}

	return nil
}

// Validate for lowercaseConstraint validates that a string is all lowercase
func (c lowercaseConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("lowercase constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string is all lowercase
	if str != strings.ToLower(str) {
		return fmt.Errorf("must be all lowercase")
	}

	return nil
}

// Validate for uppercaseConstraint validates that a string is all uppercase
func (c uppercaseConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Type check - ensure string
	if v.Kind() != reflect.String {
		return fmt.Errorf("uppercase constraint requires string value")
	}

	// 4. Get string value
	str := v.String()

	// 5. Skip empty strings
	if str == "" {
		return nil
	}

	// 6. Validation logic - check if string is all uppercase
	if str != strings.ToUpper(str) {
		return fmt.Errorf("must be all uppercase")
	}

	return nil
}

// Validate for positiveConstraint validates that a numeric value is greater than 0
func (c positiveConstraint) Validate(value any) error {
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
		return fmt.Errorf("positive constraint requires numeric value")
	}

	if numValue <= 0 {
		return fmt.Errorf("must be positive (greater than 0)")
	}

	return nil
}

// Validate for negativeConstraint validates that a numeric value is less than 0
func (c negativeConstraint) Validate(value any) error {
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
		return fmt.Errorf("negative constraint requires numeric value")
	}

	if numValue >= 0 {
		return fmt.Errorf("must be negative (less than 0)")
	}

	return nil
}

// Validate for multipleOfConstraint validates that a numeric value is divisible by factor
func (c multipleOfConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Get numeric value
	var numValue float64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = v.Float()
	default:
		return fmt.Errorf("multiple_of constraint requires numeric value")
	}

	// 4. Validation logic - check if value is divisible by factor
	remainder := math.Mod(numValue, c.factor)
	// Use small epsilon for floating point comparison
	if math.Abs(remainder) > 1e-9 && math.Abs(remainder-c.factor) > 1e-9 {
		return fmt.Errorf("must be a multiple of %v", c.factor)
	}

	return nil
}

// Validate for maxDigitsConstraint validates that a numeric value has at most maxDigits digits
func (c maxDigitsConstraint) Validate(value any) error {
	// 1. Get reflect.Value
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip validation for invalid values
	}

	// 2. Handle pointer indirection
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil // Skip validation for nil pointers
		}
		v = v.Elem()
	}

	// 3. Get numeric value as string
	var str string
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Errorf("max_digits constraint requires numeric value")
	}

	// 4. Count digits (exclude minus sign and decimal point)
	digitCount := 0
	for _, r := range str {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}

	// 5. Validation logic
	if digitCount > c.maxDigits {
		return fmt.Errorf("must have at most %d digits", c.maxDigits)
	}

	return nil
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
		case "default":
			result = append(result, defaultConstraint{value: value})
		}
	}

	return result
}

// buildMinConstraint creates a min constraint, handling context-aware type checking.
// Returns (constraint, true) on success or (nil, false) if parsing fails.
func buildMinConstraint(value string, fieldType reflect.Type) (Constraint, bool) {
	min, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}

	// Handle pointer types - check underlying type
	checkType := fieldType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	kind := checkType.Kind()
	if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
		return minLengthConstraint{minLength: min}, true
	}
	return minConstraint{min: min}, true
}

// buildMaxConstraint creates a max constraint, handling context-aware type checking.
// Returns (constraint, true) on success or (nil, false) if parsing fails.
func buildMaxConstraint(value string, fieldType reflect.Type) (Constraint, bool) {
	max, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}

	// Handle pointer types - check underlying type
	checkType := fieldType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	kind := checkType.Kind()
	if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
		return maxLengthConstraint{maxLength: max}, true
	}
	return maxConstraint{max: max}, true
}

// buildRegexConstraint compiles a regex pattern constraint.
// Panics on invalid regex pattern (fail-fast approach).
func buildRegexConstraint(pattern string) Constraint {
	compiledRegex, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("invalid regex pattern '%s': %v", pattern, err))
	}
	return regexConstraint{pattern: pattern, regex: compiledRegex}
}

// buildEnumConstraint parses space-separated enum values.
func buildEnumConstraint(value string) Constraint {
	values := strings.Fields(value)
	return enumConstraint{values: values}
}

// buildLenConstraint creates a len constraint from a numeric value.
// Returns (constraint, true) on success or (nil, false) if parsing fails.
func buildLenConstraint(value string) (Constraint, bool) {
	length, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}
	return lenConstraint{length: length}, true
}

// buildContainsConstraint creates a contains constraint with the specified substring
func buildContainsConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty substring is invalid
	}
	return containsConstraint{substring: value}, true
}

// buildExcludesConstraint creates an excludes constraint with the specified substring
func buildExcludesConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty substring is invalid
	}
	return excludesConstraint{substring: value}, true
}

// buildStartswithConstraint creates a startswith constraint with the specified prefix
func buildStartswithConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty prefix is invalid
	}
	return startswithConstraint{prefix: value}, true
}

// buildEndswithConstraint creates an endswith constraint with the specified suffix
func buildEndswithConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty suffix is invalid
	}
	return endswithConstraint{suffix: value}, true
}

// buildMultipleOfConstraint creates a multiple_of constraint with the specified factor
func buildMultipleOfConstraint(value string) (Constraint, bool) {
	factor, err := strconv.ParseFloat(value, 64)
	if err != nil || factor == 0 {
		return nil, false // Invalid or zero factor
	}
	return multipleOfConstraint{factor: factor}, true
}

// buildMaxDigitsConstraint creates a max_digits constraint with the specified maximum
func buildMaxDigitsConstraint(value string) (Constraint, bool) {
	maxDigits, err := strconv.Atoi(value)
	if err != nil || maxDigits <= 0 {
		return nil, false // Invalid or non-positive max digits
	}
	return maxDigitsConstraint{maxDigits: maxDigits}, true
}
