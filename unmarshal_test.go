package pedantigo

import (
	"testing"
)

func TestUnmarshal_ValidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" validate:"required"`
		Age   int    `json:"age" validate:"min=18"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com","age":25}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", user.Email)
	}

	if user.Age != 25 {
		t.Errorf("expected age 25, got %d", user.Age)
	}
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" validate:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":}`) // Invalid JSON

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected JSON decode error")
	}

	// Should return nil user on JSON decode errors
	if user != nil {
		t.Error("expected nil user on JSON decode error")
	}
}

func TestUnmarshal_ValidationError(t *testing.T) {
	type User struct {
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=18"`
	}

	validator := New[User]()
	// email is present but invalid (not an email), age is below min
	jsonData := []byte(`{"email":"notanemail","age":15}`)

	user, errs := validator.Unmarshal(jsonData)

	t.Logf("Got %d errors:", len(errs))
	for _, err := range errs {
		t.Logf("  - %s: %s", err.Field, err.Message)
	}

	if len(errs) == 0 {
		t.Error("expected validation errors")
	}

	// Should still return the user struct even with validation errors
	if user == nil {
		t.Error("expected non-nil user even with validation errors")
	}

	// Check we have errors for both fields (use struct field names, not JSON names)
	foundEmailError := false
	foundAgeError := false
	for _, err := range errs {
		if err.Field == "Email" {
			foundEmailError = true
		}
		if err.Field == "Age" {
			foundAgeError = true
		}
	}

	if !foundEmailError {
		t.Error("expected validation error for Email field")
	}

	if !foundAgeError {
		t.Error("expected validation error for Age field")
	}
}

func TestUnmarshal_DefaultValues(t *testing.T) {
	type User struct {
		Email  string `json:"email" validate:"required"`
		Role   string `json:"role" validate:"default=user"`
		Status string `json:"status" validate:"default=active"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com"}`) // Missing role and status

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	// Defaults should be applied
	if user.Role != "user" {
		t.Errorf("expected default role 'user', got %q", user.Role)
	}

	if user.Status != "active" {
		t.Errorf("expected default status 'active', got %q", user.Status)
	}
}

func TestUnmarshal_NestedValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" validate:"required,min=1"` // min=1 for non-empty string
	}

	type User struct {
		Email   string  `json:"email" validate:"required"`
		Address Address `json:"address"`
	}

	validator := New[User]()
	// City is present but empty - should fail min=1 constraint
	jsonData := []byte(`{"email":"test@example.com","address":{"city":""}}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for empty city (min=1)")
	}

	// Should have error for Address.City
	foundNestedError := false
	for _, err := range errs {
		if err.Field == "Address.City" || err.Field == "City" {
			foundNestedError = true
		}
	}

	if !foundNestedError {
		t.Errorf("expected validation error for nested City field, got errors: %v", errs)
	}

	if user == nil {
		t.Error("expected non-nil user even with validation errors")
	}
}
