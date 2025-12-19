package tags

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseTag_ValidConstraints tests valid constraint parsing in table-driven format.
// Covers simple constraints, constraints with values, and multiple constraint combinations.
// TestParseTag_ValidConstraints tests ParseTag validconstraints.
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

// TestParseTagWithDive_CollectionConstraintsOnly tests parsing tags with only collection-level constraints.
func TestParseTagWithDive_CollectionConstraintsOnly(t *testing.T) {
	tests := []struct {
		name        string
		tag         reflect.StructTag
		wantNil     bool
		constraints map[string]string
	}{
		{
			name:        "single_min_constraint",
			tag:         reflect.StructTag(`pedantigo:"min=3"`),
			constraints: map[string]string{"min": "3"},
		},
		{
			name:        "multiple_collection_constraints",
			tag:         reflect.StructTag(`pedantigo:"min=3,max=10"`),
			constraints: map[string]string{"min": "3", "max": "10"},
		},
		{
			name:    "empty_tag",
			tag:     reflect.StructTag(`pedantigo:""`),
			wantNil: true,
		},
		{
			name:    "no_pedantigo_tag",
			tag:     reflect.StructTag(`json:"field"`),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTagWithDive(tt.tag)

			if tt.wantNil {
				assert.Nil(t, parsed)
				return
			}

			require.NotNil(t, parsed)
			assert.False(t, parsed.DivePresent)
			assert.Equal(t, tt.constraints, parsed.CollectionConstraints)
			assert.Empty(t, parsed.KeyConstraints)
			assert.Empty(t, parsed.ElementConstraints)
		})
	}
}

// TestParseTagWithDive_ElementConstraints tests parsing tags with dive and element constraints.
func TestParseTagWithDive_ElementConstraints(t *testing.T) {
	tests := []struct {
		name                  string
		tag                   reflect.StructTag
		wantDivePresent       bool
		collectionConstraints map[string]string
		elementConstraints    map[string]string
	}{
		{
			name:                  "dive_only_with_element_constraint",
			tag:                   reflect.StructTag(`pedantigo:"dive,email"`),
			wantDivePresent:       true,
			collectionConstraints: map[string]string{},
			elementConstraints:    map[string]string{"email": ""},
		},
		{
			name:                  "dive_with_multiple_element_constraints",
			tag:                   reflect.StructTag(`pedantigo:"dive,email,min=5"`),
			wantDivePresent:       true,
			collectionConstraints: map[string]string{},
			elementConstraints:    map[string]string{"email": "", "min": "5"},
		},
		{
			name:                  "collection_and_element_constraints",
			tag:                   reflect.StructTag(`pedantigo:"min=3,dive,min=5"`),
			wantDivePresent:       true,
			collectionConstraints: map[string]string{"min": "3"},
			elementConstraints:    map[string]string{"min": "5"},
		},
		{
			name:                  "collection_max_and_element_email",
			tag:                   reflect.StructTag(`pedantigo:"max=100,dive,email,required"`),
			wantDivePresent:       true,
			collectionConstraints: map[string]string{"max": "100"},
			elementConstraints:    map[string]string{"email": "", "required": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTagWithDive(tt.tag)

			require.NotNil(t, parsed)
			assert.Equal(t, tt.wantDivePresent, parsed.DivePresent)
			assert.Equal(t, tt.collectionConstraints, parsed.CollectionConstraints)
			assert.Equal(t, tt.elementConstraints, parsed.ElementConstraints)
			assert.Empty(t, parsed.KeyConstraints)
		})
	}
}

// TestParseTagWithDive_MapKeyConstraints tests parsing tags with keys/endkeys for map validation.
func TestParseTagWithDive_MapKeyConstraints(t *testing.T) {
	tests := []struct {
		name               string
		tag                reflect.StructTag
		keyConstraints     map[string]string
		elementConstraints map[string]string
	}{
		{
			name:               "keys_with_min_constraint",
			tag:                reflect.StructTag(`pedantigo:"dive,keys,min=2,endkeys,email"`),
			keyConstraints:     map[string]string{"min": "2"},
			elementConstraints: map[string]string{"email": ""},
		},
		{
			name:               "keys_with_multiple_constraints",
			tag:                reflect.StructTag(`pedantigo:"dive,keys,min=2,max=10,endkeys,required"`),
			keyConstraints:     map[string]string{"min": "2", "max": "10"},
			elementConstraints: map[string]string{"required": ""},
		},
		{
			name:               "keys_with_pattern",
			tag:                reflect.StructTag(`pedantigo:"dive,keys,pattern=^[a-z]+$,endkeys,min=1"`),
			keyConstraints:     map[string]string{"pattern": "^[a-z]+$"},
			elementConstraints: map[string]string{"min": "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTagWithDive(tt.tag)

			require.NotNil(t, parsed)
			assert.True(t, parsed.DivePresent)
			assert.Equal(t, tt.keyConstraints, parsed.KeyConstraints)
			assert.Equal(t, tt.elementConstraints, parsed.ElementConstraints)
		})
	}
}

// TestParseTagWithDive_Panics tests that invalid tag syntax panics.
func TestParseTagWithDive_Panics(t *testing.T) {
	tests := []struct {
		name          string
		tag           reflect.StructTag
		expectedPanic string
	}{
		{
			name:          "keys_without_dive",
			tag:           reflect.StructTag(`pedantigo:"keys,min=2,endkeys"`),
			expectedPanic: "'keys' can only appear after 'dive'",
		},
		{
			name:          "endkeys_without_keys",
			tag:           reflect.StructTag(`pedantigo:"dive,endkeys"`),
			expectedPanic: "'endkeys' without preceding 'keys'",
		},
		{
			name:          "keys_without_endkeys",
			tag:           reflect.StructTag(`pedantigo:"dive,keys,min=2"`),
			expectedPanic: "'keys' without closing 'endkeys'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PanicsWithValue(t, tt.expectedPanic, func() {
				ParseTagWithDive(tt.tag)
			})
		})
	}
}

// TestParseTagWithDive_WhitespaceHandling tests that whitespace is properly trimmed.
func TestParseTagWithDive_WhitespaceHandling(t *testing.T) {
	parsed := ParseTagWithDive(reflect.StructTag(`pedantigo:"  min = 3 , dive , email  "`))

	require.NotNil(t, parsed)
	assert.True(t, parsed.DivePresent)
	assert.Equal(t, "3", parsed.CollectionConstraints["min"])
	assert.Contains(t, parsed.ElementConstraints, "email")
}

// ============================================================================
// Tests for ParseTagWithName - Custom tag name support
// ============================================================================

// TestParseTagWithName_CustomTag tests parsing with a custom tag name.
func TestParseTagWithName_CustomTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        reflect.StructTag
		tagName    string
		wantKeys   map[string]string
		wantLength int
	}{
		{
			name:       "validate_tag_required",
			tag:        reflect.StructTag(`validate:"required"`),
			tagName:    "validate",
			wantKeys:   map[string]string{"required": ""},
			wantLength: 1,
		},
		{
			name:       "binding_tag_multiple_constraints",
			tag:        reflect.StructTag(`binding:"required,email,min=5"`),
			tagName:    "binding",
			wantKeys:   map[string]string{"required": "", "email": "", "min": "5"},
			wantLength: 3,
		},
		{
			name:       "custom_tag_with_json",
			tag:        reflect.StructTag(`json:"email" custom:"required,email"`),
			tagName:    "custom",
			wantKeys:   map[string]string{"required": "", "email": ""},
			wantLength: 2,
		},
		{
			name:       "pedantigo_tag_still_works",
			tag:        reflect.StructTag(`pedantigo:"required,min=3"`),
			tagName:    "pedantigo",
			wantKeys:   map[string]string{"required": "", "min": "3"},
			wantLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := ParseTagWithName(tt.tag, tt.tagName)

			require.NotNil(t, constraints, "expected constraints map, got nil")
			assert.Len(t, constraints, tt.wantLength)

			for key, expectedVal := range tt.wantKeys {
				val, ok := constraints[key]
				require.True(t, ok, "expected constraint key %q", key)
				assert.Equal(t, expectedVal, val)
			}
		})
	}
}

// TestParseTagWithName_WrongTag_ReturnsNil tests that wrong tag name returns nil.
func TestParseTagWithName_WrongTag_ReturnsNil(t *testing.T) {
	tests := []struct {
		name    string
		tag     reflect.StructTag
		tagName string
	}{
		{
			name:    "looking_for_validate_but_has_pedantigo",
			tag:     reflect.StructTag(`pedantigo:"required"`),
			tagName: "validate",
		},
		{
			name:    "looking_for_binding_but_has_validate",
			tag:     reflect.StructTag(`validate:"required"`),
			tagName: "binding",
		},
		{
			name:    "no_matching_tag_at_all",
			tag:     reflect.StructTag(`json:"email"`),
			tagName: "validate",
		},
		{
			name:    "empty_tag_value",
			tag:     reflect.StructTag(`validate:""`),
			tagName: "validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := ParseTagWithName(tt.tag, tt.tagName)
			assert.Nil(t, constraints, "expected nil for non-matching tag")
		})
	}
}

// TestParseTagWithDiveAndName_CustomTag tests ParseTagWithDive with custom tag names.
func TestParseTagWithDiveAndName_CustomTag(t *testing.T) {
	tests := []struct {
		name                  string
		tag                   reflect.StructTag
		tagName               string
		wantDivePresent       bool
		collectionConstraints map[string]string
		elementConstraints    map[string]string
	}{
		{
			name:                  "validate_tag_with_dive",
			tag:                   reflect.StructTag(`validate:"min=3,dive,email"`),
			tagName:               "validate",
			wantDivePresent:       true,
			collectionConstraints: map[string]string{"min": "3"},
			elementConstraints:    map[string]string{"email": ""},
		},
		{
			name:                  "binding_tag_collection_only",
			tag:                   reflect.StructTag(`binding:"min=5,max=10"`),
			tagName:               "binding",
			wantDivePresent:       false,
			collectionConstraints: map[string]string{"min": "5", "max": "10"},
			elementConstraints:    map[string]string{},
		},
		{
			name:                  "custom_tag_with_multiple_element_constraints",
			tag:                   reflect.StructTag(`custom:"dive,required,email,min=5"`),
			tagName:               "custom",
			wantDivePresent:       true,
			collectionConstraints: map[string]string{},
			elementConstraints:    map[string]string{"required": "", "email": "", "min": "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTagWithDiveAndName(tt.tag, tt.tagName)

			require.NotNil(t, parsed)
			assert.Equal(t, tt.wantDivePresent, parsed.DivePresent)
			assert.Equal(t, tt.collectionConstraints, parsed.CollectionConstraints)
			assert.Equal(t, tt.elementConstraints, parsed.ElementConstraints)
		})
	}
}

// TestParseTagWithDiveAndName_WrongTag_ReturnsNil tests that wrong tag returns nil.
func TestParseTagWithDiveAndName_WrongTag_ReturnsNil(t *testing.T) {
	tests := []struct {
		name    string
		tag     reflect.StructTag
		tagName string
	}{
		{
			name:    "looking_for_validate_but_has_pedantigo",
			tag:     reflect.StructTag(`pedantigo:"dive,email"`),
			tagName: "validate",
		},
		{
			name:    "empty_tag_value",
			tag:     reflect.StructTag(`validate:""`),
			tagName: "validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTagWithDiveAndName(tt.tag, tt.tagName)
			assert.Nil(t, parsed, "expected nil for non-matching tag")
		})
	}
}

// TestParseTag_DelegatestoParseTagWithName verifies ParseTag uses default "pedantigo" tag.
func TestParseTag_DelegatesToParseTagWithName(t *testing.T) {
	tag := reflect.StructTag(`pedantigo:"required,email"`)

	// Both should return identical results
	fromParseTag := ParseTag(tag)
	fromParseTagWithName := ParseTagWithName(tag, "pedantigo")

	assert.Equal(t, fromParseTag, fromParseTagWithName)
}

// TestParseTagWithDive_DelegatesToParseTagWithDiveAndName verifies delegation.
func TestParseTagWithDive_DelegatesToParseTagWithDiveAndName(t *testing.T) {
	tag := reflect.StructTag(`pedantigo:"min=3,dive,email"`)

	// Both should return identical results
	fromParseTagWithDive := ParseTagWithDive(tag)
	fromParseTagWithDiveAndName := ParseTagWithDiveAndName(tag, "pedantigo")

	assert.Equal(t, fromParseTagWithDive.DivePresent, fromParseTagWithDiveAndName.DivePresent)
	assert.Equal(t, fromParseTagWithDive.CollectionConstraints, fromParseTagWithDiveAndName.CollectionConstraints)
	assert.Equal(t, fromParseTagWithDive.ElementConstraints, fromParseTagWithDiveAndName.ElementConstraints)
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
