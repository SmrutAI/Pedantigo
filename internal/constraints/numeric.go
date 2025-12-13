// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Numeric constraint types.
type (
	minConstraint            struct{ min int }
	maxConstraint            struct{ max int }
	minLengthConstraint      struct{ minLength int }
	maxLengthConstraint      struct{ maxLength int }
	gtConstraint             struct{ threshold float64 }
	geConstraint             struct{ threshold float64 }
	ltConstraint             struct{ threshold float64 }
	leConstraint             struct{ threshold float64 }
	positiveConstraint       struct{}
	negativeConstraint       struct{}
	multipleOfConstraint     struct{ factor float64 }
	maxDigitsConstraint      struct{ maxDigits int }
	decimalPlacesConstraint  struct{ maxPlaces int }
	disallowInfNanConstraint struct{}
)

// boundMode distinguishes between min (lower bound) and max (upper bound) checks.
type boundMode int

const (
	boundMin boundMode = iota
	boundMax
)

// validateBound is a helper that validates numeric bounds (min or max).
// For min: value must be >= bound. For max: value must be <= bound.
func validateBound(value any, bound int, mode boundMode) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	var failed bool
	var constraintName string

	switch mode {
	case boundMin:
		constraintName = "min"
		failed = checkMinViolation(v, bound)
	case boundMax:
		constraintName = "max"
		failed = checkMaxViolation(v, bound)
	}

	if !failed {
		return nil
	}
	return formatBoundError(v.Kind(), bound, mode, constraintName)
}

func checkMinViolation(v reflect.Value, bound int) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() < int64(bound)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return bound >= 0 && v.Uint() < uint64(bound) //nolint:gosec // bounds checked
	case reflect.Float32, reflect.Float64:
		return v.Float() < float64(bound)
	case reflect.String:
		return len(v.String()) < bound
	}
	return true // unsupported type is a violation
}

func checkMaxViolation(v reflect.Value, bound int) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() > int64(bound)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return bound >= 0 && v.Uint() > uint64(bound) //nolint:gosec // bounds checked
	case reflect.Float32, reflect.Float64:
		return v.Float() > float64(bound)
	case reflect.String:
		return len(v.String()) > bound
	}
	return true // unsupported type is a violation
}

func formatBoundError(kind reflect.Kind, bound int, mode boundMode, constraintName string) error {
	msgWord := "at least"
	if mode == boundMax {
		msgWord = "at most"
	}
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Errorf("must be %s %d", msgWord, bound)
	case reflect.String:
		return fmt.Errorf("must be %s %d characters", msgWord, bound)
	default:
		return fmt.Errorf("%s constraint not supported for type %s", constraintName, kind)
	}
}

// minConstraint validates that a numeric value is >= min.
func (c minConstraint) Validate(value any) error { return validateBound(value, c.min, boundMin) }

// maxConstraint validates that a numeric value is <= max.
func (c maxConstraint) Validate(value any) error { return validateBound(value, c.max, boundMax) }

// minLengthConstraint validates that a string has at least minLength characters.
func (c minLengthConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	if v.Kind() != reflect.String {
		return fmt.Errorf("min_length constraint requires string value")
	}

	if len(v.String()) < c.minLength {
		return fmt.Errorf("must be at least %d characters", c.minLength)
	}

	return nil
}

// maxLengthConstraint validates that a string has at most maxLength characters.
func (c maxLengthConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	if v.Kind() != reflect.String {
		return fmt.Errorf("max_length constraint requires string value")
	}

	if len(v.String()) > c.maxLength {
		return fmt.Errorf("must be at most %d characters", c.maxLength)
	}

	return nil
}

// gtConstraint validates that a numeric value is > threshold.
func (c gtConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("gt constraint requires numeric value")
	}

	if numValue <= c.threshold {
		return fmt.Errorf("must be greater than %v", c.threshold)
	}

	return nil
}

// geConstraint validates that a numeric value is >= threshold.
func (c geConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("ge constraint requires numeric value")
	}

	if numValue < c.threshold {
		return fmt.Errorf("must be at least %v", c.threshold)
	}

	return nil
}

// ltConstraint validates that a numeric value is < threshold.
func (c ltConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("lt constraint requires numeric value")
	}

	if numValue >= c.threshold {
		return fmt.Errorf("must be less than %v", c.threshold)
	}

	return nil
}

// leConstraint validates that a numeric value is <= threshold.
func (c leConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("le constraint requires numeric value")
	}

	if numValue > c.threshold {
		return fmt.Errorf("must be at most %v", c.threshold)
	}

	return nil
}

// positiveConstraint validates that a numeric value is greater than 0.
func (c positiveConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("positive constraint requires numeric value")
	}

	if numValue <= 0 {
		return fmt.Errorf("must be positive (greater than 0)")
	}

	return nil
}

// negativeConstraint validates that a numeric value is less than 0.
func (c negativeConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("negative constraint requires numeric value")
	}

	if numValue >= 0 {
		return fmt.Errorf("must be negative (less than 0)")
	}

	return nil
}

// multipleOfConstraint validates that a numeric value is divisible by factor.
func (c multipleOfConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	numValue, err := extractNumericValue(v)
	if err != nil {
		return fmt.Errorf("multiple_of constraint requires numeric value")
	}

	// Check if value is divisible by factor
	remainder := math.Mod(numValue, c.factor)
	// Use small epsilon for floating point comparison
	if math.Abs(remainder) > 1e-9 && math.Abs(remainder-c.factor) > 1e-9 {
		return fmt.Errorf("must be a multiple of %v", c.factor)
	}

	return nil
}

// maxDigitsConstraint validates that a numeric value has at most maxDigits digits.
func (c maxDigitsConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	// Get numeric value as string
	var str string
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Errorf("max_digits constraint requires numeric value")
	}

	// Count digits (exclude minus sign and decimal point)
	digitCount := 0
	for _, r := range str {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}

	if digitCount > c.maxDigits {
		return fmt.Errorf("must have at most %d digits", c.maxDigits)
	}

	return nil
}

// decimalPlacesConstraint validates that a numeric value has at most maxPlaces decimal places.
func (c decimalPlacesConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	// Get numeric value as string
	var str string
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Integers have no decimal places
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Unsigned integers have no decimal places
		return nil
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Errorf("decimal_places constraint requires numeric value")
	}

	// Find decimal point and count places
	decimalPlaces := 0
	if idx := strings.Index(str, "."); idx >= 0 {
		decimalPlaces = len(str) - idx - 1
	}

	if decimalPlaces > c.maxPlaces {
		return fmt.Errorf("must have at most %d decimal places", c.maxPlaces)
	}

	return nil
}

// disallowInfNanConstraint validates that a float is not Inf or NaN.
// Note: Named "disallow" (not "allow") because Go defaults to allowing Inf/NaN.
// Pydantic uses allow_inf_nan=True by default; we use disallow_inf_nan as opt-in rejection.
func (c disallowInfNanConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	// Only check float types - integers cannot be Inf/NaN
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsInf(f, 0) {
			return fmt.Errorf("infinity is not allowed")
		}
		if math.IsNaN(f) {
			return fmt.Errorf("NaN is not allowed")
		}
	}
	// Non-float types pass (integers can't be Inf/NaN)
	return nil
}

// buildMinConstraint creates a min constraint, handling context-aware type checking.
// Returns (constraint, true) on success or (nil, false) if parsing fails.
func buildMinConstraint(value string, fieldType reflect.Type) (Constraint, bool) {
	minVal, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}

	// Handle pointer types - check underlying type
	checkType := fieldType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	kind := checkType.Kind()
	if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
		return minLengthConstraint{minLength: minVal}, true
	}
	return minConstraint{min: minVal}, true
}

// buildMaxConstraint creates a max constraint, handling context-aware type checking.
// Returns (constraint, true) on success or (nil, false) if parsing fails.
func buildMaxConstraint(value string, fieldType reflect.Type) (Constraint, bool) {
	maxVal, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}

	// Handle pointer types - check underlying type
	checkType := fieldType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	kind := checkType.Kind()
	if kind == reflect.String || kind == reflect.Slice || kind == reflect.Array {
		return maxLengthConstraint{maxLength: maxVal}, true
	}
	return maxConstraint{max: maxVal}, true
}

// buildMultipleOfConstraint creates a multiple_of constraint with the specified factor.
func buildMultipleOfConstraint(value string) (Constraint, bool) {
	factor, err := strconv.ParseFloat(value, 64)
	if err != nil || factor == 0 {
		return nil, false // Invalid or zero factor
	}
	return multipleOfConstraint{factor: factor}, true
}

// buildMaxDigitsConstraint creates a max_digits constraint with the specified maximum.
func buildMaxDigitsConstraint(value string) (Constraint, bool) {
	maxDigits, err := strconv.Atoi(value)
	if err != nil || maxDigits <= 0 {
		return nil, false // Invalid or non-positive max digits
	}
	return maxDigitsConstraint{maxDigits: maxDigits}, true
}

// buildDecimalPlacesConstraint creates a decimal_places constraint with the specified maximum.
func buildDecimalPlacesConstraint(value string) (Constraint, bool) {
	maxPlaces, err := strconv.Atoi(value)
	if err != nil || maxPlaces < 0 {
		return nil, false // Invalid or negative max places
	}
	return decimalPlacesConstraint{maxPlaces: maxPlaces}, true
}
