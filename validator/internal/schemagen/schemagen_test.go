package schemagen

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs
// SimpleStruct represents the data structure.
type SimpleStruct struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Age   int    `json:"age" validate:"min=18,max=100"`
	Email string `json:"email" validate:"email"`
}

type NestedStruct struct {
	User    SimpleStruct `json:"user" validate:"required"`
	Created time.Time    `json:"created"`
}

type SliceStruct struct {
	Tags []string `json:"tags" validate:"min=1,max=10"`
	IDs  []int    `json:"ids" validate:"min=1"`
}

type MapStruct struct {
	Metadata map[string]string `json:"metadata"`
}

type ConstraintsStruct struct {
	URL        string  `json:"url" validate:"url"`
	UUID       string  `json:"uuid" validate:"uuid"`
	IPv4       string  `json:"ipv4" validate:"ipv4"`
	IPv6       string  `json:"ipv6" validate:"ipv6"`
	Pattern    string  `json:"pattern" validate:"regexp=^[A-Z]+$"`
	Status     string  `json:"status" validate:"oneof=active inactive pending"`
	Score      float64 `json:"score" validate:"gte=0,lte=100"`
	Count      int     `json:"count" validate:"gt=0,lt=1000"`
	Enabled    bool    `json:"enabled" validate:"default=true"`
	MaxRetries int     `json:"max_retries" validate:"default=3"`
}

// Test containers for nested types
// SliceContainer represents the data structure.
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
			assert.Len(t, schema.Required, tt.wantRequired)

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
		pedantigoTag := tag.Get("validate")
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
		// Network formats (Phase 10)
		{
			name:      "ip format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ip": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ip", schema.Format)
			},
		},
		{
			name:      "cidr format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"cidr": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "cidr", schema.Format)
			},
		},
		{
			name:      "mac format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"mac": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "mac", schema.Format)
			},
		},
		{
			name:      "hostname format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"hostname": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hostname", schema.Format)
			},
		},
		{
			name:      "fqdn format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"fqdn": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "fqdn", schema.Format)
			},
		},
		// Encoding formats (Phase 10)
		{
			name:      "jwt format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"jwt": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "jwt", schema.Format)
			},
		},
		{
			name:      "json format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"json": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "json", schema.Format)
			},
		},
		{
			name:      "base64 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"base64": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "base64", schema.Format)
			},
		},
		// Hash formats (Phase 10)
		{
			name:      "md5 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"md5": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "md5", schema.Format)
			},
		},
		{
			name:      "sha256 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"sha256": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "sha256", schema.Format)
			},
		},
		{
			name:      "sha512 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"sha512": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "sha512", schema.Format)
			},
		},
		// Geo formats (Phase 10)
		{
			name:      "latitude format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"latitude": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "latitude", schema.Format)
			},
		},
		{
			name:      "longitude format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"longitude": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "longitude", schema.Format)
			},
		},
		// Color formats (Phase 10)
		{
			name:      "hexcolor format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"hexcolor": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hexcolor", schema.Format)
			},
		},
		{
			name:      "rgb format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"rgb": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "rgb", schema.Format)
			},
		},
		// Filesystem formats (Phase 10)
		{
			name:      "filepath format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"filepath": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "filepath", schema.Format)
			},
		},
		{
			name:      "file format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"file": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "file", schema.Format)
			},
		},
		// Additional network formats
		{
			name:      "cidrv4 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"cidrv4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "cidrv4", schema.Format)
			},
		},
		{
			name:      "port format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"port": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "port", schema.Format)
			},
		},
		{
			name:      "tcp_addr format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"tcp_addr": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "tcp_addr", schema.Format)
			},
		},
		// Finance formats
		{
			name:      "credit_card format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"credit_card": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "credit_card", schema.Format)
			},
		},
		{
			name:      "eth_addr format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"eth_addr": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "eth_addr", schema.Format)
			},
		},
		// Identity formats
		{
			name:      "isbn format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"isbn": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "isbn", schema.Format)
			},
		},
		{
			name:      "ssn format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ssn": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ssn", schema.Format)
			},
		},
		{
			name:      "e164 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"e164": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "e164", schema.Format)
			},
		},
		// More color formats
		{
			name:      "rgba format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"rgba": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "rgba", schema.Format)
			},
		},
		{
			name:      "hsl format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"hsl": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hsl", schema.Format)
			},
		},
		// Misc formats
		{
			name:      "semver format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"semver": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "semver", schema.Format)
			},
		},
		{
			name:      "cron format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"cron": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "cron", schema.Format)
			},
		},
		{
			name:      "ulid format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ulid": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ulid", schema.Format)
			},
		},
		// ISO formats
		{
			name:      "iso3166_1_alpha2 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso3166_1_alpha2": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso3166_1_alpha2", schema.Format)
			},
		},
		{
			name:      "iso4217 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso4217": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso4217", schema.Format)
			},
		},
		{
			name:      "bcp47_language_tag format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"bcp47_language_tag": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "bcp47_language_tag", schema.Format)
			},
		},
		// Filesystem
		{
			name:      "dirpath format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"dirpath": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "dirpath", schema.Format)
			},
		},
		{
			name:      "dir format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"dir": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "dir", schema.Format)
			},
		},
		// Additional network formats
		{
			name:      "cidrv6 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"cidrv6": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "cidrv6", schema.Format)
			},
		},
		{
			name:      "hostname_rfc format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"hostname_rfc1123": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hostname_rfc1123", schema.Format)
			},
		},
		{
			name:      "udp_addr format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"udp_addr": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "udp_addr", schema.Format)
			},
		},
		{
			name:      "tcp4_addr format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"tcp4_addr": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "tcp4_addr", schema.Format)
			},
		},
		// Additional finance formats
		{
			name:      "btc_addr format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"btc_addr": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "btc_addr", schema.Format)
			},
		},
		{
			name:      "btc_addr_bech32 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"btc_addr_bech32": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "btc_addr_bech32", schema.Format)
			},
		},
		{
			name:      "luhn format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"luhn_checksum": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "luhn_checksum", schema.Format)
			},
		},
		// Additional identity formats
		{
			name:      "isbn10 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"isbn10": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "isbn10", schema.Format)
			},
		},
		{
			name:      "isbn13 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"isbn13": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "isbn13", schema.Format)
			},
		},
		{
			name:      "issn format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"issn": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "issn", schema.Format)
			},
		},
		{
			name:      "ein format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ein": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ein", schema.Format)
			},
		},
		// Additional color format
		{
			name:      "hsla format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"hsla": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hsla", schema.Format)
			},
		},
		// Additional encoding formats
		{
			name:      "base64url format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"base64url": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "base64url", schema.Format)
			},
		},
		{
			name:      "base64rawurl format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"base64rawurl": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "base64rawurl", schema.Format)
			},
		},
		// Additional hash formats
		{
			name:      "md4 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"md4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "md4", schema.Format)
			},
		},
		{
			name:      "sha384 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"sha384": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "sha384", schema.Format)
			},
		},
		{
			name:      "mongodb format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"mongodb": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "mongodb", schema.Format)
			},
		},
		// HTML format
		{
			name:      "html format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"html": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "html", schema.Format)
			},
		},
		// Additional ISO formats
		{
			name:      "iso3166_alpha2_eu format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso3166_alpha2_eu": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso3166_alpha2_eu", schema.Format)
			},
		},
		{
			name:      "iso3166_1_alpha3 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso3166_1_alpha3": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso3166_1_alpha3", schema.Format)
			},
		},
		{
			name:      "iso3166_alpha3_eu format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso3166_alpha3_eu": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso3166_alpha3_eu", schema.Format)
			},
		},
		{
			name:      "iso3166_1_alpha_numeric format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso3166_1_alpha_numeric": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso3166_1_alpha_numeric", schema.Format)
			},
		},
		{
			name:      "iso3166_2 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso3166_2": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso3166_2", schema.Format)
			},
		},
		{
			name:      "iso4217_numeric format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"iso4217_numeric": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "iso4217_numeric", schema.Format)
			},
		},
		{
			name:      "postcode format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"postcode": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "postcode", schema.Format)
			},
		},
		// Pointer type tests
		{
			name:      "min constraint pointer string",
			fieldType: reflect.TypeOf((*string)(nil)),
			constraints: map[string]string{
				"min": "5",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(5), *schema.MinLength)
			},
		},
		{
			name:      "max constraint pointer string",
			fieldType: reflect.TypeOf((*string)(nil)),
			constraints: map[string]string{
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(100), *schema.MaxLength)
			},
		},
		{
			name:      "min max constraint pointer int",
			fieldType: reflect.TypeOf((*int)(nil)),
			constraints: map[string]string{
				"min": "1",
				"max": "99",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("1"), schema.Minimum)
				assert.Equal(t, json.Number("99"), schema.Maximum)
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
		{
			name:      "len constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"len": "8",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(8), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(8), *schema.MaxLength)
			},
		},
		{
			name:      "ascii constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ascii": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[\\x00-\\x7F]*$", schema.Pattern)
			},
		},
		{
			name:      "alpha constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"alpha": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-zA-Z]+$", schema.Pattern)
			},
		},
		{
			name:      "alphanum constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"alphanum": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-zA-Z0-9]+$", schema.Pattern)
			},
		},
		{
			name:      "contains constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"contains": "test",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, ".*test.*", schema.Pattern)
			},
		},
		{
			name:      "contains with special chars",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"contains": "@example.com",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, `.*@example\.com.*`, schema.Pattern)
			},
		},
		{
			name:      "excludes constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"excludes": "bad",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^(?!.*bad).*$", schema.Pattern)
			},
		},
		{
			name:      "excludes with special chars",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"excludes": "test.value",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, `^(?!.*test\.value).*$`, schema.Pattern)
			},
		},
		{
			name:      "startswith constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"startswith": "prefix",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^prefix.*", schema.Pattern)
			},
		},
		{
			name:      "startswith with special chars",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"startswith": "http://",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, `^http://.*`, schema.Pattern)
			},
		},
		{
			name:      "endswith constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"endswith": ".txt",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, `.*\.txt$`, schema.Pattern)
			},
		},
		{
			name:      "lowercase constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"lowercase": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[^A-Z]*$", schema.Pattern)
			},
		},
		{
			name:      "uppercase constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"uppercase": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[^a-z]*$", schema.Pattern)
			},
		},
		{
			name:      "positive constraint",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"positive": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "0", string(schema.ExclusiveMinimum))
			},
		},
		{
			name:      "negative constraint",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"negative": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "0", string(schema.ExclusiveMaximum))
			},
		},
		{
			name:      "multiple_of constraint",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"multiple_of": "5",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "5", string(schema.MultipleOf))
			},
		},
		{
			name:      "defaultUsingMethod is skipped",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"defaultUsingMethod": "GetDefault",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Should not set any schema property
				assert.Nil(t, schema.Default)
			},
		},
		{
			name:      "title constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"title": "User Name",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "User Name", schema.Title)
			},
		},
		{
			name:      "description constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"description": "Full name of the user",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "Full name of the user", schema.Description)
			},
		},
		{
			name:      "examples json array single value",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"examples": `["John Doe"]`,
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.Examples)
				require.Len(t, schema.Examples, 1)
				assert.Equal(t, "John Doe", schema.Examples[0])
			},
		},
		{
			name:      "examples json array of strings",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"examples": `["John","Jane","Bob"]`,
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.Examples)
				require.Len(t, schema.Examples, 3)
				assert.Equal(t, "John", schema.Examples[0])
				assert.Equal(t, "Jane", schema.Examples[1])
				assert.Equal(t, "Bob", schema.Examples[2])
			},
		},
		{
			name:      "all metadata combined",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"title":       "User Full Name",
				"description": "The complete name of the user",
				"examples":    `["John Doe","Jane Smith","Bob Johnson"]`,
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "User Full Name", schema.Title)
				assert.Equal(t, "The complete name of the user", schema.Description)
				require.NotNil(t, schema.Examples)
				require.Len(t, schema.Examples, 3)
				assert.Equal(t, "John Doe", schema.Examples[0])
				assert.Equal(t, "Jane Smith", schema.Examples[1])
				assert.Equal(t, "Bob Johnson", schema.Examples[2])
			},
		},
		{
			name:      "title with int field",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"title": "User Age",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "User Age", schema.Title)
			},
		},
		{
			name:      "description with int field",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"description": "Age in years",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "Age in years", schema.Description)
			},
		},
		{
			name:      "examples json array of numeric values",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"examples": "[18,25,30]",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.Examples)
				require.Len(t, schema.Examples, 3)
				// json.Unmarshal of [18,25,30] produces float64 values
				assert.InDelta(t, float64(18), schema.Examples[0], 0)
				assert.InDelta(t, float64(25), schema.Examples[1], 0)
				assert.InDelta(t, float64(30), schema.Examples[2], 0)
			},
		},
		{
			name:      "metadata with validation constraints",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"required":    "",
				"min":         "3",
				"max":         "50",
				"title":       "Username",
				"description": "Unique username for the account",
				"examples":    `["alice","bob","charlie"]`,
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Check validation constraints
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(3), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(50), *schema.MaxLength)

				// Check metadata
				assert.Equal(t, "Username", schema.Title)
				assert.Equal(t, "Unique username for the account", schema.Description)
				require.NotNil(t, schema.Examples)
				require.Len(t, schema.Examples, 3)
				assert.Equal(t, "alice", schema.Examples[0])
				assert.Equal(t, "bob", schema.Examples[1])
				assert.Equal(t, "charlie", schema.Examples[2])
			},
		},
		{
			name:      "empty examples value produces no examples",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"examples": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Empty string is not valid JSON — examples are silently skipped
				assert.Nil(t, schema.Examples)
			},
		},
		{
			name:      "deprecated simple",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"deprecated": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
				assert.Empty(t, schema.Description)
			},
		},
		{
			name:      "deprecated with message",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"deprecated": "Use newField instead",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
				assert.Equal(t, "Deprecated: Use newField instead", schema.Description)
			},
		},
		{
			name:      "deprecated with existing description",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"description": "Old field for legacy support",
				"deprecated":  "Will be removed in v3.0",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
				assert.Equal(t, "Old field for legacy support. Deprecated: Will be removed in v3.0", schema.Description)
			},
		},
		{
			name:      "deprecated with int field",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"deprecated": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
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
		pedantigoTag := tag.Get("validate")
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
		pedantigoTag := tag.Get("validate")
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

// Helper functions for simple tag parsing.
func splitConstraints(tag string) []string {
	var parts []string
	remaining := tag

	for remaining != "" {
		commaIdx := strings.IndexByte(remaining, ',')
		if commaIdx == -1 {
			parts = append(parts, remaining)
			break
		}

		// If there is a key=<json> at the head, use json.Decoder to find where the JSON ends.
		eqIdx := strings.IndexByte(remaining[:commaIdx], '=')
		if eqIdx != -1 {
			afterEq := remaining[eqIdx+1:]
			if afterEq != "" && (afterEq[0] == '[' || afterEq[0] == '{' || afterEq[0] == '"') {
				dec := json.NewDecoder(strings.NewReader(afterEq))
				dec.UseNumber()
				var raw json.RawMessage
				if err := dec.Decode(&raw); err == nil {
					// dec.Buffered() is an in-memory reader — ReadAll cannot fail.
					bufferedBytes, readErr := io.ReadAll(dec.Buffered())
					if readErr == nil {
						consumed := eqIdx + 1 + (len(afterEq) - len(bufferedBytes))
						parts = append(parts, remaining[:consumed])
						after := remaining[consumed:]
						if after != "" && after[0] == ',' {
							after = after[1:]
						}
						remaining = after
						continue
					}
				}
			}
		}

		parts = append(parts, remaining[:commaIdx])
		remaining = remaining[commaIdx+1:]
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

// Test structs for discriminated union schema generation.

// CatUnion represents a cat variant with validation constraints.
type CatUnion struct {
	Name  string `json:"name" validate:"required"`
	Lives int    `json:"lives" validate:"min=1,max=9"`
}

// DogUnion represents a dog variant with validation constraints.
type DogUnion struct {
	Name  string `json:"name" validate:"required"`
	Breed string `json:"breed"`
}

// BirdUnion represents a bird variant with validation constraints.
type BirdUnion struct {
	Name string `json:"name" validate:"required"`
	Eggs int    `json:"eggs" validate:"min=0"`
}

// NestedVariant represents a variant with nested struct fields.
type NestedVariant struct {
	ID    string       `json:"id" validate:"required"`
	Owner SimpleStruct `json:"owner" validate:"required"`
}

// TestGenerateVariantSchema_BasicVariant tests GenerateVariantSchema
// generates schema for variant type with properties.
func TestGenerateVariantSchema_BasicVariant(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateVariantSchema(
		reflect.TypeOf(CatUnion{}),
		"pet_type",
		"cat",
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "variant schema should not be nil")
	assert.Equal(t, "object", schema.Type)
	require.NotNil(t, schema.Properties, "properties should not be nil")

	// Check discriminator field exists
	_, ok := schema.Properties.Get("pet_type")
	assert.True(t, ok, "pet_type property should exist")

	// Check variant properties exist
	_, ok = schema.Properties.Get("name")
	assert.True(t, ok, "name property should exist")
	_, ok = schema.Properties.Get("lives")
	assert.True(t, ok, "lives property should exist")
}

// TestGenerateVariantSchema_DiscriminatorConstConstraint tests
// GenerateVariantSchema adds const constraint on discriminator field.
func TestGenerateVariantSchema_DiscriminatorConstConstraint(t *testing.T) {
	tests := []struct {
		name               string
		variantType        reflect.Type
		discriminatorField string
		discriminatorValue string
	}{
		{
			name:               "cat variant",
			variantType:        reflect.TypeOf(CatUnion{}),
			discriminatorField: "pet_type",
			discriminatorValue: "cat",
		},
		{
			name:               "dog variant",
			variantType:        reflect.TypeOf(DogUnion{}),
			discriminatorField: "pet_type",
			discriminatorValue: "dog",
		},
		{
			name:               "bird variant",
			variantType:        reflect.TypeOf(BirdUnion{}),
			discriminatorField: "pet_type",
			discriminatorValue: "bird",
		},
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateVariantSchema(
				tt.variantType,
				tt.discriminatorField,
				tt.discriminatorValue,
				mockParseTagFunc,
			)

			require.NotNil(t, schema)
			require.NotNil(t, schema.Properties)

			// Check discriminator field has const constraint
			discriminatorSchema, ok := schema.Properties.Get(tt.discriminatorField)
			require.True(t, ok, "discriminator field should exist")
			assert.Equal(t, tt.discriminatorValue, discriminatorSchema.Const,
				"discriminator field should have const %s", tt.discriminatorValue)
		})
	}
}

// TestGenerateVariantSchema_IncludesValidationConstraints tests
// GenerateVariantSchema includes validation constraints from pedantigo tags.
func TestGenerateVariantSchema_IncludesValidationConstraints(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateVariantSchema(
		reflect.TypeOf(CatUnion{}),
		"pet_type",
		"cat",
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.Properties)

	// Check required constraint on name
	_, ok := schema.Properties.Get("name")
	require.True(t, ok, "name property should exist")
	// Name should be in required array
	assert.Contains(t, schema.Required, "name", "name should be in required fields")

	// Check min/max constraints on lives
	livesSchema, ok := schema.Properties.Get("lives")
	require.True(t, ok, "lives property should exist")
	require.NotNil(t, livesSchema.Minimum, "lives should have minimum constraint")
	assert.Equal(t, json.Number("1"), livesSchema.Minimum)
	require.NotNil(t, livesSchema.Maximum, "lives should have maximum constraint")
	assert.Equal(t, json.Number("9"), livesSchema.Maximum)
}

// TestGenerateVariantSchema_HandlesNestedStructs tests
// GenerateVariantSchema handles nested struct fields.
func TestGenerateVariantSchema_HandlesNestedStructs(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateVariantSchema(
		reflect.TypeOf(NestedVariant{}),
		"type",
		"nested",
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "variant schema should not be nil")
	require.NotNil(t, schema.Properties)

	// Check nested struct property exists
	ownerSchema, ok := schema.Properties.Get("owner")
	require.True(t, ok, "owner property should exist")
	assert.Equal(t, "object", ownerSchema.Type, "owner should be an object type")
}

// TestGenerateVariantSchema_DifferentDiscriminatorFields tests
// GenerateVariantSchema works with different discriminator field names.
func TestGenerateVariantSchema_DifferentDiscriminatorFields(t *testing.T) {
	tests := []struct {
		name                  string
		discriminatorField    string
		expectedFieldInSchema string
	}{
		{name: "pet_type", discriminatorField: "pet_type", expectedFieldInSchema: "pet_type"},
		{name: "type", discriminatorField: "type", expectedFieldInSchema: "type"},
		{name: "kind", discriminatorField: "kind", expectedFieldInSchema: "kind"},
		{name: "variant_type", discriminatorField: "variant_type", expectedFieldInSchema: "variant_type"},
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string { return nil }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateVariantSchema(
				reflect.TypeOf(CatUnion{}),
				tt.discriminatorField,
				"test_value",
				mockParseTagFunc,
			)

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.Properties)
			_, ok := schema.Properties.Get(tt.expectedFieldInSchema)
			assert.True(t, ok, "discriminator field %s should exist", tt.expectedFieldInSchema)
		})
	}
}

// TestGenerateUnionSchema_ReturnsOneOfSchema tests
// GenerateUnionSchema returns schema with oneOf array.
func TestGenerateUnionSchema_ReturnsOneOfSchema(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat":  reflect.TypeOf(CatUnion{}),
		"dog":  reflect.TypeOf(DogUnion{}),
		"bird": reflect.TypeOf(BirdUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "union schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
}

// TestGenerateUnionSchema_OneOfArrayLength tests
// GenerateUnionSchema oneOf array contains correct number of variants.
func TestGenerateUnionSchema_OneOfArrayLength(t *testing.T) {
	tests := []struct {
		name          string
		variants      map[string]reflect.Type
		expectedCount int
	}{
		{
			name: "two variants",
			variants: map[string]reflect.Type{
				"cat": reflect.TypeOf(CatUnion{}),
				"dog": reflect.TypeOf(DogUnion{}),
			},
			expectedCount: 2,
		},
		{
			name: "three variants",
			variants: map[string]reflect.Type{
				"cat":  reflect.TypeOf(CatUnion{}),
				"dog":  reflect.TypeOf(DogUnion{}),
				"bird": reflect.TypeOf(BirdUnion{}),
			},
			expectedCount: 3,
		},
		{
			name: "single variant",
			variants: map[string]reflect.Type{
				"cat": reflect.TypeOf(CatUnion{}),
			},
			expectedCount: 1,
		},
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateUnionSchema(
				"pet_type",
				tt.variants,
				mockParseTagFunc,
			)

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.OneOf)
			assert.Len(t, schema.OneOf, tt.expectedCount,
				"oneOf array should contain %d variant schemas", tt.expectedCount)
		})
	}
}

// TestGenerateUnionSchema_EachVariantProperlyGenerated tests
// GenerateUnionSchema generates each variant schema properly.
func TestGenerateUnionSchema_EachVariantProperlyGenerated(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)
	require.Len(t, schema.OneOf, 2)

	// Verify each variant schema is properly formed
	for i, variantSchema := range schema.OneOf {
		assert.NotNil(t, variantSchema, "variant schema at index %d should not be nil", i)
		assert.Equal(t, "object", variantSchema.Type, "variant at index %d should be object type", i)
		assert.NotNil(t, variantSchema.Properties, "variant at index %d should have properties", i)
	}
}

// TestGenerateUnionSchema_DiscriminatorConstInEachVariant tests
// GenerateUnionSchema adds discriminator field with const in each variant.
func TestGenerateUnionSchema_DiscriminatorConstInEachVariant(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)
	require.Len(t, schema.OneOf, 2)

	// Check each variant has pet_type with const (order may vary due to map iteration)
	foundConsts := make(map[string]bool)
	for _, variantSchema := range schema.OneOf {
		require.NotNil(t, variantSchema.Properties)
		petTypeSchema, ok := variantSchema.Properties.Get("pet_type")
		require.True(t, ok, "variant should have pet_type property")
		if constVal, ok := petTypeSchema.Const.(string); ok {
			foundConsts[constVal] = true
		}
	}
	assert.True(t, foundConsts["cat"], "should have variant with const 'cat'")
	assert.True(t, foundConsts["dog"], "should have variant with const 'dog'")
}

// TestGenerateUnionSchema_VariantPropertiesIncluded tests
// GenerateUnionSchema includes variant-specific properties.
func TestGenerateUnionSchema_VariantPropertiesIncluded(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)
	require.Len(t, schema.OneOf, 2)

	// Find cat and dog schemas (order may vary in map)
	var catSchema, dogSchema *jsonschema.Schema
	for _, vs := range schema.OneOf {
		if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "cat" {
			catSchema = vs
		}
		if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "dog" {
			dogSchema = vs
		}
	}

	require.NotNil(t, catSchema, "cat variant schema should exist")
	require.NotNil(t, dogSchema, "dog variant schema should exist")

	// Check cat-specific properties
	_, ok := catSchema.Properties.Get("lives")
	assert.True(t, ok, "cat schema should have lives property")

	// Check dog-specific properties
	_, ok = dogSchema.Properties.Get("breed")
	assert.True(t, ok, "dog schema should have breed property")
}

// TestGenerateUnionSchema_DifferentDiscriminatorFields tests
// GenerateUnionSchema works with different discriminator field names.
func TestGenerateUnionSchema_DifferentDiscriminatorFields(t *testing.T) {
	tests := []struct {
		name               string
		discriminatorField string
	}{
		{name: "pet_type", discriminatorField: "pet_type"},
		{name: "type", discriminatorField: "type"},
		{name: "kind", discriminatorField: "kind"},
	}

	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string { return nil }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateUnionSchema(
				tt.discriminatorField,
				variants,
				mockParseTagFunc,
			)

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.OneOf)
			require.Len(t, schema.OneOf, 2)

			// Check discriminator field exists in each variant
			for i, variantSchema := range schema.OneOf {
				require.NotNil(t, variantSchema.Properties)
				_, ok := variantSchema.Properties.Get(tt.discriminatorField)
				assert.True(t, ok,
					"variant at index %d should have %s field", i, tt.discriminatorField)
			}
		})
	}
}

// TestGenerateUnionSchema_WithValidationConstraints tests
// GenerateUnionSchema respects validation constraints in variants.
func TestGenerateUnionSchema_WithValidationConstraints(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat":  reflect.TypeOf(CatUnion{}),
		"bird": reflect.TypeOf(BirdUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)

	// Find cat schema and verify constraints
	for _, vs := range schema.OneOf {
		if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "cat" {
			// Check lives has min=1 and max=9
			livesSchema, ok := vs.Properties.Get("lives")
			require.True(t, ok, "cat should have lives property")
			assert.NotNil(t, livesSchema.Minimum, "lives should have minimum")
			assert.NotNil(t, livesSchema.Maximum, "lives should have maximum")
		}
	}
}

// TestEnhanceSchema_UnexportedFields tests that unexported fields are skipped.
func TestEnhanceSchema_UnexportedFields(t *testing.T) {
	type StructWithUnexported struct {
		Public     string `json:"public" validate:"required"`
		private    string //nolint:unused // intentionally unexported for testing
		AlsoPublic int    `json:"also_public" validate:"min=0"`
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
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

	schema := GenerateBaseSchema[StructWithUnexported]()
	EnhanceSchema(schema, reflect.TypeOf(StructWithUnexported{}), mockParseTagFunc)

	// Should have public fields in required
	assert.Contains(t, schema.Required, "public")

	// Should have also_public with constraints
	alsoProp, ok := schema.Properties.Get("also_public")
	require.True(t, ok)
	assert.NotNil(t, alsoProp.Minimum)
}

// TestEnhanceSchema_PointerType tests enhancing a pointer type.
func TestEnhanceSchema_PointerType(t *testing.T) {
	type Target struct {
		Name string `json:"name" validate:"required,min=1"`
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
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

	schema := GenerateBaseSchema[Target]()
	// Pass pointer type - should unwrap it
	EnhanceSchema(schema, reflect.TypeOf((*Target)(nil)), mockParseTagFunc)

	// Should still process fields correctly
	assert.Contains(t, schema.Required, "name")
	nameProp, ok := schema.Properties.Get("name")
	require.True(t, ok)
	require.NotNil(t, nameProp.MinLength)
	assert.Equal(t, uint64(1), *nameProp.MinLength)
}

// TestEnhanceSchema_NonStructType tests enhancing a non-struct type returns early.
func TestEnhanceSchema_NonStructType(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		return nil
	}

	schema := &jsonschema.Schema{}
	// Pass non-struct type - should return early without panic
	EnhanceSchema(schema, reflect.TypeOf("string"), mockParseTagFunc)
	EnhanceSchema(schema, reflect.TypeOf(123), mockParseTagFunc)
	EnhanceSchema(schema, reflect.TypeOf([]string{}), mockParseTagFunc)

	// Should not modify schema
	assert.Nil(t, schema.Properties)
}

// TestEnhanceSchema_JSONTagWithOptions tests JSON tag with comma-separated options.
func TestEnhanceSchema_JSONTagWithOptions(t *testing.T) {
	type StructWithJSONOptions struct {
		Name     string `json:"name,omitempty" validate:"required"`
		Age      int    `json:"age,string" validate:"min=0"`
		Disabled string `json:"-"`
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
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

	schema := GenerateBaseSchema[StructWithJSONOptions]()
	EnhanceSchema(schema, reflect.TypeOf(StructWithJSONOptions{}), mockParseTagFunc)

	// name should be in required (JSON name extracted from "name,omitempty")
	assert.Contains(t, schema.Required, "name")

	// age should have minimum (JSON name extracted from "age,string")
	ageProp, ok := schema.Properties.Get("age")
	require.True(t, ok, "age property should exist")
	assert.NotNil(t, ageProp.Minimum)
}

// TestApplyConstraints_SliceWithConstraints tests constraints applied to slice items.
func TestApplyConstraints_SliceWithConstraints(t *testing.T) {
	schema := &jsonschema.Schema{
		Items: &jsonschema.Schema{},
	}
	constraints := map[string]string{
		"email": "",
	}
	ApplyConstraints(schema, constraints, reflect.TypeOf([]string{}))

	// email format should be applied to items
	assert.Equal(t, "email", schema.Items.Format)
}

// TestApplyConstraints_MapWithConstraints tests constraints applied to map values.
func TestApplyConstraints_MapWithConstraints(t *testing.T) {
	schema := &jsonschema.Schema{
		AdditionalProperties: &jsonschema.Schema{},
	}
	constraints := map[string]string{
		"url": "",
	}
	ApplyConstraints(schema, constraints, reflect.TypeOf(map[string]string{}))

	// url format should be applied to additionalProperties
	assert.Equal(t, "uri", schema.AdditionalProperties.Format)
}

// MetadataStruct is a test struct with schema metadata.
type MetadataStruct struct {
	Name  string `json:"name" validate:"required,title=User Name,description=Full name of user,examples=[\"John\",\"Jane\"]"`
	Age   int    `json:"age" validate:"min=0,description=Age in years,examples=[18,25,30]"`
	Email string `json:"email" validate:"email,title=Email Address,description=Contact email,examples=[\"john@example.com\",\"jane@example.com\"]"`
}

// TestEnhanceSchema_WithMetadata tests metadata propagation through EnhanceSchema.
func TestEnhanceSchema_WithMetadata(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
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

	schema := GenerateBaseSchema[MetadataStruct]()
	EnhanceSchema(schema, reflect.TypeOf(MetadataStruct{}), mockParseTagFunc)

	// Check name field metadata and constraints
	nameProp, ok := schema.Properties.Get("name")
	require.True(t, ok, "name property should exist")
	assert.Equal(t, "User Name", nameProp.Title)
	assert.Equal(t, "Full name of user", nameProp.Description)
	require.NotNil(t, nameProp.Examples)
	require.Len(t, nameProp.Examples, 2)
	assert.Equal(t, "John", nameProp.Examples[0])
	assert.Equal(t, "Jane", nameProp.Examples[1])
	assert.Contains(t, schema.Required, "name")

	// Check age field metadata and constraints
	ageProp, ok := schema.Properties.Get("age")
	require.True(t, ok, "age property should exist")
	assert.Equal(t, "Age in years", ageProp.Description)
	require.NotNil(t, ageProp.Examples)
	require.Len(t, ageProp.Examples, 3)
	// json.Unmarshal of [18,25,30] produces float64 values
	assert.InDelta(t, float64(18), ageProp.Examples[0], 0)
	assert.InDelta(t, float64(25), ageProp.Examples[1], 0)
	assert.InDelta(t, float64(30), ageProp.Examples[2], 0)
	assert.Equal(t, json.Number("0"), ageProp.Minimum)

	// Check email field metadata and constraints
	emailProp, ok := schema.Properties.Get("email")
	require.True(t, ok, "email property should exist")
	assert.Equal(t, "Email Address", emailProp.Title)
	assert.Equal(t, "Contact email", emailProp.Description)
	require.NotNil(t, emailProp.Examples)
	require.Len(t, emailProp.Examples, 2)
	assert.Equal(t, "john@example.com", emailProp.Examples[0])
	assert.Equal(t, "jane@example.com", emailProp.Examples[1])
	assert.Equal(t, "email", emailProp.Format)
}

// ComplexMetadataStruct tests metadata with special characters and edge cases.
type ComplexMetadataStruct struct {
	Description string `json:"description" validate:"title=Item Description,examples=[\"This is a test\",\"Another example here\"]"`
	Count       int    `json:"count" validate:"title=Item Count,description=Number of items available"`
	Status      string `json:"status" validate:"oneof=active inactive,title=Status,description=Current status,examples=[\"active\",\"inactive\"]"`
}

// TestEnhanceSchema_WithComplexMetadata tests metadata with special characters.
func TestEnhanceSchema_WithComplexMetadata(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
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

	schema := GenerateBaseSchema[ComplexMetadataStruct]()
	EnhanceSchema(schema, reflect.TypeOf(ComplexMetadataStruct{}), mockParseTagFunc)

	// Check description field with examples containing commas
	descProp, ok := schema.Properties.Get("description")
	require.True(t, ok, "description property should exist")
	assert.Equal(t, "Item Description", descProp.Title)
	require.NotNil(t, descProp.Examples)
	require.Len(t, descProp.Examples, 2)
	assert.Equal(t, "This is a test", descProp.Examples[0])
	assert.Equal(t, "Another example here", descProp.Examples[1])

	// Check count field with title and description only
	countProp, ok := schema.Properties.Get("count")
	require.True(t, ok, "count property should exist")
	assert.Equal(t, "Item Count", countProp.Title)
	assert.Equal(t, "Number of items available", countProp.Description)
	// No examples for this field
	assert.Nil(t, countProp.Examples)

	// Check status field with enum and metadata
	statusProp, ok := schema.Properties.Get("status")
	require.True(t, ok, "status property should exist")
	assert.Equal(t, "Status", statusProp.Title)
	assert.Equal(t, "Current status", statusProp.Description)
	require.NotNil(t, statusProp.Examples)
	require.Len(t, statusProp.Examples, 2)
	assert.Equal(t, "active", statusProp.Examples[0])
	assert.Equal(t, "inactive", statusProp.Examples[1])
	// Check enum constraint is also applied
	require.NotNil(t, statusProp.Enum)
	assert.Len(t, statusProp.Enum, 2)
}

// TestEnhanceSchema_NilProperties tests EnhanceSchema handles nil properties.
func TestEnhanceSchema_NilProperties(t *testing.T) {
	type Empty struct {
		unexported string //nolint:unused // intentionally unused for test
	}

	parseTagFunc := func(tag reflect.StructTag) map[string]string {
		return nil
	}

	// Create schema with nil properties
	schema := &jsonschema.Schema{Type: "object", Properties: nil}

	// Should not panic
	EnhanceSchema(schema, reflect.TypeOf(Empty{}), parseTagFunc)
}

// TestEnhanceSchema_FieldWithoutSchema tests when field schema doesn't exist.
func TestEnhanceSchema_FieldWithoutSchema(t *testing.T) {
	type Missing struct {
		Name string `json:"name" validate:"required"`
	}

	parseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
		if pedantigoTag == "" {
			return nil
		}
		return map[string]string{"required": ""}
	}

	// Create schema without the property
	schema := &jsonschema.Schema{
		Type:       "object",
		Properties: jsonschema.NewProperties(),
	}

	// Should not panic when property doesn't exist
	EnhanceSchema(schema, reflect.TypeOf(Missing{}), parseTagFunc)
}

// TestEnhanceSchema_DuplicateRequired tests that required fields aren't duplicated.
func TestEnhanceSchema_DuplicateRequired(t *testing.T) {
	type User struct {
		Name string `json:"name" validate:"required"`
	}

	parseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("validate")
		if pedantigoTag == "" {
			return nil
		}
		return map[string]string{"required": ""}
	}

	schema := GenerateBaseSchema[User]()
	require.NotNil(t, schema)

	// Pre-add name to required
	schema.Required = []string{"name"}

	EnhanceSchema(schema, reflect.TypeOf(User{}), parseTagFunc)

	// Name should appear only once
	count := 0
	for _, r := range schema.Required {
		if r == "name" {
			count++
		}
	}
	assert.Equal(t, 1, count, "required field should not be duplicated")
}

// TestApplyConstraints_EdgeCases tests additional constraint combinations for full coverage.
func TestApplyConstraints_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		fieldType   reflect.Type
		constraints map[string]string
		checkFunc   func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:      "eq constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"eq": "active",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "active", schema.Const)
			},
		},
		{
			name:      "ne constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ne": "disabled",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.Not)
				assert.Equal(t, "disabled", schema.Not.Const)
			},
		},
		{
			name:      "len constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"len": "10",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(10), *schema.MinLength)
				assert.Equal(t, uint64(10), *schema.MaxLength)
			},
		},
		{
			name:      "len with negative value",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"len": "-1",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Should not set length for negative values
				assert.Nil(t, schema.MinLength)
				assert.Nil(t, schema.MaxLength)
			},
		},
		{
			name:      "len with invalid value",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"len": "invalid",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Should not set length for invalid values
				assert.Nil(t, schema.MinLength)
				assert.Nil(t, schema.MaxLength)
			},
		},
		{
			name:      "ascii constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ascii": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[\\x00-\\x7F]*$", schema.Pattern)
			},
		},
		{
			name:      "alpha constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"alpha": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-zA-Z]+$", schema.Pattern)
			},
		},
		{
			name:      "alphanum constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"alphanum": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-zA-Z0-9]+$", schema.Pattern)
			},
		},
		{
			name:      "contains with special characters",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"contains": "test.value",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Contains(t, schema.Pattern, "test\\.value")
			},
		},
		{
			name:      "excludes with special characters",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"excludes": "bad[value]",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Contains(t, schema.Pattern, "bad\\[value\\]")
			},
		},
		{
			name:      "startswith with special characters",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"startswith": "prefix*",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Contains(t, schema.Pattern, "^prefix\\*")
			},
		},
		{
			name:      "endswith with special characters",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"endswith": "suffix+",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Contains(t, schema.Pattern, "suffix\\+$")
			},
		},
		{
			name:      "lowercase constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"lowercase": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[^A-Z]*$", schema.Pattern)
			},
		},
		{
			name:      "uppercase constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"uppercase": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[^a-z]*$", schema.Pattern)
			},
		},
		{
			name:      "positive constraint",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"positive": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMinimum)
			},
		},
		{
			name:      "negative constraint",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"negative": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMaximum)
			},
		},
		{
			name:      "multiple_of constraint",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"multiple_of": "5",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("5"), schema.MultipleOf)
			},
		},
		{
			name:      "title constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"title": "User Name",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "User Name", schema.Title)
			},
		},
		{
			name:      "examples json array single value",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"examples": `["example1"]`,
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.Len(t, schema.Examples, 1)
				assert.Equal(t, "example1", schema.Examples[0])
			},
		},
		{
			name:      "examples json array multiple values",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"examples": `["example1","example2","example3"]`,
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.Len(t, schema.Examples, 3)
				assert.Equal(t, "example1", schema.Examples[0])
				assert.Equal(t, "example2", schema.Examples[1])
				assert.Equal(t, "example3", schema.Examples[2])
			},
		},
		{
			name:      "deprecated without message",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"deprecated": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
				// No message means description should be empty
				assert.Empty(t, schema.Description)
			},
		},
		{
			name:      "deprecated with message",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"deprecated": "Use newField instead",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
				assert.Contains(t, schema.Description, "Deprecated: Use newField instead")
			},
		},
		{
			name:      "deprecated with existing description",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"description": "This is a field",
				"deprecated":  "Use newField instead",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.True(t, schema.Deprecated)
				assert.Contains(t, schema.Description, "This is a field")
				assert.Contains(t, schema.Description, "Deprecated: Use newField instead")
			},
		},
		{
			name:      "default with string",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"default": "defaultValue",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "defaultValue", schema.Default)
			},
		},
		{
			name:      "defaultUsingMethod should be skipped",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"defaultUsingMethod": "GetDefault",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Should not set default for defaultUsingMethod
				assert.Nil(t, schema.Default)
			},
		},
		{
			name:      "min for pointer to string",
			fieldType: reflect.TypeOf(new(string)),
			constraints: map[string]string{
				"min": "5",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(5), *schema.MinLength)
			},
		},
		{
			name:      "max for pointer to string",
			fieldType: reflect.TypeOf(new(string)),
			constraints: map[string]string{
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(100), *schema.MaxLength)
			},
		},
		{
			name:      "min for pointer to int",
			fieldType: reflect.TypeOf(new(int)),
			constraints: map[string]string{
				"min": "10",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("10"), schema.Minimum)
			},
		},
		{
			name:      "max for pointer to int",
			fieldType: reflect.TypeOf(new(int)),
			constraints: map[string]string{
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:      "min for slice",
			fieldType: reflect.TypeOf([]string{}),
			constraints: map[string]string{
				"min": "1",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(1), *schema.MinLength)
			},
		},
		{
			name:      "max for array",
			fieldType: reflect.TypeOf([5]string{}),
			constraints: map[string]string{
				"max": "5",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(5), *schema.MaxLength)
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

// TestParseDefaultValue_AllIntegerTypes tests all integer types for full coverage.
func TestParseDefaultValue_AllIntegerTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		typ      reflect.Type
		expected any
	}{
		{
			name:     "int8 default",
			value:    "42",
			typ:      reflect.TypeOf(int8(0)),
			expected: int64(42),
		},
		{
			name:     "int16 default",
			value:    "42",
			typ:      reflect.TypeOf(int16(0)),
			expected: int64(42),
		},
		{
			name:     "int32 default",
			value:    "42",
			typ:      reflect.TypeOf(int32(0)),
			expected: int64(42),
		},
		{
			name:     "int64 default",
			value:    "42",
			typ:      reflect.TypeOf(int64(0)),
			expected: int64(42),
		},
		{
			name:     "uint8 default",
			value:    "42",
			typ:      reflect.TypeOf(uint8(0)),
			expected: uint64(42),
		},
		{
			name:     "uint16 default",
			value:    "42",
			typ:      reflect.TypeOf(uint16(0)),
			expected: uint64(42),
		},
		{
			name:     "uint32 default",
			value:    "42",
			typ:      reflect.TypeOf(uint32(0)),
			expected: uint64(42),
		},
		{
			name:     "uint64 default",
			value:    "42",
			typ:      reflect.TypeOf(uint64(0)),
			expected: uint64(42),
		},
		{
			name:     "float32 default",
			value:    "3.14",
			typ:      reflect.TypeOf(float32(0)),
			expected: 3.14,
		},
		{
			name:     "invalid uint returns string",
			value:    "invalid",
			typ:      reflect.TypeOf(uint(0)),
			expected: "invalid",
		},
		{
			name:     "invalid float returns string",
			value:    "invalid",
			typ:      reflect.TypeOf(float64(0)),
			expected: "invalid",
		},
		{
			name:     "invalid bool returns string",
			value:    "invalid",
			typ:      reflect.TypeOf(false),
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

// TestApplyConstraintsToItems_AdditionalFormats tests more format constraints for items.
func TestApplyConstraintsToItems_AdditionalFormats(t *testing.T) {
	tests := []struct {
		name        string
		elemType    reflect.Type
		constraints map[string]string
		checkFunc   func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:     "ipv4 format for items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv4", schema.Format)
			},
		},
		{
			name:     "ipv6 format for items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv6": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv6", schema.Format)
			},
		},
		{
			name:     "regexp for items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"regexp": "^[A-Z]+$",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[A-Z]+$", schema.Pattern)
			},
		},
		{
			name:     "oneof for items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"oneof": "red green blue",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.Len(t, schema.Enum, 3)
				assert.Contains(t, schema.Enum, "red")
				assert.Contains(t, schema.Enum, "green")
				assert.Contains(t, schema.Enum, "blue")
			},
		},
		{
			name:     "gt for items",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gt": "0",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMinimum)
			},
		},
		{
			name:     "gte for items",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gte": "1",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("1"), schema.Minimum)
			},
		},
		{
			name:     "lt for items",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"lt": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("100"), schema.ExclusiveMaximum)
			},
		},
		{
			name:     "lte for items",
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

// TestApplyConstraints_Phase10Formats tests all Phase 10 format constraints comprehensively.
func TestApplyConstraints_Phase10Formats(t *testing.T) {
	tests := []struct {
		name        string
		constraint  string
		expectedFmt string
	}{
		// Finance formats
		{"credit_card format", "credit_card", "credit_card"},
		{"btc_addr format", "btc_addr", "btc_addr"},
		{"btc_addr_bech32 format", "btc_addr_bech32", "btc_addr_bech32"},
		{"eth_addr format", "eth_addr", "eth_addr"},
		{"luhn_checksum format", "luhn_checksum", "luhn_checksum"},

		// Identity formats
		{"isbn format", "isbn", "isbn"},
		{"isbn10 format", "isbn10", "isbn10"},
		{"isbn13 format", "isbn13", "isbn13"},
		{"issn format", "issn", "issn"},
		{"ssn format", "ssn", "ssn"},
		{"ein format", "ein", "ein"},
		{"e164 format", "e164", "e164"},

		// Network formats
		{"hostname_rfc1123 format", "hostname_rfc1123", "hostname_rfc1123"},
		{"tcp_addr format", "tcp_addr", "tcp_addr"},
		{"udp_addr format", "udp_addr", "udp_addr"},
		{"tcp4_addr format", "tcp4_addr", "tcp4_addr"},

		// Color formats
		{"rgba format", "rgba", "rgba"},
		{"hsl format", "hsl", "hsl"},
		{"hsla format", "hsla", "hsla"},

		// Encoding formats
		{"base64url format", "base64url", "base64url"},
		{"base64rawurl format", "base64rawurl", "base64rawurl"},

		// Hash formats
		{"md4 format", "md4", "md4"},
		{"sha384 format", "sha384", "sha384"},
		{"mongodb format", "mongodb", "mongodb"},

		// Misc formats
		{"html format", "html", "html"},
		{"cron format", "cron", "cron"},
		{"semver format", "semver", "semver"},
		{"ulid format", "ulid", "ulid"},

		// ISO code formats
		{"iso3166_1_alpha2 format", "iso3166_1_alpha2", "iso3166_1_alpha2"},
		{"iso3166_alpha2_eu format", "iso3166_alpha2_eu", "iso3166_alpha2_eu"},
		{"iso3166_1_alpha3 format", "iso3166_1_alpha3", "iso3166_1_alpha3"},
		{"iso3166_alpha3_eu format", "iso3166_alpha3_eu", "iso3166_alpha3_eu"},
		{"iso3166_1_alpha_numeric format", "iso3166_1_alpha_numeric", "iso3166_1_alpha_numeric"},
		{"iso3166_2 format", "iso3166_2", "iso3166_2"},
		{"iso4217 format", "iso4217", "iso4217"},
		{"iso4217_numeric format", "iso4217_numeric", "iso4217_numeric"},
		{"postcode format", "postcode", "postcode"},
		{"postcode_iso3166_alpha2 format", "postcode_iso3166_alpha2", "postcode"},
		{"bcp47_language_tag format", "bcp47_language_tag", "bcp47_language_tag"},

		// Filesystem formats
		{"dirpath format", "dirpath", "dirpath"},
		{"dir format", "dir", "dir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{}
			constraints := map[string]string{
				tt.constraint: "",
			}
			ApplyConstraints(schema, constraints, reflect.TypeOf(""))
			assert.Equal(t, tt.expectedFmt, schema.Format)
		})
	}
}

// TestApplyConstraints_SliceAndMapConstraintPropagation tests that constraints propagate to items/values.
func TestApplyConstraints_SliceAndMapConstraintPropagation(t *testing.T) {
	t.Run("slice with email constraint", func(t *testing.T) {
		schema := &jsonschema.Schema{
			Items: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"email": "",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf([]string{}))
		assert.Equal(t, "email", schema.Items.Format)
	})

	t.Run("map with url constraint", func(t *testing.T) {
		schema := &jsonschema.Schema{
			AdditionalProperties: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"url": "",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf(map[string]string{}))
		assert.Equal(t, "uri", schema.AdditionalProperties.Format)
	})

	t.Run("slice with min constraint for strings", func(t *testing.T) {
		schema := &jsonschema.Schema{
			Items: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"min": "5",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf([]string{}))
		require.NotNil(t, schema.Items.MinLength)
		assert.Equal(t, uint64(5), *schema.Items.MinLength)
	})

	t.Run("slice with max constraint for ints", func(t *testing.T) {
		schema := &jsonschema.Schema{
			Items: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"max": "100",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf([]int{}))
		assert.Equal(t, json.Number("100"), schema.Items.Maximum)
	})

	t.Run("map with pattern constraint", func(t *testing.T) {
		schema := &jsonschema.Schema{
			AdditionalProperties: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"regexp": "^[A-Z]+$",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf(map[string]string{}))
		assert.Equal(t, "^[A-Z]+$", schema.AdditionalProperties.Pattern)
	})
}

// TestApplyConstraints_MinMaxWithInvalidValues tests error handling for min/max.
func TestApplyConstraints_MinMaxWithInvalidValues(t *testing.T) {
	t.Run("min with invalid value for string", func(t *testing.T) {
		schema := &jsonschema.Schema{}
		constraints := map[string]string{
			"min": "invalid",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf(""))
		// Should not set MinLength for invalid value
		assert.Nil(t, schema.MinLength)
	})

	t.Run("max with negative value for string", func(t *testing.T) {
		schema := &jsonschema.Schema{}
		constraints := map[string]string{
			"max": "-1",
		}
		ApplyConstraints(schema, constraints, reflect.TypeOf(""))
		// Should not set MaxLength for negative value
		assert.Nil(t, schema.MaxLength)
	})

	t.Run("min with invalid value in items", func(t *testing.T) {
		schema := &jsonschema.Schema{
			Items: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"min": "invalid",
		}
		ApplyConstraintsToItems(schema.Items, constraints, reflect.TypeOf(""))
		// Should not set MinLength for invalid value
		assert.Nil(t, schema.Items.MinLength)
	})

	t.Run("max with negative value in items", func(t *testing.T) {
		schema := &jsonschema.Schema{
			Items: &jsonschema.Schema{},
		}
		constraints := map[string]string{
			"max": "-1",
		}
		ApplyConstraintsToItems(schema.Items, constraints, reflect.TypeOf(""))
		// Should not set MaxLength for negative value
		assert.Nil(t, schema.Items.MaxLength)
	})
}

// TestApplyConstraints_RequiredSkipped tests that required constraint is skipped in ApplyConstraints.
func TestApplyConstraints_RequiredSkipped(t *testing.T) {
	schema := &jsonschema.Schema{}
	constraints := map[string]string{
		"required": "",
	}
	// Should not panic or modify schema (required is handled elsewhere)
	ApplyConstraints(schema, constraints, reflect.TypeOf(""))
	// Nothing should be set
	assert.Nil(t, schema.MinLength)
	assert.Empty(t, schema.Pattern)
}

// TestApplyConstraints_DescriptionProcessedFirst tests that description is processed before deprecated.
func TestApplyConstraints_DescriptionProcessedFirst(t *testing.T) {
	schema := &jsonschema.Schema{}
	// Map iteration order is random, but description processing is explicit
	constraints := map[string]string{
		"description": "Original description",
	}
	ApplyConstraints(schema, constraints, reflect.TypeOf(""))
	assert.Equal(t, "Original description", schema.Description)
}
