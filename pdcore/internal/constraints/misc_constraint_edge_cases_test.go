package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SmrutAI/pedantigo/v2/pdcore"
)

// TestISO3166NumericConstraint_EdgeCases tests edge cases for ISO 3166-1 numeric constraint.
func TestISO3166NumericConstraint_EdgeCases(t *testing.T) {
	t.Run("valid int country code", func(t *testing.T) {
		type Country struct {
			Code int `json:"code" validate:"iso3166_1_alpha_numeric"`
		}

		validator := New[Country]()
		data := Country{Code: 840} // USA

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("valid uint country code", func(t *testing.T) {
		type Country struct {
			Code uint `json:"code" validate:"iso3166_1_alpha_numeric"`
		}

		validator := New[Country]()
		data := Country{Code: 826} // GBR

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("uint country code exceeds 999", func(t *testing.T) {
		type Country struct {
			Code uint `json:"code" validate:"iso3166_1_alpha_numeric"`
		}

		validator := New[Country]()
		data := Country{Code: 1000} // Invalid - too large

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("invalid string type", func(t *testing.T) {
		type Country struct {
			Code string `json:"code" validate:"iso3166_1_alpha_numeric"`
		}

		validator := New[Country]()
		data := Country{Code: "840"}

		err := validator.Validate(&data)
		require.Error(t, err) // String type should fail
	})

	t.Run("invalid country code", func(t *testing.T) {
		type Country struct {
			Code int `json:"code" validate:"iso3166_1_alpha_numeric"`
		}

		validator := New[Country]()
		data := Country{Code: 999} // Invalid code

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestISO4217NumericConstraint_EdgeCases tests edge cases for ISO 4217 numeric constraint.
func TestISO4217NumericConstraint_EdgeCases(t *testing.T) {
	t.Run("valid int currency code", func(t *testing.T) {
		type Currency struct {
			Code int `json:"code" validate:"iso4217_numeric"`
		}

		validator := New[Currency]()
		data := Currency{Code: 840} // USD

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("valid uint currency code", func(t *testing.T) {
		type Currency struct {
			Code uint `json:"code" validate:"iso4217_numeric"`
		}

		validator := New[Currency]()
		data := Currency{Code: 978} // EUR

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("uint currency code exceeds 999", func(t *testing.T) {
		type Currency struct {
			Code uint `json:"code" validate:"iso4217_numeric"`
		}

		validator := New[Currency]()
		data := Currency{Code: 1000} // Invalid - too large

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("invalid string type", func(t *testing.T) {
		type Currency struct {
			Code string `json:"code" validate:"iso4217_numeric"`
		}

		validator := New[Currency]()
		data := Currency{Code: "840"}

		err := validator.Validate(&data)
		require.Error(t, err) // String type should fail
	})

	t.Run("invalid currency code", func(t *testing.T) {
		type Currency struct {
			Code int `json:"code" validate:"iso4217_numeric"`
		}

		validator := New[Currency]()
		data := Currency{Code: 1} // Invalid code (1 is not a valid ISO 4217 code)

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestLenConstraint tests the len constraint for exact length validation.
func TestLenConstraint_Coverage(t *testing.T) {
	t.Run("exact string length valid", func(t *testing.T) {
		type Data struct {
			Code string `json:"code" validate:"len=5"`
		}

		validator := New[Data]()
		data := Data{Code: "ABCDE"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("exact string length invalid short", func(t *testing.T) {
		type Data struct {
			Code string `json:"code" validate:"len=5"`
		}

		validator := New[Data]()
		data := Data{Code: "ABC"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("exact string length invalid long", func(t *testing.T) {
		type Data struct {
			Code string `json:"code" validate:"len=5"`
		}

		validator := New[Data]()
		data := Data{Code: "ABCDEFG"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestISSNConstraint_Coverage tests ISSN validation edge cases.
func TestISSNConstraint_Coverage(t *testing.T) {
	t.Run("valid ISSN", func(t *testing.T) {
		type Journal struct {
			ISSN string `json:"issn" validate:"issn"`
		}

		validator := New[Journal]()
		data := Journal{ISSN: "0317-8471"} // Valid ISSN

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("invalid ISSN wrong checksum", func(t *testing.T) {
		type Journal struct {
			ISSN string `json:"issn" validate:"issn"`
		}

		validator := New[Journal]()
		data := Journal{ISSN: "0317-8472"} // Wrong check digit

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("invalid ISSN format", func(t *testing.T) {
		type Journal struct {
			ISSN string `json:"issn" validate:"issn"`
		}

		validator := New[Journal]()
		data := Journal{ISSN: "invalid"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("empty ISSN skips validation", func(t *testing.T) {
		type Journal struct {
			ISSN string `json:"issn" validate:"issn"`
		}

		validator := New[Journal]()
		data := Journal{ISSN: ""}

		err := validator.Validate(&data)
		assert.NoError(t, err) // Empty strings skip validation
	})
}

// TestCronConstraint_Coverage tests cron expression validation edge cases.
func TestCronConstraint_Coverage(t *testing.T) {
	t.Run("valid cron expression", func(t *testing.T) {
		type Schedule struct {
			Cron string `json:"cron" validate:"cron"`
		}

		validator := New[Schedule]()
		data := Schedule{Cron: "0 0 * * *"} // Every day at midnight

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("valid cron with ranges", func(t *testing.T) {
		type Schedule struct {
			Cron string `json:"cron" validate:"cron"`
		}

		validator := New[Schedule]()
		data := Schedule{Cron: "0 9-17 * * 1-5"} // 9am-5pm weekdays

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("valid cron with steps", func(t *testing.T) {
		type Schedule struct {
			Cron string `json:"cron" validate:"cron"`
		}

		validator := New[Schedule]()
		data := Schedule{Cron: "*/15 * * * *"} // Every 15 minutes

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("invalid cron expression", func(t *testing.T) {
		type Schedule struct {
			Cron string `json:"cron" validate:"cron"`
		}

		validator := New[Schedule]()
		data := Schedule{Cron: "invalid cron"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("invalid cron range out of bounds", func(t *testing.T) {
		type Schedule struct {
			Cron string `json:"cron" validate:"cron"`
		}

		validator := New[Schedule]()
		data := Schedule{Cron: "0 25 * * *"} // Invalid hour (25 > 23)

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestOrConstraint_Coverage tests or constraint edge cases.
func TestOrConstraint_Coverage(t *testing.T) {
	t.Run("or constraint hexcolor|rgb first matches", func(t *testing.T) {
		type Data struct {
			Color string `json:"color" validate:"hexcolor|rgb"`
		}

		validator := New[Data]()
		data := Data{Color: "#FF5733"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("or constraint hexcolor|rgb second matches", func(t *testing.T) {
		type Data struct {
			Color string `json:"color" validate:"hexcolor|rgb"`
		}

		validator := New[Data]()
		data := Data{Color: "rgb(255, 87, 51)"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("or constraint hexcolor|rgb none matches", func(t *testing.T) {
		type Data struct {
			Color string `json:"color" validate:"hexcolor|rgb"`
		}

		validator := New[Data]()
		data := Data{Color: "not-a-color"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestColorConstraint_Coverage tests color constraint edge cases.
func TestColorConstraint_Coverage(t *testing.T) {
	t.Run("valid hex color", func(t *testing.T) {
		type Style struct {
			Color string `json:"color" validate:"hexcolor"`
		}

		validator := New[Style]()
		data := Style{Color: "#FF5733"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("valid short hex color", func(t *testing.T) {
		type Style struct {
			Color string `json:"color" validate:"hexcolor"`
		}

		validator := New[Style]()
		data := Style{Color: "#F53"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("invalid hex color", func(t *testing.T) {
		type Style struct {
			Color string `json:"color" validate:"hexcolor"`
		}

		validator := New[Style]()
		data := Style{Color: "red"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("valid rgb color", func(t *testing.T) {
		type Style struct {
			Color string `json:"color" validate:"rgb"`
		}

		validator := New[Style]()
		data := Style{Color: "rgb(255, 87, 51)"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("invalid rgb color out of range", func(t *testing.T) {
		type Style struct {
			Color string `json:"color" validate:"rgb"`
		}

		validator := New[Style]()
		data := Style{Color: "rgb(256, 87, 51)"} // 256 > 255

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("valid hsl color", func(t *testing.T) {
		type Style struct {
			Color string `json:"color" validate:"hsl"`
		}

		validator := New[Style]()
		data := Style{Color: "hsl(14, 100%, 53%)"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})
}

// TestUniqueConstraint_Coverage tests unique constraint edge cases.
func TestUniqueConstraint_Coverage(t *testing.T) {
	t.Run("unique slice valid", func(t *testing.T) {
		type Data struct {
			Items []string `json:"items" validate:"unique"`
		}

		validator := New[Data]()
		data := Data{Items: []string{"a", "b", "c"}}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("unique slice invalid duplicates", func(t *testing.T) {
		type Data struct {
			Items []string `json:"items" validate:"unique"`
		}

		validator := New[Data]()
		data := Data{Items: []string{"a", "b", "a"}}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("unique map values valid", func(t *testing.T) {
		type Data struct {
			Mapping map[string]int `json:"mapping" validate:"unique"`
		}

		validator := New[Data]()
		data := Data{Mapping: map[string]int{"a": 1, "b": 2, "c": 3}}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("unique map values invalid", func(t *testing.T) {
		type Data struct {
			Mapping map[string]int `json:"mapping" validate:"unique"`
		}

		validator := New[Data]()
		data := Data{Mapping: map[string]int{"a": 1, "b": 1, "c": 3}} // Duplicate value

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("unique slice of structs by field", func(t *testing.T) {
		type Item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		type Data struct {
			Items []Item `json:"items" validate:"unique=ID"`
		}

		validator := New[Data]()
		data := Data{Items: []Item{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("unique slice of structs by field invalid", func(t *testing.T) {
		type Item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		type Data struct {
			Items []Item `json:"items" validate:"unique=ID"`
		}

		validator := New[Data]()
		data := Data{Items: []Item{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}}} // Duplicate ID

		err := validator.Validate(&data)
		require.Error(t, err)
	})
}

// TestEncodingConstraint_Coverage tests encoding constraint edge cases.
func TestEncodingConstraint_Coverage(t *testing.T) {
	t.Run("valid base64 standard", func(t *testing.T) {
		type Data struct {
			Encoded string `json:"encoded" validate:"base64"`
		}

		validator := New[Data]()
		data := Data{Encoded: "SGVsbG8gV29ybGQ="}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("invalid base64", func(t *testing.T) {
		type Data struct {
			Encoded string `json:"encoded" validate:"base64"`
		}

		validator := New[Data]()
		data := Data{Encoded: "not-valid-base64!!!"}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("valid base64url", func(t *testing.T) {
		type Data struct {
			Encoded string `json:"encoded" validate:"base64url"`
		}

		validator := New[Data]()
		data := Data{Encoded: "SGVsbG8gV29ybGQ"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("valid base64rawurl", func(t *testing.T) {
		type Data struct {
			Encoded string `json:"encoded" validate:"base64rawurl"`
		}

		validator := New[Data]()
		data := Data{Encoded: "SGVsbG8gV29ybGQ"}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})
}

// TestNumericConstraint_EdgeCases tests numeric constraint edge cases.
func TestNumericConstraint_EdgeCases(t *testing.T) {
	t.Run("positive constraint valid", func(t *testing.T) {
		type Data struct {
			Value int `json:"value" validate:"positive"`
		}

		validator := New[Data]()
		data := Data{Value: 5}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("positive constraint invalid zero", func(t *testing.T) {
		type Data struct {
			Value int `json:"value" validate:"positive"`
		}

		validator := New[Data]()
		data := Data{Value: 0}

		err := validator.Validate(&data)
		require.Error(t, err)
	})

	t.Run("negative constraint valid", func(t *testing.T) {
		type Data struct {
			Value int `json:"value" validate:"negative"`
		}

		validator := New[Data]()
		data := Data{Value: -5}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("nonnegative constraint valid zero", func(t *testing.T) {
		type Data struct {
			Value int `json:"value" validate:"nonnegative"`
		}

		validator := New[Data]()
		data := Data{Value: 0}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})

	t.Run("nonpositive constraint valid zero", func(t *testing.T) {
		type Data struct {
			Value int `json:"value" validate:"nonpositive"`
		}

		validator := New[Data]()
		data := Data{Value: 0}

		err := validator.Validate(&data)
		assert.NoError(t, err)
	})
}
