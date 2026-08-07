package tags

// ParsedTag represents a structured validation tag with dive support.
// It separates constraints into collection-level, key-level, and element-level.
//
// Example tags:
//   - validate:"min=3"                    -> CollectionConstraints only
//   - validate:"dive,email"               -> ElementConstraints only (dive present)
//   - validate:"min=3,dive,min=5"         -> Both collection and element
//   - validate:"dive,keys,min=2,endkeys,email" -> Map: key + value constraints
type ParsedTag struct {
	// CollectionConstraints are constraints that apply to the collection itself
	// (before any dive tag). For slices: min/max = element count.
	// For maps: min/max = entry count.
	CollectionConstraints map[string]string

	// DivePresent indicates if the "dive" keyword was found in the tag.
	// When true, ElementConstraints apply to each element.
	DivePresent bool

	// KeyConstraints are constraints that apply to map keys only.
	// Only valid after "dive,keys" and before "endkeys".
	KeyConstraints map[string]string

	// ElementConstraints are constraints that apply to each element (slice)
	// or each value (map) after the dive tag.
	ElementConstraints map[string]string
}
