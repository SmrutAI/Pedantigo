---
sidebar_position: 0
---

# Initialization & Configuration Reference

Complete reference for all Pedantigo initialization methods and configuration options.

:::note v2: default `TagName` changed
The default struct tag (and `TagName`/`SetTagName` default, described below) changed from `pedantigo` to `validate` in v2. See the [v1 to v2 migration guide](../migration/v1-to-v2) if you're upgrading. New projects should use v2 — v1 receives no new features.
:::

## Quick Comparison {#quick-comparison}

| API             | Use Case             | Customizable       | Link                  |
|-----------------|----------------------|--------------------|-----------------------|
| Simple API      | 80% of cases         | No (uses defaults) | [→](#simple-api)      |
| Validator API   | Custom options       | Yes                | [→](#validator-api)   |
| Stream Parser   | LLM/partial JSON     | Yes                | [→](#stream-parser)   |
| Union Validator | Discriminated unions | Yes                | [→](#union-validator) |

---

## Simple API {#simple-api}

Global functions with automatic caching. Uses [DefaultOptions()](#default-options).

| Function                                     | Description                      |
|----------------------------------------------|----------------------------------|
| `validator.Unmarshal[T](data)`               | Unmarshal JSON and validate      |
| `validator.Validate[T](obj)`                 | Validate existing struct         |
| `validator.NewModel[T](input)`               | Create from map/struct           |
| `validator.Schema[T]()`                      | Get JSON Schema                  |
| `validator.SchemaJSON[T]()`                  | Get JSON Schema as bytes         |
| `validator.SchemaLLM[T]()`                   | Get JSON Schema for LLM APIs (no `$schema` or `$id`) |
| `validator.SchemaJSONLLM[T]()`               | LLM schema as bytes (no `$schema` or `$id`) |
| `validator.SchemaOpenAPI[T]()`               | Get OpenAPI 3.1 component schema |
| `validator.SchemaJSONOpenAPI[T]()`           | OpenAPI 3.1 schema as bytes      |
| `validator.Marshal[T](obj)`                  | Marshal struct to JSON           |
| `validator.MarshalWithOptions[T](obj, opts)` | Marshal with options             |
| `validator.Dict[T](obj)`                     | Convert to map[string]any        |

```go
// Simple API: automatic defaults, global caching
user, err := validator.Unmarshal[User](jsonData)

// Equivalent to:
userValidator := validator.New[User](validator.DefaultOptions())
user, err := userValidator.Unmarshal(jsonData)
```

See [Simple API Reference](/docs/api/simple-api) for detailed examples.

---

## Validator API {#validator-api}

Explicit validator instances with [custom options](#validator-options).

### Creating a Validator {#creating-validator}

| Function                 | Description                                      |
|--------------------------|--------------------------------------------------|
| `validator.New[T]()`     | Create with [defaults](#default-options)         |
| `validator.New[T](opts)` | Create with [custom options](#validator-options) |

### Validator Methods {#validator-methods}

| Method                            | Description               |
|-----------------------------------|---------------------------|
| `v.Unmarshal(data)`               | Unmarshal with validation |
| `v.Validate(obj)`                 | Validate struct           |
| `v.NewModel(input)`               | Create from map/struct    |
| `v.Schema()`                      | Get JSON Schema           |
| `v.SchemaJSON()`                  | Schema as bytes           |
| `v.SchemaLLM()`                   | LLM schema (no `$schema` or `$id`) |
| `v.SchemaJSONLLM()`               | LLM schema as bytes (no `$schema` or `$id`) |
| `v.SchemaOpenAPI()`               | OpenAPI 3.1 schema        |
| `v.SchemaJSONOpenAPI()`           | OpenAPI 3.1 as bytes      |
| `v.Marshal(obj)`                  | Marshal to JSON           |
| `v.MarshalWithOptions(obj, opts)` | Marshal with options      |
| `v.Dict(obj)`                     | Convert to map            |

```go
// Create validator with custom options
userValidator := validator.New[User](validator.Options{
StrictMissingFields: true,
ExtraFields:         validator.ExtraForbid,
})

// Reuse the same validator multiple times
user1, err := userValidator.Unmarshal(data1)
user2, err := userValidator.Unmarshal(data2)

// Schema is cached, so subsequent calls are very fast
schema := userValidator.Schema()
```

Use the Validator API when you need:

- Custom options
- Schema caching for high-performance scenarios
- Consistent behavior across multiple unmarshal calls
- Explicit control over validator creation

See [Validator Reference](/docs/api/validator) for detailed examples.

---

## Stream Parser API {#stream-parser}

For partial/streaming JSON (ideal for LLM outputs).

| Function                                       | Description                                      |
|------------------------------------------------|--------------------------------------------------|
| `validator.NewStreamParser[T]()`               | Create with [defaults](#default-options)         |
| `validator.NewStreamParser[T](opts)`           | Create with [custom options](#validator-options) |
| `validator.NewStreamParserWithValidator[T](v)` | Use existing validator                           |

### Stream Parser Methods {#stream-parser-methods}

| Method           | Description                                |
|------------------|--------------------------------------------|
| `sp.Feed(chunk)` | Process JSON chunk, returns partial result |

See [Streaming Validation](/docs/concepts/streaming) for detailed examples.

---

## Union Validator API {#union-validator}

For discriminated unions with type switching.

| Function                      | Description            |
|-------------------------------|------------------------|
| `validator.NewUnion[T](opts)` | Create union validator |

### Union Validator Methods {#union-validator-methods}

| Method               | Description                  |
|----------------------|------------------------------|
| `uv.Unmarshal(data)` | Unmarshal and return variant |

See [Unions](/docs/concepts/unions) for detailed examples.

---

## Options {#validator-options}

Configuration struct for Validator API, Stream Parser, and Union Validator.

```go
type Options struct {
// StrictMissingFields controls behavior for missing fields
// Default: true (missing fields without defaults cause errors)
StrictMissingFields bool

// ExtraFields controls how unknown JSON fields are handled
// Default: ExtraIgnore (unknown fields are silently ignored)
ExtraFields ExtraFieldsMode

// TagName overrides the global struct tag name for this validator instance
// Default: "" (uses global default "validate")
TagName string
}
```

| Field                 | Type              | Default       | Description                                      |
|-----------------------|-------------------|---------------|--------------------------------------------------|
| `StrictMissingFields` | `bool`            | `true`        | [Missing field handling](#strict-missing-fields) |
| `ExtraFields`         | `ExtraFieldsMode` | `ExtraIgnore` | [Unknown field handling](#extra-fields)          |
| `TagName`             | `string`          | `""`          | [Custom tag name](#tag-name)                     |

### Default Options {#default-options}

The default options are optimized for safety and strictness:

```go
validator.DefaultOptions()
// Returns: Options{
//     StrictMissingFields: true,
//     ExtraFields:         ExtraIgnore,
//     TagName:             "",  // Uses global default "validate"
// }
```

---

## StrictMissingFields Option {#strict-missing-fields}

Controls whether missing fields (fields not present in JSON) cause validation errors.

### StrictMissingFields: true (Default) {#strict-missing-true}

Missing fields tagged `required` (and without a `default`) are validation errors. `StrictMissingFields: true` alone does not make every missing field an error — only fields that also carry the `required` constraint:

```go
type Config struct {
Host string `json:"host" validate:"required"`
Port int    `json:"port" validate:"required"` // No default, and tagged required
}

jsonData := []byte(`{"host":"localhost"}`)
configValidator := validator.New[Config](validator.Options{
StrictMissingFields: true,
})

config, err := configValidator.Unmarshal(jsonData)
// Error: port field is missing and has no default
```

Use this mode when you want to:

- Catch missing `required` fields as validation errors
- Ensure all `required` fields are explicitly provided in JSON
- Enforce strict API contracts

### StrictMissingFields: false {#strict-missing-false}

Missing fields without defaults are left as zero values:

```go
type Config struct {
Host string `json:"host" validate:"required"`
Port int    `json:"port"` // Zero value: 0
}

jsonData := []byte(`{"host":"localhost"}`)
configValidator := validator.New[Config](validator.Options{
StrictMissingFields: false,
})

config, err := configValidator.Unmarshal(jsonData)
// Success: Port is set to 0 (zero value)
fmt.Println(config.Port) // Output: 0
```

Use this mode when you want to:

- Allow partial JSON input (missing fields get zero values)
- Make most fields optional
- Support backwards-compatible API evolution

### How `required` and `json:",omitempty"` Interact {#required-omitempty}

Understanding how Go's `json:",omitempty"` marshaling tag interacts with Pedantigo's `required` constraint:

| Struct Tags                     | In JSON Schema `required` | Unmarshal Behavior               |
|---------------------------------|---------------------------|----------------------------------|
| `validate:"required"`          | Yes                       | Error if field missing from JSON |
| `json:",omitempty"` only        | No                        | Zero value if missing (valid)    |
| Both `required` + `omitempty`   | Yes                       | Error if missing (required wins) |
| Pointer `*T` + `required`       | Yes                       | Error if missing or null         |
| Pointer `*T` without `required` | No                        | nil if missing (valid)           |
| Non-pointer without `required`  | No                        | Zero value if missing            |

**Key Points:**

1. **`json:",omitempty"` is for marshaling output** - Controls whether zero-value fields are omitted when serialising a
   struct to JSON. It does not affect validation or schema generation.
2. **`required` controls both** - Affects the schema `required` array AND validation (field must be present in JSON
   input).
3. **Pointers enable true optionality** - Use `*T` when you need to distinguish "missing" from "zero".

```go
type Config struct {
// Required: must be in JSON, appears in schema required array
Host string `json:"host" validate:"required"`

// Optional with omitempty: omitted from output if zero, NOT in required array
Port int `json:"port,omitempty"`

// Optional pointer: nil if missing, distinguishes missing from zero
Timeout *int `json:"timeout,omitempty"`

// Required + omitempty: must be in JSON input, omitted from output if zero
Name string `json:"name,omitempty" validate:"required"`
}
```

:::tip
See [Schema Generation](/docs/concepts/schema) to understand how these tags affect the generated JSON Schema.
:::

---

### `omitempty` as a Pedantigo Validation Constraint {#pedantigo-omitempty}

In addition to Go's `json:",omitempty"` marshaling tag, Pedantigo supports `omitempty` as a **validation constraint** in
the `validate` struct tag itself.

```go
type SearchRequest struct {
ActorID   string `validate:"omitempty,max=64,required_with=ActorType"`
ActorType string `validate:"omitempty,oneof=user agent system,required_with=ActorID"`
}
```

**Semantics:**

When a field is tagged with `validate:"omitempty,..."` and its value is the **zero value** for its type (empty string,
`0`, `false`, nil pointer, etc.):

- Regular constraints (`min`, `max`, `oneof`, `email`, `url`, etc.) are **skipped** — the field is treated as
  intentionally absent.
- Cross-field constraints (`required_with`, `required_if`, `eqfield`, `nefield`, etc.) **always run**, regardless of
  zero-value status, to preserve relational validation semantics.

When the field has a non-zero value, all constraints run normally.

**Example — symmetric optional pair:**

```go
type Actor struct {
// Both are optional. If one is provided, the other must be too.
ActorID   string `validate:"omitempty,max=64,required_with=ActorType"`
ActorType string `validate:"omitempty,oneof=user agent system,required_with=ActorID"`
}
```

| ActorID     | ActorType   | Result                                                                               |
|-------------|-------------|--------------------------------------------------------------------------------------|
| `""`        | `""`        | Valid — both zero, regular constraints skipped                                       |
| `"user123"` | `"user"`    | Valid — both non-zero, all constraints run                                           |
| `"user123"` | `""`        | Error on `ActorType` — `required_with=ActorID` fires because `ActorID` is non-zero   |
| `""`        | `"user"`    | Error on `ActorID` — `required_with=ActorType` fires because `ActorType` is non-zero |
| `"user123"` | `"invalid"` | Error on `ActorType` — `oneof` fires because field is non-zero                       |

**Difference from `json:",omitempty"`:**

| Tag                     | Where it appears     | Effect                                                                      |
|-------------------------|----------------------|-----------------------------------------------------------------------------|
| `json:",omitempty"`     | JSON struct tag      | Controls marshaling output: omits the field from JSON if its value is zero  |
| `validate:"omitempty"` | Pedantigo constraint | Controls validation: skips regular constraints when the field value is zero |

Both can be combined:

```go
// Omitted from JSON output when zero AND skips validation when zero
Region string `json:"region,omitempty" validate:"omitempty,oneof=us-east-1 us-west-2 eu-west-1"`
```

---

## ExtraFields Option {#extra-fields}

Controls how unknown (extra) JSON fields are handled during unmarshaling.

### ExtraIgnore (Default) {#extra-ignore}

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

userValidator := validator.New[User](validator.Options{
ExtraFields: validator.ExtraIgnore,
})

user, err := userValidator.Unmarshal(jsonData)
// Success: age and phone fields are ignored
```

Use this mode when you want to:

- Accept JSON with additional fields (future-proofing)
- Ignore client-provided metadata
- Support flexible API clients

### ExtraForbid {#extra-forbid}

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

userValidator := validator.New[User](validator.Options{
ExtraFields: validator.ExtraForbid,
})

user, err := userValidator.Unmarshal(jsonData)
// Error: unknown field "age" in JSON
var validationErr *validator.ValidationError
if errors.As(err, &validationErr) {
fmt.Println(validationErr.Errors[0].Message)
// Output: unknown field in JSON
}
```

Use this mode when you want to:

- Prevent typos in client input
- Detect API misuse
- Enforce strict schema compliance

### ExtraAllow {#extra-allow}

Unknown JSON fields are captured and stored in a designated `map[string]any` field:

```go
type User struct {
Name   string         `json:"name"`
Email  string         `json:"email"`
Extras map[string]any `json:"-" validate:"extra_fields"`
}

jsonData := []byte(`{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 30,
    "phone": "555-1234"
}`)

userValidator := validator.New[User](validator.Options{
ExtraFields: validator.ExtraAllow,
})

user, err := userValidator.Unmarshal(jsonData)
// Success: unknown fields are captured
fmt.Println(user.Extras["age"]) // Output: 30
fmt.Println(user.Extras["phone"]) // Output: 555-1234
```

**Requirements:**

1. **Struct must have an extra_fields tagged field:**
   ```go
   Extras map[string]any `json:"-" validate:"extra_fields"`
   ```

2. **The field type must be `map[string]any`** (or `map[string]interface{}`)

3. **Fail-fast validation:** If `ExtraAllow` is set but no `extra_fields` field exists, `New[T]()` panics at startup

**Nested Struct Handling:**

Each struct level independently handles its own extras. If a nested struct has an `extra_fields` field, it captures
extras at that level:

```go
type Address struct {
City   string         `json:"city"`
Extras map[string]any `json:"-" validate:"extra_fields"`
}

type User struct {
Name    string         `json:"name"`
Address Address        `json:"address"`
Extras  map[string]any `json:"-" validate:"extra_fields"`
}

jsonData := []byte(`{
    "name": "Alice",
    "extra_top": "top-level extra",
    "address": {
        "city": "NYC",
        "extra_nested": "nested extra"
    }
}`)

// Top-level extras go to User.Extras
// Nested address extras go to Address.Extras
```

If a nested struct doesn't have an `extra_fields` field, extras at that level are silently ignored (but the top-level
still requires the field when `ExtraAllow` is set).

**Round-Trip Support:**

Extra fields are preserved during marshaling:

```go
userValidator := validator.New[User](validator.Options{
ExtraFields: validator.ExtraAllow,
})

user, _ := userValidator.Unmarshal(jsonData)
// user.Extras contains captured extras

roundTripJSON, _ := userValidator.Marshal(user)
// roundTripJSON includes both struct fields AND extras
```

---

## ExtraAllow: Real-World Use Cases {#extra-allow-use-cases}

### Use Case 1: Multi-Version API Support {#extra-allow-api-versioning}

When evolving APIs, newer clients may send fields that older server versions don't recognize. ExtraAllow preserves these
fields for:

- **Forward compatibility**: Accept fields from newer clients
- **Proxy/gateway scenarios**: Pass through unknown fields to downstream services
- **Gradual migration**: Log unknown fields to understand client adoption

```go
type UserV1 struct {
Name   string         `json:"name" validate:"required"`
Email  string         `json:"email" validate:"required,email"`
Extras map[string]any `json:"-" validate:"extra_fields"`
}

// Accept requests from V2 clients that include "profile_picture", "preferences", etc.
// These are captured in Extras for logging/forwarding without breaking V1 logic.
userV1Validator := validator.New[UserV1](validator.Options{
ExtraFields: validator.ExtraAllow,
})

user, _ := userV1Validator.Unmarshal(requestBody)
if len(user.Extras) > 0 {
log.Printf("Client sent unknown fields: %v", maps.Keys(user.Extras))
// Forward to downstream service that may understand these fields
}
```

**Further Reading:**

- [Go Blog: API Compatibility](https://go.dev/blog/module-compatibility)
- [Semantic Versioning Spec](https://semver.org/)

### Use Case 2: LLM Output Capture for Prompt Evaluation {#extra-allow-llm}

When using LLMs with structured output (JSON mode), models may include unexpected fields. ExtraAllow enables:

- **Prompt debugging**: Identify when models add unrequested fields
- **Model accuracy evaluation**: Track field adherence across prompts
- **Prompt iteration**: Use captured extras to refine instructions

```go
type LLMResponse struct {
Answer     string         `json:"answer" validate:"required"`
Confidence float64        `json:"confidence"`
Extras     map[string]any `json:"-" validate:"extra_fields"`
}

llmResponseValidator := validator.New[LLMResponse](validator.Options{
ExtraFields: validator.ExtraAllow,
})

// Parse LLM output
response, err := llmResponseValidator.Unmarshal(llmOutput)
if err != nil {
// Handle validation error (missing required fields, etc.)
}

// Track extra fields for prompt improvement
if len(response.Extras) > 0 {
log.Printf("LLM included unexpected fields: %v", response.Extras)
// Metrics for evaluating model accuracy
metrics.RecordExtraFields(modelName, promptID, response.Extras)
// Example: {"reasoning": "...", "sources": [...]} - fields LLM added on its own
}
```

**Further Reading:**

- [OpenAI: Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
- [Anthropic: Tool Use](https://docs.anthropic.com/en/docs/build-with-claude/tool-use)
- [JSON Schema Spec](https://json-schema.org/specification)

---

## TagName Option {#tag-name}

Override the struct tag name for a specific validator instance.

**Default**: Uses global tag name (`"validate"` by default, or whatever was set via `SetTagName()`)

| Value         | Behavior                              |
|---------------|----------------------------------------|
| `""` (empty)  | Use global default                     |
| `"validate"`  | Use go-playground/validator style      |
| `"binding"`   | Use gin-gonic style                    |
| `"pedantigo"` | Use the v1 default tag name explicitly |

```go
// Use go-playground/validator style tags
userValidator := validator.New[User](validator.Options{
TagName: "validate",
})

type User struct {
Name string `json:"name" validate:"required,min=3"`
}
```

**Use Cases:**

- Migration from other validation libraries
- Team conventions requiring specific tag names
- Coexistence with other libraries using `pedantigo` tag

**Further Reading:**

- [Go reflect.StructTag](https://pkg.go.dev/reflect#StructTag)

---

## Choosing the Right Options {#choosing-options}

| Scenario                       | StrictMissingFields | ExtraFields   | Reason                              |
|--------------------------------|---------------------|---------------|-------------------------------------|
| REST API with strict contracts | `true`              | `ExtraForbid` | Prevent bugs from typos/API misuse  |
| Configuration file parsing     | `false`             | `ExtraIgnore` | Allow partial configs, future-proof |
| During API migration           | `false`             | `ExtraIgnore` | Accept old and new client versions  |
| Public web API                 | `true`              | `ExtraIgnore` | Strict input, future-proof          |
| Internal service API           | `true`              | `ExtraForbid` | Strict both ways                    |
| Webhook receiver               | `false`             | `ExtraIgnore` | Accept any fields from sender       |
| LLM output parsing             | `true`              | `ExtraAllow`  | Capture unexpected model outputs    |
| API gateway/proxy              | `false`             | `ExtraAllow`  | Pass through unknown fields         |

## Option Combinations {#option-combinations}

### Most Strict

```go
validator.Options{
StrictMissingFields: true,
ExtraFields:         validator.ExtraForbid,
}
// Best for: REST API validation, ensuring exact schema match
```

### Most Lenient

```go
validator.Options{
StrictMissingFields: false,
ExtraFields:         validator.ExtraIgnore,
}
// Best for: Configuration parsing, webhook receivers, flexible APIs
```

### Balanced (Default)

```go
validator.DefaultOptions()
// StrictMissingFields: true
// ExtraFields:         ExtraIgnore
// Best for: General-purpose APIs, good balance of safety and flexibility
```

### Capture Everything

```go
validator.Options{
StrictMissingFields: false,
ExtraFields:         validator.ExtraAllow,
}
// Best for: LLM output parsing, API gateways, forward-compatible services
```

---

## Complete Examples {#complete-examples}

### API Server with Strict Validation

```go
package main

import (
	"errors"
	"log"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// API validator: strict about extra fields, requires all fields
var apiValidator = validator.New[CreateUserRequest](
	validator.Options{
		StrictMissingFields: true,
		ExtraFields:         validator.ExtraForbid,
	},
)

func handleCreateUser(jsonData []byte) (*CreateUserRequest, error) {
	req, err := apiValidator.Unmarshal(jsonData)
	if err != nil {
		var validationErr *validator.ValidationError
		if errors.As(err, &validationErr) {
			return nil, validationErr
		}
		return nil, err
	}
	return req, nil
}
```

### Configuration File Parsing (Lenient)

```go
type AppConfig struct {
Database struct {
Host     string `json:"host" validate:"required"`
Port     int    `json:"port" validate:"default=5432"`
Username string `json:"username" validate:"required"`
Password string `json:"password" validate:"required"`
} `json:"database"`
Server struct {
Addr string `json:"addr" validate:"default=0.0.0.0:8080"`
} `json:"server"`
}

// Config validator: lenient about missing fields, allows extra fields
var configValidator = validator.New[AppConfig](
validator.Options{
StrictMissingFields: false,
ExtraFields:         validator.ExtraIgnore,
},
)
```

### Migration-Friendly API

```go
type UserV2 struct {
ID       string `json:"id" validate:"required,uuid"`
Username string `json:"username" validate:"required"`
Email    string `json:"email" validate:"required,email"`
Status   string `json:"status" validate:"default=active"`
}

// Migration validator: accept both old and new client requests
var migrationValidator = validator.New[UserV2](
validator.Options{
StrictMissingFields: false,
ExtraFields:         validator.ExtraIgnore,
},
)
```

---

## Further Reading {#further-reading}

**Go Documentation:**

- [Go reflect.StructTag](https://pkg.go.dev/reflect#StructTag)
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [Go Blog: API Compatibility](https://go.dev/blog/module-compatibility)

**Specifications:**

- [JSON Schema Specification](https://json-schema.org/specification)
- [Semantic Versioning](https://semver.org/)

**LLM Structured Outputs:**

- [OpenAI: Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
- [Anthropic: Tool Use](https://docs.anthropic.com/en/docs/build-with-claude/tool-use)
