---
sidebar_position: 4
---

# JSON Schema Generation

Automatically generate JSON Schemas from your Go structs. Pedantigo schemas map validation constraints to standard JSON Schema keywords for API documentation, form generation, LLM structured output, and more.

## Quick Start

Get a JSON Schema with a single function call:

```go
package main

import (
    "fmt"
    "encoding/json"
    "github.com/SmrutAI/pedantigo/v2/validator"
)

type User struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

func main() {
    // Get schema as object
    schema := validator.Schema[User]()

    // Get schema as JSON bytes
    schemaBytes, _ := validator.SchemaJSON[User]()
    fmt.Println(string(schemaBytes))
}
```

Output:
```json
{
  "title": "User",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "minLength": 2,
      "maxLength": 50
    },
    "email": {
      "type": "string",
      "format": "email"
    },
    "age": {
      "type": "integer",
      "minimum": 18,
      "maximum": 120
    }
  },
  "required": ["name", "email", "age"]
}
```

## LLM Tool Calling Compatibility

Pedantigo's default `Schema()` and `SchemaJSON()` produce **flattened schemas** optimized for LLM tool calling APIs.

### Schema Output Comparison

| Function | `$ref/$defs` | `$schema` | `$id` | Best For |
|----------|--------------|-----------|-------|----------|
| `Schema()` / `SchemaJSON()` | **No** - expanded inline | **Yes** | **Yes** | General purpose, form generation |
| `SchemaLLM()` / `SchemaJSONLLM()` | **No** - expanded inline | **No** | **No** | LLM tool calling (OpenAI, Anthropic, Gemini) |
| `SchemaOpenAPI()` / `SchemaJSONOpenAPI()` | **Yes** - uses references | **Yes** | **Yes** | OpenAPI specs, Swagger docs |

:::tip LLM Schema Recommendation
Use `SchemaLLM()` / `SchemaJSONLLM()` for LLM integrations. Both `$schema` and `$id` fields are omitted — some LLMs (like Groq) echo schema metadata fields back in their responses, causing JSON parsing failures.
:::

### Example: Flattened Schema Output

```go
type Address struct {
    Street string `json:"street" validate:"required"`
    City   string `json:"city" validate:"required"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address"`
}

schemaBytes, _ := validator.SchemaJSON[User]()
```

**Output (flattened - ready for LLM tools):**
```json
{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "address": {
      "type": "object",
      "properties": {
        "street": {"type": "string"},
        "city": {"type": "string"}
      },
      "required": ["street", "city"]
    }
  },
  "required": ["name"]
}
```

No `$ref` or `$defs` - the schema works directly with OpenAI, Anthropic, and Gemini function calling APIs.

:::tip
For LLM integrations, see [Streaming Validation](/docs/concepts/streaming) for handling streaming responses, and [LLM Streaming Examples](/docs/examples/llm-streaming) for complete integration patterns.
:::

## Simple API Functions

### Schema()

Get the JSON Schema as a `jsonschema.Schema` object.

```go
func Schema[T any]() *jsonschema.Schema
```

**Returns**: A `*jsonschema.Schema` object from the `invopop/jsonschema` package.

**Cached**: First call generates schema (~10ms), subsequent calls return cached result (sub-100ns).

**Example**:
```go
schema := validator.Schema[User]()
fmt.Println("Title:", schema.Title)
fmt.Println("Required fields:", schema.Required)
fmt.Println("Properties:", schema.Properties)
```

---

### SchemaJSON()

Get the JSON Schema as JSON bytes.

```go
func SchemaJSON[T any]() ([]byte, error)
```

**Returns**: JSON bytes, or error if serialization fails.

**Use cases**:
- Write schema to file
- Return schema in HTTP response
- Send to frontend for form generation
- Store in database

**Example**:
```go
schemaBytes, err := validator.SchemaJSON[User]()
if err != nil {
    log.Fatal(err)
}

// Write to file
os.WriteFile("user-schema.json", schemaBytes, 0644)

// Return in HTTP response
w.Header().Set("Content-Type", "application/json")
w.Write(schemaBytes)
```

---

### SchemaOpenAPI()

Get an OpenAPI 3.1 compatible component schema.

```go
func SchemaOpenAPI[T any]() *jsonschema.Schema
```

**Returns**: A JSON Schema compatible with OpenAPI 3.1 specifications.

**Important**: This generates a **component schema** (for use in `components/schemas`), not a complete OpenAPI document. Pedantigo is a validation library, not an API framework like Huma. For benchmarks, we only measure component schema generation performance.

**Features**:
- Support for nullable fields (pointer types)
- Compatible with OpenAPI 3.1 specifications (uses `$defs` syntax)
- Uses `$ref` and `$defs` for better composition

**Example**:
```go
type APIResponse struct {
    Success bool   `json:"success" validate:"required"`
    Data    *User  `json:"data"`        // nullable
    Message string `json:"message,omitempty"`
}

schema := validator.SchemaOpenAPI[APIResponse]()
// Use in OpenAPI spec
```

---

### SchemaJSONOpenAPI()

Get OpenAPI-compatible schema as JSON bytes.

```go
func SchemaJSONOpenAPI[T any]() ([]byte, error)
```

**Returns**: JSON bytes with OpenAPI enhancements.

**Example**:
```go
schemaBytes, err := validator.SchemaJSONOpenAPI[APIResponse]()
if err != nil {
    log.Fatal(err)
}

// Embed in OpenAPI YAML/JSON specification
```

---

### SchemaLLM()

Get a JSON Schema optimized for LLM APIs (no `$schema` or `$id` fields).

```go
func SchemaLLM[T any]() *jsonschema.Schema
```

**Returns**: A `*jsonschema.Schema` object with both `$schema` and `$id` cleared.

**Why strip these fields?**: Some LLMs (like Groq) echo back schema metadata fields (`$schema`, `$id`) verbatim in their JSON responses, breaking downstream parsing. Stripping both fields at the source prevents this class of failure.

**Use cases**:
- OpenAI function calling
- Anthropic tool use
- Claude structured outputs
- Gemini structured generation
- Any LLM that echoes schema metadata fields

**Example**:
```go
type ToolArgs struct {
    Query  string `json:"query" validate:"required,min=1,description=Search query"`
    Limit  int    `json:"limit" validate:"min=1,max=100,description=Max results"`
}

// Get schema for LLM function calling
schema := validator.SchemaLLM[ToolArgs]()
// schema.Version is "" → no $schema in output
// schema.ID is ""      → no $id in output
```

---

### SchemaJSONLLM()

Get JSON Schema as JSON bytes optimized for LLM APIs.

```go
func SchemaJSONLLM[T any]() ([]byte, error)
```

**Returns**: JSON bytes with both `$schema` and `$id` fields absent.

**Example**:
```go
type FunctionResponse struct {
    Thought string `json:"thought" validate:"required,description=Chain of thought"`
    Action  string `json:"action" validate:"required,oneof=search respond,description=Action to take"`
}

schemaBytes, err := validator.SchemaJSONLLM[FunctionResponse]()
if err != nil {
    log.Fatal(err)
}

// Use in OpenAI function calling API
// JSON output will NOT contain "$schema" or "$id" fields
fmt.Println(string(schemaBytes))
```

**Output**:
```json
{
  "type": "object",
  "properties": {
    "thought": {
      "type": "string",
      "description": "Chain of thought"
    },
    "action": {
      "type": "string",
      "description": "Action to take",
      "enum": ["search", "respond"]
    }
  },
  "required": ["thought", "action"]
}
```

Note: Neither `"$schema"` nor `"$id"` appear in the output.

---

## Performance & Caching

Pedantigo schemas are **automatically cached** with a 240x speedup:

| Operation | Time | Details |
|-----------|------|---------|
| First call (any type) | ~10ms | Includes reflection and generation |
| Cached calls | ~100ns | 240x faster |
| Memory overhead | Minimal | Per-type cache with type hash validation |

**How it works**:

1. **First call**: Type is reflected, constraints parsed, schema generated
2. **Cached calls**: Previous result returned instantly
3. **Thread-safe**: `sync.Map` ensures only one schema per type
4. **Invalidation**: Type hash detects if struct changes between calls

**Benchmark example**:

```go
// First call: ~10ms
schema1 := validator.Schema[User]()

// Subsequent calls: <100ns (nearly free)
for i := 0; i < 1000000; i++ {
    schema := validator.Schema[User]()
}

// Different types get separate caches
schemaUser := validator.Schema[User]()      // ~10ms
schemaProduct := validator.Schema[Product]() // ~10ms
schemaOrder := validator.Schema[Order]()    // ~10ms
```

## Constraint Mapping

Pedantigo validation constraints are automatically mapped to JSON Schema keywords. This means your validation rules are instantly documented in the generated schema.

### Core Constraints

| Constraint | JSON Schema | Description | Example |
|-----------|-----------|-------------|---------|
| `required` | `required: [...]` | Field is required | `validate:"required"` |
| `min` | `minimum` / `minLength` | Numeric min or string length | `validate:"min=18"` |
| `max` | `maximum` / `maxLength` | Numeric max or string length | `validate:"max=100"` |
| `gt` | `exclusiveMinimum` | Greater than (numeric) | `validate:"gt=0"` |
| `gte` | `minimum` | Greater than or equal | `validate:"gte=1"` |
| `lt` | `exclusiveMaximum` | Less than (numeric) | `validate:"lt=100"` |
| `lte` | `maximum` | Less than or equal | `validate:"lte=99"` |
| `len` | `minLength=X, maxLength=X` | Exact string/array length | `validate:"len=32"` |
| `oneof` / `enum` | `enum: [...]` | Must be one of values | `validate:"oneof=draft published archived"` |

### String Constraints

| Constraint | JSON Schema | Example |
|-----------|-----------|---------|
| `minLength` | `minLength` | `validate:"minLength=3"` |
| `maxLength` | `maxLength` | `validate:"maxLength=255"` |
| `pattern` | `pattern` | `validate:"pattern=^[a-z]+$"` |
| `email` | `format: "email"` | `validate:"email"` |
| `url` | `format: "uri"` | `validate:"url"` |
| `uuid` | `format: "uuid"` | `validate:"uuid"` |
| `lowercase` | `pattern` (enforced) | `validate:"lowercase"` |
| `uppercase` | `pattern` (enforced) | `validate:"uppercase"` |

### Numeric Constraints

| Constraint | JSON Schema | Example |
|-----------|-----------|---------|
| `positive` | `exclusiveMinimum: 0` | `validate:"positive"` |
| `negative` | `exclusiveMaximum: 0` | `validate:"negative"` |
| `multiple_of` | `multipleOf` | `validate:"multiple_of=5"` |
| `decimal_places` | Custom format | `validate:"decimal_places=2"` |

### Format Constraints

All standard formats map to `format` keyword:

| Constraint | Format Value | Example |
|-----------|-------------|---------|
| `email` | `"email"` | `validate:"email"` |
| `ipv4` | `"ipv4"` | `validate:"ipv4"` |
| `ipv6` | `"ipv6"` | `validate:"ipv6"` |
| `hostname` | `"hostname"` | `validate:"hostname"` |
| `fqdn` | `"fqdn"` | `validate:"fqdn"` |
| `port` | `"port"` | `validate:"port"` |
| `jwt` | `"jwt"` | `validate:"jwt"` |
| `semver` | `"semver"` | `validate:"semver"` |

### Collection Constraints

| Constraint | JSON Schema | Example |
|-----------|-----------|---------|
| `minItems` | `minItems` | `validate:"minItems=1"` |
| `maxItems` | `maxItems` | `validate:"maxItems=100"` |
| `unique` | `uniqueItems: true` | `validate:"unique"` |

## Schema Metadata Tags

Control schema generation with metadata tags in the `validate` struct tag:

### Title

Set the schema title for a field:

```go
type User struct {
    Name string `json:"name" validate:"required,title=Full Name"`
}
```

Generated schema:
```json
{
  "type": "object",
  "properties": {
    "name": {
      "title": "Full Name",
      "type": "string"
    }
  }
}
```

### Description

Add field description (appears in API docs, forms, etc.):

```go
type User struct {
    Email string `json:"email" validate:"required,email,description=User's primary email address"`
    Age   int    `json:"age" validate:"min=0,max=150,description=Age in years (0-150)"`
}
```

Generated schema:
```json
{
  "properties": {
    "email": {
      "type": "string",
      "description": "User's primary email address",
      "format": "email"
    },
    "age": {
      "type": "integer",
      "description": "Age in years (0-150)",
      "minimum": 0,
      "maximum": 150
    }
  }
}
```

### Examples

Provide example values in the schema using a JSON array literal:

```go
type Product struct {
    Name     string  `json:"name" validate:"required,examples=[\"Laptop\",\"Monitor\",\"Keyboard\"]"`
    Price    float64 `json:"price" validate:"gt=0,examples=[99.99,299.99,1999.99]"`
    Discount float64 `json:"discount" validate:"min=0,max=100,examples=[10,25,50]"`
}
```

**Syntax**: `examples=<JSON array>` — the value must be a valid JSON array literal. The array is parsed as-is, so element types follow JSON rules: strings need quotes, numbers do not.

```go
// String examples
Name string `validate:"examples=[\"Alice\",\"Bob\"]"`

// Integer examples
Age int `validate:"examples=[18,25,42]"`

// Float examples
Score float64 `validate:"examples=[0.5,0.75,1.0]"`

// Nested array examples (e.g. for [][]int fields)
Pairs [][]int `validate:"examples=[[0,1],[2,3],[4,5]]"`
```

:::note Breaking change
The previous pipe-separated syntax (`examples=Alice|Bob`) is no longer supported. All `examples=`
values must be JSON array literals.
:::

Generated schema:
```json
{
  "properties": {
    "name": {
      "type": "string",
      "examples": ["Laptop", "Monitor", "Keyboard"]
    },
    "price": {
      "type": "number",
      "examples": [99.99, 299.99, 1999.99],
      "exclusiveMinimum": 0
    }
  }
}
```

### Deprecated

Mark fields as deprecated:

```go
type User struct {
    Name     string `json:"name" validate:"required"`
    OldEmail string `json:"old_email" validate:"email,deprecated"`
    // or with message:
    LegacyID int `json:"legacy_id" validate:"deprecated=Use 'id' field instead"`
}
```

Generated schema:
```json
{
  "properties": {
    "old_email": {
      "type": "string",
      "deprecated": true,
      "description": "This field is deprecated"
    },
    "legacy_id": {
      "type": "integer",
      "deprecated": true,
      "description": "Use 'id' field instead"
    }
  }
}
```

## Complete Schema Example

Here's a realistic example showing constraint mapping and metadata:

```go
package main

import (
    "fmt"
    "github.com/SmrutAI/pedantigo/v2/validator"
)

type CreateUserRequest struct {
    // Basic fields with constraints
    Name string `json:"name" validate:"required,min=2,max=100,title=Full Name,description=User's full name"`

    // Email with format constraint
    Email string `json:"email" validate:"required,email,description=Primary email address"`

    // Numeric with range
    Age int `json:"age" validate:"min=18,max=120,description=Age in years"`

    // Enum field
    Status string `json:"status" validate:"required,oneof=active inactive suspended,description=Account status"`

    // Optional field with description
    Bio string `json:"bio,omitempty" validate:"maxLength=500,description=User biography"`

    // Tags/roles array
    Tags []string `json:"tags,omitempty" validate:"maxItems=10,unique,description=User interests and skills"`

    // URL field
    Website string `json:"website,omitempty" validate:"url,description=User's personal website"`

    // Deprecated field
    OldUsername string `json:"old_username,omitempty" validate:"deprecated=Use 'name' instead"`
}

func main() {
    schema := validator.Schema[CreateUserRequest]()

    // Use the schema for API documentation
    fmt.Printf("Title: %s\n", schema.Title)
    fmt.Printf("Required: %v\n", schema.Required)

    // Get as JSON
    schemaJSON, _ := validator.SchemaJSON[CreateUserRequest]()
    fmt.Println(string(schemaJSON))
}
```

Generated JSON:
```json
{
  "type": "object",
  "title": "CreateUserRequest",
  "properties": {
    "name": {
      "type": "string",
      "title": "Full Name",
      "description": "User's full name",
      "minLength": 2,
      "maxLength": 100
    },
    "email": {
      "type": "string",
      "description": "Primary email address",
      "format": "email"
    },
    "age": {
      "type": "integer",
      "description": "Age in years",
      "minimum": 18,
      "maximum": 120
    },
    "status": {
      "type": "string",
      "description": "Account status",
      "enum": ["active", "inactive", "suspended"]
    },
    "bio": {
      "type": "string",
      "description": "User biography",
      "maxLength": 500
    },
    "tags": {
      "type": "array",
      "description": "User interests and skills",
      "items": { "type": "string" },
      "maxItems": 10,
      "uniqueItems": true
    },
    "website": {
      "type": "string",
      "description": "User's personal website",
      "format": "uri"
    },
    "old_username": {
      "type": "string",
      "deprecated": true,
      "description": "Use 'name' instead"
    }
  },
  "required": ["name", "email", "age", "status"]
}
```

## Nested Structs

Nested structs are automatically inlined in the default schema mode:

```go
type Address struct {
    Street string `json:"street" validate:"required"`
    City   string `json:"city" validate:"required"`
    Zip    string `json:"zip" validate:"required,len=5"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address" validate:"required"`
}

schema := validator.Schema[User]()
```

Generated schema (inlined):
```json
{
  "type": "object",
  "properties": {
    "name": { "type": "string" },
    "address": {
      "type": "object",
      "properties": {
        "street": { "type": "string" },
        "city": { "type": "string" },
        "zip": {
          "type": "string",
          "minLength": 5,
          "maxLength": 5
        }
      },
      "required": ["street", "city", "zip"]
    }
  },
  "required": ["name", "address"]
}
```

For OpenAPI-style composition using `$ref` and `$defs`, use `SchemaOpenAPI()` instead.

## Use Cases

### API Documentation

Generate OpenAPI/Swagger specs automatically:

```go
type CreatePostRequest struct {
    Title   string `json:"title" validate:"required,min=5,max=200,description=Post title"`
    Content string `json:"content" validate:"required,min=10,max=10000,description=Post content"`
    Tags    []string `json:"tags" validate:"maxItems=10,unique,description=Post tags"`
}

schema := validator.SchemaJSONOpenAPI[CreatePostRequest]()
// Include in OpenAPI spec components/schemas
```

### Frontend Form Generation

Send schema to frontend for dynamic form building:

```go
// HTTP handler
func handleGetSchema(w http.ResponseWriter, r *http.Request) {
    schemaBytes, _ := validator.SchemaJSON[CreatePostRequest]()

    w.Header().Set("Content-Type", "application/json")
    w.Write(schemaBytes)
}

// Frontend receives schema and generates form fields dynamically
```

### LLM Structured Output

Provide schema to LLMs for structured generation:

```go
type TranslationResult struct {
    Original string `json:"original" validate:"required,description=Original text"`
    Translated string `json:"translated" validate:"required,description=Translated text"`
    Language string `json:"language" validate:"required,oneof=en es fr de,description=Target language code"`
    Confidence float64 `json:"confidence" validate:"min=0,max=1,description=Translation confidence 0-1"`
}

schema := validator.SchemaJSON[TranslationResult]()
// Pass to Claude, GPT, etc. for structured generation
```

### Database Schema Generation

Use generated schema to validate before persisting:

```go
// Unmarshal validates against constraints
user, err := validator.Unmarshal[User](jsonData)
if err != nil {
    return err // All validation done before database
}

// user is guaranteed valid
db.Insert(user)
```

### Runtime Type Inspection

Query the schema at runtime:

```go
schema := validator.Schema[User]()

// Check which fields are required
requiredFields := schema.Required
fmt.Printf("Required fields: %v\n", requiredFields)

// Get property details
if emailProp, ok := schema.Properties["email"]; ok {
    fmt.Printf("Email format: %s\n", emailProp.Format)
}
```

## Performance Characteristics

Schema generation is optimized for speed:

| Operation | Time | Notes |
|-----------|------|-------|
| First schema call | ~10ms | Reflection + generation + caching |
| Cached schema call | ~100ns | 240x faster |
| Serializing to JSON | ~1ms | Efficient JSON marshaling |
| Retrieving from cache | ~100ns | `sync.Map` lookup |

The 240x speedup means you can safely call `Schema[T]()` in hot paths after the first call.

## Comparison with Validation

**Schema** describes what is valid (documentation).
**Validation** checks if data is valid (enforcement).

Both use the same constraints, so they're always in sync:

```go
// These use the same constraint definitions:
schema := validator.Schema[User]()           // For documentation
user, err := validator.Unmarshal[User](data) // For validation
```

No manual schema maintenance needed - change a constraint, both schema and validation update automatically.

## Limitations & Notes

- **Embedded types**: Not supported in constraints (design choice for simplicity)
- **Custom types**: Custom constraint types are included in schema but may need custom validator implementation
- **Nullable fields**: Use `*T` (pointer) to make a field optional
- **Schema references**: `$ref` only used in OpenAPI mode (`SchemaOpenAPI`), default mode inlines

## Next Steps

- See [Validation Basics](/docs/concepts/validation) for validation constraints
- Learn [Constraints Reference](/docs/concepts/constraints) for complete constraint list
- Check [Examples](/docs/examples/basic) for real-world usage patterns
