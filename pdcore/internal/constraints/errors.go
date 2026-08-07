package constraints

import "fmt"

// ConstraintError represents a validation error with a machine-readable code.
type ConstraintError struct {
	Code    string // Machine-readable error code (e.g., "INVALID_EMAIL")
	Message string // Human-readable message
}

// Error implements the error interface.
func (e *ConstraintError) Error() string {
	return e.Message
}

// NewConstraintError creates a new ConstraintError.
func NewConstraintError(code, message string) *ConstraintError {
	return &ConstraintError{Code: code, Message: message}
}

// NewConstraintErrorf creates a new ConstraintError with formatted message.
func NewConstraintErrorf(code, format string, args ...any) *ConstraintError {
	return &ConstraintError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
