// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

// CustomValidationFunc is the signature for custom validators with parameter support.
// Following go-playground/validator pattern, validators receive:
// - value: The field value being validated
// - param: The parameter from the tag (e.g., "PRE_" from "has_prefix=PRE_"), empty if no param
//
// Returns nil if valid, error with message if invalid.
type CustomValidationFunc func(value any, param string) error

// customValidatorLookup is set by the registry package to allow constraint building
// to look up custom validators. This avoids import cycles.
var customValidatorLookup func(name string) (CustomValidationFunc, bool)

// ctxValidatorLookup is set by the registry package to check if a validator name
// is a context-aware validator. Returns true if the name is registered as a context validator.
var ctxValidatorLookup func(name string) bool

// SetCustomValidatorLookup sets the function used to look up custom validators.
// This should be called once by the registry package during initialization.
func SetCustomValidatorLookup(fn func(name string) (CustomValidationFunc, bool)) {
	customValidatorLookup = fn
}

// SetCtxValidatorLookup sets the function used to check if a validator is context-aware.
// This should be called once by the registry package during initialization.
func SetCtxValidatorLookup(fn func(name string) bool) {
	ctxValidatorLookup = fn
}

// IsContextValidator checks if the given name is a registered context-aware validator.
func IsContextValidator(name string) bool {
	if ctxValidatorLookup == nil {
		return false
	}
	return ctxValidatorLookup(name)
}

// ExtractContextValidators extracts context validator names and params from a constraints map.
// Returns a slice of ContextConstraint for validators that should be called with context.
func ExtractContextValidators(constraintsMap map[string]string) []ContextConstraint {
	var result []ContextConstraint
	for name, value := range constraintsMap {
		// Skip special prefixes
		if len(name) > 6 && name[:6] == "__or__" {
			continue
		}
		// Check if this is a context-aware validator
		if IsContextValidator(name) {
			result = append(result, ContextConstraint{
				Name:  name,
				Param: value,
			})
		}
	}
	return result
}

// customConstraint wraps a custom validator function as a Constraint.
type customConstraint struct {
	name  string               // Tag name (e.g., "us_phone")
	fn    CustomValidationFunc // The validation function
	param string               // Parameter from tag (e.g., "PRE_" from "has_prefix=PRE_")
}

// Validate implements the Constraint interface for custom validators.
func (c customConstraint) Validate(value any) error {
	// Call the wrapped validation function
	err := c.fn(value, c.param)
	if err == nil {
		return nil
	}

	// Wrap the error in a ConstraintError with CodeCustomValidation
	return NewConstraintError(CodeCustomValidation, c.name+": "+err.Error())
}

// BuildCustomConstraint attempts to build a constraint for a custom validator.
// Returns (constraint, true) if a custom validator with the given name exists,
// Returns (nil, false) if no such validator is registered.
//
// Parameters:
//   - name: The validator tag name (e.g., "us_phone").
//   - param: The parameter from the tag (e.g., "PRE_" from "has_prefix=PRE_").
func BuildCustomConstraint(name, param string) (Constraint, bool) {
	// Check if the lookup function is set
	if customValidatorLookup == nil {
		return nil, false
	}

	// Attempt to look up the custom validator
	fn, found := customValidatorLookup(name)
	if !found {
		return nil, false
	}

	// Return a customConstraint wrapping the validator function
	return customConstraint{
		name:  name,
		fn:    fn,
		param: param,
	}, true
}
