package pedantigo

// ExtraFieldsMode controls how unknown JSON fields are handled during Unmarshal.
type ExtraFieldsMode int

const (
	// ExtraIgnore ignores unknown JSON fields (default behavior).
	ExtraIgnore ExtraFieldsMode = iota
	// ExtraForbid rejects JSON with unknown fields.
	ExtraForbid
	// ExtraAllow stores unknown fields (reserved for future use).
	ExtraAllow
)

// ValidatorOptions configures validator behavior.
type ValidatorOptions struct {
	// StrictMissingFields controls whether missing fields without defaults are errors
	// When true (default): missing fields without defaults cause validation errors
	// When false: missing fields are left as zero values (user handles with pointers)
	StrictMissingFields bool

	// ExtraFields controls how unknown JSON fields are handled during Unmarshal.
	// Default is ExtraIgnore (unknown fields are silently ignored).
	ExtraFields ExtraFieldsMode

	// TagName overrides the global struct tag name for this validator instance.
	// If empty, the global tag name (set via SetTagName or defaulting to "pedantigo") is used.
	// This allows different validators to use different struct tag names.
	//
	// Example:
	//   v := pedantigo.New[User](pedantigo.ValidatorOptions{TagName: "binding"})
	//   // This validator uses `binding:"required,email"` tags
	TagName string
}

// resolveTagName determines the effective tag name for a validator.
// Returns the instance TagName if set, otherwise falls back to the global tag name.
func resolveTagName(opts ValidatorOptions) string {
	if opts.TagName != "" {
		return opts.TagName
	}
	return GetTagName()
}

// DefaultValidatorOptions returns the default validator options.
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		StrictMissingFields: true,
		ExtraFields:         ExtraIgnore,
	}
}
