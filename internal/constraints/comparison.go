package constraints

import (
	"reflect"
)

// ValidateCrossField for eqFieldConstraint: field must equal another field.
func (c eqFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) != 0 {
		return NewConstraintErrorf(CodeMustEqualField, "must equal field %s", c.targetFieldName)
	}
	return nil
}

// ValidateCrossField for neFieldConstraint: field must NOT equal another field.
func (c neFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) == 0 {
		return NewConstraintErrorf(CodeMustNotEqualField, "must not equal field %s", c.targetFieldName)
	}
	return nil
}

// ValidateCrossField for gtFieldConstraint: field must be > another field.
func (c gtFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) <= 0 {
		return NewConstraintErrorf(CodeMustBeGTField, "must be greater than field %s", c.targetFieldName)
	}
	return nil
}

// ValidateCrossField for gteFieldConstraint: field must be >= another field.
func (c gteFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) < 0 {
		return NewConstraintErrorf(CodeMustBeGTEField, "must be at least field %s", c.targetFieldName)
	}
	return nil
}

// ValidateCrossField for ltFieldConstraint: field must be < another field.
func (c ltFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) >= 0 {
		return NewConstraintErrorf(CodeMustBeLTField, "must be less than field %s", c.targetFieldName)
	}
	return nil
}

// ValidateCrossField for lteFieldConstraint: field must be <= another field.
func (c lteFieldConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	// Check type compatibility
	if err := CheckTypeCompatibility(fieldValue, targetValue); err != nil {
		return err
	}

	if Compare(fieldValue, targetValue) > 0 {
		return NewConstraintErrorf(CodeMustBeLTEField, "must be at most field %s", c.targetFieldName)
	}
	return nil
}
