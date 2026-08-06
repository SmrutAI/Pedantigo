package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SmrutAI/pedantigo"
)

// TestCrossFieldComparison_IncompatibleTypes tests that cross-field comparisons
// with incompatible types return appropriate errors.
func TestCrossFieldComparison_IncompatibleTypes(t *testing.T) {
	t.Run("eqfield: int vs string incompatible", func(t *testing.T) {
		type MixedTypes struct {
			IntField    int    `json:"int_field"`
			StringField string `json:"string_field" validate:"eqfield=IntField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			IntField:    42,
			StringField: "42",
		}

		err := validator.Validate(&data)
		require.Error(t, err)

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		foundIncompatible := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "StringField" {
				foundIncompatible = true
			}
		}
		assert.True(t, foundIncompatible, "expected incompatible types error")
	})

	t.Run("nefield: float vs string incompatible", func(t *testing.T) {
		type MixedTypes struct {
			FloatField  float64 `json:"float_field"`
			StringField string  `json:"string_field" validate:"nefield=FloatField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			FloatField:  3.14,
			StringField: "3.14",
		}

		err := validator.Validate(&data)
		require.Error(t, err)

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("gtfield: bool vs int incompatible", func(t *testing.T) {
		type MixedTypes struct {
			IntField  int  `json:"int_field"`
			BoolField bool `json:"bool_field" validate:"gtfield=IntField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			IntField:  10,
			BoolField: true,
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("gtefield: string vs float incompatible", func(t *testing.T) {
		type MixedTypes struct {
			FloatField  float64 `json:"float_field"`
			StringField string  `json:"string_field" validate:"gtefield=FloatField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			FloatField:  100.5,
			StringField: "100.5",
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("ltfield: uint vs string incompatible", func(t *testing.T) {
		type MixedTypes struct {
			UintField   uint   `json:"uint_field"`
			StringField string `json:"string_field" validate:"ltfield=UintField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			UintField:   50,
			StringField: "25",
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("ltefield: int vs bool incompatible", func(t *testing.T) {
		type MixedTypes struct {
			BoolField bool `json:"bool_field"`
			IntField  int  `json:"int_field" validate:"ltefield=BoolField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			BoolField: false,
			IntField:  0,
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("eqfield: slice vs string incompatible", func(t *testing.T) {
		type MixedTypes struct {
			SliceField  []int  `json:"slice_field"`
			StringField string `json:"string_field" validate:"eqfield=SliceField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			SliceField:  []int{1, 2, 3},
			StringField: "[1, 2, 3]",
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("nefield: map vs string incompatible", func(t *testing.T) {
		type MixedTypes struct {
			MapField    map[string]int `json:"map_field"`
			StringField string         `json:"string_field" validate:"nefield=MapField"`
		}

		validator := New[MixedTypes]()
		data := MixedTypes{
			MapField:    map[string]int{"a": 1},
			StringField: "a:1",
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestCrossFieldComparison_NestedFieldPath tests cross-field comparisons with nested field paths.
func TestCrossFieldComparison_NestedFieldPath(t *testing.T) {
	t.Run("eqfield: nested field path", func(t *testing.T) {
		type Inner struct {
			Value string `json:"value"`
		}
		type Outer struct {
			Inner   Inner  `json:"inner"`
			Compare string `json:"compare" validate:"eqfield=Inner.Value"`
		}

		validator := New[Outer]()
		data := Outer{
			Inner:   Inner{Value: "test"},
			Compare: "test",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("gtfield: nested field path numeric", func(t *testing.T) {
		type Inner struct {
			Min int `json:"min"`
		}
		type Outer struct {
			Inner Inner `json:"inner"`
			Max   int   `json:"max" validate:"gtfield=Inner.Min"`
		}

		validator := New[Outer]()
		data := Outer{
			Inner: Inner{Min: 10},
			Max:   20,
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("ltfield: nested field path invalid", func(t *testing.T) {
		type Inner struct {
			Max int `json:"max"`
		}
		type Outer struct {
			Inner Inner `json:"inner"`
			Min   int   `json:"min" validate:"ltfield=Inner.Max"`
		}

		validator := New[Outer]()
		data := Outer{
			Inner: Inner{Max: 5},
			Min:   10, // Greater than Inner.Max, should fail
		}

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestCrossFieldComparison_PointerTypes tests cross-field comparisons with pointer types.
func TestCrossFieldComparison_PointerTypes(t *testing.T) {
	t.Run("eqfield: pointer to non-pointer compatible", func(t *testing.T) {
		type PointerTest struct {
			IntPtr *int `json:"int_ptr"`
			IntVal int  `json:"int_val" validate:"eqfield=IntPtr"`
		}

		validator := New[PointerTest]()
		val := 42
		data := PointerTest{
			IntPtr: &val,
			IntVal: 42,
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("gtfield: pointer to non-pointer compatible", func(t *testing.T) {
		type PointerTest struct {
			MinPtr *int `json:"min_ptr"`
			MaxVal int  `json:"max_val" validate:"gtfield=MinPtr"`
		}

		validator := New[PointerTest]()
		val := 10
		data := PointerTest{
			MinPtr: &val,
			MaxVal: 20,
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})
}
