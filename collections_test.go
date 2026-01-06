package pedantigo

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test message constants.
const (
	errMsgValidEmail    = "must be a valid email address"
	errMsgAtLeast3Chars = "must be at least 3 characters"
	errMsgAtLeast5Chars = "must be at least 5 characters"
	errMsgRequired      = "is required"
)

// ==================== Slice Validation Tests ====================

// ==================================================
// slice element validation tests
// ==================================================

// TestSlice_ValidEmails tests slice element email validation with dive.
func TestSlice_ValidEmails(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"dive,email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","bob@example.com"]}`)

	config, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Len(t, config.Admins, 2)
}

func TestSlice_InvalidEmail_SingleElement(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"dive,email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["not-an-email"]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[0]" && fieldErr.Message == errMsgValidEmail {
			foundError = true
		}
	}

	assert.True(t, foundError, "expected error at 'Admins[0]', got %v", ve.Errors)
}

func TestSlice_InvalidEmail_MultipleElements(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"dive,email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","invalid","bob@example.com","also-invalid"]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check first error at index 1
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[1]" && fieldErr.Message == errMsgValidEmail {
			foundError1 = true
		}
	}
	assert.True(t, foundError1, "expected error at 'Admins[1]', got %v", ve.Errors)

	// Check second error at index 3
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[3]" && fieldErr.Message == errMsgValidEmail {
			foundError2 = true
		}
	}
	assert.True(t, foundError2, "expected error at 'Admins[3]', got %v", ve.Errors)
}

func TestSlice_ElementMinLength(t *testing.T) {
	// WITH dive: min=3 applies to each element's length (must be >= 3 chars)
	t.Run("with_dive_element_length", func(t *testing.T) {
		type User struct {
			Tags []string `json:"tags" pedantigo:"dive,min=3"`
		}

		validator := New[User]()
		jsonData := []byte(`{"tags":["abc","de","fgh"]}`) // "de" is only 2 chars

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		require.Len(t, ve.Errors, 1)
		assert.Equal(t, "Tags[1]", ve.Errors[0].Field)
		assert.Equal(t, errMsgAtLeast3Chars, ve.Errors[0].Message)
	})

	// WITHOUT dive: min=3 applies to slice length (must have >= 3 elements)
	t.Run("without_dive_collection_length", func(t *testing.T) {
		type User struct {
			Tags []string `json:"tags" pedantigo:"min=3"`
		}

		validator := New[User]()

		// Same data: 3 elements, but "de" is only 2 chars - should PASS because
		// without dive, we're checking element COUNT (3), not element LENGTH
		config, err := validator.Unmarshal([]byte(`{"tags":["abc","de","fgh"]}`))
		require.NoError(t, err)
		assert.Len(t, config.Tags, 3)

		// Only 2 elements - should FAIL (need 3 elements)
		_, err = validator.Unmarshal([]byte(`{"tags":["abc","de"]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Tags", ve.Errors[0].Field) // Error on collection, not element
		assert.Contains(t, ve.Errors[0].Message, "at least 3 elements")
	})
}

func TestSlice_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type User struct {
		Addresses []Address `json:"addresses" pedantigo:"dive"` // dive required to recurse into elements (like playground)
	}

	validator := New[User]()
	jsonData := []byte(`{"addresses":[{"city":"NYC","zip":"10001"},{"zip":"123"}]}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check for missing city at index 1
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Addresses[1].City" && fieldErr.Message == errMsgRequired {
			foundError1 = true
		}
	}
	assert.True(t, foundError1, "expected error at 'Addresses[1].City', got %v", ve.Errors)

	// Check for short zip at index 1
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Addresses[1].Zip" && fieldErr.Message == errMsgAtLeast5Chars {
			foundError2 = true
		}
	}
	assert.True(t, foundError2, "expected error at 'Addresses[1].Zip', got %v", ve.Errors)
}

func TestSlice_EmptySlice(t *testing.T) {
	// WITH dive: empty slice passes (no elements to validate)
	t.Run("with_dive_passes", func(t *testing.T) {
		type Config struct {
			Admins []string `json:"admins" pedantigo:"dive,email"`
		}

		validator := New[Config]()
		config, err := validator.Unmarshal([]byte(`{"admins":[]}`))
		require.NoError(t, err)
		assert.Empty(t, config.Admins)
	})

	// WITHOUT dive with min constraint: empty slice FAILS collection constraint
	t.Run("without_dive_fails_min", func(t *testing.T) {
		type Config struct {
			Admins []string `json:"admins" pedantigo:"min=1"`
		}

		validator := New[Config]()
		_, err := validator.Unmarshal([]byte(`{"admins":[]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Admins", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "at least 1")
	})
}

func TestSlice_NilSlice(t *testing.T) {
	// WITH dive: nil slice passes (no elements to validate)
	t.Run("with_dive_passes", func(t *testing.T) {
		type Config struct {
			Admins []string `json:"admins" pedantigo:"dive,email"`
		}

		validator := New[Config]()
		config, err := validator.Unmarshal([]byte(`{"admins":null}`))
		require.NoError(t, err)
		assert.Nil(t, config.Admins)
	})

	// WITHOUT dive with min constraint: nil slice has length 0, FAILS min constraint
	t.Run("without_dive_fails_min", func(t *testing.T) {
		type Config struct {
			Admins []string `json:"admins" pedantigo:"min=1"`
		}

		validator := New[Config]()
		_, err := validator.Unmarshal([]byte(`{"admins":null}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Admins", ve.Errors[0].Field)
	})
}

// ==================== Map Validation Tests ====================

// ==================================================
// map value validation tests
// ==================================================

// TestMap_ValidEmails tests map value email validation with dive.
func TestMap_ValidEmails(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"dive,email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"bob@example.com"}}`)

	config, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Len(t, config.Contacts, 2)
}

func TestMap_InvalidEmail_SingleValue(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"dive,email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"not-an-email"}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Contacts[admin]" && fieldErr.Message == errMsgValidEmail {
			foundError = true
		}
	}

	assert.True(t, foundError, "expected error at 'Contacts[admin]', got %v", ve.Errors)
}

func TestMap_InvalidEmail_MultipleValues(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"dive,email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"invalid","billing":"bob@example.com","sales":"also-invalid"}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check that we have errors for the invalid keys (exact keys may vary due to map iteration order)
	invalidKeys := map[string]bool{"support": false, "sales": false}
	for _, fieldErr := range ve.Errors {
		if fieldErr.Message == errMsgValidEmail {
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

func TestMap_ElementMinLength(t *testing.T) {
	// WITH dive: min=3 applies to each value's length (must be >= 3 chars)
	t.Run("with_dive_value_length", func(t *testing.T) {
		type Config struct {
			Tags map[string]string `json:"tags" pedantigo:"dive,min=3"`
		}

		validator := New[Config]()
		jsonData := []byte(`{"tags":{"category":"abc","type":"de","status":"fgh"}}`) // "de" is 2 chars

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		require.Len(t, ve.Errors, 1)
		assert.Equal(t, "Tags[type]", ve.Errors[0].Field)
		assert.Equal(t, errMsgAtLeast3Chars, ve.Errors[0].Message)
	})

	// WITHOUT dive: min=3 applies to map entry count (must have >= 3 entries)
	t.Run("without_dive_entry_count", func(t *testing.T) {
		type Config struct {
			Tags map[string]string `json:"tags" pedantigo:"min=3"`
		}

		validator := New[Config]()

		// Same data: 3 entries with short value "de" - should PASS because
		// without dive, we're checking entry COUNT (3), not value LENGTH
		config, err := validator.Unmarshal([]byte(`{"tags":{"category":"abc","type":"de","status":"fgh"}}`))
		require.NoError(t, err)
		assert.Len(t, config.Tags, 3)

		// Only 2 entries - should FAIL (need 3 entries)
		_, err = validator.Unmarshal([]byte(`{"tags":{"k1":"v1","k2":"v2"}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Tags", ve.Errors[0].Field) // Error on collection, not value
		assert.Contains(t, ve.Errors[0].Message, "at least 3 entries")
	})
}

func TestMap_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type Company struct {
		Offices map[string]Address `json:"offices" pedantigo:"dive"` // dive required to recurse into elements (like playground)
	}

	validator := New[Company]()
	jsonData := []byte(`{"offices":{"hq":{"city":"NYC","zip":"10001"},"branch":{"zip":"123"}}}`)

	_, err := validator.Unmarshal(jsonData)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	assert.Len(t, ve.Errors, 2)

	// Check for missing city at branch office
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Offices[branch].City" && fieldErr.Message == errMsgRequired {
			foundError1 = true
		}
	}
	assert.True(t, foundError1, "expected error at 'Offices[branch].City', got %v", ve.Errors)

	// Check for short zip at branch office
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Offices[branch].Zip" && fieldErr.Message == errMsgAtLeast5Chars {
			foundError2 = true
		}
	}
	assert.True(t, foundError2, "expected error at 'Offices[branch].Zip', got %v", ve.Errors)
}

func TestMap_EmptyMap(t *testing.T) {
	// WITH dive: empty map passes (no entries to validate)
	t.Run("with_dive_passes", func(t *testing.T) {
		type Config struct {
			Contacts map[string]string `json:"contacts" pedantigo:"dive,email"`
		}

		validator := New[Config]()
		config, err := validator.Unmarshal([]byte(`{"contacts":{}}`))
		require.NoError(t, err)
		assert.Empty(t, config.Contacts)
	})

	// WITHOUT dive with min constraint: empty map FAILS collection constraint
	t.Run("without_dive_fails_min", func(t *testing.T) {
		type Config struct {
			Contacts map[string]string `json:"contacts" pedantigo:"min=1"`
		}

		validator := New[Config]()
		_, err := validator.Unmarshal([]byte(`{"contacts":{}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Contacts", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "at least 1")
	})
}

func TestMap_NilMap(t *testing.T) {
	// WITH dive: nil map passes (no entries to validate)
	t.Run("with_dive_passes", func(t *testing.T) {
		type Config struct {
			Contacts map[string]string `json:"contacts" pedantigo:"dive,email"`
		}

		validator := New[Config]()
		config, err := validator.Unmarshal([]byte(`{"contacts":null}`))
		require.NoError(t, err)
		assert.Nil(t, config.Contacts)
	})

	// WITHOUT dive with min constraint: nil map has length 0, FAILS min constraint
	t.Run("without_dive_fails_min", func(t *testing.T) {
		type Config struct {
			Contacts map[string]string `json:"contacts" pedantigo:"min=1"`
		}

		validator := New[Config]()
		_, err := validator.Unmarshal([]byte(`{"contacts":null}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Contacts", ve.Errors[0].Field)
	})
}

// ==================== Collection-Level Constraint Tests (NO dive) ====================

// TestSlice_CollectionMinElements tests that without dive, min applies to slice length.
func TestSlice_CollectionMinElements(t *testing.T) {
	type Config struct {
		Tags []string `json:"tags" pedantigo:"min=3"` // NO dive = collection constraint
	}
	validator := New[Config]()

	// Should FAIL: only 2 elements, need 3
	_, err := validator.Unmarshal([]byte(`{"tags":["a","b"]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	// Error should be on "Tags" field, not "Tags[0]" or "Tags[1]"
	assert.Equal(t, "Tags", ve.Errors[0].Field)
	assert.Contains(t, ve.Errors[0].Message, "at least 3")

	// Should PASS: exactly 3 elements
	config, err := validator.Unmarshal([]byte(`{"tags":["a","b","c"]}`))
	require.NoError(t, err)
	assert.Len(t, config.Tags, 3)
}

// TestSlice_CollectionMaxElements tests that without dive, max applies to slice length.
func TestSlice_CollectionMaxElements(t *testing.T) {
	type Config struct {
		Tags []string `json:"tags" pedantigo:"max=2"` // NO dive = collection constraint
	}
	validator := New[Config]()

	// Should FAIL: 3 elements, max is 2
	_, err := validator.Unmarshal([]byte(`{"tags":["a","b","c"]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "Tags", ve.Errors[0].Field)
	assert.Contains(t, ve.Errors[0].Message, "at most 2")

	// Should PASS: exactly 2 elements
	config, err := validator.Unmarshal([]byte(`{"tags":["a","b"]}`))
	require.NoError(t, err)
	assert.Len(t, config.Tags, 2)
}

// TestSlice_MixedConstraints tests both collection and element constraints.
func TestSlice_MixedConstraints(t *testing.T) {
	type Config struct {
		// Collection: min 2 elements; Elements: each min 5 chars
		Tags []string `json:"tags" pedantigo:"min=2,dive,min=5"`
	}
	validator := New[Config]()

	// Should FAIL: only 1 element (collection constraint violated)
	_, err := validator.Unmarshal([]byte(`{"tags":["hello"]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "Tags", ve.Errors[0].Field)

	// Should FAIL: element too short (element constraint violated)
	_, err = validator.Unmarshal([]byte(`{"tags":["hello","hi"]}`)) // "hi" is < 5 chars
	require.Error(t, err)
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "Tags[1]", ve.Errors[0].Field)

	// Should PASS: 2 elements, each >= 5 chars
	config, err := validator.Unmarshal([]byte(`{"tags":["hello","world"]}`))
	require.NoError(t, err)
	assert.Len(t, config.Tags, 2)
}

// TestMap_CollectionMinEntries tests that without dive, min applies to entry count.
func TestMap_CollectionMinEntries(t *testing.T) {
	type Config struct {
		Tags map[string]string `json:"tags" pedantigo:"min=2"` // NO dive = entry count
	}
	validator := New[Config]()

	// Should FAIL: only 1 entry, need 2
	_, err := validator.Unmarshal([]byte(`{"tags":{"key1":"value1"}}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "Tags", ve.Errors[0].Field)
	assert.Contains(t, ve.Errors[0].Message, "at least 2")

	// Should PASS: 2 entries
	config, err := validator.Unmarshal([]byte(`{"tags":{"k1":"v1","k2":"v2"}}`))
	require.NoError(t, err)
	assert.Len(t, config.Tags, 2)
}

// ==================== Map Key Validation Tests ====================

// TestMap_KeyValidation tests keys/endkeys for validating map keys.
func TestMap_KeyValidation(t *testing.T) {
	type Config struct {
		// Keys: min 3 chars; Values: must be emails
		Contacts map[string]string `json:"contacts" pedantigo:"dive,keys,min=3,endkeys,email"`
	}
	validator := New[Config]()

	// Should FAIL: key "ab" is < 3 chars
	_, err := validator.Unmarshal([]byte(`{"contacts":{"ab":"test@example.com"}}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, ve.Errors[0].Field, "[ab]")
	assert.Contains(t, ve.Errors[0].Message, "at least 3")

	// Should FAIL: value is not valid email
	_, err = validator.Unmarshal([]byte(`{"contacts":{"admin":"not-an-email"}}`))
	require.Error(t, err)
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, ve.Errors[0].Field, "[admin]")
	assert.Contains(t, ve.Errors[0].Message, "email")

	// Should PASS: key >= 3 chars, value is valid email
	config, err := validator.Unmarshal([]byte(`{"contacts":{"admin":"admin@example.com"}}`))
	require.NoError(t, err)
	assert.Len(t, config.Contacts, 1)
}

// TestMap_KeyOnlyValidation tests key validation without value constraints.
func TestMap_KeyOnlyValidation(t *testing.T) {
	type Config struct {
		// Only key constraints, no value constraints
		Data map[string]int `json:"data" pedantigo:"dive,keys,min=2,endkeys"`
	}
	validator := New[Config]()

	// Should FAIL: key "a" is < 2 chars
	_, err := validator.Unmarshal([]byte(`{"data":{"a":100}}`))
	require.Error(t, err)

	// Should PASS: key >= 2 chars
	config, err := validator.Unmarshal([]byte(`{"data":{"ab":100}}`))
	require.NoError(t, err)
	assert.Equal(t, 100, config.Data["ab"])
}

// ==================== Error Case Tests ====================

// TestDive_PanicOnNonCollection tests that dive on a non-collection field panics.
func TestDive_PanicOnNonCollection(t *testing.T) {
	type Config struct {
		Name string `json:"name" pedantigo:"dive,min=3"` // ERROR: dive on string
	}

	// Should panic at validator creation time
	assert.Panics(t, func() {
		New[Config]()
	})
}

// TestKeys_RequiresDive tests that keys without dive panics.
func TestKeys_RequiresDive(t *testing.T) {
	type Config struct {
		// ERROR: keys without dive
		Contacts map[string]string `json:"contacts" pedantigo:"keys,min=3,endkeys,email"`
	}

	// Should panic at validator creation time
	assert.Panics(t, func() {
		New[Config]()
	})
}

// TestEndKeys_RequiresKeys tests that endkeys without keys panics.
func TestEndKeys_RequiresKeys(t *testing.T) {
	type Config struct {
		// ERROR: endkeys without keys
		Contacts map[string]string `json:"contacts" pedantigo:"dive,endkeys,email"`
	}

	// Should panic at validator creation time
	assert.Panics(t, func() {
		New[Config]()
	})
}

// TestKeys_OnlyValidForMaps tests that keys on a slice panics.
func TestKeys_OnlyValidForMaps(t *testing.T) {
	type Config struct {
		// ERROR: keys on slice
		Tags []string `json:"tags" pedantigo:"dive,keys,min=3,endkeys,email"`
	}

	// Should panic at validator creation time
	assert.Panics(t, func() {
		New[Config]()
	})
}

// ==================== Unique Constraint Tests ====================

// TestSlice_Unique tests the unique constraint on slices.
func TestSlice_Unique(t *testing.T) {
	t.Run("unique_strings", func(t *testing.T) {
		type Config struct {
			Tags []string `json:"tags" pedantigo:"unique"`
		}
		validator := New[Config]()

		// Valid: unique elements
		config, err := validator.Unmarshal([]byte(`{"tags":["a","b","c"]}`))
		require.NoError(t, err)
		assert.Len(t, config.Tags, 3)

		// Invalid: duplicates
		_, err = validator.Unmarshal([]byte(`{"tags":["a","b","a"]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Tags", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "duplicate")
	})

	t.Run("unique_ints", func(t *testing.T) {
		type Config struct {
			IDs []int `json:"ids" pedantigo:"unique"`
		}
		validator := New[Config]()

		// Valid
		_, err := validator.Unmarshal([]byte(`{"ids":[1,2,3]}`))
		require.NoError(t, err)

		// Invalid
		_, err = validator.Unmarshal([]byte(`{"ids":[1,2,1]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		type Config struct {
			Tags []string `json:"tags" pedantigo:"unique"`
		}
		validator := New[Config]()

		config, err := validator.Unmarshal([]byte(`{"tags":[]}`))
		require.NoError(t, err)
		assert.Empty(t, config.Tags)
	})

	t.Run("nil_slice_passes", func(t *testing.T) {
		type Config struct {
			Tags []string `json:"tags" pedantigo:"unique"`
		}
		validator := New[Config]()

		config, err := validator.Unmarshal([]byte(`{"tags":null}`))
		require.NoError(t, err)
		assert.Nil(t, config.Tags)
	})

	t.Run("single_element_passes", func(t *testing.T) {
		type Config struct {
			Tags []string `json:"tags" pedantigo:"unique"`
		}
		validator := New[Config]()

		config, err := validator.Unmarshal([]byte(`{"tags":["only"]}`))
		require.NoError(t, err)
		assert.Len(t, config.Tags, 1)
	})
}

// TestMap_UniqueValues tests the unique constraint on map values.
func TestMap_UniqueValues(t *testing.T) {
	t.Run("unique_values", func(t *testing.T) {
		type Config struct {
			Scores map[string]int `json:"scores" pedantigo:"unique"`
		}
		validator := New[Config]()

		// Valid: unique values
		config, err := validator.Unmarshal([]byte(`{"scores":{"a":1,"b":2,"c":3}}`))
		require.NoError(t, err)
		assert.Len(t, config.Scores, 3)

		// Invalid: duplicate values
		_, err = validator.Unmarshal([]byte(`{"scores":{"a":1,"b":1}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Scores", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "duplicate")
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		type Config struct {
			Scores map[string]int `json:"scores" pedantigo:"unique"`
		}
		validator := New[Config]()

		config, err := validator.Unmarshal([]byte(`{"scores":{}}`))
		require.NoError(t, err)
		assert.Empty(t, config.Scores)
	})
}

// TestSlice_UniqueByField tests the unique=Field constraint on struct slices.
func TestSlice_UniqueByField(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	t.Run("unique_by_id", func(t *testing.T) {
		type Config struct {
			Users []User `json:"users" pedantigo:"unique=ID"`
		}
		validator := New[Config]()

		// Valid: unique IDs
		config, err := validator.Unmarshal([]byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`))
		require.NoError(t, err)
		assert.Len(t, config.Users, 2)

		// Invalid: duplicate IDs (different names OK)
		_, err = validator.Unmarshal([]byte(`{"users":[{"id":1,"name":"Alice"},{"id":1,"name":"Bob"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Users", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "ID")
	})

	t.Run("unique_by_name", func(t *testing.T) {
		type Config struct {
			Users []User `json:"users" pedantigo:"unique=Name"`
		}
		validator := New[Config]()

		// Valid: unique names
		_, err := validator.Unmarshal([]byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`))
		require.NoError(t, err)

		// Invalid: duplicate names
		_, err = validator.Unmarshal([]byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Alice"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_struct_slice_passes", func(t *testing.T) {
		type Config struct {
			Users []User `json:"users" pedantigo:"unique=ID"`
		}
		validator := New[Config]()

		config, err := validator.Unmarshal([]byte(`{"users":[]}`))
		require.NoError(t, err)
		assert.Empty(t, config.Users)
	})
}

// TestUnique_PanicOnNonCollection tests that unique on non-collection panics.
func TestUnique_PanicOnNonCollection(t *testing.T) {
	type Config struct {
		Name string `json:"name" pedantigo:"unique"`
	}

	assert.Panics(t, func() {
		New[Config]()
	})
}

// TestSlice_UniqueWithOtherConstraints tests unique combined with other constraints.
func TestSlice_UniqueWithOtherConstraints(t *testing.T) {
	t.Run("unique_and_min", func(t *testing.T) {
		type Config struct {
			Tags []string `json:"tags" pedantigo:"unique,min=2"`
		}
		validator := New[Config]()

		// Invalid: not enough elements
		_, err := validator.Unmarshal([]byte(`{"tags":["a"]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Contains(t, ve.Errors[0].Message, "at least 2")

		// Invalid: duplicates
		_, err = validator.Unmarshal([]byte(`{"tags":["a","a"]}`))
		require.Error(t, err)

		// Valid: 2 unique elements
		config, err := validator.Unmarshal([]byte(`{"tags":["a","b"]}`))
		require.NoError(t, err)
		assert.Len(t, config.Tags, 2)
	})
}

// ==================== Numeric Dive Tests ====================
// Comprehensive dive tests for numeric constraints in nested structures.

// TestDive_GT tests greater than constraint in nested structs via dive.
func TestDive_GT(t *testing.T) {
	type Item struct {
		Value int `json:"value" pedantigo:"gt=10"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":15}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":5}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("boundary_value_invalid", func(t *testing.T) {
		// gt=10 means > 10, so exactly 10 should fail
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":10}]}`))
		require.Error(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":15},{"value":5}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Value", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_GT_Map tests greater than constraint in map values via dive.
func TestDive_GT_Map(t *testing.T) {
	type Item struct {
		Score int `json:"score" pedantigo:"gt=50"`
	}
	type Container struct {
		Scores map[string]Item `json:"scores" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_map_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"scores":{"player1":{"score":75}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_map_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"scores":{"player1":{"score":25}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Contains(t, ve.Errors[0].Field, "Scores[player1].Score")
	})

	t.Run("mixed_map_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"scores":{"p1":{"score":75},"p2":{"score":25}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"scores":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_GTE tests greater than or equal constraint in nested structs via dive.
func TestDive_GTE(t *testing.T) {
	type Item struct {
		Age int `json:"age" pedantigo:"gte=18"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"age":18},{"age":25}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"age":17}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Age", ve.Errors[0].Field)
	})

	t.Run("boundary_value_valid", func(t *testing.T) {
		// gte=18 means >= 18, so exactly 18 should pass
		_, err := validator.Unmarshal([]byte(`{"items":[{"age":18}]}`))
		require.NoError(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"age":20},{"age":15}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Age", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_GTE_Map tests greater than or equal constraint in map values via dive.
func TestDive_GTE_Map(t *testing.T) {
	type Item struct {
		Rating float64 `json:"rating" pedantigo:"gte=0.0"`
	}
	type Container struct {
		Ratings map[string]Item `json:"ratings" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_map_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"ratings":{"book1":{"rating":4.5}}}`))
		require.NoError(t, err)
	})

	t.Run("boundary_zero_valid", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"ratings":{"book1":{"rating":0.0}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_negative", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"ratings":{"book1":{"rating":-1.5}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"ratings":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_LT tests less than constraint in nested structs via dive.
func TestDive_LT(t *testing.T) {
	type Item struct {
		Discount int `json:"discount" pedantigo:"lt=100"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"discount":50}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"discount":150}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Discount", ve.Errors[0].Field)
	})

	t.Run("boundary_value_invalid", func(t *testing.T) {
		// lt=100 means < 100, so exactly 100 should fail
		_, err := validator.Unmarshal([]byte(`{"items":[{"discount":100}]}`))
		require.Error(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"discount":50},{"discount":200}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Discount", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_LT_Map tests less than constraint in map values via dive.
func TestDive_LT_Map(t *testing.T) {
	type Item struct {
		Temperature float64 `json:"temp" pedantigo:"lt=100.0"`
	}
	type Container struct {
		Readings map[string]Item `json:"readings" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_map_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"readings":{"sensor1":{"temp":75.5}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_map_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"readings":{"sensor1":{"temp":150.0}}}`))
		require.Error(t, err)
	})

	t.Run("boundary_invalid", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"readings":{"sensor1":{"temp":100.0}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"readings":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_LTE tests less than or equal constraint in nested structs via dive.
func TestDive_LTE(t *testing.T) {
	type Item struct {
		Capacity int `json:"capacity" pedantigo:"lte=1000"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"capacity":500},{"capacity":1000}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"capacity":1500}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Capacity", ve.Errors[0].Field)
	})

	t.Run("boundary_value_valid", func(t *testing.T) {
		// lte=1000 means <= 1000, so exactly 1000 should pass
		_, err := validator.Unmarshal([]byte(`{"items":[{"capacity":1000}]}`))
		require.NoError(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"capacity":800},{"capacity":1200}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Capacity", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_LTE_Map tests less than or equal constraint in map values via dive.
func TestDive_LTE_Map(t *testing.T) {
	type Item struct {
		Percentage float64 `json:"pct" pedantigo:"lte=100.0"`
	}
	type Container struct {
		Stats map[string]Item `json:"stats" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_map_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"stats":{"cpu":{"pct":75.5}}}`))
		require.NoError(t, err)
	})

	t.Run("boundary_valid", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"stats":{"memory":{"pct":100.0}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_exceeds_max", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"stats":{"disk":{"pct":101.5}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"stats":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_Positive tests positive constraint in nested structs via dive.
func TestDive_Positive(t *testing.T) {
	type Item struct {
		Amount int `json:"amount" pedantigo:"positive"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_positive_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"amount":1},{"amount":100}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_zero", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"amount":0}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Amount", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "positive")
	})

	t.Run("invalid_negative", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"amount":-5}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Amount", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"amount":10},{"amount":-1}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Amount", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_Positive_Map tests positive constraint in map values via dive.
func TestDive_Positive_Map(t *testing.T) {
	type Item struct {
		Price float64 `json:"price" pedantigo:"positive"`
	}
	type Container struct {
		Products map[string]Item `json:"products" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_positive_prices", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{"item1":{"price":9.99}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_zero_price", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{"item1":{"price":0.0}}}`))
		require.Error(t, err)
	})

	t.Run("invalid_negative_price", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{"item1":{"price":-5.0}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_Negative tests negative constraint in nested structs via dive.
func TestDive_Negative(t *testing.T) {
	type Item struct {
		Delta int `json:"delta" pedantigo:"negative"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_negative_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"delta":-1},{"delta":-100}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_zero", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"delta":0}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Delta", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "negative")
	})

	t.Run("invalid_positive", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"delta":5}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Delta", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"delta":-10},{"delta":1}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Delta", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_Negative_Map tests negative constraint in map values via dive.
func TestDive_Negative_Map(t *testing.T) {
	type Item struct {
		Adjustment float64 `json:"adj" pedantigo:"negative"`
	}
	type Container struct {
		Adjustments map[string]Item `json:"adjustments" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_negative_adjustments", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"adjustments":{"item1":{"adj":-2.5}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_zero", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"adjustments":{"item1":{"adj":0.0}}}`))
		require.Error(t, err)
	})

	t.Run("invalid_positive", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"adjustments":{"item1":{"adj":5.0}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"adjustments":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_MultipleOf tests multiple_of constraint in nested structs via dive.
func TestDive_MultipleOf(t *testing.T) {
	type Item struct {
		Quantity int `json:"quantity" pedantigo:"multiple_of=5"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_multiples", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"quantity":5},{"quantity":10},{"quantity":15}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_not_multiple", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"quantity":7}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Quantity", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "multiple of")
	})

	t.Run("zero_is_multiple", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"quantity":0}]}`))
		require.NoError(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"quantity":10},{"quantity":13}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Quantity", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_MultipleOf_Map tests multiple_of constraint in map values via dive.
func TestDive_MultipleOf_Map(t *testing.T) {
	type Item struct {
		Price float64 `json:"price" pedantigo:"multiple_of=0.25"`
	}
	type Container struct {
		Products map[string]Item `json:"products" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_multiples", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{"item1":{"price":1.25},"item2":{"price":2.50}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_not_multiple", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{"item1":{"price":1.33}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"products":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_MaxDigits tests max_digits constraint in nested structs via dive.
func TestDive_MaxDigits(t *testing.T) {
	type Item struct {
		Code int `json:"code" pedantigo:"max_digits=4"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_within_limit", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":1234},{"code":999}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_exceeds_digits", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":12345}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Code", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "at most 4 digits")
	})

	t.Run("negative_counts_digits", func(t *testing.T) {
		// -12345 has 5 digits, should fail
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":-12345}]}`))
		require.Error(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":123},{"code":12345}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Code", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_MaxDigits_Map tests max_digits constraint in map values via dive.
func TestDive_MaxDigits_Map(t *testing.T) {
	type Item struct {
		PIN int `json:"pin" pedantigo:"max_digits=6"`
	}
	type Container struct {
		Users map[string]Item `json:"users" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_within_limit", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"users":{"user1":{"pin":123456}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_exceeds_digits", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"users":{"user1":{"pin":1234567}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"users":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_DecimalPlaces tests decimal_places constraint in nested structs via dive.
func TestDive_DecimalPlaces(t *testing.T) {
	type Item struct {
		Price float64 `json:"price" pedantigo:"decimal_places=2"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_within_limit", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":9.99},{"price":10.5}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_exceeds_places", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":9.999}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Price", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "at most 2 decimal places")
	})

	t.Run("no_decimals_valid", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":10.0}]}`))
		require.NoError(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":10.50},{"price":10.555}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Price", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_DecimalPlaces_Map tests decimal_places constraint in map values via dive.
func TestDive_DecimalPlaces_Map(t *testing.T) {
	type Item struct {
		Rate float64 `json:"rate" pedantigo:"decimal_places=3"`
	}
	type Container struct {
		Rates map[string]Item `json:"rates" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_within_limit", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"rates":{"usd":{"rate":1.234}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_exceeds_places", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"rates":{"eur":{"rate":1.2345}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"rates":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_DisallowInfNan tests disallow_inf_nan constraint in nested structs via dive.
func TestDive_DisallowInfNan(t *testing.T) {
	type Item struct {
		Value float64 `json:"value" pedantigo:"disallow_inf_nan"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_normal_float", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":123.45}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_zero", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":0.0}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_negative", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":-999.99}]}`))
		require.NoError(t, err)
	})

	// Note: JSON doesn't natively support Inf/NaN, but we can test with Go's Validate method
	t.Run("invalid_infinity_via_validate", func(t *testing.T) {
		container := &Container{
			Items: []Item{{Value: math.Inf(1)}},
		}
		err := validator.Validate(container)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "infinity")
	})

	t.Run("invalid_nan_via_validate", func(t *testing.T) {
		container := &Container{
			Items: []Item{{Value: math.NaN()}},
		}
		err := validator.Validate(container)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "NaN")
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_DisallowInfNan_Map tests disallow_inf_nan constraint in map values via dive.
func TestDive_DisallowInfNan_Map(t *testing.T) {
	type Item struct {
		Measurement float64 `json:"measurement" pedantigo:"disallow_inf_nan"`
	}
	type Container struct {
		Readings map[string]Item `json:"readings" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_normal_floats", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"readings":{"sensor1":{"measurement":98.6}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_infinity_via_validate", func(t *testing.T) {
		container := &Container{
			Readings: map[string]Item{
				"sensor1": {Measurement: math.Inf(-1)},
			},
		}
		err := validator.Validate(container)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Contains(t, ve.Errors[0].Field, "Readings[sensor1].Measurement")
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"readings":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_Oneof tests oneof constraint in nested structs via dive.
func TestDive_Oneof(t *testing.T) {
	type Item struct {
		Status string `json:"status" pedantigo:"oneof=pending approved rejected"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"pending"},{"status":"approved"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Status", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "one of")
	})

	t.Run("case_sensitive", func(t *testing.T) {
		// "Pending" with capital P should fail (case-sensitive)
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"Pending"}]}`))
		require.Error(t, err)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"approved"},{"status":"unknown"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Status", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_Oneof_Map tests oneof constraint in map values via dive.
func TestDive_Oneof_Map(t *testing.T) {
	type Item struct {
		Priority string `json:"priority" pedantigo:"oneof=low medium high"`
	}
	type Container struct {
		Tasks map[string]Item `json:"tasks" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"tasks":{"task1":{"priority":"high"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"tasks":{"task1":{"priority":"critical"}}}`))
		require.Error(t, err)
	})

	t.Run("case_sensitive", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"tasks":{"task1":{"priority":"HIGH"}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"tasks":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_Oneof_Numeric tests oneof constraint with numeric values via dive.
func TestDive_Oneof_Numeric(t *testing.T) {
	type Item struct {
		Level int `json:"level" pedantigo:"oneof=1 2 3 5 8"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_numeric_values", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"level":1},{"level":5}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_numeric_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"level":4}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Level", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_OneofCI tests oneofci (case-insensitive) constraint in nested structs via dive.
func TestDive_OneofCI(t *testing.T) {
	type Item struct {
		Role string `json:"role" pedantigo:"oneofci=admin user guest"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"role":"admin"},{"role":"user"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"role":"ADMIN"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mixed_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"role":"Admin"},{"role":"USER"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"role":"superadmin"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Role", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "case-insensitive")
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"role":"GUEST"},{"role":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[1].Role", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_OneofCI_Map tests oneofci constraint in map values via dive.
func TestDive_OneofCI_Map(t *testing.T) {
	type Item struct {
		Permission string `json:"permission" pedantigo:"oneofci=read write execute"`
	}
	type Container struct {
		Files map[string]Item `json:"files" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"files":{"file1":{"permission":"read"}}}`))
		require.NoError(t, err)
	})

	t.Run("valid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"files":{"file1":{"permission":"WRITE"}}}`))
		require.NoError(t, err)
	})

	t.Run("valid_mixed_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"files":{"file1":{"permission":"Execute"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_value", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"files":{"file1":{"permission":"delete"}}}`))
		require.Error(t, err)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"files":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_CombinedNumericConstraints tests multiple numeric constraints on same field via dive.
func TestDive_CombinedNumericConstraints(t *testing.T) {
	type Item struct {
		Price float64 `json:"price" pedantigo:"positive,lte=9999.99,decimal_places=2"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_all_constraints", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":99.99}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_not_positive", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":-10.00}]}`))
		require.Error(t, err)
	})

	t.Run("invalid_exceeds_max", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":10000.00}]}`))
		require.Error(t, err)
	})

	t.Run("invalid_too_many_decimals", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"price":99.999}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_NumericConstraintsOnIntTypes tests numeric constraints work across different int types.
func TestDive_NumericConstraintsOnIntTypes(t *testing.T) {
	type Item struct {
		Int8Val  int8  `json:"int8" pedantigo:"gt=-100,lt=100"`
		Int16Val int16 `json:"int16" pedantigo:"gte=0"`
		Int32Val int32 `json:"int32" pedantigo:"positive"`
		Int64Val int64 `json:"int64" pedantigo:"multiple_of=10"`
		UintVal  uint  `json:"uint" pedantigo:"lte=1000"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_all_types", func(t *testing.T) {
		jsonData := `{"items":[{"int8":50,"int16":100,"int32":5,"int64":100,"uint":500}]}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.NoError(t, err)
	})

	t.Run("invalid_int8_boundary", func(t *testing.T) {
		jsonData := `{"items":[{"int8":100,"int16":0,"int32":1,"int64":10,"uint":100}]}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Int8Val", ve.Errors[0].Field)
	})

	t.Run("invalid_int32_not_positive", func(t *testing.T) {
		jsonData := `{"items":[{"int8":0,"int16":0,"int32":0,"int64":10,"uint":100}]}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Int32Val", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_NumericConstraintsOnFloatTypes tests numeric constraints work across different float types.
func TestDive_NumericConstraintsOnFloatTypes(t *testing.T) {
	type Item struct {
		Float32Val float32 `json:"float32" pedantigo:"gt=0.0,lt=100.0"`
		Float64Val float64 `json:"float64" pedantigo:"gte=0.0,lte=1.0,decimal_places=3"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_all_types", func(t *testing.T) {
		jsonData := `{"items":[{"float32":50.5,"float64":0.999}]}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.NoError(t, err)
	})

	t.Run("invalid_float32_out_of_range", func(t *testing.T) {
		jsonData := `{"items":[{"float32":100.0,"float64":0.5}]}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Float32Val", ve.Errors[0].Field)
	})

	t.Run("invalid_float64_too_many_decimals", func(t *testing.T) {
		jsonData := `{"items":[{"float32":50.0,"float64":0.9999}]}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Float64Val", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})
}

// ==================== Format Dive Tests ====================
//
// Tests for format constraints (url, uri, http_url, https_url, uuid, uuid3, uuid4, uuid5)
// when applied to fields inside nested structs accessed via the dive tag.

// TestDive_URL tests URL constraint in nested structs via dive.
// url accepts any scheme (http, https, ftp, file, mailto, etc.)
func TestDive_URL(t *testing.T) {
	type Link struct {
		URL string `json:"url" pedantigo:"url"`
	}
	type Container struct {
		Links []Link `json:"links" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":[{"url":"https://example.com/path"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":[{"url":"not-a-url"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Links[0].URL", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid URL")
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":[{"url":"https://valid.com"},{"url":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Links[1].URL", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":[]}`))
		require.NoError(t, err)
	})

	t.Run("various_schemes_pass", func(t *testing.T) {
		// url accepts http, https, ftp, file, mailto, etc.
		_, err := validator.Unmarshal([]byte(`{"links":[
			{"url":"http://localhost:8080"},
			{"url":"ftp://files.example.com"},
			{"url":"file:///path/to/file"}
		]}`))
		require.NoError(t, err)
	})

	t.Run("no_scheme_fails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":[{"url":"example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Links[0].URL", ve.Errors[0].Field)
	})
}

// TestDive_URLMap tests URL constraint in map values via dive.
func TestDive_URLMap(t *testing.T) {
	type Link struct {
		URL string `json:"url" pedantigo:"url"`
	}
	type Registry struct {
		Links map[string]Link `json:"links" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":{"homepage":{"url":"https://example.com"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":{"docs":{"url":"not-a-url"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":{}}`))
		require.NoError(t, err)
	})

	t.Run("multiple_map_entries_mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"links":{
			"homepage":{"url":"https://example.com"},
			"broken":{"url":"invalid-url"}
		}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})
}

// TestDive_URI tests URI constraint in nested structs via dive.
// uri accepts any scheme, similar to url but slightly different validation.
func TestDive_URI(t *testing.T) {
	type Resource struct {
		URI string `json:"uri" pedantigo:"uri"`
	}
	type Container struct {
		Resources []Resource `json:"resources" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":[{"uri":"mailto:user@example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":[{"uri":"not-a-uri"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Resources[0].URI", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid URI")
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":[{"uri":"file:///path/to/file"},{"uri":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Resources[1].URI", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":[]}`))
		require.NoError(t, err)
	})

	t.Run("various_uri_schemes_pass", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":[
			{"uri":"urn:isbn:0451450523"},
			{"uri":"mailto:test@example.com"},
			{"uri":"tel:+1-234-567-8900"}
		]}`))
		require.NoError(t, err)
	})
}

// TestDive_URIMap tests URI constraint in map values via dive.
func TestDive_URIMap(t *testing.T) {
	type Resource struct {
		URI string `json:"uri" pedantigo:"uri"`
	}
	type Registry struct {
		Resources map[string]Resource `json:"resources" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":{"book":{"uri":"urn:isbn:0451450523"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":{"contact":{"uri":"not-a-uri"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"resources":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_HTTPURL tests HTTP/HTTPS URL constraint in nested structs via dive.
// http_url only accepts http:// or https:// schemes.
func TestDive_HTTPURL(t *testing.T) {
	type Endpoint struct {
		URL string `json:"url" pedantigo:"http_url"`
	}
	type Container struct {
		Endpoints []Endpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_http_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"http://example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_https_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"https://secure.example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ftp_fails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"ftp://files.example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].URL", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid HTTP/HTTPS URL")
	})

	t.Run("invalid_no_scheme_fails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].URL", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"http://valid.com"},{"url":"ftp://invalid.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[1].URL", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[]}`))
		require.NoError(t, err)
	})

	t.Run("both_http_and_https_pass", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[
			{"url":"http://api.example.com"},
			{"url":"https://secure.api.example.com"}
		]}`))
		require.NoError(t, err)
	})
}

// TestDive_HTTPURLMap tests HTTP/HTTPS URL constraint in map values via dive.
func TestDive_HTTPURLMap(t *testing.T) {
	type Endpoint struct {
		URL string `json:"url" pedantigo:"http_url"`
	}
	type Registry struct {
		Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":{"api":{"url":"https://api.example.com"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_scheme_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":{"files":{"url":"ftp://files.example.com"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_HTTPSURL tests HTTPS-only URL constraint in nested structs via dive.
// https_url only accepts https:// scheme (rejects http://).
func TestDive_HTTPSURL(t *testing.T) {
	type SecureEndpoint struct {
		URL string `json:"url" pedantigo:"https_url"`
	}
	type Container struct {
		Endpoints []SecureEndpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_https_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"https://secure.example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_http_fails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"http://example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].URL", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid HTTPS URL")
	})

	t.Run("invalid_ftp_fails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"ftp://files.example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].URL", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"url":"https://valid.com"},{"url":"http://invalid.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[1].URL", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_HTTPSURLMap tests HTTPS-only URL constraint in map values via dive.
func TestDive_HTTPSURLMap(t *testing.T) {
	type SecureEndpoint struct {
		URL string `json:"url" pedantigo:"https_url"`
	}
	type Registry struct {
		Endpoints map[string]SecureEndpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":{"api":{"url":"https://api.example.com"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_http_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":{"api":{"url":"http://api.example.com"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID tests UUID constraint (any version) in nested structs via dive.
func TestDive_UUID(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}
	type Container struct {
		Entities []Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_uuid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"550e8400-e29b-41d4-a716-446655440000"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_uuid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"not-a-uuid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid UUID")
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[
			{"id":"550e8400-e29b-41d4-a716-446655440000"},
			{"id":"invalid-uuid"}
		]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[1].ID", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[]}`))
		require.NoError(t, err)
	})

	t.Run("various_uuid_versions_pass", func(t *testing.T) {
		// uuid accepts any version (v3, v4, v5, etc.)
		_, err := validator.Unmarshal([]byte(`{"entities":[
			{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"},
			{"id":"550e8400-e29b-41d4-a716-446655440000"},
			{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}
		]}`))
		require.NoError(t, err)
	})

	t.Run("missing_dashes_fails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"550e8400e29b41d4a716446655440000"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
	})
}

// TestDive_UUIDMap tests UUID constraint in map values via dive.
func TestDive_UUIDMap(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}
	type Registry struct {
		Entities map[string]Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"user1":{"id":"550e8400-e29b-41d4-a716-446655440000"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"user2":{"id":"not-a-uuid"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID3 tests UUID version 3 constraint in nested structs via dive.
// uuid3 requires version byte (position 14) to be '3'.
func TestDive_UUID3(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid3"`
	}
	type Container struct {
		Entities []Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_uuid3_in_slice", func(t *testing.T) {
		// UUID v3 with version byte '3' at position 14
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_uuid4_as_uuid3_fails", func(t *testing.T) {
		// UUID v4 (has '4' at position 14)
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"550e8400-e29b-41d4-a716-446655440000"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid UUID version 3")
	})

	t.Run("invalid_uuid5_as_uuid3_fails", func(t *testing.T) {
		// UUID v5 (has '5' at position 14)
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[
			{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"},
			{"id":"550e8400-e29b-41d4-a716-446655440000"}
		]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[1].ID", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID3Map tests UUID version 3 constraint in map values via dive.
func TestDive_UUID3Map(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid3"`
	}
	type Registry struct {
		Entities map[string]Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"item1":{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_version_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"item2":{"id":"550e8400-e29b-41d4-a716-446655440000"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID4 tests UUID version 4 constraint in nested structs via dive.
// uuid4 requires version byte (position 14) to be '4'.
func TestDive_UUID4(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid4"`
	}
	type Container struct {
		Entities []Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_uuid4_in_slice", func(t *testing.T) {
		// UUID v4 with version byte '4' at position 14
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"f47ac10b-58cc-4372-a567-0e02b2c3d479"}]}`))
		require.NoError(t, err)
	})

	t.Run("another_valid_uuid4", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"550e8400-e29b-41d4-a716-446655440000"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_uuid3_as_uuid4_fails", func(t *testing.T) {
		// UUID v3 (has '3' at position 14)
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid UUID version 4")
	})

	t.Run("invalid_uuid5_as_uuid4_fails", func(t *testing.T) {
		// UUID v5 (has '5' at position 14)
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[
			{"id":"f47ac10b-58cc-4372-a567-0e02b2c3d479"},
			{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"}
		]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[1].ID", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID4Map tests UUID version 4 constraint in map values via dive.
func TestDive_UUID4Map(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid4"`
	}
	type Registry struct {
		Entities map[string]Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"item1":{"id":"f47ac10b-58cc-4372-a567-0e02b2c3d479"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_version_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"item2":{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID5 tests UUID version 5 constraint in nested structs via dive.
// uuid5 requires version byte (position 14) to be '5'.
func TestDive_UUID5(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid5"`
	}
	type Container struct {
		Entities []Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_uuid5_in_slice", func(t *testing.T) {
		// UUID v5 with version byte '5' at position 14
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_uuid3_as_uuid5_fails", func(t *testing.T) {
		// UUID v3 (has '3' at position 14)
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"6ba7b810-9dad-31d0-80b4-00c04fd430c8"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
		assert.Contains(t, ve.Errors[0].Message, "must be a valid UUID version 5")
	})

	t.Run("invalid_uuid4_as_uuid5_fails", func(t *testing.T) {
		// UUID v4 (has '4' at position 14)
		_, err := validator.Unmarshal([]byte(`{"entities":[{"id":"550e8400-e29b-41d4-a716-446655440000"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[0].ID", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[
			{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"},
			{"id":"550e8400-e29b-41d4-a716-446655440000"}
		]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Entities[1].ID", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":[]}`))
		require.NoError(t, err)
	})
}

// TestDive_UUID5Map tests UUID version 5 constraint in map values via dive.
func TestDive_UUID5Map(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid5"`
	}
	type Registry struct {
		Entities map[string]Entity `json:"entities" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("valid_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"item1":{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}}}`))
		require.NoError(t, err)
	})

	t.Run("invalid_version_in_map", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{"item2":{"id":"550e8400-e29b-41d4-a716-446655440000"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty_map_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"entities":{}}`))
		require.NoError(t, err)
	})
}

// TestDive_AllFormatsSlice tests all format constraints in a single slice dive scenario.
func TestDive_AllFormatsSlice(t *testing.T) {
	type Resource struct {
		URL      string `json:"url" pedantigo:"url"`
		URI      string `json:"uri" pedantigo:"uri"`
		HTTPURL  string `json:"http_url" pedantigo:"http_url"`
		HTTPSURL string `json:"https_url" pedantigo:"https_url"`
		UUID     string `json:"uuid" pedantigo:"uuid"`
		UUID3    string `json:"uuid3" pedantigo:"uuid3"`
		UUID4    string `json:"uuid4" pedantigo:"uuid4"`
		UUID5    string `json:"uuid5" pedantigo:"uuid5"`
	}
	type Container struct {
		Resources []Resource `json:"resources" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("all_valid_formats", func(t *testing.T) {
		jsonData := `{
			"resources": [{
				"url": "https://example.com/path",
				"uri": "urn:isbn:0451450523",
				"http_url": "http://api.example.com",
				"https_url": "https://secure.example.com",
				"uuid": "550e8400-e29b-41d4-a716-446655440000",
				"uuid3": "6ba7b810-9dad-31d0-80b4-00c04fd430c8",
				"uuid4": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				"uuid5": "886313e1-3b8a-5372-9b90-0c9aee199e5d"
			}]
		}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.NoError(t, err)
	})

	t.Run("invalid_url_format", func(t *testing.T) {
		jsonData := `{
			"resources": [{
				"url": "not-a-url",
				"uri": "urn:isbn:0451450523",
				"http_url": "http://api.example.com",
				"https_url": "https://secure.example.com",
				"uuid": "550e8400-e29b-41d4-a716-446655440000",
				"uuid3": "6ba7b810-9dad-31d0-80b4-00c04fd430c8",
				"uuid4": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				"uuid5": "886313e1-3b8a-5372-9b90-0c9aee199e5d"
			}]
		}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Resources[0].URL", ve.Errors[0].Field)
	})

	t.Run("invalid_uuid4_version", func(t *testing.T) {
		jsonData := `{
			"resources": [{
				"url": "https://example.com/path",
				"uri": "urn:isbn:0451450523",
				"http_url": "http://api.example.com",
				"https_url": "https://secure.example.com",
				"uuid": "550e8400-e29b-41d4-a716-446655440000",
				"uuid3": "6ba7b810-9dad-31d0-80b4-00c04fd430c8",
				"uuid4": "6ba7b810-9dad-31d0-80b4-00c04fd430c8",
				"uuid5": "886313e1-3b8a-5372-9b90-0c9aee199e5d"
			}]
		}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Resources[0].UUID4", ve.Errors[0].Field)
	})
}

// TestDive_AllFormatsMap tests all format constraints in a single map dive scenario.
func TestDive_AllFormatsMap(t *testing.T) {
	type Resource struct {
		URL      string `json:"url" pedantigo:"url"`
		UUID     string `json:"uuid" pedantigo:"uuid"`
		HTTPSURL string `json:"https_url" pedantigo:"https_url"`
	}
	type Registry struct {
		Resources map[string]Resource `json:"resources" pedantigo:"dive"`
	}

	validator := New[Registry]()

	t.Run("all_valid_formats_in_map", func(t *testing.T) {
		jsonData := `{
			"resources": {
				"resource1": {
					"url": "https://example.com",
					"uuid": "550e8400-e29b-41d4-a716-446655440000",
					"https_url": "https://secure.example.com"
				}
			}
		}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.NoError(t, err)
	})

	t.Run("invalid_https_url_in_map", func(t *testing.T) {
		jsonData := `{
			"resources": {
				"resource1": {
					"url": "https://example.com",
					"uuid": "550e8400-e29b-41d4-a716-446655440000",
					"https_url": "http://insecure.example.com"
				}
			}
		}`
		_, err := validator.Unmarshal([]byte(jsonData))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})
}

// ==================== Network Dive Tests ====================
// Tests for dive tag with network constraints in nested structs (slices and maps).

// TestDive_IP tests IP constraint in nested structs via dive.
func TestDive_IP(t *testing.T) {
	type Server struct {
		Address string `json:"address" pedantigo:"ip"`
	}
	type Container struct {
		Servers []Server `json:"servers" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ipv4_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"192.168.1.1"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_ipv6_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_ipv6_short_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"::1"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"not-an-ip"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"192.168.1.1"},{"address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid_ipv4", func(t *testing.T) {
		type MapContainer struct {
			Servers map[string]Server `json:"servers" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"servers":{"primary":{"address":"192.168.1.1"},"secondary":{"address":"10.0.0.1"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Servers map[string]Server `json:"servers" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"servers":{"primary":{"address":"invalid-ip"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[primary].Address", ve.Errors[0].Field)
	})
}

// TestDive_IPv4 tests IPv4 constraint in nested structs via dive.
func TestDive_IPv4(t *testing.T) {
	type Server struct {
		Address string `json:"address" pedantigo:"ipv4"`
	}
	type Container struct {
		Servers []Server `json:"servers" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ipv4_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"192.168.1.1"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_ipv4", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"192.168.1.1"},{"address":"10.0.0.1"},{"address":"172.16.0.1"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ipv6_rejected", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"2001:0db8:85a3::1"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"not-an-ip"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"192.168.1.1"},{"address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Servers map[string]Server `json:"servers" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"servers":{"primary":{"address":"192.168.1.1"},"secondary":{"address":"10.0.0.1"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Servers map[string]Server `json:"servers" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"servers":{"primary":{"address":"2001:db8::1"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[primary].Address", ve.Errors[0].Field)
	})
}

// TestDive_IPv6 tests IPv6 constraint in nested structs via dive.
func TestDive_IPv6(t *testing.T) {
	type Server struct {
		Address string `json:"address" pedantigo:"ipv6"`
	}
	type Container struct {
		Servers []Server `json:"servers" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ipv6_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_ipv6_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"::1"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_ipv6", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"2001:db8::1"},{"address":"::1"},{"address":"fe80::1"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ipv4_rejected", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"192.168.1.1"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"not-an-ip"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[{"address":"2001:db8::1"},{"address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"servers":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Servers map[string]Server `json:"servers" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"servers":{"primary":{"address":"2001:db8::1"},"secondary":{"address":"::1"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Servers map[string]Server `json:"servers" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"servers":{"primary":{"address":"192.168.1.1"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Servers[primary].Address", ve.Errors[0].Field)
	})
}

// TestDive_CIDR tests CIDR constraint in nested structs via dive.
func TestDive_CIDR(t *testing.T) {
	type Network struct {
		Subnet string `json:"subnet" pedantigo:"cidr"`
	}
	type Container struct {
		Networks []Network `json:"networks" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_cidrv4_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_cidrv6_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"2001:db8::/32"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_cidr", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"},{"subnet":"2001:db8::/32"},{"subnet":"10.0.0.0/8"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_prefix", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[0].Subnet", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"not-a-cidr"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[0].Subnet", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"},{"subnet":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[1].Subnet", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Networks map[string]Network `json:"networks" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"networks":{"lan":{"subnet":"192.168.1.0/24"},"wan":{"subnet":"2001:db8::/32"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Networks map[string]Network `json:"networks" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"networks":{"lan":{"subnet":"invalid-cidr"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[lan].Subnet", ve.Errors[0].Field)
	})
}

// TestDive_CIDRv4 tests CIDRv4 constraint in nested structs via dive.
func TestDive_CIDRv4(t *testing.T) {
	type Network struct {
		Subnet string `json:"subnet" pedantigo:"cidrv4"`
	}
	type Container struct {
		Networks []Network `json:"networks" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_cidrv4_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_cidrv4", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"},{"subnet":"10.0.0.0/8"},{"subnet":"172.16.0.0/16"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_cidrv6_rejected", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"2001:db8::/32"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[0].Subnet", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"not-a-cidr"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[0].Subnet", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"},{"subnet":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[1].Subnet", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Networks map[string]Network `json:"networks" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"networks":{"lan":{"subnet":"192.168.1.0/24"},"wan":{"subnet":"10.0.0.0/8"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Networks map[string]Network `json:"networks" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"networks":{"lan":{"subnet":"2001:db8::/32"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[lan].Subnet", ve.Errors[0].Field)
	})
}

// TestDive_CIDRv6 tests CIDRv6 constraint in nested structs via dive.
func TestDive_CIDRv6(t *testing.T) {
	type Network struct {
		Subnet string `json:"subnet" pedantigo:"cidrv6"`
	}
	type Container struct {
		Networks []Network `json:"networks" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_cidrv6_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"2001:db8::/32"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_cidrv6", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"2001:db8::/32"},{"subnet":"fe80::/10"},{"subnet":"::1/128"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_cidrv4_rejected", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"192.168.1.0/24"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[0].Subnet", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"not-a-cidr"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[0].Subnet", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[{"subnet":"2001:db8::/32"},{"subnet":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[1].Subnet", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"networks":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Networks map[string]Network `json:"networks" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"networks":{"lan":{"subnet":"2001:db8::/32"},"wan":{"subnet":"fe80::/10"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Networks map[string]Network `json:"networks" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"networks":{"lan":{"subnet":"192.168.1.0/24"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Networks[lan].Subnet", ve.Errors[0].Field)
	})
}

// TestDive_MAC tests MAC constraint in nested structs via dive.
func TestDive_MAC(t *testing.T) {
	type Device struct {
		MACAddress string `json:"mac_address" pedantigo:"mac"`
	}
	type Container struct {
		Devices []Device `json:"devices" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_mac_colon_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[{"mac_address":"00:1B:44:11:3A:B7"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mac_hyphen_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[{"mac_address":"00-1B-44-11-3A-B7"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_mac", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[{"mac_address":"00:1B:44:11:3A:B7"},{"mac_address":"AA-BB-CC-DD-EE-FF"},{"mac_address":"01:23:45:67:89:ab"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[{"mac_address":"not-a-mac"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Devices[0].MACAddress", ve.Errors[0].Field)
	})

	t.Run("invalid_short_mac", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[{"mac_address":"00:1B:44"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Devices[0].MACAddress", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[{"mac_address":"00:1B:44:11:3A:B7"},{"mac_address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Devices[1].MACAddress", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"devices":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Devices map[string]Device `json:"devices" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"devices":{"eth0":{"mac_address":"00:1B:44:11:3A:B7"},"eth1":{"mac_address":"AA-BB-CC-DD-EE-FF"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Devices map[string]Device `json:"devices" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"devices":{"eth0":{"mac_address":"invalid-mac"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Devices[eth0].MACAddress", ve.Errors[0].Field)
	})
}

// TestDive_Hostname tests hostname constraint in nested structs via dive.
func TestDive_Hostname(t *testing.T) {
	type Host struct {
		Name string `json:"name" pedantigo:"hostname"`
	}
	type Container struct {
		Hosts []Host `json:"hosts" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hostname_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"localhost"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_single_letter", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"a"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_hyphen", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"my-server"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_dot", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"server.example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[0].Name", ve.Errors[0].Field)
	})

	t.Run("invalid_starts_with_digit", func(t *testing.T) {
		// RFC 952 requires hostname to start with letter
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"1server"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[0].Name", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"server1"},{"name":"invalid.host"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[1].Name", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Hosts map[string]Host `json:"hosts" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"hosts":{"primary":{"name":"server1"},"secondary":{"name":"server2"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Hosts map[string]Host `json:"hosts" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"hosts":{"primary":{"name":"server.example.com"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[primary].Name", ve.Errors[0].Field)
	})
}

// TestDive_HostnameRFC1123 tests hostname_rfc1123 constraint in nested structs via dive.
func TestDive_HostnameRFC1123(t *testing.T) {
	type Host struct {
		Name string `json:"name" pedantigo:"hostname_rfc1123"`
	}
	type Container struct {
		Hosts []Host `json:"hosts" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hostname_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"my-host"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_starts_with_digit", func(t *testing.T) {
		// RFC 1123 allows starting with digit
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"1server"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_hyphen", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"server-01"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_dot", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"server.example.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[0].Name", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[{"name":"server1"},{"name":"invalid.host"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[1].Name", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"hosts":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Hosts map[string]Host `json:"hosts" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"hosts":{"primary":{"name":"1server"},"secondary":{"name":"server-01"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Hosts map[string]Host `json:"hosts" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"hosts":{"primary":{"name":"server.example.com"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Hosts[primary].Name", ve.Errors[0].Field)
	})
}

// TestDive_FQDN tests fqdn constraint in nested structs via dive.
func TestDive_FQDN(t *testing.T) {
	type Domain struct {
		Name string `json:"name" pedantigo:"fqdn"`
	}
	type Container struct {
		Domains []Domain `json:"domains" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_fqdn_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[{"name":"www.example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_fqdn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[{"name":"www.example.com"},{"name":"mail.server.org"},{"name":"api.v2.company.co.uk"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_trailing_dot", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[{"name":"www.example.com."}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_single_label", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[{"name":"localhost"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Domains[0].Name", ve.Errors[0].Field)
	})

	t.Run("invalid_ip_address", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[{"name":"192.168.1.1"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Domains[0].Name", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[{"name":"www.example.com"},{"name":"localhost"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Domains[1].Name", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"domains":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Domains map[string]Domain `json:"domains" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"domains":{"web":{"name":"www.example.com"},"mail":{"name":"mail.example.com"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Domains map[string]Domain `json:"domains" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"domains":{"web":{"name":"localhost"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Domains[web].Name", ve.Errors[0].Field)
	})
}

// TestDive_Port tests port constraint in nested structs via dive.
func TestDive_Port(t *testing.T) {
	type Service struct {
		Port int `json:"port" pedantigo:"port"`
	}
	type Container struct {
		Services []Service `json:"services" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_port_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":80}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_ports", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":80},{"port":443},{"port":8080}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_port_zero", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":0}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_port_max", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":65535}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_negative_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":-1}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Services[0].Port", ve.Errors[0].Field)
	})

	t.Run("invalid_port_too_high", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":65536}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Services[0].Port", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[{"port":80},{"port":99999}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Services[1].Port", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"services":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Services map[string]Service `json:"services" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"services":{"http":{"port":80},"https":{"port":443}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Services map[string]Service `json:"services" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"services":{"http":{"port":70000}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Services[http].Port", ve.Errors[0].Field)
	})
}

// TestDive_TCPAddr tests tcp_addr constraint in nested structs via dive.
func TestDive_TCPAddr(t *testing.T) {
	type Endpoint struct {
		Address string `json:"address" pedantigo:"tcp_addr"`
	}
	type Container struct {
		Endpoints []Endpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_tcp_addr_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_hostname", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:443"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_tcp_addr", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"},{"address":"localhost:443"},{"address":"example.com:80"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_ipv6", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"[::1]:8080"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"not-a-tcp-addr"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"},{"address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"primary":{"address":"192.168.1.1:8080"},"secondary":{"address":"localhost:443"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"primary":{"address":"192.168.1.1"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[primary].Address", ve.Errors[0].Field)
	})
}

// TestDive_TCP4Addr tests tcp4_addr constraint in nested structs via dive.
func TestDive_TCP4Addr(t *testing.T) {
	type Endpoint struct {
		Address string `json:"address" pedantigo:"tcp4_addr"`
	}
	type Container struct {
		Endpoints []Endpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_tcp4_addr_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_tcp4_addr", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"},{"address":"10.0.0.1:443"},{"address":"172.16.0.1:80"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_hostname_rejected", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:8080"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_ipv6_rejected", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"[::1]:8080"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_missing_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"},{"address":"localhost:443"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"primary":{"address":"192.168.1.1:8080"},"secondary":{"address":"10.0.0.1:443"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"primary":{"address":"localhost:8080"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[primary].Address", ve.Errors[0].Field)
	})
}

// TestDive_UDPAddr tests udp_addr constraint in nested structs via dive.
func TestDive_UDPAddr(t *testing.T) {
	type Endpoint struct {
		Address string `json:"address" pedantigo:"udp_addr"`
	}
	type Container struct {
		Endpoints []Endpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_udp_addr_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:53"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_hostname", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:53"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_udp_addr", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:53"},{"address":"localhost:123"},{"address":"example.com:161"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_ipv6", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"[::1]:53"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"not-a-udp-addr"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:53"},{"address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"dns":{"address":"192.168.1.1:53"},"ntp":{"address":"localhost:123"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"dns":{"address":"192.168.1.1"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[dns].Address", ve.Errors[0].Field)
	})
}

// TestDive_HostnamePort tests hostname_port constraint in nested structs via dive.
func TestDive_HostnamePort(t *testing.T) {
	type Endpoint struct {
		Address string `json:"address" pedantigo:"hostname_port"`
	}
	type Container struct {
		Endpoints []Endpoint `json:"endpoints" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hostname_port_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:8080"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_fqdn_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"example.com:443"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_ip_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"192.168.1.1:8080"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_multiple_hostname_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:8080"},{"address":"example.com:443"},{"address":"192.168.1.1:80"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_port_zero", func(t *testing.T) {
		// hostname_port requires port 1-65535 (NOT 0)
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:0"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_missing_port", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"not-a-hostname-port"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[0].Address", ve.Errors[0].Field)
	})

	t.Run("mixed_validity", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[{"address":"localhost:8080"},{"address":"invalid"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[1].Address", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"endpoints":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"web":{"address":"localhost:8080"},"api":{"address":"example.com:443"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Endpoints map[string]Endpoint `json:"endpoints" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"endpoints":{"web":{"address":"localhost"}}}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Endpoints[web].Address", ve.Errors[0].Field)
	})
}

// =============================================================================
// ENUM DIVE TESTS (oneof, oneofci with nested structs)
// =============================================================================

// TestOneOf_NestedDive tests that oneof constraint works correctly
// on fields inside nested structs accessed via dive.
func TestOneOf_NestedDive(t *testing.T) {
	type Item struct {
		Status string `json:"status" pedantigo:"oneof=pending active completed"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid status in nested slice", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"status":"active"}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "valid enum value should not error")
		assert.Equal(t, "active", result.Items[0].Status)
	})

	t.Run("invalid status in nested slice", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"status":"invalid"}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "invalid enum value should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Items[0].Status" {
				found = true
				assert.Contains(t, fieldErr.Message, "must be one of: pending, active, completed")
			}
		}
		assert.True(t, found, "expected error for Items[0].Status, got: %v", ve.Errors)
	})

	t.Run("case sensitivity test - uppercase should fail", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"status":"ACTIVE"}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "case-sensitive oneof should reject uppercase")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Items[0].Status" {
				found = true
				assert.Contains(t, fieldErr.Message, "must be one of: pending, active, completed")
			}
		}
		assert.True(t, found, "expected error for Items[0].Status with uppercase value, got: %v", ve.Errors)
	})

	t.Run("multiple items with mixed validity", func(t *testing.T) {
		// First item valid, second item invalid
		jsonData := []byte(`{"items":[{"status":"pending"},{"status":"wrong"}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "second item with invalid status should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Items[1].Status" {
				found = true
				assert.Contains(t, fieldErr.Message, "must be one of: pending, active, completed")
			}
		}
		assert.True(t, found, "expected error for Items[1].Status, got: %v", ve.Errors)
	})

	t.Run("empty slice should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "empty slice should not error")
		assert.Empty(t, result.Items)
	})

	t.Run("all valid values should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"status":"pending"},{"status":"active"},{"status":"completed"}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "all valid enum values should not error")
		assert.Len(t, result.Items, 3)
		assert.Equal(t, "pending", result.Items[0].Status)
		assert.Equal(t, "active", result.Items[1].Status)
		assert.Equal(t, "completed", result.Items[2].Status)
	})
}

// TestOneOfCI_NestedDive tests that oneofci (case-insensitive) constraint works correctly
// on fields inside nested structs accessed via dive.
func TestOneOfCI_NestedDive(t *testing.T) {
	type Item struct {
		Role string `json:"role" pedantigo:"oneofci=admin user guest"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("exact match passes", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"role":"admin"}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "exact match should not error")
		assert.Equal(t, "admin", result.Items[0].Role)
	})

	t.Run("uppercase match passes", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"role":"ADMIN"}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "uppercase match should not error for case-insensitive")
		assert.Equal(t, "ADMIN", result.Items[0].Role)
	})

	t.Run("mixed case match passes", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"role":"AdMiN"}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "mixed case match should not error for case-insensitive")
		assert.Equal(t, "AdMiN", result.Items[0].Role)
	})

	t.Run("invalid value fails", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"role":"superuser"}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "invalid enum value should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Items[0].Role" {
				found = true
				assert.Contains(t, fieldErr.Message, "must be one of (case-insensitive): admin, user, guest")
			}
		}
		assert.True(t, found, "expected error for Items[0].Role, got: %v", ve.Errors)
	})

	t.Run("multiple items with mixed validity", func(t *testing.T) {
		// First valid (uppercase), second valid (mixed case), third invalid
		jsonData := []byte(`{"items":[{"role":"USER"},{"role":"GuesT"},{"role":"invalid"}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "third item with invalid role should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Items[2].Role" {
				found = true
				assert.Contains(t, fieldErr.Message, "must be one of (case-insensitive): admin, user, guest")
			}
		}
		assert.True(t, found, "expected error for Items[2].Role, got: %v", ve.Errors)
	})

	t.Run("empty slice should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "empty slice should not error")
		assert.Empty(t, result.Items)
	})

	t.Run("all valid values with different cases should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"role":"admin"},{"role":"USER"},{"role":"GuesT"}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "all valid enum values with different cases should not error")
		assert.Len(t, result.Items, 3)
		assert.Equal(t, "admin", result.Items[0].Role)
		assert.Equal(t, "USER", result.Items[1].Role)
		assert.Equal(t, "GuesT", result.Items[2].Role)
	})
}

// TestOneOf_NestedMapDive tests that oneof constraint works correctly
// in map value structs accessed via dive.
func TestOneOf_NestedMapDive(t *testing.T) {
	type Config struct {
		Environment string `json:"environment" pedantigo:"oneof=dev staging production"`
	}
	type Configs struct {
		ByRegion map[string]Config `json:"by_region" pedantigo:"dive"`
	}

	validator := New[Configs]()

	t.Run("valid environment in map should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_region":{"us-east":{"environment":"production"}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "valid enum value in map should not error")
		assert.Equal(t, "production", result.ByRegion["us-east"].Environment)
	})

	t.Run("invalid environment in map should fail", func(t *testing.T) {
		jsonData := []byte(`{"by_region":{"us-west":{"environment":"local"}}}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "invalid enum value in map should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		// Error should mention the field path including map key
		assert.NotEmpty(t, ve.Errors, "expected validation errors")
	})

	t.Run("multiple map entries with mixed validity", func(t *testing.T) {
		jsonData := []byte(`{"by_region":{"us-east":{"environment":"dev"},"eu-west":{"environment":"invalid"}}}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "invalid enum value in one map entry should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("empty map should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_region":{}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "empty map should not error")
		assert.Empty(t, result.ByRegion)
	})

	t.Run("all valid environments in map should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_region":{"us-east":{"environment":"production"},"us-west":{"environment":"staging"},"eu":{"environment":"dev"}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "all valid enum values in map should not error")
		assert.Len(t, result.ByRegion, 3)
		assert.Equal(t, "production", result.ByRegion["us-east"].Environment)
		assert.Equal(t, "staging", result.ByRegion["us-west"].Environment)
		assert.Equal(t, "dev", result.ByRegion["eu"].Environment)
	})
}

// TestOneOfCI_NestedMapDive tests that oneofci constraint works correctly
// in map value structs accessed via dive.
func TestOneOfCI_NestedMapDive(t *testing.T) {
	type Permission struct {
		Level string `json:"level" pedantigo:"oneofci=read write admin"`
	}
	type Permissions struct {
		ByUser map[string]Permission `json:"by_user" pedantigo:"dive"`
	}

	validator := New[Permissions]()

	t.Run("exact match in map passes", func(t *testing.T) {
		jsonData := []byte(`{"by_user":{"alice":{"level":"read"}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "exact match should not error")
		assert.Equal(t, "read", result.ByUser["alice"].Level)
	})

	t.Run("uppercase match in map passes", func(t *testing.T) {
		jsonData := []byte(`{"by_user":{"bob":{"level":"WRITE"}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "uppercase match should not error for case-insensitive")
		assert.Equal(t, "WRITE", result.ByUser["bob"].Level)
	})

	t.Run("mixed case match in map passes", func(t *testing.T) {
		jsonData := []byte(`{"by_user":{"charlie":{"level":"AdMiN"}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "mixed case match should not error for case-insensitive")
		assert.Equal(t, "AdMiN", result.ByUser["charlie"].Level)
	})

	t.Run("invalid level in map fails", func(t *testing.T) {
		jsonData := []byte(`{"by_user":{"dave":{"level":"execute"}}}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "invalid enum value in map should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("multiple map entries with different cases should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_user":{"alice":{"level":"read"},"bob":{"level":"WRITE"},"charlie":{"level":"AdMiN"}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "all valid enum values with different cases should not error")
		assert.Len(t, result.ByUser, 3)
		assert.Equal(t, "read", result.ByUser["alice"].Level)
		assert.Equal(t, "WRITE", result.ByUser["bob"].Level)
		assert.Equal(t, "AdMiN", result.ByUser["charlie"].Level)
	})

	t.Run("empty map should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_user":{}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "empty map should not error")
		assert.Empty(t, result.ByUser)
	})
}

// =============================================================================
// STRING CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Alpha tests alpha constraint in nested structs via dive.
func TestDive_Alpha(t *testing.T) {
	type Item struct {
		Code string `json:"code" pedantigo:"alpha"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_alpha", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC"},{"code":"xyz"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_numbers", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC123"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Code", ve.Errors[0].Field)
	})

	t.Run("invalid_with_space", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"AB C"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"ABC"},"b":{"code":"XYZ"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"ABC123"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Alphanum tests alphanum constraint in nested structs via dive.
func TestDive_Alphanum(t *testing.T) {
	type Item struct {
		Code string `json:"code" pedantigo:"alphanum"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_alphanum", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC123"},{"code":"xyz789"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_special", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC@123"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Code", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"ABC123"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"ABC-123"}}}`))
		require.Error(t, err)
	})
}

// TestDive_ASCII tests ascii constraint in nested structs via dive.
func TestDive_ASCII(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"ascii"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ascii", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"Hello World 123!"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_unicode", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"Hello 世界"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"ASCII only"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"café"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Contains tests contains constraint in nested structs via dive.
func TestDive_Contains(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"contains=@"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_contains", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"user@example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_char", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"userexample.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"test@test"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"no at symbol"}}}`))
		require.Error(t, err)
	})
}

// TestDive_StartsWith tests startswith constraint in nested structs via dive.
func TestDive_StartsWith(t *testing.T) {
	type Item struct {
		Path string `json:"path" pedantigo:"startswith=/api/"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_startswith", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"path":"/api/users"},{"path":"/api/products"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_wrong_prefix", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"path":"/web/users"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Path", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"path":"/api/test"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"path":"/v1/test"}}}`))
		require.Error(t, err)
	})
}

// TestDive_EndsWith tests endswith constraint in nested structs via dive.
func TestDive_EndsWith(t *testing.T) {
	type Item struct {
		Filename string `json:"filename" pedantigo:"endswith=.json"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_endswith", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"filename":"config.json"},{"filename":"data.json"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_wrong_suffix", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"filename":"config.yaml"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Filename", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"filename":"test.json"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"filename":"test.xml"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Lowercase tests lowercase constraint in nested structs via dive.
func TestDive_Lowercase(t *testing.T) {
	type Item struct {
		Tag string `json:"tag" pedantigo:"lowercase"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tag":"hello"},{"tag":"world"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_has_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tag":"Hello"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Tag", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"tag":"lowercase"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"tag":"UPPERCASE"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Uppercase tests uppercase constraint in nested structs via dive.
func TestDive_Uppercase(t *testing.T) {
	type Item struct {
		Code string `json:"code" pedantigo:"uppercase"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC"},{"code":"XYZ"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_has_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"Abc"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Code", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"UPPERCASE"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"lowercase"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Len tests len constraint in nested structs via dive.
func TestDive_Len(t *testing.T) {
	type Item struct {
		Code string `json:"code" pedantigo:"len=5"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_exact_length", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABCDE"},{"code":"12345"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Code", ve.Errors[0].Field)
	})

	t.Run("invalid_too_long", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABCDEFG"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"ABCDE"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"AB"}}}`))
		require.Error(t, err)
	})
}

// TestDive_MinLen tests min constraint on strings in nested structs via dive.
func TestDive_MinLen(t *testing.T) {
	type Item struct {
		Name string `json:"name" pedantigo:"min=3"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_min_length", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"name":"John"},{"name":"Alice"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_exact_min", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"name":"Bob"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"name":"Jo"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Name", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"name":"John"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"name":"Jo"}}}`))
		require.Error(t, err)
	})
}

// TestDive_MaxLen tests max constraint on strings in nested structs via dive.
func TestDive_MaxLen(t *testing.T) {
	type Item struct {
		Code string `json:"code" pedantigo:"max=5"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_under_max", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABC"},{"code":"XY"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_exact_max", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABCDE"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_long", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"code":"ABCDEFG"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Code", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"ABC"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"code":"TOOLONGCODE"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Regexp tests regexp constraint in nested structs via dive.
func TestDive_Regexp(t *testing.T) {
	type Item struct {
		SKU string `json:"sku" pedantigo:"regexp=^[A-Z]{3}-[0-9]{4}$"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_pattern", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"sku":"ABC-1234"},{"sku":"XYZ-9999"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_wrong_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"sku":"AB-123"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].SKU", ve.Errors[0].Field)
	})

	t.Run("invalid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"sku":"abc-1234"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"sku":"ABC-1234"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"sku":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Excludes tests excludes constraint in nested structs via dive.
func TestDive_Excludes(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"excludes=<script>"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_no_script", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"Hello World"},{"text":"Safe content"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_has_script", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"<script>alert('xss')</script>"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"safe text"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"has <script> tag"}}}`))
		require.Error(t, err)
	})
}

// TestDive_PrintASCII tests printascii constraint in nested structs via dive.
func TestDive_PrintASCII(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"printascii"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_printable", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"Hello World!"},{"text":"Test 123"}]}`))
		require.NoError(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"Printable ASCII"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_Numeric tests numeric constraint in nested structs via dive.
func TestDive_Numeric(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"numeric"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_numeric_string", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"12345"},{"value":"67890"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_decimal", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"123.45"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_letters", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"123abc"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"12345"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"abc"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Number tests number constraint in nested structs via dive.
// Note: "number" constraint only accepts pure digits (0-9), no decimals or signs.
// Use "numeric" constraint for signed decimals.
func TestDive_Number(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"number"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_number_string", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"12345"},{"value":"67890"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_zero", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"0"},{"value":"007"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_letters", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"12abc"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("invalid_with_decimal", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"123.45"}]}`))
		require.Error(t, err)
	})

	t.Run("invalid_with_negative", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"-123"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"12345"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"not a number"}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// ENCODING CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Base64 tests base64 constraint in nested structs via dive.
func TestDive_Base64(t *testing.T) {
	type Item struct {
		Data string `json:"data" pedantigo:"base64"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_base64", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"SGVsbG8gV29ybGQ="},{"data":"dGVzdA=="}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_base64", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"not-valid-base64!!!"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Data", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"SGVsbG8="}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"###invalid###"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Base64URL tests base64url constraint in nested structs via dive.
func TestDive_Base64URL(t *testing.T) {
	type Item struct {
		Data string `json:"data" pedantigo:"base64url"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_base64url", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"SGVsbG8tV29ybGQ_"},{"data":"dGVzdA=="}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_base64url", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"not valid base64url!!!"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Data", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"SGVsbG8="}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"+++invalid+++"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Hexadecimal tests hexadecimal constraint in nested structs via dive.
func TestDive_Hexadecimal(t *testing.T) {
	type Item struct {
		Hex string `json:"hex" pedantigo:"hexadecimal"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hex", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hex":"DEADBEEF"},{"hex":"0123456789abcdef"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_hex", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hex":"GHIJK"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hex", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hex":"CAFE"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hex":"ZZZZ"}}}`))
		require.Error(t, err)
	})
}

// TestDive_HexColor tests hexcolor constraint in nested structs via dive.
func TestDive_HexColor(t *testing.T) {
	type Item struct {
		Color string `json:"color" pedantigo:"hexcolor"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hexcolor", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"#FF5733"},{"color":"#fff"},{"color":"#AABBCC"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_hexcolor", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"#GGGGGG"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Color", ve.Errors[0].Field)
	})

	t.Run("invalid_no_hash", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"FF5733"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"color":"#123ABC"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"color":"red"}}}`))
		require.Error(t, err)
	})
}

// TestDive_JSON tests json constraint in nested structs via dive.
func TestDive_JSON(t *testing.T) {
	type Item struct {
		Data string `json:"data" pedantigo:"json"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_json_object", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"{\"key\":\"value\"}"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_json_array", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"[1,2,3]"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"{invalid json}"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Data", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"{\"valid\":true}"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"not json"}}}`))
		require.Error(t, err)
	})
}

// TestDive_JWT tests jwt constraint in nested structs via dive.
func TestDive_JWT(t *testing.T) {
	type Item struct {
		Token string `json:"token" pedantigo:"jwt"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_jwt", func(t *testing.T) {
		// A valid JWT structure (header.payload.signature) - test token, not a real secret
		_, err := validator.Unmarshal([]byte(`{"items":[{"token":"aaa.bbb.ccc"}]}`)) // gitleaks:allow test token
		require.NoError(t, err)
	})

	t.Run("invalid_jwt", func(t *testing.T) {
		// Uses $ character which isn't valid in base64url
		_, err := validator.Unmarshal([]byte(`{"items":[{"token":"invalid$token.test.here"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Token", ve.Errors[0].Field)
	})

	t.Run("invalid_missing_parts", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"token":"only.two"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		// Test token with valid JWT structure (header.payload.signature)
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"token":"xxx.yyy.zzz"}}}`)) // gitleaks:allow test token
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"token":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Alphaspace tests alphaspace constraint in nested structs via dive.
func TestDive_Alphaspace(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"alphaspace"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_alphaspace", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello world"},{"value":"ABC XYZ"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_numbers", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello123"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("invalid_with_special_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello!world"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"hello world"},"b":{"value":"ABC XYZ"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"hello123"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Alphanumspace tests alphanumspace constraint in nested structs via dive.
func TestDive_Alphanumspace(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"alphanumspace"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_alphanumspace", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello world 123"},{"value":"ABC XYZ 789"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_special_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello!world"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("invalid_with_punctuation", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello, world"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"hello world 123"},"b":{"value":"ABC XYZ 789"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"hello!world"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Alphaunicode tests alphaunicode constraint in nested structs via dive.
func TestDive_Alphaunicode(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"alphaunicode"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_alphaunicode", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"héllo"},{"value":"世界"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_numbers", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello123"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("invalid_with_space", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello world"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"héllo"},"b":{"value":"世界"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"hello123"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Alphanumunicode tests alphanumunicode constraint in nested structs via dive.
func TestDive_Alphanumunicode(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"alphanumunicode"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_alphanumunicode", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"héllo123"},{"value":"世界789"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_with_special_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello!world"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("invalid_with_space", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"hello world"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"héllo123"},"b":{"value":"世界789"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"hello!world"}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// IDENTITY CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Email tests email constraint in nested structs via dive.
func TestDive_Email(t *testing.T) {
	type Item struct {
		Email string `json:"email" pedantigo:"email"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_emails", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"email":"user@example.com"},{"email":"test.user@domain.org"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_no_at", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"email":"userexample.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Email", ve.Errors[0].Field)
	})

	t.Run("invalid_no_domain", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"email":"user@"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"email":"test@test.com"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"email":"invalid-email"}}}`))
		require.Error(t, err)
	})
}

// TestDive_E164 tests e164 constraint in nested structs via dive.
func TestDive_E164(t *testing.T) {
	type Item struct {
		Phone string `json:"phone" pedantigo:"e164"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_e164", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"phone":"+14155552671"},{"phone":"+442071234567"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_no_plus", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"phone":"14155552671"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Phone", ve.Errors[0].Field)
	})

	t.Run("invalid_with_spaces", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"phone":"+1 415 555 2671"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"phone":"+14155552671"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"phone":"555-1234"}}}`))
		require.Error(t, err)
	})
}

// TestDive_ISBN tests isbn constraint in nested structs via dive.
func TestDive_ISBN(t *testing.T) {
	type Item struct {
		ISBN string `json:"isbn" pedantigo:"isbn"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_isbn13", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"978-3-16-148410-0"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_isbn10", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"0-306-40615-2"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_isbn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"123-456-789"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].ISBN", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"isbn":"978-3-16-148410-0"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"isbn":"not-an-isbn"}}}`))
		require.Error(t, err)
	})
}

// TestDive_ISBN10 tests isbn10 constraint in nested structs via dive.
func TestDive_ISBN10(t *testing.T) {
	type Item struct {
		ISBN string `json:"isbn" pedantigo:"isbn10"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_isbn10", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"0306406152"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_isbn13_as_isbn10", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"9783161484100"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"isbn":"0306406152"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_ISBN13 tests isbn13 constraint in nested structs via dive.
func TestDive_ISBN13(t *testing.T) {
	type Item struct {
		ISBN string `json:"isbn" pedantigo:"isbn13"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_isbn13", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"9783161484100"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_isbn10_as_isbn13", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"isbn":"0306406152"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"isbn":"9783161484100"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_ULID tests ulid constraint in nested structs via dive.
func TestDive_ULID(t *testing.T) {
	type Item struct {
		ID string `json:"id" pedantigo:"ulid"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ulid", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"id":"01ARZ3NDEKTSV4RRFFQ69G5FAV"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ulid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"id":"01ARZ3NDEK"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].ID", ve.Errors[0].Field)
	})

	t.Run("invalid_ulid_bad_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"id":"ILOU0000000000000000000000"}]}`)) // I, L, O, U are invalid
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"id":"01ARZ3NDEKTSV4RRFFQ69G5FAV"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"id":"not-a-ulid"}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// FINANCE CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_CreditCard tests credit_card constraint in nested structs via dive.
func TestDive_CreditCard(t *testing.T) {
	type Item struct {
		Card string `json:"card" pedantigo:"credit_card"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_visa", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card":"4111111111111111"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mastercard", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card":"5500000000000004"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_luhn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card":"4111111111111112"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Card", ve.Errors[0].Field)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card":"411111"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"card":"4111111111111111"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"card":"1234567890123456"}}}`))
		require.Error(t, err)
	})
}

// TestDive_SSN tests ssn constraint in nested structs via dive.
func TestDive_SSN(t *testing.T) {
	type Item struct {
		SSN string `json:"ssn" pedantigo:"ssn"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ssn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ssn":"123-45-6789"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ssn":"12345-6789"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].SSN", ve.Errors[0].Field)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ssn":"123-45-678"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"ssn":"123-45-6789"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"ssn":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_BtcAddr tests btc_addr constraint in nested structs via dive.
func TestDive_BtcAddr(t *testing.T) {
	type Item struct {
		Address string `json:"address" pedantigo:"btc_addr"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_p2pkh", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_p2sh", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_checksum", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN3"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_bech32", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"address":"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"address":"invalid-btc-address"}}}`))
		require.Error(t, err)
	})
}

// TestDive_BtcAddrBech32 tests btc_addr_bech32 constraint in nested structs via dive.
func TestDive_BtcAddrBech32(t *testing.T) {
	type Item struct {
		Address string `json:"address" pedantigo:"btc_addr_bech32"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_mainnet_p2wpkh", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mainnet_p2wsh", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_testnet", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_p2pkh", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_checksum", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t5"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"address":"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"address":"bc1invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_EthAddr tests eth_addr constraint in nested structs via dive.
func TestDive_EthAddr(t *testing.T) {
	type Item struct {
		Address string `json:"address" pedantigo:"eth_addr"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"0x742d35cc6634c0532925a3b844bc9e7595f8fee5"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mixed_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"0x742d35Cc6634C0532925a3b844Bc9e7595f8fEe5"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_all_zeros", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"0x0000000000000000000000000000000000000000"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_no_prefix", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"742d35cc6634c0532925a3b844bc9e7595f8fee5"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Address", ve.Errors[0].Field)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"0x742d35cc6634c0532925a3b844bc9e7595f8fe"}]}`))
		require.Error(t, err)
	})

	t.Run("invalid_char", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"address":"0x742d35cc6634c0532925a3b844bc9e7595f8feeg"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"address":"0xffffffffffffffffffffffffffffffffffffffff"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"address":"0xinvalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_LuhnChecksum tests luhn_checksum constraint in nested structs via dive.
func TestDive_LuhnChecksum(t *testing.T) {
	type Item struct {
		CardNumber string `json:"card_number" pedantigo:"luhn_checksum"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_visa", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"4111111111111111"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mastercard", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"5500000000000004"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_amex", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"378282246310005"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_short_number", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"79927398713"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_checksum", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"4111111111111112"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].CardNumber", ve.Errors[0].Field)
	})

	t.Run("invalid_with_letters", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"411111111111a111"}]}`))
		require.Error(t, err)
	})

	t.Run("invalid_with_spaces", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"card_number":"4111 1111 1111 1111"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"card_number":"79927398713"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"card_number":"1234567890"}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// HASH CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_MD5 tests md5 constraint in nested structs via dive.
func TestDive_MD5(t *testing.T) {
	type Item struct {
		Hash string `json:"hash" pedantigo:"md5"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_md5", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"d41d8cd98f00b204e9800998ecf8427e"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"d41d8cd98f00b204"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("invalid_bad_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"g41d8cd98f00b204e9800998ecf8427e"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"d41d8cd98f00b204e9800998ecf8427e"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_SHA256 tests sha256 constraint in nested structs via dive.
func TestDive_SHA256(t *testing.T) {
	type Item struct {
		Hash string `json:"hash" pedantigo:"sha256"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_sha256", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"e3b0c44298fc1c149afbf4c8996fb924"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_SHA512 tests sha512 constraint in nested structs via dive.
func TestDive_SHA512(t *testing.T) {
	type Item struct {
		Hash string `json:"hash" pedantigo:"sha512"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_sha512", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"cf83e1357eefb8bdf1542850d66d8007"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"}}}`))
		require.NoError(t, err)
	})
}

// =============================================================================
// DATETIME CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Datetime tests datetime constraint in nested structs via dive.
func TestDive_Datetime(t *testing.T) {
	type Item struct {
		Date string `json:"date" pedantigo:"datetime=2006-01-02"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_datetime", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"date":"2024-01-15"},{"date":"2023-12-31"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"date":"01-15-2024"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Date", ve.Errors[0].Field)
	})

	t.Run("invalid_not_date", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"date":"not-a-date"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"date":"2024-06-15"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"date":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Timezone tests timezone constraint in nested structs via dive.
func TestDive_Timezone(t *testing.T) {
	type Item struct {
		TZ string `json:"tz" pedantigo:"timezone"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_timezone", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tz":"America/New_York"},{"tz":"Europe/London"},{"tz":"UTC"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_timezone", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tz":"Invalid/Timezone"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].TZ", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"tz":"America/Los_Angeles"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"tz":"Not/Real"}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// GEO CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Latitude tests latitude constraint in nested structs via dive.
func TestDive_Latitude(t *testing.T) {
	type Item struct {
		Lat float64 `json:"lat" pedantigo:"latitude"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_latitude", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"lat":40.7128},{"lat":-33.8688},{"lat":0}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_high", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"lat":91.0}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Lat", ve.Errors[0].Field)
	})

	t.Run("invalid_too_low", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"lat":-91.0}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"lat":51.5074}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"lat":200.0}}}`))
		require.Error(t, err)
	})
}

// TestDive_Longitude tests longitude constraint in nested structs via dive.
func TestDive_Longitude(t *testing.T) {
	type Item struct {
		Lng float64 `json:"lng" pedantigo:"longitude"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_longitude", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"lng":-74.0060},{"lng":151.2093},{"lng":0}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_high", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"lng":181.0}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Lng", ve.Errors[0].Field)
	})

	t.Run("invalid_too_low", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"lng":-181.0}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"lng":-0.1278}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"lng":200.0}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// COLOR CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_RGB tests rgb constraint in nested structs via dive.
func TestDive_RGB(t *testing.T) {
	type Item struct {
		Color string `json:"color" pedantigo:"rgb"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_rgb", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"rgb(255, 0, 0)"},{"color":"rgb(0,128,255)"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_rgb", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"rgb(300, 0, 0)"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Color", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"color":"rgb(0, 0, 0)"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_RGBA tests rgba constraint in nested structs via dive.
func TestDive_RGBA(t *testing.T) {
	type Item struct {
		Color string `json:"color" pedantigo:"rgba"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_rgba", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"rgba(255, 0, 0, 0.5)"},{"color":"rgba(0,128,255,1)"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_rgba", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"rgba(255, 0, 0, 2)"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Color", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"color":"rgba(0, 0, 0, 0)"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_HSL tests hsl constraint in nested structs via dive.
func TestDive_HSL(t *testing.T) {
	type Item struct {
		Color string `json:"color" pedantigo:"hsl"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hsl", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"hsl(0, 100%, 50%)"},{"color":"hsl(240,50%,50%)"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_hsl", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"hsl(400, 100%, 50%)"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Color", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"color":"hsl(120, 100%, 50%)"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_HSLA tests hsla constraint in nested structs via dive.
func TestDive_HSLA(t *testing.T) {
	type Item struct {
		Color string `json:"color" pedantigo:"hsla"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_hsla", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"hsla(0, 100%, 50%, 0.5)"},{"color":"hsla(240,50%,50%,1)"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_hsla", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"color":"hsla(0, 100%, 50%, 2)"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Color", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"color":"hsla(120, 100%, 50%, 0.5)"}}}`))
		require.NoError(t, err)
	})
}

// =============================================================================
// COMPARISON CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Eq tests eq constraint in nested structs via dive.
func TestDive_Eq(t *testing.T) {
	type Item struct {
		Status string `json:"status" pedantigo:"eq=active"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_eq", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"active"},{"status":"active"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ne", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"inactive"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Status", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"status":"active"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"status":"disabled"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Ne tests ne constraint in nested structs via dive.
func TestDive_Ne(t *testing.T) {
	type Item struct {
		Status string `json:"status" pedantigo:"ne=deleted"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_ne", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"active"},{"status":"pending"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_eq", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"deleted"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Status", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"status":"active"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"status":"deleted"}}}`))
		require.Error(t, err)
	})
}

// TestDive_EqIgnoreCase tests eq_ignore_case constraint in nested structs via dive.
func TestDive_EqIgnoreCase(t *testing.T) {
	type Item struct {
		Type string `json:"type" pedantigo:"eq_ignore_case=premium"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_exact", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"type":"premium"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"type":"PREMIUM"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mixed", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"type":"PrEmIuM"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_different", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"type":"basic"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Type", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"type":"PREMIUM"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_NeIgnoreCase tests ne_ignore_case constraint in nested structs via dive.
func TestDive_NeIgnoreCase(t *testing.T) {
	type Item struct {
		Status string `json:"status" pedantigo:"ne_ignore_case=deleted"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_different", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"active"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_different_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"ACTIVE"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_exact_match", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"deleted"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Status", ve.Errors[0].Field)
	})

	t.Run("invalid_case_insensitive_match", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"status":"DELETED"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"status":"active"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"status":"DeLeTed"}}}`))
		require.Error(t, err)
	})
}

// TestDive_ContainsRune tests containsrune constraint in nested structs via dive.
func TestDive_ContainsRune(t *testing.T) {
	type Item struct {
		Value string `json:"value" pedantigo:"containsrune=@"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_contains_rune", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"test@example.com"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_rune", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":"testexample.com"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"user@domain"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":"nodomain"}}}`))
		require.Error(t, err)
	})
}

// =============================================================================
// MISC CONSTRAINT DIVE TESTS
// =============================================================================

// TestDive_Semver tests semver constraint in nested structs via dive.
func TestDive_Semver(t *testing.T) {
	type Item struct {
		Version string `json:"version" pedantigo:"semver"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_semver", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"version":"1.0.0"},{"version":"2.1.3"},{"version":"0.0.1-alpha"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_semver", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"version":"1.0"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Version", ve.Errors[0].Field)
	})

	t.Run("invalid_not_semver", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"version":"version1"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"version":"1.2.3"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"version":"v1"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Image tests that image constraint works correctly on fields inside
// nested structs accessed via dive. Requires actual image files on disk.
func TestDive_Image(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "test_image_dive_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create valid PNG image file with proper magic bytes
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R', // IHDR chunk
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, 0x00, 0x00, 0x00, // bit depth, color type, etc.
		0x90, 0x77, 0x53, 0xDE, // IHDR CRC
	}
	validImagePath := filepath.Join(tmpDir, "valid.png")
	require.NoError(t, os.WriteFile(validImagePath, pngData, 0o600))

	// Create invalid file (not an image)
	invalidFilePath := filepath.Join(tmpDir, "invalid.txt")
	require.NoError(t, os.WriteFile(invalidFilePath, []byte("not an image"), 0o600))

	type Item struct {
		Path string `json:"path" pedantigo:"image"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_image_path", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"path":"` + validImagePath + `"}]}`)
		_, err := validator.Unmarshal(jsonData)
		require.NoError(t, err)
	})

	t.Run("invalid_not_image", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"path":"` + invalidFilePath + `"}]}`)
		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Path", ve.Errors[0].Field)
	})

	t.Run("invalid_file_not_found", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"path":"/nonexistent/path/to/image.png"}]}`)
		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Path", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		jsonData := []byte(`{"items":{"a":{"path":"` + validImagePath + `"}}}`)
		_, err := mapValidator.Unmarshal(jsonData)
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		jsonData := []byte(`{"items":{"a":{"path":"` + invalidFilePath + `"}}}`)
		_, err := mapValidator.Unmarshal(jsonData)
		require.Error(t, err)
	})
}

// TestDive_Md4 tests md4 constraint in nested structs via dive.
func TestDive_Md4(t *testing.T) {
	type Item struct {
		Hash string `json:"hash" pedantigo:"md4"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_md4", func(t *testing.T) {
		// 32 hex characters
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"d41d8cd98f00b204e9800998ecf8427e"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"d41d8cd98f00b204"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("invalid_bad_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"g41d8cd98f00b204e9800998ecf8427e"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"d41d8cd98f00b204e9800998ecf8427e"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"not-a-valid-md4"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Sha384 tests sha384 constraint in nested structs via dive.
func TestDive_Sha384(t *testing.T) {
	type Item struct {
		Hash string `json:"hash" pedantigo:"sha384"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_sha384", func(t *testing.T) {
		// 96 hex characters
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"38b060a751ac96384cd9327eb1b1e36a"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("invalid_bad_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"hash":"zzz060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Hash", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"hash":"not-a-valid-sha384"}}}`))
		require.Error(t, err)
	})
}

// TestDive_StripWhitespace tests strip_whitespace validator in nested structs via dive.
// strip_whitespace validates that a string has no leading/trailing whitespace.
func TestDive_StripWhitespace(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"strip_whitespace"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_no_whitespace", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello world"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_leading_whitespace", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"  hello"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("invalid_trailing_whitespace", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello  "}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"trimmed"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"  untrimmed  "}}}`))
		require.Error(t, err)
	})
}

// TestDive_ToLower tests to_lower validator in nested structs via dive.
// to_lower validates that a string is all lowercase.
func TestDive_ToLower(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"to_lower"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello world"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"HELLO"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("invalid_mixed_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"HeLLo"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"lowercase"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"UPPERCASE"}}}`))
		require.Error(t, err)
	})
}

// TestDive_ToUpper tests to_upper validator in nested structs via dive.
// to_upper validates that a string is all uppercase.
func TestDive_ToUpper(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"to_upper"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"HELLO WORLD"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("invalid_mixed_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"HeLLo"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"UPPERCASE"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"lowercase"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Base32 tests base32 constraint in nested structs via dive.
func TestDive_Base32(t *testing.T) {
	type Item struct {
		Data string `json:"data" pedantigo:"base32"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_base32", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"JBSWY3DPEHPK3PXP"},{"data":"MZXW6YTBOI======"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_base32", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"not-valid-base32!"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Data", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"JBSWY3DPEHPK3PXP"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"###invalid###"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Base64rawurl tests base64rawurl constraint in nested structs via dive.
func TestDive_Base64rawurl(t *testing.T) {
	type Item struct {
		Data string `json:"data" pedantigo:"base64rawurl"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_base64rawurl", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"SGVsbG8"},{"data":"dGVzdA"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_base64rawurl", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"not-valid-base64rawurl!!!"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Data", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"SGVsbG8"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"###invalid###"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Datauri tests datauri constraint in nested structs via dive.
func TestDive_Datauri(t *testing.T) {
	type Item struct {
		Data string `json:"data" pedantigo:"datauri"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_datauri", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"data:text/plain;base64,SGVsbG8="},{"data":"data:image/png;base64,iVBORw0KGgo="}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_datauri", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"data":"not-a-valid-datauri"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Data", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"data:text/plain;base64,SGVsbG8="}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"data":"invalid-data-uri"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Multibyte tests that multibyte constraint works correctly on fields inside
// nested structs accessed via dive. Validates presence of multibyte characters.
func TestDive_Multibyte(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"multibyte"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_japanese", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"日本語"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_chinese", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"你好"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_emoji", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello 👋"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ascii_only", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"привет"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"text":"plain"}}}`))
		require.Error(t, err)
	})
}

// TestDive_UrnRfc2141 tests that urn_rfc2141 constraint works correctly on fields inside
// nested structs accessed via dive. Validates URN format per RFC 2141.
func TestDive_UrnRfc2141(t *testing.T) {
	type Item struct {
		URN string `json:"urn" pedantigo:"urn_rfc2141"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_isbn_urn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"urn":"urn:isbn:0451450523"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_uuid_urn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"urn":"urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_prefix", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"urn":"isbn:0451450523"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].URN", ve.Errors[0].Field)
	})

	t.Run("invalid_not_urn", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"urn":"https://example.com"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"urn":"urn:ietf:rfc:2141"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"urn":"not-a-urn"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Mongodb tests that mongodb constraint works correctly on fields inside
// nested structs accessed via dive. Validates MongoDB ObjectId (24 hex chars).
func TestDive_Mongodb(t *testing.T) {
	type Item struct {
		ObjectID string `json:"object_id" pedantigo:"mongodb"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_lowercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"object_id":"507f1f77bcf86cd799439011"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_uppercase", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"object_id":"507F1F77BCF86CD799439011"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_mixed_case", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"object_id":"5d6ede6a0ba62570afcedd3a"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"object_id":"507f1f77bcf86cd7994390"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].ObjectID", ve.Errors[0].Field)
	})

	t.Run("invalid_non_hex", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"object_id":"507f1f77bcf86cd799439xyz"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"object_id":"000000000000000000000000"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"object_id":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Html tests that html constraint works correctly on fields inside
// nested structs accessed via dive. Validates presence of HTML tags.
func TestDive_Html(t *testing.T) {
	type Item struct {
		Content string `json:"content" pedantigo:"html"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_div_tag", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"content":"<div>Hello</div>"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_with_attributes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"content":"<a href=\"http://example.com\">link</a>"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_comment", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"content":"<!-- comment -->"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_no_tags", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"content":"plain text"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Content", ve.Errors[0].Field)
	})

	t.Run("invalid_escaped_html", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"content":"&lt;div&gt;"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"content":"<p>test</p>"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"content":"no html here"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Cron tests that cron constraint works correctly on fields inside
// nested structs accessed via dive. Validates cron expressions (5 fields).
func TestDive_Cron(t *testing.T) {
	type Item struct {
		Schedule string `json:"schedule" pedantigo:"cron"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_all_wildcards", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"schedule":"* * * * *"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_midnight_daily", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"schedule":"0 0 * * *"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_every_5_minutes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"schedule":"*/5 * * * *"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_complex", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"schedule":"*/15 9-17 * * 1-5"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_only_3_fields", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"schedule":"* * *"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Schedule", ve.Errors[0].Field)
	})

	t.Run("invalid_out_of_range", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"schedule":"60 * * * *"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"schedule":"0 0 1 * *"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"schedule":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Ein tests that ein constraint works correctly on fields inside
// nested structs accessed via dive. Validates U.S. EIN format (XX-XXXXXXX).
func TestDive_Ein(t *testing.T) {
	type Item struct {
		EIN string `json:"ein" pedantigo:"ein"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_standard", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ein":"12-3456789"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_all_zeros", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ein":"00-0000000"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_all_nines", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ein":"99-9999999"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_missing_dash", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ein":"123456789"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].EIN", ve.Errors[0].Field)
	})

	t.Run("invalid_wrong_format", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"ein":"123-456789"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"ein":"12-3456789"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"ein":"invalid"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Issn tests that issn constraint works correctly on fields inside
// nested structs accessed via dive. Validates ISSN format (ISO 3297).
func TestDive_Issn(t *testing.T) {
	type Item struct {
		ISSN string `json:"issn" pedantigo:"issn"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_with_dash", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"issn":"0378-5955"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_no_dash", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"issn":"03785955"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_alternative", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"issn":"2049-3630"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_bad_checksum", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"issn":"0378-5956"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].ISSN", ve.Errors[0].Field)
	})

	t.Run("invalid_too_short", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"issn":"1234"}]}`))
		require.Error(t, err)
	})

	t.Run("empty_slice_passes", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"issn":"0317-8471"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"issn":"1234567890"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Min tests min constraint in nested structs via dive.
func TestDive_Min(t *testing.T) {
	type Item struct {
		Value int `json:"value" pedantigo:"min=5"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":10}]}`))
		require.NoError(t, err)
	})

	t.Run("boundary_valid", func(t *testing.T) {
		// min=5 means >= 5, so exactly 5 should pass
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":5}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":3}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":10}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":2}}}`))
		require.Error(t, err)
	})
}

// TestDive_Max tests max constraint in nested structs via dive.
func TestDive_Max(t *testing.T) {
	type Item struct {
		Value int `json:"value" pedantigo:"max=100"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":50}]}`))
		require.NoError(t, err)
	})

	t.Run("boundary_valid", func(t *testing.T) {
		// max=100 means <= 100, so exactly 100 should pass
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":100}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_in_slice", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"value":150}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Value", ve.Errors[0].Field)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":50}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"a":{"value":200}}}`))
		require.Error(t, err)
	})
}

// TestDive_ContainsAny tests containsany constraint in nested structs via dive.
func TestDive_ContainsAny(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"containsany=abc"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_contains_a", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"alpha"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_contains_b", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"beta"}]}`))
		require.NoError(t, err)
	})

	t.Run("valid_contains_c", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"citrus"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_no_match", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"xyz"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"apple"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_ExcludesAll tests excludesall constraint in nested structs via dive.
func TestDive_ExcludesAll(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"excludesall=xyz"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_no_excluded_chars", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello world"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_contains_x", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"extra"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("invalid_contains_y", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"yes"}]}`))
		require.Error(t, err)
	})

	t.Run("invalid_contains_z", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"zoo"}]}`))
		require.Error(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"abc"}}}`))
		require.NoError(t, err)
	})
}

// TestDive_ExcludesRune tests excludesrune constraint in nested structs via dive.
func TestDive_ExcludesRune(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"excludesrune=@"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_no_at_sign", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"hello world"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_contains_at", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"test@example"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"no at here"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"has@sign"}}}`))
		require.Error(t, err)
	})
}

// TestDive_EndsNotWith tests endsnotwith constraint in nested structs via dive.
func TestDive_EndsNotWith(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"endsnotwith=.tmp"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_different_ending", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"file.txt"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_ends_with_tmp", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"backup.tmp"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"document.pdf"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"cache.tmp"}}}`))
		require.Error(t, err)
	})
}

// TestDive_StartsNotWith tests startsnotwith constraint in nested structs via dive.
func TestDive_StartsNotWith(t *testing.T) {
	type Item struct {
		Text string `json:"text" pedantigo:"startsnotwith=_"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_normal_start", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"normal"}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_starts_with_underscore", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"text":"_private"}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Text", ve.Errors[0].Field)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"public"}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"text":"_hidden"}}}`))
		require.Error(t, err)
	})
}

// TestDive_Unique tests unique constraint in nested structs via dive.
func TestDive_Unique(t *testing.T) {
	type Item struct {
		Tags []string `json:"tags" pedantigo:"unique"`
	}
	type Container struct {
		Items []Item `json:"items" pedantigo:"dive"`
	}

	validator := New[Container]()

	t.Run("valid_unique_tags", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tags":["a","b","c"]}]}`))
		require.NoError(t, err)
	})

	t.Run("invalid_duplicate_tags", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tags":["a","b","a"]}]}`))
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "Items[0].Tags", ve.Errors[0].Field)
	})

	t.Run("empty_array_valid", func(t *testing.T) {
		_, err := validator.Unmarshal([]byte(`{"items":[{"tags":[]}]}`))
		require.NoError(t, err)
	})

	t.Run("map_valid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"tags":["x","y","z"]}}}`))
		require.NoError(t, err)
	})

	t.Run("map_invalid", func(t *testing.T) {
		type MapContainer struct {
			Items map[string]Item `json:"items" pedantigo:"dive"`
		}
		mapValidator := New[MapContainer]()
		_, err := mapValidator.Unmarshal([]byte(`{"items":{"key":{"tags":["x","x"]}}}`))
		require.Error(t, err)
	})
}
