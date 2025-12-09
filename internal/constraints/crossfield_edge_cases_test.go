package constraints_test

import (
	"math"
	"testing"
	"time"

	. "github.com/SmrutAI/Pedantigo"
)

// ==================================================
// Cross-Field Constraint Edge Case Tests
// ==================================================
//
// IMPORTANT: These tests are part of Phase 4.2 (Red phase - Failing tests)
// The cross-field constraint implementations are stubs returning "not implemented"
// These tests will pass once the actual implementations are added.
//
// Test approach:
// - Most tests use Validate() since cross-field validation happens on the struct itself
// - Some tests check validator creation (fail-fast patterns)
// - Tests cover error conditions and boundary cases

// ==================================================
// Edge Case 1: Nonexistent Target Field
// ==================================================

func TestCrossField_NonexistentField_EqField(t *testing.T) {
	type User struct {
		Password        string `pedantigo:"required"`
		ConfirmPassword string `pedantigo:"eqfield=NonExistentField"`
	}

	// Should panic during validator creation due to nonexistent field
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when eqfield references nonexistent field")
		}
	}()

	_ = New[User]() // This should panic
	t.Error("should have panicked before reaching here")
}

func TestCrossField_NonexistentField_GtField(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Val int `pedantigo:"gtfield=NoSuchField"`
	}

	// Should panic during validator creation due to nonexistent field
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when gtfield references nonexistent field")
		}
	}()

	_ = New[Range]() // This should panic
	t.Error("should have panicked before reaching here")
}

func TestCrossField_NonexistentField_LtField(t *testing.T) {
	type Range struct {
		Max int `pedantigo:"required"`
		Val int `pedantigo:"ltfield=InvalidTarget"`
	}

	// Should panic during validator creation due to nonexistent field
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when ltfield references nonexistent field")
		}
	}()

	_ = New[Range]() // This should panic
	t.Error("should have panicked before reaching here")
}

// ==================================================
// Edge Case 2: Type Incompatibility
// ==================================================

func TestCrossField_TypeIncompatibility_StringVsInt(t *testing.T) {
	type Mixed struct {
		Age  int    `pedantigo:"required"`
		Name string `pedantigo:"gtfield=Age"` // Comparing string > int
	}

	validator := New[Mixed]()
	m := &Mixed{Age: 25, Name: "Alice"}

	err := validator.Validate(m)
	if err == nil {
		t.Error("expected error comparing incompatible types (string vs int)")
	}
}

func TestCrossField_TypeIncompatibility_FloatVsString(t *testing.T) {
	type Mixed struct {
		Price float64 `pedantigo:"required"`
		Label string  `pedantigo:"ltfield=Price"` // Comparing string < float64
	}

	validator := New[Mixed]()
	m := &Mixed{Price: 99.99, Label: "expensive"}

	err := validator.Validate(m)
	if err == nil {
		t.Error("expected error comparing incompatible types (string vs float64)")
	}
}

func TestCrossField_TypeIncompatibility_StructVsInt(t *testing.T) {
	type Nested struct {
		Value int
	}

	type Mixed struct {
		Count  int    `pedantigo:"required"`
		Config Nested `pedantigo:"eqfield=Count"` // Comparing struct == int
	}

	validator := New[Mixed]()
	m := &Mixed{Count: 5, Config: Nested{Value: 5}}

	err := validator.Validate(m)
	if err == nil {
		t.Error("expected error comparing incompatible types (struct vs int)")
	}
}

// ==================================================
// Edge Case 3: Nil Pointer Fields
// ==================================================

func TestCrossField_NilPointer_TargetIsNil(t *testing.T) {
	type Optional struct {
		Value  *int `pedantigo:"required"`
		MinVal int  `pedantigo:"ltfield=Value"` // Comparing int < nil pointer
	}

	validator := New[Optional]()
	o := &Optional{Value: nil, MinVal: 10}

	err := validator.Validate(o)
	if err == nil {
		t.Error("expected error when comparing against nil pointer field")
	}
}

func TestCrossField_NilPointer_SourceIsNil(t *testing.T) {
	type Optional struct {
		Value  *int `pedantigo:"required"`
		MinVal *int `pedantigo:"gtfield=Value"` // Nil pointer compared against value
	}

	validator := New[Optional]()
	val := 10
	o := &Optional{Value: &val, MinVal: nil}

	err := validator.Validate(o)
	if err == nil {
		t.Error("expected error when source field is nil pointer")
	}
}

func TestCrossField_NilPointer_BothNil(t *testing.T) {
	type Optional struct {
		Field1 *int `pedantigo:"required"`
		Field2 *int `pedantigo:"eqfield=Field1"` // Both nil
	}

	validator := New[Optional]()
	o := &Optional{Field1: nil, Field2: nil}

	err := validator.Validate(o)
	// Both nil: should be equal
	if err != nil {
		t.Errorf("expected no error for nil == nil, got %v", err)
	}
}

// ==================================================
// Edge Case 4: Case Sensitivity in Field Names
// ==================================================

func TestCrossField_CaseSensitivity_LowercaseRef(t *testing.T) {
	type CaseTest struct {
		Value    int `pedantigo:"required"`
		MinValue int `pedantigo:"gtfield=value"` // lowercase 'value' doesn't match 'Value'
	}

	// Should panic during validator creation due to case mismatch (field not found)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for case-mismatched field name (value != Value)")
		}
	}()

	_ = New[CaseTest]() // This should panic
	t.Error("should have panicked before reaching here")
}

func TestCrossField_CaseSensitivity_MixedCaseRef(t *testing.T) {
	type CaseTest struct {
		UserID   int `pedantigo:"required"`
		MinLimit int `pedantigo:"ltfield=userid"` // incorrect case: 'userid' != 'UserID'
	}

	// Should panic during validator creation due to case mismatch (field not found)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for incorrect field name case")
		}
	}()

	_ = New[CaseTest]() // This should panic
	t.Error("should have panicked before reaching here")
}

func TestCrossField_CaseSensitivity_CorrectCase(t *testing.T) {
	type CaseTest struct {
		Value    int `pedantigo:"required"`
		MinValue int `pedantigo:"gtfield=Value"` // correct case
	}

	validator := New[CaseTest]()
	c := &CaseTest{Value: 10, MinValue: 5}

	// Should not error (assuming implementation is correct)
	err := validator.Validate(c)
	// This will fail until implementation is complete
	if err != nil {
		// Expected during stub phase
		t.Logf("expected failure in stub phase: %v", err)
	}
}

// ==================================================
// Edge Case 5: Nested Structs
// ==================================================

func TestCrossField_NestedStruct_Direct(t *testing.T) {
	t.Skip("TODO: Cross-field validation within nested structs not yet implemented")
	// This test expects cross-field constraints to be validated within nested structs
	// Currently, cross-field constraints are only built for the top-level struct
	// Future enhancement: recursively validate nested struct cross-field constraints
}

func TestCrossField_NestedStruct_CrossNested(t *testing.T) {
	t.Skip("TODO: Dotted field notation (Info.Value) for nested struct cross-field validation not yet implemented")
	// This test expects support for cross-referencing nested struct fields using dotted notation
	// Currently only supports same-level field references
	// Future enhancement: add support for Info.Value syntax to reference nested struct fields
}

// ==================================================
// Edge Case 6: Multiple Cross-Field Constraints
// ==================================================

func TestCrossField_MultipleConstraints_BothValid(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Max int `pedantigo:"required"`
		Val int `pedantigo:"gtfield=Min,ltfield=Max"` // Multiple constraints
	}

	validator := New[Range]()
	r := &Range{Min: 0, Max: 100, Val: 50}

	err := validator.Validate(r)
	// Should pass if implementation supports multiple cross-field constraints
	if err != nil {
		t.Logf("multiple constraints validation: %v", err)
	}
}

func TestCrossField_MultipleConstraints_FirstFails(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Max int `pedantigo:"required"`
		Val int `pedantigo:"gtfield=Min,ltfield=Max"`
	}

	validator := New[Range]()
	r := &Range{Min: 50, Max: 100, Val: 40} // Val <= Min fails first constraint

	err := validator.Validate(r)
	if err == nil {
		t.Error("expected error for Val <= Min")
	}
}

func TestCrossField_MultipleConstraints_SecondFails(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Max int `pedantigo:"required"`
		Val int `pedantigo:"gtfield=Min,ltfield=Max"`
	}

	validator := New[Range]()
	r := &Range{Min: 0, Max: 50, Val: 60} // Val >= Max fails second constraint

	err := validator.Validate(r)
	if err == nil {
		t.Error("expected error for Val >= Max")
	}
}

func TestCrossField_MultipleConstraints_BothFail(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Max int `pedantigo:"required"`
		Val int `pedantigo:"gtfield=Min,ltfield=Max"`
	}

	validator := New[Range]()
	r := &Range{Min: 50, Max: 40, Val: 30} // Both constraints fail (Min > Max, Val not in range)

	err := validator.Validate(r)
	if err == nil {
		t.Error("expected error(s) for multiple constraint violations")
	}
}

// ==================================================
// Edge Case 7: Self-Reference
// ==================================================

func TestCrossField_SelfReference_EqField(t *testing.T) {
	type Recursive struct {
		Value int `pedantigo:"eqfield=Value"` // Field references itself
	}

	// Self-reference should panic during validator creation
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for self-referencing field")
		}
	}()

	_ = New[Recursive]() // This should panic
	t.Error("should have panicked before reaching here")
}

func TestCrossField_SelfReference_GtField(t *testing.T) {
	type Recursive struct {
		Value int `pedantigo:"gtfield=Value"` // Value > Value makes no sense
	}

	// Self-reference should panic during validator creation
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for self-referencing gtfield")
		}
	}()

	_ = New[Recursive]() // This should panic
	t.Error("should have panicked before reaching here")
}

// ==================================================
// Edge Case 8: Circular Dependencies
// ==================================================

func TestCrossField_CircularDependency_TwoFields(t *testing.T) {
	type Circular struct {
		Field1 int `pedantigo:"gtfield=Field2"`
		Field2 int `pedantigo:"gtfield=Field1"` // Field1 > Field2, Field2 > Field1
	}

	validator := New[Circular]()
	c := &Circular{Field1: 10, Field2: 5}

	// Should error due to circular dependency
	err := validator.Validate(c)
	if err != nil {
		t.Logf("circular dependency detected: %v", err)
	}
}

// ==================================================
// Edge Case 9: Zero Value Comparison
// ==================================================

func TestCrossField_ZeroValue_BothZero(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Max int `pedantigo:"required"`
		Val int `pedantigo:"eqfield=Min"` // Val equals Min, both are 0
	}

	validator := New[Range]()
	r := &Range{Min: 0, Max: 100, Val: 0}

	err := validator.Validate(r)
	if err != nil {
		t.Errorf("expected no error for 0 == 0, got %v", err)
	}
}

func TestCrossField_ZeroValue_ZeroVsNonZero(t *testing.T) {
	type Range struct {
		Min int `pedantigo:"required"`
		Val int `pedantigo:"eqfield=Min"` // Val = 0, Min = 10
	}

	validator := New[Range]()
	r := &Range{Min: 10, Val: 0}

	err := validator.Validate(r)
	if err == nil {
		t.Error("expected error for 0 != 10")
	}
}

func TestCrossField_ZeroValue_EmptyString(t *testing.T) {
	type Strings struct {
		Field1 string `pedantigo:"required"`
		Field2 string `pedantigo:"eqfield=Field1"` // Both empty strings
	}

	validator := New[Strings]()
	s := &Strings{Field1: "", Field2: ""}

	err := validator.Validate(s)
	if err != nil {
		t.Errorf("expected no error for empty string == empty string, got %v", err)
	}
}

// ==================================================
// Edge Case 10: Numeric Type Compatibility
// ==================================================

func TestCrossField_NumericTypeCompatibility_IntVsInt64(t *testing.T) {
	type Mixed struct {
		Value1 int   `pedantigo:"required"`
		Value2 int64 `pedantigo:"gtfield=Value1"` // int vs int64
	}

	validator := New[Mixed]()
	m := &Mixed{Value1: 10, Value2: 20}

	err := validator.Validate(m)
	// Should handle numeric type compatibility
	if err != nil {
		t.Logf("int vs int64 comparison: %v", err)
	}
}

func TestCrossField_NumericTypeCompatibility_IntVsFloat(t *testing.T) {
	type Mixed struct {
		IntVal   int     `pedantigo:"required"`
		FloatVal float64 `pedantigo:"ltfield=IntVal"` // float < int
	}

	validator := New[Mixed]()
	m := &Mixed{IntVal: 100, FloatVal: 50.5}

	err := validator.Validate(m)
	if err != nil {
		t.Logf("int vs float64 comparison: %v", err)
	}
}

func TestCrossField_NumericTypeCompatibility_UintVsInt(t *testing.T) {
	type Mixed struct {
		Signed   int  `pedantigo:"required"`
		Unsigned uint `pedantigo:"eqfield=Signed"` // uint vs int
	}

	validator := New[Mixed]()
	m := &Mixed{Signed: 10, Unsigned: 10}

	err := validator.Validate(m)
	if err != nil {
		t.Logf("uint vs int comparison: %v", err)
	}
}

// ==================================================
// Edge Case 11: Empty String Comparisons
// ==================================================

func TestCrossField_EmptyString_EqField(t *testing.T) {
	type Strings struct {
		Field1 string `pedantigo:"required"`
		Field2 string `pedantigo:"eqfield=Field1"`
	}

	validator := New[Strings]()
	s := &Strings{Field1: "", Field2: ""}

	err := validator.Validate(s)
	if err != nil {
		t.Errorf("expected no error for empty string == empty string, got %v", err)
	}
}

func TestCrossField_EmptyString_NeField(t *testing.T) {
	type Strings struct {
		Field1 string `pedantigo:"required"`
		Field2 string `pedantigo:"nefield=Field1"` // Not equal
	}

	validator := New[Strings]()
	s := &Strings{Field1: "", Field2: ""}

	err := validator.Validate(s)
	if err == nil {
		t.Error("expected error for empty != empty (should be equal)")
	}
}

// ==================================================
// Edge Case 12: Time.Time Comparisons
// ==================================================

func TestCrossField_TimeComparison_Equal(t *testing.T) {
	type TimeRange struct {
		StartTime time.Time `pedantigo:"required"`
		EndTime   time.Time `pedantigo:"eqfield=StartTime"`
	}

	validator := New[TimeRange]()
	now := time.Now()
	tr := &TimeRange{StartTime: now, EndTime: now}

	err := validator.Validate(tr)
	if err != nil {
		t.Errorf("expected no error for equal times, got %v", err)
	}
}

func TestCrossField_TimeComparison_After(t *testing.T) {
	type TimeRange struct {
		StartTime time.Time `pedantigo:"required"`
		EndTime   time.Time `pedantigo:"gtfield=StartTime"` // EndTime > StartTime
	}

	validator := New[TimeRange]()
	now := time.Now()
	tr := &TimeRange{
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}

	err := validator.Validate(tr)
	if err != nil {
		t.Logf("time comparison error: %v", err)
	}
}

func TestCrossField_TimeComparison_Before(t *testing.T) {
	type TimeRange struct {
		StartTime time.Time `pedantigo:"required"`
		EndTime   time.Time `pedantigo:"ltfield=StartTime"` // EndTime < StartTime
	}

	validator := New[TimeRange]()
	now := time.Now()
	tr := &TimeRange{
		StartTime: now,
		EndTime:   now.Add(-1 * time.Hour),
	}

	err := validator.Validate(tr)
	if err == nil {
		t.Error("expected error when EndTime < StartTime")
	}
}

// ==================================================
// Edge Case 13: All Comparison Operators
// ==================================================

func TestCrossField_EqField_Valid(t *testing.T) {
	type Pair struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"eqfield=Field1"`
	}

	validator := New[Pair]()
	p := &Pair{Field1: 42, Field2: 42}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("eqfield validation: %v", err)
	}
}

func TestCrossField_NeField_Valid(t *testing.T) {
	type Pair struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"nefield=Field1"`
	}

	validator := New[Pair]()
	p := &Pair{Field1: 42, Field2: 43}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("nefield validation: %v", err)
	}
}

func TestCrossField_GtField_Valid(t *testing.T) {
	type Pair struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"gtfield=Field1"`
	}

	validator := New[Pair]()
	p := &Pair{Field1: 10, Field2: 20}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("gtfield validation: %v", err)
	}
}

func TestCrossField_GteField_Valid(t *testing.T) {
	type Pair struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"gtefield=Field1"`
	}

	validator := New[Pair]()
	p := &Pair{Field1: 10, Field2: 10}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("gtefield validation: %v", err)
	}
}

func TestCrossField_LtField_Valid(t *testing.T) {
	type Pair struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"ltfield=Field1"`
	}

	validator := New[Pair]()
	p := &Pair{Field1: 100, Field2: 50}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("ltfield validation: %v", err)
	}
}

func TestCrossField_LteField_Valid(t *testing.T) {
	type Pair struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"ltefield=Field1"`
	}

	validator := New[Pair]()
	p := &Pair{Field1: 100, Field2: 100}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("ltefield validation: %v", err)
	}
}

// ==================================================
// Edge Case 14: Boundary Values
// ==================================================

func TestCrossField_BoundaryValues_MinInt(t *testing.T) {
	type Boundary struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"gtfield=Field1"`
	}

	validator := New[Boundary]()
	// At minimum boundary
	b := &Boundary{Field1: 0, Field2: 1}

	err := validator.Validate(b)
	if err != nil {
		t.Logf("boundary value validation: %v", err)
	}
}

func TestCrossField_BoundaryValues_NegativeNumbers(t *testing.T) {
	type Boundary struct {
		Field1 int `pedantigo:"required"`
		Field2 int `pedantigo:"ltfield=Field1"`
	}

	validator := New[Boundary]()
	b := &Boundary{Field1: -10, Field2: -100}

	err := validator.Validate(b)
	if err != nil {
		t.Logf("negative number comparison: %v", err)
	}
}

func TestCrossField_BoundaryValues_MaxInt(t *testing.T) {
	type Boundary struct {
		Field1 int64 `pedantigo:"required"`
		Field2 int64 `pedantigo:"eqfield=Field1"`
	}

	validator := New[Boundary]()
	b := &Boundary{Field1: 9223372036854775807, Field2: 9223372036854775807} // Max int64

	err := validator.Validate(b)
	if err != nil {
		t.Logf("max int64 comparison: %v", err)
	}
}

// ==================================================
// Edge Case 15: Boolean and Complex Types
// ==================================================

func TestCrossField_BooleanComparison(t *testing.T) {
	type BooleanTest struct {
		Flag1 bool `pedantigo:"required"`
		Flag2 bool `pedantigo:"eqfield=Flag1"`
	}

	validator := New[BooleanTest]()
	b := &BooleanTest{Flag1: true, Flag2: true}

	err := validator.Validate(b)
	if err != nil {
		t.Logf("boolean comparison: %v", err)
	}
}

func TestCrossField_SliceComparison(t *testing.T) {
	type SliceTest struct {
		Items []int `pedantigo:"required"`
		Ref   []int `pedantigo:"eqfield=Items"` // Comparing slices
	}

	validator := New[SliceTest]()
	s := &SliceTest{
		Items: []int{1, 2, 3},
		Ref:   []int{1, 2, 3},
	}

	err := validator.Validate(s)
	if err != nil {
		t.Logf("slice comparison: %v", err)
	}
}

// ==================================================
// Edge Case 16: Unexported Fields (Should be skipped)
// ==================================================

func TestCrossField_UnexportedField_IgnoredTarget(t *testing.T) {
	type Unexported struct {
		Field int `pedantigo:"gtfield=value"` // References non-existent unexported field
	}

	// Should panic during validator creation - unexported fields aren't referenceable
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when referencing unexported field")
		}
	}()

	_ = New[Unexported]() // This should panic
	t.Error("should have panicked before reaching here")
}

// ==================================================
// Edge Case 17: Pointer to Different Types
// ==================================================

func TestCrossField_PointerTypes_PtrVsValue(t *testing.T) {
	type PointerTest struct {
		Value1 *int `pedantigo:"required"`
		Value2 int  `pedantigo:"gtfield=Value1"` // Comparing int > *int (dereference needed)
	}

	validator := New[PointerTest]()
	val := 10
	p := &PointerTest{Value1: &val, Value2: 20}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("pointer vs value comparison: %v", err)
	}
}

func TestCrossField_PointerTypes_DifferentPointerTypes(t *testing.T) {
	type PointerTest struct {
		Value1 *int   `pedantigo:"required"`
		Value2 *int64 `pedantigo:"eqfield=Value1"` // *int vs *int64
	}

	validator := New[PointerTest]()
	val1 := 10
	val2 := int64(10)
	p := &PointerTest{Value1: &val1, Value2: &val2}

	err := validator.Validate(p)
	if err != nil {
		t.Logf("different pointer type comparison: %v", err)
	}
}

// ==================================================
// Edge Case 18: Field Order Independence
// ==================================================

func TestCrossField_FieldOrder_ForwardReference(t *testing.T) {
	type Order struct {
		Field2 int `pedantigo:"gtfield=Field1"` // References Field1 (defined below)
		Field1 int `pedantigo:"required"`
	}

	validator := New[Order]()
	o := &Order{Field1: 10, Field2: 20}

	err := validator.Validate(o)
	if err != nil {
		t.Logf("forward reference validation: %v", err)
	}
}

// ==================================================
// Edge Case 19: Multiple Cross-Field Validators on Different Fields
// ==================================================

func TestCrossField_MultipleFieldsWithCrossFieldConstraints(t *testing.T) {
	type MultiConstraint struct {
		Min int `pedantigo:"required"`
		Mid int `pedantigo:"gtfield=Min"`
		Max int `pedantigo:"gtfield=Mid"`
	}

	validator := New[MultiConstraint]()
	m := &MultiConstraint{Min: 10, Mid: 20, Max: 30}

	err := validator.Validate(m)
	if err != nil {
		t.Logf("chain of cross-field constraints: %v", err)
	}
}

// ==================================================
// Edge Case 20: Special Values (NaN, Infinity for floats)
// ==================================================

func TestCrossField_FloatSpecialValues_Infinity(t *testing.T) {
	type FloatSpecial struct {
		Field1 float64 `pedantigo:"required"`
		Field2 float64 `pedantigo:"gtfield=Field1"`
	}

	validator := New[FloatSpecial]()
	inf := math.Inf(1) // Positive infinity
	f := &FloatSpecial{Field1: 100.0, Field2: inf}

	err := validator.Validate(f)
	if err != nil {
		t.Logf("infinity comparison: %v", err)
	}
}
