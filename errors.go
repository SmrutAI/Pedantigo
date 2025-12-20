package pedantigo

import "fmt"

// Error message constants for validation errors.
const (
	// ErrMsgUnknownField is returned when ExtraForbid encounters unknown JSON fields.
	ErrMsgUnknownField = "unknown field in JSON"

	// ErrMsgConstMismatch is returned when a value doesn't match the expected constant.
	ErrMsgConstMismatch = "must be equal to %s"

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
