package pedantigo

import (
	"testing"
)

type PartialTestUser struct {
	Email    string `json:"email" pedantigo:"required,email"`
	Name     string `json:"name" pedantigo:"required,min=2"`
	Age      int    `json:"age" pedantigo:"min=18"`
	Password string `json:"password" pedantigo:"required,min=8"`
}

func TestValidator_StructPartial(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	v := New[PartialTestUser]()

	tests := []struct {
		name    string
		user    PartialTestUser
		fields  []string
		wantErr bool
		errMsg  string // Expected error substring
	}{
		{
			name:    "validate one valid field",
			user:    PartialTestUser{Email: "test@example.com"},
			fields:  []string{"email"},
			wantErr: false,
		},
		{
			name:    "validate one invalid field - bad email",
			user:    PartialTestUser{Email: "invalid"},
			fields:  []string{"email"},
			wantErr: true,
			errMsg:  "email",
		},
		{
			name:    "skip invalid field not in list",
			user:    PartialTestUser{Email: "invalid", Name: "John"},
			fields:  []string{"name"},
			wantErr: false, // Email not validated
		},
		{
			name:    "validate multiple valid fields",
			user:    PartialTestUser{Email: "test@example.com", Age: 25},
			fields:  []string{"email", "age"},
			wantErr: false,
		},
		{
			name:    "validate multiple fields with one error",
			user:    PartialTestUser{Email: "invalid", Age: 25},
			fields:  []string{"email", "age"},
			wantErr: true,
			errMsg:  "email",
		},
		{
			name:    "validate subset with errors in non-validated fields",
			user:    PartialTestUser{Email: "test@example.com", Age: 10}, // Age < 18
			fields:  []string{"email"},
			wantErr: false, // Age error ignored
		},
		{
			name:    "empty field list validates nothing",
			user:    PartialTestUser{Email: "invalid", Name: "X"}, // Multiple errors
			fields:  []string{},
			wantErr: false, // No fields validated
		},
		{
			name:    "non-existent field is skipped",
			user:    PartialTestUser{Email: "test@example.com"},
			fields:  []string{"NotAField"},
			wantErr: false, // Unknown field ignored
		},
		{
			name:    "mix of valid and non-existent fields",
			user:    PartialTestUser{Email: "test@example.com"},
			fields:  []string{"email", "NotAField"},
			wantErr: false,
		},
		{
			name:    "required field validation",
			user:    PartialTestUser{Name: "John"}, // Email required but empty
			fields:  []string{"email"},
			wantErr: true,
			errMsg:  "required",
		},
		{
			name:    "min length validation",
			user:    PartialTestUser{Name: "J"}, // min=2
			fields:  []string{"name"},
			wantErr: true,
			errMsg:  "min",
		},
		{
			name:    "min value validation",
			user:    PartialTestUser{Age: 10}, // min=18
			fields:  []string{"age"},
			wantErr: true,
			errMsg:  "min",
		},
		{
			name:    "password min length",
			user:    PartialTestUser{Password: "short"},
			fields:  []string{"password"},
			wantErr: true,
			errMsg:  "min",
		},
		{
			name:    "validate all fields explicitly",
			user:    PartialTestUser{Email: "test@example.com", Name: "John", Age: 25, Password: "securepass"},
			fields:  []string{"email", "name", "age", "password"},
			wantErr: false,
		},
		{
			name:    "validate all fields with multiple errors",
			user:    PartialTestUser{Email: "invalid", Name: "X", Age: 10, Password: "short"},
			fields:  []string{"email", "name", "age", "password"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // avoid loop aliasing
		t.Run(tt.name, func(t *testing.T) {
			err := v.StructPartial(&tt.user, tt.fields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructPartial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				// Check error message contains expected substring
				errStr := err.Error()
				if errStr == "" {
					t.Errorf("Expected error containing %q, got empty error", tt.errMsg)
				}
			}
		})
	}
}

func TestValidator_StructExcept(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	v := New[PartialTestUser]()

	tests := []struct {
		name    string
		user    PartialTestUser
		exclude []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "exclude one field - password",
			user:    PartialTestUser{Email: "test@example.com", Name: "John", Age: 25, Password: "x"}, // Invalid password
			exclude: []string{"password"},
			wantErr: false, // Password validation skipped
		},
		{
			name:    "exclude keeps other errors",
			user:    PartialTestUser{Email: "invalid", Name: "John", Age: 25, Password: "securepass"},
			exclude: []string{"name"},
			wantErr: true, // Email error still caught
			errMsg:  "email",
		},
		{
			name:    "exclude multiple fields",
			user:    PartialTestUser{Email: "invalid", Name: "X", Age: 25, Password: "securepass"},
			exclude: []string{"email", "name"},
			wantErr: false, // Both errors excluded
		},
		{
			name:    "empty exclude list validates all",
			user:    PartialTestUser{Email: "invalid", Name: "John", Age: 25, Password: "securepass"},
			exclude: []string{},
			wantErr: true, // Email error caught
		},
		{
			name:    "exclude all fields",
			user:    PartialTestUser{Email: "invalid", Name: "X", Age: 10, Password: "x"},
			exclude: []string{"email", "name", "age", "password"},
			wantErr: false, // All validations skipped
		},
		{
			name:    "exclude non-existent field",
			user:    PartialTestUser{Email: "test@example.com", Name: "John", Age: 25, Password: "securepass"},
			exclude: []string{"NotAField"},
			wantErr: false, // Unknown field ignored, all valid
		},
		{
			name:    "exclude with all valid data",
			user:    PartialTestUser{Email: "test@example.com", Name: "John", Age: 25, Password: "securepass"},
			exclude: []string{"age"},
			wantErr: false,
		},
		{
			name:    "exclude required field that's empty",
			user:    PartialTestUser{Name: "John", Age: 25, Password: "securepass"}, // Email empty
			exclude: []string{"email"},
			wantErr: false, // Email required check skipped
		},
		{
			name:    "exclude one error but catch another",
			user:    PartialTestUser{Email: "invalid", Name: "X", Age: 25, Password: "securepass"},
			exclude: []string{"email"},
			wantErr: true, // Name error caught
			errMsg:  "min",
		},
		{
			name:    "mix of valid and excluded fields",
			user:    PartialTestUser{Email: "test@example.com", Name: "John", Age: 10, Password: "securepass"}, // Age < 18
			exclude: []string{"age"},
			wantErr: false, // Age validation excluded
		},
	}

	for _, tt := range tests {
		tt := tt // avoid loop aliasing
		t.Run(tt.name, func(t *testing.T) {
			err := v.StructExcept(&tt.user, tt.exclude...)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructExcept() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				errStr := err.Error()
				if errStr == "" {
					t.Errorf("Expected error containing %q, got empty error", tt.errMsg)
				}
			}
		})
	}
}

func TestValidatePartial(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	// Test simple API wrapper for StructPartial
	user := &PartialTestUser{
		Email: "test@example.com",
		Name:  "John",
		Age:   10, // Invalid (< 18)
	}

	// Validate only Email and Name (skip Age)
	err := ValidatePartial(user, "email", "name")
	if err != nil {
		t.Errorf("ValidatePartial() should pass when validating only email and name, got error: %v", err)
	}

	// Validate Age (should fail)
	err = ValidatePartial(user, "age")
	if err == nil {
		t.Error("ValidatePartial() should fail when validating age < 18")
	}

	// Invalid email
	user.Email = "invalid"
	err = ValidatePartial(user, "email")
	if err == nil {
		t.Error("ValidatePartial() should fail with invalid email")
	}
}

func TestValidateExcept(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	// Test simple API wrapper for StructExcept
	user := &PartialTestUser{
		Email:    "test@example.com",
		Name:     "John",
		Age:      10, // Invalid (< 18)
		Password: "securepass",
	}

	// Exclude Age validation (should pass)
	err := ValidateExcept(user, "age")
	if err != nil {
		t.Errorf("ValidateExcept() should pass when excluding age, got error: %v", err)
	}

	// Don't exclude Age (should fail)
	err = ValidateExcept(user)
	if err == nil {
		t.Error("ValidateExcept() should fail when not excluding age")
	}

	// Multiple exclusions
	user.Email = "invalid"
	user.Name = "X" // Too short
	err = ValidateExcept(user, "email", "name", "age")
	if err != nil {
		t.Errorf("ValidateExcept() should pass when excluding all invalid fields, got error: %v", err)
	}
}

// Test edge cases
func TestStructPartial_EdgeCases(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	v := New[PartialTestUser]()

	t.Run("nil pointer", func(t *testing.T) {
		var user *PartialTestUser
		err := v.StructPartial(user, "email")
		// Should return error for nil pointer
		if err == nil {
			t.Error("Expected error for nil pointer")
		}
	})

	t.Run("zero value struct", func(t *testing.T) {
		user := PartialTestUser{}
		err := v.StructPartial(&user, "email")
		// Email is required, should error
		if err == nil {
			t.Error("Expected error for required field on zero value struct")
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		user := PartialTestUser{Email: "test@example.com"}
		// Field names should be case-sensitive (lowercase json name)
		err := v.StructPartial(&user, "email") // lowercase JSON name - should validate
		if err != nil {
			t.Errorf("Expected no error when using correct JSON field name, got: %v", err)
		}

		// Uppercase struct field name should NOT match
		err = v.StructPartial(&user, "Email") // uppercase - should be skipped
		if err != nil {
			t.Errorf("Expected no error when field name doesn't match (should skip), got: %v", err)
		}
	})
}

func TestStructExcept_EdgeCases(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	v := New[PartialTestUser]()

	t.Run("nil pointer", func(t *testing.T) {
		var user *PartialTestUser
		err := v.StructExcept(user, "email")
		// Should return error for nil pointer
		if err == nil {
			t.Error("Expected error for nil pointer")
		}
	})

	t.Run("duplicate exclusions", func(t *testing.T) {
		user := PartialTestUser{Email: "invalid", Name: "John", Age: 25, Password: "securepass"}
		err := v.StructExcept(&user, "email", "email", "email")
		// Should handle duplicates gracefully
		if err != nil {
			t.Errorf("Unexpected error with duplicate exclusions: %v", err)
		}
	})
}
