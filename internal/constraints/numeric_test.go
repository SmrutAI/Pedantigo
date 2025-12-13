package constraints

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMaxConstraint tests maxConstraint.Validate() for numeric values.
func TestMaxConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		max     int
		wantErr bool
	}{
		// Valid cases - below max
		{name: "int below max", value: 50, max: 100, wantErr: false},
		{name: "int at max", value: 100, max: 100, wantErr: false},
		{name: "int zero with positive max", value: 0, max: 100, wantErr: false},
		{name: "int negative with positive max", value: -50, max: 100, wantErr: false},

		// Invalid cases - exceeds max
		{name: "int exceeds max", value: 150, max: 100, wantErr: true},
		{name: "int exceeds negative max", value: -3, max: -5, wantErr: true},

		// Float values
		{name: "float below max", value: 3.5, max: 10, wantErr: false},
		{name: "float at max", value: 10.0, max: 10, wantErr: false},
		{name: "float exceeds max", value: 10.1, max: 10, wantErr: true},

		// Negative max
		{name: "negative value with negative max (valid)", value: -10, max: -5, wantErr: false},
		{name: "negative value with negative max (invalid)", value: -3, max: -5, wantErr: true},

		// Different numeric types
		{name: "int8 below max", value: int8(50), max: 100, wantErr: false},
		{name: "int8 exceeds max", value: int8(127), max: 100, wantErr: true},
		{name: "int16 below max", value: int16(50), max: 100, wantErr: false},
		{name: "int32 below max", value: int32(50), max: 100, wantErr: false},
		{name: "int64 below max", value: int64(50), max: 100, wantErr: false},
		{name: "uint below max", value: uint(50), max: 100, wantErr: false},
		{name: "uint exceeds max", value: uint(150), max: 100, wantErr: true},
		{name: "uint8 below max", value: uint8(50), max: 100, wantErr: false},
		{name: "uint16 below max", value: uint16(50), max: 100, wantErr: false},
		{name: "uint32 below max", value: uint32(50), max: 100, wantErr: false},
		{name: "uint64 below max", value: uint64(50), max: 100, wantErr: false},
		{name: "float32 below max", value: float32(3.5), max: 10, wantErr: false},
		{name: "float32 exceeds max", value: float32(10.5), max: 10, wantErr: true},
		{name: "float64 below max", value: float64(3.5), max: 10, wantErr: false},
		{name: "float64 exceeds max", value: float64(10.5), max: 10, wantErr: true},

		// String (length check)
		{name: "string below max length", value: "hello", max: 10, wantErr: false},
		{name: "string at max length", value: "helloworld", max: 10, wantErr: false},
		{name: "string exceeds max length", value: "hello world!", max: 10, wantErr: true},
		{name: "empty string with max length", value: "", max: 10, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), max: 100, wantErr: false},

		// Invalid type
		{name: "invalid type - bool", value: true, max: 100, wantErr: true},
		{name: "invalid type - slice", value: []int{1, 2, 3}, max: 100, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := maxConstraint{max: tt.max}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestMaxLengthConstraint tests maxLengthConstraint.Validate() for strings.
func TestMaxLengthConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		maxLength int
		wantErr   bool
	}{
		// Valid cases
		{name: "string below max length", value: "hello", maxLength: 10, wantErr: false},
		{name: "string at max length", value: "helloworld", maxLength: 10, wantErr: false},
		{name: "empty string", value: "", maxLength: 10, wantErr: false},
		{name: "single char at max length", value: "a", maxLength: 1, wantErr: false},

		// Invalid cases - exceeds max length
		{name: "string exceeds max length", value: "hello world!", maxLength: 10, wantErr: true},
		{name: "string far exceeds max length", value: "this is a very long string", maxLength: 5, wantErr: true},
		{name: "string one char over max", value: "ab", maxLength: 1, wantErr: true},

		// Edge cases
		{name: "max length zero with empty string", value: "", maxLength: 0, wantErr: false},
		{name: "max length zero with non-empty string", value: "a", maxLength: 0, wantErr: true},
		{name: "large max length", value: "hello", maxLength: 1000, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), maxLength: 10, wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, maxLength: 10, wantErr: true},
		{name: "invalid type - bool", value: true, maxLength: 10, wantErr: true},
		{name: "invalid type - slice", value: []string{"a", "b"}, maxLength: 10, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := maxLengthConstraint{maxLength: tt.maxLength}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestLtConstraint tests ltConstraint.Validate() for < threshold.
func TestLtConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value < threshold
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: false},
		{name: "int at boundary", value: 19, threshold: 20, wantErr: false},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: false},
		{name: "negative less than zero", value: -10, threshold: 0, wantErr: false},
		{name: "negative less than positive", value: -5, threshold: 10, wantErr: false},

		// Invalid cases - value >= threshold
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: true},
		{name: "int exceeds threshold", value: 25, threshold: 20, wantErr: true},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: true},
		{name: "float exceeds threshold", value: 6.5, threshold: 5.0, wantErr: true},

		// Different numeric types
		{name: "int8 less than threshold", value: int8(10), threshold: 20, wantErr: false},
		{name: "int8 exceeds threshold", value: int8(25), threshold: 20, wantErr: true},
		{name: "uint less than threshold", value: uint(10), threshold: 20, wantErr: false},
		{name: "uint exceeds threshold", value: uint(25), threshold: 20, wantErr: true},
		{name: "float32 less than threshold", value: float32(3.14), threshold: 5.0, wantErr: false},
		{name: "float32 exceeds threshold", value: float32(6.5), threshold: 5.0, wantErr: true},

		// Zero and near-zero thresholds
		{name: "negative less than zero threshold", value: -10, threshold: 0, wantErr: false},
		{name: "zero equals zero threshold", value: 0, threshold: 0, wantErr: true},
		{name: "positive exceeds zero threshold", value: 5, threshold: 0, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
		{name: "invalid type - bool", value: true, threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ltConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestLeConstraint tests leConstraint.Validate() for <= threshold.
func TestLeConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value <= threshold
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: false},
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: false},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: false},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: false},
		{name: "negative less than zero", value: -10, threshold: 0, wantErr: false},
		{name: "zero equals zero", value: 0, threshold: 0, wantErr: false},

		// Invalid cases - value > threshold
		{name: "int exceeds threshold", value: 25, threshold: 20, wantErr: true},
		{name: "float exceeds threshold", value: 6.5, threshold: 5.0, wantErr: true},
		{name: "positive exceeds zero", value: 5, threshold: 0, wantErr: true},

		// Different numeric types
		{name: "int8 equals threshold", value: int8(20), threshold: 20, wantErr: false},
		{name: "int8 exceeds threshold", value: int8(25), threshold: 20, wantErr: true},
		{name: "uint less than threshold", value: uint(10), threshold: 20, wantErr: false},
		{name: "uint equals threshold", value: uint(20), threshold: 20, wantErr: false},
		{name: "uint exceeds threshold", value: uint(25), threshold: 20, wantErr: true},
		{name: "float32 equals threshold", value: float32(5.0), threshold: 5.0, wantErr: false},
		{name: "float32 exceeds threshold", value: float32(6.5), threshold: 5.0, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
		{name: "invalid type - bool", value: true, threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := leConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestMinConstraint tests minConstraint.Validate() for numeric values
// Added to ensure comprehensive coverage of all constraints mentioned
// TestMinConstraint tests MinConstraint validation.
func TestMinConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		min     int
		wantErr bool
	}{
		// Valid cases - above min
		{name: "int above min", value: 50, min: 10, wantErr: false},
		{name: "int at min", value: 10, min: 10, wantErr: false},
		{name: "int zero with negative min", value: 0, min: -10, wantErr: false},

		// Invalid cases - below min
		{name: "int below min", value: 5, min: 10, wantErr: true},
		{name: "int far below min", value: -50, min: 10, wantErr: true},

		// Float values
		{name: "float above min", value: 5.0, min: 3, wantErr: false},
		{name: "float at min", value: 3.0, min: 3, wantErr: false},
		{name: "float below min", value: 2.5, min: 3, wantErr: true},

		// Different numeric types
		{name: "int8 above min", value: int8(50), min: 10, wantErr: false},
		{name: "uint above min", value: uint(50), min: 10, wantErr: false},
		{name: "float32 above min", value: float32(5.0), min: 3, wantErr: false},

		// String (length check)
		{name: "string above min length", value: "hello", min: 3, wantErr: false},
		{name: "string at min length", value: "hel", min: 3, wantErr: false},
		{name: "string below min length", value: "hi", min: 3, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), min: 10, wantErr: false},

		// Invalid types
		{name: "invalid type - bool", value: true, min: 10, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := minConstraint{min: tt.min}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestGtConstraint tests gtConstraint.Validate() for > threshold.
func TestGtConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value > threshold
		{name: "int greater than threshold", value: 25, threshold: 20, wantErr: false},
		{name: "float greater than threshold", value: 6.5, threshold: 5.0, wantErr: false},
		{name: "negative greater than negative", value: -5, threshold: -10, wantErr: false},

		// Invalid cases - value <= threshold
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: true},
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: true},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: true},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: true},

		// Different numeric types
		{name: "int8 greater than threshold", value: int8(25), threshold: 20, wantErr: false},
		{name: "uint greater than threshold", value: uint(25), threshold: 20, wantErr: false},
		{name: "float32 greater than threshold", value: float32(6.5), threshold: 5.0, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := gtConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestGeConstraint tests geConstraint.Validate() for >= threshold.
func TestGeConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		threshold float64
		wantErr   bool
	}{
		// Valid cases - value >= threshold
		{name: "int greater than threshold", value: 25, threshold: 20, wantErr: false},
		{name: "int equals threshold", value: 20, threshold: 20, wantErr: false},
		{name: "float greater than threshold", value: 6.5, threshold: 5.0, wantErr: false},
		{name: "float equals threshold", value: 5.0, threshold: 5.0, wantErr: false},

		// Invalid cases - value < threshold
		{name: "int less than threshold", value: 10, threshold: 20, wantErr: true},
		{name: "float less than threshold", value: 3.14, threshold: 5.0, wantErr: true},

		// Different numeric types
		{name: "int8 greater than threshold", value: int8(25), threshold: 20, wantErr: false},
		{name: "int8 equals threshold", value: int8(20), threshold: 20, wantErr: false},
		{name: "uint greater than threshold", value: uint(25), threshold: 20, wantErr: false},
		{name: "float32 equals threshold", value: float32(5.0), threshold: 5.0, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*int)(nil), threshold: 20, wantErr: false},

		// Invalid types
		{name: "invalid type - string", value: "hello", threshold: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := geConstraint{threshold: tt.threshold}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestMinLengthConstraint tests minLengthConstraint.Validate() for strings.
func TestMinLengthConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		minLength int
		wantErr   bool
	}{
		// Valid cases
		{name: "string above min length", value: "hello", minLength: 3, wantErr: false},
		{name: "string at min length", value: "hel", minLength: 3, wantErr: false},
		{name: "long string", value: "this is a long string", minLength: 5, wantErr: false},
		{name: "single char at min", value: "a", minLength: 1, wantErr: false},

		// Invalid cases - below min length
		{name: "string below min length", value: "hi", minLength: 3, wantErr: true},
		{name: "empty string below min", value: "", minLength: 1, wantErr: true},

		// Edge cases
		{name: "min length zero with empty string", value: "", minLength: 0, wantErr: false},
		{name: "min length zero with non-empty", value: "a", minLength: 0, wantErr: false},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), minLength: 3, wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, minLength: 3, wantErr: true},
		{name: "invalid type - bool", value: true, minLength: 3, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := minLengthConstraint{minLength: tt.minLength}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestBuildMaxConstraint tests buildMaxConstraint builder function.
func TestBuildMaxConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldType reflect.Type
		wantType  string
		wantOk    bool
	}{
		// String type should create maxLengthConstraint
		{name: "string field", value: "10", fieldType: reflect.TypeOf(""), wantType: "maxLengthConstraint", wantOk: true},
		{name: "slice field", value: "5", fieldType: reflect.TypeOf([]int{}), wantType: "maxLengthConstraint", wantOk: true},
		{name: "array field", value: "3", fieldType: reflect.TypeOf([3]int{}), wantType: "maxLengthConstraint", wantOk: true},

		// Numeric types should create maxConstraint
		{name: "int field", value: "100", fieldType: reflect.TypeOf(0), wantType: "maxConstraint", wantOk: true},
		{name: "int64 field", value: "100", fieldType: reflect.TypeOf(int64(0)), wantType: "maxConstraint", wantOk: true},
		{name: "uint field", value: "100", fieldType: reflect.TypeOf(uint(0)), wantType: "maxConstraint", wantOk: true},
		{name: "float64 field", value: "100", fieldType: reflect.TypeOf(float64(0)), wantType: "maxConstraint", wantOk: true},

		// Pointer types should unwrap
		{name: "pointer to string", value: "10", fieldType: reflect.TypeOf((*string)(nil)), wantType: "maxLengthConstraint", wantOk: true},
		{name: "pointer to int", value: "100", fieldType: reflect.TypeOf((*int)(nil)), wantType: "maxConstraint", wantOk: true},

		// Invalid value should fail
		{name: "invalid value", value: "not-a-number", fieldType: reflect.TypeOf(0), wantType: "", wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, ok := buildMaxConstraint(tt.value, tt.fieldType)

			assert.Equal(t, tt.wantOk, ok)

			if !ok {
				return
			}

			switch tt.wantType {
			case "maxLengthConstraint":
				_, isMaxLength := constraint.(maxLengthConstraint)
				assert.True(t, isMaxLength, "expected maxLengthConstraint, got %T", constraint)
			case "maxConstraint":
				_, isMax := constraint.(maxConstraint)
				assert.True(t, isMax, "expected maxConstraint, got %T", constraint)
			}
		})
	}
}

// TestPositiveConstraint tests positiveConstraint.Validate() for positive number validation.
func TestPositiveConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - positive numbers
		{name: "positive int", value: 1, wantErr: false},
		{name: "large positive int", value: 1000000, wantErr: false},
		{name: "positive float", value: 0.1, wantErr: false},
		{name: "positive float64", value: float64(3.14), wantErr: false},
		{name: "positive uint", value: uint(5), wantErr: false},
		{name: "positive int8", value: int8(127), wantErr: false},
		{name: "positive int64", value: int64(9999), wantErr: false},

		// Invalid cases - zero and negative
		{name: "zero int", value: 0, wantErr: true},
		{name: "zero float", value: 0.0, wantErr: true},
		{name: "negative int", value: -1, wantErr: true},
		{name: "negative float", value: -0.5, wantErr: true},
		{name: "large negative", value: -1000000, wantErr: true},

		// Edge cases
		{name: "nil pointer", value: (*int)(nil), wantErr: false},
		{name: "pointer to positive", value: intPtr(5), wantErr: false},
		{name: "pointer to zero", value: intPtr(0), wantErr: true},
		{name: "pointer to negative", value: intPtr(-5), wantErr: true},

		// Invalid types
		{name: "invalid type - string", value: "123", wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := positiveConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// Helper function for int pointers.
func intPtr(i int) *int {
	return &i
}

// TestNegativeConstraint tests negativeConstraint.Validate() for negative number validation.
func TestNegativeConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - negative numbers
		{name: "negative int", value: -1, wantErr: false},
		{name: "large negative int", value: -1000000, wantErr: false},
		{name: "negative float", value: -0.1, wantErr: false},
		{name: "negative float64", value: float64(-3.14), wantErr: false},
		{name: "negative int8", value: int8(-127), wantErr: false},
		{name: "negative int64", value: int64(-9999), wantErr: false},

		// Invalid cases - zero and positive
		{name: "zero int", value: 0, wantErr: true},
		{name: "zero float", value: 0.0, wantErr: true},
		{name: "positive int", value: 1, wantErr: true},
		{name: "positive float", value: 0.5, wantErr: true},
		{name: "large positive", value: 1000000, wantErr: true},
		{name: "positive uint", value: uint(5), wantErr: true},

		// Edge cases
		{name: "nil pointer", value: (*int)(nil), wantErr: false},
		{name: "pointer to negative", value: intPtr(-5), wantErr: false},
		{name: "pointer to zero", value: intPtr(0), wantErr: true},
		{name: "pointer to positive", value: intPtr(5), wantErr: true},

		// Invalid types
		{name: "invalid type - string", value: "-123", wantErr: true},
		{name: "invalid type - bool", value: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := negativeConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

func TestMultipleOfConstraint(t *testing.T) {
	tests := []struct {
		name    string
		factor  float64
		value   any
		wantErr bool
	}{
		// Valid cases - exact multiples (integer factor)
		{name: "10 is multiple of 5", factor: 5, value: 10, wantErr: false},
		{name: "15 is multiple of 5", factor: 5, value: 15, wantErr: false},
		{name: "0 is multiple of 5", factor: 5, value: 0, wantErr: false},
		{name: "100 is multiple of 10", factor: 10, value: 100, wantErr: false},
		{name: "-10 is multiple of 5", factor: 5, value: -10, wantErr: false},
		{name: "negative multiple of negative", factor: -5, value: -10, wantErr: false},

		// Valid cases - float factor
		{name: "1.5 is multiple of 0.5", factor: 0.5, value: 1.5, wantErr: false},
		{name: "3.0 is multiple of 0.5", factor: 0.5, value: 3.0, wantErr: false},
		{name: "0.25 is multiple of 0.25", factor: 0.25, value: 0.25, wantErr: false},

		// Invalid cases - not exact multiples
		{name: "7 is not multiple of 5", factor: 5, value: 7, wantErr: true},
		{name: "11 is not multiple of 3", factor: 3, value: 11, wantErr: true},
		{name: "1.6 is not multiple of 0.5", factor: 0.5, value: 1.6, wantErr: true},
		{name: "10 is not multiple of 3", factor: 3, value: 10, wantErr: true},

		// Various numeric types
		{name: "int64 multiple", factor: 5, value: int64(25), wantErr: false},
		{name: "int32 multiple", factor: 5, value: int32(20), wantErr: false},
		{name: "uint multiple", factor: 5, value: uint(15), wantErr: false},
		{name: "float32 multiple", factor: 0.5, value: float32(2.5), wantErr: false},
		{name: "float64 multiple", factor: 0.5, value: float64(3.5), wantErr: false},

		// Edge cases
		{name: "nil pointer", factor: 5, value: (*int)(nil), wantErr: false},
		{name: "pointer to multiple", factor: 5, value: intPtr(10), wantErr: false},
		{name: "pointer to non-multiple", factor: 5, value: intPtr(7), wantErr: true},

		// Invalid types
		{name: "invalid type - string", factor: 5, value: "10", wantErr: true},
		{name: "invalid type - bool", factor: 5, value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := multipleOfConstraint{factor: tt.factor}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

func TestMaxDigitsConstraint(t *testing.T) {
	tests := []struct {
		name      string
		maxDigits int
		value     any
		wantErr   bool
	}{
		// Valid cases - within max digits
		{name: "1 digit within max 3", maxDigits: 3, value: 5, wantErr: false},
		{name: "2 digits within max 3", maxDigits: 3, value: 42, wantErr: false},
		{name: "3 digits at max 3", maxDigits: 3, value: 123, wantErr: false},
		{name: "zero within max", maxDigits: 3, value: 0, wantErr: false},
		{name: "negative 2 digits within max 3", maxDigits: 3, value: -42, wantErr: false},
		{name: "float 3.14 within max 5", maxDigits: 5, value: 3.14, wantErr: false},
		{name: "1 integer digit within max 1", maxDigits: 1, value: 5, wantErr: false},

		// Invalid cases - exceeds max digits
		{name: "4 digits exceeds max 3", maxDigits: 3, value: 1234, wantErr: true},
		{name: "5 digits exceeds max 3", maxDigits: 3, value: 12345, wantErr: true},
		{name: "6 digits exceeds max 5", maxDigits: 5, value: 123456, wantErr: true},
		{name: "negative 4 digits exceeds max 3", maxDigits: 3, value: -1234, wantErr: true},
		{name: "float 123.45 has 5 digits exceeds max 3", maxDigits: 3, value: 123.45, wantErr: true},

		// Various numeric types
		{name: "int64 within max", maxDigits: 5, value: int64(12345), wantErr: false},
		{name: "int64 exceeds max", maxDigits: 3, value: int64(12345), wantErr: true},
		{name: "uint within max", maxDigits: 3, value: uint(123), wantErr: false},
		{name: "uint exceeds max", maxDigits: 2, value: uint(123), wantErr: true},
		{name: "float32 within max", maxDigits: 4, value: float32(12.5), wantErr: false},
		{name: "float64 within max", maxDigits: 6, value: float64(123.456), wantErr: false},

		// Edge cases
		{name: "nil pointer", maxDigits: 3, value: (*int)(nil), wantErr: false},
		{name: "pointer to valid", maxDigits: 3, value: intPtr(123), wantErr: false},
		{name: "pointer to invalid", maxDigits: 3, value: intPtr(1234), wantErr: true},

		// Invalid types
		{name: "invalid type - string", maxDigits: 3, value: "123", wantErr: true},
		{name: "invalid type - bool", maxDigits: 3, value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := maxDigitsConstraint{maxDigits: tt.maxDigits}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

func TestDecimalPlacesConstraint(t *testing.T) {
	tests := []struct {
		name      string
		maxPlaces int
		value     any
		wantErr   bool
	}{
		// Valid cases - within max decimal places
		{name: "integer no decimals", maxPlaces: 2, value: 123, wantErr: false},
		{name: "float with 1 decimal within max 2", maxPlaces: 2, value: 12.3, wantErr: false},
		{name: "float with 2 decimals at max 2", maxPlaces: 2, value: 12.34, wantErr: false},
		{name: "zero value", maxPlaces: 2, value: 0, wantErr: false},
		{name: "negative with 1 decimal", maxPlaces: 2, value: -12.3, wantErr: false},
		{name: "float with max 0 decimals no decimal", maxPlaces: 0, value: 123.0, wantErr: false},

		// Invalid cases - exceeds max decimal places
		{name: "float with 3 decimals exceeds max 2", maxPlaces: 2, value: 12.345, wantErr: true},
		{name: "float with 4 decimals exceeds max 2", maxPlaces: 2, value: 12.3456, wantErr: true},
		{name: "float with 1 decimal exceeds max 0", maxPlaces: 0, value: 12.3, wantErr: true},
		{name: "negative with 3 decimals exceeds max 2", maxPlaces: 2, value: -12.345, wantErr: true},

		// Various numeric types
		{name: "int64 no decimals", maxPlaces: 2, value: int64(12345), wantErr: false},
		{name: "uint no decimals", maxPlaces: 2, value: uint(123), wantErr: false},
		{name: "float32 within max", maxPlaces: 2, value: float32(12.5), wantErr: false},
		{name: "float64 within max", maxPlaces: 3, value: float64(12.345), wantErr: false},
		{name: "float64 exceeds max", maxPlaces: 2, value: float64(12.345), wantErr: true},

		// Edge cases
		{name: "nil pointer", maxPlaces: 2, value: (*float64)(nil), wantErr: false},
		{name: "pointer to valid", maxPlaces: 2, value: func() *float64 { f := 12.34; return &f }(), wantErr: false},
		{name: "pointer to invalid", maxPlaces: 2, value: func() *float64 { f := 12.345; return &f }(), wantErr: true},

		// Invalid types
		{name: "invalid type - string", maxPlaces: 2, value: "12.34", wantErr: true},
		{name: "invalid type - bool", maxPlaces: 2, value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := decimalPlacesConstraint{maxPlaces: tt.maxPlaces}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestDisallowInfNanConstraint tests disallowInfNanConstraint.Validate() for Inf/NaN rejection.
func TestDisallowInfNanConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases - normal numbers
		{name: "positive int", value: 42, wantErr: false},
		{name: "negative int", value: -42, wantErr: false},
		{name: "zero int", value: 0, wantErr: false},
		{name: "float64 normal", value: 3.14, wantErr: false},
		{name: "float32 normal", value: float32(3.14), wantErr: false},
		{name: "negative float", value: -123.456, wantErr: false},
		{name: "zero float", value: 0.0, wantErr: false},
		{name: "very large float", value: 1e308, wantErr: false},
		{name: "very small float", value: 1e-308, wantErr: false},

		// Invalid cases - Inf
		{name: "positive infinity float64", value: math.Inf(1), wantErr: true},
		{name: "negative infinity float64", value: math.Inf(-1), wantErr: true},
		{name: "positive infinity float32", value: float32(math.Inf(1)), wantErr: true},
		{name: "negative infinity float32", value: float32(math.Inf(-1)), wantErr: true},

		// Invalid cases - NaN
		{name: "NaN float64", value: math.NaN(), wantErr: true},
		{name: "NaN float32", value: float32(math.NaN()), wantErr: true},

		// Edge cases - nil/pointer
		{name: "nil pointer", value: (*float64)(nil), wantErr: false},
		{name: "pointer to normal", value: func() *float64 { f := 3.14; return &f }(), wantErr: false},
		{name: "pointer to Inf", value: func() *float64 { f := math.Inf(1); return &f }(), wantErr: true},
		{name: "pointer to negative Inf", value: func() *float64 { f := math.Inf(-1); return &f }(), wantErr: true},
		{name: "pointer to NaN", value: func() *float64 { f := math.NaN(); return &f }(), wantErr: true},

		// Non-float types (should pass - constraint only applies to floats)
		{name: "int type", value: int(100), wantErr: false},
		{name: "int8 type", value: int8(100), wantErr: false},
		{name: "int16 type", value: int16(100), wantErr: false},
		{name: "int32 type", value: int32(100), wantErr: false},
		{name: "int64 type", value: int64(100), wantErr: false},
		{name: "uint type", value: uint(100), wantErr: false},
		{name: "uint8 type", value: uint8(100), wantErr: false},
		{name: "uint16 type", value: uint16(100), wantErr: false},
		{name: "uint32 type", value: uint32(100), wantErr: false},
		{name: "uint64 type", value: uint64(100), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := disallowInfNanConstraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}
