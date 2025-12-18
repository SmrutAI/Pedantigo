---
sidebar_position: 1
---

# Simple API

The recommended API for most Pedantigo use cases. All functions work automatically with caching - no setup required.

## Overview

The Simple API provides zero-setup validation through automatic caching. Call any function directly without creating a validator first:

```go
// No setup needed - just call the function
user, err := pedantigo.Unmarshal[User](jsonData)
schema := pedantigo.Schema[User]()
err := pedantigo.Validate(user)
```

All functions use `sync.Map` to cache validators per type:
- **First call**: Creates validator, generates schema (~10ms)
- **Subsequent calls**: Returns from cache (under 100ns, 240x faster)
- **Thread-safe**: Concurrent goroutines safely share the cache

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
    Name  string `json:"name" pedantigo:"required"`
    Email string `json:"email" pedantigo:"email"`
    Age   int    `json:"age" pedantigo:"min=18"`
}

data := []byte(`{"name":"Alice","email":"alice@example.com","age":25}`)
user, err := pedantigo.Unmarshal[User](data)
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
    Host string `pedantigo:"required"`
    Port int    `pedantigo:"min=1,max=65535"`
}

config := &Config{
    Host: "localhost",
    Port: 8080,
}

if err := pedantigo.Validate(config); err != nil {
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
- `string` - JSON string
- `map[string]any` - Key-value map (kwargs pattern)
- `T` - Struct value (validates it)
- `*T` - Struct pointer (validates it)

**Returns**: Pointer to created struct, or error.

**Example - From JSON**:
```go
type User struct {
    Email string `json:"email" pedantigo:"email"`
    Age   int    `json:"age" pedantigo:"min=18"`
}

user, err := pedantigo.NewModel[User]([]byte(`{"email":"bob@example.com","age":30}`))
```

**Example - From Map (kwargs)**:
```go
user, err := pedantigo.NewModel[User](map[string]any{
    "email": "charlie@example.com",
    "age":   25,
})
```

**Example - From Struct**:
```go
existing := User{Email: "david@example.com", Age: 35}
user, err := pedantigo.NewModel[User](existing) // Validates and returns pointer
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
    Name  string `json:"name" pedantigo:"required,min=1"`
    Price float64 `json:"price" pedantigo:"min=0"`
}

schema := pedantigo.Schema[Product]()

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
schemaBytes, err := pedantigo.SchemaJSON[User]()
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

Get an OpenAPI-compatible JSON Schema.

```go
func SchemaOpenAPI[T any]() *jsonschema.Schema
```

Returns a schema with OpenAPI-specific enhancements:
- Support for nullable fields
- Compatible with OpenAPI 3.0+ specifications

**Example**:
```go
type APIResponse struct {
    Success bool      `json:"success" pedantigo:"required"`
    Data    *User     `json:"data"` // nullable
    Message string    `json:"message"`
}

schema := pedantigo.SchemaOpenAPI[APIResponse]()
// Use in OpenAPI specification
```

---

### SchemaJSONOpenAPI

Get OpenAPI-compatible schema as JSON bytes.

```go
func SchemaJSONOpenAPI[T any]() ([]byte, error)
```

OpenAPI-enhanced schema serialized as JSON bytes.

**Example**:
```go
schemaBytes, err := pedantigo.SchemaJSONOpenAPI[APIResponse]()
if err != nil {
    log.Fatal(err)
}

// Embed in OpenAPI YAML/JSON spec
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

jsonData, err := pedantigo.Marshal(user)
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
opts := pedantigo.ForContext("api")      // Create context-based options
opts := pedantigo.DefaultMarshalOptions() // Create default options
```

**Example - Context-based Exclusion**:
```go
type User struct {
    Name     string `json:"name" pedantigo:"required"`
    Email    string `json:"email" pedantigo:"email"`
    Password string `json:"password" pedantigo:"exclude:api"`
}

user := &User{
    Name:     "Frank",
    Email:    "frank@example.com",
    Password: "secret123",
}

// Serialize for API response (excludes password)
opts := pedantigo.ForContext("api")
jsonData, err := pedantigo.MarshalWithOptions(user, opts)
// {"name":"Frank","email":"frank@example.com"}
// Password field is omitted
```

**Example - With OmitZero**:
```go
type Config struct {
    Host     string `json:"host" pedantigo:"required"`
    Port     int    `json:"port" pedantigo:"omitzero"`
    Timeout  int    `json:"timeout" pedantigo:"omitzero"`
}

config := &Config{
    Host: "localhost",
    Port: 0,        // Zero value (will be omitted)
    Timeout: 0,     // Zero value (will be omitted)
}

opts := MarshalOptions{OmitZero: true}
jsonData, _ := pedantigo.MarshalWithOptions(config, opts)
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

dict, err := pedantigo.Dict(user)
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

All Simple API functions benefit from automatic caching:

| Operation | Time | Notes |
|-----------|------|-------|
| First call (validator creation) | ~10ms | Includes tag parsing and reflection |
| Schema generation (first call) | ~10ms | Reflection-based generation |
| Schema generation (cached) | ~100ns | 240x faster with cache |
| Validation (cached) | ~500ns | Per-struct validation |
| Unmarshal (cached) | ~2-5µs | JSON parsing + validation |

**Example benchmark**:
```go
// First call: ~10-15ms
schema1 := pedantigo.Schema[User]()

// Subsequent calls: <100ns
for i := 0; i < 1000000; i++ {
    schema := pedantigo.Schema[User]() // Nearly free
}
```

---

## Struct Tags

Use the `pedantigo` struct tag (not `validate`) for Simple API:

```go
type User struct {
    // Field name for JSON serialization
    Name string `json:"name" pedantigo:"required,min=1,max=100"`

    // Email validation
    Email string `json:"email" pedantigo:"required,email"`

    // Numeric constraints
    Age int `json:"age" pedantigo:"min=0,max=150"`

    // Pattern matching
    Phone string `json:"phone" pedantigo:"pattern=^\\d{10}$"`

    // Field exclusion by context
    Password string `json:"password" pedantigo:"exclude:api,exclude:logs"`

    // Zero value omission
    Score int `json:"score" pedantigo:"omitzero"`

    // Default values
    Status string `json:"status" pedantigo:"default=pending"`
}
```

Common constraints:
- `required` - Field must be present (for Unmarshal)
- `email` - Valid email format
- `min=N` - Minimum numeric value or string length
- `max=N` - Maximum numeric value or string length
- `pattern=REGEX` - Regex pattern match
- `exclude:CONTEXT` - Omit in MarshalWithOptions for that context
- `omitzero` - Omit zero values in Marshal
- `default=VALUE` - Default value if missing

See [Validation Constraints](../concepts/constraints) for complete list.

---

## When to Use Validator API Instead

The Simple API covers 80% of use cases. Use the [Validator API](./validator.md) for advanced scenarios:

- **Discriminated unions**: Use `pedantigo.NewUnion[T](opts...)`
- **Custom field deserialization**: Implement `CustomDeserializer` interface
- **Cross-field validation**: Implement `Validate()` method with custom logic
- **Reusable validators**: Create once, use many times (tiny optimization)
- **Dynamic schema manipulation**: Modify generated schema before use

```go
// Simple API (80% of cases)
user, err := pedantigo.Unmarshal[User](data)

// Validator API (advanced cases)
validator := pedantigo.New[User]()
user, err := validator.Unmarshal(data)
schema := validator.Schema()
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
user, err := pedantigo.Unmarshal[User](data)
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
        user, err := pedantigo.Unmarshal[User](data)
        // ...
    }()
}
wg.Wait()
```

The cache uses `sync.Map` to ensure only one validator is created per type, even with concurrent access.
