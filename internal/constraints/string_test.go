package constraints

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUrlConstraint tests urlConstraint.Validate() for valid URLs.
func TestUrlConstraint(t *testing.T) {
	runSimpleConstraintTests(t, urlConstraint{}, []simpleTestCase{
		// Valid HTTP URLs
		{"http URL simple", "http://example.com", false},
		{"http URL with path", "http://example.com/path", false},
		{"http URL with query", "http://example.com/path?key=value", false},
		{"http URL with subdomain", "http://api.example.com", false},
		{"http URL with port", "http://example.com:8080", false},
		{"http URL with complex path", "http://example.com/path/to/resource?id=123&name=test", false},
		// Valid HTTPS URLs
		{"https URL simple", "https://example.com", false},
		{"https URL with path", "https://example.com/path", false},
		{"https URL with query", "https://example.com/path?key=value", false},
		{"https URL with subdomain", "https://api.example.com", false},
		{"https URL with port", "https://example.com:443", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid schemes
		{"ftp scheme", "ftp://example.com", true},
		{"file scheme", "file:///etc/passwd", true},
		{"data scheme", "data:text/plain,hello", true},
		// Invalid URLs
		{"invalid URL - missing host", "http://", true},
		{"invalid URL - no scheme", "example.com", true},
		{"invalid URL - malformed", "http://exa mple.com", true},
		{"invalid URL - only path", "/path/to/resource", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestUriConstraint tests uriConstraint.Validate() for valid URIs (any scheme).
func TestUriConstraint(t *testing.T) {
	runSimpleConstraintTests(t, uriConstraint{}, []simpleTestCase{
		// Valid HTTP/HTTPS URIs
		{"http URI simple", "http://example.com", false},
		{"https URI simple", "https://example.com", false},
		{"https URI with path", "https://example.com/path", false},
		// Valid database URIs
		{"postgres URI", "postgres://user:pass@localhost:5432/smrut", false},
		{"postgresql URI", "postgresql://user@localhost/db", false},
		{"mysql URI", "mysql://root:password@127.0.0.1:3306/mydb", false},
		{"redis URI", "redis://localhost:6379/0", false},
		{"mongodb URI", "mongodb://localhost:27017/testdb", false},
		{"mongodb+srv URI", "mongodb+srv://cluster0.example.mongodb.net/mydb", false},
		// Other valid URI schemes
		{"ftp URI", "ftp://ftp.example.com/file.txt", false},
		{"sftp URI", "sftp://user@host/path", false},
		{"file URI", "file:///etc/passwd", false},
		{"data URI", "data:text/plain,hello", false},
		{"mailto URI", "mailto:user@example.com", false},
		{"tel URI", "tel:+1234567890", false},
		{"ssh URI", "ssh://git@github.com/user/repo.git", false},
		{"s3 URI", "s3://bucket-name/key", false},
		{"amqp URI", "amqp://guest:guest@localhost:5672/", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid URIs - missing scheme
		{"no scheme", "example.com", true},
		{"only path", "/path/to/resource", true},
		{"relative path", "path/to/resource", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestUuidConstraint tests uuidConstraint.Validate() for valid UUIDs.
func TestUuidConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid UUIDs (v4 format - the regex validates the format, not version)
		{name: "valid UUID lowercase", value: "550e8400-e29b-41d4-a716-446655440000", wantErr: false},
		{name: "valid UUID uppercase", value: "550E8400-E29B-41D4-A716-446655440000", wantErr: false},
		{name: "valid UUID mixed case", value: "550e8400-E29B-41D4-a716-446655440000", wantErr: false},
		{name: "valid UUID all zeros", value: "00000000-0000-0000-0000-000000000000", wantErr: false},
		{name: "valid UUID all f's", value: "ffffffff-ffff-ffff-ffff-ffffffffffff", wantErr: false},
		{name: "valid UUID numeric heavy", value: "12345678-1234-1234-1234-123456789012", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// Invalid UUID formats
		{name: "invalid UUID - no dashes", value: "550e8400e29b41d4a716446655440000", wantErr: true},
		{name: "invalid UUID - wrong dash positions", value: "550e8400-e29b-41d4a716-446655440000", wantErr: true},
		{name: "invalid UUID - too short", value: "550e8400-e29b-41d4-a716", wantErr: true},
		{name: "invalid UUID - too long", value: "550e8400-e29b-41d4-a716-446655440000-extra", wantErr: true},
		{name: "invalid UUID - invalid hex char", value: "550e8400-e29b-41d4-a716-44665544000g", wantErr: true},
		{name: "invalid UUID - spaces", value: "550e8400 -e29b-41d4-a716-446655440000", wantErr: true},
		{name: "invalid UUID - extra dashes", value: "550e8400--e29b-41d4-a716-446655440000", wantErr: true},

		// Non-UUID formats
		{name: "random string", value: "not-a-uuid-at-all-right-here", wantErr: true},
		{name: "numeric string", value: "12345678-1234-1234-1234-12345678901", wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := uuidConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestRegexConstraint tests regexConstraint.Validate() for custom patterns.
func TestRegexConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		pattern string
		wantErr bool
	}{
		// Phone number pattern
		{name: "valid phone", value: "123-456-7890", pattern: `^\d{3}-\d{3}-\d{4}$`, wantErr: false},
		{name: "invalid phone - no dashes", value: "1234567890", pattern: `^\d{3}-\d{3}-\d{4}$`, wantErr: true},
		{name: "invalid phone - wrong format", value: "12-34-5678", pattern: `^\d{3}-\d{3}-\d{4}$`, wantErr: true},

		// Alphanumeric pattern
		{name: "valid alphanumeric", value: "abc123", pattern: `^[a-zA-Z0-9]+$`, wantErr: false},
		{name: "invalid alphanumeric - spaces", value: "abc 123", pattern: `^[a-zA-Z0-9]+$`, wantErr: true},
		{name: "invalid alphanumeric - special chars", value: "abc-123", pattern: `^[a-zA-Z0-9]+$`, wantErr: true},

		// Hex color pattern
		{name: "valid hex color", value: "#FF5733", pattern: `^#[0-9A-Fa-f]{6}$`, wantErr: false},
		{name: "valid hex color lowercase", value: "#ff5733", pattern: `^#[0-9A-Fa-f]{6}$`, wantErr: false},
		{name: "invalid hex color - no hash", value: "FF5733", pattern: `^#[0-9A-Fa-f]{6}$`, wantErr: true},
		{name: "invalid hex color - invalid char", value: "#FF573G", pattern: `^#[0-9A-Fa-f]{6}$`, wantErr: true},

		// Email-like pattern
		{name: "valid email pattern", value: "user@example.com", pattern: `^[a-zA-Z0-9.]+@[a-zA-Z0-9.]+$`, wantErr: false},
		{name: "invalid email pattern - no domain", value: "user@", pattern: `^[a-zA-Z0-9.]+@[a-zA-Z0-9.]+$`, wantErr: true},

		// Simple digit pattern
		{name: "valid digits", value: "12345", pattern: `^\d+$`, wantErr: false},
		{name: "invalid digits - letter", value: "1234a", pattern: `^\d+$`, wantErr: true},

		// Empty string - should be skipped
		{name: "empty string", value: "", pattern: `^\d+$`, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), pattern: `^\d+$`, wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, pattern: `^\d+$`, wantErr: true},
		{name: "invalid type - bool", value: true, pattern: `^\d+$`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := regexp.Compile(tt.pattern)
			require.NoError(t, err, "failed to compile regex")
			constraint := regexConstraint{pattern: tt.pattern, regex: regex}
			err = constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestEmailConstraint tests emailConstraint.Validate() for email format.
func TestEmailConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid emails
		{name: "simple email", value: "user@example.com", wantErr: false},
		{name: "email with dots", value: "user.name@example.com", wantErr: false},
		{name: "email with plus", value: "user+tag@example.com", wantErr: false},
		{name: "email with dash", value: "user-name@example.com", wantErr: false},
		{name: "email with underscore", value: "user_name@example.com", wantErr: false},
		{name: "email with number", value: "user123@example.com", wantErr: false},
		{name: "email with percent", value: "user%test@example.com", wantErr: false},
		{name: "email with subdomain", value: "user@mail.example.com", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// Invalid emails
		{name: "no at symbol", value: "userexample.com", wantErr: true},
		{name: "double at", value: "user@@example.com", wantErr: true},
		{name: "no domain", value: "user@", wantErr: true},
		{name: "no local part", value: "@example.com", wantErr: true},
		{name: "no TLD", value: "user@example", wantErr: true},
		{name: "space in email", value: "user @example.com", wantErr: true},
		{name: "invalid char", value: "user#@example.com", wantErr: true},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := emailConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestBuildRegexConstraint tests buildRegexConstraint builder function.
func TestBuildRegexConstraint(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		shouldPanic bool
		testValue   string
		wantErr     bool
	}{
		// Valid patterns
		{name: "valid pattern lowercase", pattern: "^[a-z]+$", shouldPanic: false, testValue: "hello", wantErr: false},
		{name: "valid pattern digits", pattern: "^[0-9]+$", shouldPanic: false, testValue: "12345", wantErr: false},
		{name: "valid pattern email-like", pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", shouldPanic: false, testValue: "test@example.com", wantErr: false},

		// Invalid value should fail validation
		{name: "pattern mismatch", pattern: "^[a-z]+$", shouldPanic: false, testValue: "HELLO", wantErr: true},

		// Invalid pattern should panic
		{name: "invalid regex", pattern: "[unclosed", shouldPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					_ = buildRegexConstraint(tt.pattern)
				})
				return
			}

			constraint := buildRegexConstraint(tt.pattern)
			require.NotNil(t, constraint, "expected non-nil constraint")

			// Test validation if we have a test value
			if tt.testValue != "" {
				err := constraint.Validate(tt.testValue)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

// TestLenConstraint tests lenConstraint.Validate() for exact string length.
func TestLenConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		length  int
		wantErr bool
	}{
		// Valid cases - exact match
		{name: "exact match 5 chars", value: "hello", length: 5, wantErr: false},
		{name: "exact match 1 char", value: "a", length: 1, wantErr: false},
		{name: "exact match 10 chars", value: "helloworld", length: 10, wantErr: false},
		{name: "exact match 20 chars", value: "abcdefghijklmnopqrst", length: 20, wantErr: false},
		{name: "zero length with empty string", value: "", length: 0, wantErr: false},
		{name: "unicode exact match", value: "café", length: 4, wantErr: false},
		{name: "emoji exact match", value: "👍🎉", length: 2, wantErr: false},

		// Invalid cases - length mismatch
		{name: "too short by 1", value: "hi", length: 5, wantErr: true},
		{name: "too long by 1", value: "hello!", length: 5, wantErr: true},
		{name: "too short by many", value: "hi", length: 10, wantErr: true},
		{name: "too long by many", value: "hello world!", length: 5, wantErr: true},
		{name: "non-empty string with zero length", value: "a", length: 0, wantErr: true},

		// Edge cases
		{name: "nil pointer", value: (*string)(nil), length: 5, wantErr: false},
		{name: "large length match", value: "this is a very long string that we need to validate", length: 51, wantErr: false},
		{name: "spaces and special chars", value: "  hello!  ", length: 10, wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, length: 3, wantErr: true},
		{name: "invalid type - bool", value: true, length: 4, wantErr: true},
		{name: "invalid type - slice", value: []string{"a", "b"}, length: 2, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := lenConstraint{length: tt.length}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestAsciiConstraint tests asciiConstraint.Validate() for ASCII-only strings.
func TestAsciiConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - ASCII characters
		{name: "basic ASCII letters", value: "hello", wantErr: false},
		{name: "ASCII digits", value: "12345", wantErr: false},
		{name: "ASCII symbols", value: "!@#$%", wantErr: false},
		{name: "mixed ASCII", value: "Hello123!", wantErr: false},
		{name: "spaces and newlines", value: "hello\nworld\t!", wantErr: false},

		// Invalid cases - non-ASCII
		{name: "unicode accented", value: "café", wantErr: true},
		{name: "emoji", value: "hello 👍", wantErr: true},
		{name: "chinese characters", value: "你好", wantErr: true},
		{name: "cyrillic", value: "привет", wantErr: true},
		{name: "mixed ASCII and unicode", value: "hello世界", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), wantErr: false},
		{name: "single ASCII char", value: "a", wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := asciiConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

func TestAlphaConstraint(t *testing.T) {
	runSimpleConstraintTests(t, alphaConstraint{}, []simpleTestCase{
		// Valid cases - alphabetic characters only
		{"lowercase letters", "hello", false},
		{"uppercase letters", "WORLD", false},
		{"mixed case", "HelloWorld", false},
		{"single letter lowercase", "a", false},
		{"single letter uppercase", "Z", false},
		{"long alphabetic string", "thequickbrownfoxjumpsoverthelazydog", false},
		// Invalid cases - non-alphabetic characters
		{"contains digits", "hello123", true},
		{"contains spaces", "hello world", true},
		{"contains symbols", "hello!", true},
		{"only digits", "12345", true},
		{"unicode accented", "café", true},
		{"emoji", "hello👍", true},
		// Edge cases
		{"empty string", "", false},
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

func TestAlphanumConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - alphanumeric characters only
		{name: "lowercase letters only", value: "hello", wantErr: false},
		{name: "uppercase letters only", value: "WORLD", wantErr: false},
		{name: "digits only", value: "12345", wantErr: false},
		{name: "mixed letters and digits", value: "hello123", wantErr: false},
		{name: "mixed case and digits", value: "Hello123World", wantErr: false},
		{name: "single letter", value: "a", wantErr: false},
		{name: "single digit", value: "5", wantErr: false},

		// Invalid cases - non-alphanumeric characters
		{name: "contains spaces", value: "hello world", wantErr: true},
		{name: "contains symbols", value: "hello!", wantErr: true},
		{name: "contains underscore", value: "hello_world", wantErr: true},
		{name: "contains hyphen", value: "hello-123", wantErr: true},
		{name: "unicode accented", value: "café123", wantErr: true},
		{name: "emoji", value: "hello123👍", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := alphanumConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

func TestContainsConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		substring string
		wantErr   bool
	}{
		// Valid cases - substring present
		{name: "substring at start", value: "helloworld", substring: "hello", wantErr: false},
		{name: "substring in middle", value: "helloworld", substring: "low", wantErr: false},
		{name: "substring at end", value: "helloworld", substring: "world", wantErr: false},
		{name: "full match", value: "hello", substring: "hello", wantErr: false},
		{name: "single character", value: "hello", substring: "e", wantErr: false},
		{name: "repeated substring", value: "hello hello", substring: "hello", wantErr: false},
		{name: "substring with numbers", value: "test123", substring: "123", wantErr: false},

		// Invalid cases - substring absent or case mismatch
		{name: "substring not found", value: "hello", substring: "world", wantErr: true},
		{name: "case mismatch", value: "hello", substring: "HELLO", wantErr: true},
		{name: "partial match", value: "hello", substring: "helloworld", wantErr: true},
		{name: "similar characters", value: "hello", substring: "helo", wantErr: true},
		{name: "substring with space", value: "hello", substring: "hel lo", wantErr: true},
		{name: "empty value non-empty substring", value: "", substring: "test", wantErr: true},
		{name: "unicode mismatch", value: "hello", substring: "café", wantErr: true},

		// Edge cases
		{name: "empty string empty substring", value: "", substring: "", wantErr: false},
		{name: "non-empty with empty substring", value: "hello", substring: "", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), substring: "test", wantErr: false},
		{name: "unicode substring present", value: "hello café", substring: "café", wantErr: false},
		{name: "special characters", value: "test@example.com", substring: "@", wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, substring: "123", wantErr: true},
		{name: "invalid type - bool", value: true, substring: "true", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := containsConstraint{substring: tt.substring}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestExcludesConstraint tests excludesConstraint.Validate() for substring exclusion.
func TestExcludesConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		substring string
		wantErr   bool
	}{
		// Valid cases - substring NOT present
		{name: "substring not found", value: "hello", substring: "world", wantErr: false},
		{name: "case mismatch (no match)", value: "hello", substring: "HELLO", wantErr: false},
		{name: "partial match not counted", value: "hello", substring: "helloworld", wantErr: false},
		{name: "similar characters", value: "hello", substring: "helo", wantErr: false},
		{name: "substring with space", value: "hello", substring: "hel lo", wantErr: false},
		{name: "unicode mismatch", value: "hello", substring: "café", wantErr: false},

		// Invalid cases - substring IS present (should fail)
		{name: "substring at start", value: "helloworld", substring: "hello", wantErr: true},
		{name: "substring in middle", value: "helloworld", substring: "low", wantErr: true},
		{name: "substring at end", value: "helloworld", substring: "world", wantErr: true},
		{name: "full match", value: "hello", substring: "hello", wantErr: true},
		{name: "single character present", value: "hello", substring: "e", wantErr: true},
		{name: "substring with numbers", value: "test123", substring: "123", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", substring: "test", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), substring: "test", wantErr: false},
		{name: "unicode substring present", value: "hello café", substring: "café", wantErr: true},
		{name: "special characters present", value: "test@example.com", substring: "@", wantErr: true},

		// Invalid types
		{name: "invalid type - int", value: 123, substring: "123", wantErr: true},
		{name: "invalid type - bool", value: true, substring: "true", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := excludesConstraint{substring: tt.substring}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// prefixSuffixTestCase holds test cases for startswith/endswith constraints.
type prefixSuffixTestCase struct {
	name    string
	value   any
	param   string
	wantErr bool
}

// TestStartswithEndswithConstraints tests startswith and endswith constraints together.
func TestStartswithEndswithConstraints(t *testing.T) {
	// Common edge cases and invalid types
	startswithTests := []prefixSuffixTestCase{
		// Valid cases - starts with prefix
		{"starts with prefix", "helloworld", "hello", false},
		{"full match", "hello", "hello", false},
		{"single char prefix", "hello", "h", false},
		{"numeric prefix", "123abc", "123", false},
		{"unicode prefix", "café au lait", "café", false},
		{"special char prefix", "@user hello", "@user", false},
		// Invalid cases - doesn't start with prefix
		{"prefix in middle", "world hello", "hello", true},
		{"prefix at end", "worldhello", "hello", true},
		{"case mismatch", "Hello", "hello", true},
		{"prefix not found", "hello", "world", true},
		{"prefix longer than string", "hi", "hello", true},
		// Edge cases
		{"empty string", "", "test", false},
		{"nil pointer", (*string)(nil), "test", false},
		// Invalid types
		{"invalid type - int", 123, "123", true},
		{"invalid type - bool", true, "true", true},
	}
	endswithTests := []prefixSuffixTestCase{
		// Valid cases - ends with suffix
		{"ends with suffix", "helloworld", "world", false},
		{"full match", "world", "world", false},
		{"single char suffix", "hello", "o", false},
		{"numeric suffix", "abc123", "123", false},
		{"unicode suffix", "au lait café", "café", false},
		{"special char suffix", "hello @user", "@user", false},
		// Invalid cases - doesn't end with suffix
		{"suffix in middle", "world hello", "world", true},
		{"suffix at start", "worldhello", "world", true},
		{"case mismatch", "helloWorld", "world", true},
		{"suffix not found", "hello", "world", true},
		{"suffix longer than string", "hi", "hello", true},
		// Edge cases
		{"empty string", "", "test", false},
		{"nil pointer", (*string)(nil), "test", false},
		// Invalid types
		{"invalid type - int", 123, "123", true},
		{"invalid type - bool", true, "true", true},
	}

	t.Run("startswith", func(t *testing.T) {
		for _, tt := range startswithTests {
			t.Run(tt.name, func(t *testing.T) {
				c := startswithConstraint{prefix: tt.param}
				checkConstraintError(t, c.Validate(tt.value), tt.wantErr)
			})
		}
	})
	t.Run("endswith", func(t *testing.T) {
		for _, tt := range endswithTests {
			t.Run(tt.name, func(t *testing.T) {
				c := endswithConstraint{suffix: tt.param}
				checkConstraintError(t, c.Validate(tt.value), tt.wantErr)
			})
		}
	})
}

// TestLowercaseConstraint tests lowercaseConstraint.Validate() for lowercase validation.
func TestLowercaseConstraint(t *testing.T) {
	runSimpleConstraintTests(t, lowercaseConstraint{}, []simpleTestCase{
		// Valid cases - all lowercase
		{"all lowercase letters", "hello", false},
		{"lowercase with numbers", "hello123", false},
		{"lowercase with spaces", "hello world", false},
		{"lowercase with special chars", "hello@world!", false},
		{"lowercase with hyphens", "hello-world", false},
		{"numbers only", "12345", false},
		{"special chars only", "@#$%", false},
		// Invalid cases - contains uppercase
		{"single uppercase", "Hello", true},
		{"all uppercase", "HELLO", true},
		{"mixed case", "hElLo", true},
		{"uppercase at end", "hellO", true},
		{"camelCase", "helloWorld", true},
		// Edge cases
		{"empty string", "", false},
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestUppercaseConstraint tests uppercaseConstraint.Validate() for uppercase validation.
func TestUppercaseConstraint(t *testing.T) {
	runSimpleConstraintTests(t, uppercaseConstraint{}, []simpleTestCase{
		// Valid cases - all uppercase
		{"all uppercase letters", "HELLO", false},
		{"uppercase with numbers", "HELLO123", false},
		{"uppercase with spaces", "HELLO WORLD", false},
		{"uppercase with special chars", "HELLO@WORLD!", false},
		{"uppercase with hyphens", "HELLO-WORLD", false},
		{"numbers only", "12345", false},
		{"special chars only", "@#$%", false},
		// Invalid cases - contains lowercase
		{"single lowercase", "HELLo", true},
		{"all lowercase", "hello", true},
		{"mixed case", "HeLLo", true},
		{"lowercase at start", "hELLO", true},
		{"camelCase", "helloWorld", true},
		// Edge cases
		{"empty string", "", false},
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestStripWhitespaceConstraint tests stripWhitespaceConstraint.Validate()
// This constraint checks if string has NO leading/trailing whitespace.
func TestStripWhitespaceConstraint(t *testing.T) {
	runSimpleConstraintTests(t, stripWhitespaceConstraint{}, []simpleTestCase{
		// Valid cases - no leading/trailing whitespace
		{"no whitespace", "hello", false},
		{"no whitespace with internal spaces", "hello world", false},
		{"no whitespace with tabs inside", "hello\tworld", false},
		{"numbers only", "12345", false},
		{"special chars", "@#$%", false},
		{"mixed content", "hello 123 world!", false},
		{"single char", "a", false},

		// Invalid cases - has leading/trailing whitespace
		{"leading space", " hello", true},
		{"trailing space", "hello ", true},
		{"both leading and trailing", " hello ", true},
		{"leading tab", "\thello", true},
		{"trailing tab", "hello\t", true},
		{"leading newline", "\nhello", true},
		{"trailing newline", "hello\n", true},
		{"multiple leading spaces", "   hello", true},
		{"multiple trailing spaces", "hello   ", true},
		{"only whitespace", "   ", true},

		// Edge cases
		{"empty string", "", false},
		{"nil pointer", (*string)(nil), false},

		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}
