package constraints

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstraintError_Interface(t *testing.T) {
	t.Run("implements_error_interface", func(t *testing.T) {
		var err error = &ConstraintError{
			Code:    CodeRequired,
			Message: "field is required",
		}

		require.Error(t, err)
		assert.Equal(t, "field is required", err.Error())
	})

	t.Run("can_be_unwrapped_with_errors_as", func(t *testing.T) {
		constraintErr := &ConstraintError{
			Code:    CodeInvalidEmail,
			Message: "invalid email format",
		}

		var target *ConstraintError
		require.ErrorAs(t, constraintErr, &target)
		assert.Equal(t, CodeInvalidEmail, target.Code)
		assert.Equal(t, "invalid email format", target.Message)
	})
}

func TestConstraintError_Error(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		message     string
		wantMessage string
	}{
		{
			name:        "required_error",
			code:        CodeRequired,
			message:     "field is required",
			wantMessage: "field is required",
		},
		{
			name:        "email_error",
			code:        CodeInvalidEmail,
			message:     "must be a valid email address",
			wantMessage: "must be a valid email address",
		},
		{
			name:        "min_length_error",
			code:        CodeMinLength,
			message:     "must be at least 5 characters",
			wantMessage: "must be at least 5 characters",
		},
		{
			name:        "empty_message",
			code:        CodeRequired,
			message:     "",
			wantMessage: "",
		},
		{
			name:        "multiline_message",
			code:        CodePatternMismatch,
			message:     "value does not match pattern\nexpected: ^[a-z]+$",
			wantMessage: "value does not match pattern\nexpected: ^[a-z]+$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ConstraintError{
				Code:    tt.code,
				Message: tt.message,
			}

			assert.Equal(t, tt.wantMessage, err.Error())
		})
	}
}

func TestNewConstraintError(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		message     string
		wantCode    string
		wantMessage string
	}{
		{
			name:        "required_constraint",
			code:        CodeRequired,
			message:     "field is required",
			wantCode:    CodeRequired,
			wantMessage: "field is required",
		},
		{
			name:        "invalid_email",
			code:        CodeInvalidEmail,
			message:     "must be a valid email",
			wantCode:    CodeInvalidEmail,
			wantMessage: "must be a valid email",
		},
		{
			name:        "min_value",
			code:        CodeMinValue,
			message:     "must be at least 10",
			wantCode:    CodeMinValue,
			wantMessage: "must be at least 10",
		},
		{
			name:        "empty_code_and_message",
			code:        "",
			message:     "",
			wantCode:    "",
			wantMessage: "",
		},
		{
			name:        "special_characters_in_message",
			code:        CodePatternMismatch,
			message:     "pattern: ^[a-zA-Z0-9_-]+$ (special chars: !@#$%)",
			wantCode:    CodePatternMismatch,
			wantMessage: "pattern: ^[a-zA-Z0-9_-]+$ (special chars: !@#$%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConstraintError(tt.code, tt.message)

			require.NotNil(t, err)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
			assert.Equal(t, tt.wantMessage, err.Error())
		})
	}
}

func TestNewConstraintErrorf(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		format      string
		args        []any
		wantCode    string
		wantMessage string
	}{
		{
			name:        "min_length_formatted",
			code:        CodeMinLength,
			format:      "must be at least %d characters",
			args:        []any{5},
			wantCode:    CodeMinLength,
			wantMessage: "must be at least 5 characters",
		},
		{
			name:        "max_length_formatted",
			code:        CodeMaxLength,
			format:      "must be at most %d characters",
			args:        []any{100},
			wantCode:    CodeMaxLength,
			wantMessage: "must be at most 100 characters",
		},
		{
			name:        "min_value_formatted",
			code:        CodeMinValue,
			format:      "must be at least %d",
			args:        []any{18},
			wantCode:    CodeMinValue,
			wantMessage: "must be at least 18",
		},
		{
			name:        "multiple_format_args",
			code:        CodePatternMismatch,
			format:      "value %q does not match pattern %q",
			args:        []any{"abc123", "^[a-z]+$"},
			wantCode:    CodePatternMismatch,
			wantMessage: `value "abc123" does not match pattern "^[a-z]+$"`,
		},
		{
			name:        "no_format_args",
			code:        CodeRequired,
			format:      "field is required",
			args:        []any{},
			wantCode:    CodeRequired,
			wantMessage: "field is required",
		},
		{
			name:        "string_formatting",
			code:        CodeMustContain,
			format:      "must contain %s",
			args:        []any{"@"},
			wantCode:    CodeMustContain,
			wantMessage: "must contain @",
		},
		{
			name:        "float_formatting",
			code:        CodeMultipleOf,
			format:      "must be a multiple of %.2f",
			args:        []any{0.5},
			wantCode:    CodeMultipleOf,
			wantMessage: "must be a multiple of 0.50",
		},
		{
			name:        "complex_formatting",
			code:        CodeExclusiveMin,
			format:      "must be greater than %d (exclusive), got %d",
			args:        []any{10, 5},
			wantCode:    CodeExclusiveMin,
			wantMessage: "must be greater than 10 (exclusive), got 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConstraintErrorf(tt.code, tt.format, tt.args...)

			require.NotNil(t, err)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
			assert.Equal(t, tt.wantMessage, err.Error())
		})
	}
}

func TestErrorCodes_AllDefined(t *testing.T) {
	t.Run("all_error_codes_non_empty", func(t *testing.T) {
		codes := map[string]string{
			// Required constraints
			"CodeRequired":        CodeRequired,
			"CodeRequiredIf":      CodeRequiredIf,
			"CodeRequiredUnless":  CodeRequiredUnless,
			"CodeRequiredWith":    CodeRequiredWith,
			"CodeRequiredWithout": CodeRequiredWithout,

			// Format constraints
			"CodeInvalidEmail":    CodeInvalidEmail,
			"CodeInvalidURL":      CodeInvalidURL,
			"CodeInvalidUUID":     CodeInvalidUUID,
			"CodeInvalidIPv4":     CodeInvalidIPv4,
			"CodeInvalidIPv6":     CodeInvalidIPv6,
			"CodeInvalidIP":       CodeInvalidIP,
			"CodePatternMismatch": CodePatternMismatch,

			// Length constraints
			"CodeMinLength":   CodeMinLength,
			"CodeMaxLength":   CodeMaxLength,
			"CodeExactLength": CodeExactLength,

			// Numeric constraints
			"CodeMinValue":       CodeMinValue,
			"CodeMaxValue":       CodeMaxValue,
			"CodeExclusiveMin":   CodeExclusiveMin,
			"CodeExclusiveMax":   CodeExclusiveMax,
			"CodeMustBePositive": CodeMustBePositive,
			"CodeMustBeNegative": CodeMustBeNegative,
			"CodeMultipleOf":     CodeMultipleOf,

			// String constraints
			"CodeMustBeASCII":     CodeMustBeASCII,
			"CodeMustBeAlpha":     CodeMustBeAlpha,
			"CodeMustBeAlphanum":  CodeMustBeAlphanum,
			"CodeMustContain":     CodeMustContain,
			"CodeMustNotContain":  CodeMustNotContain,
			"CodeMustStartWith":   CodeMustStartWith,
			"CodeMustEndWith":     CodeMustEndWith,
			"CodeMustBeLowercase": CodeMustBeLowercase,
			"CodeMustBeUppercase": CodeMustBeUppercase,

			// Enum/const constraints
			"CodeInvalidEnum":   CodeInvalidEnum,
			"CodeConstMismatch": CodeConstMismatch,

			// Collection constraints
			"CodeNotUnique": CodeNotUnique,

			// Cross-field constraints
			"CodeMustEqualField":    CodeMustEqualField,
			"CodeMustNotEqualField": CodeMustNotEqualField,
			"CodeMustBeGTField":     CodeMustBeGTField,
			"CodeMustBeGTEField":    CodeMustBeGTEField,
			"CodeMustBeLTField":     CodeMustBeLTField,
			"CodeMustBeLTEField":    CodeMustBeLTEField,
			"CodeExcludedIf":        CodeExcludedIf,
			"CodeExcludedUnless":    CodeExcludedUnless,
			"CodeExcludedWith":      CodeExcludedWith,
			"CodeExcludedWithout":   CodeExcludedWithout,

			// Type errors
			"CodeUnknownField": CodeUnknownField,
		}

		for name, code := range codes {
			assert.NotEmpty(t, code, "error code %s should not be empty", name)
		}
	})

	t.Run("error_codes_follow_screaming_snake_case", func(t *testing.T) {
		codes := []string{
			CodeRequired,
			CodeRequiredIf,
			CodeRequiredUnless,
			CodeRequiredWith,
			CodeRequiredWithout,
			CodeInvalidEmail,
			CodeInvalidURL,
			CodeInvalidUUID,
			CodeInvalidIPv4,
			CodeInvalidIPv6,
			CodeInvalidIP,
			CodePatternMismatch,
			CodeMinLength,
			CodeMaxLength,
			CodeExactLength,
			CodeMinValue,
			CodeMaxValue,
			CodeExclusiveMin,
			CodeExclusiveMax,
			CodeMustBePositive,
			CodeMustBeNegative,
			CodeMultipleOf,
			CodeMustBeASCII,
			CodeMustBeAlpha,
			CodeMustBeAlphanum,
			CodeMustContain,
			CodeMustNotContain,
			CodeMustStartWith,
			CodeMustEndWith,
			CodeMustBeLowercase,
			CodeMustBeUppercase,
			CodeInvalidEnum,
			CodeConstMismatch,
			CodeNotUnique,
			CodeMustEqualField,
			CodeMustNotEqualField,
			CodeMustBeGTField,
			CodeMustBeGTEField,
			CodeMustBeLTField,
			CodeMustBeLTEField,
			CodeExcludedIf,
			CodeExcludedUnless,
			CodeExcludedWith,
			CodeExcludedWithout,
			CodeUnknownField,
		}

		for _, code := range codes {
			// SCREAMING_SNAKE_CASE means:
			// - All uppercase letters
			// - Underscores for word separation
			// - No lowercase letters
			// - No spaces or other special characters
			assert.Equal(t, strings.ToUpper(code), code,
				"error code %q should be all uppercase (SCREAMING_SNAKE_CASE)", code)
			assert.NotContains(t, code, " ",
				"error code %q should not contain spaces", code)
			assert.NotContains(t, code, "-",
				"error code %q should not contain hyphens (use underscores)", code)

			// Should only contain uppercase letters, underscores, and numbers
			for _, char := range code {
				valid := (char >= 'A' && char <= 'Z') || char == '_' || (char >= '0' && char <= '9')
				assert.True(t, valid,
					"error code %q contains invalid character %q (should only have A-Z, _, 0-9)", code, char)
			}
		}
	})

	t.Run("error_codes_are_unique", func(t *testing.T) {
		codes := []string{
			CodeRequired,
			CodeRequiredIf,
			CodeRequiredUnless,
			CodeRequiredWith,
			CodeRequiredWithout,
			CodeInvalidEmail,
			CodeInvalidURL,
			CodeInvalidUUID,
			CodeInvalidIPv4,
			CodeInvalidIPv6,
			CodeInvalidIP,
			CodePatternMismatch,
			CodeMinLength,
			CodeMaxLength,
			CodeExactLength,
			CodeMinValue,
			CodeMaxValue,
			CodeExclusiveMin,
			CodeExclusiveMax,
			CodeMustBePositive,
			CodeMustBeNegative,
			CodeMultipleOf,
			CodeMustBeASCII,
			CodeMustBeAlpha,
			CodeMustBeAlphanum,
			CodeMustContain,
			CodeMustNotContain,
			CodeMustStartWith,
			CodeMustEndWith,
			CodeMustBeLowercase,
			CodeMustBeUppercase,
			CodeInvalidEnum,
			CodeConstMismatch,
			CodeNotUnique,
			CodeMustEqualField,
			CodeMustNotEqualField,
			CodeMustBeGTField,
			CodeMustBeGTEField,
			CodeMustBeLTField,
			CodeMustBeLTEField,
			CodeExcludedIf,
			CodeExcludedUnless,
			CodeExcludedWith,
			CodeExcludedWithout,
			CodeUnknownField,
		}

		seen := make(map[string]bool)
		for _, code := range codes {
			assert.False(t, seen[code], "duplicate error code found: %q", code)
			seen[code] = true
		}

		// Verify we have all expected error codes (45 defined in error_codes.go)
		assert.Len(t, seen, 45, "expected 45 unique error codes")
	})
}

func TestErrorCodes_ExpectedValues(t *testing.T) {
	// Verify specific error codes have expected string values
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Required", CodeRequired, "REQUIRED"},
		{"InvalidEmail", CodeInvalidEmail, "INVALID_EMAIL"},
		{"MinLength", CodeMinLength, "MIN_LENGTH"},
		{"MaxValue", CodeMaxValue, "MAX_VALUE"},
		{"PatternMismatch", CodePatternMismatch, "PATTERN_MISMATCH"},
		{"MustBePositive", CodeMustBePositive, "MUST_BE_POSITIVE"},
		{"NotUnique", CodeNotUnique, "NOT_UNIQUE"},
		{"UnknownField", CodeUnknownField, "UNKNOWN_FIELD"},
		{"RequiredIf", CodeRequiredIf, "REQUIRED_IF"},
		{"ExclusiveMax", CodeExclusiveMax, "EXCLUSIVE_MAX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.code)
		})
	}
}

func TestConstraintError_UsageExample(t *testing.T) {
	t.Run("example_required_field", func(t *testing.T) {
		err := NewConstraintError(CodeRequired, "field 'email' is required")

		assert.Equal(t, CodeRequired, err.Code)
		assert.Equal(t, "field 'email' is required", err.Message)
		assert.Equal(t, "field 'email' is required", err.Error())
	})

	t.Run("example_min_length_formatted", func(t *testing.T) {
		minLen := 8
		err := NewConstraintErrorf(CodeMinLength, "password must be at least %d characters", minLen)

		assert.Equal(t, CodeMinLength, err.Code)
		assert.Equal(t, "password must be at least 8 characters", err.Message)
	})

	t.Run("example_pattern_mismatch", func(t *testing.T) {
		value := "user@123"
		pattern := "^[a-z]+$"
		err := NewConstraintErrorf(CodePatternMismatch, "username %q does not match pattern %q", value, pattern)

		assert.Equal(t, CodePatternMismatch, err.Code)
		assert.Contains(t, err.Message, "user@123")
		assert.Contains(t, err.Message, "^[a-z]+$")
	})
}
