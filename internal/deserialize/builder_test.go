package deserialize

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildFieldDeserializers_BasicTypes tests deserializer creation for primitive types
func TestBuildFieldDeserializers_BasicTypes(t *testing.T) {
	tests := []struct {
		name               string
		structType         any
		expectedFieldCount int
		expectedFields     []string
		wantErr            bool
	}{
		{
			name: "string field",
			structType: struct {
				Name string
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Name"},
		},
		{
			name: "int field",
			structType: struct {
				Age int
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Age"},
		},
		{
			name: "float64 field",
			structType: struct {
				Score float64
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Score"},
		},
		{
			name: "bool field",
			structType: struct {
				Enabled bool
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Enabled"},
		},
		{
			name: "multiple primitive fields",
			structType: struct {
				Name    string
				Age     int
				Score   float64
				Enabled bool
			}{},
			expectedFieldCount: 4,
			expectedFields:     []string{"Name", "Age", "Score", "Enabled"},
		},
		{
			name: "pointer to string",
			structType: struct {
				Name *string
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Name"},
		},
		{
			name: "pointer to int",
			structType: struct {
				Age *int
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Age"},
		},
		{
			name: "mixed pointers and values",
			structType: struct {
				Name  *string
				Age   int
				Score *float64
			}{},
			expectedFieldCount: 3,
			expectedFields:     []string{"Name", "Age", "Score"},
		},
		{
			name: "pointer to struct",
			structType: (*struct {
				Name string
			})(nil),
			expectedFieldCount: 1,
			expectedFields:     []string{"Name"},
		},
		{
			name: "no fields",
			structType: struct {
			}{},
			expectedFieldCount: 0,
			expectedFields:     []string{},
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			assert.Equal(t, tt.expectedFieldCount, len(deserializers), "deserializer count mismatch")

			for _, fieldName := range tt.expectedFields {
				_, exists := deserializers[fieldName]
				assert.True(t, exists, "expected deserializer for field %q", fieldName)
			}
		})
	}
}

// TestBuildFieldDeserializers_JSONTags tests JSON tag handling
func TestBuildFieldDeserializers_JSONTags(t *testing.T) {
	tests := []struct {
		name           string
		structType     any
		expectedFields []string
		wantErr        bool
	}{
		{
			name: "json tag renames field",
			structType: struct {
				Name string `json:"user_name"`
			}{},
			expectedFields: []string{"user_name"},
		},
		{
			name: "json tag with options",
			structType: struct {
				Name string `json:"user_name,omitempty"`
			}{},
			expectedFields: []string{"user_name"},
		},
		{
			name: "json tag ignored with dash",
			structType: struct {
				Name string `json:"-"`
			}{},
			expectedFields: []string{},
		},
		{
			name: "mixed json tags",
			structType: struct {
				Name  string `json:"user_name"`
				Age   int    `json:"age"`
				Score float64
			}{},
			expectedFields: []string{"user_name", "age", "Score"},
		},
		{
			name: "json tag with empty name uses field name",
			structType: struct {
				Name string `json:""`
			}{},
			expectedFields: []string{"Name"},
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			assert.Equal(t, len(tt.expectedFields), len(deserializers), "deserializer count mismatch")

			for _, fieldName := range tt.expectedFields {
				_, exists := deserializers[fieldName]
				assert.True(t, exists, "expected deserializer for field %q", fieldName)
			}
		})
	}
}

// TestBuildFieldDeserializers_UnexportedFields tests that unexported fields are skipped
func TestBuildFieldDeserializers_UnexportedFields(t *testing.T) {
	tests := []struct {
		name           string
		structType     any
		expectedFields []string
	}{
		{
			name: "skip unexported field",
			structType: struct {
				Name string
				age  int // unexported
			}{},
			expectedFields: []string{"Name"},
		},
		{
			name: "all exported fields",
			structType: struct {
				Name string
				Age  int
			}{},
			expectedFields: []string{"Name", "Age"},
		},
		{
			name: "all unexported fields",
			structType: struct {
				name string
				age  int
			}{},
			expectedFields: []string{},
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			assert.Equal(t, len(tt.expectedFields), len(deserializers), "expected %d deserializers, got %d; got fields: %v", len(tt.expectedFields), len(deserializers), getMapKeys(deserializers))

			for _, fieldName := range tt.expectedFields {
				_, exists := deserializers[fieldName]
				assert.True(t, exists, "expected deserializer for field %q", fieldName)
			}
		})
	}
}

// TestBuildFieldDeserializers_Defaults tests default value handling
func TestBuildFieldDeserializers_Defaults(t *testing.T) {
	tests := []struct {
		name                 string
		structType           any
		strictMissingFields  bool
		shouldPanic          bool
		panicMessageContains string
	}{
		{
			name: "string default in strict mode",
			structType: struct {
				Name string `pedantigo:"default=guest"`
			}{},
			strictMissingFields: true,
			shouldPanic:         false,
		},
		{
			name: "int default in strict mode",
			structType: struct {
				Port int `pedantigo:"default=8080"`
			}{},
			strictMissingFields: true,
			shouldPanic:         false,
		},
		{
			name: "bool default in strict mode",
			structType: struct {
				Enabled bool `pedantigo:"default=true"`
			}{},
			strictMissingFields: true,
			shouldPanic:         false,
		},
		{
			name: "float64 default in strict mode",
			structType: struct {
				Timeout float64 `pedantigo:"default=30.5"`
			}{},
			strictMissingFields: true,
			shouldPanic:         false,
		},
		{
			name: "empty string default",
			structType: struct {
				Name string `pedantigo:"default="`
			}{},
			strictMissingFields: true,
			shouldPanic:         false,
		},
		{
			name: "default tag in non-strict mode panics",
			structType: struct {
				Name string `pedantigo:"default=guest"`
			}{},
			strictMissingFields:  false,
			shouldPanic:          true,
			panicMessageContains: "has 'default=' tag but StrictMissingFields is false",
		},
		{
			name: "multiple defaults",
			structType: struct {
				Name    string `pedantigo:"default=guest"`
				Port    int    `pedantigo:"default=8080"`
				Enabled bool   `pedantigo:"default=true"`
			}{},
			strictMissingFields: true,
			shouldPanic:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := BuilderOptions{StrictMissingFields: tt.strictMissingFields}
			typ := reflect.TypeOf(tt.structType)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					BuildFieldDeserializers(typ, opts, nil, nil)
				}, "expected panic")
			} else {
				assert.NotPanics(t, func() {
					BuildFieldDeserializers(typ, opts, nil, nil)
				}, "unexpected panic")
			}
		})
	}
}

// TestBuildFieldDeserializers_DefaultUsingMethod tests defaultUsingMethod tag handling
func TestBuildFieldDeserializers_DefaultUsingMethod(t *testing.T) {
	tests := []struct {
		name                 string
		structType           any
		strictMissingFields  bool
		shouldPanic          bool
		panicMessageContains string
	}{
		{
			name: "defaultUsingMethod with non-strict mode panics",
			structType: struct {
				Port int `pedantigo:"defaultUsingMethod=GetPort"`
			}{},
			strictMissingFields:  false,
			shouldPanic:          true,
			panicMessageContains: "has 'defaultUsingMethod=' tag but StrictMissingFields is false",
		},
		{
			name: "defaultUsingMethod with missing method panics",
			structType: struct {
				Age int `pedantigo:"defaultUsingMethod=NonExistentMethod"`
			}{},
			strictMissingFields:  true,
			shouldPanic:          true,
			panicMessageContains: "method NonExistentMethod not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := BuilderOptions{StrictMissingFields: tt.strictMissingFields}
			typ := reflect.TypeOf(tt.structType)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					BuildFieldDeserializers(typ, opts, nil, nil)
				}, "expected panic")
			}
		})
	}
}

// TestBuildFieldDeserializers_Collections tests slice and map handling
func TestBuildFieldDeserializers_Collections(t *testing.T) {
	tests := []struct {
		name               string
		structType         any
		expectedFieldCount int
		expectedFields     []string
	}{
		{
			name: "slice of strings",
			structType: struct {
				Tags []string
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Tags"},
		},
		{
			name: "slice of ints",
			structType: struct {
				Scores []int
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Scores"},
		},
		{
			name: "map with string keys and int values",
			structType: struct {
				Mapping map[string]int
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Mapping"},
		},
		{
			name: "map with string keys and string values",
			structType: struct {
				Config map[string]string
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Config"},
		},
		{
			name: "slice of slices",
			structType: struct {
				Matrix [][]int
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Matrix"},
		},
		{
			name: "pointer to slice",
			structType: struct {
				Tags *[]string
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Tags"},
		},
		{
			name: "pointer to map",
			structType: struct {
				Config *map[string]string
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Config"},
		},
		{
			name: "mixed collections",
			structType: struct {
				Tags    []string
				Scores  []int
				Mapping map[string]int
				Config  map[string]string
			}{},
			expectedFieldCount: 4,
			expectedFields:     []string{"Tags", "Scores", "Mapping", "Config"},
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			assert.Equal(t, tt.expectedFieldCount, len(deserializers), "deserializer count mismatch")

			for _, fieldName := range tt.expectedFields {
				_, exists := deserializers[fieldName]
				assert.True(t, exists, "expected deserializer for field %q", fieldName)
			}
		})
	}
}

// TestBuildFieldDeserializers_NonStructType tests handling of non-struct types
func TestBuildFieldDeserializers_NonStructType(t *testing.T) {
	tests := []struct {
		name               string
		inputType          any
		expectedFieldCount int
	}{
		{
			name:               "string type",
			inputType:          "string",
			expectedFieldCount: 0,
		},
		{
			name:               "int type",
			inputType:          42,
			expectedFieldCount: 0,
		},
		{
			name:               "slice type",
			inputType:          []int{},
			expectedFieldCount: 0,
		},
		{
			name:               "map type",
			inputType:          map[string]int{},
			expectedFieldCount: 0,
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.inputType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			assert.Equal(t, tt.expectedFieldCount, len(deserializers), "deserializer count mismatch")
		})
	}
}

// TestBuildFieldDeserializers_FieldDeserializerCallable tests that deserializers are callable
func TestBuildFieldDeserializers_FieldDeserializerCallable(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	opts := BuilderOptions{StrictMissingFields: true}
	typ := reflect.TypeOf(TestStruct{})

	var mockSetFieldCalled bool
	mockSetField := func(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error {
		mockSetFieldCalled = true
		return nil
	}

	deserializers := BuildFieldDeserializers(typ, opts, mockSetField, nil)

	// Create instance for testing
	instance := TestStruct{}
	outPtr := reflect.ValueOf(&instance).Elem()

	// Test that deserializer can be called with a present field
	nameDeserializer, exists := deserializers["Name"]
	require.True(t, exists, "expected Name deserializer to exist")

	err := nameDeserializer(&outPtr, "John")
	assert.NoError(t, err, "unexpected error calling deserializer")
	assert.True(t, mockSetFieldCalled, "expected setFieldValueFunc to be called")

	// Test that deserializer can be called with missing field (sentinel)
	mockSetFieldCalled = false
	err = nameDeserializer(&outPtr, FieldMissingSentinel)
	assert.NoError(t, err, "unexpected error calling deserializer with sentinel")
	assert.False(t, mockSetFieldCalled, "expected setFieldValueFunc NOT to be called for missing field")
}

// TestBuildFieldDeserializers_RequiredFieldValidation tests required field tag validation
func TestBuildFieldDeserializers_RequiredFieldValidation(t *testing.T) {
	type TestStruct struct {
		Name string `pedantigo:"required"`
		Age  int
	}

	tests := []struct {
		name                string
		strictMissingFields bool
		fieldValue          any
		expectError         bool
	}{
		{
			name:                "required field missing in strict mode returns error",
			strictMissingFields: true,
			fieldValue:          FieldMissingSentinel,
			expectError:         true,
		},
		{
			name:                "required field present in strict mode no error",
			strictMissingFields: true,
			fieldValue:          "John",
			expectError:         false,
		},
		{
			name:                "required field missing in non-strict mode no error",
			strictMissingFields: false,
			fieldValue:          FieldMissingSentinel,
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := BuilderOptions{StrictMissingFields: tt.strictMissingFields}
			typ := reflect.TypeOf(TestStruct{})

			deserializers := BuildFieldDeserializers(typ, opts, func(fv reflect.Value, iv any, ft reflect.Type) error {
				return nil
			}, nil)

			instance := TestStruct{}
			outPtr := reflect.ValueOf(&instance).Elem()

			nameDeserializer, exists := deserializers["Name"]
			require.True(t, exists, "expected Name deserializer to exist")

			err := nameDeserializer(&outPtr, tt.fieldValue)

			if tt.expectError {
				assert.Error(t, err, "expected error for required field")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

// TestBuildFieldDeserializers_ConstraintsParsed tests that constraints are parsed from tags
func TestBuildFieldDeserializers_ConstraintsParsed(t *testing.T) {
	tests := []struct {
		name           string
		structType     any
		expectedFields []string
	}{
		{
			name: "field with email constraint",
			structType: struct {
				Email string `pedantigo:"email"`
			}{},
			expectedFields: []string{"Email"},
		},
		{
			name: "field with min constraint",
			structType: struct {
				Age int `pedantigo:"min=18"`
			}{},
			expectedFields: []string{"Age"},
		},
		{
			name: "field with multiple constraints",
			structType: struct {
				Email string `pedantigo:"required,email,min=5,max=100"`
			}{},
			expectedFields: []string{"Email"},
		},
		{
			name: "field without pedantigo tag",
			structType: struct {
				Name string `json:"name"`
			}{},
			expectedFields: []string{"name"},
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			for _, fieldName := range tt.expectedFields {
				_, exists := deserializers[fieldName]
				assert.True(t, exists, "expected deserializer for field %q", fieldName)
			}
		})
	}
}

// TestBuildFieldDeserializers_EdgeCases tests various edge cases
func TestBuildFieldDeserializers_EdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		structType         any
		expectedFieldCount int
		expectedFields     []string
	}{
		{
			name: "embedded struct",
			structType: struct {
				Name   string
				Nested struct {
					Value string
				}
			}{},
			expectedFieldCount: 2,
			expectedFields:     []string{"Name", "Nested"},
		},
		{
			name: "field with multiple json tag options",
			structType: struct {
				Name string `json:"user_name,omitempty"`
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"user_name"},
		},
		{
			name: "field with type alias",
			structType: struct {
				CustomType int `json:"custom"`
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"custom"},
		},
		{
			name: "interface{} field",
			structType: struct {
				Data interface{}
			}{},
			expectedFieldCount: 1,
			expectedFields:     []string{"Data"},
		},
	}

	opts := BuilderOptions{StrictMissingFields: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			deserializers := BuildFieldDeserializers(typ, opts, nil, nil)

			assert.Equal(t, tt.expectedFieldCount, len(deserializers), "deserializer count mismatch")

			for _, fieldName := range tt.expectedFields {
				_, exists := deserializers[fieldName]
				assert.True(t, exists, "expected deserializer for field %q", fieldName)
			}
		})
	}
}

// TestValidateDefaultMethod_ValidMethod tests valid method validation
func TestValidateDefaultMethod_ValidMethod(t *testing.T) {
	tests := []struct {
		name        string
		structType  any
		methodName  string
		fieldType   reflect.Type
		shouldErr   bool
		errContains string
	}{
		{
			name:        "method does not exist",
			structType:  struct{ Age int }{},
			methodName:  "NonExistentMethod",
			fieldType:   reflect.TypeOf(int(0)),
			shouldErr:   true,
			errContains: "method NonExistentMethod not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			err := ValidateDefaultMethod(typ, tt.methodName, tt.fieldType)

			if tt.shouldErr {
				require.Error(t, err, "expected error, got nil")
				assert.Contains(t, err.Error(), tt.errContains, "error message should contain expected text")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

// TestValidateDefaultMethod_SignatureValidation tests method signature validation
func TestValidateDefaultMethod_SignatureValidation(t *testing.T) {
	tests := []struct {
		name       string
		structType any
		methodName string
		fieldType  reflect.Type
		shouldErr  bool
		contains   string
	}{
		{
			name:       "method with wrong return count",
			structType: struct{}{},
			methodName: "GetValue",
			fieldType:  reflect.TypeOf(""),
			shouldErr:  true,
			contains:   "not found", // Empty struct has no methods, so error is "not found"
		},
		{
			name:       "method does not exist",
			structType: struct{}{},
			methodName: "DoesNotExist",
			fieldType:  reflect.TypeOf(""),
			shouldErr:  true,
			contains:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			err := ValidateDefaultMethod(typ, tt.methodName, tt.fieldType)

			if tt.shouldErr {
				require.Error(t, err, "expected error, got nil")
				assert.Contains(t, err.Error(), tt.contains, "error message should contain expected text")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

// Test types for ValidateDefaultMethod tests
type ValidMethodStruct struct{ Name string }

func (v *ValidMethodStruct) GetName() (string, error) { return "test", nil }

type WrongInputArgsStruct struct{ Age int }

func (w *WrongInputArgsStruct) GetAge(arg int) (int, error) { return 0, nil }

type WrongOutputCountStruct struct{ Score float64 }

func (w *WrongOutputCountStruct) GetScore() float64 { return 0.0 }

type WrongFirstReturnStruct struct{ ID int }

func (w *WrongFirstReturnStruct) GetID() (string, error) { return "", nil } // Returns string instead of int

type WrongSecondReturnStruct struct{ Count int }

func (w *WrongSecondReturnStruct) GetCount() (int, string) { return 0, "" } // Returns string instead of error

type ThreeReturnsStruct struct{ Data string }

func (t *ThreeReturnsStruct) GetData() (string, error, bool) { return "", nil, false }

// TestValidateDefaultMethod_ComprehensiveCases tests all validation paths
func TestValidateDefaultMethod_ComprehensiveCases(t *testing.T) {

	tests := []struct {
		name        string
		structType  any
		methodName  string
		fieldType   reflect.Type
		shouldErr   bool
		errContains string
	}{
		{
			name:        "valid method signature",
			structType:  ValidMethodStruct{},
			methodName:  "GetName",
			fieldType:   reflect.TypeOf(""),
			shouldErr:   false,
			errContains: "",
		},
		{
			name:        "method with input arguments",
			structType:  WrongInputArgsStruct{},
			methodName:  "GetAge",
			fieldType:   reflect.TypeOf(0),
			shouldErr:   true,
			errContains: "should take no arguments",
		},
		{
			name:        "method with single return value",
			structType:  WrongOutputCountStruct{},
			methodName:  "GetScore",
			fieldType:   reflect.TypeOf(0.0),
			shouldErr:   true,
			errContains: "should return (value, error), got 1 return values",
		},
		{
			name:        "method with three return values",
			structType:  ThreeReturnsStruct{},
			methodName:  "GetData",
			fieldType:   reflect.TypeOf(""),
			shouldErr:   true,
			errContains: "should return (value, error), got 3 return values",
		},
		{
			name:        "method with wrong first return type",
			structType:  WrongFirstReturnStruct{},
			methodName:  "GetID",
			fieldType:   reflect.TypeOf(0), // Expects int, but method returns string
			shouldErr:   true,
			errContains: "should return int as first value, got string",
		},
		{
			name:        "method with wrong second return type",
			structType:  WrongSecondReturnStruct{},
			methodName:  "GetCount",
			fieldType:   reflect.TypeOf(0),
			shouldErr:   true,
			errContains: "should return error as second value, got string",
		},
		{
			name:        "method does not exist",
			structType:  ValidMethodStruct{},
			methodName:  "NonExistent",
			fieldType:   reflect.TypeOf(""),
			shouldErr:   true,
			errContains: "method NonExistent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.structType)
			err := ValidateDefaultMethod(typ, tt.methodName, tt.fieldType)

			if tt.shouldErr {
				require.Error(t, err, "expected error, got nil")
				assert.Contains(t, err.Error(), tt.errContains, "error message should contain expected text")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

// Helper functions

func getMapKeys(m map[string]FieldDeserializer) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
