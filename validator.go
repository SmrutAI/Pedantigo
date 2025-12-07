package pedantigo

import "reflect"

// Validator validates structs of type T
type Validator[T any] struct {
	typ reflect.Type
}

// New creates a new Validator for type T
func New[T any]() *Validator[T] {
	var zero T
	return &Validator[T]{
		typ: reflect.TypeOf(zero),
	}
}

// Validate validates a struct and returns any validation errors
func (v *Validator[T]) Validate(obj *T) ValidationErrors {
	// TODO: implement validation tree walker
	return nil
}

// Unmarshal unmarshals JSON data, applies defaults, and validates
func (v *Validator[T]) Unmarshal(data []byte) (*T, ValidationErrors) {
	// TODO: implement unmarshal + validate
	return nil, nil
}

// Marshal validates and marshals struct to JSON
func (v *Validator[T]) Marshal(obj *T) ([]byte, ValidationErrors) {
	// TODO: implement validate + marshal
	return nil, nil
}
