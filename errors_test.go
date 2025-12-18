package pedantigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldError(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		code        string
		message     string
		value       interface{}
		wantField   string
		wantCode    string
		wantMessage string
		wantValue   interface{}
	}{
		{
			name:        "email_required",
			field:       "Email",
			code:        "REQUIRED",
			message:     "is required",
			value:       "test@example.com",
			wantField:   "Email",
			wantCode:    "REQUIRED",
			wantMessage: "is required",
			wantValue:   "test@example.com",
		},
		{
			name:        "age_minimum_constraint",
			field:       "Age",
			code:        "MIN_VALUE",
			message:     "must be at least 18",
			value:       int(15),
			wantField:   "Age",
			wantCode:    "MIN_VALUE",
			wantMessage: "must be at least 18",
			wantValue:   int(15),
		},
		{
			name:        "name_too_short",
			field:       "Name",
			code:        "MIN_LENGTH",
			message:     "too short",
			value:       "Jo",
			wantField:   "Name",
			wantCode:    "MIN_LENGTH",
			wantMessage: "too short",
			wantValue:   "Jo",
		},
		{
			name:        "nil_value",
			field:       "Profile",
			code:        "REQUIRED",
			message:     "is required",
			value:       nil,
			wantField:   "Profile",
			wantCode:    "REQUIRED",
			wantMessage: "is required",
			wantValue:   nil,
		},
		{
			name:        "empty_string_value",
			field:       "Description",
			code:        "MIN_LENGTH",
			message:     "cannot be empty",
			value:       "",
			wantField:   "Description",
			wantCode:    "MIN_LENGTH",
			wantMessage: "cannot be empty",
			wantValue:   "",
		},
		{
			name:        "email_format_error",
			field:       "Email",
			code:        "INVALID_EMAIL",
			message:     "must be a valid email",
			value:       "not-an-email",
			wantField:   "Email",
			wantCode:    "INVALID_EMAIL",
			wantMessage: "must be a valid email",
			wantValue:   "not-an-email",
		},
		{
			name:        "url_format_error",
			field:       "Website",
			code:        "INVALID_URL",
			message:     "must be a valid URL",
			value:       "not a url",
			wantField:   "Website",
			wantCode:    "INVALID_URL",
			wantMessage: "must be a valid URL",
			wantValue:   "not a url",
		},
		{
			name:        "empty_code",
			field:       "CustomField",
			code:        "",
			message:     "custom validation failed",
			value:       "value",
			wantField:   "CustomField",
			wantCode:    "",
			wantMessage: "custom validation failed",
			wantValue:   "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FieldError{
				Field:   tt.field,
				Code:    tt.code,
				Message: tt.message,
				Value:   tt.value,
			}

			assert.Equal(t, tt.wantField, err.Field)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
			assert.Equal(t, tt.wantValue, err.Value)
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name           string
		errors         []FieldError
		wantErrorMsg   string
		wantErrorCount int
		validateErrors func(t *testing.T, ve *ValidationError)
	}{
		{
			name:           "empty_errors_list",
			errors:         []FieldError{},
			wantErrorMsg:   "no errors found",
			wantErrorCount: 0,
			validateErrors: nil,
		},
		{
			name: "single_field_error",
			errors: []FieldError{
				{Field: "Email", Message: "is required"},
			},
			wantErrorMsg:   "Email: is required",
			wantErrorCount: 1,
			validateErrors: func(t *testing.T, ve *ValidationError) {
				assert.Equal(t, "Email", ve.Errors[0].Field)
				assert.Equal(t, "is required", ve.Errors[0].Message)
			},
		},
		{
			name: "two_field_errors",
			errors: []FieldError{
				{Field: "Email", Message: "is required"},
				{Field: "Age", Message: "must be at least 18"},
			},
			wantErrorMsg:   "Email: is required (and 1 more errors)",
			wantErrorCount: 2,
			validateErrors: func(t *testing.T, ve *ValidationError) {
				assert.Equal(t, "Email", ve.Errors[0].Field)
				assert.Equal(t, "Age", ve.Errors[1].Field)
				assert.Equal(t, "must be at least 18", ve.Errors[1].Message)
			},
		},
		{
			name: "three_field_errors",
			errors: []FieldError{
				{Field: "Email", Message: "is required"},
				{Field: "Age", Message: "must be at least 18"},
				{Field: "Name", Message: "too short"},
			},
			wantErrorMsg:   "Email: is required (and 2 more errors)",
			wantErrorCount: 3,
			validateErrors: func(t *testing.T, ve *ValidationError) {
				assert.Equal(t, "Email", ve.Errors[0].Field)
				assert.Equal(t, "Name", ve.Errors[2].Field)
			},
		},
		{
			name: "many_field_errors",
			errors: []FieldError{
				{Field: "Email", Message: "is required"},
				{Field: "Age", Message: "must be at least 18"},
				{Field: "Name", Message: "too short"},
				{Field: "Phone", Message: "invalid format"},
				{Field: "Address", Message: "is required"},
			},
			wantErrorMsg:   "Email: is required (and 4 more errors)",
			wantErrorCount: 5,
			validateErrors: func(t *testing.T, ve *ValidationError) {
				assert.Equal(t, "Phone", ve.Errors[3].Field)
				assert.Equal(t, "Address", ve.Errors[4].Field)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{
				Errors: tt.errors,
			}

			// Test Error() method
			assert.Equal(t, tt.wantErrorMsg, ve.Error())

			// Test error count
			assert.Len(t, ve.Errors, tt.wantErrorCount)

			// Run additional validations if provided
			if tt.validateErrors != nil {
				tt.validateErrors(t, ve)
			}
		})
	}
}

func TestFieldError_HasCode(t *testing.T) {
	t.Run("field_error_with_code", func(t *testing.T) {
		fe := FieldError{
			Field:   "email",
			Code:    "INVALID_EMAIL",
			Message: "must be a valid email",
			Value:   "not-an-email",
		}

		assert.Equal(t, "email", fe.Field)
		assert.Equal(t, "INVALID_EMAIL", fe.Code)
		assert.Equal(t, "must be a valid email", fe.Message)
		assert.Equal(t, "not-an-email", fe.Value)
	})

	t.Run("field_error_with_empty_code", func(t *testing.T) {
		fe := FieldError{
			Field:   "customField",
			Code:    "",
			Message: "custom validation failed",
			Value:   123,
		}

		assert.Equal(t, "customField", fe.Field)
		assert.Empty(t, fe.Code)
		assert.Equal(t, "custom validation failed", fe.Message)
		assert.Equal(t, 123, fe.Value)
	})

	t.Run("multiple_different_codes", func(t *testing.T) {
		testCases := []struct {
			code     string
			expected string
		}{
			{"REQUIRED", "REQUIRED"},
			{"MIN_LENGTH", "MIN_LENGTH"},
			{"MAX_VALUE", "MAX_VALUE"},
			{"INVALID_EMAIL", "INVALID_EMAIL"},
			{"PATTERN_MISMATCH", "PATTERN_MISMATCH"},
		}

		for _, tc := range testCases {
			fe := FieldError{
				Field:   "testField",
				Code:    tc.code,
				Message: "test message",
			}
			assert.Equal(t, tc.expected, fe.Code)
		}
	})
}

func TestValidationError_WithCodes(t *testing.T) {
	t.Run("single_error_with_code", func(t *testing.T) {
		ve := &ValidationError{
			Errors: []FieldError{
				{
					Field:   "email",
					Code:    "REQUIRED",
					Message: "is required",
					Value:   nil,
				},
			},
		}

		assert.Len(t, ve.Errors, 1)
		assert.Equal(t, "REQUIRED", ve.Errors[0].Code)
		assert.Equal(t, "email", ve.Errors[0].Field)
		assert.Equal(t, "is required", ve.Errors[0].Message)
	})

	t.Run("multiple_errors_with_different_codes", func(t *testing.T) {
		ve := &ValidationError{
			Errors: []FieldError{
				{
					Field:   "email",
					Code:    "INVALID_EMAIL",
					Message: "must be a valid email",
					Value:   "not-an-email",
				},
				{
					Field:   "age",
					Code:    "MIN_VALUE",
					Message: "must be at least 18",
					Value:   15,
				},
				{
					Field:   "password",
					Code:    "MIN_LENGTH",
					Message: "must be at least 8 characters",
					Value:   "short",
				},
			},
		}

		assert.Len(t, ve.Errors, 3)

		// Verify each error has correct code
		assert.Equal(t, "INVALID_EMAIL", ve.Errors[0].Code)
		assert.Equal(t, "MIN_VALUE", ve.Errors[1].Code)
		assert.Equal(t, "MIN_LENGTH", ve.Errors[2].Code)

		// Verify error message format
		expectedMsg := "email: must be a valid email (and 2 more errors)"
		assert.Equal(t, expectedMsg, ve.Error())
	})

	t.Run("validation_error_codes_preserved", func(t *testing.T) {
		errors := []FieldError{
			{Field: "username", Code: "MIN_LENGTH", Message: "too short", Value: "ab"},
			{Field: "email", Code: "INVALID_EMAIL", Message: "invalid format", Value: "bad@"},
			{Field: "age", Code: "MAX_VALUE", Message: "too old", Value: 150},
			{Field: "phone", Code: "PATTERN_MISMATCH", Message: "invalid pattern", Value: "123"},
		}

		ve := &ValidationError{Errors: errors}

		// All codes should be preserved
		for i, err := range ve.Errors {
			assert.Equal(t, errors[i].Code, err.Code,
				"error code should be preserved at index %d", i)
		}
	})

	t.Run("mixed_empty_and_non_empty_codes", func(t *testing.T) {
		ve := &ValidationError{
			Errors: []FieldError{
				{Field: "field1", Code: "REQUIRED", Message: "required"},
				{Field: "field2", Code: "", Message: "custom error"},
				{Field: "field3", Code: "INVALID_EMAIL", Message: "bad email"},
			},
		}

		assert.Equal(t, "REQUIRED", ve.Errors[0].Code)
		assert.Empty(t, ve.Errors[1].Code)
		assert.Equal(t, "INVALID_EMAIL", ve.Errors[2].Code)
	})
}

func TestFieldError_CodeUsagePatterns(t *testing.T) {
	t.Run("code_for_programmatic_error_handling", func(t *testing.T) {
		ve := &ValidationError{
			Errors: []FieldError{
				{Field: "email", Code: "REQUIRED", Message: "email is required"},
				{Field: "age", Code: "MIN_VALUE", Message: "age must be at least 18"},
				{Field: "username", Code: "INVALID_FORMAT", Message: "invalid username format"},
			},
		}

		// Example: Filter errors by code
		var requiredErrors []FieldError
		for _, err := range ve.Errors {
			if err.Code == "REQUIRED" {
				requiredErrors = append(requiredErrors, err)
			}
		}

		assert.Len(t, requiredErrors, 1)
		assert.Equal(t, "email", requiredErrors[0].Field)
	})

	t.Run("code_enables_error_categorization", func(t *testing.T) {
		errors := []FieldError{
			{Field: "email", Code: "INVALID_EMAIL", Message: "bad email"},
			{Field: "url", Code: "INVALID_URL", Message: "bad url"},
			{Field: "uuid", Code: "INVALID_UUID", Message: "bad uuid"},
			{Field: "name", Code: "MIN_LENGTH", Message: "too short"},
			{Field: "age", Code: "MIN_VALUE", Message: "too young"},
		}

		ve := &ValidationError{Errors: errors}

		// Count format validation errors (INVALID_*)
		formatErrors := 0
		for _, err := range ve.Errors {
			if len(err.Code) >= 8 && err.Code[:8] == "INVALID_" {
				formatErrors++
			}
		}

		assert.Equal(t, 3, formatErrors, "should have 3 format validation errors")
	})

	t.Run("code_supports_internationalization", func(t *testing.T) {
		// Codes enable mapping to localized messages
		fe := FieldError{
			Field:   "email",
			Code:    "REQUIRED",
			Message: "is required", // English message
			Value:   nil,
		}

		// Simulate i18n lookup by code
		i18nMessages := map[string]string{
			"REQUIRED":      "est obligatoire",            // French
			"INVALID_EMAIL": "doit être un e-mail valide", // French
		}

		localizedMsg := i18nMessages[fe.Code]
		assert.Equal(t, "est obligatoire", localizedMsg)
	})
}

func TestValidationError_ErrorCodePropagation(t *testing.T) {
	t.Run("error_codes_survive_error_collection", func(t *testing.T) {
		// Simulate collecting errors from multiple validation steps
		collectedErrors := []FieldError{
			// Step 1: Required field validation
			{
				Field:   "email",
				Code:    "REQUIRED",
				Message: "email is required",
			},
			// Step 2: Format validation
			{
				Field:   "phone",
				Code:    "PATTERN_MISMATCH",
				Message: "phone number format invalid",
			},
			// Step 3: Range validation
			{
				Field:   "age",
				Code:    "MIN_VALUE",
				Message: "age must be at least 18",
			},
		}

		ve := &ValidationError{Errors: collectedErrors}

		// All codes should be intact
		assert.Equal(t, "REQUIRED", ve.Errors[0].Code)
		assert.Equal(t, "PATTERN_MISMATCH", ve.Errors[1].Code)
		assert.Equal(t, "MIN_VALUE", ve.Errors[2].Code)
	})
}
