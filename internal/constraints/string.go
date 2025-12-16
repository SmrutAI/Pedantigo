// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// String constraint types.
type (
	emailConstraint struct{}
	urlConstraint   struct{}
	uuidConstraint  struct{}
	regexConstraint struct {
		pattern string
		regex   *regexp.Regexp
	}
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
)

// emailConstraint validates that a string is a valid email format.
func (c emailConstraint) Validate(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("email constraint requires string value")
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !emailRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidEmail, "must be a valid email address")
	}

	return nil
}

// urlConstraint validates that a string is a valid URL (http or https only).
func (c urlConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("url constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse the URL
	parsedURL, err := url.Parse(str)
	if err != nil {
		return NewConstraintError(CodeInvalidURL, "must be a valid URL (http or https)")
	}

	// Check scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return NewConstraintError(CodeInvalidURL, "must be a valid URL (http or https)")
	}

	// Check host is non-empty
	if parsedURL.Host == "" {
		return NewConstraintError(CodeInvalidURL, "must be a valid URL (http or https)")
	}

	return nil
}

// uuidConstraint validates that a string is a valid UUID.
func (c uuidConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("uuid constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate UUID format using regex
	if !uuidRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidUUID, "must be a valid UUID")
	}

	return nil
}

// regexConstraint validates that a string matches a custom regex pattern.
func (c regexConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("regex constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate against the compiled regex
	if !c.regex.MatchString(str) {
		return NewConstraintErrorf(CodePatternMismatch, "must match pattern '%s'", c.pattern)
	}

	return nil
}

// lenConstraint validates that a string has exact length.
func (c lenConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	if v.Kind() != reflect.String {
		return fmt.Errorf("len constraint requires string value")
	}

	str := v.String()

	// Note: len constraint validates empty strings (len=0 is valid)
	// Do NOT skip empty strings like other constraints

	// Validation logic - count runes, not bytes (for Unicode support)
	runeCount := len([]rune(str))
	if runeCount != c.length {
		return NewConstraintErrorf(CodeExactLength, "must be exactly %d characters", c.length)
	}

	return nil
}

// asciiConstraint validates that a string contains only ASCII characters.
func (c asciiConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ascii constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check all runes are ASCII (0-127)
	for _, r := range str {
		if r > 127 {
			return NewConstraintError(CodeMustBeASCII, "must contain only ASCII characters")
		}
	}

	return nil
}

// alphaConstraint validates that a string contains only alphabetic characters.
func (c alphaConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("alpha constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string matches alphabetic pattern
	if !alphaRegex.MatchString(str) {
		return NewConstraintError(CodeMustBeAlpha, "must contain only alphabetic characters")
	}

	return nil
}

// alphanumConstraint validates that a string contains only alphanumeric characters.
func (c alphanumConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("alphanum constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string matches alphanumeric pattern
	if !alphanumRegex.MatchString(str) {
		return NewConstraintError(CodeMustBeAlphanum, "must contain only alphanumeric characters")
	}

	return nil
}

// containsConstraint validates that a string contains a specific substring.
func (c containsConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("contains constraint %w", err)
	}

	// Skip empty strings only if substring is non-empty
	if str == "" && c.substring != "" {
		return NewConstraintErrorf(CodeMustContain, "must contain '%s'", c.substring)
	}

	// Check if string contains substring
	if !strings.Contains(str, c.substring) {
		return NewConstraintErrorf(CodeMustContain, "must contain '%s'", c.substring)
	}

	return nil
}

// excludesConstraint validates that a string does NOT contain a specific substring.
func (c excludesConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("excludes constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string does NOT contain substring
	if strings.Contains(str, c.substring) {
		return NewConstraintErrorf(CodeMustNotContain, "must not contain '%s'", c.substring)
	}

	return nil
}

// startswithConstraint validates that a string starts with a specific prefix.
func (c startswithConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("startswith constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string starts with prefix
	if !strings.HasPrefix(str, c.prefix) {
		return NewConstraintErrorf(CodeMustStartWith, "must start with '%s'", c.prefix)
	}

	return nil
}

// endswithConstraint validates that a string ends with a specific suffix.
func (c endswithConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("endswith constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string ends with suffix
	if !strings.HasSuffix(str, c.suffix) {
		return NewConstraintErrorf(CodeMustEndWith, "must end with '%s'", c.suffix)
	}

	return nil
}

// lowercaseConstraint validates that a string is all lowercase.
func (c lowercaseConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("lowercase constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string is all lowercase
	if str != strings.ToLower(str) {
		return NewConstraintError(CodeMustBeLowercase, "must be all lowercase")
	}

	return nil
}

// uppercaseConstraint validates that a string is all uppercase.
func (c uppercaseConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("uppercase constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string is all uppercase
	if str != strings.ToUpper(str) {
		return NewConstraintError(CodeMustBeUppercase, "must be all uppercase")
	}

	return nil
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

// buildLenConstraint creates a len constraint from a numeric value.
// Returns (constraint, true) on success or (nil, false) if parsing fails.
func buildLenConstraint(value string) (Constraint, bool) {
	length, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}
	return lenConstraint{length: length}, true
}

// buildContainsConstraint creates a contains constraint with the specified substring.
func buildContainsConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty substring is invalid
	}
	return containsConstraint{substring: value}, true
}

// buildExcludesConstraint creates an excludes constraint with the specified substring.
func buildExcludesConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty substring is invalid
	}
	return excludesConstraint{substring: value}, true
}

// buildStartswithConstraint creates a startswith constraint with the specified prefix.
func buildStartswithConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty prefix is invalid
	}
	return startswithConstraint{prefix: value}, true
}

// buildEndswithConstraint creates an endswith constraint with the specified suffix.
func buildEndswithConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty suffix is invalid
	}
	return endswithConstraint{suffix: value}, true
}
