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
import "github.com/SmrutAI/pedantigo"

type User struct {
    Email string `pedantigo:"required,email"`
    Age   int    `pedantigo:"required,min=18"`
}

// Create validator with default options
validator := pedantigo.New[User]()
```

Default options:
- `StrictMissingFields: true` - Missing required fields are errors
- `ExtraFields: ExtraIgnore` - Unknown JSON fields are silently ignored

### Custom Options

```go
import "github.com/SmrutAI/pedantigo"

// Create with custom options
validator := pedantigo.New[User](pedantigo.ValidatorOptions{
    StrictMissingFields: false,  // Allow missing fields (use pointers for optional)
    ExtraFields:         pedantigo.ExtraForbid,  // Reject unknown fields
})
```

#### ValidatorOptions

```go
type ValidatorOptions struct {
    // StrictMissingFields controls whether missing fields without defaults cause errors
    // Default: true (missing fields are errors)
    // Set to false if using pointers for optional fields
    StrictMissingFields bool

    // ExtraFields controls how unknown JSON fields are handled during Unmarshal
    // Options:
    //   - ExtraIgnore (default): Unknown fields are silently ignored
    //   - ExtraForbid: Unknown fields cause validation errors
    //   - ExtraAllow: Reserved for future use
    ExtraFields ExtraFieldsMode
}
```

## Validator Methods

### Unmarshal

Unmarshal JSON data and validate it in a single operation.

```go
// Unmarshal JSON and validate
user, err := validator.Unmarshal([]byte(`{"email": "test@example.com", "age": 25}`))
if err != nil {
    // Handle validation error
    fmt.Printf("Validation failed: %v\n", err)
}

// Access validated data
fmt.Printf("Email: %s, Age: %d\n", user.Email, user.Age)
```

**Behavior:**
- Parses JSON and validates fields according to struct tags
- Applies defaults and defaultFactory functions
- Returns error slice on validation failure
- Reuses field deserializers for efficiency

### Validate

Validate an existing struct instance.

```go
user := &User{
    Email: "invalid-email",
    Age: 15,
}

err := validator.Validate(user)
if err != nil {
    // Handle validation error
    fmt.Printf("Validation failed: %v\n", err)
}
```

**Use Case:** Post-construction validation or programmatic struct creation.

### NewModel

Create a validated instance from various input types.

```go
// From JSON bytes
user, err := validator.NewModel([]byte(`{"email": "test@example.com", "age": 25}`))

// From map (kwargs pattern)
user, err := validator.NewModel(map[string]any{
    "email": "test@example.com",
    "age": 25,
})

// From existing struct (validates it)
existing := User{Email: "test@example.com", Age: 25}
user, err := validator.NewModel(existing)

// From pointer
existing := &User{Email: "test@example.com", Age: 25}
user, err := validator.NewModel(existing)
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
schema := validator.Schema()
// schema is *jsonschema.Schema

// Access schema properties
fmt.Printf("Title: %s\n", schema.Title)
fmt.Printf("Type: %s\n", schema.Type)
```

**Performance:** First call generates schema (~10ms). Subsequent calls return from cache (under 100ns).

#### SchemaJSON

Get the JSON Schema as JSON bytes.

```go
schemaBytes, err := validator.SchemaJSON()
if err != nil {
    // Handle error
}

// Use for API documentation, storage, etc.
fmt.Println(string(schemaBytes))
```

#### SchemaOpenAPI

Get OpenAPI-compatible JSON Schema.

```go
schema := validator.SchemaOpenAPI()
// Includes OpenAPI-specific enhancements like nullable support
```

#### SchemaJSONOpenAPI

Get OpenAPI-compatible JSON Schema as JSON bytes.

```go
schemaBytes, err := validator.SchemaJSONOpenAPI()
if err != nil {
    // Handle error
}
```

### Marshal Methods

Validate and convert struct to JSON.

```go
user := &User{
    Email: "test@example.com",
    Age: 25,
}

// With default options
jsonData, err := validator.Marshal(user)
if err != nil {
    // Handle validation or marshal error
}

// With custom options (context-based field exclusion)
opts := pedantigo.ForContext("api")  // Excludes fields marked with exclude:api
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
validator := pedantigo.New[User]()

// In request handler
func handleUserCreation(w http.ResponseWriter, r *http.Request) {
    var data []byte
    // ... read request body ...

    user, err := validator.Unmarshal(data)
    if err != nil {
        // Handle error
    }
    // ... process user ...
}

// Reuse same validator for all requests
func handleUserUpdate(w http.ResponseWriter, r *http.Request) {
    var data []byte
    // ... read request body ...

    user, err := validator.Unmarshal(data)
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
adminValidator := pedantigo.New[User](pedantigo.ValidatorOptions{
    StrictMissingFields: true,
    ExtraFields:         pedantigo.ExtraForbid,
})

// Lenient validation for imports
importValidator := pedantigo.New[User](pedantigo.ValidatorOptions{
    StrictMissingFields: false,
    ExtraFields:         pedantigo.ExtraIgnore,
})

// Use as appropriate
adminUser, err := adminValidator.Unmarshal(data)
importedUser, err := importValidator.Unmarshal(data)
```

### Schema Caching

Validators cache schemas internally per type:

```go
// First call - generates schema (~10ms)
schema1 := validator.Schema()

// Subsequent calls - from cache (<100ns)
schema2 := validator.Schema()

// Same cached schema is returned
fmt.Println(schema1 == schema2)  // true
```

## Performance Considerations

### Validator Creation Cost

Creating a validator is moderately expensive (~1-2ms):
- Parses all struct tags
- Builds field deserializer closures
- Validates defaultFactory functions
- Sets up cross-field validation

**Best Practice:** Create once at initialization, reuse for multiple operations.

### Reuse vs. Cache Lookup

Comparison for high-throughput paths:

```
Validator instance reuse:       ~500ns per unmarshal
Simple API (cache lookup):     ~700ns per unmarshal (includes lock acquisition)
```

For performance-critical code, maintain a validator instance.

### Memory Considerations

Each validator maintains:
- Field deserializer closures
- Schema cache (lazy initialized)
- Cross-field constraint cache

Memory overhead is minimal (~10-50KB per validator instance).

## Comparison with Simple API

| Feature | Validator API | Simple API |
|---------|---------------|-----------|
| Setup | Explicit instance | Automatic caching |
| Configuration | Custom options | Default only |
| Performance | Fastest (no cache lookup) | Fast (global cache) |
| Reusability | Manual management | Automatic |
| Use Case | High-throughput, custom config | General purpose |
| Code Example | `validator.Unmarshal(data)` | `pedantigo.Unmarshal[User](data)` |

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
user, err := validator.Unmarshal(data)
if err != nil {
    // err is a validation error
    fmt.Printf("Validation failed: %v\n", err)
}

// Check specific validation errors
errs := user, err  // err is a FieldError slice
for _, fieldErr := range errs {
    fmt.Printf("Field %s: %s\n", fieldErr.Field, fieldErr.Message)
}
```

See [Errors](./errors.md) for detailed error handling.

## Advanced Topics

### Custom Validator Registration

Register custom validation functions per validator instance:

```go
validator := pedantigo.New[User]()
// Custom validators can be registered at validator creation
// See ValidatorOptions for details
```

### Discriminated Unions

For complex validation scenarios with union types:

```go
// Create union validator (advanced feature)
validator := pedantigo.NewUnion[T](opts...)
```

Refer to advanced examples for union validation patterns.

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/SmrutAI/pedantigo"
)

type User struct {
    Email    string `pedantigo:"required,email"`
    Age      int    `pedantigo:"required,min=18,max=120"`
    Username string `pedantigo:"required,min=3,max=50"`
}

func main() {
    // Create validator with custom options
    validator := pedantigo.New[User](pedantigo.ValidatorOptions{
        StrictMissingFields: true,
        ExtraFields:         pedantigo.ExtraForbid,
    })

    // Example 1: Unmarshal JSON
    jsonData := []byte(`{
        "email": "alice@example.com",
        "age": 28,
        "username": "alice_wonder"
    }`)

    user, err := validator.Unmarshal(jsonData)
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

    err = validator.Validate(invalidUser)
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
    }

    // Example 3: Get schema
    schema := validator.Schema()
    fmt.Printf("Schema: %+v\n", schema)

    // Example 4: Marshal to JSON
    validUser := &User{
        Email:    "bob@example.com",
        Age:      30,
        Username: "bob_builder",
    }

    jsonOutput, err := validator.Marshal(validUser)
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
