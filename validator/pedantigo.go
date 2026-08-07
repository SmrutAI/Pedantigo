// package validator provides Pydantic-inspired validation for Go.
//
// Pedantigo offers two APIs: a Simple API for most use cases and a Validator API
// for advanced scenarios requiring custom options.
//
// # Simple API (Recommended)
//
// Global functions with automatic caching - no setup needed:
//
//	type User struct {
//	    Email string `json:"email" validate:"required,email"`
//	    Age   int    `json:"age" validate:"min=18,max=120"`
//	}
//
//	// Parse JSON and validate
//	user, err := vl.Unmarshal[User](jsonData)
//
//	// Create from JSON, map, or struct
//	user, err := vl.NewModel[User](input)
//
//	// Get cached JSON Schema
//	schema := vl.Schema[User]()
//
// # Validator API (Advanced)
//
// For custom options like strict mode or extra field handling:
//
//	vl := validator.New[User](validator.Options{
//	    StrictMissingFields: true,
//	    ExtraFields:         validator.ExtraForbid,
//	})
//	user, err := vl.Unmarshal(jsonData)
//
// # Key Features
//
//   - 100+ built-in validation constraints
//   - JSON Schema generation with 240x caching speedup
//   - Streaming validation for partial JSON (LLM support)
//   - Discriminated unions with type-safe handling
//   - Cross-field validation
//   - Custom validator registration
//
// See https://pedantigo.dev for complete documentation.
package validator

// Validatable is an interface for types that implement custom validation.
// When a struct implements this interface, its Validate method is called
// after all field-level validations pass.
//
// Example:
//
//	type DateRange struct {
//	    Start time.Time `json:"start" validate:"required"`
//	    End   time.Time `json:"end" validate:"required"`
//	}
//
//	func (d *DateRange) Validate() error {
//	    if d.End.Before(d.Start) {
//	        return errors.New("end must be after start")
//	    }
//	    return nil
//	}
type Validatable interface {
	Validate() error
}
