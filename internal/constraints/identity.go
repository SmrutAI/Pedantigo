// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"regexp"
	"strings"
)

// Identity/publishing constraint types.
type (
	isbnConstraint   struct{} // isbn: validates ISBN-10 or ISBN-13 (ISO 2108)
	isbn10Constraint struct{} // isbn10: validates 10-digit ISBN checksum
	isbn13Constraint struct{} // isbn13: validates 13-digit ISBN (EAN) checksum
	issnConstraint   struct{} // issn: validates 8-digit ISSN checksum (ISO 3297)
	ssnConstraint    struct{} // ssn: validates U.S. SSN format XXX-XX-XXXX
	einConstraint    struct{} // ein: validates U.S. EIN format XX-XXXXXXX
	e164Constraint   struct{} // e164: validates E.164 phone format +[1-9][0-9]{1,14}
)

// Precompiled regex patterns for identity validators.
var (
	// issnRegex matches 8-digit ISSN with optional hyphen after 4th digit, last can be X.
	issnRegex = regexp.MustCompile(`^\d{4}-?\d{3}[\dXx]$`)
	// ssnRegex matches U.S. SSN format XXX-XX-XXXX.
	ssnRegex = regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	// einRegex matches U.S. EIN format XX-XXXXXXX.
	einRegex = regexp.MustCompile(`^\d{2}-\d{7}$`)
	// e164Regex matches E.164 phone format: + followed by 1-15 digits, first digit not 0.
	e164Regex = regexp.MustCompile(`^\+[1-9]\d{0,14}$`)
)

// isbn10Valid validates a 10-digit ISBN checksum.
// ISBN-10 checksum: sum of (digit * position) mod 11 == 0
// Last digit can be 'X' representing 10.
func isbn10Valid(s string) bool {
	cleaned := strings.ReplaceAll(s, "-", "")
	if len(cleaned) != 10 {
		return false
	}
	sum := 0
	for i, r := range cleaned {
		var digit int
		switch {
		case i == 9 && (r == 'X' || r == 'x'):
			digit = 10
		case r >= '0' && r <= '9':
			digit = int(r - '0')
		default:
			return false
		}
		sum += digit * (10 - i)
	}
	return sum%11 == 0
}

// isbn13Valid validates a 13-digit ISBN (EAN) checksum.
// ISBN-13 checksum: alternating weights 1 and 3, sum mod 10 == 0.
func isbn13Valid(s string) bool {
	cleaned := strings.ReplaceAll(s, "-", "")
	if len(cleaned) != 13 {
		return false
	}
	sum := 0
	for i, r := range cleaned {
		if r < '0' || r > '9' {
			return false
		}
		digit := int(r - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	return sum%10 == 0
}

// issnValid validates an 8-digit ISSN checksum.
// ISSN checksum: sum of (digit * (8-position)) mod 11 == 0.
func issnValid(s string) bool {
	cleaned := strings.ReplaceAll(s, "-", "")
	if len(cleaned) != 8 {
		return false
	}
	sum := 0
	for i, r := range cleaned {
		var digit int
		switch {
		case i == 7 && (r == 'X' || r == 'x'):
			digit = 10
		case r >= '0' && r <= '9':
			digit = int(r - '0')
		default:
			return false
		}
		sum += digit * (8 - i)
	}
	return sum%11 == 0
}

// isbnConstraint validates that a string is a valid ISBN-10 or ISBN-13.
func (c isbnConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("isbn constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if isbn10Valid(str) || isbn13Valid(str) {
		return nil
	}
	return NewConstraintError(CodeInvalidISBN, "must be a valid ISBN (10 or 13 digits)")
}

// isbn10Constraint validates that a string is a valid 10-digit ISBN.
func (c isbn10Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("isbn10 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !isbn10Valid(str) {
		return NewConstraintError(CodeInvalidISBN10, "must be a valid ISBN-10")
	}
	return nil
}

// isbn13Constraint validates that a string is a valid 13-digit ISBN (EAN).
func (c isbn13Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("isbn13 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !isbn13Valid(str) {
		return NewConstraintError(CodeInvalidISBN13, "must be a valid ISBN-13")
	}
	return nil
}

// issnConstraint validates that a string is a valid 8-digit ISSN.
func (c issnConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("issn constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// First check format with regex
	if !issnRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidISSN, "must be a valid ISSN")
	}

	// Then validate checksum
	if !issnValid(str) {
		return NewConstraintError(CodeInvalidISSN, "must be a valid ISSN")
	}
	return nil
}

// ssnConstraint validates that a string is a valid U.S. Social Security Number.
func (c ssnConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ssn constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !ssnRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidSSN, "must be a valid SSN (XXX-XX-XXXX)")
	}
	return nil
}

// einConstraint validates that a string is a valid U.S. Employer Identification Number.
func (c einConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ein constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !einRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidEIN, "must be a valid EIN (XX-XXXXXXX)")
	}
	return nil
}

// e164Constraint validates that a string is a valid E.164 international phone number.
func (c e164Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("e164 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !e164Regex.MatchString(str) {
		return NewConstraintError(CodeInvalidE164, "must be a valid E.164 phone number")
	}
	return nil
}
