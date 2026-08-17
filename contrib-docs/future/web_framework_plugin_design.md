# Web Framework Plugin Design

Analysis of which Go web frameworks expose a global hook automatic pedantigo
integration can use, for both request decoding and response encoding.
"Automatic" means one setup call at startup — no per-handler code.

## Framework Capability Matrix

| Framework | Direction | Hook(s) | Automatic? | Details |
|---|---|---|---|---|
| Echo | Request | `Binder` | Yes | Built — `plugins/web/pedantigoecho/v2` |
| Echo | Response | `JSONSerializer` | Yes | [1](#1-echo-response-plugin) |
| Gin | Request | `codec/json.API` + `binding.Validator` | Yes, needs both overridden | [2](#2-gin-request-plugin) |
| Gin | Response | `codec/json.API` | Yes | [3](#3-gin-response-renderers) |
| Chi | Request | none in core | No | [5](#5-chi-has-no-hook) |
| Chi | Response | none in core | No | [5](#5-chi-has-no-hook) |

## 1. Echo Response Plugin

1. Mirrors the existing `Binder` plugin's shape: implement `JSONSerializer`
   (`Serialize`/`Deserialize`), set it on `e.JSONSerializer`.
2. `Serialize` looks up the registered `Validator[T]` for the response type
   and calls its existing `Marshal()` method.
3. `Validator[T].Marshal()` already validates before marshaling
   (`validator.go:782`) — no new pedantigo-core code needed, only the plugin
   wrapper.

## 2. Gin Request Plugin

1. Two separate global seams exist, not one: `codec/json.API` (decode) and
   `binding.Validator` (validation). Both route through the same call.
2. `binding.jsonBinding`'s `decodeJSON()` calls both, sequentially,
   unconditionally: decode via `json.API.NewDecoder`, then
   `binding.Validator.ValidateStruct(obj)`. Overriding `json.API` alone does
   not disable the second call.
3. See [4](#4-gin-redundant-validator-pass---and-the-non-json-validation-gap)
   for why the second call must be redirected to real pedantigo validation,
   not just neutralized — for six of Gin's nine binding types, it is the
   only validation call that ever runs.

## 3. Gin Response Renderers

All of Gin's JSON-family renderers route through `json.API` for the core
encode step; each adds its own post-processing on top. Overriding `json.API`
does not lose any of this.

| Renderer | Codec call | Extra processing |
|---|---|---|
| `JSON` | `Marshal` | none |
| `IndentedJSON` | `MarshalIndent` | 4-space formatting |
| `SecureJSON` | `Marshal` | prepends security prefix on array output |
| `JsonpJSON` | `Marshal` | wraps in JSONP callback |
| `AsciiJSON` | `Marshal` | escapes non-ASCII to `\uXXXX` |
| `PureJSON` | `NewEncoder` | disables HTML escaping |

Non-JSON bindings (Form, Query, URI, Header, XML, YAML, TOML, MsgPack,
ProtoBuf, BSON) never touch `json.API` — unaffected by the override.

## 4. Gin Redundant Validator Pass — and the Non-JSON Validation Gap

1. go-playground/validator's default tag name is `validate` — identical to
   pedantigo's default tag. Left at its default, `binding.Validator` would
   attempt to re-parse and enforce the same `validate:"..."` tags pedantigo
   already evaluated during decode.
2. This is not a safe no-op: pedantigo-only tag syntax (`defaultFactory=...`,
   custom keywords) is not valid go-playground syntax and can fail the
   second pass.
3. **An unconditional no-op is wrong, not just redundant-but-safe.** Verified
   by reading every binding implementation
   (`binding/{json,xml,query,yaml,toml,header,form,uri}.go`): all of them
   except `plain.go` call the same shared `validate(obj)` helper, which calls
   `binding.Validator.ValidateStruct(obj)`. Only the JSON path ever reaches
   pedantigo at all (via `json.API`). An unconditional no-op `ValidateStruct`
   would silently remove **all** validation — not just the redundant
   go-playground pass — for XML, YAML, TOML, Query, Form, Header, and Uri
   requests, since nothing else ever validated those.
4. Fix: `ValidateStruct` calls a new pedantigo-core primitive,
   `validator.ValidateInto(obj any) error` — see
   [8.1](#81-validatestruct-real-validation-not-a-no-op). This makes
   pedantigo's rules the actual validation for every content type Gin
   supports (JSON gets a second, redundant-but-harmless pass; the other six
   get real validation they never had before).

## 5. Chi Has No Hook

1. Core chi ships no `Binder`/`Renderer` abstraction. Deliberate minimalism —
   "100% compatible with net/http," confirmed in chi's own README.
2. `go-chi/render`'s `Bind()`/`Renderer` conventions require an explicit call
   in every handler and every response type implementing an interface method
   — ruled out, not automatic.
3. `middleware.WrapResponseWriter`-based workarounds (buffer + rewrite) act
   on raw bytes after the handler runs, with no Go type information —
   structurally weaker than a typed hook, not viable for a pedantigo plugin.

## 6. Gin Error-Handling Path

1. `c.Bind()` (via `MustBindWith`) is type-agnostic: any non-nil error from
   the binding becomes a 400 (or 413 for `*http.MaxBytesError`), regardless
   of whether it came from decode or validation. No dependency on error type.
2. `c.ShouldBindJSON()` does not auto-abort; handler code commonly
   type-asserts the returned error as `validator.ValidationErrors` to build
   structured per-field responses. That assertion must be updated to
   pedantigo's `*validator.ValidationError` type once pedantigo does the
   validation — the concrete error type changes.

## 7. Blocker Status

What must be resolved before each plugin works correctly, and where the fix
has to live — hook existence (the matrix above) does not by itself mean a
plugin is ready to build.

| Feature | Blocked? | On what | Fix lives in |
|---|---|---|---|
| Echo request | No — shipped | — | — |
| Echo response | Yes | `MarshalInto` does not exist yet | Pedantigo core |
| Gin request | Yes | `binding.Validator` adapter must call `ValidateInto` — [4](#4-gin-redundant-validator-pass---and-the-non-json-validation-gap) | Pedantigo core (`ValidateInto`) + plugin package |
| Gin response | Yes | `MarshalInto` — same gap as Echo response | Pedantigo core |
| Chi (either direction) | Yes, permanently | No hook in the framework — [5](#5-chi-has-no-hook) | Not fixable |

A blocker is anything that must be resolved before the feature is correct,
regardless of whether the fix is new pedantigo-core API or plugin-local code.
"Doesn't need new pedantigo-core API" is not the same as "not blocked" — the
Gin request no-op validator is small and plugin-local, but skipping it leaves
the plugin broken for any struct using pedantigo-only tag syntax that
go-playground's validator can't parse.

## 8. Gin Request Plugin — Verified Implementation Plan

Verified against real clones of both dependencies (`others/gin`,
`others/validator`), not summaries. Seven items, each with its concrete fix.

| # | Problem | Verified against | Fix |
|---|---|---|---|
| 1 | Non-JSON bindings never validated at all, tag-name collision on JSON | `binding.go:72` default `Validator`; `binding/{xml,query,yaml,toml,header,form,uri}.go` all call `validate(obj)`; `others/validator` default tag `validate` | Real `StructValidator` backed by `ValidateInto` — [8.1](#81-validatestruct-real-validation-not-a-no-op) |
| 2 | `Decoder` needs 3 methods, not 1 | `codec/json/json.go` — `Decoder` interface (`UseNumber()`, `DisallowUnknownFields()` return nothing; `Decode(v any) error` returns an error) | Implement all 3 per their actual signatures — [8.2](#82-decoder-implementation) |
| 3 | `UseNumber`/`DisallowUnknownFields` have no pedantigo per-call equivalent | `binding/json.go:19,25` — both default `false`, exported `var`s | Checked once in `NewBinder()`, panics there if either is `true` — [8.2](#82-decoder-implementation) |
| 4 | Does the `json.API` fallback capture the right default? | `codec/json/{json,go_json,jsoniter,sonic}.go` — one `init()` runs before any user code | Confirmed safe, no fix needed |
| 5 | `ShouldBindJSON` handlers type-assert on go-playground's error type | `binding/json.go:49` — `decodeJSON` returns before `validate()` on decode failure | `AsValidationError()` helper — [8.3](#83-error-migration-helper) |
| 6 | `c.Bind()` behavior after plugin install | `context.go` `MustBindWith` — type-agnostic | Confirmed no change needed |
| 7 | An internal pedantigo panic (e.g. registry bug) must never cross a Gin interface boundary | `binding.go` doc comment: `ValidateStruct` "should never panic, even if the configuration is not right"; `Decode(v any) error` and `ValidateStruct(any) error` both have an `error` return channel to use instead | `recover()` inside both, converted to an explanatory `error` — [8.1](#81-validatestruct-real-validation-not-a-no-op), [8.2](#82-decoder-implementation) |

### 8.1 ValidateStruct — Real Validation, Not a No-op

Two pieces: a new pedantigo-core primitive, and the Gin-side adapter that
uses it.

**Pedantigo core (`validator/registry.go`)** — `ValidateInto` mirrors
`UnmarshalInto`'s registry lookup exactly, with one deliberate divergence:
it must never panic, because `StructValidator.ValidateStruct` is called
automatically for every bound value regardless of whether its type was ever
`Register()`'d, and Gin's own interface contract forbids panicking from
this method.

```go
// ValidateInto looks up the registered Validator[T] for obj's concrete type
// and validates it. Unlike UnmarshalInto, it does not panic when no
// validator is registered for the type — it returns nil instead, matching
// binding.StructValidator's documented "never panic" contract, since this
// function is meant to be called from framework validator adapters for
// every value they see, not just ones a caller deliberately registered.
func ValidateInto(obj any) error {
    typ := reflect.TypeOf(obj)
    cached, ok := validatorCache.Load(typ)
    if !ok {
        return nil
    }
    return cached.(validatableInto).validateInto(obj)
}

// validatableInto is a non-generic interface that allows type-erased
// validation, mirroring unmarshalable/unmarshalInto.
type validatableInto interface {
    validateInto(obj any) error
}
```

`Validator[T]` implements `validatableInto` the same way it implements
`unmarshalable` (`validator.go`, next to `unmarshalInto`):

```go
func (v *Validator[T]) validateInto(obj any) error {
    typed, ok := obj.(*T)
    if !ok {
        return fmt.Errorf("validator: ValidateInto got %T, want %T", obj, typed)
    }
    return v.Validate(typed)
}
```

**Gin adapter (`plugins/web/pedantigogin/v2`)** — calls `ValidateInto`, recovering any
internal panic into an explanatory `error` rather than letting it cross the
`StructValidator` boundary:

```go
type pedantigoValidator struct{}

func (pedantigoValidator) ValidateStruct(obj any) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("pedantigo: internal panic during validation of %T: %v", obj, r)
        }
    }()
    return validator.ValidateInto(obj)
}

func (pedantigoValidator) Engine() any { return nil }
```

For JSON, this is a second, redundant-but-harmless pedantigo validation pass
(the struct was already validated once during `UnmarshalInto`). For every
other content type Gin binds (XML, YAML, TOML, Query, Form, Header, Uri),
this is the *only* validation that ever runs on it — before this fix, an
unconditional no-op here would have silently skipped validation for all of
them, not just skipped a redundant go-playground pass. `plain.go` never
calls `validate(obj)` at all (confirmed by reading it in full) — Plain
bindings stay unvalidated by Gin's own design, nothing this adapter can
change.

### 8.2 Decoder Implementation

```go
type pedantigoDecoder struct{ r io.Reader }

// UseNumber and DisallowUnknownFields return nothing — Gin gives them no
// error channel, so they must never panic either. Both are unreachable in
// practice: NewBinder() already refused to install if Gin's own
// EnableDecoderUseNumber/EnableDecoderDisallowUnknownFields flags were set
// to true, so Gin never calls these when it matters.
func (d *pedantigoDecoder) UseNumber()             {}
func (d *pedantigoDecoder) DisallowUnknownFields() {}

func (d *pedantigoDecoder) Decode(v any) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("pedantigo: internal panic during decode into %T: %v", v, r)
        }
    }()
    data, readErr := io.ReadAll(d.r)
    if readErr != nil {
        return readErr
    }
    return validator.UnmarshalInto(data, v)
}
```

The loud failure for `UseNumber`/`DisallowUnknownFields` moves to setup
time, where a panic doesn't cross any Gin interface contract because
`NewBinder()` is a plain function Gin never calls — full function in
[8.4](#84-setup--newbinder-no-global-tag-name-default).

Both flags default to `false` in Gin, matching pedantigo's own default
(`ExtraIgnore`) — the check only fires for apps that explicitly opt into
either flag before calling `NewBinder()`. An app that sets either flag
*after* `NewBinder()` has already run is not caught by this check — an
acknowledged, currently deferred gap, same ordering-hazard class as the
tag-name issue this design already solves for `Register()`.

### 8.3 Error Migration Helper

```go
func AsValidationError(err error) (*validator.ValidationError, bool) {
    var ve *validator.ValidationError
    return ve, errors.As(err, &ve)
}
```

### 8.4 Setup — `NewBinder()`, No Global Tag-Name Default

Named `NewBinder()` to match the naming already used by the shipped Echo
plugin (`pedantigoecho.NewBinder() *PedantigoBinder`, set via
`e.Binder = pedantigoecho.NewBinder()`). The shape necessarily differs —
Echo has one field (`e.Binder`) to assign a constructed value to; Gin has no
equivalent injection point, only two independent package-level globals
(`json.API`, `binding.Validator`), so the Gin version performs those
assignments itself as a side effect rather than returning a value for the
caller to assign. The name stays consistent even though the call shape does
not.

An earlier version of this design had `NewBinder()` call
`validator.SetTagName("binding")` as a default. **Rejected — broken by Go's
own initialization order, not just a rough edge.** `Register()` calls are
written as package-level variable initializers
(`var _ = validator.Register(validator.New[T]())`), which Go runs — across
every imported package — before `main()` starts. `NewBinder()` called from
`main()` always runs *after* those, never before, so `SetTagName`'s
panic-on-late-call guard would fire in any real program using this pattern.
There is also no reliable way to force `NewBinder()`'s package to initialize
before sibling packages that call `Register()` — Go does not guarantee
relative init order between packages with no import relationship.

`NewBinder()`'s two global reassignments (`json.API`, `binding.Validator`)
have no ordering dependency — safe to call from `main()` regardless of when
`Register()` calls fire. `validator.SetTagName` is not used at all, for the
reasons above. Instead, `NewBinder()` calls `RequireSingleRegisteredTagName`
([8.5](#85-register-single-tag-name-constraint)), which has no such ordering
problem — it works whether called before, after, or interleaved with
`Register()` calls.

The expected tag name is a configurable parameter, not hardcoded, defaulting
to `"binding"` to match Gin's own ecosystem convention. The full function,
including the setup-time panic checks from [8.2](#82-decoder-implementation):

```go
func NewBinder(opts ...Option) {
    if binding.EnableDecoderUseNumber {
        panic("pedantigo/gin: EnableDecoderUseNumber has no pedantigo equivalent")
    }
    if binding.EnableDecoderDisallowUnknownFields {
        panic("pedantigo/gin: EnableDecoderDisallowUnknownFields is not honored " +
            "per-call — configure ExtraForbid on this type's Validator[T] instead")
    }
    cfg := binderOptions{tagName: "binding"}
    for _, opt := range opts {
        opt(&cfg)
    }
    validator.RequireSingleRegisteredTagName(cfg.tagName)
    json.API = pedantigoCodec{fallback: json.API}
    binding.Validator = pedantigoValidator{}
}

func WithTagName(name string) Option {
    return func(o *binderOptions) { o.tagName = name }
}
```

Default usage — no configuration, `binding:"..."` tags:
```go
pedantigogin.NewBinder()
```

Overridden — e.g. to match go-playground's `validate` convention instead:
```go
pedantigogin.NewBinder(pedantigogin.WithTagName("validate"))
```

### 8.5 Register() Single-Tag-Name Constraint

Core pedantigo (`registry.go`), not plugin-specific — applies identically to
Gin and Echo, and to any future framework plugin.

**Constraint:** all types passed through `Register()` in a process must use
exactly one tag name. `validator.New[T](...)` used standalone, without
`Register()`, is unrestricted — any tag name, no cross-type consistency
requirement, no interaction with this constraint at all. The restriction
exists only because `Register()`'s purpose is enabling framework plugins to
look up a type at runtime by `reflect.Type` alone (`UnmarshalInto`,
`MarshalInto`) — automatic binding has no way to know which tag name applies
to an arbitrary runtime type unless every registered type agrees on one.

**Enforcement — one shared `atomic.Value`, checked from two entry points:**

```go
var registeredTagName atomic.Value // the one tag name every Register()'d type must use

func Register[T any](v *Validator[T]) *Validator[T] {
    // ...existing duplicate-type panic...
    if got, ok := registeredTagName.Load().(string); ok {
        if v.tagName != got {
            var zero T
            panic(fmt.Sprintf(
                "pedantigo: %s registered with tag name %q, but Register() has "+
                    "already fixed the tag name to %q — Register() only supports "+
                    "a single tag name per process",
                reflect.TypeOf(zero), v.tagName, got))
        }
    } else {
        registeredTagName.Store(v.tagName)
    }
    return v
}

// RequireSingleRegisteredTagName is called once by a framework plugin's setup
// (Gin's NewBinder(), Echo's binder setup). If a type was already Register()'d,
// verifies its tag name matches want. If nothing has been Register()'d yet,
// seeds want as the required tag name, so anything Register()'d afterward is
// checked against it automatically by Register() itself.
func RequireSingleRegisteredTagName(want string) {
    if got, ok := registeredTagName.Load().(string); ok {
        if got != want {
            panic(fmt.Sprintf("pedantigo: registered tag name %q does not match "+
                "this framework setup's expected tag name %q", got, want))
        }
        return
    }
    registeredTagName.Store(want)
}
```

**Cases this covers, all via the same variable, no separate bookkeeping:**

| Case | Behavior |
|---|---|
| First-ever `Register()` call | Establishes the one allowed tag name for the process |
| Later `Register()` call, matching tag name | Passes |
| Later `Register()` call, different tag name | Panics, at that call site, immediately |
| Framework setup runs, nothing registered yet | Passes silently, seeds `want` for future `Register()` calls |
| Framework setup runs, types already registered, all matching `want` | Passes |
| Framework setup runs, types already registered with a different name | Panics |
| `Register()` called after framework setup, matching `want` | Passes |
| `Register()` called after framework setup, different tag name | Panics — caught by `Register()`'s own check, not a separate late check |
| Two frameworks' setups called with different `want` values | Second call panics — only one active expectation per process |
| `validator.New[T](...)` without `Register()`, any tag name, never | Unaffected — this constraint only ever touches `Register()`'d types |
