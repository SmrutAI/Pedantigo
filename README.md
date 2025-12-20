# Pedantigo

[![CI](https://github.com/SmrutAI/pedantigo/actions/workflows/ci.yml/badge.svg)](https://github.com/SmrutAI/pedantigo/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/tushar2708/67408111d1830ed523e2661d9ee2a442/raw/pedantigo-coverage.json)](https://github.com/SmrutAI/pedantigo)
[![Go Report Card](https://goreportcard.com/badge/github.com/SmrutAI/pedantigo)](https://goreportcard.com/report/github.com/SmrutAI/pedantigo)

Type-safe JSON validation and schema generation for Go.

## Installation

```bash
go get github.com/SmrutAI/pedantigo
```

Requires Go 1.21+

## When to Use Pedantigo

| Use Case | Why Pedantigo? |
|----------|----------------|
| **API Request Validation** | Validate incoming JSON, return structured errors |
| **LLM Structured Output** | Generate JSON Schema for function calling, validate responses |
| **Configuration Files** | Parse config with defaults, fail fast on invalid values |
| **Data Pipeline Input** | Ensure data quality at ingestion with detailed error paths |

Pedantigo combines JSON unmarshaling with validation in a single step. Define constraints once in struct tags, get validated data and JSON Schema automatically.

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

> **Two Ways to Validate:**
> - `Unmarshal(jsonBytes)` — Parse JSON and validate in one step
> - `Validate(structPtr)` — Validate an existing Go struct

## Feature Coverage

See [API_PARITY.md](API_PARITY.md) for detailed feature comparison with Pydantic v2 and go-playground/validator.

## Core Validation

### Creating a Validator

Use `New[T]()` to create a type-safe validator:

```go
validator := pedantigo.New[User]()
```

The validator is built once and can be reused. It pre-compiles all validation rules for performance.

### Reuse Validators for Performance

Create validators ONCE and reuse them. Don't create a new validator every time.

```go
// DO: Create once, reuse many times
var userValidator = pedantigo.New[User]()  // Package-level

// OR store in a struct
type Service struct {
    userValidator *pedantigo.Validator[User]
}

func NewService() *Service {
    return &Service{userValidator: pedantigo.New[User]()}
}

// DON'T: Create per-request (loses caching benefit!)
func HandleRequest(data []byte) (*User, error) {
    v := pedantigo.New[User]()  // Wasteful! Rebuilds every time
    return v.Unmarshal(data)
}
```

**Why reuse?** `New[T]()` parses struct tags and compiles validation rules. Creating it once avoids repeated reflection overhead. Schema generation (`validator.Schema()`) is also cached.

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

NOTE: Unlike JSON, for structs, required fields cannot be verified for missing values. (In Go, structs never have missing values)

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

| Constraint         | Description                                        | Example                                    |
|--------------------|----------------------------------------------------|--------------------------------------------|
| `required`         | Field must be present in JSON                      | `pedantigo:"required"`                     |
| `min`              | Minimum value (numbers) or length (strings/slices) | `pedantigo:"min=18"`                       |
| `max`              | Maximum value (numbers) or length (strings/slices) | `pedantigo:"max=100"`                      |
| `gt`               | Greater than (numbers only)                        | `pedantigo:"gt=0"`                         |
| `gte`              | Greater than or equal (numbers only)               | `pedantigo:"gte=1"`                        |
| `lt`               | Less than (numbers only)                           | `pedantigo:"lt=100"`                       |
| `lte`              | Less than or equal (numbers only)                  | `pedantigo:"lte=99"`                       |
| `email`            | Valid email address                                | `pedantigo:"email"`                        |
| `url`              | Valid URL                                          | `pedantigo:"url"`                          |
| `uuid`             | Valid UUID                                         | `pedantigo:"uuid"`                         |
| `ipv4`             | Valid IPv4 address                                 | `pedantigo:"ipv4"`                         |
| `ipv6`             | Valid IPv6 address                                 | `pedantigo:"ipv6"`                         |
| `ip`               | Valid IP address (IPv4 or IPv6)                    | `pedantigo:"ip"`                           |
| `cidr`             | Valid CIDR notation                                | `pedantigo:"cidr"`                         |
| `mac`              | Valid MAC address                                  | `pedantigo:"mac"`                          |
| `hostname`         | Valid RFC 952 hostname                             | `pedantigo:"hostname"`                     |
| `fqdn`             | Valid fully qualified domain name                  | `pedantigo:"fqdn"`                         |
| `port`             | Valid port number (0-65535)                        | `pedantigo:"port"`                         |
| `regexp`           | Match regular expression                           | `pedantigo:"regexp=^[A-Z]+$"`              |
| `oneof`            | Value must be one of specified options             | `pedantigo:"oneof=red green blue"`         |
| `eqfield`          | Field equals another field                         | `pedantigo:"eqfield=Password"`             |
| `nefield`          | Field not equal to another field                   | `pedantigo:"nefield=OldPassword"`          |
| `gtfield`          | Greater than another field                         | `pedantigo:"gtfield=MinPrice"`             |
| `gtefield`         | Greater than or equal to another field             | `pedantigo:"gtefield=StartDate"`           |
| `ltfield`          | Less than another field                            | `pedantigo:"ltfield=MaxPrice"`             |
| `ltefield`         | Less than or equal to another field                | `pedantigo:"ltefield=EndDate"`             |
| `required_if`      | Required if another field has value                | `pedantigo:"required_if=Country:USA"`      |
| `required_unless`  | Required unless another field has value            | `pedantigo:"required_unless=Type:guest"`   |
| `required_with`    | Required if another field is present               | `pedantigo:"required_with=Address"`        |
| `required_without` | Required if another field is absent                | `pedantigo:"required_without=Email"`       |
| `excluded_if`      | Excluded if another field has value                | `pedantigo:"excluded_if=Type admin"`       |
| `excluded_unless`  | Excluded unless another field has value            | `pedantigo:"excluded_unless=Role user"`    |
| `excluded_with`    | Excluded if another field is present               | `pedantigo:"excluded_with=TempToken"`      |
| `excluded_without` | Excluded if another field is absent                | `pedantigo:"excluded_without=PermanentID"` |
| `len`              | Exact length (strings/slices)                      | `pedantigo:"len=10"`                       |
| `alpha`            | Letters only                                       | `pedantigo:"alpha"`                        |
| `alphanum`         | Letters and numbers only                           | `pedantigo:"alphanum"`                     |
| `ascii`            | ASCII characters only                              | `pedantigo:"ascii"`                        |
| `lowercase`        | Must be lowercase                                  | `pedantigo:"lowercase"`                    |
| `uppercase`        | Must be uppercase                                  | `pedantigo:"uppercase"`                    |
| `contains`         | Must contain substring                             | `pedantigo:"contains=@"`                   |
| `excludes`         | Must not contain substring                         | `pedantigo:"excludes=<"`                   |
| `startswith`       | Must start with prefix                             | `pedantigo:"startswith=http"`              |
| `endswith`         | Must end with suffix                               | `pedantigo:"endswith=.com"`                |
| `positive`         | Must be > 0 (numbers only)                         | `pedantigo:"positive"`                     |
| `negative`         | Must be < 0 (numbers only)                         | `pedantigo:"negative"`                     |
| `multiple_of`      | Must be divisible by value                         | `pedantigo:"multiple_of=5"`                |
| `max_digits`       | Maximum total digits                               | `pedantigo:"max_digits=10"`                |
| `decimal_places`   | Maximum decimal places                             | `pedantigo:"decimal_places=2"`             |
| `credit_card`      | Valid credit card number (Luhn)                    | `pedantigo:"credit_card"`                  |
| `isbn`             | Valid ISBN-10 or ISBN-13                           | `pedantigo:"isbn"`                         |
| `ssn`              | Valid U.S. SSN (XXX-XX-XXXX)                       | `pedantigo:"ssn"`                          |
| `e164`             | Valid E.164 phone number                           | `pedantigo:"e164"`                         |
| `latitude`         | Valid latitude (-90 to 90)                         | `pedantigo:"latitude"`                     |
| `longitude`        | Valid longitude (-180 to 180)                      | `pedantigo:"longitude"`                    |
| `hexcolor`         | Valid hex color (#RGB or #RRGGBB)                  | `pedantigo:"hexcolor"`                     |
| `jwt`              | Valid JWT format                                   | `pedantigo:"jwt"`                          |
| `json`             | Valid JSON string                                  | `pedantigo:"json"`                         |
| `base64`           | Valid base64 encoding                              | `pedantigo:"base64"`                       |
| `md5`              | Valid MD5 hash (32 hex chars)                      | `pedantigo:"md5"`                          |
| `sha256`           | Valid SHA256 hash (64 hex chars)                   | `pedantigo:"sha256"`                       |
| `semver`           | Valid semantic version (X.Y.Z)                     | `pedantigo:"semver"`                       |
| `ulid`             | Valid ULID (26 chars)                              | `pedantigo:"ulid"`                         |
| `cron`             | Valid cron expression                              | `pedantigo:"cron"`                         |

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

Use cross-field tags to compare or conditionally require fields:

```go
type PriceRange struct {
    MinPrice int `json:"min_price" pedantigo:"required,min=0"`
    MaxPrice int `json:"max_price" pedantigo:"required,gtfield=MinPrice"`
}

type Registration struct {
    Password        string `json:"password" pedantigo:"required,min=8"`
    PasswordConfirm string `json:"password_confirm" pedantigo:"required,eqfield=Password"`
}

type Address struct {
    Country    string `json:"country"`
    PostalCode string `json:"postal_code" pedantigo:"required_if=Country:USA"`
}
```

For custom validation logic, implement the `Validatable` interface:

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

## Error Codes

Every validation error includes a machine-readable error code for programmatic handling:

```go
user, err := validator.Unmarshal(jsonData)
if err != nil {
    ve := err.(*pedantigo.ValidationError)
    for _, fe := range ve.Errors {
        switch fe.Code {
        case "REQUIRED":
            // Handle missing required field
        case "INVALID_EMAIL":
            // Handle invalid email format
        case "MIN_VALUE":
            // Handle value below minimum
        default:
            // Handle other errors
        }
        fmt.Printf("Field: %s, Code: %s, Message: %s\n", fe.Field, fe.Code, fe.Message)
    }
}
```

Common error codes include:
- `REQUIRED`, `REQUIRED_IF`, `REQUIRED_WITH` - Missing field errors
- `INVALID_EMAIL`, `INVALID_URL`, `INVALID_UUID` - Format errors
- `MIN_VALUE`, `MAX_VALUE`, `MIN_LENGTH`, `MAX_LENGTH` - Range errors
- `PATTERN_MISMATCH` - Regex validation failed
- `INVALID_ENUM` - Value not in allowed set

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

## Advanced: OpenAPI/Swagger Schema (Optional)

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

### Schema Metadata

Add titles, descriptions, and examples to improve schema quality for LLM prompt engineering:

```go
type UserInput struct {
    Name  string `json:"name" pedantigo:"required,title=User Name,description=Full name of the user,example=John Doe"`
    Email string `json:"email" pedantigo:"required,email,title=Email Address,description=Primary contact email"`
    Tags  []string `json:"tags" pedantigo:"examples=work|personal|urgent"` // Multiple examples with pipe separator
}

validator := pedantigo.New[UserInput]()
schema := validator.Schema()
```

Generated schema includes metadata:
```json
{
  "properties": {
    "name": {
      "type": "string",
      "title": "User Name",
      "description": "Full name of the user",
      "examples": ["John Doe"]
    },
    "email": {
      "type": "string",
      "format": "email",
      "title": "Email Address",
      "description": "Primary contact email"
    },
    "tags": {
      "type": "array",
      "items": {"type": "string"},
      "examples": ["work", "personal", "urgent"]
    }
  }
}
```

## Advanced: Marshal with Options (Optional)

Control JSON output with field exclusion and empty value handling:

```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Password string `json:"password"`
    Nickname string `json:"nickname"`
    Bio      string `json:"bio"`
}

user := &User{
    ID:       1,
    Name:     "John",
    Password: "secret123",
    Nickname: "",  // empty
    Bio:      "",  // empty
}

validator := pedantigo.New[User]()

// Exclude sensitive fields, omit empty optional fields
data, err := validator.MarshalWithOptions(user, pedantigo.MarshalOptions{
    Exclude:   []string{"password"},        // Never include password
    OmitEmpty: []string{"nickname", "bio"}, // Omit if empty
})
// Result: {"id":1,"name":"John"}
```

Options:
- `Exclude` - Fields to never include in output
- `OmitEmpty` - Fields to omit when they have zero values

## Advanced: Streaming JSON (Optional)

Parse incomplete/chunked JSON from LLM streaming responses:

```go
type ToolCall struct {
    Name string         `json:"name" pedantigo:"required"`
    Args map[string]any `json:"args"`
}

parser := pedantigo.NewStreamParser[ToolCall]()

// Simulate LLM streaming chunks
chunks := []string{
    `{"name": "get_`,
    `weather", "args": {"city": "NYC`,
    `"}}`,
}

for _, chunk := range chunks {
    result, state, err := parser.Feed(chunk)
    if err != nil {
        log.Fatal(err)
    }

    if state.IsComplete {
        fmt.Printf("Complete: %+v\n", result)
        // Complete: {Name:get_weather Args:map[city:NYC]}
    } else {
        fmt.Printf("Partial (confidence: %.0f%%)\n", state.Confidence*100)
    }
}
```

StreamParser features:
- Accumulates JSON chunks until complete
- Reports parsing confidence (0.0 to 1.0)
- Validates completed JSON against struct constraints
- Handles nested objects and arrays

## Advanced: Performance Mode (Optional)

A lot of gophers like the zero-values, and don't want to have even the slightest performance drop that comes with additional validations.
For them, we have a bypass to continue using Go's zero-value based validations.

Skip required-field checking and default-value application for better performance when using Go's zero-value semantics.

### When to Use

Use `StrictMissingFields: false` when:
- You can handle optionality with pointers (`*int`, `*bool`)
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

## Advanced: Extra Fields Handling (Optional)

Control how unknown JSON fields are handled during unmarshaling.

### Available Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| `ExtraIgnore` | Silently discard unknown fields | Default Go behavior |
| `ExtraForbid` | Return error on unknown fields | Strict API validation |
| `ExtraAllow` | Store unknown fields for inspection | Flexible data handling |

### Usage

```go
type User struct {
    Name string `json:"name" pedantigo:"required"`
    Age  int    `json:"age"`
}

// Default: ignores unknown fields
validator := pedantigo.New[User]()

// Strict mode: reject unknown fields
strictValidator := pedantigo.New[User](pedantigo.ValidatorOptions{
    ExtraFields: pedantigo.ExtraForbid,
})

jsonData := []byte(`{"name": "John", "age": 30, "unknown_field": true}`)

// ExtraIgnore → succeeds, unknown_field discarded
// ExtraForbid → error: "unknown field in JSON"
```

### When to Use

- **ExtraIgnore** (default): API evolution, backward compatibility
- **ExtraForbid**: Strict API contracts, prevent typos in field names
- **ExtraAllow**: Audit logging, pass-through data

## Advanced: Discriminated Unions (Optional)

Validate JSON where a field determines which variant type applies. Like Pydantic's `Discriminator` or TypeScript's discriminated unions.

### When to Use

- API responses with different shapes based on `type` field
- Polymorphic data (e.g., different payment methods, notification types)
- Any tagged union pattern

### Usage

```go
// Define variant types
type Cat struct {
    Name  string `json:"name" pedantigo:"required"`
    Lives int    `json:"lives" pedantigo:"min=1,max=9"`
}

type Dog struct {
    Name  string `json:"name" pedantigo:"required"`
    Breed string `json:"breed"`
}

// Create union validator with discriminator field
validator, err := pedantigo.NewUnion[any](pedantigo.UnionOptions{
    DiscriminatorField: "pet_type",
    Variants: []pedantigo.UnionVariant{
        pedantigo.VariantFor[Cat]("cat"),
        pedantigo.VariantFor[Dog]("dog"),
    },
})
if err != nil {
    log.Fatal(err)
}

// Unmarshal dispatches based on discriminator
catJSON := []byte(`{"pet_type": "cat", "name": "Whiskers", "lives": 9}`)
result, err := validator.Unmarshal(catJSON)
if err != nil {
    // Validation error or unknown variant
}

cat := result.(Cat) // Type assertion to concrete type
fmt.Printf("Cat: %s has %d lives\n", cat.Name, cat.Lives)

dogJSON := []byte(`{"pet_type": "dog", "name": "Rex", "breed": "German Shepherd"}`)
result, err = validator.Unmarshal(dogJSON)
dog := result.(Dog)
fmt.Printf("Dog: %s is a %s\n", dog.Name, dog.Breed)
```

### Schema Generation

Union validators generate JSON Schema with `oneOf`:

```go
schema := validator.Schema()
jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
```

Output:

```json
{
  "oneOf": [
    {
      "type": "object",
      "properties": {
        "pet_type": {"const": "cat"},
        "name": {"type": "string"},
        "lives": {"type": "integer", "minimum": 1, "maximum": 9}
      },
      "required": ["name"]
    },
    {
      "type": "object",
      "properties": {
        "pet_type": {"const": "dog"},
        "name": {"type": "string"},
        "breed": {"type": "string"}
      },
      "required": ["name"]
    }
  ]
}
```

### Validate Existing Values

```go
cat := Cat{Name: "Whiskers", Lives: 9}
err := validator.Validate(cat)
```

### Error Handling

```go
// Missing discriminator field
json := []byte(`{"name": "Unknown"}`)
_, err := validator.Unmarshal(json)
// Error: discriminator field "pet_type" is missing

// Unknown discriminator value
json = []byte(`{"pet_type": "fish", "name": "Nemo"}`)
_, err = validator.Unmarshal(json)
// Error: unknown discriminator value "fish" for field "pet_type"

// Variant validation failure
json = []byte(`{"pet_type": "cat", "name": "Whiskers", "lives": 15}`)
_, err = validator.Unmarshal(json)
// Error: lives: must be at most 9
```

## Controversies

Some design decisions differ from Pydantic due to Go's type system:

- **[Why No BaseModel?](documents/nuances/why_not_basemodel.md)** — External validators over embedding. BaseModel adds initialization boilerplate with minimal benefit; `validator.Validate(&user)` is more idiomatic than `user.Validate()`.

- **[Why No `.model_rebuild()`?](documents/nuances/model_rebuild.md)** — Go resolves types at compile-time using pointers; no runtime forward reference resolution needed.

- **[How to create Computed Fields](documents/nuances/computed_derived_fields.md)** — Go uses `MarshalJSON()` interface instead of decorators. More boilerplate, but zero runtime overhead.

I will revisit these based on what the community prefers.

## License

MIT
