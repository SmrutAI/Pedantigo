package constraints

// ContextConstraint represents a context-aware validator name and its parameter.
type ContextConstraint struct {
	Name  string // validator name (looked up at runtime)
	Param string // parameter value
}

// CachedField holds pre-built validation data for a single struct field.
// Built once at validator creation time, used on every Validate() call.
type CachedField struct {
	Name       string // struct field name
	JSONName   string // JSON field name (from json tag or field name)
	FieldIndex int    // index in parent struct for O(1) access

	// Pre-built constraints (from tags before dive)
	Constraints           []Constraint
	CrossFieldConstraints []CrossFieldConstraint // eqfield, gtfield, etc.
	ContextConstraints    []ContextConstraint    // context-aware validators (looked up at runtime)

	// For collections with dive
	HasDive            bool
	ElementConstraints []Constraint // constraints after dive
	KeyConstraints     []Constraint // for map keys (between keys/endkeys)

	// Field type info
	IsCollection bool // slice or map
	IsMap        bool // specifically a map
	IsRequired   bool // has required tag (for nested struct validation)

	// For nested structs (recursive cache)
	NestedCache *FieldCache
}

// FieldCache holds cached validation data for all fields in a struct.
type FieldCache struct {
	Fields []CachedField // indexed by struct field order
}

// NewFieldCache creates a new instance of FieldCache.
func NewFieldCache() *FieldCache {
	return &FieldCache{
		Fields: make([]CachedField, 0),
	}
}
