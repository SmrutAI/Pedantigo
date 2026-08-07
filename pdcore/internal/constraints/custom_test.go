package constraints

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomConstraint_Validate_Success tests that valid values pass custom validation.
func TestCustomConstraint_Validate_Success(t *testing.T) {
	tests := []struct {
		name      string
		validator CustomValidationFunc
		value     any
		param     string
	}{
		{
			name: "validator accepts string starting with 'valid'",
			validator: func(value any, param string) error {
				str, ok := value.(string)
				if !ok {
					return errors.New("must be string")
				}
				if len(str) >= 5 && str[:5] == "valid" {
					return nil
				}
				return errors.New("must start with 'valid'")
			},
			value: "valid_data",
			param: "",
		},
		{
			name: "validator accepts any string",
			validator: func(value any, param string) error {
				_, ok := value.(string)
				if !ok {
					return errors.New("must be string")
				}
				return nil
			},
			value: "any_string",
			param: "",
		},
		{
			name: "validator with param accepts matching prefix",
			validator: func(value any, param string) error {
				str, ok := value.(string)
				if !ok {
					return errors.New("must be string")
				}
				if len(str) >= len(param) && str[:len(param)] == param {
					return nil
				}
				return errors.New("prefix mismatch")
			},
			value: "PRE_data",
			param: "PRE_",
		},
		{
			name: "validator accepts int greater than param",
			validator: func(value any, param string) error {
				num, ok := value.(int)
				if !ok {
					return errors.New("must be int")
				}
				if param == "10" && num > 10 {
					return nil
				}
				return errors.New("must be > 10")
			},
			value: 15,
			param: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := customConstraint{
				name:  "test_validator",
				fn:    tt.validator,
				param: tt.param,
			}
			err := c.Validate(tt.value)
			assert.NoError(t, err)
		})
	}
}

// TestCustomConstraint_Validate_Failure tests that invalid values fail custom validation.
func TestCustomConstraint_Validate_Failure(t *testing.T) {
	tests := []struct {
		name          string
		validator     CustomValidationFunc
		value         any
		param         string
		expectErrCode string
		expectErrMsg  string
	}{
		{
			name: "validator rejects string not starting with 'valid'",
			validator: func(value any, param string) error {
				str, ok := value.(string)
				if !ok {
					return errors.New("must be string")
				}
				if len(str) >= 5 && str[:5] == "valid" {
					return nil
				}
				return errors.New("must start with 'valid'")
			},
			value:         "invalid_data",
			param:         "",
			expectErrCode: CodeCustomValidation,
			expectErrMsg:  "test_validator",
		},
		{
			name: "validator rejects non-string type",
			validator: func(value any, param string) error {
				_, ok := value.(string)
				if !ok {
					return errors.New("must be string")
				}
				return nil
			},
			value:         123,
			param:         "",
			expectErrCode: CodeCustomValidation,
			expectErrMsg:  "must be string",
		},
		{
			name: "validator with param rejects wrong prefix",
			validator: func(value any, param string) error {
				str, ok := value.(string)
				if !ok {
					return errors.New("must be string")
				}
				if len(str) >= len(param) && str[:len(param)] == param {
					return nil
				}
				return errors.New("prefix mismatch")
			},
			value:         "WRONG_data",
			param:         "PRE_",
			expectErrCode: CodeCustomValidation,
			expectErrMsg:  "prefix mismatch",
		},
		{
			name: "validator rejects int not greater than param",
			validator: func(value any, param string) error {
				num, ok := value.(int)
				if !ok {
					return errors.New("must be int")
				}
				if param == "10" && num > 10 {
					return nil
				}
				return errors.New("must be > 10")
			},
			value:         5,
			param:         "10",
			expectErrCode: CodeCustomValidation,
			expectErrMsg:  "must be > 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := customConstraint{
				name:  "test_validator",
				fn:    tt.validator,
				param: tt.param,
			}
			err := c.Validate(tt.value)
			require.Error(t, err)

			// Check that error is a ConstraintError with correct code
			var ce *ConstraintError
			if errors.As(err, &ce) {
				assert.Equal(t, tt.expectErrCode, ce.Code, "error code mismatch")
				assert.Contains(t, ce.Message, tt.expectErrMsg, "error message should contain expected text")
			} else {
				t.Errorf("expected ConstraintError, got %T", err)
			}
		})
	}
}

// TestCustomConstraint_Validate_WithParam tests validators with parameters.
func TestCustomConstraint_Validate_WithParam(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		value   any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "has_prefix with PRE_ - valid",
			param:   "PRE_",
			value:   "PRE_data",
			wantErr: false,
		},
		{
			name:    "has_prefix with PRE_ - invalid",
			param:   "PRE_",
			value:   "data",
			wantErr: true,
			errMsg:  "must start with",
		},
		{
			name:    "has_prefix with POST_ - valid",
			param:   "POST_",
			value:   "POST_data",
			wantErr: false,
		},
		{
			name:    "has_prefix with empty param - any string valid",
			param:   "",
			value:   "anything",
			wantErr: false,
		},
		{
			name:    "min_length param=5 - valid",
			param:   "5",
			value:   "hello",
			wantErr: false,
		},
		{
			name:    "min_length param=5 - invalid",
			param:   "5",
			value:   "hi",
			wantErr: true,
			errMsg:  "too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validator CustomValidationFunc
			if tt.name[:10] == "has_prefix" {
				// Prefix validator
				validator = func(value any, param string) error {
					str, ok := value.(string)
					if !ok {
						return errors.New("must be string")
					}
					if len(str) >= len(param) && str[:len(param)] == param {
						return nil
					}
					return errors.New("must start with " + param)
				}
			} else {
				// Min length validator
				validator = func(value any, param string) error {
					str, ok := value.(string)
					if !ok {
						return errors.New("must be string")
					}
					minLen := 0
					if param == "5" {
						minLen = 5
					}
					if len(str) >= minLen {
						return nil
					}
					return errors.New("too short")
				}
			}

			c := customConstraint{
				name:  "test_validator",
				fn:    validator,
				param: tt.param,
			}
			err := c.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCustomConstraint_Validate_NilValue tests validation of nil values.
func TestCustomConstraint_Validate_NilValue(t *testing.T) {
	tests := []struct {
		name      string
		validator CustomValidationFunc
		value     any
		wantErr   bool
	}{
		{
			name: "nil pointer - validator handles nil gracefully",
			validator: func(value any, param string) error {
				if value == nil {
					return nil // Skip validation for nil interface
				}
				// Check for nil pointer using reflection (nil pointer in interface != nil interface)
				v := reflect.ValueOf(value)
				if v.Kind() == reflect.Ptr && v.IsNil() {
					return nil // Skip validation for nil pointer
				}
				// Dereference pointer if needed
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}
				if v.Kind() != reflect.String {
					return errors.New("must be string")
				}
				if v.String() == "" {
					return errors.New("must not be empty")
				}
				return nil
			},
			value:   (*string)(nil),
			wantErr: false,
		},
		{
			name: "nil value - validator rejects nil",
			validator: func(value any, param string) error {
				if value == nil {
					return errors.New("value cannot be nil")
				}
				return nil
			},
			value:   nil,
			wantErr: true,
		},
		{
			name: "nil interface - validator handles",
			validator: func(value any, param string) error {
				if value == nil {
					return nil
				}
				return nil
			},
			value:   nil,
			wantErr: false,
		},
		{
			name: "non-nil string pointer - validates content",
			validator: func(value any, param string) error {
				// For pointers, validation typically happens on dereferenced value
				// But custom validators receive the raw value
				str, ok := value.(*string)
				if !ok {
					return errors.New("expected string pointer")
				}
				if str == nil {
					return nil // nil pointer is okay
				}
				if *str == "" {
					return errors.New("string cannot be empty")
				}
				return nil
			},
			value:   ptrTo("hello"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := customConstraint{
				name:  "test_validator",
				fn:    tt.validator,
				param: "",
			}
			err := c.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBuildCustomConstraint_Found tests building constraint for registered validator.
func TestBuildCustomConstraint_Found(t *testing.T) {
	// Setup: Create a mock lookup that returns a known validator
	mockValidator := func(value any, param string) error {
		return nil // Always valid
	}

	mockLookup := func(name string) (CustomValidationFunc, bool) {
		if name == "known_validator" {
			return mockValidator, true
		}
		return nil, false
	}

	// Inject the mock lookup
	originalLookup := customValidatorLookup
	SetCustomValidatorLookup(mockLookup)
	defer func() {
		customValidatorLookup = originalLookup
	}()

	// Test
	constraint, found := BuildCustomConstraint("known_validator", "")
	assert.True(t, found, "should find known validator")
	assert.NotNil(t, constraint, "should return constraint")

	// Verify the constraint is of correct type
	cc, ok := constraint.(customConstraint)
	require.True(t, ok, "should be customConstraint type")
	assert.Equal(t, "known_validator", cc.name)
	assert.Empty(t, cc.param)
}

// TestBuildCustomConstraint_NotFound tests building constraint for unknown validator.
func TestBuildCustomConstraint_NotFound(t *testing.T) {
	// Setup: Create a mock lookup that returns not found
	mockLookup := func(name string) (CustomValidationFunc, bool) {
		return nil, false // Always not found
	}

	// Inject the mock lookup
	originalLookup := customValidatorLookup
	SetCustomValidatorLookup(mockLookup)
	defer func() {
		customValidatorLookup = originalLookup
	}()

	// Test
	constraint, found := BuildCustomConstraint("unknown_validator", "")
	assert.False(t, found, "should not find unknown validator")
	assert.Nil(t, constraint, "should return nil constraint")
}

// TestBuildCustomConstraint_WithParam tests building constraint with parameter.
func TestBuildCustomConstraint_WithParam(t *testing.T) {
	tests := []struct {
		name          string
		validatorName string
		param         string
		expectFound   bool
	}{
		{
			name:          "found with simple param",
			validatorName: "has_prefix",
			param:         "PRE_",
			expectFound:   true,
		},
		{
			name:          "found with numeric param",
			validatorName: "min_length",
			param:         "10",
			expectFound:   true,
		},
		{
			name:          "found with empty param",
			validatorName: "required_if",
			param:         "",
			expectFound:   true,
		},
		{
			name:          "not found returns false",
			validatorName: "nonexistent",
			param:         "anything",
			expectFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock lookup
			mockValidator := func(value any, param string) error {
				return nil
			}

			mockLookup := func(name string) (CustomValidationFunc, bool) {
				if tt.expectFound {
					return mockValidator, true
				}
				return nil, false
			}

			originalLookup := customValidatorLookup
			SetCustomValidatorLookup(mockLookup)
			defer func() {
				customValidatorLookup = originalLookup
			}()

			// Test
			constraint, found := BuildCustomConstraint(tt.validatorName, tt.param)
			assert.Equal(t, tt.expectFound, found)

			if tt.expectFound {
				require.NotNil(t, constraint)
				cc, ok := constraint.(customConstraint)
				require.True(t, ok)
				assert.Equal(t, tt.validatorName, cc.name)
				assert.Equal(t, tt.param, cc.param)
			} else {
				assert.Nil(t, constraint)
			}
		})
	}
}

// TestSetCustomValidatorLookup tests the lookup injection.
func TestSetCustomValidatorLookup(t *testing.T) {
	// Save original lookup
	originalLookup := customValidatorLookup
	defer func() {
		customValidatorLookup = originalLookup
	}()

	// Create a test validator
	testValidator := func(value any, param string) error {
		return errors.New("test error")
	}

	// Create a new lookup function
	newLookup := func(name string) (CustomValidationFunc, bool) {
		if name == "test_validator" {
			return testValidator, true
		}
		return nil, false
	}

	// Set the new lookup
	SetCustomValidatorLookup(newLookup)

	// Verify the lookup was set by using BuildCustomConstraint
	constraint, found := BuildCustomConstraint("test_validator", "")
	assert.True(t, found)
	assert.NotNil(t, constraint)

	// Verify a different name returns not found
	constraint2, found2 := BuildCustomConstraint("other_validator", "")
	assert.False(t, found2)
	assert.Nil(t, constraint2)
}

// TestCustomConstraint_ErrorCode tests error code is correct.
func TestCustomConstraint_ErrorCode(t *testing.T) {
	tests := []struct {
		name             string
		validator        CustomValidationFunc
		value            any
		param            string
		expectedCode     string
		expectedMsgParts []string
	}{
		{
			name: "validator error returns CodeCustomValidation",
			validator: func(value any, param string) error {
				return errors.New("validation failed")
			},
			value:            "test",
			param:            "",
			expectedCode:     CodeCustomValidation,
			expectedMsgParts: []string{"validation failed"},
		},
		{
			name: "validator with param error includes param context",
			validator: func(value any, param string) error {
				return errors.New("must be greater than " + param)
			},
			value:            5,
			param:            "10",
			expectedCode:     CodeCustomValidation,
			expectedMsgParts: []string{"must be greater than 10"},
		},
		{
			name: "validator type error",
			validator: func(value any, param string) error {
				_, ok := value.(string)
				if !ok {
					return errors.New("must be string type")
				}
				return nil
			},
			value:            123,
			param:            "",
			expectedCode:     CodeCustomValidation,
			expectedMsgParts: []string{"must be string type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := customConstraint{
				name:  "my_validator",
				fn:    tt.validator,
				param: tt.param,
			}
			err := c.Validate(tt.value)
			require.Error(t, err)

			// Check that error is ConstraintError with correct code
			var ce *ConstraintError
			require.ErrorAs(t, err, &ce, "error should be ConstraintError")
			assert.Equal(t, tt.expectedCode, ce.Code, "error code should be CodeCustomValidation")

			// Check error message contains expected parts
			for _, part := range tt.expectedMsgParts {
				assert.Contains(t, ce.Message, part, "error message should contain expected text")
			}
		})
	}
}

// TestCustomConstraint_ErrorMessage tests error messages include validator name.
func TestCustomConstraint_ErrorMessage(t *testing.T) {
	validatorName := "us_phone_number"
	validator := func(value any, param string) error {
		return errors.New("invalid phone format")
	}

	c := customConstraint{
		name:  validatorName,
		fn:    validator,
		param: "",
	}

	err := c.Validate("not-a-phone")
	require.Error(t, err)

	var ce *ConstraintError
	require.ErrorAs(t, err, &ce)
	assert.Equal(t, CodeCustomValidation, ce.Code)

	// Error message should include validator name and original error
	assert.Contains(t, ce.Message, validatorName, "error should include validator name")
	assert.Contains(t, ce.Message, "invalid phone format", "error should include original error message")
}

// TestBuildCustomConstraint_NilLookup tests behavior when lookup is not set.
func TestBuildCustomConstraint_NilLookup(t *testing.T) {
	// Save and clear the lookup
	originalLookup := customValidatorLookup
	customValidatorLookup = nil
	defer func() {
		customValidatorLookup = originalLookup
	}()

	// Should return (nil, false) when lookup is not set
	constraint, found := BuildCustomConstraint("any_validator", "")
	assert.False(t, found)
	assert.Nil(t, constraint)
}
