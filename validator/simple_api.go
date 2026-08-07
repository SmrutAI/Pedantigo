package validator

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/SmrutAI/pedantigo/v2/validator/internal/constraints"
	"github.com/SmrutAI/pedantigo/v2/validator/internal/tags"
)

// Unmarshal unmarshals JSON data into a validated struct of type T.
// It uses a cached validator for type T, creating one if necessary.
// This is equivalent to calling New[T]().Unmarshal(data) but with automatic caching.
//
// Example:
//
//	user, err := vl.Unmarshal[User](jsonData)
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
//	if err := vl.Validate(user); err != nil {
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
//	user, err := vl.NewModel[User](jsonData)
//
//	// From map (kwargs pattern)
//	user, err := vl.NewModel[User](map[string]any{
//	    "email": "test@example.com",
//	    "age": 25,
//	})
//
//	// From existing struct (validates it)
//	existing := User{Email: "test@example.com"}
//	user, err := vl.NewModel[User](existing)
func NewModel[T any](input any) (*T, error) {
	return getOrCreateValidator[T]().NewModel(input)
}

// Schema returns the JSON Schema for type T using a cached validator.
// The schema is cached within the validator for maximum performance.
//
// Example:
//
//	schema := vl.Schema[User]()
//	// schema contains the full JSON Schema object
func Schema[T any]() *jsonschema.Schema {
	return getOrCreateValidator[T]().Schema()
}

// SchemaJSON returns the JSON Schema for type T as JSON bytes.
// The schema is cached within the validator for maximum performance.
//
// Example:
//
//	schemaBytes, err := validator.SchemaJSON[User]()
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
//	schema := validator.SchemaOpenAPI[User]()
//	// Use in OpenAPI specification
func SchemaOpenAPI[T any]() *jsonschema.Schema {
	return getOrCreateValidator[T]().SchemaOpenAPI()
}

// SchemaJSONOpenAPI returns an OpenAPI-compatible JSON Schema as JSON bytes.
// This version includes OpenAPI-specific enhancements like nullable support.
//
// Example:
//
//	schemaBytes, err := validator.SchemaJSONOpenAPI[User]()
//	if err != nil {
//	    // Handle error
//	}
func SchemaJSONOpenAPI[T any]() ([]byte, error) {
	return getOrCreateValidator[T]().SchemaJSONOpenAPI()
}

// SchemaLLM returns a JSON Schema optimized for LLM APIs (no $schema field).
// The schema is cached within the validator for maximum performance.
// Use this for: OpenAI function calling, Anthropic tool use, Claude structured outputs.
//
// Example:
//
//	schema := validator.SchemaLLM[User]()
//	// schema contains the JSON Schema object without $schema field
func SchemaLLM[T any]() *jsonschema.Schema {
	return getOrCreateValidator[T]().SchemaLLM()
}

// SchemaJSONLLM returns a JSON Schema as JSON bytes optimized for LLM APIs.
// The $schema field is omitted because some LLMs (like Groq) echo it back in responses.
// Use this for: OpenAI function calling, Anthropic tool use, Claude structured outputs.
//
// Example:
//
//	schemaBytes, err := validator.SchemaJSONLLM[User]()
//	if err != nil {
//	    // Handle error
//	}
//	// JSON output will NOT contain "$schema" field
func SchemaJSONLLM[T any]() ([]byte, error) {
	return getOrCreateValidator[T]().SchemaJSONLLM()
}

// Marshal validates and marshals a struct to JSON using default options.
// It uses a cached validator for type T, creating one if necessary.
//
// Example:
//
//	user := &User{Email: "test@example.com", Age: 25}
//	jsonData, err := vl.Marshal(user)
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
//	opts := validator.ForContext("api") // Excludes password if tagged with exclude:api
//	jsonData, err := validator.MarshalWithOptions(user, opts)
func MarshalWithOptions[T any](obj *T, opts MarshalOptions) ([]byte, error) {
	return getOrCreateValidator[T]().MarshalWithOptions(obj, opts)
}

// Dict converts a struct into a map[string]interface{}.
// It uses a cached validator for type T, creating one if necessary.
//
// Example:
//
//	user := &User{Email: "test@example.com", Age: 25}
//	dict, err := validator.Dict(user)
//	// dict["email"] == "test@example.com"
//	// dict["age"] == 25
func Dict[T any](obj *T) (map[string]interface{}, error) {
	return getOrCreateValidator[T]().Dict(obj)
}

// Var validates a single value against the provided constraints.
// This allows validating values without defining a struct.
//
// Example:
//
//	err := validator.Var("test@example.com", "required,email")
//	err := validator.Var(25, "min=18,max=120")
func Var(value any, tag string) error {
	if tag == "" {
		return nil // No constraints to validate
	}

	// Parse the tag using the global tag name
	tagName := GetTagName()
	// Escape backslashes for Go's struct tag parser
	// Go's tag.Get() treats \x as escape sequences, so we need to double them
	escapedTag := strings.ReplaceAll(tag, `\`, `\\`)
	// Create a fake struct tag for parsing
	structTag := reflect.StructTag(tagName + `:` + `"` + escapedTag + `"`)
	constraintsMap := tags.ParseTagWithName(structTag, tagName)

	if len(constraintsMap) == 0 {
		return nil
	}

	// Check if "required" is in the constraints
	_, isRequired := constraintsMap["required"]

	// Check required first
	if isRequired {
		if value == nil {
			return &ValidationError{
				Errors: []FieldError{{
					Field:   fieldNameValue,
					Code:    codeRequired,
					Message: ErrMsgValueRequired,
				}},
			}
		}
		// Check if value is zero
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Pointer && rv.IsNil() {
			return &ValidationError{
				Errors: []FieldError{{
					Field:   fieldNameValue,
					Code:    codeRequired,
					Message: ErrMsgValueRequired,
				}},
			}
		}
		if isZeroValue(rv) {
			return &ValidationError{
				Errors: []FieldError{{
					Field:   fieldNameValue,
					Code:    codeRequired,
					Message: ErrMsgValueRequired,
				}},
			}
		}
	}

	// Skip validation for nil values (optional fields)
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return nil
	}

	// Get the type for building constraints
	var typ reflect.Type
	if rv.Kind() == reflect.Pointer {
		typ = rv.Elem().Type()
	} else {
		typ = rv.Type()
	}

	// Build constraints
	constrs := constraints.BuildConstraints(constraintsMap, typ)

	// Run validations
	var errs []FieldError
	for _, c := range constrs {
		err := c.Validate(value)
		if err == nil {
			continue
		}
		code := codeValidationFailed
		message := err.Error()

		// Extract code from ConstraintError if available
		var constraintErr *constraints.ConstraintError
		if errors.As(err, &constraintErr) {
			code = constraintErr.Code
			message = constraintErr.Message
		}

		errs = append(errs, FieldError{
			Field:   fieldNameValue,
			Code:    code,
			Message: message,
		})
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// ValidatePartial validates only the specified fields using a cached validator.
// Field names should match JSON field names (from json tags).
// Fields not in the list are skipped entirely.
//
// Example:
//
//	user := &User{Email: "invalid", Age: 15}
//	// Only validate email field, skip age validation
//	err := validator.ValidatePartial(user, "email")
func ValidatePartial[T any](obj *T, fields ...string) error {
	return getOrCreateValidator[T]().StructPartial(obj, fields...)
}

// ValidateExcept validates all fields except specified ones using cached validator.
// Field names should match JSON field names (from json tags).
// Excluded fields are skipped entirely.
//
// Example:
//
//	user := &User{Email: "test@example.com", Age: 15}
//	// Validate all fields except age
//	err := validator.ValidateExcept(user, "age")
func ValidateExcept[T any](obj *T, excludeFields ...string) error {
	return getOrCreateValidator[T]().StructExcept(obj, excludeFields...)
}

// ValidateCtx validates with context support for context-aware validators.
// It uses a cached validator for type T, creating one if necessary.
// Context-aware validators registered with RegisterValidationCtx will receive
// the provided context.
//
// Example:
//
//	ctx := context.WithValue(context.Background(), "db", dbConn)
//	user := &User{Username: "john"}
//	if err := validator.ValidateCtx(ctx, user); err != nil {
//	    // Handle validation errors
//	}
func ValidateCtx[T any](ctx context.Context, obj *T) error {
	return getOrCreateValidator[T]().ValidateCtx(ctx, obj)
}

// UnmarshalCtx unmarshals and validates with context.
// It uses a cached validator for type T, creating one if necessary.
// This allows context-aware validators to access the context during unmarshal.
//
// Example:
//
//	ctx := context.WithValue(context.Background(), "db", dbConn)
//	user, err := validator.UnmarshalCtx[User](ctx, jsonData)
//	if err != nil {
//	    // Handle validation errors
//	}
func UnmarshalCtx[T any](ctx context.Context, data []byte) (*T, error) {
	return getOrCreateValidator[T]().UnmarshalCtx(ctx, data)
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
// For the "required" constraint, only empty strings are considered zero.
// Numeric zeros (0, 0.0) and boolean false are valid non-zero values.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil() || v.Len() == 0
	default:
		// For numeric types and bools, any value (including 0 and false) is valid
		return false
	}
}
