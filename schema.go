package pedantigo

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/SmrutAI/Pedantigo/schema"
	"github.com/invopop/jsonschema"
)

// Schema generates a JSON Schema from the validator's type T
// The schema includes all validation constraints mapped to JSON Schema properties
func (v *Validator[T]) Schema() *jsonschema.Schema {
	// Fast path: read lock check for cached schema
	v.schemaMu.RLock()
	if v.cachedSchema != nil {
		cached := v.cachedSchema
		v.schemaMu.RUnlock()
		return cached
	}
	v.schemaMu.RUnlock()

	// Slow path: generate and cache
	v.schemaMu.Lock()
	defer v.schemaMu.Unlock()

	// Double-check (another goroutine may have cached it while we waited for the lock)
	if v.cachedSchema != nil {
		return v.cachedSchema
	}

	// Generate base schema using schema package
	actualSchema := schema.GenerateBaseSchema[T]()

	// Enhance schema with our custom constraints
	schema.EnhanceSchema(actualSchema, v.typ, parseTag)

	// Cache result
	v.cachedSchema = actualSchema
	return actualSchema
}

// enhanceSchema recursively enhances a JSON Schema with validation constraints
func (v *Validator[T]) enhanceSchema(schema *jsonschema.Schema, typ reflect.Type) {
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
		constraints := parseTag(field.Tag)
		if constraints == nil {
			// No constraints, but check for nested structs/slices/maps
			v.enhanceNestedTypes(fieldSchema, field.Type)
			continue
		}

		// Apply constraints to field schema
		v.applyConstraints(fieldSchema, constraints, field.Type)

		// Check for required constraint
		if _, hasRequired := constraints["required"]; hasRequired {
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
		v.enhanceNestedTypes(fieldSchema, field.Type)
	}
}

// enhanceNestedTypes handles nested structs, slices, and maps
func (v *Validator[T]) enhanceNestedTypes(schema *jsonschema.Schema, typ reflect.Type) {
	switch typ.Kind() {
	case reflect.Struct:
		// Recursively enhance nested struct
		if typ != reflect.TypeOf((*time.Time)(nil)).Elem() {
			// Clear required fields set by jsonschema for nested structs
			schema.Required = nil
			v.enhanceSchema(schema, typ)
		}

	case reflect.Slice:
		// Enhance array items
		if schema.Items != nil {
			elemType := typ.Elem()
			if elemType.Kind() == reflect.Struct {
				// Clear required fields for nested struct items
				schema.Items.Required = nil
				v.enhanceSchema(schema.Items, elemType)
			}
		}

	case reflect.Map:
		// Enhance map values
		if schema.AdditionalProperties != nil {
			valueType := typ.Elem()
			if valueType.Kind() == reflect.Struct {
				// Clear required fields for nested struct values
				schema.AdditionalProperties.Required = nil
				v.enhanceSchema(schema.AdditionalProperties, valueType)
			}
		}
	}
}

// applyConstraints applies validation constraints to a JSON Schema
func (v *Validator[T]) applyConstraints(schema *jsonschema.Schema, constraints map[string]string, fieldType reflect.Type) {
	for name, value := range constraints {
		switch name {
		case "required":
			// Already handled in enhanceSchema
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
			schema.Default = parseDefaultValue(value, fieldType)

		case "defaultUsingMethod":
			// Skip - this is runtime behavior, not schema
			continue
		}
	}

	// For slices, apply constraints to items as well
	if fieldType.Kind() == reflect.Slice && schema.Items != nil {
		v.applyConstraintsToItems(schema.Items, constraints, fieldType.Elem())
	}

	// For maps, apply constraints to additionalProperties as well
	if fieldType.Kind() == reflect.Map && schema.AdditionalProperties != nil {
		v.applyConstraintsToItems(schema.AdditionalProperties, constraints, fieldType.Elem())
	}
}

// applyConstraintsToItems applies constraints to array items or map values
func (v *Validator[T]) applyConstraintsToItems(schema *jsonschema.Schema, constraints map[string]string, elemType reflect.Type) {
	// Skip constraints that don't apply to elements
	for name, value := range constraints {
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

// parseDefaultValue converts a string default value to the appropriate type
func parseDefaultValue(value string, typ reflect.Type) any {
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

// SchemaJSON generates JSON Schema as JSON bytes for LLM APIs
// Returns expanded schema with nested objects inlined (no $ref/$defs)
// Use this for: OpenAI function calling, Anthropic tool use, Claude structured outputs
func (v *Validator[T]) SchemaJSON() ([]byte, error) {
	// Fast path: read lock check for cached JSON
	v.schemaMu.RLock()
	if v.cachedSchemaJSON != nil {
		cached := v.cachedSchemaJSON
		v.schemaMu.RUnlock()
		return cached, nil
	}
	// Check if schema is cached (we'll marshal it)
	if v.cachedSchema != nil {
		schema := v.cachedSchema
		v.schemaMu.RUnlock()

		// Marshal outside lock
		jsonBytes, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return nil, err
		}

		// Cache the JSON bytes
		v.schemaMu.Lock()
		v.cachedSchemaJSON = jsonBytes
		v.schemaMu.Unlock()

		return jsonBytes, nil
	}
	v.schemaMu.RUnlock()

	// Slow path: generate schema and JSON, then cache both
	v.schemaMu.Lock()
	defer v.schemaMu.Unlock()

	// Double-check both caches
	if v.cachedSchemaJSON != nil {
		return v.cachedSchemaJSON, nil
	}

	// Generate schema WITHOUT calling Schema() to avoid deadlock
	var zero T
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: true,
	}
	baseSchema := reflector.Reflect(zero)

	actualSchema := baseSchema
	if baseSchema.Properties == nil && len(baseSchema.Definitions) > 0 {
		for _, def := range baseSchema.Definitions {
			if def.Type == "object" && def.Properties != nil {
				actualSchema = def
				break
			}
		}
	}

	actualSchema.Required = nil
	v.enhanceSchema(actualSchema, v.typ)

	// Cache schema
	v.cachedSchema = actualSchema

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(actualSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	// Cache JSON bytes
	v.cachedSchemaJSON = jsonBytes
	return jsonBytes, nil
}

// SchemaOpenAPI generates a JSON Schema with $ref support for OpenAPI/Swagger specs
// Returns schema with $ref/$defs for type reusability and cleaner documentation
// Use this for: OpenAPI 3.0 specs, Swagger documentation, API documentation tools
func (v *Validator[T]) SchemaOpenAPI() *jsonschema.Schema {
	// Fast path: read lock check for cached OpenAPI schema
	v.schemaMu.RLock()
	if v.cachedOpenAPI != nil {
		cached := v.cachedOpenAPI
		v.schemaMu.RUnlock()
		return cached
	}
	v.schemaMu.RUnlock()

	// Slow path: generate and cache
	v.schemaMu.Lock()
	defer v.schemaMu.Unlock()

	// Double-check (another goroutine may have cached it while we waited for the lock)
	if v.cachedOpenAPI != nil {
		return v.cachedOpenAPI
	}

	var zero T
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,  // Expand root struct inline
		DoNotReference: false, // Allow $ref/$defs for nested types
	}
	baseSchema := reflector.Reflect(zero)

	// Enhance all schemas (root and definitions) with constraints
	v.enhanceSchemaWithDefs(baseSchema, v.typ)

	// Cache result
	v.cachedOpenAPI = baseSchema
	return baseSchema
}

// SchemaJSONOpenAPI generates JSON Schema as JSON bytes for OpenAPI/Swagger specs
// Returns schema with $ref/$defs for type reusability
// Use this for: OpenAPI 3.0 specs, Swagger documentation, API documentation tools
func (v *Validator[T]) SchemaJSONOpenAPI() ([]byte, error) {
	// Fast path: read lock check for cached OpenAPI JSON
	v.schemaMu.RLock()
	if v.cachedOpenAPIJSON != nil {
		cached := v.cachedOpenAPIJSON
		v.schemaMu.RUnlock()
		return cached, nil
	}
	// Check if OpenAPI schema is cached (we'll marshal it)
	if v.cachedOpenAPI != nil {
		schema := v.cachedOpenAPI
		v.schemaMu.RUnlock()

		// Marshal outside lock
		jsonBytes, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return nil, err
		}

		// Cache the JSON bytes
		v.schemaMu.Lock()
		v.cachedOpenAPIJSON = jsonBytes
		v.schemaMu.Unlock()

		return jsonBytes, nil
	}
	v.schemaMu.RUnlock()

	// Slow path: generate OpenAPI schema and JSON, then cache both
	v.schemaMu.Lock()
	defer v.schemaMu.Unlock()

	// Double-check both caches
	if v.cachedOpenAPIJSON != nil {
		return v.cachedOpenAPIJSON, nil
	}

	// Generate OpenAPI schema WITHOUT calling SchemaOpenAPI() to avoid deadlock
	var zero T
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: false, // Allow $ref/$defs
	}
	baseSchema := reflector.Reflect(zero)

	v.enhanceSchemaWithDefs(baseSchema, v.typ)

	// Cache OpenAPI schema
	v.cachedOpenAPI = baseSchema

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(baseSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	// Cache JSON bytes
	v.cachedOpenAPIJSON = jsonBytes
	return jsonBytes, nil
}

// enhanceSchemaWithDefs enhances both root schema and all definitions
func (v *Validator[T]) enhanceSchemaWithDefs(schema *jsonschema.Schema, typ reflect.Type) {
	// Clear the required fields set by jsonschema library
	// We'll add our own based on pedantigo:"required" tags
	schema.Required = nil

	// Enhance root schema
	v.enhanceSchema(schema, typ)

	// Enhance all definitions
	for name, def := range schema.Definitions {
		def.Required = nil
		// Find the type for this definition
		if defTyp := v.findTypeForDefinition(typ, name); defTyp != nil {
			v.enhanceSchema(def, defTyp)
		}
	}
}

// findTypeForDefinition finds the reflect.Type for a definition by name
func (v *Validator[T]) findTypeForDefinition(typ reflect.Type, defName string) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	// Check if this is the type we're looking for
	if typ.Name() == defName {
		return typ
	}

	// Search through struct fields for nested types
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check if this field type matches
		if fieldType.Name() == defName {
			return fieldType
		}

		// Recursively search nested structs
		if fieldType.Kind() == reflect.Struct {
			if found := v.findTypeForDefinition(fieldType, defName); found != nil {
				return found
			}
		}

		// Search in slice element types
		if fieldType.Kind() == reflect.Slice {
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			if elemType.Name() == defName {
				return elemType
			}
			if elemType.Kind() == reflect.Struct {
				if found := v.findTypeForDefinition(elemType, defName); found != nil {
					return found
				}
			}
		}

		// Search in map value types
		if fieldType.Kind() == reflect.Map {
			valueType := fieldType.Elem()
			if valueType.Kind() == reflect.Ptr {
				valueType = valueType.Elem()
			}
			if valueType.Name() == defName {
				return valueType
			}
			if valueType.Kind() == reflect.Struct {
				if found := v.findTypeForDefinition(valueType, defName); found != nil {
					return found
				}
			}
		}
	}

	return nil
}
