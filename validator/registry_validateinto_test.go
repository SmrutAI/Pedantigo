package validator

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type validateIntoValidRequest struct {
	Name string `validate:"required,min=1"`
}

// validateIntoValidRequest is Register()ed with no explicit Options, matching
// how every real caller of this package actually registers a type (e.g.
// smritea-cloud's `var _ = validator.Register(validator.New[T]())` package-level
// declarations): relying on the default "validate" tag resolved via
// GetTagName(). Passing an explicit TagName here would test a configuration
// no real caller uses.
//
// registerValidateIntoValidRequest is called from inside each test function
// that needs this type, not from a package-level var. Two things force this
// shape:
//
//  1. Register() panics if called twice for the same type (registry.go:
//     "Register[T]() may be called exactly once per type"). This type is
//     used by two tests below (TestValidateInto_RegisteredType_Valid and
//     TestValidateInto_RegisteredType_Invalid); each test must still be
//     runnable alone via `go test -run <name>`, so each needs its own call
//     guaranteeing the type is registered -- but only the first such call
//     across the whole test run may actually perform the registration.
//     sync.Once.Do gives exactly that: safe to call from every test that
//     needs the type, a no-op after the first successful call.
//  2. This file lives inside the validator package itself (an internal test
//     file, like every other *_test.go here), and Go runs all of a package's
//     var initializers before any of that package's own init() functions.
//     A package-level `var _ = Register(New[validateIntoValidRequest]())`
//     would therefore call New() -> resolveTagName() -> GetTagName() before
//     config.go's `func init() { globalTagName.Store(DefaultTagName) }` has
//     run, panicking on a nil interface type assertion. Calling Register()
//     from inside a test function body runs strictly after all package init
//     has completed, so it never hits that race.
var validateIntoValidRequestOnce sync.Once

func registerValidateIntoValidRequest() {
	validateIntoValidRequestOnce.Do(func() {
		Register(New[validateIntoValidRequest]())
	})
}

func TestValidateInto_RegisteredType_Valid(t *testing.T) {
	registerValidateIntoValidRequest()

	err := ValidateInto(&validateIntoValidRequest{Name: "x"})
	require.NoError(t, err)
}

func TestValidateInto_RegisteredType_Invalid(t *testing.T) {
	registerValidateIntoValidRequest()

	err := ValidateInto(&validateIntoValidRequest{})
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.NotEmpty(t, ve.Errors)
	require.Equal(t, "Name", ve.Errors[0].Field)
}

func TestValidateInto_UnregisteredType_ReturnsNilNotPanic(t *testing.T) {
	type unregisteredValidateIntoRequest struct {
		Name string `validate:"required"`
	}

	var (
		req unregisteredValidateIntoRequest
		err error
	)
	require.NotPanics(t, func() {
		err = ValidateInto(&req)
	})
	require.NoError(t, err)
}

func TestValidateInto_NonPointerValue_ReturnsError(t *testing.T) {
	err := ValidateInto(validateIntoValidRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a pointer")
}

func TestValidateInto_NilTypedPointer_ReturnsError(t *testing.T) {
	var req *validateIntoValidRequest
	err := ValidateInto(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a non-nil pointer")
}

func TestValidateInto_UntypedNil_ReturnsError(t *testing.T) {
	err := ValidateInto(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a pointer")
}
