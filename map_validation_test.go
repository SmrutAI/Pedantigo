package pedantigo

import (
	"testing"
)

// ==================================================
// map value validation tests
// ==================================================

func TestMap_ValidEmails(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"bob@example.com"}}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid emails, got %v", errs)
	}

	if len(config.Contacts) != 2 {
		t.Errorf("expected 2 contacts, got %d", len(config.Contacts))
	}
}

func TestMap_InvalidEmail_SingleValue(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"not-an-email"}}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for invalid email in map")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Contacts[admin]" && err.Message == "must be a valid email address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Contacts[admin]', got %v", errs)
	}
}

func TestMap_InvalidEmail_MultipleValues(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"invalid","billing":"bob@example.com","sales":"also-invalid"}}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(errs), errs)
	}

	// Check that we have errors for the invalid keys (exact keys may vary due to map iteration order)
	invalidKeys := map[string]bool{"support": false, "sales": false}
	for _, err := range errs {
		if err.Message == "must be a valid email address" {
			switch err.Field {
			case "Contacts[support]":
				invalidKeys["support"] = true
			case "Contacts[sales]":
				invalidKeys["sales"] = true
			}
		}
	}

	if !invalidKeys["support"] {
		t.Errorf("expected error at 'Contacts[support]', got %v", errs)
	}
	if !invalidKeys["sales"] {
		t.Errorf("expected error at 'Contacts[sales]', got %v", errs)
	}
}

func TestMap_MinLength(t *testing.T) {
	type Config struct {
		Tags map[string]string `json:"tags" pedantigo:"min=3"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"tags":{"category":"abc","type":"de","status":"fgh"}}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 1 {
		t.Errorf("expected 1 validation error, got %d: %v", len(errs), errs)
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Tags[type]" && err.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Tags[type]', got %v", errs)
	}
}

func TestMap_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type Company struct {
		Offices map[string]Address `json:"offices"`
	}

	validator := New[Company]()
	jsonData := []byte(`{"offices":{"hq":{"city":"NYC","zip":"10001"},"branch":{"zip":"123"}}}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(errs), errs)
	}

	// Check for missing city at branch office
	foundError1 := false
	for _, err := range errs {
		if err.Field == "Offices[branch].City" && err.Message == "is required" {
			foundError1 = true
		}
	}
	if !foundError1 {
		t.Errorf("expected error at 'Offices[branch].City', got %v", errs)
	}

	// Check for short zip at branch office
	foundError2 := false
	for _, err := range errs {
		if err.Field == "Offices[branch].Zip" && err.Message == "must be at least 5 characters" {
			foundError2 = true
		}
	}
	if !foundError2 {
		t.Errorf("expected error at 'Offices[branch].Zip', got %v", errs)
	}
}

func TestMap_EmptyMap(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{}}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty map, got %v", errs)
	}

	if len(config.Contacts) != 0 {
		t.Errorf("expected empty contacts map, got %d elements", len(config.Contacts))
	}
}

func TestMap_NilMap(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":null}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil map, got %v", errs)
	}

	if config.Contacts != nil {
		t.Errorf("expected nil contacts map, got %v", config.Contacts)
	}
}
