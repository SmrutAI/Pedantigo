package pedantigo

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegisterValidation_Success tests successful registration of a custom validator.
func TestRegisterValidation_Success(t *testing.T) {
	validatorFunc := func(value any) error {
		s, ok := value.(string)
		if !ok {
			return errors.New("expected string")
		}
		if len(s) < 5 {
			return errors.New("must be at least 5 characters")
		}
		return nil
	}

	err := RegisterValidation("custom_min5", validatorFunc)
	require.NoError(t, err)

	// Verify it was registered
	fn, ok := GetCustomValidator("custom_min5")
	assert.True(t, ok, "validator should be registered")
	assert.NotNil(t, fn)
}

// TestRegisterValidation_EmptyName tests that empty name returns error.
func TestRegisterValidation_EmptyName(t *testing.T) {
	validatorFunc := func(value any) error {
		return nil
	}

	err := RegisterValidation("", validatorFunc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestRegisterValidation_NilFunc tests that nil function returns error.
func TestRegisterValidation_NilFunc(t *testing.T) {
	err := RegisterValidation("custom_validator", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestRegisterValidation_BuiltInConflict tests that overriding built-in validators fails.
func TestRegisterValidation_BuiltInConflict(t *testing.T) {
	builtInNames := []string{
		"email", "required", "min", "max", "url", "uuid",
		"regexp", "ipv4", "ipv6", "len", "gt", "gte",
		"lt", "lte", "oneof", "const", "ascii", "alpha",
	}

	validatorFunc := func(value any) error {
		return nil
	}

	for _, name := range builtInNames {
		t.Run(name, func(t *testing.T) {
			err := RegisterValidation(name, validatorFunc)
			require.Error(t, err, "should not allow overriding built-in validator: %s", name)
			assert.Contains(t, err.Error(), "built-in")
		})
	}
}

// TestRegisterStructValidation_Success tests successful struct-level validator registration.
func TestRegisterStructValidation_Success(t *testing.T) {
	type Account struct {
		Balance float64 `json:"balance"`
		Limit   float64 `json:"limit"`
	}

	structValidator := func(a *Account) error {
		if a.Balance > a.Limit {
			return errors.New("balance exceeds limit")
		}
		return nil
	}

	err := RegisterStructValidation[Account](structValidator)
	assert.NoError(t, err)
}

// TestRegisterStructValidation_NilFunc tests that nil function returns error.
func TestRegisterStructValidation_NilFunc(t *testing.T) {
	type TestStruct struct {
		Value string `json:"value"`
	}

	err := RegisterStructValidation[TestStruct](nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestGetCustomValidator_Exists tests retrieval of registered validator.
func TestGetCustomValidator_Exists(t *testing.T) {
	validatorFunc := func(value any) error {
		return errors.New("test error")
	}

	err := RegisterValidation("test_validator", validatorFunc)
	require.NoError(t, err)

	fn, ok := GetCustomValidator("test_validator")
	assert.True(t, ok, "validator should exist")
	assert.NotNil(t, fn)

	// Verify the function works
	err = fn("test")
	require.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestGetCustomValidator_NotExists tests retrieval of non-existent validator.
func TestGetCustomValidator_NotExists(t *testing.T) {
	fn, ok := GetCustomValidator("nonexistent_validator_12345")
	assert.False(t, ok, "validator should not exist")
	assert.Nil(t, fn)
}

// TestConcurrentRegistration tests concurrent validator registration without races.
func TestConcurrentRegistration(t *testing.T) {
	var wg sync.WaitGroup

	// Concurrent registrations should not race
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("custom_%d", idx)
			_ = RegisterValidation(name, func(value any) error {
				return nil
			})
		}(i)
	}

	wg.Wait()

	// Verify all validators were registered
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("custom_%d", i)
		fn, ok := GetCustomValidator(name)
		assert.True(t, ok, "validator %s should exist", name)
		assert.NotNil(t, fn)
	}
}

// TestRegisterValidation_ConcurrentWithValidation tests concurrent registration and validation.
func TestRegisterValidation_ConcurrentWithValidation(t *testing.T) {
	type User struct {
		Code string `json:"code" pedantigo:"custom_code"`
	}

	// First, register the validator
	err := RegisterValidation("custom_code", func(value any) error {
		s, ok := value.(string)
		if !ok {
			return errors.New("expected string")
		}
		if len(s) != 6 {
			return errors.New("code must be 6 characters")
		}
		return nil
	})
	require.NoError(t, err)

	var wg sync.WaitGroup

	// Now validate concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user := &User{Code: "ABC123"}
			_ = Validate[User](user)
		}()
	}

	wg.Wait()
}

// TestRegisterStructValidation_ConcurrentAccess tests concurrent struct validation access.
func TestRegisterStructValidation_ConcurrentAccess(t *testing.T) {
	type Account struct {
		Balance float64 `json:"balance"`
		Limit   float64 `json:"limit"`
	}

	err := RegisterStructValidation[Account](func(a *Account) error {
		if a.Balance > a.Limit {
			return errors.New("balance exceeds limit")
		}
		return nil
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			account := &Account{Balance: float64(idx), Limit: 100}
			_ = Validate[Account](account)
		}(i)
	}

	wg.Wait()
}

// TestRegisterValidation_MultipleCustomValidators tests registering multiple validators.
func TestRegisterValidation_MultipleCustomValidators(t *testing.T) {
	validators := map[string]ValidationFunc{
		"is_even": func(value any) error {
			num, ok := value.(int)
			if !ok {
				return errors.New("expected int")
			}
			if num%2 != 0 {
				return errors.New("must be even")
			}
			return nil
		},
		"is_positive": func(value any) error {
			num, ok := value.(int)
			if !ok {
				return errors.New("expected int")
			}
			if num <= 0 {
				return errors.New("must be positive")
			}
			return nil
		},
		"divisible_by_3": func(value any) error {
			num, ok := value.(int)
			if !ok {
				return errors.New("expected int")
			}
			if num%3 != 0 {
				return errors.New("must be divisible by 3")
			}
			return nil
		},
	}

	// Register all validators
	for name, fn := range validators {
		err := RegisterValidation(name, fn)
		require.NoError(t, err, "failed to register %s", name)
	}

	// Verify all were registered
	for name := range validators {
		fn, ok := GetCustomValidator(name)
		assert.True(t, ok, "validator %s should exist", name)
		assert.NotNil(t, fn, "validator %s should not be nil", name)
	}
}

// TestRegisterStructValidation_MultipleDifferentTypes tests registering validators for different types.
func TestRegisterStructValidation_MultipleDifferentTypes(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Product struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	type Order struct {
		Total float64 `json:"total"`
		Items int     `json:"items"`
	}

	// Register validators for different types
	err := RegisterStructValidation[User](func(u *User) error {
		if u.Age < 0 {
			return errors.New("age cannot be negative")
		}
		return nil
	})
	require.NoError(t, err)

	err = RegisterStructValidation[Product](func(p *Product) error {
		if p.Price < 0 {
			return errors.New("price cannot be negative")
		}
		return nil
	})
	require.NoError(t, err)

	err = RegisterStructValidation[Order](func(o *Order) error {
		if o.Items <= 0 {
			return errors.New("order must have at least one item")
		}
		return nil
	})
	require.NoError(t, err)
}

// TestGetCustomValidator_CaseSensitive tests that validator names are case-sensitive.
func TestGetCustomValidator_CaseSensitive(t *testing.T) {
	validatorFunc := func(value any) error {
		return nil
	}

	err := RegisterValidation("TestValidator", validatorFunc)
	require.NoError(t, err)

	// Exact match should work
	fn, ok := GetCustomValidator("TestValidator")
	assert.True(t, ok, "exact case should match")
	assert.NotNil(t, fn)

	// Different case should not work
	fn, ok = GetCustomValidator("testvalidator")
	assert.False(t, ok, "different case should not match")
	assert.Nil(t, fn)

	fn, ok = GetCustomValidator("TESTVALIDATOR")
	assert.False(t, ok, "different case should not match")
	assert.Nil(t, fn)
}

// TestRegisterValidation_OverwriteCustomValidator tests overwriting a custom validator.
func TestRegisterValidation_OverwriteCustomValidator(t *testing.T) {
	// Register first validator
	firstValidator := func(value any) error {
		return errors.New("first validator")
	}

	err := RegisterValidation("overwrite_test", firstValidator)
	require.NoError(t, err)

	// Verify first validator works
	fn, ok := GetCustomValidator("overwrite_test")
	require.True(t, ok)
	err = fn("test")
	require.EqualError(t, err, "first validator")

	// Register second validator with same name (should overwrite or error)
	secondValidator := func(value any) error {
		return errors.New("second validator")
	}

	err = RegisterValidation("overwrite_test", secondValidator)
	// Implementation can choose to allow or deny overwriting
	// We test both behaviors are acceptable
	if err == nil {
		// Overwriting is allowed - verify the new validator is active
		fn, ok := GetCustomValidator("overwrite_test")
		require.True(t, ok)
		err = fn("test")
		assert.EqualError(t, err, "second validator", "should use the new validator")
	} else {
		// Overwriting is denied - verify error message
		assert.Contains(t, err.Error(), "already registered")
	}
}

// TestRegisterStructValidation_DuplicateRegistration tests registering struct validator twice.
func TestRegisterStructValidation_DuplicateRegistration(t *testing.T) {
	type TestStruct struct {
		Value int `json:"value"`
	}

	firstValidator := func(ts *TestStruct) error {
		return errors.New("first validator")
	}

	err := RegisterStructValidation[TestStruct](firstValidator)
	require.NoError(t, err)

	// Try to register again
	secondValidator := func(ts *TestStruct) error {
		return errors.New("second validator")
	}

	err = RegisterStructValidation[TestStruct](secondValidator)
	// Implementation can choose to allow or deny duplicate registration
	// We test that the behavior is documented (either succeeds or fails with clear error)
	if err != nil {
		assert.Contains(t, err.Error(), "already registered")
	}
}

// TestValidationFunc_TypeAssertion tests that custom validators handle type assertions.
func TestValidationFunc_TypeAssertion(t *testing.T) {
	stringValidator := func(value any) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		if s == "" {
			return errors.New("string cannot be empty")
		}
		return nil
	}

	err := RegisterValidation("string_nonempty", stringValidator)
	require.NoError(t, err)

	fn, ok := GetCustomValidator("string_nonempty")
	require.True(t, ok)

	// Test with correct type
	err = fn("hello")
	require.NoError(t, err)

	err = fn("")
	require.EqualError(t, err, "string cannot be empty")

	// Test with incorrect type
	err = fn(123)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected string")
}

// TestStructLevelFunc_CrossFieldValidation tests struct-level validation with multiple fields.
func TestStructLevelFunc_CrossFieldValidation(t *testing.T) {
	type DateRange struct {
		StartDate int `json:"start_date"`
		EndDate   int `json:"end_date"`
	}

	rangeValidator := func(dr *DateRange) error {
		if dr.EndDate < dr.StartDate {
			return errors.New("end_date must be after start_date")
		}
		if dr.EndDate-dr.StartDate > 365 {
			return errors.New("date range cannot exceed 365 days")
		}
		return nil
	}

	err := RegisterStructValidation[DateRange](rangeValidator)
	assert.NoError(t, err)
}

// TestRegisterValidation_ComplexValidator tests a complex custom validator.
func TestRegisterValidation_ComplexValidator(t *testing.T) {
	// Password validator: must have uppercase, lowercase, digit, and special char
	passwordValidator := func(value any) error {
		s, ok := value.(string)
		if !ok {
			return errors.New("expected string")
		}

		if len(s) < 8 {
			return errors.New("password must be at least 8 characters")
		}

		var hasUpper, hasLower, hasDigit, hasSpecial bool
		for _, ch := range s {
			switch {
			case ch >= 'A' && ch <= 'Z':
				hasUpper = true
			case ch >= 'a' && ch <= 'z':
				hasLower = true
			case ch >= '0' && ch <= '9':
				hasDigit = true
			case ch == '!' || ch == '@' || ch == '#' || ch == '$' || ch == '%':
				hasSpecial = true
			}
		}

		if !hasUpper {
			return errors.New("password must contain uppercase letter")
		}
		if !hasLower {
			return errors.New("password must contain lowercase letter")
		}
		if !hasDigit {
			return errors.New("password must contain digit")
		}
		if !hasSpecial {
			return errors.New("password must contain special character (!@#$%)")
		}

		return nil
	}

	err := RegisterValidation("strong_password", passwordValidator)
	require.NoError(t, err)

	fn, ok := GetCustomValidator("strong_password")
	require.True(t, ok)

	// Test valid password
	err = fn("Passw0rd!")
	require.NoError(t, err)

	// Test invalid passwords
	testCases := []struct {
		password string
		errMsg   string
	}{
		{"short", "at least 8 characters"},
		{"alllowercase1!", "uppercase letter"},
		{"ALLUPPERCASE1!", "lowercase letter"},
		{"NoDigitsHere!", "digit"},
		{"NoSpecial1Char", "special character"},
	}

	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			err := fn(tc.password)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}
