// Package schemagen provides JSON Schema generation and enhancement utilities for pedantigo validators.
package schemagen

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
)

// Format constraint name constants.
const (
	fmtEmail = "email"
	fmtURL   = "url"
	fmtUUID  = "uuid"
	fmtIPv4  = "ipv4"
	fmtIPv6  = "ipv6"

	// Network formats (Phase 10).
	fmtIP          = "ip"
	fmtCIDR        = "cidr"
	fmtCIDRv4      = "cidrv4"
	fmtCIDRv6      = "cidrv6"
	fmtMAC         = "mac"
	fmtHostname    = "hostname"
	fmtHostnameRFC = "hostname_rfc1123"
	fmtFQDN        = "fqdn"
	fmtPort        = "port"
	fmtTCPAddr     = "tcp_addr"
	fmtUDPAddr     = "udp_addr"
	fmtTCP4Addr    = "tcp4_addr"

	// Finance formats (Phase 10).
	fmtCreditCard    = "credit_card"
	fmtBTCAddr       = "btc_addr"
	fmtBTCAddrBech32 = "btc_addr_bech32"
	fmtETHAddr       = "eth_addr"
	fmtLuhnChecksum  = "luhn_checksum"

	// Identity formats (Phase 10).
	fmtISBN   = "isbn"
	fmtISBN10 = "isbn10"
	fmtISBN13 = "isbn13"
	fmtISSN   = "issn"
	fmtSSN    = "ssn"
	fmtEIN    = "ein"
	fmtE164   = "e164"

	// Geo formats (Phase 10).
	fmtLatitude  = "latitude"
	fmtLongitude = "longitude"

	// Color formats (Phase 10).
	fmtHexColor = "hexcolor"
	fmtRGB      = "rgb"
	fmtRGBA     = "rgba"
	fmtHSL      = "hsl"
	fmtHSLA     = "hsla"

	// Encoding formats (Phase 10).
	fmtJWT          = "jwt"
	fmtJSON         = "json"
	fmtBase64       = "base64"
	fmtBase64URL    = "base64url"
	fmtBase64RawURL = "base64rawurl"

	// Hash formats (Phase 10).
	fmtMD4     = "md4"
	fmtMD5     = "md5"
	fmtSHA256  = "sha256"
	fmtSHA384  = "sha384"
	fmtSHA512  = "sha512"
	fmtMongoDB = "mongodb"

	// Misc formats (Phase 10).
	fmtHTML   = "html"
	fmtCron   = "cron"
	fmtSemver = "semver"
	fmtULID   = "ulid"
)

// GenerateBaseSchema creates base JSON schema for a type (all nested structs inlined).
func GenerateBaseSchema[T any]() *jsonschema.Schema {
	var zero T
	reflector := jsonschema.Reflector{
		ExpandedStruct: true, // Expand root struct inline
		DoNotReference: true, // Inline ALL nested structs without creating $ref
	}
	baseSchema := reflector.Reflect(zero)

	// If the schema is a reference, unwrap it and return the actual definition
	actualSchema := baseSchema
	if baseSchema.Properties == nil && len(baseSchema.Definitions) > 0 {
		// The jsonschema library creates a reference schema with definitions
		// Find the actual struct schema in the definitions
		for _, def := range baseSchema.Definitions {
			if def.Type == "object" && def.Properties != nil {
				actualSchema = def
				break
			}
		}
	}

	// Clear the required fields set by jsonschema library
	// We'll add our own based on pedantigo:"required" tags
	actualSchema.Required = nil

	return actualSchema
}

// GenerateOpenAPIBaseSchema creates base JSON schema with $ref support for OpenAPI.
func GenerateOpenAPIBaseSchema[T any]() *jsonschema.Schema {
	var zero T
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,  // Expand root struct inline
		DoNotReference: false, // Allow $ref/$defs for nested types
	}
	return reflector.Reflect(zero)
}

// EnhanceSchema recursively enhances a JSON Schema with validation constraints
// parseTagFunc should parse struct tags and return constraint map, or nil if no constraints
// typReflect is the reflect.Type of the struct being enhanced
// EnhanceSchema implements the functionality.
func EnhanceSchema(schema *jsonschema.Schema, typ reflect.Type, parseTagFunc func(reflect.StructTag) map[string]string) {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return
	}

	// Iterate through struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON field name
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name
		if jsonTag != "" && jsonTag != "-" {
			if name, _, found := strings.Cut(jsonTag, ","); found {
				fieldName = name
			} else {
				fieldName = jsonTag
			}
		}

		// Get field's schema property
		if schema.Properties == nil {
			continue
		}
		fieldSchema, ok := schema.Properties.Get(fieldName)
		if !ok || fieldSchema == nil {
			continue
		}

		// Parse validation constraints
		constraintsMap := parseTagFunc(field.Tag)
		if constraintsMap == nil {
			// No constraints, but check for nested structs/slices/maps
			EnhanceNestedTypes(fieldSchema, field.Type, parseTagFunc)
			continue
		}

		// Apply constraints to field schema
		ApplyConstraints(fieldSchema, constraintsMap, field.Type)

		// Check for required constraint
		if _, hasRequired := constraintsMap["required"]; hasRequired {
			// Add to required array if not already there
			found := false
			for _, req := range schema.Required {
				if req == fieldName {
					found = true
					break
				}
			}
			if !found {
				schema.Required = append(schema.Required, fieldName)
			}
		}

		// Handle nested types
		EnhanceNestedTypes(fieldSchema, field.Type, parseTagFunc)
	}
}

// EnhanceNestedTypes handles nested structs, slices, and maps.
func EnhanceNestedTypes(schema *jsonschema.Schema, typ reflect.Type, parseTagFunc func(reflect.StructTag) map[string]string) {
	switch typ.Kind() {
	case reflect.Struct:
		// Recursively enhance nested struct
		if typ != reflect.TypeOf((*time.Time)(nil)).Elem() {
			// Clear required fields set by jsonschema for nested structs
			schema.Required = nil
			EnhanceSchema(schema, typ, parseTagFunc)
		}

	case reflect.Slice:
		// Enhance array items
		if schema.Items != nil {
			elemType := typ.Elem()
			if elemType.Kind() == reflect.Struct {
				// Clear required fields for nested struct items
				schema.Items.Required = nil
				EnhanceSchema(schema.Items, elemType, parseTagFunc)
			}
		}

	case reflect.Map:
		// Enhance map values
		if schema.AdditionalProperties != nil {
			valueType := typ.Elem()
			if valueType.Kind() == reflect.Struct {
				// Clear required fields for nested struct values
				schema.AdditionalProperties.Required = nil
				EnhanceSchema(schema.AdditionalProperties, valueType, parseTagFunc)
			}
		}
	}
}

// ApplyConstraints applies validation constraints to a JSON Schema.
func ApplyConstraints(schema *jsonschema.Schema, constraintsMap map[string]string, fieldType reflect.Type) {
	for name, value := range constraintsMap {
		switch name {
		case "required":
			// Already handled in EnhanceSchema
			continue

		case "min":
			applyMinConstraint(schema, value, fieldType)

		case "max":
			applyMaxConstraint(schema, value, fieldType)

		case "gt":
			// gt → exclusiveMinimum (exclusive)
			schema.ExclusiveMinimum = json.Number(value)

		case "gte":
			// gte → minimum (inclusive)
			schema.Minimum = json.Number(value)

		case "lt":
			// lt → exclusiveMaximum (exclusive)
			schema.ExclusiveMaximum = json.Number(value)

		case "lte":
			// lte → maximum (inclusive)
			schema.Maximum = json.Number(value)

		case fmtEmail, fmtURL, fmtUUID, fmtIPv4, fmtIPv6,
			// Network formats (Phase 10).
			fmtIP, fmtCIDR, fmtCIDRv4, fmtCIDRv6, fmtMAC, fmtHostname, fmtHostnameRFC, fmtFQDN,
			fmtPort, fmtTCPAddr, fmtUDPAddr, fmtTCP4Addr,
			// Finance formats (Phase 10).
			fmtCreditCard, fmtBTCAddr, fmtBTCAddrBech32, fmtETHAddr, fmtLuhnChecksum,
			// Identity formats (Phase 10).
			fmtISBN, fmtISBN10, fmtISBN13, fmtISSN, fmtSSN, fmtEIN, fmtE164,
			// Geo formats (Phase 10).
			fmtLatitude, fmtLongitude,
			// Color formats (Phase 10).
			fmtHexColor, fmtRGB, fmtRGBA, fmtHSL, fmtHSLA,
			// Encoding formats (Phase 10).
			fmtJWT, fmtJSON, fmtBase64, fmtBase64URL, fmtBase64RawURL,
			// Hash formats (Phase 10).
			fmtMD4, fmtMD5, fmtSHA256, fmtSHA384, fmtSHA512, fmtMongoDB,
			// Misc formats (Phase 10).
			fmtHTML, fmtCron, fmtSemver, fmtULID:
			applyFormatConstraint(schema, name)

		case "regexp":
			// regexp → pattern
			schema.Pattern = value

		case "oneof":
			// oneof → enum array (space-separated values)
			values := strings.Fields(value)
			enumValues := make([]any, len(values))
			for i, v := range values {
				enumValues[i] = v
			}
			schema.Enum = enumValues

		case "len":
			// len → minLength + maxLength (exact length)
			if length, err := strconv.Atoi(value); err == nil && length >= 0 {
				l := uint64(length) //nolint:gosec // bounds checked above
				schema.MinLength = &l
				schema.MaxLength = &l
			}

		case "ascii":
			// ascii → pattern for ASCII characters only (0x00-0x7F)
			schema.Pattern = "^[\\x00-\\x7F]*$"

		case "alpha":
			// alpha → pattern for alphabetic characters only (a-z, A-Z)
			schema.Pattern = "^[a-zA-Z]+$"

		case "alphanum":
			// alphanum → pattern for alphanumeric characters only (a-z, A-Z, 0-9)
			schema.Pattern = "^[a-zA-Z0-9]+$"

		case "contains":
			// contains → pattern for substring presence (with escaped special characters)
			escapedSubstring := regexp.QuoteMeta(value)
			schema.Pattern = ".*" + escapedSubstring + ".*"

		case "excludes":
			// excludes → pattern using negative lookahead to exclude substring
			escapedSubstring := regexp.QuoteMeta(value)
			schema.Pattern = "^(?!.*" + escapedSubstring + ").*$"

		case "startswith":
			// startswith → pattern anchored at start
			escapedPrefix := regexp.QuoteMeta(value)
			schema.Pattern = "^" + escapedPrefix + ".*"

		case "endswith":
			// endswith → pattern anchored at end
			escapedSuffix := regexp.QuoteMeta(value)
			schema.Pattern = ".*" + escapedSuffix + "$"

		case "lowercase":
			// lowercase → pattern excluding uppercase letters
			schema.Pattern = "^[^A-Z]*$"

		case "uppercase":
			// uppercase → pattern excluding lowercase letters
			schema.Pattern = "^[^a-z]*$"

		case "positive":
			// positive → exclusiveMinimum of 0
			schema.ExclusiveMinimum = json.Number("0")

		case "negative":
			// negative → exclusiveMaximum of 0
			schema.ExclusiveMaximum = json.Number("0")

		case "multiple_of":
			// multiple_of → multipleOf (JSON Schema keyword)
			schema.MultipleOf = json.Number(value)

		case "default":
			// default → default value
			schema.Default = ParseDefaultValue(value, fieldType)

		case "defaultUsingMethod":
			// Skip - this is runtime behavior, not schema
			continue
		}
	}

	// For slices, apply constraints to items as well
	if fieldType.Kind() == reflect.Slice && schema.Items != nil {
		ApplyConstraintsToItems(schema.Items, constraintsMap, fieldType.Elem())
	}

	// For maps, apply constraints to additionalProperties as well
	if fieldType.Kind() == reflect.Map && schema.AdditionalProperties != nil {
		ApplyConstraintsToItems(schema.AdditionalProperties, constraintsMap, fieldType.Elem())
	}
}

// ApplyConstraintsToItems applies constraints to array items or map values.
func ApplyConstraintsToItems(schema *jsonschema.Schema, constraintsMap map[string]string, elemType reflect.Type) {
	// Skip constraints that don't apply to elements.
	for name, value := range constraintsMap {
		switch name {
		case fmtEmail:
			schema.Format = fmtEmail
		case fmtURL:
			schema.Format = "uri"
		case fmtUUID:
			schema.Format = fmtUUID
		case fmtIPv4:
			schema.Format = fmtIPv4
		case fmtIPv6:
			schema.Format = fmtIPv6
		case "regexp":
			schema.Pattern = value
		case "oneof":
			values := strings.Fields(value)
			enumValues := make([]any, len(values))
			for i, v := range values {
				enumValues[i] = v
			}
			schema.Enum = enumValues
		case "min":
			// Context-aware for element type
			kind := elemType.Kind()
			if kind == reflect.String {
				if minLength, err := strconv.Atoi(value); err == nil && minLength >= 0 {
					ml := uint64(minLength) //nolint:gosec // bounds checked above
					schema.MinLength = &ml
				}
			} else {
				schema.Minimum = json.Number(value)
			}
		case "max":
			// Context-aware for element type
			kind := elemType.Kind()
			if kind == reflect.String {
				if maxLength, err := strconv.Atoi(value); err == nil && maxLength >= 0 {
					ml := uint64(maxLength) //nolint:gosec // bounds checked above
					schema.MaxLength = &ml
				}
			} else {
				schema.Maximum = json.Number(value)
			}
		case "gt":
			schema.ExclusiveMinimum = json.Number(value)
		case "gte":
			schema.Minimum = json.Number(value)
		case "lt":
			schema.ExclusiveMaximum = json.Number(value)
		case "lte":
			schema.Maximum = json.Number(value)
		}
	}
}

// ParseDefaultValue converts a string default value to the appropriate type.
func ParseDefaultValue(value string, typ reflect.Type) any {
	switch typ.Kind() {
	case reflect.String:
		return value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(value, 10, 64); err == nil {
			return u
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return value
}

// applyMinConstraint applies min constraint context-aware to field type.
// For strings/arrays: sets minLength, for numbers: sets minimum.
func applyMinConstraint(schema *jsonschema.Schema, value string, fieldType reflect.Type) {
	checkType := fieldType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	kind := checkType.Kind()
	if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
		// min → minLength for strings/arrays
		if minLength, err := strconv.Atoi(value); err == nil && minLength >= 0 {
			ml := uint64(minLength) //nolint:gosec // bounds checked above
			schema.MinLength = &ml
		}
	} else {
		// min → minimum for numbers
		schema.Minimum = json.Number(value)
	}
}

// applyMaxConstraint applies max constraint context-aware to field type.
// For strings/arrays: sets maxLength, for numbers: sets maximum.
func applyMaxConstraint(schema *jsonschema.Schema, value string, fieldType reflect.Type) {
	checkType := fieldType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	kind := checkType.Kind()
	if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
		// max → maxLength for strings/arrays
		if maxLength, err := strconv.Atoi(value); err == nil && maxLength >= 0 {
			ml := uint64(maxLength) //nolint:gosec // bounds checked above
			schema.MaxLength = &ml
		}
	} else {
		// max → maximum for numbers
		schema.Maximum = json.Number(value)
	}
}

// applyFormatConstraint maps constraint names to JSON Schema format values.
func applyFormatConstraint(schema *jsonschema.Schema, constraintName string) {
	switch constraintName {
	case fmtEmail:
		schema.Format = fmtEmail
	case fmtURL:
		schema.Format = "uri"
	case fmtUUID:
		schema.Format = fmtUUID
	case fmtIPv4:
		schema.Format = fmtIPv4
	case fmtIPv6:
		schema.Format = fmtIPv6

	// Network formats (Phase 10).
	case fmtIP:
		schema.Format = fmtIP
	case fmtCIDR:
		schema.Format = fmtCIDR
	case fmtCIDRv4:
		schema.Format = fmtCIDRv4
	case fmtCIDRv6:
		schema.Format = fmtCIDRv6
	case fmtMAC:
		schema.Format = fmtMAC
	case fmtHostname:
		schema.Format = fmtHostname
	case fmtHostnameRFC:
		schema.Format = fmtHostnameRFC
	case fmtFQDN:
		schema.Format = fmtFQDN
	case fmtPort:
		schema.Format = fmtPort
	case fmtTCPAddr:
		schema.Format = fmtTCPAddr
	case fmtUDPAddr:
		schema.Format = fmtUDPAddr
	case fmtTCP4Addr:
		schema.Format = fmtTCP4Addr

	// Finance formats (Phase 10).
	case fmtCreditCard:
		schema.Format = fmtCreditCard
	case fmtBTCAddr:
		schema.Format = fmtBTCAddr
	case fmtBTCAddrBech32:
		schema.Format = fmtBTCAddrBech32
	case fmtETHAddr:
		schema.Format = fmtETHAddr
	case fmtLuhnChecksum:
		schema.Format = fmtLuhnChecksum

	// Identity formats (Phase 10).
	case fmtISBN:
		schema.Format = fmtISBN
	case fmtISBN10:
		schema.Format = fmtISBN10
	case fmtISBN13:
		schema.Format = fmtISBN13
	case fmtISSN:
		schema.Format = fmtISSN
	case fmtSSN:
		schema.Format = fmtSSN
	case fmtEIN:
		schema.Format = fmtEIN
	case fmtE164:
		schema.Format = fmtE164

	// Geo formats (Phase 10).
	case fmtLatitude:
		schema.Format = fmtLatitude
	case fmtLongitude:
		schema.Format = fmtLongitude

	// Color formats (Phase 10).
	case fmtHexColor:
		schema.Format = fmtHexColor
	case fmtRGB:
		schema.Format = fmtRGB
	case fmtRGBA:
		schema.Format = fmtRGBA
	case fmtHSL:
		schema.Format = fmtHSL
	case fmtHSLA:
		schema.Format = fmtHSLA

	// Encoding formats (Phase 10).
	case fmtJWT:
		schema.Format = fmtJWT
	case fmtJSON:
		schema.Format = fmtJSON
	case fmtBase64:
		schema.Format = fmtBase64
	case fmtBase64URL:
		schema.Format = fmtBase64URL
	case fmtBase64RawURL:
		schema.Format = fmtBase64RawURL

	// Hash formats (Phase 10).
	case fmtMD4:
		schema.Format = fmtMD4
	case fmtMD5:
		schema.Format = fmtMD5
	case fmtSHA256:
		schema.Format = fmtSHA256
	case fmtSHA384:
		schema.Format = fmtSHA384
	case fmtSHA512:
		schema.Format = fmtSHA512
	case fmtMongoDB:
		schema.Format = fmtMongoDB

	// Misc formats (Phase 10).
	case fmtHTML:
		schema.Format = fmtHTML
	case fmtCron:
		schema.Format = fmtCron
	case fmtSemver:
		schema.Format = fmtSemver
	case fmtULID:
		schema.Format = fmtULID
	}
}

// GenerateVariantSchema creates a JSON Schema for a single union variant.
// It generates the base schema for the variant type and adds a const constraint
// on the discriminator field.
// Parameters:
//   - variantType: the reflect.Type of the variant struct
//   - discriminatorField: the JSON field name used as discriminator
//   - discriminatorValue: the const value for this variant
//   - parseTagFunc: function to parse validation tags
//
// Implementation.
func GenerateVariantSchema(variantType reflect.Type, discriminatorField, discriminatorValue string, parseTagFunc func(reflect.StructTag) map[string]string) *jsonschema.Schema {
	// Handle pointer types
	if variantType.Kind() == reflect.Ptr {
		variantType = variantType.Elem()
	}

	// Generate base schema for the variant type using jsonschema reflector
	// Set DoNotReference: true to inline all nested structs
	reflector := jsonschema.Reflector{
		ExpandedStruct: true, // Expand root struct inline
		DoNotReference: true, // Inline ALL nested structs without creating $ref
	}
	// Create a zero value of the type - Reflect() needs a value, not a reflect.Type
	variantZero := reflect.New(variantType).Interface()
	variantSchema := reflector.Reflect(variantZero)

	// If the schema is a reference, unwrap it and return the actual definition
	if variantSchema.Properties == nil && len(variantSchema.Definitions) > 0 {
		// The jsonschema library creates a reference schema with definitions
		// Find the actual struct schema in the definitions
		for _, def := range variantSchema.Definitions {
			if def.Type == "object" && def.Properties != nil {
				variantSchema = def
				break
			}
		}
	}

	// Clear the required fields set by jsonschema library
	variantSchema.Required = nil

	// Add the discriminator field with const constraint
	// Create the const schema for the discriminator field
	discriminatorSchema := &jsonschema.Schema{
		Const: discriminatorValue,
	}

	// Set the discriminator field in Properties
	variantSchema.Properties.Set(discriminatorField, discriminatorSchema)

	// Apply validation constraints using EnhanceSchema
	EnhanceSchema(variantSchema, variantType, parseTagFunc)

	return variantSchema
}

// GenerateUnionSchema creates a JSON Schema with oneOf for discriminated unions.
// Parameters:
//   - discriminatorField: the JSON field name used as discriminator
//   - variants: map of discriminator values to variant types
//   - parseTagFunc: function to parse validation tags
//
// Implementation.
func GenerateUnionSchema(discriminatorField string, variants map[string]reflect.Type, parseTagFunc func(reflect.StructTag) map[string]string) *jsonschema.Schema {
	// Create an empty schema to hold the oneOf array
	unionSchema := &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{},
	}

	// Generate a schema for each variant and add to oneOf array
	for discriminatorValue, variantType := range variants {
		variantSchema := GenerateVariantSchema(variantType, discriminatorField, discriminatorValue, parseTagFunc)
		unionSchema.OneOf = append(unionSchema.OneOf, variantSchema)
	}

	return unionSchema
}
