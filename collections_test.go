package pedantigo

import (
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
