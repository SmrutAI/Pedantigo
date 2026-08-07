package constraints

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// boolStringTrue is the string form of a true boolean value in a skip_unless comparison.
const boolStringTrue = "true"

// CrossFieldConstraint represents a validation constraint that compares two fields.
type CrossFieldConstraint interface {
	ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error
}

// SkipConstraint is evaluated BEFORE other constraints.
// If ShouldSkip returns true, ALL validation on this field is bypassed.
type SkipConstraint interface {
	ShouldSkip(structValue reflect.Value) bool
}

// Cross-field constraint types.
type (
	eqFieldConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	neFieldConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	gtFieldConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	gteFieldConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	ltFieldConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	lteFieldConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	requiredIfConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
		compareValue    string
	}
	requiredUnlessConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
		compareValue    string
	}
	requiredWithConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	requiredWithoutConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	excludedIfConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
		compareValue    string
	}
	excludedUnlessConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
		compareValue    string
	}
	excludedWithConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	excludedWithoutConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
	}
	skipUnlessConstraint struct {
		targetFieldName string     // Keep for error messages
		targetFieldPath *FieldPath // Replace targetFieldIndex
		compareValue    string

		// Precomputed for performance
		isSimplePath     bool                     // true for sibling fields (single-level path)
		directFieldIndex int                      // O(1) access for simple paths
		compareFn        func(reflect.Value) bool // Type-specific comparison
	}

	// requiredWithAllConstraint: field is required if ALL listed fields are present (non-zero).
	requiredWithAllConstraint struct {
		targetFieldNames []string     // Field names for error messages
		targetFieldPaths []*FieldPath // Pre-computed paths for validation
	}

	// requiredWithoutAllConstraint: field is required if ALL listed fields are absent (zero).
	requiredWithoutAllConstraint struct {
		targetFieldNames []string
		targetFieldPaths []*FieldPath
	}

	// excludedWithAllConstraint: field must be absent if ALL listed fields are present.
	excludedWithAllConstraint struct {
		targetFieldNames []string
		targetFieldPaths []*FieldPath
	}

	// excludedWithoutAllConstraint: field must be absent if ALL listed fields are absent.
	excludedWithoutAllConstraint struct {
		targetFieldNames []string
		targetFieldPaths []*FieldPath
	}
)

// BuildCrossFieldConstraintsForField builds cross-field and skip constraint instances from parsed tags.
// Returns both types separately so validator can handle skip constraints before other validations.
func BuildCrossFieldConstraintsForField(constraints map[string]string, structType reflect.Type, fieldIndex int) ([]CrossFieldConstraint, []SkipConstraint) {
	var result []CrossFieldConstraint
	var skipResult []SkipConstraint

	fieldName := structType.Field(fieldIndex).Name

	for name, value := range constraints {
		switch name {
		case "eqfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "eqfield")
			result = append(result, eqFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "nefield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "nefield")
			result = append(result, neFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "gtfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "gtfield")
			result = append(result, gtFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "gtefield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "gtefield")
			result = append(result, gteFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "ltfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "ltfield")
			result = append(result, ltFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "ltefield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "ltefield")
			result = append(result, lteFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "eqcsfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "eqcsfield")
			result = append(result, eqFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "necsfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "necsfield")
			result = append(result, neFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "gtcsfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "gtcsfield")
			result = append(result, gtFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "gtecsfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "gtecsfield")
			result = append(result, gteFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "ltcsfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "ltcsfield")
			result = append(result, ltFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "ltecsfield":
			fp := resolveAndValidateField(structType, value, fieldIndex, fieldName, "ltecsfield")
			result = append(result, lteFieldConstraint{targetFieldName: value, targetFieldPath: fp})
		case "required_if":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				fp := ParseFieldPath(structType, fieldName)
				result = append(result, requiredIfConstraint{targetFieldName: fieldName, targetFieldPath: fp, compareValue: compareValue})
			}
		case "required_unless":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				fp := ParseFieldPath(structType, fieldName)
				result = append(result, requiredUnlessConstraint{targetFieldName: fieldName, targetFieldPath: fp, compareValue: compareValue})
			}
		case "required_with":
			fp := ParseFieldPath(structType, value)
			result = append(result, requiredWithConstraint{targetFieldName: value, targetFieldPath: fp})
		case "required_without":
			fp := ParseFieldPath(structType, value)
			result = append(result, requiredWithoutConstraint{targetFieldName: value, targetFieldPath: fp})
		case "excluded_if":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				fp := ParseFieldPath(structType, fieldName)
				result = append(result, excludedIfConstraint{targetFieldName: fieldName, targetFieldPath: fp, compareValue: compareValue})
			}
		case "excluded_unless":
			if fieldName, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				fp := ParseFieldPath(structType, fieldName)
				result = append(result, excludedUnlessConstraint{targetFieldName: fieldName, targetFieldPath: fp, compareValue: compareValue})
			}
		case "excluded_with":
			fp := ParseFieldPath(structType, value)
			result = append(result, excludedWithConstraint{targetFieldName: value, targetFieldPath: fp})
		case "excluded_without":
			fp := ParseFieldPath(structType, value)
			result = append(result, excludedWithoutConstraint{targetFieldName: value, targetFieldPath: fp})
		case "skip_unless":
			if targetField, compareValue, ok := parseConditionalConstraint(value, " "); ok {
				fp := ParseFieldPath(structType, targetField)

				// Precompute optimizations for performance
				isSimple := len(fp.Parts) == 1
				directIdx := -1
				if isSimple {
					directIdx = fp.IndexAtLevel[0]
				}

				// Create type-specific comparator (precomputed at build time)
				var compareFn func(reflect.Value) bool
				if field, found := structType.FieldByName(targetField); found {
					switch field.Type.Kind() {
					case reflect.String:
						compareFn = func(v reflect.Value) bool { return v.String() == compareValue }
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						expected, _ := strconv.ParseInt(compareValue, 10, 64)
						compareFn = func(v reflect.Value) bool { return v.Int() == expected }
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						expected, _ := strconv.ParseUint(compareValue, 10, 64)
						compareFn = func(v reflect.Value) bool { return v.Uint() == expected }
					case reflect.Bool:
						expected := compareValue == boolStringTrue
						compareFn = func(v reflect.Value) bool { return v.Bool() == expected }
					default:
						// Fallback: string comparison via fmt.Sprint
						compareFn = func(v reflect.Value) bool { return fmt.Sprint(v.Interface()) == compareValue }
					}
				}

				skipResult = append(skipResult, skipUnlessConstraint{
					targetFieldName:  targetField,
					targetFieldPath:  fp,
					compareValue:     compareValue,
					isSimplePath:     isSimple,
					directFieldIndex: directIdx,
					compareFn:        compareFn,
				})
			}
		case "required_with_all":
			names, paths := parseMultiFieldConstraint(structType, value)
			if len(paths) > 0 {
				result = append(result, requiredWithAllConstraint{
					targetFieldNames: names,
					targetFieldPaths: paths,
				})
			}
		case "required_without_all":
			names, paths := parseMultiFieldConstraint(structType, value)
			if len(paths) > 0 {
				result = append(result, requiredWithoutAllConstraint{
					targetFieldNames: names,
					targetFieldPaths: paths,
				})
			}
		case "excluded_with_all":
			names, paths := parseMultiFieldConstraint(structType, value)
			if len(paths) > 0 {
				result = append(result, excludedWithAllConstraint{
					targetFieldNames: names,
					targetFieldPaths: paths,
				})
			}
		case "excluded_without_all":
			names, paths := parseMultiFieldConstraint(structType, value)
			if len(paths) > 0 {
				result = append(result, excludedWithoutAllConstraint{
					targetFieldNames: names,
					targetFieldPaths: paths,
				})
			}
		}
	}

	return result, skipResult
}

// ============================================================================
// Helper Functions for Cross-Field Validation
// ============================================================================

// CheckTypeCompatibility checks if two values can be compared.
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
		if aVal.Kind() == reflect.Pointer || bVal.Kind() == reflect.Pointer {
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

// Dereference removes pointer indirection from a type.
func Dereference(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// IsNumericType checks if a type is numeric.
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
// Compare compares two values.
func Compare(a, b any) int {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Check if both are nil pointers
	aIsNil := aVal.Kind() == reflect.Pointer && aVal.IsNil()
	bIsNil := bVal.Kind() == reflect.Pointer && bVal.IsNil()

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
	if aVal.Kind() == reflect.Pointer {
		aVal = aVal.Elem()
	}
	if bVal.Kind() == reflect.Pointer {
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

// CompareToString converts any value to string for comparison.
func CompareToString(value any) string {
	val := reflect.ValueOf(value)

	// Handle pointer types
	if val.Kind() == reflect.Pointer {
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

// resolveAndValidateField resolves a field path, validates it exists and is not self-referencing, panics on error.
func resolveAndValidateField(structType reflect.Type, targetFieldName string, currentFieldIndex int, currentFieldName, constraintName string) *FieldPath {
	// Use ParseFieldPath which already panics on invalid fields
	fp := ParseFieldPath(structType, targetFieldName)

	// Check for self-reference (only for single-level paths)
	if len(fp.Parts) == 1 && fp.IndexAtLevel[0] == currentFieldIndex {
		panic(fmt.Sprintf("field %s cannot reference itself in %s constraint", currentFieldName, constraintName))
	}

	return fp
}

// parseConditionalConstraint parses "field:value" or "field value" syntax.
// Returns (fieldName, compareValue, true) on success, ("", "", false) on failure.
func parseConditionalConstraint(value, separator string) (fieldName, compareValue string, ok bool) {
	parts := strings.SplitN(value, separator, 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// ============================================================================
// ValidateCrossField methods for *_all constraints
// ============================================================================

// requiredWithAllConstraint: field is required if ALL listed fields are present (non-zero).
func (c requiredWithAllConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	// Check if ALL target fields are present (non-zero)
	allPresent := true
	for _, fp := range c.targetFieldPaths {
		targetValue, err := fp.ResolveValue(structValue)
		if err != nil || IsZeroValue(targetValue) {
			allPresent = false
			break
		}
	}

	if allPresent {
		// All target fields are present - this field must also be present
		if IsZeroValue(fieldValue) {
			return NewConstraintErrorf(CodeRequiredWithAll,
				"is required when all of [%s] are present",
				strings.Join(c.targetFieldNames, ", "))
		}
	}
	return nil
}

// requiredWithoutAllConstraint: field is required if ALL listed fields are absent (zero).
func (c requiredWithoutAllConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	// Check if ALL target fields are absent (zero)
	allAbsent := true
	for _, fp := range c.targetFieldPaths {
		targetValue, err := fp.ResolveValue(structValue)
		if err == nil && !IsZeroValue(targetValue) {
			allAbsent = false
			break
		}
	}

	if allAbsent {
		// All target fields are absent - this field must be present
		if IsZeroValue(fieldValue) {
			return NewConstraintErrorf(CodeRequiredWithoutAll,
				"is required when all of [%s] are absent",
				strings.Join(c.targetFieldNames, ", "))
		}
	}
	return nil
}

// excludedWithAllConstraint: field must be absent if ALL listed fields are present.
func (c excludedWithAllConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	allPresent := true
	for _, fp := range c.targetFieldPaths {
		targetValue, err := fp.ResolveValue(structValue)
		if err != nil || IsZeroValue(targetValue) {
			allPresent = false
			break
		}
	}

	if allPresent {
		if !IsZeroValue(fieldValue) {
			return NewConstraintErrorf(CodeExcludedWithAll,
				"must be absent when all of [%s] are present",
				strings.Join(c.targetFieldNames, ", "))
		}
	}
	return nil
}

// excludedWithoutAllConstraint: field must be absent if ALL listed fields are absent.
func (c excludedWithoutAllConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	allAbsent := true
	for _, fp := range c.targetFieldPaths {
		targetValue, err := fp.ResolveValue(structValue)
		if err == nil && !IsZeroValue(targetValue) {
			allAbsent = false
			break
		}
	}

	if allAbsent {
		if !IsZeroValue(fieldValue) {
			return NewConstraintErrorf(CodeExcludedWithoutAll,
				"must be absent when all of [%s] are absent",
				strings.Join(c.targetFieldNames, ", "))
		}
	}
	return nil
}

// skipUnlessConstraint: field validation is skipped unless another field matches a value.
// Implements SkipConstraint interface for proper skip behavior.

// ShouldSkip returns true if validation should be skipped (condition NOT met).
func (c skipUnlessConstraint) ShouldSkip(structValue reflect.Value) bool {
	// Early return if compareFn not set (stub behavior)
	if c.compareFn == nil {
		return false
	}

	var target reflect.Value
	switch {
	case c.isSimplePath:
		// O(1) direct field access - no loop
		target = structValue.Field(c.directFieldIndex)
	default:
		// Nested path - use full resolution
		val, err := c.targetFieldPath.ResolveValue(structValue)
		if err != nil {
			return true // Skip if target doesn't exist
		}
		target = reflect.ValueOf(val)
	}

	// Condition NOT met → skip; Condition met → don't skip
	return !c.compareFn(target)
}

// ValidateCrossField is kept for backwards compatibility.
// The actual skip logic is in ShouldSkip().
func (c skipUnlessConstraint) ValidateCrossField(fieldValue any, structValue reflect.Value, fieldName string) error {
	// This method is now a no-op - skip logic moved to ShouldSkip()
	return nil
}

// parseMultiFieldConstraint parses space-separated field names and returns paths.
// Example: "FieldA FieldB FieldC" -> (["FieldA", "FieldB", "FieldC"], [path1, path2, path3]).
func parseMultiFieldConstraint(structType reflect.Type, value string) ([]string, []*FieldPath) {
	parts := strings.Fields(value) // Split by whitespace
	if len(parts) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(parts))
	paths := make([]*FieldPath, 0, len(parts))

	for _, fieldName := range parts {
		fieldName = strings.TrimSpace(fieldName)
		if fieldName == "" {
			continue
		}
		fp := ParseFieldPath(structType, fieldName)
		names = append(names, fieldName)
		paths = append(paths, fp)
	}

	return names, paths
}
