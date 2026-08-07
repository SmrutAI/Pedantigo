package validator

import "fmt"

// Error code constants for validation errors.
const (
	// codeValidationFailed is the default error code when no specific code is available.
	codeValidationFailed = "VALIDATION_FAILED"

	// codeRequired is the error code for a missing required value.
	codeRequired = "REQUIRED"

	// fieldNameValue is the FieldError.Field value used by Var, which validates a
	// standalone value rather than a named struct field.
	fieldNameValue = "value"

	// fieldNameRoot is the FieldError.Field value for errors that apply to the
	// whole struct being validated rather than one of its fields.
	fieldNameRoot = "root"
)

// Error message constants for validation errors.
const (
	// ErrMsgFieldRequired is returned when a required field or value is missing.
	ErrMsgFieldRequired = "is required"

	// ErrMsgNilPointer is returned when Validate/StructPartial/StructExcept receives a nil pointer.
	ErrMsgNilPointer = "cannot validate nil pointer"

	// ErrMsgValueRequired is returned by Var when a required standalone value is missing.
	ErrMsgValueRequired = "value is required"

	// ErrMsgUnknownField is returned when ExtraForbid encounters unknown JSON fields.
	ErrMsgUnknownField = "unknown field in JSON"

	// ErrMsgEqMismatch is returned when a value doesn't match the expected value.
	ErrMsgEqMismatch = "must be equal to %s"

	// ErrMsgNeMismatch is returned when a value matches a forbidden value.
	ErrMsgNeMismatch = "must not be equal to %s"

	// ErrMsgMissingDiscriminator is returned when discriminator field is missing from JSON.
	ErrMsgMissingDiscriminator = "discriminator field %q is missing"

	// ErrMsgUnknownDiscriminator is returned when discriminator value doesn't match any variant.
	ErrMsgUnknownDiscriminator = "unknown discriminator value %q for field %q"

	// ErrMsgExtraFieldRequired is returned when ExtraAllow mode is enabled but no extra_fields field exists.
	ErrMsgExtraFieldRequired = "ExtraAllow mode requires a field with `pedantigo:\"extra_fields\"` tag of type map[string]any"
)

// FieldError represents a single field validation error.
type FieldError struct {
	Field   string // Field path (e.g., "user.email")
	Code    string // Machine-readable error code (e.g., "INVALID_EMAIL")
	Message string // Human-readable error message
	Value   any    // The value that failed validation
}

// ValidationError represents one or more validation errors
// It implements the error interface for idiomatic Go error handling
// ValidationError represents an error condition.
type ValidationError struct {
	Errors []FieldError
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors found"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("%s: %s", e.Errors[0].Field, e.Errors[0].Message)
	}
	return fmt.Sprintf("%s: %s (and %d more errors)",
		e.Errors[0].Field, e.Errors[0].Message, len(e.Errors)-1)
}
