package pedantigo

import (
	"encoding/json"
	"reflect"

	"github.com/SmrutAI/Pedantigo/internal/tags"
	"github.com/SmrutAI/Pedantigo/schemagen"
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
	actualSchema := schemagen.GenerateBaseSchema[T]()

	// Enhance schema with our custom constraints
	schemagen.EnhanceSchema(actualSchema, v.typ, tags.ParseTag)

	// Cache result
	v.cachedSchema = actualSchema
	return actualSchema
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
	schemagen.EnhanceSchema(actualSchema, v.typ, tags.ParseTag)

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
	schemagen.EnhanceSchema(schema, typ, tags.ParseTag)

	// Enhance all definitions
	for name, def := range schema.Definitions {
		def.Required = nil
		// Find the type for this definition
		if defTyp := v.findTypeForDefinition(typ, name); defTyp != nil {
			schemagen.EnhanceSchema(def, defTyp, tags.ParseTag)
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
