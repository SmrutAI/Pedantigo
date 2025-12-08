package pedantigo

import (
	"testing"
)

// ==================================================
// enum constraint tests
// ==================================================

func TestEnum_ValidString(t *testing.T) {
	type User struct {
		Role string `json:"role" pedantigo:"oneof=admin user guest"`
	}

	validator := New[User]()
	jsonData := []byte(`{"role":"admin"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid enum value, got %v", errs)
	}

	if user.Role != "admin" {
		t.Errorf("expected role 'admin', got %s", user.Role)
	}
}

func TestEnum_InvalidString(t *testing.T) {
	type User struct {
		Role string `json:"role" pedantigo:"oneof=admin user guest"`
	}

	validator := New[User]()
	jsonData := []byte(`{"role":"superadmin"}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for invalid enum value")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Role" && err.Message == "must be one of: admin, user, guest" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be one of' error, got %v", errs)
	}
}

func TestEnum_ValidInteger(t *testing.T) {
	type Status struct {
		Code int `json:"code" pedantigo:"oneof=200 201 204"`
	}

	validator := New[Status]()
	jsonData := []byte(`{"code":200}`)

	status, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid enum value, got %v", errs)
	}

	if status.Code != 200 {
		t.Errorf("expected code 200, got %d", status.Code)
	}
}

func TestEnum_InvalidInteger(t *testing.T) {
	type Status struct {
		Code int `json:"code" pedantigo:"oneof=200 201 204"`
	}

	validator := New[Status]()
	jsonData := []byte(`{"code":404}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for invalid enum value")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Code" && err.Message == "must be one of: 200, 201, 204" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be one of' error, got %v", errs)
	}
}

func TestEnum_InSlice(t *testing.T) {
	type Config struct {
		Roles []string `json:"roles" pedantigo:"oneof=admin user guest"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"roles":["admin","user","superadmin"]}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 1 {
		t.Errorf("expected 1 validation error, got %d: %v", len(errs), errs)
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Roles[2]" && err.Message == "must be one of: admin, user, guest" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Roles[2]', got %v", errs)
	}
}

func TestEnum_InMap(t *testing.T) {
	type Config struct {
		Permissions map[string]string `json:"permissions" pedantigo:"oneof=read write execute"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"permissions":{"file":"read","script":"delete"}}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) != 1 {
		t.Errorf("expected 1 validation error, got %d: %v", len(errs), errs)
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Permissions[script]" && err.Message == "must be one of: read, write, execute" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Permissions[script]', got %v", errs)
	}
}

func TestEnum_Schema(t *testing.T) {
	type User struct {
		Role string `json:"role" pedantigo:"oneof=admin user guest"`
	}

	validator := New[User]()
	schema := validator.Schema()

	roleProp, ok := schema.Properties.Get("role")
	if !ok || roleProp == nil {
		t.Fatal("expected 'role' property to exist")
	}

	if len(roleProp.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(roleProp.Enum))
	}

	expectedValues := map[string]bool{"admin": false, "user": false, "guest": false}
	for _, val := range roleProp.Enum {
		strVal, ok := val.(string)
		if !ok {
			t.Errorf("expected enum value to be string, got %T", val)
			continue
		}
		if _, exists := expectedValues[strVal]; exists {
			expectedValues[strVal] = true
		}
	}

	for val, found := range expectedValues {
		if !found {
			t.Errorf("expected enum value '%s' not found", val)
		}
	}
}
