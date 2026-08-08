---
sidebar_position: 3
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Performance Optimization

How Pedantigo achieves its speed, and how to get the most out of it. Every number on this page comes from a real
benchmark run in the companion [pedantigo-benchmarks](https://github.com/SmrutAI/pedantigo-benchmarks) repository —
see the full [Benchmark Results](https://github.com/SmrutAI/pedantigo-benchmarks/blob/main/BENCHMARK.md) for the
complete cross-library comparison and raw data behind every figure here.

## How Is Pedantigo So Fast?

Compared to [go-playground/validator](https://github.com/go-playground/validator) — the most direct competitor,
since both use struct tags — Pedantigo wins on two concrete, code-level design choices.

### 1. Generics remove a lookup playground is structurally forced to do

Playground's entry point takes an untyped value: `func (v *Validate) Struct(s interface{}) error`. Because the
concrete type is erased behind `interface{}`, playground cannot know at compile time which struct's rules apply —
it has to ask at runtime, on **every single call**:

```go
// go-playground/validator internals (validator.go)
cs, ok := v.v.structCache.Get(typ) // map[reflect.Type]*cStruct lookup, guarded by atomic.Value
```

This lookup runs not just once per `Struct()` call but once per nested struct too — `validateStruct` calls itself
recursively for every embedded struct field, repeating the same lookup each time.

Pedantigo's `Validator[T]` is generic over `T`. When you call `validator.New[User]()`, the Go compiler already
knows the type — there is nothing to look up. The resulting `*Validator[T]` carries its own precomputed
`fieldCache` as a plain struct field, built once inside `New()`. `Validate()` reads that field directly — no
map read, no type dispatch, at call time, ever.

### 2. The cache itself has a flatter, more cache-friendly shape

Pedantigo's field cache is a plain slice of value structs — one contiguous block of memory:

```go
type FieldCache struct {
    Fields []CachedField // indexed by struct field order
}
```

Playground's equivalent is a slice of *pointers* to individually heap-allocated structs, and each field's parsed
constraints form a *linked list* rather than a slice:

```go
// go-playground/validator internals (cache.go)
type cStruct struct {
    fields []*cField // pointer per field, not value
}
type cTag struct {
    next *cTag // constraints on one field are a linked list
}
```

Validating one field with several constraints in playground means chasing several separate pointers scattered
across the heap. Pedantigo walks one contiguous slice.

### What this does *not* mean

Neither library eliminates reflection outright. Both still call `reflect.Value.Field(i)` to actually read each
field's value at validate time — generics don't remove that. What generics remove specifically is the *"which
type is this, and where's its cache"* dispatch step, not the field read itself.

### Measured consequence

| Benchmark | Pedantigo | Playground | Gap |
|-----------|-----------|------------|-----|
| `Validate_Simple` | 560 ns / 10 allocs | 801 ns / 7 allocs | 1.43x faster — despite *more* allocations |
| `Validate_Complex` | 963 ns / 15 allocs | 1.43 µs / 9 allocs | 1.48x faster |
| `JSONValidate_Simple` | 1.41 µs / 19 allocs | 1.67 µs / 16 allocs | 1.19x faster |
| `Marshal_Simple` | 711 ns / 11 allocs | 973 ns / 9 allocs | 1.37x faster |

Playground consistently does *fewer* allocations yet is still slower — consistent with the cost being per-call
lookup and pointer-chasing overhead, not allocation count. See the [full benchmark
report](https://github.com/SmrutAI/pedantigo-benchmarks/blob/main/BENCHMARK.md) for every library and category,
including `Unmarshal` (single-call decode+validate), where Pedantigo is also fastest against its two closest peers
(godasse, godantic).

---

## What `New()` Actually Does

`New[T]()` does the expensive work once. It walks the struct via reflection, resolves every constraint tag, and
builds the `fieldCache` described above, plus the JSON field deserializers used by `Unmarshal`. That one-time cost
is real:

| Struct | Time | Allocations |
|--------|------|--------------|
| Simple (5 fields) | 6.70 µs | 129 |
| Complex (nested) | 16.18 µs | 299 |

Every other operation — `Validate`, `Unmarshal`, `Marshal`, `Schema` — reuses that precomputed cache instead of
redoing this work, which is why they run in the hundreds-of-ns to low-µs range:

| Operation | Simple | Complex |
|-----------|--------|---------|
| `Validate` | 560 ns / 10 allocs | 963 ns / 15 allocs |
| `Unmarshal` (decode + validate, one call) | 1.94 µs / 39 allocs | 6.47 µs / 122 allocs |
| `Marshal` | 711 ns / 11 allocs | — |
| `Schema()` — first call | 11.38 µs / 227 allocs | — |
| `Schema()` — cached | 9 ns / 0 allocs | — |

Schema caching is the largest single win on this page: **~1,264x** faster once cached (11.38 µs → 9 ns), not the
"240x" figure quoted elsewhere in older docs — that number was never benchmarked and should be treated as
superseded by the figures on this page.

This only pays off if the `*Validator[T]` returned by `New` is built once and reused — not recreated per request.
See [Best Practices](#best-practices) below.

---

## Simple API vs Validator API

Both APIs exist for different reasons — and it is *not* "one is slow, one is fast." Measured directly (3 runs
each, `Validate` on the same struct):

| | ns/op | allocs |
|---|-------|--------|
| Simple API — `validator.Validate(&user)` (goes through the internal `sync.Map` cache lookup) | ~571 ns | 10 |
| Validator API — direct instance, no lookup | ~568 ns | 10 |

The difference is ~3 ns — statistically noise, not a real cost. A `sync.Map.Load` keyed by a stable
`reflect.Type` is far cheaper than the actual validation work that follows it, so it doesn't show up as a
measurable difference in practice.

<Tabs>
<TabItem value="simple" label="Simple API" default>

**What it's actually for**: ergonomic parity with playground-style call patterns, for straightforward migration
and everyday use. It mirrors the shape of `validate.Struct(x)` — call the function, get a result, no setup:

```go
user, err := validator.Unmarshal[User](jsonData)
err = validator.Validate(&user)
schema := validator.Schema[User]()
```

It is a thin wrapper: the first call for a given type builds a `*Validator[T]` via `New[T]()` and caches it in a
package-level `sync.Map`; every later call for that type reuses the cached instance. This is *not* a performance
compromise for the common case — see the measured numbers above.

</TabItem>
<TabItem value="validator" label="Validator API">

**What it's actually for**: cases the Simple API's one-cache-per-type model structurally cannot express —

- **Custom `Options`** (`StrictMissingFields`, `ExtraForbid`, custom tag name) — the Simple API always uses
  default options; if you need different behavior, you need your own instance.
- **Multiple differently-configured validators for the same type** — e.g. a strict one for API input and a
  lenient one for internal reprocessing.
- **Explicit `Register()` control** for framework integrations (the Echo Binder plugin, `UnmarshalInto`) that
  look up a validator by `reflect.Type` at runtime.
- **Paying `New()`'s one-time cost explicitly**, at startup, under your own control, rather than implicitly on
  first use.

```go
userValidator := validator.New[User](validator.Options{
    ExtraFields: validator.ExtraForbid,
})
user, err := userValidator.Unmarshal(jsonData)
```

</TabItem>
</Tabs>

---

## Best Practices

### 1. Hold your validator at module level, not per-call

`New[T]()`'s cost (6.70 µs–16.18 µs, see above) is only paid once if you follow this pattern:

```go
// Good - built once at package init, reused for the life of the program
var userValidator = validator.New[User]()

func handleCreateUser(body []byte) (*User, error) {
    return userValidator.Unmarshal(body) // reuses the cached field constraints
}
```

```go
// Bad - pays the full New() cost on every single call
func handleCreateUser(body []byte) (*User, error) {
    userValidator := validator.New[User]()
    return userValidator.Unmarshal(body)
}
```

If you only ever call the package-level Simple API functions (`validator.Unmarshal[User](...)`, etc.), you get
this same amortization automatically via the internal cache — no module-level variable needed.

### 2. Use `Register()` when a framework needs to find your validator by type

Some integrations — the Echo Binder plugin, `validator.UnmarshalInto` — only have a `reflect.Type` at runtime, not
your module-level variable. `Register()` makes a validator discoverable that way:

```go
var _ = validator.Register(validator.New[User]())
```

`Register[T]` may be called **exactly once per type** — a second call for the same type panics, by design. A type
could have multiple differently-configured validators (different `Options`), and Pedantigo has no way to guess
which one a framework plugin should resolve to, so it refuses to silently pick one:

```go
var _ = validator.Register(validator.New[User]())
var _ = validator.Register(validator.New[User]()) // panics: already registered
```

Call `Register` from exactly one package-level `var` declaration per type. If a type doesn't need framework
lookup, plain `New()` plus a module-level variable (pattern 1 above) is enough — `Register` is not required just
to get caching.

### 3. Cache schemas the same way

`Schema()`, `SchemaOpenAPI()`, and their JSON variants are cached internally after the first call per validator
instance, but if you serve the same schema repeatedly (e.g. an HTTP endpoint), avoid the repeated cache-hit call
entirely by storing the result once:

```go
var userValidator = validator.New[User]()
var userSchema = userValidator.Schema() // computed once at init

func sendSchema(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(userSchema)
}
```

### 4. Profile before optimizing further

The gap between the Simple API and Validator API is noise for the common case (see above) — don't switch to the
Validator API purely for speed unless you've actually measured a hot path where it matters. Do switch if you need
custom `Options` or framework registration, regardless of speed.

```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

---

## Further Reading

- [Full benchmark results](https://github.com/SmrutAI/pedantigo-benchmarks/blob/main/BENCHMARK.md) — every
  library, every category, raw numbers.
- [Simple API reference](../api/simple-api.md)
- [Validator API reference](../api/validator.md)
