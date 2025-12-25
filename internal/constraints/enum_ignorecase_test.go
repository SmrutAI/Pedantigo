package constraints

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEqIgnoreCaseConstraint tests eqIgnoreCaseConstraint.Validate() for case-insensitive equality.
func TestEqIgnoreCaseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		param   string
		wantErr bool
	}{
		// Exact match cases
		{name: "exact match lowercase", value: "hello", param: "hello", wantErr: false},
		{name: "exact match uppercase", value: "HELLO", param: "HELLO", wantErr: false},
		{name: "exact match mixed case", value: "HeLLo", param: "HeLLo", wantErr: false},

		// Case-insensitive match cases
		{name: "value lowercase param uppercase", value: "hello", param: "HELLO", wantErr: false},
		{name: "value uppercase param lowercase", value: "HELLO", param: "hello", wantErr: false},
		{name: "mixed case match 1", value: "HeLLo", param: "hello", wantErr: false},
		{name: "mixed case match 2", value: "hello", param: "HeLLo", wantErr: false},
		{name: "mixed case match 3", value: "WoRlD", param: "wOrLd", wantErr: false},

		// Mismatch cases
		{name: "different strings lowercase", value: "hello", param: "world", wantErr: true},
		{name: "different strings mixed case", value: "Hello", param: "World", wantErr: true},
		{name: "different length", value: "hello", param: "hi", wantErr: true},
		{name: "similar but different", value: "hello", param: "hallo", wantErr: true},

		// Empty string cases
		{name: "both empty", value: "", param: "", wantErr: false},
		{name: "value empty param not", value: "", param: "test", wantErr: false}, // empty strings skip validation
		{name: "value not empty param empty", value: "test", param: "", wantErr: true},

		// Nil pointer cases - should skip validation
		{name: "nil string pointer", value: (*string)(nil), param: "test", wantErr: false},
		{name: "nil int pointer", value: (*int)(nil), param: "42", wantErr: false},

		// Numeric string comparisons (case-insensitive string comparison)
		{name: "numeric strings equal", value: "123", param: "123", wantErr: false},
		{name: "numeric strings not equal", value: "123", param: "456", wantErr: true},

		// Whitespace cases
		{name: "whitespace preserved not equal", value: "hello world", param: "helloworld", wantErr: true},
		{name: "whitespace preserved equal", value: "hello world", param: "HELLO WORLD", wantErr: false},
		{name: "leading space not equal", value: " hello", param: "hello", wantErr: true},
		{name: "trailing space not equal", value: "hello ", param: "hello", wantErr: true},

		// Special characters
		{name: "special chars equal", value: "test@example.com", param: "TEST@EXAMPLE.COM", wantErr: false},
		{name: "special chars with numbers", value: "User123", param: "user123", wantErr: false},
		{name: "underscores", value: "hello_world", param: "HELLO_WORLD", wantErr: false},
		{name: "hyphens", value: "test-case", param: "TEST-CASE", wantErr: false},

		// Unicode cases
		{name: "unicode basic", value: "café", param: "CAFÉ", wantErr: false},
		{name: "unicode cyrillic", value: "привет", param: "ПРИВЕТ", wantErr: false},
		{name: "unicode greek", value: "γεια", param: "ΓΕΙΑ", wantErr: false},

		// Integer values (converted to string for comparison)
		{name: "int equal", value: 42, param: "42", wantErr: false},
		{name: "int not equal", value: 42, param: "43", wantErr: true},
		{name: "int zero", value: 0, param: "0", wantErr: false},
		{name: "int negative", value: -10, param: "-10", wantErr: false},

		// Integer type variations
		{name: "int8 equal", value: int8(42), param: "42", wantErr: false},
		{name: "int16 equal", value: int16(42), param: "42", wantErr: false},
		{name: "int32 equal", value: int32(42), param: "42", wantErr: false},
		{name: "int64 equal", value: int64(42), param: "42", wantErr: false},

		// Unsigned integer values
		{name: "uint equal", value: uint(42), param: "42", wantErr: false},
		{name: "uint not equal", value: uint(42), param: "43", wantErr: true},
		{name: "uint8 equal", value: uint8(42), param: "42", wantErr: false},
		{name: "uint16 equal", value: uint16(42), param: "42", wantErr: false},
		{name: "uint32 equal", value: uint32(42), param: "42", wantErr: false},
		{name: "uint64 equal", value: uint64(42), param: "42", wantErr: false},

		// Float values
		{name: "float equal", value: 3.14, param: "3.14", wantErr: false},
		{name: "float not equal", value: 3.14, param: "2.71", wantErr: true},
		{name: "float32 equal", value: float32(2.5), param: "2.5", wantErr: false},
		{name: "float64 equal", value: float64(3.14), param: "3.14", wantErr: false},

		// Boolean values
		{name: "bool true lowercase", value: true, param: "true", wantErr: false},
		{name: "bool true uppercase", value: true, param: "TRUE", wantErr: false},
		{name: "bool true mixed case", value: true, param: "True", wantErr: false},
		{name: "bool false lowercase", value: false, param: "false", wantErr: false},
		{name: "bool false uppercase", value: false, param: "FALSE", wantErr: false},
		{name: "bool mismatch", value: true, param: "false", wantErr: true},

		// Invalid types - should return error
		{name: "slice type", value: []string{"hello"}, param: "hello", wantErr: true},
		{name: "map type", value: map[string]int{"key": 1}, param: "key", wantErr: true},
		{name: "struct type", value: struct{ Name string }{"test"}, param: "test", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := eqIgnoreCaseConstraint{value: tt.param}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestNeIgnoreCaseConstraint tests neIgnoreCaseConstraint.Validate() for case-insensitive inequality.
func TestNeIgnoreCaseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		param   string
		wantErr bool
	}{
		// Different strings - should pass
		{name: "different strings lowercase", value: "hello", param: "world", wantErr: false},
		{name: "different strings mixed case", value: "Hello", param: "World", wantErr: false},
		{name: "different length", value: "hello", param: "hi", wantErr: false},
		{name: "similar but different", value: "hello", param: "hallo", wantErr: false},

		// Case-insensitive same strings - should fail
		{name: "exact match lowercase", value: "hello", param: "hello", wantErr: true},
		{name: "exact match uppercase", value: "HELLO", param: "HELLO", wantErr: true},
		{name: "value lowercase param uppercase", value: "hello", param: "HELLO", wantErr: true},
		{name: "value uppercase param lowercase", value: "HELLO", param: "hello", wantErr: true},
		{name: "mixed case match 1", value: "HeLLo", param: "hello", wantErr: true},
		{name: "mixed case match 2", value: "hello", param: "HeLLo", wantErr: true},
		{name: "mixed case match 3", value: "WoRlD", param: "wOrLd", wantErr: true},

		// Empty string cases - empty strings skip validation (handled by required constraint)
		{name: "both empty", value: "", param: "", wantErr: false},                      // empty skips validation
		{name: "value empty param not", value: "", param: "test", wantErr: false},       // empty strings skip validation
		{name: "value not empty param empty", value: "test", param: "", wantErr: false}, // comparing to empty param passes

		// Nil pointer cases - should skip validation
		{name: "nil string pointer", value: (*string)(nil), param: "test", wantErr: false},
		{name: "nil int pointer", value: (*int)(nil), param: "42", wantErr: false},

		// Numeric string comparisons
		{name: "numeric strings different", value: "123", param: "456", wantErr: false},
		{name: "numeric strings same", value: "123", param: "123", wantErr: true},

		// Whitespace cases
		{name: "whitespace makes different", value: "hello world", param: "helloworld", wantErr: false},
		{name: "whitespace preserved same", value: "hello world", param: "HELLO WORLD", wantErr: true},
		{name: "leading space different", value: " hello", param: "hello", wantErr: false},
		{name: "trailing space different", value: "hello ", param: "hello", wantErr: false},

		// Special characters
		{name: "special chars same", value: "test@example.com", param: "TEST@EXAMPLE.COM", wantErr: true},
		{name: "special chars different", value: "test@example.com", param: "other@example.com", wantErr: false},
		{name: "underscores same", value: "hello_world", param: "HELLO_WORLD", wantErr: true},
		{name: "underscores different", value: "hello_world", param: "goodbye_world", wantErr: false},

		// Unicode cases
		{name: "unicode same", value: "café", param: "CAFÉ", wantErr: true},
		{name: "unicode different", value: "café", param: "tea", wantErr: false},
		{name: "unicode cyrillic same", value: "привет", param: "ПРИВЕТ", wantErr: true},
		{name: "unicode greek different", value: "γεια", param: "καλημέρα", wantErr: false},

		// Integer values
		{name: "int different", value: 42, param: "43", wantErr: false},
		{name: "int same", value: 42, param: "42", wantErr: true},
		{name: "int zero different", value: 0, param: "1", wantErr: false},
		{name: "int negative same", value: -10, param: "-10", wantErr: true},

		// Integer type variations
		{name: "int8 same", value: int8(42), param: "42", wantErr: true},
		{name: "int16 different", value: int16(42), param: "43", wantErr: false},
		{name: "int32 same", value: int32(42), param: "42", wantErr: true},
		{name: "int64 different", value: int64(42), param: "43", wantErr: false},

		// Unsigned integer values
		{name: "uint different", value: uint(42), param: "43", wantErr: false},
		{name: "uint same", value: uint(42), param: "42", wantErr: true},
		{name: "uint8 different", value: uint8(42), param: "43", wantErr: false},
		{name: "uint16 same", value: uint16(42), param: "42", wantErr: true},

		// Float values
		{name: "float different", value: 3.14, param: "2.71", wantErr: false},
		{name: "float same", value: 3.14, param: "3.14", wantErr: true},
		{name: "float32 different", value: float32(2.5), param: "3.5", wantErr: false},
		{name: "float64 same", value: float64(3.14), param: "3.14", wantErr: true},

		// Boolean values
		{name: "bool same true lowercase", value: true, param: "true", wantErr: true},
		{name: "bool same true uppercase", value: true, param: "TRUE", wantErr: true},
		{name: "bool same false mixed", value: false, param: "False", wantErr: true},
		{name: "bool different", value: true, param: "false", wantErr: false},

		// Invalid types - should return error
		{name: "slice type", value: []string{"hello"}, param: "hello", wantErr: true},
		{name: "map type", value: map[string]int{"key": 1}, param: "key", wantErr: true},
		{name: "struct type", value: struct{ Name string }{"test"}, param: "test", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := neIgnoreCaseConstraint{value: tt.param}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestBuildEqIgnoreCaseConstraint tests buildEqIgnoreCaseConstraint builder function.
func TestBuildEqIgnoreCaseConstraint(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		testValue any
		wantOk    bool
		wantErr   bool
	}{
		// Valid parameter cases
		{name: "simple lowercase", param: "hello", testValue: "HELLO", wantOk: true, wantErr: false},
		{name: "simple uppercase", param: "WORLD", testValue: "world", wantOk: true, wantErr: false},
		{name: "mixed case", param: "HeLLo", testValue: "hello", wantOk: true, wantErr: false},
		{name: "with spaces", param: "hello world", testValue: "HELLO WORLD", wantOk: true, wantErr: false},
		{name: "with special chars", param: "test@example.com", testValue: "TEST@EXAMPLE.COM", wantOk: true, wantErr: false},
		{name: "numeric string", param: "12345", testValue: "12345", wantOk: true, wantErr: false},

		// Empty parameter - should return false for ok
		{name: "empty param", param: "", testValue: "test", wantOk: false, wantErr: false},

		// Test mismatches
		{name: "mismatch", param: "hello", testValue: "world", wantOk: true, wantErr: true},
		{name: "case insensitive mismatch", param: "hello", testValue: "WORLD", wantOk: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, ok := buildEqIgnoreCaseConstraint(tt.param)

			if tt.wantOk {
				require.True(t, ok, "buildEqIgnoreCaseConstraint should return ok=true")
				require.NotNil(t, constraint, "buildEqIgnoreCaseConstraint should return non-nil constraint")

				err := constraint.Validate(tt.testValue)
				if tt.wantErr {
					require.Error(t, err, "Validate should return error for mismatch")
				} else {
					require.NoError(t, err, "Validate should return no error for match")
				}
			} else {
				require.False(t, ok, "buildEqIgnoreCaseConstraint should return ok=false for empty param")
				require.Nil(t, constraint, "buildEqIgnoreCaseConstraint should return nil constraint for empty param")
			}
		})
	}
}

// TestBuildNeIgnoreCaseConstraint tests buildNeIgnoreCaseConstraint builder function.
func TestBuildNeIgnoreCaseConstraint(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		testValue any
		wantOk    bool
		wantErr   bool
	}{
		// Valid parameter cases with different values
		{name: "different strings", param: "hello", testValue: "world", wantOk: true, wantErr: false},
		{name: "different case insensitive", param: "HELLO", testValue: "WORLD", wantOk: true, wantErr: false},
		{name: "with spaces different", param: "hello world", testValue: "goodbye world", wantOk: true, wantErr: false},

		// Same values - should fail validation
		{name: "exact match", param: "hello", testValue: "hello", wantOk: true, wantErr: true},
		{name: "case insensitive match", param: "hello", testValue: "HELLO", wantOk: true, wantErr: true},
		{name: "mixed case match", param: "HeLLo", testValue: "hEllO", wantOk: true, wantErr: true},

		// Empty parameter - should return false for ok
		{name: "empty param", param: "", testValue: "test", wantOk: false, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, ok := buildNeIgnoreCaseConstraint(tt.param)

			if tt.wantOk {
				require.True(t, ok, "buildNeIgnoreCaseConstraint should return ok=true")
				require.NotNil(t, constraint, "buildNeIgnoreCaseConstraint should return non-nil constraint")

				err := constraint.Validate(tt.testValue)
				if tt.wantErr {
					require.Error(t, err, "Validate should return error for same values")
				} else {
					require.NoError(t, err, "Validate should return no error for different values")
				}
			} else {
				require.False(t, ok, "buildNeIgnoreCaseConstraint should return ok=false for empty param")
				require.Nil(t, constraint, "buildNeIgnoreCaseConstraint should return nil constraint for empty param")
			}
		})
	}
}

// TestEqIgnoreCaseConstraint_EdgeCases tests edge cases for eqIgnoreCaseConstraint.
func TestEqIgnoreCaseConstraint_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		param   string
		wantErr bool
	}{
		// Very long strings
		{name: "long strings match", value: "thisisaverylongstringtotestcaseinsensitivecomparison", param: "THISISAVERYLONGSTRINGTOTESTCASEINSENSITIVECOMPARISON", wantErr: false},
		{name: "long strings mismatch", value: "thisisaverylongstringtotestcaseinsensitivecomparison", param: "THISISADIFFERENTVERYLONGSTRING", wantErr: true},

		// Numbers with different formats
		{name: "float scientific notation", value: "1.23e10", param: "1.23E10", wantErr: false},
		{name: "negative numbers", value: "-123", param: "-123", wantErr: false},
		{name: "positive sign", value: "+123", param: "+123", wantErr: false},

		// Single character
		{name: "single char match", value: "a", param: "A", wantErr: false},
		{name: "single char mismatch", value: "a", param: "B", wantErr: true},

		// Multiple word strings
		{name: "sentence match", value: "The Quick Brown Fox", param: "the quick brown fox", wantErr: false},
		{name: "sentence mismatch", value: "The Quick Brown Fox", param: "the slow brown fox", wantErr: true},

		// Strings with punctuation
		{name: "punctuation match", value: "Hello, World!", param: "HELLO, WORLD!", wantErr: false},
		{name: "punctuation mismatch", value: "Hello, World!", param: "Hello World", wantErr: true},

		// Tab and newline characters
		{name: "tab match", value: "hello\tworld", param: "HELLO\tWORLD", wantErr: false},
		{name: "newline match", value: "hello\nworld", param: "HELLO\nWORLD", wantErr: false},
		{name: "tab mismatch", value: "hello\tworld", param: "hello world", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := eqIgnoreCaseConstraint{value: tt.param}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestNeIgnoreCaseConstraint_EdgeCases tests edge cases for neIgnoreCaseConstraint.
func TestNeIgnoreCaseConstraint_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		param   string
		wantErr bool
	}{
		// Very long strings
		{name: "long strings different", value: "thisisaverylongstringtotestcaseinsensitivecomparison", param: "THISISADIFFERENTVERYLONGSTRING", wantErr: false},
		{name: "long strings same", value: "thisisaverylongstringtotestcaseinsensitivecomparison", param: "THISISAVERYLONGSTRINGTOTESTCASEINSENSITIVECOMPARISON", wantErr: true},

		// Numbers with different formats
		{name: "float scientific same", value: "1.23e10", param: "1.23E10", wantErr: true},
		{name: "float scientific different", value: "1.23e10", param: "4.56E10", wantErr: false},

		// Single character
		{name: "single char different", value: "a", param: "B", wantErr: false},
		{name: "single char same", value: "a", param: "A", wantErr: true},

		// Multiple word strings
		{name: "sentence different", value: "The Quick Brown Fox", param: "the slow brown fox", wantErr: false},
		{name: "sentence same", value: "The Quick Brown Fox", param: "the quick brown fox", wantErr: true},

		// Strings with punctuation
		{name: "punctuation different", value: "Hello, World!", param: "Hello World", wantErr: false},
		{name: "punctuation same", value: "Hello, World!", param: "HELLO, WORLD!", wantErr: true},

		// Tab and newline characters
		{name: "tab different content", value: "hello\tworld", param: "hello world", wantErr: false},
		{name: "newline same", value: "hello\nworld", param: "HELLO\nWORLD", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := neIgnoreCaseConstraint{value: tt.param}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}
