package schemagen

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs
type SimpleStruct struct {
	Name  string `json:"name" pedantigo:"required,min=3,max=50"`
	Age   int    `json:"age" pedantigo:"min=18,max=100"`
	Email string `json:"email" pedantigo:"email"`
}

type NestedStruct struct {
	User    SimpleStruct `json:"user" pedantigo:"required"`
	Created time.Time    `json:"created"`
}

type SliceStruct struct {
	Tags []string `json:"tags" pedantigo:"min=1,max=10"`
	IDs  []int    `json:"ids" pedantigo:"min=1"`
}

type MapStruct struct {
	Metadata map[string]string `json:"metadata"`
}

type ConstraintsStruct struct {
	URL        string  `json:"url" pedantigo:"url"`
	UUID       string  `json:"uuid" pedantigo:"uuid"`
	IPv4       string  `json:"ipv4" pedantigo:"ipv4"`
	IPv6       string  `json:"ipv6" pedantigo:"ipv6"`
	Pattern    string  `json:"pattern" pedantigo:"regexp=^[A-Z]+$"`
	Status     string  `json:"status" pedantigo:"oneof=active inactive pending"`
	Score      float64 `json:"score" pedantigo:"gte=0,lte=100"`
	Count      int     `json:"count" pedantigo:"gt=0,lt=1000"`
	Enabled    bool    `json:"enabled" pedantigo:"default=true"`
	MaxRetries int     `json:"max_retries" pedantigo:"default=3"`
}

// Test containers for nested types
type SliceContainer struct {
	Items []SimpleStruct `json:"items"`
}

type MapContainer struct {
	Data map[string]SimpleStruct `json:"data"`
}

func TestGenerateBaseSchema(t *testing.T) {
	tests := []struct {
		name         string
		generateFunc func() *jsonschema.Schema
		wantType     string
		checkProps   []string
		wantRequired int
		// Optional: extra validation for specific properties
		checkPropType map[string]string // property name -> expected type
	}{
		{
			name:         "simple struct",
			generateFunc: func() *jsonschema.Schema { return GenerateBaseSchema[SimpleStruct]() },
			wantType:     "object",
			checkProps:   []string{"name", "age", "email"},
			wantRequired: 0, // Cleared by GenerateBaseSchema
		},
		{
			name:         "nested struct",
			generateFunc: func() *jsonschema.Schema { return GenerateBaseSchema[NestedStruct]() },
			wantType:     "object",
			checkProps:   []string{"user", "created"},
			wantRequired: 0,
			checkPropType: map[string]string{
				"user": "object",
			},
		},
		{
			name:         "slice struct",
			generateFunc: func() *jsonschema.Schema { return GenerateBaseSchema[SliceStruct]() },
			wantType:     "object",
			checkProps:   []string{"tags", "ids"},
			wantRequired: 0,
			checkPropType: map[string]string{
				"tags": "array",
			},
		},
	}

	// Common validation logic - runs for EVERY test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.generateFunc()

			require.NotNil(t, schema, "schema is nil")
			assert.Equal(t, tt.wantType, schema.Type)
			require.NotNil(t, schema.Properties, "properties is nil")

			// Check all expected properties exist
			for _, prop := range tt.checkProps {
				_, ok := schema.Properties.Get(prop)
				assert.True(t, ok, "%s property not found", prop)
			}

			// Check required count
			assert.Equal(t, tt.wantRequired, len(schema.Required))

			// Optional: check specific property types
			for propName, expectedType := range tt.checkPropType {
				prop, ok := schema.Properties.Get(propName)
				require.True(t, ok, "%s property not found", propName)
				assert.Equal(t, expectedType, prop.Type)
			}
		})
	}
}

func TestGenerateOpenAPIBaseSchema(t *testing.T) {
	schema := GenerateOpenAPIBaseSchema[SimpleStruct]()

	require.NotNil(t, schema, "schema is nil")
	assert.Equal(t, "object", schema.Type)
}

func TestEnhanceSchema(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("pedantigo")
		if pedantigoTag == "" {
			return nil
		}

		// Simple parser for testing
		constraints := make(map[string]string)
		parts := splitConstraints(pedantigoTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	tests := []struct {
		name         string
		structType   reflect.Type
		wantRequired []string
		// Optional: custom check (max 4-5 lines)
		checkFunc func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:         "enhance simple struct with constraints",
			structType:   reflect.TypeOf(SimpleStruct{}),
			wantRequired: []string{"name"},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Check name constraints
				nameProp, _ := schema.Properties.Get("name")
				require.NotNil(t, nameProp.MinLength)
				assert.Equal(t, uint64(3), *nameProp.MinLength)
				require.NotNil(t, nameProp.MaxLength)
				assert.Equal(t, uint64(50), *nameProp.MaxLength)

				// Check age constraints
				ageProp, _ := schema.Properties.Get("age")
				assert.Equal(t, json.Number("18"), ageProp.Minimum)
				assert.Equal(t, json.Number("100"), ageProp.Maximum)

				// Check email format
				emailProp, _ := schema.Properties.Get("email")
				assert.Equal(t, "email", emailProp.Format)
			},
		},
		{
			name:         "enhance nested struct",
			structType:   reflect.TypeOf(NestedStruct{}),
			wantRequired: []string{"user"},
		},
	}

	// Common validation logic
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateBaseSchema[SimpleStruct]()
			if tt.structType == reflect.TypeOf(NestedStruct{}) {
				schema = GenerateBaseSchema[NestedStruct]()
			}

			EnhanceSchema(schema, tt.structType, mockParseTagFunc)

			// Check required fields
			for _, reqField := range tt.wantRequired {
				assert.Contains(t, schema.Required, reqField)
			}

			// Optional custom validation
			if tt.checkFunc != nil {
				tt.checkFunc(t, schema)
			}
		})
	}
}

func TestApplyConstraints(t *testing.T) {
	tests := []struct {
		name        string
		fieldType   reflect.Type
		constraints map[string]string
		checkFunc   func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:      "min max for string",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"min": "5",
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(5), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(100), *schema.MaxLength)
			},
		},
		{
			name:      "min max for int",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"min": "10",
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("10"), schema.Minimum)
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:      "gt gte lt lte",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gt":  "0",
				"lte": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMinimum)
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:      "format constraints",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"email": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "email", schema.Format)
			},
		},
		{
			name:      "url format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"url": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uri", schema.Format)
			},
		},
		{
			name:      "uuid format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"uuid": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uuid", schema.Format)
			},
		},
		{
			name:      "ipv4 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv4", schema.Format)
			},
		},
		{
			name:      "ipv6 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv6": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv6", schema.Format)
			},
		},
		{
			name:      "regexp pattern",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"regexp": "^[A-Z]+$",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[A-Z]+$", schema.Pattern)
			},
		},
		{
			name:      "oneof enum",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"oneof": "red green blue",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Len(t, schema.Enum, 3)
			},
		},
		{
			name:      "default value",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"default": "hello",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hello", schema.Default)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{}
			ApplyConstraints(schema, tt.constraints, tt.fieldType)
			tt.checkFunc(t, schema)
		})
	}
}

func TestParseDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		typ      reflect.Type
		expected any
	}{
		{
			name:     "string default",
			value:    "test",
			typ:      reflect.TypeOf(""),
			expected: "test",
		},
		{
			name:     "int default",
			value:    "42",
			typ:      reflect.TypeOf(0),
			expected: int64(42),
		},
		{
			name:     "uint default",
			value:    "42",
			typ:      reflect.TypeOf(uint(0)),
			expected: uint64(42),
		},
		{
			name:     "float default",
			value:    "3.14",
			typ:      reflect.TypeOf(0.0),
			expected: 3.14,
		},
		{
			name:     "bool default true",
			value:    "true",
			typ:      reflect.TypeOf(false),
			expected: true,
		},
		{
			name:     "bool default false",
			value:    "false",
			typ:      reflect.TypeOf(false),
			expected: false,
		},
		{
			name:     "invalid int returns string",
			value:    "invalid",
			typ:      reflect.TypeOf(0),
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDefaultValue(tt.value, tt.typ)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnhanceNestedTypes(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("pedantigo")
		if pedantigoTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(pedantigoTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	tests := []struct {
		name             string
		structType       reflect.Type
		checkProp        string // Property name to check
		expectItems      bool   // Expect Items field (for slices)
		expectAdditional bool   // Expect AdditionalProperties (for maps)
		skipCheck        bool   // Skip property checks (e.g., time.Time)
	}{
		{
			name:        "enhance slice of structs",
			structType:  reflect.TypeOf(SliceContainer{}),
			checkProp:   "items",
			expectItems: true,
		},
		{
			name:             "enhance map of structs",
			structType:       reflect.TypeOf(MapContainer{}),
			checkProp:        "data",
			expectAdditional: true,
		},
		{
			name:       "skip time.Time enhancement",
			structType: reflect.TypeOf(time.Time{}),
			skipCheck:  true,
		},
	}

	// Common validation logic
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCheck {
				// Should not panic
				schema := &jsonschema.Schema{}
				EnhanceNestedTypes(schema, tt.structType, mockParseTagFunc)
				return
			}

			// Generate and enhance schema
			var schema *jsonschema.Schema
			if tt.structType == reflect.TypeOf(SliceContainer{}) {
				schema = GenerateBaseSchema[SliceContainer]()
			} else if tt.structType == reflect.TypeOf(MapContainer{}) {
				schema = GenerateBaseSchema[MapContainer]()
			}

			EnhanceSchema(schema, tt.structType, mockParseTagFunc)

			// Check property exists
			prop, ok := schema.Properties.Get(tt.checkProp)
			require.True(t, ok, "%s property not found", tt.checkProp)

			// Check expected fields (max 4 lines)
			if tt.expectItems {
				require.NotNil(t, prop.Items, "items schema is nil")
			}
			if tt.expectAdditional {
				require.NotNil(t, prop.AdditionalProperties, "additionalProperties schema is nil")
			}
		})
	}
}

func TestApplyConstraintsToItems(t *testing.T) {
	tests := []struct {
		name        string
		elemType    reflect.Type
		constraints map[string]string
		checkFunc   func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:     "format constraints for items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"email": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "email", schema.Format)
			},
		},
		{
			name:     "min max for string items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"min": "5",
				"max": "50",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(5), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(50), *schema.MaxLength)
			},
		},
		{
			name:     "min max for int items",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"min": "1",
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("1"), schema.Minimum)
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:     "url format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"url": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uri", schema.Format)
			},
		},
		{
			name:     "uuid format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"uuid": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uuid", schema.Format)
			},
		},
		{
			name:     "ipv4 format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv4", schema.Format)
			},
		},
		{
			name:     "ipv6 format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv6": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv6", schema.Format)
			},
		},
		{
			name:     "regexp pattern constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"regexp": "^[a-z]+$",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-z]+$", schema.Pattern)
			},
		},
		{
			name:     "oneof enum constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"oneof": "red green blue",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Len(t, schema.Enum, 3)
				expected := []string{"red", "green", "blue"}
				for i, val := range schema.Enum {
					assert.Equal(t, expected[i], val)
				}
			},
		},
		{
			name:     "gt exclusive minimum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gt": "0",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMinimum)
			},
		},
		{
			name:     "gte inclusive minimum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gte": "1",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("1"), schema.Minimum)
			},
		},
		{
			name:     "lt exclusive maximum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"lt": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("100"), schema.ExclusiveMaximum)
			},
		},
		{
			name:     "lte inclusive maximum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"lte": "99",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("99"), schema.Maximum)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{}
			ApplyConstraintsToItems(schema, tt.constraints, tt.elemType)
			tt.checkFunc(t, schema)
		})
	}
}

func TestFullIntegration(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("pedantigo")
		if pedantigoTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(pedantigoTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateBaseSchema[ConstraintsStruct]()
	typ := reflect.TypeOf(ConstraintsStruct{})
	EnhanceSchema(schema, typ, mockParseTagFunc)

	// Marshal to JSON to verify structure
	jsonData, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err, "failed to marshal schema")

	// Basic sanity checks
	assert.Equal(t, "object", schema.Type)
	require.NotNil(t, schema.Properties, "properties is nil")

	// Verify at least one property was enhanced
	urlProp, ok := schema.Properties.Get("url")
	require.True(t, ok, "url property not found")
	assert.Equal(t, "uri", urlProp.Format)

	t.Logf("Generated schema:\n%s", string(jsonData))
}

// Helper functions for simple tag parsing
func splitConstraints(tag string) []string {
	var parts []string
	current := ""
	for _, ch := range tag {
		if ch == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func splitKeyValue(part string) (key, value string, found bool) {
	for i, ch := range part {
		if ch == '=' {
			return part[:i], part[i+1:], true
		}
	}
	return "", "", false
}
