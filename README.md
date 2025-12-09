# Pedantigo

Type-safe JSON validation and schema generation for Go.

## Installation

```bash
go get github.com/SmrutAI/Pedantigo
```

Requires Go 1.25.4+

## Quick Start

```go
type User struct {
    Email string `json:"email" pedantigo:"required,email"`
    Age   int    `json:"age" pedantigo:"min=18"`
}

validator := pedantigo.New[User]()
user, err := validator.Unmarshal(jsonData)
if err != nil {
    // Handle validation errors
}
```

## Feature Coverage

See [API_PARITY.md](API_PARITY.md) for detailed feature comparison with Pydantic v2 and go-playground/validator.

## Core Validation

### Creating a Validator

Use `New[T]()` to create a type-safe validator:

```go
validator := pedantigo.New[User]()
```

The validator is built once and can be reused. It pre-compiles all validation rules for performance.

### Validation Tags

Add validation rules using the `pedantigo` struct tag:

```go
type User struct {
    Name     string `json:"name" pedantigo:"required,min=3,max=50"`
    Email    string `json:"email" pedantigo:"required,email"`
    Age      int    `json:"age" pedantigo:"min=18,max=120"`
    Website  string `json:"website" pedantigo:"url"`
    Role     string `json:"role" pedantigo:"oneof=admin user guest"`
    Password string `json:"password" pedantigo:"min=8,regexp=^[a-zA-Z0-9]+$"`
}
```

### Unmarshal and Validate

`Unmarshal()` parses JSON and validates in one call:

```go
jsonData := []byte(`{"email":"john@example.com","age":25}`)
user, err := validator.Unmarshal(jsonData)

if err != nil {
    if ve, ok := err.(*pedantigo.ValidationError); ok {
        for _, fieldErr := range ve.Errors {
            fmt.Printf("%s: %s\n", fieldErr.Field, fieldErr.Message)
        }
    }
    return err
}

// user is valid and ready to use
fmt.Printf("User: %+v\n", user)
```

### Validate Existing Structs

Use `Validate()` on structs you created manually:

```go
user := &User{
    Email: "invalid-email",
    Age:   15,
}

err := validator.Validate(user)
if err != nil {
    ve := err.(*pedantigo.ValidationError)
    // ve.Errors contains: Email must be valid, Age must be at least 18
}
```

**Important**: `Validate()` works on Go structs, not JSON data. It **cannot distinguish between "missing" and "zero value"** because Go initializes all struct fields to their zero values (`0`, `false`, `""`).

- For `int` fields: `0` is indistinguishable from "not set"
- For `bool` fields: `false` is indistinguishable from "not set"
- For `string` fields: `""` is indistinguishable from "not set"

**If you need to detect missing fields**, use `Unmarshal()` instead, which parses JSON and can distinguish between:
- `{"age": 0}` (age explicitly set to 0)
- `{}` (age missing from JSON)

Alternatively, use pointer types (`*int`, `*bool`, `*string`) where `nil` indicates "not set".

### Available Constraints

| Constraint | Description | Example |
|------------|-------------|---------|
| `required` | Field must be present in JSON | `pedantigo:"required"` |
| `min` | Minimum value (numbers) or length (strings/slices) | `pedantigo:"min=18"` |
| `max` | Maximum value (numbers) or length (strings/slices) | `pedantigo:"max=100"` |
| `gt` | Greater than (numbers only) | `pedantigo:"gt=0"` |
| `gte` | Greater than or equal (numbers only) | `pedantigo:"gte=1"` |
| `lt` | Less than (numbers only) | `pedantigo:"lt=100"` |
| `lte` | Less than or equal (numbers only) | `pedantigo:"lte=99"` |
| `email` | Valid email address | `pedantigo:"email"` |
| `url` | Valid URL | `pedantigo:"url"` |
| `uuid` | Valid UUID | `pedantigo:"uuid"` |
| `ipv4` | Valid IPv4 address | `pedantigo:"ipv4"` |
| `ipv6` | Valid IPv6 address | `pedantigo:"ipv6"` |
| `regexp` | Match regular expression | `pedantigo:"regexp=^[A-Z]+$"` |
| `oneof` | Value must be one of specified options | `pedantigo:"oneof=red green blue"` |

Combine multiple constraints with commas: `pedantigo:"required,min=3,max=50"`

### Default Values

Set default values for missing fields:

```go
type Config struct {
    Port    int    `json:"port" pedantigo:"default=8080"`
    Host    string `json:"host" pedantigo:"default=localhost"`
    Timeout int    `json:"timeout" pedantigo:"default=30"`
}

// JSON: {}
// Result: Port=8080, Host="localhost", Timeout=30
```

Use `defaultUsingMethod` to compute defaults dynamically:

```go
type Session struct {
    ID        string    `json:"id" pedantigo:"defaultUsingMethod=GenerateID"`
    CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=Now"`
}

func (s *Session) GenerateID() (string, error) {
    return uuid.New().String(), nil
}

func (s *Session) Now() (time.Time, error) {
    return time.Now(), nil
}
```

Methods must have signature `func(*T) (FieldType, error)`.

### Cross-Field Validation

Implement the `Validatable` interface for custom validation logic:

```go
type Registration struct {
    Password        string `json:"password" pedantigo:"required,min=8"`
    PasswordConfirm string `json:"password_confirm" pedantigo:"required"`
}

func (r *Registration) Validate() error {
    if r.Password != r.PasswordConfirm {
        return &pedantigo.ValidationError{
            Errors: []pedantigo.FieldError{{
                Field:   "password_confirm",
                Message: "passwords must match",
            }},
        }
    }
    return nil
}
```

## Schema Generation

Generate JSON Schema for LLM function calling and structured outputs.

### Basic Usage

```go
type WeatherQuery struct {
    City string `json:"city" pedantigo:"required"`
    Unit string `json:"unit" pedantigo:"oneof=celsius fahrenheit"`
}

validator := pedantigo.New[WeatherQuery]()
schema := validator.Schema()

// Or get JSON bytes directly
jsonBytes, _ := validator.SchemaJSON()
```

### LLM Integration

Use schemas with OpenAI function calling:

```go
type ExtractInfo struct {
    Name  string `json:"name" pedantigo:"required"`
    Email string `json:"email" pedantigo:"required,email"`
    Age   int    `json:"age" pedantigo:"min=0,max=150"`
}

validator := pedantigo.New[ExtractInfo]()
schemaJSON, _ := validator.SchemaJSON()

// Pass schemaJSON to OpenAI's function calling parameter
// Or Anthropic's tool definition
```

Validation tags automatically map to JSON Schema properties:
- `required` → `required` array
- `min`/`max` → `minimum`/`maximum` (numbers) or `minLength`/`maxLength` (strings)
- `email` → `format: "email"`
- `url` → `format: "uri"`
- `oneof` → `enum` array

### Nested Structures

Schemas support nested structs, slices, and maps:

```go
type Address struct {
    Street string `json:"street" pedantigo:"required"`
    City   string `json:"city" pedantigo:"required"`
    Zip    string `json:"zip" pedantigo:"min=5,max=10"`
}

type User struct {
    Name      string    `json:"name" pedantigo:"required"`
    Address   Address   `json:"address" pedantigo:"required"`
    Emails    []string  `json:"emails" pedantigo:"min=1,email"`
    Metadata  map[string]string `json:"metadata"`
}

validator := pedantigo.New[User]()
schema := validator.Schema()
// Generates fully nested schema with all constraints
```

## Advanced: OpenAPI Schema (Optional)

For OpenAPI specifications and Swagger documentation, use schemas with `$ref` for reusable type definitions.

### When to Use

- Building OpenAPI 3.0 specifications
- Generating Swagger UI documentation
- API documentation tools that support `$ref`

### Usage

```go
type Product struct {
    Name  string  `json:"name" pedantigo:"required,min=3"`
    Price float64 `json:"price" pedantigo:"required,min=0"`
}

type Order struct {
    Products []Product `json:"products" pedantigo:"required,min=1"`
    Total    float64   `json:"total" pedantigo:"required,min=0"`
}

validator := pedantigo.New[Order]()

// Generate schema with $ref/$defs
schema := validator.SchemaOpenAPI()
jsonBytes, _ := validator.SchemaJSONOpenAPI()
```

### Difference from Default Schema

**Default (`Schema()`)**: Expands all nested types inline. Used by LLM APIs that don't support `$ref`.

```json
{
  "type": "object",
  "properties": {
    "products": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string", "minLength": 3},
          "price": {"type": "number", "minimum": 0}
        }
      }
    }
  }
}
```

**OpenAPI (`SchemaOpenAPI()`)**: Uses `$ref` to reference reusable definitions.

```json
{
  "type": "object",
  "properties": {
    "products": {
      "type": "array",
      "items": {"$ref": "#/$defs/Product"}
    }
  },
  "$defs": {
    "Product": {
      "type": "object",
      "properties": {
        "name": {"type": "string", "minLength": 3},
        "price": {"type": "number", "minimum": 0}
      }
    }
  }
}
```

Constraints are applied to all definitions, including referenced types.

## Advanced: Performance Mode (Optional)

Skip required-field checking and default-value application for better performance when using Go's zero-value semantics.

### When to Use

Use `StrictMissingFields: false` when:
- You handle optionality with pointers (`*int`, `*bool`)
- You prefer zero values over explicit defaults
- You don't need "field required" errors

### Usage

```go
type Config struct {
    Port    *int  `json:"port" pedantigo:"min=1024"`     // nil = not provided
    Enabled *bool `json:"enabled"`                       // nil = not provided
    Name    string `json:"name" pedantigo:"min=3"`       // "" = zero value
}

validator := pedantigo.New[Config](pedantigo.ValidatorOptions{
    StrictMissingFields: false,
})

// JSON: {}
config, err := validator.Unmarshal(jsonData)

if err != nil {
    // Port = nil, Enabled = nil, Name = ""
    // No "required field" errors
    // Validation constraints still run on provided values
    return err
}
```

### Behavior Changes

With `StrictMissingFields: false`:

1. **Skips 2-step unmarshal**: Uses direct `json.Unmarshal` (faster)
2. **No required-field errors**: Missing fields get zero values
3. **No default values**: `default=` and `defaultUsingMethod=` tags are ignored
4. **Validators still run**: Constraints validate zero values and provided values
5. **Nil pointers skip validation**: `*int` with `min=1024` → nil pointer passes

### Zero Values vs Pointers

**Non-pointer fields** with constraints may fail on zero values:

```go
type User struct {
    Age int `json:"age" pedantigo:"min=18"`
}

// JSON: {}
// Age = 0 → fails validation (0 < 18)
```

**Pointer fields** skip validation when nil:

```go
type User struct {
    Age *int `json:"age" pedantigo:"min=18"`
}

// JSON: {}
// Age = nil → validation skipped ✓

// JSON: {"age": 15}
// Age = 15 → fails validation (15 < 18)
```

### Safety Check

Attempting to use `default=` or `defaultUsingMethod=` tags with `StrictMissingFields: false` panics at validator creation:

```go
type Config struct {
    Port int `json:"port" pedantigo:"default=8080"`
}

validator := pedantigo.New[Config](pedantigo.ValidatorOptions{
    StrictMissingFields: false,
})
// Panics: field Config.Port has 'default=' tag but StrictMissingFields is false
```

This prevents silent bugs from ignored default values.

### Default Behavior

By default, `StrictMissingFields: true`:
- Required fields must be present in JSON
- Default values are applied to missing fields
- 2-step unmarshal for accurate missing-field detection

```go
// These are equivalent:
validator := pedantigo.New[User]()
validator := pedantigo.New[User](pedantigo.ValidatorOptions{
    StrictMissingFields: true,
})
```

## License

MIT
