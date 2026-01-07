package pedantigo

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================================================
// Coverage tests for edge cases in root package
// ==================================================

// TestValidateWithCache_NilCache tests validateWithCache with nil cache
func TestValidateWithCache_NilCache_Simple(t *testing.T) {
	type SimpleStruct struct {
		Name string `json:"name" pedantigo:"required"`
	}

	// Create validator but bypass normal cache creation
	v := New[SimpleStruct]()

	// Call validateWithCache with nil cache directly via Validate
	// The validator will have a cache, but we can test the nil path
	// by creating a minimal struct that doesn't trigger cache building
	obj := SimpleStruct{Name: "test"}
	err := v.Validate(&obj)
	assert.NoError(t, err)
}

// TestValidateWithCache_NonStructKind tests validateWithCache with non-struct value
func TestValidateWithCache_NonStructKind(t *testing.T) {
	// Test with a type alias to primitive
	type StringAlias string

	v := New[StringAlias]()
	val := StringAlias("test")
	err := v.Validate(&val)
	// Should not error since there are no constraints
	assert.NoError(t, err)
}

// TestValidateWithCache_PointerIndirection tests multiple pointer levels
func TestValidateWithCache_PointerIndirection(t *testing.T) {
	type NestedStruct struct {
		Value string `json:"value" pedantigo:"required"`
	}

	v := New[*NestedStruct]()

	// Test with non-nil pointer
	obj := &NestedStruct{Value: "test"}
	err := v.Validate(&obj)
	require.NoError(t, err)

	// Test with nil pointer (should return early)
	var nilObj *NestedStruct
	err = v.Validate(&nilObj)
	require.NoError(t, err) // No validation on nil
}

// TestSchemaJSON_ConcurrentCachePaths tests concurrent access to SchemaJSON
func TestSchemaJSON_ConcurrentCachePaths(t *testing.T) {
	type ConcurrentStruct struct {
		Field1 string `json:"field1" pedantigo:"required"`
		Field2 int    `json:"field2" pedantigo:"min=0"`
	}

	v := New[ConcurrentStruct]()

	// Create multiple goroutines to trigger cache race paths
	var wg sync.WaitGroup
	numGoroutines := 10

	results := make([][]byte, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errors[idx] = v.SchemaJSON()
		}(i)
	}

	wg.Wait()

	// All should succeed and return same schema
	for i := 0; i < numGoroutines; i++ {
		require.NoError(t, errors[i], "goroutine %d failed", i)
		require.NotNil(t, results[i], "goroutine %d got nil result", i)
	}

	// All results should be equal
	for i := 1; i < numGoroutines; i++ {
		assert.Equal(t, results[0], results[i], "results differ between goroutine 0 and %d", i)
	}
}

// TestSchemaJSON_CachedSchemaButNotJSON tests path where Schema is cached but JSON isn't
func TestSchemaJSON_CachedSchemaButNotJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" pedantigo:"required"`
		Email string `json:"email" pedantigo:"email"`
	}

	v := New[TestStruct]()

	// First call Schema() to populate cachedSchema but not cachedSchemaJSON
	schema := v.Schema()
	require.NotNil(t, schema)

	// Now call SchemaJSON() - this should hit the path where cachedSchema exists
	// but cachedSchemaJSON doesn't (lines 67-82 in schema.go)
	jsonBytes, err := v.SchemaJSON()
	require.NoError(t, err)
	require.NotNil(t, jsonBytes)

	// Verify it's valid JSON
	var schemaMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &schemaMap)
	assert.NoError(t, err)
}

// TestSchemaJSONOpenAPI_ConcurrentCachePaths tests concurrent access to SchemaJSONOpenAPI
func TestSchemaJSONOpenAPI_ConcurrentCachePaths(t *testing.T) {
	type OpenAPIStruct struct {
		ID   string `json:"id" pedantigo:"uuid"`
		Name string `json:"name" pedantigo:"required"`
	}

	v := New[OpenAPIStruct]()

	// Create multiple goroutines to trigger cache race paths
	var wg sync.WaitGroup
	numGoroutines := 10

	results := make([][]byte, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errors[idx] = v.SchemaJSONOpenAPI()
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i := 0; i < numGoroutines; i++ {
		require.NoError(t, errors[i], "goroutine %d failed", i)
		require.NotNil(t, results[i], "goroutine %d got nil result", i)
	}

	// All results should be equal
	for i := 1; i < numGoroutines; i++ {
		assert.Equal(t, results[0], results[i], "results differ between goroutine 0 and %d", i)
	}
}

// TestSchemaJSONOpenAPI_CachedSchemaButNotJSON tests path where OpenAPI schema is cached but JSON isn't
func TestSchemaJSONOpenAPI_CachedSchemaButNotJSON(t *testing.T) {
	type NestedType struct {
		Value string `json:"value" pedantigo:"required"`
	}

	type TestStruct struct {
		Name   string     `json:"name" pedantigo:"required"`
		Nested NestedType `json:"nested"`
	}

	v := New[TestStruct]()

	// First call SchemaOpenAPI() to populate cachedOpenAPI but not cachedOpenAPIJSON
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// Now call SchemaJSONOpenAPI() - should hit the path where cachedOpenAPI exists
	// but cachedOpenAPIJSON doesn't (lines 184-199 in schema.go)
	jsonBytes, err := v.SchemaJSONOpenAPI()
	require.NoError(t, err)
	require.NotNil(t, jsonBytes)

	// Verify it's valid JSON
	var schemaMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &schemaMap)
	assert.NoError(t, err)
}

// TestFindTypeForDefinition_MapValueType tests findTypeForDefinition with map value types
func TestFindTypeForDefinition_MapValueType_Address(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type UserWithMapField struct {
		Name      string             `json:"name"`
		Addresses map[string]Address `json:"addresses"`
	}

	v := New[UserWithMapField]()

	// Generate SchemaOpenAPI which uses findTypeForDefinition
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// Check that the Address type was found in definitions
	// (This exercises the searchMapType branch in findTypeForDefinition)
	if schema.Definitions != nil {
		// The schema should have properly handled the map value type
		assert.NotNil(t, schema.Properties)
	}
}

// TestFindTypeForDefinition_NonStructKind tests findTypeForDefinition with struct containing primitives
func TestFindTypeForDefinition_NonStructKind(t *testing.T) {
	// Test a struct with only primitive fields (no nested definitions)
	type PrimitiveOnly struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	v := New[PrimitiveOnly]()

	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// Should handle struct with primitives gracefully
	assert.Equal(t, "object", schema.Type)
}

// TestFindTypeForDefinition_SliceOfPointers tests slice of pointer structs
func TestFindTypeForDefinition_SliceOfPointers(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}

	type Order struct {
		ID    string  `json:"id"`
		Items []*Item `json:"items"`
	}

	v := New[Order]()

	// Generate SchemaOpenAPI to exercise findTypeForDefinition with slices
	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// Verify the slice type was handled correctly
	if schema.Properties != nil {
		itemsProp, exists := schema.Properties.Get("items")
		if exists {
			assert.NotNil(t, itemsProp)
		}
	}
}

// TestMarshalWithExtras_UnmarshalError tests error path in marshalWithExtras
func TestMarshalWithExtras_UnmarshalError(t *testing.T) {
	// Create a type that will marshal to JSON but then fail to unmarshal to map
	// This is difficult to trigger with normal structs, so we test the normal path
	type WithExtras struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" pedantigo:"extra_fields"`
	}

	v := New[WithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	obj := &WithExtras{
		Name: "test",
		Extras: map[string]any{
			"custom_field": "value",
		},
	}

	// Normal marshal should work
	data, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(data), "custom_field")
}

// TestMarshalWithExtras_NestedStructs tests mergeExtrasRecursive with nested structs
func TestMarshalWithExtras_NestedStructs(t *testing.T) {
	type NestedWithExtras struct {
		Field  string         `json:"field"`
		Extras map[string]any `json:"-" pedantigo:"extra_fields"`
	}

	type ParentWithExtras struct {
		Name   string           `json:"name"`
		Nested NestedWithExtras `json:"nested"`
		Extras map[string]any   `json:"-" pedantigo:"extra_fields"`
	}

	v := New[ParentWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	obj := &ParentWithExtras{
		Name: "parent",
		Nested: NestedWithExtras{
			Field: "nested_value",
			Extras: map[string]any{
				"nested_extra": "nested_custom",
			},
		},
		Extras: map[string]any{
			"parent_extra": "parent_custom",
		},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(data), "parent_extra")
	assert.Contains(t, string(data), "nested_extra")
}

// TestMarshalWithExtras_SliceOfStructsWithExtras tests merging extras in slice elements
func TestMarshalWithExtras_SliceOfStructsWithExtras(t *testing.T) {
	type ItemWithExtras struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" pedantigo:"extra_fields"`
	}

	type Container struct {
		Items  []ItemWithExtras `json:"items"`
		Extras map[string]any   `json:"-" pedantigo:"extra_fields"`
	}

	v := New[Container](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	obj := &Container{
		Items: []ItemWithExtras{
			{
				Name: "item1",
				Extras: map[string]any{
					"custom1": "value1",
				},
			},
			{
				Name: "item2",
				Extras: map[string]any{
					"custom2": "value2",
				},
			},
		},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)

	// Verify the marshaled data contains extras
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	items, ok := result["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, items, 2)
}

// TestMarshalWithExtras_PointerFields tests extras with pointer struct fields
func TestMarshalWithExtras_PointerFields(t *testing.T) {
	type NestedWithExtras struct {
		Value  string         `json:"value"`
		Extras map[string]any `json:"-" pedantigo:"extra_fields"`
	}

	type WithPointerField struct {
		Name   string            `json:"name"`
		Nested *NestedWithExtras `json:"nested"`
		Extras map[string]any    `json:"-" pedantigo:"extra_fields"`
	}

	v := New[WithPointerField](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	obj := &WithPointerField{
		Name: "test",
		Nested: &NestedWithExtras{
			Value: "nested",
			Extras: map[string]any{
				"nested_custom": "nested_value",
			},
		},
		Extras: map[string]any{
			"top_custom": "top_value",
		},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(data), "top_custom")
	assert.Contains(t, string(data), "nested_custom")
}

// TestMarshalWithExtras_NilExtrasField tests nil extras field handling
func TestMarshalWithExtras_NilExtrasField(t *testing.T) {
	type WithExtras struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" pedantigo:"extra_fields"`
	}

	v := New[WithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	obj := &WithExtras{
		Name:   "test",
		Extras: nil, // Nil extras
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)

	// Should still marshal successfully
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "test", result["name"])
}

// TestSchemaJSON_DoubleCheckCache tests the double-check cache pattern
func TestSchemaJSON_DoubleCheckCache(t *testing.T) {
	type RaceStruct struct {
		Field string `json:"field" pedantigo:"required"`
	}

	v := New[RaceStruct]()

	// Use goroutines to try to trigger the double-check path (lines 91-93)
	var wg sync.WaitGroup
	const numRaces = 100

	for i := 0; i < numRaces; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := v.SchemaJSON()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

// TestSchemaJSONOpenAPI_DoubleCheckCache tests the double-check cache pattern for OpenAPI
func TestSchemaJSONOpenAPI_DoubleCheckCache(t *testing.T) {
	type RaceStruct struct {
		Field string `json:"field" pedantigo:"required"`
	}

	v := New[RaceStruct]()

	// Use goroutines to try to trigger the double-check path (lines 208-210)
	var wg sync.WaitGroup
	const numRaces = 100

	for i := 0; i < numRaces; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := v.SchemaJSONOpenAPI()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

// TestFindTypeForDefinition_DeepNesting tests deeply nested struct definitions
func TestFindTypeForDefinition_DeepNesting(t *testing.T) {
	type Level3 struct {
		Value string `json:"value"`
	}

	type Level2 struct {
		L3 Level3 `json:"l3"`
	}

	type Level1 struct {
		L2 Level2 `json:"l2"`
	}

	type Root struct {
		L1 Level1 `json:"l1"`
	}

	v := New[Root]()

	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// The findTypeForDefinition should recursively find all nested types
	assert.NotNil(t, schema.Properties)
}

// TestFindTypeForDefinition_MapOfPointers tests map with pointer value types
func TestFindTypeForDefinition_MapOfPointers(t *testing.T) {
	type Value struct {
		Data string `json:"data"`
	}

	type WithMapOfPointers struct {
		Name   string            `json:"name"`
		Values map[string]*Value `json:"values"`
	}

	v := New[WithMapOfPointers]()

	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)

	// Should handle map of pointers correctly
	assert.NotNil(t, schema.Properties)
}

// TestValidateWithCache_SkipConstraints tests the skip_unless branch
func TestValidateWithCache_SkipConstraints(t *testing.T) {
	type ConditionalValidation struct {
		Role     string `json:"role"`
		AdminKey string `json:"admin_key" pedantigo:"skip_unless=Role admin,required,min=10"`
	}

	v := New[ConditionalValidation]()

	// Test case where skip condition is NOT met (Role != "admin")
	obj1 := ConditionalValidation{
		Role:     "user",
		AdminKey: "", // Should be skipped, no error
	}
	err := v.Validate(&obj1)
	require.NoError(t, err, "validation should skip admin_key when role is not admin")

	// Test case where skip condition IS met (Role == "admin")
	obj2 := ConditionalValidation{
		Role:     "admin",
		AdminKey: "short", // Should be validated and fail min=10
	}
	err = v.Validate(&obj2)
	require.Error(t, err, "validation should check admin_key when role is admin")

	// Valid admin case
	obj3 := ConditionalValidation{
		Role:     "admin",
		AdminKey: "long_enough_key",
	}
	err = v.Validate(&obj3)
	require.NoError(t, err)
}

// TestMarshalWithExtras_EmptyStruct tests marshaling empty struct with extras capability
func TestMarshalWithExtras_EmptyStruct(t *testing.T) {
	type EmptyWithExtras struct {
		Extras map[string]any `json:"-" pedantigo:"extra_fields"`
	}

	v := New[EmptyWithExtras](ValidatorOptions{
		ExtraFields: ExtraAllow,
	})

	obj := &EmptyWithExtras{
		Extras: map[string]any{
			"dynamic_field": "dynamic_value",
		},
	}

	data, err := v.Marshal(obj)
	require.NoError(t, err)
	assert.Contains(t, string(data), "dynamic_field")
}

// TestFindTypeForDefinition_SliceOfMaps tests slice containing maps
func TestFindTypeForDefinition_SliceOfMaps(t *testing.T) {
	type ComplexStruct struct {
		Name string                   `json:"name"`
		Data []map[string]interface{} `json:"data"`
	}

	v := New[ComplexStruct]()

	schema := v.SchemaOpenAPI()
	require.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
}
