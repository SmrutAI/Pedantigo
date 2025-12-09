// Package schema provides JSON Schema generation and enhancement utilities for pedantigo validators.
package schema

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
)

// GenerateBaseSchema creates base JSON schema for a type (all nested structs inlined)
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

// GenerateOpenAPIBaseSchema creates base JSON schema with $ref support for OpenAPI
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

// EnhanceNestedTypes handles nested structs, slices, and maps
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

// ApplyConstraints applies validation constraints to a JSON Schema
func ApplyConstraints(schema *jsonschema.Schema, constraintsMap map[string]string, fieldType reflect.Type) {
	for name, value := range constraintsMap {
		switch name {
		case "required":
			// Already handled in EnhanceSchema
			continue

		case "min":
			// Context-aware: numeric min for numbers, length min for strings/arrays
			// Handle pointer types
			checkType := fieldType
			if checkType.Kind() == reflect.Ptr {
				checkType = checkType.Elem()
			}
			kind := checkType.Kind()
			if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
				// min → minLength for strings/arrays
				if minLength, err := strconv.Atoi(value); err == nil {
					ml := uint64(minLength)
					schema.MinLength = &ml
				}
			} else {
				// min → minimum for numbers
				schema.Minimum = json.Number(value)
			}

		case "max":
			// Context-aware: numeric max for numbers, length max for strings/arrays
			// Handle pointer types
			checkType := fieldType
			if checkType.Kind() == reflect.Ptr {
				checkType = checkType.Elem()
			}
			kind := checkType.Kind()
			if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
				// max → maxLength for strings/arrays
				if maxLength, err := strconv.Atoi(value); err == nil {
					ml := uint64(maxLength)
					schema.MaxLength = &ml
				}
			} else {
				// max → maximum for numbers
				schema.Maximum = json.Number(value)
			}

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

		case "email":
			// email → format: email
			schema.Format = "email"

		case "url":
			// url → format: uri
			schema.Format = "uri"

		case "uuid":
			// uuid → format: uuid
			schema.Format = "uuid"

		case "ipv4":
			// ipv4 → format: ipv4
			schema.Format = "ipv4"

		case "ipv6":
			// ipv6 → format: ipv6
			schema.Format = "ipv6"

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

// ApplyConstraintsToItems applies constraints to array items or map values
func ApplyConstraintsToItems(schema *jsonschema.Schema, constraintsMap map[string]string, elemType reflect.Type) {
	// Skip constraints that don't apply to elements
	for name, value := range constraintsMap {
		switch name {
		case "email":
			schema.Format = "email"
		case "url":
			schema.Format = "uri"
		case "uuid":
			schema.Format = "uuid"
		case "ipv4":
			schema.Format = "ipv4"
		case "ipv6":
			schema.Format = "ipv6"
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
				if minLength, err := strconv.Atoi(value); err == nil {
					ml := uint64(minLength)
					schema.MinLength = &ml
				}
			} else {
				schema.Minimum = json.Number(value)
			}
		case "max":
			// Context-aware for element type
			kind := elemType.Kind()
			if kind == reflect.String {
				if maxLength, err := strconv.Atoi(value); err == nil {
					ml := uint64(maxLength)
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

// ParseDefaultValue converts a string default value to the appropriate type
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
