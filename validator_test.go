package pedantigo

import (
	"encoding/json"
	"testing"
)

// NOTE: 'required' is only checked during Unmarshal (missing JSON keys), not Validate()
// Validate() only checks value constraints (min, max, email, etc.)

func TestValidator_Required_Present(t *testing.T) {
	type User struct {
		Email string `pedantigo:"required"`
	}

	validator := New[User]()
	user := &User{Email: "test@example.com"}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

func TestValidator_Min_BelowMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 15}

	err := validator.Validate(user)
	if err == nil {
		t.Error("expected validation error for value below minimum")
	}
}

func TestValidator_Min_AtMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 18}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

func TestValidator_Max_AboveMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 150}

	err := validator.Validate(user)
	if err == nil {
		t.Error("expected validation error for value above maximum")
	}
}

func TestValidator_Max_AtMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 120}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

func TestValidator_MinMax_InRange(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{Age: 25}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

// Test type for cross-field validation
type testPasswordChange struct {
	Password string `pedantigo:"required"`
	Confirm  string `pedantigo:"required"`
}

func (vpc *testPasswordChange) Validate() error {
	if vpc.Password != vpc.Confirm {
		return &ValidationError{
			Errors: []FieldError{{
				Field:   "Confirm",
				Message: "passwords do not match",
			}},
		}
	}
	return nil
}

func TestValidator_CrossField_PasswordConfirmation(t *testing.T) {
	validator := New[testPasswordChange]()
	pwd := &testPasswordChange{
		Password: "secret123",
		Confirm:  "different",
	}

	err := validator.Validate(pwd)
	if err == nil {
		t.Error("expected validation error for password mismatch")
	}

	// Should have cross-field error
	foundCrossFieldError := false
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Confirm" && fieldErr.Message == "passwords do not match" {
			foundCrossFieldError = true
		}
	}

	if !foundCrossFieldError {
		t.Error("expected cross-field validation error")
	}
}

// TestMarshal_Valid verifies that Marshal returns JSON for valid structs
func TestMarshal_Valid(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	data, err := validator.Marshal(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify JSON is valid and contains expected fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Marshal returned invalid JSON: %v", err)
	}

	if result["name"] != "John Doe" {
		t.Errorf("expected name='John Doe', got %v", result["name"])
	}
	if result["email"] != "john@example.com" {
		t.Errorf("expected email='john@example.com', got %v", result["email"])
	}
	if result["age"] != float64(25) {
		t.Errorf("expected age=25, got %v", result["age"])
	}
}

// TestMarshal_Invalid verifies that Marshal returns validation errors for invalid structs
func TestMarshal_Invalid(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{
		Name:  "J",          // Too short (min=2)
		Email: "notanemail", // Invalid email
		Age:   15,           // Too young (min=18)
	}

	data, err := validator.Marshal(user)

	// Should return validation error, not JSON
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if data != nil {
		t.Errorf("expected nil data when validation fails, got %d bytes", len(data))
	}

	// Verify it's a ValidationError with multiple field errors
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if len(ve.Errors) != 3 {
		t.Errorf("expected 3 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}

	// Check that errors are for the expected fields
	errorFields := make(map[string]bool)
	for _, fieldErr := range ve.Errors {
		errorFields[fieldErr.Field] = true
	}

	if !errorFields["Name"] {
		t.Error("expected validation error for Name field")
	}
	if !errorFields["Email"] {
		t.Error("expected validation error for Email field")
	}
	if !errorFields["Age"] {
		t.Error("expected validation error for Age field")
	}
}

// TestMarshal_Nil verifies that Marshal handles nil pointer appropriately
func TestMarshal_Nil(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"min=2"`
	}

	validator := New[User]()

	// Pass nil pointer
	data, err := validator.Marshal(nil)

	// Should handle nil gracefully (either return error or marshal "null")
	// Let's check what actually happens - if Validate accepts nil, json.Marshal will return "null"
	// If Validate rejects nil, we'll get a validation error
	if err != nil {
		// Validation error is acceptable for nil
		t.Logf("Marshal(nil) returned error: %v", err)
	} else if string(data) != "null" {
		t.Errorf("expected Marshal(nil) to return 'null', got %q", string(data))
	}
}
