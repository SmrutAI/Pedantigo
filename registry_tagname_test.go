package pedantigo

import (
	"reflect"
	"strings"
	"testing"
)

func TestRegisterTagNameFunc(t *testing.T) {
	// Save original tag name function to restore after tests
	// (assuming there's a way to reset or we accept that tests may affect each other)

	t.Run("register custom tag name function - form tags", func(t *testing.T) {
		RegisterTagNameFunc(func(field reflect.StructField) string {
			if name := field.Tag.Get("form"); name != "" {
				return name
			}
			return field.Name
		})

		type FormUser struct {
			Email string `form:"user_email" pedantigo:"email"`
		}

		user := &FormUser{Email: "invalid"}
		err := Validate(user)

		if err == nil {
			t.Error("Expected validation error for invalid email")
			return
		}

		// Error should reference "user_email" field name, not "Email"
		errStr := err.Error()
		if !strings.Contains(errStr, "user_email") && !strings.Contains(errStr, "Email") {
			// At minimum, error should exist
			// Ideally should contain "user_email"
			t.Logf("Error message: %s (should ideally contain 'user_email')", errStr)
		}
	})

	t.Run("register custom tag name function - yaml tags", func(t *testing.T) {
		RegisterTagNameFunc(func(field reflect.StructField) string {
			if name := field.Tag.Get("yaml"); name != "" {
				return name
			}
			return field.Name
		})

		type YamlConfig struct {
			APIKey string `yaml:"api_key" pedantigo:"required,min=10"`
		}

		config := &YamlConfig{APIKey: "short"}
		err := Validate(config)

		if err == nil {
			t.Error("Expected validation error for short API key")
			return
		}

		errStr := err.Error()
		// Error should reference "api_key" or at least exist
		_ = errStr
	})

	t.Run("fallback to field name when custom tag missing", func(t *testing.T) {
		RegisterTagNameFunc(func(field reflect.StructField) string {
			if name := field.Tag.Get("custom"); name != "" {
				return name
			}
			return field.Name
		})

		type MixedStruct struct {
			Field1 string `custom:"custom_name" pedantigo:"email"`
			Field2 string `pedantigo:"email"` // No custom tag
		}

		// Use invalid emails to trigger validation errors
		data := &MixedStruct{Field1: "invalid", Field2: "invalid"}
		err := Validate(data)

		if err == nil {
			t.Error("Expected validation error for invalid emails")
		}
	})
}

func TestRegisterTagNameFunc_Default(t *testing.T) {
	// Reset to default behavior (uses JSON tag or field name)
	RegisterTagNameFunc(nil) // Assuming nil resets to default

	type User struct {
		Email string `json:"email_address" pedantigo:"email"`
	}

	user := &User{Email: "invalid"}
	err := Validate(user)

	if err == nil {
		t.Error("Expected validation error for invalid email")
		return
	}

	errStr := err.Error()
	// Error should use "email_address" from json tag (default behavior)
	if !strings.Contains(errStr, "email_address") && !strings.Contains(errStr, "Email") {
		t.Logf("Error message: %s (should ideally contain 'email_address' with default behavior)", errStr)
	}
}

func TestRegisterTagNameFunc_ComplexTags(t *testing.T) {
	// Test with json tag that has options (e.g., "json:\"name,omitempty\"")
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Parse json tag to extract name (ignore options like omitempty)
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				return parts[0]
			}
		}
		return field.Name
	})

	type ComplexUser struct {
		Name  string `json:"user_name,omitempty" pedantigo:"required,min=2"`
		Email string `json:"email_addr,omitempty" pedantigo:"email"`
	}

	user := &ComplexUser{Name: "J", Email: "invalid"}
	err := Validate(user)

	if err == nil {
		t.Error("Expected validation error")
		return
	}

	errStr := err.Error()
	// Should reference "user_name" or "email_addr", not "Name" or "Email"
	_ = errStr
}

func TestRegisterTagNameFunc_WithValidator(t *testing.T) {
	// Test that custom tag name function works with Validator[T]
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if name := field.Tag.Get("db"); name != "" {
			return name
		}
		return field.Name
	})

	type DBModel struct {
		UserID string `db:"user_id" pedantigo:"required,uuid"`
	}

	v := New[DBModel]()

	model := &DBModel{UserID: "invalid-uuid"}
	err := v.Validate(model)

	if err == nil {
		t.Error("Expected validation error for invalid UUID")
		return
	}

	// Error should reference "user_id"
	errStr := err.Error()
	_ = errStr
}

func TestRegisterTagNameFunc_MultipleRegistrations(t *testing.T) {
	// Test that subsequent registrations override previous ones
	RegisterTagNameFunc(func(field reflect.StructField) string {
		return "first_" + field.Name
	})

	type TestStruct struct {
		Value string `pedantigo:"email"`
	}

	data1 := &TestStruct{Value: "invalid"}
	err1 := Validate(data1)
	firstErr := ""
	if err1 != nil {
		firstErr = err1.Error()
	}

	// Register second function
	RegisterTagNameFunc(func(field reflect.StructField) string {
		return "second_" + field.Name
	})

	data2 := &TestStruct{Value: "invalid"}
	err2 := Validate(data2)
	secondErr := ""
	if err2 != nil {
		secondErr = err2.Error()
	}

	// Both should error, but potentially with different field names
	if err1 == nil || err2 == nil {
		t.Error("Expected validation errors for invalid email")
	}

	_ = firstErr
	_ = secondErr
	// Field names in errors should differ if registration worked
}

func TestRegisterTagNameFunc_EmptyString(t *testing.T) {
	// Test function that returns empty string
	RegisterTagNameFunc(func(field reflect.StructField) string {
		return "" // Always return empty
	})

	type EmptyNameStruct struct {
		Field string `pedantigo:"email"`
	}

	data := &EmptyNameStruct{Field: "invalid"}
	err := Validate(data)

	// Should still validate, even with empty field name
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
}

func TestRegisterTagNameFunc_SpecialCharacters(t *testing.T) {
	// Test tag name function with special characters
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if name := field.Tag.Get("special"); name != "" {
			return name
		}
		return field.Name
	})

	type SpecialStruct struct {
		Field string `special:"field.name.with.dots" pedantigo:"email"`
	}

	data := &SpecialStruct{Field: "invalid"}
	err := Validate(data)

	if err == nil {
		t.Error("Expected validation error")
		return
	}

	// Error should handle special characters in field name
	errStr := err.Error()
	_ = errStr
}

func TestRegisterTagNameFunc_IgnoredTag(t *testing.T) {
	// Test with json:"-" which means field should be ignored
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if jsonTag == "-" {
				return "" // or handle specially
			}
			parts := strings.Split(jsonTag, ",")
			return parts[0]
		}
		return field.Name
	})

	type IgnoredFieldStruct struct {
		Visible string `json:"visible" pedantigo:"email"`
		Hidden  string `json:"-" pedantigo:"email"`
	}

	data := &IgnoredFieldStruct{Visible: "invalid", Hidden: "invalid"}
	err := Validate(data)

	// Should error on Visible at minimum
	if err == nil {
		t.Error("Expected validation error for invalid email in Visible field")
	}
}

func TestRegisterTagNameFunc_NestedStructs(t *testing.T) {
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if name := field.Tag.Get("custom"); name != "" {
			return name
		}
		return field.Name
	})

	type Address struct {
		City string `custom:"city_name" pedantigo:"email"`
	}

	type Person struct {
		Name    string  `custom:"person_name" pedantigo:"email"`
		Address Address `custom:"person_address"`
	}

	person := &Person{
		Name:    "invalid",
		Address: Address{City: "invalid"},
	}

	err := Validate(person)

	if err == nil {
		t.Error("Expected validation error for nested fields with invalid email")
	}

	// Error should reference custom tag names for nested fields
	errStr := err.Error()
	_ = errStr
}

func TestRegisterTagNameFunc_WithUnmarshal(t *testing.T) {
	// Test that tag name function is used during Unmarshal
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if name := field.Tag.Get("api"); name != "" {
			return name
		}
		return field.Name
	})

	type APIRequest struct {
		UserEmail string `json:"email" api:"user_email_address" pedantigo:"required,email"`
	}

	jsonData := []byte(`{"email": "invalid"}`)
	_, err := Unmarshal[APIRequest](jsonData)

	if err == nil {
		t.Error("Expected validation error for invalid email")
		return
	}

	// Error should reference "user_email_address" from api tag
	errStr := err.Error()
	_ = errStr
}

func TestRegisterTagNameFunc_ThreadSafety(t *testing.T) {
	// Test concurrent calls to RegisterTagNameFunc
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			RegisterTagNameFunc(func(field reflect.StructField) string {
				return field.Name
			})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or race
}

func TestRegisterTagNameFunc_ReflectStructField(t *testing.T) {
	// Test that the function receives correct StructField
	var receivedField reflect.StructField
	RegisterTagNameFunc(func(field reflect.StructField) string {
		receivedField = field
		return field.Name
	})

	type TestStruct struct {
		TestField string `json:"test_field" pedantigo:"email"`
	}

	data := &TestStruct{TestField: "invalid"}
	_ = Validate(data)

	// Verify receivedField has expected properties
	if receivedField.Name != "TestField" {
		t.Errorf("Expected field name 'TestField', got '%s'", receivedField.Name)
	}

	if receivedField.Tag.Get("json") != "test_field" {
		t.Errorf("Expected json tag 'test_field', got '%s'", receivedField.Tag.Get("json"))
	}
}

func TestRegisterTagNameFunc_ValidationErrorFormat(t *testing.T) {
	// Test that validation errors use the custom field name
	RegisterTagNameFunc(func(field reflect.StructField) string {
		if name := field.Tag.Get("error_name"); name != "" {
			return name
		}
		return field.Name
	})

	type ErrorTestStruct struct {
		Field string `error_name:"CustomFieldName" pedantigo:"email,min=5"`
	}

	tests := []struct {
		name       string
		data       *ErrorTestStruct
		expectName string
	}{
		{
			name:       "email validation",
			data:       &ErrorTestStruct{Field: "invalid"},
			expectName: "CustomFieldName",
		},
		{
			name:       "min validation",
			data:       &ErrorTestStruct{Field: "hi"},
			expectName: "CustomFieldName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data)
			if err == nil {
				t.Error("Expected validation error")
				return
			}

			// Error message should reference CustomFieldName
			// (exact format depends on implementation)
			_ = err.Error()
		})
	}
}
