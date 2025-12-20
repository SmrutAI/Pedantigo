---
sidebar_position: 2
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Quick Start

Get started with Pedantigo in minutes. This guide shows you the essential patterns for struct validation in Go.

## Step 1: Install Pedantigo

```bash
go get github.com/smrutai/pedantigo
```

## Step 2: Define Your Struct

Add `pedantigo` struct tags to define validation rules:

```go
package main

import "github.com/smrutai/pedantigo"

type User struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Website  string `json:"website,omitempty" pedantigo:"url"`
}
```

## Two Ways to Create Validated Instances

Pedantigo offers two complementary methods for creating validated instances:

### From JSON Bytes: `Unmarshal`

Use when you have JSON data (e.g., HTTP request bodies):

```go
jsonData := []byte(`{
    "email": "user@example.com",
    "age": 25,
    "username": "johndoe"
}`)

user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    // Handle validation errors
}
```

### From Any Input: `NewModel`

Use when you have maps, structs, or want maximum flexibility:

```go
// From JSON bytes
user, err := pedantigo.NewModel[User](jsonData)

// From map (great for testing!)
user, err := pedantigo.NewModel[User](map[string]any{
    "email":    "user@example.com",
    "age":      25,
    "username": "johndoe",
})

// From existing struct (validates and returns copy)
existingUser := User{Email: "test@example.com", Age: 25, Username: "test"}
user, err := pedantigo.NewModel[User](existingUser)
```

:::info When to Use Each
| Method | Use When |
|--------|----------|
| `Unmarshal` | You have JSON bytes (HTTP request body, file data) |
| `NewModel` | You have a map, struct, or want input flexibility |
:::

## Simple API vs Validator API

Pedantigo provides two APIs: a Simple API for most use cases, and a Validator API for advanced scenarios.

<Tabs>
  <TabItem value="simple" label="Simple API (Recommended)" default>

```go
package main

import "github.com/smrutai/pedantigo"

func main() {
    // Global functions with automatic caching - no setup needed

    // Unmarshal and validate JSON
    user, err := pedantigo.Unmarshal[User](jsonData)

    // Create from any input
    user, err := pedantigo.NewModel[User](inputData)

    // Get JSON Schema
    schema := pedantigo.Schema[User]()

    // Validate existing struct
    errs := pedantigo.Validate(user)
}
```

**Recommended for 80% of use cases** - automatic schema caching, clean API, minimal setup.

  </TabItem>
  <TabItem value="validator" label="Validator API (Advanced)">

```go
package main

import "github.com/smrutai/pedantigo"

func main() {
    // Explicit validator for custom options
    validator := pedantigo.New[User](pedantigo.ValidatorOptions{
        StrictMissingFields: true,
        ExtraFields:         pedantigo.ExtraForbid,
    })

    // Use validator methods
    user, err := validator.Unmarshal(jsonData)
    schema := validator.Schema()
    errs := validator.Validate(&user)
}
```

  </TabItem>
</Tabs>

:::tip When to Use Validator API
Use the explicit Validator when you need:
- Custom options (`StrictMissingFields`, `ExtraForbid`)
- Streaming validation for large JSON
- Discriminated unions (type-based routing)
- Custom validator registration

See [Initialization & Configuration](/docs/api/initialization) for all initialization methods and options.
:::

## Error Handling

Pedantigo returns detailed validation errors with field paths:

```go
user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    if validationErr, ok := err.(*pedantigo.ValidationError); ok {
        // Get all validation errors
        for _, fieldErr := range validationErr.Errors {
            fmt.Printf("Field: %s, Error: %s\n", fieldErr.Field, fieldErr.Message)
        }
    } else {
        // Handle other errors (JSON syntax, etc.)
        fmt.Printf("Error: %v\n", err)
    }
}
```

Example output:
```
Field: email, Error: must be a valid email address
Field: age, Error: must be at least 18
Field: username, Error: field is required
```

## Common Constraints Quick Reference

| Constraint | Example | Description |
|------------|---------|-------------|
| `required` | `pedantigo:"required"` | Field must be present (not missing) |
| `email` | `pedantigo:"email"` | Must be valid email format |
| `url` | `pedantigo:"url"` | Must be valid URL |
| `min`/`max` | `pedantigo:"min=0,max=100"` | Numeric range (int, float) |
| `minLength`/`maxLength` | `pedantigo:"minLength=3,maxLength=20"` | String length |
| `minItems`/`maxItems` | `pedantigo:"minItems=1,maxItems=10"` | Array/slice size |
| `pattern` | `pedantigo:"pattern=^[a-z]+$"` | Regex pattern match |
| `oneof` | `pedantigo:"oneof=red green blue"` | Must be one of values |
| `uuid` | `pedantigo:"uuid"` | Valid UUID format |
| `alpha` | `pedantigo:"alpha"` | Only letters |
| `alphanumeric` | `pedantigo:"alphanumeric"` | Letters and numbers only |
| `numeric` | `pedantigo:"numeric"` | Only numbers (string) |

See the [Constraints Reference](/docs/concepts/constraints) for the complete list.

## Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/smrutai/pedantigo"
)

type User struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
    Username string `json:"username" pedantigo:"required,minLength=3,maxLength=20,alphanumeric"`
    Website  string `json:"website,omitempty" pedantigo:"url"`
    Role     string `json:"role" pedantigo:"required,oneof=admin user guest"`
}

func main() {
    // Example 1: Unmarshal JSON
    jsonData := []byte(`{
        "email": "john@example.com",
        "age": 25,
        "username": "johndoe",
        "website": "https://example.com",
        "role": "user"
    }`)

    user, err := pedantigo.Unmarshal[User](jsonData)
    if err != nil {
        log.Fatalf("Validation failed: %v", err)
    }
    fmt.Printf("Valid user: %+v\n", user)

    // Example 2: Create from map
    userData := map[string]any{
        "email":    "jane@example.com",
        "age":      30,
        "username": "janedoe",
        "role":     "admin",
    }

    user2, err := pedantigo.NewModel[User](userData)
    if err != nil {
        log.Fatalf("Validation failed: %v", err)
    }
    fmt.Printf("Valid user: %+v\n", user2)

    // Example 3: Get JSON Schema
    schema := pedantigo.Schema[User]()
    schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
    fmt.Printf("JSON Schema:\n%s\n", schemaJSON)
}
```

## Next Steps

Now that you know the basics, explore more features:

- **[Constraints Reference](/docs/concepts/constraints)** - Complete list of validation rules
- **[JSON Schema Generation](/docs/concepts/schema)** - Auto-generate schemas with caching
- **[Custom Validators](/docs/advanced/custom-validators)** - Write your own validation logic
- **[Cross-Field Validation](/docs/concepts/cross-field)** - Validate relationships between fields
- **[Streaming Validation](/docs/concepts/streaming)** - Handle large JSON efficiently
- **[Error Handling](/docs/api/errors)** - Advanced error patterns
