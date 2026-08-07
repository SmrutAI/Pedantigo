---
sidebar_position: 3
title: v1 to v2
description: What changed in Pedantigo v2 and how to upgrade
---

# Migrating from v1 to v2

Two things change when you move from v1 to v2:

1. **Import path and package name.** v1's `github.com/SmrutAI/pedantigo` (package `pedantigo`) becomes `github.com/SmrutAI/pedantigo/v2/pdcore` (package `pdcore`).
2. **Default struct tag.** v1's implicit default tag `pedantigo` becomes `validate`.

Everything else — constraint syntax, the Simple API, the Core API — is unchanged.

## What you need to do

1. Update your import: `github.com/SmrutAI/pedantigo` → `github.com/SmrutAI/pedantigo/v2/pdcore`.
2. Update your package qualifier in code: `pedantigo.` → `pdcore.` (find-and-replace across your codebase; watch for the `pedantigo` struct tag name literal, which is unrelated and must NOT be renamed).
3. Handle the tag rename — pick Option A or Option B below.
4. (Optional) Adopt new v2-only capabilities — `UnmarshalInto`/`Register` for framework integrations, or the Echo Binder plugin — see [New in v2](#new-in-v2).

That's it. No other API changes.

---

## Import path and package name

Per [Go's module versioning rules](https://go.dev/ref/mod), a v2+ release requires a new import path:

```go
// v1
import "github.com/SmrutAI/pedantigo"

// v2
import "github.com/SmrutAI/pedantigo/v2/pdcore"
```

```bash
go get github.com/SmrutAI/pedantigo/v2/pdcore
```

**New projects must start with v2.** v1 (`v1.1.4` and earlier) remains available and frozen at its last tag purely for existing users already depending on it — no further v1.x.x patches or features are planned, and all new feature development happens in v2 only. There is no forced upgrade timeline for existing v1 users; move to v2 when ready.

All of v2's code lives in the `pdcore` sub-package (`github.com/SmrutAI/pedantigo/v2/pdcore`, package qualifier `pdcore`) — never at the bare `github.com/SmrutAI/pedantigo/v2` path. This keeps the core validation library free of any framework dependency; optional framework integrations (like the [Echo Binder](/plugins/web/echo)) live in their own separately-versioned modules under `plugins/`. Do NOT use import aliases — use `pdcore` directly.

---

## Why the tag rename happened

`validate` is the tag name used by [go-playground/validator](https://github.com/go-playground/validator), the most widely used Go validation library, and Pedantigo already tracks its constraint syntax closely. Matching the tag name too means structs written for that ecosystem now work under Pedantigo with zero tag edits.

It also makes Pedantigo-validated structs visible to tooling that reads `validate` tags. [swaggo/swag](https://github.com/swaggo/swag), the standard tool for generating OpenAPI specs from Go source, reads `validate:"required"` to mark fields required (and has partial support for `oneof`/`min`/`max`). A struct tagged `pedantigo:"..."` was invisible to swag; the same struct tagged `validate:"..."` now produces a more accurate OpenAPI spec, and by extension more accurate generated client SDKs, with no extra annotation work.

---

## What breaks if you don't act

If your structs rely on the implicit default — no `SetTagName()` call, tags written as `pedantigo:"..."` — upgrading to v2 does **not** produce a compile error or a panic. It fails silently: Pedantigo v2 looks for `validate` tags by default, finds none on your fields, and treats them as unconstrained. Validation that used to run stops running, with no error to signal it.

This is the dangerous case. Audit for it before upgrading — search your codebase for `pedantigo:"` tags.

**Caution, the reverse case:** if your codebase already has unrelated `validate:"..."` tags on structs — left over from a different library, or dead annotations — Pedantigo v2 will now read and enforce them by default. Check for this too.

---

## Handling the tag rename

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

Constraint syntax, the Simple API (`pdcore.Unmarshal`, `pdcore.Validate`), the Core API (`pdcore.New[T]()`), schema generation, and custom validator registration all work exactly as they did in v1 — just under the `pdcore` package name.

---

## New in v2

These are optional. Nothing below is required to migrate — adopt them when useful.

### UnmarshalInto and Register

`pdcore.UnmarshalInto(data []byte, target any) error` is a non-generic variant of `Unmarshal` for framework integrations that only have a `reflect.Type`/`any` at the call site (e.g. a web framework's binder). It looks up the validator registered for `target`'s concrete type and validates against it.

To make a type visible to `UnmarshalInto`, register it once, at package init time:

```go
var _ = pdcore.Register(pdcore.New[MyRequest]())
```

`pdcore.New[T]()` alone does NOT populate the cache `UnmarshalInto` reads from — only wrapping it in `pdcore.Register(...)` does. `Register` may be called exactly once per type; a second call for the same type panics.

### Echo Binder plugin

A plugin at `github.com/SmrutAI/pedantigo/v2/plugins/web/echo` replaces Echo's `DefaultBinder` with one that calls `pdcore.UnmarshalInto` on POST/PUT/PATCH bodies, so `c.Bind()` validates automatically. See [Echo Binder Plugin](/plugins/web/echo) for full details.
