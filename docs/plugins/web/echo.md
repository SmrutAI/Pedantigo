---
sidebar_position: 1
---

# Echo Binder Plugin

Replaces Echo's `DefaultBinder` so every `c.Bind()` call on POST/PUT/PATCH
runs `pdcore.UnmarshalInto` — enforcing required fields, defaults, and all
`validate` constraints automatically. GET/DELETE/HEAD requests fall back to
Echo's own `DefaultBinder` (path + query params only) — pedantigo is never
invoked for those methods.

## Install

```bash
go get github.com/SmrutAI/pedantigo/v2/plugins/web/echo@v2.0.0
```

## Setup

```go
import pedantigoecho "github.com/SmrutAI/pedantigo/v2/plugins/web/echo"

e := echo.New()
e.Binder = pedantigoecho.NewBinder()
```

## How it works

- **POST/PUT/PATCH:** reads the request body → calls
  `pdcore.UnmarshalInto(body, target)` → enforces required fields, defaults,
  and constraints. Returns an `echo.HTTPError` with status 400 on validation
  failure.
- **GET/DELETE/HEAD:** falls back to Echo's `DefaultBinder`
  (`BindPathParams` + `BindQueryParams`). No pedantigo validation runs for
  these methods.
- **Empty POST/PUT/PATCH body:** returns `nil` (no error).

## Registration requirement

Every request struct passed to `c.Bind()` must be registered once, at
package init time, with `pdcore.Register(pdcore.New[T]())`:

```go
var _ = pdcore.Register(pdcore.New[CreateRequest]())
```

If a type is never registered, `UnmarshalInto` panics with a message naming
the missing type and this exact fix. If `Register` is called twice for the
same type, it panics too — a type may only be registered once; find the
duplicate call and remove it.

## Before / after

Before (manual `io.ReadAll` + `Unmarshal`):

```go
var createValidator = pdcore.New[CreateRequest]()

func (h *Handler) Create(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
	}
	req, err := createValidator.Unmarshal(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// use req...
}
```

After (with the Echo Binder installed):

```go
var _ = pdcore.Register(pdcore.New[CreateRequest]())

func (h *Handler) Create(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	// use &req...
}
```
