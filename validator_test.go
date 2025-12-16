package pedantigo

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test message constants.
const testMsgIsRequired = "is required"

// ==================== Core Validation Tests ====================
// NOTE: 'required' is only checked during Unmarshal (missing JSON keys), not Validate()
// Validate() only checks value constraints (min, max, email, etc.)

// TestValidator_Required_Present tests Validator required present.
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

// Test type for cross-field validation.
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
	require.Error(t, err)

	// Should have cross-field error
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	foundCrossFieldError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Confirm" && fieldErr.Message == "passwords do not match" {
			foundCrossFieldError = true
		}
	}

	assert.True(t, foundCrossFieldError, "expected cross-field validation error")
}

// TestMarshal_Valid verifies that Marshal returns JSON for valid structs.
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
	assert.InDelta(t, float64(25), result["age"], 1e-9)
}

// TestMarshal_Invalid verifies that Marshal returns validation errors for invalid structs.
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
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

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

// TestMarshal_Nil verifies that Marshal handles nil pointer appropriately.
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

// TestUnmarshal_ValidJSON tests Unmarshal validjson.
func TestUnmarshal_ValidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com","age":25}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)
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
	require.Error(t, err)
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

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

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
	require.NoError(t, err)
	require.NotNil(t, user)

	// Defaults should be applied.
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

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

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
// TestPointer_ExplicitValue tests Pointer explicitvalue.
func TestPointer_ExplicitValue(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice","age":25}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Name)
	assert.Equal(t, "Alice", *user.Name)

	require.NotNil(t, user.Age)
	assert.Equal(t, 25, *user.Age)
}

// Test pointer field with explicit zero value (should create pointer to zero)
// TestPointer_ExplicitZero tests Pointer explicitzero.
func TestPointer_ExplicitZero(t *testing.T) {
	type Config struct {
		Port    *int    `json:"port"`
		Enabled *bool   `json:"enabled"`
		Name    *string `json:"name"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0,"enabled":false,"name":""}`)

	config, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	// Explicit zeros should create pointers to zero values
	require.NotNil(t, config.Port)
	assert.Equal(t, 0, *config.Port)

	require.NotNil(t, config.Enabled)
	assert.False(t, *config.Enabled)

	require.NotNil(t, config.Name)
	assert.Empty(t, *config.Name)
}

// Test pointer field with explicit null (should be nil pointer)
// TestPointer_ExplicitNull tests Pointer explicitnull.
func TestPointer_ExplicitNull(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null,"age":null}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	// Explicit null should result in nil pointers.
	assert.Nil(t, user.Name)
	assert.Nil(t, user.Age)
}

// Test pointer field missing from JSON (should be nil pointer)
// TestPointer_Missing tests Pointer missing.
func TestPointer_Missing(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Both fields missing

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	// Missing fields should result in nil pointers
	assert.Nil(t, user.Name)
	assert.Nil(t, user.Age)
}

// Test required pointer field with explicit value
// TestPointer_RequiredWithValue tests Pointer requiredwithvalue.
func TestPointer_RequiredWithValue(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice"}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Name)
	assert.Equal(t, "Alice", *user.Name)
}

// Test required pointer field missing (should fail)
// TestPointer_RequiredMissing tests Pointer requiredmissing.
func TestPointer_RequiredMissing(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Missing required field

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	// Check for required field error
	foundRequiredError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "name" && fieldErr.Message == testMsgIsRequired {
			foundRequiredError = true
		}
	}

	assert.True(t, foundRequiredError, "expected 'is required' error for name field")
}

// Test required pointer field with explicit null (should pass - field is present)
// TestPointer_RequiredWithNull tests Pointer requiredwithnull.
func TestPointer_RequiredWithNull(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null}`) // Field present but null

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	// Required means "field must be present", not "value can't be nil"
	assert.Nil(t, user.Name)
}

// Test pointer field with default value
// TestPointer_WithDefault tests Pointer withdefault.
func TestPointer_WithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{}`) // Missing port field

	config, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	// Default should be applied to missing field
	require.NotNil(t, config.Port)
	assert.Equal(t, 8080, *config.Port)
}

// Test pointer field with explicit zero and default (should keep zero)
// TestPointer_ExplicitZeroWithDefault tests Pointer explicitzerowithdefault.
func TestPointer_ExplicitZeroWithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0}`) // Explicit zero

	config, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	// Explicit zero should be kept, not replaced with default
	require.NotNil(t, config.Port)
	assert.Equal(t, 0, *config.Port)
}

// Test nested struct with pointer fields
// TestPointer_NestedStruct tests Pointer nestedstruct.
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
	require.NoError(t, err)

	require.NotNil(t, user.Address)

	require.NotNil(t, user.Address.Street)
	assert.Equal(t, "123 Main St", *user.Address.Street)

	// City was explicitly null
	assert.Nil(t, user.Address.City)
}

// ==================== Deserializer Tests ====================

// Test type for defaultUsingMethod
// UserWithTimestamp represents the data structure.
type UserWithTimestamp struct {
	Email     string    `json:"email" pedantigo:"required"`
	Role      string    `json:"role" pedantigo:"default=user"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=SetCreationTime"`
}

// Method that provides dynamic default value
// SetCreationTime sets the field value.
func (u *UserWithTimestamp) SetCreationTime() (time.Time, error) {
	// Return a fixed time for testing (not time.Now())
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), nil
}

// Test type with invalid method signature (should panic at New() time)
// InvalidMethodType represents the data structure.
type InvalidMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=WrongSignature"`
}

// Wrong signature: returns only value, no error
// WrongSignature implements the method.
func (i *InvalidMethodType) WrongSignature() time.Time {
	return time.Now()
}

// Test type with non-existent method
// NonExistentMethodType represents the data structure.
type NonExistentMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=DoesNotExist"`
}

// TestDeserializer_UnmarshalBehavior validates deserializer behavior across various scenarios:
// defaults, missing fields, explicit values, required fields, and validator options.
// TestDeserializer_UnmarshalBehavior tests Deserializer unmarshalbehavior.
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
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "active" && fieldErr.Message == testMsgIsRequired {
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if result != nil || !tt.wantErr {
				tt.assertions(t, result)
			}
		})
	}
}

// TestDeserializer_ValidatorSetup validates fail-fast validation during New().
// Invalid method signatures or non-existent methods should panic at validator creation time.
// TestDeserializer_ValidatorSetup tests Deserializer validatorsetup.
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
// TestValidatorOptions_StrictMissingFields tests ValidatorOptions strictmissingfields.
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if err != nil && tt.expectErrFields != nil {
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
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
// TestValidatorOptions_PointerFields tests ValidatorOptions pointerfields.
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
				assert.Empty(t, settings.Name)
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if err != nil && tt.expectErr {
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
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
// TestValidatorOptions_PanicOnIncompatibleTags tests ValidatorOptions paniconincompatibletags.
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
					assert.Contains(t, panicMsg, expectedStr, "panic message missing '%s', got: %s", expectedStr, panicMsg)
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

// ==================== Dict Tests ====================

// TestValidator_Dict tests the Dict() method with various struct types and field combinations.
func TestValidator_Dict(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type User struct {
		Name    string   `json:"name"`
		Email   string   `json:"email"`
		Age     int      `json:"age"`
		Active  bool     `json:"active"`
		Tags    []string `json:"tags"`
		Address Address  `json:"address"`
	}

	type ConfigWithPointers struct {
		Port    *int    `json:"port"`
		Enabled *bool   `json:"enabled"`
		Name    *string `json:"name"`
	}

	tests := []struct {
		name     string
		obj      any
		expected map[string]interface{}
	}{
		{
			name: "basic struct with primitives",
			obj: &User{
				Name:   "Alice",
				Email:  "alice@example.com",
				Age:    25,
				Active: true,
				Tags:   []string{"admin", "developer"},
				Address: Address{
					Street: "123 Main St",
					City:   "Springfield",
				},
			},
			expected: map[string]interface{}{
				"name":   "Alice",
				"email":  "alice@example.com",
				"age":    float64(25),
				"active": true,
				"tags":   []interface{}{"admin", "developer"},
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "Springfield",
				},
			},
		},
		{
			name: "struct with zero values",
			obj: &User{
				Name:   "",
				Email:  "",
				Age:    0,
				Active: false,
				Tags:   nil,
				Address: Address{
					Street: "",
					City:   "",
				},
			},
			expected: map[string]interface{}{
				"name":   "",
				"email":  "",
				"age":    float64(0),
				"active": false,
				"tags":   nil,
				"address": map[string]interface{}{
					"street": "",
					"city":   "",
				},
			},
		},
		{
			name: "struct with pointer fields - all nil",
			obj: &ConfigWithPointers{
				Port:    nil,
				Enabled: nil,
				Name:    nil,
			},
			expected: map[string]interface{}{
				"port":    nil,
				"enabled": nil,
				"name":    nil,
			},
		},
		{
			name: "struct with pointer fields - all set",
			obj: func() *ConfigWithPointers {
				port := 8080
				enabled := true
				name := "myapp"
				return &ConfigWithPointers{
					Port:    &port,
					Enabled: &enabled,
					Name:    &name,
				}
			}(),
			expected: map[string]interface{}{
				"port":    float64(8080),
				"enabled": true,
				"name":    "myapp",
			},
		},
		{
			name: "struct with mixed nil and non-nil pointers",
			obj: func() *ConfigWithPointers {
				port := 3000
				name := "service"
				return &ConfigWithPointers{
					Port:    &port,
					Enabled: nil,
					Name:    &name,
				}
			}(),
			expected: map[string]interface{}{
				"port":    float64(3000),
				"enabled": nil,
				"name":    "service",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dict map[string]interface{}
			var err error

			// Use type assertion to call Dict with correct type
			switch v := tt.obj.(type) {
			case *User:
				validator := New[User]()
				dict, err = validator.Dict(v)
			case *ConfigWithPointers:
				validator := New[ConfigWithPointers]()
				dict, err = validator.Dict(v)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, dict)
		})
	}
}

// ErrIntentionalMarshalFailure is a test error for Dict error handling.
var ErrIntentionalMarshalFailure = errors.New("intentional marshal error")

// Helper type with custom MarshalJSON that returns an error.
type UnmarshalableStruct struct {
	Name string
}

func (u UnmarshalableStruct) MarshalJSON() ([]byte, error) {
	return nil, ErrIntentionalMarshalFailure
}

// TestValidator_Dict_UnmarshalableType tests Dict with type that has failing MarshalJSON.
func TestValidator_Dict_UnmarshalableType(t *testing.T) {
	validator := New[UnmarshalableStruct]()
	obj := &UnmarshalableStruct{Name: "test"}

	dict, err := validator.Dict(obj)

	// Should return error (either from Marshal or Unmarshal of nil bytes)
	require.Error(t, err)
	assert.Nil(t, dict)
	// Error could be from json.Unmarshal when it tries to parse nil bytes
	// This is still validating that our error handling works
}

// ==================== Extra Fields Tests ====================

// TestExtraFields_DefaultBehavior tests that default behavior ignores unknown fields.
func TestExtraFields_DefaultBehavior(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"required"`
	}

	// Default options (ExtraIgnore)
	validator := New[User]()

	// JSON with extra field "age" not in struct
	jsonData := []byte(`{"name": "Alice", "age": 30}`)
	user, err := validator.Unmarshal(jsonData)

	require.NoError(t, err)
	assert.Equal(t, "Alice", user.Name)
}

// TestExtraFields_Ignore tests ExtraIgnore explicitly ignores unknown fields.
func TestExtraFields_Ignore(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"required"`
	}

	validator := New[User](ValidatorOptions{
		StrictMissingFields: true,
		ExtraFields:         ExtraIgnore,
	})

	// JSON with extra fields
	jsonData := []byte(`{"name": "Bob", "email": "bob@test.com", "age": 25}`)
	user, err := validator.Unmarshal(jsonData)

	require.NoError(t, err)
	assert.Equal(t, "Bob", user.Name)
}

// TestExtraFields_Forbid tests ExtraForbid rejects unknown fields.
func TestExtraFields_Forbid(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"required"`
	}

	validator := New[User](ValidatorOptions{
		StrictMissingFields: true,
		ExtraFields:         ExtraForbid,
	})

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "no extra fields - valid",
			json:    `{"name": "Alice"}`,
			wantErr: false,
		},
		{
			name:    "one extra field - error",
			json:    `{"name": "Alice", "age": 30}`,
			wantErr: true,
		},
		{
			name:    "multiple extra fields - error",
			json:    `{"name": "Alice", "age": 30, "email": "test@example.com"}`,
			wantErr: true,
		},
		{
			name:    "nested extra field in root - error",
			json:    `{"name": "Alice", "metadata": {"key": "value"}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := validator.Unmarshal([]byte(tt.json))

			if tt.wantErr {
				require.Error(t, err)

				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				require.Len(t, ve.Errors, 1)
				assert.Equal(t, "root", ve.Errors[0].Field)
				assert.Equal(t, ErrMsgUnknownField, ve.Errors[0].Message)

				// User should still be partially populated even with error
				assert.NotNil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "Alice", user.Name)
			}
		})
	}
}

// TestExtraFields_Forbid_NestedStruct tests ExtraForbid with nested structs.
func TestExtraFields_Forbid_NestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
	}
	type User struct {
		Name    string  `json:"name" pedantigo:"required"`
		Address Address `json:"address"`
	}

	validator := New[User](ValidatorOptions{
		StrictMissingFields: true,
		ExtraFields:         ExtraForbid,
	})

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid nested struct",
			json:    `{"name": "Alice", "address": {"city": "NYC"}}`,
			wantErr: false,
		},
		{
			name:    "extra field at root",
			json:    `{"name": "Alice", "address": {"city": "NYC"}, "extra": "field"}`,
			wantErr: true,
		},
		{
			name:    "extra field in nested struct",
			json:    `{"name": "Alice", "address": {"city": "NYC", "country": "USA"}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Unmarshal([]byte(tt.json))

			if tt.wantErr {
				require.Error(t, err)

				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				require.Len(t, ve.Errors, 1)
				assert.Equal(t, "root", ve.Errors[0].Field)
				assert.Equal(t, ErrMsgUnknownField, ve.Errors[0].Message)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestExtraFields_Forbid_WithStrictMissingFieldsFalse tests interaction with StrictMissingFields=false.
func TestExtraFields_Forbid_WithStrictMissingFieldsFalse(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	validator := New[User](ValidatorOptions{
		StrictMissingFields: false,
		ExtraFields:         ExtraForbid,
	})

	// Valid: no extra fields, missing 'age' is OK with StrictMissingFields=false
	user, err := validator.Unmarshal([]byte(`{"name": "Alice"}`))
	require.NoError(t, err)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, 0, user.Age) // Zero value

	// Invalid: extra field
	_, err = validator.Unmarshal([]byte(`{"name": "Alice", "extra": "field"}`))
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	require.Len(t, ve.Errors, 1)
	assert.Equal(t, "root", ve.Errors[0].Field)
	assert.Equal(t, "JSON decode error: "+ErrMsgUnknownField, ve.Errors[0].Message)
}

// ==================== MarshalWithOptions Tests ====================

// Test types for marshal options.
type UserWithSensitiveData struct {
	ID       int    `json:"id" pedantigo:"min=1"`
	Name     string `json:"name" pedantigo:"min=2"`
	Email    string `json:"email" pedantigo:"email"`
	Password string `json:"password" pedantigo:"exclude:response,log,min=8"`
	Token    string `json:"token" pedantigo:"exclude:log"`
	Port     int    `json:"port" pedantigo:"omitzero"`
	Debug    bool   `json:"debug" pedantigo:"omitzero"`
}

func TestValidator_MarshalWithOptions_ExcludeContext(t *testing.T) {
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Password string `json:"password" pedantigo:"exclude:response"`
		Token    string `json:"token" pedantigo:"exclude:log"`
	}

	validator := New[User]()
	user := &User{
		ID:       1,
		Name:     "Alice",
		Password: "secret123",
		Token:    "token456",
	}

	// Test 1: Marshal with "response" context - should exclude password
	optsResponse := ForContext("response")
	dataResponse, err := validator.MarshalWithOptions(user, optsResponse)
	require.NoError(t, err)

	var resultResponse map[string]interface{}
	require.NoError(t, json.Unmarshal(dataResponse, &resultResponse))

	assert.InDelta(t, float64(1), resultResponse["id"], 0.001)
	assert.Equal(t, "Alice", resultResponse["name"])
	assert.NotContains(t, resultResponse, "password", "password should be excluded in 'response' context")
	assert.Equal(t, "token456", resultResponse["token"], "token should be included in 'response' context")

	// Test 2: Marshal with "log" context - should exclude token
	optsLog := ForContext("log")
	dataLog, err := validator.MarshalWithOptions(user, optsLog)
	require.NoError(t, err)

	var resultLog map[string]interface{}
	require.NoError(t, json.Unmarshal(dataLog, &resultLog))

	assert.InDelta(t, float64(1), resultLog["id"], 0.001)
	assert.Equal(t, "Alice", resultLog["name"])
	assert.Equal(t, "secret123", resultLog["password"], "password should be included in 'log' context")
	assert.NotContains(t, resultLog, "token", "token should be excluded in 'log' context")

	// Test 3: Marshal with no context - should include all fields
	optsNone := DefaultMarshalOptions()
	dataNone, err := validator.MarshalWithOptions(user, optsNone)
	require.NoError(t, err)

	var resultNone map[string]interface{}
	require.NoError(t, json.Unmarshal(dataNone, &resultNone))

	assert.InDelta(t, float64(1), resultNone["id"], 0.001)
	assert.Equal(t, "Alice", resultNone["name"])
	assert.Equal(t, "secret123", resultNone["password"], "password should be included with no context")
	assert.Equal(t, "token456", resultNone["token"], "token should be included with no context")
}

func TestValidator_MarshalWithOptions_OmitZero(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"min=2"`
		Port    int    `json:"port" pedantigo:"omitzero"`
		Timeout int    `json:"timeout" pedantigo:"omitzero"`
		Retries int    `json:"retries"`
	}

	validator := New[Config]()

	tests := []struct {
		name          string
		config        *Config
		opts          MarshalOptions
		shouldHave    []string
		shouldNotHave []string
	}{
		{
			name: "omitzero enabled - omits zero values with tag",
			config: &Config{
				Name:    "myapp",
				Port:    0, // Zero with omitzero tag
				Timeout: 0, // Zero with omitzero tag
				Retries: 0, // Zero without omitzero tag
			},
			opts: MarshalOptions{
				Context:  "",
				OmitZero: true,
			},
			shouldHave:    []string{"name", "retries"},
			shouldNotHave: []string{"port", "timeout"},
		},
		{
			name: "omitzero enabled - includes non-zero values",
			config: &Config{
				Name:    "myapp",
				Port:    8080, // Non-zero with omitzero tag
				Timeout: 30,   // Non-zero with omitzero tag
				Retries: 3,
			},
			opts: MarshalOptions{
				Context:  "",
				OmitZero: true,
			},
			shouldHave:    []string{"name", "port", "timeout", "retries"},
			shouldNotHave: []string{},
		},
		{
			name: "omitzero disabled - includes all fields",
			config: &Config{
				Name:    "myapp",
				Port:    0, // Zero with omitzero tag
				Timeout: 0, // Zero with omitzero tag
				Retries: 0,
			},
			opts: MarshalOptions{
				Context:  "",
				OmitZero: false, // Disabled
			},
			shouldHave:    []string{"name", "port", "timeout", "retries"},
			shouldNotHave: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := validator.MarshalWithOptions(tt.config, tt.opts)
			require.NoError(t, err)

			var result map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &result))

			for _, field := range tt.shouldHave {
				assert.Contains(t, result, field, "field %s should be present", field)
			}

			for _, field := range tt.shouldNotHave {
				assert.NotContains(t, result, field, "field %s should be omitted", field)
			}
		})
	}
}

func TestValidator_MarshalWithOptions_CombinedExcludeAndOmitZero(t *testing.T) {
	user := &UserWithSensitiveData{
		ID:       1,
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
		Token:    "token456",
		Port:     0,     // Zero with omitzero tag
		Debug:    false, // Zero with omitzero tag
	}

	validator := New[UserWithSensitiveData]()

	// Marshal with "response" context and OmitZero enabled
	opts := MarshalOptions{
		Context:  "response",
		OmitZero: true,
	}

	data, err := validator.MarshalWithOptions(user, opts)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	// Should include: id, name, email, token
	assert.InDelta(t, float64(1), result["id"], 0.001)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "alice@example.com", result["email"])
	assert.Equal(t, "token456", result["token"])

	// Should exclude password (excluded by context)
	assert.NotContains(t, result, "password")

	// Should exclude port and debug (omitted by omitzero)
	assert.NotContains(t, result, "port")
	assert.NotContains(t, result, "debug")
}

func TestValidator_MarshalWithOptions_ValidationErrors(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{
		Name:  "A",          // Too short (min=2)
		Email: "notanemail", // Invalid email
		Age:   15,           // Too young (min=18)
	}

	// MarshalWithOptions should still validate and return errors
	opts := DefaultMarshalOptions()
	data, err := validator.MarshalWithOptions(user, opts)

	// Should return validation error
	require.Error(t, err)
	assert.Nil(t, data)

	// Verify it's a ValidationError with multiple field errors
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 3)
}

func TestValidator_MarshalWithOptions_PointerFields(t *testing.T) {
	type Config struct {
		Name    string `json:"name"`
		Port    *int   `json:"port" pedantigo:"omitzero"`
		Enabled *bool  `json:"enabled" pedantigo:"omitzero"`
	}

	validator := New[Config]()

	// Test with nil pointers
	config1 := &Config{
		Name:    "app",
		Port:    nil,
		Enabled: nil,
	}

	opts := MarshalOptions{
		Context:  "",
		OmitZero: true,
	}

	data1, err := validator.MarshalWithOptions(config1, opts)
	require.NoError(t, err)

	var result1 map[string]interface{}
	require.NoError(t, json.Unmarshal(data1, &result1))

	assert.Equal(t, "app", result1["name"])
	assert.NotContains(t, result1, "port", "nil pointer with omitzero should be omitted")
	assert.NotContains(t, result1, "enabled", "nil pointer with omitzero should be omitted")

	// Test with non-nil pointers
	port := 8080
	enabled := true
	config2 := &Config{
		Name:    "app",
		Port:    &port,
		Enabled: &enabled,
	}

	data2, err := validator.MarshalWithOptions(config2, opts)
	require.NoError(t, err)

	var result2 map[string]interface{}
	require.NoError(t, json.Unmarshal(data2, &result2))

	assert.Equal(t, "app", result2["name"])
	assert.InDelta(t, float64(8080), result2["port"], 0.001)
	assert.Equal(t, true, result2["enabled"])
}

func TestValidator_MarshalWithOptions_NestedStructs(t *testing.T) {
	type Address struct {
		Street     string `json:"street"`
		City       string `json:"city"`
		PostalCode string `json:"postal_code" pedantigo:"exclude:summary"`
	}

	type User struct {
		Name     string  `json:"name"`
		Password string  `json:"password" pedantigo:"exclude:response"`
		Address  Address `json:"address"`
	}

	validator := New[User]()
	user := &User{
		Name:     "Alice",
		Password: "secret123",
		Address: Address{
			Street:     "123 Main St",
			City:       "NYC",
			PostalCode: "10001",
		},
	}

	// Marshal with "response" context
	opts := ForContext("response")
	data, err := validator.MarshalWithOptions(user, opts)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	// Should include name
	assert.Equal(t, "Alice", result["name"])

	// Should exclude password
	assert.NotContains(t, result, "password")

	// Should include nested address with all fields (postal_code not excluded in "response" context)
	address, ok := result["address"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "123 Main St", address["street"])
	assert.Equal(t, "NYC", address["city"])
	assert.Equal(t, "10001", address["postal_code"])

	// Marshal with "summary" context
	optsSummary := ForContext("summary")
	dataSummary, err := validator.MarshalWithOptions(user, optsSummary)
	require.NoError(t, err)

	var resultSummary map[string]interface{}
	require.NoError(t, json.Unmarshal(dataSummary, &resultSummary))

	// Nested address should exclude postal_code in "summary" context
	addressSummary, ok := resultSummary["address"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "123 Main St", addressSummary["street"])
	assert.Equal(t, "NYC", addressSummary["city"])
	assert.NotContains(t, addressSummary, "postal_code")
}

func TestValidator_Marshal_BackwardCompatible(t *testing.T) {
	type User struct {
		ID       int    `json:"id" pedantigo:"min=1"`
		Name     string `json:"name" pedantigo:"min=2"`
		Password string `json:"password" pedantigo:"exclude:response"`
	}

	validator := New[User]()
	user := &User{
		ID:       1,
		Name:     "Alice",
		Password: "secret123",
	}

	// Marshal() without options should still work and include all fields
	data, err := validator.Marshal(user)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	assert.InDelta(t, float64(1), result["id"], 0.001)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "secret123", result["password"], "Marshal() without options should include password")
}

func TestValidator_MarshalWithOptions_NilPointer(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	validator := New[User]()

	// Pass nil pointer
	opts := DefaultMarshalOptions()
	data, err := validator.MarshalWithOptions(nil, opts)

	// Should handle nil appropriately (either error or marshal "null")
	if err != nil {
		// Validation error is acceptable for nil
		t.Logf("MarshalWithOptions(nil) returned error: %v", err)
	} else {
		assert.Equal(t, "null", string(data))
	}
}

func TestValidator_MarshalWithOptions_MultipleExclusionContexts(t *testing.T) {
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Password string `json:"password" pedantigo:"exclude:response|log|audit"`
	}

	validator := New[User]()
	user := &User{
		ID:       1,
		Name:     "Alice",
		Password: "secret123",
	}

	contexts := []string{"response", "log", "audit"}

	for _, ctx := range contexts {
		t.Run("context_"+ctx, func(t *testing.T) {
			opts := ForContext(ctx)
			data, err := validator.MarshalWithOptions(user, opts)
			require.NoError(t, err)

			var result map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &result))

			assert.InDelta(t, float64(1), result["id"], 0.001)
			assert.Equal(t, "Alice", result["name"])
			assert.NotContains(t, result, "password", "password should be excluded in '%s' context", ctx)
		})
	}
}
