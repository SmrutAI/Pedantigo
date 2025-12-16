package pedantigo

import (
	"reflect"
	"sync"

	"github.com/invopop/jsonschema"
)

var (
	// validatorCache stores cached validators per type.
	// Stores map[reflect.Type]any (*Validator[T]).
	validatorCache sync.Map
)

// getOrCreateValidator returns a cached validator for type T, creating one if needed.
// This is an internal helper used by the simple API functions.
// Thread-safe: uses LoadOrStore to ensure only one validator is created per type.
func getOrCreateValidator[T any]() *Validator[T] {
	var zero T
	typ := reflect.TypeOf(zero)

	// Fast path: check if already cached
	if cached, ok := validatorCache.Load(typ); ok {
		return cached.(*Validator[T])
	}

	// Slow path: create new validator
	validator := New[T]()

	// Atomically store and return the existing value if another goroutine beat us
	actual, _ := validatorCache.LoadOrStore(typ, validator)
	return actual.(*Validator[T])
}

// Unmarshal unmarshals JSON data into a validated struct of type T.
// It uses a cached validator for type T, creating one if necessary.
// This is equivalent to calling New[T]().Unmarshal(data) but with automatic caching.
//
// Example:
//
//	user, err := pedantigo.Unmarshal[User](jsonData)
//	if err != nil {
//	    // Handle validation errors
//	}
func Unmarshal[T any](data []byte) (*T, error) {
	return getOrCreateValidator[T]().Unmarshal(data)
}

// Validate validates an existing struct using cached validators.
// This is equivalent to calling New[T]().Validate(obj) but with automatic caching.
//
// Example:
//
//	user := &User{Email: "invalid"}
//	if err := pedantigo.Validate(user); err != nil {
//	    // Handle validation errors
//	}
func Validate[T any](obj *T) error {
	return getOrCreateValidator[T]().Validate(obj)
}

// NewModel creates a validated instance of T from various input types.
// Accepts: []byte (JSON), T (struct), *T (pointer), or map[string]any (kwargs).
// It uses a cached validator for type T, creating one if necessary.
//
// Example:
//
//	// From JSON bytes
//	user, err := pedantigo.NewModel[User](jsonData)
//
//	// From map (kwargs pattern)
//	user, err := pedantigo.NewModel[User](map[string]any{
//	    "email": "test@example.com",
//	    "age": 25,
//	})
//
//	// From existing struct (validates it)
//	existing := User{Email: "test@example.com"}
//	user, err := pedantigo.NewModel[User](existing)
func NewModel[T any](input any) (*T, error) {
	return getOrCreateValidator[T]().NewModel(input)
}

// Schema returns the JSON Schema for type T using a cached validator.
// The schema is cached within the validator for maximum performance.
//
// Example:
//
//	schema := pedantigo.Schema[User]()
//	// schema contains the full JSON Schema object
func Schema[T any]() *jsonschema.Schema {
	return getOrCreateValidator[T]().Schema()
}

// SchemaJSON returns the JSON Schema for type T as JSON bytes.
// The schema is cached within the validator for maximum performance.
//
// Example:
//
//	schemaBytes, err := pedantigo.SchemaJSON[User]()
//	if err != nil {
//	    // Handle error
//	}
func SchemaJSON[T any]() ([]byte, error) {
	return getOrCreateValidator[T]().SchemaJSON()
}

// SchemaOpenAPI returns an OpenAPI-compatible JSON Schema for type T.
// This version includes OpenAPI-specific enhancements like nullable support.
//
// Example:
//
//	schema := pedantigo.SchemaOpenAPI[User]()
//	// Use in OpenAPI specification
func SchemaOpenAPI[T any]() *jsonschema.Schema {
	return getOrCreateValidator[T]().SchemaOpenAPI()
}

// SchemaJSONOpenAPI returns an OpenAPI-compatible JSON Schema as JSON bytes.
// This version includes OpenAPI-specific enhancements like nullable support.
//
// Example:
//
//	schemaBytes, err := pedantigo.SchemaJSONOpenAPI[User]()
//	if err != nil {
//	    // Handle error
//	}
func SchemaJSONOpenAPI[T any]() ([]byte, error) {
	return getOrCreateValidator[T]().SchemaJSONOpenAPI()
}

// Marshal validates and marshals a struct to JSON using default options.
// It uses a cached validator for type T, creating one if necessary.
//
// Example:
//
//	user := &User{Email: "test@example.com", Age: 25}
//	jsonData, err := pedantigo.Marshal(user)
//	if err != nil {
//	    // Handle validation or marshal error
//	}
func Marshal[T any](obj *T) ([]byte, error) {
	return getOrCreateValidator[T]().Marshal(obj)
}

// MarshalWithOptions validates and marshals a struct to JSON with custom options.
// Options allow context-based field exclusion and omitzero behavior.
// It uses a cached validator for type T, creating one if necessary.
//
// Example:
//
//	user := &User{Email: "test@example.com", Password: "secret"}
//	opts := pedantigo.ForContext("api") // Excludes password if tagged with exclude:api
//	jsonData, err := pedantigo.MarshalWithOptions(user, opts)
func MarshalWithOptions[T any](obj *T, opts MarshalOptions) ([]byte, error) {
	return getOrCreateValidator[T]().MarshalWithOptions(obj, opts)
}

// Dict converts a struct into a map[string]interface{}.
// It uses a cached validator for type T, creating one if necessary.
//
// Example:
//
//	user := &User{Email: "test@example.com", Age: 25}
//	dict, err := pedantigo.Dict(user)
//	// dict["email"] == "test@example.com"
//	// dict["age"] == 25
func Dict[T any](obj *T) (map[string]interface{}, error) {
	return getOrCreateValidator[T]().Dict(obj)
}
