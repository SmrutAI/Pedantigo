package pedantigo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================================================
// Test Structs for ExtraAllow Mode
// ==================================================

// UserWithExtras is a top-level struct with extra_fields tag.
type UserWithExtras struct {
	Name   string         `json:"name" pedantigo:"required"`
	Age    int            `json:"age"`
	Extras map[string]any `json:"-" pedantigo:"extra_fields"`
}

// NestedWithExtra is a nested struct with extras.
type NestedWithExtra struct {
	Value  string         `json:"value"`
	Extras map[string]any `json:"-" pedantigo:"extra_fields"`
}

// NestedNoExtra is a nested struct without extras.
type NestedNoExtra struct {
	Value string `json:"value"`
}

// TopOnlyExtras is where top-level only has extras.
type TopOnlyExtras struct {
	Name   string         `json:"name"`
	Nested NestedNoExtra  `json:"nested"`
	Extras map[string]any `json:"-" pedantigo:"extra_fields"`
}

// TopNoExtras is where nested only has extras.
type TopNoExtras struct {
	Name   string          `json:"name"`
	Nested NestedWithExtra `json:"nested"`
}

// BothHaveExtras is where both top-level and nested have extras.
type BothHaveExtras struct {
	Name   string          `json:"name"`
	Nested NestedWithExtra `json:"nested"`
	Extras map[string]any  `json:"-" pedantigo:"extra_fields"`
}

// NoExtraFieldStruct is missing extra_fields field (for panic test).
type NoExtraFieldStruct struct {
	Name string `json:"name"`
}

// MultipleNestedExtras has multiple nested structs with extras.
type MultipleNestedExtras struct {
	Name    string          `json:"name"`
	Primary NestedWithExtra `json:"primary"`
	Backup  NestedWithExtra `json:"backup"`
	Extras  map[string]any  `json:"-" pedantigo:"extra_fields"`
}

// SliceWithExtras is a struct for slice testing.
type SliceWithExtras struct {
	Items  []NestedWithExtra `json:"items"`
	Extras map[string]any    `json:"-" pedantigo:"extra_fields"`
}

// DeepNestingLevel3 represents three levels of nesting (level 3).
type DeepNestingLevel3 struct {
	Data   string         `json:"data"`
	Extras map[string]any `json:"-" pedantigo:"extra_fields"`
}

type DeepNestingLevel2 struct {
	Info   string            `json:"info"`
	Level3 DeepNestingLevel3 `json:"level3"`
	Extras map[string]any    `json:"-" pedantigo:"extra_fields"`
}

type DeepNestingLevel1 struct {
	Title  string            `json:"title"`
	Level2 DeepNestingLevel2 `json:"level2"`
	Extras map[string]any    `json:"-" pedantigo:"extra_fields"`
}

// ==================================================
// Creation Tests
// ==================================================

// TestNew_ExtraAllow_WithField_Success tests that New() succeeds when ExtraAllow
// is set and the struct has a field with pedantigo:"extra_fields" tag.
func TestNew_ExtraAllow_WithField_Success(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})
	require.NotNil(t, validator)
}

// TestNew_ExtraAllow_NoField_Panics tests that New() panics when ExtraAllow
// is set but the struct doesn't have an extra_fields field.
func TestNew_ExtraAllow_NoField_Panics(t *testing.T) {
	require.Panics(t, func() {
		New[NoExtraFieldStruct](ValidatorOptions{
			ExtraFields: ExtraAllow,
		})
	}, "expected panic when ExtraAllow is set without extra_fields field")
}

// TestNew_ExtraIgnore_Unchanged tests that ExtraIgnore mode works as before.
func TestNew_ExtraIgnore_Unchanged(t *testing.T) {
	// Should work fine without extra_fields field
	validator := New[NoExtraFieldStruct](ValidatorOptions{
		ExtraFields: ExtraIgnore,
	})
	require.NotNil(t, validator)
}

// TestNew_ExtraForbid_Unchanged tests that ExtraForbid mode works as before.
func TestNew_ExtraForbid_Unchanged(t *testing.T) {
	// Should work fine without extra_fields field
	validator := New[NoExtraFieldStruct](ValidatorOptions{
		ExtraFields: ExtraForbid,
	})
	require.NotNil(t, validator)
}

// ==================================================
// Unmarshal Tests
// ==================================================

// TestUnmarshal_ExtraAllow_CapturesExtras tests that ExtraAllow captures unknown fields.
func TestUnmarshal_ExtraAllow_CapturesExtras(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "John",
		"age": 30,
		"unknown1": "value1",
		"unknown2": 123,
		"unknown3": true
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 30, user.Age)

	require.NotNil(t, user.Extras)
	assert.Equal(t, "value1", user.Extras["unknown1"])
	assert.InEpsilon(t, float64(123), user.Extras["unknown2"], 0.01) // JSON numbers are float64
	assert.Equal(t, true, user.Extras["unknown3"])
}

// TestUnmarshal_ExtraAllow_NoExtras_EmptyMap tests that when there are no extra fields,
// the Extras map is still initialized (empty, not nil).
func TestUnmarshal_ExtraAllow_NoExtras_EmptyMap(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "Jane",
		"age": 25
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, "Jane", user.Name)
	assert.Equal(t, 25, user.Age)
	require.NotNil(t, user.Extras)
	assert.Empty(t, user.Extras)
}

// TestUnmarshal_ExtraAllow_NestedMapValues tests capturing nested objects as map[string]any.
func TestUnmarshal_ExtraAllow_NestedMapValues(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "Bob",
		"age": 40,
		"metadata": {
			"nested_key": "nested_value",
			"nested_num": 456
		}
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Extras)
	metadata, ok := user.Extras["metadata"].(map[string]any)
	require.True(t, ok, "expected metadata to be map[string]any")
	assert.Equal(t, "nested_value", metadata["nested_key"])
	assert.InEpsilon(t, float64(456), metadata["nested_num"], 0.01)
}

// TestUnmarshal_ExtraAllow_ArrayValues tests capturing array values in extras.
func TestUnmarshal_ExtraAllow_ArrayValues(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "Alice",
		"age": 28,
		"tags": ["tag1", "tag2", "tag3"]
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Extras)
	tags, ok := user.Extras["tags"].([]any)
	require.True(t, ok, "expected tags to be []any")
	assert.Len(t, tags, 3)
	assert.Equal(t, "tag1", tags[0])
	assert.Equal(t, "tag2", tags[1])
	assert.Equal(t, "tag3", tags[2])
}

// TestUnmarshal_ExtraAllow_NilValues tests capturing explicit null values.
func TestUnmarshal_ExtraAllow_NilValues(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "Charlie",
		"age": 35,
		"nullable_field": null
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Extras)
	val, exists := user.Extras["nullable_field"]
	assert.True(t, exists, "expected nullable_field to exist in extras")
	assert.Nil(t, val)
}

// ==================================================
// Marshal/Dict Tests
// ==================================================

// TestMarshal_ExtraAllow_IncludesExtras tests that Marshal includes extra fields.
func TestMarshal_ExtraAllow_IncludesExtras(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	user := &UserWithExtras{
		Name: "David",
		Age:  45,
		Extras: map[string]any{
			"custom1": "value1",
			"custom2": 789,
		},
	}

	jsonData, err := validator.Marshal(user)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	assert.Equal(t, "David", result["name"])
	assert.InEpsilon(t, float64(45), result["age"], 0.01)
	assert.Equal(t, "value1", result["custom1"])
	assert.InEpsilon(t, float64(789), result["custom2"], 0.01)
}

// TestDict_ExtraAllow_IncludesExtras tests that Dict includes extra fields.
func TestDict_ExtraAllow_IncludesExtras(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	user := &UserWithExtras{
		Name: "Eve",
		Age:  50,
		Extras: map[string]any{
			"metadata": map[string]any{
				"source": "import",
			},
		},
	}

	dict, err := validator.Dict(user)
	require.NoError(t, err)
	require.NotNil(t, dict)

	assert.Equal(t, "Eve", dict["name"])
	assert.InEpsilon(t, float64(50), dict["age"], 0.01) // JSON numbers are float64

	metadata, ok := dict["metadata"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "import", metadata["source"])
}

// TestExtras_DontOverrideStructFields tests that extras don't override defined struct fields.
func TestExtras_DontOverrideStructFields(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	// Create struct with Extras containing keys that match struct fields
	user := &UserWithExtras{
		Name: "Frank",
		Age:  55,
		Extras: map[string]any{
			"name": "ShouldNotOverride", // Should NOT override the Name field
			"age":  999,                 // Should NOT override the Age field
		},
	}

	jsonData, err := validator.Marshal(user)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	// Struct fields should take precedence
	assert.Equal(t, "Frank", result["name"])
	assert.InEpsilon(t, float64(55), result["age"], 0.01)
}

// ==================================================
// Round-Trip Tests (CRITICAL)
// ==================================================

// TestRoundTrip_TopLevelOnly_ExtrasPreserved tests round-trip with top-level extras only.
func TestRoundTrip_TopLevelOnly_ExtrasPreserved(t *testing.T) {
	validator := New[TopOnlyExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"name": "Test",
		"nested": {"value": "nested_val"},
		"extra_top": "top_extra_value",
		"extra_num": 42
	}`)

	// Unmarshal
	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)

	// Marshal back
	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	// Parse both JSONs to compare
	var original, roundTrip map[string]any
	require.NoError(t, json.Unmarshal(originalJSON, &original))
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	assert.Equal(t, original["name"], roundTrip["name"])
	assert.Equal(t, original["extra_top"], roundTrip["extra_top"])
	assert.Equal(t, original["extra_num"], roundTrip["extra_num"])
}

// TestRoundTrip_NestedOnly_ExtrasPreserved tests round-trip with only nested extras populated.
// Note: Uses BothHaveExtras because ExtraAllow requires top-level to have extra_fields field.
// We only populate nested extras to test that nested extras work independently.
func TestRoundTrip_NestedOnly_ExtrasPreserved(t *testing.T) {
	validator := New[BothHaveExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	// Only nested has extra fields, top level has none
	originalJSON := []byte(`{
		"name": "TestNested",
		"nested": {
			"value": "val",
			"extra_nested": "nested_extra_value"
		}
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)

	// Verify nested extras were captured
	require.NotNil(t, obj.Nested.Extras)
	assert.Equal(t, "nested_extra_value", obj.Nested.Extras["extra_nested"])

	// Top-level extras should be empty (but not nil)
	require.NotNil(t, obj.Extras)
	assert.Empty(t, obj.Extras)

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var original, roundTrip map[string]any
	require.NoError(t, json.Unmarshal(originalJSON, &original))
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	assert.Equal(t, original["name"], roundTrip["name"])

	originalNested := original["nested"].(map[string]any)
	roundTripNested := roundTrip["nested"].(map[string]any)
	assert.Equal(t, originalNested["value"], roundTripNested["value"])
	assert.Equal(t, originalNested["extra_nested"], roundTripNested["extra_nested"])
}

// TestRoundTrip_TopAndNested_BothExtrasPreserved tests round-trip with both top and nested extras.
func TestRoundTrip_TopAndNested_BothExtrasPreserved(t *testing.T) {
	validator := New[BothHaveExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"name": "Both",
		"nested": {
			"value": "nested_val",
			"extra_nested": "nested_extra"
		},
		"extra_top": "top_extra"
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var original, roundTrip map[string]any
	require.NoError(t, json.Unmarshal(originalJSON, &original))
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	// Top-level extras
	assert.Equal(t, original["extra_top"], roundTrip["extra_top"])

	// Nested extras
	originalNested := original["nested"].(map[string]any)
	roundTripNested := roundTrip["nested"].(map[string]any)
	assert.Equal(t, originalNested["extra_nested"], roundTripNested["extra_nested"])
}

// TestRoundTrip_TopHasExtras_NestedDoesNot_TopPreserved tests top extras preserved,
// nested has no extras.
func TestRoundTrip_TopHasExtras_NestedDoesNot_TopPreserved(t *testing.T) {
	validator := New[TopOnlyExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"name": "TopOnly",
		"nested": {"value": "nested_val"},
		"extra_top": "should_be_preserved"
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)
	require.Equal(t, "should_be_preserved", obj.Extras["extra_top"])

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))
	assert.Equal(t, "should_be_preserved", roundTrip["extra_top"])
}

// TestRoundTrip_NestedHasExtras_TopDoesNot_NestedPreserved tests nested extras preserved,
// top has no extras populated (but struct has extra_fields to satisfy ExtraAllow requirement).
func TestRoundTrip_NestedHasExtras_TopDoesNot_NestedPreserved(t *testing.T) {
	// Note: Uses BothHaveExtras because ExtraAllow requires top-level to have extra_fields field
	validator := New[BothHaveExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"name": "NestedOnly",
		"nested": {
			"value": "val",
			"extra_nested": "nested_should_persist"
		}
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)
	require.Equal(t, "nested_should_persist", obj.Nested.Extras["extra_nested"])

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	nestedMap := roundTrip["nested"].(map[string]any)
	assert.Equal(t, "nested_should_persist", nestedMap["extra_nested"])
}

// TestRoundTrip_MultipleNestedStructs_AllExtrasPreserved tests multiple nested structs
// each with their own extras.
//
//nolint:dupl // Intentionally similar pattern to other round-trip tests
func TestRoundTrip_MultipleNestedStructs_AllExtrasPreserved(t *testing.T) {
	validator := New[MultipleNestedExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"name": "Multi",
		"primary": {
			"value": "primary_val",
			"extra_primary": "primary_extra"
		},
		"backup": {
			"value": "backup_val",
			"extra_backup": "backup_extra"
		},
		"extra_top": "top_extra"
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var original, roundTrip map[string]any
	require.NoError(t, json.Unmarshal(originalJSON, &original))
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	// Top extras
	assert.Equal(t, original["extra_top"], roundTrip["extra_top"])

	// Primary extras
	originalPrimary := original["primary"].(map[string]any)
	roundTripPrimary := roundTrip["primary"].(map[string]any)
	assert.Equal(t, originalPrimary["extra_primary"], roundTripPrimary["extra_primary"])

	// Backup extras
	originalBackup := original["backup"].(map[string]any)
	roundTripBackup := roundTrip["backup"].(map[string]any)
	assert.Equal(t, originalBackup["extra_backup"], roundTripBackup["extra_backup"])
}

// TestRoundTrip_SliceOfStructsWithExtras_AllPreserved tests slice of structs with extras.
func TestRoundTrip_SliceOfStructsWithExtras_AllPreserved(t *testing.T) {
	validator := New[SliceWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"items": [
			{"value": "item1", "extra1": "extra_val1"},
			{"value": "item2", "extra2": "extra_val2"}
		],
		"extra_top": "top_level_extra"
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var original, roundTrip map[string]any
	require.NoError(t, json.Unmarshal(originalJSON, &original))
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	// Top-level extras
	assert.Equal(t, original["extra_top"], roundTrip["extra_top"])

	// Slice item extras
	originalItems := original["items"].([]any)
	roundTripItems := roundTrip["items"].([]any)

	item1Orig := originalItems[0].(map[string]any)
	item1Round := roundTripItems[0].(map[string]any)
	assert.Equal(t, item1Orig["extra1"], item1Round["extra1"])

	item2Orig := originalItems[1].(map[string]any)
	item2Round := roundTripItems[1].(map[string]any)
	assert.Equal(t, item2Orig["extra2"], item2Round["extra2"])
}

// TestRoundTrip_DeepNesting_ThreeLevels_AllExtrasPreserved tests three levels of nesting.
//
//nolint:dupl // Intentionally similar pattern to other round-trip tests
func TestRoundTrip_DeepNesting_ThreeLevels_AllExtrasPreserved(t *testing.T) {
	validator := New[DeepNestingLevel1](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	originalJSON := []byte(`{
		"title": "Level1",
		"level2": {
			"info": "Level2",
			"level3": {
				"data": "Level3",
				"extra_l3": "level3_extra"
			},
			"extra_l2": "level2_extra"
		},
		"extra_l1": "level1_extra"
	}`)

	obj, err := validator.Unmarshal(originalJSON)
	require.NoError(t, err)

	roundTripJSON, err := validator.Marshal(obj)
	require.NoError(t, err)

	var original, roundTrip map[string]any
	require.NoError(t, json.Unmarshal(originalJSON, &original))
	require.NoError(t, json.Unmarshal(roundTripJSON, &roundTrip))

	// Level 1 extras
	assert.Equal(t, original["extra_l1"], roundTrip["extra_l1"])

	// Level 2 extras
	originalL2 := original["level2"].(map[string]any)
	roundTripL2 := roundTrip["level2"].(map[string]any)
	assert.Equal(t, originalL2["extra_l2"], roundTripL2["extra_l2"])

	// Level 3 extras
	originalL3 := originalL2["level3"].(map[string]any)
	roundTripL3 := roundTripL2["level3"].(map[string]any)
	assert.Equal(t, originalL3["extra_l3"], roundTripL3["extra_l3"])
}

// ==================================================
// Edge Cases
// ==================================================

// TestUnmarshal_ExtraAllow_UnicodeKeys tests Unicode keys in extras.
func TestUnmarshal_ExtraAllow_UnicodeKeys(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "Test",
		"age": 30,
		"日本語キー": "unicode value",
		"émoji": "🚀"
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Extras)
	assert.Equal(t, "unicode value", user.Extras["日本語キー"])
	assert.Equal(t, "🚀", user.Extras["émoji"])
}

// TestUnmarshal_ExtraAllow_LargeNumbers tests large numbers in extras.
func TestUnmarshal_ExtraAllow_LargeNumbers(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "BigNum",
		"age": 30,
		"big_int": 9007199254740991,
		"big_float": 1.7976931348623157e+308
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Extras)
	assert.InEpsilon(t, float64(9007199254740991), user.Extras["big_int"], 0.01)
	assert.InEpsilon(t, 1.7976931348623157e+308, user.Extras["big_float"], 0.01)
}

// TestUnmarshal_ExtraAllow_EmptyExtrasMap tests explicitly setting Extras to empty map
// doesn't cause issues.
func TestUnmarshal_ExtraAllow_EmptyExtrasMap(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	user := &UserWithExtras{
		Name:   "Empty",
		Age:    30,
		Extras: map[string]any{},
	}

	jsonData, err := validator.Marshal(user)
	require.NoError(t, err)

	// Round-trip
	result, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "Empty", result.Name)
	assert.Equal(t, 30, result.Age)
	require.NotNil(t, result.Extras)
	assert.Empty(t, result.Extras)
}

// TestUnmarshal_ExtraAllow_NullExtraValue tests null values in extras are captured correctly.
func TestUnmarshal_ExtraAllow_NullExtraValue(t *testing.T) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "NullTest",
		"age": 30,
		"null_field": null,
		"non_null_field": "value"
	}`)

	user, err := validator.Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, user.Extras)
	assert.Contains(t, user.Extras, "null_field")
	assert.Nil(t, user.Extras["null_field"])
	assert.Equal(t, "value", user.Extras["non_null_field"])
}

// ==================================================
// Benchmarks
// ==================================================

// BenchmarkUnmarshal_ExtraIgnore_Baseline establishes baseline performance with ExtraIgnore.
func BenchmarkUnmarshal_ExtraIgnore_Baseline(b *testing.B) {
	type SimpleUser struct {
		Name string `json:"name" pedantigo:"required"`
		Age  int    `json:"age"`
	}

	validator := New[SimpleUser](ValidatorOptions{
		ExtraFields: ExtraIgnore,
	})

	jsonData := []byte(`{
		"name": "Benchmark",
		"age": 30,
		"unknown1": "value1",
		"unknown2": 123
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnmarshal_ExtraAllow_Enabled benchmarks ExtraAllow with capturing extras.
func BenchmarkUnmarshal_ExtraAllow_Enabled(b *testing.B) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "Benchmark",
		"age": 30,
		"unknown1": "value1",
		"unknown2": 123
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnmarshal_ExtraAllow_ZeroOverhead benchmarks ExtraAllow with no extra fields.
// Should have minimal overhead compared to ExtraIgnore baseline.
func BenchmarkUnmarshal_ExtraAllow_ZeroOverhead(b *testing.B) {
	validator := New[UserWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	// JSON with no extra fields
	jsonData := []byte(`{
		"name": "Benchmark",
		"age": 30
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
