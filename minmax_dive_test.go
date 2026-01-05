package pedantigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinMaxNumericInNestedDiveStruct tests that min/max constraints work correctly
// on numeric fields inside nested structs accessed via dive.
func TestMinMaxNumericInNestedDiveStruct(t *testing.T) {
	type Fact struct {
		Content    string  `json:"content" pedantigo:"required"`
		Importance float32 `json:"importance" pedantigo:"required,min=0,max=1"`
	}
	type Schema struct {
		Facts []Fact `json:"facts" pedantigo:"required,dive"`
	}

	validator := New[Schema]()

	t.Run("value exceeds max should fail", func(t *testing.T) {
		jsonData := []byte(`{"facts":[{"content":"test","importance":1.5}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "value exceeding max should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Facts[0].Importance" {
				found = true
				assert.Contains(t, fieldErr.Message, "at most 1")
			}
		}
		assert.True(t, found, "expected error for Facts[0].Importance, got: %v", ve.Errors)
	})

	t.Run("value below min should fail", func(t *testing.T) {
		jsonData := []byte(`{"facts":[{"content":"test","importance":-0.1}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "value below min should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Facts[0].Importance" {
				found = true
				assert.Contains(t, fieldErr.Message, "at least 0")
			}
		}
		assert.True(t, found, "expected error for Facts[0].Importance, got: %v", ve.Errors)
	})

	t.Run("value in range should pass", func(t *testing.T) {
		jsonData := []byte(`{"facts":[{"content":"test","importance":0.5}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "value in range should not error")
		assert.InDelta(t, float32(0.5), result.Facts[0].Importance, 0.0001)
	})

	t.Run("zero value (min boundary) should pass", func(t *testing.T) {
		jsonData := []byte(`{"facts":[{"content":"test","importance":0}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "zero value at min boundary should not error")
		assert.InDelta(t, float32(0.0), result.Facts[0].Importance, 0.0001)
	})

	t.Run("max boundary value should pass", func(t *testing.T) {
		jsonData := []byte(`{"facts":[{"content":"test","importance":1}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err, "max boundary value should not error")
		assert.InDelta(t, float32(1.0), result.Facts[0].Importance, 0.0001)
	})

	t.Run("multiple facts with mixed validity", func(t *testing.T) {
		// First fact valid, second fact exceeds max
		jsonData := []byte(`{"facts":[{"content":"good","importance":0.5},{"content":"bad","importance":2.0}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err, "second fact exceeding max should error")

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)

		found := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Facts[1].Importance" {
				found = true
				assert.Contains(t, fieldErr.Message, "at most 1")
			}
		}
		assert.True(t, found, "expected error for Facts[1].Importance, got: %v", ve.Errors)
	})
}

// TestMinMaxIntInNestedDiveStruct tests min/max on int fields in nested structs.
func TestMinMaxIntInNestedDiveStruct(t *testing.T) {
	type Item struct {
		Name     string `json:"name" pedantigo:"required"`
		Quantity int    `json:"quantity" pedantigo:"required,min=0,max=100"`
	}
	type Order struct {
		Items []Item `json:"items" pedantigo:"required,dive"`
	}

	validator := New[Order]()

	t.Run("quantity exceeds max should fail", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"name":"widget","quantity":150}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)

		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.NotEmpty(t, ve.Errors)
	})

	t.Run("negative quantity should fail", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"name":"widget","quantity":-5}]}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)
	})

	t.Run("zero quantity should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"name":"widget","quantity":0}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err)
		assert.Equal(t, 0, result.Items[0].Quantity)
	})

	t.Run("valid quantity should pass", func(t *testing.T) {
		jsonData := []byte(`{"items":[{"name":"widget","quantity":50}]}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err)
		assert.Equal(t, 50, result.Items[0].Quantity)
	})
}

// TestMinMaxInNestedMapStruct tests min/max on fields in map value structs.
func TestMinMaxInNestedMapStruct(t *testing.T) {
	type Score struct {
		Value float64 `json:"value" pedantigo:"required,min=0,max=100"`
	}
	type Scores struct {
		BySubject map[string]Score `json:"by_subject" pedantigo:"dive"`
	}

	validator := New[Scores]()

	t.Run("score exceeds max should fail", func(t *testing.T) {
		jsonData := []byte(`{"by_subject":{"math":{"value":105}}}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)
	})

	t.Run("negative score should fail", func(t *testing.T) {
		jsonData := []byte(`{"by_subject":{"math":{"value":-10}}}`)

		_, err := validator.Unmarshal(jsonData)
		require.Error(t, err)
	})

	t.Run("valid scores should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_subject":{"math":{"value":95},"english":{"value":88}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err)
		assert.InDelta(t, 95.0, result.BySubject["math"].Value, 0.0001)
		assert.InDelta(t, 88.0, result.BySubject["english"].Value, 0.0001)
	})

	t.Run("zero score should pass", func(t *testing.T) {
		jsonData := []byte(`{"by_subject":{"failed":{"value":0}}}`)

		result, err := validator.Unmarshal(jsonData)
		require.NoError(t, err)
		assert.InDelta(t, 0.0, result.BySubject["failed"].Value, 0.0001)
	})
}
