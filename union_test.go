package pedantigo

import (
	"reflect"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test discriminator values.
const (
	discCat = "cat"
	discDog = "dog"
)

// Test structs for discriminated union validation

// Cat represents a cat variant in the pet union.
type Cat struct {
	Name  string `json:"name" validate:"required"`
	Lives int    `json:"lives" validate:"min=1,max=9"`
}

// Dog represents a dog variant in the pet union.
type Dog struct {
	Name  string `json:"name" validate:"required"`
	Breed string `json:"breed"`
}

// Bird represents a bird variant in the pet union.
type Bird struct {
	Name string `json:"name" validate:"required"`
	Eggs int    `json:"eggs" validate:"min=0"`
}

// Pet is the union container type (for type parameter T).
type Pet struct{}

// TestNewUnion_EmptyDiscriminatorField tests that NewUnion returns error for empty discriminator field.
func TestNewUnion_EmptyDiscriminatorField(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
		},
	}

	validator, err := NewUnion[Pet](opts)

	require.Error(t, err)
	assert.Nil(t, validator)
	assert.Equal(t, "discriminator field is required", err.Error())
}

// TestNewUnion_EmptyVariants tests that NewUnion returns error for empty variants list.
func TestNewUnion_EmptyVariants(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "pet_type",
		Variants:           []UnionVariant{},
	}

	validator, err := NewUnion[Pet](opts)

	require.Error(t, err)
	assert.Nil(t, validator)
	assert.Equal(t, "at least one variant is required", err.Error())
}

// TestNewUnion_EmptyVariantDiscriminatorValue tests that NewUnion returns error
// when a variant has empty discriminator value.
func TestNewUnion_EmptyVariantDiscriminatorValue(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			{
				DiscriminatorValue: "",
				Type:               reflect.TypeOf(Cat{}),
			},
		},
	}

	validator, err := NewUnion[Pet](opts)

	require.Error(t, err)
	assert.Nil(t, validator)
	assert.Equal(t, "variant discriminator value cannot be empty", err.Error())
}

// TestNewUnion_NilVariantType tests that NewUnion returns error when variant type is nil.
func TestNewUnion_NilVariantType(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			{
				DiscriminatorValue: "cat",
				Type:               nil,
			},
		},
	}

	validator, err := NewUnion[Pet](opts)

	require.Error(t, err)
	assert.Nil(t, validator)
	assert.Equal(t, "variant type cannot be nil", err.Error())
}

// TestNewUnion_DuplicateDiscriminatorValues tests that NewUnion returns error
// when multiple variants have the same discriminator value.
func TestNewUnion_DuplicateDiscriminatorValues(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("cat"), // Duplicate discriminator value
		},
	}

	validator, err := NewUnion[Pet](opts)

	require.Error(t, err)
	assert.Nil(t, validator)
	assert.Equal(t, "duplicate discriminator value: cat", err.Error())
}

// TestNewUnion_ValidOptions tests that NewUnion succeeds with valid options.
func TestNewUnion_ValidOptions(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
			VariantFor[Bird]("bird"),
		},
	}

	validator, err := NewUnion[Pet](opts)

	require.NoError(t, err)
	require.NotNil(t, validator)
	assert.Equal(t, "pet_type", validator.options.DiscriminatorField)
	assert.Len(t, validator.variants, 3)
}

// TestNewUnion_SingleVariant tests that NewUnion works with a single variant.
func TestNewUnion_SingleVariant(t *testing.T) {
	opts := UnionOptions{
		DiscriminatorField: "type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
		},
	}

	validator, err := NewUnion[Pet](opts)

	require.NoError(t, err)
	require.NotNil(t, validator)
	assert.Len(t, validator.variants, 1)
}

// TestVariantFor_CreatesCorrectDiscriminatorValue tests that VariantFor helper
// creates a variant with the correct discriminator value.
func TestVariantFor_CreatesCorrectDiscriminatorValue(t *testing.T) {
	tests := []struct {
		name          string
		discriminator string
	}{
		{name: "cat variant", discriminator: "cat"},
		{name: "dog variant", discriminator: "dog"},
		{name: "bird variant", discriminator: "bird"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variant := VariantFor[Cat](tt.discriminator)
			assert.Equal(t, tt.discriminator, variant.DiscriminatorValue)
		})
	}
}

// TestVariantFor_CreatesCorrectType tests that VariantFor helper creates
// a variant with the correct reflect.Type.
func TestVariantFor_CreatesCorrectType(t *testing.T) {
	tests := []struct {
		name         string
		variantFunc  func() UnionVariant
		expectedType reflect.Type
	}{
		{
			name: "cat type",
			variantFunc: func() UnionVariant {
				return VariantFor[Cat]("cat")
			},
			expectedType: reflect.TypeOf(Cat{}),
		},
		{
			name: "dog type",
			variantFunc: func() UnionVariant {
				return VariantFor[Dog]("dog")
			},
			expectedType: reflect.TypeOf(Dog{}),
		},
		{
			name: "bird type",
			variantFunc: func() UnionVariant {
				return VariantFor[Bird]("bird")
			},
			expectedType: reflect.TypeOf(Bird{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variant := tt.variantFunc()
			assert.Equal(t, tt.expectedType, variant.Type)
		})
	}
}

// TestUnmarshal_ValidJSONWithKnownDiscriminator tests that Unmarshal correctly
// unmarshals JSON with a known discriminator value to the correct variant type.
func TestUnmarshal_ValidJSONWithKnownDiscriminator(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name            string
		data            []byte
		expectedVariant string
		expectedType    reflect.Type
		shouldValidate  bool
	}{
		{
			name:            "valid cat json",
			data:            []byte(`{"pet_type":"cat","name":"Whiskers","lives":7}`),
			expectedVariant: "cat",
			expectedType:    reflect.TypeOf(Cat{}),
			shouldValidate:  true,
		},
		{
			name:            "valid dog json",
			data:            []byte(`{"pet_type":"dog","name":"Buddy","breed":"Golden Retriever"}`),
			expectedVariant: "dog",
			expectedType:    reflect.TypeOf(Dog{}),
			shouldValidate:  true,
		},
		{
			name:            "cat with minimal fields",
			data:            []byte(`{"pet_type":"cat","name":"Mittens","lives":1}`),
			expectedVariant: "cat",
			expectedType:    reflect.TypeOf(Cat{}),
			shouldValidate:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, reflect.TypeOf(result))
		})
	}
}

// TestUnmarshal_MissingDiscriminatorField tests that Unmarshal returns appropriate error
// when discriminator field is missing from JSON.
func TestUnmarshal_MissingDiscriminatorField(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "completely missing discriminator",
			data: []byte(`{"name":"Whiskers","lives":7}`),
		},
		{
			name: "empty object",
			data: []byte(`{}`),
		},
		{
			name: "null discriminator field",
			data: []byte(`{"pet_type":null,"name":"Whiskers"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			require.Error(t, err)
			assert.Nil(t, result)
			// Error message should reference missing discriminator
			assert.Contains(t, err.Error(), "pet_type")
		})
	}
}

// TestUnmarshal_UnknownDiscriminatorValue tests that Unmarshal returns appropriate error
// when discriminator value doesn't match any variant.
func TestUnmarshal_UnknownDiscriminatorValue(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name                 string
		data                 []byte
		unknownDiscriminator string
	}{
		{
			name:                 "unknown variant fish",
			data:                 []byte(`{"pet_type":"fish","name":"Nemo"}`),
			unknownDiscriminator: "fish",
		},
		{
			name:                 "typo in discriminator",
			data:                 []byte(`{"pet_type":"catt","name":"Whiskers","lives":7}`),
			unknownDiscriminator: "catt",
		},
		{
			name:                 "empty string discriminator",
			data:                 []byte(`{"pet_type":"","name":"Unknown"}`),
			unknownDiscriminator: "",
		},
		{
			name:                 "number as discriminator",
			data:                 []byte(`{"pet_type":123,"name":"Thing"}`),
			unknownDiscriminator: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			require.Error(t, err)
			assert.Nil(t, result)
			// Error message should mention unknown discriminator and field name
			assert.Contains(t, err.Error(), "pet_type")
		})
	}
}

// TestUnionUnmarshal_InvalidJSON tests that Unmarshal returns error for invalid JSON.
func TestUnionUnmarshal_InvalidJSON(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "malformed json",
			data: []byte(`{"pet_type":"cat"invalid`),
		},
		{
			name: "empty data",
			data: []byte(``),
		},
		{
			name: "not json",
			data: []byte(`not json at all`),
		},
		{
			name: "truncated json",
			data: []byte(`{"pet_type":"cat","name":"Whiskers"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			require.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

// TestUnmarshal_ValidatesVariantAfterUnmarshal tests that Unmarshal applies
// validation constraints to the unmarshaled variant.
func TestUnmarshal_ValidatesVariantAfterUnmarshal(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          []byte
		shouldSucceed bool
		description   string
	}{
		{
			name:          "valid cat with name and valid lives",
			data:          []byte(`{"pet_type":"cat","name":"Whiskers","lives":5}`),
			shouldSucceed: true,
			description:   "passes validation",
		},
		{
			name:          "cat with missing required name",
			data:          []byte(`{"pet_type":"cat","lives":5}`),
			shouldSucceed: false,
			description:   "fails because name is required",
		},
		{
			name:          "cat with lives below min constraint",
			data:          []byte(`{"pet_type":"cat","name":"Whiskers","lives":0}`),
			shouldSucceed: false,
			description:   "fails because lives must be min=1",
		},
		{
			name:          "cat with lives above max constraint",
			data:          []byte(`{"pet_type":"cat","name":"Whiskers","lives":10}`),
			shouldSucceed: false,
			description:   "fails because lives must be max=9",
		},
		{
			name:          "valid dog with name",
			data:          []byte(`{"pet_type":"dog","name":"Buddy"}`),
			shouldSucceed: true,
			description:   "passes validation with optional breed",
		},
		{
			name:          "dog with missing required name",
			data:          []byte(`{"pet_type":"dog","breed":"Labrador"}`),
			shouldSucceed: false,
			description:   "fails because name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			if tt.shouldSucceed {
				require.NoError(t, err, tt.description)
				assert.NotNil(t, result, tt.description)
			} else {
				require.Error(t, err, tt.description)
			}
		})
	}
}

// TestValidate_ValidVariantPassesValidation tests that Validate passes
// when given a valid variant instance.
func TestValidate_ValidVariantPassesValidation(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name   string
		object any
	}{
		{
			name: "valid cat",
			object: Cat{
				Name:  "Whiskers",
				Lives: 5,
			},
		},
		{
			name: "valid dog",
			object: Dog{
				Name:  "Buddy",
				Breed: "Golden Retriever",
			},
		},
		{
			name: "cat with max lives",
			object: Cat{
				Name:  "Mittens",
				Lives: 9,
			},
		},
		{
			name: "cat with min lives",
			object: Cat{
				Name:  "Garfield",
				Lives: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.object)
			require.NoError(t, err)
		})
	}
}

// TestValidate_InvalidVariantFailsValidation tests that Validate fails
// when given an invalid variant instance with constraint violations.
func TestValidate_InvalidVariantFailsValidation(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		object      any
		description string
	}{
		{
			name: "cat with missing name",
			object: Cat{
				Name:  "",
				Lives: 5,
			},
			description: "name is required",
		},
		{
			name: "cat with lives below min",
			object: Cat{
				Name:  "Whiskers",
				Lives: 0,
			},
			description: "lives must be >= 1",
		},
		{
			name: "cat with lives above max",
			object: Cat{
				Name:  "Whiskers",
				Lives: 10,
			},
			description: "lives must be <= 9",
		},
		{
			name: "dog with missing name",
			object: Dog{
				Name:  "",
				Breed: "Poodle",
			},
			description: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.object)
			require.Error(t, err, tt.description)
		})
	}
}

// TestValidate_WrongTypeFailsValidation tests that Validate fails
// when given a type that is not one of the union variants.
func TestValidate_WrongTypeFailsValidation(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name   string
		object any
	}{
		{
			name:   "bird type not in union",
			object: Bird{Name: "Tweety", Eggs: 3},
		},
		{
			name:   "string type",
			object: "not a variant",
		},
		{
			name:   "integer type",
			object: 42,
		},
		{
			name:   "nil value",
			object: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.object)
			require.Error(t, err)
		})
	}
}

// TestUnmarshal_MultipleVariants tests Unmarshal with multiple variant types
// to ensure correct routing based on discriminator value.
func TestUnmarshal_MultipleVariants(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
			VariantFor[Bird]("bird"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          []byte
		expectedType  reflect.Type
		shouldSucceed bool
	}{
		{
			name:          "cat variant",
			data:          []byte(`{"pet_type":"cat","name":"Whiskers","lives":7}`),
			expectedType:  reflect.TypeOf(Cat{}),
			shouldSucceed: true,
		},
		{
			name:          "dog variant",
			data:          []byte(`{"pet_type":"dog","name":"Buddy","breed":"Lab"}`),
			expectedType:  reflect.TypeOf(Dog{}),
			shouldSucceed: true,
		},
		{
			name:          "bird variant",
			data:          []byte(`{"pet_type":"bird","name":"Tweety","eggs":4}`),
			expectedType:  reflect.TypeOf(Bird{}),
			shouldSucceed: true,
		},
		{
			name:          "unknown variant",
			data:          []byte(`{"pet_type":"snake","name":"Sly"}`),
			expectedType:  nil,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedType, reflect.TypeOf(result))
			} else {
				require.Error(t, err)
				assert.Nil(t, result)
			}
		})
	}
}

// TestUnmarshal_DiscriminatorFieldWithDifferentNames tests that Unmarshal
// correctly uses different discriminator field names.
func TestUnmarshal_DiscriminatorFieldWithDifferentNames(t *testing.T) {
	tests := []struct {
		name               string
		discriminatorField string
		data               []byte
		shouldSucceed      bool
	}{
		{
			name:               "pet_type discriminator",
			discriminatorField: "pet_type",
			data:               []byte(`{"pet_type":"cat","name":"Whiskers","lives":7}`),
			shouldSucceed:      true,
		},
		{
			name:               "type discriminator",
			discriminatorField: "type",
			data:               []byte(`{"type":"cat","name":"Whiskers","lives":7}`),
			shouldSucceed:      true,
		},
		{
			name:               "kind discriminator",
			discriminatorField: "kind",
			data:               []byte(`{"kind":"cat","name":"Whiskers","lives":7}`),
			shouldSucceed:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewUnion[Pet](UnionOptions{
				DiscriminatorField: tt.discriminatorField,
				Variants: []UnionVariant{
					VariantFor[Cat]("cat"),
				},
			})
			require.NoError(t, err)

			result, err := validator.Unmarshal(tt.data)

			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestValidate_MultipleVariantTypes tests Validate with different variant types
// to ensure proper routing and validation per variant.
func TestValidate_MultipleVariantTypes(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
			VariantFor[Bird]("bird"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		object        any
		shouldSucceed bool
		description   string
	}{
		{
			name: "valid cat",
			object: Cat{
				Name:  "Whiskers",
				Lives: 7,
			},
			shouldSucceed: true,
			description:   "cat with valid fields",
		},
		{
			name: "valid dog",
			object: Dog{
				Name:  "Buddy",
				Breed: "Golden",
			},
			shouldSucceed: true,
			description:   "dog with valid fields",
		},
		{
			name: "valid bird",
			object: Bird{
				Name: "Tweety",
				Eggs: 4,
			},
			shouldSucceed: true,
			description:   "bird with valid fields",
		},
		{
			name: "invalid cat - missing name",
			object: Cat{
				Name:  "",
				Lives: 7,
			},
			shouldSucceed: false,
			description:   "cat missing required name",
		},
		{
			name: "invalid dog - missing name",
			object: Dog{
				Name:  "",
				Breed: "Poodle",
			},
			shouldSucceed: false,
			description:   "dog missing required name",
		},
		{
			name: "invalid bird - negative eggs",
			object: Bird{
				Name: "Tweety",
				Eggs: -1,
			},
			shouldSucceed: false,
			description:   "bird with negative eggs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.object)

			if tt.shouldSucceed {
				require.NoError(t, err, tt.description)
			} else {
				require.Error(t, err, tt.description)
			}
		})
	}
}

// TestUnmarshal_ExtraFieldsInJSON tests that Unmarshal handles extra fields
// in JSON that aren't part of the variant struct.
func TestUnmarshal_ExtraFieldsInJSON(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
		},
	})
	require.NoError(t, err)

	data := []byte(`{
		"pet_type": "cat",
		"name": "Whiskers",
		"lives": 7,
		"extra_field": "should be ignored",
		"another_extra": 123
	}`)

	result, err := validator.Unmarshal(data)

	// Should succeed - extra fields are typically ignored in JSON unmarshaling
	// unless strict validation is enabled
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestUnmarshal_CaseSensitivityOfDiscriminator tests discriminator value matching.
func TestUnmarshal_CaseSensitivityOfDiscriminator(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          []byte
		shouldSucceed bool
		description   string
	}{
		{
			name:          "exact match lowercase",
			data:          []byte(`{"pet_type":"cat","name":"Whiskers","lives":7}`),
			shouldSucceed: true,
			description:   "lowercase cat matches",
		},
		{
			name:          "uppercase cat",
			data:          []byte(`{"pet_type":"CAT","name":"Whiskers","lives":7}`),
			shouldSucceed: false,
			description:   "uppercase CAT does not match lowercase cat",
		},
		{
			name:          "mixed case cat",
			data:          []byte(`{"pet_type":"Cat","name":"Whiskers","lives":7}`),
			shouldSucceed: false,
			description:   "mixed case Cat does not match lowercase cat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			if tt.shouldSucceed {
				require.NoError(t, err, tt.description)
				assert.NotNil(t, result, tt.description)
			} else {
				require.Error(t, err, tt.description)
				assert.Nil(t, result, tt.description)
			}
		})
	}
}

// TestUnmarshal_DiscriminatorValueNotString tests discriminator values that are not strings.
func TestUnmarshal_DiscriminatorValueNotString(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("1"),
			VariantFor[Dog]("2"),
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          []byte
		shouldSucceed bool
		description   string
	}{
		{
			name:          "numeric discriminator as string",
			data:          []byte(`{"pet_type":"1","name":"Whiskers","lives":7}`),
			shouldSucceed: true,
			description:   "numeric string discriminator",
		},
		{
			name:          "numeric discriminator as number",
			data:          []byte(`{"pet_type":1,"name":"Whiskers","lives":7}`),
			shouldSucceed: true,
			description:   "numeric discriminator converted to string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal(tt.data)

			if tt.shouldSucceed {
				require.NoError(t, err, tt.description)
				assert.NotNil(t, result, tt.description)
			} else {
				require.Error(t, err, tt.description)
			}
		})
	}
}

// TestUnionValidator_Schema_TwoVariants tests Schema() returns oneOf with two variants.
func TestUnionValidator_Schema_TwoVariants(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	schema := validator.Schema()

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
	assert.Len(t, schema.OneOf, 2, "oneOf should contain 2 variant schemas")
}

// TestUnionValidator_Schema_ThreeVariants tests Schema() returns oneOf with three variants.
func TestUnionValidator_Schema_ThreeVariants(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
			VariantFor[Bird]("bird"),
		},
	})
	require.NoError(t, err)

	schema := validator.Schema()

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
	assert.Len(t, schema.OneOf, 3, "oneOf should contain 3 variant schemas")
}

// TestUnionValidator_Schema_SingleVariant tests Schema() returns oneOf with single variant.
func TestUnionValidator_Schema_SingleVariant(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
		},
	})
	require.NoError(t, err)

	schema := validator.Schema()

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
	assert.Len(t, schema.OneOf, 1, "oneOf should contain 1 variant schema")
}

// TestUnionValidator_Schema_DiscriminatorFieldWithConst tests each variant
// schema has discriminator field with const constraint.
func TestUnionValidator_Schema_DiscriminatorFieldWithConst(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	schema := validator.Schema()

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
	require.Len(t, schema.OneOf, 2, "oneOf should have 2 schemas")

	// Find cat and dog schemas by discriminator const value (order may vary due to map iteration)
	var catSchema, dogSchema *jsonschema.Schema
	for _, vs := range schema.OneOf {
		if vs.Properties != nil {
			if pt, ok := vs.Properties.Get("pet_type"); ok {
				switch pt.Const {
				case discCat:
					catSchema = vs
				case discDog:
					dogSchema = vs
				}
			}
		}
	}

	// Verify cat schema
	require.NotNil(t, catSchema, "cat schema should exist")
	require.NotNil(t, catSchema.Properties, "cat schema properties should not be nil")
	petTypeSchema, ok := catSchema.Properties.Get("pet_type")
	require.True(t, ok, "pet_type property should exist in cat schema")
	assert.Equal(t, "cat", petTypeSchema.Const, "cat schema pet_type should have const 'cat'")

	// Verify dog schema
	require.NotNil(t, dogSchema, "dog schema should exist")
	require.NotNil(t, dogSchema.Properties, "dog schema properties should not be nil")
	petTypeSchema, ok = dogSchema.Properties.Get("pet_type")
	require.True(t, ok, "pet_type property should exist in dog schema")
	assert.Equal(t, "dog", petTypeSchema.Const, "dog schema pet_type should have const 'dog'")
}

// TestUnionValidator_Schema_VariantPropertiesIncluded tests each variant
// schema includes properties from the variant struct.
func TestUnionValidator_Schema_VariantPropertiesIncluded(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	schema := validator.Schema()

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
	require.Len(t, schema.OneOf, 2)

	// Find cat and dog schemas by discriminator const value (order may vary)
	var catSchema, dogSchema *jsonschema.Schema
	for _, vs := range schema.OneOf {
		if vs.Properties != nil {
			if pt, ok := vs.Properties.Get("pet_type"); ok {
				switch pt.Const {
				case discCat:
					catSchema = vs
				case discDog:
					dogSchema = vs
				}
			}
		}
	}

	// Check cat variant has name and lives properties
	require.NotNil(t, catSchema, "cat schema should exist")
	require.NotNil(t, catSchema.Properties)
	_, ok := catSchema.Properties.Get("name")
	assert.True(t, ok, "cat schema should have name property")
	_, ok = catSchema.Properties.Get("lives")
	assert.True(t, ok, "cat schema should have lives property")

	// Check dog variant has name and breed properties
	require.NotNil(t, dogSchema, "dog schema should exist")
	require.NotNil(t, dogSchema.Properties)
	_, ok = dogSchema.Properties.Get("name")
	assert.True(t, ok, "dog schema should have name property")
	_, ok = dogSchema.Properties.Get("breed")
	assert.True(t, ok, "dog schema should have breed property")
}

// TestUnionValidator_Schema_VariantConstraintsApplied tests variant schemas
// include validation constraints from pedantigo tags.
func TestUnionValidator_Schema_VariantConstraintsApplied(t *testing.T) {
	validator, err := NewUnion[Pet](UnionOptions{
		DiscriminatorField: "pet_type",
		Variants: []UnionVariant{
			VariantFor[Cat]("cat"),
			VariantFor[Dog]("dog"),
		},
	})
	require.NoError(t, err)

	schema := validator.Schema()

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)

	// Find cat schema by discriminator const value (order may vary)
	var catSchema *jsonschema.Schema
	for _, vs := range schema.OneOf {
		if vs.Properties != nil {
			if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "cat" {
				catSchema = vs
				break
			}
		}
	}
	require.NotNil(t, catSchema, "cat schema should exist")
	require.NotNil(t, catSchema.Properties)

	// Check that name is in required (from validate:"required")
	// and lives has min/max constraints
	livesSchema, ok := catSchema.Properties.Get("lives")
	require.True(t, ok, "lives property should exist")

	// Min constraint should be 1
	assert.NotNil(t, livesSchema.Minimum, "lives should have minimum constraint")
	// Max constraint should be 9
	assert.NotNil(t, livesSchema.Maximum, "lives should have maximum constraint")
}

// TestUnionValidator_Schema_DifferentDiscriminatorFieldName tests Schema()
// uses the configured discriminator field name.
func TestUnionValidator_Schema_DifferentDiscriminatorFieldName(t *testing.T) {
	tests := []struct {
		name             string
		discriminator    string
		expectedInSchema string
	}{
		{
			name:             "pet_type discriminator",
			discriminator:    "pet_type",
			expectedInSchema: "pet_type",
		},
		{
			name:             "type discriminator",
			discriminator:    "type",
			expectedInSchema: "type",
		},
		{
			name:             "kind discriminator",
			discriminator:    "kind",
			expectedInSchema: "kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewUnion[Pet](UnionOptions{
				DiscriminatorField: tt.discriminator,
				Variants: []UnionVariant{
					VariantFor[Cat]("cat"),
				},
			})
			require.NoError(t, err)

			schema := validator.Schema()

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.OneOf)
			require.Len(t, schema.OneOf, 1)
			catSchema := schema.OneOf[0]
			require.NotNil(t, catSchema.Properties)

			_, ok := catSchema.Properties.Get(tt.expectedInSchema)
			assert.True(t, ok, "discriminator field %s should exist in schema", tt.expectedInSchema)
		})
	}
}
