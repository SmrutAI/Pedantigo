// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Encoding format constraint types.
type (
	jwtConstraint          struct{} // jwt: validates JWT format (3 base64url parts)
	jsonConstraint         struct{} // json: validates JSON string (json.Valid)
	base64Constraint       struct{} // base64: validates base64 encoding (RFC 4648)
	base64urlConstraint    struct{} // base64url: validates base64url encoding (RFC 4648 section 5)
	base64rawurlConstraint struct{} // base64rawurl: validates base64 raw URL encoding (RFC 4648 section 3.2)
)

// Pre-compiled regex for JWT format validation.
var jwtRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)

// Validate checks if the value is a valid JWT (3 base64url parts separated by dots).
func (c jwtConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("jwt constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// JWT must have exactly 3 parts separated by dots
	parts := strings.Split(str, ".")
	if len(parts) != 3 {
		return fmt.Errorf("must be a valid JWT (3 base64url parts)")
	}

	// Each part must be non-empty and match base64url pattern
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("must be a valid JWT (3 base64url parts)")
		}
	}

	// Validate overall format using regex
	if !jwtRegex.MatchString(str) {
		return fmt.Errorf("must be a valid JWT (3 base64url parts)")
	}

	return nil
}

// Validate checks if the value is a valid JSON string.
func (c jsonConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("json constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !json.Valid([]byte(str)) {
		return fmt.Errorf("must be valid JSON")
	}

	return nil
}

// Validate checks if the value is valid base64 encoding (RFC 4648).
func (c base64Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("base64 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	_, decodeErr := base64.StdEncoding.DecodeString(str)
	if decodeErr != nil {
		return fmt.Errorf("must be valid base64 encoding")
	}

	return nil
}

// Validate checks if the value is valid base64url encoding (RFC 4648 section 5).
func (c base64urlConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("base64url constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Check for + or / which are standard base64, not base64url
	if strings.ContainsAny(str, "+/") {
		return fmt.Errorf("must be valid base64url encoding")
	}

	// Try decoding with URL encoding (which allows padding)
	_, decodeErr := base64.URLEncoding.DecodeString(str)
	if decodeErr != nil {
		// Also try without padding
		_, decodeErr = base64.RawURLEncoding.DecodeString(str)
		if decodeErr != nil {
			return fmt.Errorf("must be valid base64url encoding")
		}
	}

	return nil
}

// Validate checks if the value is valid base64 raw URL encoding (RFC 4648 section 3.2).
func (c base64rawurlConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("base64rawurl constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Raw URL encoding must not have padding
	if strings.Contains(str, "=") {
		return fmt.Errorf("must be valid base64 raw URL encoding (no padding)")
	}

	// Check for + or / which are standard base64, not base64url
	if strings.ContainsAny(str, "+/") {
		return fmt.Errorf("must be valid base64 raw URL encoding (no padding)")
	}

	// Try decoding with RawURLEncoding (no padding)
	_, decodeErr := base64.RawURLEncoding.DecodeString(str)
	if decodeErr != nil {
		return fmt.Errorf("must be valid base64 raw URL encoding (no padding)")
	}

	return nil
}
