package deserialize

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Nested struct fallback (non-map[string]any input) ====================

type coverageAddr struct {
	City string `json:"city"`
	Age  int    `json:"age"`
}

// TestSetFieldValueWithOptions_NestedStructFallback_Success covers the
// re-marshal/re-unmarshal fallback used when the nested struct's input value
// is a map but not concretely map[string]any (e.g. map[string]string).
func TestSetFieldValueWithOptions_NestedStructFallback_Success(t *testing.T) {
	var target coverageAddr
	fieldValue := reflect.ValueOf(&target).Elem()

	input := map[string]string{"city": "NYC"}
	err := SetFieldValueWithOptions(fieldValue, input, reflect.TypeOf(coverageAddr{}), recursiveSetFuncNoop, FieldOptions{})

	require.NoError(t, err)
	assert.Equal(t, "NYC", target.City)
}

// TestSetFieldValueWithOptions_NestedStructFallback_MarshalError covers the
// json.Marshal failure branch of the fallback path (channels aren't marshalable).
func TestSetFieldValueWithOptions_NestedStructFallback_MarshalError(t *testing.T) {
	var target coverageAddr
	fieldValue := reflect.ValueOf(&target).Elem()

	input := map[string]chan int{"city": make(chan int)}
	err := SetFieldValueWithOptions(fieldValue, input, reflect.TypeOf(coverageAddr{}), recursiveSetFuncNoop, FieldOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal nested struct")
}

// TestSetFieldValueWithOptions_NestedStructFallback_UnmarshalError covers the
// json.Unmarshal failure branch of the fallback path (type mismatch).
func TestSetFieldValueWithOptions_NestedStructFallback_UnmarshalError(t *testing.T) {
	var target coverageAddr
	fieldValue := reflect.ValueOf(&target).Elem()

	// "age" is an int field; a string value fails json.Unmarshal type checking.
	input := map[string]string{"age": "not-a-number"}
	err := SetFieldValueWithOptions(fieldValue, input, reflect.TypeOf(coverageAddr{}), recursiveSetFuncNoop, FieldOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal nested struct")
}

// ==================== deserializeStructFields: static "default" tag ====================

type coverageStaticDefault struct {
	Port int `json:"port" validate:"default=42"`
}

func TestDeserializeStructFields_StaticDefault(t *testing.T) {
	var target coverageStaticDefault
	structValue := reflect.ValueOf(&target).Elem()

	err := deserializeStructFields(structValue, reflect.TypeOf(coverageStaticDefault{}), map[string]any{}, recursiveSetFuncNoop, FieldOptions{TagName: "validate"})

	require.NoError(t, err)
	assert.Equal(t, 42, target.Port)
}

// ==================== deserializeStructFields: "defaultUsingMethod" tag ====================

type coverageMethodDefault struct {
	Port int `json:"port" validate:"defaultUsingMethod=GetPort"`
}

func (m *coverageMethodDefault) GetPort() (int, error) {
	return 9090, nil
}

func (m *coverageMethodDefault) GetPortErr() (int, error) {
	return 0, errors.New("boom")
}

func TestDeserializeStructFields_DefaultUsingMethod_Success(t *testing.T) {
	var target coverageMethodDefault
	structValue := reflect.ValueOf(&target).Elem()

	err := deserializeStructFields(structValue, reflect.TypeOf(coverageMethodDefault{}), map[string]any{}, recursiveSetFuncNoop, FieldOptions{TagName: "validate"})

	require.NoError(t, err)
	assert.Equal(t, 9090, target.Port)
}

func TestDeserializeStructFields_DefaultUsingMethod_MethodError(t *testing.T) {
	var target coverageMethodDefaultErr
	structValue := reflect.ValueOf(&target).Elem()

	err := deserializeStructFields(structValue, reflect.TypeOf(coverageMethodDefaultErr{}), map[string]any{}, recursiveSetFuncNoop, FieldOptions{TagName: "validate"})

	require.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

type coverageMethodDefaultErr struct {
	Port int `json:"port" validate:"defaultUsingMethod=GetPortErr"`
}

func (m *coverageMethodDefaultErr) GetPortErr() (int, error) {
	return 0, errors.New("boom")
}

func TestDeserializeStructFields_DefaultUsingMethod_MethodNotFound(t *testing.T) {
	type coverageMethodMissing struct {
		Port int `json:"port" validate:"defaultUsingMethod=NoSuchMethod"`
	}
	var target coverageMethodMissing
	structValue := reflect.ValueOf(&target).Elem()

	err := deserializeStructFields(structValue, reflect.TypeOf(coverageMethodMissing{}), map[string]any{}, recursiveSetFuncNoop, FieldOptions{TagName: "validate"})

	require.NoError(t, err)
	assert.Equal(t, 0, target.Port) // left at zero value since method wasn't found
}

func TestDeserializeStructFields_DefaultUsingMethod_NotAddressable(t *testing.T) {
	// A non-addressable struct value: structValue.CanAddr() is false, so the
	// method lookup is skipped entirely and the field is left at zero value.
	structValue := reflect.ValueOf(coverageMethodDefault{})

	err := deserializeStructFields(structValue, reflect.TypeOf(coverageMethodDefault{}), map[string]any{}, recursiveSetFuncNoop, FieldOptions{TagName: "validate"})

	require.NoError(t, err)
}

// ==================== setSliceField: struct elements ====================

type coverageItem struct {
	Name string `json:"name" validate:"required"`
}

func TestSetSliceField_StructElement_NotAMap(t *testing.T) {
	var target []coverageItem
	fieldValue := reflect.ValueOf(&target).Elem()

	// Element input is a map, but not map[string]any -> "expected map for struct element".
	inVal := reflect.ValueOf([]map[string]int{{"x": 1}})
	err := setSliceField(fieldValue, inVal, reflect.TypeOf(target), recursiveSetFuncNoop, FieldOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected map for struct element")
}

func TestSetSliceField_StructElement_CollectsRequiredErrors(t *testing.T) {
	var target []coverageItem
	fieldValue := reflect.ValueOf(&target).Elem()

	inVal := reflect.ValueOf([]any{
		map[string]any{}, // missing required "name"
	})
	err := setSliceField(fieldValue, inVal, reflect.TypeOf(target), recursiveSetFuncNoop, FieldOptions{StrictMissingFields: true, FieldName: "Items"})

	require.Error(t, err)
	var multiErr *MultiRequiredFieldError
	require.ErrorAs(t, err, &multiErr)
	require.Len(t, multiErr.Errors, 1)
	assert.Equal(t, "Items[0].Name", multiErr.Errors[0].Field)
}

// ==================== setMapField: key conversion and struct values ====================

type coverageMapKey int

func TestSetMapField_KeyConvertibleNotAssignable(t *testing.T) {
	var target map[coverageMapKey]string
	fieldValue := reflect.ValueOf(&target).Elem()

	inVal := reflect.ValueOf(map[int]string{1: "one"})
	err := setMapField(fieldValue, inVal, reflect.TypeOf(target), recursiveSetFuncNoop, FieldOptions{})

	require.NoError(t, err)
	assert.Equal(t, "one", target[coverageMapKey(1)])
}

type coverageIncompatibleKey struct{}

func TestSetMapField_KeyConversionError(t *testing.T) {
	var target map[coverageIncompatibleKey]string
	fieldValue := reflect.ValueOf(&target).Elem()

	// string keys cannot convert to a struct{} key type.
	inVal := reflect.ValueOf(map[string]string{"a": "one"})
	err := setMapField(fieldValue, inVal, reflect.TypeOf(target), recursiveSetFuncNoop, FieldOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert map key")
}

func TestSetMapField_StructValue_NotAMap(t *testing.T) {
	var target map[string]coverageItem
	fieldValue := reflect.ValueOf(&target).Elem()

	// Value is a map, but not map[string]any -> "expected map for struct value".
	inVal := reflect.ValueOf(map[string]map[int]string{"k": {1: "v"}})
	err := setMapField(fieldValue, inVal, reflect.TypeOf(target), recursiveSetFuncNoop, FieldOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected map for struct value")
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
