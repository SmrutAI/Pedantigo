package constraints

import (
	"reflect"
)

// uniqueConstraint validates that collection elements are unique.
// For slices: no duplicate elements.
// For maps: no duplicate values.
// For struct slices with field param: no duplicate field values.
type uniqueConstraint struct {
	field string // optional: for struct slices, e.g. "ID"
}

// Validate checks if collection elements are unique.
func (c uniqueConstraint) Validate(value any) error {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return c.validateSlice(v)
	case reflect.Map:
		return c.validateMap(v)
	default:
		// Non-collection types pass (validation is handled at validator.go level)
		return nil
	}
}

// validateSlice validates uniqueness of slice elements.
func (c uniqueConstraint) validateSlice(v reflect.Value) error {
	if v.Len() == 0 {
		return nil
	}

	seen := make(map[any]bool)
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		var key any
		if c.field != "" {
			// Struct slice: extract field value
			key = c.extractFieldValue(elem, c.field)
		} else {
			// Simple slice: use element as key
			key = c.toComparable(elem)
		}

		if key == nil {
			continue // Skip non-comparable or nil elements
		}

		if seen[key] {
			if c.field != "" {
				return NewConstraintErrorf(CodeNotUnique, "duplicate value for field %s", c.field)
			}
			return NewConstraintError(CodeNotUnique, "contains duplicate values")
		}
		seen[key] = true
	}
	return nil
}

// validateMap validates uniqueness of map values.
func (c uniqueConstraint) validateMap(v reflect.Value) error {
	if v.Len() == 0 {
		return nil
	}

	seen := make(map[any]bool)
	iter := v.MapRange()
	for iter.Next() {
		val := iter.Value()
		key := c.toComparable(val)

		if key == nil {
			continue // Skip non-comparable values
		}

		if seen[key] {
			return NewConstraintError(CodeNotUnique, "contains duplicate values")
		}
		seen[key] = true
	}
	return nil
}

// extractFieldValue extracts a field value from a struct element.
func (c uniqueConstraint) extractFieldValue(elem reflect.Value, fieldName string) any {
	// Dereference pointer if needed
	if elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			return nil
		}
		elem = elem.Elem()
	}

	if elem.Kind() != reflect.Struct {
		return nil
	}

	field := elem.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	return c.toComparable(field)
}

// toComparable converts a reflect.Value to a comparable any type.
func (c uniqueConstraint) toComparable(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Only comparable types can be map keys
	if v.Type().Comparable() {
		return v.Interface()
	}
	return nil
}
