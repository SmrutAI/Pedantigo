# API Feature Parity

This document tracks Pedantigo's feature coverage compared to Pydantic v2 and go-playground/validator v10, including
JSON Schema standard support.

**VALIDATION BASICS**  

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Required fields              | √         | √           | √            | √                                  |
| Optional fields              | √         | √           | √            | √                                  |
| Default values (static)      | √         | √           | ×            | √                                  |
| Default values (dynamic)     | √         | √           | ×            | ×                                  |
| Field presence detection     | √         | √           | √            | √                                  |
| Zero vs missing distinction  | √         | √           | Partial      | ×                                  |

**STRING CONSTRAINTS**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Min/Max length               | √         | √           | √            | √                                  |
| Exact length                 | √         | Partial     | √            | √                                  |
| Email                        | √         | √           | √            | √                                  |
| URL                          | √         | √           | √            | √                                  |
| UUID                         | √         | √           | √            | √                                  |
| Regex/Pattern                | √         | √           | √            | √                                  |
| Enum/OneOf                   | √         | √           | √            | √                                  |
| Alpha/Alphanumeric           | √         | ×           | √            | √                                  |
| ASCII only                   | √         | ×           | √            | √                                  |
| Contains/Excludes            | √         | ×           | √            | √                                  |
| Starts/Ends with             | √         | ×           | √            | √                                  |
| Case validation              | √         | ×           | √            | ×                                  |
| Strip whitespace             | ×         | √           | ×            | ×                                  |
| String transform             | ×         | √           | ×            | ×                                  |

**NUMERIC CONSTRAINTS**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Min/Max value                | √         | √           | √            | √                                  |
| Greater/Less than            | √         | √           | √            | √                                  |
| Greater/Less or equal        | √         | √           | √            | √                                  |
| Multiple of                  | √         | √           | ×            | √                                  |                                    |
| Decimal precision            | √         | √           | ×            | ×                                  |                                    |
| Allow inf/nan                | √         | √           | ×            | ×                                  | Inverted: `disallow_inf_nan` (opt-in rejection) |
| Strict types                 | ×         | √           | ×            | ×                                  |                                    |
| Positive/Negative            | √         | √           | ×            | ×                                  |                                    |

**FORMAT VALIDATORS**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| IPv4/IPv6                    | √         | √           | √            | √                                  |
| IP (any)                     | ×         | √           | √            | √                                  |
| CIDR                         | ×         | √           | √            | √                                  |
| MAC address                  | ×         | ×           | √            | √                                  |
| Hostname                     | ×         | ×           | √            | √                                  |
| Port                         | ×         | ×           | √            | ×                                  |
| TCP/UDP address              | ×         | ×           | √            | ×                                  |
| Credit card                  | ×         | √           | √            | ×                                  |
| Bitcoin/Ethereum             | ×         | ×           | √            | ×                                  |
| ISBN/ISSN                    | ×         | ×           | √            | ×                                  |
| SSN/EIN                      | ×         | ×           | √            | ×                                  |
| Phone (E.164)                | ×         | Partial     | √            | ×                                  |
| Lat/Long                     | ×         | ×           | √            | ×                                  |
| Colors (hex, RGB, HSL)       | ×         | Partial     | √            | ×                                  |
| HTML                         | ×         | ×           | √            | ×                                  |
| JWT                          | ×         | Partial     | √            | ×                                  |
| JSON string                  | ×         | √           | √            | ×                                  |
| Hashes (MD5, SHA*)           | ×         | ×           | √            | ×                                  |
| Base64                       | ×         | √           | √            | √                                  |
| MongoDB ID                   | ×         | ×           | √            | ×                                  |
| Cron                         | ×         | ×           | √            | ×                                  |
| Semver                       | ×         | Partial     | √            | ×                                  |
| ULID                         | ×         | ×           | √            | ×                                  |
| Country codes                | ×         | Partial     | √            | ×                                  |
| Currency codes               | ×         | Partial     | √            | ×                                  |
| Language codes               | ×         | Partial     | √            | ×                                  |
| Postal codes                 | ×         | ×           | √            | ×                                  |

**COLLECTION VALIDATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Array/Slice min/max          | √         | √           | √            | √                                  |
| Element validation           | √         | √           | √            | √                                  |
| Map validation               | √         | √           | √            | √                                  |
| Map key validation           | Partial   | √           | √            | √                                  |
| Unique items                 | ×         | √           | √            | √                                  |
| Set types                    | ×         | √           | ×            | ×                                  |
| Tuple types                  | ×         | √           | ×            | √                                  |

**CROSS-FIELD VALIDATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Struct-level validators      | √         | √           | √            | ×                                  |
| Field comparisons            | √         | √           | √            | ×                                  |
| Cross-struct validation      | ×         | √           | √            | ×                                  |
| Conditional required         | √         | √           | √            | √                                  |
| Conditional exclusion        | √         | √           | √            | √                                  |
| Before validators            | ×         | √           | ×            | ×                                  |
| After validators             | √         | √           | ×            | ×                                  |
| Wrap validators              | ×         | √           | ×            | ×                                  |

**TYPE SUPPORT**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Primitives                   | √         | √           | √            | √                                  |
| Pointers/Optional            | √         | √           | √            | √                                  |
| Nested structs               | √         | √           | √            | √                                  |
| Slices/Lists                 | √         | √           | √            | √                                  |
| Maps/Dicts                   | √         | √           | √            | √                                  |
| time.Time/datetime           | √         | √           | Partial      | √                                  |
| time.Duration                | ×         | √           | ×            | √                                  |
| Secret types                 | ×         | √           | ×            | ×                                  |
| Path types                   | ×         | √           | Partial      | ×                                  |
| Literal types                | ×         | √           | ×            | √                                  |
| Union types                  | ×         | √           | ×            | √                                  |
| Discriminated unions         | ×         | √           | ×            | √                                  |
| Generic structs              | ×         | √           | ×            | ×                                  |
| Enum types                   | Partial   | √           | Partial      | √                                  |
| Decimal                      | ×         | √           | ×            | √                                  |

**JSON OPERATIONS**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Unmarshal + validate         | √         | √           | ×            | ×                                  |
| Marshal to JSON              | √         | √           | ×            | ×                                  |
| Marshal with field exclusion | ×         | √           | ×            | ×                                  |
| Marshal with field selection | ×         | √           | ×            | ×                                  |
| Marshal omitting zero values | ×         | √           | Partial      | ×                                  |
| Marshal using JSON tags      | Partial   | √           | √            | ×                                  |
| Custom MarshalJSON methods   | ×         | √           | √            | ×                                  |
| Streaming JSON               | ×         | ×           | ×            | ×                                  |
| Partial JSON repair          | ×         | ×           | ×            | ×                                  |

**SCHEMA GENERATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| JSON Schema                  | √         | √           | ×            | √                                  |
| OpenAPI ($ref)               | √         | √           | ×            | √                                  |
| Schema caching               | √         | √           | ×            | ×                                  |
| Schema examples              | ×         | √           | ×            | √                                  |
| Schema title                 | ×         | √           | ×            | √                                  |
| Field descriptions           | ×         | √           | ×            | √                                  |
| Deprecated fields            | ×         | √           | ×            | √                                  |

**FIELD CONFIGURATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| JSON tag aliases             | ×         | √           | √            | ×                                  |
| Validation-only aliases      | ×         | √           | ×            | ×                                  |
| Serialization-only aliases   | ×         | √           | ×            | ×                                  |
| Alias generator              | ×         | √           | ×            | ×                                  |
| Immutable fields             | ×         | √           | ×            | ×                                  |
| Computed fields              | ×         | √           | ×            | ×                                  |
| Discriminator field          | ×         | √           | ×            | √                                  |

**STRUCT CONFIGURATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Strict mode                  | Partial   | √           | ×            | ×                                  |
| Extra fields forbid          | ×         | √           | ×            | ×                                  |
| Extra fields allow           | ×         | √           | ×            | √                                  |
| Extra fields ignore          | √         | √           | ×            | ×                                  |
| Validate on assignment       | ×         | √           | ×            | ×                                  |
| Validate defaults            | √         | √           | ×            | ×                                  |
| ORM mode                     | ×         | √           | ×            | ×                                  |
| Arbitrary types              | ×         | √           | ×            | ×                                  |
| Immutable structs            | ×         | √           | ×            | ×                                  |

**ERROR HANDLING**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Multiple errors              | √         | √           | √            | ×                                  |
| Field paths                  | √         | √           | √            | ×                                  |
| Custom messages              | Partial   | √           | √            | ×                                  |
| Error codes                  | ×         | √           | ×            | ×                                  |
| i18n/l10n                    | ×         | Partial     | √            | ×                                  |
| Custom error types           | ×         | √           | ×            | ×                                  |

**CUSTOM VALIDATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Custom validators            | √         | √           | √            | ×                                  |
| Validator registration       | ×         | √           | √            | ×                                  |
| Alias tags                   | ×         | Partial     | √            | ×                                  |
| Validator context            | ×         | √           | √            | ×                                  |
| Struct-level                 | √         | √           | √            | ×                                  |
| Plugin system                | ×         | √           | ×            | ×                                  |

**ADVANCED FEATURES**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Type adapters                | ×         | √           | ×            | ×                                  |
| Root models                  | ×         | √           | ×            | ×                                  |
| Dataclass support            | ×         | √           | ×            | ×                                  |
| Config management            | ×         | √           | ×            | ×                                  |
| Environment variables        | ×         | √           | ×            | ×                                  |
| Struct copying               | ×         | √           | ×            | ×                                  |
| Struct field reflection      | ×         | √           | ×            | ×                                  |
| Recursive structs            | √         | √           | √            | √                                  |

**Summary**: 73/137 features implemented (53%)

**Legend**: √ = Supported, × = Not supported, Partial = Limited support

**JSON Schema Resources**:

- Specification: https://json-schema.org/specification
- Understanding JSON Schema: https://json-schema.org/understanding-json-schema/reference
- Latest Draft (2020-12): https://json-schema.org/draft/2020-12/schema
