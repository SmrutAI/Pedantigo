package pedantigo

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Core Validation Tests ====================
// NOTE: 'required' is only checked during Unmarshal (missing JSON keys), not Validate()
// Validate() only checks value constraints (min, max, email, etc.)

func TestValidator_Required_Present(t *testing.T) {
	type User struct {
		Email string `pedantigo:"required"`
	}

	validator := New[User]()
	user := &User{Email: "test@example.com"}

	err := validator.Validate(user)
	assert.NoError(t, err)
}

func TestValidator_Min_BelowMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 15}

	err := validator.Validate(user)
	assert.Error(t, err)
}

func TestValidator_Min_AtMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 18}

	err := validator.Validate(user)
	assert.NoError(t, err)
}

func TestValidator_Max_AboveMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 150}

	err := validator.Validate(user)
	assert.Error(t, err)
}

func TestValidator_Max_AtMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 120}

	err := validator.Validate(user)
	assert.NoError(t, err)
}

func TestValidator_MinMax_InRange(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{Age: 25}

	err := validator.Validate(user)
	assert.NoError(t, err)
}

// Test type for cross-field validation
type testPasswordChange struct {
	Password string `pedantigo:"required"`
	Confirm  string `pedantigo:"required"`
}

func (vpc *testPasswordChange) Validate() error {
	if vpc.Password != vpc.Confirm {
		return &ValidationError{
			Errors: []FieldError{{
				Field:   "Confirm",
				Message: "passwords do not match",
			}},
		}
	}
	return nil
}

func TestValidator_CrossField_PasswordConfirmation(t *testing.T) {
	validator := New[testPasswordChange]()
	pwd := &testPasswordChange{
		Password: "secret123",
		Confirm:  "different",
	}

	err := validator.Validate(pwd)
	assert.Error(t, err)

	// Should have cross-field error
	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	foundCrossFieldError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Confirm" && fieldErr.Message == "passwords do not match" {
			foundCrossFieldError = true
		}
	}

	assert.True(t, foundCrossFieldError, "expected cross-field validation error")
}

// TestMarshal_Valid verifies that Marshal returns JSON for valid structs
func TestMarshal_Valid(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	data, err := validator.Marshal(user)
	require.NoError(t, err)

	// Verify JSON is valid and contains expected fields
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
	assert.Equal(t, float64(25), result["age"])
}

// TestMarshal_Invalid verifies that Marshal returns validation errors for invalid structs
func TestMarshal_Invalid(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{
		Name:  "J",          // Too short (min=2)
		Email: "notanemail", // Invalid email
		Age:   15,           // Too young (min=18)
	}

	data, err := validator.Marshal(user)

	// Should return validation error, not JSON
	require.Error(t, err)
	assert.Nil(t, data)

	// Verify it's a ValidationError with multiple field errors
	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	assert.Len(t, ve.Errors, 3)

	// Check that errors are for the expected fields
	errorFields := make(map[string]bool)
	for _, fieldErr := range ve.Errors {
		errorFields[fieldErr.Field] = true
	}

	assert.True(t, errorFields["Name"], "expected validation error for Name field")
	assert.True(t, errorFields["Email"], "expected validation error for Email field")
	assert.True(t, errorFields["Age"], "expected validation error for Age field")
}

// TestMarshal_Nil verifies that Marshal handles nil pointer appropriately
func TestMarshal_Nil(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"min=2"`
	}

	validator := New[User]()

	// Pass nil pointer
	data, err := validator.Marshal(nil)

	// Should handle nil gracefully (either return error or marshal "null")
	// Let's check what actually happens - if Validate accepts nil, json.Marshal will return "null"
	// If Validate rejects nil, we'll get a validation error
	if err != nil {
		// Validation error is acceptable for nil
		t.Logf("Marshal(nil) returned error: %v", err)
	} else {
		assert.Equal(t, "null", string(data))
	}
}

// ==================== Unmarshal Tests ====================

func TestUnmarshal_ValidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com","age":25}`)

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, 25, user.Age)
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":}`) // Invalid JSON

	user, err := validator.Unmarshal(jsonData)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUnmarshal_ValidationError(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required,email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	// email is present but invalid (not an email), age is below min
	jsonData := []byte(`{"email":"notanemail","age":15}`)

	user, err := validator.Unmarshal(jsonData)

	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	t.Logf("Got %d errors:", len(ve.Errors))
	for _, fieldErr := range ve.Errors {
		t.Logf("  - %s: %s", fieldErr.Field, fieldErr.Message)
	}

	// Should still return the user struct even with validation errors
	assert.NotNil(t, user)

	// Check we have errors for both fields (use struct field names, not JSON names)
	foundEmailError := false
	foundAgeError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Email" {
			foundEmailError = true
		}
		if fieldErr.Field == "Age" {
			foundAgeError = true
		}
	}

	assert.True(t, foundEmailError, "expected validation error for Email field")
	assert.True(t, foundAgeError, "expected validation error for Age field")
}

func TestUnmarshal_DefaultValues(t *testing.T) {
	type User struct {
		Email  string `json:"email" pedantigo:"required"`
		Role   string `json:"role" pedantigo:"default=user"`
		Status string `json:"status" pedantigo:"default=active"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com"}`) // Missing role and status

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	require.NotNil(t, user)

	// Defaults should be applied
	assert.Equal(t, "user", user.Role)
	assert.Equal(t, "active", user.Status)
}

func TestUnmarshal_NestedValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required,min=1"` // min=1 for non-empty string
	}

	type User struct {
		Email   string  `json:"email" pedantigo:"required"`
		Address Address `json:"address"`
	}

	validator := New[User]()
	// City is present but empty - should fail min=1 constraint
	jsonData := []byte(`{"email":"test@example.com","address":{"city":""}}`)

	user, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	// Should have error for Address.City
	foundNestedError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Address.City" || fieldErr.Field == "City" {
			foundNestedError = true
		}
	}

	assert.True(t, foundNestedError, "expected validation error for nested City field")
	assert.NotNil(t, user)
}

// ==================== Pointer Tests ====================

// Test pointer field with explicit value
func TestPointer_ExplicitValue(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice","age":25}`)

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	require.NotNil(t, user.Name)
	assert.Equal(t, "Alice", *user.Name)

	require.NotNil(t, user.Age)
	assert.Equal(t, 25, *user.Age)
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

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	// Explicit zeros should create pointers to zero values
	require.NotNil(t, config.Port)
	assert.Equal(t, 0, *config.Port)

	require.NotNil(t, config.Enabled)
	assert.Equal(t, false, *config.Enabled)

	require.NotNil(t, config.Name)
	assert.Equal(t, "", *config.Name)
}

// Test pointer field with explicit null (should be nil pointer)
func TestPointer_ExplicitNull(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null,"age":null}`)

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	// Explicit null should result in nil pointers
	assert.Nil(t, user.Name)
	assert.Nil(t, user.Age)
}

// Test pointer field missing from JSON (should be nil pointer)
func TestPointer_Missing(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Both fields missing

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	// Missing fields should result in nil pointers
	assert.Nil(t, user.Name)
	assert.Nil(t, user.Age)
}

// Test required pointer field with explicit value
func TestPointer_RequiredWithValue(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice"}`)

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	require.NotNil(t, user.Name)
	assert.Equal(t, "Alice", *user.Name)
}

// Test required pointer field missing (should fail)
func TestPointer_RequiredMissing(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Missing required field

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	// Check for required field error
	foundRequiredError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "name" && fieldErr.Message == "is required" {
			foundRequiredError = true
		}
	}

	assert.True(t, foundRequiredError, "expected 'is required' error for name field")
}

// Test required pointer field with explicit null (should pass - field is present)
func TestPointer_RequiredWithNull(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null}`) // Field present but null

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	// Required means "field must be present", not "value can't be nil"
	assert.Nil(t, user.Name)
}

// Test pointer field with default value
func TestPointer_WithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{}`) // Missing port field

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	// Default should be applied to missing field
	require.NotNil(t, config.Port)
	assert.Equal(t, 8080, *config.Port)
}

// Test pointer field with explicit zero and default (should keep zero)
func TestPointer_ExplicitZeroWithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0}`) // Explicit zero

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	// Explicit zero should be kept, not replaced with default
	require.NotNil(t, config.Port)
	assert.Equal(t, 0, *config.Port)
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

	user, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)

	require.NotNil(t, user.Address)

	require.NotNil(t, user.Address.Street)
	assert.Equal(t, "123 Main St", *user.Address.Street)

	// City was explicitly null
	assert.Nil(t, user.Address.City)
}

// ==================== Deserializer Tests ====================

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

// Test type with invalid method signature (should panic at New() time)
type InvalidMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=WrongSignature"`
}

// Wrong signature: returns only value, no error
func (i *InvalidMethodType) WrongSignature() time.Time {
	return time.Now()
}

// Test type with non-existent method
type NonExistentMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=DoesNotExist"`
}

// TestDeserializer_UnmarshalBehavior validates deserializer behavior across various scenarios:
// defaults, missing fields, explicit values, required fields, and validator options.
func TestDeserializer_UnmarshalBehavior(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"required"`
		Port    int    `json:"port" pedantigo:"default=8080"`
		Timeout int    `json:"timeout" pedantigo:"default=30"`
	}

	type Settings struct {
		Name   string `json:"name" pedantigo:"required"`
		Active bool   `json:"active" pedantigo:"required"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		validatorFn func() (any, error)
		wantErr     bool
		assertions  func(*testing.T, any)
	}{
		{
			name:     "missing fields with defaults",
			jsonData: []byte(`{"name":"myapp"}`),
			validatorFn: func() (any, error) {
				v := New[Config]()
				return v.Unmarshal([]byte(`{"name":"myapp"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				assert.Equal(t, 8080, config.Port)
				assert.Equal(t, 30, config.Timeout)
				assert.Equal(t, "myapp", config.Name)
			},
		},
		{
			name:     "explicit zero values not replaced with defaults",
			jsonData: []byte(`{"name":"myapp","port":0,"timeout":0}`),
			validatorFn: func() (any, error) {
				v := New[Config]()
				return v.Unmarshal([]byte(`{"name":"myapp","port":0,"timeout":0}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				// Explicit zeros should be kept, NOT replaced with defaults
				assert.Equal(t, 0, config.Port)
				assert.Equal(t, 0, config.Timeout)
			},
		},
		{
			name:     "explicit false value passes required validation",
			jsonData: []byte(`{"name":"test","active":false}`),
			validatorFn: func() (any, error) {
				v := New[Settings]()
				return v.Unmarshal([]byte(`{"name":"test","active":false}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				settings := result.(*Settings)
				assert.False(t, settings.Active)
			},
		},
		{
			name:     "missing required field fails validation",
			jsonData: []byte(`{"name":"test"}`),
			validatorFn: func() (any, error) {
				v := New[Settings]()
				return v.Unmarshal([]byte(`{"name":"test"}`))
			},
			wantErr: true,
			assertions: func(t *testing.T, result any) {
				// Error case - check error message through direct validation
				v := New[Settings]()
				_, err := v.Unmarshal([]byte(`{"name":"test"}`))
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "active" && fieldErr.Message == "is required" {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected 'is required' error for field 'active', got errors: %+v", ve.Errors)
			},
		},
		{
			name:     "defaultUsingMethod called for missing fields",
			jsonData: []byte(`{"email":"test@example.com"}`),
			validatorFn: func() (any, error) {
				v := New[UserWithTimestamp]()
				return v.Unmarshal([]byte(`{"email":"test@example.com"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				user := result.(*UserWithTimestamp)
				expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				assert.True(t, user.CreatedAt.Equal(expectedTime), "expected created_at to be %v, got %v", expectedTime, user.CreatedAt)
				assert.Equal(t, "user", user.Role)
			},
		},
		{
			name:     "defaultUsingMethod not called for explicit values",
			jsonData: []byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`),
			validatorFn: func() (any, error) {
				v := New[UserWithTimestamp]()
				return v.Unmarshal([]byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				user := result.(*UserWithTimestamp)
				explicitTime := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)
				assert.True(t, user.CreatedAt.Equal(explicitTime), "expected created_at to be %v (explicit), got %v", explicitTime, user.CreatedAt)
			},
		},
		{
			name:     "strict mode applies defaults for missing fields",
			jsonData: []byte(`{"name":"myapp"}`),
			validatorFn: func() (any, error) {
				v := New[Config](ValidatorOptions{StrictMissingFields: true})
				return v.Unmarshal([]byte(`{"name":"myapp"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				assert.Equal(t, 8080, config.Port)
				assert.Equal(t, 30, config.Timeout)
			},
		},
		{
			name:     "relaxed mode skips constraints on zero values",
			jsonData: []byte(`{"name":"myapp","age":0}`),
			validatorFn: func() (any, error) {
				type Profile struct {
					Name string `json:"name" pedantigo:"required"`
					Age  int    `json:"age" pedantigo:"min=1"`
				}
				v := New[Profile](ValidatorOptions{StrictMissingFields: false})
				return v.Unmarshal([]byte(`{"name":"myapp","age":0}`))
			},
			wantErr: true,
			assertions: func(t *testing.T, result any) {
				// Age=0 is explicit, so validation should still run and fail
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.validatorFn()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if result != nil || !tt.wantErr {
				tt.assertions(t, result)
			}
		})
	}
}

// TestDeserializer_ValidatorSetup validates fail-fast validation during New().
// Invalid method signatures or non-existent methods should panic at validator creation time.
func TestDeserializer_ValidatorSetup(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectPanic bool
	}{
		{
			name: "invalid method signature panics",
			setup: func() {
				_ = New[InvalidMethodType]()
			},
			expectPanic: true,
		},
		{
			name: "non-existent method panics",
			setup: func() {
				_ = New[NonExistentMethodType]()
			},
			expectPanic: true,
		},
		{
			name: "valid method signature succeeds",
			setup: func() {
				_ = New[UserWithTimestamp]()
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					r := recover()
					assert.NotNil(t, r, "expected panic but none occurred")
				}()
				tt.setup()
			} else {
				// Should not panic
				defer func() {
					r := recover()
					assert.Nil(t, r, "unexpected panic: %v", r)
				}()
				tt.setup()
			}
		})
	}
}

// ==================== Validator Options Tests ====================

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
				assert.Equal(t, "John", user.Name)
				assert.Equal(t, "john@example.com", user.Email)
				assert.Equal(t, 25, user.Age)
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

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if err != nil && tt.expectErrFields != nil {
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)
				assert.Len(t, ve.Errors, len(tt.expectErrFields))
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
				assert.Nil(t, settings.Port)
				assert.Nil(t, settings.Enabled)
				// Name should have zero value ""
				assert.Equal(t, "", settings.Name)
			},
		},
		{
			name:      "pointer_fields_with_valid_values",
			jsonInput: `{"port":8080,"enabled":true,"name":"John"}`,
			expectErr: false,
			checkValues: func(t *testing.T, settings *Settings) {
				require.NotNil(t, settings.Port)
				assert.Equal(t, 8080, *settings.Port)
				require.NotNil(t, settings.Enabled)
				assert.True(t, *settings.Enabled)
				assert.Equal(t, "John", settings.Name)
			},
		},
		{
			name:            "pointer_field_invalid_value",
			jsonInput:       `{"port":80}`,
			expectErr:       true,
			expectErrFields: []string{"Port", "Name"}, // Port too low, Name missing/empty
			checkValues: func(t *testing.T, settings *Settings) {
				// Pointer should still be set even with validation error
				require.NotNil(t, settings.Port)
				assert.Equal(t, 80, *settings.Port)
			},
		},
		{
			name:      "pointer_fields_partial_missing",
			jsonInput: `{"port":2048,"name":"Alice"}`,
			expectErr: false,
			checkValues: func(t *testing.T, settings *Settings) {
				require.NotNil(t, settings.Port)
				assert.Equal(t, 2048, *settings.Port)
				assert.Nil(t, settings.Enabled)
				assert.Equal(t, "Alice", settings.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Settings](ValidatorOptions{
				StrictMissingFields: false,
			})

			settings, err := validator.Unmarshal([]byte(tt.jsonInput))

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if err != nil && tt.expectErr {
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)
				assert.Len(t, ve.Errors, len(tt.expectErrFields))
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
				r := recover()
				require.NotNil(t, r, "expected panic but didn't panic")
				panicMsg := r.(string)
				// Verify all expected strings are in panic message
				for _, expectedStr := range tt.expectPanicStrs {
					assert.True(t, strings.Contains(panicMsg, expectedStr), "panic message missing '%s', got: %s", expectedStr, panicMsg)
				}
				// Verify at least one expected field is mentioned
				foundField := false
				for _, expectedField := range tt.expectPanicFields {
					if strings.Contains(panicMsg, expectedField) {
						foundField = true
						break
					}
				}
				assert.True(t, foundField, "panic message should mention one of %v, got: %s", tt.expectPanicFields, panicMsg)
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
