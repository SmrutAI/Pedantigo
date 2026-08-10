package validator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/SmrutAI/pedantigo/v2/validator/internal/constraints"
	"github.com/SmrutAI/pedantigo/v2/validator/internal/tags"
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
			// Convert validator.ValidationFunc to constraints.CustomValidationFunc
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

	// simpleAPICache stores validators lazily built for the Simple API's package-level
	// convenience functions (Validate, Unmarshal[T], Schema[T], etc.). Every entry is
	// disposable: losing one costs a single New[T]() rebuild on next use, invisible to
	// the caller. RegisterValidation/RegisterValidationCtx/RegisterTagNameFunc/RegisterAlias
	// wipe this cache (via clearValidatorCache) because a new custom validator, alias, or
	// tag-name function may make an already-built entry's field cache stale, and there is
	// no cheap way to know which entries are affected.
	// Stores map[reflect.Type]any (*Validator[T]).
	simpleAPICache sync.Map

	// registeredValidators stores validators explicitly bound via Register[T](), for
	// framework-plugin lookup (UnmarshalInto, ValidateInto, the Echo Binder, Gin's
	// NewBinder). Entries here are permanent for the process's lifetime — Register[T]()
	// may be called exactly once per type, and nothing else may ever remove an entry.
	// This is a deliberately separate map from simpleAPICache: the two have incompatible
	// lifetimes, and sharing storage let RegisterValidation/RegisterAlias/etc. silently
	// evict Register()'d bindings as a side effect neither side intended.
	// Stores map[reflect.Type]any (*Validator[T]).
	registeredValidators sync.Map

	// registeredTagName stores the single effective tag name used by all
	// Register()-visible validators in this process. nil means nothing has
	// been registered/required yet.
	registeredTagName atomic.Pointer[string]

	// tagNameFunc stores the custom tag name resolution function.
	tagNameFunc atomic.Pointer[TagNameFunc]
)

// getOrCreateValidator returns a cached validator for type T, creating one if needed.
// Thread-safe: uses LoadOrStore to ensure only one validator is created per type.
func getOrCreateValidator[T any]() *Validator[T] {
	var zero T
	typ := reflect.TypeOf(zero)

	if cached, ok := simpleAPICache.Load(typ); ok {
		return cached.(*Validator[T])
	}

	vl := New[T]()
	actual, _ := simpleAPICache.LoadOrStore(typ, vl) //nolint:not-an-error // LoadOrStore returns bool, not error
	return actual.(*Validator[T])
}

// UnmarshalInto unmarshals JSON data into the target using the cached validator
// for target's type. target must be a non-nil pointer to a struct whose type has
// been registered via Register() (which populates the validator cache).
//
// If no validator is cached for target's type, UnmarshalInto panics with a message
// naming the missing type and the registration call needed.
//
// This function enables framework integrations (Echo Binder, Gin middleware) where
// the target type is known only at runtime via reflect.Type, not at compile time via T.
//
// Example:
//
//	var _ = validator.Register(validator.New[MyRequest]())  // register at init time
//
//	// later, at runtime:
//	var req MyRequest
//	err := validator.UnmarshalInto(jsonBody, &req)
func UnmarshalInto(data []byte, target any) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		panic("validator: UnmarshalInto target must be a non-nil pointer")
	}
	typ := rv.Elem().Type()
	cached, ok := registeredValidators.Load(typ)
	if !ok {
		panic(fmt.Sprintf(
			"validator: no validator registered for type %s.%s. "+
				"UnmarshalInto (and any framework plugin built on it, e.g. the Echo Binder) "+
				"can only validate types that were explicitly registered via Register(). "+
				"Fix: add this once, at package init time, for this type:\n"+
				"    var _ = validator.Register(validator.New[%s]())",
			typ.PkgPath(), typ.Name(), typ.Name()))
	}
	return cached.(unmarshalable).unmarshalInto(data, target)
}

// ValidateInto looks up the registered Validator[T] for obj's concrete type and validates it.
// obj must be a non-nil pointer — every real caller (Gin's binding functions populate the
// target via reflection, which requires addressability) already guarantees this, so a caller
// that violates it has a bug, and ValidateInto returns a clear error rather than silently
// reporting success. It does not panic (unlike UnmarshalInto's equivalent check), because the
// StructValidator interface this feeds into must never panic — returning a non-nil error is
// fully compatible with that contract.
//
// Unlike UnmarshalInto, ValidateInto does not error when the pointed-to type was never
// registered — it returns nil, since this is called from framework validator adapters for
// every value they see, not just ones a caller deliberately registered via Register().
func ValidateInto(obj any) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("validator: ValidateInto requires a pointer, got %T", obj)
	}
	if rv.IsNil() {
		return fmt.Errorf("validator: ValidateInto requires a non-nil pointer, got nil %T", obj)
	}
	typ := rv.Elem().Type()
	cached, ok := registeredValidators.Load(typ)
	if !ok {
		return nil
	}
	return cached.(validatableInto).validateInto(obj)
}

// Register makes v the instance that framework-plugin binders (UnmarshalInto,
// the Echo Binder, etc.) will find for type T. Custom code that only ever calls
// v.Unmarshal() / v.Validate() directly never needs this.
//
// Register may be called exactly once per type T. A second call for the same
// type — even with an identical instance — panics. Duplicate registration is
// the caller's responsibility to avoid; pedantigo enforces it rather than
// silently picking a winner, because a type may legitimately have multiple
// differently-configured validators (different Options), and only
// one of them can be "the" plugin-visible instance.
func Register[T any](v *Validator[T]) *Validator[T] {
	checkOrSeedTagName(v.tagName)

	typ := v.typ
	_, loaded := registeredValidators.LoadOrStore(typ, v)
	if loaded {
		panic(fmt.Sprintf(
			"validator: validator for type %s.%s is already registered. "+
				"Register[T]() may be called exactly once per type — this is enforced "+
				"because a type may have multiple validators built with different "+
				"Options (e.g. different TagName or StrictMissingFields), and "+
				"pedantigo cannot know which one should be visible to plugin-based "+
				"lookups like UnmarshalInto/the Echo Binder if more than one is registered. "+
				"Fix: find the other validator.Register(validator.New[%s](...)) call for this "+
				"type and remove one of them, so Register is called from exactly one "+
				"package-level var declaration.",
			typ.PkgPath(), typ.Name(), typ.Name()))
	}
	return v
}

// RequireSingleRegisteredTagName is called once by a framework plugin's setup.
// If a type was already Register()'d, it verifies the tag name matches want.
// If nothing has been Register()'d yet, it seeds want as the required tag name.
func RequireSingleRegisteredTagName(want string) {
	checkOrSeedTagName(want)
}

// checkOrSeedTagName is the single race-free entry point for registeredTagName.
// CompareAndSwap(nil, &name) atomically claims the "nothing registered yet" state —
// exactly one concurrent caller can ever win that transition. Every other caller
// (whether it lost the race or the tag was already set long ago) falls through and
// compares against whatever is now there. A plain Load-then-Store here would leave a
// window where two concurrent callers with different tag names could both observe nil
// and both store, silently defeating the single-tag-name-per-process guarantee.
func checkOrSeedTagName(name string) {
	if registeredTagName.CompareAndSwap(nil, &name) {
		return
	}
	got := registeredTagName.Load()
	if *got != name {
		panic(fmt.Sprintf(
			"pedantigo: registered tag name %q does not match this framework setup's expected tag name %q",
			*got, name))
	}
}

// unmarshalable is a non-generic interface that allows type-erased unmarshal.
// Implemented by Validator[T] (see validator.go) to enable UnmarshalInto.
type unmarshalable interface {
	unmarshalInto(data []byte, target any) error
}

// validatableInto is a non-generic interface that allows type-erased validation.
// Implemented by Validator[T] to enable ValidateInto.
type validatableInto interface {
	validateInto(obj any) error
}

func resetRegisteredTagNameForTesting() {
	registeredTagName.Store(nil)
}

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
	simpleAPICache.Delete(t)
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
//	validator.RegisterValidationCtx("db_unique", func(ctx context.Context, value any, param string) error {
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
//	validator.RegisterTagNameFunc(func(field reflect.StructField) string {
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
//	validator.RegisterAlias("iscolor", "hexcolor|rgb|rgba|hsl|hsla")
//	// Now `iscolor` expands to an OR constraint for all color formats
//
//	validator.RegisterAlias("username", "required,alphanum,min=3,max=20")
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

// clearValidatorCache clears the Simple API's disposable cache to pick up new registrations.
// This ensures that newly registered custom validators/aliases/tag-name functions are used
// by validators built after this call. It deliberately never touches registeredValidators —
// Register()'d bindings must survive this unconditionally; see registeredValidators' doc comment.
func clearValidatorCache() {
	simpleAPICache.Range(func(key, value any) bool {
		simpleAPICache.Delete(key)
		return true
	})
}

// isBuiltInValidator returns true if the name is a built-in validator.
// Built-in validators include: required, email, min, max, len, regex, etc.
func isBuiltInValidator(name string) bool {
	builtInValidators := map[string]bool{
		// Core
		constraints.CRequired: true, "const": true,
		// When present, skips all constraints if the field is at its zero value (empty string, nil pointer/slice/map).
		"omitempty": true,
		// String
		constraints.CMin: true, constraints.CMax: true, constraints.CLen: true, "regex": true, constraints.CRegexp: true, "pattern": true,
		constraints.CEmail: true, constraints.CUrl: true, constraints.CUri: true, constraints.CUuid: true,
		constraints.CAlpha: true, constraints.CAlphanum: true, constraints.CAlphanumunicode: true,
		constraints.CAscii: true, constraints.CContains: true, constraints.CExcludes: true,
		constraints.CStartswith: true, constraints.CEndswith: true, constraints.CLowercase: true, constraints.CUppercase: true,
		constraints.COneof: true, constraints.COneofci: true, "enum": true,
		// Built-in aliases
		"iscolor": true,
		// Numeric
		constraints.CGt: true, constraints.CGte: true, constraints.CLt: true, constraints.CLte: true,
		"multipleOf": true, constraints.CPositive: true, constraints.CNegative: true,
		// Network
		constraints.CIp: true, constraints.CIpv4: true, constraints.CIpv6: true, constraints.CCidr: true,
		constraints.CMac: true, constraints.CHostname: true, constraints.CFqdn: true, constraints.CPort: true,
		// Format
		constraints.CDatetime: true, "date": true, "time": true,
		constraints.CBase64: true, constraints.CJson: true, constraints.CJwt: true,
		"creditcard": true, constraints.CIsbn: true, constraints.CSsn: true,
		// Collections
		"dive": true, "keys": true, "endkeys": true, constraints.CUnique: true,
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
