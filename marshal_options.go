package pedantigo

// MarshalOptions configures Marshal behavior.
type MarshalOptions struct {
	// Context specifies which exclusion context to apply.
	// Fields tagged with pedantigo:"exclude:context" will be omitted.
	// Empty string means no context-based exclusion.
	Context string

	// OmitZero controls whether fields with omitzero tag and zero values are omitted.
	// Default: true (honor omitzero tags)
	OmitZero bool
}

// DefaultMarshalOptions returns sensible defaults.
func DefaultMarshalOptions() MarshalOptions {
	return MarshalOptions{
		Context:  "",
		OmitZero: true,
	}
}

// ForContext creates MarshalOptions for a specific exclusion context.
func ForContext(ctx string) MarshalOptions {
	return MarshalOptions{
		Context:  ctx,
		OmitZero: true,
	}
}
