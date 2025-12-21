---
sidebar_position: 2
---

# Constraints Overview

Constraints are validation rules applied to struct fields via struct tags. Pedantigo provides a comprehensive set of built-in constraints covering everything from basic requirements to complex format validation.

## Constraint Syntax

Constraints are specified using the `pedantigo` struct tag. Multiple constraints can be combined with commas, and some accept parameters:

```go
type User struct {
    // Basic constraint
    Name     string `json:"name" pedantigo:"required"`

    // Multiple constraints
    Email    string `json:"email" pedantigo:"required,email"`

    // Constraints with parameters
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
    Username string `json:"username" pedantigo:"minLength=3,maxLength=20,pattern=^[a-z0-9_]+$"`
}
```

## Constraint Categories

### Core Constraints

The fundamental constraints applicable across multiple types:

| Constraint | Parameter | Description | Example |
|-----------|-----------|-------------|---------|
| `required` | None | Field must be present in the input | `pedantigo:"required"` |
| `min` | Numeric | Minimum value (numeric) or length (string) | `pedantigo:"min=18"` |
| `max` | Numeric | Maximum value (numeric) or length (string) | `pedantigo:"max=100"` |
| `gt` | Numeric | Greater than | `pedantigo:"gt=0"` |
| `gte` | Numeric | Greater than or equal | `pedantigo:"gte=1"` |
| `lt` | Numeric | Less than | `pedantigo:"lt=100"` |
| `lte` | Numeric | Less than or equal | `pedantigo:"lte=99"` |
| `eq` | Value | Must equal exact value | `pedantigo:"eq=active"` |
| `ne` | Value | Must NOT equal value | `pedantigo:"ne=banned"` |
| `oneof` | Space-separated values | Must be one of specified values | `pedantigo:"oneof=red green blue"` |
| `oneofci` | Space-separated values | Case-insensitive oneof | `pedantigo:"oneofci=admin user guest"` |
| `len` | Numeric | Exact length (strings/arrays) | `pedantigo:"len=32"` |

See the [Numeric Constraints](/docs/constraints/numeric) page for detailed numeric range examples.

### String Constraints

Specialized constraints for string validation:

| Constraint | Parameter | Description | Example |
|-----------|-----------|-------------|---------|
| `minLength` | Numeric | Minimum string length | `pedantigo:"minLength=3"` |
| `maxLength` | Numeric | Maximum string length | `pedantigo:"maxLength=100"` |
| `alpha` | None | Only alphabetic characters | `pedantigo:"alpha"` |
| `alphanum` | None | Only letters and numbers | `pedantigo:"alphanum"` |
| `ascii` | None | Only ASCII characters | `pedantigo:"ascii"` |
| `lowercase` | None | Must be lowercase | `pedantigo:"lowercase"` |
| `uppercase` | None | Must be uppercase | `pedantigo:"uppercase"` |
| `contains` | String | Must contain substring | `pedantigo:"contains=test"` |
| `excludes` | String | Must not contain substring | `pedantigo:"excludes=forbidden"` |
| `startswith` | String | Must start with prefix | `pedantigo:"startswith=https://"` |
| `endswith` | String | Must end with suffix | `pedantigo:"endswith=.com"` |
| `strip_whitespace` | None | No leading/trailing whitespace | `pedantigo:"strip_whitespace"` |
| `pattern` | Regex | Match regex pattern | `pedantigo:"pattern=^[a-z]+$"` |
| `regexp` | Regex | Match regex pattern (alias) | `pedantigo:"regexp=^[a-z]+$"` |

See the [String Constraints](/docs/constraints/string) page for detailed string validation examples.

### Numeric Constraints

Additional constraints specific to numeric types:

| Constraint | Parameter | Description | Example |
|-----------|-----------|-------------|---------|
| `positive` | None | Must be greater than zero | `pedantigo:"positive"` |
| `negative` | None | Must be less than zero | `pedantigo:"negative"` |
| `multiple_of` | Numeric | Must be divisible by value | `pedantigo:"multiple_of=5"` |
| `max_digits` | Numeric | Maximum total digits | `pedantigo:"max_digits=8"` |
| `decimal_places` | Numeric | Maximum decimal places | `pedantigo:"decimal_places=2"` |
| `disallow_inf_nan` | None | Reject infinity and NaN values | `pedantigo:"disallow_inf_nan"` |

See the [Numeric Constraints](/docs/constraints/numeric) page for detailed numeric validation examples.

### Format Constraints

Constraints for common data formats:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `email` | Valid email address format | `pedantigo:"email"` |
| `url` | Valid URL format | `pedantigo:"url"` |
| `uri` | Valid URI format | `pedantigo:"uri"` |
| `uuid` | Valid UUID (any version) | `pedantigo:"uuid"` |
| `ipv4` | Valid IPv4 address | `pedantigo:"ipv4"` |
| `ipv6` | Valid IPv6 address | `pedantigo:"ipv6"` |
| `ip` | Valid IPv4 or IPv6 address | `pedantigo:"ip"` |
| `cidr` | Valid CIDR notation (IPv4 or IPv6) | `pedantigo:"cidr"` |
| `cidrv4` | Valid IPv4 CIDR notation | `pedantigo:"cidrv4"` |
| `cidrv6` | Valid IPv6 CIDR notation | `pedantigo:"cidrv6"` |
| `mac` | Valid MAC address | `pedantigo:"mac"` |
| `hostname` | Valid hostname | `pedantigo:"hostname"` |
| `hostname_rfc1123` | Valid RFC 1123 hostname | `pedantigo:"hostname_rfc1123"` |
| `fqdn` | Valid fully qualified domain name | `pedantigo:"fqdn"` |
| `port` | Valid port number (0-65535) | `pedantigo:"port"` |
| `tcp_addr` | Valid TCP address | `pedantigo:"tcp_addr"` |
| `udp_addr` | Valid UDP address | `pedantigo:"udp_addr"` |
| `tcp4_addr` | Valid TCP4 address | `pedantigo:"tcp4_addr"` |
| `json` | Valid JSON string | `pedantigo:"json"` |
| `jwt` | Valid JSON Web Token | `pedantigo:"jwt"` |
| `semver` | Valid semantic version | `pedantigo:"semver"` |
| `cron` | Valid cron expression | `pedantigo:"cron"` |
| `datetime` | Matches Go time layout | `pedantigo:"datetime=2006-01-02"` |
| `ulid` | Valid ULID format | `pedantigo:"ulid"` |

See the [Format Constraints](/docs/constraints/format) page for detailed format validation examples.

### Encoding Constraints

Constraints for encoded data formats:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `base64` | Valid base64 encoding | `pedantigo:"base64"` |
| `base64url` | Valid URL-safe base64 encoding | `pedantigo:"base64url"` |
| `base64rawurl` | Valid raw URL-safe base64 encoding | `pedantigo:"base64rawurl"` |

### Hash Constraints

Constraints for validating hash format strings:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `md5` | Valid MD5 hash format | `pedantigo:"md5"` |
| `sha256` | Valid SHA256 hash format | `pedantigo:"sha256"` |
| `sha384` | Valid SHA384 hash format | `pedantigo:"sha384"` |
| `sha512` | Valid SHA512 hash format | `pedantigo:"sha512"` |
| `mongodb` | Valid MongoDB ObjectID format | `pedantigo:"mongodb"` |

### Finance Constraints

Constraints for financial and cryptocurrency identifiers:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `credit_card` | Valid credit card number (Luhn check) | `pedantigo:"credit_card"` |
| `luhn_checksum` | Valid Luhn checksum | `pedantigo:"luhn_checksum"` |
| `btc_addr` | Valid Bitcoin address (P2PKH/P2SH) | `pedantigo:"btc_addr"` |
| `btc_addr_bech32` | Valid Bitcoin bech32 address | `pedantigo:"btc_addr_bech32"` |
| `eth_addr` | Valid Ethereum address | `pedantigo:"eth_addr"` |

### Identity Constraints

Constraints for identification numbers and codes:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `isbn` | Valid ISBN (10 or 13) | `pedantigo:"isbn"` |
| `isbn10` | Valid ISBN-10 | `pedantigo:"isbn10"` |
| `isbn13` | Valid ISBN-13 | `pedantigo:"isbn13"` |
| `issn` | Valid ISSN format | `pedantigo:"issn"` |
| `ssn` | Valid US Social Security Number | `pedantigo:"ssn"` |
| `ein` | Valid US Employer Identification Number | `pedantigo:"ein"` |
| `e164` | Valid E.164 phone number format | `pedantigo:"e164"` |

### Geographic Constraints

Constraints for geographic coordinates:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `latitude` | Valid latitude (-90 to 90) | `pedantigo:"latitude"` |
| `longitude` | Valid longitude (-180 to 180) | `pedantigo:"longitude"` |

### Color Constraints

Constraints for color format validation:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `hexcolor` | Valid hex color (#RGB or #RRGGBB) | `pedantigo:"hexcolor"` |
| `rgb` | Valid RGB color | `pedantigo:"rgb"` |
| `rgba` | Valid RGBA color | `pedantigo:"rgba"` |
| `hsl` | Valid HSL color | `pedantigo:"hsl"` |
| `hsla` | Valid HSLA color | `pedantigo:"hsla"` |

### ISO Constraints

Constraints for ISO standard codes and formats:

| Constraint | Parameter | Description | Example |
|-----------|-----------|-------------|---------|
| `iso3166_alpha2` | None | ISO 3166-1 alpha-2 country code | `pedantigo:"iso3166_alpha2"` |
| `iso3166_alpha2_eu` | None | ISO 3166-1 alpha-2 EU country code | `pedantigo:"iso3166_alpha2_eu"` |
| `iso3166_alpha3` | None | ISO 3166-1 alpha-3 country code | `pedantigo:"iso3166_alpha3"` |
| `iso3166_alpha3_eu` | None | ISO 3166-1 alpha-3 EU country code | `pedantigo:"iso3166_alpha3_eu"` |
| `iso3166_numeric` | None | ISO 3166-1 numeric country code | `pedantigo:"iso3166_numeric"` |
| `iso3166_2` | None | ISO 3166-2 subdivision code | `pedantigo:"iso3166_2"` |
| `iso4217` | None | ISO 4217 currency code | `pedantigo:"iso4217"` |
| `iso4217_numeric` | None | ISO 4217 numeric currency code | `pedantigo:"iso4217_numeric"` |
| `postcode` | Country code | Postal code for specific country | `pedantigo:"postcode=US"` |
| `postcode_iso3166_alpha2` | Country code | Postal code (alias for `postcode`) | `pedantigo:"postcode_iso3166_alpha2=GB"` |
| `bcp47` | None | BCP 47 language tag | `pedantigo:"bcp47"` |

### Filesystem Constraints

Constraints for file and directory path validation:

| Constraint | Description | Example |
|-----------|-------------|---------|
| `filepath` | Valid file path | `pedantigo:"filepath"` |
| `dirpath` | Valid directory path | `pedantigo:"dirpath"` |
| `file` | Path must point to existing file | `pedantigo:"file"` |
| `dir` | Path must point to existing directory | `pedantigo:"dir"` |

### Collection Constraints

Constraints for arrays, slices, and maps:

| Constraint | Parameter | Description | Example |
|-----------|-----------|-------------|---------|
| `minItems` | Numeric | Minimum number of items | `pedantigo:"minItems=1"` |
| `maxItems` | Numeric | Maximum number of items | `pedantigo:"maxItems=100"` |
| `unique` | None | All items must be unique | `pedantigo:"unique"` |

See the [Collection Constraints](/docs/constraints/collection) page for detailed collection validation examples.

### Default Values

| Constraint | Parameter | Description | Example |
|-----------|-----------|-------------|---------|
| `default` | Value | Default value if field missing | `pedantigo:"default=active"` |

## Complete Example

Here's a realistic example combining constraints from multiple categories:

```go
package main

import (
    "log"
    "github.com/smrutai/pedantigo"
)

type UserProfile struct {
    // Core constraints
    ID       string `json:"id" pedantigo:"required,uuid"`
    Email    string `json:"email" pedantigo:"required,email"`
    Username string `json:"username" pedantigo:"required,minLength=3,maxLength=20,alphanum"`

    // String constraints
    Bio      string `json:"bio,omitempty" pedantigo:"maxLength=500"`
    Website  string `json:"website,omitempty" pedantigo:"url"`

    // Numeric constraints
    Age      int    `json:"age" pedantigo:"min=13,max=120"`
    Rating   float64 `json:"rating,omitempty" pedantigo:"min=0,max=5,decimal_places=1"`

    // Geographic constraints
    Latitude  float64 `json:"latitude,omitempty" pedantigo:"latitude"`
    Longitude float64 `json:"longitude,omitempty" pedantigo:"longitude"`

    // ISO constraints
    Country  string `json:"country,omitempty" pedantigo:"iso3166_alpha2"`
    Currency string `json:"currency,omitempty" pedantigo:"iso4217"`

    // Collection constraints
    Tags     []string `json:"tags,omitempty" pedantigo:"maxItems=10,unique"`
    Roles    []string `json:"roles" pedantigo:"minItems=1,oneof=admin moderator user"`

    // Enum constraint
    Status   string `json:"status" pedantigo:"required,oneof=active inactive suspended"`
}

func main() {
    jsonData := []byte(`{
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "alice@example.com",
        "username": "alice123",
        "bio": "Software engineer and open source enthusiast",
        "website": "https://example.com",
        "age": 28,
        "rating": 4.5,
        "latitude": 40.7128,
        "longitude": -74.0060,
        "country": "US",
        "currency": "USD",
        "tags": ["golang", "rust", "web"],
        "roles": ["user"],
        "status": "active"
    }`)

    user, err := pedantigo.Unmarshal[UserProfile](jsonData)
    if err != nil {
        log.Fatalf("Validation failed: %v", err)
    }

    // user is now a fully validated UserProfile
    log.Printf("User: %+v", user)
}
```

## Validation Error Handling

When validation fails, Pedantigo returns detailed errors for each field:

```go
user, err := pedantigo.Unmarshal[UserProfile](invalidData)
if err != nil {
    if validationErr, ok := err.(*pedantigo.ValidationError); ok {
        for _, fieldErr := range validationErr.Errors {
            fmt.Printf("Field: %s, Error: %s\n", fieldErr.Field, fieldErr.Message)
        }
    }
}
```

Example validation error output:
```
Field: email, Error: must be a valid email address
Field: age, Error: must be at least 13
Field: rating, Error: must be at most 5
Field: roles, Error: field is required
```

## Context-Aware Constraints

Some constraints behave differently depending on the field type:

- **`min`/`max`**: For numeric types, validates value range. For strings, validates length. For arrays, validates item count.
- **`gt`/`gte`/`lt`/`lte`**: Works with numeric types and comparable values.
- **`len`**: For strings, validates character count. For arrays/slices, validates element count.

## Performance Considerations

Constraint validation in Pedantigo is highly optimized:

- **Format constraints** (email, URL, UUID, etc.) use compiled regex patterns cached at startup
- **ISO code validation** uses precompiled lookup tables
- **Numeric constraints** perform simple arithmetic comparisons
- **String constraints** use efficient string operations

See [Schema Generation](/docs/concepts/schema) for caching strategy that provides 240x speedup.

## Next Steps

Learn more about specific constraint categories:

- **[String Constraints](/docs/constraints/string)** - Detailed string validation
- **[Numeric Constraints](/docs/constraints/numeric)** - Detailed numeric validation
- **[Format Constraints](/docs/constraints/format)** - Email, URL, UUID, and more
- **[Collection Constraints](/docs/constraints/collection)** - Array and slice validation
- **[Cross-Field Validation](/docs/concepts/cross-field)** - Validate relationships between fields
