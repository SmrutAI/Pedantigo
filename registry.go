package pedantigo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/SmrutAI/pedantigo/internal/constraints"
	"github.com/SmrutAI/pedantigo/internal/tags"
)

// ValidationFunc is the signature for custom field-level validation functions.
// It receives the field value and param string, returns an error if validation fails.
type ValidationFunc func(value any, param string) error

// ValidationFuncCtx is the signature for context-aware custom validators.
type ValidationFuncCtx func(ctx context.Context, value any, param string) error

// TagNameFunc is the signature for custom field name resolution.
type TagNameFunc func(field reflect.StructField) string

func init() {
	// Wire up custom validator lookup to constraints package
	constraints.SetCustomValidatorLookup(func(name string) (constraints.CustomValidationFunc, bool) {
		if fn, ok := GetCustomValidator(name); ok {
			// Convert pedantigo.ValidationFunc to constraints.CustomValidationFunc
			// Both have the same signature: func(value any, param string) error
			return constraints.CustomValidationFunc(fn), true
		}
		return nil, false
	})

	// Wire up context validator lookup to constraints package
	constraints.SetCtxValidatorLookup(func(name string) bool {
		_, ok := ctxValidators.Load(name)
		return ok
	})

	// Wire up alias lookup to tags package
	tags.SetAliasLookup(GetAlias)
}

// StructLevelFunc is the signature for struct-level validation functions.
// It receives the entire struct and returns an error if validation fails.
type StructLevelFunc[T any] func(obj *T) error

var (
	// customValidators stores registered custom field validators.
	// Stores map[string]ValidationFunc.
	customValidators sync.Map

	// ctxValidators stores context-aware custom validators.
	ctxValidators sync.Map

	// structValidators stores registered struct-level validators.
	// Stores map[reflect.Type]any.
	structValidators sync.Map

	// aliases stores registered tag aliases.
	// Stores map[string]string where key is alias name, value is expansion.
	aliases sync.Map

	// tagNameFunc stores the custom tag name resolution function.
	tagNameFunc atomic.Pointer[TagNameFunc]
)

// Built-in aliases for validator compatibility.
func init() {
	// iscolor is an alias for all color formats (validator compatibility)
	aliases.Store("iscolor", "hexcolor|rgb|rgba|hsl|hsla")
}

// RegisterValidation registers a custom field-level validator with the given name.
// The validator function will be called during validation for fields tagged with this name.
// Returns an error if the name is empty, the function is nil, or if the name conflicts
// with a built-in validator.
func RegisterValidation(name string, fn ValidationFunc) error {
	if name == "" {
		return errors.New("validator name cannot be empty")
	}
	if fn == nil {
		return errors.New("validator function cannot be nil")
	}
	if isBuiltInValidator(name) {
		return fmt.Errorf("cannot override built-in validator: %s", name)
	}

	customValidators.Store(name, fn)
	clearValidatorCache()
	return nil
}

// RegisterStructValidation registers a struct-level validator for type T.
// The validator function will be called after field-level validation succeeds.
// Returns an error if the function is nil or if a validator is already registered for type T.
func RegisterStructValidation[T any](fn StructLevelFunc[T]) error {
	if fn == nil {
		return errors.New("validator function cannot be nil")
	}

	var zero T
	t := reflect.TypeOf(zero)
	structValidators.Store(t, fn)
	validatorCache.Delete(t)
	return nil
}

// GetCustomValidator retrieves a registered custom validator by name.
// Returns the validator function and true if found, nil and false otherwise.
func GetCustomValidator(name string) (ValidationFunc, bool) {
	if v, ok := customValidators.Load(name); ok {
		return v.(ValidationFunc), true
	}
	return nil, false
}

// RegisterValidationCtx registers a context-aware custom validator.
// The validator will receive the context passed to ValidateCtx.
//
// Example:
//
//	pedantigo.RegisterValidationCtx("db_unique", func(ctx context.Context, value any, param string) error {
//	    db := ctx.Value("db").(*sql.DB)
//	    // Check uniqueness in database
//	    return nil
//	})
func RegisterValidationCtx(name string, fn ValidationFuncCtx) error {
	if name == "" {
		return errors.New("validator name cannot be empty")
	}
	if fn == nil {
		return errors.New("validator function cannot be nil")
	}

	ctxValidators.Store(name, fn)
	clearValidatorCache()
	return nil
}

// GetContextValidator returns a registered context-aware validator by name.
// Returns (validator, true) if found, (nil, false) if not registered.
func GetContextValidator(name string) (ValidationFuncCtx, bool) {
	if v, ok := ctxValidators.Load(name); ok {
		return v.(ValidationFuncCtx), true
	}
	return nil, false
}

// RegisterTagNameFunc sets a custom function for resolving field names.
// This affects how field names appear in validation error messages.
//
// Example:
//
//	pedantigo.RegisterTagNameFunc(func(field reflect.StructField) string {
//	    if name := field.Tag.Get("form"); name != "" {
//	        return name
//	    }
//	    return field.Name
//	})
func RegisterTagNameFunc(fn TagNameFunc) {
	if fn == nil {
		tagNameFunc.Store(nil)
		clearValidatorCache()
		return
	}
	tagNameFunc.Store(&fn)
	clearValidatorCache()
}

// RegisterAlias registers a tag alias that expands to other tags.
// This allows creating shorthand names for common tag combinations.
//
// Example:
//
//	pedantigo.RegisterAlias("iscolor", "hexcolor|rgb|rgba|hsl|hsla")
//	// Now `iscolor` expands to an OR constraint for all color formats
//
//	pedantigo.RegisterAlias("username", "required,alphanum,min=3,max=20")
//	// Now `username` expands to multiple constraints
//
// Returns an error if the alias name conflicts with a built-in validator.
func RegisterAlias(alias, expandsTo string) error {
	if alias == "" {
		return errors.New("alias name cannot be empty")
	}
	if expandsTo == "" {
		return errors.New("alias tags cannot be empty")
	}
	if isBuiltInValidator(alias) {
		return fmt.Errorf("cannot override built-in validator: %s", alias)
	}

	aliases.Store(alias, expandsTo)
	clearValidatorCache()
	return nil
}

// GetAlias retrieves a registered alias expansion.
// Returns the expansion and true if found, empty string and false otherwise.
func GetAlias(name string) (string, bool) {
	if v, ok := aliases.Load(name); ok {
		return v.(string), true
	}
	return "", false
}

// clearValidatorCache clears all cached validators to pick up new registrations.
// This ensures that newly registered validators are used by existing validator instances.
func clearValidatorCache() {
	validatorCache.Range(func(key, value any) bool {
		validatorCache.Delete(key)
		return true
	})
}

// isBuiltInValidator returns true if the name is a built-in validator.
// Built-in validators include: required, email, min, max, len, regex, etc.
func isBuiltInValidator(name string) bool {
	builtInValidators := map[string]bool{
		// Core
		"required": true, "const": true,
		// Reserved for future use (not implemented, but prevents custom validator conflicts)
		"omitempty": true,
		// String
		"min": true, "max": true, "len": true, "regex": true, "regexp": true, "pattern": true,
		"email": true, "url": true, "uri": true, "uuid": true,
		"alpha": true, "alphanum": true, "alphanumunicode": true,
		"ascii": true, "contains": true, "excludes": true,
		"startswith": true, "endswith": true, "lowercase": true, "uppercase": true,
		"oneof": true, "oneofci": true, "enum": true,
		// Built-in aliases
		"iscolor": true,
		// Numeric
		"gt": true, "gte": true, "lt": true, "lte": true,
		"multipleOf": true, "positive": true, "negative": true,
		// Network
		"ip": true, "ipv4": true, "ipv6": true, "cidr": true,
		"mac": true, "hostname": true, "fqdn": true, "port": true,
		// Format
		"datetime": true, "date": true, "time": true,
		"base64": true, "json": true, "jwt": true,
		"creditcard": true, "isbn": true, "ssn": true,
		// Collections
		"dive": true, "keys": true, "endkeys": true, "unique": true,
		// Cross-field
		"eqfield": true, "nefield": true, "gtfield": true, "ltfield": true,
		"required_if": true, "excluded_if": true,
	}
	return builtInValidators[name]
}

// getTagNameFunc returns the current tag name function or nil if not set.
func getTagNameFunc() TagNameFunc {
	ptr := tagNameFunc.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}

// resolveFieldName returns the field name using the custom function or defaults.
// Default behavior: use JSON tag if present, otherwise use field name.
func resolveFieldName(field *reflect.StructField) string {
	if fn := getTagNameFunc(); fn != nil {
		if name := fn(*field); name != "" {
			return name
		}
	}
	// Default: use JSON tag if present
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		if comma := strings.Index(jsonTag, ","); comma != -1 {
			jsonTag = jsonTag[:comma]
		}
		if jsonTag != "" && jsonTag != "-" {
			return jsonTag
		}
	}
	return field.Name
}
