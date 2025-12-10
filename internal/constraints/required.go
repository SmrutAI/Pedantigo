package constraints

import (
	"fmt"
	"reflect"
)

// requiredIfConstraint: field is required if another field equals a specific value
// ValidateCrossField validates the field against another field in the struct
func (c requiredIfConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if CompareToString(targetValue) == c.compareValue {
		// Condition is met - field must be non-zero
		if IsZeroValue(fieldValue) {
			return fmt.Errorf("is required when %s equals '%s'", c.targetFieldName, c.compareValue)
		}
	}
	return nil
}

// requiredUnlessConstraint: field is required unless another field equals a specific value
// ValidateCrossField validates the field against another field in the struct
func (c requiredUnlessConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if CompareToString(targetValue) != c.compareValue {
		// Condition is met - field must be non-zero
		if IsZeroValue(fieldValue) {
			return fmt.Errorf("is required unless %s equals '%s'", c.targetFieldName, c.compareValue)
		}
	}
	return nil
}

// requiredWithConstraint: field is required if another field is non-zero
// ValidateCrossField validates the field against another field in the struct
func (c requiredWithConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if !IsZeroValue(targetValue) {
		// Target field is present - this field must also be present
		if IsZeroValue(fieldValue) {
			return fmt.Errorf("is required when %s is present", c.targetFieldName)
		}
	}
	return nil
}

// requiredWithoutConstraint: field is required if another field is zero
// ValidateCrossField validates the field against another field in the struct
func (c requiredWithoutConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	targetValue := structValue.Field(c.targetFieldIndex).Interface()

	if IsZeroValue(targetValue) {
		// Target field is absent - this field must be present
		if IsZeroValue(fieldValue) {
			return fmt.Errorf("is required when %s is absent", c.targetFieldName)
		}
	}
	return nil
}
