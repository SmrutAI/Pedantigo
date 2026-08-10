package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type validateIntoValidRequest struct {
	Name string `validate:"required,min=1"`
}

var _ = Register(New[validateIntoValidRequest](Options{
	StrictMissingFields: true,
	ExtraFields:         ExtraIgnore,
	TagName:             DefaultTagName,
}))

func TestValidateInto_RegisteredType_Valid(t *testing.T) {
	err := ValidateInto(&validateIntoValidRequest{Name: "x"})
	require.NoError(t, err)
}

func TestValidateInto_RegisteredType_Invalid(t *testing.T) {
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
