---
sidebar_position: 1
title: From go-playground/validator
description: What actually changes when migrating from go-playground/validator to Pedantigo
---

# Migrating from go-playground/validator

Pedantigo's default struct tag is `validate` — the same tag name go-playground/validator uses. If you're migrating, **your existing struct tags need no changes.** Tag syntax is ~127/147 identical (see [API Parity](./api-parity.md)), and any validator-only tag pedantigo doesn't recognize (`omitnil`, `structonly`, etc.) is silently ignored rather than erroring.

Migration is two steps. Step 1 gets you running with a near-rename; step 2 is an optional, later upgrade for high-throughput code paths.

**Caution:** if your codebase already has unrelated `validate:"..."` tags on structs (leftover from another library, or dead annotations), Pedantigo will now read and enforce them by default.

---

## Step 1: Switch to the Simple API

```go
// go-playground/validator
validate := validator.New()
err := validate.Struct(user)

// pedantigo Simple API — drop-in, no setup (internally cached per type via sync.Map)
err := validator.Validate(&user)
// or, to parse + validate in one step:
user, err := validator.Unmarshal[User](jsonBytes)
```

This is a near-rename: no validator to construct, no cache to manage. Benchmarked cost is a ~200ns `sync.Map` lookup per call, ~2-5µs total with unmarshal — negligible below roughly 100k req/sec. This is the recommended default for 99% of applications ([full benchmarks](/docs/advanced/performance)).

## Step 2 (optional): Upgrade to the Core API for hot paths

```go
// Declare once, at package/startup scope
var userValidator = validator.New[User]()

// Call directly — skips the Simple API's cache lookup
err := userValidator.Validate(user)
user, err := userValidator.Unmarshal(jsonBytes)
```

Benchmarked saving is ~200ns/call (the `sync.Map` lookup Step 1 pays on every call). Only worth it if profiling shows that lookup in a flame graph — the performance guide's own recommendation is "profile first." Everything else (tags, custom validators, `omitempty`, cross-field constraints) behaves identically between the Simple and Core API; only the construction call-site changes.

## Behavior difference: `omitempty` is a real constraint, not a parser special case

- **Regular constraints** (`min`, `max`, `oneof`, `email`, etc.) are skipped when the field is at its zero value — same effect as validator.
- **Cross-field constraints** (`required_with`, `required_if`, `eqfield`, etc.) always run, even for zero-value fields — this is the one behavioral difference to test for.

```go
Email string `validate:"omitempty,email"` // identical effect in both libraries
```

See [omitempty as a Validation Constraint](/docs/api/initialization#pedantigo-omitempty) for the full reference.

## API difference: custom validator registration signatures

```go
// validator
validate.RegisterValidation("custom", customFunc)

// pedantigo
validator.RegisterValidation("custom", func(value any, param string) error {
    return nil // return an error to fail validation
})
```

---

## Tags that are safe no-ops

These validator-only tags are silently ignored by pedantigo — harmless to leave, but fine to delete for clarity:

`omitnil`, `omitzero`, `-`, `structonly`, `nostructlevel`, `isdefault`. (See "Behavior difference" above for `omitempty`, which is not in this list.)

---

## RegisterAlias

```go
// validator
validate.RegisterAlias("is_active", "oneof=active enabled")

// pedantigo (identical)
validator.RegisterAlias("is_active", "oneof=active enabled")
```

(See "API difference" above for custom validator registration via `RegisterValidation`.)

---

## API Differences

These validator APIs have slightly different signatures in Pedantigo:

### Var() - Single Value Validation

```go
// validator
err := validate.Var(email, "required,email")

// pedantigo
err := validator.Var(email, "required,email")
```

### StructPartial / StructExcept

```go
// validator
err := validate.StructPartial(user, "Username", "Email")
err := validate.StructExcept(user, "Password")

// pedantigo
err := validator.ValidatePartial(&user, "Username", "Email")
err := validator.ValidateExcept(&user, "Password")
```

### RegisterValidationCtx - Context-Aware Validators

```go
// validator
validate.RegisterValidationCtx("db_unique", func(ctx context.Context, fl validator.FieldLevel) bool {
    // ...
})

// pedantigo
validator.RegisterValidationCtx("db_unique", func(ctx context.Context, value any, param string) error {
    // Return error instead of bool
    return nil
})

// Usage
err := validator.ValidateCtx(ctx, &user)
```

### RegisterTagNameFunc

```go
// validator
validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
    return fld.Tag.Get("json")
})

// pedantigo
validator.RegisterTagNameFunc(func(field reflect.StructField) string {
    return field.Tag.Get("json")
})
```

---

## What You Gain

Pedantigo provides features not available in validator:

| Feature                | Description                                    |
|------------------------|------------------------------------------------|
| JSON Schema generation | `validator.Schema[User]()`                     |
| Unmarshal + Validate   | Single step: `validator.Unmarshal[User](json)` |
| Streaming validation   | Parse partial JSON for LLM output              |
| Discriminated unions   | `Union[TypeA, TypeB, TypeC]`                   |
| ExtraAllow mode        | Capture unknown JSON fields                    |
| Secret types           | `Secret[string]` masks in logs                 |
| Transformers           | `strip_whitespace`, `to_lower`, `to_upper`     |
| Default values         | `default=value`                                |

---

## Checklist

- [ ] Swap `validator.New()` + `.Struct()` calls for the Simple API (Step 1) — struct tags need no edits
- [ ] Run `go test ./...` and confirm behavior on zero-value fields with cross-field constraints (the one behavioral difference — see above)
- [ ] Delete any now-unnecessary validator-only tags for clarity (optional — they're harmless no-ops either way)
- [ ] Profile before considering Step 2 (Core API) — only relevant at high request volume

---

## Troubleshooting

### "unknown constraint" error

Check if the constraint is supported in the [API Parity](./api-parity.md) comparison. If not, implement a custom validator.

### Different validation behavior

Pedantigo may have stricter or different validation for some formats. Test edge cases and adjust if needed.