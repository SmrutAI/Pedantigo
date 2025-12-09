package pedantigo

import (
	"strings"
	"testing"
)

// TestValidatorOptions_StrictMissingFields tests the StrictMissingFields behavior
// with various configuration combinations and JSON inputs.
func TestValidatorOptions_StrictMissingFields(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"required,min=2"`
		Email string `json:"email" pedantigo:"required,email"`
		Age   int    `json:"age" pedantigo:"required,min=18"`
	}

	tests := []struct {
		name            string
		strictMode      bool
		jsonInput       string
		expectErr       bool
		expectErrFields []string                       // Expected field names in ValidationError
		checkValues     func(t *testing.T, user *User) // Verify parsed values
	}{
		{
			name:       "StrictMissingFields_false_valid_values",
			strictMode: false,
			jsonInput:  `{"name":"John","email":"john@example.com","age":25}`,
			expectErr:  false,
			checkValues: func(t *testing.T, user *User) {
				if user.Name != "John" {
					t.Errorf("expected Name='John', got %q", user.Name)
				}
				if user.Email != "john@example.com" {
					t.Errorf("expected Email='john@example.com', got %q", user.Email)
				}
				if user.Age != 25 {
					t.Errorf("expected Age=25, got %d", user.Age)
				}
			},
		},
		{
			name:            "StrictMissingFields_false_zero_values_fail_min",
			strictMode:      false,
			jsonInput:       `{}`,
			expectErr:       true,
			expectErrFields: []string{"Name", "Age"},
			checkValues:     nil, // Error checking happens in test loop
		},
		{
			name:            "StrictMissingFields_false_invalid_email_and_age",
			strictMode:      false,
			jsonInput:       `{"email":"notanemail","age":15}`,
			expectErr:       true,
			expectErrFields: []string{"Name", "Email", "Age"}, // Name missing (zero value "") also fails min=2
		},
		{
			name:            "StrictMissingFields_true_required_field_missing",
			strictMode:      true,
			jsonInput:       `{}`,
			expectErr:       true,
			expectErrFields: []string{"name", "email", "age"}, // All required fields fail when missing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[User](ValidatorOptions{
				StrictMissingFields: tt.strictMode,
			})
			user, err := validator.Unmarshal([]byte(tt.jsonInput))

			if (err != nil) != tt.expectErr {
				t.Errorf("expectErr=%v, got err=%v", tt.expectErr, err)
			}

			if err != nil && tt.expectErrFields != nil {
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if len(ve.Errors) != len(tt.expectErrFields) {
					t.Errorf("expected %d errors, got %d: %v", len(tt.expectErrFields), len(ve.Errors), ve.Errors)
				}
			}

			if !tt.expectErr && tt.checkValues != nil {
				tt.checkValues(t, user)
			}
		})
	}
}

// TestValidatorOptions_PointerFields tests pointer field behavior with StrictMissingFields=false.
// Pointers to primitive types allow optional fields (nil when missing) while still validating when present.
func TestValidatorOptions_PointerFields(t *testing.T) {
	type Settings struct {
		Port    *int   `json:"port" pedantigo:"min=1024"`
		Enabled *bool  `json:"enabled"`
		Name    string `json:"name" pedantigo:"min=3"`
	}

	tests := []struct {
		name            string
		jsonInput       string
		expectErr       bool
		expectErrFields []string // Expected field names in ValidationError
		checkValues     func(t *testing.T, settings *Settings)
	}{
		{
			name:            "pointer_fields_all_missing",
			jsonInput:       `{}`,
			expectErr:       true,
			expectErrFields: []string{"Name"}, // Only Name should error (non-pointer zero value)
			checkValues: func(t *testing.T, settings *Settings) {
				// Port and Enabled should be nil (pointers skip validation when missing)
				if settings.Port != nil {
					t.Errorf("expected Port to be nil, got %v", *settings.Port)
				}
				if settings.Enabled != nil {
					t.Errorf("expected Enabled to be nil, got %v", *settings.Enabled)
				}
				// Name should have zero value ""
				if settings.Name != "" {
					t.Errorf("expected Name to be empty string, got %q", settings.Name)
				}
			},
		},
		{
			name:      "pointer_fields_with_valid_values",
			jsonInput: `{"port":8080,"enabled":true,"name":"John"}`,
			expectErr: false,
			checkValues: func(t *testing.T, settings *Settings) {
				if settings.Port == nil || *settings.Port != 8080 {
					t.Errorf("expected Port=8080, got %v", settings.Port)
				}
				if settings.Enabled == nil || *settings.Enabled != true {
					t.Errorf("expected Enabled=true, got %v", settings.Enabled)
				}
				if settings.Name != "John" {
					t.Errorf("expected Name='John', got %q", settings.Name)
				}
			},
		},
		{
			name:            "pointer_field_invalid_value",
			jsonInput:       `{"port":80}`,
			expectErr:       true,
			expectErrFields: []string{"Port", "Name"}, // Port too low, Name missing/empty
			checkValues: func(t *testing.T, settings *Settings) {
				// Pointer should still be set even with validation error
				if settings.Port == nil || *settings.Port != 80 {
					t.Errorf("expected Port=80 (even with error), got %v", settings.Port)
				}
			},
		},
		{
			name:      "pointer_fields_partial_missing",
			jsonInput: `{"port":2048,"name":"Alice"}`,
			expectErr: false,
			checkValues: func(t *testing.T, settings *Settings) {
				if settings.Port == nil || *settings.Port != 2048 {
					t.Errorf("expected Port=2048, got %v", settings.Port)
				}
				if settings.Enabled != nil {
					t.Errorf("expected Enabled to be nil, got %v", *settings.Enabled)
				}
				if settings.Name != "Alice" {
					t.Errorf("expected Name='Alice', got %q", settings.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Settings](ValidatorOptions{
				StrictMissingFields: false,
			})

			settings, err := validator.Unmarshal([]byte(tt.jsonInput))

			if (err != nil) != tt.expectErr {
				t.Errorf("expectErr=%v, got err=%v", tt.expectErr, err)
			}

			if err != nil && tt.expectErr {
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if len(ve.Errors) != len(tt.expectErrFields) {
					t.Errorf("expected %d errors, got %d: %v", len(tt.expectErrFields), len(ve.Errors), ve.Errors)
				}
			}

			if tt.checkValues != nil {
				tt.checkValues(t, settings)
			}
		})
	}
}

// TestValidatorOptions_PanicOnIncompatibleTags tests that creating a validator
// with StrictMissingFields=false and default/defaultUsingMethod tags panics.
// These combinations are incompatible because defaults only make sense when
// StrictMissingFields=true (missing field handling is disabled).
func TestValidatorOptions_PanicOnIncompatibleTags(t *testing.T) {
	tests := []struct {
		name              string
		testCase          string   // discriminator for which struct to use
		expectPanicFields []string // Expected field names in panic message
		expectPanicStrs   []string // Expected strings in panic message
	}{
		{
			name:              "panic_on_default_tag",
			testCase:          "single_default",
			expectPanicFields: []string{"Theme"},
			expectPanicStrs:   []string{"default=", "StrictMissingFields is false"},
		},
		{
			name:              "panic_on_multiple_default_tags",
			testCase:          "multiple_defaults",
			expectPanicFields: []string{"Name", "Port", "Enabled"},
			expectPanicStrs:   []string{"default=", "StrictMissingFields is false"},
		},
		{
			name:              "panic_on_defaultUsingMethod_tag",
			testCase:          "default_using_method",
			expectPanicFields: []string{"ID"},
			expectPanicStrs:   []string{"defaultUsingMethod=", "StrictMissingFields is false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic but didn't panic")
				} else {
					panicMsg := r.(string)
					// Verify all expected strings are in panic message
					for _, expectedStr := range tt.expectPanicStrs {
						if !strings.Contains(panicMsg, expectedStr) {
							t.Errorf("panic message missing '%s', got: %s", expectedStr, panicMsg)
						}
					}
					// Verify at least one expected field is mentioned
					foundField := false
					for _, expectedField := range tt.expectPanicFields {
						if strings.Contains(panicMsg, expectedField) {
							foundField = true
							break
						}
					}
					if !foundField {
						t.Errorf("panic message should mention one of %v, got: %s", tt.expectPanicFields, panicMsg)
					}
				}
			}()

			// Use testCase discriminator to handle different struct types
			switch tt.testCase {
			case "single_default":
				type Settings struct {
					Theme    string `json:"theme" pedantigo:"default=dark"`
					Language string `json:"language"`
				}
				_ = New[Settings](ValidatorOptions{
					StrictMissingFields: false,
				})

			case "multiple_defaults":
				type Config struct {
					Name    string `json:"name" pedantigo:"default=unnamed"`
					Port    int    `json:"port" pedantigo:"default=8080"`
					Enabled bool   `json:"enabled" pedantigo:"default=true"`
				}
				_ = New[Config](ValidatorOptions{
					StrictMissingFields: false,
				})

			case "default_using_method":
				type Product struct {
					ID   string `json:"id" pedantigo:"defaultUsingMethod=GenerateID"`
					Name string `json:"name"`
				}
				_ = New[Product](ValidatorOptions{
					StrictMissingFields: false,
				})
			}
		})
	}
}
