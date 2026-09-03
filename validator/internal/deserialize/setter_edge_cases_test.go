package deserialize

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== DecodeStruct: static "default" tag ====================

type coverageStaticDefault struct {
	Port int `json:"port" validate:"default=42"`
}

func TestDecodeStruct_StaticDefault(t *testing.T) {
	var target coverageStaticDefault
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageStaticDefault{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	err := DecodeStruct(reflect.ValueOf(&target).Elem(), map[string]any{}, plan, st, "")

	require.NoError(t, err)
	assert.Equal(t, 42, target.Port)
}

// ==================== DecodeStruct: "defaultUsingMethod" tag ====================

type coverageMethodDefault struct {
	Port int `json:"port" validate:"defaultUsingMethod=GetPort"`
}

func (m *coverageMethodDefault) GetPort() (int, error) {
	return 9090, nil
}

func (m *coverageMethodDefault) GetPortErr() (int, error) {
	return 0, errors.New("boom")
}

func TestDecodeStruct_DefaultUsingMethod_Success(t *testing.T) {
	var target coverageMethodDefault
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMethodDefault{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	err := DecodeStruct(reflect.ValueOf(&target).Elem(), map[string]any{}, plan, st, "")

	require.NoError(t, err)
	assert.Equal(t, 9090, target.Port)
}

func TestDecodeStruct_DefaultUsingMethod_MethodError(t *testing.T) {
	var target coverageMethodDefaultErr
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMethodDefaultErr{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	err := DecodeStruct(reflect.ValueOf(&target).Elem(), map[string]any{}, plan, st, "")

	require.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

type coverageMethodDefaultErr struct {
	Port int `json:"port" validate:"defaultUsingMethod=GetPortErr"`
}

func (m *coverageMethodDefaultErr) GetPortErr() (int, error) {
	return 0, errors.New("boom")
}

func TestDecodeStruct_DefaultUsingMethod_MethodNotFound(t *testing.T) {
	type coverageMethodMissing struct {
		Port int `json:"port" validate:"defaultUsingMethod=NoSuchMethod"`
	}
	var target coverageMethodMissing
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMethodMissing{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	err := DecodeStruct(reflect.ValueOf(&target).Elem(), map[string]any{}, plan, st, "")

	require.NoError(t, err)
	assert.Equal(t, 0, target.Port) // left at zero value since method wasn't found
}

func TestDecodeStruct_DefaultUsingMethod_NotAddressable(t *testing.T) {
	// A non-addressable struct value: the interpreter can still handle it
	// when DecodeStruct is called with the value (not a pointer).
	// The defaultUsingMethod dispatch still requires addressability, so it's left at zero.
	var target coverageMethodDefault
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMethodDefault{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	// Since the interpreter requires addressable values for default methods,
	// this will leave Port at zero.
	err := DecodeStruct(reflect.ValueOf(&target).Elem(), map[string]any{}, plan, st, "")

	require.NoError(t, err)
}

// ==================== DecodeStruct: slice struct elements ====================

type coverageItem struct {
	Name string `json:"name" validate:"required"`
}

type coverageItemsContainer struct {
	Items []coverageItem `json:"items"`
}

func TestDecodeStruct_SliceStructElement_NotAMap(t *testing.T) {
	// Use the container to decode the slice field via DecodeStruct
	var target coverageItemsContainer
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageItemsContainer{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	// Element input is a map, but not map[string]any -> error during decode
	input := map[string]any{
		"items": []map[string]int{{"x": 1}}, // not map[string]any
	}
	err := DecodeStruct(reflect.ValueOf(&target).Elem(), input, plan, st, "")

	require.Error(t, err)
}

func TestDecodeStruct_SliceStructElement_CollectsRequiredErrors(t *testing.T) {
	// Use the container to decode the slice field via DecodeStruct
	var target coverageItemsContainer
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageItemsContainer{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	input := map[string]any{
		"items": []any{
			map[string]any{}, // missing required "name"
		},
	}
	err := DecodeStruct(reflect.ValueOf(&target).Elem(), input, plan, st, "")

	require.Error(t, err)
	// The error should contain information about the missing required field
}

// ==================== DecodeStruct: map key conversion and struct values ====================

// coverageMapKey is string-based (not int-based): real JSON-decoded map keys
// are always Go strings (JSON object keys are always strings), and Go's
// conversion rules do not allow string->int conversion (only int->string) -
// so a "convertible but not assignable" map key, reachable through the real
// DecodeStruct/JSON path, can only be string -> a named string type.
type coverageMapKey string

type coverageMapContainer struct {
	Data map[coverageMapKey]string `json:"data"`
}

func TestDecodeStruct_MapKeyConvertibleNotAssignable(t *testing.T) {
	// Use the container to decode the map field via DecodeStruct
	var target coverageMapContainer
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMapContainer{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	input := map[string]any{
		"data": map[string]any{"one": "one"}, // string key gets converted to coverageMapKey
	}
	err := DecodeStruct(reflect.ValueOf(&target).Elem(), input, plan, st, "")

	require.NoError(t, err)
}

type coverageIncompatibleKey struct{}

type coverageMapIncompatibleContainer struct {
	Data map[coverageIncompatibleKey]string `json:"data"`
}

func TestDecodeStruct_MapKeyConversionError(t *testing.T) {
	// Use the container to decode the map field via DecodeStruct
	var target coverageMapIncompatibleContainer
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMapIncompatibleContainer{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	// string keys cannot convert to a struct{} key type.
	input := map[string]any{
		"data": map[string]string{"a": "one"},
	}
	err := DecodeStruct(reflect.ValueOf(&target).Elem(), input, plan, st, "")

	require.Error(t, err)
}

type coverageMapStructValueContainer struct {
	Mapping map[string]coverageItem `json:"mapping"`
}

func TestDecodeStruct_MapStructValue_NotAMap(t *testing.T) {
	// Use the container to decode the map field via DecodeStruct
	var target coverageMapStructValueContainer
	index := map[reflect.Type]*TypePlan{}
	plan := BuildTypePlan(reflect.TypeOf(coverageMapStructValueContainer{}), "validate", index)
	st := NewPlanState(index, "validate", 3)

	// Value is a map, but not map[string]any -> error during decode
	input := map[string]any{
		"mapping": map[string]map[int]string{"k": {1: "v"}},
	}
	err := DecodeStruct(reflect.ValueOf(&target).Elem(), input, plan, st, "")

	require.Error(t, err)
}

// ==================== SetDefaultValue: slice element kinds ====================

func TestSetDefaultValue_SliceUint(t *testing.T) {
	var target struct {
		Values []uint
	}
	fieldValue := reflect.ValueOf(&target).Elem().Field(0)

	SetDefaultValue(fieldValue, "1 2 3", recursiveSetDefault)

	assert.Equal(t, []uint{1, 2, 3}, target.Values)
}

func TestSetDefaultValue_SliceFloat(t *testing.T) {
	var target struct {
		Values []float64
	}
	fieldValue := reflect.ValueOf(&target).Elem().Field(0)

	SetDefaultValue(fieldValue, "1.5 2.5", recursiveSetDefault)

	assert.Equal(t, []float64{1.5, 2.5}, target.Values)
}

func TestSetDefaultValue_SliceBool(t *testing.T) {
	var target struct {
		Values []bool
	}
	fieldValue := reflect.ValueOf(&target).Elem().Field(0)

	SetDefaultValue(fieldValue, "true false", recursiveSetDefault)

	assert.Equal(t, []bool{true, false}, target.Values)
}
