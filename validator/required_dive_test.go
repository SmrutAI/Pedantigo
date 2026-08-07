package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequiredZeroValueInNestedDiveStruct tests that required fields inside
// nested structs (via dive) correctly distinguish between:
// - Key present with zero value (should PASS)
// - Key missing from JSON (should FAIL)
//
// Bug: Previously used IsZero() which cannot distinguish these cases.
func TestRequiredZeroValueInNestedDiveStruct(t *testing.T) {
	// Test struct matching the bug report
	type Fact struct {
		Content    string  `json:"content" validate:"required"`
		Importance float32 `json:"importance" validate:"required"`
	}
	type Schema struct {
		Facts []Fact `json:"facts" validate:"required,dive"`
	}

	vl := New[Schema]()

	t.Run("zero float32 with key present should pass", func(t *testing.T) {
		// importance: 0.0 is present in JSON - should pass
		jsonData := []byte(`{"facts":[{"content":"test","importance":0.0}]}`)

		result, err := vl.Unmarshal(jsonData)
		require.NoError(t, err, "zero float32 with key present should not error")
		assert.InDelta(t, float32(0.0), result.Facts[0].Importance, 0.0001)
	})

	t.Run("missing importance key should fail", func(t *testing.T) {
		// importance key is missing - should fail
		jsonData := []byte(`{"facts":[{"content":"test"}]}`)

		_, err := vl.Unmarshal(jsonData)
		require.Error(t, err, "missing required field should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		// Should have error for missing Importance field with full path
		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Facts[0].Importance" {
				found = true
				assert.Equal(t, "is required", fieldErr.Message)
			}
		}
		assert.True(t, found, "expected error for Facts[0].Importance, got: %v", ve.Errors)
	})
}

// TestRequiredZeroValueIntInNestedDiveStruct tests zero int values.
func TestRequiredZeroValueIntInNestedDiveStruct(t *testing.T) {
	type Item struct {
		Name  string `json:"name" validate:"required"`
		Count int    `json:"count" validate:"required"`
	}
	type Container struct {
		Items []Item `json:"items" validate:"required,dive"`
	}

	vl := New[Container]()

	t.Run("zero int with key present should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"name":"item1","count":0}]}`)

		result, err := vl.Unmarshal(jsonData)
		require.NoError(t, err, "zero int with key present should not error")
		assert.Equal(t, 0, result.Items[0].Count)
	})

	t.Run("missing count key should fail", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"name":"item1"}]}`)

		_, err := vl.Unmarshal(jsonData)
		require.Error(t, err, "missing required field should error")
	})
}

// TestRequiredZeroValueBoolInNestedDiveStruct tests false bool values.
func TestRequiredZeroValueBoolInNestedDiveStruct(t *testing.T) {
	type Setting struct {
		Name   string `json:"name" validate:"required"`
		Active bool   `json:"active" validate:"required"`
	}
	type Config struct {
		Settings []Setting `json:"settings" validate:"required,dive"`
	}

	vl := New[Config]()

	t.Run("false bool with key present should pass", func(t *testing.T) {
		jsonData := []byte(`{"settings":[{"name":"feature","active":false}]}`)

		result, err := vl.Unmarshal(jsonData)
		require.NoError(t, err, "false bool with key present should not error")
		assert.False(t, result.Settings[0].Active)
	})

	t.Run("missing active key should fail", func(t *testing.T) {
		jsonData := []byte(`{"settings":[{"name":"feature"}]}`)

		_, err := vl.Unmarshal(jsonData)
		require.Error(t, err, "missing required field should error")
	})
}

// TestRequiredEmptyStringInNestedDiveStruct tests empty string values.
// Note: Empty string with key present should pass required validation.
func TestRequiredEmptyStringInNestedDiveStruct(t *testing.T) {
	type Note struct {
		Title   string `json:"title" validate:"required"`
		Content string `json:"content" validate:"required"`
	}
	type Notebook struct {
		Notes []Note `json:"notes" validate:"required,dive"`
	}

	vl := New[Notebook]()

	t.Run("empty string with key present should pass", func(t *testing.T) {
		// content: "" is explicitly provided - should pass required
		jsonData := []byte(`{"notes":[{"title":"My Note","content":""}]}`)

		result, err := vl.Unmarshal(jsonData)
		require.NoError(t, err, "empty string with key present should not error")
		assert.Empty(t, result.Notes[0].Content)
	})

	t.Run("missing content key should fail", func(t *testing.T) {
		jsonData := []byte(`{"notes":[{"title":"My Note"}]}`)

		_, err := vl.Unmarshal(jsonData)
		require.Error(t, err, "missing required field should error")
	})
}

// TestRequiredZeroValueInNestedMapStruct tests required in map values.
func TestRequiredZeroValueInNestedMapStruct(t *testing.T) {
	type Data struct {
		Value int `json:"value" validate:"required"`
	}
	type Registry struct {
		Items map[string]Data `json:"items" validate:"dive"`
	}

	vl := New[Registry]()

	t.Run("zero int in map value with key present should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":{"key1":{"value":0}}}`)

		result, err := vl.Unmarshal(jsonData)
		require.NoError(t, err, "zero int in map value should not error")
		assert.Equal(t, 0, result.Items["key1"].Value)
	})

	t.Run("missing value key in map entry should fail", func(t *testing.T) {
		jsonData := []byte(`{"items":{"key1":{}}}`)

		_, err := vl.Unmarshal(jsonData)
		require.Error(t, err, "missing required field in map value should error")
	})
}
