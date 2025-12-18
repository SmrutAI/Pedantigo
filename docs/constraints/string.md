---
sidebar_position: 1
---

# String Constraints

String validation rules in Pedantigo for validating text content, length, character composition, and patterns.

## Length Constraints

### `min` / `max`

Validates the **number of characters** (runes) in the string.

```go
type Post struct {
    // Username must be 3-20 characters
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    // Bio can be empty or up to 500 characters
    Bio      string `json:"bio" pedantigo:"max=500"`
}
```

**Behavior:**
- Counts Unicode characters (runes), not bytes
- Empty strings skip validation (only required constraint enforces non-empty)

### `len`

Validates that a string has an **exact number of characters**.

```go
type License struct {
    // License key must be exactly 32 characters
    Key string `json:"key" pedantigo:"len=32"`
}
```

**Behavior:**
- Counts Unicode characters (runes)
- Validates even empty strings (len=0 is valid for empty strings)
- Returns error if character count does not match exactly

## Character Composition Constraints

### `alpha`

Validates that a string contains **only alphabetic characters** (A-Z, a-z).

```go
type Person struct {
    // FirstName must be letters only
    FirstName string `json:"first_name" pedantigo:"required,alpha"`
    LastName  string `json:"last_name" pedantigo:"required,alpha"`
}
```

**Valid examples:** "John", "Marie", "Alexandre"
**Invalid examples:** "John123", "Marie-Anne", "St. James"

### `alphanum`

Validates that a string contains **only alphabetic characters and digits** (A-Z, a-z, 0-9).

```go
type Username struct {
    // Username can contain letters and numbers only
    Handle string `json:"handle" pedantigo:"required,alphanum,min=3,max=20"`
}
```

**Valid examples:** "user123", "test456abc", "Username42"
**Invalid examples:** "user-123", "user_name", "user@123"

### `ascii`

Validates that a string contains **only ASCII characters** (character codes 0-127).

```go
type APIKey struct {
    // API key must be ASCII-compatible
    Token string `json:"token" pedantigo:"required,ascii"`
}
```

**Valid examples:** "abc123!@#", "Bearer-Token-xyz"
**Invalid examples:** "café", "Москва", "北京" (non-ASCII characters)

## Content Constraints

### `contains`

Validates that a string **contains a specific substring**.

```go
type DocumentTag struct {
    // Tag must contain the word "archive"
    Label string `json:"label" pedantigo:"required,contains=archive"`
}
```

**Valid examples:** "archive-2024", "old-archive", "archive"
**Invalid examples:** "archived", "archiv", "archive-" is valid but "ark-chive" is not

### `excludes`

Validates that a string **does NOT contain a specific substring**.

```go
type Comment struct {
    // Comment must not contain profanity
    Text string `json:"text" pedantigo:"required,excludes=badword"`
}
```

**Valid examples:** "This is a great post", "I love this!"
**Invalid examples:** "This contains badword in it"

### `startswith`

Validates that a string **starts with a specific prefix**.

```go
type OrderID struct {
    // Order ID must start with "ORD-"
    ID string `json:"id" pedantigo:"required,startswith=ORD-"`
}
```

**Valid examples:** "ORD-12345", "ORD-ABC-001"
**Invalid examples:** "12345-ORD", "ORDER-12345"

### `endswith`

Validates that a string **ends with a specific suffix**.

```go
type EmailAddress struct {
    // Email must end with company domain
    Email string `json:"email" pedantigo:"required,endswith=@company.com"`
}
```

**Valid examples:** "john@company.com", "support@company.com"
**Invalid examples:** "john@example.com", "company.com"

## Case Constraints

### `lowercase`

Validates that a string is **entirely lowercase**.

```go
type Hostname struct {
    // Hostname must be lowercase
    Domain string `json:"domain" pedantigo:"required,lowercase"`
}
```

**Valid examples:** "example-domain.com", "test123"
**Invalid examples:** "Example-Domain.com", "TEST123", "example-Domain"

### `uppercase`

Validates that a string is **entirely uppercase**.

```go
type CountryCode struct {
    // Country code must be uppercase
    Code string `json:"code" pedantigo:"required,uppercase,len=2"`
}
```

**Valid examples:** "US", "GB", "CA"
**Invalid examples:** "us", "Us", "uS"

## Whitespace Handling

### `strip_whitespace`

Validates that a string has **no leading or trailing whitespace** in validation mode.

```go
type TagName struct {
    // Tag must not have leading/trailing whitespace
    Tag string `json:"tag" pedantigo:"required,strip_whitespace"`
}
```

**Valid examples:** "golang", "best-practices", "test123"
**Invalid examples:** " golang", "golang ", " golang "

**Note:** This validates that whitespace is already stripped. It does not strip the whitespace automatically.

## Pattern Matching

### `regexp`

Validates that a string matches a **custom regular expression pattern**.

```go
type PhoneNumber struct {
    // Phone must match pattern (flexible for various formats)
    Phone string `json:"phone" pedantigo:"regexp=^\\+?[0-9\\s\\-\\(\\)]+$"`
}

type Username struct {
    // Username: letters, numbers, hyphens, underscores, 3-20 chars
    Username string `json:"username" pedantigo:"regexp=^[a-zA-Z0-9_-]{3,20}$"`
}

type Zipcode struct {
    // US Zipcode: 5 or 9 digits
    Zipcode string `json:"zipcode" pedantigo:"regexp=^\\d{5}(-\\d{4})?$"`
}
```

**Pattern Syntax:**
- Uses Go's `regexp` package syntax
- Escape special characters with backslashes
- Use `^` to match start, `$` to match end
- Use `[]` for character classes, `{n,m}` for repetition

**Valid regex examples:**
- `^[a-z0-9]+$` - lowercase letters and numbers only
- `^\\d{10}$` - exactly 10 digits
- `^[A-Z][a-z]*$` - starts with uppercase, then lowercase
- `^\\w+@\\w+\\.\\w+$` - email-like pattern

## Format Constraints

These constraints validate common string formats (documented in detail in [Format Constraints](/docs/constraints/format)):

- `email` - Valid email format
- `url` - Valid URL (http/https)
- `uuid` - Valid UUID format
- `uri` - Valid URI format
- `ipv4` - Valid IPv4 address
- `ipv6` - Valid IPv6 address

## Complete Example

Here's a comprehensive example showing multiple string constraints working together:

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type UserProfile struct {
    // Username: lowercase alphanumeric, 3-20 chars, no leading spaces
    Username string `json:"username" pedantigo:"required,lowercase,alphanum,min=3,max=20"`

    // Email: valid format
    Email string `json:"email" pedantigo:"required,email"`

    // Bio: optional, max 500 chars
    Bio string `json:"bio,omitempty" pedantigo:"max=500"`

    // Country: exactly 2 uppercase letters
    Country string `json:"country" pedantigo:"required,uppercase,len=2,alpha"`

    // Website: optional URL starting with https
    Website string `json:"website,omitempty" pedantigo:"url,startswith=https://"`

    // Nickname: letters only, 2-50 chars
    Nickname string `json:"nickname" pedantigo:"alpha,min=2,max=50"`

    // Status: one of specific values
    Status string `json:"status" pedantigo:"required,oneof=active inactive suspended"`

    // ApiKey: 32 ASCII characters
    ApiKey string `json:"api_key" pedantigo:"required,len=32,ascii"`

    // AccountID: exactly 10 digits
    AccountID string `json:"account_id" pedantigo:"required,regexp=^\\d{10}$"`

    // Preferences: no leading/trailing whitespace
    Preferences string `json:"preferences,omitempty" pedantigo:"strip_whitespace"`
}

func main() {
    // Valid profile
    profileJSON := []byte(`{
        "username": "johndoe123",
        "email": "john@example.com",
        "bio": "Software developer passionate about Go",
        "country": "US",
        "website": "https://example.com",
        "nickname": "John",
        "status": "active",
        "api_key": "abcdef1234567890abcdef1234567890",
        "account_id": "1234567890"
    }`)

    profile, err := pedantigo.Unmarshal[UserProfile](profileJSON)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("Valid profile: %+v\n", profile)

    // Invalid profile - multiple constraint violations
    invalidJSON := []byte(`{
        "username": "John_Doe_123",
        "email": "not-an-email",
        "country": "USA",
        "website": "http://example.com",
        "status": "unknown",
        "api_key": "short",
        "account_id": "abc123"
    }`)

    _, err = pedantigo.Unmarshal[UserProfile](invalidJSON)
    if err != nil {
        fmt.Printf("Validation errors:\n%v\n", err)
        // Output will show all constraint violations:
        // - username: must contain only alphanumeric characters (has underscore)
        // - email: must be a valid email address
        // - country: must be exactly 2 characters (has 3)
        // - website: must start with 'https://' (has 'http://')
        // - status: must be one of [active inactive suspended]
        // - api_key: must be exactly 32 characters
        // - account_id: must match pattern ^\d{10}$
    }
}
```

## Combining Constraints

String constraints can be combined to create powerful validation rules:

```go
type SecurePassword struct {
    // Must be 12+ chars, contain uppercase, numbers, special chars
    Password string `json:"password" pedantigo:"required,min=12,regexp=^(?=.*[A-Z])(?=.*\\d)(?=.*[!@#$%])`
}

type EuropeanPhoneNumber struct {
    // Format: +CC-XXX-XXX-XXXX where CC is country code
    Phone string `json:"phone" pedantigo:"required,regexp=^\\+\\d{1,3}-\\d{3}-\\d{3}-\\d{4}$"`
}

type FileExtension struct {
    // Must be exactly 3 chars, lowercase, representing file type
    Extension string `json:"ext" pedantigo:"required,len=3,lowercase,alpha"`
}

type SlugURL struct {
    // URL-friendly slug: lowercase, hyphens, no spaces or special chars
    Slug string `json:"slug" pedantigo:"required,lowercase,regexp=^[a-z0-9-]+$"`
}
```

## Validation Behavior Notes

- **Empty strings:** Most string constraints skip validation for empty strings. Use `required` to enforce non-empty values.
- **Unicode support:** `min`, `max`, and `len` count Unicode characters (runes), not bytes.
- **Case sensitivity:** `lowercase`, `uppercase`, `contains`, `startswith`, and `endswith` are case-sensitive.
- **Whitespace:** Internal whitespace is preserved; only `strip_whitespace` validates absence of leading/trailing space.
- **Nil values:** Nil pointers are skipped in validation; use `required` to enforce non-nil.

## Quick Reference Table

| Constraint | Example | Description |
|------------|---------|-------------|
| `min=N` | `min=3` | Minimum length in characters |
| `max=N` | `max=100` | Maximum length in characters |
| `len=N` | `len=32` | Exact length in characters |
| `alpha` | `alpha` | Letters only (A-Z, a-z) |
| `alphanum` | `alphanum` | Letters and numbers only |
| `ascii` | `ascii` | ASCII characters only (0-127) |
| `contains=str` | `contains=@` | Must contain substring |
| `excludes=str` | `excludes=admin` | Must NOT contain substring |
| `startswith=str` | `startswith=https://` | Must start with prefix |
| `endswith=str` | `endswith=.com` | Must end with suffix |
| `lowercase` | `lowercase` | All lowercase letters |
| `uppercase` | `uppercase` | All uppercase letters |
| `strip_whitespace` | `strip_whitespace` | No leading/trailing whitespace |
| `regexp=pattern` | `regexp=^[0-9]{5}$` | Match regex pattern |
| `email` | `email` | Valid email format |
| `url` | `url` | Valid URL (http/https) |
| `uuid` | `uuid` | Valid UUID format |
