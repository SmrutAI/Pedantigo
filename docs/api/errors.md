---
sidebar_position: 3
---

# Error Handling

How to handle validation errors in Pedantigo.

## Overview

Pedantigo collects all validation errors in a single pass and returns them as a `ValidationError` type. This means you get complete feedback about all validation failures at once, rather than discovering errors one by one.

## ValidationError Type

The main error type wraps multiple field-level errors:

```go
type ValidationError struct {
    Errors []FieldError
}

// Error implements the error interface
func (e *ValidationError) Error() string
```

The `Error()` method returns a human-readable summary:
- If one error: `"field: message"`
- If multiple errors: `"field: message (and N more errors)"`

## FieldError Type

Each error in the `Errors` slice is a `FieldError`:

```go
type FieldError struct {
    Field   string // Field path (e.g., "email", "address.city", "items[0].price")
    Code    string // Machine-readable error code (e.g., "INVALID_EMAIL")
    Message string // Human-readable error message
    Value   any    // The actual value that failed validation
}
```

### Field Path Format

The `Field` string describes the location of the error:

| Example | Meaning |
|---------|---------|
| `"email"` | Top-level field named `email` |
| `"user.email"` | Nested field `email` in nested struct `user` |
| `"address.city"` | Field `city` in nested struct `address` |
| `"items[0]"` | First element of array/slice `items` |
| `"items[0].name"` | Field `name` in first element of `items` array |

## Error Handling Patterns

### Type Assertion

Check if an error is a validation error using type assertion:

```go
user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    if validationErr, ok := err.(*pedantigo.ValidationError); ok {
        // Handle validation errors
        for _, fieldErr := range validationErr.Errors {
            fmt.Printf("Field: %s, Message: %s\n", fieldErr.Field, fieldErr.Message)
        }
    } else {
        // Handle other errors (JSON syntax, etc.)
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Using errors.As (Recommended)

Use the `errors.As()` function for cleaner error handling:

```go
import (
    "errors"
    "github.com/smrutai/pedantigo"
)

user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    var validationErr *pedantigo.ValidationError
    if errors.As(err, &validationErr) {
        // Handle validation errors
        for _, fieldErr := range validationErr.Errors {
            fmt.Printf("[%s] %s\n", fieldErr.Field, fieldErr.Message)
        }
    } else {
        // Handle other error types
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

### Iterating All Errors

```go
type User struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
}

user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    var validationErr *pedantigo.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Printf("Validation failed with %d error(s):\n", len(validationErr.Errors))
        for i, fieldErr := range validationErr.Errors {
            fmt.Printf("%d. Field '%s':\n", i+1, fieldErr.Field)
            fmt.Printf("   Message: %s\n", fieldErr.Message)
            fmt.Printf("   Value: %v\n", fieldErr.Value)
            if fieldErr.Code != "" {
                fmt.Printf("   Code: %s\n", fieldErr.Code)
            }
        }
    }
}
```

### Filtering Errors by Field

```go
var validationErr *pedantigo.ValidationError
if errors.As(err, &validationErr) {
    // Get all errors for a specific field
    emailErrors := make([]pedantigo.FieldError, 0)
    for _, fieldErr := range validationErr.Errors {
        if fieldErr.Field == "email" {
            emailErrors = append(emailErrors, fieldErr)
        }
    }

    if len(emailErrors) > 0 {
        fmt.Printf("Email validation failed: %s\n", emailErrors[0].Message)
    }
}
```

## Common Error Messages

### String Constraints

| Constraint | Error Message | Example |
|------------|---------------|---------|
| `required` | `field is required` | Missing required field |
| `email` | `must be a valid email address` | Invalid email format |
| `url` | `must be a valid URL` | Invalid URL format |
| `uuid` | `must be a valid UUID` | Invalid UUID format |
| `min=N` | `must be at least N characters` | String too short |
| `max=N` | `must be at most N characters` | String too long |
| `alpha` | `must contain only alphabetic characters` | Contains non-letters |
| `alphanum` | `must contain only alphanumeric characters` | Contains special chars |
| `pattern` | `must match pattern .*` | Regex pattern mismatch |

### Numeric Constraints

| Constraint | Error Message | Example |
|------------|---------------|---------|
| `min=N` | `must be at least N` | Value too small |
| `max=N` | `must be at most N` | Value too large |
| `gt=N` | `must be greater than N` | Value not greater |
| `lt=N` | `must be less than N` | Value not less |
| `gte=N` | `must be at least N` | Value too small |
| `lte=N` | `must be at most N` | Value too large |

### Array/Slice Constraints

| Constraint | Error Message | Example |
|------------|---------------|---------|
| `minItems=N` | `must have at least N items` | Array too short |
| `maxItems=N` | `must have at most N items` | Array too long |
| `unique` | `items must be unique` | Duplicate items |

## Error Handling Example

Here's a complete example showing error handling with nested structs:

```go
package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"

    "github.com/smrutai/pedantigo"
)

type Address struct {
    Street string `json:"street" pedantigo:"required,min=5"`
    City   string `json:"city" pedantigo:"required,min=2"`
    Zip    string `json:"zip" pedantigo:"required,regexp=^\\d{5}$"`
}

type User struct {
    Email   string  `json:"email" pedantigo:"required,email"`
    Age     int     `json:"age" pedantigo:"required,min=18,max=120"`
    Address Address `json:"address" pedantigo:"required"`
}

func handleValidationError(err error) {
    var validationErr *pedantigo.ValidationError
    if !errors.As(err, &validationErr) {
        log.Fatalf("Unexpected error: %v", err)
        return
    }

    // Organize errors by field
    errorsByField := make(map[string][]pedantigo.FieldError)
    for _, fieldErr := range validationErr.Errors {
        errorsByField[fieldErr.Field] = append(errorsByField[fieldErr.Field], fieldErr)
    }

    // Display errors grouped by field
    fmt.Printf("Validation failed with %d error(s):\n\n", len(validationErr.Errors))
    for field, errors := range errorsByField {
        fmt.Printf("Field: %s\n", field)
        for i, err := range errors {
            fmt.Printf("  %d. %s\n", i+1, err.Message)
            if err.Code != "" {
                fmt.Printf("     Code: %s\n", err.Code)
            }
        }
        fmt.Println()
    }
}

func main() {
    jsonData := []byte(`{
        "email": "not-an-email",
        "age": 15,
        "address": {
            "street": "Main",
            "city": "NY",
            "zip": "not-a-zip"
        }
    }`)

    user, err := pedantigo.Unmarshal[User](jsonData)
    if err != nil {
        handleValidationError(err)
        return
    }

    // Continue with valid user
    fmt.Printf("Valid user: %+v\n", user)
}
```

## Non-Validation Errors

Pedantigo can also return non-validation errors in certain cases:

```go
// JSON syntax error
_, err := pedantigo.Unmarshal[User]([]byte(`{invalid json}`))
// err will be a json.SyntaxError (not ValidationError)

// Type mismatch (if type conversion fails)
_, err := pedantigo.Unmarshal[User]([]byte(`{"age": "not a number"}`))
// err will be a json.UnmarshalTypeError (not ValidationError)
```

Always check the actual error type before assuming it's a `ValidationError`.

## Error Behavior with Different API Methods

### Unmarshal

Returns both validation errors and JSON parsing errors:

```go
user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    // Could be ValidationError, json.SyntaxError, or json.UnmarshalTypeError
}
```

### Validate

Returns only validation errors:

```go
err := pedantigo.Validate(user)
if err != nil {
    // Always a ValidationError (if non-nil)
    var validationErr *pedantigo.ValidationError
    errors.As(err, &validationErr) // Always succeeds
}
```

### NewModel

Returns both validation errors and type conversion errors:

```go
user, err := pedantigo.NewModel[User](data)
if err != nil {
    // Could be ValidationError or type conversion error
}
```

## Best Practices

1. **Always check error type**: Use `errors.As()` to safely check for `ValidationError`
2. **Collect all errors**: Process the entire `Errors` slice to show users all problems at once
3. **Use error codes**: When available, use `FieldError.Code` for programmatic error handling
4. **Preserve error context**: Log the `Value` field to debug what value caused the error
5. **Return meaningful messages**: Format errors user-friendly in your API responses

```go
// Return validation errors as JSON in HTTP response
func handleUserCreation(w http.ResponseWriter, r *http.Request) {
    var jsonData []byte
    // ... read request body into jsonData ...

    user, err := pedantigo.Unmarshal[User](jsonData)
    if err != nil {
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]any{
                "error": "validation_failed",
                "errors": validationErr.Errors,
            })
            return
        }
        // Handle non-validation errors...
    }

    // Process valid user...
}
```

:::tip
See [Validation Basics](/docs/concepts/validation) for more details on how validation works.
:::
