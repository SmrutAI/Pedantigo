package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUniqueConstraint_Slice tests uniqueness validation for slices.
func TestUniqueConstraint_Slice(t *testing.T) {
	c := uniqueConstraint{}

	tests := []simpleTestCase{
		// String slices
		{name: "unique strings", value: []string{"a", "b", "c"}, wantErr: false},
		{name: "duplicate strings", value: []string{"a", "b", "a"}, wantErr: true},
		{name: "all same strings", value: []string{"a", "a", "a"}, wantErr: true},
		{name: "case sensitive", value: []string{"a", "A", "b"}, wantErr: false},

		// Int slices
		{name: "unique ints", value: []int{1, 2, 3}, wantErr: false},
		{name: "duplicate ints", value: []int{1, 2, 1}, wantErr: true},
		{name: "unique ints negative", value: []int{-1, 0, 1}, wantErr: false},

		// Float slices
		{name: "unique floats", value: []float64{1.1, 2.2, 3.3}, wantErr: false},
		{name: "duplicate floats", value: []float64{1.1, 2.2, 1.1}, wantErr: true},

		// Edge cases
		{name: "empty slice", value: []string{}, wantErr: false},
		{name: "single element", value: []int{1}, wantErr: false},
		{name: "nil slice", value: []string(nil), wantErr: false},

		// Pointer to slice
		{name: "pointer to unique slice", value: &[]int{1, 2, 3}, wantErr: false},
		{name: "pointer to duplicate slice", value: &[]int{1, 2, 1}, wantErr: true},

		// Bool slice
		{name: "unique bools", value: []bool{true, false}, wantErr: false},
		{name: "duplicate bools", value: []bool{true, true}, wantErr: true},

		// Nil value
		{name: "nil value", value: nil, wantErr: false},
	}

	runSimpleConstraintTests(t, c, tests)
}

// TestUniqueConstraint_Array tests uniqueness validation for arrays.
func TestUniqueConstraint_Array(t *testing.T) {
	c := uniqueConstraint{}

	t.Run("unique array", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		err := c.Validate(arr)
		assert.NoError(t, err)
	})

	t.Run("duplicate array", func(t *testing.T) {
		arr := [3]int{1, 2, 1}
		err := c.Validate(arr)
		assert.Error(t, err)
	})

	t.Run("empty array", func(t *testing.T) {
		arr := [0]int{}
		err := c.Validate(arr)
		assert.NoError(t, err)
	})
}

// TestUniqueConstraint_Map tests uniqueness validation for map values.
func TestUniqueConstraint_Map(t *testing.T) {
	c := uniqueConstraint{}

	tests := []simpleTestCase{
		// Map with string keys, int values
		{name: "unique int values", value: map[string]int{"a": 1, "b": 2, "c": 3}, wantErr: false},
		{name: "duplicate int values", value: map[string]int{"a": 1, "b": 1}, wantErr: true},

		// Map with int keys, string values
		{name: "unique string values", value: map[int]string{1: "a", 2: "b"}, wantErr: false},
		{name: "duplicate string values", value: map[int]string{1: "a", 2: "a"}, wantErr: true},

		// Edge cases
		{name: "empty map", value: map[string]int{}, wantErr: false},
		{name: "single entry", value: map[string]int{"a": 1}, wantErr: false},
		{name: "nil map", value: map[string]int(nil), wantErr: false},

		// Pointer to map
		{name: "pointer to unique map", value: &map[string]int{"a": 1, "b": 2}, wantErr: false},
		{name: "pointer to duplicate map", value: &map[string]int{"a": 1, "b": 1}, wantErr: true},
	}

	runSimpleConstraintTests(t, c, tests)
}

// TestUniqueConstraint_StructSlice tests uniqueness validation by struct field.
func TestUniqueConstraint_StructSlice(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	c := uniqueConstraint{field: "ID"}

	t.Run("unique IDs", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		err := c.Validate(users)
		assert.NoError(t, err)
	})

	t.Run("duplicate IDs", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}, {ID: 1, Name: "Bob"}}
		err := c.Validate(users)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ID")
	})

	t.Run("same name different ID", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Alice"}}
		err := c.Validate(users)
		assert.NoError(t, err) // Only checking ID uniqueness
	})

	t.Run("empty slice", func(t *testing.T) {
		users := []User{}
		err := c.Validate(users)
		assert.NoError(t, err)
	})

	t.Run("single element", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}}
		err := c.Validate(users)
		assert.NoError(t, err)
	})
}

// TestUniqueConstraint_StructSliceByName tests uniqueness by string field.
func TestUniqueConstraint_StructSliceByName(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	c := uniqueConstraint{field: "Name"}

	t.Run("unique names", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		err := c.Validate(users)
		assert.NoError(t, err)
	})

	t.Run("duplicate names", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Alice"}}
		err := c.Validate(users)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Name")
	})
}

// TestUniqueConstraint_PointerStructSlice tests uniqueness for pointer element slices.
func TestUniqueConstraint_PointerStructSlice(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	c := uniqueConstraint{field: "ID"}

	t.Run("unique pointer elements", func(t *testing.T) {
		users := []*User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		err := c.Validate(users)
		assert.NoError(t, err)
	})

	t.Run("duplicate pointer elements", func(t *testing.T) {
		users := []*User{{ID: 1, Name: "Alice"}, {ID: 1, Name: "Bob"}}
		err := c.Validate(users)
		assert.Error(t, err)
	})

	t.Run("nil pointer element", func(t *testing.T) {
		users := []*User{{ID: 1, Name: "Alice"}, nil, {ID: 2, Name: "Bob"}}
		err := c.Validate(users)
		assert.NoError(t, err) // nil elements are skipped
	})
}

// TestUniqueConstraint_NonExistentField tests behavior when field doesn't exist.
func TestUniqueConstraint_NonExistentField(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	c := uniqueConstraint{field: "NonExistent"}

	t.Run("field not found", func(t *testing.T) {
		users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		err := c.Validate(users)
		// Should not error - non-existent fields return nil (skip)
		assert.NoError(t, err)
	})
}

// TestUniqueConstraint_NonCollectionTypes tests behavior on non-collection types.
func TestUniqueConstraint_NonCollectionTypes(t *testing.T) {
	c := uniqueConstraint{}

	tests := []simpleTestCase{
		{name: "string", value: "hello", wantErr: false},
		{name: "int", value: 42, wantErr: false},
		{name: "bool", value: true, wantErr: false},
		{name: "struct", value: struct{ X int }{X: 1}, wantErr: false},
	}

	// Non-collection types should pass validation (handled at validator.go level)
	// The constraint itself just returns nil for non-collections
	runSimpleConstraintTests(t, c, tests)
}

// TestUniqueConstraint_NilPointer tests behavior with nil pointers.
func TestUniqueConstraint_NilPointer(t *testing.T) {
	c := uniqueConstraint{}

	t.Run("nil pointer to slice", func(t *testing.T) {
		var ptr *[]int = nil
		err := c.Validate(ptr)
		assert.NoError(t, err)
	})

	t.Run("nil pointer to map", func(t *testing.T) {
		var ptr *map[string]int = nil
		err := c.Validate(ptr)
		assert.NoError(t, err)
	})
}

// TestUniqueConstraint_TypeVariants tests different numeric type variants.
func TestUniqueConstraint_TypeVariants(t *testing.T) {
	c := uniqueConstraint{}

	t.Run("int8 slice", func(t *testing.T) {
		err := c.Validate([]int8{1, 2, 3})
		require.NoError(t, err)
		err = c.Validate([]int8{1, 2, 1})
		require.Error(t, err)
	})

	t.Run("int16 slice", func(t *testing.T) {
		err := c.Validate([]int16{1, 2, 3})
		require.NoError(t, err)
		err = c.Validate([]int16{1, 2, 1})
		require.Error(t, err)
	})

	t.Run("int32 slice", func(t *testing.T) {
		err := c.Validate([]int32{1, 2, 3})
		require.NoError(t, err)
		err = c.Validate([]int32{1, 2, 1})
		require.Error(t, err)
	})

	t.Run("int64 slice", func(t *testing.T) {
		err := c.Validate([]int64{1, 2, 3})
		require.NoError(t, err)
		err = c.Validate([]int64{1, 2, 1})
		require.Error(t, err)
	})

	t.Run("uint slice", func(t *testing.T) {
		err := c.Validate([]uint{1, 2, 3})
		require.NoError(t, err)
		err = c.Validate([]uint{1, 2, 1})
		require.Error(t, err)
	})

	t.Run("uint8 slice", func(t *testing.T) {
		err := c.Validate([]uint8{1, 2, 3})
		require.NoError(t, err)
		err = c.Validate([]uint8{1, 2, 1})
		require.Error(t, err)
	})

	t.Run("float32 slice", func(t *testing.T) {
		err := c.Validate([]float32{1.1, 2.2, 3.3})
		require.NoError(t, err)
		err = c.Validate([]float32{1.1, 2.2, 1.1})
		require.Error(t, err)
	})
}

// TestBuildConstraints_Unique tests that BuildConstraints correctly creates unique constraint.
func TestBuildConstraints_Unique(t *testing.T) {
	t.Run("unique without field", func(t *testing.T) {
		constraints := BuildConstraints(map[string]string{"unique": ""}, nil)
		assert.Len(t, constraints, 1)
	})

	t.Run("unique with field", func(t *testing.T) {
		constraints := BuildConstraints(map[string]string{"unique": "ID"}, nil)
		assert.Len(t, constraints, 1)
	})
}
