package pedantigo

import "fmt"

// FieldError represents a single field validation error
type FieldError struct {
	Field   string // Field path (e.g., "user.email")
	Message string // Human-readable error message
	Value   any    // The value that failed validation
}

// ValidationError represents one or more validation errors
// It implements the error interface for idiomatic Go error handling
// ValidationError represents an error condition
type ValidationError struct {
	Errors []FieldError
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("%s: %s", e.Errors[0].Field, e.Errors[0].Message)
	}
	return fmt.Sprintf("%s: %s (and %d more errors)",
		e.Errors[0].Field, e.Errors[0].Message, len(e.Errors)-1)
}
