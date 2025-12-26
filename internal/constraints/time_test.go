package constraints

import (
	"testing"
)

// TestTimezoneConstraint tests timezoneConstraint.Validate() for valid IANA timezone names.
func TestTimezoneConstraint(t *testing.T) {
	runSimpleConstraintTests(t, timezoneConstraint{}, []simpleTestCase{
		// Valid IANA timezones - major regions
		{"valid UTC", "UTC", false},
		{"valid America/New_York", "America/New_York", false},
		{"valid America/Los_Angeles", "America/Los_Angeles", false},
		{"valid America/Chicago", "America/Chicago", false},
		{"valid America/Denver", "America/Denver", false},
		{"valid Europe/London", "Europe/London", false},
		{"valid Europe/Paris", "Europe/Paris", false},
		{"valid Europe/Berlin", "Europe/Berlin", false},
		{"valid Asia/Tokyo", "Asia/Tokyo", false},
		{"valid Asia/Shanghai", "Asia/Shanghai", false},
		{"valid Asia/Dubai", "Asia/Dubai", false},
		{"valid Asia/Kolkata", "Asia/Kolkata", false},
		{"valid Australia/Sydney", "Australia/Sydney", false},
		{"valid Australia/Melbourne", "Australia/Melbourne", false},
		{"valid Pacific/Auckland", "Pacific/Auckland", false},

		// Valid special timezones
		{"valid Local", "Local", false},
		{"valid GMT", "GMT", false},

		// Valid less common but real timezones
		{"valid America/Argentina/Buenos_Aires", "America/Argentina/Buenos_Aires", false},
		{"valid Africa/Cairo", "Africa/Cairo", false},
		{"valid Antarctica/McMurdo", "Antarctica/McMurdo", false},

		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},

		// Note: Some abbreviations like EST, MST are valid in Go's time package
		// They are defined in the Go timezone database as aliases
		{"valid EST abbreviation", "EST", false}, // Accepted by time.LoadLocation
		{"valid MST abbreviation", "MST", false}, // Accepted by time.LoadLocation

		// Invalid abbreviations (not in Go's database)
		{"invalid PST abbreviation", "PST", true},
		{"invalid CST abbreviation", "CST", true},
		{"invalid EDT abbreviation", "EDT", true},
		{"invalid PDT abbreviation", "PDT", true},
		{"invalid BST abbreviation", "BST", true},
		{"invalid IST abbreviation", "IST", true},

		// Invalid timezones - malformed or non-existent
		{"invalid random string", "foobar", true},
		{"invalid partial timezone", "America", true},
		{"invalid made up zone", "Invalid/Zone", true},
		{"invalid numeric string", "12345", true},
		{"invalid with spaces", "America / New York", true},
		{"invalid mixed separators", "America\\New_York", true},
		{"invalid trailing slash", "America/New_York/", true},
		{"invalid leading slash", "/America/New_York", true},
		// Note: Case sensitivity (lowercase "america/new_york") and double slashes
		// ("America//New_York") have OS-dependent behavior due to filesystem differences.
		// macOS (case-insensitive fs) accepts lowercase; Linux (case-sensitive) rejects it.
		// These cases are intentionally omitted to ensure cross-platform test consistency.
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},

		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 3.14, true},
		{"invalid type - slice", []string{"UTC"}, true},
	})
}
