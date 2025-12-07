// Package pedantigo provides Pydantic-inspired validation for Go with struct tags,
// JSON schema generation, and performance optimizations.
//
// Basic usage:
//
//	type User struct {
//	    Email string `json:"email" validate:"required,email"`
//	    Age   int    `json:"age" validate:"min=18,max=120"`
//	}
//
//	validator := pedantigo.New[User]()
//	user, errs := validator.Unmarshal(jsonData)
package pedantigo

// Validatable is an interface for types that implement custom validation
type Validatable interface {
	Validate() error
}
