---
sidebar_position: 3
title: v1 to v2
description: What changed in Pedantigo v2 and how to upgrade
---

# Migrating from v1 to v2

Pedantigo v2 changes the **default struct tag** from `pedantigo` to `validate`. This is the only breaking change in v2 — everything else (constraint syntax, the Simple API, the Core API) is unchanged.

---

## Why this changed

`validate` is the tag name used by [go-playground/validator](https://github.com/go-playground/validator), the most widely used Go validation library, and Pedantigo already tracks its constraint syntax closely. Matching the tag name too means structs written for that ecosystem now work under Pedantigo with zero tag edits.

It also makes Pedantigo-validated structs visible to tooling that reads `validate` tags. [swaggo/swag](https://github.com/swaggo/swag), the standard tool for generating OpenAPI specs from Go source, reads `validate:"required"` to mark fields required (and has partial support for `oneof`/`min`/`max`). A struct tagged `pedantigo:"..."` was invisible to swag; the same struct tagged `validate:"..."` now produces a more accurate OpenAPI spec, and by extension more accurate generated client SDKs, with no extra annotation work.

---

## What breaks

If your structs rely on the implicit default — no `SetTagName()` call, tags written as `pedantigo:"..."` — upgrading to v2 does **not** produce a compile error or a panic. It fails silently: Pedantigo v2 looks for `validate` tags by default, finds none on your fields, and treats them as unconstrained. Validation that used to run stops running, with no error to signal it.

This is the dangerous case. Audit for it before upgrading — search your codebase for `pedantigo:"` tags.

**Caution, the reverse case:** if your codebase already has unrelated `validate:"..."` tags on structs — left over from a different library, or dead annotations — Pedantigo v2 will now read and enforce them by default. Check for this too.

---

## Import path

Per [Go's module versioning rules](https://go.dev/ref/mod), a v2+ release requires a new import path:

```go
// v1
import "github.com/SmrutAI/pedantigo"

// v2
import "github.com/SmrutAI/pedantigo/v2"
```

```bash
go get github.com/SmrutAI/pedantigo/v2
```

v1 (`v1.1.4` and earlier) remains available and frozen at its last tag — no further v1.x.x patches are planned. There is no forced upgrade timeline; move to v2 when ready.

---

## How to migrate

Pick one:

**Option A — rename your tags (recommended).** Change `pedantigo:"..."` to `validate:"..."` throughout your structs. This aligns with the new default and the ecosystem-compatibility benefit above.

```go
// Before
Email string `json:"email" pedantigo:"required,email"`

// After
Email string `json:"email" validate:"required,email"`
```

**Option B — keep your tags, override the default.** Call `SetTagName("pedantigo")` once, before creating any validator, to preserve v1 behavior without touching your structs:

```go
func init() {
    pdcore.SetTagName("pedantigo")
}
```

Option A is the better long-term choice — it gets you the tooling compatibility this release is for. Option B is a valid stopgap if you need to defer the tag rename.

---

## Everything else is unchanged

Constraint syntax, the Simple API (`pdcore.Unmarshal`, `pdcore.Validate`), the Core API (`pdcore.New[T]()`), schema generation, and custom validator registration all work exactly as they did in v1. This is a one-line-of-behavior change, not a rewrite.

---

## Package restructure — `pdcore` sub-package

All library code moved from the root package to `pdcore/`. The import path and package qualifier change:

```go
// v2.0.0 (before restructure)
import "github.com/SmrutAI/pedantigo/v2"
pedantigo.New[User]()
pedantigo.Unmarshal[User](data)
pedantigo.SecretStr

// v2.x.x (after restructure)
import "github.com/SmrutAI/pedantigo/v2/pdcore"
pdcore.New[User]()
pdcore.Unmarshal[User](data)
pdcore.SecretStr
```

Install the new path:

```bash
go get github.com/SmrutAI/pedantigo/v2/pdcore
```

All types, functions, and constants are the same. Only the package name changed. Do NOT use import aliases — use `pdcore` directly.

New API added: `pdcore.UnmarshalInto(data []byte, target any) error` — a non-generic variant for framework integrations. Panics if no validator is registered for the target type.

**CRITICAL — registration pattern (this is NOT bare `New[T]()`):** To make a type visible to `UnmarshalInto`/framework plugins, it must be explicitly registered:

```go
var _ = pdcore.Register(pdcore.New[MyRequest]())
```

`pdcore.New[T]()` alone does NOT populate the cache `UnmarshalInto` reads from — only wrapping it in `pdcore.Register(...)` does. `Register` may be called exactly once per type; a second call for the same type panics. See `docs/plugins/web/echo.md` for the full explanation and the Echo Binder that consumes this.

---

## Echo Binder plugin

A plugin at `github.com/SmrutAI/pedantigo/v2/plugins/web/echo` replaces Echo's `DefaultBinder` with one that calls `pdcore.UnmarshalInto` on POST/PUT/PATCH bodies. See `docs/plugins/web/echo.md` for details.
