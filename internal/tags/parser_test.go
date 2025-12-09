package tags

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseTag_ValidConstraints tests valid constraint parsing in table-driven format.
// Covers simple constraints, constraints with values, and multiple constraint combinations.
func TestParseTag_ValidConstraints(t *testing.T) {
	tests := []struct {
		name       string
		tag        reflect.StructTag
		wantKeys   map[string]string // constraint key -> expected value (empty string for simple constraints)
		wantLength int               // expected number of constraints
	}{
		{
			name:       "single_simple_constraint_required",
			tag:        reflect.StructTag(`pedantigo:"required"`),
			wantKeys:   map[string]string{"required": ""},
			wantLength: 1,
		},
		{
			name:       "single_simple_constraint_email",
			tag:        reflect.StructTag(`pedantigo:"email"`),
			wantKeys:   map[string]string{"email": ""},
			wantLength: 1,
		},
		{
			name:       "multiple_simple_constraints",
			tag:        reflect.StructTag(`pedantigo:"required,email"`),
			wantKeys:   map[string]string{"required": "", "email": ""},
			wantLength: 2,
		},
		{
			name:       "single_constraint_with_value_min",
			tag:        reflect.StructTag(`pedantigo:"min=18"`),
			wantKeys:   map[string]string{"min": "18"},
			wantLength: 1,
		},
		{
			name:       "single_constraint_with_value_default",
			tag:        reflect.StructTag(`pedantigo:"default=active"`),
			wantKeys:   map[string]string{"default": "active"},
			wantLength: 1,
		},
		{
			name:       "multiple_constraints_with_values",
			tag:        reflect.StructTag(`pedantigo:"min=18,max=120"`),
			wantKeys:   map[string]string{"min": "18", "max": "120"},
			wantLength: 2,
		},
		{
			name:       "mixed_simple_and_valued_constraints",
			tag:        reflect.StructTag(`pedantigo:"required,email,min=18"`),
			wantKeys:   map[string]string{"required": "", "email": "", "min": "18"},
			wantLength: 3,
		},
		{
			name:       "constraint_value_with_alphanumeric",
			tag:        reflect.StructTag(`pedantigo:"pattern=[a-z]+"`),
			wantKeys:   map[string]string{"pattern": "[a-z]+"},
			wantLength: 1,
		},
		{
			name:       "constraint_with_whitespace_around_equals",
			tag:        reflect.StructTag(`pedantigo:"min = 5 , max = 10"`),
			wantKeys:   map[string]string{"min": "5", "max": "10"},
			wantLength: 2,
		},
		{
			name:       "constraints_with_trailing_comma",
			tag:        reflect.StructTag(`pedantigo:"required,email,"`),
			wantKeys:   map[string]string{"required": "", "email": ""},
			wantLength: 2,
		},
		{
			name:       "complex_combination_all_types",
			tag:        reflect.StructTag(`pedantigo:"required,email,min=3,max=100,default=user@example.com"`),
			wantKeys:   map[string]string{"required": "", "email": "", "min": "3", "max": "100", "default": "user@example.com"},
			wantLength: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := ParseTag(tt.tag)

			// Check non-nil for valid pedantigo tags
			require.NotNil(t, constraints, "expected constraints map, got nil")

			// Check length
			assert.Len(t, constraints, tt.wantLength, "expected %d constraints, got %d", tt.wantLength, len(constraints))

			// Check each expected key and value
			for key, expectedVal := range tt.wantKeys {
				val, ok := constraints[key]
				require.True(t, ok, "expected constraint key %q, not found in %v", key, constraints)
				assert.Equal(t, expectedVal, val, "constraint %q: expected value %q, got %q", key, expectedVal, val)
			}
		})
	}
}

// TestParseTag_InvalidInputs tests edge cases and missing/invalid tags.
func TestParseTag_InvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		tag       reflect.StructTag
		wantNil   bool              // whether expecting nil result
		wantEmpty bool              // whether expecting empty map
		wantKeys  map[string]string // constraints to verify (if applicable)
	}{
		{
			name:      "no_pedantigo_tag",
			tag:       reflect.StructTag(`json:"email"`),
			wantNil:   true,
			wantEmpty: false,
		},
		{
			name:      "empty_struct_tag",
			tag:       reflect.StructTag(``),
			wantNil:   true,
			wantEmpty: false,
		},
		{
			name:      "pedantigo_with_empty_value",
			tag:       reflect.StructTag(`pedantigo:""`),
			wantNil:   true,
			wantEmpty: false,
		},
		{
			name:      "only_whitespace_in_tag",
			tag:       reflect.StructTag(`pedantigo:"   "`),
			wantNil:   false,
			wantEmpty: true,
			wantKeys:  map[string]string{},
		},
		{
			name:      "multiple_other_tags_no_pedantigo",
			tag:       reflect.StructTag(`json:"name" db:"user_name" sql:"varchar(255)"`),
			wantNil:   true,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := ParseTag(tt.tag)

			// Check nil expectation
			if tt.wantNil {
				assert.Nil(t, constraints, "expected nil constraints, got %v", constraints)
				return
			}
			require.NotNil(t, constraints, "expected non-nil constraints, got nil")

			// Check empty expectation
			if tt.wantEmpty {
				assert.Empty(t, constraints, "expected empty constraints, got %v", constraints)
			} else {
				assert.NotEmpty(t, constraints, "expected non-empty constraints, got empty")
			}

			// Verify any specified keys
			for key, expectedVal := range tt.wantKeys {
				val, ok := constraints[key]
				require.True(t, ok, "expected constraint key %q, not found in %v", key, constraints)
				assert.Equal(t, expectedVal, val, "constraint %q: expected value %q, got %q", key, expectedVal, val)
			}
		})
	}
}
