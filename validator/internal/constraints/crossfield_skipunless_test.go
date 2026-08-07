package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SmrutAI/pedantigo/v2/validator"
)

// TestSkipUnless tests the skip_unless constraint for conditional validation.
// skip_unless SKIPS ALL validation on the field when the condition is NOT met.
// This enables discriminated union patterns where only relevant fields are validated.
func TestSkipUnless(t *testing.T) {
	t.Run("condition met - validation proceeds", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" validate:"skip_unless=Status active"`
		}

		vl := New[Form]()

		// Condition met (Status == "active"), skip_unless passes
		data := Form{
			Status: "active",
			Data:   "some data",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("condition not met - validation skipped", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" validate:"skip_unless=Status active"`
		}

		vl := New[Form]()

		// Condition NOT met (Status != "active"), skip_unless passes (skips)
		data := Form{
			Status: "inactive",
			Data:   "",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("boolean condition - true", func(t *testing.T) {
		type Form struct {
			Enabled bool   `json:"enabled"`
			APIKey  string `json:"api_key" validate:"skip_unless=Enabled true"`
		}

		vl := New[Form]()

		// Condition met (Enabled == true)
		data := Form{
			Enabled: true,
			APIKey:  "valid-key",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("boolean condition - false skips", func(t *testing.T) {
		type Form struct {
			Enabled bool   `json:"enabled"`
			APIKey  string `json:"api_key" validate:"skip_unless=Enabled true"`
		}

		vl := New[Form]()

		// Condition NOT met (Enabled == false)
		data := Form{
			Enabled: false,
			APIKey:  "",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("integer condition", func(t *testing.T) {
		type Form struct {
			Level  int    `json:"level"`
			Reward string `json:"reward" validate:"skip_unless=Level 5"`
		}

		vl := New[Form]()

		// Condition met (Level == 5)
		data := Form{
			Level:  5,
			Reward: "gold",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("integer condition not met", func(t *testing.T) {
		type Form struct {
			Level  int    `json:"level"`
			Reward string `json:"reward" validate:"skip_unless=Level 5"`
		}

		vl := New[Form]()

		// Condition NOT met (Level != 5)
		data := Form{
			Level:  3,
			Reward: "",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("case sensitive match", func(t *testing.T) {
		type Form struct {
			Type string `json:"type"`
			Data string `json:"data" validate:"skip_unless=Type Active"`
		}

		vl := New[Form]()

		// Condition NOT met (case mismatch: "active" != "Active")
		data := Form{
			Type: "active", // lowercase, doesn't match "Active"
			Data: "",
		}

		err := vl.Validate(&data)
		assert.NoError(t, err) // Skipped because case doesn't match
	})
}

func TestSkipUnlessErrorCases(t *testing.T) {
	t.Run("target field missing - panics at validator creation", func(t *testing.T) {
		type Form struct {
			Data string `json:"data" validate:"skip_unless=NonExistentField value"`
		}

		// ParseFieldPath panics when field doesn't exist
		// This is intentional - it catches misconfiguration at startup
		assert.Panics(t, func() {
			_ = New[Form]()
		}, "should panic when target field doesn't exist")
	})
}

// TestSkipUnless_SkipsAllConstraints verifies that when condition is NOT met,
// ALL other constraints on the field are skipped (required, min, email, etc.)
func TestSkipUnless_SkipsAllConstraints(t *testing.T) {
	t.Run("skip_unless skips required when condition not met", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" validate:"skip_unless=Status active,required,min=5"`
		}

		vl := New[Form]()

		// Condition NOT met (Status != "active")
		// Even though Data is empty and required+min=5 would fail,
		// skip_unless should skip ALL validation
		data := Form{
			Status: "inactive",
			Data:   "", // Would fail required and min=5, but should be skipped
		}

		err := vl.Validate(&data)
		require.NoError(t, err, "skip_unless should skip required when condition not met")
	})

	t.Run("skip_unless validates when condition is met", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Data   string `json:"data" validate:"skip_unless=Status active,required,min=5"`
		}

		vl := New[Form]()

		// Condition met (Status == "active")
		// Now required+min=5 should be enforced
		data := Form{
			Status: "active",
			Data:   "", // Fails required
		}

		err := vl.Validate(&data)
		require.Error(t, err, "skip_unless should validate when condition is met")
	})

	t.Run("skip_unless skips email validation when condition not met", func(t *testing.T) {
		type Form struct {
			ContactMethod string `json:"contact_method"`
			Email         string `json:"email" validate:"skip_unless=ContactMethod email,required,email"`
		}

		vl := New[Form]()

		// Condition NOT met (ContactMethod != "email")
		data := Form{
			ContactMethod: "phone",
			Email:         "not-an-email", // Invalid email, but should be skipped
		}

		err := vl.Validate(&data)
		require.NoError(t, err, "skip_unless should skip email validation")
	})

	t.Run("skip_unless validates email when condition is met", func(t *testing.T) {
		type Form struct {
			ContactMethod string `json:"contact_method"`
			Email         string `json:"email" validate:"skip_unless=ContactMethod email,required,email"`
		}

		vl := New[Form]()

		// Condition met (ContactMethod == "email")
		data := Form{
			ContactMethod: "email",
			Email:         "not-an-email", // Invalid email - should fail
		}

		err := vl.Validate(&data)
		require.Error(t, err, "skip_unless should validate email when condition is met")
	})
}

// TestSkipUnless_DiscriminatedUnion tests the classic discriminated union pattern.
// This is the primary use case for skip_unless.
func TestSkipUnless_DiscriminatedUnion(t *testing.T) {
	type TV struct {
		Channel int `json:"channel" validate:"required,min=1,max=999"`
	}
	type Fan struct {
		Speed int `json:"speed" validate:"required,min=1,max=5"`
	}
	type Suite struct {
		SuiteType string `json:"suite_type" validate:"required,oneof=tv fan"`
		TV        TV     `json:"tv" validate:"skip_unless=SuiteType tv"`
		Fan       Fan    `json:"fan" validate:"skip_unless=SuiteType fan"`
	}

	vl := New[Suite]()

	t.Run("tv type validates TV, skips Fan", func(t *testing.T) {
		data := Suite{
			SuiteType: "tv",
			TV:        TV{Channel: 42},
			Fan:       Fan{Speed: 0}, // Invalid (required, min=1), but should be skipped
		}

		err := vl.Validate(&data)
		require.NoError(t, err, "Fan should be skipped when SuiteType=tv")
	})

	t.Run("fan type validates Fan, skips TV", func(t *testing.T) {
		data := Suite{
			SuiteType: "fan",
			TV:        TV{Channel: 0}, // Invalid (required, min=1), but should be skipped
			Fan:       Fan{Speed: 3},
		}

		err := vl.Validate(&data)
		require.NoError(t, err, "TV should be skipped when SuiteType=fan")
	})

	t.Run("tv type fails when TV is invalid", func(t *testing.T) {
		data := Suite{
			SuiteType: "tv",
			TV:        TV{Channel: 0}, // Invalid (0 < min=1) - validation runs
			Fan:       Fan{Speed: 0},  // Invalid but skipped
		}

		err := vl.Validate(&data)
		require.Error(t, err, "TV validation should run when SuiteType=tv")
	})

	t.Run("fan type fails when Fan is invalid", func(t *testing.T) {
		data := Suite{
			SuiteType: "fan",
			TV:        TV{Channel: 0}, // Invalid but skipped
			Fan:       Fan{Speed: 0},  // Invalid (0 < min=1) - validation runs
		}

		err := vl.Validate(&data)
		require.Error(t, err, "Fan validation should run when SuiteType=fan")
	})
}

// TestSkipUnless_MultipleConstraints tests that ALL constraints are skipped.
func TestSkipUnless_MultipleConstraints(t *testing.T) {
	type Form struct {
		Type     string `json:"type"`
		Username string `json:"username" validate:"skip_unless=Type user,required,min=3,max=20,alphanum"`
	}

	vl := New[Form]()

	t.Run("all constraints skipped when condition not met", func(t *testing.T) {
		data := Form{
			Type:     "guest", // Not "user"
			Username: "",      // Would fail: required, min=3
		}

		err := vl.Validate(&data)
		require.NoError(t, err)
	})

	t.Run("all constraints validated when condition is met", func(t *testing.T) {
		data := Form{
			Type:     "user",
			Username: "ab", // Fails min=3
		}

		err := vl.Validate(&data)
		require.Error(t, err)
	})
}

// TestSkipUnless_TypeComparisons tests different field types for comparison.
func TestSkipUnless_TypeComparisons(t *testing.T) {
	t.Run("uint comparison", func(t *testing.T) {
		type Form struct {
			Level uint   `json:"level"`
			Badge string `json:"badge" validate:"skip_unless=Level 10,required,min=1"`
		}

		vl := New[Form]()

		// Condition not met
		data := Form{Level: 5, Badge: ""}
		err := vl.Validate(&data)
		require.NoError(t, err)

		// Condition met
		data2 := Form{Level: 10, Badge: ""}
		err = vl.Validate(&data2)
		require.Error(t, err)
	})

	t.Run("bool comparison false", func(t *testing.T) {
		type Form struct {
			Disabled bool   `json:"disabled"`
			Reason   string `json:"reason" validate:"skip_unless=Disabled true,required,min=5"`
		}

		vl := New[Form]()

		// Condition not met (Disabled=false, not "true")
		data := Form{Disabled: false, Reason: ""}
		err := vl.Validate(&data)
		require.NoError(t, err)

		// Condition met
		data2 := Form{Disabled: true, Reason: ""}
		err = vl.Validate(&data2)
		require.Error(t, err)
	})
}
