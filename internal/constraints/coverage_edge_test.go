package constraints

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShouldSkip_NilCompareFn tests the edge case where compareFn is nil.
// This covers crossfield.go:567 (line 569-571).
func TestShouldSkip_NilCompareFn(t *testing.T) {
	type TestStruct struct {
		Status string
		Value  string
	}

	// Create a skipUnlessConstraint with nil compareFn
	c := skipUnlessConstraint{
		targetFieldName:  "Status",
		targetFieldPath:  nil, // Not needed for this test
		compareValue:     "active",
		isSimplePath:     true,
		directFieldIndex: 0,
		compareFn:        nil, // This is the key: nil compareFn
	}

	structValue := reflect.ValueOf(TestStruct{Status: "active", Value: "test"})

	// When compareFn is nil, ShouldSkip should return false (don't skip)
	result := c.ShouldSkip(structValue)
	assert.False(t, result, "ShouldSkip should return false when compareFn is nil")
}

// TestIsContextValidator_NilLookup tests when ctxValidatorLookup is nil.
// This covers custom.go:33 (lines 34-35).
func TestIsContextValidator_NilLookup(t *testing.T) {
	// Save original lookup
	originalLookup := ctxValidatorLookup
	defer func() {
		ctxValidatorLookup = originalLookup
	}()

	// Set lookup to nil
	ctxValidatorLookup = nil

	// When lookup is nil, IsContextValidator should return false
	result := IsContextValidator("some_validator")
	assert.False(t, result, "IsContextValidator should return false when ctxValidatorLookup is nil")
}

// TestExtractContextValidators_OrPrefix tests that __or__ prefixed constraints are skipped.
// This covers custom.go:42 (lines 46-47).
func TestExtractContextValidators_OrPrefix(t *testing.T) {
	// Setup mock context validator lookup
	mockCtxLookup := func(name string) bool {
		return name == "ctx_validator"
	}

	originalLookup := ctxValidatorLookup
	SetCtxValidatorLookup(mockCtxLookup)
	defer func() {
		ctxValidatorLookup = originalLookup
	}()

	constraintsMap := map[string]string{
		"ctx_validator":      "param1",
		"__or__email|url":    "",        // Should be skipped due to __or__ prefix
		"__or__min=5|max=10": "special", // Should be skipped
		"regular":            "param2",  // Not a context validator, so not extracted
	}

	result := ExtractContextValidators(constraintsMap)

	// Should only extract ctx_validator, not the __or__ prefixed ones
	assert.Len(t, result, 1, "should extract exactly 1 context validator")
	assert.Equal(t, "ctx_validator", result[0].Name)
	assert.Equal(t, "param1", result[0].Param)
}

// TestAppendCollectionConstraint_Default tests the default constraint case.
// This covers constraints.go:498 (line 503).
func TestAppendCollectionConstraint_Default(t *testing.T) {
	result := []Constraint{}

	// Test unique constraint (already covered elsewhere)
	result = appendCollectionConstraint(result, "unique", "ID")
	assert.Len(t, result, 1)

	// Test default constraint (this is the uncovered line 503)
	result = appendCollectionConstraint(result, "default", "test_value")
	assert.Len(t, result, 2)

	// Verify the default constraint was added
	dc, ok := result[1].(defaultConstraint)
	require.True(t, ok, "should be defaultConstraint type")
	assert.Equal(t, "test_value", dc.value)
}

// TestAppendGeoConstraint_Longitude tests the longitude constraint case.
// This covers constraints.go:582 (line 587).
func TestAppendGeoConstraint_Longitude(t *testing.T) {
	result := []Constraint{}

	// Test latitude constraint (already covered)
	result = appendGeoConstraint(result, "latitude")
	assert.Len(t, result, 1)
	_, ok := result[0].(latitudeConstraint)
	assert.True(t, ok, "should be latitudeConstraint")

	// Test longitude constraint (this is the uncovered line 587)
	result = appendGeoConstraint(result, "longitude")
	assert.Len(t, result, 2)
	_, ok = result[1].(longitudeConstraint)
	assert.True(t, ok, "should be longitudeConstraint")
}

// TestIsNumericType_NilType tests IsNumericType with nil input.
// This covers crossfield.go:335 (line 336).
func TestIsNumericType_NilType(t *testing.T) {
	// Pass nil type
	result := IsNumericType(nil)
	assert.False(t, result, "IsNumericType should return false for nil type")

	// Test with valid types for comparison
	assert.True(t, IsNumericType(reflect.TypeOf(int(0))))
	assert.True(t, IsNumericType(reflect.TypeOf(float64(0))))
	assert.False(t, IsNumericType(reflect.TypeOf("")))
}

// TestCompare_DefaultFallback tests the default return case in Compare.
// This covers crossfield.go:351 (line 398).
func TestCompare_DefaultFallback(t *testing.T) {
	// Test with incompatible types that fall through to default case
	type CustomStruct struct {
		Value int
	}

	// Comparing two custom structs should hit the default case (line 398)
	a := CustomStruct{Value: 1}
	b := CustomStruct{Value: 2}

	result := Compare(a, b)
	assert.Equal(t, 0, result, "Compare should return 0 for incomparable types (default case)")

	// Test with struct and string (different types, not numeric or string comparison)
	result = Compare(CustomStruct{Value: 1}, "string")
	assert.Equal(t, 0, result, "Compare should return 0 for incomparable types")

	// Test with bool (not handled in string or numeric switch)
	result = Compare(true, false)
	assert.Equal(t, 0, result, "Compare should return 0 for bool types (not in numeric/string switch)")
}

// TestCompare_NumericDefaultBranch tests the default branch in numeric comparison.
// This ensures line 398 and 409 are covered.
func TestCompare_NumericDefaultBranch(t *testing.T) {
	// Create a complex type that passes initial checks but fails numeric extraction
	type ComplexType struct {
		Data []byte
	}

	a := ComplexType{Data: []byte{1, 2, 3}}
	b := ComplexType{Data: []byte{4, 5, 6}}

	// This should hit the default case in the switch statements (lines 398, 409)
	result := Compare(a, b)
	assert.Equal(t, 0, result, "Compare should return 0 for non-comparable complex types")
}

// TestBuildConstraints_WithUniqueAndDefault tests collection constraints are built correctly.
func TestBuildConstraints_WithUniqueAndDefault(t *testing.T) {
	constraintsMap := map[string]string{
		"unique":  "ID",
		"default": "default_value",
	}

	result := BuildConstraints(constraintsMap, nil)

	// Should have both unique and default constraints
	assert.Len(t, result, 2, "should build both unique and default constraints")

	// Check that we have one of each type
	var hasUnique, hasDefault bool
	for _, c := range result {
		switch c.(type) {
		case uniqueConstraint:
			hasUnique = true
		case defaultConstraint:
			hasDefault = true
		}
	}

	assert.True(t, hasUnique, "should have unique constraint")
	assert.True(t, hasDefault, "should have default constraint")
}

// TestBuildConstraints_WithGeoConstraints tests geo constraints are built correctly.
func TestBuildConstraints_WithGeoConstraints(t *testing.T) {
	constraintsMap := map[string]string{
		"latitude":  "",
		"longitude": "",
	}

	result := BuildConstraints(constraintsMap, nil)

	// Should have both latitude and longitude constraints
	assert.Len(t, result, 2, "should build both latitude and longitude constraints")

	// Check that we have one of each type
	var hasLatitude, hasLongitude bool
	for _, c := range result {
		switch c.(type) {
		case latitudeConstraint:
			hasLatitude = true
		case longitudeConstraint:
			hasLongitude = true
		}
	}

	assert.True(t, hasLatitude, "should have latitude constraint")
	assert.True(t, hasLongitude, "should have longitude constraint")
}

// TestExtractContextValidators_MultipleContextValidators tests extracting multiple context validators.
func TestExtractContextValidators_MultipleContextValidators(t *testing.T) {
	// Setup mock that recognizes multiple validators as context-aware
	mockCtxLookup := func(name string) bool {
		return name == "ctx_val1" || name == "ctx_val2" || name == "ctx_val3"
	}

	originalLookup := ctxValidatorLookup
	SetCtxValidatorLookup(mockCtxLookup)
	defer func() {
		ctxValidatorLookup = originalLookup
	}()

	constraintsMap := map[string]string{
		"ctx_val1":            "param1",
		"ctx_val2":            "param2",
		"__or__skip_this":     "",
		"ctx_val3":            "param3",
		"regular_constraint":  "value",
		"__or__skip_this_too": "value",
	}

	result := ExtractContextValidators(constraintsMap)

	// Should extract exactly 3 context validators, skipping __or__ prefixed ones
	assert.Len(t, result, 3, "should extract exactly 3 context validators")

	// Verify all three are present
	names := make(map[string]string)
	for _, cv := range result {
		names[cv.Name] = cv.Param
	}

	assert.Equal(t, "param1", names["ctx_val1"])
	assert.Equal(t, "param2", names["ctx_val2"])
	assert.Equal(t, "param3", names["ctx_val3"])
}

// TestSetCtxValidatorLookup tests setting the context validator lookup function.
func TestSetCtxValidatorLookup(t *testing.T) {
	// Save original
	originalLookup := ctxValidatorLookup
	defer func() {
		ctxValidatorLookup = originalLookup
	}()

	// Create a test lookup
	testLookup := func(name string) bool {
		return name == "test_ctx_validator"
	}

	// Set the lookup
	SetCtxValidatorLookup(testLookup)

	// Verify it was set by testing IsContextValidator
	assert.True(t, IsContextValidator("test_ctx_validator"))
	assert.False(t, IsContextValidator("other_validator"))
}
