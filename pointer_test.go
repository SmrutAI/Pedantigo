package pedantigo

import (
	"testing"
)

// Test pointer field with explicit value
func TestPointer_ExplicitValue(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice","age":25}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	if user.Name == nil {
		t.Fatal("expected non-nil Name pointer")
	}
	if *user.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", *user.Name)
	}

	if user.Age == nil {
		t.Fatal("expected non-nil Age pointer")
	}
	if *user.Age != 25 {
		t.Errorf("expected age 25, got %d", *user.Age)
	}
}

// Test pointer field with explicit zero value (should create pointer to zero)
func TestPointer_ExplicitZero(t *testing.T) {
	type Config struct {
		Port    *int    `json:"port"`
		Enabled *bool   `json:"enabled"`
		Name    *string `json:"name"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0,"enabled":false,"name":""}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	// Explicit zeros should create pointers to zero values
	if config.Port == nil {
		t.Fatal("expected non-nil Port pointer for explicit 0")
	}
	if *config.Port != 0 {
		t.Errorf("expected port 0, got %d", *config.Port)
	}

	if config.Enabled == nil {
		t.Fatal("expected non-nil Enabled pointer for explicit false")
	}
	if *config.Enabled != false {
		t.Errorf("expected enabled false, got %v", *config.Enabled)
	}

	if config.Name == nil {
		t.Fatal("expected non-nil Name pointer for explicit empty string")
	}
	if *config.Name != "" {
		t.Errorf("expected name empty string, got %q", *config.Name)
	}
}

// Test pointer field with explicit null (should be nil pointer)
func TestPointer_ExplicitNull(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null,"age":null}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	// Explicit null should result in nil pointers
	if user.Name != nil {
		t.Errorf("expected nil Name pointer for explicit null, got %v", *user.Name)
	}

	if user.Age != nil {
		t.Errorf("expected nil Age pointer for explicit null, got %v", *user.Age)
	}
}

// Test pointer field missing from JSON (should be nil pointer)
func TestPointer_Missing(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Both fields missing

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	// Missing fields should result in nil pointers
	if user.Name != nil {
		t.Errorf("expected nil Name pointer for missing field, got %v", *user.Name)
	}

	if user.Age != nil {
		t.Errorf("expected nil Age pointer for missing field, got %v", *user.Age)
	}
}

// Test required pointer field with explicit value
func TestPointer_RequiredWithValue(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	if user.Name == nil {
		t.Fatal("expected non-nil Name pointer")
	}
	if *user.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", *user.Name)
	}
}

// Test required pointer field missing (should fail)
func TestPointer_RequiredMissing(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Missing required field

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for missing required field")
	}

	// Check for required field error
	foundRequiredError := false
	for _, err := range errs {
		if err.Field == "name" && err.Message == "is required" {
			foundRequiredError = true
		}
	}

	if !foundRequiredError {
		t.Errorf("expected 'is required' error for name field, got %v", errs)
	}
}

// Test required pointer field with explicit null (should pass - field is present)
func TestPointer_RequiredWithNull(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null}`) // Field present but null

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors (field is present), got %v", errs)
	}

	// Required means "field must be present", not "value can't be nil"
	if user.Name != nil {
		t.Errorf("expected nil Name pointer for explicit null, got %v", *user.Name)
	}
}

// Test pointer field with default value
func TestPointer_WithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{}`) // Missing port field

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	// Default should be applied to missing field
	if config.Port == nil {
		t.Fatal("expected non-nil Port pointer with default")
	}
	if *config.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", *config.Port)
	}
}

// Test pointer field with explicit zero and default (should keep zero)
func TestPointer_ExplicitZeroWithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0}`) // Explicit zero

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	// Explicit zero should be kept, not replaced with default
	if config.Port == nil {
		t.Fatal("expected non-nil Port pointer for explicit 0")
	}
	if *config.Port != 0 {
		t.Errorf("expected port 0 (not default), got %d", *config.Port)
	}
}

// Test nested struct with pointer fields
func TestPointer_NestedStruct(t *testing.T) {
	type Address struct {
		Street *string `json:"street"`
		City   *string `json:"city"`
	}

	type User struct {
		Name    string   `json:"name"`
		Address *Address `json:"address"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice","address":{"street":"123 Main St","city":null}}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	if user.Address == nil {
		t.Fatal("expected non-nil Address pointer")
	}

	if user.Address.Street == nil {
		t.Fatal("expected non-nil Street pointer")
	}
	if *user.Address.Street != "123 Main St" {
		t.Errorf("expected street '123 Main St', got %q", *user.Address.Street)
	}

	// City was explicitly null
	if user.Address.City != nil {
		t.Errorf("expected nil City pointer for explicit null, got %v", *user.Address.City)
	}
}
