package pedantigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidate_OmitEmpty_WithCrossFieldConstraints verifies the omitempty + cross-field
// constraint interaction: for zero-value omitempty fields, regular constraints (oneof, max,
// min) are skipped, but cross-field constraints (required_with, required_if, eqfield) always
// run regardless of the field's zero-value status.
func TestValidate_OmitEmpty_WithCrossFieldConstraints(t *testing.T) {
	// -------------------------------------------------------------------------
	// Group 1: required_with + omitempty
	// Both ActorID and ActorType are omitempty fields referencing each other via
	// required_with. This creates a symmetric: "if one is present, the other must
	// be too" — without requiring either when both are absent.
	// -------------------------------------------------------------------------

	type omitEmptyScopeStruct struct {
		ActorID   string `pedantigo:"omitempty,max=64,required_with=ActorType"`
		ActorType string `pedantigo:"omitempty,oneof=user agent system,required_with=ActorID"`
	}

	t.Run("required_with/both_zero_no_error", func(t *testing.T) {
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "", ActorType: ""})
		assert.NoError(t, err)
	})

	t.Run("required_with/actor_id_zero_actor_type_nonzero_errors", func(t *testing.T) {
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "", ActorType: "user"})
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		// ActorID is omitempty+zero, but required_with=ActorType fires because ActorType
		// is non-zero. The error is reported on ActorID.
		found := false
		for i := range ve.Errors {
			if ve.Errors[i].Field == "ActorID" {
				assert.Contains(t, ve.Errors[i].Message, "is required when ActorType is present")
				found = true
			}
		}
		assert.True(t, found, "expected error on field ActorID")
	})

	t.Run("required_with/actor_id_nonzero_actor_type_zero_errors", func(t *testing.T) {
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "user123", ActorType: ""})
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		// ActorType is omitempty+zero, but required_with=ActorID fires because ActorID
		// is non-zero. The error is reported on ActorType.
		found := false
		for i := range ve.Errors {
			if ve.Errors[i].Field == "ActorType" {
				assert.Contains(t, ve.Errors[i].Message, "is required when ActorID is present")
				found = true
			}
		}
		assert.True(t, found, "expected error on field ActorType")
	})

	t.Run("required_with/both_nonzero_valid", func(t *testing.T) {
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "user123", ActorType: "user"})
		assert.NoError(t, err)
	})

	t.Run("required_with/both_nonzero_invalid_oneof_errors", func(t *testing.T) {
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "user123", ActorType: "invalid"})
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		// ActorType is non-zero so regular constraints run: oneof fires because "invalid"
		// is not in the allowed set. required_with on ActorType does NOT fire because
		// ActorID is non-zero (so ActorType passes that check). The error must be from
		// oneof, not from required_with.
		found := false
		for i := range ve.Errors {
			if ve.Errors[i].Field == "ActorType" {
				// Must be an oneof error, not a required_with error.
				assert.NotContains(t, ve.Errors[i].Message, "is required when")
				assert.Contains(t, ve.Errors[i].Message, "must be one of")
				found = true
			}
		}
		assert.True(t, found, "expected oneof error on field ActorType")
	})

	// -------------------------------------------------------------------------
	// Group 2: Regular constraints still skip for omitempty zero fields
	// These three cases document intent explicitly: zero-value omitempty fields
	// do not trigger oneof, max, or other regular constraints.
	// -------------------------------------------------------------------------

	t.Run("omitempty_skips_oneof_when_zero", func(t *testing.T) {
		// Both zero: oneof on ActorType does NOT fire (omitempty + zero = skip regular).
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "", ActorType: ""})
		assert.NoError(t, err)
	})

	t.Run("omitempty_skips_max_when_zero", func(t *testing.T) {
		// Both zero: max=64 on ActorID does NOT fire (omitempty + zero = skip regular).
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "", ActorType: ""})
		assert.NoError(t, err)
	})

	t.Run("omitempty_runs_oneof_when_nonzero", func(t *testing.T) {
		// ActorType is non-zero "invalid": oneof fires because the field has a value.
		v := New[omitEmptyScopeStruct]()
		err := v.Validate(&omitEmptyScopeStruct{ActorID: "user123", ActorType: "invalid"})
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		found := false
		for i := range ve.Errors {
			if ve.Errors[i].Field == "ActorType" && ve.Errors[i].Message != "" {
				found = true
			}
		}
		assert.True(t, found, "expected validation error on ActorType for invalid oneof value")
	})

	// -------------------------------------------------------------------------
	// Group 3: required_if + omitempty
	// Query is omitempty and conditionally required when Mode equals "manual".
	// The condition check (required_if) always runs even when Query is zero.
	// -------------------------------------------------------------------------

	type omitEmptyConditionalStruct struct {
		Mode  string `pedantigo:"oneof=auto manual"`
		Query string `pedantigo:"omitempty,required_if=Mode manual"`
	}

	t.Run("required_if/condition_not_met_no_error", func(t *testing.T) {
		// Mode is "auto" — condition not met, so Query can remain empty.
		v := New[omitEmptyConditionalStruct]()
		err := v.Validate(&omitEmptyConditionalStruct{Mode: "auto", Query: ""})
		assert.NoError(t, err)
	})

	t.Run("required_if/condition_met_zero_field_errors", func(t *testing.T) {
		// Mode is "manual" — condition met. Query is zero (omitempty), but
		// required_if always runs and fires an error.
		v := New[omitEmptyConditionalStruct]()
		err := v.Validate(&omitEmptyConditionalStruct{Mode: "manual", Query: ""})
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		found := false
		for i := range ve.Errors {
			if ve.Errors[i].Field == "Query" {
				assert.Contains(t, ve.Errors[i].Message, "is required when Mode equals 'manual'")
				found = true
			}
		}
		assert.True(t, found, "expected required_if error on field Query")
	})

	t.Run("required_if/condition_met_nonzero_field_no_error", func(t *testing.T) {
		// Mode is "manual" and Query is provided — no error.
		v := New[omitEmptyConditionalStruct]()
		err := v.Validate(&omitEmptyConditionalStruct{Mode: "manual", Query: "some query"})
		assert.NoError(t, err)
	})

	// -------------------------------------------------------------------------
	// Group 4: Nested struct pointer + omitempty
	// Inner is a pointer field with omitempty. A nil pointer is the zero value:
	// omitempty skips recursion entirely (no panic, no min=1 error from Inner.Value).
	// A non-nil pointer is non-zero: recursion runs and Inner.Value is validated.
	// -------------------------------------------------------------------------

	type omitEmptyInnerStruct struct {
		Value string `pedantigo:"min=1"`
	}

	type omitEmptyOuterStruct struct {
		Inner *omitEmptyInnerStruct `pedantigo:"omitempty"`
		Name  string                `pedantigo:"omitempty,max=10"`
	}

	t.Run("nested_struct/nil_pointer_omitempty_skips_recursion", func(t *testing.T) {
		// Inner is nil (zero value for pointer) and Name is empty (zero value for string).
		// omitempty on Inner skips all recursion — no panic and no error from Inner.Value.
		v := New[omitEmptyOuterStruct]()
		err := v.Validate(&omitEmptyOuterStruct{Inner: nil, Name: ""})
		assert.NoError(t, err)
	})

	t.Run("nested_struct/nonnil_pointer_recursion_runs_and_errors", func(t *testing.T) {
		// Inner is non-nil (non-zero) so recursion runs. Inner.Value is "" which fails
		// min=1. The error path should be "Inner.Value".
		v := New[omitEmptyOuterStruct]()
		err := v.Validate(&omitEmptyOuterStruct{
			Inner: &omitEmptyInnerStruct{Value: ""},
			Name:  "",
		})
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		found := false
		for i := range ve.Errors {
			if ve.Errors[i].Field == "Inner.Value" {
				found = true
			}
		}
		assert.True(t, found, "expected validation error on field Inner.Value")
	})

	t.Run("nested_struct/nonnil_pointer_valid", func(t *testing.T) {
		// Inner is non-nil and Inner.Value is non-empty — recursion runs and passes.
		v := New[omitEmptyOuterStruct]()
		err := v.Validate(&omitEmptyOuterStruct{
			Inner: &omitEmptyInnerStruct{Value: "ok"},
			Name:  "",
		})
		assert.NoError(t, err)
	})
}
