package validator

import (
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"

	"github.com/SmrutAI/pedantigo/v2/validator/internal/schemagen"
	"github.com/SmrutAI/pedantigo/v2/validator/internal/tags"
)

// parseTagFunc creates a closure that parses tags using the validator's configured tag name.
// This enables custom tag name support (e.g., overriding the default "validate" with "binding").
func (v *Validator[T]) parseTagFunc() func(reflect.StructTag) map[string]string {
	return func(tag reflect.StructTag) map[string]string {
		return tags.ParseTagWithName(tag, v.tagName)
	}
}

// Schema generates a JSON Schema from the validator's type T
// The schema includes all validation constraints mapped to JSON Schema properties
// Schema implements the method.
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
	actualSchema := schemagen.GenerateBaseSchema[T]()

	// Enhance schema with our custom constraints using the configured tag name
	schemagen.EnhanceSchema(actualSchema, v.typ, v.parseTagFunc())

	// Cache result
	v.cachedSchema = actualSchema
	return actualSchema
}

// SchemaJSON generates JSON Schema as JSON bytes for LLM APIs
// Returns expanded schema with nested objects inlined (no $ref/$defs)
// Use this for: OpenAI function calling, Anthropic tool use, Claude structured outputs
// SchemaJSON implements the method.
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
	schemagen.EnhanceSchema(actualSchema, v.typ, v.parseTagFunc())

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

// SchemaOpenAPI generates a JSON Schema compatible with OpenAPI 3.1 specifications.
// Returns schema with $ref/$defs for type reusability and cleaner documentation.
//
// Note: This generates a component schema (for use in components/schemas), not a complete
// OpenAPI document. Pedantigo is a validation library, not an API framework like Huma.
// Use this for: OpenAPI 3.1 specs, API documentation tools, embedding in OpenAPI documents.
// SchemaOpenAPI implements the method.
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

// SchemaJSONOpenAPI generates JSON Schema as JSON bytes for OpenAPI/Swagger specs.
// Returns schema with $ref/$defs for type reusability.
// Use this for: OpenAPI 3.0 specs, Swagger documentation, API documentation tools.
// SchemaJSONOpenAPI implements the method.
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

// enhanceSchemaWithDefs enhances both root schema and all definitions.
func (v *Validator[T]) enhanceSchemaWithDefs(schema *jsonschema.Schema, typ reflect.Type) {
	// Clear the required fields set by jsonschema library
	// We'll add our own based on validate:"required" tags (or custom tag name)
	schema.Required = nil

	// Get parse function once and reuse for all schemas
	parseFunc := v.parseTagFunc()

	// Enhance root schema
	schemagen.EnhanceSchema(schema, typ, parseFunc)

	// Enhance all definitions
	for name, def := range schema.Definitions {
		def.Required = nil
		// Find the type for this definition
		if defTyp := v.findTypeForDefinition(typ, name); defTyp != nil {
			schemagen.EnhanceSchema(def, defTyp, parseFunc)
		}
	}
}

// findTypeForDefinition finds the reflect.Type for a definition by name.
func (v *Validator[T]) findTypeForDefinition(typ reflect.Type, defName string) reflect.Type {
	if typ.Kind() == reflect.Pointer {
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

		if fieldType.Kind() == reflect.Pointer {
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
			if found := v.searchSliceType(fieldType, defName); found != nil {
				return found
			}
		}

		// Search in map value types
		if fieldType.Kind() == reflect.Map {
			if found := v.searchMapType(fieldType, defName); found != nil {
				return found
			}
		}
	}

	return nil
}

// searchSliceType searches for a type within slice element types.
func (v *Validator[T]) searchSliceType(fieldType reflect.Type, defName string) reflect.Type {
	elemType := fieldType.Elem()
	if elemType.Kind() == reflect.Pointer {
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
	return nil
}

// searchMapType searches for a type within map value types.
func (v *Validator[T]) searchMapType(fieldType reflect.Type, defName string) reflect.Type {
	valueType := fieldType.Elem()
	if valueType.Kind() == reflect.Pointer {
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
	return nil
}

// SchemaLLM generates a JSON Schema optimized for LLM APIs (no $schema field).
// Returns expanded schema with nested objects inlined (no $ref/$defs).
// The $schema field is omitted because some LLMs (like Groq) echo it back in responses.
// Use this for: OpenAI function calling, Anthropic tool use, Claude structured outputs.
//
// Note: This has independent caching from Schema() to allow different optimizations.
func (v *Validator[T]) SchemaLLM() *jsonschema.Schema {
	// Fast path: read lock check for cached LLM schema
	v.schemaMu.RLock()
	if v.cachedSchemaLLM != nil {
		cached := v.cachedSchemaLLM
		v.schemaMu.RUnlock()
		return cached
	}
	v.schemaMu.RUnlock()

	// Slow path: generate and cache
	v.schemaMu.Lock()
	defer v.schemaMu.Unlock()

	// Double-check (another goroutine may have cached it while we waited for the lock)
	if v.cachedSchemaLLM != nil {
		return v.cachedSchemaLLM
	}

	// Generate base schema using schema package
	actualSchema := schemagen.GenerateBaseSchema[T]()

	// Enhance schema with our custom constraints using the configured tag name
	schemagen.EnhanceSchema(actualSchema, v.typ, v.parseTagFunc())

	// Clear $schema and $id fields for LLM compatibility.
	// Version ($schema) and ID ($id) both have omitempty JSON tags,
	// so setting them to empty values omits them from marshaled output.
	actualSchema.Version = ""
	actualSchema.ID = ""

	// Cache result
	v.cachedSchemaLLM = actualSchema
	return actualSchema
}

// SchemaJSONLLM generates JSON Schema as JSON bytes optimized for LLM APIs.
// Returns expanded schema with nested objects inlined (no $ref/$defs) and no $schema field.
// Use this for: OpenAI function calling, Anthropic tool use, Claude structured outputs.
//
// Note: This has independent caching from SchemaJSON() to allow different optimizations.
func (v *Validator[T]) SchemaJSONLLM() ([]byte, error) {
	// Fast path: read lock check for cached LLM JSON
	v.schemaMu.RLock()
	if v.cachedLLMJSON != nil {
		cached := v.cachedLLMJSON
		v.schemaMu.RUnlock()
		return cached, nil
	}
	// Check if LLM schema is cached (we'll marshal it)
	if v.cachedSchemaLLM != nil {
		schema := v.cachedSchemaLLM
		v.schemaMu.RUnlock()

		// Marshal outside lock
		jsonBytes, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return nil, err
		}

		// Cache the JSON bytes
		v.schemaMu.Lock()
		v.cachedLLMJSON = jsonBytes
		v.schemaMu.Unlock()

		return jsonBytes, nil
	}
	v.schemaMu.RUnlock()

	// Slow path: generate LLM schema and JSON, then cache both
	v.schemaMu.Lock()
	defer v.schemaMu.Unlock()

	// Double-check both caches
	if v.cachedLLMJSON != nil {
		return v.cachedLLMJSON, nil
	}

	// Generate schema WITHOUT calling SchemaLLM() to avoid deadlock
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
	schemagen.EnhanceSchema(actualSchema, v.typ, v.parseTagFunc())

	// Clear $schema and $id fields for LLM compatibility
	actualSchema.Version = ""
	actualSchema.ID = ""

	// Cache LLM schema
	v.cachedSchemaLLM = actualSchema

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(actualSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	// Cache JSON bytes
	v.cachedLLMJSON = jsonBytes
	return jsonBytes, nil
}
