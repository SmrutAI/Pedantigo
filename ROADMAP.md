# ROADMAP — pedantigo

Planned work that is decided but not yet implemented. Each item states the
problem, the design, the exact code it touches, its dependencies, and its
acceptance criteria, so any contributor can pick it up without re-deriving the
decision.

## Walker-aware custom (de)serialization

Context for the two items below. pedantigo's `Unmarshal` owns deserialization
(a reflection walker in `validator/internal/deserialize`) specifically so it can
detect a field's **absence** in the raw JSON and thereby enforce `required` and
inject `default` — the capability that distinguishes it from validate-only
libraries (see `CLAUDE.md` → "Core Semantics"). The cost of owning
deserialization is that the walker must re-implement what `encoding/json` does,
**including `json.Unmarshaler` dispatch — which it currently does not**. A field
whose Go `Kind` does not match the JSON shape the walker assumes for that `Kind`
(e.g. a `struct`-typed field whose wire form is a bare string or a bare array,
implemented via a custom `UnmarshalJSON`) therefore fails in
`SetFieldValueWithOptions` (`validator/internal/deserialize/setter.go`) at the
generic `AssignableTo`/`ConvertibleTo` fallthrough. pedantigo's own `SecretStr`
(`validator/secret.go`) has this latent defect; it is masked only because its
in-struct test (`validator/secret_test.go` → `TestSecretStr_InStruct`) decodes
via stdlib `json.Unmarshal`, never through `Validator.Unmarshal`.

### Phase 1 — `WalkerDecoder` decode hook (primary feature)

Give a custom type control over how it is filled from the decoded JSON value,
**without** losing `required`/`default` enforcement on anything nested inside it.

- **Interface** (defined in `internal/deserialize`, re-exported as a type alias
  in package `validator` to avoid the `validator → deserialize` import cycle):

  ```go
  type WalkerDecoder interface {
      // decoded is the already-json-decoded generic value for this field
      // (map[string]any | []any | string | float64 | bool | nil), reusing the
      // walker's single top-level decode. For any nested struct/slice, the
      // implementation MUST call recurse(dst, decoded) instead of json.Unmarshal,
      // so required/default/constraints run at every deeper level.
      DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error
  }
  ```

- **Hook site**: `SetFieldValueWithOptions` (`internal/deserialize/setter.go`),
  immediately after the nil-handling block and **before** the `time.Time`
  branch. Both top-level fields (via the field deserializer) and nested values
  (via `recursiveSetFunc`) funnel through this function, so one check covers all
  depths. The `recurse` argument adapts the walker's internal
  `func(reflect.Value, any, reflect.Type) error` to a pointer-based
  `func(dst any, decoded any) error`.

- **Detection**: a runtime O(1) type assertion
  (`fieldValue.Addr().Interface().(WalkerDecoder)`) — same cost class as the
  existing `any(obj).(Validatable)` check in `Validate` (`validator.go`), and
  consistent with the deserialize path already re-parsing tags per call for
  absent nested fields (`setter.go`, `deserializeStructFields` →
  `ParseTagWithName`). Requires `fieldValue` to be addressable (holds at every
  current call site: top-level via `reflect.ValueOf(&obj).Elem()`, slice
  elements via `MakeSlice`, nested via `reflect.New(elemType).Elem()`).

- **Leaf fallback**: if a field does not implement `WalkerDecoder` but does
  implement `json.Unmarshaler`, the walker re-marshals the decoded value back to
  bytes and calls `UnmarshalJSON`. This fixes `SecretStr`/`SecretBytes` through
  the walker for free. Known limitation: a leaf type needing exact original-byte
  fidelity (e.g. number precision) should implement `WalkerDecoder` and inspect
  the decoded value directly rather than rely on this re-marshal fallback.

- **Marshal**: no change. `Marshal` (`validator.go`) delegates to
  `json.Marshal(obj)`, which already honors `MarshalJSON`. (`MarshalWithOptions`
  uses the separate `internal/serialize` walker; audit it before relying on
  custom `MarshalJSON` under context-exclusion marshal.)

- **Test-coverage gap to close**: add tests that route a `WalkerDecoder` type
  **and** a plain `json.Unmarshaler` type (e.g. `SecretStr`) through
  `Validator.Unmarshal`/`UnmarshalInto` with `StrictMissingFields: true`,
  asserting nested `required`/`default` still fire. This is the exact path with
  no existing coverage today.

### Phase 2 — schema fidelity for `WalkerDecoder` types (this item)

`Schema()`/`SchemaJSON()`/`SchemaOpenAPI()`/`SchemaLLM()` currently derive a
field's schema from its **reflected Go type** (via `invopop/jsonschema`), which
does not match a `WalkerDecoder` type's wire shape: a `struct` whose wire form is
`string | array` emits an object schema of the struct's Go fields instead of the
intended `oneOf`. Until this lands, `WalkerDecoder` types emit Go-shape schema.

This blocks nothing in the decode path (Phase 1 is independent) and is deferred
to a follow-up, but pedantigo is a general library, so schema output must
eventually be correct for these types.

- **Design**: an **optional** companion interface a `WalkerDecoder` type may also
  implement:

  ```go
  type SchemaShaper interface {
      SchemaShape() *jsonschema.Schema
  }
  ```

  When present, the schema generator uses the type's returned schema for that
  field instead of reflecting its Go struct. This mirrors the existing
  discriminated-union path — `UnionValidator.Schema()` (`validator/union.go`)
  already emits `oneOf` via `schemagen.GenerateUnionSchema` — so reuse that
  approach rather than inventing a second schema mechanism.

- **Caching (Rule 5)**: the result must be cached per `Validator[T]` instance
  with the existing double-checked-locking pattern (`schemaMu`, and the
  `cachedSchema`/`cachedSchemaJSON`/`cachedOpenAPI`/`cachedSchemaLLM` fields in
  `validator.go`) — never recomputed per call. A `SchemaShaper`'s output is a
  pure function of the type, so it caches identically to every other derived
  schema artifact.

- **Touch points**: `validator/schema.go` (and the `Schema*` methods on
  `Validator[T]`) — detect `SchemaShaper` at schema-build time and splice its
  output in place of the reflected field schema, including inside the LLM
  (`SchemaJSONLLM`, no `$schema`/`$id`) and OpenAPI variants.

- **Dependency**: Phase 1 (`WalkerDecoder`) must land first — `SchemaShaper` is
  meaningful only for types that already opt out of Go-shape decoding.

- **Acceptance**: a `WalkerDecoder` + `SchemaShaper` type emits its declared
  wire schema (e.g. `oneOf[string, array]`) across `Schema`, `SchemaJSON`,
  `SchemaOpenAPI`, and `SchemaLLM`; a `WalkerDecoder` type without `SchemaShaper`
  falls back to today's reflected-Go-shape schema (no regression).
