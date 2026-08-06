package deserialize

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for detection tests

type ValidExtraField struct {
	Name   string         `json:"name"`
	Extras map[string]any `json:"-" validate:"extra_fields"`
}

type NoExtraField struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type WrongType struct {
	Name   string `json:"name"`
	Extras string `json:"-" validate:"extra_fields"` // Wrong type!
}

type MultipleExtraFields struct {
	Name    string         `json:"name"`
	Extras1 map[string]any `json:"-" validate:"extra_fields"`
	Extras2 map[string]any `json:"-" validate:"extra_fields"` // Duplicate!
}

type PointerMapField struct {
	Name   string          `json:"name"`
	Extras *map[string]any `json:"-" validate:"extra_fields"` // Pointer to map - should fail
}

type privateExtraField struct {
	Name   string
	extras map[string]any `json:"-" validate:"extra_fields"` //nolint:unused // private field - ignored
}

type MapStringInterface struct {
	Name   string                 `json:"name"`
	Extras map[string]interface{} `json:"-" validate:"extra_fields"` // interface{} is alias for any
}

type WrongMapKeyType struct {
	Name   string      `json:"name"`
	Extras map[int]any `json:"-" validate:"extra_fields"` // Wrong key type!
}

type WrongMapValueType struct {
	Name   string            `json:"name"`
	Extras map[string]string `json:"-" validate:"extra_fields"` // Wrong value type!
}

// Tests

func TestDetectExtraField_ValidField_ReturnsInfo(t *testing.T) {
	typ := reflect.TypeOf(ValidExtraField{})
	result := DetectExtraField(typ, "validate")

	require.NotNil(t, result, "Should detect extra_fields field")
	assert.Equal(t, 1, result.FieldIndex, "Extra field should be at index 1")
	assert.Equal(t, "Extras", result.FieldName, "Field name should be 'Extras'")
}

func TestDetectExtraField_MapStringInterface_ReturnsInfo(t *testing.T) {
	// interface{} is an alias for any, should be accepted
	typ := reflect.TypeOf(MapStringInterface{})
	result := DetectExtraField(typ, "validate")

	require.NotNil(t, result, "Should detect extra_fields field with map[string]interface{}")
	assert.Equal(t, 1, result.FieldIndex, "Extra field should be at index 1")
	assert.Equal(t, "Extras", result.FieldName, "Field name should be 'Extras'")
}

func TestDetectExtraField_NoExtraField_ReturnsNil(t *testing.T) {
	typ := reflect.TypeOf(NoExtraField{})
	result := DetectExtraField(typ, "validate")

	assert.Nil(t, result, "Should return nil when no extra_fields field exists")
}

func TestDetectExtraField_WrongType_Panics(t *testing.T) {
	typ := reflect.TypeOf(WrongType{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "validate")
		},
		"Should panic when field type is not map[string]any",
	)
}

func TestDetectExtraField_WrongMapKeyType_Panics(t *testing.T) {
	typ := reflect.TypeOf(WrongMapKeyType{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "validate")
		},
		"Should panic when map key type is not string",
	)
}

func TestDetectExtraField_WrongMapValueType_Panics(t *testing.T) {
	typ := reflect.TypeOf(WrongMapValueType{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "validate")
		},
		"Should panic when map value type is not any/interface{}",
	)
}

func TestDetectExtraField_MultipleExtraFields_Panics(t *testing.T) {
	typ := reflect.TypeOf(MultipleExtraFields{})

	require.PanicsWithValue(t,
		"multiple fields tagged with pedantigo:\"extra_fields\" found: only one is allowed",
		func() {
			DetectExtraField(typ, "validate")
		},
		"Should panic when multiple extra_fields tags exist",
	)
}

func TestDetectExtraField_PointerToMapStringAny_Panics(t *testing.T) {
	typ := reflect.TypeOf(PointerMapField{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "validate")
		},
		"Should panic when field is pointer to map[string]any",
	)
}

func TestDetectExtraField_PrivateField_Ignored(t *testing.T) {
	// Private fields should be ignored even if they have the tag
	typ := reflect.TypeOf(privateExtraField{})
	result := DetectExtraField(typ, "validate")

	assert.Nil(t, result, "Should ignore private fields with extra_fields tag")
}

func TestDetectExtraField_DifferentTagName_ReturnsNil(t *testing.T) {
	// Field is tagged with "binding", but we look up using "validate" - should not match.
	type BindingTaggedField struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" binding:"extra_fields"`
	}
	typ := reflect.TypeOf(BindingTaggedField{})
	result := DetectExtraField(typ, "validate") // Different tag name than the struct uses

	assert.Nil(t, result, "Should return nil when using different tag name")
}

func TestDetectExtraField_EmptyTagValue_ReturnsNil(t *testing.T) {
	type EmptyTag struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:""` // Empty tag value
	}

	typ := reflect.TypeOf(EmptyTag{})
	result := DetectExtraField(typ, "validate")

	assert.Nil(t, result, "Should return nil when tag value is empty")
}

func TestDetectExtraField_WrongTagValue_ReturnsNil(t *testing.T) {
	type WrongTagValue struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:"something_else"` // Wrong tag value
	}

	typ := reflect.TypeOf(WrongTagValue{})
	result := DetectExtraField(typ, "validate")

	assert.Nil(t, result, "Should return nil when tag value is not 'extra_fields'")
}

// TestDetectExtraField_PointerType tests DetectExtraField dereferences pointer types.
func TestDetectExtraField_PointerType(t *testing.T) {
	type HasExtra struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" validate:"extra_fields"`
	}

	// Pass pointer to struct type
	typ := reflect.TypeOf(&HasExtra{})
	result := DetectExtraField(typ, "validate")

	assert.NotNil(t, result, "Should handle pointer types by dereferencing")
	assert.Equal(t, 1, result.FieldIndex)
}

// TestDetectExtraField_NonStructType tests DetectExtraField returns nil for non-struct types.
func TestDetectExtraField_NonStructType(t *testing.T) {
	// Test with string type
	result := DetectExtraField(reflect.TypeOf(""), "validate")
	assert.Nil(t, result, "Should return nil for string type")

	// Test with int type
	result = DetectExtraField(reflect.TypeOf(0), "validate")
	assert.Nil(t, result, "Should return nil for int type")

	// Test with slice type
	result = DetectExtraField(reflect.TypeOf([]string{}), "validate")
	assert.Nil(t, result, "Should return nil for slice type")
}
