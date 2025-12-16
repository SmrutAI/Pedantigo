// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Miscellaneous format constraint types.
type (
	htmlConstraint   struct{} // html: validates contains HTML tags
	cronConstraint   struct{} // cron: validates cron expression (5 fields)
	semverConstraint struct{} // semver: validates semantic version X.Y.Z
	ulidConstraint   struct{} // ulid: validates 26 char Crockford base32 ULID
)

// Pre-compiled regex patterns for misc validation.
var (
	// HTML tag detection - matches opening tags with optional attributes.
	htmlRegex = regexp.MustCompile(`<[a-zA-Z!][a-zA-Z0-9]*[^>]*>|<!--[\s\S]*?-->`)

	// Semantic versioning regex (strict adherence to semver.org).
	semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

	// ULID regex - 26 characters from Crockford base32 alphabet (excludes I, L, O, U).
	ulidRegex = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{26}$`)
)

// Validate checks if the value contains HTML tags.
func (c htmlConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("html constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !htmlRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidHTML, "must contain HTML tags")
	}

	return nil
}

// Validate checks if the value is a valid cron expression (5 fields).
func (c cronConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("cron constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Trim and split by whitespace
	fields := strings.Fields(str)
	if len(fields) != 5 {
		return NewConstraintError(CodeInvalidCron, "must be a valid cron expression (5 fields)")
	}

	// Validate each field
	// Field limits: minute (0-59), hour (0-23), day (1-31), month (1-12), weekday (0-7)
	fieldLimits := []struct {
		min, max int
		name     string
	}{
		{0, 59, "minute"},
		{0, 23, "hour"},
		{1, 31, "day"},
		{1, 12, "month"},
		{0, 7, "weekday"},
	}

	for i, field := range fields {
		if !isValidCronField(field, fieldLimits[i].min, fieldLimits[i].max) {
			return NewConstraintError(CodeInvalidCron, "must be a valid cron expression (5 fields)")
		}
	}

	return nil
}

// isValidCronField validates a single cron field against its limits.
func isValidCronField(field string, minVal, maxVal int) bool {
	// Wildcard is always valid
	if field == "*" {
		return true
	}

	// Handle step notation first (*/5 or 1-10/2 or 0-30/5)
	if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) != 2 {
			return false
		}
		// Validate base part (could be * or range or single value)
		base := parts[0]
		if base != "*" && !isValidCronRange(base, minVal, maxVal) {
			return false
		}
		// Validate step value
		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return false
		}
		return true
	}

	// Handle named weekdays (SUN, MON, etc.) and months (JAN, FEB, etc.)
	if isAlpha(field) {
		// Accept common cron names for weekdays
		weekdays := map[string]bool{
			"SUN": true, "MON": true, "TUE": true, "WED": true,
			"THU": true, "FRI": true, "SAT": true,
		}
		return weekdays[strings.ToUpper(field)]
	}

	// Handle range (1-5) or list (1,2,3) or single value
	return isValidCronRange(field, minVal, maxVal)
}

// isValidCronRange validates a cron field that may contain ranges or lists.
func isValidCronRange(field string, minVal, maxVal int) bool {
	// Handle list (1,2,3)
	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, part := range parts {
			if !isValidCronRange(part, minVal, maxVal) {
				return false
			}
		}
		return true
	}

	// Handle range (1-5)
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) != 2 {
			return false
		}
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return false
		}
		return start >= minVal && end <= maxVal && start <= end
	}

	// Single value
	val, err := strconv.Atoi(field)
	if err != nil {
		return false
	}
	return val >= minVal && val <= maxVal
}

// isAlpha returns true if the string contains only alphabetic characters.
func isAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

// Validate checks if the value is a valid semantic version (X.Y.Z).
func (c semverConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("semver constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !semverRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidSemver, "must be a valid semantic version (X.Y.Z)")
	}

	return nil
}

// Validate checks if the value is a valid ULID (26 char Crockford base32).
func (c ulidConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ulid constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !ulidRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidULID, "must be a valid ULID (26 char Crockford base32)")
	}

	return nil
}
