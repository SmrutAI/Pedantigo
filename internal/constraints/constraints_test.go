package constraints

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaxConstraint tests maxConstraint.Validate() for numeric values
func TestMaxConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		max     int
		wantErr bool
	}{
		// Valid cases - below max
		{name: "int below max", value: 50, max: 100, wantErr: false},
		{name: "int at max", value: 100, max: 100, wantErr: false},
		{name: "int zero with positive max", value: 0, max: 100, wantErr: false},
		{name: "int negative with positive max", value: -50, max: 100, wantErr: false},

		// Invalid cases - exceeds max
		{name: "int exceeds max", value: 150, max: 100, wantErr: true},
		{name: "int exceeds negative max", value: -3, max: -5, wantErr: true},

		// Float values
		{name: "float below max", value: 3.5, max: 10, wantErr: false},
		{name: "float at max", value: 10.0, max: 10, wantErr: false},
		{name: "float exceeds max", value: 10.1, max: 10, wantErr: true},

		// Negative max
		{name: "negative value with negative max (valid)", value: -10, max: -5, wantErr: false},
		{name: "negative value with negative max (invalid)", value: -3, max: -5, wantErr: true},

		// Different numeric types
		{name: "int8 below max", value: int8(50), max: 100, wantErr: false},
		{name: "int8 exceeds max", value: int8(127), max: 100, wantErr: true},
		{name: "int16 below max", value: int16(50), max: 100, wantErr: false},
		{name: "int32 below max", value: int32(50), max: 100, wantErr: false},
		{name: "int64 below max", value: int64(50), max: 100, wantErr: false},
		{name: "uint below max", value: uint(50), max: 100, wantErr: false},
		{name: "uint exceeds max", value: uint(150), max: 100, wantErr: true},
		{name: "uint8 below max", value: uint8(50), max: 100, wantErr: false},
		{name: "uint16 below max", value: uint16(50), max: 100, wantErr: false},
		{name: "uint32 below max", value: uint32(50), max: 100, wantErr: false},
		{name: "uint64 below max", value: uint64(50), max: 100, wantErr: false},
		{name: "float32 below max", value: float32(3.5), max: 10, wantErr: false},
		{name: "float32 exceeds max", value: float32(10.5), max: 10, wantErr: true},
		{name: "float64 below max", value: float64(3.5), max: 10, wantErr: false},
		{name: "float64 exceeds max", value: float64(10.5), max: 10, wantErr: true},

		// String (length check)
		{name: "string below max length", value: "hello", max: 10, wantErr: false},
		{name: "string at max length", value: "helloworld", max: 10, wantErr: false},
		{name: "string exceeds max length", value: "hello world!", max: 10, wantErr: true},
		{name: "empty string with max length", value: "", max: 10, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), max: 100, wantErr: false},

		// Invalid type
		{name: "invalid type - bool", value: true, max: 100, wantErr: true},
		{name: "invalid type - slice", value: []int{1, 2, 3}, max: 100, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := maxConstraint{max: tt.max}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMaxLengthConstraint tests maxLengthConstraint.Validate() for strings
func TestMaxLengthConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		maxLength int
		wantErr   bool
	}{
		// Valid cases
		{name: "string below max length", value: "hello", maxLength: 10, wantErr: false},
		{name: "string at max length", value: "helloworld", maxLength: 10, wantErr: false},
		{name: "empty string", value: "", maxLength: 10, wantErr: false},
		{name: "single char at max length", value: "a", maxLength: 1, wantErr: false},

		// Invalid cases - exceeds max length
		{name: "string exceeds max length", value: "hello world!", maxLength: 10, wantErr: true},
		{name: "string far exceeds max length", value: "this is a very long string", maxLength: 5, wantErr: true},
		{name: "string one char over max", value: "ab", maxLength: 1, wantErr: true},

		// Edge cases
		{name: "max length zero with empty string", value: "", maxLength: 0, wantErr: false},
		{name: "max length zero with non-empty string", value: "a", maxLength: 0, wantErr: true},
		{name: "large max length", value: "hello", maxLength: 1000, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), maxLength: 10, wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, maxLength: 10, wantErr: true},
		{name: "invalid type - bool", value: true, maxLength: 10, wantErr: true},
		{name: "invalid type - slice", value: []string{"a", "b"}, maxLength: 10, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := maxLengthConstraint{maxLength: tt.maxLength}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLtConstraint tests ltConstraint.Validate() for < threshold
func TestLtConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value < threshold
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: false},
		{name: "int at boundary", value: 19, threshold: 20, wantErr: false},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: false},
		{name: "negative less than zero", value: -10, threshold: 0, wantErr: false},
		{name: "negative less than positive", value: -5, threshold: 10, wantErr: false},

		// Invalid cases - value >= threshold
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: true},
		{name: "int exceeds threshold", value: 25, threshold: 20, wantErr: true},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: true},
		{name: "float exceeds threshold", value: 6.5, threshold: 5.0, wantErr: true},

		// Different numeric types
		{name: "int8 less than threshold", value: int8(10), threshold: 20, wantErr: false},
		{name: "int8 exceeds threshold", value: int8(25), threshold: 20, wantErr: true},
		{name: "uint less than threshold", value: uint(10), threshold: 20, wantErr: false},
		{name: "uint exceeds threshold", value: uint(25), threshold: 20, wantErr: true},
		{name: "float32 less than threshold", value: float32(3.14), threshold: 5.0, wantErr: false},
		{name: "float32 exceeds threshold", value: float32(6.5), threshold: 5.0, wantErr: true},

		// Zero and near-zero thresholds
		{name: "negative less than zero threshold", value: -10, threshold: 0, wantErr: false},
		{name: "zero equals zero threshold", value: 0, threshold: 0, wantErr: true},
		{name: "positive exceeds zero threshold", value: 5, threshold: 0, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
		{name: "invalid type - bool", value: true, threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ltConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLeConstraint tests leConstraint.Validate() for <= threshold
func TestLeConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value <= threshold
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: false},
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: false},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: false},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: false},
		{name: "negative less than zero", value: -10, threshold: 0, wantErr: false},
		{name: "zero equals zero", value: 0, threshold: 0, wantErr: false},

		// Invalid cases - value > threshold
		{name: "int exceeds threshold", value: 25, threshold: 20, wantErr: true},
		{name: "float exceeds threshold", value: 6.5, threshold: 5.0, wantErr: true},
		{name: "positive exceeds zero", value: 5, threshold: 0, wantErr: true},

		// Different numeric types
		{name: "int8 equals threshold", value: int8(20), threshold: 20, wantErr: false},
		{name: "int8 exceeds threshold", value: int8(25), threshold: 20, wantErr: true},
		{name: "uint less than threshold", value: uint(10), threshold: 20, wantErr: false},
		{name: "uint equals threshold", value: uint(20), threshold: 20, wantErr: false},
		{name: "uint exceeds threshold", value: uint(25), threshold: 20, wantErr: true},
		{name: "float32 equals threshold", value: float32(5.0), threshold: 5.0, wantErr: false},
		{name: "float32 exceeds threshold", value: float32(6.5), threshold: 5.0, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
		{name: "invalid type - bool", value: true, threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := leConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUrlConstraint tests urlConstraint.Validate() for valid URLs
func TestUrlConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid HTTP URLs
		{name: "http URL simple", value: "http://example.com", wantErr: false},
		{name: "http URL with path", value: "http://example.com/path", wantErr: false},
		{name: "http URL with query", value: "http://example.com/path?key=value", wantErr: false},
		{name: "http URL with subdomain", value: "http://api.example.com", wantErr: false},
		{name: "http URL with port", value: "http://example.com:8080", wantErr: false},
		{name: "http URL with complex path", value: "http://example.com/path/to/resource?id=123&name=test", wantErr: false},

		// Valid HTTPS URLs
		{name: "https URL simple", value: "https://example.com", wantErr: false},
		{name: "https URL with path", value: "https://example.com/path", wantErr: false},
		{name: "https URL with query", value: "https://example.com/path?key=value", wantErr: false},
		{name: "https URL with subdomain", value: "https://api.example.com", wantErr: false},
		{name: "https URL with port", value: "https://example.com:443", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// Invalid schemes
		{name: "ftp scheme", value: "ftp://example.com", wantErr: true},
		{name: "file scheme", value: "file:///etc/passwd", wantErr: true},
		{name: "data scheme", value: "data:text/plain,hello", wantErr: true},

		// Invalid URLs
		{name: "invalid URL - missing host", value: "http://", wantErr: true},
		{name: "invalid URL - no scheme", value: "example.com", wantErr: true},
		{name: "invalid URL - malformed", value: "http://exa mple.com", wantErr: true},
		{name: "invalid URL - only path", value: "/path/to/resource", wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := urlConstraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUuidConstraint tests uuidConstraint.Validate() for valid UUIDs
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRegexConstraint tests regexConstraint.Validate() for custom patterns
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIpv4Constraint tests ipv4Constraint.Validate() for valid IPv4 addresses
func TestIpv4Constraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid IPv4 addresses
		{name: "valid IPv4 - localhost", value: "127.0.0.1", wantErr: false},
		{name: "valid IPv4 - private range", value: "192.168.1.1", wantErr: false},
		{name: "valid IPv4 - private range 10", value: "10.0.0.1", wantErr: false},
		{name: "valid IPv4 - private range 172", value: "172.16.0.1", wantErr: false},
		{name: "valid IPv4 - zeros", value: "0.0.0.0", wantErr: false},
		{name: "valid IPv4 - broadcast", value: "255.255.255.255", wantErr: false},
		{name: "valid IPv4 - google DNS", value: "8.8.8.8", wantErr: false},
		{name: "valid IPv4 - public IP", value: "1.1.1.1", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// Invalid IPv4 addresses
		{name: "invalid IPv4 - out of range", value: "256.1.1.1", wantErr: true},
		{name: "invalid IPv4 - too few octets", value: "192.168.1", wantErr: true},
		{name: "invalid IPv4 - too many octets", value: "192.168.1.1.1", wantErr: true},
		{name: "invalid IPv4 - letters", value: "192.168.a.1", wantErr: true},
		{name: "invalid IPv4 - empty octets", value: "192.168..1", wantErr: true},

		// IPv6 addresses - should fail
		{name: "IPv6 address fails", value: "::1", wantErr: true},
		{name: "IPv6 full address fails", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", wantErr: true},
		{name: "IPv6 compressed fails", value: "2001:db8::1", wantErr: true},

		// Other invalid formats
		{name: "hostname not IP", value: "example.com", wantErr: true},
		{name: "CIDR notation not IP", value: "192.168.1.0/24", wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ipv4Constraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIpv6Constraint tests ipv6Constraint.Validate() for valid IPv6 addresses
func TestIpv6Constraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid IPv6 addresses - full form
		{name: "valid IPv6 - localhost loopback", value: "::1", wantErr: false},
		{name: "valid IPv6 - unspecified", value: "::", wantErr: false},
		{name: "valid IPv6 - full form", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", wantErr: false},

		// Valid IPv6 addresses - compressed form
		{name: "valid IPv6 - compressed", value: "2001:db8::1", wantErr: false},
		{name: "valid IPv6 - compressed form 2", value: "2001:db8::8a2e:370:7334", wantErr: false},
		{name: "valid IPv6 - link-local", value: "fe80::1", wantErr: false},
		{name: "valid IPv6 - multicast", value: "ff02::1", wantErr: false},

		// Valid IPv6 addresses - zone ID variants (some may fail depending on implementation)
		{name: "valid IPv6 - with numbers", value: "1234:5678:90ab:cdef:1234:5678:90ab:cdef", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// IPv4 addresses - should fail
		{name: "IPv4 localhost fails", value: "127.0.0.1", wantErr: true},
		{name: "IPv4 private fails", value: "192.168.1.1", wantErr: true},
		{name: "IPv4 google DNS fails", value: "8.8.8.8", wantErr: true},

		// Invalid IPv6 addresses
		{name: "invalid IPv6 - too many colons", value: "2001::db8:::1", wantErr: true},
		{name: "invalid IPv6 - invalid hex", value: "2001:db8::gggg", wantErr: true},
		{name: "invalid IPv6 - too many groups", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334:extra", wantErr: true},
		{name: "invalid IPv6 - incomplete", value: "2001:db8:", wantErr: true},

		// Other invalid formats
		{name: "hostname not IP", value: "example.com", wantErr: true},
		{name: "IPv6 with port fails", value: "[2001:db8::1]:8080", wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ipv6Constraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEnumConstraint tests enumConstraint.Validate() for allowed values
func TestEnumConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		values  []string
		wantErr bool
	}{
		// String values
		{name: "string in enum", value: "active", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string in enum - first value", value: "pending", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string in enum - last value", value: "completed", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string not in enum", value: "invalid", values: []string{"pending", "active", "completed"}, wantErr: true},
		{name: "string case sensitive", value: "Active", values: []string{"pending", "active", "completed"}, wantErr: true},

		// Int values
		{name: "int in enum", value: 1, values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int in enum - zero", value: 0, values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int not in enum", value: 5, values: []string{"0", "1", "2"}, wantErr: true},
		{name: "int8 in enum", value: int8(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int16 in enum", value: int16(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int32 in enum", value: int32(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int64 in enum", value: int64(1), values: []string{"0", "1", "2"}, wantErr: false},

		// Uint values
		{name: "uint in enum", value: uint(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint not in enum", value: uint(5), values: []string{"0", "1", "2"}, wantErr: true},
		{name: "uint8 in enum", value: uint8(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint16 in enum", value: uint16(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint32 in enum", value: uint32(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint64 in enum", value: uint64(1), values: []string{"0", "1", "2"}, wantErr: false},

		// Float values
		{name: "float in enum", value: 3.14, values: []string{"2.71", "3.14", "2.0"}, wantErr: false},
		{name: "float not in enum", value: 1.5, values: []string{"2.71", "3.14", "2.0"}, wantErr: true},
		{name: "float32 in enum", value: float32(2.5), values: []string{"2.5", "1.5"}, wantErr: false},
		{name: "float64 in enum", value: float64(3.14), values: []string{"3.14", "2.71"}, wantErr: false},

		// Bool values
		{name: "bool true in enum", value: true, values: []string{"true", "false"}, wantErr: false},
		{name: "bool false in enum", value: false, values: []string{"true", "false"}, wantErr: false},

		// Empty enum
		{name: "value not in empty enum", value: "test", values: []string{}, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), values: []string{"a", "b"}, wantErr: false},

		// Invalid types
		{name: "invalid type - slice", value: []string{"a"}, values: []string{"a", "b"}, wantErr: true},
		{name: "invalid type - map", value: map[string]int{"a": 1}, values: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := enumConstraint{values: tt.values}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultConstraint tests defaultConstraint.Validate() - no-op validator
func TestDefaultConstraint(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "string value", value: "test"},
		{name: "int value", value: 42},
		{name: "float value", value: 3.14},
		{name: "bool value", value: true},
		{name: "nil value", value: nil},
		{name: "empty string", value: ""},
		{name: "zero int", value: 0},
		{name: "slice", value: []int{1, 2, 3}},
		{name: "map", value: map[string]int{"a": 1}},
	}

	constraint := defaultConstraint{value: "default"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// defaultConstraint always returns nil - it's a no-op validator
			err := constraint.Validate(tt.value)
			assert.NoError(t, err)
		})
	}
}

// TestMinConstraint tests minConstraint.Validate() for numeric values
// Added to ensure comprehensive coverage of all constraints mentioned
// TestMinConstraint tests MinConstraint validation
func TestMinConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		min     int
		wantErr bool
	}{
		// Valid cases - above min
		{name: "int above min", value: 50, min: 10, wantErr: false},
		{name: "int at min", value: 10, min: 10, wantErr: false},
		{name: "int zero with negative min", value: 0, min: -10, wantErr: false},

		// Invalid cases - below min
		{name: "int below min", value: 5, min: 10, wantErr: true},
		{name: "int far below min", value: -50, min: 10, wantErr: true},

		// Float values
		{name: "float above min", value: 5.0, min: 3, wantErr: false},
		{name: "float at min", value: 3.0, min: 3, wantErr: false},
		{name: "float below min", value: 2.5, min: 3, wantErr: true},

		// Different numeric types
		{name: "int8 above min", value: int8(50), min: 10, wantErr: false},
		{name: "uint above min", value: uint(50), min: 10, wantErr: false},
		{name: "float32 above min", value: float32(5.0), min: 3, wantErr: false},

		// String (length check)
		{name: "string above min length", value: "hello", min: 3, wantErr: false},
		{name: "string at min length", value: "hel", min: 3, wantErr: false},
		{name: "string below min length", value: "hi", min: 3, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), min: 10, wantErr: false},

		// Invalid types
		{name: "invalid type - bool", value: true, min: 10, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := minConstraint{min: tt.min}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGtConstraint tests gtConstraint.Validate() for > threshold
func TestGtConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value > threshold
		{name: "int greater than threshold", value: 25, threshold: 20, wantErr: false},
		{name: "float greater than threshold", value: 6.5, threshold: 5.0, wantErr: false},
		{name: "negative greater than negative", value: -5, threshold: -10, wantErr: false},

		// Invalid cases - value <= threshold
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: true},
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: true},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: true},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: true},

		// Different numeric types
		{name: "int8 greater than threshold", value: int8(25), threshold: 20, wantErr: false},
		{name: "uint greater than threshold", value: uint(25), threshold: 20, wantErr: false},
		{name: "float32 greater than threshold", value: float32(6.5), threshold: 5.0, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := gtConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGeConstraint tests geConstraint.Validate() for >= threshold
func TestGeConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value >= threshold
		{name: "int greater than threshold", value: 25, threshold: 20, wantErr: false},
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: false},
		{name: "float greater than threshold", value: 6.5, threshold: 5.0, wantErr: false},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: false},

		// Invalid cases - value < threshold
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: true},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: true},

		// Different numeric types
		{name: "int8 greater than threshold", value: int8(25), threshold: 20, wantErr: false},
		{name: "int8 equals threshold", value: int8(20), threshold: 20, wantErr: false},
		{name: "uint greater than threshold", value: uint(25), threshold: 20, wantErr: false},
		{name: "float32 equals threshold", value: float32(5.0), threshold: 5.0, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := geConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMinLengthConstraint tests minLengthConstraint.Validate() for strings
func TestMinLengthConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		minLength int
		wantErr   bool
	}{
		// Valid cases
		{name: "string above min length", value: "hello", minLength: 3, wantErr: false},
		{name: "string at min length", value: "hel", minLength: 3, wantErr: false},
		{name: "long string", value: "this is a long string", minLength: 5, wantErr: false},
		{name: "single char at min", value: "a", minLength: 1, wantErr: false},

		// Invalid cases - below min length
		{name: "string below min length", value: "hi", minLength: 3, wantErr: true},
		{name: "empty string below min", value: "", minLength: 1, wantErr: true},

		// Edge cases
		{name: "min length zero with empty string", value: "", minLength: 0, wantErr: false},
		{name: "min length zero with non-empty", value: "a", minLength: 0, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), minLength: 3, wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, minLength: 3, wantErr: true},
		{name: "invalid type - bool", value: true, minLength: 3, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := minLengthConstraint{minLength: tt.minLength}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEmailConstraint tests emailConstraint.Validate() for email format
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBuildMaxConstraint tests buildMaxConstraint builder function
func TestBuildMaxConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldType reflect.Type
		wantType  string
		wantOk    bool
	}{
		// String type should create maxLengthConstraint
		{name: "string field", value: "10", fieldType: reflect.TypeOf(""), wantType: "maxLengthConstraint", wantOk: true},
		{name: "slice field", value: "5", fieldType: reflect.TypeOf([]int{}), wantType: "maxLengthConstraint", wantOk: true},
		{name: "array field", value: "3", fieldType: reflect.TypeOf([3]int{}), wantType: "maxLengthConstraint", wantOk: true},

		// Numeric types should create maxConstraint
		{name: "int field", value: "100", fieldType: reflect.TypeOf(0), wantType: "maxConstraint", wantOk: true},
		{name: "int64 field", value: "100", fieldType: reflect.TypeOf(int64(0)), wantType: "maxConstraint", wantOk: true},
		{name: "uint field", value: "100", fieldType: reflect.TypeOf(uint(0)), wantType: "maxConstraint", wantOk: true},
		{name: "float64 field", value: "100", fieldType: reflect.TypeOf(float64(0)), wantType: "maxConstraint", wantOk: true},

		// Pointer types should unwrap
		{name: "pointer to string", value: "10", fieldType: reflect.TypeOf((*string)(nil)), wantType: "maxLengthConstraint", wantOk: true},
		{name: "pointer to int", value: "100", fieldType: reflect.TypeOf((*int)(nil)), wantType: "maxConstraint", wantOk: true},

		// Invalid value should fail
		{name: "invalid value", value: "not-a-number", fieldType: reflect.TypeOf(0), wantType: "", wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, ok := buildMaxConstraint(tt.value, tt.fieldType)

			assert.Equal(t, tt.wantOk, ok)

			if !ok {
				return
			}

			switch tt.wantType {
			case "maxLengthConstraint":
				_, isMaxLength := constraint.(maxLengthConstraint)
				assert.True(t, isMaxLength, "expected maxLengthConstraint, got %T", constraint)
			case "maxConstraint":
				_, isMax := constraint.(maxConstraint)
				assert.True(t, isMax, "expected maxConstraint, got %T", constraint)
			}
		})
	}
}

// TestBuildRegexConstraint tests buildRegexConstraint builder function
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

// TestBuildEnumConstraint tests buildEnumConstraint builder function
func TestBuildEnumConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		testValue any
		wantErr   bool
	}{
		// Valid enum values
		{name: "single value match", value: "red", testValue: "red", wantErr: false},
		{name: "multiple values match first", value: "red green blue", testValue: "red", wantErr: false},
		{name: "multiple values match middle", value: "red green blue", testValue: "green", wantErr: false},
		{name: "multiple values match last", value: "red green blue", testValue: "blue", wantErr: false},

		// Invalid enum values
		{name: "single value mismatch", value: "red", testValue: "blue", wantErr: true},
		{name: "multiple values mismatch", value: "red green blue", testValue: "yellow", wantErr: true},
		{name: "case sensitive mismatch", value: "red", testValue: "RED", wantErr: true},

		// Empty and whitespace handling
		{name: "multiple spaces", value: "red   green   blue", testValue: "green", wantErr: false},
		{name: "trailing spaces", value: "red green blue  ", testValue: "blue", wantErr: false},
		{name: "leading spaces", value: "  red green blue", testValue: "red", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := buildEnumConstraint(tt.value)
			require.NotNil(t, constraint, "expected non-nil constraint")

			err := constraint.Validate(tt.testValue)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestToFloat64_AllNumericTypes tests toFloat64 with all numeric type cases
func TestToFloat64_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected float64
	}{
		// Signed integers
		{name: "int", value: int(42), expected: 42.0},
		{name: "int8", value: int8(42), expected: 42.0},
		{name: "int16", value: int16(42), expected: 42.0},
		{name: "int32", value: int32(42), expected: 42.0},
		{name: "int64", value: int64(42), expected: 42.0},
		// Unsigned integers
		{name: "uint", value: uint(42), expected: 42.0},
		{name: "uint8", value: uint8(42), expected: 42.0},
		{name: "uint16", value: uint16(42), expected: 42.0},
		{name: "uint32", value: uint32(42), expected: 42.0},
		{name: "uint64", value: uint64(42), expected: 42.0},
		// Floats
		{name: "float32", value: float32(42.5), expected: 42.5},
		{name: "float64", value: float64(42.5), expected: 42.5},
		// Non-numeric (returns 0)
		{name: "string", value: "test", expected: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			result := toFloat64(val)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCheckTypeCompatibility_BoolAndTime tests missing branches in CheckTypeCompatibility
func TestCheckTypeCompatibility_BoolAndTime(t *testing.T) {
	tests := []struct {
		name    string
		a       any
		b       any
		wantErr bool
	}{
		// Bool types
		{name: "bool compatible", a: true, b: false, wantErr: false},
		{name: "bool vs int incompatible", a: true, b: 42, wantErr: true},
		// Time types
		{name: "time.Time compatible", a: time.Now(), b: time.Now(), wantErr: false},
		{name: "time vs string incompatible", a: time.Now(), b: "test", wantErr: true},
		// Nil cases
		{name: "both nil", a: nil, b: nil, wantErr: false},
		{name: "one nil non-pointer", a: nil, b: 42, wantErr: true},
		{name: "nil vs string", a: "test", b: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckTypeCompatibility(tt.a, tt.b)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDereference_PointerLevels tests Dereference with various pointer levels
func TestDereference_PointerLevels(t *testing.T) {
	tests := []struct {
		name     string
		getType  func() reflect.Type
		expected reflect.Kind
	}{
		{
			name:     "non-pointer",
			getType:  func() reflect.Type { return reflect.TypeOf(42) },
			expected: reflect.Int,
		},
		{
			name: "single pointer",
			getType: func() reflect.Type {
				x := 42
				return reflect.TypeOf(&x)
			},
			expected: reflect.Int,
		},
		{
			name: "double pointer",
			getType: func() reflect.Type {
				x := 42
				p1 := &x
				return reflect.TypeOf(&p1)
			},
			expected: reflect.Int,
		},
		{
			name: "triple pointer",
			getType: func() reflect.Type {
				x := 42
				p1 := &x
				p2 := &p1
				return reflect.TypeOf(&p2)
			},
			expected: reflect.Int,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Dereference(tt.getType())
			assert.Equal(t, tt.expected, result.Kind())
		})
	}
}

// TestCompareToString_BoolAndDefault tests missing branches in CompareToString
func TestCompareToString_BoolAndDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		// Bool cases
		{name: "bool true", value: true, expected: "true"},
		{name: "bool false", value: false, expected: "false"},
		// Pointer to bool
		{name: "pointer to bool", value: func() *bool { b := true; return &b }(), expected: "true"},
		{name: "nil pointer", value: (*int)(nil), expected: ""},
		// Default case (non-standard types)
		{name: "struct default", value: struct{ X int }{X: 42}, expected: "{42}"},
		{name: "slice default", value: []int{1, 2, 3}, expected: "[1 2 3]"},
		// Already covered types (sanity check)
		{name: "string", value: "test", expected: "test"},
		{name: "int", value: 42, expected: "42"},
		{name: "uint", value: uint(42), expected: "42"},
		{name: "float", value: 42.5, expected: "42.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareToString(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildConstraints_MissingBranches tests uncovered constraint types in BuildConstraints
func TestBuildConstraints_MissingBranches(t *testing.T) {
	tests := []struct {
		name          string
		constraints   map[string]string
		fieldType     reflect.Type
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "gt constraint",
			constraints:   map[string]string{"gt": "10.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"gtConstraint"},
		},
		{
			name:          "gte constraint",
			constraints:   map[string]string{"gte": "20.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"geConstraint"},
		},
		{
			name:          "lt constraint",
			constraints:   map[string]string{"lt": "30.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"ltConstraint"},
		},
		{
			name:          "lte constraint",
			constraints:   map[string]string{"lte": "40.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"leConstraint"},
		},
		{
			name:          "ipv4 constraint",
			constraints:   map[string]string{"ipv4": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"ipv4Constraint"},
		},
		{
			name:          "ipv6 constraint",
			constraints:   map[string]string{"ipv6": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"ipv6Constraint"},
		},
		{
			name:          "default constraint",
			constraints:   map[string]string{"default": "test"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"defaultConstraint"},
		},
		{
			name:          "gt with invalid float",
			constraints:   map[string]string{"gt": "invalid"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "gte with invalid float",
			constraints:   map[string]string{"gte": "not-a-number"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "lt with invalid float",
			constraints:   map[string]string{"lt": "xyz"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "lte with invalid float",
			constraints:   map[string]string{"lte": "abc"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "email constraint",
			constraints:   map[string]string{"email": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"emailConstraint"},
		},
		{
			name:          "url constraint",
			constraints:   map[string]string{"url": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"urlConstraint"},
		},
		{
			name:          "uuid constraint",
			constraints:   map[string]string{"uuid": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"uuidConstraint"},
		},
		{
			name:          "regexp constraint",
			constraints:   map[string]string{"regexp": "^[a-z]+$"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"regexConstraint"},
		},
		{
			name:          "oneof constraint",
			constraints:   map[string]string{"oneof": "red green blue"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"enumConstraint"},
		},
		{
			name:          "required constraint (skipped)",
			constraints:   map[string]string{"required": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "multiple constraints",
			constraints:   map[string]string{"gt": "5", "lte": "100", "ipv4": "", "default": "10"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 4,
			expectedTypes: []string{"gtConstraint", "leConstraint", "ipv4Constraint", "defaultConstraint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildConstraints(tt.constraints, tt.fieldType)
			assert.Equal(t, tt.expectedCount, len(result))

			// Verify constraint types (order may vary due to map iteration)
			if len(tt.expectedTypes) > 0 {
				foundTypes := make(map[string]bool)
				for _, c := range result {
					typeName := reflect.TypeOf(c).Name()
					foundTypes[typeName] = true
				}
				for _, expectedType := range tt.expectedTypes {
					assert.True(t, foundTypes[expectedType], "Expected constraint type %s not found", expectedType)
				}
			}
		})
	}
}

// TestParseConditionalConstraint_ErrorPath tests the false return branch
func TestParseConditionalConstraint_ErrorPath(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		separator  string
		wantOk     bool
		wantFirst  string
		wantSecond string
	}{
		{
			name:       "valid two parts",
			value:      "field:value",
			separator:  ":",
			wantOk:     true,
			wantFirst:  "field",
			wantSecond: "value",
		},
		{
			name:       "no separator",
			value:      "fieldvalue",
			separator:  ":",
			wantOk:     false,
			wantFirst:  "",
			wantSecond: "",
		},
		{
			name:       "empty value",
			value:      "",
			separator:  ":",
			wantOk:     false,
			wantFirst:  "",
			wantSecond: "",
		},
		{
			name:       "only separator",
			value:      ":",
			separator:  ":",
			wantOk:     true,
			wantFirst:  "",
			wantSecond: "",
		},
		{
			name:       "multiple separators (splits on first)",
			value:      "field:value:extra",
			separator:  ":",
			wantOk:     true,
			wantFirst:  "field",
			wantSecond: "value:extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, second, ok := parseConditionalConstraint(tt.value, tt.separator)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantFirst, first)
			assert.Equal(t, tt.wantSecond, second)
		})
	}
}

// TestLenConstraint tests lenConstraint.Validate() for exact string length
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAsciiConstraint tests asciiConstraint.Validate() for ASCII-only strings
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAlphaConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - alphabetic characters only
		{name: "lowercase letters", value: "hello", wantErr: false},
		{name: "uppercase letters", value: "WORLD", wantErr: false},
		{name: "mixed case", value: "HelloWorld", wantErr: false},
		{name: "single letter lowercase", value: "a", wantErr: false},
		{name: "single letter uppercase", value: "Z", wantErr: false},
		{name: "long alphabetic string", value: "thequickbrownfoxjumpsoverthelazydog", wantErr: false},

		// Invalid cases - non-alphabetic characters
		{name: "contains digits", value: "hello123", wantErr: true},
		{name: "contains spaces", value: "hello world", wantErr: true},
		{name: "contains symbols", value: "hello!", wantErr: true},
		{name: "only digits", value: "12345", wantErr: true},
		{name: "unicode accented", value: "café", wantErr: true},
		{name: "emoji", value: "hello👍", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := alphaConstraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestExcludesConstraint tests excludesConstraint.Validate() for substring exclusion
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

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStartswithConstraint tests startswithConstraint.Validate() for prefix validation
func TestStartswithConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		prefix  string
		wantErr bool
	}{
		// Valid cases - starts with prefix
		{name: "starts with prefix", value: "helloworld", prefix: "hello", wantErr: false},
		{name: "full match", value: "hello", prefix: "hello", wantErr: false},
		{name: "single char prefix", value: "hello", prefix: "h", wantErr: false},
		{name: "numeric prefix", value: "123abc", prefix: "123", wantErr: false},
		{name: "unicode prefix", value: "café au lait", prefix: "café", wantErr: false},
		{name: "special char prefix", value: "@user hello", prefix: "@user", wantErr: false},

		// Invalid cases - doesn't start with prefix
		{name: "prefix in middle", value: "world hello", prefix: "hello", wantErr: true},
		{name: "prefix at end", value: "worldhello", prefix: "hello", wantErr: true},
		{name: "case mismatch", value: "Hello", prefix: "hello", wantErr: true},
		{name: "prefix not found", value: "hello", prefix: "world", wantErr: true},
		{name: "prefix longer than string", value: "hi", prefix: "hello", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", prefix: "test", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), prefix: "test", wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, prefix: "123", wantErr: true},
		{name: "invalid type - bool", value: true, prefix: "true", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := startswithConstraint{prefix: tt.prefix}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEndswithConstraint tests endswithConstraint.Validate() for suffix validation
func TestEndswithConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		suffix  string
		wantErr bool
	}{
		// Valid cases - ends with suffix
		{name: "ends with suffix", value: "helloworld", suffix: "world", wantErr: false},
		{name: "full match", value: "world", suffix: "world", wantErr: false},
		{name: "single char suffix", value: "hello", suffix: "o", wantErr: false},
		{name: "numeric suffix", value: "abc123", suffix: "123", wantErr: false},
		{name: "unicode suffix", value: "au lait café", suffix: "café", wantErr: false},
		{name: "special char suffix", value: "hello @user", suffix: "@user", wantErr: false},

		// Invalid cases - doesn't end with suffix
		{name: "suffix in middle", value: "world hello", suffix: "world", wantErr: true},
		{name: "suffix at start", value: "worldhello", suffix: "world", wantErr: true},
		{name: "case mismatch", value: "helloWorld", suffix: "world", wantErr: true},
		{name: "suffix not found", value: "hello", suffix: "world", wantErr: true},
		{name: "suffix longer than string", value: "hi", suffix: "hello", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", suffix: "test", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), suffix: "test", wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, suffix: "123", wantErr: true},
		{name: "invalid type - bool", value: true, suffix: "true", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := endswithConstraint{suffix: tt.suffix}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLowercaseConstraint tests lowercaseConstraint.Validate() for lowercase validation
func TestLowercaseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - all lowercase
		{name: "all lowercase letters", value: "hello", wantErr: false},
		{name: "lowercase with numbers", value: "hello123", wantErr: false},
		{name: "lowercase with spaces", value: "hello world", wantErr: false},
		{name: "lowercase with special chars", value: "hello@world!", wantErr: false},
		{name: "lowercase with hyphens", value: "hello-world", wantErr: false},
		{name: "numbers only", value: "12345", wantErr: false},
		{name: "special chars only", value: "@#$%", wantErr: false},

		// Invalid cases - contains uppercase
		{name: "single uppercase", value: "Hello", wantErr: true},
		{name: "all uppercase", value: "HELLO", wantErr: true},
		{name: "mixed case", value: "hElLo", wantErr: true},
		{name: "uppercase at end", value: "hellO", wantErr: true},
		{name: "camelCase", value: "helloWorld", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := lowercaseConstraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUppercaseConstraint tests uppercaseConstraint.Validate() for uppercase validation
func TestUppercaseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - all uppercase
		{name: "all uppercase letters", value: "HELLO", wantErr: false},
		{name: "uppercase with numbers", value: "HELLO123", wantErr: false},
		{name: "uppercase with spaces", value: "HELLO WORLD", wantErr: false},
		{name: "uppercase with special chars", value: "HELLO@WORLD!", wantErr: false},
		{name: "uppercase with hyphens", value: "HELLO-WORLD", wantErr: false},
		{name: "numbers only", value: "12345", wantErr: false},
		{name: "special chars only", value: "@#$%", wantErr: false},

		// Invalid cases - contains lowercase
		{name: "single lowercase", value: "HELLo", wantErr: true},
		{name: "all lowercase", value: "hello", wantErr: true},
		{name: "mixed case", value: "HeLLo", wantErr: true},
		{name: "lowercase at start", value: "hELLO", wantErr: true},
		{name: "camelCase", value: "helloWorld", wantErr: true},

		// Edge cases
		{name: "empty string", value: "", wantErr: false},
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := uppercaseConstraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPositiveConstraint tests positiveConstraint.Validate() for positive number validation
func TestPositiveConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - positive numbers
		{name: "positive int", value: 1, wantErr: false},
		{name: "large positive int", value: 1000000, wantErr: false},
		{name: "positive float", value: 0.1, wantErr: false},
		{name: "positive float64", value: float64(3.14), wantErr: false},
		{name: "positive uint", value: uint(5), wantErr: false},
		{name: "positive int8", value: int8(127), wantErr: false},
		{name: "positive int64", value: int64(9999), wantErr: false},

		// Invalid cases - zero and negative
		{name: "zero int", value: 0, wantErr: true},
		{name: "zero float", value: 0.0, wantErr: true},
		{name: "negative int", value: -1, wantErr: true},
		{name: "negative float", value: -0.5, wantErr: true},
		{name: "large negative", value: -1000000, wantErr: true},

		// Edge cases
		{name: "nil pointer", value: (*int)(nil), wantErr: false},
		{name: "pointer to positive", value: intPtr(5), wantErr: false},
		{name: "pointer to zero", value: intPtr(0), wantErr: true},
		{name: "pointer to negative", value: intPtr(-5), wantErr: true},

		// Invalid types
		{name: "invalid type - string", value: "123", wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := positiveConstraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function for int pointers
func intPtr(i int) *int {
	return &i
}

// TestNegativeConstraint tests negativeConstraint.Validate() for negative number validation
func TestNegativeConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - negative numbers
		{name: "negative int", value: -1, wantErr: false},
		{name: "large negative int", value: -1000000, wantErr: false},
		{name: "negative float", value: -0.1, wantErr: false},
		{name: "negative float64", value: float64(-3.14), wantErr: false},
		{name: "negative int8", value: int8(-127), wantErr: false},
		{name: "negative int64", value: int64(-9999), wantErr: false},

		// Invalid cases - zero and positive
		{name: "zero int", value: 0, wantErr: true},
		{name: "zero float", value: 0.0, wantErr: true},
		{name: "positive int", value: 1, wantErr: true},
		{name: "positive float", value: 0.5, wantErr: true},
		{name: "large positive", value: 1000000, wantErr: true},
		{name: "positive uint", value: uint(5), wantErr: true},

		// Edge cases
		{name: "nil pointer", value: (*int)(nil), wantErr: false},
		{name: "pointer to negative", value: intPtr(-5), wantErr: false},
		{name: "pointer to zero", value: intPtr(0), wantErr: true},
		{name: "pointer to positive", value: intPtr(5), wantErr: true},

		// Invalid types
		{name: "invalid type - string", value: "-123", wantErr: true},
		{name: "invalid type - bool", value: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := negativeConstraint{}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultipleOfConstraint(t *testing.T) {
	tests := []struct {
		name    string
		factor  float64
		value   any
		wantErr bool
	}{
		// Valid cases - exact multiples (integer factor)
		{name: "10 is multiple of 5", factor: 5, value: 10, wantErr: false},
		{name: "15 is multiple of 5", factor: 5, value: 15, wantErr: false},
		{name: "0 is multiple of 5", factor: 5, value: 0, wantErr: false},
		{name: "100 is multiple of 10", factor: 10, value: 100, wantErr: false},
		{name: "-10 is multiple of 5", factor: 5, value: -10, wantErr: false},
		{name: "negative multiple of negative", factor: -5, value: -10, wantErr: false},

		// Valid cases - float factor
		{name: "1.5 is multiple of 0.5", factor: 0.5, value: 1.5, wantErr: false},
		{name: "3.0 is multiple of 0.5", factor: 0.5, value: 3.0, wantErr: false},
		{name: "0.25 is multiple of 0.25", factor: 0.25, value: 0.25, wantErr: false},

		// Invalid cases - not exact multiples
		{name: "7 is not multiple of 5", factor: 5, value: 7, wantErr: true},
		{name: "11 is not multiple of 3", factor: 3, value: 11, wantErr: true},
		{name: "1.6 is not multiple of 0.5", factor: 0.5, value: 1.6, wantErr: true},
		{name: "10 is not multiple of 3", factor: 3, value: 10, wantErr: true},

		// Various numeric types
		{name: "int64 multiple", factor: 5, value: int64(25), wantErr: false},
		{name: "int32 multiple", factor: 5, value: int32(20), wantErr: false},
		{name: "uint multiple", factor: 5, value: uint(15), wantErr: false},
		{name: "float32 multiple", factor: 0.5, value: float32(2.5), wantErr: false},
		{name: "float64 multiple", factor: 0.5, value: float64(3.5), wantErr: false},

		// Edge cases
		{name: "nil pointer", factor: 5, value: (*int)(nil), wantErr: false},
		{name: "pointer to multiple", factor: 5, value: intPtr(10), wantErr: false},
		{name: "pointer to non-multiple", factor: 5, value: intPtr(7), wantErr: true},

		// Invalid types
		{name: "invalid type - string", factor: 5, value: "10", wantErr: true},
		{name: "invalid type - bool", factor: 5, value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := multipleOfConstraint{factor: tt.factor}
			err := constraint.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
