// Package pedantigo provides Pydantic-inspired validation for Go.
package pedantigo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// SecretStr masks sensitive string data in JSON output and logs.
// Use SecretStr for passwords, API keys, tokens, and other sensitive data.
// The actual value is preserved internally and accessible via Value().
//
// Example:
//
//	type Config struct {
//	    APIKey SecretStr `json:"api_key" validate:"required"`
//	}
//
//	// JSON output: {"api_key": "**********"}
//	// String() output: "**********"
//	// Value() output: actual API key
type SecretStr struct {
	value string
}

// NewSecretStr creates a new SecretStr from a plain string.
func NewSecretStr(s string) SecretStr {
	return SecretStr{value: s}
}

// Value returns the actual secret string value.
// Use this method to access the underlying secret for processing.
func (s SecretStr) Value() string {
	return s.value
}

// String returns a masked representation (safe for logs).
// Implements fmt.Stringer interface.
func (s SecretStr) String() string {
	return "**********"
}

// MarshalJSON returns a masked value for JSON serialization.
// The actual secret is never exposed in JSON output.
func (s SecretStr) MarshalJSON() ([]byte, error) {
	return json.Marshal("**********")
}

// UnmarshalJSON stores the actual value from JSON input.
// The value is preserved internally for later access via Value().
func (s *SecretStr) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	s.value = v
	return nil
}

// SecretBytes masks sensitive byte data in JSON output and logs.
// Use SecretBytes for binary secrets like encryption keys.
// The JSON input must be base64-encoded.
//
// Example:
//
//	type Config struct {
//	    EncryptionKey SecretBytes `json:"encryption_key" validate:"required"`
//	}
type SecretBytes struct {
	value []byte
}

// NewSecretBytes creates a new SecretBytes from a byte slice.
func NewSecretBytes(b []byte) SecretBytes {
	return SecretBytes{value: b}
}

// Value returns the actual secret bytes.
func (s SecretBytes) Value() []byte {
	return s.value
}

// String returns a masked representation (safe for logs).
func (s SecretBytes) String() string {
	return "**********"
}

// MarshalJSON returns a masked value for JSON serialization.
func (s SecretBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal("**********")
}

// UnmarshalJSON stores the actual value from JSON input.
// Expects base64-encoded string in JSON.
func (s *SecretBytes) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v == "" {
		s.value = []byte{}
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}
	s.value = decoded
	return nil
}
