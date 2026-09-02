package validator

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- UnmarshalInto on a Register()ed type that also owns a custom
// UnmarshalJSON: does UnmarshalInto call that method, or bypass it? ---

// uibBlock and uibLeaf are each Register()ed with no explicit Options, matching
// how every real caller of this package actually registers a type (e.g.
// smritea-cloud's `var _ = validator.Register(validator.New[T]())` package-level
// declarations): relying on the default "validate" tag resolved via
// GetTagName(). Passing an explicit TagName here would test a configuration
// no real caller uses.
//
// registerUibBlock/registerUibLeaf are called from inside each test function
// that needs the corresponding type, not from a package-level var. Two
// things force this shape:
//
//  1. Register() panics if called twice for the same type (registry.go:
//     "Register[T]() may be called exactly once per type"). uibBlock is used
//     by two tests below, and uibLeaf by two more; each test must still be
//     runnable alone via `go test -run <name>`, so each needs its own call
//     guaranteeing its type is registered -- but only the first such call
//     across the whole test run may actually perform the registration.
//     sync.Once.Do gives exactly that: safe to call from every test that
//     needs the type, a no-op after the first successful call.
//  2. This file lives inside the validator package itself (an internal test
//     file, like every other *_test.go here), and Go runs all of a package's
//     var initializers before any of that package's own init() functions.
//     A package-level `var _ = Register(New[uibBlock]())` would therefore
//     call New() -> resolveTagName() -> GetTagName() before config.go's
//     `func init() { globalTagName.Store(DefaultTagName) }` has run,
//     panicking on a nil interface type assertion. Calling Register() from
//     inside a test function body runs strictly after all package init has
//     completed, so it never hits that race.
var (
	uibBlockOnce sync.Once
	uibLeafOnce  sync.Once
)

func registerUibBlock() {
	uibBlockOnce.Do(func() {
		Register(New[uibBlock]())
	})
}

func registerUibLeaf() {
	uibLeafOnce.Do(func() {
		Register(New[uibLeaf]())
	})
}

type uibBlock struct {
	Type        string `json:"type" validate:"required"`
	CustomFired bool   `json:"-"`
	Extra       map[string]json.RawMessage
}

func (b *uibBlock) UnmarshalJSON(data []byte) error {
	b.CustomFired = true
	type shadow struct {
		Type string `json:"type"`
	}
	var s shadow
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.Type = s.Type
	raw := map[string]json.RawMessage{}
	_ = json.Unmarshal(data, &raw)
	b.Extra = map[string]json.RawMessage{}
	for k, v := range raw {
		if k != "type" {
			b.Extra[k] = v
		}
	}
	return nil
}

func TestUnmarshalInto_BypassesOwnTypesCustomUnmarshalJSON(t *testing.T) {
	registerUibBlock()

	var b uibBlock
	err := UnmarshalInto([]byte(`{"type":"text","foo":1}`), &b)
	require.NoError(t, err)

	require.False(t, b.CustomFired,
		"UnmarshalInto dispatched through the walker's own field-by-field decode "+
			"for uibBlock, bypassing uibBlock's own UnmarshalJSON entirely -- "+
			"WalkerDecoder/json.Unmarshaler dispatch (setter.go) only fires for "+
			"FIELDS of an already-walked struct, never for the top-level T passed "+
			"to Unmarshal/UnmarshalInto itself (verified: zero references to "+
			"WalkerDecoder or Unmarshaler in validator.go).")

	require.Equal(t, "text", b.Type, "the walker's native field decode still "+
		"populates uibBlock's own declared fields correctly")
	require.Empty(t, b.Extra, "the unknown key \"foo\" is silently dropped: "+
		"Extra is only ever populated by the bypassed custom UnmarshalJSON, "+
		"and the walker has no native knowledge of an Extra-flatten field")
}

func TestUnmarshalInto_StillEnforcesRequiredOnNativeFields(t *testing.T) {
	registerUibBlock()

	// Even though the custom UnmarshalJSON method above is bypassed, the
	// walker's OWN native required-check on uibBlock.Type (a directly
	// declared field) still fires correctly.
	var b uibBlock
	err := UnmarshalInto([]byte(`{"foo":1}`), &b)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Equal(t, "type", ve.Errors[0].Field)
}

// --- Route 2: manual []json.RawMessage + per-element UnmarshalInto, against
// a type registered WITHOUT any custom UnmarshalJSON, used to decode an
// untagged-union-shaped field (string | array) by hand. ---

type uibLeaf struct {
	Kind string `json:"kind" validate:"required"`
	Text string `json:"text"`
}

func decodeUibLeafArray(data []byte) ([]uibLeaf, error) {
	registerUibLeaf()

	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, err
	}
	leaves := make([]uibLeaf, 0, len(raws))
	var errs []FieldError
	for i, raw := range raws {
		var leaf uibLeaf
		if err := UnmarshalInto(raw, &leaf); err != nil {
			var ve *ValidationError
			if errors.As(err, &ve) {
				for _, fe := range ve.Errors {
					errs = append(errs, FieldError{
						Field:   fieldPathForElement(i, fe.Field),
						Message: fe.Message,
						Code:    fe.Code,
					})
				}
				continue
			}
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	if len(errs) > 0 {
		return leaves, &ValidationError{Errors: errs}
	}
	return leaves, nil
}

func fieldPathForElement(i int, field string) string {
	return "[" + itoa(i) + "]." + field
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	digits := ""
	for i > 0 {
		digits = string(rune('0'+i%10)) + digits
		i /= 10
	}
	return digits
}

func TestUnmarshalInto_ManualLoopEnforcesRequiredPerElement(t *testing.T) {
	leaves, err := decodeUibLeafArray([]byte(`[{"nope":true},{"kind":"text","text":"hi"}]`))
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Equal(t, "[0].kind", ve.Errors[0].Field)

	// The valid sibling element still decoded correctly despite the other
	// element's error.
	require.Len(t, leaves, 1)
	require.Equal(t, "text", leaves[0].Kind)
	require.Equal(t, "hi", leaves[0].Text)
}

func TestUnmarshalInto_ManualLoopAllValid(t *testing.T) {
	leaves, err := decodeUibLeafArray([]byte(`[{"kind":"text","text":"hi"},{"kind":"image","text":"ignored"}]`))
	require.NoError(t, err)
	require.Len(t, leaves, 2)
	require.Equal(t, "text", leaves[0].Kind)
	require.Equal(t, "image", leaves[1].Kind)
}
