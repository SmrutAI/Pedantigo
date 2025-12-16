package deserialize

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== Primitive Types ====================

// TestSetFieldValue_PrimitiveTypes tests SetFieldValue primitivetypes.
func TestSetFieldValue_PrimitiveTypes(t *testing.T) {
	type TestStruct struct {
		Name   string
		Age    int
		Score  float64
		Active bool
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
	}{
		// String field tests
		{name: "set string valid", fieldName: "Name", value: "John", wantErr: false},
		{name: "set string empty", fieldName: "Name", value: "", wantErr: false},
		{name: "set string with spaces", fieldName: "Name", value: "John Doe", wantErr: false},

		// Int field tests
		{name: "set int positive", fieldName: "Age", value: 25, wantErr: false},
		{name: "set int zero", fieldName: "Age", value: 0, wantErr: false},
		{name: "set int negative", fieldName: "Age", value: -5, wantErr: false},

		// Float field tests
		{name: "set float positive", fieldName: "Score", value: 95.5, wantErr: false},
		{name: "set float zero", fieldName: "Score", value: 0.0, wantErr: false},
		{name: "set float negative", fieldName: "Score", value: -10.5, wantErr: false},

		// Bool field tests
		{name: "set bool true", fieldName: "Active", value: true, wantErr: false},
		{name: "set bool false", fieldName: "Active", value: false, wantErr: false},

		// Type mismatch tests
		{name: "string field with int", fieldName: "Name", value: 123, wantErr: true},
		{name: "int field with string", fieldName: "Age", value: "not an int", wantErr: true},
		{name: "float field with string", fieldName: "Score", value: "not a float", wantErr: true},
		{name: "bool field with string", fieldName: "Active", value: "not a bool", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				switch tt.fieldName {
				case "Name":
					assert.Equal(t, tt.value.(string), field.String())
				case "Age":
					assert.Equal(t, int64(tt.value.(int)), field.Int())
				case "Score":
					assert.InDelta(t, tt.value.(float64), field.Float(), 1e-9)
				case "Active":
					assert.Equal(t, tt.value.(bool), field.Bool())
				}
			}
		})
	}
}

// ==================== Integer Types ====================

// TestSetFieldValue_IntegerTypes tests SetFieldValue integertypes.
func TestSetFieldValue_IntegerTypes(t *testing.T) {
	type TestStruct struct {
		Int8Field   int8
		Int16Field  int16
		Int32Field  int32
		Int64Field  int64
		UintField   uint
		Uint8Field  uint8
		Uint16Field uint16
		Uint32Field uint32
		Uint64Field uint64
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
	}{
		// Signed integers
		{name: "int8 valid", fieldName: "Int8Field", value: int8(127), wantErr: false},
		{name: "int16 valid", fieldName: "Int16Field", value: int16(32767), wantErr: false},
		{name: "int32 valid", fieldName: "Int32Field", value: int32(2147483647), wantErr: false},
		{name: "int64 valid", fieldName: "Int64Field", value: int64(9223372036854775807), wantErr: false},

		// Unsigned integers
		{name: "uint valid", fieldName: "UintField", value: uint(100), wantErr: false},
		{name: "uint8 valid", fieldName: "Uint8Field", value: uint8(255), wantErr: false},
		{name: "uint16 valid", fieldName: "Uint16Field", value: uint16(65535), wantErr: false},
		{name: "uint32 valid", fieldName: "Uint32Field", value: uint32(4294967295), wantErr: false},
		{name: "uint64 valid", fieldName: "Uint64Field", value: uint64(18446744073709551615), wantErr: false},

		// Type conversion tests
		{name: "int64 to int8 convertible", fieldName: "Int8Field", value: int64(50), wantErr: false},
		{name: "int to int32 convertible", fieldName: "Int32Field", value: int(100), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Pointer Types ====================

// TestSetFieldValue_Pointers tests SetFieldValue pointers.
func TestSetFieldValue_Pointers(t *testing.T) {
	type TestStruct struct {
		StringPtr *string
		IntPtr    *int
		FloatPtr  *float64
		BoolPtr   *bool
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
		expectNil bool
	}{
		// String pointer tests
		{name: "string pointer valid", fieldName: "StringPtr", value: "hello", wantErr: false, expectNil: false},
		{name: "string pointer nil", fieldName: "StringPtr", value: nil, wantErr: false, expectNil: true},
		{name: "string pointer empty", fieldName: "StringPtr", value: "", wantErr: false, expectNil: false},

		// Int pointer tests
		{name: "int pointer valid", fieldName: "IntPtr", value: 42, wantErr: false, expectNil: false},
		{name: "int pointer nil", fieldName: "IntPtr", value: nil, wantErr: false, expectNil: true},
		{name: "int pointer zero", fieldName: "IntPtr", value: 0, wantErr: false, expectNil: false},

		// Float pointer tests
		{name: "float pointer valid", fieldName: "FloatPtr", value: 3.14, wantErr: false, expectNil: false},
		{name: "float pointer nil", fieldName: "FloatPtr", value: nil, wantErr: false, expectNil: true},

		// Bool pointer tests
		{name: "bool pointer true", fieldName: "BoolPtr", value: true, wantErr: false, expectNil: false},
		{name: "bool pointer false", fieldName: "BoolPtr", value: false, wantErr: false, expectNil: false},
		{name: "bool pointer nil", fieldName: "BoolPtr", value: nil, wantErr: false, expectNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectNil {
					assert.True(t, field.IsNil(), "expected nil pointer")
				} else {
					assert.False(t, field.IsNil(), "expected non-nil pointer")
				}
			}
		})
	}
}

// ==================== Slice Types ====================

// TestSetFieldValue_Slices tests SetFieldValue slices.
func TestSetFieldValue_Slices(t *testing.T) {
	type TestStruct struct {
		StringSlice []string
		IntSlice    []int
		FloatSlice  []float64
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
		expectLen int
	}{
		// String slice tests
		{name: "string slice with values", fieldName: "StringSlice", value: []any{"a", "b", "c"}, wantErr: false, expectLen: 3},
		{name: "string slice empty", fieldName: "StringSlice", value: []any{}, wantErr: false, expectLen: 0},
		{name: "string slice nil", fieldName: "StringSlice", value: nil, wantErr: false, expectLen: 0},

		// Int slice tests
		{name: "int slice with values", fieldName: "IntSlice", value: []any{1, 2, 3}, wantErr: false, expectLen: 3},
		{name: "int slice single", fieldName: "IntSlice", value: []any{42}, wantErr: false, expectLen: 1},
		{name: "int slice nil", fieldName: "IntSlice", value: nil, wantErr: false, expectLen: 0},

		// Float slice tests
		{name: "float slice with values", fieldName: "FloatSlice", value: []any{1.1, 2.2, 3.3}, wantErr: false, expectLen: 3},
		{name: "float slice nil", fieldName: "FloatSlice", value: nil, wantErr: false, expectLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.value != nil {
					assert.Equal(t, tt.expectLen, field.Len())
				}
			}
		})
	}
}

// ==================== Map Types ====================

// TestSetFieldValue_Maps tests SetFieldValue maps.
func TestSetFieldValue_Maps(t *testing.T) {
	type TestStruct struct {
		StringMap map[string]string
		IntMap    map[string]int
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
		expectLen int
	}{
		// String map tests
		{name: "string map with values", fieldName: "StringMap", value: map[string]any{"key1": "val1", "key2": "val2"}, wantErr: false, expectLen: 2},
		{name: "string map empty", fieldName: "StringMap", value: map[string]any{}, wantErr: false, expectLen: 0},
		{name: "string map nil", fieldName: "StringMap", value: nil, wantErr: false, expectLen: 0},

		// Int map tests
		{name: "int map with values", fieldName: "IntMap", value: map[string]any{"a": 1, "b": 2}, wantErr: false, expectLen: 2},
		{name: "int map nil", fieldName: "IntMap", value: nil, wantErr: false, expectLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.value != nil {
					assert.Equal(t, tt.expectLen, field.Len())
				}
			}
		})
	}
}

// ==================== Type Conversion ====================

// TestSetFieldValue_TypeConversion tests SetFieldValue typeconversion.
func TestSetFieldValue_TypeConversion(t *testing.T) {
	type TestStruct struct {
		IntField     int
		UintField    uint
		Float64Field float64
		Float32Field float32
		StringField  string
		BoolField    bool
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
	}{
		// JSON number to int conversion
		{name: "int32 to int", fieldName: "IntField", value: int32(42), wantErr: false},
		{name: "int64 to int", fieldName: "IntField", value: int64(42), wantErr: false},
		{name: "uint to int", fieldName: "IntField", value: uint(42), wantErr: false},

		// JSON number to uint conversion
		{name: "int to uint", fieldName: "UintField", value: 42, wantErr: false},
		{name: "int32 to uint", fieldName: "UintField", value: int32(42), wantErr: false},
		{name: "int64 to uint", fieldName: "UintField", value: int64(42), wantErr: false},

		// JSON number to float64 conversion
		{name: "int to float64", fieldName: "Float64Field", value: 42, wantErr: false},
		{name: "int64 to float64", fieldName: "Float64Field", value: int64(42), wantErr: false},
		{name: "float32 to float64", fieldName: "Float64Field", value: float32(3.14), wantErr: false},

		// JSON number to float32 conversion
		{name: "int to float32", fieldName: "Float32Field", value: 42, wantErr: false},
		{name: "float64 to float32", fieldName: "Float32Field", value: 3.14, wantErr: false},

		// Invalid conversions
		{name: "bool to int invalid", fieldName: "IntField", value: true, wantErr: true},
		{name: "string to int invalid", fieldName: "IntField", value: "123", wantErr: true},
		{name: "slice to int invalid", fieldName: "IntField", value: []int{1, 2, 3}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Time Type ====================

// TestSetFieldValue_TimeType tests SetFieldValue timetype.
func TestSetFieldValue_TimeType(t *testing.T) {
	type TestStruct struct {
		CreatedAt time.Time
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
		checkFn func(*TestStruct) bool
	}{
		{
			name:    "RFC3339 format valid",
			value:   "2023-01-15T10:30:00Z",
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return s.CreatedAt.Year() == 2023 && s.CreatedAt.Month() == 1 && s.CreatedAt.Day() == 15
			},
		},
		{
			name:    "RFC3339 format with timezone",
			value:   "2023-12-25T15:45:30+00:00",
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return s.CreatedAt.Year() == 2023 && s.CreatedAt.Month() == 12 && s.CreatedAt.Day() == 25
			},
		},
		{
			name:    "invalid format",
			value:   "invalid-date",
			wantErr: true,
		},
		{
			name:    "wrong type",
			value:   12345,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName("CreatedAt")

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFn != nil {
					assert.True(t, tt.checkFn(s), "time value check failed")
				}
			}
		})
	}
}

// ==================== Struct Types ====================

// TestSetFieldValue_NestedStruct tests SetFieldValue nestedstruct.
func TestSetFieldValue_NestedStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type TestStruct struct {
		Address Address
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
		checkFn func(*TestStruct) bool
	}{
		{
			name:    "nested struct valid",
			value:   map[string]any{"street": "123 Main St", "city": "New York"},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return s.Address.Street == "123 Main St" && s.Address.City == "New York"
			},
		},
		{
			name:    "nested struct partial",
			value:   map[string]any{"street": "456 Oak Ave"},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return s.Address.Street == "456 Oak Ave" && s.Address.City == ""
			},
		},
		{
			name:    "nested struct empty",
			value:   map[string]any{},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return s.Address.Street == "" && s.Address.City == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName("Address")

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFn != nil {
					assert.True(t, tt.checkFn(s), "struct value check failed")
				}
			}
		})
	}
}

// ==================== Unset Field Handling ====================

// TestSetFieldValue_UnsetField tests SetFieldValue unsetfield.
func TestSetFieldValue_UnsetField(t *testing.T) {
	type TestStruct struct {
		_unexported string //nolint:unused // unexported field for testing CanSet()
		Exported    string
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
	}{
		// Unexported fields cannot be set
		{name: "unexported field", fieldName: "_unexported", value: "test", wantErr: false}, // CanSet() returns false, so no error

		// Exported fields can be set
		{name: "exported field", fieldName: "Exported", value: "test", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Slice of Structs ====================

// TestSetFieldValue_SliceOfStructs tests SetFieldValue sliceofstructs.
func TestSetFieldValue_SliceOfStructs(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type TestStruct struct {
		Items []Item
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
		checkFn func(*TestStruct) bool
	}{
		{
			name: "slice of structs with values",
			value: []any{
				map[string]any{"id": 1, "name": "Item1"},
				map[string]any{"id": 2, "name": "Item2"},
			},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return len(s.Items) == 2 && s.Items[0].ID == 1 && s.Items[0].Name == "Item1"
			},
		},
		{
			name:    "slice of structs empty",
			value:   []any{},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return len(s.Items) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName("Items")

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFn != nil {
					assert.True(t, tt.checkFn(s), "slice of structs value check failed")
				}
			}
		})
	}
}

// ==================== Map with Struct Values ====================

// TestSetFieldValue_MapWithStructValues tests SetFieldValue mapwithstructvalues.
func TestSetFieldValue_MapWithStructValues(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type TestStruct struct {
		Users map[string]User
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
		checkFn func(*TestStruct) bool
	}{
		{
			name: "map with struct values",
			value: map[string]any{
				"alice": map[string]any{"name": "Alice", "age": 30},
				"bob":   map[string]any{"name": "Bob", "age": 25},
			},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return len(s.Users) == 2 && s.Users["alice"].Name == "Alice" && s.Users["alice"].Age == 30
			},
		},
		{
			name:    "map with struct values empty",
			value:   map[string]any{},
			wantErr: false,
			checkFn: func(s *TestStruct) bool {
				return len(s.Users) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName("Users")

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFn != nil {
					assert.True(t, tt.checkFn(s), "map with struct values check failed")
				}
			}
		})
	}
}

// ==================== Edge Cases ====================

// TestSetFieldValue_EdgeCases tests SetFieldValue edgecases.
func TestSetFieldValue_EdgeCases(t *testing.T) {
	type TestStruct struct {
		String string
		Int    int
		Float  float64
		Slice  []string
		Map    map[string]int
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
	}{
		// Zero values should not error
		{name: "string zero value", fieldName: "String", value: "", wantErr: false},
		{name: "int zero value", fieldName: "Int", value: 0, wantErr: false},
		{name: "float zero value", fieldName: "Float", value: 0.0, wantErr: false},

		// Very large numbers
		{name: "large int", fieldName: "Int", value: 9223372036854775807, wantErr: false}, // Valid on 64-bit (int=int64), would overflow on 32-bit

		// Special float values
		{name: "float infinity", fieldName: "Float", value: math.Inf(1), wantErr: false},
		{name: "float negative infinity", fieldName: "Float", value: math.Inf(-1), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Pointer Slice ====================

// TestSetFieldValue_PointerSlice tests SetFieldValue pointerslice.
func TestSetFieldValue_PointerSlice(t *testing.T) {
	type TestStruct struct {
		StringPtrs []*string
		IntPtrs    []*int
	}

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantErr   bool
		expectLen int
	}{
		{
			name:      "pointer slice with values",
			fieldName: "StringPtrs",
			value:     []any{"hello", "world"},
			wantErr:   false,
			expectLen: 2,
		},
		{
			name:      "pointer slice with nil elements",
			fieldName: "IntPtrs",
			value:     []any{nil, 42, nil},
			wantErr:   false,
			expectLen: 3,
		},
		{
			name:      "pointer slice empty",
			fieldName: "StringPtrs",
			value:     []any{},
			wantErr:   false,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.value != nil {
					assert.Equal(t, tt.expectLen, field.Len())
				}
			}
		})
	}
}

// ==================== Pointer Pointer ====================

// TestSetFieldValue_PointerPointer tests SetFieldValue pointerpointer.
func TestSetFieldValue_PointerPointer(t *testing.T) {
	type TestStruct struct {
		DoublePtr **string
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "pointer pointer valid", value: "hello", wantErr: false},
		{name: "pointer pointer nil", value: nil, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName("DoublePtr")

			err := SetFieldValue(field, tt.value, field.Type(), recursiveSetFuncNoop)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== SetDefaultValue Tests ====================

// TestSetDefaultValue_StringDefaults tests SetDefaultValue stringdefaults.
func TestSetDefaultValue_StringDefaults(t *testing.T) {
	type TestStruct struct {
		Name    string
		Message string
		Empty   string
	}

	tests := []struct {
		name         string
		fieldName    string
		defaultValue string
		expected     string
	}{
		{name: "string simple", fieldName: "Name", defaultValue: "John", expected: "John"},
		{name: "string empty", fieldName: "Empty", defaultValue: "", expected: ""},
		{name: "string with spaces", fieldName: "Message", defaultValue: "Hello World", expected: "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			SetDefaultValue(field, tt.defaultValue, recursiveSetDefault)

			assert.Equal(t, tt.expected, field.String())
		})
	}
}

func TestSetDefaultValue_IntDefaults(t *testing.T) {
	type TestStruct struct {
		Int     int
		Int8    int8
		Int16   int16
		Int32   int32
		Int64   int64
		Invalid int
	}

	tests := []struct {
		name         string
		fieldName    string
		defaultValue string
		expected     int64
	}{
		{name: "int positive", fieldName: "Int", defaultValue: "42", expected: 42},
		{name: "int zero", fieldName: "Int", defaultValue: "0", expected: 0},
		{name: "int negative", fieldName: "Int", defaultValue: "-100", expected: -100},
		{name: "int8 max", fieldName: "Int8", defaultValue: "127", expected: 127},
		{name: "int16 max", fieldName: "Int16", defaultValue: "32767", expected: 32767},
		{name: "int32 max", fieldName: "Int32", defaultValue: "2147483647", expected: 2147483647},
		{name: "int64 max", fieldName: "Int64", defaultValue: "9223372036854775807", expected: 9223372036854775807},
		{name: "int invalid parse", fieldName: "Invalid", defaultValue: "not-a-number", expected: 0}, // Silent failure
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			SetDefaultValue(field, tt.defaultValue, recursiveSetDefault)

			assert.Equal(t, tt.expected, field.Int())
		})
	}
}

func TestSetDefaultValue_UintDefaults(t *testing.T) {
	type TestStruct struct {
		Uint    uint
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Invalid uint
	}

	tests := []struct {
		name         string
		fieldName    string
		defaultValue string
		expected     uint64
	}{
		{name: "uint positive", fieldName: "Uint", defaultValue: "42", expected: 42},
		{name: "uint zero", fieldName: "Uint", defaultValue: "0", expected: 0},
		{name: "uint8 max", fieldName: "Uint8", defaultValue: "255", expected: 255},
		{name: "uint16 max", fieldName: "Uint16", defaultValue: "65535", expected: 65535},
		{name: "uint32 max", fieldName: "Uint32", defaultValue: "4294967295", expected: 4294967295},
		{name: "uint64 max", fieldName: "Uint64", defaultValue: "18446744073709551615", expected: 18446744073709551615},
		{name: "uint invalid parse", fieldName: "Invalid", defaultValue: "not-a-number", expected: 0}, // Silent failure
		{name: "uint negative invalid", fieldName: "Invalid", defaultValue: "-1", expected: 0},        // Silent failure on negative
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			SetDefaultValue(field, tt.defaultValue, recursiveSetDefault)

			assert.Equal(t, tt.expected, field.Uint())
		})
	}
}

func TestSetDefaultValue_FloatDefaults(t *testing.T) {
	type TestStruct struct {
		Float32 float32
		Float64 float64
		Invalid float64
	}

	tests := []struct {
		name         string
		fieldName    string
		defaultValue string
		expected     float64
	}{
		{name: "float64 positive", fieldName: "Float64", defaultValue: "3.14", expected: 3.14},
		{name: "float64 zero", fieldName: "Float64", defaultValue: "0.0", expected: 0.0},
		{name: "float64 negative", fieldName: "Float64", defaultValue: "-10.5", expected: -10.5},
		{name: "float32 valid", fieldName: "Float32", defaultValue: "2.5", expected: 2.5},
		{name: "float invalid parse", fieldName: "Invalid", defaultValue: "not-a-float", expected: 0.0}, // Silent failure
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			SetDefaultValue(field, tt.defaultValue, recursiveSetDefault)

			assert.InDelta(t, tt.expected, field.Float(), 1e-9)
		})
	}
}

func TestSetDefaultValue_BoolDefaults(t *testing.T) {
	type TestStruct struct {
		Active  bool
		Enabled bool
		Invalid bool
	}

	tests := []struct {
		name         string
		fieldName    string
		defaultValue string
		expected     bool
	}{
		{name: "bool true", fieldName: "Active", defaultValue: "true", expected: true},
		{name: "bool false", fieldName: "Enabled", defaultValue: "false", expected: false},
		{name: "bool 1", fieldName: "Active", defaultValue: "1", expected: true},
		{name: "bool 0", fieldName: "Enabled", defaultValue: "0", expected: false},
		{name: "bool invalid parse", fieldName: "Invalid", defaultValue: "not-a-bool", expected: false}, // Silent failure
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			SetDefaultValue(field, tt.defaultValue, recursiveSetDefault)

			assert.Equal(t, tt.expected, field.Bool())
		})
	}
}

func TestSetDefaultValue_PointerDefaults(t *testing.T) {
	type TestStruct struct {
		StringPtr *string
		IntPtr    *int
		FloatPtr  *float64
		BoolPtr   *bool
	}

	tests := []struct {
		name         string
		fieldName    string
		defaultValue string
		checkFn      func(*TestStruct) bool
	}{
		{
			name:         "pointer string",
			fieldName:    "StringPtr",
			defaultValue: "hello",
			checkFn: func(s *TestStruct) bool {
				return s.StringPtr != nil && *s.StringPtr == "hello"
			},
		},
		{
			name:         "pointer int",
			fieldName:    "IntPtr",
			defaultValue: "42",
			checkFn: func(s *TestStruct) bool {
				return s.IntPtr != nil && *s.IntPtr == 42
			},
		},
		{
			name:         "pointer float",
			fieldName:    "FloatPtr",
			defaultValue: "3.14",
			checkFn: func(s *TestStruct) bool {
				return s.FloatPtr != nil && *s.FloatPtr == 3.14
			},
		},
		{
			name:         "pointer bool",
			fieldName:    "BoolPtr",
			defaultValue: "true",
			checkFn: func(s *TestStruct) bool {
				return s.BoolPtr != nil && *s.BoolPtr == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{}
			v := reflect.ValueOf(s).Elem()
			field := v.FieldByName(tt.fieldName)

			SetDefaultValue(field, tt.defaultValue, recursiveSetDefault)

			assert.True(t, tt.checkFn(s), "pointer default value check failed")
		})
	}
}

func TestSetDefaultValue_UnsettableField(t *testing.T) {
	type TestStruct struct {
		_unexported string //nolint:unused // unexported field for testing CanSet()
		Exported    string
	}

	s := &TestStruct{}
	v := reflect.ValueOf(s).Elem()

	// Try to set unexported field - should silently fail (CanSet() returns false)
	unexported := v.FieldByName("_unexported")
	SetDefaultValue(unexported, "test", recursiveSetDefault)
	// No panic should occur, field remains zero value

	// Set exported field - should work
	exported := v.FieldByName("Exported")
	SetDefaultValue(exported, "success", recursiveSetDefault)

	assert.Equal(t, "success", exported.String())
}

// ==================== Helper Functions ====================

// recursiveSetFuncNoop is a no-op recursive set function for testing.
func recursiveSetFuncNoop(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error {
	// For primitive types, delegate to SetFieldValue
	return SetFieldValue(fieldValue, inValue, fieldType, recursiveSetFuncNoop)
}

// recursiveSetDefault is a helper for SetDefaultValue recursive calls.
func recursiveSetDefault(fieldValue reflect.Value, defaultValue string) {
	SetDefaultValue(fieldValue, defaultValue, recursiveSetDefault)
}

// ==================== isValidConversion Tests ====================

// ==================== Duration Type ====================

// TestSetFieldValue_Duration tests time.Duration handling in SetFieldValue.
func TestSetFieldValue_Duration(t *testing.T) {
	type DurationStruct struct {
		Timeout time.Duration
	}

	tests := []struct {
		name    string
		input   any
		want    time.Duration
		wantErr bool
	}{
		// String format - Go duration strings
		{"string 1h30m", "1h30m", 90 * time.Minute, false},
		{"string 500ms", "500ms", 500 * time.Millisecond, false},
		{"string 2h45m30s", "2h45m30s", 2*time.Hour + 45*time.Minute + 30*time.Second, false},
		{"string 1s", "1s", time.Second, false},
		{"string 100ns", "100ns", 100 * time.Nanosecond, false},
		{"string 1m30s", "1m30s", 90 * time.Second, false},
		{"string 0s", "0s", 0, false},
		{"string negative -1h", "-1h", -1 * time.Hour, false},
		// Invalid strings
		{"invalid string hello", "hello", 0, true},
		{"invalid string 1x", "1x", 0, true},
		{"invalid string empty", "", 0, true}, // Empty string is invalid duration
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result DurationStruct
			resultVal := reflect.ValueOf(&result).Elem()
			fieldVal := resultVal.Field(0)

			var recursiveSetFunc func(fv reflect.Value, iv any, ft reflect.Type) error
			recursiveSetFunc = func(fv reflect.Value, iv any, ft reflect.Type) error {
				return SetFieldValue(fv, iv, ft, recursiveSetFunc)
			}

			err := SetFieldValue(fieldVal, tt.input, reflect.TypeOf(time.Duration(0)), recursiveSetFunc)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Timeout != tt.want {
				t.Errorf("SetFieldValue() got = %v, want %v", result.Timeout, tt.want)
			}
		})
	}
}

// ==================== isValidConversion Tests ====================

// Test_isValidConversion tests  isvalidconversion.
func Test_isValidConversion(t *testing.T) {
	tests := []struct {
		name     string
		from     reflect.Type
		to       reflect.Type
		expected bool
	}{
		// Numeric to numeric conversions - should be allowed
		{name: "int to int64", from: reflect.TypeOf(int(0)), to: reflect.TypeOf(int64(0)), expected: true},
		{name: "int32 to int", from: reflect.TypeOf(int32(0)), to: reflect.TypeOf(int(0)), expected: true},
		{name: "uint to uint64", from: reflect.TypeOf(uint(0)), to: reflect.TypeOf(uint64(0)), expected: true},
		{name: "int to float64", from: reflect.TypeOf(int(0)), to: reflect.TypeOf(float64(0)), expected: true},
		{name: "float32 to float64", from: reflect.TypeOf(float32(0)), to: reflect.TypeOf(float64(0)), expected: true},
		{name: "uint8 to int16", from: reflect.TypeOf(uint8(0)), to: reflect.TypeOf(int16(0)), expected: true},

		// Numeric to string conversions - should be blocked (would convert to rune)
		{name: "int to string blocked", from: reflect.TypeOf(int(0)), to: reflect.TypeOf(""), expected: false},
		{name: "int64 to string blocked", from: reflect.TypeOf(int64(0)), to: reflect.TypeOf(""), expected: false},
		{name: "uint to string blocked", from: reflect.TypeOf(uint(0)), to: reflect.TypeOf(""), expected: false},
		{name: "float64 to string blocked", from: reflect.TypeOf(float64(0)), to: reflect.TypeOf(""), expected: false},

		// String to numeric conversions - should be blocked (panics at runtime)
		{name: "string to int blocked", from: reflect.TypeOf(""), to: reflect.TypeOf(int(0)), expected: false},
		{name: "string to int64 blocked", from: reflect.TypeOf(""), to: reflect.TypeOf(int64(0)), expected: false},
		{name: "string to uint blocked", from: reflect.TypeOf(""), to: reflect.TypeOf(uint(0)), expected: false},
		{name: "string to float64 blocked", from: reflect.TypeOf(""), to: reflect.TypeOf(float64(0)), expected: false},

		// Same kind conversions - should be allowed
		{name: "string to string", from: reflect.TypeOf(""), to: reflect.TypeOf(""), expected: true},
		{name: "bool to bool", from: reflect.TypeOf(false), to: reflect.TypeOf(false), expected: true},
		{name: "slice to slice", from: reflect.TypeOf([]int{}), to: reflect.TypeOf([]int{}), expected: true},
		{name: "map to map", from: reflect.TypeOf(map[string]int{}), to: reflect.TypeOf(map[string]int{}), expected: true},

		// Different non-numeric kinds - should be blocked
		{name: "string to bool blocked", from: reflect.TypeOf(""), to: reflect.TypeOf(false), expected: false},
		{name: "bool to string blocked", from: reflect.TypeOf(false), to: reflect.TypeOf(""), expected: false},
		{name: "slice to map blocked", from: reflect.TypeOf([]int{}), to: reflect.TypeOf(map[string]int{}), expected: false},
		{name: "struct to string blocked", from: reflect.TypeOf(struct{}{}), to: reflect.TypeOf(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidConversion(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}
