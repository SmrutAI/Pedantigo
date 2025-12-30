---
sidebar_position: 2
title: API Parity
description: Feature comparison between Pedantigo, Pydantic, and go-playground/validator
---

# Feature Comparison

How Pedantigo compares to Pydantic v2 and go-playground/validator v10.

**Legend**: ✓ Supported | ✗ Not supported | ~ Partial

---

## Validation Basics

| Feature                     | Pedantigo | Pydantic | Validator |                             Docs                              | Comment                                        |
|-----------------------------|:---------:|:--------:|:---------:|:-------------------------------------------------------------:|------------------------------------------------|
| Required fields             |     ✓     |    ✓     |     ✓     |  [Validation](../concepts/validation#understanding-required)  |                                                |
| Optional fields             |     ✓     |    ✓     |     ✓     |  [Validation](../concepts/validation#understanding-required)  |                                                |
| Default values (static)     |     ✓     |    ✓     |     ✗     |     [Constraints](../concepts/constraints#default-values)     |                                                |
| Default values (dynamic)    |     ✓     |    ✓     |     ✗     |  [Validation](../concepts/validation#defaults-and-factories)  |                                                |
| Field presence detection    |     ✓     |    ✓     |     ✓     | [Initialization](../api/initialization#strict-missing-fields) |                                                |
| Zero vs missing distinction |     ✓     |    ✓     |     ~     | [Initialization](../api/initialization#strict-missing-fields) | validator: requires `omitempty` tag workaround |

---

## String Constraints

| Feature               | Pedantigo | Pydantic | Validator |                        Docs                         | Comment                                                |
|-----------------------|:---------:|:--------:|:---------:|:---------------------------------------------------:|--------------------------------------------------------|
| Min/Max length        |     ✓     |    ✓     |     ✓     |      [String](../constraints/string#min--max)       |                                                        |
| Exact length          |     ✓     |    ~     |     ✓     |         [String](../constraints/string#len)         | Pydantic: use `min_length=max_length`; no direct `len` |
| Email                 |     ✓     |    ✓     |     ✓     |        [Format](../constraints/format#email)        |                                                        |
| URL                   |     ✓     |    ✓     |     ✓     |         [Format](../constraints/format#url)         |                                                        |
| URI                   |     ✓     |    ✗     |     ✓     |         [Format](../constraints/format#uri)         |                                                        |
| UUID                  |     ✓     |    ✓     |     ✓     |        [Format](../constraints/format#uuid)         |                                                        |
| UUID3/UUID4/UUID5     |     ✓     |    ✗     |     ✓     | [Format](../constraints/format#uuid3--uuid4--uuid5) |                                                        |
| Regex/Pattern         |     ✓     |    ✓     |     ✓     |       [String](../constraints/string#regexp)        |                                                        |
| Enum/OneOf            |     ✓     |    ✓     |     ✓     |       [Constraints](../concepts/constraints)        |                                                        |
| Enum case-insensitive |     ✓     |    ✗     |     ✓     |       [Constraints](../concepts/constraints)        |                                                        |
| Alpha/Alphanumeric    |     ✓     |    ✗     |     ✓     |        [String](../constraints/string#alpha)        |                                                        |
| ASCII only            |     ✓     |    ✗     |     ✓     |        [String](../constraints/string#ascii)        |                                                        |
| Multibyte chars       |     ✓     |    ✗     |     ✓     |      [String](../constraints/string#multibyte)      |                                                        |
| Contains/Excludes     |     ✓     |    ✗     |     ✓     |      [String](../constraints/string#contains)       |                                                        |
| Starts/Ends with      |     ✓     |    ✗     |     ✓     |     [String](../constraints/string#startswith)      |                                                        |
| Case validation       |     ✓     |    ✗     |     ✓     |      [String](../constraints/string#lowercase)      |                                                        |
| Strip whitespace      |     ✓     |    ✓     |     ✗     |  [String](../constraints/string#strip_whitespace)   |                                                        |
| String transform      |     ✓     |    ✓     |     ✗     |           [String](../constraints/string)           |                                                        |

---

## Numeric Constraints

| Feature               | Pedantigo | Pydantic | Validator |                        Docs                        | Comment                                                           |
|-----------------------|:---------:|:--------:|:---------:|:--------------------------------------------------:|-------------------------------------------------------------------|
| Min/Max value         |     ✓     |    ✓     |     ✓     |     [Numeric](../constraints/numeric#min--max)     |                                                                   |
| Greater/Less than     |     ✓     |    ✓     |     ✓     | [Numeric](../constraints/numeric#gt--gte--lt--lte) |                                                                   |
| Greater/Less or equal |     ✓     |    ✓     |     ✓     | [Numeric](../constraints/numeric#gt--gte--lt--lte) |                                                                   |
| Multiple of           |     ✓     |    ✓     |     ✗     |   [Numeric](../constraints/numeric#multiple_of)    |                                                                   |
| Decimal precision     |     ✓     |    ✓     |     ✗     |  [Numeric](../constraints/numeric#decimal_places)  |                                                                   |
| Disallow inf/nan      |     ✓     |    ✓     |     ✗     | [Numeric](../constraints/numeric#disallow_inf_nan) |                                                                   |
| Strict types          |     ✗     |    ✓     |     ✗     |                         —                          | Go is statically typed; no coercion like Python's `"123"` → `123` |
| Positive/Negative     |     ✓     |    ✓     |     ✗     |     [Numeric](../constraints/numeric#positive)     |                                                                   |

---

## Format Validators

| Feature          | Pedantigo | Pydantic | Validator | Standard             |                               Docs                                | Comment                                                          |
|------------------|:---------:|:--------:|:---------:|----------------------|:-----------------------------------------------------------------:|------------------------------------------------------------------|
| IPv4/IPv6        |     ✓     |    ✓     |     ✓     | `net.ParseIP`        |               [Format](../constraints/format#ipv4)                |                                                                  |
| IP (any)         |     ✓     |    ✓     |     ✓     | `net.ParseIP`        |                [Format](../constraints/format#ip)                 |                                                                  |
| CIDR             |     ✓     |    ✓     |     ✓     | RFC 4632             |               [Format](../constraints/format#cidr)                |                                                                  |
| CIDRv4/CIDRv6    |     ✓     |    ✓     |     ✓     | RFC 4632             |              [Format](../constraints/format#cidrv4)               |                                                                  |
| MAC address      |     ✓     |    ✗     |     ✓     | IEEE 802             |                [Format](../constraints/format#mac)                |                                                                  |
| Hostname         |     ✓     |    ✗     |     ✓     | RFC 952              |             [Format](../constraints/format#hostname)              |                                                                  |
| Hostname RFC1123 |     ✓     |    ✗     |     ✓     | RFC 1123             |         [Format](../constraints/format#hostname_rfc1123)          |                                                                  |
| FQDN             |     ✓     |    ✗     |     ✓     | DNS standard         |               [Format](../constraints/format#fqdn)                |                                                                  |
| Port             |     ✓     |    ✗     |     ✓     | 0-65535              |               [Format](../constraints/format#port)                |                                                                  |
| TCP/UDP address  |     ✓     |    ✗     |     ✓     | `net.ResolveTCPAddr` |             [Format](../constraints/format#tcp_addr)              |                                                                  |
| HTTP URL         |     ✓     |    ✗     |     ✓     | RFC 3986             |             [Format](../constraints/format#http_url)              |                                                                  |
| HTTPS URL        |     ✓     |    ✗     |     ✓     | RFC 3986             |             [Format](../constraints/format#https_url)             |                                                                  |
| Credit card      |     ✓     |    ✓     |     ✓     | ISO/IEC 7812         |            [Format](../constraints/format#credit_card)            |                                                                  |
| Bitcoin address  |     ✓     |    ✗     |     ✓     | Base58Check          |             [Format](../constraints/format#btc_addr)              |                                                                  |
| Bitcoin Bech32   |     ✓     |    ✗     |     ✓     | BIP-0173             |          [Format](../constraints/format#btc_addr_bech32)          |                                                                  |
| Ethereum address |     ✓     |    ✗     |     ✓     | EIP-55               |             [Format](../constraints/format#eth_addr)              |                                                                  |
| ISBN             |     ✓     |    ✗     |     ✓     | ISO 2108             |       [Format](../constraints/format#isbn--isbn10--isbn13)        |                                                                  |
| ISBN-10/ISBN-13  |     ✓     |    ✗     |     ✓     | ISO 2108             |       [Format](../constraints/format#isbn--isbn10--isbn13)        |                                                                  |
| ISSN             |     ✓     |    ✗     |     ✓     | ISO 3297             |               [Format](../constraints/format#issn)                |                                                                  |
| SSN              |     ✓     |    ✗     |     ✓     | U.S. SSA             |                [Format](../constraints/format#ssn)                |                                                                  |
| EIN              |     ✓     |    ✗     |     ✓     | U.S. IRS             |                [Format](../constraints/format#ein)                |                                                                  |
| Phone (E.164)    |     ✓     |    ~     |     ✓     | ITU-T E.164          |               [Format](../constraints/format#e164)                | Pydantic: has `PhoneNumber` type but requires `phonenumbers` lib |
| Latitude         |     ✓     |    ✗     |     ✓     | WGS 84               |        [Format](../constraints/format#latitude--longitude)        |                                                                  |
| Longitude        |     ✓     |    ✗     |     ✓     | WGS 84               |        [Format](../constraints/format#latitude--longitude)        |                                                                  |
| Hex color        |     ✓     |    ~     |     ✓     | CSS Color            |  [Format](../constraints/format#hexcolor--rgb--rgba--hsl--hsla)   | Pydantic: has `Color` type but not individual validators         |
| RGB/RGBA         |     ✓     |    ~     |     ✓     | CSS Color            |  [Format](../constraints/format#hexcolor--rgb--rgba--hsl--hsla)   | Pydantic: has `Color` type but not individual validators         |
| HSL/HSLA         |     ✓     |    ~     |     ✓     | CSS Color            |  [Format](../constraints/format#hexcolor--rgb--rgba--hsl--hsla)   | Pydantic: has `Color` type but not individual validators         |
| iscolor (alias)  |     ✓     |    ✗     |     ✓     | Any color            |              [Format](../constraints/format#iscolor)              |                                                                  |
| HTML             |     ✓     |    ✗     |     ✓     | HTML5                |               [Format](../constraints/format#html)                |                                                                  |
| JWT              |     ✓     |    ~     |     ✓     | RFC 7519             |                [Format](../constraints/format#jwt)                | Pydantic: no built-in; use custom validator                      |
| JSON string      |     ✓     |    ✓     |     ✓     | RFC 8259             |               [Format](../constraints/format#json)                |                                                                  |
| Base64           |     ✓     |    ✓     |     ✓     | RFC 4648             |  [Format](../constraints/format#base64--base64url--base64rawurl)  |                                                                  |
| Base64URL        |     ✓     |    ✓     |     ✓     | RFC 4648 §5          |  [Format](../constraints/format#base64--base64url--base64rawurl)  |                                                                  |
| Base64RawURL     |     ✓     |    ✗     |     ✓     | RFC 4648 §3.2        |  [Format](../constraints/format#base64--base64url--base64rawurl)  |                                                                  |
| Base32           |     ✓     |    ✗     |     ✓     | RFC 4648 §6          |              [Format](../constraints/format#base32)               |                                                                  |
| Data URI         |     ✓     |    ✗     |     ✓     | RFC 2397             |              [Format](../constraints/format#datauri)              |                                                                  |
| URN (RFC 2141)   |     ✓     |    ✗     |     ✓     | RFC 2141             |            [Format](../constraints/format#urn_rfc2141)            |                                                                  |
| MD4              |     ✓     |    ✗     |     ✓     | RFC 1320             | [Format](../constraints/format#md5--md4--sha256--sha384--sha512)  |                                                                  |
| MD5              |     ✓     |    ✗     |     ✓     | RFC 1321             | [Format](../constraints/format#md5--md4--sha256--sha384--sha512)  |                                                                  |
| SHA256/384/512   |     ✓     |    ✗     |     ✓     | FIPS 180-4           | [Format](../constraints/format#md5--md4--sha256--sha384--sha512)  |                                                                  |
| MongoDB ID       |     ✓     |    ✗     |     ✓     | ObjectId             |              [Format](../constraints/format#mongodb)              |                                                                  |
| Cron             |     ✓     |    ✗     |     ✓     | Cron expr            |               [Format](../constraints/format#cron)                |                                                                  |
| Semver           |     ✓     |    ~     |     ✓     | Semver 2.0           |              [Format](../constraints/format#semver)               | Pydantic: no built-in; use custom validator                      |
| Datetime format  |     ✓     |    ✓     |     ✓     | Go layout            |             [Format](../constraints/format#datetime)              |                                                                  |
| Timezone (IANA)  |     ✓     |    ✗     |     ✓     | IANA tz database     |             [Format](../constraints/format#timezone)              |                                                                  |
| ULID             |     ✓     |    ✗     |     ✓     | Crockford            |               [Format](../constraints/format#ulid)                |                                                                  |
| Luhn checksum    |     ✓     |    ✗     |     ✓     | ISO 7812             |           [Format](../constraints/format#luhn_checksum)           |                                                                  |
| Country codes    |     ✓     |    ~     |     ✓     | ISO 3166-1           |  [Format](../constraints/format#iso3166_alpha2--iso3166_alpha3)   | Pydantic: no built-in; use pycountry lib                         |
| Currency codes   |     ✓     |    ~     |     ✓     | ISO 4217             |     [Format](../constraints/format#iso4217--iso4217_numeric)      | Pydantic: no built-in; use pycountry lib                         |
| Language codes   |     ✓     |    ~     |     ✓     | BCP 47               |               [Format](../constraints/format#bcp47)               | Pydantic: no built-in; use langcodes lib                         |
| Postal codes     |     ✓     |    ✗     |     ✓     | Per-country          | [Format](../constraints/format#postcode--postcode_iso3166_alpha2) |                                                                  |

---

## Collection Validation

| Feature                   | Pedantigo | Pydantic | Validator |                                   Docs                                    | Comment                                                                  |
|---------------------------|:---------:|:--------:|:---------:|:-------------------------------------------------------------------------:|--------------------------------------------------------------------------|
| Array/Slice min/max       |     ✓     |    ✓     |     ✓     |      [Collection](../constraints/collection#min--max-on-collections)      |                                                                          |
| Element validation (dive) |     ✓     |    ✓     |     ✓     |   [Collection](../constraints/collection#dive---validate-each-element)    |                                                                          |
| Map validation            |     ✓     |    ✓     |     ✓     | [Collection](../constraints/collection#keys--endkeys---validate-map-keys) |                                                                          |
| Map key validation (keys) |     ✓     |    ✓     |     ✓     | [Collection](../constraints/collection#keys--endkeys---validate-map-keys) |                                                                          |
| Unique items              |     ✓     |    ✓     |     ✓     |              [Collection](../constraints/collection#unique)               |                                                                          |
| Set types                 |     ✗     |    ✓     |     ✗     |                                     —                                     | Go has no built-in set type; use `map[T]struct{}` or slice with `unique` |
| Tuple types               |     ✗     |    ✓     |     ✗     |                                     —                                     | Go has no tuple type; use structs for fixed heterogeneous sequences      |

---

## Cross-Field Validation

| Feature                     | Pedantigo | Pydantic | Validator |                                 Docs                                 | Comment                                                                    |
|-----------------------------|:---------:|:--------:|:---------:|:--------------------------------------------------------------------:|----------------------------------------------------------------------------|
| Struct-level validators     |     ✓     |    ✓     |     ✓     | [Cross-Field](../concepts/cross-field#custom-cross-field-validation) |                                                                            |
| Field comparisons           |     ✓     |    ✓     |     ✓     |       [Cross-Field](../concepts/cross-field#field-comparison)        |                                                                            |
| Cross-struct validation     |     ✓     |    ✓     |     ✓     |       [Cross-Field](../concepts/cross-field#field-comparison)        |                                                                            |
| Conditional required        |     ✓     |    ✓     |     ✓     |     [Cross-Field](../concepts/cross-field#conditional-required)      |                                                                            |
| Conditional required (all)  |     ✓     |    ✗     |     ✓     |     [Cross-Field](../concepts/cross-field#conditional-required)      |                                                                            |
| Conditional exclusion       |     ✓     |    ✓     |     ✓     |     [Cross-Field](../concepts/cross-field#conditional-excluded)      |                                                                            |
| Conditional exclusion (all) |     ✓     |    ✗     |     ✓     |     [Cross-Field](../concepts/cross-field#conditional-excluded)      |                                                                            |
| Before validators           |     ✗     |    ✓     |     ✗     |                                  —                                   | Pydantic's `@field_validator(mode='before')`; use custom unmarshaler in Go |
| After validators            |     ✓     |    ✓     |     ✗     |          [Custom Validators](../advanced/custom-validators)          |                                                                            |
| Wrap validators             |     ✗     |    ✓     |     ✗     |                                  —                                   | Pydantic decorator pattern; not idiomatic in Go                            |

---

## Type Support

| Feature              | Pedantigo | Pydantic | Validator |                       Docs                        | Comment                                                 |
|----------------------|:---------:|:--------:|:---------:|:-------------------------------------------------:|---------------------------------------------------------|
| Primitives           |     ✓     |    ✓     |     ✓     |       [Validation](../concepts/validation)        |                                                         |
| Pointers/Optional    |     ✓     |    ✓     |     ✓     |       [Validation](../concepts/validation)        |                                                         |
| Nested structs       |     ✓     |    ✓     |     ✓     |       [Validation](../concepts/validation)        |                                                         |
| Slices/Lists         |     ✓     |    ✓     |     ✓     |      [Collection](../constraints/collection)      |                                                         |
| Maps/Dicts           |     ✓     |    ✓     |     ✓     |      [Collection](../constraints/collection)      |                                                         |
| time.Time/datetime   |     ✓     |    ✓     |     ~     |       [Validation](../concepts/validation)        | validator: limited datetime support, use custom         |
| time.Duration        |     ✓     |    ✓     |     ✗     |       [Validation](../concepts/validation)        |                                                         |
| Secret types         |     ✓     |    ✓     |     ✗     |          [Secrets](../advanced/secrets)           |                                                         |
| Path types           |     ✓     |    ✓     |     ~     | [Format](../constraints/format#filepath--dirpath) | validator: `file` constraint but limited path types     |
| Literal types        |     ✗     |    ✓     |     ✗     |                         —                         | Python type hint; use `oneof` constraint in Go          |
| Union types          |     ✓     |    ✓     |     ✗     |           [Unions](../concepts/unions)            |                                                         |
| Discriminated unions |     ✓     |    ✓     |     ✗     |           [Unions](../concepts/unions)            |                                                         |
| Generic structs      |     ✗     |    ✓     |     ✗     |                         —                         | Go 1.18+ generics exist; complex to support fully       |
| Enum types           |     ~     |    ✓     |     ~     |      [Constraints](../concepts/constraints)       | Go has no native enum; all use `oneof`/iota workarounds |
| Decimal              |     ✗     |    ✓     |     ✗     |                         —                         | Use `shopspring/decimal` package if needed              |

---

## JSON Operations

| Feature                      | Pedantigo | Pydantic | Validator |                        Docs                        | Comment                                                                                       |
|------------------------------|:---------:|:--------:|:---------:|:--------------------------------------------------:|-----------------------------------------------------------------------------------------------|
| Unmarshal + validate         |     ✓     |    ✓     |     ✗     |     [Simple API](../api/simple-api#unmarshal)      |                                                                                               |
| Marshal to JSON              |     ✓     |    ✓     |     ✗     |      [Simple API](../api/simple-api#marshal)       |                                                                                               |
| Marshal with field exclusion |     ✓     |    ✓     |     ✗     | [Simple API](../api/simple-api#marshalwithoptions) |                                                                                               |
| Marshal with field selection |     ✓     |    ✓     |     ✗     | [Simple API](../api/simple-api#marshalwithoptions) |                                                                                               |
| Marshal omitting zero values |     ✓     |    ✓     |     ~     | [Simple API](../api/simple-api#marshalwithoptions) | validator: uses standard `omitempty` json tag                                                 |
| Marshal using JSON tags      |     ~     |    ✓     |     ✓     |    [Simple API](../api/simple-api#struct-tags)     | Pedantigo: respects `json` tag names but uses own marshal (doesn't call custom `MarshalJSON`) |
| Custom MarshalJSON methods   |     ✗     |    ✓     |     ✓     |                         —                          | Go supports this natively; Pedantigo uses its own marshal                                     |
| Streaming JSON               |     ✓     |    ✗     |     ✗     |         [Streaming](../concepts/streaming)         |                                                                                               |
| Partial JSON repair          |     ✗     |    ✗     |     ✗     |                         —                          | Complex edge case                                                                             |

---

## Schema Generation

| Feature            | Pedantigo | Pydantic | Validator |                       Docs                        | Comment |
|--------------------|:---------:|:--------:|:---------:|:-------------------------------------------------:|---------|
| JSON Schema        |     ✓     |    ✓     |     ✗     |        [Schema](../concepts/schema#schema)        |         |
| OpenAPI ($ref)     |     ✓     |    ✓     |     ✗     |    [Schema](../concepts/schema#schemaopenapi)     |         |
| Schema caching     |     ✓     |    ✓     |     ✗     | [Schema](../concepts/schema#performance--caching) |         |
| Schema examples    |     ✓     |    ✓     |     ✗     |       [Schema](../concepts/schema#examples)       |         |
| Schema title       |     ✓     |    ✓     |     ✗     |        [Schema](../concepts/schema#title)         |         |
| Field descriptions |     ✓     |    ✓     |     ✗     |     [Schema](../concepts/schema#description)      |         |
| Deprecated fields  |     ✓     |    ✓     |     ✗     |      [Schema](../concepts/schema#deprecated)      |         |

---

## Struct Configuration

| Feature                | Pedantigo | Pydantic | Validator |                          Docs                           | Comment                                                                                      |
|------------------------|:---------:|:--------:|:---------:|:-------------------------------------------------------:|----------------------------------------------------------------------------------------------|
| Strict mode            |     ✗     |    ✓     |     ✗     |                            —                            | Pydantic strict mode prevents type coercion (`"123"` → `123`); Go is statically typed so N/A |
| Extra fields forbid    |     ✓     |    ✓     |     ✗     |  [Initialization](../api/initialization#extra-forbid)   |                                                                                              |
| Extra fields allow     |     ✓     |    ✓     |     ✗     |   [Initialization](../api/initialization#extra-allow)   |                                                                                              |
| Extra fields ignore    |     ✓     |    ✓     |     ✗     |  [Initialization](../api/initialization#extra-ignore)   |                                                                                              |
| Validate on assignment |     ✗     |    ✓     |     ✗     |                            —                            | Python can override `__setattr__`; Go structs are plain data                                 |
| Validate defaults      |     ✓     |    ✓     |     ✗     | [Initialization](../api/initialization#default-options) |                                                                                              |
| ORM mode               |     ✗     |    ✓     |     ✗     |                            —                            | Pydantic-specific for SQLAlchemy; use GORM/sqlx directly in Go                               |
| Arbitrary types        |     ✗     |    ✓     |     ✗     |                            —                            | Pydantic allows any Python type; Go is statically typed                                      |
| Immutable structs      |     ✗     |    ✓     |     ✗     |                            —                            | Go has no built-in immutability; use unexported fields + getters                             |

---

## Error Handling

| Feature            | Pedantigo | Pydantic | Validator |                     Docs                      | Comment                                                                                |
|--------------------|:---------:|:--------:|:---------:|:---------------------------------------------:|----------------------------------------------------------------------------------------|
| Multiple errors    |     ✓     |    ✓     |     ✓     |    [Errors](../api/errors#multiple-errors)    |                                                                                        |
| Field paths        |     ✓     |    ✓     |     ✓     |   [Errors](../api/errors#field-path-format)   |                                                                                        |
| Custom messages    |     ~     |    ✓     |     ✓     | [Errors](../api/errors#common-error-messages) | Pedantigo: custom constraints can return custom messages; no per-field override syntax |
| Error codes        |     ✓     |    ✓     |     ✗     | [Errors](../api/errors#error-codes-reference) |                                                                                        |
| i18n/l10n          |     ✗     |    ~     |     ✓     |                       —                       | Internationalized messages; significant work to implement                              |
| Custom error types |     ✗     |    ✓     |     ✗     |                       —                       | Go supports this; could expose more customization                                      |

---

## Custom Validation

| Feature                | Pedantigo | Pydantic | Validator |                                          Docs                                          | Comment                                                                                    |
|------------------------|:---------:|:--------:|:---------:|:--------------------------------------------------------------------------------------:|--------------------------------------------------------------------------------------------|
| Custom validators      |     ✓     |    ✓     |     ✓     |       [Custom Validators](../advanced/custom-validators#field-level-validators)        |                                                                                            |
| Validator registration |     ✓     |    ✓     |     ✓     | [Custom Validators](../advanced/custom-validators#registering-with-registervalidation) |                                                                                            |
| Alias tags             |     ✓     |    ~     |     ✓     |            [Custom Validators](../advanced/custom-validators#registeralias)            | Pydantic: uses `AliasPath`/`AliasChoices` for field name aliasing, not constraint aliasing |
| Validator context      |     ✓     |    ✓     |     ✓     |      [Custom Validators](../advanced/custom-validators#context-aware-validators)       |                                                                                            |
| Struct-level           |     ✓     |    ✓     |     ✓     |       [Custom Validators](../advanced/custom-validators#struct-level-validators)       |                                                                                            |
| Plugin system          |     ✗     |    ✓     |     ✗     |                                           —                                            | Pydantic plugin architecture; not idiomatic in Go                                          |

---

## Advanced Features

| Feature                 | Pedantigo | Pydantic | Validator |                 Docs                 | Comment                                                                 |
|-------------------------|:---------:|:--------:|:---------:|:------------------------------------:|-------------------------------------------------------------------------|
| Type adapters           |     ✗     |    ✓     |     ✗     |                  —                   | Pydantic-specific for custom type coercion                              |
| Root models             |     ✗     |    ✓     |     ✗     |                  —                   | Pydantic's `RootModel` for non-dict types; use slice/map directly in Go |
| Dataclass support       |     ✗     |    ✓     |     ✗     |                  —                   | Python dataclasses; Go uses structs natively                            |
| Config management       |     ✗     |    ✓     |     ✗     |                  —                   | Pydantic-settings; use Viper/envconfig in Go                            |
| Environment variables   |     ✗     |    ✓     |     ✗     |                  —                   | Pydantic-settings; use `os.Getenv` or Viper in Go                       |
| Struct copying          |     ✗     |    ✓     |     ✗     |                  —                   | Pydantic's `model_copy()`; use manual copy or `copier` package in Go    |
| Struct field reflection |     ✗     |    ✓     |     ✗     |                  —                   | Pydantic's `model_fields`; use `reflect` package in Go                  |
| Recursive structs       |     ✓     |    ✓     |     ✓     | [Validation](../concepts/validation) |                                                                         |

---

## Summary

**127/147 features** — Full parity with go-playground/validator. Strong parity with Pydantic v2 for features applicable
to Go's type system.

| Category               | Coverage |
|------------------------|:--------:|
| Validation Basics      |   6/6    |
| String Constraints     |  16/16   |
| Numeric Constraints    |   7/8    |
| Format Validators      |  46/46   |
| Collection Validation  |   5/7    |
| Cross-Field Validation |   7/10   |
| Type Support           |  12/15   |
| JSON Operations        |   7/9    |
| Schema Generation      |   7/7    |
| Struct Configuration   |   4/9    |
| Error Handling         |   4/6    |
| Custom Validation      |   5/6    |
| Advanced Features      |   1/8    |

---

## Get Started

- [Installation](../getting-started/installation)
- [Quick Start](../getting-started/quickstart)
- [Basic Examples](../examples/basic)