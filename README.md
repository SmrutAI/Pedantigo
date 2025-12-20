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

## Quick Example

```go
type User struct {
    Email string `json:"email" pedantigo:"required,email"`
    Age   int    `json:"age" pedantigo:"min=18"`
}

// Parse and validate JSON
user, err := pedantigo.Unmarshal[User](jsonData)
if err != nil {
    // Handle validation errors with field paths and error codes
    if ve, ok := err.(*pedantigo.ValidationError); ok {
        for _, fe := range ve.Errors {
            fmt.Printf("%s: %s\n", fe.Field, fe.Message)
        }
    }
}

// Generate JSON Schema (for LLM tools, OpenAPI, etc.)
schemaBytes, _ := pedantigo.SchemaJSON[User]()
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
