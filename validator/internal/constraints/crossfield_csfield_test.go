package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

func TestCsFieldConstraints(t *testing.T) {
	// Test eqcsfield
	t.Run("eqcsfield", func(t *testing.T) {
		type Inner struct {
			Value string `json:"value"`
		}
		type Form struct {
			Inner Inner  `json:"inner"`
			Match string `json:"match" validate:"eqcsfield=Inner.Value"`
		}
		vl := validator.New[Form]()

		// Should pass when values match
		err := vl.Validate(&Form{
			Inner: Inner{Value: "test"},
			Match: "test",
		})
		require.NoError(t, err)

		// Should fail when values don't match
		err = vl.Validate(&Form{
			Inner: Inner{Value: "test"},
			Match: "different",
		})
		require.Error(t, err)
	})

	// Test necsfield
	t.Run("necsfield", func(t *testing.T) {
		type Inner struct {
			Value string `json:"value"`
		}
		type Form struct {
			Inner    Inner  `json:"inner"`
			NotMatch string `json:"not_match" validate:"necsfield=Inner.Value"`
		}
		vl := validator.New[Form]()

		// Should pass when values are different
		err := vl.Validate(&Form{
			Inner:    Inner{Value: "test"},
			NotMatch: "different",
		})
		require.NoError(t, err)

		// Should fail when values match
		err = vl.Validate(&Form{
			Inner:    Inner{Value: "test"},
			NotMatch: "test",
		})
		require.Error(t, err)
	})

	// Test gtcsfield
	t.Run("gtcsfield", func(t *testing.T) {
		type Inner struct {
			Value int `json:"value"`
		}
		type Form struct {
			Inner   Inner `json:"inner"`
			Greater int   `json:"greater" validate:"gtcsfield=Inner.Value"`
		}
		vl := validator.New[Form]()

		// Should pass when greater
		err := vl.Validate(&Form{
			Inner:   Inner{Value: 10},
			Greater: 15,
		})
		require.NoError(t, err)

		// Should fail when not greater
		err = vl.Validate(&Form{
			Inner:   Inner{Value: 10},
			Greater: 5,
		})
		require.Error(t, err)
	})

	// Test gtecsfield
	t.Run("gtecsfield", func(t *testing.T) {
		type Inner struct {
			Value int `json:"value"`
		}
		type Form struct {
			Inner       Inner `json:"inner"`
			GreaterOrEq int   `json:"greater_or_eq" validate:"gtecsfield=Inner.Value"`
		}
		vl := validator.New[Form]()

		// Should pass when equal
		err := vl.Validate(&Form{
			Inner:       Inner{Value: 10},
			GreaterOrEq: 10,
		})
		require.NoError(t, err)

		// Should pass when greater
		err = vl.Validate(&Form{
			Inner:       Inner{Value: 10},
			GreaterOrEq: 15,
		})
		require.NoError(t, err)

		// Should fail when less
		err = vl.Validate(&Form{
			Inner:       Inner{Value: 10},
			GreaterOrEq: 5,
		})
		require.Error(t, err)
	})

	// Test ltcsfield
	t.Run("ltcsfield", func(t *testing.T) {
		type Inner struct {
			Value int `json:"value"`
		}
		type Form struct {
			Inner Inner `json:"inner"`
			Less  int   `json:"less" validate:"ltcsfield=Inner.Value"`
		}
		vl := validator.New[Form]()

		// Should pass when less
		err := vl.Validate(&Form{
			Inner: Inner{Value: 10},
			Less:  5,
		})
		require.NoError(t, err)

		// Should fail when greater
		err = vl.Validate(&Form{
			Inner: Inner{Value: 10},
			Less:  15,
		})
		require.Error(t, err)
	})

	// Test ltecsfield
	t.Run("ltecsfield", func(t *testing.T) {
		type Inner struct {
			Value int `json:"value"`
		}
		type Form struct {
			Inner    Inner `json:"inner"`
			LessOrEq int   `json:"less_or_eq" validate:"ltecsfield=Inner.Value"`
		}
		vl := validator.New[Form]()

		// Should pass when equal
		err := vl.Validate(&Form{
			Inner:    Inner{Value: 10},
			LessOrEq: 10,
		})
		require.NoError(t, err)

		// Should pass when less
		err = vl.Validate(&Form{
			Inner:    Inner{Value: 10},
			LessOrEq: 5,
		})
		require.NoError(t, err)

		// Should fail when greater
		err = vl.Validate(&Form{
			Inner:    Inner{Value: 10},
			LessOrEq: 15,
		})
		require.Error(t, err)
	})
}
