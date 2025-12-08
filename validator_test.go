package pedantigo

import (
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

	errs := validator.Validate(user)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidator_Min_BelowMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 15}

	errs := validator.Validate(user)
	if len(errs) == 0 {
		t.Error("expected validation error for value below minimum")
	}
}

func TestValidator_Min_AtMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 18}

	errs := validator.Validate(user)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidator_Max_AboveMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 150}

	errs := validator.Validate(user)
	if len(errs) == 0 {
		t.Error("expected validation error for value above maximum")
	}
}

func TestValidator_Max_AtMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 120}

	errs := validator.Validate(user)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidator_MinMax_InRange(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{Age: 25}

	errs := validator.Validate(user)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

// Test type for cross-field validation
type testPasswordChange struct {
	Password string `pedantigo:"required"`
	Confirm  string `pedantigo:"required"`
}

func (vpc *testPasswordChange) Validate() error {
	if vpc.Password != vpc.Confirm {
		return NewFieldError("Confirm", "passwords do not match")
	}
	return nil
}

func TestValidator_CrossField_PasswordConfirmation(t *testing.T) {
	validator := New[testPasswordChange]()
	pwd := &testPasswordChange{
		Password: "secret123",
		Confirm:  "different",
	}

	errs := validator.Validate(pwd)
	if len(errs) == 0 {
		t.Error("expected validation error for password mismatch")
	}

	// Should have cross-field error
	foundCrossFieldError := false
	for _, err := range errs {
		if err.Field == "Confirm" && err.Message == "passwords do not match" {
			foundCrossFieldError = true
		}
	}

	if !foundCrossFieldError {
		t.Error("expected cross-field validation error")
	}
}
