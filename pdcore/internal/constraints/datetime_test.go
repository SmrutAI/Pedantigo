package constraints

import (
	"testing"
)

// TestDatetimeConstraint tests datetimeConstraint.Validate() for datetime validation.
func TestDatetimeConstraint(t *testing.T) {
	tests := []struct {
		name    string
		layout  string
		value   any
		wantErr bool
	}{
		// Test 1: Valid date format (YYYY-MM-DD)
		{name: "valid date YYYY-MM-DD", layout: "2006-01-02", value: "2024-12-21", wantErr: false},

		// Test 2: Invalid date format (wrong order)
		{name: "invalid date format", layout: "2006-01-02", value: "21-12-2024", wantErr: true},

		// Test 3: Invalid date values (month 13)
		{name: "invalid date values", layout: "2006-01-02", value: "2024-13-45", wantErr: true},

		// Test 4: Valid RFC3339
		{name: "valid RFC3339", layout: "2006-01-02T15:04:05Z07:00", value: "2024-12-21T10:30:00Z", wantErr: false},

		// Test 5: Valid with timezone offset
		{name: "valid with timezone", layout: "2006-01-02T15:04:05Z07:00", value: "2024-12-21T10:30:00+05:30", wantErr: false},

		// Test 6: Time only
		{name: "time only", layout: "15:04:05", value: "10:30:00", wantErr: false},

		// Test 7: Empty string passes (not required)
		{name: "empty string passes", layout: "2006-01-02", value: "", wantErr: false},

		// Test 8: Nil pointer passes
		{name: "nil pointer passes", layout: "2006-01-02", value: (*string)(nil), wantErr: false},

		// Test 9: Custom layout (Jan 2, 2006)
		{name: "custom layout", layout: "Jan 2, 2006", value: "Dec 21, 2024", wantErr: false},

		// Test 10: Leap year valid
		{name: "leap year valid", layout: "2006-01-02", value: "2024-02-29", wantErr: false},

		// Test 11: Non-leap year Feb 29 invalid
		{name: "non-leap year Feb 29 invalid", layout: "2006-01-02", value: "2023-02-29", wantErr: true},

		// Additional test cases for comprehensive coverage

		// Test 12: Invalid type - int
		{name: "invalid type - int", layout: "2006-01-02", value: 123, wantErr: true},

		// Test 13: Invalid type - bool
		{name: "invalid type - bool", layout: "2006-01-02", value: true, wantErr: true},

		// Test 14: Wrong format but valid date
		{name: "wrong format but valid date", layout: "2006-01-02", value: "2024/12/21", wantErr: true},

		// Test 15: Valid datetime with seconds
		{name: "valid datetime with seconds", layout: "2006-01-02 15:04:05", value: "2024-12-21 14:30:45", wantErr: false},

		// Test 16: Invalid time - hour 25
		{name: "invalid time - hour 25", layout: "15:04:05", value: "25:30:00", wantErr: true},

		// Test 17: Invalid time - minute 60
		{name: "invalid time - minute 60", layout: "15:04:05", value: "10:60:00", wantErr: true},

		// Test 18: Valid 12-hour format with AM/PM
		{name: "valid 12-hour AM", layout: "3:04 PM", value: "2:30 AM", wantErr: false},
		{name: "valid 12-hour PM", layout: "3:04 PM", value: "2:30 PM", wantErr: false},

		// Test 19: Month name format
		{name: "month name format", layout: "January 2, 2006", value: "December 21, 2024", wantErr: false},

		// Test 20: Short month name format
		{name: "short month name format", layout: "Jan 02 2006", value: "Dec 21 2024", wantErr: false},

		// Test 21: Invalid month name
		{name: "invalid month name", layout: "January 2, 2006", value: "InvalidMonth 21, 2024", wantErr: true},

		// Test 22: Valid date with day of week
		{name: "valid with day of week", layout: "Monday, January 2, 2006", value: "Saturday, December 21, 2024", wantErr: false},

		// Test 23: Partial string (too short)
		{name: "partial string too short", layout: "2006-01-02", value: "2024-12", wantErr: true},

		// Test 24: Extra characters after valid date
		{name: "extra characters", layout: "2006-01-02", value: "2024-12-21extra", wantErr: true},

		// Test 25: Valid ISO 8601 with microseconds
		{name: "ISO 8601 with microseconds", layout: "2006-01-02T15:04:05.000000Z07:00", value: "2024-12-21T10:30:00.123456Z", wantErr: false},

		// Test 26: Unix timestamp layout (unusual but valid)
		{name: "unix date format", layout: "Mon Jan _2 15:04:05 MST 2006", value: "Sat Dec 21 10:30:00 UTC 2024", wantErr: false},

		// Test 27: Valid but edge case - Jan 1, year 1
		{name: "year 1 date", layout: "2006-01-02", value: "0001-01-01", wantErr: false},

		// Test 28: Invalid day for month (April 31)
		{name: "invalid day for month", layout: "2006-01-02", value: "2024-04-31", wantErr: true},

		// Test 29: Valid end of month (February 28 non-leap)
		{name: "feb 28 non-leap", layout: "2006-01-02", value: "2023-02-28", wantErr: false},

		// Test 30: Zero-padded formats
		{name: "zero-padded day", layout: "2006-01-02", value: "2024-01-01", wantErr: false},
		{name: "zero-padded month", layout: "2006-01-02", value: "2024-01-15", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := datetimeConstraint{layout: tt.layout}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestBuildDatetimeConstraint tests buildDatetimeConstraint builder function.
func TestBuildDatetimeConstraint(t *testing.T) {
	tests := []struct {
		name      string
		layout    string
		wantOk    bool
		testValue string
		wantErr   bool
	}{
		// Valid layouts
		{name: "valid YYYY-MM-DD layout", layout: "2006-01-02", wantOk: true, testValue: "2024-12-21", wantErr: false},
		{name: "valid RFC3339 layout", layout: "2006-01-02T15:04:05Z07:00", wantOk: true, testValue: "2024-12-21T10:30:00Z", wantErr: false},
		{name: "valid time layout", layout: "15:04:05", wantOk: true, testValue: "14:30:45", wantErr: false},
		{name: "valid custom layout", layout: "Jan 2, 2006", wantOk: true, testValue: "Dec 21, 2024", wantErr: false},

		// Invalid value should fail validation
		{name: "layout mismatch", layout: "2006-01-02", wantOk: true, testValue: "21-12-2024", wantErr: true},

		// Empty layout should return false
		{name: "empty layout", layout: "", wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, ok := buildDatetimeConstraint(tt.layout)

			if !tt.wantOk {
				if ok {
					t.Errorf("buildDatetimeConstraint(%q) returned ok=true, want ok=false", tt.layout)
				}
				return
			}

			if !ok {
				t.Errorf("buildDatetimeConstraint(%q) returned ok=false, want ok=true", tt.layout)
				return
			}

			if constraint == nil {
				t.Errorf("buildDatetimeConstraint(%q) returned nil constraint", tt.layout)
				return
			}

			// Test validation if we have a test value
			if tt.testValue != "" {
				err := constraint.Validate(tt.testValue)
				if tt.wantErr {
					if err == nil {
						t.Errorf("Validate(%q) with layout %q: expected error, got nil", tt.testValue, tt.layout)
					}
				} else {
					if err != nil {
						t.Errorf("Validate(%q) with layout %q: expected no error, got %v", tt.testValue, tt.layout, err)
					}
				}
			}
		})
	}
}
