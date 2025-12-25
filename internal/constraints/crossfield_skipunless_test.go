package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SmrutAI/pedantigo"
)

// TestSkipUnless tests the skip_unless constraint for conditional validation.
// Note: skip_unless is a cross-field constraint that acts as a "gate".
// It does NOT skip other constraints - it just passes/fails based on whether
// the condition is met. For full skip behavior, use omitempty or pointer fields.
func TestSkipUnless(t *testing.T) {
	t.Run("condition met - validation proceeds", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" pedantigo:"skip_unless=Status active"`
		}

		validator := New[Form]()

		// Condition met (Status == "active"), skip_unless passes
		data := Form{
			Status: "active",
			Data:   "some data",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("condition not met - validation skipped", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" pedantigo:"skip_unless=Status active"`
		}

		validator := New[Form]()

		// Condition NOT met (Status != "active"), skip_unless passes (skips)
		data := Form{
			Status: "inactive",
			Data:   "",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("boolean condition - true", func(t *testing.T) {
		type Form struct {
			Enabled bool   `json:"enabled"`
			APIKey  string `json:"api_key" pedantigo:"skip_unless=Enabled true"`
		}

		validator := New[Form]()

		// Condition met (Enabled == true)
		data := Form{
			Enabled: true,
			APIKey:  "valid-key",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("boolean condition - false skips", func(t *testing.T) {
		type Form struct {
			Enabled bool   `json:"enabled"`
			APIKey  string `json:"api_key" pedantigo:"skip_unless=Enabled true"`
		}

		validator := New[Form]()

		// Condition NOT met (Enabled == false)
		data := Form{
			Enabled: false,
			APIKey:  "",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("integer condition", func(t *testing.T) {
		type Form struct {
			Level  int    `json:"level"`
			Reward string `json:"reward" pedantigo:"skip_unless=Level 5"`
		}

		validator := New[Form]()

		// Condition met (Level == 5)
		data := Form{
			Level:  5,
			Reward: "gold",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("integer condition not met", func(t *testing.T) {
		type Form struct {
			Level  int    `json:"level"`
			Reward string `json:"reward" pedantigo:"skip_unless=Level 5"`
		}

		validator := New[Form]()

		// Condition NOT met (Level != 5)
		data := Form{
			Level:  3,
			Reward: "",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("case sensitive match", func(t *testing.T) {
		type Form struct {
			Type string `json:"type"`
			Data string `json:"data" pedantigo:"skip_unless=Type Active"`
		}

		validator := New[Form]()

		// Condition NOT met (case mismatch: "active" != "Active")
		data := Form{
			Type: "active", // lowercase, doesn't match "Active"
			Data: "",
		}

		err := validator.Validate(&data)
		assert.NoError(t, err) // Skipped because case doesn't match
	})
}

func TestSkipUnlessErrorCases(t *testing.T) {
	t.Run("target field missing - panics at validator creation", func(t *testing.T) {
		type Form struct {
			Data string `json:"data" pedantigo:"skip_unless=NonExistentField value"`
		}

		// ParseFieldPath panics when field doesn't exist
		// This is intentional - it catches misconfiguration at startup
		assert.Panics(t, func() {
			_ = New[Form]()
		}, "should panic when target field doesn't exist")
	})
}

func TestSkipUnlessWithValidation(t *testing.T) {
	// Note: skip_unless by itself doesn't skip other constraints.
	// It's a gate that passes when the condition isn't met.
	// To have conditional validation, combine with required_if or use pointers.

	t.Run("combined with required_if for conditional validation", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" pedantigo:"required_if=Status active,min=5"`
		}

		validator := New[Form]()

		// Condition met - requires data and min length
		data := Form{
			Status: "active",
			Data:   "valid data",
		}

		err := validator.Validate(&data)
		require.NoError(t, err)

		// Condition met but data too short
		data2 := Form{
			Status: "active",
			Data:   "abc",
		}

		err = validator.Validate(&data2)
		require.Error(t, err)
	})
}
