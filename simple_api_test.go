package pedantigo

import (
	"encoding/json"
	"fmt"
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
		Name  string `json:"name" pedantigo:"required"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=0"`
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
		Email string `json:"email" pedantigo:"required"`
		Age   int    `json:"age" pedantigo:"min=18"`
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
		Host string `pedantigo:"required"`
		Port int    `pedantigo:"min=1,max=65535"`
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
		Port int `pedantigo:"min=1,max=65535"`
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
		Name  string `json:"name" pedantigo:"required"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=0"`
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
		ID    string  `json:"id" pedantigo:"required"`
		Name  string  `json:"name" pedantigo:"required"`
		Price float64 `json:"price" pedantigo:"min=0"`
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
		Title   string `json:"title" pedantigo:"required"`
		Content string `json:"content" pedantigo:"required"`
		Author  string `json:"author" pedantigo:"required"`
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
		ISBN   string `json:"isbn" pedantigo:"required"`
		Title  string `json:"title" pedantigo:"required"`
		Author string `json:"author" pedantigo:"required"`
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
	// Note: The library uses pedantigo:"exclude:context" format
	type Account struct {
		Username string `json:"username" pedantigo:"required"`
		Email    string `json:"email" pedantigo:"email"`
		Password string `json:"password" pedantigo:"exclude:response"`
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
		Street  string `json:"street" pedantigo:"required"`
		City    string `json:"city" pedantigo:"required"`
		ZipCode string `json:"zip_code" pedantigo:"required"`
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
		Name  string `json:"name" pedantigo:"required"`
		Email string `json:"email" pedantigo:"email"`
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
		Host string `pedantigo:"required"`
		Port int    `pedantigo:"min=1,max=65535"`
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
		ID    string  `json:"id" pedantigo:"required"`
		Name  string  `json:"name" pedantigo:"required"`
		Price float64 `json:"price" pedantigo:"min=0"`
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
		OrderID  string  `json:"order_id" pedantigo:"required"`
		Total    float64 `json:"total" pedantigo:"min=0"`
		Customer string  `json:"customer" pedantigo:"required"`
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
		FieldA string `json:"field_a" pedantigo:"required"`
	}
	type TypeB struct {
		FieldB int `json:"field_b" pedantigo:"min=0"`
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
		Name string `json:"name" pedantigo:"required"`
		URL  string `json:"url" pedantigo:"url"`
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)
	validatorChan := make(chan *Validator[Service], 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// This will trigger getOrCreateValidator
			validator := getOrCreateValidator[Service]()
			if validator == nil {
				errChan <- fmt.Errorf("validator is nil")
			} else {
				validatorChan <- validator
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
	for validator := range validatorChan {
		if firstValidator == nil {
			firstValidator = validator
		} else {
			assert.Same(t, firstValidator, validator, "All validators should be the same cached instance")
		}
	}
}
