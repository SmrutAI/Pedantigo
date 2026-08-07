// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"time"
)

// Time-related constraint types.
type (
	timezoneConstraint struct{} // timezone: validates IANA timezone names
)

// Validate checks if the value is a valid IANA timezone.
func (c timezoneConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("timezone constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Use time.LoadLocation to validate IANA timezone
	_, loadErr := time.LoadLocation(str)
	if loadErr != nil {
		return NewConstraintError(CodeInvalidTimezone, "must be a valid IANA timezone")
	}

	return nil
}
