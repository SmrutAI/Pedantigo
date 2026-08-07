---
sidebar_position: 1
---

# Plugins

Pedantigo's core (`pdcore`) has zero dependency on any web framework, RPC
library, or LLM tooling. Framework integrations live in `plugins/`, each as
its own Go module with its own `go.mod` — so using a plugin never pulls an
unwanted dependency into projects that only need core validation.

## How plugins work

Every plugin is a thin adapter around two core functions:

- `pdcore.Register[T](v *pdcore.Validator[T]) *pdcore.Validator[T]` — opts a
  specific, already-configured validator into a type-erased lookup table.
- `pdcore.UnmarshalInto(data []byte, target any) error` — looks up the
  registered validator for `target`'s type at runtime and validates against
  it, without the caller needing to know the concrete type at compile time.

A plugin's binder/adapter calls `UnmarshalInto` on your behalf whenever the
framework needs to deserialize and validate a request. You register your
types once, at startup, with `pdcore.Register(pdcore.New[T]())`.

## Categories

- **[Web Frameworks](/plugins/web)** — HTTP framework integrations (Echo
  available today).
- **[RPC](/plugins/rpc)** — coming soon.
- **[LLM](/plugins/llm)** — coming soon.

## Roadmap

Pedantigo's plugin system is deliberately structured by category (`web/`,
`rpc/`, `llm/`) so new framework and library integrations can be added
without restructuring existing docs. The Echo Binder is the first web
framework plugin; RPC and LLM-tooling plugins are planned.
