package constraints

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// CrossFieldConstraint represents a validation constraint that compares two fields
type CrossFieldConstraint interface {
	ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error
}

// Cross-field constraint types
type (
	eqFieldConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	neFieldConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	gtFieldConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	gteFieldConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	ltFieldConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	lteFieldConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	requiredIfConstraint struct {
		targetFieldName  string
		targetFieldIndex int
		compareValue     string
	}
	requiredUnlessConstraint struct {
		targetFieldName  string
		targetFieldIndex int
		compareValue     string
	}
	requiredWithConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	requiredWithoutConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	excludedIfConstraint struct {
		targetFieldName  string
		targetFieldIndex int
		compareValue     string
	}
	excludedUnlessConstraint struct {
		targetFieldName  string
		targetFieldIndex int
		compareValue     string
	}
	excludedWithConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
	excludedWithoutConstraint struct {
		targetFieldName  string
		targetFieldIndex int
	}
)

// BuildCrossFieldConstraintsForField builds cross-field constraint instances from parsed tags
func BuildCrossFieldConstraintsForField(constraints map[string]string, structType reflect.Type, fieldIndex int) []CrossFieldConstraint {
	var result []CrossFieldConstraint

	fieldName := structType.Field(fieldIndex).Name

	for name, value := range constraints {
		switch name {
		case "eqfield":
			targetIdx := resolveAndValidateField(structType, value, fieldIndex, fieldName, "eqfield")
			result = append(result, eqFieldConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "nefield":
			targetIdx := resolveAndValidateField(structType, value, fieldIndex, fieldName, "nefield")
			result = append(result, neFieldConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "gtfield":
			targetIdx := resolveAndValidateField(structType, value, fieldIndex, fieldName, "gtfield")
			result = append(result, gtFieldConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "gtefield":
			targetIdx := resolveAndValidateField(structType, value, fieldIndex, fieldName, "gtefield")
			result = append(result, gteFieldConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "ltfield":
			targetIdx := resolveAndValidateField(structType, value, fieldIndex, fieldName, "ltfield")
			result = append(result, ltFieldConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "ltefield":
			targetIdx := resolveAndValidateField(structType, value, fieldIndex, fieldName, "ltefield")
			result = append(result, lteFieldConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "required_if":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, ":"); ok {
				targetIdx := resolveFieldIndexSilent(structType, fieldName)
				result = append(result, requiredIfConstraint{targetFieldName: fieldName, targetFieldIndex: targetIdx, compareValue: compareValue})
			}
		case "required_unless":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, ":"); ok {
				targetIdx := resolveFieldIndexSilent(structType, fieldName)
				result = append(result, requiredUnlessConstraint{targetFieldName: fieldName, targetFieldIndex: targetIdx, compareValue: compareValue})
			}
		case "required_with":
			targetIdx := resolveFieldIndexSilent(structType, value)
			result = append(result, requiredWithConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "required_without":
			targetIdx := resolveFieldIndexSilent(structType, value)
			result = append(result, requiredWithoutConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "excluded_if":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				targetIdx := resolveFieldIndexSilent(structType, fieldName)
				result = append(result, excludedIfConstraint{targetFieldName: fieldName, targetFieldIndex: targetIdx, compareValue: compareValue})
			}
		case "excluded_unless":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				targetIdx := resolveFieldIndexSilent(structType, fieldName)
				result = append(result, excludedUnlessConstraint{targetFieldName: fieldName, targetFieldIndex: targetIdx, compareValue: compareValue})
			}
		case "excluded_with":
			targetIdx := resolveFieldIndexSilent(structType, value)
			result = append(result, excludedWithConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		case "excluded_without":
			targetIdx := resolveFieldIndexSilent(structType, value)
			result = append(result, excludedWithoutConstraint{targetFieldName: value, targetFieldIndex: targetIdx})
		}
	}

	return result
}

// resolveFieldIndexSilent resolves a field name to its index, returning -1 if not found
func resolveFieldIndexSilent(structType reflect.Type, fieldName string) int {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		// Only match exported fields with exact case-sensitive name match
		if field.IsExported() && field.Name == fieldName {
			return i
		}
	}
	return -1 // Field not found or not exported
}

// ============================================================================
// Helper Functions for Cross-Field Validation
// ============================================================================

// CheckTypeCompatibility checks if two values can be compared
func CheckTypeCompatibility(a, b any) error {
	aType := Dereference(reflect.TypeOf(a))
	bType := Dereference(reflect.TypeOf(b))

	// Handle nil values
	if a == nil && b == nil {
		return nil // Both nil are compatible
	}

	if a == nil || b == nil {
		// One is nil, check if we can compare
		// Only allow if both are pointer types (one nil, one not)
		aVal := reflect.ValueOf(a)
		bVal := reflect.ValueOf(b)
		if aVal.Kind() == reflect.Ptr || bVal.Kind() == reflect.Ptr {
			// At least one is a pointer type, this is incompatible
			return fmt.Errorf("cannot compare nil with non-nil value")
		}
		return fmt.Errorf("cannot compare nil with non-nil value")
	}

	// Check if both are numeric types
	if IsNumericType(aType) && IsNumericType(bType) {
		return nil // Numeric types are always compatible
	}

	// Check if both are strings
	if aType.Kind() == reflect.String && bType.Kind() == reflect.String {
		return nil
	}

	// Check if both are bools
	if aType.Kind() == reflect.Bool && bType.Kind() == reflect.Bool {
		return nil
	}

	// Check if both are time.Time
	if aType == reflect.TypeOf(time.Time{}) && bType == reflect.TypeOf(time.Time{}) {
		return nil
	}

	return fmt.Errorf("cannot compare types %v and %v", aType, bType)
}

// Dereference removes pointer indirection from a type
func Dereference(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// IsNumericType checks if a type is numeric
func IsNumericType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// Compare returns -1 if a < b, 0 if a == b, 1 if a > b
// This works for strings and numeric types
// Compare compares two values
func Compare(a, b any) int {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Check if both are nil pointers
	aIsNil := aVal.Kind() == reflect.Ptr && aVal.IsNil()
	bIsNil := bVal.Kind() == reflect.Ptr && bVal.IsNil()

	if aIsNil && bIsNil {
		return 0 // Both nil are equal
	}
	if aIsNil {
		return -1 // nil is less than non-nil
	}
	if bIsNil {
		return 1 // non-nil is greater than nil
	}

	// Dereference pointers
	if aVal.Kind() == reflect.Ptr {
		aVal = aVal.Elem()
	}
	if bVal.Kind() == reflect.Ptr {
		bVal = bVal.Elem()
	}

	// String comparison
	if aVal.Kind() == reflect.String && bVal.Kind() == reflect.String {
		if aVal.String() < bVal.String() {
			return -1
		} else if aVal.String() > bVal.String() {
			return 1
		}
		return 0
	}

	// Numeric comparison
	var aNum, bNum float64

	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		aNum = float64(aVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		aNum = float64(aVal.Uint())
	case reflect.Float32, reflect.Float64:
		aNum = aVal.Float()
	default:
		return 0 // Can't compare
	}

	switch bVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bNum = float64(bVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bNum = float64(bVal.Uint())
	case reflect.Float32, reflect.Float64:
		bNum = bVal.Float()
	default:
		return 0 // Can't compare
	}

	if aNum < bNum {
		return -1
	} else if aNum > bNum {
		return 1
	}
	return 0
}

// CompareToString converts any value to string for comparison
func CompareToString(value any) string {
	val := reflect.ValueOf(value)

	// Handle pointer types
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	default:
		return fmt.Sprintf("%v", value)
	}
}

// resolveAndValidateField resolves a field, validates it exists and is not self-referencing, panics on error
func resolveAndValidateField(structType reflect.Type, targetFieldName string, currentFieldIndex int, currentFieldName string, constraintName string) int {
	targetIdx := resolveFieldIndexSilent(structType, targetFieldName)
	if targetIdx == -1 {
		panic(fmt.Sprintf("field %s references non-existent field %s in %s constraint", currentFieldName, targetFieldName, constraintName))
	}
	if targetIdx == currentFieldIndex {
		panic(fmt.Sprintf("field %s cannot reference itself in %s constraint", currentFieldName, constraintName))
	}
	return targetIdx
}

// parseConditionalConstraint parses "field:value" or "field value" syntax
// Returns (fieldName, compareValue, true) on success, ("", "", false) on failure
func parseConditionalConstraint(value string, separator string) (string, string, bool) {
	parts := strings.SplitN(value, separator, 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
