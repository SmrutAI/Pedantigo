package pedantigo

import "sync/atomic"

// DefaultTagName is the default struct tag name used by Pedantigo.
// This is exported for reference but users should use GetTagName() for the current value.
const DefaultTagName = "pedantigo"

var (
	// globalTagName stores the current struct tag name (default: "pedantigo").
	globalTagName atomic.Value

	// validatorCreated tracks if any validator has been created.
	// This is used to enforce that SetTagName is called before any validators.
	validatorCreated atomic.Bool
)

func init() {
	globalTagName.Store(DefaultTagName)
}

// SetTagName sets the global default struct tag name.
//
// IMPORTANT: This function MUST be called in init() or at the very start of main(),
// BEFORE any other Pedantigo functions are called. Calling it after any validator
// has been created will cause a panic.
//
// This allows Pedantigo to be used with existing struct tags from other validation
// libraries like go-playground/validator.
//
// Example:
//
//	func init() {
//	    pedantigo.SetTagName("validate") // Now uses `validate:"required,email"`
//	}
//
//	type User struct {
//	    Email string `json:"email" validate:"required,email"`
//	}
func SetTagName(name string) {
	if validatorCreated.Load() {
		panic("pedantigo: SetTagName must be called before any validators are created. " +
			"Call it in init() or at the start of main().")
	}

	if name == "" {
		name = DefaultTagName
	}
	globalTagName.Store(name)
}

// GetTagName returns the current global tag name.
// By default, this is "pedantigo".
func GetTagName() string {
	return globalTagName.Load().(string)
}

// markValidatorCreated is called when a new validator is created.
// This marks that SetTagName can no longer be called safely.
func markValidatorCreated() {
	validatorCreated.Store(true)
}

// hasValidatorBeenCreated returns whether any validator has been created.
// This is exported for testing purposes only.
func hasValidatorBeenCreated() bool {
	return validatorCreated.Load()
}

// resetTagNameForTesting resets the global tag name to default.
// This should ONLY be used in tests.
func resetTagNameForTesting() {
	globalTagName.Store(DefaultTagName)
}

// resetValidatorCreatedForTesting resets the validatorCreated flag.
// This should ONLY be used in tests.
func resetValidatorCreatedForTesting() {
	validatorCreated.Store(false)
}
