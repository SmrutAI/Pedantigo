---
sidebar_position: 1
title: From go-playground/validator
description: Step-by-step guide to migrate from go-playground/validator to Pedantigo
---

# Migrating from go-playground/validator

This guide helps you migrate from [go-playground/validator](https://github.com/go-playground/validator) to Pedantigo.

---

## Quick Migration (One Line)

For most codebases, migration requires only one line:

```go
func init() {
    pedantigo.SetTagName("validate")
}
```

This tells Pedantigo to read your existing `validate:"..."` struct tags instead of `pedantigo:"..."`.

Your existing structs work unchanged:

```go
type User struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

// Works with Pedantigo
user, err := pedantigo.Unmarshal[User](jsonData)
```

---

## Supported Tags (Zero Changes Needed)

These validator tags work identically in Pedantigo:

### Core Constraints
`required`, `min`, `max`, `len`, `eq`, `ne`, `gt`, `gte`, `lt`, `lte`

### String Constraints
`email`, `url`, `uri`, `uuid`, `uuid3`, `uuid4`, `uuid5`, `alpha`, `alphanum`, `alphaunicode`, `alphanumunicode`, `numeric`, `number`, `hexadecimal`, `ascii`, `printascii`, `multibyte`, `lowercase`, `uppercase`, `contains`, `excludes`, `startswith`, `endswith`, `startsnotwith`, `endsnotwith`, `containsany`, `containsrune`, `excludesall`, `excludesrune`, `eq_ignore_case`, `ne_ignore_case`

### Enum/Choice
`oneof`, `oneofci`

### Field Comparisons
`eqfield`, `nefield`, `gtfield`, `gtefield`, `ltfield`, `ltefield`, `eqcsfield`, `necsfield`, `gtcsfield`, `gtecsfield`, `ltcsfield`, `ltecsfield`

### Conditional Validation
`required_if`, `required_unless`, `required_with`, `required_without`, `required_with_all`, `required_without_all`, `excluded_if`, `excluded_unless`, `excluded_with`, `excluded_without`, `excluded_with_all`, `excluded_without_all`, `skip_unless`

### Network
`ip`, `ipv4`, `ipv6`, `cidr`, `cidrv4`, `cidrv6`, `mac`, `hostname`, `hostname_rfc1123`, `fqdn`, `tcp_addr`, `udp_addr`, `hostname_port`, `http_url`, `https_url`

### Format Validators
`datetime`, `timezone`, `credit_card`, `isbn`, `isbn10`, `isbn13`, `issn`, `ssn`, `ein`, `e164`, `base64`, `base64url`, `base64rawurl`, `base32`, `datauri`, `urn_rfc2141`, `json`, `jwt`, `html`, `hexcolor`, `rgb`, `rgba`, `hsl`, `hsla`, `latitude`, `longitude`, `md4`, `md5`, `sha256`, `sha384`, `sha512`, `mongodb`, `cron`, `semver`, `ulid`, `luhn_checksum`, `bitcoin_addr`, `bitcoin_addr_bech32`, `ethereum_addr`, `image`

### ISO Codes
`iso3166_1_alpha2`, `iso3166_1_alpha3`, `iso3166_1_alpha_numeric`, `iso4217`, `bcp47_language_tag`, `postcode_iso3166_alpha2`

### Collections
`dive`, `unique`

### OR Operator
`hexcolor|rgb|rgba` (validates if ANY matches)

### Aliases
`iscolor` (expands to `hexcolor|rgb|rgba|hsl|hsla`)

---

## Tags You Can Remove

These validator tags are **not needed** in Pedantigo:

| Tag             | Why Not Needed                                                                        |
|-----------------|---------------------------------------------------------------------------------------|
| `omitempty`     | Pedantigo tracks JSON field presence. Empty values skip validation unless `required`. |
| `omitnil`       | Nil pointers are handled automatically.                                               |
| `omitzero`      | Zero values skip validation unless `required`.                                        |
| `-`             | Simply don't add a tag.                                                               |
| `structonly`    | Not needed - Pedantigo validates all fields by default.                               |
| `nostructlevel` | Not needed.                                                                           |
| `isdefault`     | Not needed - check for zero value in code.                                            |

**Example migration:**

```go
// validator
Email string `validate:"omitempty,email"`

// pedantigo (same behavior)
Email string `validate:"email"`
```

---

## Custom Validator Registration

### RegisterAlias

```go
// validator
validate.RegisterAlias("is_active", "oneof=active enabled")

// pedantigo (identical)
pedantigo.RegisterAlias("is_active", "oneof=active enabled")
```

### Custom Validators

```go
// validator
validate.RegisterValidation("custom", customFunc)

// pedantigo
pedantigo.RegisterConstraint("custom", func(value string) (constraints.Constraint, bool) {
    return &myCustomConstraint{}, true
})
```

---

## API Differences

These validator APIs have slightly different signatures in Pedantigo:

### Var() - Single Value Validation

```go
// validator
err := validate.Var(email, "required,email")

// pedantigo
err := pedantigo.Var(email, "required,email")
```

### StructPartial / StructExcept

```go
// validator
err := validate.StructPartial(user, "Username", "Email")
err := validate.StructExcept(user, "Password")

// pedantigo
err := pedantigo.StructPartial(&user, "Username", "Email")
err := pedantigo.StructExcept(&user, "Password")
```

### RegisterValidationCtx - Context-Aware Validators

```go
// validator
validate.RegisterValidationCtx("db_unique", func(ctx context.Context, fl validator.FieldLevel) bool {
    // ...
})

// pedantigo
pedantigo.RegisterValidationCtx("db_unique", func(ctx context.Context, value any, param string) error {
    // Return error instead of bool
    return nil
})

// Usage
err := pedantigo.ValidateCtx(ctx, &user)
```

### RegisterTagNameFunc

```go
// validator
validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
    return fld.Tag.Get("json")
})

// pedantigo
pedantigo.RegisterTagNameFunc(func(field reflect.StructField) string {
    return field.Tag.Get("json")
})
```

---

## What You Gain

Pedantigo provides features not available in validator:

| Feature                | Description                                    |
|------------------------|------------------------------------------------|
| JSON Schema generation | `pedantigo.Schema[User]()`                     |
| Unmarshal + Validate   | Single step: `pedantigo.Unmarshal[User](json)` |
| Streaming validation   | Parse partial JSON for LLM output              |
| Discriminated unions   | `Union[TypeA, TypeB, TypeC]`                   |
| ExtraAllow mode        | Capture unknown JSON fields                    |
| Secret types           | `Secret[string]` masks in logs                 |
| Transformers           | `strip_whitespace`, `to_lower`, `to_upper`     |
| Default values         | `default=value`                                |

---

## Step-by-Step Migration

1. **Add the tag override:**
   ```go
   func init() {
       pedantigo.SetTagName("validate")
   }
   ```

2. **Replace validation calls:**
   ```go
   // Before (validator)
   validate := validator.New()
   err := validate.Struct(user)

   // After (pedantigo)
   user, err := pedantigo.Unmarshal[User](jsonData)
   // or
   err := pedantigo.Validate(&user)
   ```

3. **Remove unnecessary tags:**
   - Delete `omitempty`, `omitnil`, `omitzero`
   - Delete `-` tags (just remove the tag entirely)

4. **Test your structs:**
   ```go
   go test ./...
   ```

5. **Optional: Migrate tag name:**
   Once validated, you can gradually rename `validate` to `pedantigo` tags if desired.

---

## Troubleshooting

### "unknown constraint" error

Check if the constraint is supported in the [API Parity](./api-parity.md) comparison. If not, implement a custom validator.

### Different validation behavior

Pedantigo may have stricter or different validation for some formats. Test edge cases and adjust if needed.