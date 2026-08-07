package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartsNotWithConstraint tests startsnotwithConstraint.Validate() for prefix negation.
func TestStartsNotWithConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		prefix  string
		wantErr bool
	}{
		// Valid cases - does NOT start with prefix
		{"valid - no prefix match", "hello", "world", false},
		{"valid - different prefix", "test", "abc", false},
		{"valid - case mismatch", "Hello", "hello", false},
		{"valid - prefix at end", "worldhello", "hello", false},
		{"valid - prefix in middle", "abchellodef", "hello", false},
		{"valid - partial match", "hello", "helloworld", false},

		// Invalid cases - DOES start with prefix
		{"invalid - has prefix", "hello", "hel", true},
		{"invalid - full match", "hello", "hello", true},
		{"invalid - single char prefix", "hello", "h", true},
		{"invalid - numeric prefix", "123abc", "123", true},
		{"invalid - unicode prefix", "café au lait", "café", true},

		// Edge cases
		{"empty string", "", "test", false},
		{"nil pointer", (*string)(nil), "test", false},

		// Invalid types
		{"invalid type - int", 123, "123", true},
		{"invalid type - bool", true, "true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := startsnotwithConstraint{prefix: tt.prefix}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestEndsNotWithConstraint tests endsnotwithConstraint.Validate() for suffix negation.
func TestEndsNotWithConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		suffix  string
		wantErr bool
	}{
		// Valid cases - does NOT end with suffix
		{"valid - no suffix match", "hello", "world", false},
		{"valid - different suffix", "test", "abc", false},
		{"valid - case mismatch", "Hello", "HELLO", false},
		{"valid - suffix at start", "helloworld", "hello", false},
		{"valid - suffix in middle", "abchellodef", "hello", false},
		{"valid - partial match", "hello", "worldhello", false},

		// Invalid cases - DOES end with suffix
		{"invalid - has suffix", "hello", "llo", true},
		{"invalid - full match", "world", "world", true},
		{"invalid - single char suffix", "hello", "o", true},
		{"invalid - numeric suffix", "abc123", "123", true},
		{"invalid - unicode suffix", "au lait café", "café", true},

		// Edge cases
		{"empty string", "", "test", false},
		{"nil pointer", (*string)(nil), "test", false},

		// Invalid types
		{"invalid type - int", 123, "123", true},
		{"invalid type - bool", true, "true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := endsnotwithConstraint{suffix: tt.suffix}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestContainsAnyConstraint tests containsanyConstraint.Validate() for character set matching.
func TestContainsAnyConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		chars   string
		wantErr bool
	}{
		// Valid cases - contains at least one character from set
		{"valid - has vowel", "hello", "aeiou", false},
		{"valid - has multiple vowels", "beautiful", "aeiou", false},
		{"valid - has char at start", "xyz", "xyz", false},
		{"valid - has char in middle", "test", "es", false},
		{"valid - has char at end", "hello", "o", false},
		{"valid - has digit", "test123", "0123456789", false},
		{"valid - has special char", "hello!", "!?.", false},
		{"valid - has unicode", "café", "é", false},

		// Invalid cases - doesn't contain any character from set
		{"invalid - no vowels", "bcdfg", "aeiou", true},
		{"invalid - no match", "test", "xyz", true},
		{"invalid - no digits", "hello", "0123456789", true},
		{"invalid - no special chars", "hello", "!@#$%", true},
		{"invalid - case mismatch", "hello", "AEIOU", true},

		// Edge cases
		{"empty string", "", "aeiou", false},
		{"nil pointer", (*string)(nil), "aeiou", false},

		// Invalid types
		{"invalid type - int", 123, "123", true},
		{"invalid type - bool", true, "true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := containsanyConstraint{chars: tt.chars}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestExcludesAllConstraint tests excludesallConstraint.Validate() for character set exclusion.
func TestExcludesAllConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		chars   string
		wantErr bool
	}{
		// Valid cases - does NOT contain any character from set
		{"valid - no vowels", "bcdfg", "aeiou", false},
		{"valid - no digits", "hello", "0123456789", false},
		{"valid - no special chars", "hello", "!@#$%", false},
		{"valid - case mismatch", "hello", "AEIOU", false},
		{"valid - no match", "test", "xyz", false},

		// Invalid cases - contains at least one character from set
		{"invalid - has vowel", "hello", "aeiou", true},
		{"invalid - has multiple vowels", "beautiful", "aeiou", true},
		{"invalid - has char at start", "xyz", "xyz", true},
		{"invalid - has char in middle", "test", "es", true},
		{"invalid - has char at end", "hello", "o", true},
		{"invalid - has digit", "test123", "0123456789", true},
		{"invalid - has special char", "hello!", "!?.", true},
		{"invalid - has unicode", "café", "é", true},

		// Edge cases
		{"empty string", "", "aeiou", false},
		{"nil pointer", (*string)(nil), "aeiou", false},

		// Invalid types
		{"invalid type - int", 123, "123", true},
		{"invalid type - bool", true, "true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := excludesallConstraint{chars: tt.chars}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestExcludesRuneConstraint tests excludesruneConstraint.Validate() for single rune exclusion.
func TestExcludesRuneConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		r       rune
		wantErr bool
	}{
		// Valid cases - does NOT contain the specific rune
		{"valid - no match", "hello", 'x', false},
		{"valid - different char", "test", 'z', false},
		{"valid - case mismatch", "hello", 'H', false},
		{"valid - no digit", "hello", '5', false},
		{"valid - no special char", "hello", '!', false},
		{"valid - no unicode", "hello", 'é', false},

		// Invalid cases - DOES contain the specific rune
		{"invalid - has rune at start", "hello", 'h', true},
		{"invalid - has rune in middle", "hello", 'l', true},
		{"invalid - has rune at end", "hello", 'o', true},
		{"invalid - has digit", "test123", '1', true},
		{"invalid - has special char", "hello!", '!', true},
		{"invalid - has unicode", "café", 'é', true},
		{"invalid - has space", "hello world", ' ', true},

		// Edge cases
		{"empty string", "", 'a', false},
		{"nil pointer", (*string)(nil), 'a', false},

		// Invalid types
		{"invalid type - int", 123, '1', true},
		{"invalid type - bool", true, 't', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := excludesruneConstraint{r: tt.r}
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestBuildStartsNotWithConstraint tests buildStartsnotwithConstraint builder function.
func TestBuildStartsNotWithConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantOk    bool
		testValue string
		wantErr   bool
	}{
		{"valid builder - non-empty prefix", "test", true, "hello", false},
		{"valid builder - single char", "h", true, "world", false},
		{"invalid builder - empty prefix", "", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildStartsnotwithConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				require.NotNil(t, c)
				err := c.Validate(tt.testValue)
				checkConstraintError(t, err, tt.wantErr)
			}
		})
	}
}

// TestBuildEndsNotWithConstraint tests buildEndsnotwithConstraint builder function.
func TestBuildEndsNotWithConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantOk    bool
		testValue string
		wantErr   bool
	}{
		{"valid builder - non-empty suffix", "test", true, "hello", false},
		{"valid builder - single char ends with o", "o", true, "hello", true},      // "hello" ends with "o" -> error
		{"valid builder - single char not ends with o", "o", true, "world", false}, // "world" ends with "d" not "o" -> pass
		{"invalid builder - empty suffix", "", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildEndsnotwithConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				require.NotNil(t, c)
				err := c.Validate(tt.testValue)
				checkConstraintError(t, err, tt.wantErr)
			}
		})
	}
}

// TestBuildContainsAnyConstraint tests buildContainsanyConstraint builder function.
func TestBuildContainsAnyConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantOk    bool
		testValue string
		wantErr   bool
	}{
		{"valid builder - vowels", "aeiou", true, "hello", false},
		{"valid builder - digits", "0123456789", true, "test", true},
		{"invalid builder - empty chars", "", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildContainsanyConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				require.NotNil(t, c)
				err := c.Validate(tt.testValue)
				checkConstraintError(t, err, tt.wantErr)
			}
		})
	}
}

// TestBuildExcludesAllConstraint tests buildExcludesallConstraint builder function.
func TestBuildExcludesAllConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantOk    bool
		testValue string
		wantErr   bool
	}{
		{"valid builder - vowels", "aeiou", true, "bcdfg", false},
		{"valid builder - digits", "0123456789", true, "hello123", true},
		{"invalid builder - empty chars", "", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildExcludesallConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				require.NotNil(t, c)
				err := c.Validate(tt.testValue)
				checkConstraintError(t, err, tt.wantErr)
			}
		})
	}
}

// TestBuildExcludesRuneConstraint tests buildExcludesruneConstraint builder function.
func TestBuildExcludesRuneConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantOk    bool
		testValue string
		wantErr   bool
	}{
		{"valid builder - single char not in string", "x", true, "hello", false},
		{"valid builder - single char in string", "e", true, "hello", true},                   // "hello" contains "e" -> error
		{"valid builder - multi-char (uses first) not in string", "abc", true, "test", false}, // "test" doesn't contain "a" -> pass
		{"valid builder - multi-char (uses first) in string", "abc", true, "cat", true},       // "cat" contains "a" -> error
		{"valid builder - unicode not in string", "é", true, "hello", false},
		{"invalid builder - empty string", "", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildExcludesruneConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				require.NotNil(t, c)
				err := c.Validate(tt.testValue)
				checkConstraintError(t, err, tt.wantErr)
			}
		})
	}
}
