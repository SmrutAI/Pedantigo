// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// String constraint types.
type (
	emailConstraint   struct{}
	urlConstraint     struct{}
	httpURLConstraint struct{}
	uriConstraint     struct{}
	uuidConstraint    struct{}
	regexConstraint   struct {
		pattern string
		regex   *regexp.Regexp
	}
	lenConstraint             struct{ length int }
	asciiConstraint           struct{}
	alphaConstraint           struct{}
	alphanumConstraint        struct{}
	alphaspaceConstraint      struct{}
	alphanumspaceConstraint   struct{}
	printasciiConstraint      struct{}
	numericConstraint         struct{}
	numberConstraint          struct{}
	hexadecimalConstraint     struct{}
	alphaunicodeConstraint    struct{}
	alphanumunicodeConstraint struct{}
	containsConstraint        struct{ substring string }
	excludesConstraint        struct{ substring string }
	startswithConstraint      struct{ prefix string }
	endswithConstraint        struct{ suffix string }
	startsnotwithConstraint   struct{ prefix string }
	endsnotwithConstraint     struct{ suffix string }
	containsanyConstraint     struct{ chars string }
	excludesallConstraint     struct{ chars string }
	excludesruneConstraint    struct{ r rune }
	containsRuneConstraint    struct{ r rune }
	lowercaseConstraint       struct{}
	uppercaseConstraint       struct{}
	stripWhitespaceConstraint struct{}
	uuid3Constraint           struct{} // uuid3: validates UUID version 3
	uuid4Constraint           struct{} // uuid4: validates UUID version 4
	uuid5Constraint           struct{} // uuid5: validates UUID version 5
	multibyteConstraint       struct{} // multibyte: validates string has multibyte chars
	urnRfc2141Constraint      struct{} // urn_rfc2141: validates URN format (RFC 2141)
	httpsURLConstraint        struct{} // https_url: validates HTTPS URL only
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

// urlConstraint validates that a string is a valid URL (any scheme).
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
		return NewConstraintError(CodeInvalidURL, "must be a valid URL")
	}

	// Changed: Accept ANY scheme, not just http/https
	if parsedURL.Scheme == "" {
		return NewConstraintError(CodeInvalidURL, "must be a valid URL")
	}
	// Allow URLs without host for schemes like file:// or mailto:
	// Just require a scheme and either host or path
	if parsedURL.Host == "" && parsedURL.Path == "" && parsedURL.Opaque == "" {
		return NewConstraintError(CodeInvalidURL, "must be a valid URL")
	}

	return nil
}

// httpURLConstraint validates that a string is a valid HTTP/HTTPS URL.
func (c httpURLConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("http_url constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse the URL
	parsedURL, err := url.Parse(str)
	if err != nil {
		return NewConstraintError(CodeInvalidHTTPURL, "must be a valid HTTP/HTTPS URL")
	}

	// Check scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return NewConstraintError(CodeInvalidHTTPURL, "must be a valid HTTP/HTTPS URL")
	}

	// Check host is non-empty
	if parsedURL.Host == "" {
		return NewConstraintError(CodeInvalidHTTPURL, "must be a valid HTTP/HTTPS URL")
	}

	return nil
}

// uriConstraint validates that a string is a valid URI (any scheme allowed).
// This accepts database URIs like postgres://, mysql://, redis://, etc.
func (c uriConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("uri constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse the URI
	parsedURI, err := url.Parse(str)
	if err != nil {
		return NewConstraintError(CodeInvalidURI, "must be a valid URI")
	}

	// Check scheme is non-empty
	if parsedURI.Scheme == "" {
		return NewConstraintError(CodeInvalidURI, "must be a valid URI")
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

// alphaspaceConstraint validates that a string contains only ASCII letters and spaces.
func (c alphaspaceConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("alphaspace constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check each rune is ASCII letter or space
	for _, r := range str {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		if !isLetter && r != ' ' {
			return NewConstraintError(CodeMustBeAlphaSpace, "must contain only ASCII letters and spaces")
		}
	}

	return nil
}

// alphanumspaceConstraint validates that a string contains only ASCII letters, numbers, and spaces.
func (c alphanumspaceConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("alphanumspace constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check each rune is ASCII letter, digit, or space
	for _, r := range str {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit && r != ' ' {
			return NewConstraintError(CodeMustBeAlphanumSpace, "must contain only ASCII letters, numbers, and spaces")
		}
	}

	return nil
}

// printasciiConstraint validates that a string contains only printable ASCII characters (0x20-0x7E).
func (c printasciiConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("printascii constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check each rune is in printable ASCII range
	for _, r := range str {
		if r < 0x20 || r > 0x7E {
			return NewConstraintError(CodeMustBePrintableASCII, "must contain only printable ASCII characters")
		}
	}

	return nil
}

// numericConstraint validates that a string contains only numeric digits.
func (c numericConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("numeric constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string matches numeric pattern (supports signed decimals)
	if !numericRegex.MatchString(str) {
		return NewConstraintError(CodeMustBeNumeric, "must be a valid numeric value")
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

// startsnotwithConstraint validates that a string does NOT start with a specific prefix.
func (c startsnotwithConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("startsnotwith constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string DOES NOT start with prefix
	if strings.HasPrefix(str, c.prefix) {
		return NewConstraintErrorf(CodeMustNotStartWith, "must not start with '%s'", c.prefix)
	}

	return nil
}

// endsnotwithConstraint validates that a string does NOT end with a specific suffix.
func (c endsnotwithConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("endsnotwith constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string DOES NOT end with suffix
	if strings.HasSuffix(str, c.suffix) {
		return NewConstraintErrorf(CodeMustNotEndWith, "must not end with '%s'", c.suffix)
	}

	return nil
}

// containsanyConstraint validates that a string contains at least one character from a set.
func (c containsanyConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("containsany constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string contains at least one character from the set
	if !strings.ContainsAny(str, c.chars) {
		return NewConstraintErrorf(CodeMustContainAny, "must contain at least one of '%s'", c.chars)
	}

	return nil
}

// excludesallConstraint validates that a string does NOT contain any character from a set.
func (c excludesallConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("excludesall constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string does NOT contain any character from the set
	if strings.ContainsAny(str, c.chars) {
		return NewConstraintErrorf(CodeMustExcludeAll, "must not contain any of '%s'", c.chars)
	}

	return nil
}

// excludesruneConstraint validates that a string does NOT contain a specific rune.
func (c excludesruneConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("excludesrune constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string does NOT contain the specific rune
	if strings.ContainsRune(str, c.r) {
		return NewConstraintErrorf(CodeMustExcludeRune, "must not contain rune '%c'", c.r)
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

// stripWhitespaceConstraint validates that a string has no leading/trailing whitespace.
// Used in Validate() mode to check if string is already stripped.
func (c stripWhitespaceConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("strip_whitespace constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check if string has leading/trailing whitespace
	if str != strings.TrimSpace(str) {
		return NewConstraintError(CodeMustBeStripped, "must not have leading or trailing whitespace")
	}

	return nil
}

// numberConstraint validates that a string contains only unsigned integer digits (0-9).
// Unlike numeric, this rejects signs, decimals, and scientific notation.
func (c numberConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("number constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check all characters are digits (0-9)
	for _, r := range str {
		if r < '0' || r > '9' {
			return NewConstraintError(CodeMustBeNumber, "must contain only digits (0-9)")
		}
	}

	return nil
}

// hexadecimalConstraint validates that a string is a valid hexadecimal string.
// Accepts optional 0x or 0X prefix.
func (c hexadecimalConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("hexadecimal constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	s := str
	// Remove optional 0x/0X prefix
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}

	// After removing prefix, string must not be empty
	if s == "" {
		return NewConstraintError(CodeMustBeHexadecimal, "must be a valid hexadecimal string")
	}

	// Check all remaining characters are valid hex digits
	for _, r := range s {
		isDigit := r >= '0' && r <= '9'
		isLowerHex := r >= 'a' && r <= 'f'
		isUpperHex := r >= 'A' && r <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return NewConstraintError(CodeMustBeHexadecimal, "must be a valid hexadecimal string")
		}
	}

	return nil
}

// alphaunicodeConstraint validates that a string contains only Unicode letters.
func (c alphaunicodeConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("alphaunicode constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check all characters are Unicode letters
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return NewConstraintError(CodeMustBeAlphaUnicode, "must contain only Unicode letters")
		}
	}

	return nil
}

// alphanumunicodeConstraint validates that a string contains only Unicode letters and numbers.
func (c alphanumunicodeConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("alphanumunicode constraint %w", err)
	}

	if str == "" {
		return nil // Skip empty strings
	}

	// Check all characters are Unicode letters or numbers
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return NewConstraintError(CodeMustBeAlphanumUnicode, "must contain only Unicode letters and numbers")
		}
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

// buildStartsnotwithConstraint creates a startsnotwith constraint with the specified prefix.
func buildStartsnotwithConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty prefix is invalid
	}
	return startsnotwithConstraint{prefix: value}, true
}

// buildEndsnotwithConstraint creates an endsnotwith constraint with the specified suffix.
func buildEndsnotwithConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty suffix is invalid
	}
	return endsnotwithConstraint{suffix: value}, true
}

// buildContainsanyConstraint creates a containsany constraint with the specified character set.
func buildContainsanyConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty character set is invalid
	}
	return containsanyConstraint{chars: value}, true
}

// buildExcludesallConstraint creates an excludesall constraint with the specified character set.
func buildExcludesallConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty character set is invalid
	}
	return excludesallConstraint{chars: value}, true
}

// buildExcludesruneConstraint creates an excludesrune constraint with the first rune from the value.
func buildExcludesruneConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false // Empty string is invalid
	}
	runes := []rune(value)
	return excludesruneConstraint{r: runes[0]}, true
}

// containsRuneConstraint validates that a string contains a specific rune.
func (c containsRuneConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("containsrune constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !strings.ContainsRune(str, c.r) {
		return NewConstraintError(CodeContainsRune, fmt.Sprintf("must contain rune %q", string(c.r)))
	}

	return nil
}

// buildContainsRuneConstraint creates a containsrune constraint.
// The value parameter should be a single character string or Unicode code point.
func buildContainsRuneConstraint(value string) (Constraint, bool) {
	if value == "" {
		return nil, false
	}
	// Get the first rune from the string
	r, _ := utf8.DecodeRuneInString(value)
	if r == utf8.RuneError {
		return nil, false
	}
	return containsRuneConstraint{r: r}, true
}

// uuid3Constraint validates that a string is a valid UUID version 3.
func (c uuid3Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("uuid3 constraint %w", err)
	}
	if str == "" {
		return nil
	}

	// Check UUID format first
	if !uuidRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidUUIDv3, "must be a valid UUID version 3")
	}

	// Check version byte (position 14 must be '3')
	if str[14] != '3' {
		return NewConstraintError(CodeInvalidUUIDv3, "must be a valid UUID version 3")
	}

	// Check variant byte (position 19 must be 8, 9, a, b, A, B for RFC 4122)
	v := str[19]
	if v != '8' && v != '9' && v != 'a' && v != 'b' && v != 'A' && v != 'B' {
		return NewConstraintError(CodeInvalidUUIDv3, "must be a valid UUID version 3")
	}

	return nil
}

// uuid4Constraint validates that a string is a valid UUID version 4.
func (c uuid4Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("uuid4 constraint %w", err)
	}
	if str == "" {
		return nil
	}

	// Check UUID format first
	if !uuidRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidUUIDv4, "must be a valid UUID version 4")
	}

	// Check version byte (position 14 must be '4')
	if str[14] != '4' {
		return NewConstraintError(CodeInvalidUUIDv4, "must be a valid UUID version 4")
	}

	// Check variant byte (position 19 must be 8, 9, a, b, A, B for RFC 4122)
	v := str[19]
	if v != '8' && v != '9' && v != 'a' && v != 'b' && v != 'A' && v != 'B' {
		return NewConstraintError(CodeInvalidUUIDv4, "must be a valid UUID version 4")
	}

	return nil
}

// uuid5Constraint validates that a string is a valid UUID version 5.
func (c uuid5Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("uuid5 constraint %w", err)
	}
	if str == "" {
		return nil
	}

	// Check UUID format first
	if !uuidRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidUUIDv5, "must be a valid UUID version 5")
	}

	// Check version byte (position 14 must be '5')
	if str[14] != '5' {
		return NewConstraintError(CodeInvalidUUIDv5, "must be a valid UUID version 5")
	}

	// Check variant byte (position 19 must be 8, 9, a, b, A, B for RFC 4122)
	v := str[19]
	if v != '8' && v != '9' && v != 'a' && v != 'b' && v != 'A' && v != 'B' {
		return NewConstraintError(CodeInvalidUUIDv5, "must be a valid UUID version 5")
	}

	return nil
}

// multibyteConstraint validates that a string contains at least one multibyte character.
func (c multibyteConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("multibyte constraint %w", err)
	}
	if str == "" {
		return nil
	}

	// Check if string contains any multibyte character (> 127 in UTF-8)
	for _, r := range str {
		if r > 127 {
			return nil // Found a multibyte character
		}
	}

	return NewConstraintError(CodeInvalidMultibyte, "must contain at least one multibyte character")
}

// urnRfc2141Constraint validates that a string is a valid URN per RFC 2141.
func (c urnRfc2141Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("urn_rfc2141 constraint %w", err)
	}
	if str == "" {
		return nil
	}

	// Validate URN format: urn:<NID>:<NSS>
	if !urnRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidURN, "must be a valid URN (RFC 2141)")
	}

	return nil
}

// httpsURLConstraint validates that a string is a valid HTTPS URL.
func (c httpsURLConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("https_url constraint %w", err)
	}
	if str == "" {
		return nil
	}

	// Parse the URL
	parsedURL, parseErr := url.Parse(str)
	if parseErr != nil {
		return NewConstraintError(CodeInvalidHTTPSURL, "must be a valid HTTPS URL")
	}

	// Check scheme is https (case-insensitive)
	if !strings.EqualFold(parsedURL.Scheme, "https") {
		return NewConstraintError(CodeInvalidHTTPSURL, "must be a valid HTTPS URL")
	}

	// Check host is non-empty
	if parsedURL.Host == "" {
		return NewConstraintError(CodeInvalidHTTPSURL, "must be a valid HTTPS URL")
	}

	return nil
}
