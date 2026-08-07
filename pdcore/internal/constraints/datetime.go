package constraints

import (
	"fmt"
	"time"
)

// datetimeConstraint validates strings against a Go time layout format.
type datetimeConstraint struct {
	layout string // Go time layout (e.g., "2006-01-02")
}

// Validate checks if the value is a valid datetime string matching the layout.
func (c datetimeConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // Skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("datetime constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings handled by required constraint
	}

	_, err = time.Parse(c.layout, str)
	if err != nil {
		return NewConstraintErrorf(CodeInvalidDatetime,
			"must be a valid datetime matching layout %q", c.layout)
	}
	return nil
}

// buildDatetimeConstraint creates a datetime constraint from the layout value.
func buildDatetimeConstraint(layout string) (Constraint, bool) {
	if layout == "" {
		return nil, false
	}
	return datetimeConstraint{layout: layout}, true
}
