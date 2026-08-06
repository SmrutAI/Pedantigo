package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	pedantigo "github.com/SmrutAI/pedantigo"
)

// TestOrConstraint_FirstMatches tests OR constraint when first option matches.
func TestOrConstraint_FirstMatches(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "#FF0000"})
	require.NoError(t, err)
}

// TestOrConstraint_SecondMatches tests OR constraint when second option matches.
func TestOrConstraint_SecondMatches(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "rgb(255,0,0)"})
	require.NoError(t, err)
}

// TestOrConstraint_NoMatch tests OR constraint when all options fail.
func TestOrConstraint_NoMatch(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "invalid"})
	require.Error(t, err)
}

// TestOrConstraint_TripleOR tests OR constraint with three options (third matches).
func TestOrConstraint_TripleOR_ThirdMatches(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb|rgba"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "rgba(255,0,0,0.5)"})
	require.NoError(t, err)
}

// TestOrConstraint_TripleOR_FirstMatches tests OR constraint with three options (first matches).
func TestOrConstraint_TripleOR_FirstMatches(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb|rgba"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "#ABC123"})
	require.NoError(t, err)
}

// TestOrConstraint_TripleOR_NoMatch tests OR constraint with three options (all fail).
func TestOrConstraint_TripleOR_NoMatch(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb|rgba"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "not-a-color"})
	require.Error(t, err)
}

// TestOrConstraint_EmptyString tests OR constraint with empty string (should pass, not required).
func TestOrConstraint_EmptyString(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: ""})
	require.NoError(t, err)
}

// TestOrConstraint_MixedTypes_EmailMatches tests OR with different constraint types (email matches).
func TestOrConstraint_MixedTypes_EmailMatches(t *testing.T) {
	type Form struct {
		Contact string `json:"contact" validate:"email|url"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Contact: "test@example.com"})
	require.NoError(t, err)
}

// TestOrConstraint_MixedTypes_URLMatches tests OR with different constraint types (url matches).
func TestOrConstraint_MixedTypes_URLMatches(t *testing.T) {
	type Form struct {
		Contact string `json:"contact" validate:"email|url"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Contact: "https://example.com"})
	require.NoError(t, err)
}

// TestOrConstraint_MixedTypes_NoMatch tests OR with different constraint types (both fail).
func TestOrConstraint_MixedTypes_NoMatch(t *testing.T) {
	type Form struct {
		Contact string `json:"contact" validate:"email|url"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Contact: "not-an-email-or-url"})
	require.Error(t, err)
}

// TestOrConstraint_WithRequired_Valid tests OR combined with required constraint (valid input).
func TestOrConstraint_WithRequired_Valid(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"required,hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "#FF0000"})
	require.NoError(t, err)
}

// TestOrConstraint_WithRequired_MissingFails tests OR combined with required constraint.
// Note: In pedantigo, 'required' only checks for missing JSON fields, not empty strings.
func TestOrConstraint_WithRequired_MissingFails(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"required,hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	// Test via Unmarshal - missing field should fail required
	_, err := validator.Unmarshal([]byte(`{}`))
	require.Error(t, err)
}

// TestOrConstraint_WithRequired_EmptyPasses tests that empty string passes (required only checks for missing).
func TestOrConstraint_WithRequired_EmptyPasses(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"required,hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	// Empty string passes - 'required' only checks for missing fields, not empty values
	_, err := validator.Unmarshal([]byte(`{"color": ""}`))
	require.NoError(t, err)
}

// TestOrConstraint_WithRequired_InvalidFails tests OR combined with required (invalid value fails OR).
func TestOrConstraint_WithRequired_InvalidFails(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"required,hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: "invalid"})
	require.Error(t, err)
}

// TestOrConstraint_NilPointer tests OR constraint with nil pointer (should pass).
func TestOrConstraint_NilPointer(t *testing.T) {
	type Form struct {
		Color *string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{Color: nil})
	require.NoError(t, err)
}

// TestOrConstraint_PointerToValidValue tests OR constraint with pointer to valid value.
func TestOrConstraint_PointerToValidValue(t *testing.T) {
	type Form struct {
		Color *string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	validColor := "#FF0000"
	err := validator.Validate(&Form{Color: &validColor})
	require.NoError(t, err)
}

// TestOrConstraint_PointerToInvalidValue tests OR constraint with pointer to invalid value.
func TestOrConstraint_PointerToInvalidValue(t *testing.T) {
	type Form struct {
		Color *string `json:"color" validate:"hexcolor|rgb"`
	}
	validator := pedantigo.New[Form]()
	invalidColor := "not-valid"
	err := validator.Validate(&Form{Color: &invalidColor})
	require.Error(t, err)
}

// TestOrConstraint_MultipleFields tests OR constraint across multiple struct fields.
func TestOrConstraint_MultipleFields(t *testing.T) {
	type Form struct {
		PrimaryColor   string `json:"primary_color" validate:"hexcolor|rgb"`
		SecondaryColor string `json:"secondary_color" validate:"hexcolor|rgba"`
		Contact        string `json:"contact" validate:"email|url"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{
		PrimaryColor:   "#FF0000",
		SecondaryColor: "rgba(0,255,0,0.8)",
		Contact:        "admin@example.com",
	})
	require.NoError(t, err)
}

// TestOrConstraint_MultipleFields_OneFails tests OR constraint with one field failing.
func TestOrConstraint_MultipleFields_OneFails(t *testing.T) {
	type Form struct {
		PrimaryColor   string `json:"primary_color" validate:"hexcolor|rgb"`
		SecondaryColor string `json:"secondary_color" validate:"hexcolor|rgba"`
	}
	validator := pedantigo.New[Form]()
	err := validator.Validate(&Form{
		PrimaryColor:   "#FF0000",
		SecondaryColor: "invalid-color",
	})
	require.Error(t, err)
}

// TestOrConstraint_WithUnmarshal tests OR constraint via Unmarshal (integration test).
func TestOrConstraint_WithUnmarshal(t *testing.T) {
	type Form struct {
		Color string `json:"color" validate:"hexcolor|rgb|rgba"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid hexcolor",
			json:    `{"color": "#FF0000"}`,
			wantErr: false,
		},
		{
			name:    "valid rgb",
			json:    `{"color": "rgb(255,0,0)"}`,
			wantErr: false,
		},
		{
			name:    "valid rgba",
			json:    `{"color": "rgba(255,0,0,0.5)"}`,
			wantErr: false,
		},
		{
			name:    "invalid color",
			json:    `{"color": "not-a-color"}`,
			wantErr: true,
		},
		{
			name:    "empty color",
			json:    `{"color": ""}`,
			wantErr: false,
		},
	}

	validator := pedantigo.New[Form]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Unmarshal([]byte(tt.json))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestOrConstraint_ComplexCombination tests OR with other string constraints.
func TestOrConstraint_ComplexCombination(t *testing.T) {
	type Form struct {
		Identifier string `json:"identifier" validate:"uuid|email"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid uuid v4",
			value:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid email",
			value:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "invalid both",
			value:   "not-uuid-or-email",
			wantErr: true,
		},
	}

	validator := pedantigo.New[Form]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&Form{Identifier: tt.value})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestOrConstraint_AlphaOrNumeric tests OR with character class constraints.
func TestOrConstraint_AlphaOrNumeric(t *testing.T) {
	type Form struct {
		Code string `json:"code" validate:"alpha|numeric"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "alpha only",
			value:   "ABCDEF",
			wantErr: false,
		},
		{
			name:    "numeric only",
			value:   "123456",
			wantErr: false,
		},
		{
			name:    "alphanumeric (fails both)",
			value:   "ABC123",
			wantErr: true,
		},
		{
			name:    "special chars (fails both)",
			value:   "ABC-123",
			wantErr: true,
		},
	}

	validator := pedantigo.New[Form]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&Form{Code: tt.value})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
