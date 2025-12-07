package pedantigo

// constraint represents a validation constraint
type constraint interface {
	Validate(value any) error
}

// Built-in constraint types (to be implemented)
type (
	requiredConstraint struct{}
	minConstraint      struct{ min int }
	maxConstraint      struct{ max int }
	emailConstraint    struct{}
)
