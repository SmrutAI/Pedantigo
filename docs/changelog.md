---
sidebar_position: 100
---

# Changelog

All notable changes to Pedantigo are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-12-18

### Added

#### Core Validation
- Struct tag-based validation with `pedantigo:"..."` tags
- 100+ built-in validation constraints
- Detailed error messages with field paths
- Support for nested structs and slices

#### Simple API (Recommended)
- `Unmarshal[T]()` - Parse JSON and validate in one call
- `NewModel[T]()` - Create validated instances from JSON, maps, or structs
- `Validate[T]()` - Validate existing struct instances
- `Schema[T]()` - Get cached JSON Schema
- `SchemaJSON[T]()` - Get JSON Schema as bytes
- `Marshal[T]()` - Validate and marshal to JSON
- `Dict[T]()` - Convert struct to map

#### Validator API (Advanced)
- `New[T]()` - Create validator with custom options
- `ValidatorOptions` - Configure strict mode, extra fields handling
- `MarshalOptions` - Context-based field exclusion

#### JSON Schema Generation
- Automatic schema generation from struct tags
- Built-in caching (240x speedup after first call)
- OpenAPI-compatible schema generation
- Thread-safe concurrent access

#### Streaming Validation
- `NewStreamParser[T]()` - Parse partial JSON from streams
- Real-time validation for LLM responses
- Automatic JSON repair for incomplete data

#### Discriminated Unions
- `NewUnion[T]()` - Type-safe union handling
- Multiple discriminator field support
- Automatic variant detection

#### Cross-Field Validation
- `eqfield`, `nefield` - Field equality comparisons
- `ltfield`, `lefield`, `gtfield`, `gefield` - Field numeric comparisons
- `required_if`, `required_unless` - Conditional requirements
- `Validatable` interface for custom cross-field logic

#### Advanced Features
- `SecretStr` / `SecretBytes` - Safe handling of sensitive data
- `RegisterValidation()` - Custom field validators
- `RegisterStructValidation[T]()` - Custom struct validators
- Context-based field exclusion for marshaling

### String Constraints
- `required`, `email`, `url`, `uuid`, `alpha`, `alphanumeric`, `numeric`
- `minLength`, `maxLength`, `startswith`, `endswith`, `contains`
- `pattern` (regex), `lowercase`, `uppercase`, `ascii`
- `base64`, `hexadecimal`, `json`, `jwt`, `semver`

### Numeric Constraints
- `min`, `max`, `gt`, `gte`, `lt`, `lte`
- `positive`, `negative`, `nonnegative`, `nonpositive`
- `multipleOf`, `divisibleBy`

### Format Constraints
- `datetime`, `date`, `time`, `duration`
- `ipv4`, `ipv6`, `cidr`, `mac`
- `latitude`, `longitude`
- `isbn`, `isbn10`, `isbn13`
- `credit_card`, `ssn`
- `country_code`, `currency_code`, `postal_code`

### Collection Constraints
- `minItems`, `maxItems`, `unique`
- `dive` - Validate slice/array elements
- `keys`, `values` - Validate map keys/values

---

## [Unreleased]

### Planned
- Additional format validators
- Performance improvements
- Extended OpenAPI support
