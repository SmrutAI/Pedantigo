package pedantigo

import (
	"testing"
)

func TestFieldError(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		message     string
		value       interface{}
		wantField   string
		wantMessage string
		wantValue   interface{}
	}{
		{
			name:        "email_required",
			field:       "Email",
			message:     "is required",
			value:       "test@example.com",
			wantField:   "Email",
			wantMessage: "is required",
			wantValue:   "test@example.com",
		},
		{
			name:        "age_minimum_constraint",
			field:       "Age",
			message:     "must be at least 18",
			value:       int(15),
			wantField:   "Age",
			wantMessage: "must be at least 18",
			wantValue:   int(15),
		},
		{
			name:        "name_too_short",
			field:       "Name",
			message:     "too short",
			value:       "Jo",
			wantField:   "Name",
			wantMessage: "too short",
			wantValue:   "Jo",
		},
		{
			name:        "nil_value",
			field:       "Profile",
			message:     "is required",
			value:       nil,
			wantField:   "Profile",
			wantMessage: "is required",
			wantValue:   nil,
		},
		{
			name:        "empty_string_value",
			field:       "Description",
			message:     "cannot be empty",
			value:       "",
			wantField:   "Description",
			wantMessage: "cannot be empty",
			wantValue:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FieldError{
				Field:   tt.field,
				Message: tt.message,
				Value:   tt.value,
			}

			if err.Field != tt.wantField {
				t.Errorf("field: expected %q, got %q", tt.wantField, err.Field)
			}

			if err.Message != tt.wantMessage {
				t.Errorf("message: expected %q, got %q", tt.wantMessage, err.Message)
			}

			if err.Value != tt.wantValue {
				t.Errorf("value: expected %v, got %v", tt.wantValue, err.Value)
			}
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
			wantErrorMsg:   "validation failed",
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
				if ve.Errors[0].Field != "Email" {
					t.Errorf("first error field: expected 'Email', got %q", ve.Errors[0].Field)
				}
				if ve.Errors[0].Message != "is required" {
					t.Errorf("first error message: expected 'is required', got %q", ve.Errors[0].Message)
				}
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
				if ve.Errors[0].Field != "Email" {
					t.Errorf("first error field: expected 'Email', got %q", ve.Errors[0].Field)
				}
				if ve.Errors[1].Field != "Age" {
					t.Errorf("second error field: expected 'Age', got %q", ve.Errors[1].Field)
				}
				if ve.Errors[1].Message != "must be at least 18" {
					t.Errorf("second error message: expected 'must be at least 18', got %q", ve.Errors[1].Message)
				}
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
				if ve.Errors[0].Field != "Email" {
					t.Errorf("first error field: expected 'Email', got %q", ve.Errors[0].Field)
				}
				if ve.Errors[2].Field != "Name" {
					t.Errorf("third error field: expected 'Name', got %q", ve.Errors[2].Field)
				}
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
				if ve.Errors[3].Field != "Phone" {
					t.Errorf("fourth error field: expected 'Phone', got %q", ve.Errors[3].Field)
				}
				if ve.Errors[4].Field != "Address" {
					t.Errorf("fifth error field: expected 'Address', got %q", ve.Errors[4].Field)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{
				Errors: tt.errors,
			}

			// Test Error() method
			if ve.Error() != tt.wantErrorMsg {
				t.Errorf("error message: expected %q, got %q", tt.wantErrorMsg, ve.Error())
			}

			// Test error count
			if len(ve.Errors) != tt.wantErrorCount {
				t.Errorf("error count: expected %d, got %d", tt.wantErrorCount, len(ve.Errors))
			}

			// Run additional validations if provided
			if tt.validateErrors != nil {
				tt.validateErrors(t, ve)
			}
		})
	}
}
