---
sidebar_position: 1
---

# Validator API

The Validator API provides advanced configuration and performance optimization for validation workflows. This is the underlying API used by the Simple API and offers fine-grained control for specialized use cases.

:::tip Simple API Recommended
For most applications, use the [Simple API](./simple-api) instead. It provides automatic caching and requires less code.
:::

## Overview

The Validator API creates reusable validator instances with custom configuration. Each validator is specific to a struct type and applies the same rules consistently across requests.

### When to Use Validator API

- Need custom validation options (StrictMissingFields, ExtraForbid)
- Reusing the same validator across many requests
- Performance-critical paths (avoid global cache lookup)
- Implementing custom validator registration
- Building discriminated unions (advanced feature)

## Creating a Validator

### Default Options

```go
import "github.com/SmrutAI/pedantigo/v2/validator"

type User struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"required,min=18"`
}

// Create validator with default options
userValidator := validator.New[User]()
```

Default options:
- `StrictMissingFields: true` - Missing required fields are errors
- `ExtraFields: ExtraIgnore` - Unknown JSON fields are silently ignored

### Custom Options

```go
import "github.com/SmrutAI/pedantigo/v2/validator"

// Create with custom options
userValidator := validator.New[User](validator.Options{
    StrictMissingFields: false,  // Allow missing fields (use pointers for optional)
    ExtraFields:         validator.ExtraForbid,  // Reject unknown fields
})
```

#### Options

```go
type Options struct {
    // StrictMissingFields controls whether missing fields without defaults cause errors
    // Default: true (missing fields are errors)
    // Set to false if using pointers for optional fields
    StrictMissingFields bool

    // ExtraFields controls how unknown JSON fields are handled during Unmarshal
    // Options:
    //   - ExtraIgnore (default): Unknown fields are silently ignored
    //   - ExtraForbid: Unknown fields cause validation errors
    //   - ExtraAllow: Unknown fields are captured into a struct field tagged `extra_fields` (a map[string]any), recursively into nested structs
    ExtraFields ExtraFieldsMode
}
```

## Validator Methods

### Unmarshal

Unmarshal JSON data and validate it in a single operation.

```go
userValidator := validator.New[User]()

// Unmarshal JSON and validate
user, err := userValidator.Unmarshal([]byte(`{"email": "test@example.com", "age": 25}`))
if err != nil {
    // Handle validation error
    fmt.Printf("Validation failed: %v\n", err)
}

// Access validated data
fmt.Printf("Email: %s, Age: %d\n", user.Email, user.Age)
```

**Behavior:**
- Parses JSON and validates fields according to struct tags
- Applies static `default=VALUE` defaults for missing fields
- Returns a `*ValidationError` (whose `.Errors` field holds the individual `FieldError`s) on validation failure, not an error slice directly
- Reuses field deserializers for efficiency

### Validate

Validate an existing struct instance.

```go
userValidator := validator.New[User]()

user := &User{
    Email: "invalid-email",
    Age: 15,
}

err := userValidator.Validate(user)
if err != nil {
    // Handle validation error
    fmt.Printf("Validation failed: %v\n", err)
}
```

**Use Case:** Post-construction validation or programmatic struct creation.

### NewModel

Create a validated instance from various input types.

```go
userValidator := validator.New[User]()

// From JSON bytes
user, err := userValidator.NewModel([]byte(`{"email": "test@example.com", "age": 25}`))

// From map (kwargs pattern)
user, err := userValidator.NewModel(map[string]any{
    "email": "test@example.com",
    "age": 25,
})

// From existing struct (validates it)
existing := User{Email: "test@example.com", Age: 25}
user, err := userValidator.NewModel(existing)

// From pointer
existing := &User{Email: "test@example.com", Age: 25}
user, err := userValidator.NewModel(existing)
```

**Accepts:**
- `[]byte` - JSON data
- `map[string]any` - Field values
- `T` - Struct value
- `*T` - Struct pointer

### Schema Methods

#### Schema

Get the JSON Schema as a Go object.

```go
userValidator := validator.New[User]()

schema := userValidator.Schema()
// schema is *jsonschema.Schema

// Access schema properties
fmt.Printf("Title: %s\n", schema.Title)
fmt.Printf("Type: %s\n", schema.Type)
```

**Caching:** the schema is generated once and cached on the validator instance — see [Performance
Optimization](../advanced/performance.md) for the measured cost.

#### SchemaJSON

Get the JSON Schema as JSON bytes.

```go
userValidator := validator.New[User]()

schemaBytes, err := userValidator.SchemaJSON()
if err != nil {
    // Handle error
}

// Use for API documentation, storage, etc.
fmt.Println(string(schemaBytes))
```

#### SchemaOpenAPI

Get OpenAPI 3.1 compatible component schema.

```go
userValidator := validator.New[User]()

schema := userValidator.SchemaOpenAPI()
// Returns component schema with $defs (OpenAPI 3.1 / JSON Schema Draft 2020-12)
```

#### SchemaJSONOpenAPI

Get OpenAPI 3.1 compatible component schema as JSON bytes.

```go
userValidator := validator.New[User]()

schemaBytes, err := userValidator.SchemaJSONOpenAPI()
if err != nil {
    // Handle error
}
// Embed in OpenAPI 3.1 spec under components/schemas
```

#### SchemaLLM

Get JSON Schema optimized for LLM APIs (no `$schema` or `$id` fields).

```go
schema := userValidator.SchemaLLM()
// Returns schema without $schema or $id fields for LLM tool calling
```

Both `$schema` and `$id` are cleared because some LLMs (like Groq) echo schema metadata fields back in their responses, causing JSON parsing failures.

#### SchemaJSONLLM

Get JSON Schema as JSON bytes for LLM APIs.

```go
userValidator := validator.New[User]()

schemaBytes, err := userValidator.SchemaJSONLLM()
if err != nil {
    // Handle error
}
// JSON has no $schema or $id fields - cleaner for LLM integration
```

### Marshal Methods

Validate and convert struct to JSON.

```go
userValidator := validator.New[User]()

user := &User{
    Email: "test@example.com",
    Age: 25,
}

// With default options
jsonData, err := userValidator.Marshal(user)
if err != nil {
    // Handle validation or marshal error
}

// With custom options (context-based field exclusion)
opts := validator.ForContext("api")  // Excludes fields marked with exclude:api
jsonData, err := validator.MarshalWithOptions(user, opts)
```

**Behavior:**
- Validates struct before marshaling
- Supports context-based field inclusion/exclusion
- Supports omitzero tag for sparse output

### Dict

Convert struct to map.

```go
user := &User{
    Email: "test@example.com",
    Age: 25,
}

dict, err := validator.Dict(user)
if err != nil {
    // Handle error
}

// Access as map
fmt.Printf("Email: %s\n", dict["email"])
fmt.Printf("Age: %d\n", dict["age"])
```

**Use Case:** API responses, dynamic field access, data transformation.

## Usage Patterns

### Reusing Validators

Create once, use many times for best performance:

```go
// At initialization
userValidator := validator.New[User]()

// In request handler
func handleUserCreation(w http.ResponseWriter, r *http.Request) {
    var data []byte
    // ... read request body ...

    user, err := userValidator.Unmarshal(data)
    if err != nil {
        // Handle error
    }
    // ... process user ...
}

// Reuse same validator for all requests
func handleUserUpdate(w http.ResponseWriter, r *http.Request) {
    var data []byte
    // ... read request body ...

    user, err := userValidator.Unmarshal(data)
    if err != nil {
        // Handle error
    }
    // ... process user ...
}
```

### Multiple Validators

Use different validators with different configurations:

```go
// Strict validation for admin operations
adminValidator := validator.New[User](validator.Options{
    StrictMissingFields: true,
    ExtraFields:         validator.ExtraForbid,
})

// Lenient validation for imports
importValidator := validator.New[User](validator.Options{
    StrictMissingFields: false,
    ExtraFields:         validator.ExtraIgnore,
})

// Use as appropriate
adminUser, err := adminValidator.Unmarshal(data)
importedUser, err := importValidator.Unmarshal(data)
```

### Schema Caching

Validators cache schemas internally per type — see [Performance Optimization](../advanced/performance.md) for
the measured cost:

```go
// First call - generates and caches the schema
schema1 := userValidator.Schema()

// Subsequent calls - returned from cache
schema2 := userValidator.Schema()

// Same cached schema is returned
fmt.Println(schema1 == schema2)  // true
```

## Performance Considerations

`New()` does real, one-time work — parses struct tags, builds field deserializers, and sets up cross-field
validation — so create a validator once at initialization and reuse it, rather than calling `New()` per request.
See [Performance Optimization](../advanced/performance.md) for the actual measured cost of `New()`, the real
(negligible) difference between the Simple API and Validator API for repeated calls, and the caching patterns
that make either fast.

## Comparison with Simple API

| Feature | Validator API | Simple API |
|---------|---------------|-----------|
| Setup | Explicit instance | Automatic caching |
| Configuration | Custom options | Default only |
| Reusability | Manual management | Automatic |
| Use Case | High-throughput, custom config | General purpose |
| Code Example | `userValidator.Unmarshal(data)` | `validator.Unmarshal[User](data)` |

### When to Switch to Simple API

If any of these apply, use the Simple API instead:
- Configuration not needed (using defaults)
- Multiple different types validated
- Simplicity over micro-optimization
- General-purpose application code

### When to Use Validator API

If any of these apply, use the Validator API:
- High-throughput request handler
- Custom validator options needed
- Validator reused across many requests
- Building a framework or library
- Discriminated unions or advanced features

## Error Handling

Both methods return errors on validation failure:

```go
userValidator := validator.New[User]()

user, err := userValidator.Unmarshal(data)
if err != nil {
    // err is a validation error
    fmt.Printf("Validation failed: %v\n", err)
}

// Check specific validation errors
var ve *validator.ValidationError
if errors.As(err, &ve) {
    for _, fieldErr := range ve.Errors {
        fmt.Printf("Field %s: %s\n", fieldErr.Field, fieldErr.Message)
    }
}
```

See [Errors](./errors.md) for detailed error handling.

## Advanced Topics

### Custom Validator Registration

Register custom validation functions per validator instance:

```go
userValidator := validator.New[User]()
// Custom validators can be registered at validator creation
// See Options for details
```

### Discriminated Unions

For complex validation scenarios with union types:

```go
// Create union validator (advanced feature)
unionValidator, err := validator.NewUnion[T](validator.UnionOptions{
    // ... variant definitions
})
```

Refer to advanced examples for union validation patterns.

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/SmrutAI/pedantigo/v2/validator"
)

type User struct {
    Email    string `validate:"required,email"`
    Age      int    `validate:"required,min=18,max=120"`
    Username string `validate:"required,min=3,max=50"`
}

func main() {
    // Create validator with custom options
    userValidator := validator.New[User](validator.Options{
        StrictMissingFields: true,
        ExtraFields:         validator.ExtraForbid,
    })

    // Example 1: Unmarshal JSON
    jsonData := []byte(`{
        "email": "alice@example.com",
        "age": 28,
        "username": "alice_wonder"
    }`)

    user, err := userValidator.Unmarshal(jsonData)
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
        return
    }
    fmt.Printf("Valid user: %+v\n", user)

    // Example 2: Validate existing struct
    invalidUser := &User{
        Email:    "not-an-email",
        Age:      17,
        Username: "ab",
    }

    err = userValidator.Validate(invalidUser)
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
    }

    // Example 3: Get schema
    schema := userValidator.Schema()
    fmt.Printf("Schema: %+v\n", schema)

    // Example 4: Marshal to JSON
    validUser := &User{
        Email:    "bob@example.com",
        Age:      30,
        Username: "bob_builder",
    }

    jsonOutput, err := userValidator.Marshal(validUser)
    if err != nil {
        fmt.Printf("Marshal error: %v\n", err)
        return
    }
    fmt.Printf("JSON output: %s\n", string(jsonOutput))
}
```

## See Also

- [Simple API](./simple-api) - Recommended for most use cases
- [Initialization & Configuration](./initialization) - All initialization methods and options
- [Errors](./errors) - Error handling and types
- [Constraints Reference](../concepts/constraints) - Tag syntax and constraints
