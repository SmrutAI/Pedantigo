---
sidebar_position: 1
---

# Validation Basics

How validation works in Pedantigo.

## Overview

Pedantigo validates Go structs using struct tags, similar to how Go's built-in `encoding/json` package works. Validation automatically happens when you unmarshal JSON data, and you can also validate manually created structs.

## The Validation Pipeline

Validation follows a clear, predictable pipeline:

1. **Tag Parsing** - Struct tags are parsed to extract validation constraints (fail-fast at creation time)
2. **Deserialization** - JSON is unmarshaled to a map, preserving which keys were present
3. **Field Validation** - Each field's constraints are checked (required, min, max, format, etc.)
4. **Default Application** - Missing fields get default values or `defaultFactory` results
5. **Cross-Field Validation** - If your struct implements `Validatable`, its `Validate()` method is called
6. **Error Collection** - All errors are gathered and returned as a `ValidationError`

This means **all validation errors are collected at once** - you don't get just the first error.

## Using Struct Tags

Validation is controlled with the `pedantigo` struct tag:

```go
type User struct {
    Name     string `json:"name" pedantigo:"required,min=2,max=50"`
    Email    string `json:"email" pedantigo:"required,email"`
    Age      int    `json:"age" pedantigo:"min=0,max=150"`
    Status   string `json:"status" pedantigo:"required,enum=active|inactive|pending"`
    Verified bool   `json:"verified" pedantigo:"default=false"`
}
```

**Key syntax rules:**
- Constraints are comma-separated: `pedantigo:"constraint1,constraint2=value"`
- Constraints with values use `=`: `min=2`, `max=50`, `default=false`
- Enum values are pipe-separated: `enum=active|inactive|pending`
- Multiple constraints stack: `required,email,max=100`

## Validation Methods

### Simple API (Recommended)

The Simple API uses automatic caching for best performance:

#### Unmarshal + Validate

```go
type User struct {
    Email string `json:"email" pedantigo:"required,email"`
    Age   int    `json:"age" pedantigo:"min=18"`
}

jsonData := []byte(`{"email":"alice@example.com","age":25}`)
user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    // Handle validation errors
    var ve *pedantigo.ValidationError
    if errors.As(err, &ve) {
        for _, fieldErr := range ve.Errors {
            fmt.Printf("Field %s: %s\n", fieldErr.Field, fieldErr.Message)
        }
    }
    return
}
// user is valid and ready to use
```

#### Validate Existing Structs

For structs created manually (not from JSON), use `pedantigo.Validate()`:

```go
user := &User{
    Email: "bob@example.com",
    Age:   30,
}

if err := pedantigo.Validate(user); err != nil {
    // Handle validation errors
}
```

#### Create from Any Input

```go
// From a map
userMap := map[string]any{
    "email": "charlie@example.com",
    "age":   35,
}
user, err := pedantigo.NewModel[User](userMap)

// From a struct
user2 := User{Email: "dave@example.com", Age: 40}
user3, err := pedantigo.NewModel[User](user2)
```

### Advanced: Validator Object API

For advanced features or performance-critical code that creates validators once:

```go
// Create once, reuse many times
validator := pedantigo.New[User]()

// Unmarshal with this validator
user, err := validator.Unmarshal(jsonData)

// Get cached schema
schema := validator.Schema()

// Validate existing structs
err = validator.ValidateValue(user)
```

## Understanding `required`

The `required` constraint has special behavior:

**During Unmarshal (from JSON):**
- `required` means the field must be present in the JSON
- An explicit `null` value will fail validation if `required` is set
- Missing JSON keys are treated as missing fields

**During Validate (existing struct):**
- `required` is NOT checked
- Only format and constraint validations apply (min, max, email, etc.)
- Existing structs are assumed to be properly initialized

This distinction matters because:
- JSON unmarshal needs to know which fields were explicitly provided
- Direct struct validation assumes the struct was already initialized

Example:

```go
type Config struct {
    APIKey string `pedantigo:"required,min=10"`
}

// This passes (Unmarshal): JSON key is present
config, _ := pedantigo.Unmarshal[Config]([]byte(`{"apiKey":""}`))

// This fails (Unmarshal): JSON key is missing
config, err := pedantigo.Unmarshal[Config]([]byte(`{}`))
// Error: APIKey is required

// This passes (Validate): required not checked on existing struct
config := &Config{APIKey: ""}
pedantigo.Validate(config) // No error for empty string
```

## Field-Level vs Cross-Field Validation

### Field-Level Validation

Each field is validated independently using constraints:

```go
type Product struct {
    Name  string `pedantigo:"required,max=100"`
    Price float64 `pedantigo:"gt=0,lt=1000000"`
    Stock int `pedantigo:"min=0,max=1000"`
}
```

Field-level constraints are applied during the main validation phase.

### Cross-Field Validation

For validation that involves multiple fields, implement the `Validatable` interface:

```go
type DateRange struct {
    StartDate time.Time `json:"start_date" pedantigo:"required"`
    EndDate   time.Time `json:"end_date" pedantigo:"required"`
}

func (d *DateRange) Validate() error {
    if d.EndDate.Before(d.StartDate) {
        return errors.New("EndDate must be after StartDate")
    }
    return nil
}
```

Your `Validate()` method is called **after** all field constraints pass. This allows you to:
- Check relationships between fields
- Implement business logic validation
- Return detailed error messages

If your `Validate()` method returns an error, it's included in the `ValidationError.Errors` slice.

See [Cross-Field Validation](/docs/concepts/cross-field) for more details.

## Error Collection

Pedantigo collects all validation errors in one pass, not stopping at the first error:

```go
type User struct {
    Email string `json:"email" pedantigo:"required,email,max=100"`
    Age   int    `json:"age" pedantigo:"required,min=18,max=120"`
    Name  string `json:"name" pedantigo:"required,min=2"`
}

jsonData := []byte(`{
    "email": "not-an-email-and-this-is-way-too-long@example.com",
    "age": 10,
    "name": "A"
}`)

user, err := pedantigo.Unmarshal[User](jsonData)
// err contains 4 validation errors:
// - email: invalid email format
// - email: must be at most 100 characters
// - age: must be at least 18
// - name: must be at least 2 characters
```

This helps users fix all problems at once, rather than discovering them one by one.

## The Validatable Interface

To implement custom validation, your type must satisfy:

```go
type Validatable interface {
    Validate() error
}
```

Example:

```go
type User struct {
    FirstName string `json:"first_name" pedantigo:"required"`
    LastName  string `json:"last_name" pedantigo:"required"`
    Email     string `json:"email" pedantigo:"required,email"`
    Age       int    `json:"age" pedantigo:"min=0,max=150"`
}

func (u *User) Validate() error {
    if u.FirstName == u.LastName {
        return errors.New("FirstName and LastName cannot be the same")
    }
    return nil
}
```

Your `Validate()` method is called after all field-level validation passes. If it returns an error:
- The error message is added to the validation error list
- The error is associated with the struct (no specific field)

## Defaults and Factories

Fields can have default values:

```go
type Config struct {
    Port      int    `json:"port" pedantigo:"default=8080"`
    Debug     bool   `json:"debug" pedantigo:"default=false"`
    Environment string `json:"env" pedantigo:"default=development"`
}
```

Defaults are applied only for missing fields during `Unmarshal`. For dynamic defaults, use `defaultFactory`:

```go
type Session struct {
    Token   string    `json:"token" pedantigo:"required,defaultFactory=generateToken"`
    Created time.Time `json:"created" pedantigo:"defaultFactory=now"`
}

func generateToken() string {
    return uuid.New().String()
}

func now() time.Time {
    return time.Now()
}
```

The factory function is called only if the field is missing in the JSON.

## Common Validation Patterns

### Required Fields

```go
type LoginRequest struct {
    Username string `json:"username" pedantigo:"required,min=3"`
    Password string `json:"password" pedantigo:"required,min=8"`
}
```

### Email Validation

```go
type User struct {
    Email string `json:"email" pedantigo:"required,email"`
}
```

### Numeric Ranges

```go
type Product struct {
    Price float64 `json:"price" pedantigo:"gt=0,lt=1000000"`
    Stock int `json:"stock" pedantigo:"min=0,max=10000"`
}
```

### String Length

```go
type Post struct {
    Title   string `json:"title" pedantigo:"required,min=5,max=200"`
    Content string `json:"content" pedantigo:"required,min=10,max=10000"`
}
```

### Enum Validation

```go
type Order struct {
    Status string `json:"status" pedantigo:"required,enum=pending|processing|shipped|delivered"`
}
```

### URL and Format Validation

```go
type Website struct {
    URL      string `json:"url" pedantigo:"required,url"`
    IPv4     string `json:"ipv4" pedantigo:"ipv4"`
    IPv6     string `json:"ipv6" pedantigo:"ipv6"`
    UUID     string `json:"uuid" pedantigo:"uuid"`
}
```

## Next Steps

- See [Constraints Reference](/docs/constraints/string) for the complete list of available constraints
- Learn about [Cross-Field Validation](/docs/concepts/cross-field) for complex business logic
- Check out [Examples](/docs/examples/basic) for real-world patterns
