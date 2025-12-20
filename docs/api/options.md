---
sidebar_position: 4
---

# Configuration Options

Customize Pedantigo's validation behavior with options.

## Overview

Pedantigo offers configuration options to customize how validators behave during JSON unmarshaling. Options are passed when creating explicit validators:

```go
// Simple API: automatic defaults
user, err := pedantigo.Unmarshal[User](jsonData)

// Validator API: customize behavior
validator := pedantigo.New[User](pedantigo.ValidatorOptions{
    StrictMissingFields: true,
    ExtraFields:         pedantigo.ExtraForbid,
})
user, err := validator.Unmarshal(jsonData)
```

## ValidatorOptions Type

Configuration is provided via the `ValidatorOptions` struct:

```go
type ValidatorOptions struct {
    // StrictMissingFields controls behavior for missing fields
    // Default: true (missing fields without defaults cause errors)
    StrictMissingFields bool

    // ExtraFields controls how unknown JSON fields are handled
    // Default: ExtraIgnore (unknown fields are silently ignored)
    ExtraFields ExtraFieldsMode
}
```

### Default Options

The default options are optimized for safety and strictness:

```go
pedantigo.DefaultValidatorOptions()
// Returns: ValidatorOptions{
//     StrictMissingFields: true,
//     ExtraFields:         ExtraIgnore,
// }
```

## StrictMissingFields Option

Controls whether missing fields (fields not present in JSON) cause validation errors.

### StrictMissingFields: true (Default)

Missing fields without defaults are validation errors:

```go
type Config struct {
    Host string `json:"host" pedantigo:"required"`
    Port int    `json:"port"`  // No default
}

jsonData := []byte(`{"host":"localhost"}`)
validator := pedantigo.New[Config](pedantigo.ValidatorOptions{
    StrictMissingFields: true,
})

config, err := validator.Unmarshal(jsonData)
// Error: port field is missing and has no default
```

Use this mode when you want to:
- Catch missing fields as validation errors
- Ensure all fields are explicitly provided in JSON
- Enforce strict API contracts

### StrictMissingFields: false

Missing fields without defaults are left as zero values:

```go
type Config struct {
    Host string `json:"host" pedantigo:"required"`
    Port int    `json:"port"`  // Zero value: 0
}

jsonData := []byte(`{"host":"localhost"}`)
validator := pedantigo.New[Config](pedantigo.ValidatorOptions{
    StrictMissingFields: false,
})

config, err := validator.Unmarshal(jsonData)
// Success: Port is set to 0 (zero value)
fmt.Println(config.Port) // Output: 0
```

Use this mode when you want to:
- Allow partial JSON input (missing fields get zero values)
- Make most fields optional
- Support backwards-compatible API evolution

## How `required` and `omitempty` Interact

Understanding how Go's `json:",omitempty"` interacts with Pedantigo's `required` tag:

| Struct Tags | In JSON Schema `required` | Unmarshal Behavior |
|-------------|---------------------------|-------------------|
| `pedantigo:"required"` | Yes | Error if field missing from JSON |
| `json:",omitempty"` only | No | Zero value if missing (valid) |
| Both `required` + `omitempty` | Yes | Error if missing (required wins) |
| Pointer `*T` + `required` | Yes | Error if missing or null |
| Pointer `*T` without `required` | No | nil if missing (valid) |
| Non-pointer without `required` | No | Zero value if missing |

### Key Points

1. **`omitempty` is for output** - It controls JSON marshaling, not schema generation
2. **`required` controls both** - Affects schema `required` array AND validation
3. **Pointers enable true optionality** - Use `*T` when you need to distinguish "missing" from "zero"

### Example

```go
type Config struct {
    // Required: must be in JSON, appears in schema required array
    Host string `json:"host" pedantigo:"required"`

    // Optional with omitempty: omitted from output if zero, NOT in required array
    Port int `json:"port,omitempty"`

    // Optional pointer: nil if missing, distinguishes missing from zero
    Timeout *int `json:"timeout,omitempty"`

    // Required + omitempty: must be in JSON input, omitted from output if zero
    Name string `json:"name,omitempty" pedantigo:"required"`
}
```

:::tip
See [Schema Generation](/docs/concepts/schema) to understand how these tags affect the generated JSON Schema.
:::

## ExtraFields Option

Controls how unknown (extra) JSON fields are handled during unmarshaling.

### ExtraIgnore (Default)

Unknown JSON fields are silently ignored:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

jsonData := []byte(`{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 30,
    "phone": "555-1234"
}`)

validator := pedantigo.New[User](pedantigo.ValidatorOptions{
    ExtraFields: pedantigo.ExtraIgnore,
})

user, err := validator.Unmarshal(jsonData)
// Success: age and phone fields are ignored
```

Use this mode when you want to:
- Accept JSON with additional fields (future-proofing)
- Ignore client-provided metadata
- Support flexible API clients

### ExtraForbid

Unknown JSON fields cause validation errors:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

jsonData := []byte(`{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 30
}`)

validator := pedantigo.New[User](pedantigo.ValidatorOptions{
    ExtraFields: pedantigo.ExtraForbid,
})

user, err := validator.Unmarshal(jsonData)
// Error: unknown field "age" in JSON
var validationErr *pedantigo.ValidationError
if errors.As(err, &validationErr) {
    fmt.Println(validationErr.Errors[0].Message)
    // Output: unknown field in JSON
}
```

Use this mode when you want to:
- Prevent typos in client input
- Detect API misuse
- Enforce strict schema compliance

## Complete Examples

### API Server with Strict Validation

```go
package main

import (
    "encoding/json"
    "errors"
    "log"

    "github.com/smrutai/pedantigo"
)

type CreateUserRequest struct {
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Email    string `json:"email" pedantigo:"required,email"`
    Password string `json:"password" pedantigo:"required,min=8"`
}

// API validator: strict about extra fields, requires all fields
var apiValidator = pedantigo.New[CreateUserRequest](
    pedantigo.ValidatorOptions{
        StrictMissingFields: true,
        ExtraFields:         pedantigo.ExtraForbid,
    },
)

func handleCreateUser(jsonData []byte) (*CreateUserRequest, error) {
    req, err := apiValidator.Unmarshal(jsonData)
    if err != nil {
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            return nil, validationErr
        }
        return nil, err
    }
    return req, nil
}

func main() {
    // Valid request
    valid := []byte(`{
        "username": "alice",
        "email": "alice@example.com",
        "password": "securepass123"
    }`)

    user, err := handleCreateUser(valid)
    if err != nil {
        log.Fatalf("Validation failed: %v", err)
    }
    log.Printf("Created user: %s\n", user.Username)

    // Invalid: missing field
    invalid := []byte(`{
        "username": "bob",
        "email": "bob@example.com"
    }`)

    _, err = handleCreateUser(invalid)
    // Error: password field is required

    // Invalid: extra fields
    typo := []byte(`{
        "username": "charlie",
        "email": "charlie@example.com",
        "password": "securepass123",
        "user_id": 42
    }`)

    _, err = handleCreateUser(typo)
    // Error: unknown field "user_id" in JSON
}
```

### Configuration File Parsing (Lenient)

```go
type AppConfig struct {
    Database struct {
        Host     string `json:"host" pedantigo:"required"`
        Port     int    `json:"port" pedantigo:"default=5432"`
        Username string `json:"username" pedantigo:"required"`
        Password string `json:"password" pedantigo:"required"`
    } `json:"database"`
    Server struct {
        Addr string `json:"addr" pedantigo:"default=0.0.0.0:8080"`
    } `json:"server"`
}

// Config validator: lenient about missing fields, allows extra fields
var configValidator = pedantigo.New[AppConfig](
    pedantigo.ValidatorOptions{
        StrictMissingFields: false, // Missing fields get zero values
        ExtraFields:         pedantigo.ExtraIgnore, // Extra fields ignored
    },
)

func loadConfig(jsonData []byte) (*AppConfig, error) {
    config, err := configValidator.Unmarshal(jsonData)
    if err != nil {
        // Still check for validation errors
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            return nil, validationErr
        }
        return nil, err
    }
    return config, nil
}

func main() {
    // Minimal config: missing fields use defaults
    jsonData := []byte(`{
        "database": {
            "host": "db.example.com",
            "username": "admin",
            "password": "secret"
        }
    }`)

    config, err := loadConfig(jsonData)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // config.Server.Addr is "0.0.0.0:8080" (default)
    // config.Database.Port is 5432 (default)
    // Extra fields in config are ignored
}
```

### Migration-Friendly API

```go
type UserV2 struct {
    ID       string `json:"id" pedantigo:"required,uuid"`
    Username string `json:"username" pedantigo:"required"`
    Email    string `json:"email" pedantigo:"required,email"`
    // New field: optional during migration
    Status   string `json:"status" pedantigo:"default=active"`
}

// Migration validator: accept both old and new client requests
var migrationValidator = pedantigo.New[UserV2](
    pedantigo.ValidatorOptions{
        StrictMissingFields: false, // Old clients don't send 'status'
        ExtraFields:         pedantigo.ExtraIgnore, // Old clients send extra fields
    },
)

func handleUserUpdate(jsonData []byte) (*UserV2, error) {
    user, err := migrationValidator.Unmarshal(jsonData)
    if err != nil {
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            return nil, validationErr
        }
        return nil, err
    }
    return user, nil
}

func main() {
    // Old client request (missing 'status' field)
    oldClient := []byte(`{
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "alice",
        "email": "alice@example.com"
    }`)

    user, err := handleUserUpdate(oldClient)
    // Success: Status defaults to "active"

    // New client request (includes 'status', sends extra metadata)
    newClient := []byte(`{
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "bob",
        "email": "bob@example.com",
        "status": "inactive",
        "metadata": {"source": "mobile"}
    }`)

    user, err = handleUserUpdate(newClient)
    // Success: metadata field is ignored
}
```

## Choosing the Right Options

| Scenario | StrictMissingFields | ExtraFields | Reason |
|----------|--------------------|-----------|----|
| REST API with strict contracts | `true` | `ExtraForbid` | Prevent bugs from typos/API misuse |
| Configuration file parsing | `false` | `ExtraIgnore` | Allow partial configs, future-proof |
| During API migration | `false` | `ExtraIgnore` | Accept old and new client versions |
| Public web API | `true` | `ExtraIgnore` | Strict input, future-proof |
| Internal service API | `true` | `ExtraForbid` | Strict both ways |
| Webhook receiver | `false` | `ExtraIgnore` | Accept any fields from sender |

## Option Combinations

### Most Strict
```go
pedantigo.ValidatorOptions{
    StrictMissingFields: true,
    ExtraFields:         pedantigo.ExtraForbid,
}
// Best for: REST API validation, ensuring exact schema match
```

### Most Lenient
```go
pedantigo.ValidatorOptions{
    StrictMissingFields: false,
    ExtraFields:         pedantigo.ExtraIgnore,
}
// Best for: Configuration parsing, webhook receivers, flexible APIs
```

### Balanced (Default)
```go
pedantigo.DefaultValidatorOptions()
// StrictMissingFields: true
// ExtraFields:         ExtraIgnore
// Best for: General-purpose APIs, good balance of safety and flexibility
```

## API Comparison

### Simple API (Uses Defaults)

```go
// Always uses DefaultValidatorOptions()
user, err := pedantigo.Unmarshal[User](jsonData)

// Equivalent to:
validator := pedantigo.New[User](pedantigo.DefaultValidatorOptions())
user, err := validator.Unmarshal(jsonData)
```

### Validator API (Customizable)

```go
// Create validator with custom options
validator := pedantigo.New[User](pedantigo.ValidatorOptions{
    StrictMissingFields: true,
    ExtraFields:         pedantigo.ExtraForbid,
})

// Reuse the same validator multiple times
user1, err := validator.Unmarshal(data1)
user2, err := validator.Unmarshal(data2)

// Schema is cached, so subsequent calls are very fast
schema := validator.Schema()
```

Use the Validator API when you need:
- Custom options
- Schema caching for high-performance scenarios
- Consistent behavior across multiple unmarshal calls
- Explicit control over validator creation

:::info
See [Validator API](/docs/api/validator) for more details on the full API.
:::

:::tip
For most use cases, the Simple API with default options is sufficient. Use the Validator API only when you need custom behavior or performance optimization.
:::
