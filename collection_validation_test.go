package pedantigo

import (
	"testing"
)

// ==================================================
// slice element validation tests
// ==================================================

func TestSlice_ValidEmails(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","bob@example.com"]}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid emails, got %v", errs)
	}

	if len(config.Admins) != 2 {
		t.Errorf("expected 2 admins, got %d", len(config.Admins))
	}
}

func TestSlice_InvalidEmail_SingleElement(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["not-an-email"]}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for invalid email in slice")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Admins[0]" && err.Message == "must be a valid email address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Admins[0]', got %v", errs)
	}
}

func TestSlice_InvalidEmail_MultipleElements(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","invalid","bob@example.com","also-invalid"]}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(errs), errs)
	}

	// Check first error at index 1
	foundError1 := false
	for _, err := range errs {
		if err.Field == "Admins[1]" && err.Message == "must be a valid email address" {
			foundError1 = true
		}
	}
	if !foundError1 {
		t.Errorf("expected error at 'Admins[1]', got %v", errs)
	}

	// Check second error at index 3
	foundError2 := false
	for _, err := range errs {
		if err.Field == "Admins[3]" && err.Message == "must be a valid email address" {
			foundError2 = true
		}
	}
	if !foundError2 {
		t.Errorf("expected error at 'Admins[3]', got %v", errs)
	}
}

func TestSlice_MinLength(t *testing.T) {
	type User struct {
		Tags []string `json:"tags" pedantigo:"min=3"`
	}

	validator := New[User]()
	jsonData := []byte(`{"tags":["abc","de","fgh"]}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 1 {
		t.Errorf("expected 1 validation error, got %d: %v", len(errs), errs)
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Tags[1]" && err.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Tags[1]', got %v", errs)
	}
}

func TestSlice_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type User struct {
		Addresses []Address `json:"addresses"`
	}

	validator := New[User]()
	jsonData := []byte(`{"addresses":[{"city":"NYC","zip":"10001"},{"zip":"123"}]}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(errs), errs)
	}

	// Check for missing city at index 1
	foundError1 := false
	for _, err := range errs {
		if err.Field == "Addresses[1].City" && err.Message == "is required" {
			foundError1 = true
		}
	}
	if !foundError1 {
		t.Errorf("expected error at 'Addresses[1].City', got %v", errs)
	}

	// Check for short zip at index 1
	foundError2 := false
	for _, err := range errs {
		if err.Field == "Addresses[1].Zip" && err.Message == "must be at least 5 characters" {
			foundError2 = true
		}
	}
	if !foundError2 {
		t.Errorf("expected error at 'Addresses[1].Zip', got %v", errs)
	}
}

func TestSlice_EmptySlice(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":[]}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty slice, got %v", errs)
	}

	if len(config.Admins) != 0 {
		t.Errorf("expected empty admins slice, got %d elements", len(config.Admins))
	}
}

func TestSlice_NilSlice(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":null}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil slice, got %v", errs)
	}

	if config.Admins != nil {
		t.Errorf("expected nil admins slice, got %v", config.Admins)
	}
}
