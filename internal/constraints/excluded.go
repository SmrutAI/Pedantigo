package constraints

import (
	"fmt"
	"reflect"
)

// excludedIfConstraint: field must be absent (zero) if another field equals a specific value
// ValidateCrossField validates the field against another field in the struct
func (c excludedIfConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if CompareToString(targetValue) == c.compareValue {
		// Condition is met - field must be zero
		if !IsZeroValue(fieldValue) {
			return fmt.Errorf("must be absent when %s equals '%s'", c.targetFieldName, c.compareValue)
		}
	}
	return nil
}

// excludedUnlessConstraint: field must be absent unless another field equals a specific value
// ValidateCrossField validates the field against another field in the struct
func (c excludedUnlessConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if CompareToString(targetValue) != c.compareValue {
		// Condition is met - field must be zero
		if !IsZeroValue(fieldValue) {
			return fmt.Errorf("must be absent unless %s equals '%s'", c.targetFieldName, c.compareValue)
		}
	}
	return nil
}

// excludedWithConstraint: field must be absent if another field is non-zero
// ValidateCrossField validates the field against another field in the struct
func (c excludedWithConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if !IsZeroValue(targetValue) {
		// Target field is present - this field must be absent
		if !IsZeroValue(fieldValue) {
			return fmt.Errorf("must be absent when %s is present", c.targetFieldName)
		}
	}
	return nil
}

// excludedWithoutConstraint: field must be absent if another field is zero
// ValidateCrossField validates the field against another field in the struct
func (c excludedWithoutConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if IsZeroValue(targetValue) {
		// Target field is absent - this field must also be absent
		if !IsZeroValue(fieldValue) {
			return fmt.Errorf("must be absent when %s is absent", c.targetFieldName)
		}
	}
	return nil
}
