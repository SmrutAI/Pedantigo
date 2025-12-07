package pedantigo

import (
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "Email",
		Message: "is required",
	}

	expected := "Email: is required"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationErrors_Error_Single(t *testing.T) {
	errs := ValidationErrors{
		{Field: "Email", Message: "is required"},
	}

	expected := "Email: is required"
	if errs.Error() != expected {
		t.Errorf("expected %q, got %q", expected, errs.Error())
	}
}

func TestValidationErrors_Error_Multiple(t *testing.T) {
	errs := ValidationErrors{
		{Field: "Email", Message: "is required"},
		{Field: "Age", Message: "must be at least 18"},
	}

	result := errs.Error()
	// Should show first error + count of remaining
	if result == "" {
		t.Error("expected non-empty error message for multiple errors")
	}

	// Should contain the first error
	if len(result) < len("Email: is required") {
		t.Errorf("error message too short: %q", result)
	}
}

func TestValidationErrors_Messages(t *testing.T) {
	errs := ValidationErrors{
		{Field: "Email", Message: "is required"},
		{Field: "Age", Message: "must be at least 18"},
	}

	msgs := errs.Messages()
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}

	if msgs[0] != "Email: is required" {
		t.Errorf("expected first message 'Email: is required', got %q", msgs[0])
	}

	if msgs[1] != "Age: must be at least 18" {
		t.Errorf("expected second message 'Age: must be at least 18', got %q", msgs[1])
	}
}

func TestNewFieldError(t *testing.T) {
	err := NewFieldError("Email", "invalid format")

	if err.Field != "Email" {
		t.Errorf("expected field 'Email', got %q", err.Field)
	}

	if err.Message != "invalid format" {
		t.Errorf("expected message 'invalid format', got %q", err.Message)
	}
}
