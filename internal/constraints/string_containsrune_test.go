package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainsRuneConstraint tests containsRuneConstraint.Validate() for single rune presence.
func TestContainsRuneConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		r       rune
		wantErr bool
	}{
		// Valid cases - string contains the specific rune
		{"contains @ symbol", "test@example.com", '@', false},
		{"contains unicode é", "café", 'é', false},
		{"contains emoji", "hello🎉", '🎉', false},
		{"contains exclamation", "hello!", '!', false},
		{"contains space", "hello world", ' ', false},
		{"contains digit", "test123", '1', false},
		{"contains unicode combining mark", "cafe\u0301", '\u0301', false},
		{"contains first char", "abc", 'a', false},
		{"contains middle char", "abc", 'b', false},
		{"contains last char", "abc", 'c', false},
		{"contains hyphen", "hello-world", '-', false},
		{"contains underscore", "hello_world", '_', false},
		{"contains dot", "file.txt", '.', false},
		{"contains dollar", "price$100", '$', false},
		{"contains question mark", "really?", '?', false},
		{"contains Chinese character", "你好世界", '好', false},
		{"contains Arabic character", "مرحبا", 'ح', false},
		{"contains Cyrillic character", "привет", 'и', false},

		// Invalid cases - string does NOT contain the specific rune
		{"missing @ symbol", "hello", '@', true},
		{"missing unicode", "hello", 'é', true},
		{"missing emoji", "hello", '🎉', true},
		{"missing exclamation", "hello", '!', true},
		{"missing space", "helloworld", ' ', true},
		{"missing digit", "hello", '1', true},
		{"case mismatch uppercase", "hello", 'H', true},
		{"case mismatch lowercase", "HELLO", 'h', true},
		{"similar but different char", "hello", 'x', true},
		{"missing unicode in ASCII", "hello", 'ñ', true},

		// Edge cases
		{"empty string", "", '@', false},            // skip validation
		{"nil pointer", (*string)(nil), '@', false}, // skip validation

		// Invalid types
		{"invalid type - int", 123, '@', true},
		{"invalid type - bool", true, 't', true},
		{"invalid type - float", 3.14, '.', true},
		{"invalid type - slice", []string{"test"}, 't', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := containsRuneConstraint{r: tt.r}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestContainsRuneConstraint_ErrorCode tests that the correct error code is returned.
func TestContainsRuneConstraint_ErrorCode(t *testing.T) {
	c := containsRuneConstraint{r: '@'}
	err := c.Validate("hello")

	require.Error(t, err, "expected error for missing rune")

	var ce *ConstraintError
	require.ErrorAs(t, err, &ce, "expected ConstraintError type")
	assert.Equal(t, CodeContainsRune, ce.Code, "expected CodeContainsRune error code")
}

// TestContainsRuneConstraint_ErrorMessage tests that error messages are descriptive.
func TestContainsRuneConstraint_ErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		r           rune
		value       string
		wantContain string
	}{
		{"missing @", '@', "hello", "@"},
		{"missing emoji", '🎉', "hello", "🎉"},
		{"missing é", 'é', "hello", "é"},
		{"missing space", ' ', "hello", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := containsRuneConstraint{r: tt.r}
			err := c.Validate(tt.value)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantContain, "error message should mention the rune")
		})
	}
}

// TestBuildContainsRuneConstraint tests buildContainsRuneConstraint builder function.
func TestBuildContainsRuneConstraint(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		wantOk     bool
		expectRune rune
		testValue  string
		wantErr    bool
	}{
		// Valid builder cases
		{"valid builder - single char @", "@", true, '@', "test@example.com", false},
		{"valid builder - single char e", "e", true, 'e', "hello", false},
		{"valid builder - unicode é", "é", true, 'é', "café", false},
		{"valid builder - emoji", "🎉", true, '🎉', "party🎉", false},
		{"valid builder - space", " ", true, ' ', "hello world", false},
		{"valid builder - digit", "1", true, '1', "test123", false},
		{"valid builder - exclamation", "!", true, '!', "wow!", false},

		// Multi-char strings use first rune
		{"multi-char uses first", "abc", true, 'a', "cat", false},
		{"multi-char uses first unicode", "éx", true, 'é', "café", false},
		{"multi-char emoji uses first", "🎉🎊", true, '🎉', "hello🎉", false},

		// Test that validation works correctly
		{"validation passes when rune present", "@", true, '@', "test@", false},
		{"validation fails when rune missing", "@", true, '@', "hello", true},
		{"unicode validation passes", "ñ", true, 'ñ', "español", false},
		{"unicode validation fails", "ñ", true, 'ñ', "hello", true},

		// Invalid builder cases
		{"invalid builder - empty string", "", false, 0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildContainsRuneConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok, "builder ok status should match")

			if ok {
				require.NotNil(t, c, "expected non-nil constraint")

				// Verify the constraint has the expected rune
				containsRune, ok := c.(containsRuneConstraint)
				require.True(t, ok, "expected containsRuneConstraint type")
				assert.Equal(t, tt.expectRune, containsRune.r, "rune should match expected")

				// Test validation if we have a test value
				if tt.testValue != "" {
					err := c.Validate(tt.testValue)
					checkConstraintError(t, err, tt.wantErr)
				}
			}
		})
	}
}

// TestBuildContainsRuneConstraint_EmptyString tests that empty string returns false.
func TestBuildContainsRuneConstraint_EmptyString(t *testing.T) {
	c, ok := buildContainsRuneConstraint("")
	assert.False(t, ok, "empty string should return false")
	assert.Nil(t, c, "constraint should be nil for empty string")
}

// TestBuildContainsRuneConstraint_MultiRune tests multi-rune strings use first rune.
func TestBuildContainsRuneConstraint_MultiRune(t *testing.T) {
	// Test that "abc" uses 'a' (first rune)
	c, ok := buildContainsRuneConstraint("abc")
	require.True(t, ok, "multi-char string should build successfully")
	require.NotNil(t, c)

	containsRune, ok := c.(containsRuneConstraint)
	require.True(t, ok, "expected containsRuneConstraint type")
	assert.Equal(t, 'a', containsRune.r, "should use first rune 'a'")

	// Verify it validates correctly
	require.NoError(t, c.Validate("cat"), "should pass when 'a' is present")
	require.Error(t, c.Validate("dog"), "should fail when 'a' is missing")
}

// TestContainsRuneConstraint_UnicodeEdgeCases tests various Unicode scenarios.
func TestContainsRuneConstraint_UnicodeEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		r       rune
		value   string
		wantErr bool
	}{
		// Combining characters
		{"combining acute accent", '\u0301', "e\u0301", false},
		{"combining diaeresis", '\u0308', "o\u0308", false},

		// Emoji variations
		{"emoji with skin tone", '👍', "👍🏻", false},  // Base emoji present
		{"emoji zwj sequence", '👨', "👨‍👩‍👧", false}, // Component present

		// Zero-width characters
		{"zero-width space", '\u200B', "hello\u200Bworld", false},
		{"zero-width joiner", '\u200D', "test\u200D", false},

		// Special ASCII
		{"null character", '\x00', "test\x00", false},
		{"newline", '\n', "hello\nworld", false},
		{"tab", '\t', "hello\tworld", false},
		{"carriage return", '\r', "hello\r", false},

		// High Unicode planes
		{"musical symbol", '𝄞', "𝄞 music", false},
		{"math symbol", '∑', "∑(1..n)", false},
		{"ancient Greek", 'Ω', "Ω omega", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := containsRuneConstraint{r: tt.r}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestContainsRuneConstraint_TypeValidation tests invalid input types.
func TestContainsRuneConstraint_TypeValidation(t *testing.T) {
	c := containsRuneConstraint{r: 't'}

	tests := []struct {
		name  string
		value any
	}{
		{"int", 123},
		{"int64", int64(456)},
		{"float32", float32(3.14)},
		{"float64", float64(2.71)},
		{"bool", true},
		{"slice", []string{"test"}},
		{"map", map[string]string{"key": "value"}},
		{"struct", struct{ Name string }{Name: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Validate(tt.value)
			assert.Error(t, err, "should error for non-string type")
		})
	}
}

// TestContainsRuneConstraint_PointerHandling tests nil and non-nil string pointers.
func TestContainsRuneConstraint_PointerHandling(t *testing.T) {
	c := containsRuneConstraint{r: '@'}

	// Nil pointer should skip validation
	var nilStr *string
	err := c.Validate(nilStr)
	require.NoError(t, err, "nil pointer should skip validation")

	// Non-nil pointer with rune present
	str1 := "test@example.com"
	err = c.Validate(&str1)
	require.NoError(t, err, "pointer to string with rune should pass")

	// Non-nil pointer without rune
	str2 := "hello"
	err = c.Validate(&str2)
	require.Error(t, err, "pointer to string without rune should fail")
}

// TestContainsRuneConstraint_EmptyStringSkipsValidation tests empty string behavior.
func TestContainsRuneConstraint_EmptyStringSkipsValidation(t *testing.T) {
	c := containsRuneConstraint{r: '@'}

	// Empty string should skip validation (handled by required constraint)
	err := c.Validate("")
	assert.NoError(t, err, "empty string should skip validation")
}

// TestContainsRuneConstraint_CaseSensitivity tests that matching is case-sensitive.
func TestContainsRuneConstraint_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name    string
		r       rune
		value   string
		wantErr bool
	}{
		{"lowercase a in lowercase", 'a', "abc", false},
		{"lowercase a not in uppercase", 'a', "ABC", true},
		{"uppercase A in uppercase", 'A', "ABC", false},
		{"uppercase A not in lowercase", 'A', "abc", true},
		{"é in café", 'é', "café", false},
		{"é not in cafe", 'é', "cafe", true},
		{"É in CAFÉ", 'É', "CAFÉ", false},
		{"É not in café", 'É', "café", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := containsRuneConstraint{r: tt.r}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}
