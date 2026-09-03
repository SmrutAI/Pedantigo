# Spec: Zero per-call reflection + safe cyclic/recursive handling for Unmarshal and Validate

Status: ready-for-agent
Area: `validator/internal/deserialize`, `validator` (New[T] build path, validateWithCache)
Related: ROADMAP.md ("Walker-aware custom (de)serialization"); the WalkerDecoder feature.

This spec has **two phases**, delivered in order:
- **Phase A — Deserialize plan (behavior-preserving).** Precompute a nested plan at
  `New[T]()` so the Unmarshal hot path does zero per-call reflection.
- **Phase B — Validate deep-recursion (behavior-changing).** Deep-validate recursive
  types via a runtime visited-pointer cycle guard, and add a security depth cap.

Both phases share one build-time cycle-handling mechanism (register-before-populate
back-edge) and are documented together in a single concept doc with mermaid flowcharts.

## Problem Statement

Two problems, one area:

1. **Performance (Phase A).** A developer using pedantigo's `Unmarshal` on a request with
   nested structs pays a reflection cost **on every call**, not once. For each nested
   struct the walker re-iterates the struct's fields (`reflect.Type.NumField`) and
   re-parses its struct tags every single Unmarshal, and for slices of custom-decoder
   elements it runs `reflect.PointerTo(...).Implements(...)` per element per call. This is
   per-call reflection over the entire message tree — the opposite of pedantigo's core
   promise that all expensive reflection happens once at validator creation. The validate
   path already caches a recursive field plan at `New[T]()`; the deserialize path never
   got the equivalent.

2. **Recursive-type correctness + safety (Phase B).** The validate path stops at a type
   cycle (`buildFieldConstraints` returns nil `NestedCache`, from the #15 stack-overflow
   fix), so recursive types are under-validated below the first cycle. Separately, a
   validation/deserialization library that recurses on untrusted, deeply-nested input is a
   documented DoS target (see Security): a small, deeply-nested payload can exhaust the
   stack and crash the process.

## Solution

- **Phase A:** build a precomputed, recursive **deserialize plan** once at `New[T]()`,
  stored on the `Validator[T]`. The hot path becomes an interpreter over a flat plan:
  fields looked up by precomputed JSON key, precomputed required/default rules applied,
  custom decoders dispatched via a precomputed kind flag. The only hot-path reflection is
  a single flat (non-nested) type assertion per custom-decoder field. Observable behavior
  (values, required/default/absence, error strings and paths) is unchanged.

- **Phase B:** replace the validate path's nil-on-cycle with the same
  register-before-populate **back-edge** used by Phase A, and add a runtime **current-path
  visited-pointer set** so validation follows recursive data to its full depth without
  infinite-looping on cyclic Go values. Add a configurable **max-recursion-depth cap**
  (default 3) on self-referential recursion that returns a clear error when exceeded — a
  security control against deeply-nested/recursive payloads.

## User Stories

1. As a Go developer using pedantigo, I want `Unmarshal` of a deeply nested request to run without per-call reflection, so that my hot path is as fast as the library promises.
2. As a Go developer, I want repeated `Unmarshal` calls for the same type to reuse work computed once, so that throughput scales with request volume, not with struct depth.
3. As a Go developer, I want `required` on a nested struct field still enforced by absence detection, so that a missing key is rejected exactly as before.
4. As a Go developer, I want `default` / `defaultUsingMethod` on nested fields still applied only when the key is truly absent, so that an explicit zero value is preserved.
5. As a Go developer with a recursive type (e.g. a tree node holding a slice of itself), I want `Unmarshal` to keep decoding to the full depth of my data, so that nothing I could decode before now fails.
6. As a Go developer using a custom `WalkerDecoder` type inside a slice, I want its `DecodeWalk` still invoked per element, so that string-or-array and flatten-extra shapes keep working.
7. As a Go developer using a `json.Unmarshaler` leaf type (e.g. `SecretStr`) through the walker, I want it still honored, so that custom leaf decoding is unaffected.
8. As a Go developer using `ExtraAllow`, I want unknown-field capture to keep working without paying per-call reflection, so that forward-compat capture is both correct and fast.
9. As a Go developer, I want the exact same error message and field path (e.g. `Items[0].City`, `MultiRequiredFieldError` phrasing) on a validation/decoding failure, so that my error handling and tests do not break.
10. As a pedantigo maintainer, I want the deserialize path to obey the same "reflection once at `New[T]()`" rule the validate path obeys, so that the codebase is internally consistent and the performance guarantee holds everywhere.
11. As a pedantigo maintainer, I want one plan mechanism for the whole type tree (top-level and nested), so that there are not two parallel decoding models to maintain.
12. As a pedantigo maintainer, I want recursive/self-referential types handled by a back-edge in the plan graph, so that plan construction cannot stack-overflow and runtime recursion follows the data.
13. As a pedantigo maintainer, I want the plan built per `Validator[T]` instance (deduplicated per type via the existing validator cache), so that no new global cache and no `reflect.Type` keying discipline is introduced.
14. As a downstream consumer (the gateway), I want the WalkerDecoder-based DTO decoding to keep behaving identically, so that my Anthropic request parsing is unaffected by this internal change.
15. As a pedantigo maintainer, I want the change landed in small, independently testable steps, so that any regression on this safety-critical path bisects to one diff.
16. As a pedantigo maintainer, I want the existing test suite to remain the safety net, so that a green suite after each step is evidence the refactor preserved behavior.
17. As a pedantigo maintainer, I want a benchmark that actually exercises the core-API Unmarshal walker on a nested payload, so that this path stops being invisible to the benchmark suite.
18. As a Go developer validating a recursive structure (e.g. a tree), I want its constraints checked at every level of my data, so that deep invalid fields are caught, not silently skipped (Phase B).
19. As a Go developer, I want validation of a genuinely cyclic in-memory value to terminate safely (not infinite-loop), so that a back-reference cannot hang or crash my process (Phase B).
20. As an operator of a service using pedantigo, I want deeply-nested/recursive untrusted payloads rejected with an error rather than crashing the process, so that this DoS class is mitigated (Phase B).
21. As a Go developer whose data legitimately nests recursively beyond the default, I want to raise the depth limit via an option, so that the safe default does not block my valid use case (Phase B).
22. As a new contributor, I want a concept doc with flow diagrams for the cyclic/nested algorithm (both paths), so that I can understand the model without reading the whole implementation.

## Implementation Decisions

### Shared

- **Model: flat data plan + interpreter.** Introduce a recursive, data-only plan
  (a per-field record slice with a pointer to the child type's plan) mirroring the
  validate path's cached field metadata. The hot path is an interpreter over the flat
  slice, not a graph of per-field closures. Rationale: consistency with the validate
  path and the library rule favoring flat slices over pointer-chains. This shape came
  out of the design session; the decision-bearing fields are:
  - per field: parent field index; JSON key name (from the `json` tag); `isRequired`;
    static default; `defaultUsingMethod` name (signature validated at build, invoked
    per run — inherently dynamic); `decoderKind` ∈ {none, walkerDecoder, jsonUnmarshaler};
    `isSlice` / `isMap` / `isPointer`; element `decoderKind`; pointer to the nested
    type plan; string-transform flags.
  - per type: the ordered field slice, plus extra-field info (presence + index) for
    `ExtraAllow`.
- **Location: on the `Validator[T]` instance**, built inside `New[T]()`. Per-type
  dedup already comes from the existing per-type validator cache; no new package-level
  `sync.Map` and no `reflect.Type`-vs-`t.String()` keying concern.
- **One mechanism for the whole tree.** Replace the current top-level per-field
  closures with the same plan used for nested levels, so top-level and nested decoding
  share one path.
- **`decoderKind` precomputed once.** The single `reflect.PointerTo(t).Implements(...)`
  check per type runs at build time and is frozen into `decoderKind`. This removes the
  per-element `Implements` call the WalkerDecoder change introduced in the slice path.
- **Hot-path reflection budget.** The interpreter may use `reflect.Value.Field(i)` for
  value access (as the validate path does) and, only for a field whose cached
  `decoderKind` is non-none, exactly one flat `reflect.Value.Addr().Interface().(iface)`
  type assertion. Nothing nested, and none of `PointerTo` / `Implements` / `NumField` /
  `StructField.Tag.Get` in the hot path.
- **Recursive types: cyclic plan graph via back-edge, through a shared helper.** A single
  generic register-before-populate utility (in-progress `map[reflect.Type]*node`; on a
  cycle, return the in-progress node) is used by BOTH the deserialize plan builder and the
  validate path's `buildFieldConstraints` — one implementation of the cycle discipline, one
  test. During build, register a type's node before populating it; on a cycle, point the
  child at the in-progress node (a back-edge). At runtime the interpreter follows the
  pointer and recurses to the data's depth (decoded JSON is a finite tree). This prevents
  build-time stack overflow and preserves full-depth decoding; in Phase B it replaces the
  validate path's nil-on-cycle so validation also reaches full data depth.
- **`ExtraAllow` folded in.** Precompute per-type extra-field presence and index into
  the plan so the unknown-field capture path is also reflection-free, preserving its
  exact "which key is extra" and merge semantics.
- **No public API change.** `Unmarshal`, `UnmarshalInto`, `WalkerDecoder`, and the
  options surface are unchanged. No new exported types unless strictly required.
- **Field-name source unchanged.** The input-map key is resolved from the `json` tag,
  not `tagNameFunc` (which only affects validation error field names), so the plan is a
  pure function of the type with no dependency on the globals that
  `clearValidatorCache` wipes.
- **Staged rollout, full suite after each step:** (1) nested plan replacing the per-call
  nested reflection; (2) unify top-level closures into the plan; (3) fold in `ExtraAllow`;
  (4) validate deep-recursion + visited-pointer guard + depth cap (Phase B).

### Phase B — validate deep-recursion, cycle safety, and the depth cap

- **Runtime cycle detection: current-path visited-pointer set.** Track only pointers on the
  ACTIVE recursion path (add-on-enter, remove-on-leave — stack semantics), keyed by pointer
  identity, and only for "hard" reference-carrying kinds (pointer/interface/slice/map) as
  `reflect.DeepEqual` does, since a cycle must pass through one. A cycle = a pointer already
  on the path → stop recursing that branch (no error; the cycle is broken, remaining
  constraints still run). Shared, non-cyclic sub-objects (diamonds) are re-validated
  (removed on leave) — matching "validate all reachable".
- **State home: the pooled per-call `validateContext`.** The visited-pointer set lives in
  the existing pooled context (Rule 4): per-`Validate`-call, thread-safe, reset between
  calls. Not on the `Validator[T]` instance.
- **Preserve the `#15` re-entrancy guard.** The existing `v.validating` guard that prevents
  infinite recursion when a user's `Validatable.Validate()` calls back into pedantigo stays
  intact and independent of the new field-recursion guard.
- **Max-recursion-depth cap (security control).** A configurable limit on **self-referential
  (cyclic-type) recursion depth only** — acyclic nesting through DISTINCT types stays
  unbounded (finite by the type graph). "Depth" is the number of nested instances of the
  same recursive type along the current path, counting the outermost as depth 1.
  **Default = 3.** Exceeding it returns a clear error (e.g. `max recursion depth 3 exceeded
  at <path>`) — never a silent stop (no data loss on Unmarshal, no under-validation on
  Validate). Configurable via the existing option builders; callers whose data legitimately
  nests deeper raise it. Applies to BOTH the Unmarshal plan interpreter and the Validate
  traversal.

## Security

This addresses a real, cross-ecosystem attack class — **uncontrolled recursion on
deeply-nested input** (CWE-674 / CWE-121). An attacker who observes a recursive type in an
API's client payloads can send a small but deeply-nested/recursive JSON body to exhaust the
stack and crash the process. **Size limits alone are insufficient** — a 1000-deep array is
only ~2 KB — so explicit depth enforcement is the reliable defense. Documented incidents:

- **Jackson (Java) — CVE-2025-52999** (CVSS 8.7 HIGH): unbounded recursion on deeply-nested
  JSON → `StackOverflowError`; fixed by a default max depth (1000) that throws instead of
  crashing. https://www.herodevs.com/blog-posts/cve-2025-52999-denial-of-service-via-stack-overflow-in-jackson-core
- **Newtonsoft Json.NET (.NET) — CVE-2024-21907**: deeply-nested JSON to
  `JsonConvert.DeserializeObject` → StackOverflow DoS; network-based, unauthenticated.
  https://www.sentinelone.com/vulnerability-database/cve-2024-21907/
- **Oj (Ruby) — CVE-2026-54592**: stack buffer overflow via deeply-nested input in
  `Oj::Doc#each_child` (parser imposes no nesting-depth limit).
  https://github.com/ohler55/oj/security/advisories/GHSA-3m6q-jj5j-38c9
- **Unleash (Node.js) — CVE-2026-63462**: unauthenticated remote DoS by posting deeply-nested
  JSON to an OpenAPI-*validated* endpoint; recursion in the validation/error path exhausts
  the stack. A ~10 KB payload crashes the server and it **bypassed the body-size limit** —
  the direct analog of pedantigo's Validate path.
  https://cvereports.com/reports/CVE-2026-63462

And in Go itself — the most relevant precedent, since even the standard library's own
decoders were hit and had to add explicit recursion-depth handling:

- **`encoding/xml` `Unmarshal` — CVE-2022-30633** (CVSS 7.5 High): deeply-nested XML into a
  struct with an `any`-tagged field → panic via stack exhaustion (fixed in Go 1.17.12 / 1.18.4).
  https://github.com/golang/go/issues/53611
- **`encoding/xml` `Decoder.Skip` — CVE-2022-28131** (CVSS 7.5 High): deeply-nested XML →
  stack exhaustion.
  https://github.com/golang/go/issues/53614
- **`encoding/gob` `Decoder.Decode` — CVE-2022-30635**: deeply-nested structures → stack
  exhaustion.
  https://www.cvedetails.com/cve/CVE-2022-30635/
- **`encoding/gob` `Decoder.Decode` — CVE-2024-34156**: the 2022 gob fix was incomplete; the
  same DoS re-opened and had to be re-patched — evidence this class is easy to get wrong.
  https://github.com/golang/go/issues/69139
- **`go/parser` — CVE-2022-1962**: deeply-nested types/declarations → uncontrolled recursion
  → panic.
  https://osv.dev/vulnerability/CVE-2022-1962
- **`encoding/json`** (no CVE, but the documented risk): a ~5–17 MB nested JSON blows the
  stack — which is why Go's json scanner carries its own `maxNestingDepth` guard.
  https://github.com/golang/go/issues/31789

A validation library runs on already-decoded Go values, where the json scanner's
`maxNestingDepth` no longer applies — precisely the un-guarded surface the depth cap covers,
and none of the incumbent libraries (go-playground/validator, ozzo-validation, Huma) bound it.

Threat model and controls in pedantigo:
- **Unmarshal, Step 1 (decode to `map[string]any`, `validator.go:570`):** inherits Go stdlib
  `encoding/json`'s scanner `maxNestingDepth` guard, so the decode-to-map step is already
  protected against stack overflow.
- **Unmarshal, Step 2 (plan interpreter) and the entire Validate path:** pedantigo's OWN
  recursion — the Validate path runs on already-constructed Go values with no stdlib scanner
  in front, and is the direct analog of the Unleash CVE. Here pedantigo's
  **max-recursion-depth cap (default 3, error on exceed)** and the **visited-pointer cycle
  guard** are the defense: recursive/cyclic attacker input is rejected with an error (or a
  real in-memory cycle is broken) instead of crashing.
- The cap targets self-referential recursion specifically because that is the only way an
  attacker gets attacker-controlled unbounded depth against a fixed schema; distinct-type
  nesting is bounded by the type graph.

The concept doc (task #15) must present the depth cap as a security feature with this threat
model and these case links.

## Testing Decisions

- **What makes a good test here:** assert only externally observable behavior of the
  public Unmarshal boundary — the decoded struct value, the returned error's type,
  message text, and field path — for a given JSON input. Never assert plan internals
  (field-record shape, cache pointers); those are implementation details free to change.
- **Seam (single, highest):** the public boundary — `validator.Unmarshal[T]`,
  `Validator[T].Unmarshal`, `UnmarshalInto` for Phase A; `Validator[T].Validate` /
  `validator.Validate` for Phase B. The existing seam; no new seam introduced.
- **Prior art:** `validator/simple_api_test.go` (Unmarshal via the Simple API),
  `validator/secret_test.go` (custom `json.Unmarshaler` types),
  `validator/walker_decoder_test.go` (WalkerDecoder + json.Unmarshaler leaf + nested
  required through `Unmarshal`), and the circular-reference suite in
  `validator/validator_test.go` (the `#15` tests) already drive this seam.
- **Phase A unit tests (mandatory):** nested-struct required/absence; static default and
  `defaultUsingMethod` on nested fields; slices and maps of structs; slices whose element
  type implements `WalkerDecoder`; `json.Unmarshaler` leaf inside a struct; `ExtraAllow`
  at multiple nesting levels; a recursive type decoded deeper than one level; exact error
  text and field path for a missing-required nested field.
- **Phase B unit tests (mandatory):** a recursive tree with an INVALID field several levels
  deep is now caught (the new behavior); a genuinely cyclic in-memory value terminates with
  `NoError` on valid data (cycle broken by the visited-pointer set); a diamond (shared,
  non-cyclic sub-object) is validated on each path; self-referential nesting exceeding the
  configured cap returns the depth-exceeded error; raising the cap via the option allows
  deeper nesting.
- **Existing circular-reference tests must stay green at the default cap of 3.** The `#15`
  suite's deepest tree is 3 levels (`root → child2 → grandchild`, `validator_test.go:2488`),
  within the default; **no edit to those tests is required**. Per the standing rule, a
  pre-existing test failing during this work is a real regression to investigate, never to
  be edited/suppressed unless proven a deliberate, correct behavior change.
- **No golden-master / snapshot suite.** Rely on pedantigo's existing high coverage
  plus the added unit tests. A pre-existing test that fails during this refactor is to
  be investigated and treated as a real regression; it must not be edited, deleted, or
  suppressed to make the suite green unless the change has been proven to be a
  deliberate, correct behavior change.
- **Benchmark:** add a core-API Unmarshal-walker benchmark on a nested payload (handled
  in the benchmark-migration task) so the corrected numbers are measurable against the
  pre-refactor baseline.

## Out of Scope

- The **public API** beyond the additive depth-limit option.
- **Schema generation** for WalkerDecoder types (separate SchemaShaper task).
- The **benchmark harness v1→v2 migration** and its never-cache setup (separate task);
  this spec only assumes a walker benchmark exists to measure Phase A.
- A **non-recursive-type (total-depth) cap** and converting recursion to an explicit
  worklist — the cap here is scoped to self-referential recursion; broader iterative
  hardening, if wanted, is a separate task.
- Any change to the **Validatable re-entrancy guard** (must be left intact).

## Further Notes

- This refactor also closes the specific defect the WalkerDecoder change introduced:
  a per-element `reflect.PointerTo(elemType).Implements(...)` in the slice path. That
  was a planning error (the reflection was specified in the plan, not invented by the
  implementer); precomputing `decoderKind` removes it.
- The design follows established precedents: Go `encoding/json` (per-type field metadata
  computed once, cached, register-before-populate cycle safety), Go `reflect.DeepEqual`
  (pointer-identity visited set, only "hard" reference kinds tracked), and Tailscale
  `deephash` (current-path/stack visited set — O(depth) and deterministic). The deliberate
  divergence from the validate path's old behavior is the back-edge (decode/validate to
  full data depth) plus the depth cap.
- Publishing to an external issue tracker was not configured for this repo; this spec
  lives as a contributor design doc. Convert to a tracked issue if/when a tracker is
  wired up.
