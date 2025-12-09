package constraints

import (
	"fmt"
	"reflect"
)

// eqFieldConstraint: field must equal another field
func (c eqFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) != 0 {
		return fmt.Errorf("must equal field %s", c.targetFieldName)
	}
	return nil
}

// neFieldConstraint: field must NOT equal another field
func (c neFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) == 0 {
		return fmt.Errorf("must not equal field %s", c.targetFieldName)
	}
	return nil
}

// gtFieldConstraint: field must be > another field
func (c gtFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) <= 0 {
		return fmt.Errorf("must be greater than field %s", c.targetFieldName)
	}
	return nil
}

// gteFieldConstraint: field must be >= another field
func (c gteFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) < 0 {
		return fmt.Errorf("must be at least field %s", c.targetFieldName)
	}
	return nil
}

// ltFieldConstraint: field must be < another field
func (c ltFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) >= 0 {
		return fmt.Errorf("must be less than field %s", c.targetFieldName)
	}
	return nil
}

// lteFieldConstraint: field must be <= another field
func (c lteFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) > 0 {
		return fmt.Errorf("must be at most field %s", c.targetFieldName)
	}
	return nil
}
