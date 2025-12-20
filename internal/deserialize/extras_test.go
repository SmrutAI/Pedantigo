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
	Extras map[string]any `json:"-" pedantigo:"extra_fields"`
}

type NoExtraField struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type WrongType struct {
	Name   string `json:"name"`
	Extras string `json:"-" pedantigo:"extra_fields"` // Wrong type!
}

type MultipleExtraFields struct {
	Name    string         `json:"name"`
	Extras1 map[string]any `json:"-" pedantigo:"extra_fields"`
	Extras2 map[string]any `json:"-" pedantigo:"extra_fields"` // Duplicate!
}

type PointerMapField struct {
	Name   string          `json:"name"`
	Extras *map[string]any `json:"-" pedantigo:"extra_fields"` // Pointer to map - should fail
}

type privateExtraField struct {
	Name   string
	extras map[string]any `json:"-" pedantigo:"extra_fields"` //nolint:unused // private field - ignored
}

type MapStringInterface struct {
	Name   string                 `json:"name"`
	Extras map[string]interface{} `json:"-" pedantigo:"extra_fields"` // interface{} is alias for any
}

type WrongMapKeyType struct {
	Name   string      `json:"name"`
	Extras map[int]any `json:"-" pedantigo:"extra_fields"` // Wrong key type!
}

type WrongMapValueType struct {
	Name   string            `json:"name"`
	Extras map[string]string `json:"-" pedantigo:"extra_fields"` // Wrong value type!
}

// Tests

func TestDetectExtraField_ValidField_ReturnsInfo(t *testing.T) {
	typ := reflect.TypeOf(ValidExtraField{})
	result := DetectExtraField(typ, "pedantigo")

	require.NotNil(t, result, "Should detect extra_fields field")
	assert.Equal(t, 1, result.FieldIndex, "Extra field should be at index 1")
	assert.Equal(t, "Extras", result.FieldName, "Field name should be 'Extras'")
}

func TestDetectExtraField_MapStringInterface_ReturnsInfo(t *testing.T) {
	// interface{} is an alias for any, should be accepted
	typ := reflect.TypeOf(MapStringInterface{})
	result := DetectExtraField(typ, "pedantigo")

	require.NotNil(t, result, "Should detect extra_fields field with map[string]interface{}")
	assert.Equal(t, 1, result.FieldIndex, "Extra field should be at index 1")
	assert.Equal(t, "Extras", result.FieldName, "Field name should be 'Extras'")
}

func TestDetectExtraField_NoExtraField_ReturnsNil(t *testing.T) {
	typ := reflect.TypeOf(NoExtraField{})
	result := DetectExtraField(typ, "pedantigo")

	assert.Nil(t, result, "Should return nil when no extra_fields field exists")
}

func TestDetectExtraField_WrongType_Panics(t *testing.T) {
	typ := reflect.TypeOf(WrongType{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "pedantigo")
		},
		"Should panic when field type is not map[string]any",
	)
}

func TestDetectExtraField_WrongMapKeyType_Panics(t *testing.T) {
	typ := reflect.TypeOf(WrongMapKeyType{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "pedantigo")
		},
		"Should panic when map key type is not string",
	)
}

func TestDetectExtraField_WrongMapValueType_Panics(t *testing.T) {
	typ := reflect.TypeOf(WrongMapValueType{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "pedantigo")
		},
		"Should panic when map value type is not any/interface{}",
	)
}

func TestDetectExtraField_MultipleExtraFields_Panics(t *testing.T) {
	typ := reflect.TypeOf(MultipleExtraFields{})

	require.PanicsWithValue(t,
		"multiple fields tagged with pedantigo:\"extra_fields\" found: only one is allowed",
		func() {
			DetectExtraField(typ, "pedantigo")
		},
		"Should panic when multiple extra_fields tags exist",
	)
}

func TestDetectExtraField_PointerToMapStringAny_Panics(t *testing.T) {
	typ := reflect.TypeOf(PointerMapField{})

	require.PanicsWithValue(t,
		"field 'Extras' tagged with pedantigo:\"extra_fields\" must be of type map[string]any",
		func() {
			DetectExtraField(typ, "pedantigo")
		},
		"Should panic when field is pointer to map[string]any",
	)
}

func TestDetectExtraField_PrivateField_Ignored(t *testing.T) {
	// Private fields should be ignored even if they have the tag
	typ := reflect.TypeOf(privateExtraField{})
	result := DetectExtraField(typ, "pedantigo")

	assert.Nil(t, result, "Should ignore private fields with extra_fields tag")
}

func TestDetectExtraField_DifferentTagName_ReturnsNil(t *testing.T) {
	// When using a different tag name, should not detect the field
	typ := reflect.TypeOf(ValidExtraField{})
	result := DetectExtraField(typ, "validate") // Different tag name

	assert.Nil(t, result, "Should return nil when using different tag name")
}

func TestDetectExtraField_EmptyTagValue_ReturnsNil(t *testing.T) {
	type EmptyTag struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" pedantigo:""` // Empty tag value
	}

	typ := reflect.TypeOf(EmptyTag{})
	result := DetectExtraField(typ, "pedantigo")

	assert.Nil(t, result, "Should return nil when tag value is empty")
}

func TestDetectExtraField_WrongTagValue_ReturnsNil(t *testing.T) {
	type WrongTagValue struct {
		Name   string         `json:"name"`
		Extras map[string]any `json:"-" pedantigo:"something_else"` // Wrong tag value
	}

	typ := reflect.TypeOf(WrongTagValue{})
	result := DetectExtraField(typ, "pedantigo")

	assert.Nil(t, result, "Should return nil when tag value is not 'extra_fields'")
}
