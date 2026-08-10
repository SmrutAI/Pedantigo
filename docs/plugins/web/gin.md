---
sidebar_position: 2
---

# Gin Request Plugin

Installs Pedantigo into both Gin request-binding seams:

- `codec/json.API` for JSON request decoding
- `binding.Validator` for struct validation after query/form/header/URI binding

That split is why a Gin plugin needs more than a custom binder. Replacing only
one hook would leave part of Gin's binding pipeline outside Pedantigo.

## Install

```bash
go get github.com/SmrutAI/pedantigo/v2/plugins/web/gin@v2.0.0
```

## Setup

```go
import (
	"github.com/gin-gonic/gin"

	pedantigogin "github.com/SmrutAI/pedantigo/v2/plugins/web/gin"
)

func main() {
	pedantigogin.NewBinder()

	r := gin.Default()
	// routes...
}
```

## What it changes

- `c.ShouldBindJSON`, `c.BindJSON`, and other JSON body binders decode through
  `validator.UnmarshalInto`.
- `c.ShouldBindQuery`, `c.ShouldBind`, `c.ShouldBindHeader`, and
  `c.ShouldBindUri` validate through `validator.ValidateInto` after Gin populates
  the target struct.
- Internal Pedantigo panics are recovered at the Gin boundary and returned as
  ordinary `error`s instead of escaping through Gin.

## What Pedantigo can and cannot enforce in Gin

Gin's request binding paths are not symmetric. JSON is decoded by Pedantigo
from raw bytes, while query/form/header/URI binding happens inside Gin first
and only then passes a populated struct to Pedantigo.

| Gin path | Pedantigo hook | Sees raw input before struct creation? | Can detect missing `required` fields? | Can apply `default=` / `defaultUsingMethod=` for missing fields? | Can validate ordinary value constraints (`email`, `min`, etc.)? |
|---|---|---:|---:|---:|---:|
| JSON body (`ShouldBindJSON`, `BindJSON`) | `codec/json.API.NewDecoder` → `validator.UnmarshalInto` | Yes | Yes | Yes | Yes |
| Query (`ShouldBindQuery`) | `binding.Validator.ValidateStruct` → `validator.ValidateInto` | No | No | No | Yes |
| Form (`ShouldBind`, `ShouldBindWith(Form)`) | `binding.Validator.ValidateStruct` → `validator.ValidateInto` | No | No | No | Yes |
| Header (`ShouldBindHeader`) | `binding.Validator.ValidateStruct` → `validator.ValidateInto` | No | No | No | Yes |
| URI (`ShouldBindUri`) | `binding.Validator.ValidateStruct` → `validator.ValidateInto` | No | No | No | Yes |

For the non-JSON rows above, once Gin has already populated the struct,
Pedantigo no longer knows whether a zero value came from:

- a field that was absent from the request, or
- a field that was present but parsed to the type's zero value

So both of these are genuine limitations for those paths and are not something
the plugin can reconstruct after the fact:

- missing-field `required`
- defaults that depend on a field being absent from the request

## Registration requirement

Every request struct that should be validated by the plugin must be registered
once with Pedantigo:

```go
type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

var _ = validator.Register(
	validator.New[CreateUserRequest](validator.Options{TagName: "binding"}),
)
```

Why registration is required:

- JSON requests use `validator.UnmarshalInto`, which looks up the registered
  validator for the runtime target type.
- Query/form/header/URI requests use `validator.ValidateInto`, which performs
  the same runtime lookup after Gin has already filled the struct.

If a JSON target type is not registered, Pedantigo's internal panic is
recovered by the plugin and returned as a normal bind error. If a non-JSON
target type is not registered, `ValidateInto` returns `nil`, so Gin treats the
bind as successful and no Pedantigo validation runs.

## Tag names

The plugin defaults to Gin's conventional `binding:"..."` tag namespace.

```go
pedantigogin.NewBinder()
```

If your registered validators use a different tag name, set it once at plugin
setup:

```go
pedantigogin.NewBinder(pedantigogin.WithTagName("validate"))
```

This calls `validator.RequireSingleRegisteredTagName(...)` during setup, so all
plugin-visible registered validators in the process must agree on one tag name.
Mixing `binding`-registered and `validate`-registered request structs in the
same process will panic at setup or registration time.

## Error handling

Validation failures come back as normal bind errors. To detect Pedantigo's
structured validation error, use `AsValidationError`:

```go
var req CreateUserRequest
if err := c.ShouldBindJSON(&req); err != nil {
	if ve, ok := pedantigogin.AsValidationError(err); ok {
		return c.JSON(http.StatusBadRequest, ve.Errors)
	}
	return c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
```

## Unsupported Gin decoder flags

`NewBinder()` panics at setup time if either of these Gin globals was already
enabled:

- `binding.EnableDecoderUseNumber`
- `binding.EnableDecoderDisallowUnknownFields`

Those flags customize Gin's default JSON decoder path. This plugin replaces that
path with `validator.UnmarshalInto`, so the flags cannot be forwarded safely.
