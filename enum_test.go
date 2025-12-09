package pedantigo

import (
	"testing"
)

// ==================================================
// enum constraint tests
// ==================================================

func TestEnum(t *testing.T) {
	tests := []struct {
		name       string
		testType   string // "string_valid", "string_invalid", "int_valid", "int_invalid", "slice", "map", "schema"
		json       string
		expectErr  bool
		errorField string
		errorMsg   string
	}{
		// String enum tests
		{
			name:      "valid string enum",
			testType:  "string_valid",
			json:      `{"role":"admin"}`,
			expectErr: false,
		},
		{
			name:       "invalid string enum",
			testType:   "string_invalid",
			json:       `{"role":"superadmin"}`,
			expectErr:  true,
			errorField: "Role",
			errorMsg:   "must be one of: admin, user, guest",
		},
		// Integer enum tests
		{
			name:      "valid integer enum",
			testType:  "int_valid",
			json:      `{"code":200}`,
			expectErr: false,
		},
		{
			name:       "invalid integer enum",
			testType:   "int_invalid",
			json:       `{"code":404}`,
			expectErr:  true,
			errorField: "Code",
			errorMsg:   "must be one of: 200, 201, 204",
		},
		// Collection tests
		{
			name:       "enum in slice",
			testType:   "slice",
			json:       `{"roles":["admin","user","superadmin"]}`,
			expectErr:  true,
			errorField: "Roles[2]",
			errorMsg:   "must be one of: admin, user, guest",
		},
		{
			name:       "enum in map",
			testType:   "map",
			json:       `{"permissions":{"file":"read","script":"delete"}}`,
			expectErr:  true,
			errorField: "Permissions[script]",
			errorMsg:   "must be one of: read, write, execute",
		},
		// Schema test
		{
			name:      "schema generation",
			testType:  "schema",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case "string_valid":
				type User struct {
					Role string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				validator := New[User]()
				user, err := validator.Unmarshal([]byte(tt.json))
				if err != nil {
					t.Errorf("expected no errors for valid enum value, got %v", err)
				}
				if user.Role != "admin" {
					t.Errorf("expected role 'admin', got %s", user.Role)
				}

			case "string_invalid":
				type User struct {
					Role string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				validator := New[User]()
				_, err := validator.Unmarshal([]byte(tt.json))
				if err == nil {
					t.Error("expected validation error for invalid enum value")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				if !foundError {
					t.Errorf("expected error field=%s msg=%s, got %v", tt.errorField, tt.errorMsg, ve.Errors)
				}

			case "int_valid":
				type Status struct {
					Code int `json:"code" pedantigo:"oneof=200 201 204"`
				}
				validator := New[Status]()
				status, err := validator.Unmarshal([]byte(tt.json))
				if err != nil {
					t.Errorf("expected no errors for valid enum value, got %v", err)
				}
				if status.Code != 200 {
					t.Errorf("expected code 200, got %d", status.Code)
				}

			case "int_invalid":
				type Status struct {
					Code int `json:"code" pedantigo:"oneof=200 201 204"`
				}
				validator := New[Status]()
				_, err := validator.Unmarshal([]byte(tt.json))
				if err == nil {
					t.Error("expected validation error for invalid enum value")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				if !foundError {
					t.Errorf("expected error field=%s msg=%s, got %v", tt.errorField, tt.errorMsg, ve.Errors)
				}

			case "slice":
				type Config struct {
					Roles []string `json:"roles" pedantigo:"oneof=admin user guest"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if len(ve.Errors) != 1 {
					t.Errorf("expected 1 validation error, got %d: %v", len(ve.Errors), ve.Errors)
				}
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				if !foundError {
					t.Errorf("expected error at field=%s, got %v", tt.errorField, ve.Errors)
				}

			case "map":
				type Config struct {
					Permissions map[string]string `json:"permissions" pedantigo:"oneof=read write execute"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if len(ve.Errors) != 1 {
					t.Errorf("expected 1 validation error, got %d: %v", len(ve.Errors), ve.Errors)
				}
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				if !foundError {
					t.Errorf("expected error at field=%s, got %v", tt.errorField, ve.Errors)
				}

			case "schema":
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
		})
	}
}
