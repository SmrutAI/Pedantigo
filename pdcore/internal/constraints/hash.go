// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"regexp"
)

// Hash format constraint types.
type (
	md4Constraint     struct{} // md4: validates 32 hex char hash
	md5Constraint     struct{} // md5: validates 32 hex char hash
	sha256Constraint  struct{} // sha256: validates 64 hex char hash
	sha384Constraint  struct{} // sha384: validates 96 hex char hash
	sha512Constraint  struct{} // sha512: validates 128 hex char hash
	mongodbConstraint struct{} // mongodb: validates 24 hex char MongoDB ObjectId
)

// Pre-compiled regex patterns for hash validation.
var (
	md4Regex     = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	md5Regex     = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	sha256Regex  = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	sha384Regex  = regexp.MustCompile(`^[a-fA-F0-9]{96}$`)
	sha512Regex  = regexp.MustCompile(`^[a-fA-F0-9]{128}$`)
	mongodbRegex = regexp.MustCompile(`^[a-fA-F0-9]{24}$`)
)

// Validate checks if the value is a valid MD4 hash (32 hex characters).
func (c md4Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("md4 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !md4Regex.MatchString(str) {
		return NewConstraintError(CodeInvalidMD4, "must be a valid MD4 hash (32 hex characters)")
	}

	return nil
}

// Validate checks if the value is a valid MD5 hash (32 hex characters).
func (c md5Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("md5 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !md5Regex.MatchString(str) {
		return NewConstraintError(CodeInvalidMD5, "must be a valid MD5 hash (32 hex characters)")
	}

	return nil
}

// Validate checks if the value is a valid SHA256 hash (64 hex characters).
func (c sha256Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("sha256 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !sha256Regex.MatchString(str) {
		return NewConstraintError(CodeInvalidSHA256, "must be a valid SHA256 hash (64 hex characters)")
	}

	return nil
}

// Validate checks if the value is a valid SHA384 hash (96 hex characters).
func (c sha384Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("sha384 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !sha384Regex.MatchString(str) {
		return NewConstraintError(CodeInvalidSHA384, "must be a valid SHA384 hash (96 hex characters)")
	}

	return nil
}

// Validate checks if the value is a valid SHA512 hash (128 hex characters).
func (c sha512Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("sha512 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !sha512Regex.MatchString(str) {
		return NewConstraintError(CodeInvalidSHA512, "must be a valid SHA512 hash (128 hex characters)")
	}

	return nil
}

// Validate checks if the value is a valid MongoDB ObjectId (24 hex characters).
func (c mongodbConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("mongodb constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !mongodbRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidMongoDB, "must be a valid MongoDB ObjectId (24 hex characters)")
	}

	return nil
}
