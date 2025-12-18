---
sidebar_position: 1
slug: /
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Welcome to Pedantigo

**Type-safe JSON validation for Go, inspired by Pydantic**

Pedantigo brings Pydantic's elegant validation patterns to Go with a reflection-based design that feels natural in the Go ecosystem.

## Why Pedantigo?

Pedantigo was built while developing [smrut.ai](https://smrut.ai), where reliable JSON validation is critical for AI agent interactions.

When building Go applications that consume JSON from external sources (APIs, AI agents, user input), you need validation that's:

- **Type-safe**: Catch errors at unmarshal time, not deep in your application
- **Declarative**: Define rules alongside your struct fields with tags
- **Fast**: Cached JSON schemas deliver 240x performance gains
- **Production-ready**: Battle-tested constraints covering email, URLs, regex, ranges, and more

## Quick Example

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type User struct {
    Email string `json:"email" pedantigo:"required,email"`
    Age   int    `json:"age" pedantigo:"min=18,max=120"`
    Role  string `json:"role" pedantigo:"oneof=admin user guest"`
}

func main() {
    jsonData := []byte(`{"email": "user@example.com", "age": 25, "role": "admin"}`)

    // Parse, validate, and unmarshal in one call
    user, err := pedantigo.Unmarshal[User](jsonData)
    if err != nil {
        fmt.Println("Validation failed:", err)
        return
    }

    fmt.Printf("Welcome %s!\n", user.Email)
}
```

## Two Ways to Validate

Pedantigo offers two API styles to match your use case:

<Tabs>
<TabItem value="simple" label="Simple API (Recommended)" default>

**Best for most use cases** - Global schema cache, minimal boilerplate:

```go
// Parse JSON + validate
user, err := pedantigo.Unmarshal[User](jsonData)

// Create from JSON/map/struct with validation
user, err := pedantigo.NewModel[User](input)

// Validate existing struct
err := pedantigo.Validate[User](existingUser)

// Get cached JSON Schema
schema := pedantigo.Schema[User]()
```

The Simple API uses a global schema cache that's automatically managed. Perfect for typical validation workflows.

</TabItem>
<TabItem value="validator" label="Validator API">

**For advanced features** - Explicit validator instances with custom options:

```go
// Create validator instance
validator := pedantigo.New[User]()

// Use validator methods
user, err := validator.Unmarshal(jsonData)
schema := validator.JSONSchema()
```

Use the Validator API when you need:
- Discriminated unions (`NewUnion[T](opts...)`)
- Custom validator registration per instance
- Fine-grained control over caching
- Plugin integration (advanced)

</TabItem>
</Tabs>

## Features at a Glance

| Feature | Description |
|---------|-------------|
| **100+ Built-in Constraints** | Email, URL, regex, ranges, string length, array constraints, and more |
| **JSON Schema Generation** | Auto-generate JSON schemas from Go structs with validation metadata |
| **240x Performance** | First-call schema caching delivers sub-100ns lookups vs ~10ms generation |
| **Streaming Validation** | Validate partial JSON for progressive parsing scenarios |
| **Discriminated Unions** | Type-safe union handling with discriminator fields |
| **Cross-Field Validation** | Implement `Validate()` interface for complex multi-field rules |
| **Custom Validators** | Register your own constraints globally or per-validator |
| **Zero Dependencies** | Only uses `invopop/jsonschema` + Go standard library |

## What You'll Learn

This documentation covers everything from basic validation to advanced patterns:

- **[Getting Started](./getting-started/installation)** - Install Pedantigo and validate your first struct
- **[Core Concepts](./concepts/validation)** - Understand how validation works under the hood
- **[API Reference](./api/simple-api)** - Complete guide to Simple API and Validator API
- **[Constraints](./constraints/string)** - Browse all 100+ built-in validation rules
- **[Advanced Features](./advanced/custom-validators)** - Custom validators, unions, streaming, and more

:::tip Ready to start?

Head to [Installation](./getting-started/installation) to add Pedantigo to your project, or jump to [Quick Start](./getting-started/quickstart) for a guided tutorial.

:::
