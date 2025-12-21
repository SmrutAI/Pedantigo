package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnumConstraint tests enumConstraint.Validate() for allowed values.
func TestEnumConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		values  []string
		wantErr bool
	}{
		// String values
		{name: "string in enum", value: "active", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string in enum - first value", value: "pending", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string in enum - last value", value: "completed", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string not in enum", value: "invalid", values: []string{"pending", "active", "completed"}, wantErr: true},
		{name: "string case sensitive", value: "Active", values: []string{"pending", "active", "completed"}, wantErr: true},

		// Int values
		{name: "int in enum", value: 1, values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int in enum - zero", value: 0, values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int not in enum", value: 5, values: []string{"0", "1", "2"}, wantErr: true},
		{name: "int8 in enum", value: int8(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int16 in enum", value: int16(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int32 in enum", value: int32(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int64 in enum", value: int64(1), values: []string{"0", "1", "2"}, wantErr: false},

		// Uint values
		{name: "uint in enum", value: uint(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint not in enum", value: uint(5), values: []string{"0", "1", "2"}, wantErr: true},
		{name: "uint8 in enum", value: uint8(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint16 in enum", value: uint16(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint32 in enum", value: uint32(1), values: []string{"0", "1", "2"}, wantErr: false},
		{name: "uint64 in enum", value: uint64(1), values: []string{"0", "1", "2"}, wantErr: false},

		// Float values
		{name: "float in enum", value: 3.14, values: []string{"2.71", "3.14", "2.0"}, wantErr: false},
		{name: "float not in enum", value: 1.5, values: []string{"2.71", "3.14", "2.0"}, wantErr: true},
		{name: "float32 in enum", value: float32(2.5), values: []string{"2.5", "1.5"}, wantErr: false},
		{name: "float64 in enum", value: float64(3.14), values: []string{"3.14", "2.71"}, wantErr: false},

		// Bool values
		{name: "bool true in enum", value: true, values: []string{"true", "false"}, wantErr: false},
		{name: "bool false in enum", value: false, values: []string{"true", "false"}, wantErr: false},

		// Empty enum
		{name: "value not in empty enum", value: "test", values: []string{}, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), values: []string{"a", "b"}, wantErr: false},

		// Invalid types
		{name: "invalid type - slice", value: []string{"a"}, values: []string{"a", "b"}, wantErr: true},
		{name: "invalid type - map", value: map[string]int{"a": 1}, values: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := enumConstraint{values: tt.values}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestDefaultConstraint tests defaultConstraint.Validate() - no-op validator.
func TestDefaultConstraint(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "string value", value: "test"},
		{name: "int value", value: 42},
		{name: "float value", value: 3.14},
		{name: "bool value", value: true},
		{name: "nil value", value: nil},
		{name: "empty string", value: ""},
		{name: "zero int", value: 0},
		{name: "slice", value: []int{1, 2, 3}},
		{name: "map", value: map[string]int{"a": 1}},
	}

	constraint := defaultConstraint{value: "default"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// defaultConstraint always returns nil - it's a no-op validator
			err := constraint.Validate(tt.value)
			assert.NoError(t, err)
		})
	}
}

// TestEnumCIConstraint tests enumCIConstraint.Validate() for case-insensitive matching.
func TestEnumCIConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		values  []string
		wantErr bool
	}{
		// String values - case insensitive
		{name: "string exact match", value: "active", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string uppercase match", value: "ACTIVE", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string mixed case match", value: "Active", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string mixed case match 2", value: "aCtIvE", values: []string{"pending", "active", "completed"}, wantErr: false},
		{name: "string not in enum", value: "invalid", values: []string{"pending", "active", "completed"}, wantErr: true},

		// Int values (case doesn't apply)
		{name: "int in enum", value: 1, values: []string{"0", "1", "2"}, wantErr: false},
		{name: "int not in enum", value: 5, values: []string{"0", "1", "2"}, wantErr: true},

		// Bool values
		{name: "bool TRUE in enum", value: true, values: []string{"true", "false"}, wantErr: false},
		{name: "bool FALSE in enum", value: false, values: []string{"TRUE", "FALSE"}, wantErr: false},

		// Empty enum
		{name: "value not in empty enum", value: "test", values: []string{}, wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), values: []string{"a", "b"}, wantErr: false},

		// Invalid types
		{name: "invalid type - slice", value: []string{"a"}, values: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := enumCIConstraint{values: tt.values}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestBuildEnumCIConstraint tests buildEnumCIConstraint builder function.
func TestBuildEnumCIConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		testValue any
		wantErr   bool
	}{
		// Case insensitive matches
		{name: "exact match", value: "red green blue", testValue: "red", wantErr: false},
		{name: "uppercase match", value: "red green blue", testValue: "RED", wantErr: false},
		{name: "mixed case match", value: "red green blue", testValue: "GrEeN", wantErr: false},
		{name: "enum values uppercase", value: "RED GREEN BLUE", testValue: "red", wantErr: false},

		// Non-matches
		{name: "value not in enum", value: "red green blue", testValue: "yellow", wantErr: true},

		// Whitespace handling
		{name: "multiple spaces", value: "red   green   blue", testValue: "GREEN", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := buildEnumCIConstraint(tt.value)
			require.NotNil(t, constraint, "expected non-nil constraint")

			err := constraint.Validate(tt.testValue)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBuildEnumConstraint tests buildEnumConstraint builder function.
func TestBuildEnumConstraint(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		testValue any
		wantErr   bool
	}{
		// Valid enum values
		{name: "single value match", value: "red", testValue: "red", wantErr: false},
		{name: "multiple values match first", value: "red green blue", testValue: "red", wantErr: false},
		{name: "multiple values match middle", value: "red green blue", testValue: "green", wantErr: false},
		{name: "multiple values match last", value: "red green blue", testValue: "blue", wantErr: false},

		// Invalid enum values
		{name: "single value mismatch", value: "red", testValue: "blue", wantErr: true},
		{name: "multiple values mismatch", value: "red green blue", testValue: "yellow", wantErr: true},
		{name: "case sensitive mismatch", value: "red", testValue: "RED", wantErr: true},

		// Empty and whitespace handling
		{name: "multiple spaces", value: "red   green   blue", testValue: "green", wantErr: false},
		{name: "trailing spaces", value: "red green blue  ", testValue: "blue", wantErr: false},
		{name: "leading spaces", value: "  red green blue", testValue: "red", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := buildEnumConstraint(tt.value)
			require.NotNil(t, constraint, "expected non-nil constraint")

			err := constraint.Validate(tt.testValue)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
