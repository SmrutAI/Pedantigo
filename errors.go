package pedantigo

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string // Field path (e.g., "user.email")
	Message string // Human-readable error message
	Value   any    // The value that failed validation
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	return ve[0].Error() + " (and " + string(rune(len(ve)-1)) + " more errors)"
}

// Messages returns all error messages
func (ve ValidationErrors) Messages() []string {
	msgs := make([]string, len(ve))
	for i, err := range ve {
		msgs[i] = err.Error()
	}
	return msgs
}

// NewFieldError creates a new ValidationError for a specific field
func NewFieldError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
