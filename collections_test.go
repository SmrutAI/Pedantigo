package pedantigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Slice Validation Tests ====================

// ==================================================
// slice element validation tests
// ==================================================

// TestSlice_ValidEmails tests Slice validemails
func TestSlice_ValidEmails(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","bob@example.com"]}`)

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	assert.Len(t, config.Admins, 2)
}

func TestSlice_InvalidEmail_SingleElement(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["not-an-email"]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[0]" && fieldErr.Message == "must be a valid email address" {
			foundError = true
		}
	}

	assert.True(t, foundError, "expected error at 'Admins[0]', got %v", ve.Errors)
}

func TestSlice_InvalidEmail_MultipleElements(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","invalid","bob@example.com","also-invalid"]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check first error at index 1
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[1]" && fieldErr.Message == "must be a valid email address" {
			foundError1 = true
		}
	}
	assert.True(t, foundError1, "expected error at 'Admins[1]', got %v", ve.Errors)

	// Check second error at index 3
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[3]" && fieldErr.Message == "must be a valid email address" {
			foundError2 = true
		}
	}
	assert.True(t, foundError2, "expected error at 'Admins[3]', got %v", ve.Errors)
}

func TestSlice_MinLength(t *testing.T) {
	type User struct {
		Tags []string `json:"tags" pedantigo:"min=3"`
	}

	validator := New[User]()
	jsonData := []byte(`{"tags":["abc","de","fgh"]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 1)

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Tags[1]" && fieldErr.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	assert.True(t, foundError, "expected error at 'Tags[1]', got %v", ve.Errors)
}

func TestSlice_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type User struct {
		Addresses []Address `json:"addresses"`
	}

	validator := New[User]()
	jsonData := []byte(`{"addresses":[{"city":"NYC","zip":"10001"},{"zip":"123"}]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check for missing city at index 1
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Addresses[1].City" && fieldErr.Message == "is required" {
			foundError1 = true
		}
	}
	assert.True(t, foundError1, "expected error at 'Addresses[1].City', got %v", ve.Errors)

	// Check for short zip at index 1
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Addresses[1].Zip" && fieldErr.Message == "must be at least 5 characters" {
			foundError2 = true
		}
	}
	assert.True(t, foundError2, "expected error at 'Addresses[1].Zip', got %v", ve.Errors)
}

func TestSlice_EmptySlice(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":[]}`)

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	assert.Len(t, config.Admins, 0)
}

func TestSlice_NilSlice(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":null}`)

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	assert.Nil(t, config.Admins)
}

// ==================== Map Validation Tests ====================

// ==================================================
// map value validation tests
// ==================================================

// TestMap_ValidEmails tests Map validemails
func TestMap_ValidEmails(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"bob@example.com"}}`)

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	assert.Len(t, config.Contacts, 2)
}

func TestMap_InvalidEmail_SingleValue(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"not-an-email"}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Contacts[admin]" && fieldErr.Message == "must be a valid email address" {
			foundError = true
		}
	}

	assert.True(t, foundError, "expected error at 'Contacts[admin]', got %v", ve.Errors)
}

func TestMap_InvalidEmail_MultipleValues(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"invalid","billing":"bob@example.com","sales":"also-invalid"}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check that we have errors for the invalid keys (exact keys may vary due to map iteration order)
	invalidKeys := map[string]bool{"support": false, "sales": false}
	for _, fieldErr := range ve.Errors {
		if fieldErr.Message == "must be a valid email address" {
			switch fieldErr.Field {
			case "Contacts[support]":
				invalidKeys["support"] = true
			case "Contacts[sales]":
				invalidKeys["sales"] = true
			}
		}
	}

	assert.True(t, invalidKeys["support"], "expected error at 'Contacts[support]', got %v", ve.Errors)
	assert.True(t, invalidKeys["sales"], "expected error at 'Contacts[sales]', got %v", ve.Errors)
}

func TestMap_MinLength(t *testing.T) {
	type Config struct {
		Tags map[string]string `json:"tags" pedantigo:"min=3"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"tags":{"category":"abc","type":"de","status":"fgh"}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 1)

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Tags[type]" && fieldErr.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	assert.True(t, foundError, "expected error at 'Tags[type]', got %v", ve.Errors)
}

func TestMap_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type Company struct {
		Offices map[string]Address `json:"offices"`
	}

	validator := New[Company]()
	jsonData := []byte(`{"offices":{"hq":{"city":"NYC","zip":"10001"},"branch":{"zip":"123"}}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check for missing city at branch office
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Offices[branch].City" && fieldErr.Message == "is required" {
			foundError1 = true
		}
	}
	assert.True(t, foundError1, "expected error at 'Offices[branch].City', got %v", ve.Errors)

	// Check for short zip at branch office
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Offices[branch].Zip" && fieldErr.Message == "must be at least 5 characters" {
			foundError2 = true
		}
	}
	assert.True(t, foundError2, "expected error at 'Offices[branch].Zip', got %v", ve.Errors)
}

func TestMap_EmptyMap(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{}}`)

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	assert.Len(t, config.Contacts, 0)
}

func TestMap_NilMap(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":null}`)

	config, err := validator.Unmarshal(jsonData)
	assert.NoError(t, err)
	assert.Nil(t, config.Contacts)
}
