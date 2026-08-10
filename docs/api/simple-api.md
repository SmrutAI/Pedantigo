---
sidebar_position: 1
---

# Simple API

The recommended API for most Pedantigo use cases. All functions work automatically with caching - no setup required.

## Overview

The Simple API provides zero-setup validation through automatic caching. Call any function directly without creating a validator first:

```go
// No setup needed - just call the function
user, err := validator.Unmarshal[User](jsonData)
schema := validator.Schema[User]()
err = validator.Validate(&user)
```

All functions use `sync.Map` to cache validators per type — see [Performance
Optimization](../advanced/performance.md) for the measured cost of the first call vs. cached calls. Concurrent
goroutines safely share the cache.

See [Initialization & Configuration](./initialization) for comparison with other APIs and custom options.

## Functions

### Unmarshal

Parse JSON data into a validated struct.

```go
func Unmarshal[T any](data []byte) (*T, error)
```

Unmarshals JSON bytes into a struct of type T with automatic validation.

**Returns**: Pointer to unmarshaled struct, or error with validation details.

**Example**:
```go
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
    Age   int    `json:"age" validate:"min=18"`
}

data := []byte(`{"name":"Alice","email":"alice@example.com","age":25}`)
user, err := validator.Unmarshal[User](data)
if err != nil {
    // Handle validation errors
    log.Printf("Validation failed: %v", err)
    return
}

fmt.Printf("User: %s (%s), age %d\n", user.Name, user.Email, user.Age)
```

**Validates**:
- Missing required fields
- Invalid JSON format
- Constraint violations (email format, min/max values, patterns, etc.)
- Type mismatches

---

### UnmarshalInto

`UnmarshalInto(data []byte, target any) error` is a non-generic variant of `Unmarshal` for framework integrations that only have a `reflect.Type`/`any` at the call site (e.g. a web framework's binder). It looks up the validator registered for `target`'s concrete type and validates against it.

```go
var _ = validator.Register(validator.New[User]())

var u User
err := validator.UnmarshalInto(data, &u)
```

The target type must have been registered first via `validator.Register(validator.New[T]())` — `UnmarshalInto` panics with an actionable message naming the missing type if it wasn't. This is the mechanism the Echo Binder plugin (`docs/plugins/web/echo.md`) and the Gin request plugin's JSON path (`docs/plugins/web/gin.md`) use internally.

---

### ValidateInto

`ValidateInto(target any) error` is the runtime-typed companion to `Validate`.
It exists for framework integrations that only know the destination object at
runtime after the framework has already populated it.

```go
var _ = validator.Register(validator.New[QueryRequest](validator.Options{TagName: "binding"}))

var req QueryRequest
err := validator.ValidateInto(&req)
```

Behavior:

- looks up the registered validator for the concrete element type behind `target`
- returns `nil` for `nil`, non-pointer values, and unregistered types
- returns `*ValidationError` for ordinary validation failures

This is the mechanism the Gin request plugin (`docs/plugins/web/gin.md`) uses
for query/form/header/URI validation after Gin's own binders populate the
struct.

When using framework plugins or other runtime type lookups, all registered validators in the same process must share one effective tag name. For example, registering one request type with `TagName: "validate"` and another with `TagName: "binding"` will panic once both are made visible through `Register()`.

---

### Validate

Validate an existing struct instance.

```go
func Validate[T any](obj *T) error
```

Validates a struct that already exists in memory (not from JSON).

**Note**: Checks value constraints only, not field presence (since the struct is already created). Use `Unmarshal` if you need to detect missing fields.

**Returns**: Error with validation details, or nil if valid.

**Example**:
```go
type Config struct {
    Host string `validate:"required"`
    Port int    `validate:"min=1,max=65535"`
}

config := &Config{
    Host: "localhost",
    Port: 8080,
}

if err := validator.Validate(config); err != nil {
    log.Printf("Config invalid: %v", err)
    return
}
```

---

### NewModel

Create a validated instance from any input type.

```go
func NewModel[T any](input any) (*T, error)
```

Accepts multiple input formats and creates a validated struct:
- `[]byte` - JSON data
- `map[string]any` - Key-value map (kwargs pattern)
- `T` - Struct value (validates it)
- `*T` - Struct pointer (validates it)

**Returns**: Pointer to created struct, or error.

**Example - From JSON**:
```go
type User struct {
    Email string `json:"email" validate:"email"`
    Age   int    `json:"age" validate:"min=18"`
}

user, err := validator.NewModel[User]([]byte(`{"email":"bob@example.com","age":30}`))
```

**Example - From Map (kwargs)**:
```go
user, err := validator.NewModel[User](map[string]any{
    "email": "charlie@example.com",
    "age":   25,
})
```

**Example - From Struct**:
```go
existing := User{Email: "david@example.com", Age: 35}
user, err := validator.NewModel[User](existing) // Validates and returns pointer
```

---

### Schema

Get the JSON Schema for type T.

```go
func Schema[T any]() *jsonschema.Schema
```

Returns the JSON Schema as a `*jsonschema.Schema` object (from `invopop/jsonschema`).

**Caching**: Schema is generated once and cached within the validator. Subsequent calls are extremely fast.

**Example**:
```go
type Product struct {
    Name  string `json:"name" validate:"required,min=1"`
    Price float64 `json:"price" validate:"min=0"`
}

schema := validator.Schema[Product]()

// Use the schema object
fmt.Println("Product schema title:", schema.Title)
fmt.Println("Required fields:", schema.Required)
```

---

### SchemaJSON

Get the JSON Schema as JSON bytes.

```go
func SchemaJSON[T any]() ([]byte, error)
```

Returns the JSON Schema serialized as JSON bytes.

**Example**:
```go
schemaBytes, err := validator.SchemaJSON[User]()
if err != nil {
    log.Fatal(err)
}

// Write to file or send in HTTP response
os.WriteFile("user-schema.json", schemaBytes, 0644)
```

**Use cases**:
- Store schema in files
- Expose schema via HTTP API
- Send schema to frontend for form validation
- Integrate with OpenAPI generators

---

### SchemaOpenAPI

Get an OpenAPI 3.1 compatible component schema.

```go
func SchemaOpenAPI[T any]() *jsonschema.Schema
```

Returns a JSON Schema compatible with OpenAPI 3.1 specifications:
- Support for nullable fields
- Uses `$defs` syntax (OpenAPI 3.1 / JSON Schema Draft 2020-12)

**Note**: This generates a component schema (for `components/schemas`), not a complete OpenAPI document. Pedantigo is a validation library, not an API framework.

**Example**:
```go
type APIResponse struct {
    Success bool      `json:"success" validate:"required"`
    Data    *User     `json:"data"` // nullable
    Message string    `json:"message"`
}

schema := validator.SchemaOpenAPI[APIResponse]()
// Embed in OpenAPI 3.1 specification's components/schemas
```

---

### SchemaJSONOpenAPI

Get OpenAPI 3.1 compatible component schema as JSON bytes.

```go
func SchemaJSONOpenAPI[T any]() ([]byte, error)
```

Component schema serialized as JSON bytes, ready to embed in OpenAPI 3.1 specifications.

**Example**:
```go
schemaBytes, err := validator.SchemaJSONOpenAPI[APIResponse]()
if err != nil {
    log.Fatal(err)
}

// Embed in OpenAPI 3.1 YAML/JSON spec under components/schemas
```

---

### SchemaLLM

Get JSON Schema optimized for LLM APIs (no `$schema` or `$id` fields).

```go
func SchemaLLM[T any]() *jsonschema.Schema
```

Returns schema with both `$schema` and `$id` cleared. Some LLMs (like Groq) echo back schema metadata fields in responses, causing JSON parsing failures.

**Use cases**:
- OpenAI function calling
- Anthropic tool use (Claude)
- Gemini structured outputs
- Any LLM that echoes schema metadata fields back

**Example**:
```go
schema := validator.SchemaLLM[ToolArgs]()
// schema has no $schema or $id fields - cleaner for LLM integration
```

---

### SchemaJSONLLM

Get JSON Schema as JSON bytes for LLM APIs.

```go
func SchemaJSONLLM[T any]() ([]byte, error)
```

Returns JSON bytes with both `$schema` and `$id` fields absent.

**Example**:
```go
schemaBytes, err := validator.SchemaJSONLLM[ToolArgs]()
if err != nil {
    log.Fatal(err)
}

// Send to LLM API - no $schema or $id fields in JSON output
```

---

### Marshal

Validate and serialize a struct to JSON.

```go
func Marshal[T any](obj *T) ([]byte, error)
```

Validates the struct and marshals it to JSON using default options.

**Returns**: JSON bytes, or error if validation fails.

**Example**:
```go
user := &User{
    Name:  "Eve",
    Email: "eve@example.com",
    Age:   28,
}

jsonData, err := validator.Marshal(user)
if err != nil {
    log.Printf("Marshal failed: %v", err)
    return
}

fmt.Println(string(jsonData))
// {"name":"Eve","email":"eve@example.com","age":28}
```

---

### MarshalWithOptions

Validate and serialize with custom options.

```go
func MarshalWithOptions[T any](obj *T, opts MarshalOptions) ([]byte, error)
```

Marshals with options for:
- **Context-based field exclusion**: Omit fields based on context (e.g., hide passwords in API responses)
- **OmitZero behavior**: Control whether zero values are included

**MarshalOptions**:
```go
type MarshalOptions struct {
    // Context for field exclusion
    Context string  // e.g., "api", "response", "log"

    // Whether to honor omitzero tags
    OmitZero bool   // Default: true
}
```

**Helper functions**:
```go
opts := validator.ForContext("api")      // Create context-based options
opts := validator.DefaultMarshalOptions() // Create default options
```

**Example - Context-based Exclusion**:
```go
type User struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"email"`
    Password string `json:"password" validate:"exclude:api"`
}

user := &User{
    Name:     "Frank",
    Email:    "frank@example.com",
    Password: "secret123",
}

// Serialize for API response (excludes password)
opts := validator.ForContext("api")
jsonData, err := validator.MarshalWithOptions(user, opts)
// {"name":"Frank","email":"frank@example.com"}
// Password field is omitted
```

**Example - With OmitZero**:
```go
type Config struct {
    Host     string `json:"host" validate:"required"`
    Port     int    `json:"port" validate:"omitzero"`
    Timeout  int    `json:"timeout" validate:"omitzero"`
}

config := &Config{
    Host: "localhost",
    Port: 0,        // Zero value (will be omitted)
    Timeout: 0,     // Zero value (will be omitted)
}

opts := MarshalOptions{OmitZero: true}
jsonData, _ := validator.MarshalWithOptions(config, opts)
// {"host":"localhost"}
// port and timeout omitted because they're zero and OmitZero is true
```

---

### Dict

Convert a struct to `map[string]interface{}`.

```go
func Dict[T any](obj *T) (map[string]interface{}, error)
```

Serializes a struct into a map. Useful for dynamic access or intermediate transformations.

**Example**:
```go
user := &User{
    Name:  "Grace",
    Email: "grace@example.com",
    Age:   32,
}

dict, err := validator.Dict(user)
if err != nil {
    log.Fatal(err)
}

fmt.Println(dict["name"])  // "Grace"
fmt.Println(dict["age"])   // 32

// Use map for conditional logic
if dict["age"].(int) >= 18 {
    fmt.Println("User is an adult")
}
```

---

## Performance Characteristics

All Simple API functions benefit from automatic caching — see [Performance
Optimization](../advanced/performance.md) for verified benchmark numbers and how the caching works.

---

## Struct Tags

Use the `validate` struct tag for Simple API:

```go
type User struct {
    // Field name for JSON serialization
    Name string `json:"name" validate:"required,min=1,max=100"`

    // Email validation
    Email string `json:"email" validate:"required,email"`

    // Numeric constraints
    Age int `json:"age" validate:"min=0,max=150"`

    // Pattern matching
    Phone string `json:"phone" validate:"regexp=^\\d{10}$"`

    // Field exclusion by context
    Password string `json:"password" validate:"exclude:api|logs"`

    // Zero value omission
    Score int `json:"score" validate:"omitzero"`

    // Default values
    Status string `json:"status" validate:"default=pending"`
}
```

Common constraints:
- `required` - Field must be present (for Unmarshal)
- `email` - Valid email format
- `min=N` - Minimum numeric value or string length
- `max=N` - Maximum numeric value or string length
- `regexp=REGEX` - Regex pattern match
- `exclude:CONTEXT` - Omit in MarshalWithOptions for that context (pipe-separate multiple: `exclude:api|logs`)
- `omitzero` - Omit zero values in Marshal
- `default=VALUE` - Default value if missing

See [Validation Constraints](../concepts/constraints) for complete list.

---

## When to Use Validator API Instead

The Simple API covers 80% of use cases. Use the [Validator API](./validator.md) for advanced scenarios:

- **Discriminated unions**: Use `validator.NewUnion[T](opts...)`
- **Custom field deserialization**: Implement `CustomDeserializer` interface
- **Cross-field validation**: Implement `Validate()` method with custom logic
- **Reusable validators**: Create once, use many times (tiny optimization)
- **Dynamic schema manipulation**: Modify generated schema before use

```go
// Simple API (80% of cases)
user, err := validator.Unmarshal[User](data)

// Validator API (advanced cases)
userValidator := validator.New[User]()
user, err := userValidator.Unmarshal(data)
schema := userValidator.Schema()
// ... customize schema ...
```

---

## Error Handling

All functions return `error` on validation failure. Errors include:
- Field path (e.g., `user.email`)
- Constraint that failed (e.g., `email`)
- Actual value and expected constraint

**Example error message**:
```
validation error: field "email": constraint "email" failed for value "invalid"
validation error: field "age": constraint "min" failed (minimum: 18, actual: 10)
```

Handle errors appropriately:
```go
user, err := validator.Unmarshal[User](data)
if err != nil {
    // Log the error
    log.Errorf("Validation failed: %v", err)

    // Return to client (with HTTP 400 Bad Request)
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

---

## Thread Safety

All Simple API functions are thread-safe:

```go
// Multiple goroutines can call simultaneously
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        user, err := validator.Unmarshal[User](data)
        // ...
    }()
}
wg.Wait()
```

The cache uses `sync.Map` to ensure only one validator is created per type, even with concurrent access.
