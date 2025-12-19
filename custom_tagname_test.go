package pedantigo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Integration Tests for Custom Tag Name Support
// ============================================================================

// TestCustomTagName_PlaygroundCompatibility tests using "validate" tags
// like go-playground/validator for seamless migration.
func TestCustomTagName_PlaygroundCompatibility(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Set global tag name to "validate" (playground compatible)
	SetTagName("validate")

	type User struct {
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=18,max=120"`
	}

	v := New[User]()

	// Test validation with valid data
	valid, err := v.Unmarshal([]byte(`{"email": "test@example.com", "age": 25}`))
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", valid.Email)
	assert.Equal(t, 25, valid.Age)

	// Test validation with invalid email
	_, err = v.Unmarshal([]byte(`{"email": "invalid", "age": 25}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")

	// Test validation with age below min
	_, err = v.Unmarshal([]byte(`{"email": "test@example.com", "age": 10}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Age") // Field name in error

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_InstanceOverridesGlobal tests that per-instance TagName
// overrides the global setting.
func TestCustomTagName_InstanceOverridesGlobal(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Set global to "validate"
	SetTagName("validate")

	type Data struct {
		// Has both validate and custom tags with different constraints
		Value string `json:"value" validate:"required" custom:"min=5"`
	}

	// Create validator with custom tag name override
	v := New[Data](ValidatorOptions{TagName: "custom"})

	// With TagName: "custom", only min=5 constraint applies
	// Value "ab" should fail min=5 check
	_, err := v.Unmarshal([]byte(`{"value":"ab"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Value") // Field name in error

	// Value "abcdef" (6 chars) should pass min=5
	valid, err := v.Unmarshal([]byte(`{"value":"abcdef"}`))
	require.NoError(t, err)
	assert.Equal(t, "abcdef", valid.Value)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_DefaultIsPedantigo verifies default tag is "pedantigo".
func TestCustomTagName_DefaultIsPedantigo(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	type User struct {
		Name string `json:"name" pedantigo:"required,min=2"`
	}

	v := New[User]()

	// Valid name
	valid, err := v.Unmarshal([]byte(`{"name": "John"}`))
	require.NoError(t, err)
	assert.Equal(t, "John", valid.Name)

	// Name too short
	_, err = v.Unmarshal([]byte(`{"name": "J"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name") // Field name in error

	// Missing required field
	_, err = v.Unmarshal([]byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_BindingTag tests using "binding" tag like Gin framework.
func TestCustomTagName_BindingTag(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("binding")

	type LoginRequest struct {
		Username string `json:"username" binding:"required,min=3"`
		Password string `json:"password" binding:"required,min=8"`
	}

	v := New[LoginRequest]()

	// Valid request
	valid, err := v.Unmarshal([]byte(`{"username": "john", "password": "secretpw123"}`))
	require.NoError(t, err)
	assert.Equal(t, "john", valid.Username)

	// Password too short
	_, err = v.Unmarshal([]byte(`{"username": "john", "password": "short"}`))
	require.Error(t, err)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_DiveWithCustomTag verifies dive works with custom tags.
func TestCustomTagName_DiveWithCustomTag(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type EmailList struct {
		Emails []string `json:"emails" validate:"min=1,dive,email"`
	}

	v := New[EmailList]()

	// Valid emails
	valid, err := v.Unmarshal([]byte(`{"emails": ["a@b.com", "c@d.com"]}`))
	require.NoError(t, err)
	assert.Len(t, valid.Emails, 2)

	// Invalid email in list
	_, err = v.Unmarshal([]byte(`{"emails": ["a@b.com", "not-an-email"]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_NestedStructs verifies custom tags work with nested structs.
func TestCustomTagName_NestedStructs(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type Address struct {
		City    string `json:"city" validate:"required"`
		ZipCode string `json:"zip_code" validate:"len=5"`
	}

	type Person struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	v := New[Person]()

	// Valid nested struct
	valid, err := v.Unmarshal([]byte(`{"name": "John", "address": {"city": "NYC", "zip_code": "10001"}}`))
	require.NoError(t, err)
	assert.Equal(t, "John", valid.Name)
	assert.Equal(t, "NYC", valid.Address.City)

	// Invalid zip code length
	_, err = v.Unmarshal([]byte(`{"name": "John", "address": {"city": "NYC", "zip_code": "123"}}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ZipCode") // Field name in error

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_Validate verifies Validate() method works with custom tags.
func TestCustomTagName_Validate(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type User struct {
		Email string `json:"email" validate:"required,email"`
	}

	v := New[User]()

	// Valid struct
	user := &User{Email: "test@example.com"}
	err := v.Validate(user)
	require.NoError(t, err)

	// Invalid email
	invalidUser := &User{Email: "not-an-email"}
	err = v.Validate(invalidUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_Schema verifies Schema() generation works with custom tags.
func TestCustomTagName_Schema(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type Product struct {
		Name  string `json:"name" validate:"required,min=1,max=100"`
		Price int    `json:"price" validate:"min=0"`
	}

	v := New[Product]()
	schema := v.Schema()

	// Schema should include constraints from the "validate" tag
	require.NotNil(t, schema)
	require.NotNil(t, schema.Properties)

	nameSchema, ok := schema.Properties.Get("name")
	require.True(t, ok)
	require.NotNil(t, nameSchema)
	assert.NotNil(t, nameSchema.MinLength) // From min=1
	assert.NotNil(t, nameSchema.MaxLength) // From max=100

	priceSchema, ok := schema.Properties.Get("price")
	require.True(t, ok)
	require.NotNil(t, priceSchema)
	assert.NotNil(t, priceSchema.Minimum) // From min=0

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_UnionValidator tests union validator with custom tags.
func TestCustomTagName_UnionValidator(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type Cat struct {
		Name  string `json:"name" validate:"required"`
		Lives int    `json:"lives" validate:"min=1,max=9"`
	}
	type Dog struct {
		Name  string `json:"name" validate:"required"`
		Breed string `json:"breed"`
	}

	v, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	// Valid cat
	cat, err := v.Unmarshal([]byte(`{"type":"cat","name":"Whiskers","lives":9}`))
	require.NoError(t, err)
	assert.Equal(t, "Whiskers", cat.(Cat).Name)

	// Invalid: lives > 9
	_, err = v.Unmarshal([]byte(`{"type":"cat","name":"Whiskers","lives":15}`))
	require.Error(t, err)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_UnionSchema tests union schema generation with custom tags.
func TestCustomTagName_UnionSchema(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type Cat struct {
		Name  string `json:"name" validate:"required,min=2"`
		Lives int    `json:"lives" validate:"min=1,max=9"`
	}

	v, err := NewUnion[any](UnionOptions{
		DiscriminatorField: "type",
		Variants:           []UnionVariant{VariantFor[Cat]("cat")},
	})
	require.NoError(t, err)

	schema := v.Schema()
	require.NotNil(t, schema)
	require.NotEmpty(t, schema.OneOf)

	catSchema := schema.OneOf[0]
	livesSchema, ok := catSchema.Properties.Get("lives")
	require.True(t, ok)
	assert.Equal(t, json.Number("1"), livesSchema.Minimum)
	assert.Equal(t, json.Number("9"), livesSchema.Maximum)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_SchemaOpenAPI tests OpenAPI schema with custom tags.
func TestCustomTagName_SchemaOpenAPI(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type Product struct {
		Name  string `json:"name" validate:"required,min=1,max=100"`
		Price int    `json:"price" validate:"min=0"`
	}

	v := New[Product]()
	schema := v.SchemaOpenAPI()

	require.NotNil(t, schema)
	nameSchema, ok := schema.Properties.Get("name")
	require.True(t, ok)
	assert.NotNil(t, nameSchema.MinLength)
	assert.NotNil(t, nameSchema.MaxLength)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_SchemaJSON tests JSON schema bytes with custom tags.
func TestCustomTagName_SchemaJSON(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type User struct {
		Email string `json:"email" validate:"required,email"`
	}

	v := New[User]()
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)

	// Verify JSON contains expected constraints
	assert.Contains(t, string(jsonBytes), `"format": "email"`)
	assert.Contains(t, string(jsonBytes), `"required"`)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_MarshalWithOptions tests serialization with custom tags.
func TestCustomTagName_MarshalWithOptions(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate")

	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Password string `json:"password" validate:"exclude:api"`
	}

	v := New[User]()
	user := &User{ID: 1, Name: "John", Password: "secret"}

	// With "api" context, password should be excluded
	data, err := v.MarshalWithOptions(user, ForContext("api"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "password")
	assert.NotContains(t, string(data), "secret")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestCustomTagName_InstanceOverride_Schema tests per-instance TagName in schema.
func TestCustomTagName_InstanceOverride_Schema(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("validate") // Global is "validate"

	type Data struct {
		// Has different constraints in different tags
		Value string `json:"value" validate:"min=10" custom:"min=5"`
	}

	// Use instance override to "custom"
	v := New[Data](ValidatorOptions{TagName: "custom"})
	schema := v.Schema()

	valueSchema, ok := schema.Properties.Get("value")
	require.True(t, ok)
	// Should be min=5 from "custom" tag, not min=10 from "validate" tag
	assert.Equal(t, uint64(5), *valueSchema.MinLength)

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}
