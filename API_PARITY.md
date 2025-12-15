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

| Feature          | Pedantigo | Pydantic v2 | Go Validator | JSON Schema? | Standard/RFC                 |
|------------------|-----------|-------------|--------------|--------------|------------------------------|
| IPv4/IPv6        | √         | √           | √            | √            | Go `net.ParseIP`             |
| IP (any)         | √         | √           | √            | √            | Go `net.ParseIP`             |
| CIDR             | √         | √           | √            | √            | RFC 4632                     |
| CIDRv4/CIDRv6    | √         | √           | √            | √            | RFC 4632                     |
| MAC address      | √         | ×           | √            | √            | IEEE 802                     |
| Hostname         | √         | ×           | √            | √            | RFC 952                      |
| Hostname RFC1123 | √         | ×           | √            | √            | RFC 1123                     |
| FQDN             | √         | ×           | √            | √            | DNS standard                 |
| Port             | √         | ×           | √            | ×            | 0-65535                      |
| TCP/UDP address  | √         | ×           | √            | ×            | `net.ResolveTCPAddr`         |
| Credit card      | √         | √           | √            | ×            | ISO/IEC 7812 (Luhn)          |
| Bitcoin address  | √         | ×           | √            | ×            | P2PKH/P2SH, Base58Check      |
| Bitcoin Bech32   | √         | ×           | √            | ×            | BIP-0173                     |
| Ethereum address | √         | ×           | √            | ×            | EIP-55                       |
| ISBN             | √         | ×           | √            | ×            | ISO 2108                     |
| ISBN-10/ISBN-13  | √         | ×           | √            | ×            | ISO 2108                     |
| ISSN             | √         | ×           | √            | ×            | ISO 3297                     |
| SSN              | √         | ×           | √            | ×            | U.S. SSA format              |
| EIN              | √         | ×           | √            | ×            | U.S. IRS format              |
| Phone (E.164)    | √         | Partial     | √            | ×            | ITU-T E.164                  |
| Latitude         | √         | ×           | √            | ×            | WGS 84 (-90 to +90)          |
| Longitude        | √         | ×           | √            | ×            | WGS 84 (-180 to +180)        |
| Hex color        | √         | Partial     | √            | ×            | CSS Color spec               |
| RGB/RGBA         | √         | Partial     | √            | ×            | CSS Color spec               |
| HSL/HSLA         | √         | Partial     | √            | ×            | CSS Color spec               |
| HTML             | √         | ×           | √            | ×            | HTML5                        |
| JWT              | √         | Partial     | √            | ×            | RFC 7519                     |
| JSON string      | √         | √           | √            | ×            | RFC 8259                     |
| Base64           | √         | √           | √            | √            | RFC 4648                     |
| Base64URL        | √         | √           | √            | √            | RFC 4648 §5                  |
| Base64RawURL     | √         | ×           | √            | √            | RFC 4648 §3.2                |
| MD4              | √         | ×           | √            | ×            | RFC 1320                     |
| MD5              | √         | ×           | √            | ×            | RFC 1321                     |
| SHA256/384/512   | √         | ×           | √            | ×            | FIPS 180-4                   |
| MongoDB ID       | √         | ×           | √            | ×            | MongoDB ObjectId             |
| Cron             | √         | ×           | √            | ×            | Cron expression              |
| Semver           | √         | Partial     | √            | ×            | Semver 2.0.0                 |
| ULID             | √         | ×           | √            | ×            | ULID spec (Crockford base32) |
| Luhn checksum    | √         | ×           | √            | ×            | ISO/IEC 7812                 |
| Country codes    | ×         | Partial     | √            | ×            | ISO 3166-1                   |
| Currency codes   | ×         | Partial     | √            | ×            | ISO 4217                     |
| Language codes   | ×         | Partial     | √            | ×            | BCP 47                       |
| Postal codes     | ×         | ×           | √            | ×            | ISO 3166 per-country         |

**COLLECTION VALIDATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Array/Slice min/max          | √         | √           | √            | √                                  |
| Element validation (dive)    | √         | √           | √            | √                                  |
| Map validation               | √         | √           | √            | √                                  |
| Map key validation (keys)    | √         | √           | √            | √                                  |
| Unique items                 | √         | √           | √            | √                                  |
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
| Union types                  | √         | √           | ×            | √                                  |
| Discriminated unions         | √         | √           | ×            | √                                  |
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
| Discriminator field          | √         | √           | ×            | √                                  |

**STRUCT CONFIGURATION**

| Feature                      | Pedantigo | Pydantic v2 | Go Validator | Supported by JSON Schema standard? |
|------------------------------|-----------|-------------|--------------|------------------------------------|
| Strict mode                  | Partial   | √           | ×            | ×                                  |
| Extra fields forbid          | √         | √           | ×            | ×                                  |
| Extra fields allow           | √         | √           | ×            | √                                  |
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

**Summary**: 119/141 features implemented (84%)

**Legend**: √ = Supported, × = Not supported, Partial = Limited support

**JSON Schema Resources**:

- Specification: https://json-schema.org/specification
- Understanding JSON Schema: https://json-schema.org/understanding-json-schema/reference
- Latest Draft (2020-12): https://json-schema.org/draft/2020-12/schema
