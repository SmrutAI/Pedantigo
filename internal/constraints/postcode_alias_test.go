package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	pedantigo "github.com/SmrutAI/pedantigo"
)

// TestPostcodeISO3166Alpha2Alias tests the postcode_iso3166_alpha2 alias.
func TestPostcodeISO3166Alpha2Alias(t *testing.T) {
	// Test 1: US postcode via alias - valid
	t.Run("US postcode via alias valid", func(t *testing.T) {
		type Form struct {
			Zip string `json:"zip" validate:"postcode_iso3166_alpha2=US"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Zip: "90210"})
		require.NoError(t, err)
	})

	// Test 2: US postcode via alias - invalid
	t.Run("US postcode via alias invalid", func(t *testing.T) {
		type Form struct {
			Zip string `json:"zip" validate:"postcode_iso3166_alpha2=US"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Zip: "invalid"})
		require.Error(t, err)
	})

	// Test 3: UK postcode via alias
	t.Run("UK postcode via alias", func(t *testing.T) {
		type Form struct {
			Postcode string `json:"postcode" validate:"postcode_iso3166_alpha2=GB"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Postcode: "SW1A 1AA"})
		require.NoError(t, err)
	})

	// Test 4: Both syntaxes produce same result
	t.Run("alias identical to original", func(t *testing.T) {
		type FormOriginal struct {
			Zip string `json:"zip" validate:"postcode=US"`
		}
		type FormAlias struct {
			Zip string `json:"zip" validate:"postcode_iso3166_alpha2=US"`
		}

		validatorOrig := pedantigo.New[FormOriginal]()
		validatorAlias := pedantigo.New[FormAlias]()

		testCases := []string{"90210", "12345-6789", "invalid", ""}
		for _, tc := range testCases {
			errOrig := validatorOrig.Validate(&FormOriginal{Zip: tc})
			errAlias := validatorAlias.Validate(&FormAlias{Zip: tc})

			if errOrig == nil {
				require.NoError(t, errAlias, "mismatch for input: %s", tc)
			} else {
				require.Error(t, errAlias, "mismatch for input: %s", tc)
			}
		}
	})

	// Test 5: Empty string passes (not required)
	t.Run("empty string passes", func(t *testing.T) {
		type Form struct {
			Zip string `json:"zip" validate:"postcode_iso3166_alpha2=US"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Zip: ""})
		require.NoError(t, err)
	})
}
