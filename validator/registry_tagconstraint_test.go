package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequireSingleRegisteredTagName_SeedsWhenEmpty(t *testing.T) {
	resetRegisteredTagNameForTesting()

	require.NotPanics(t, func() {
		RequireSingleRegisteredTagName("binding")
	})

	type seededBindingRequest struct {
		Name string `binding:"required"`
	}

	require.NotPanics(t, func() {
		Register(New[seededBindingRequest](Options{TagName: "binding"}))
	})
}

func TestRequireSingleRegisteredTagName_PassesWhenMatching(t *testing.T) {
	resetRegisteredTagNameForTesting()

	type matchingValidateRequest struct {
		Name string `validate:"required"`
	}

	Register(New[matchingValidateRequest]())

	require.NotPanics(t, func() {
		RequireSingleRegisteredTagName("validate")
	})
}

func TestRequireSingleRegisteredTagName_PanicsWhenMismatched(t *testing.T) {
	resetRegisteredTagNameForTesting()

	type mismatchedValidateRequest struct {
		Name string `validate:"required"`
	}

	Register(New[mismatchedValidateRequest]())

	require.Panics(t, func() {
		RequireSingleRegisteredTagName("binding")
	})
}

func TestRequireSingleRegisteredTagName_CalledTwiceDifferentWant(t *testing.T) {
	resetRegisteredTagNameForTesting()

	RequireSingleRegisteredTagName("binding")

	require.Panics(t, func() {
		RequireSingleRegisteredTagName("validate")
	})
}

func TestRegister_PanicsOnTagNameMismatchAfterSeed(t *testing.T) {
	resetRegisteredTagNameForTesting()

	RequireSingleRegisteredTagName("binding")

	type mismatchedSeedRequest struct {
		Name string `validate:"required"`
	}

	require.Panics(t, func() {
		Register(New[mismatchedSeedRequest]())
	})
}

func TestRegister_SecondTypeMatchingTagPasses(t *testing.T) {
	resetRegisteredTagNameForTesting()

	type firstMatchingRegisteredRequest struct {
		Name string `validate:"required"`
	}
	type secondMatchingRegisteredRequest struct {
		Name string `validate:"required"`
	}

	require.NotPanics(t, func() {
		Register(New[firstMatchingRegisteredRequest]())
	})
	require.NotPanics(t, func() {
		Register(New[secondMatchingRegisteredRequest]())
	})
}

func TestRegister_SecondTypeMismatchedTagPanics(t *testing.T) {
	resetRegisteredTagNameForTesting()

	type firstRegisteredValidateRequest struct {
		Name string `validate:"required"`
	}
	type secondRegisteredBindingRequest struct {
		Name string `binding:"required"`
	}

	Register(New[firstRegisteredValidateRequest]())

	require.Panics(t, func() {
		Register(New[secondRegisteredBindingRequest](Options{TagName: "binding"}))
	})
}
