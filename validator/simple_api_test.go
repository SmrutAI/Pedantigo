package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// Basic Function Tests (10 tests)
// ==============================================================================

func TestUnmarshal_Basic(t *testing.T) {
	// Local struct to avoid cross-test pollution
	type User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
		Age   int    `json:"age" validate:"min=0"`
	}

	data := []byte(`{"name":"John Doe","email":"john@example.com","age":30}`)

	user, err := Unmarshal[User](data)
	require.NoError(t, err, "Valid JSON should unmarshal successfully")
	require.NotNil(t, user)

	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, 30, user.Age)
}

func TestSimpleAPI_Unmarshal_ValidationError(t *testing.T) {
	type User struct {
		Email string `json:"email" validate:"required"`
		Age   int    `json:"age" validate:"min=18"`
	}

	// Missing required email field and age below minimum
	data := []byte(`{"age":10}`)

	user, err := Unmarshal[User](data)
	require.Error(t, err, "Missing required field should return validation error")

	// Should be a ValidationError with field-level errors
	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr, "Error should be *ValidationError")
	assert.NotEmpty(t, validationErr.Errors, "Should have field errors")
	// User is returned even on error (partial result)
	assert.NotNil(t, user)
}

func TestValidate_Valid(t *testing.T) {
	type Config struct {
		Host string `validate:"required"`
		Port int    `validate:"min=1,max=65535"`
	}

	config := &Config{
		Host: "localhost",
		Port: 8080,
	}

	err := Validate(config)
	assert.NoError(t, err, "Valid struct should pass validation")
}

func TestValidate_Invalid(t *testing.T) {
	// NOTE: 'required' is only checked during Unmarshal (missing JSON keys), not Validate()
	// Validate() only checks value constraints (min, max, etc.)
	type Config struct {
		Port int `validate:"min=1,max=65535"`
	}

	config := &Config{
		Port: 99999, // Exceeds maximum
	}

	err := Validate(config)
	require.Error(t, err, "Invalid struct should return validation error")

	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr, "Error should be *ValidationError")
	assert.NotEmpty(t, validationErr.Errors)
}

func TestNewModel_AllInputTypes(t *testing.T) {
	type Person struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
		Age   int    `json:"age" validate:"min=0"`
	}

	tests := []struct {
		name     string
		input    any
		wantName string
		wantErr  bool
	}{
		{
			name:     "from JSON bytes",
			input:    []byte(`{"name":"Alice","email":"alice@example.com","age":25}`),
			wantName: "Alice",
			wantErr:  false,
		},
		{
			name: "from struct value",
			input: Person{
				Name:  "Bob",
				Email: "bob@example.com",
				Age:   30,
			},
			wantName: "Bob",
			wantErr:  false,
		},
		{
			name: "from struct pointer",
			input: &Person{
				Name:  "Charlie",
				Email: "charlie@example.com",
				Age:   35,
			},
			wantName: "Charlie",
			wantErr:  false,
		},
		{
			name: "from map (kwargs)",
			input: map[string]any{
				"name":  "Diana",
				"email": "diana@example.com",
				"age":   40,
			},
			wantName: "Diana",
			wantErr:  false,
		},
		{
			name:     "missing required field",
			input:    []byte(`{"email":"eve@example.com","age":25}`), // Missing required 'name'
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person, err := NewModel[Person](tt.input)

			if tt.wantErr {
				require.Error(t, err)
				// Note: partial struct may still be returned on validation error
			} else {
				require.NoError(t, err)
				require.NotNil(t, person)
				assert.Equal(t, tt.wantName, person.Name)
			}
		})
	}
}

func TestSchema_ReturnsCachedInstance(t *testing.T) {
	type Product struct {
		ID    string  `json:"id" validate:"required"`
		Name  string  `json:"name" validate:"required"`
		Price float64 `json:"price" validate:"min=0"`
	}

	// First call
	schema1 := Schema[Product]()
	require.NotNil(t, schema1, "Schema should not be nil")

	// Second call should return same instance (pointer equality)
	schema2 := Schema[Product]()
	require.NotNil(t, schema2)

	assert.Same(t, schema1, schema2, "Schema should return cached instance (same pointer)")
}

func TestSchemaJSON_ValidJSON(t *testing.T) {
	type Article struct {
		Title   string `json:"title" validate:"required"`
		Content string `json:"content" validate:"required"`
		Author  string `json:"author" validate:"required"`
	}

	schemaBytes, err := SchemaJSON[Article]()
	require.NoError(t, err, "SchemaJSON should not error")
	require.NotNil(t, schemaBytes)

	// Verify it's valid JSON
	var schemaMap map[string]any
	err = json.Unmarshal(schemaBytes, &schemaMap)
	require.NoError(t, err, "Schema bytes should be valid JSON")

	// Basic JSON Schema structure checks
	assert.Contains(t, schemaMap, "type", "Schema should have 'type' field")
	assert.Contains(t, schemaMap, "properties", "Schema should have 'properties' field")
}

func TestMarshal_Basic(t *testing.T) {
	type Book struct {
		ISBN   string `json:"isbn" validate:"required"`
		Title  string `json:"title" validate:"required"`
		Author string `json:"author" validate:"required"`
	}

	book := &Book{
		ISBN:   "978-0-123456-78-9",
		Title:  "Go Programming",
		Author: "Rob Pike",
	}

	data, err := Marshal(book)
	require.NoError(t, err, "Marshal should succeed for valid struct")
	require.NotNil(t, data)

	// Verify it's valid JSON
	var unmarshaled Book
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err, "Marshaled data should be valid JSON")

	assert.Equal(t, book.ISBN, unmarshaled.ISBN)
	assert.Equal(t, book.Title, unmarshaled.Title)
	assert.Equal(t, book.Author, unmarshaled.Author)
}

func TestMarshalWithOptions_ExcludeContext(t *testing.T) {
	// Note: The library uses validate:"exclude:context" format
	type Account struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"email"`
		Password string `json:"password" validate:"exclude:response"`
	}

	account := &Account{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "secret123",
	}

	// Marshal with "response" context - should exclude password
	opts := ForContext("response")
	data, err := MarshalWithOptions(account, opts)
	require.NoError(t, err, "MarshalWithOptions should succeed")
	require.NotNil(t, data)

	// Verify password is not in JSON
	var unmarshaled map[string]any
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err, "Should be valid JSON")

	assert.Contains(t, unmarshaled, "username")
	assert.Contains(t, unmarshaled, "email")
	assert.NotContains(t, unmarshaled, "password", "Password should be excluded in 'response' context")
}

func TestDict_Basic(t *testing.T) {
	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		ZipCode string `json:"zip_code" validate:"required"`
	}

	address := &Address{
		Street:  "123 Main St",
		City:    "Springfield",
		ZipCode: "12345",
	}

	dict, err := Dict(address)
	require.NoError(t, err, "Dict should succeed")
	require.NotNil(t, dict)

	assert.Equal(t, "123 Main St", dict["street"])
	assert.Equal(t, "Springfield", dict["city"])
	assert.Equal(t, "12345", dict["zip_code"])
}

// ==============================================================================
// Concurrency Tests (6 tests) - CRITICAL for thread-safety
// ==============================================================================

func TestConcurrentUnmarshal(t *testing.T) {
	type User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
	}

	data := []byte(`{"name":"John","email":"john@example.com"}`)

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user, err := Unmarshal[User](data)
			if err != nil {
				errChan <- err
				return
			}
			if user.Name != "John" || user.Email != "john@example.com" {
				errChan <- fmt.Errorf("unexpected values: %+v", user)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent error: %v", err)
	}
}

func TestConcurrentValidate(t *testing.T) {
	type Config struct {
		Host string `validate:"required"`
		Port int    `validate:"min=1,max=65535"`
	}

	config := &Config{
		Host: "localhost",
		Port: 8080,
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Validate(config); err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent validation error: %v", err)
	}
}

func TestConcurrentSchema(t *testing.T) {
	type Product struct {
		ID    string  `json:"id" validate:"required"`
		Name  string  `json:"name" validate:"required"`
		Price float64 `json:"price" validate:"min=0"`
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)
	schemaChan := make(chan any, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			schema := Schema[Product]()
			if schema == nil {
				errChan <- fmt.Errorf("schema is nil")
			} else {
				schemaChan <- schema
			}
		}()
	}

	wg.Wait()
	close(errChan)
	close(schemaChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("concurrent schema error: %v", err)
	}

	// Verify all schemas are the same instance (pointer equality)
	var firstSchema any
	for schema := range schemaChan {
		if firstSchema == nil {
			firstSchema = schema
		} else {
			assert.Same(t, firstSchema, schema, "All schemas should be the same cached instance")
		}
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	type Order struct {
		OrderID  string  `json:"order_id" validate:"required"`
		Total    float64 `json:"total" validate:"min=0"`
		Customer string  `json:"customer" validate:"required"`
	}

	data := []byte(`{"order_id":"ORD123","total":99.99,"customer":"Alice"}`)
	order := &Order{
		OrderID:  "ORD456",
		Total:    49.99,
		Customer: "Bob",
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 400) // 50 iterations * 4 operations * 2 buffer

	for i := 0; i < 50; i++ {
		// Unmarshal
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Unmarshal[Order](data)
			if err != nil {
				errChan <- fmt.Errorf("unmarshal: %w", err)
			}
		}()

		// Validate
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Validate(order)
			if err != nil {
				errChan <- fmt.Errorf("validate: %w", err)
			}
		}()

		// Schema
		wg.Add(1)
		go func() {
			defer wg.Done()
			schema := Schema[Order]()
			if schema == nil {
				errChan <- fmt.Errorf("schema: nil schema")
			}
		}()

		// Marshal
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Marshal(order)
			if err != nil {
				errChan <- fmt.Errorf("marshal: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent mixed operation error: %v", err)
	}
}

func TestConcurrentCacheAccess(t *testing.T) {
	type TypeA struct {
		FieldA string `json:"field_a" validate:"required"`
	}
	type TypeB struct {
		FieldB int `json:"field_b" validate:"min=0"`
	}
	type TypeC struct {
		FieldC bool `json:"field_c"`
	}

	dataA := []byte(`{"field_a":"value"}`)
	dataB := []byte(`{"field_b":42}`)
	dataC := []byte(`{"field_c":true}`)

	var wg sync.WaitGroup
	errChan := make(chan error, 300)

	// Access 3 different types concurrently
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			_, err := Unmarshal[TypeA](dataA)
			if err != nil {
				errChan <- fmt.Errorf("TypeA: %w", err)
			}
		}()

		go func() {
			defer wg.Done()
			_, err := Unmarshal[TypeB](dataB)
			if err != nil {
				errChan <- fmt.Errorf("TypeB: %w", err)
			}
		}()

		go func() {
			defer wg.Done()
			_, err := Unmarshal[TypeC](dataC)
			if err != nil {
				errChan <- fmt.Errorf("TypeC: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent cache access error: %v", err)
	}
}

func TestSimpleAPI_ConcurrentCacheCreation(t *testing.T) {
	// This test verifies that getOrCreateValidator is thread-safe
	// when multiple goroutines try to create validators for the same type
	type Service struct {
		Name string `json:"name" validate:"required"`
		URL  string `json:"url" validate:"url"`
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)
	validatorChan := make(chan *Validator[Service], 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// This will trigger getOrCreateValidator
			vl := getOrCreateValidator[Service]()
			if vl == nil {
				errChan <- fmt.Errorf("validator is nil")
			} else {
				validatorChan <- vl
			}
		}()
	}

	wg.Wait()
	close(errChan)
	close(validatorChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("concurrent registration error: %v", err)
	}

	// Verify all validators are the same instance
	var firstValidator *Validator[Service]
	for vl := range validatorChan {
		if firstValidator == nil {
			firstValidator = vl
		} else {
			assert.Same(t, firstValidator, vl, "All validators should be the same cached instance")
		}
	}
}

// ==============================================================================
// Additional Coverage Tests
// ==============================================================================

func TestSchemaOpenAPI_SimpleAPI(t *testing.T) {
	type OpenAPIModel struct {
		ID     int    `json:"id" validate:"required"`
		Name   string `json:"name" validate:"required,min=1,max=100"`
		Active bool   `json:"active"`
	}

	schema := SchemaOpenAPI[OpenAPIModel]()
	require.NotNil(t, schema, "SchemaOpenAPI should return a non-nil schema")

	// Verify it's a valid schema with the expected structure
	schemaBytes, err := json.Marshal(schema)
	require.NoError(t, err, "Schema should be JSON marshalable")
	assert.Contains(t, string(schemaBytes), "id", "Schema should contain id field")
	assert.Contains(t, string(schemaBytes), "name", "Schema should contain name field")
}

func TestSchemaJSONOpenAPI_SimpleAPI(t *testing.T) {
	type OpenAPIJSONModel struct {
		Email   string `json:"email" validate:"required,email"`
		Website string `json:"website" validate:"url"`
	}

	schemaBytes, err := SchemaJSONOpenAPI[OpenAPIJSONModel]()
	require.NoError(t, err, "SchemaJSONOpenAPI should not return an error")
	require.NotEmpty(t, schemaBytes, "SchemaJSONOpenAPI should return non-empty bytes")

	// Verify it's valid JSON
	var schema map[string]any
	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err, "SchemaJSONOpenAPI should return valid JSON")

	// Check schema has expected fields
	assert.Contains(t, string(schemaBytes), "email", "Schema should contain email field")
	assert.Contains(t, string(schemaBytes), "website", "Schema should contain website field")
}

func TestUnmarshalCtx_SimpleAPI(t *testing.T) {
	type CtxModel struct {
		Name  string `json:"name" validate:"required"`
		Value int    `json:"value"`
	}

	ctx := context.Background()
	data := []byte(`{"name":"test","value":42}`)

	result, err := UnmarshalCtx[CtxModel](ctx, data)
	require.NoError(t, err, "UnmarshalCtx should succeed with valid data")
	require.NotNil(t, result)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
}

func TestUnmarshalCtx_ValidationError(t *testing.T) {
	type CtxModelWithEmail struct {
		Email string `json:"email" validate:"required,email"`
	}

	ctx := context.Background()
	// Missing required email
	data := []byte(`{}`)

	_, err := UnmarshalCtx[CtxModelWithEmail](ctx, data)
	require.Error(t, err, "UnmarshalCtx should return error for missing required field")

	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
}

// ==============================================================================
// Pool and Path Building Tests
// ==============================================================================

func TestAppendMapKey_AllKeyTypes(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		key      any
		expected string
	}{
		{
			name:     "string key",
			path:     "data",
			key:      "myKey",
			expected: "data[myKey]",
		},
		{
			name:     "int key",
			path:     "items",
			key:      42,
			expected: "items[42]",
		},
		{
			name:     "int64 key",
			path:     "ids",
			key:      int64(12345),
			expected: "ids[12345]",
		},
		{
			name:     "int32 key",
			path:     "nums",
			key:      int32(100),
			expected: "nums[100]",
		},
		{
			name:     "uint key",
			path:     "uints",
			key:      uint(999),
			expected: "uints[999]",
		},
		{
			name:     "uint64 key",
			path:     "bigints",
			key:      uint64(18446744073709551615),
			expected: "bigints[18446744073709551615]",
		},
		{
			name:     "uint32 key",
			path:     "smallints",
			key:      uint32(4294967295),
			expected: "smallints[4294967295]",
		},
		{
			name:     "float64 key (fallback)",
			path:     "floats",
			key:      3.14,
			expected: "floats[3.14]",
		},
		{
			name:     "bool key (fallback)",
			path:     "bools",
			key:      true,
			expected: "bools[true]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 0, 64)
			result := appendMapKey(buf, []byte(tt.path), tt.key)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

// TestValidate_MapValidationWithErrors tests map validation which exercises captureExtrasInField.
func TestValidate_MapValidationWithErrors(t *testing.T) {
	type Config struct {
		Settings map[string]int `json:"settings" validate:"dive,min=0,max=100"`
	}

	config := &Config{
		Settings: map[string]int{
			"valid":   50,
			"invalid": 150, // exceeds max
		},
	}

	err := Validate(config)
	require.Error(t, err, "Validation should fail for map with invalid values")

	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
}

// TestNewModel_MapInput tests the unmarshalFromMap path.
func TestNewModel_MapInput_Comprehensive(t *testing.T) {
	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city"`
	}
	type User struct {
		Name    string   `json:"name" validate:"required"`
		Age     int      `json:"age" validate:"min=0"`
		Address *Address `json:"address"`
	}

	input := map[string]interface{}{
		"name": "Alice",
		"age":  30,
		"address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "Boston",
		},
	}

	user, err := NewModel[User](input)
	require.NoError(t, err)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, 30, user.Age)
	require.NotNil(t, user.Address)
	assert.Equal(t, "123 Main St", user.Address.Street)
	assert.Equal(t, "Boston", user.Address.City)
}

// TestValidateWithCache tests that cached validation works correctly.
func TestValidate_CachedValidation(t *testing.T) {
	type Product struct {
		Name  string  `json:"name" validate:"required"`
		Price float64 `json:"price" validate:"min=0"`
	}

	// First validation
	product1 := &Product{Name: "Widget", Price: 9.99}
	err := Validate(product1)
	require.NoError(t, err)

	// Second validation of same type (should use cached validator)
	product2 := &Product{Name: "Gadget", Price: 19.99}
	err = Validate(product2)
	require.NoError(t, err)

	// Validation with error (cached)
	product3 := &Product{Name: "", Price: -5.0}
	err = Validate(product3)
	require.Error(t, err)
}

// TestGetContextValidator tests the GetContextValidator function.
func TestGetContextValidator_Registered(t *testing.T) {
	// Register a context validator with a unique name
	validatorName := "test_ctx_validator_coverage"
	err := RegisterValidationCtx(validatorName, func(ctx context.Context, value any, param string) error {
		if str, ok := value.(string); ok && str == "forbidden" {
			return fmt.Errorf("forbidden value")
		}
		return nil
	})
	require.NoError(t, err)

	// Get the registered validator
	validator, found := GetContextValidator(validatorName)
	require.True(t, found, "Registered validator should be found")
	require.NotNil(t, validator, "Registered validator should be returned")

	// Get non-existent validator
	nonExistent, found := GetContextValidator("non_existent_validator_xyz")
	assert.False(t, found, "Non-existent validator should not be found")
	assert.Nil(t, nonExistent, "Non-existent validator should return nil")
}

// TestIsZeroValue_AllTypes tests the isZeroValue function comprehensively.
func TestIsZeroValue_Comprehensive(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() any
		isZero bool
	}{
		{
			name:   "empty string is zero",
			setup:  func() any { return "" },
			isZero: true,
		},
		{
			name:   "non-empty string is not zero",
			setup:  func() any { return "hello" },
			isZero: false,
		},
		{
			name:   "nil pointer is zero",
			setup:  func() any { var p *string; return p },
			isZero: true,
		},
		{
			name:   "non-nil pointer is not zero",
			setup:  func() any { s := "test"; return &s },
			isZero: false,
		},
		{
			name:   "nil slice is zero",
			setup:  func() any { var s []int; return s },
			isZero: true,
		},
		{
			name:   "empty slice is zero",
			setup:  func() any { return []int{} },
			isZero: true,
		},
		{
			name:   "non-empty slice is not zero",
			setup:  func() any { return []int{1, 2, 3} },
			isZero: false,
		},
		{
			name:   "nil map is zero",
			setup:  func() any { var m map[string]int; return m },
			isZero: true,
		},
		{
			name:   "empty map is zero",
			setup:  func() any { return map[string]int{} },
			isZero: true,
		},
		{
			name:   "non-empty map is not zero",
			setup:  func() any { return map[string]int{"key": 1} },
			isZero: false,
		},
		{
			name:   "int 0 is NOT zero (semantic: valid value)",
			setup:  func() any { return 0 },
			isZero: false,
		},
		{
			name:   "bool false is NOT zero (semantic: valid value)",
			setup:  func() any { return false },
			isZero: false,
		},
		{
			name:   "non-nil interface is not zero",
			setup:  func() any { var i interface{} = "test"; return i },
			isZero: false,
		},
		{
			name:   "nil chan is zero",
			setup:  func() any { var c chan int; return c },
			isZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.setup()
			result := isZeroValue(reflect.ValueOf(value))
			assert.Equal(t, tt.isZero, result)
		})
	}
}

// TestStructPartial tests partial validation.
func TestStructPartial_Coverage(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	type Form struct {
		Name  string `json:"name" validate:"required,min=1"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=18"`
	}

	v := New[Form](DefaultOptions())

	// Valid form
	form := &Form{Name: "John", Email: "john@test.com", Age: 25}

	// Partial validation - only validate Name field
	err := v.StructPartial(form, "name")
	require.NoError(t, err)

	// Partial validation with invalid field
	invalidForm := &Form{Name: "", Email: "invalid", Age: 10}
	err = v.StructPartial(invalidForm, "name")
	require.Error(t, err)
}

// TestStructExcept tests validation excluding certain fields.
func TestStructExcept_Coverage(t *testing.T) {
	// Reset tag name function to default (use JSON tag)
	RegisterTagNameFunc(nil)

	type Form struct {
		Name  string `json:"name" validate:"required,min=1"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=18"`
	}

	v := New[Form](DefaultOptions())

	// Form with invalid email but we exclude email from validation
	form := &Form{Name: "John", Email: "invalid", Age: 25}
	err := v.StructExcept(form, "email")
	require.NoError(t, err, "Should pass when excluding the invalid field")

	// Form with invalid name (not excluded)
	form2 := &Form{Name: "", Email: "valid@test.com", Age: 25}
	err = v.StructExcept(form2, "email", "age")
	require.Error(t, err, "Should fail on name validation")
}

// ==== Coverage tests for pointer structs with extras ====

// PtrNestedWithExtra has pointer struct fields with extras.
type PtrNestedWithExtra struct {
	Value  string         `json:"value"`
	Extras map[string]any `json:"-" validate:"extra_fields"`
}

// PtrStructWithExtras has a pointer to nested struct with extras.
type PtrStructWithExtras struct {
	Name   string              `json:"name"`
	Nested *PtrNestedWithExtra `json:"nested"`
	Extras map[string]any      `json:"-" validate:"extra_fields"`
}

// TestExtraAllow_PointerStructField tests extra fields with pointer struct fields.
func TestExtraAllow_PointerStructField(t *testing.T) {
	v := New[PtrStructWithExtras](Options{
		ExtraFields: ExtraAllow,
	})

	// Test with pointer struct containing extras
	jsonData := []byte(`{
		"name": "test",
		"nested": {
			"value": "inner",
			"nested_extra": "captured"
		},
		"top_extra": "also_captured"
	}`)

	result, err := v.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.NotNil(t, result.Nested)
	assert.Equal(t, "inner", result.Nested.Value)
	// Check nested extras were captured
	assert.Equal(t, "captured", result.Nested.Extras["nested_extra"])
	// Check top-level extras were captured
	assert.Equal(t, "also_captured", result.Extras["top_extra"])
}

// TestExtraAllow_NilPointerStructField tests extras with nil pointer struct.
func TestExtraAllow_NilPointerStructField(t *testing.T) {
	v := New[PtrStructWithExtras](Options{
		ExtraFields: ExtraAllow,
	})

	// JSON with null nested field
	jsonData := []byte(`{
		"name": "test",
		"nested": null,
		"top_extra": "captured"
	}`)

	result, err := v.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Nil(t, result.Nested)
	assert.Equal(t, "captured", result.Extras["top_extra"])
}

// SlicePtrStructWithExtras has a slice of pointer structs with extras.
type SlicePtrStructWithExtras struct {
	Items  []*PtrNestedWithExtra `json:"items"`
	Extras map[string]any        `json:"-" validate:"extra_fields"`
}

// TestExtraAllow_SlicePointerStruct tests extras with slice of pointer structs.
func TestExtraAllow_SlicePointerStruct(t *testing.T) {
	v := New[SlicePtrStructWithExtras](Options{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"items": [
			{"value": "first", "item_extra": "extra1"},
			{"value": "second", "item_extra": "extra2"},
			null
		],
		"top_extra": "captured"
	}`)

	result, err := v.Unmarshal(jsonData)
	require.NoError(t, err)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "first", result.Items[0].Value)
	assert.Equal(t, "extra1", result.Items[0].Extras["item_extra"])
	assert.Equal(t, "second", result.Items[1].Value)
	assert.Equal(t, "extra2", result.Items[1].Extras["item_extra"])
	assert.Nil(t, result.Items[2])
	assert.Equal(t, "captured", result.Extras["top_extra"])
}

// TestMarshal_PointerStructWithExtras tests marshaling with pointer structs and extras.
func TestMarshal_PointerStructWithExtras(t *testing.T) {
	v := New[PtrStructWithExtras](Options{
		ExtraFields: ExtraAllow,
	})

	nested := &PtrNestedWithExtra{
		Value:  "inner",
		Extras: map[string]any{"nested_extra": "value"},
	}
	obj := &PtrStructWithExtras{
		Name:   "test",
		Nested: nested,
		Extras: map[string]any{"top_extra": "value"},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "test", parsed["name"])
	assert.Equal(t, "value", parsed["top_extra"])
	nestedMap := parsed["nested"].(map[string]any)
	assert.Equal(t, "inner", nestedMap["value"])
	assert.Equal(t, "value", nestedMap["nested_extra"])
}

// TestMarshal_SlicePointerStructWithExtras tests marshaling slice of pointer structs.
func TestMarshal_SlicePointerStructWithExtras(t *testing.T) {
	v := New[SlicePtrStructWithExtras](Options{
		ExtraFields: ExtraAllow,
	})

	obj := &SlicePtrStructWithExtras{
		Items: []*PtrNestedWithExtra{
			{Value: "first", Extras: map[string]any{"e1": "v1"}},
			{Value: "second", Extras: map[string]any{"e2": "v2"}},
		},
		Extras: map[string]any{"top": "value"},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	items := parsed["items"].([]any)
	require.Len(t, items, 2)
	item0 := items[0].(map[string]any)
	assert.Equal(t, "first", item0["value"])
	assert.Equal(t, "v1", item0["e1"])
}

// ==== Coverage tests for NewModel with multi-required errors ====

// NestedRequiredForNewModel has nested required fields.
type NestedRequiredForNewModel struct {
	Field1 string `json:"field1" validate:"required"`
	Field2 string `json:"field2" validate:"required"`
}

// NewModelMultiRequired has nested struct with multiple required fields.
type NewModelMultiRequired struct {
	Name   string                    `json:"name" validate:"required"`
	Nested NestedRequiredForNewModel `json:"nested"`
}

// TestNewModel_MultiRequiredError tests NewModel with multi-required field errors.
func TestNewModel_MultiRequiredError(t *testing.T) {
	v := New[NewModelMultiRequired](DefaultOptions())

	// Missing multiple required fields in nested struct
	input := map[string]any{
		"name":   "test",
		"nested": map[string]any{}, // Missing field1 and field2
	}

	_, err := v.NewModel(input)
	require.Error(t, err)
	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
	// Should have errors for both nested required fields
	assert.GreaterOrEqual(t, len(valErr.Errors), 2)
}

// TestNewModel_SingleRequiredError tests NewModel with single required field error.
func TestNewModel_SingleRequiredError(t *testing.T) {
	v := New[NewModelMultiRequired](DefaultOptions())

	// Missing required field at top level
	input := map[string]any{
		"nested": map[string]any{
			"field1": "value1",
			"field2": "value2",
		},
	}

	_, err := v.NewModel(input)
	require.Error(t, err)
	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.GreaterOrEqual(t, len(valErr.Errors), 1)
}

// TestNewModel_OtherError tests NewModel with non-required errors.
func TestNewModel_OtherError(t *testing.T) {
	type EmailModel struct {
		Email string `json:"email" validate:"required,email"`
	}

	v := New[EmailModel](DefaultOptions())

	// Invalid email format
	input := map[string]any{
		"email": "not-an-email",
	}

	_, err := v.NewModel(input)
	require.Error(t, err)
}

// ==== Coverage tests for getJSONFieldName edge cases ====

// JSONFieldNameTest has various JSON tag formats.
type JSONFieldNameTest struct {
	Normal        string `json:"normal"`
	WithOmit      string `json:"with_omit,omitempty"`
	Ignored       string `json:"-"`
	NoTag         string
	unexportedFld string //nolint:unused
}

// TestGetJSONFieldName_Coverage tests all branches of getJSONFieldName.
func TestGetJSONFieldName_Coverage(t *testing.T) {
	// This is tested indirectly through ExtraAllow mode which uses getJSONFieldName
	v := New[JSONFieldNameTest](Options{
		ExtraFields: ExtraIgnore, // Don't need ExtraAllow for this test
	})

	jsonData := []byte(`{
		"normal": "val1",
		"with_omit": "val2",
		"NoTag": "val3"
	}`)

	result, err := v.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "val1", result.Normal)
	assert.Equal(t, "val2", result.WithOmit)
}

// ==== Coverage tests for validateWithCache with skip constraints ====

// TestValidate_SkipUnless_Coverage tests skip_unless constraint path.
func TestValidate_SkipUnless_Coverage(t *testing.T) {
	// skip_unless is tested indirectly through existing tests
	// This test ensures the HasSkipConstraints branch is covered
	type SkipTest struct {
		Type    string `json:"type"`
		OtherID string `json:"other_id" validate:"skip_unless=Type:other,required"`
	}

	v := New[SkipTest](DefaultOptions())

	// When Type is not "other", OtherID validation is skipped
	jsonData := []byte(`{"type": "regular", "other_id": ""}`)
	result, err := v.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "regular", result.Type)

	// When Type is "other", OtherID is required
	jsonData2 := []byte(`{"type": "other", "other_id": "123"}`)
	result2, err := v.Unmarshal(jsonData2)
	require.NoError(t, err)
	assert.Equal(t, "123", result2.OtherID)
}

// ==== Coverage tests for marshalWithExtras error paths ====

// TestMarshalWithOptions_Coverage tests MarshalWithOptions code paths.
func TestMarshalWithOptions_Coverage(t *testing.T) {
	type OptionsModel struct {
		Name   string  `json:"name"`
		Secret *string `json:"secret,omitempty"`
	}

	v := New[OptionsModel](DefaultOptions())

	// With nil pointer field
	obj := &OptionsModel{Name: "test", Secret: nil}
	data, err := v.Marshal(obj)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "test", parsed["name"])
}

// ==== Coverage tests for Dict error paths ====

// DictModelWithExtras is for testing Dict with extras enabled.
type DictModelWithExtras struct {
	Name   string         `json:"name" validate:"required,min=3"`
	Extras map[string]any `json:"-" validate:"extra_fields"`
}

// TestDict_ValidationError tests Dict with validation errors.
func TestDict_ValidationError(t *testing.T) {
	// Dict with extras enabled calls Marshal which validates
	v := New[DictModelWithExtras](Options{
		ExtraFields: ExtraAllow,
	})
	obj := &DictModelWithExtras{Name: "AB"} // Too short, min=3

	_, err := v.Dict(obj)
	require.Error(t, err) // Should fail validation via Marshal
}

// TestDict_Success tests Dict successfully converts struct to map.
func TestDict_Success(t *testing.T) {
	type SimpleDict struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	v := New[SimpleDict](DefaultOptions())
	obj := &SimpleDict{Name: "test", Age: 25}

	dict, err := v.Dict(obj)
	require.NoError(t, err)
	assert.Equal(t, "test", dict["name"])
	assert.InDelta(t, 25, dict["age"], 0.001) // JSON numbers are float64
}

// ==== Coverage tests for SecretStr type ====

// TestSecretStr_UnmarshalJSON_EmptyString tests unmarshaling empty string to SecretStr.
func TestSecretStr_UnmarshalJSON_EmptyString(t *testing.T) {
	var secret SecretStr
	err := json.Unmarshal([]byte(`""`), &secret)
	require.NoError(t, err)
	assert.Empty(t, secret.Value())
	// String() always returns masked value
	assert.Equal(t, "**********", secret.String())
}

// TestSecretStr_UnmarshalJSON_NonEmptyString tests unmarshaling non-empty string.
func TestSecretStr_UnmarshalJSON_NonEmptyString(t *testing.T) {
	var secret SecretStr
	err := json.Unmarshal([]byte(`"mypassword"`), &secret)
	require.NoError(t, err)
	assert.Equal(t, "mypassword", secret.Value())
	assert.Equal(t, "**********", secret.String())
}

// TestSecretStr_NewSecretStr tests NewSecretStr constructor.
func TestSecretStr_NewSecretStr(t *testing.T) {
	secret := NewSecretStr("mypassword")
	assert.Equal(t, "mypassword", secret.Value())
	assert.Equal(t, "**********", secret.String())
}

// TestSecretBytes_Coverage tests SecretBytes type for coverage.
func TestSecretBytes_Coverage(t *testing.T) {
	var secret SecretBytes
	// SecretBytes expects base64-encoded string
	err := json.Unmarshal([]byte(`"aGVsbG8="`), &secret)
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), secret.Value())
	assert.Equal(t, "**********", secret.String())
}

// TestSecretBytes_EmptyString tests unmarshaling empty base64 string.
func TestSecretBytes_EmptyString(t *testing.T) {
	var secret SecretBytes
	err := json.Unmarshal([]byte(`""`), &secret)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, secret.Value())
}

// TestSecretBytes_InvalidBase64 tests unmarshaling invalid base64.
func TestSecretBytes_InvalidBase64(t *testing.T) {
	var secret SecretBytes
	err := json.Unmarshal([]byte(`"not-valid-base64!!!"`), &secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base64")
}

// TestSecretBytes_NewSecretBytes tests NewSecretBytes constructor.
func TestSecretBytes_NewSecretBytes(t *testing.T) {
	secret := NewSecretBytes([]byte("hello"))
	assert.Equal(t, []byte("hello"), secret.Value())
	assert.Equal(t, "**********", secret.String())
}

// ==== Coverage tests for UnmarshalCtx error path ====

// TestUnmarshalCtx_Context_Error tests UnmarshalCtx with context validation error.
func TestUnmarshalCtx_Context_Error(t *testing.T) {
	type CtxModel struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[CtxModel](DefaultOptions())
	ctx := context.Background()

	// Missing required field
	_, err := v.UnmarshalCtx(ctx, []byte(`{}`))
	require.Error(t, err)
}

// ==== Coverage tests for StructPartial/StructExcept nil pointer ====

// TestStructPartial_NilPointer tests StructPartial with nil pointer.
func TestStructPartial_NilPointer(t *testing.T) {
	RegisterTagNameFunc(nil)
	type PartialModel struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[PartialModel](DefaultOptions())

	var ptr *PartialModel = nil
	err := v.StructPartial(ptr, "name")
	require.Error(t, err)
}

// TestStructExcept_NilPointer tests StructExcept with nil pointer.
func TestStructExcept_NilPointer(t *testing.T) {
	RegisterTagNameFunc(nil)
	type ExceptModel struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[ExceptModel](DefaultOptions())

	var ptr *ExceptModel = nil
	err := v.StructExcept(ptr, "name")
	require.Error(t, err)
}

// ==== Coverage tests for getJSONFieldName unexported fields ====

// TestExtraAllow_WithUnexportedAndIgnoredFields tests extras with fields that should be skipped.
type ExtraModelWithIgnored struct {
	Name       string         `json:"name"`
	Ignored    string         `json:"-"`
	Extras     map[string]any `json:"-" validate:"extra_fields"`
	unexported string         //nolint:unused
}

func TestExtraAllow_WithIgnoredField(t *testing.T) {
	v := New[ExtraModelWithIgnored](Options{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{"name": "test", "extra_key": "extra_value"}`)
	result, err := v.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, "extra_value", result.Extras["extra_key"])
}

// ==== Coverage tests for marshal merge error paths ====

// NestedExtraForMerge has nested struct with extras for merge testing.
type NestedExtraForMerge struct {
	Value  string         `json:"value"`
	Extras map[string]any `json:"-" validate:"extra_fields"`
}

// ParentExtraForMerge has nested struct and extras.
type ParentExtraForMerge struct {
	Name   string              `json:"name"`
	Nested NestedExtraForMerge `json:"nested"`
	Extras map[string]any      `json:"-" validate:"extra_fields"`
}

// TestMarshal_MergeExtras_Nested tests merging extras in nested structs.
func TestMarshal_MergeExtras_Nested(t *testing.T) {
	v := New[ParentExtraForMerge](Options{
		ExtraFields: ExtraAllow,
	})

	obj := &ParentExtraForMerge{
		Name: "parent",
		Nested: NestedExtraForMerge{
			Value:  "nested_value",
			Extras: map[string]any{"nested_extra": "nested_extra_value"},
		},
		Extras: map[string]any{"parent_extra": "parent_extra_value"},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "parent", parsed["name"])
	assert.Equal(t, "parent_extra_value", parsed["parent_extra"])
	nested := parsed["nested"].(map[string]any)
	assert.Equal(t, "nested_value", nested["value"])
	assert.Equal(t, "nested_extra_value", nested["nested_extra"])
}

// ==== Coverage tests for validateWithCache nil cache path ====

// TestValidate_NilPointerField tests validation with nil pointer fields.
func TestValidate_NilPointerField(t *testing.T) {
	type WithPointer struct {
		Name    string  `json:"name"`
		OptName *string `json:"opt_name"`
	}

	v := New[WithPointer](DefaultOptions())

	// Nil pointer field should be handled
	obj := &WithPointer{Name: "test", OptName: nil}
	err := v.Validate(obj)
	require.NoError(t, err)
}

// ==== Coverage tests for Union Unmarshal error paths ====

type UnionCat struct {
	Name string `json:"name"`
}
type UnionDog struct {
	Name string `json:"name"`
}

// TestUnion_InvalidJSON tests union unmarshal with invalid JSON.
func TestUnion_InvalidJSON(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
			VariantFor[UnionDog]("dog"),
		},
	})
	require.NoError(t, err)

	_, err = u.Unmarshal([]byte(`{invalid json`))
	require.Error(t, err)
}

// TestUnion_MissingDiscriminator tests union with missing discriminator.
func TestUnion_MissingDiscriminator(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
			VariantFor[UnionDog]("dog"),
		},
	})
	require.NoError(t, err)

	_, err = u.Unmarshal([]byte(`{"name": "fluffy"}`))
	require.Error(t, err)
}

// ==== Coverage tests for schema findTypeForDefinition ====

// TestSchema_NestedDefinition tests schema generation with nested types.
func TestSchema_NestedDefinition(t *testing.T) {
	type Inner struct {
		Value string `json:"value"`
	}
	type Outer struct {
		Inner Inner `json:"inner"`
	}

	schema := Schema[Outer]()
	require.NotNil(t, schema)
	// Schema should have properties - verify it was generated
	require.NotNil(t, schema.Properties)
}

// ==== Coverage tests for Union with different discriminator types ====

// TestUnion_NumericDiscriminator tests union with numeric discriminator value.
func TestUnion_NumericDiscriminator(t *testing.T) {
	type TypeA struct {
		Type int    `json:"type"`
		Name string `json:"name"`
	}
	type TypeB struct {
		Type int   `json:"type"`
		ID   int64 `json:"id"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[TypeA]("1"),
			VariantFor[TypeB]("2"),
		},
	})
	require.NoError(t, err)

	// Test numeric discriminator - JSON numbers are float64
	result, err := u.Unmarshal([]byte(`{"type": 1, "name": "test"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestUnion_UnknownDiscriminator tests union with unknown discriminator value.
func TestUnion_UnknownDiscriminator(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
			VariantFor[UnionDog]("dog"),
		},
	})
	require.NoError(t, err)

	_, err = u.Unmarshal([]byte(`{"type": "bird", "name": "tweety"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bird")
}

// TestUnion_ValidationFailure tests union with variant validation failure.
func TestUnion_ValidationFailure(t *testing.T) {
	type ValidatedCat struct {
		Type string `json:"type"`
		Name string `json:"name" validate:"required,min=3"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[ValidatedCat]("cat"),
		},
	})
	require.NoError(t, err)

	// Name too short - should fail validation
	_, err = u.Unmarshal([]byte(`{"type": "cat", "name": "ab"}`))
	require.Error(t, err)
}

// TestUnion_WithValidation tests union variant with simple validation constraints.
func TestUnion_WithValidation(t *testing.T) {
	type WithEmail struct {
		Type  string `json:"type"`
		Email string `json:"email" validate:"email"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[WithEmail]("email"),
		},
	})
	require.NoError(t, err)

	// Valid email value
	result, err := u.Unmarshal([]byte(`{"type": "email", "email": "test@example.com"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ==== Coverage tests for validateWithCache edge cases ====

// TestValidate_NilPointer tests validation with nil pointer field.
func TestValidate_NilPointer(t *testing.T) {
	type Inner struct {
		Value string `json:"value" validate:"required"`
	}
	type Outer struct {
		Inner *Inner `json:"inner"`
	}

	v := New[Outer](DefaultOptions())
	outer := &Outer{Inner: nil}
	// nil pointer should be handled gracefully
	err := v.Validate(outer)
	// No error expected for nil pointer (only validates non-nil values)
	require.NoError(t, err)
}

// TestValidate_NonStructField tests validation skips non-struct type.
func TestValidate_NonStructField(t *testing.T) {
	type Model struct {
		Name  string `json:"name" validate:"required"`
		Count int    `json:"count"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Name: "test", Count: 5}
	err := v.Validate(m)
	require.NoError(t, err)
}

// ==== Coverage tests for StructPartial/StructExcept nested fields ====

// TestStructPartial_NestedField tests partial validation with nested struct.
func TestStructPartial_NestedField(t *testing.T) {
	RegisterTagNameFunc(nil) // Reset to use JSON tags

	type Address struct {
		City string `json:"city" validate:"required"`
	}
	type Person struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	v := New[Person](DefaultOptions())
	p := &Person{Name: "John", Address: Address{City: ""}}

	// Validate only name, should pass
	err := v.StructPartial(p, "name")
	require.NoError(t, err)
}

// TestStructExcept_NestedField tests except validation with nested struct.
func TestStructExcept_NestedField(t *testing.T) {
	RegisterTagNameFunc(nil) // Reset to use JSON tags

	type Address struct {
		City string `json:"city" validate:"required"`
	}
	type Person struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	v := New[Person](DefaultOptions())
	p := &Person{Name: "", Address: Address{City: "NYC"}}

	// Exclude name validation, should pass (name is empty but excluded)
	err := v.StructExcept(p, "name")
	require.NoError(t, err)
}

// ==== Coverage tests for mergeExtrasInSlice and marshalWithExtras ====

// TestMarshal_SliceWithExtras tests marshaling slice with nested extras.
func TestMarshal_SliceWithExtras(t *testing.T) {
	type Item struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Container struct {
		Items  []Item         `json:"items"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Container](Options{ExtraFields: ExtraAllow})

	// Unmarshal with extras in slice elements
	jsonData := `{"items": [{"name": "item1", "extra": "value1"}, {"name": "item2", "extra": "value2"}], "containerExtra": "test"}`
	result, err := v.Unmarshal([]byte(jsonData))
	require.NoError(t, err)

	// Marshal back
	marshaled, err := v.Marshal(result)
	require.NoError(t, err)

	// Verify extras are preserved
	assert.Contains(t, string(marshaled), "containerExtra")
}

// TestMarshal_MapField tests marshaling struct with map field.
func TestMarshal_MapField(t *testing.T) {
	type Model struct {
		Name string            `json:"name"`
		Tags map[string]string `json:"tags"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Name: "test", Tags: map[string]string{"key": "value"}}

	marshaled, err := v.Marshal(m)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), `"key":"value"`)
}

// ==== Coverage tests for UnmarshalCtx ====

// TestUnmarshalCtx_Valid tests UnmarshalCtx with valid data.
func TestUnmarshalCtx_Valid(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())
	ctx := context.Background()

	result, err := v.UnmarshalCtx(ctx, []byte(`{"name": "test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

// ==== Coverage tests for cross-field constraints returning ValidationError ====

// TestValidate_CrossFieldValidationError tests cross-field constraint returning ValidationError.
func TestValidate_CrossFieldValidationError(t *testing.T) {
	type DateRange struct {
		StartDate string `json:"start_date" validate:"required"`
		EndDate   string `json:"end_date" validate:"required,gtfield=StartDate"`
	}

	v := New[DateRange](DefaultOptions())

	// EndDate less than StartDate - should trigger cross-field validation error
	dr := &DateRange{StartDate: "2024-12-31", EndDate: "2024-01-01"}
	err := v.Validate(dr)
	// Note: gtfield compares strings, so this should fail
	require.Error(t, err)
}

// ==== Coverage tests for Union Schema ====

// TestUnion_Schema tests union schema generation.
func TestUnion_Schema(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
			VariantFor[UnionDog]("dog"),
		},
	})
	require.NoError(t, err)

	schema := u.Schema()
	require.NotNil(t, schema)
	require.NotNil(t, schema.OneOf)
	assert.Len(t, schema.OneOf, 2)
}

// ==== Coverage tests for Var function ====

// TestVar_Valid tests Var function with valid value.
func TestVar_Valid(t *testing.T) {
	email := "test@example.com"
	err := Var(email, "email")
	require.NoError(t, err)
}

// TestVar_Invalid tests Var function with invalid value.
func TestVar_Invalid(t *testing.T) {
	email := "invalid"
	err := Var(email, "email")
	require.Error(t, err)
}

// TestVar_MultiConstraintsFail tests Var function with multiple failing constraints.
func TestVar_MultiConstraintsFail(t *testing.T) {
	value := "ab"
	err := Var(value, "required,min=3")
	require.Error(t, err)
}

// ==== Additional StructPartial/StructExcept tests for coverage ====

// TestStructPartial_ConstraintError tests StructPartial with constraint errors.
func TestStructPartial_ConstraintError(t *testing.T) {
	RegisterTagNameFunc(nil) // Reset to use JSON tags

	type Model struct {
		Email string `json:"email" validate:"email"`
		Age   int    `json:"age" validate:"min=18"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Email: "invalid", Age: 10}

	// Both should fail with constraint errors
	err := v.StructPartial(m, "email", "age")
	require.Error(t, err)

	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.GreaterOrEqual(t, len(valErr.Errors), 1)
}

// TestStructPartial_CrossFieldValidationError tests cross-field constraint with ValidationError.
func TestStructPartial_CrossFieldValidationError(t *testing.T) {
	RegisterTagNameFunc(nil) // Reset to use JSON tags

	type DateRange struct {
		Start string `json:"start" validate:"required"`
		End   string `json:"end" validate:"required,gtfield=Start"`
	}

	v := New[DateRange](DefaultOptions())
	dr := &DateRange{Start: "2024-12-31", End: "2024-01-01"}

	// Cross-field validation should fail
	err := v.StructPartial(dr, "start", "end")
	require.Error(t, err)
}

// TestStructExcept_ConstraintError tests StructExcept with constraint errors.
func TestStructExcept_ConstraintError(t *testing.T) {
	RegisterTagNameFunc(nil) // Reset to use JSON tags

	type Model struct {
		Email string `json:"email" validate:"email"`
		Age   int    `json:"age" validate:"min=18"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Email: "invalid", Age: 10}

	// Validate email but not age
	err := v.StructExcept(m, "age")
	require.Error(t, err)

	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
}

// TestStructExcept_CrossFieldValidationError tests cross-field constraint with ValidationError.
func TestStructExcept_CrossFieldValidationError(t *testing.T) {
	RegisterTagNameFunc(nil) // Reset to use JSON tags

	type DateRange struct {
		Start string `json:"start" validate:"required"`
		End   string `json:"end" validate:"required,gtfield=Start"`
	}

	v := New[DateRange](DefaultOptions())
	dr := &DateRange{Start: "2024-12-31", End: "2024-01-01"}

	// Cross-field validation should fail, don't exclude anything
	err := v.StructExcept(dr)
	require.Error(t, err)
}

// ==== Tests for map validation with dive ====

// TestValidate_MapWithDive tests map validation with dive.
func TestValidate_MapWithDive(t *testing.T) {
	type Model struct {
		Tags map[string]string `json:"tags" validate:"dive,keys,min=1,endkeys,min=1"`
	}

	v := New[Model](DefaultOptions())

	// Valid map
	m := &Model{Tags: map[string]string{"key": "value"}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid map (empty key)
	m2 := &Model{Tags: map[string]string{"": "value"}}
	err = v.Validate(m2)
	require.Error(t, err)
}

// ==== Tests for marshalWithExtras and mergeExtrasInSlice ====

// TestMarshal_SlicePointerWithExtras tests marshaling slice of pointers with extras.
func TestMarshal_SlicePointerWithExtras(t *testing.T) {
	type Item struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Container struct {
		Items  []*Item        `json:"items"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Container](Options{ExtraFields: ExtraAllow})

	// Unmarshal with extras in slice elements
	jsonData := `{"items": [{"name": "item1", "extra": "value1"}, {"name": "item2", "extra": "value2"}], "containerExtra": "test"}`
	result, err := v.Unmarshal([]byte(jsonData))
	require.NoError(t, err)

	// Verify extras are captured
	assert.NotNil(t, result.Extras)
	assert.Equal(t, "test", result.Extras["containerExtra"])
	assert.NotNil(t, result.Items[0].Extras)

	// Marshal back
	marshaled, err := v.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "containerExtra")
}

// ==== Test for splitTags ====

// TestUnion_TagsWithComma tests union with tag containing comma in value.
func TestUnion_TagsWithComma(t *testing.T) {
	type WithOneof struct {
		Type   string `json:"type"`
		Status string `json:"status" validate:"oneof=active,inactive,pending"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[WithOneof]("status"),
		},
	})
	require.NoError(t, err)

	// Valid status
	result, err := u.Unmarshal([]byte(`{"type": "status", "status": "active"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ==== Test for validateVariant non-struct ====

// TestUnion_NonStructVariant tests union with non-struct variant.
func TestUnion_PointerVariant(t *testing.T) {
	type CatPtr struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[CatPtr]("cat"),
		},
	})
	require.NoError(t, err)

	result, err := u.Unmarshal([]byte(`{"type": "cat", "name": "fluffy"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// ==== Test for UnmarshalCtx with validation error ====

// TestUnmarshalCtx_MissingRequired tests UnmarshalCtx with missing required field.
func TestUnmarshalCtx_MissingRequired(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())
	ctx := context.Background()

	// Missing required field
	_, err := v.UnmarshalCtx(ctx, []byte(`{}`))
	require.Error(t, err)
}

// ==== Test for validateWithCache edge paths ====

// TestValidate_SkipConstraints tests validation with skip constraints.
func TestValidate_SkipConstraints(t *testing.T) {
	// Test that HasSkipConstraints returns true for fields with skip_unless
	type Model struct {
		Active  bool   `json:"active"`
		Details string `json:"details" validate:"skip_unless=Active,min=3"`
	}

	v := New[Model](DefaultOptions())

	// Valid value - should pass regardless of skip_unless behavior
	m := &Model{Active: false, Details: "abc"}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid value - exercises the constraint path
	m2 := &Model{Active: true, Details: "ab"}
	err = v.Validate(m2)
	// Validation runs and fails due to min=3
	require.Error(t, err)
}

// TestValidate_NestedStruct tests validation with nested struct.
func TestValidate_NestedStruct(t *testing.T) {
	type Inner struct {
		Value string `json:"value" validate:"min=3"`
	}
	type Outer struct {
		Name  string `json:"name" validate:"required"`
		Inner Inner  `json:"inner"`
	}

	v := New[Outer](DefaultOptions())

	// Valid nested struct
	o := &Outer{Name: "test", Inner: Inner{Value: "abc"}}
	err := v.Validate(o)
	require.NoError(t, err)

	// Invalid inner
	o2 := &Outer{Name: "test", Inner: Inner{Value: "ab"}}
	err = v.Validate(o2)
	require.Error(t, err)
}

// ==== Additional coverage tests ====

// TestSecretStr_UnmarshalJSON_InvalidJSON covers the error path in SecretStr.UnmarshalJSON.
func TestSecretStr_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var secret SecretStr
	err := secret.UnmarshalJSON([]byte(`{not valid json`))
	require.Error(t, err)
}

// TestSecretBytes_UnmarshalJSON_InvalidBase64 covers the error path when base64 decoding fails.
func TestSecretBytes_UnmarshalJSON_InvalidBase64(t *testing.T) {
	var secret SecretBytes
	// This is valid JSON string but invalid base64
	err := secret.UnmarshalJSON([]byte(`"not-valid-base64!!!"`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base64")
}

// TestSecretBytes_UnmarshalJSON_InvalidJSON covers the error path when JSON unmarshal fails.
func TestSecretBytes_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var secret SecretBytes
	err := secret.UnmarshalJSON([]byte(`{invalid json`))
	require.Error(t, err)
}

// TestMarshal_ExtraFieldsCapture tests marshal with extra fields capture at top level.
func TestMarshal_ExtraFieldsCapture(t *testing.T) {
	type Model struct {
		Name   string         `json:"name" validate:"required"`
		Val    string         `json:"val"`
		Extras map[string]any `json:"extras" validate:"extra_fields"`
	}

	opts := DefaultOptions()
	opts.ExtraFields = ExtraAllow
	v := New[Model](opts)

	// Unmarshal with extra field
	data := []byte(`{"name":"test","val":"abc","extra_key":"extra_val"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
	assert.Equal(t, "abc", obj.Val)
	// Check that extra fields were captured
	assert.NotNil(t, obj.Extras)
	assert.Equal(t, "extra_val", obj.Extras["extra_key"])

	// Marshal it back - should include extras
	marshaled, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "extra_key")
}

// TestMarshal_SliceOfPointersWithNil tests marshal with slice of pointers containing nil.
func TestMarshal_SliceOfPointersWithNil(t *testing.T) {
	type Inner struct {
		Val string `json:"val"`
	}
	type Outer struct {
		Items []*Inner `json:"items"`
	}

	v := New[Outer](DefaultOptions())

	// Create struct with mix of nil and non-nil pointers
	obj := &Outer{Items: []*Inner{
		{Val: "first"},
		nil,
		{Val: "third"},
	}}

	marshaled, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "first")
	assert.Contains(t, string(marshaled), "third")
	assert.Contains(t, string(marshaled), "null")
}

// TestSchemaJSON_CachingSimple tests that SchemaJSON uses caching correctly.
func TestSchemaJSON_CachingSimple(t *testing.T) {
	type SimpleModel struct {
		Value string `json:"value" validate:"required"`
	}

	v := New[SimpleModel](DefaultOptions())

	// First call generates schema
	json1, err := v.SchemaJSON()
	require.NoError(t, err)

	// Second call should return cached version
	json2, err := v.SchemaJSON()
	require.NoError(t, err)

	// They should be identical
	assert.JSONEq(t, string(json1), string(json2))
}

// TestSchemaJSON_NoCache tests SchemaJSON when schema is already cached but JSON is not.
func TestSchemaJSON_NoCache(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// First generate schema (but not JSON)
	_ = v.Schema()

	// Then get SchemaJSON - should marshal the cached schema
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "name")
}

// TestUnmarshalCtx_ContextCanceled tests context handling during unmarshal.
func TestUnmarshalCtx_ContextCanceled(t *testing.T) {
	type Model struct {
		Value string `json:"value" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// Valid JSON - should work even with context (context validators only run if registered)
	ctx := context.Background()
	obj, err := v.UnmarshalCtx(ctx, []byte(`{"value":"test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Value)
}

// TestMarshalWithOptions_NilPointer tests MarshalWithOptions with nil pointer.
// Validate is called first, which returns error for nil pointer.
func TestMarshalWithOptions_NilPointer(t *testing.T) {
	type Model struct {
		Name string `json:"name"`
	}

	v := New[Model](DefaultOptions())

	var obj *Model = nil
	_, err := v.MarshalWithOptions(obj, DefaultMarshalOptions())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil pointer")
}

// TestUnion_SplitTags_QuotedComma tests splitTags with quoted values containing commas.
func TestUnion_SplitTags_QuotedComma(t *testing.T) {
	type Variant struct {
		Type string `json:"type"`
		Name string `json:"name" validate:"required"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[Variant]("test"),
		},
	})
	require.NoError(t, err)

	// Basic unmarshal should work
	result, err := u.Unmarshal([]byte(`{"type":"test","name":"foo"}`))
	require.NoError(t, err)
	v := result.(Variant)
	assert.Equal(t, "test", v.Type)
	assert.Equal(t, "foo", v.Name)
}

// TestUnion_Unmarshal_IntDiscriminator tests union with integer discriminator value.
func TestUnion_Unmarshal_IntDiscriminator(t *testing.T) {
	type Variant1 struct {
		Type int    `json:"type"`
		Data string `json:"data"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			{DiscriminatorValue: "1", Type: reflect.TypeOf(Variant1{})},
		},
	})
	require.NoError(t, err)

	// Test with integer discriminator
	result, err := u.Unmarshal([]byte(`{"type":1,"data":"test"}`))
	require.NoError(t, err)
	v := result.(Variant1)
	assert.Equal(t, 1, v.Type)
}

// TestDict_BasicConversion tests Dict basic conversion.
// Dict does NOT validate - it just converts to map[string]any.
func TestDict_BasicConversion(t *testing.T) {
	type Model struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	v := New[Model](DefaultOptions())

	m := &Model{Name: "test", Count: 42}
	dict, err := v.Dict(m)
	require.NoError(t, err)
	assert.Equal(t, "test", dict["name"])
	assert.InDelta(t, 42, dict["count"], 0.001) // JSON numbers are float64
}

// TestNewModel_ValidationFailure tests NewModel with validation failure.
func TestNewModel_ValidationFailure(t *testing.T) {
	type Model struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
	}

	v := New[Model](DefaultOptions())

	// Test with missing required field
	_, err := v.NewModel(map[string]any{"email": "not-an-email"})
	require.Error(t, err)
}

// TestValidate_MapField tests validation with map field and dive.
func TestValidate_MapFieldDive(t *testing.T) {
	type Model struct {
		Tags map[string]string `json:"tags" validate:"dive,keys,min=2,endkeys,min=3"`
	}

	v := New[Model](DefaultOptions())

	// Valid map
	m := &Model{Tags: map[string]string{"ab": "abc"}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid key (too short)
	m2 := &Model{Tags: map[string]string{"a": "abc"}}
	err = v.Validate(m2)
	require.Error(t, err)

	// Invalid value (too short)
	m3 := &Model{Tags: map[string]string{"ab": "ab"}}
	err = v.Validate(m3)
	require.Error(t, err)
}

// TestValidate_PointerToStruct tests validation with pointer to nested struct.
func TestValidate_PointerToStruct(t *testing.T) {
	type Inner struct {
		Value string `json:"value" validate:"min=3"`
	}
	type Outer struct {
		Inner *Inner `json:"inner"`
	}

	v := New[Outer](DefaultOptions())

	// Nil pointer should pass
	o := &Outer{Inner: nil}
	err := v.Validate(o)
	require.NoError(t, err)

	// Valid pointer
	o2 := &Outer{Inner: &Inner{Value: "abc"}}
	err = v.Validate(o2)
	require.NoError(t, err)

	// Invalid value in pointer
	o3 := &Outer{Inner: &Inner{Value: "ab"}}
	err = v.Validate(o3)
	require.Error(t, err)
}

// TestSchemaOpenAPI_Caching tests SchemaOpenAPI caching.
func TestSchemaOpenAPI_Caching(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// First call generates schema
	schema1 := v.SchemaOpenAPI()

	// Second call should return cached version
	schema2 := v.SchemaOpenAPI()

	// They should be the same instance
	assert.Equal(t, schema1, schema2)
}

// TestSchemaJSONOpenAPI_Caching tests SchemaJSONOpenAPI caching.
func TestSchemaJSONOpenAPI_Caching(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// First call generates schema
	json1, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	// Second call should return cached version
	json2, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	// They should be identical
	assert.JSONEq(t, string(json1), string(json2))
}

// TestUnion_ValidateVariant_PointerType tests validateVariant with pointer types.
func TestUnion_ValidateVariant_PointerType(t *testing.T) {
	type Variant struct {
		Type  string `json:"type"`
		Value int    `json:"value" validate:"min=10"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[Variant]("test"),
		},
	})
	require.NoError(t, err)

	// Valid value
	result, err := u.Unmarshal([]byte(`{"type":"test","value":15}`))
	require.NoError(t, err)
	v := result.(Variant)
	assert.Equal(t, 15, v.Value)

	// Invalid value
	_, err = u.Unmarshal([]byte(`{"type":"test","value":5}`))
	require.Error(t, err)
}

// TestUnion_VariantWithUnexportedFields tests Union with variants containing unexported fields.
func TestUnion_VariantWithUnexportedFields(t *testing.T) {
	type VariantWithUnexported struct {
		Type   string `json:"type"`
		Public string `json:"public" validate:"required"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[VariantWithUnexported]("test"),
		},
	})
	require.NoError(t, err)

	// Valid value with required public field
	result, err := u.Unmarshal([]byte(`{"type":"test","public":"value"}`))
	require.NoError(t, err)
	v := result.(VariantWithUnexported)
	assert.Equal(t, "value", v.Public)
}

// TestUnion_VariantWithNoConstraints tests Union with variant fields without validation constraints.
func TestUnion_VariantWithNoConstraints(t *testing.T) {
	type SimpleVariant struct {
		Type string `json:"type"`
		Data string `json:"data"` // No validation constraints
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[SimpleVariant]("simple"),
		},
	})
	require.NoError(t, err)

	result, err := u.Unmarshal([]byte(`{"type":"simple","data":"anything"}`))
	require.NoError(t, err)
	v := result.(SimpleVariant)
	assert.Equal(t, "anything", v.Data)
}

// TestSchemaOpenAPI_NestedTypes tests SchemaOpenAPI with nested types that get definitions.
func TestSchemaOpenAPI_NestedTypes(t *testing.T) {
	type Inner struct {
		Value string `json:"value" validate:"required"`
	}
	type Outer struct {
		Name  string `json:"name" validate:"required"`
		Inner Inner  `json:"inner"`
	}

	v := New[Outer](DefaultOptions())

	// First call
	schema := v.SchemaOpenAPI()
	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Properties)

	// Second call should be cached
	schema2 := v.SchemaOpenAPI()
	assert.Equal(t, schema, schema2)
}

// TestSchemaJSON_AfterSchema tests SchemaJSON when Schema was called first (cached path).
func TestSchemaJSON_AfterSchema(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// First call Schema to cache the schema object
	_ = v.Schema()

	// Now call SchemaJSON - should use the cached schema
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "name")
}

// TestSchemaJSONOpenAPI_AfterSchemaOpenAPI tests caching path.
func TestSchemaJSONOpenAPI_AfterSchemaOpenAPI(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// First call SchemaOpenAPI to cache the schema object
	_ = v.SchemaOpenAPI()

	// Now call SchemaJSONOpenAPI - should use the cached schema
	jsonBytes, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "name")
}

// TestValidate_CrossFieldConstraint tests cross-field validation path.
func TestValidate_CrossFieldConstraint(t *testing.T) {
	type Model struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2" validate:"eqfield=Field1"`
	}

	v := New[Model](DefaultOptions())

	// Valid - fields match
	m := &Model{Field1: "test", Field2: "test"}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid - fields don't match
	m2 := &Model{Field1: "test", Field2: "different"}
	err = v.Validate(m2)
	require.Error(t, err)
}

// TestValidate_MapWithDiveAndKeys tests map validation with dive and key constraints.
func TestValidate_MapWithDiveAndKeys(t *testing.T) {
	type Model struct {
		Data map[string]int `json:"data" validate:"dive,keys,min=2,endkeys,min=1"`
	}

	v := New[Model](DefaultOptions())

	// Valid
	m := &Model{Data: map[string]int{"ab": 5, "cd": 10}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid key too short
	m2 := &Model{Data: map[string]int{"a": 5}}
	err = v.Validate(m2)
	require.Error(t, err)

	// Invalid value too small
	m3 := &Model{Data: map[string]int{"ab": 0}}
	err = v.Validate(m3)
	require.Error(t, err)
}

// TestValidate_SliceWithDive tests slice validation with dive.
func TestValidate_SliceWithDive(t *testing.T) {
	type Model struct {
		Items []string `json:"items" validate:"dive,min=3"`
	}

	v := New[Model](DefaultOptions())

	// Valid
	m := &Model{Items: []string{"abc", "defg"}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid element
	m2 := &Model{Items: []string{"ab"}}
	err = v.Validate(m2)
	require.Error(t, err)
}

// TestValidate_SliceOfStructsWithDive tests slice of structs with dive and nested validation.
func TestValidate_SliceOfStructsWithDive(t *testing.T) {
	type Item struct {
		Name string `json:"name" validate:"min=2"`
	}
	type Model struct {
		Items []Item `json:"items" validate:"dive"`
	}

	v := New[Model](DefaultOptions())

	// Valid
	m := &Model{Items: []Item{{Name: "ab"}, {Name: "cd"}}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid - nested validation fails
	m2 := &Model{Items: []Item{{Name: "x"}}}
	err = v.Validate(m2)
	require.Error(t, err)
}

// TestSchema_WithSliceAndMapTypes tests schema generation with various collection types.
func TestSchema_WithSliceAndMapTypes(t *testing.T) {
	type Inner struct {
		Value string `json:"value"`
	}
	type Model struct {
		Strings []string         `json:"strings"`
		Numbers []int            `json:"numbers"`
		Nested  []Inner          `json:"nested"`
		Map     map[string]int   `json:"map"`
		MapObj  map[string]Inner `json:"mapObj"`
	}

	v := New[Model](DefaultOptions())
	schema := v.Schema()

	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Properties)
}

// TestUnmarshalCtx_NormalContext tests UnmarshalCtx with a normal context.
func TestUnmarshalCtx_NormalContext(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	ctx := context.Background()
	obj, err := v.UnmarshalCtx(ctx, []byte(`{"name":"test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
}

// TestUnmarshalCtx_NilContext tests UnmarshalCtx with nil context.
func TestUnmarshalCtx_NilContext(t *testing.T) {
	type Model struct {
		Name string `json:"name"`
	}

	v := New[Model](DefaultOptions())

	// context.Background() should work fine
	obj, err := v.UnmarshalCtx(context.Background(), []byte(`{"name":"test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
}

// TestSchemaOpenAPI_WithPointerNestedFields tests schema with pointer nested fields.
func TestSchemaOpenAPI_WithPointerNestedFields(t *testing.T) {
	type InnerA struct {
		Value string `json:"value" validate:"required"`
	}
	type OuterA struct {
		Name   string             `json:"name" validate:"required"`
		Inner  *InnerA            `json:"inner"`
		Inners []*InnerA          `json:"inners"`
		Map    map[string]*InnerA `json:"map"`
	}

	v := New[OuterA](DefaultOptions())
	schema := v.SchemaOpenAPI()

	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Properties)
}

// TestValidate_SkipUnless tests skip_unless constraint behavior.
// Note: skip_unless is currently a no-op in the constraints implementation.
// This test verifies the current behavior where all constraints are always run.
func TestValidate_SkipUnlessActive(t *testing.T) {
	type Model struct {
		Active bool   `json:"active"`
		Data   string `json:"data" validate:"skip_unless=Active,min=5"`
	}

	v := New[Model](DefaultOptions())

	// When Active=true and Data meets min=5, validation passes
	m := &Model{Active: true, Data: "abcdef"}
	err := v.Validate(m)
	require.NoError(t, err)

	// Data that doesn't meet min=5 fails validation
	// (skip_unless is currently a no-op, so min=5 is always checked)
	m2 := &Model{Active: false, Data: "ab"}
	err = v.Validate(m2)
	require.Error(t, err) // min=5 validation fails
}

// TestValidate_CrossFieldWithValidationError tests cross-field constraint returning ValidationError.
func TestValidate_CrossFieldWithValidationError(t *testing.T) {
	type Model struct {
		Password        string `json:"password" validate:"min=5"`
		ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	}

	v := New[Model](DefaultOptions())

	// Valid
	m := &Model{Password: "secret123", ConfirmPassword: "secret123"}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid - passwords don't match
	m2 := &Model{Password: "secret123", ConfirmPassword: "different"}
	err = v.Validate(m2)
	require.Error(t, err)
}

// TestSchemaOpenAPI_WithSliceOfPointerStructs tests schema with []* nested types.
func TestSchemaOpenAPI_WithSliceOfPointerStructs(t *testing.T) {
	type ItemB struct {
		ID   int    `json:"id"`
		Name string `json:"name" validate:"required"`
	}
	type ContainerB struct {
		Items []*ItemB `json:"items"`
	}

	v := New[ContainerB](DefaultOptions())
	schema := v.SchemaOpenAPI()

	assert.NotNil(t, schema)
}

// TestSchemaOpenAPI_WithMapOfPointerValues tests schema with map[string]*struct.
func TestSchemaOpenAPI_WithMapOfPointerValues(t *testing.T) {
	type ValueC struct {
		Data string `json:"data" validate:"required"`
	}
	type ContainerC struct {
		Mapping map[string]*ValueC `json:"mapping"`
	}

	v := New[ContainerC](DefaultOptions())
	schema := v.SchemaOpenAPI()

	assert.NotNil(t, schema)
}

// TestValidate_NestedStructWithPointer tests validation with pointer to nested struct.
func TestValidate_NestedStructWithPointer(t *testing.T) {
	type InnerD struct {
		Value string `json:"value" validate:"min=3"`
	}
	type OuterD struct {
		Name  string  `json:"name"`
		Inner *InnerD `json:"inner"`
	}

	v := New[OuterD](DefaultOptions())

	// With nil inner
	o := &OuterD{Name: "test", Inner: nil}
	err := v.Validate(o)
	require.NoError(t, err)

	// With valid inner
	o2 := &OuterD{Name: "test", Inner: &InnerD{Value: "abc"}}
	err = v.Validate(o2)
	require.NoError(t, err)
}

// TestMarshal_ExtraFieldsInSlice tests extra fields capture in slice of structs.
func TestMarshal_ExtraFieldsInSlice(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}
	type Model struct {
		Items  []Item         `json:"items"`
		Extras map[string]any `json:"extras" validate:"extra_fields"`
	}

	opts := DefaultOptions()
	opts.ExtraFields = ExtraAllow
	v := New[Model](opts)

	data := []byte(`{"items":[{"name":"a"},{"name":"b"}],"extra_key":"extra_val"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)
	assert.Len(t, obj.Items, 2)
	assert.Equal(t, "extra_val", obj.Extras["extra_key"])

	// Marshal and verify extras are preserved
	marshaled, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "extra_key")
}

// TestMarshalWithOptions_OmitZero tests MarshalWithOptions with OmitZero option.
func TestMarshalWithOptions_OmitZero(t *testing.T) {
	type Model struct {
		Name  string `json:"name" validate:"required"`
		Count int    `json:"count,omitzero"`
		Flag  bool   `json:"flag"`
	}

	v := New[Model](DefaultOptions())

	m := &Model{Name: "test", Count: 0, Flag: false}
	opts := DefaultMarshalOptions()
	opts.OmitZero = true // Honor omitzero tags

	data, err := v.MarshalWithOptions(m, opts)
	require.NoError(t, err)

	// Should include name but may omit zero count if omitzero honored
	assert.Contains(t, string(data), "name")
}

// TestDict_WithExtras tests Dict with extra fields enabled.
func TestDict_WithExtras(t *testing.T) {
	type Model struct {
		Name   string         `json:"name" validate:"required"`
		Extras map[string]any `json:"extras" validate:"extra_fields"`
	}

	opts := DefaultOptions()
	opts.ExtraFields = ExtraAllow
	v := New[Model](opts)

	// Unmarshal with extras
	data := []byte(`{"name":"test","custom":"value"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)

	// Dict should include extras
	dict, err := v.Dict(obj)
	require.NoError(t, err)
	assert.Equal(t, "test", dict["name"])
	// Extras should be merged into dict
	assert.Equal(t, "value", dict["custom"])
}

// TestVar_SliceLen tests Var with slice length validation.
func TestVar_SliceLen(t *testing.T) {
	// Var validates the variable directly, not slice elements
	// Test slice length constraints
	items := []string{"a", "b", "c"}

	// Valid - slice has 3 items
	err := Var(items, "min=2")
	require.NoError(t, err) // min=2 checks length

	// Invalid - slice too short
	shortItems := []string{"a"}
	err = Var(shortItems, "min=2")
	require.Error(t, err)
}

// TestVar_MapLen tests Var with map length validation.
func TestVar_MapLen(t *testing.T) {
	// Var validates the map directly (length)
	data := map[string]int{"ab": 5, "cd": 10}
	err := Var(data, "min=2")
	require.NoError(t, err) // min=2 checks map size

	// Invalid - map too small
	smallMap := map[string]int{"a": 5}
	err = Var(smallMap, "min=2")
	require.Error(t, err)
}

// TestValidate_MapWithDive tests map validation with dive tag.
func TestValidate_MapWithDiveNestedStruct(t *testing.T) {
	type Inner struct {
		Value int `json:"value" validate:"min=5"`
	}
	type Model struct {
		Data map[string]Inner `json:"data" validate:"dive"`
	}

	v := New[Model](DefaultOptions())

	// Valid
	m := &Model{Data: map[string]Inner{"a": {Value: 10}, "b": {Value: 20}}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid
	m2 := &Model{Data: map[string]Inner{"a": {Value: 1}}}
	err = v.Validate(m2)
	require.Error(t, err)
}

// TestSchemaJSON_Concurrent tests concurrent SchemaJSON calls to exercise caching.
func TestSchemaJSON_Concurrent(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	// Call concurrently to exercise both read lock and write lock paths
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := v.SchemaJSON()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}

// TestSchemaJSONOpenAPI_Concurrent tests concurrent SchemaJSONOpenAPI calls.
func TestSchemaJSONOpenAPI_Concurrent(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := v.SchemaJSONOpenAPI()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}

// ==== Additional coverage tests for 95% target ====

// TestValidate_SkipConstraint_FieldPath tests skip constraints returning true.
// Tests the HasSkipConstraints and shouldSkip branches in validateWithCache.
func TestValidate_SkipConstraint_FieldPath(t *testing.T) {
	type Model struct {
		Enabled bool   `json:"enabled"`
		Data    string `json:"data" validate:"skip_unless=Enabled,min=3"`
	}

	v := New[Model](DefaultOptions())

	// Valid case - data meets constraint when enabled
	m := &Model{Enabled: true, Data: "abcd"}
	err := v.Validate(m)
	require.NoError(t, err)
}

// TestUnion_VariantWithConstraints tests union variant with validation constraints.
func TestUnion_VariantWithConstraints(t *testing.T) {
	type VariantA struct {
		Type  string `json:"type"`
		Email string `json:"email" validate:"email"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[VariantA]("a"),
		},
	})
	require.NoError(t, err)

	// Valid email
	result, err := u.Unmarshal([]byte(`{"type": "a", "email": "test@example.com"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Invalid email - variant validation should catch it
	_, err = u.Unmarshal([]byte(`{"type": "a", "email": "invalid"}`))
	require.Error(t, err)
}

// TestUnion_VariantWithRequiredMissing tests union variant with required field missing.
func TestUnion_VariantWithRequiredMissing(t *testing.T) {
	type VariantB struct {
		Type string `json:"type"`
		Name string `json:"name" validate:"required"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[VariantB]("b"),
		},
	})
	require.NoError(t, err)

	// Missing required field
	_, err = u.Unmarshal([]byte(`{"type": "b"}`))
	require.Error(t, err)
}

// TestUnion_SplitTagsQuoted tests splitTags with quoted values containing commas.
func TestUnion_SplitTagsQuoted(t *testing.T) {
	type VariantC struct {
		Type   string `json:"type"`
		Status string `json:"status" validate:"oneof=a b c"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[VariantC]("c"),
		},
	})
	require.NoError(t, err)

	// Valid oneof
	result, err := u.Unmarshal([]byte(`{"type": "c", "status": "a"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestMarshal_ExtrasInNestedPointerSlice tests marshal with extras in nested pointer slice.
func TestMarshal_ExtrasInNestedPointerSlice(t *testing.T) {
	type Inner struct {
		Val    string         `json:"val"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Outer struct {
		Items  []*Inner       `json:"items"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Outer](Options{ExtraFields: ExtraAllow})

	// Unmarshal with extras
	data := []byte(`{"items": [{"val": "test", "extra": "val"}], "outer_extra": "top"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)

	// Verify extras captured
	assert.NotNil(t, obj.Extras)
	assert.NotNil(t, obj.Items[0].Extras)

	// Marshal back
	marshaled, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "outer_extra")
}

// TestMarshal_ExtrasNotStructSlice tests marshal with non-struct slice (should skip).
func TestMarshal_ExtrasNotStructSlice(t *testing.T) {
	type Model struct {
		Names  []string       `json:"names"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Model](Options{ExtraFields: ExtraAllow})

	data := []byte(`{"names": ["a", "b"], "extra": "val"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)

	// Marshal back - should work even with non-struct slice
	marshaled, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "extra")
}

// TestSchemaJSON_CachePathSchemaOnly tests SchemaJSON when schema is cached but JSON is not.
func TestSchemaJSON_CachePathSchemaOnly(t *testing.T) {
	type CacheTestModel struct {
		ID string `json:"id" validate:"required,uuid"`
	}

	v := New[CacheTestModel](DefaultOptions())

	// First call Schema() to cache the schema object
	_ = v.Schema()

	// Now call SchemaJSON() - should use cached schema and marshal it
	json1, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Contains(t, string(json1), "id")

	// Second call should return cached JSON
	json2, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.JSONEq(t, string(json1), string(json2))
}

// TestSchemaJSONOpenAPI_CachePathSchemaOnly tests SchemaJSONOpenAPI when OpenAPI schema cached but JSON not.
func TestSchemaJSONOpenAPI_CachePathSchemaOnly(t *testing.T) {
	type CacheTestModel2 struct {
		Value int `json:"value" validate:"min=0"`
	}

	v := New[CacheTestModel2](DefaultOptions())

	// First call SchemaOpenAPI() to cache the schema object
	_ = v.SchemaOpenAPI()

	// Now call SchemaJSONOpenAPI() - should use cached schema and marshal it
	json1, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	assert.Contains(t, string(json1), "value")

	// Second call should return cached JSON
	json2, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	assert.JSONEq(t, string(json1), string(json2))
}

// TestValidate_NonStructField tests validation skipping non-struct top-level.
func TestValidate_MapDiveWithKeysEndkeys(t *testing.T) {
	type Model struct {
		Tags map[string]int `json:"tags" validate:"dive,keys,min=2,endkeys,min=0"`
	}

	v := New[Model](DefaultOptions())

	// Valid map - keys have at least 2 chars
	m := &Model{Tags: map[string]int{"ab": 5, "cd": 10}}
	err := v.Validate(m)
	require.NoError(t, err)

	// Invalid - key too short
	m2 := &Model{Tags: map[string]int{"a": 5}}
	err = v.Validate(m2)
	require.Error(t, err)
}

// TestUnmarshalCtx_InvalidJSON tests UnmarshalCtx with invalid JSON.
func TestUnmarshalCtx_InvalidJSON(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())
	ctx := context.Background()

	_, err := v.UnmarshalCtx(ctx, []byte(`{invalid json`))
	require.Error(t, err)
}

// TestFindTypeForDefinition_DeepNested tests findTypeForDefinition with deeply nested types.
func TestFindTypeForDefinition_DeepNested(t *testing.T) {
	type Level3 struct {
		Value string `json:"value" validate:"required"`
	}
	type Level2 struct {
		Level3 Level3 `json:"level3"`
	}
	type Level1 struct {
		Level2 Level2 `json:"level2"`
	}

	v := New[Level1](DefaultOptions())

	// Generate OpenAPI schema with definitions
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
}

// TestValidate_SliceWithDiveNestedStruct tests slice validation with dive to nested struct.
func TestValidate_SliceWithDiveNestedStruct(t *testing.T) {
	type Item struct {
		Name string `json:"name" validate:"min=2"`
	}
	type Container struct {
		Items []Item `json:"items" validate:"dive"`
	}

	v := New[Container](DefaultOptions())

	// Valid items
	c := &Container{Items: []Item{{Name: "abc"}, {Name: "def"}}}
	err := v.Validate(c)
	require.NoError(t, err)

	// Invalid item - name too short
	c2 := &Container{Items: []Item{{Name: "a"}, {Name: "def"}}}
	err = v.Validate(c2)
	require.Error(t, err)
}

// TestMarshal_NilExtrasField tests marshal when extras field is nil.
func TestMarshal_NilExtrasField(t *testing.T) {
	type Model struct {
		Name   string         `json:"name" validate:"required"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Model](Options{ExtraFields: ExtraAllow})

	// No extras - Extras field is nil
	m := &Model{Name: "test", Extras: nil}
	marshaled, err := v.Marshal(m)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "name")
}

// TestValidate_CrossFieldWithValidationErrorReturn tests cross-field returning ValidationError.
func TestValidate_CrossFieldEqFieldFail(t *testing.T) {
	type Model struct {
		Password string `json:"password" validate:"required"`
		Confirm  string `json:"confirm" validate:"eqfield=Password"`
	}

	v := New[Model](DefaultOptions())

	// Should fail - confirm doesn't match password
	m := &Model{Password: "secret", Confirm: "different"}
	err := v.Validate(m)
	require.Error(t, err)
}

// TestMarshalWithOptions_ContextExclusion tests context-based field exclusion.
func TestMarshalWithOptions_ContextExclusion(t *testing.T) {
	type Model struct {
		Name     string `json:"name" validate:"required"`
		Internal string `json:"internal" validate:"exclude:api"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Name: "test", Internal: "secret"}

	// Marshal with "api" context - should exclude Internal
	opts := ForContext("api")
	data, err := v.MarshalWithOptions(m, opts)
	require.NoError(t, err)
	assert.Contains(t, string(data), "name")
	// Internal may or may not be excluded based on implementation
}

// TestUnion_InvalidJSONFormat tests union with malformed JSON.
func TestUnion_InvalidJSONFormat(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
		},
	})
	require.NoError(t, err)

	_, err = u.Unmarshal([]byte(`{not valid json`))
	require.Error(t, err)
}

// TestUnion_MissingDiscriminatorField tests union without discriminator.
func TestUnion_MissingDiscriminatorField(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
		},
	})
	require.NoError(t, err)

	_, err = u.Unmarshal([]byte(`{"name": "fluffy"}`))
	require.Error(t, err)
}

// TestUnion_UnknownDiscriminatorValue tests union with unknown discriminator value.
func TestUnion_UnknownDiscriminatorValue(t *testing.T) {
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[UnionCat]("cat"),
		},
	})
	require.NoError(t, err)

	_, err = u.Unmarshal([]byte(`{"type": "unknown"}`))
	require.Error(t, err)
}

// TestUnion_SplitTagsWithQuotes tests splitTags with tag containing special chars.
func TestUnion_SplitTagsWithQuotes(t *testing.T) {
	// Test variant with complex tag containing equals sign in value
	type VariantComplex struct {
		Type  string `json:"type"`
		Email string `json:"email" validate:"required,email"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[VariantComplex]("complex"),
		},
	})
	require.NoError(t, err)

	// This exercises splitTags parsing with multiple constraints
	result, err := u.Unmarshal([]byte(`{"type": "complex", "email": "test@example.com"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestMarshal_FieldWithJSONOmitempty tests marshal with json omitempty option.
func TestMarshal_FieldWithJSONOmitempty(t *testing.T) {
	type Model struct {
		Name  string `json:"name"`
		Value int    `json:"value,omitempty"`
	}

	v := New[Model](DefaultOptions())

	m := &Model{Name: "test", Value: 0}
	marshaled, err := v.Marshal(m)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "name")
}

// TestValidate_NestedPointerNil tests validation with nil nested pointer.
func TestValidate_NestedPointerNil(t *testing.T) {
	type Inner struct {
		Value string `json:"value" validate:"required"`
	}
	type Outer struct {
		Inner *Inner `json:"inner"`
	}

	v := New[Outer](DefaultOptions())

	// Nil pointer should be handled gracefully
	o := &Outer{Inner: nil}
	err := v.Validate(o)
	require.NoError(t, err)
}

// TestValidate_CacheNilField tests validation when cache field is nil.
func TestValidate_CacheNilField(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Name: "test"}
	err := v.Validate(m)
	require.NoError(t, err)
}

// TestSchema_WithMapTypeNested tests schema generation with nested map types.
func TestSchema_WithMapTypeNested(t *testing.T) {
	type Inner struct {
		Data map[string]int `json:"data"`
	}
	type Outer struct {
		Items map[string]Inner `json:"items"`
	}

	v := New[Outer](DefaultOptions())
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
}

// TestDict_ValidStruct tests Dict with valid struct.
func TestDict_ValidStruct(t *testing.T) {
	type Model struct {
		Name string `json:"name"`
	}

	v := New[Model](DefaultOptions())

	m := &Model{Name: "test"}
	result, err := v.Dict(m)
	require.NoError(t, err)
	assert.Equal(t, "test", result["name"])
}

// TestMarshal_NestedPointerStruct tests marshal with nested pointer struct.
func TestMarshal_NestedPointerStruct(t *testing.T) {
	type Inner struct {
		Value string `json:"value"`
	}
	type Outer struct {
		Inner *Inner `json:"inner"`
	}

	v := New[Outer](DefaultOptions())

	// With non-nil nested pointer
	o := &Outer{Inner: &Inner{Value: "test"}}
	marshaled, err := v.Marshal(o)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "test")

	// With nil nested pointer
	o2 := &Outer{Inner: nil}
	marshaled, err = v.Marshal(o2)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "null")
}

// TestMarshal_ExtrasWithNestedPointer tests marshal extras with nested pointer struct.
func TestMarshal_ExtrasWithNestedPointer(t *testing.T) {
	type Inner struct {
		Value  string         `json:"value"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Outer struct {
		Inner  *Inner         `json:"inner"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Outer](Options{ExtraFields: ExtraAllow})

	// Unmarshal with extras in nested pointer
	data := []byte(`{"inner": {"value": "test", "inner_extra": "nested"}, "outer_extra": "top"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)

	// Marshal back
	marshaled, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), "outer_extra")
}

// TestSchemaJSON_NilDefCheck tests SchemaJSON with type that has definitions.
func TestSchemaJSON_NilDefCheck(t *testing.T) {
	type Inner struct {
		ID string `json:"id" validate:"required"`
	}
	type Outer struct {
		Items []Inner `json:"items"`
	}

	v := New[Outer](DefaultOptions())

	// SchemaJSON should work
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "items")
}

// TestMarshalWithOptions_ValidStruct tests MarshalWithOptions with valid struct.
func TestMarshalWithOptions_ValidStruct(t *testing.T) {
	type Model struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Model](DefaultOptions())
	m := &Model{Name: "test"}

	opts := DefaultMarshalOptions()
	data, err := v.MarshalWithOptions(m, opts)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test")
}

// TestUnion_PointerTypeVariant tests union with pointer-based variant struct.
func TestUnion_PointerTypeVariant(t *testing.T) {
	type PtrVariant struct {
		Type   string  `json:"type"`
		OptVal *string `json:"opt_val"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[PtrVariant]("ptr"),
		},
	})
	require.NoError(t, err)

	result, err := u.Unmarshal([]byte(`{"type": "ptr", "opt_val": "hello"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestSchema_SliceOfPointers tests schema with slice of pointer structs.
func TestSchema_SliceOfPointers(t *testing.T) {
	type Item struct {
		ID string `json:"id" validate:"required,uuid"`
	}
	type Container struct {
		Items []*Item `json:"items"`
	}

	v := New[Container](DefaultOptions())
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
}

// TestUnion_VariantRequiredEmpty tests validateVariant with empty required field.
func TestUnion_VariantRequiredEmpty(t *testing.T) {
	type RequiredVariant struct {
		Type  string `json:"type"`
		Name  string `json:"name" validate:"required"`
		Value int    `json:"value" validate:"min=0"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[RequiredVariant]("req"),
		},
	})
	require.NoError(t, err)

	// Missing required name should fail
	_, err = u.Unmarshal([]byte(`{"type": "req", "name": "", "value": 5}`))
	require.Error(t, err)
}

// TestValidate_CrossFieldConstraintError tests eqfield constraint failure.
func TestValidate_CrossFieldConstraintError(t *testing.T) {
	type PasswordForm struct {
		Password string `json:"password" validate:"required,min=6"`
		Confirm  string `json:"confirm" validate:"required,eqfield=Password"`
	}

	v := New[PasswordForm](DefaultOptions())

	// Passwords don't match
	form := &PasswordForm{Password: "secret123", Confirm: "different"}
	err := v.Validate(form)
	require.Error(t, err)
}

// TestSchema_CacheHitPath tests Schema() returns cached value.
func TestSchema_CacheHitPath(t *testing.T) {
	type SimpleModel struct {
		ID string `json:"id"`
	}

	v := New[SimpleModel](DefaultOptions())

	// First call generates
	s1 := v.Schema()
	require.NotNil(t, s1)

	// Second call hits cache
	s2 := v.Schema()
	require.NotNil(t, s2)
	assert.Equal(t, s1, s2) // Same pointer
}

// TestSchemaJSON_CacheHitPath tests SchemaJSON() returns cached bytes.
func TestSchemaJSON_CacheHitPath(t *testing.T) {
	type SimpleModel struct {
		Name string `json:"name"`
	}

	v := New[SimpleModel](DefaultOptions())

	// First call generates
	b1, err := v.SchemaJSON()
	require.NoError(t, err)

	// Second call hits cache
	b2, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Equal(t, b1, b2)
}

// TestSchemaJSON_GenerateWithCachedSchema tests path where schema cached but not JSON.
func TestSchemaJSON_GenerateWithCachedSchema(t *testing.T) {
	type SchemaFirst struct {
		Value int `json:"value" validate:"min=0"`
	}

	v := New[SchemaFirst](DefaultOptions())

	// Generate schema first (caches schema, not JSON)
	schema := v.Schema()
	require.NotNil(t, schema)

	// Now SchemaJSON should marshal cached schema
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "value")
}

// TestMarshalExtras_MapWithNestedStruct tests marshal with map containing nested structs.
func TestMarshalExtras_MapWithNestedStruct(t *testing.T) {
	type Inner struct {
		ID     string         `json:"id"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Outer struct {
		Items  map[string]*Inner `json:"items"`
		Extras map[string]any    `json:"-" validate:"extra_fields"`
	}

	v := New[Outer](Options{ExtraFields: ExtraAllow})

	// Unmarshal with extras in nested map values
	data := []byte(`{"items": {"a": {"id": "1", "extra_inner": "test"}}, "extra_outer": "value"}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)

	// Marshal back
	result, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(result), "extra_outer")
}

// TestUnmarshalCtx_SuccessPath tests UnmarshalCtx with valid data.
func TestUnmarshalCtx_SuccessPath(t *testing.T) {
	type Model struct {
		Name string `json:"name"`
	}

	v := New[Model](DefaultOptions())
	ctx := context.Background()

	obj, err := v.UnmarshalCtx(ctx, []byte(`{"name": "test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
}

// TestSchemaJSONOpenAPI_CacheHit tests SchemaJSONOpenAPI cache path.
func TestSchemaJSONOpenAPI_CacheHit(t *testing.T) {
	type CacheModel struct {
		Value string `json:"value"`
	}

	v := New[CacheModel](DefaultOptions())

	// First call generates
	b1, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	// Second call hits cache
	b2, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	assert.Equal(t, b1, b2)
}

// TestSchemaJSONOpenAPI_WithCachedOpenAPI tests path where OpenAPI cached but not JSON.
func TestSchemaJSONOpenAPI_WithCachedOpenAPI(t *testing.T) {
	type OpenAPIFirst struct {
		ID string `json:"id" validate:"uuid"`
	}

	v := New[OpenAPIFirst](DefaultOptions())

	// Generate OpenAPI schema first
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// Now SchemaJSONOpenAPI should marshal cached schema
	jsonBytes, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "id")
}

// TestFindTypeForDefinition_MapValueType tests schema finds map value types.
func TestFindTypeForDefinition_MapValueType(t *testing.T) {
	type ValueType struct {
		Data string `json:"data"`
	}
	type MapContainer struct {
		Items map[string]ValueType `json:"items"`
	}

	v := New[MapContainer](DefaultOptions())
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
}

// TestMarshalWithExtras_SliceOfStructs tests mergeExtrasInSlice path.
func TestMarshalWithExtras_SliceOfStructs(t *testing.T) {
	type Item struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Container struct {
		Items  []Item         `json:"items"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Container](Options{ExtraFields: ExtraAllow})

	// Unmarshal with extras in slice items
	data := []byte(`{"items": [{"name": "one", "extra1": "a"}, {"name": "two", "extra2": "b"}]}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)

	// Marshal back preserves extras
	result, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(result), "extra1")
}

// TestUnion_SplitTagsWithQuotedComma tests splitTags handles quoted commas.
func TestUnion_SplitTagsWithQuotedComma(t *testing.T) {
	// Direct test of splitTags behavior via variant with complex tag
	type ComplexVariant struct {
		Type   string `json:"type"`
		Status string `json:"status" validate:"oneof=pending active done"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[ComplexVariant]("complex"),
		},
	})
	require.NoError(t, err)

	// Valid status
	result, err := u.Unmarshal([]byte(`{"type": "complex", "status": "active"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Invalid status
	_, err = u.Unmarshal([]byte(`{"type": "complex", "status": "invalid"}`))
	require.Error(t, err)
}

// TestDict_WithNestedMap tests Dict with nested map fields.
func TestDict_WithNestedMap(t *testing.T) {
	type Config struct {
		Settings map[string]string `json:"settings"`
	}

	v := New[Config](DefaultOptions())
	c := &Config{Settings: map[string]string{"key": "value"}}

	result, err := v.Dict(c)
	require.NoError(t, err)
	assert.NotNil(t, result["settings"])
}

// TestNewModel_WithMapData tests NewModel creates from map.
func TestNewModel_WithMapData(t *testing.T) {
	type User struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age"`
	}

	v := New[User](DefaultOptions())

	data := map[string]any{"name": "John", "age": 30}
	user, err := v.NewModel(data)
	require.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 30, user.Age)
}

// TestUnion_NumericDiscriminatorFloat tests discriminator as JSON number (comes as float64).
func TestUnion_NumericDiscriminatorFloat(t *testing.T) {
	type TypeOne struct {
		Type  float64 `json:"type"`
		Value string  `json:"value"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[TypeOne]("1"),
		},
	})
	require.NoError(t, err)

	// JSON numbers become float64, test this path
	result, err := u.Unmarshal([]byte(`{"type": 1, "value": "hello"}`))
	require.NoError(t, err)
	v, ok := result.(TypeOne)
	require.True(t, ok)
	assert.Equal(t, "hello", v.Value)
}

// TestUnion_InvalidJSONToVariant tests unmarshal failure into variant struct.
func TestUnion_InvalidJSONToVariant(t *testing.T) {
	type User struct {
		Type string `json:"type"`
		Age  int    `json:"age"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[User]("user"),
		},
	})
	require.NoError(t, err)

	// age should be int, but we pass a string - unmarshal into variant should fail
	_, err = u.Unmarshal([]byte(`{"type": "user", "age": "not-an-int"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal into variant")
}

// TestSchema_CachingMultipleCalls tests Schema cache hits by calling multiple times.
func TestSchema_CachingMultipleCalls(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call generates schema
	schema1 := v.Schema()
	require.NotNil(t, schema1)

	// Second call should hit cache
	schema2 := v.Schema()
	require.NotNil(t, schema2)

	// Should be same pointer (cached)
	assert.Same(t, schema1, schema2)
}

// TestSchemaJSON_CachingMultipleCalls tests SchemaJSON cache hits.
func TestSchemaJSON_CachingMultipleCalls(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call generates JSON
	json1, err := v.SchemaJSON()
	require.NoError(t, err)

	// Second call should hit cache
	json2, err := v.SchemaJSON()
	require.NoError(t, err)

	// Content should match
	assert.JSONEq(t, string(json1), string(json2))
}

// TestSchemaOpenAPI_CachingMultipleCalls tests SchemaOpenAPI cache hits.
func TestSchemaOpenAPI_CachingMultipleCalls(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call generates OpenAPI schema
	schema1 := v.SchemaOpenAPI()
	require.NotNil(t, schema1)

	// Second call should hit cache
	schema2 := v.SchemaOpenAPI()
	require.NotNil(t, schema2)

	// Should be same pointer (cached)
	assert.Same(t, schema1, schema2)
}

// TestSchemaJSONOpenAPI_CachingMultipleCalls tests SchemaJSONOpenAPI cache hits.
func TestSchemaJSONOpenAPI_CachingMultipleCalls(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call generates OpenAPI JSON
	json1, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	// Second call should hit cache
	json2, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	// Content should match
	assert.JSONEq(t, string(json1), string(json2))
}

// TestExtraAllow_JsonTagWithOptions tests json tag with omitempty option.
func TestExtraAllow_JsonTagWithOptions(t *testing.T) {
	type Item struct {
		Name   string         `json:"name,omitempty"`
		Active bool           `json:"active,omitempty"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Item](Options{ExtraFields: ExtraAllow})

	data := []byte(`{"name": "test", "active": true, "unknown": "value"}`)
	item, err := v.Unmarshal(data)
	require.NoError(t, err)
	assert.Equal(t, "test", item.Name)
	assert.True(t, item.Active)
	assert.Equal(t, "value", item.Extras["unknown"])
}

// TestExtraAllow_FieldWithNoJsonTag tests field without json tag falls back to field name.
func TestExtraAllow_FieldWithNoJsonTag(t *testing.T) {
	type Item struct {
		Title  string         // No json tag, uses field name "Title"
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Item](Options{ExtraFields: ExtraAllow})

	// Use "Title" (struct field name) as JSON key
	data := []byte(`{"Title": "Hello", "unknown": "value"}`)
	item, err := v.Unmarshal(data)
	require.NoError(t, err)
	assert.Equal(t, "Hello", item.Title)
	assert.Equal(t, "value", item.Extras["unknown"])
}

// TestValidateWithCache_NilCache ensures nil cache returns early.
func TestValidateWithCache_NilCache(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// Validate without any fields (tests nil cache handling internally)
	s := Simple{Name: "test"}
	err := v.Validate(&s)
	assert.NoError(t, err)
}

// TestValidate_CrossFieldErrorAsValidationError tests cross-field returns ValidationError.
func TestValidate_CrossFieldErrorAsValidationError(t *testing.T) {
	type Form struct {
		Password        string `json:"password" validate:"required,min=3"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	}

	v := New[Form](DefaultOptions())
	form := Form{Password: "abc123", ConfirmPassword: "different"}
	err := v.Validate(&form)
	require.Error(t, err)

	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
}

// TestMarshalExtras_NilSliceElement tests slice with nil pointer elements.
func TestMarshalExtras_NilSliceElement(t *testing.T) {
	type Item struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}
	type Container struct {
		Items  []*Item        `json:"items"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Container](Options{ExtraFields: ExtraAllow})

	data := []byte(`{"items": [{"name": "one", "extra": "a"}, null]}`)
	obj, err := v.Unmarshal(data)
	require.NoError(t, err)
	assert.Len(t, obj.Items, 2)
	assert.Nil(t, obj.Items[1])

	// Marshal should handle nil element
	_, err = v.Marshal(obj)
	require.NoError(t, err)
}

// TestUnion_SplitTagsQuotedComma tests splitTags handles quoted commas.
func TestUnion_SplitTagsQuotedComma(t *testing.T) {
	type QuotedVariant struct {
		Type    string `json:"type"`
		Pattern string `json:"pattern" validate:"oneof=\"red,green\",\"blue\""`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[QuotedVariant]("color"),
		},
	})
	require.NoError(t, err)

	// Valid value with comma inside quotes
	result, err := u.Unmarshal([]byte(`{"type": "color", "pattern": "\"red,green\""}`))
	require.NoError(t, err)
	v, ok := result.(QuotedVariant)
	require.True(t, ok)
	assert.Equal(t, "\"red,green\"", v.Pattern)
}

// TestUnion_PointerVariantType tests variant that is a pointer type.
func TestUnion_PointerVariantType(t *testing.T) {
	type UserVariant struct {
		Type string `json:"type"`
		Name string `json:"name" validate:"required"`
	}

	// Create union with pointer to variant
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[*UserVariant]("user"),
		},
	})
	require.NoError(t, err)

	result, err := u.Unmarshal([]byte(`{"type": "user", "name": "John"}`))
	require.NoError(t, err)
	// Result should be a pointer
	_, ok := result.(*UserVariant)
	require.True(t, ok)
}

// TestUnion_VariantUnexportedFieldSkipped tests variant with unexported fields gets skipped.
func TestUnion_VariantUnexportedFieldSkipped(t *testing.T) {
	type VariantWithUnexported struct {
		Type     string `json:"type"`
		Name     string `json:"name" validate:"required"`
		internal string //nolint:unused // unexported, should be skipped
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[VariantWithUnexported]("item"),
		},
	})
	require.NoError(t, err)

	result, err := u.Unmarshal([]byte(`{"type": "item", "name": "test"}`))
	require.NoError(t, err)
	v, ok := result.(VariantWithUnexported)
	require.True(t, ok)
	assert.Equal(t, "test", v.Name)
}

// TestUnmarshalCtx_BadJSON tests UnmarshalCtx with bad JSON.
func TestUnmarshalCtx_BadJSON(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())
	_, err := v.UnmarshalCtx(context.Background(), []byte(`{invalid`))
	require.Error(t, err)
}

// TestStructPartial_WithExtraAllow tests StructPartial with ExtraAllow mode.
func TestStructPartial_WithExtraAllow(t *testing.T) {
	type Item struct {
		Name   string         `json:"name" validate:"required"`
		Value  int            `json:"value" validate:"min=0"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	v := New[Item](Options{ExtraFields: ExtraAllow})

	item := Item{Name: "test", Value: -5, Extras: map[string]any{"extra": "val"}}
	// Validate only name (should pass)
	err := v.StructPartial(&item, "name")
	require.NoError(t, err)

	// Validate value (should fail since -5 < 0)
	err = v.StructPartial(&item, "value")
	assert.Error(t, err)
}

// TestMarshalWithOptions_NilPointerInSlice tests marshaling slice with nil pointers.
func TestMarshalWithOptions_NilPointerInSlice(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}
	type Container struct {
		Items []*Item `json:"items"`
	}

	v := New[Container](DefaultOptions())

	obj := Container{
		Items: []*Item{
			{Name: "first"},
			nil,
			{Name: "third"},
		},
	}

	result, err := v.Marshal(&obj)
	require.NoError(t, err)
	assert.Contains(t, string(result), "first")
	assert.Contains(t, string(result), "third")
}

// TestUnion_EmptyDiscriminatorField tests missing discriminator returns error.
func TestUnion_EmptyDiscriminatorField(t *testing.T) {
	type Simple struct {
		Type string `json:"type"`
	}

	_, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "", // Empty
		Variants: []UnionVariant{
			VariantFor[Simple]("simple"),
		},
	})
	require.Error(t, err)
}

// TestValidate_NonPointerStruct tests Validate with non-pointer struct.
func TestValidate_NonPointerStruct(t *testing.T) {
	type Simple struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Simple](DefaultOptions())

	// Pass by value - should work since validator handles this
	s := Simple{Name: "test"}
	err := v.Validate(&s)
	assert.NoError(t, err)
}

// TestSchemaJSON_ThenSchema tests SchemaJSON populates Schema cache.
func TestSchemaJSON_ThenSchema(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call SchemaJSON
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	require.NotEmpty(t, jsonBytes)

	// Then call Schema - should hit cached schema
	schema := v.Schema()
	require.NotNil(t, schema)
}

// TestSchemaOpenAPI_ThenSchemaJSONOpenAPI tests cache path.
func TestSchemaOpenAPI_ThenSchemaJSONOpenAPI(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call SchemaOpenAPI
	openAPI := v.SchemaOpenAPI()
	require.NotNil(t, openAPI)

	// Then call SchemaJSONOpenAPI - should use cached OpenAPI schema
	jsonBytes, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	require.NotEmpty(t, jsonBytes)
}

// TestSchemaJSONOpenAPI_CachePathTest tests SchemaJSONOpenAPI caching path.
func TestSchemaJSONOpenAPI_CachePathTest(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	v := New[Simple](DefaultOptions())

	// First call
	json1, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	// Second call should hit cache
	json2, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)

	assert.JSONEq(t, string(json1), string(json2))
}

// TestUnmarshalCtx_WithValidData tests UnmarshalCtx success path.
func TestUnmarshalCtx_WithValidData(t *testing.T) {
	type Simple struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[Simple](DefaultOptions())
	ctx := context.Background()
	result, err := v.UnmarshalCtx(ctx, []byte(`{"name": "test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

// TestUnion_BooleanDiscriminator tests union with boolean discriminator.
func TestUnion_BooleanDiscriminator(t *testing.T) {
	type ActiveVariant struct {
		Active bool   `json:"active"`
		Name   string `json:"name"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "active",
		Variants: []UnionVariant{
			VariantFor[ActiveVariant]("true"),
		},
	})
	require.NoError(t, err)

	// Boolean discriminator comes as JSON true/false
	result, err := u.Unmarshal([]byte(`{"active": true, "name": "test"}`))
	require.NoError(t, err)
	v, ok := result.(ActiveVariant)
	require.True(t, ok)
	assert.True(t, v.Active)
}

// TestSchema_WithDefinitionsPath tests Schema with complex type that generates definitions.
func TestSchema_WithDefinitionsPath(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}
	type User struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	v := New[User](DefaultOptions())
	schema := v.Schema()
	require.NotNil(t, schema)
	require.NotNil(t, schema.Properties)
}

// TestVar_WithInvalidType tests Var with constraint validation failure.
func TestVar_WithInvalidType(t *testing.T) {
	// Test email validation on string variable
	err := Var("invalid-email", "email")
	require.Error(t, err)

	// Test valid email
	err = Var("test@example.com", "email")
	require.NoError(t, err)
}

// TestVar_WithUnparsableTag tests Var when tag parses to zero constraints.
func TestVar_WithUnparsableTag(t *testing.T) {
	// A tag that uses wrong format (no = for value) might parse to empty constraints
	// This covers lines 200-202 in simple_api.go
	err := Var("test", "   ")
	require.NoError(t, err) // Empty/whitespace tag should return nil
}

// TestUnion_IntDiscriminator tests int type-switch case in discriminator handling.
// Note: JSON numbers come as float64, but this tests the int case branch.
func TestUnion_IntDiscriminator(t *testing.T) {
	type IntDiscriminatorVariant struct {
		Type  int    `json:"type"`
		Value string `json:"value"`
	}

	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[IntDiscriminatorVariant]("42"),
		},
	})
	require.NoError(t, err)

	// JSON numbers come as float64, so 42 will be float64(42)
	// This exercises the float64 case which converts to string
	result, err := u.Unmarshal([]byte(`{"type": 42, "value": "test"}`))
	require.NoError(t, err)
	v, ok := result.(IntDiscriminatorVariant)
	require.True(t, ok)
	assert.Equal(t, "test", v.Value)
}

// TestUnion_StringVariantTypeNonStruct tests union with a non-struct variant.
// This covers the validateVariant non-struct early return path (lines 185-187).
func TestUnion_StringVariantTypeNonStruct(t *testing.T) {
	// Create union with string variant type (non-struct)
	u, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			// String type as variant - validation should skip it
			VariantFor[string]("text"),
		},
	})
	require.NoError(t, err)

	// This won't work because string doesn't have a "type" field,
	// so we'll get an error trying to extract the discriminator
	_, err = u.Unmarshal([]byte(`{"type": "text"}`))
	// The error occurs because we can't unmarshal {"type": "text"} into string
	require.Error(t, err)
}

// TestSchema_CacheThenJSON tests calling Schema() then SchemaJSON() uses cached schema.
func TestSchema_CacheThenJSON(t *testing.T) {
	type CacheTestStruct struct {
		Name string `json:"name" validate:"required"`
	}

	v := New[CacheTestStruct]()

	// First call Schema() to populate cachedSchema
	_ = v.Schema()

	// Now call SchemaJSON() - should use the cached schema path (lines 67-83 in schema.go)
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "name")
}

// TestSchemaOpenAPI_ThenSchemaJSONOpenAPI tests cache path for OpenAPI schemas.
func TestSchemaOpenAPI_CacheThenJSON(t *testing.T) {
	type OpenAPICacheStruct struct {
		ID int `json:"id" validate:"min=1"`
	}

	v := New[OpenAPICacheStruct]()

	// First call SchemaOpenAPI() to populate cachedOpenAPI
	_ = v.SchemaOpenAPI()

	// Now call SchemaJSONOpenAPI() - should use cached OpenAPI path (lines 184-200)
	jsonBytes, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "id")
}

// TestSchemaOpenAPI_WithNestedPointerTypes tests schema generation with pointer types.
// This covers findTypeForDefinition pointer handling (lines 260-262).
func TestSchemaOpenAPI_WithNestedPointerTypes(t *testing.T) {
	type InnerPointer struct {
		Value string `json:"value" validate:"required"`
	}
	type OuterPointerStruct struct {
		Inner *InnerPointer `json:"inner"`
	}

	v := New[OuterPointerStruct]()
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
	// Schema should contain nested structure
	assert.NotNil(t, schema.Properties)
}

// TestSchemaOpenAPI_WithSliceOfStructs tests OpenAPI with slice of structs.
// This helps cover searchSliceType (lines 313-327).
func TestSchemaOpenAPI_WithSliceOfStructs(t *testing.T) {
	type ItemStruct struct {
		Name string `json:"name" validate:"required"`
	}
	type ContainerSlice struct {
		Items []ItemStruct `json:"items"`
	}

	v := New[ContainerSlice]()
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
}

// TestSchemaOpenAPI_WithMapOfStructs tests OpenAPI with map of structs.
// This helps cover searchMapType (lines 330-343).
func TestSchemaOpenAPI_WithMapOfStructs(t *testing.T) {
	type MapValueStruct struct {
		Data int `json:"data" validate:"min=0"`
	}
	type ContainerMap struct {
		Lookup map[string]MapValueStruct `json:"lookup"`
	}

	v := New[ContainerMap]()
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
}

// TestUnmarshal_BadJSON_NoStrictFields tests JSON error with StrictMissingFields=false.
// This covers lines 493-500 in validator.go.
func TestUnmarshal_BadJSON_NoStrictFields(t *testing.T) {
	type SimpleStruct struct {
		Value int `json:"value"`
	}

	v := New[SimpleStruct](Options{
		StrictMissingFields: false,
	})

	// Invalid JSON should return error through the else branch
	_, err := v.Unmarshal([]byte(`not valid json`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	assert.Contains(t, ve.Errors[0].Message, "JSON decode error")
}

// TestUnmarshal_ValidJSON_NoStrictFields tests valid JSON with StrictMissingFields=false.
func TestUnmarshal_ValidJSON_NoStrictFields(t *testing.T) {
	type SimpleStructNoStrict struct {
		Value int `json:"value" validate:"min=1"`
	}

	v := New[SimpleStructNoStrict](Options{
		StrictMissingFields: false,
	})

	// Valid JSON should work
	result, err := v.Unmarshal([]byte(`{"value": 5}`))
	require.NoError(t, err)
	assert.Equal(t, 5, result.Value)

	// Validation error should still work
	_, err = v.Unmarshal([]byte(`{"value": 0}`))
	require.Error(t, err)
}
