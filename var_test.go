package pedantigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================================================
// Basic Constraint Tests
// ==================================================

// TestVar_Email tests email constraint validation.
func TestVar_Email(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid emails
		{"valid email simple", "test@example.com", "email", false},
		{"valid email subdomain", "user@mail.example.com", "email", false},
		{"valid email with plus", "user+tag@example.com", "email", false},
		{"valid email with dots", "first.last@example.com", "email", false},
		{"valid email with numbers", "user123@example456.com", "email", false},

		// Invalid emails
		{"invalid email no at", "invalid", "email", true},
		{"invalid email no domain", "user@", "email", true},
		{"invalid email no user", "@example.com", "email", true},
		{"invalid email spaces", "user @example.com", "email", true},
		{"invalid email double at", "user@@example.com", "email", true},

		// Empty string (optional by default)
		{"empty string optional", "", "email", false},

		// Nil value (optional by default)
		{"nil pointer optional", (*string)(nil), "email", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Required tests required constraint validation.
func TestVar_Required(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid required values
		{"required string present", "hello", "required", false},
		{"required int non-zero", 42, "required", false},
		{"required int zero is valid", 0, "required", false}, // Zero is valid, just not empty string
		{"required bool true", true, "required", false},
		{"required bool false", false, "required", false},

		// Invalid required values
		{"required string empty", "", "required", true},
		{"required nil", nil, "required", true},
		{"required nil pointer", (*string)(nil), "required", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// Numeric Constraint Tests
// ==================================================

// TestVar_Min tests minimum value constraint.
func TestVar_Min(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid min values
		{"min int valid", 25, "min=18", false},
		{"min int equal", 18, "min=18", false},
		{"min int large", 100, "min=18", false},
		{"min float valid", 25.5, "min=18", false},
		{"min int8", int8(25), "min=18", false},
		{"min int16", int16(25), "min=18", false},
		{"min int32", int32(25), "min=18", false},
		{"min int64", int64(25), "min=18", false},
		{"min uint", uint(25), "min=18", false},
		{"min float32", float32(25.5), "min=18", false},
		{"min float64", float64(25.5), "min=18", false},

		// Invalid min values
		{"min int invalid", 15, "min=18", true},
		{"min int negative", -5, "min=18", true},
		{"min float invalid", 10.5, "min=18", true},

		// String length (min works on string length too)
		{"min string length valid", "hello", "min=3", false},
		{"min string length equal", "hi", "min=2", false},
		{"min string length invalid", "a", "min=5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Max tests maximum value constraint.
func TestVar_Max(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid max values
		{"max int valid", 100, "max=120", false},
		{"max int equal", 120, "max=120", false},
		{"max int small", 50, "max=120", false},
		{"max float valid", 100.5, "max=120", false},

		// Invalid max values
		{"max int invalid", 150, "max=120", true},
		{"max int large", 200, "max=120", true},
		{"max float invalid", 130.5, "max=120", true},

		// String length (max works on string length too)
		{"max string length valid", "hello", "max=10", false},
		{"max string length equal", "hello", "max=5", false},
		{"max string length invalid", "toolongstring", "max=5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Range tests min and max together.
func TestVar_Range(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"range valid middle", 50, "min=18,max=120", false},
		{"range valid min boundary", 18, "min=18,max=120", false},
		{"range valid max boundary", 120, "min=18,max=120", false},
		{"range invalid below min", 10, "min=18,max=120", true},
		{"range invalid above max", 150, "min=18,max=120", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// String Constraint Tests
// ==================================================

// TestVar_UUID tests UUID constraint validation.
func TestVar_UUID(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid UUIDs
		{"uuid v4 valid", "550e8400-e29b-41d4-a716-446655440000", "uuid", false},
		{"uuid v1 valid", "c9bf9e58-0000-1000-8000-00805f9b34fb", "uuid", false},
		{"uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", "uuid", false},
		{"uuid lowercase", "550e8400-e29b-41d4-a716-446655440000", "uuid", false},

		// Invalid UUIDs
		{"uuid invalid format", "not-a-uuid", "uuid", true},
		{"uuid missing dashes", "550e8400e29b41d4a716446655440000", "uuid", true},
		{"uuid too short", "550e8400-e29b-41d4", "uuid", true},
		{"uuid with extra chars", "550e8400-e29b-41d4-a716-446655440000x", "uuid", true},
		{"uuid empty string", "", "uuid", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_URL tests URL constraint validation.
func TestVar_URL(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid URLs
		{"url https", "https://example.com", "url", false},
		{"url http", "http://example.com", "url", false},
		{"url with path", "https://example.com/path/to/resource", "url", false},
		{"url with query", "https://example.com?key=value", "url", false},
		{"url with fragment", "https://example.com#section", "url", false},
		{"url with port", "https://example.com:8080", "url", false},
		{"url localhost", "http://localhost:3000", "url", false},

		// Invalid URLs
		{"url invalid no scheme", "not-a-url", "url", true},
		{"url invalid no domain", "https://", "url", true},
		{"url invalid spaces", "https://exa mple.com", "url", true},
		{"url empty string", "", "url", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Oneof tests oneof constraint validation.
func TestVar_Oneof(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid oneof values
		{"oneof valid admin", "admin", "oneof=admin user guest", false},
		{"oneof valid user", "user", "oneof=admin user guest", false},
		{"oneof valid guest", "guest", "oneof=admin user guest", false},

		// Invalid oneof values
		{"oneof invalid superuser", "superuser", "oneof=admin user guest", true},
		// Note: oneof validates even empty strings
		{"oneof invalid empty", "", "oneof=admin user guest", true}, // Empty not in list
		{"oneof invalid case", "Admin", "oneof=admin user guest", true},

		// Numeric oneof
		{"oneof int valid", 1, "oneof=1 2 3", false},
		{"oneof int invalid", 5, "oneof=1 2 3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Len tests exact length constraint.
func TestVar_Len(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Valid len values
		{"len string valid", "hello", "len=5", false},
		{"len string equal", "hi", "len=2", false},

		// Invalid len values
		{"len string invalid short", "hi", "len=5", true},
		{"len string invalid long", "toolong", "len=5", true},

		// Note: len constraint in Pedantigo only works on strings, not slices
		// For slice length, use minlen/maxlen or a custom validator
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Alpha tests alphabetic characters constraint.
func TestVar_Alpha(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"alpha valid lowercase", "hello", "alpha", false},
		{"alpha valid uppercase", "HELLO", "alpha", false},
		{"alpha valid mixed", "HelloWorld", "alpha", false},
		{"alpha invalid with numbers", "hello123", "alpha", true},
		{"alpha invalid with spaces", "hello world", "alpha", true},
		{"alpha invalid with special chars", "hello!", "alpha", true},
		{"alpha empty string", "", "alpha", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Alphanum tests alphanumeric characters constraint.
func TestVar_Alphanum(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"alphanum valid letters", "hello", "alphanum", false},
		{"alphanum valid numbers", "12345", "alphanum", false},
		{"alphanum valid mixed", "hello123", "alphanum", false},
		{"alphanum invalid spaces", "hello world", "alphanum", true},
		{"alphanum invalid special", "hello!", "alphanum", true},
		{"alphanum empty string", "", "alphanum", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Regexp tests regular expression pattern constraint.
func TestVar_Regexp(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"regexp valid uppercase code", "ABC", "regexp=^[A-Z]{3}$", false},
		{"regexp valid date format", "2023-12-25", `regexp=^\d{4}-\d{2}-\d{2}$`, false},
		{"regexp invalid lowercase", "abc", "regexp=^[A-Z]{3}$", true},
		{"regexp invalid length", "AB", "regexp=^[A-Z]{3}$", true},
		{"regexp invalid format", "12-25-2023", `regexp=^\d{4}-\d{2}-\d{2}$`, true},
		{"regexp empty string", "", "regexp=^[A-Z]{3}$", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// String Pattern Tests
// ==================================================

// TestVar_Contains tests substring contains constraint.
func TestVar_Contains(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"contains valid", "hello world", "contains=world", false},
		{"contains valid at start", "hello world", "contains=hello", false},
		{"contains valid at end", "hello world", "contains=world", false},
		{"contains invalid", "hello world", "contains=foo", true},
		{"contains case sensitive", "hello world", "contains=WORLD", true},
		// Note: contains checks even empty strings
		{"contains empty string", "", "contains=test", true}, // Empty string doesn't contain "test"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Startswith tests string prefix constraint.
func TestVar_Startswith(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"startswith valid", "hello world", "startswith=hello", false},
		{"startswith invalid", "hello world", "startswith=world", true},
		{"startswith case sensitive", "hello world", "startswith=Hello", true},
		{"startswith empty string", "", "startswith=test", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Endswith tests string suffix constraint.
func TestVar_Endswith(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"endswith valid", "hello world", "endswith=world", false},
		{"endswith invalid", "hello world", "endswith=hello", true},
		{"endswith case sensitive", "hello world", "endswith=World", true},
		{"endswith empty string", "", "endswith=test", false}, // Empty is optional
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// Multiple Constraints Tests
// ==================================================

// TestVar_MultipleConstraints tests combining multiple constraints.
func TestVar_MultipleConstraints(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Both constraints pass
		{"multi pass all", "test@example.com", "required,email", false},
		{"multi pass min max", 50, "min=18,max=120", false},

		// First constraint fails
		{"multi fail required", "", "required,email", true},
		{"multi fail min", 10, "min=18,max=120", true},

		// Second constraint fails
		{"multi fail email", "invalid", "required,email", true},
		{"multi fail max", 150, "min=18,max=120", true},

		// Complex combination
		{"multi complex valid", "admin", "required,oneof=admin user guest", false},
		{"multi complex invalid", "superuser", "required,oneof=admin user guest", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// Edge Cases and Special Tests
// ==================================================

// TestVar_EmptyTag tests that empty tag passes validation.
func TestVar_EmptyTag(t *testing.T) {
	err := Var("anything", "")
	require.NoError(t, err, "Empty tag should pass validation")

	err = Var(123, "")
	require.NoError(t, err, "Empty tag should pass validation for any type")

	err = Var(nil, "")
	assert.NoError(t, err, "Empty tag should pass even for nil")
}

// TestVar_NilValues tests nil value handling.
func TestVar_NilValues(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Nil with required should fail
		{"nil required fail", nil, "required", true},

		// Nil with optional constraints should pass
		{"nil email pass", nil, "email", false},
		{"nil url pass", nil, "url", false},
		{"nil uuid pass", nil, "uuid", false},

		// Nil pointer with required should fail
		{"nil pointer required fail", (*string)(nil), "required", true},

		// Nil pointer with optional should pass
		{"nil pointer email pass", (*string)(nil), "email", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_IntTypes tests various integer types.
func TestVar_IntTypes(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"int", int(25), "min=18", false},
		{"int8", int8(25), "min=18", false},
		{"int16", int16(25), "min=18", false},
		{"int32", int32(25), "min=18", false},
		{"int64", int64(25), "min=18", false},
		{"uint", uint(25), "min=18", false},
		{"uint8", uint8(25), "min=18", false},
		{"uint16", uint16(25), "min=18", false},
		{"uint32", uint32(25), "min=18", false},
		{"uint64", uint64(25), "min=18", false},
		{"float32", float32(25.5), "min=18", false},
		{"float64", float64(25.5), "min=18", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_InvalidType tests that invalid types fail appropriately.
func TestVar_InvalidType(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		// Email expects string
		{"email with int", 123, "email", true},
		{"email with bool", true, "email", true},

		// URL expects string
		{"url with int", 123, "url", true},

		// UUID expects string
		{"uuid with int", 123, "uuid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// Additional Format Constraints
// ==================================================

// TestVar_Numeric tests numeric-only constraint.
// Note: Pedantigo's numeric constraint allows signed decimals (e.g., -123.45)
// Use regexp=^\d+$ if you need digits-only validation.
func TestVar_Numeric(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"numeric valid digits", "12345", "numeric", false},
		{"numeric invalid letters", "abc123", "numeric", true},
		{"numeric valid decimal", "123.45", "numeric", false}, // Pedantigo allows decimals
		{"numeric valid negative", "-123", "numeric", false},
		{"numeric valid signed decimal", "+123.45", "numeric", false},
		{"numeric invalid non-numeric", "12.34.56", "numeric", true},
		{"numeric empty", "", "numeric", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Lowercase tests lowercase constraint.
func TestVar_Lowercase(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"lowercase valid", "hello", "lowercase", false},
		{"lowercase invalid", "Hello", "lowercase", true},
		{"lowercase invalid all caps", "HELLO", "lowercase", true},
		{"lowercase empty", "", "lowercase", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// TestVar_Uppercase tests uppercase constraint.
func TestVar_Uppercase(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"uppercase valid", "HELLO", "uppercase", false},
		{"uppercase invalid", "Hello", "uppercase", true},
		{"uppercase invalid all lower", "hello", "uppercase", true},
		{"uppercase empty", "", "uppercase", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var(%v, %q) error = %v, wantErr %v", tt.value, tt.tag, err, tt.wantErr)
			}
		})
	}
}

// ==================================================
// Benchmark Tests
// ==================================================

// BenchmarkVar_Email benchmarks email validation.
func BenchmarkVar_Email(b *testing.B) {
	value := "test@example.com"
	tag := "email"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Var(value, tag)
	}
}

// BenchmarkVar_MultipleConstraints benchmarks multiple constraint validation.
func BenchmarkVar_MultipleConstraints(b *testing.B) {
	value := "test@example.com"
	tag := "required,email,len=16"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Var(value, tag)
	}
}

// BenchmarkVar_NumericRange benchmarks numeric range validation.
func BenchmarkVar_NumericRange(b *testing.B) {
	value := 50
	tag := "min=18,max=120"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Var(value, tag)
	}
}
