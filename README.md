# Pedantigo

[![CI](https://github.com/SmrutAI/pedantigo/actions/workflows/ci.yml/badge.svg)](https://github.com/SmrutAI/pedantigo/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/tushar2708/67408111d1830ed523e2661d9ee2a442/raw/pedantigo-coverage.json)](https://github.com/SmrutAI/pedantigo)
[![Go Report Card](https://goreportcard.com/badge/github.com/SmrutAI/pedantigo)](https://goreportcard.com/report/github.com/SmrutAI/pedantigo)

Type-safe JSON validation and schema generation for Go.

> **v2 breaking change:** the default struct tag changed from `pedantigo` to `validate`. This is to be in sync with wider range of community tools like swaggo that work with `validate` tags
> If you're upgrading from v1, see the [migration guide](docs/migration/v1-to-v2.md) before you update.
>
> **New projects should start with v2.** v1 is maintained only for existing users — it receives no new features. All new feature development happens in v2.

## Installation

```bash
go get github.com/SmrutAI/pedantigo/v2/pdcore
```

Requires Go 1.21+

## Quick Example

```go
type User struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18"`
}

// Parse and validate JSON
user, err := pdcore.Unmarshal[User](jsonData)
if err != nil {
    // Handle validation errors with field paths and error codes
    if ve, ok := err.(*pdcore.ValidationError); ok {
        for _, fe := range ve.Errors {
            fmt.Printf("%s: %s\n", fe.Field, fe.Message)
        }
    }
}

// Generate JSON Schema for LLM tools (no $schema field)
schemaBytes, _ := pdcore.SchemaJSONLLM[User]()
```

## Features

| Feature | Description |
|---------|-------------|
| **100+ Constraints** | Email, URL, UUID, regex, numeric ranges, string length, and more |
| **JSON Schema Generation** | Auto-generate schemas for LLM tool calling and OpenAPI |
| **240x Caching Speedup** | Schema generation cached for high performance |
| **Streaming Validation** | Validate partial JSON from LLM streams |
| **Discriminated Unions** | Type-safe polymorphic data handling |
| **Cross-Field Validation** | Validate relationships between fields |
| **Zero Dependencies** | Only `invopop/jsonschema` + Go stdlib |

## Documentation

**Full documentation at [pedantigo.dev](https://pedantigo.dev)**

- [Getting Started](https://pedantigo.dev/docs/getting-started/quickstart) - Installation and first steps
- [Validation Concepts](https://pedantigo.dev/docs/concepts/validation) - How validation works
- [JSON Schema Generation](https://pedantigo.dev/docs/concepts/schema) - Generate schemas for LLM tools
- [Constraint Reference](https://pedantigo.dev/docs/concepts/constraints) - All 100+ validation rules
- [Streaming Validation](https://pedantigo.dev/docs/concepts/streaming) - Handle LLM streaming responses
- [API Reference](https://pedantigo.dev/docs/api/simple-api) - Complete API documentation
- [Benchmarks](https://github.com/smrutAI/pedantigo-benchmarks) - Performance comparisons with other validation libraries

## Use Cases

| Use Case | Why Pedantigo? |
|----------|----------------|
| **API Request Validation** | Validate incoming JSON, return structured errors |
| **LLM Structured Output** | Generate JSON Schema for function calling, validate responses |
| **Configuration Files** | Parse config with defaults, fail fast on invalid values |
| **Data Pipeline Input** | Ensure data quality at ingestion with detailed error paths |

## Feature Coverage

See [API_PARITY.md](API_PARITY.md) for detailed comparison with Pydantic v2 and go-playground/validator.

## License

MIT
