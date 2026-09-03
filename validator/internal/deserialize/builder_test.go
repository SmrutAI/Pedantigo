package deserialize

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateDefaultMethod_ValidMethod tests valid method validation.
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

// TestValidateDefaultMethod_SignatureValidation tests method signature validation.
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
// ValidMethodStruct represents the data structure.
type ValidMethodStruct struct{ Name string }

func (v *ValidMethodStruct) GetName() (string, error) { return "test", nil }

type WrongInputArgsStruct struct{ Age int }

func (w *WrongInputArgsStruct) GetAge(arg int) (int, error) { return 0, nil }

type WrongOutputCountStruct struct{ Score float64 }

func (w *WrongOutputCountStruct) GetScore() float64 { return 0.0 }

type WrongFirstReturnStruct struct{ ID int }

func (w *WrongFirstReturnStruct) GetID() (string, error) { return "", nil } // Returns string instead of int

type WrongSecondReturnStruct struct{ Count int }

func (w *WrongSecondReturnStruct) GetCount() (count int, errMsg string) { return 0, "" } // Returns string instead of error

type ThreeReturnsStruct struct{ Data string }

func (t *ThreeReturnsStruct) GetData() (string, error, bool) { return "", nil, false }

// TestValidateDefaultMethod_ComprehensiveCases tests all validation paths.
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

// TestApplyStringTransformations tests string transformation logic.
func TestApplyStringTransformations(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		transforms StringTransformations
		want       string
	}{
		{
			name:       "strip_whitespace",
			input:      "  hello  ",
			transforms: StringTransformations{StripWhitespace: true},
			want:       "hello",
		},
		{
			name:       "to_upper",
			input:      "hello",
			transforms: StringTransformations{ToUpper: true},
			want:       "HELLO",
		},
		{
			name:       "to_lower",
			input:      "HELLO",
			transforms: StringTransformations{ToLower: true},
			want:       "hello",
		},
		{
			name:       "strip_whitespace and to_lower",
			input:      "  HELLO WORLD  ",
			transforms: StringTransformations{StripWhitespace: true, ToLower: true},
			want:       "hello world",
		},
		{
			name:       "strip_whitespace and to_upper",
			input:      "  hello world  ",
			transforms: StringTransformations{StripWhitespace: true, ToUpper: true},
			want:       "HELLO WORLD",
		},
		{
			name:       "to_lower takes precedence over to_upper",
			input:      "Hello",
			transforms: StringTransformations{ToLower: true, ToUpper: true},
			want:       "hello",
		},
		{
			name:       "no transformations",
			input:      "  Hello  ",
			transforms: StringTransformations{},
			want:       "  Hello  ",
		},
	}

	for _, tt := range tests {
		tt := tt // avoid G601
		t.Run(tt.name, func(t *testing.T) {
			value := reflect.ValueOf(&tt.input).Elem()
			applyStringTransformations(value, tt.transforms)
			assert.Equal(t, tt.want, tt.input)
		})
	}

	// Test nil pointer case (should not panic)
	t.Run("nil pointer does not panic", func(t *testing.T) {
		var nilPtr *string
		value := reflect.ValueOf(&nilPtr).Elem()
		assert.NotPanics(t, func() {
			applyStringTransformations(value, StringTransformations{StripWhitespace: true})
		})
	})

	// Test pointer to string
	t.Run("pointer to string with strip_whitespace", func(t *testing.T) {
		str := "  hello  "
		ptr := &str
		value := reflect.ValueOf(&ptr).Elem()
		applyStringTransformations(value, StringTransformations{StripWhitespace: true})
		assert.Equal(t, "hello", *ptr)
	})

	// Test non-string field (should be no-op)
	t.Run("non-string field is no-op", func(t *testing.T) {
		num := 42
		value := reflect.ValueOf(&num).Elem()
		applyStringTransformations(value, StringTransformations{StripWhitespace: true, ToUpper: true})
		assert.Equal(t, 42, num) // Should remain unchanged
	})
}
