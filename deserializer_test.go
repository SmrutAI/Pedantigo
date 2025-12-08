package pedantigo

import (
	"testing"
	"time"
)

// Test that missing fields with defaults get the default value
func TestDeserializer_MissingFieldWithDefault(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"required"`
		Port    int    `json:"port" pedantigo:"default=8080"`
		Timeout int    `json:"timeout" pedantigo:"default=30"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"name":"myapp"}`) // port and timeout are missing

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	// Missing fields should get their default values
	if config.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", config.Port)
	}

	if config.Timeout != 30 {
		t.Errorf("expected default timeout 30, got %d", config.Timeout)
	}

	if config.Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", config.Name)
	}
}

// Test that explicit zero values with defaults keep the zero (not the default)
func TestDeserializer_ExplicitZeroWithDefault(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"required"`
		Port    int    `json:"port" pedantigo:"default=8080"`
		Timeout int    `json:"timeout" pedantigo:"default=30"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"name":"myapp","port":0,"timeout":0}`) // explicit zeros

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	// Explicit zeros should be kept, NOT replaced with defaults
	if config.Port != 0 {
		t.Errorf("expected port 0 (not default), got %d", config.Port)
	}

	if config.Timeout != 0 {
		t.Errorf("expected timeout 0 (not default), got %d", config.Timeout)
	}
}

// Test that explicit false values pass required validation
func TestDeserializer_RequiredWithExplicitFalse(t *testing.T) {
	type Settings struct {
		Name   string `json:"name" pedantigo:"required"`
		Active bool   `json:"active" pedantigo:"required"`
	}

	validator := New[Settings]()
	jsonData := []byte(`{"name":"test","active":false}`) // explicit false

	settings, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors for explicit false, got %v", errs)
	}

	if settings == nil {
		t.Fatal("expected non-nil settings")
	}

	// Explicit false should pass validation
	if settings.Active != false {
		t.Errorf("expected active=false, got %v", settings.Active)
	}
}

// Test that missing required field fails validation
func TestDeserializer_MissingRequiredField(t *testing.T) {
	type Settings struct {
		Name   string `json:"name" pedantigo:"required"`
		Active bool   `json:"active" pedantigo:"required"`
	}

	validator := New[Settings]()
	jsonData := []byte(`{"name":"test"}`) // active is missing

	settings, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for missing required field")
	}

	// Should have error for active field
	foundError := false
	for _, err := range errs {
		if err.Field == "active" {
			foundError = true
			if err.Message != "is required" {
				t.Errorf("expected error message 'is required', got %q", err.Message)
			}
		}
	}

	if !foundError {
		t.Errorf("expected error for field 'active', got errors: %+v", errs)
	}

	if settings == nil {
		t.Error("expected non-nil settings even with validation errors")
	}
}

// Test type for defaultUsingMethod
type UserWithTimestamp struct {
	Email     string    `json:"email" pedantigo:"required"`
	Role      string    `json:"role" pedantigo:"default=user"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=SetCreationTime"`
}

// Method that provides dynamic default value
func (u *UserWithTimestamp) SetCreationTime() (time.Time, error) {
	// Return a fixed time for testing (not time.Now())
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), nil
}

// Test that defaultUsingMethod is called for missing fields
func TestDeserializer_DefaultUsingMethod(t *testing.T) {
	validator := New[UserWithTimestamp]()
	jsonData := []byte(`{"email":"test@example.com"}`) // created_at is missing

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	// Method should have been called to set created_at
	expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !user.CreatedAt.Equal(expectedTime) {
		t.Errorf("expected created_at to be %v, got %v", expectedTime, user.CreatedAt)
	}

	// Static default should also be applied
	if user.Role != "user" {
		t.Errorf("expected default role 'user', got %q", user.Role)
	}
}

// Test that defaultUsingMethod is NOT called for explicit values
func TestDeserializer_DefaultUsingMethod_NotCalledForExplicit(t *testing.T) {
	validator := New[UserWithTimestamp]()
	explicitTime := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)
	jsonData := []byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	// Explicit value should be kept, method should NOT be called
	if !user.CreatedAt.Equal(explicitTime) {
		t.Errorf("expected created_at to be %v (explicit), got %v", explicitTime, user.CreatedAt)
	}
}

// Test type with invalid method signature (should panic at New() time)
type InvalidMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=WrongSignature"`
}

// Wrong signature: returns only value, no error
func (i *InvalidMethodType) WrongSignature() time.Time {
	return time.Now()
}

// Test that invalid method signature panics at New() time (fail-fast)
func TestDeserializer_InvalidMethodSignature_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid method signature")
		}
	}()

	// This should panic because WrongSignature doesn't return (value, error)
	_ = New[InvalidMethodType]()
}

// Test type with non-existent method
type NonExistentMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=DoesNotExist"`
}

// Test that non-existent method panics at New() time (fail-fast)
func TestDeserializer_NonExistentMethod_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-existent method")
		}
	}()

	// This should panic because DoesNotExist method doesn't exist
	_ = New[NonExistentMethodType]()
}

// Test StrictMissingFields option (relaxed mode)
func TestDeserializer_RelaxedMode(t *testing.T) {
	type Config struct {
		Name string `json:"name" pedantigo:"required"`
		Port int    `json:"port" pedantigo:"required"`
	}

	// Create validator with StrictMissingFields=false (relaxed mode)
	validator := New[Config](ValidatorOptions{StrictMissingFields: false})
	jsonData := []byte(`{"name":"myapp"}`) // port is missing

	config, errs := validator.Unmarshal(jsonData)
	// In relaxed mode, missing required fields without defaults should NOT error
	if len(errs) != 0 {
		t.Errorf("expected no validation errors in relaxed mode, got %v", errs)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	// Port should be zero value
	if config.Port != 0 {
		t.Errorf("expected port 0 (zero value), got %d", config.Port)
	}
}

// Test StrictMissingFields option (strict mode - default)
func TestDeserializer_StrictMode(t *testing.T) {
	type Config struct {
		Name string `json:"name" pedantigo:"required"`
		Port int    `json:"port" pedantigo:"required"`
	}

	// Create validator with default options (StrictMissingFields=true)
	validator := New[Config]()
	jsonData := []byte(`{"name":"myapp"}`) // port is missing

	config, errs := validator.Unmarshal(jsonData)
	// In strict mode, missing required fields should error
	if len(errs) == 0 {
		t.Error("expected validation error for missing required field in strict mode")
	}

	// Should have error for port field
	foundError := false
	for _, err := range errs {
		if err.Field == "port" && err.Message == "is required" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'is required' error for field 'port', got errors: %v", errs)
	}

	if config == nil {
		t.Error("expected non-nil config even with validation errors")
	}
}
