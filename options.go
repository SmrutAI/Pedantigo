package pedantigo

// ValidatorOptions configures validator behavior
type ValidatorOptions struct {
	// StrictMissingFields controls whether missing fields without defaults are errors
	// When true (default): missing fields without defaults cause validation errors
	// When false: missing fields are left as zero values (user handles with pointers)
	StrictMissingFields bool
}

// DefaultValidatorOptions returns the default validator options
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		StrictMissingFields: true,
	}
}
