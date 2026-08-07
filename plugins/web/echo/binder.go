// Package echo provides a pedantigo-backed Binder for the Echo web framework.
//
// Usage:
//
//	import pedantigoecho "github.com/SmrutAI/pedantigo/v2/plugins/web/echo"
//
//	e := echo.New()
//	e.Binder = pedantigoecho.NewBinder()
package echo

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

// PedantigoBinder replaces Echo's DefaultBinder. For POST/PUT/PATCH requests,
// it reads the JSON body and runs validator.UnmarshalInto (which enforces required,
// defaults, and all validate constraints). For GET/DELETE/HEAD, it falls back
// to Echo's DefaultBinder (path params + query params only).
type PedantigoBinder struct {
	fallback echo.DefaultBinder
}

// NewBinder creates a PedantigoBinder.
func NewBinder() *PedantigoBinder {
	return &PedantigoBinder{}
}

// Bind implements echo.Binder. For POST/PUT/PATCH it reads the body and calls
// validator.UnmarshalInto. For GET/DELETE/HEAD it delegates to Echo's DefaultBinder.
func (b *PedantigoBinder) Bind(i interface{}, c echo.Context) error {
	if err := b.fallback.BindPathParams(c, i); err != nil {
		return err
	}

	method := c.Request().Method
	if method == http.MethodGet || method == http.MethodDelete ||
		method == http.MethodHead {
		return b.fallback.BindQueryParams(c, i)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
	}
	if len(body) == 0 {
		return nil
	}
	if err := validator.UnmarshalInto(body, i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
