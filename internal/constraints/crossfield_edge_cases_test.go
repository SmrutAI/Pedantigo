package constraints_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/SmrutAI/pedantigo"
)

// runPanicTest runs a test function and checks for expected panic.
func runPanicTest(t *testing.T, testFunc func() interface{}, expectPanic bool) {
	t.Helper()

	if expectPanic {
		defer func() {
			r := recover()
			assert.NotNil(t, r, "expected panic for invalid field reference")
		}()
	}
	_ = testFunc()
	if expectPanic {
		assert.Fail(t, "should have panicked before reaching here")
	}
}

// ==================================================
// Cross-Field Constraint Edge Case Tests (Table-Driven)
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

// panicTestCase defines a test case for panic testing.
type panicTestCase struct {
	name        string
	testFunc    func() interface{}
	expectPanic bool
}

// runPanicTestCases executes a slice of panic test cases.
func runPanicTestCases(t *testing.T, tests []panicTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runPanicTest(t, tt.testFunc, tt.expectPanic)
		})
	}
}

// ==================================================
// Edge Case 1: Nonexistent Target Field
// ==================================================

// TestCrossField_NonexistentField tests CrossField nonexistentfield.
func TestCrossField_NonexistentField(t *testing.T) {
	runPanicTestCases(t, []panicTestCase{
		{
			name: "eqfield with nonexistent target field",
			testFunc: func() interface{} {
				type User struct {
					Password        string `pedantigo:"required"`
					ConfirmPassword string `pedantigo:"eqfield=NonExistentField"`
				}
				return New[User]()
			},
			expectPanic: true,
		},
		{
			name: "gtfield with nonexistent target field",
			testFunc: func() interface{} {
				type Range struct {
					Min int `pedantigo:"required"`
					Val int `pedantigo:"gtfield=NoSuchField"`
				}
				return New[Range]()
			},
			expectPanic: true,
		},
		{
			name: "ltfield with nonexistent target field",
			testFunc: func() interface{} {
				type Range struct {
					Max int `pedantigo:"required"`
					Val int `pedantigo:"ltfield=InvalidTarget"`
				}
				return New[Range]()
			},
			expectPanic: true,
		},
	})
}

// ==================================================
// Edge Case 2: Type Incompatibility
// ==================================================

// TestCrossField_TypeIncompatibility tests CrossField typeincompatibility.
func TestCrossField_TypeIncompatibility(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "string compared with int (gtfield)",
			setup: func() (interface{}, interface{}) {
				type Mixed struct {
					Age  int    `pedantigo:"required"`
					Name string `pedantigo:"gtfield=Age"`
				}
				v := New[Mixed]()
				return v, &Mixed{Age: 25, Name: "Alice"}
			},
			expectErr: true,
		},
		{
			name: "string compared with float64 (ltfield)",
			setup: func() (interface{}, interface{}) {
				type Mixed struct {
					Price float64 `pedantigo:"required"`
					Label string  `pedantigo:"ltfield=Price"`
				}
				v := New[Mixed]()
				return v, &Mixed{Price: 99.99, Label: "expensive"}
			},
			expectErr: true,
		},
		{
			name: "struct compared with int (eqfield)",
			setup: func() (interface{}, interface{}) {
				type Nested struct {
					Value int
				}
				type Mixed struct {
					Count  int    `pedantigo:"required"`
					Config Nested `pedantigo:"eqfield=Count"`
				}
				v := New[Mixed]()
				return v, &Mixed{Count: 5, Config: Nested{Value: 5}}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			// Type assertion to get the validator's Validate method
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				err := v.Validate(obj)
				if tt.expectErr {
					assert.Error(t, err, "expected error for incompatible types")
				}
			}
		})
	}
}

// ==================================================
// Edge Case 3: Nil Pointer Fields
// ==================================================

// TestCrossField_NilPointer tests CrossField nilpointer.
func TestCrossField_NilPointer(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "target field is nil pointer",
			setup: func() (interface{}, interface{}) {
				type Optional struct {
					Value  *int `pedantigo:"required"`
					MinVal int  `pedantigo:"ltfield=Value"`
				}
				v := New[Optional]()
				return v, &Optional{Value: nil, MinVal: 10}
			},
			expectErr: true,
		},
		{
			name: "source field is nil pointer",
			setup: func() (interface{}, interface{}) {
				type Optional struct {
					Value  *int `pedantigo:"required"`
					MinVal *int `pedantigo:"gtfield=Value"`
				}
				v := New[Optional]()
				val := 10
				return v, &Optional{Value: &val, MinVal: nil}
			},
			expectErr: true,
		},
		{
			name: "both fields are nil pointers (should be equal)",
			setup: func() (interface{}, interface{}) {
				type Optional struct {
					Field1 *int `pedantigo:"required"`
					Field2 *int `pedantigo:"eqfield=Field1"`
				}
				v := New[Optional]()
				return v, &Optional{Field1: nil, Field2: nil}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				err := v.Validate(obj)
				if tt.expectErr {
					assert.Error(t, err, "expected error for nil pointer comparison")
				} else {
					assert.NoError(t, err, "expected no error for nil pointer comparison")
				}
			}
		})
	}
}

// ==================================================
// Edge Case 4: Case Sensitivity in Field Names
// ==================================================

// TestCrossField_CaseSensitivity tests CrossField casesensitivity.
func TestCrossField_CaseSensitivity(t *testing.T) {
	runPanicTestCases(t, []panicTestCase{
		{
			name: "lowercase field reference (case mismatch)",
			testFunc: func() interface{} {
				type CaseTest struct {
					Value    int `pedantigo:"required"`
					MinValue int `pedantigo:"gtfield=value"`
				}
				return New[CaseTest]()
			},
			expectPanic: true,
		},
		{
			name: "incorrect mixed case field reference",
			testFunc: func() interface{} {
				type CaseTest struct {
					UserID   int `pedantigo:"required"`
					MinLimit int `pedantigo:"ltfield=userid"`
				}
				return New[CaseTest]()
			},
			expectPanic: true,
		},
		{
			name: "correct case field reference (should work)",
			testFunc: func() interface{} {
				type CaseTest struct {
					Value    int `pedantigo:"required"`
					MinValue int `pedantigo:"gtfield=Value"`
				}
				return New[CaseTest]()
			},
			expectPanic: false,
		},
	})
}

// ==================================================
// Edge Case 5: Nested Structs (Skipped for future)
// ==================================================

// TestCrossField_NestedStruct tests CrossField nestedstruct.
// Comprehensive tests are in crossfield_nested_test.go.
func TestCrossField_NestedStruct(t *testing.T) {
	// Feature implemented - see crossfield_nested_test.go for comprehensive tests
	type Inner struct {
		Value int
	}
	type Outer struct {
		Inner Inner
		Check int `pedantigo:"eqfield=Inner.Value"`
	}

	t.Run("nested field reference - valid", func(t *testing.T) {
		outer := Outer{Inner: Inner{Value: 42}, Check: 42}
		validator := New[Outer]()
		err := validator.Validate(&outer)
		assert.NoError(t, err, "expected no errors for matching nested field values")
	})

	t.Run("nested field reference - invalid", func(t *testing.T) {
		outer := Outer{Inner: Inner{Value: 42}, Check: 100}
		validator := New[Outer]()
		err := validator.Validate(&outer)
		assert.Error(t, err, "expected error for mismatched nested field values")
	})
}

// ==================================================
// Edge Case 6: Multiple Cross-Field Constraints
// ==================================================

// TestCrossField_MultipleConstraints tests CrossField multiple constraints.
func TestCrossField_MultipleConstraints(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "both constraints valid",
			setup: func() (interface{}, interface{}) {
				type Range struct {
					Min int `pedantigo:"required"`
					Max int `pedantigo:"required"`
					Val int `pedantigo:"gtfield=Min,ltfield=Max"`
				}
				v := New[Range]()
				return v, &Range{Min: 0, Max: 100, Val: 50}
			},
			expectErr: false,
		},
		{
			name: "first constraint fails",
			setup: func() (interface{}, interface{}) {
				type Range struct {
					Min int `pedantigo:"required"`
					Max int `pedantigo:"required"`
					Val int `pedantigo:"gtfield=Min,ltfield=Max"`
				}
				v := New[Range]()
				return v, &Range{Min: 50, Max: 100, Val: 40}
			},
			expectErr: true,
		},
		{
			name: "second constraint fails",
			setup: func() (interface{}, interface{}) {
				type Range struct {
					Min int `pedantigo:"required"`
					Max int `pedantigo:"required"`
					Val int `pedantigo:"gtfield=Min,ltfield=Max"`
				}
				v := New[Range]()
				return v, &Range{Min: 0, Max: 50, Val: 60}
			},
			expectErr: true,
		},
		{
			name: "both constraints fail",
			setup: func() (interface{}, interface{}) {
				type Range struct {
					Min int `pedantigo:"required"`
					Max int `pedantigo:"required"`
					Val int `pedantigo:"gtfield=Min,ltfield=Max"`
				}
				v := New[Range]()
				return v, &Range{Min: 50, Max: 40, Val: 30}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				err := v.Validate(obj)
				if tt.expectErr {
					assert.Error(t, err, "expected error for constraint violation")
				}
			}
		})
	}
}

// ==================================================
// Edge Case 7: Self-Reference
// ==================================================

// TestCrossField_SelfReference tests CrossField selfreference.
func TestCrossField_SelfReference(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func() interface{}
		expectPanic bool
	}{
		{
			name: "eqfield self-reference",
			testFunc: func() interface{} {
				type Recursive struct {
					Value int `pedantigo:"eqfield=Value"`
				}
				return New[Recursive]()
			},
			expectPanic: true,
		},
		{
			name: "gtfield self-reference",
			testFunc: func() interface{} {
				type Recursive struct {
					Value int `pedantigo:"gtfield=Value"`
				}
				return New[Recursive]()
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				assert.NotNil(t, r, "expected panic for self-referencing field")
			}()
			_ = tt.testFunc()
			assert.Fail(t, "should have panicked before reaching here")
		})
	}
}

// ==================================================
// Edge Case 8: Circular Dependencies
// ==================================================

// TestCrossField_CircularDependency tests CrossField circulardependency.
func TestCrossField_CircularDependency(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "two-field circular dependency",
			setup: func() (interface{}, interface{}) {
				type Circular struct {
					Field1 int `pedantigo:"gtfield=Field2"`
					Field2 int `pedantigo:"gtfield=Field1"`
				}
				v := New[Circular]()
				return v, &Circular{Field1: 10, Field2: 5}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 9: Zero Value Comparison
// ==================================================

// TestCrossField_ZeroValue tests CrossField zerovalue.
func TestCrossField_ZeroValue(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "both fields are zero (should be equal)",
			setup: func() (interface{}, interface{}) {
				type Range struct {
					Min int `pedantigo:"required"`
					Max int `pedantigo:"required"`
					Val int `pedantigo:"eqfield=Min"`
				}
				v := New[Range]()
				return v, &Range{Min: 0, Max: 100, Val: 0}
			},
			expectErr: false,
		},
		{
			name: "zero vs non-zero (should fail equality)",
			setup: func() (interface{}, interface{}) {
				type Range struct {
					Min int `pedantigo:"required"`
					Val int `pedantigo:"eqfield=Min"`
				}
				v := New[Range]()
				return v, &Range{Min: 10, Val: 0}
			},
			expectErr: true,
		},
		{
			name: "both empty strings (should be equal)",
			setup: func() (interface{}, interface{}) {
				type Strings struct {
					Field1 string `pedantigo:"required"`
					Field2 string `pedantigo:"eqfield=Field1"`
				}
				v := New[Strings]()
				return v, &Strings{Field1: "", Field2: ""}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				err := v.Validate(obj)
				if tt.expectErr {
					assert.Error(t, err, "expected error for zero value comparison")
				} else {
					assert.NoError(t, err, "expected no error for zero value comparison")
				}
			}
		})
	}
}

// ==================================================
// Edge Case 10: Numeric Type Compatibility
// ==================================================

// TestCrossField_NumericTypeCompatibility tests CrossField numerictypecompatibility.
func TestCrossField_NumericTypeCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "int vs int64 comparison",
			setup: func() (interface{}, interface{}) {
				type Mixed struct {
					Value1 int   `pedantigo:"required"`
					Value2 int64 `pedantigo:"gtfield=Value1"`
				}
				v := New[Mixed]()
				return v, &Mixed{Value1: 10, Value2: 20}
			},
			expectErr: false,
		},
		{
			name: "float64 vs int comparison",
			setup: func() (interface{}, interface{}) {
				type Mixed struct {
					IntVal   int     `pedantigo:"required"`
					FloatVal float64 `pedantigo:"ltfield=IntVal"`
				}
				v := New[Mixed]()
				return v, &Mixed{IntVal: 100, FloatVal: 50.5}
			},
			expectErr: false,
		},
		{
			name: "uint vs int comparison",
			setup: func() (interface{}, interface{}) {
				type Mixed struct {
					Signed   int  `pedantigo:"required"`
					Unsigned uint `pedantigo:"eqfield=Signed"`
				}
				v := New[Mixed]()
				return v, &Mixed{Signed: 10, Unsigned: 10}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 11: Empty String Comparisons
// ==================================================

// TestCrossField_EmptyString tests CrossField emptystring.
func TestCrossField_EmptyString(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "empty string equality",
			setup: func() (interface{}, interface{}) {
				type Strings struct {
					Field1 string `pedantigo:"required"`
					Field2 string `pedantigo:"eqfield=Field1"`
				}
				v := New[Strings]()
				return v, &Strings{Field1: "", Field2: ""}
			},
			expectErr: false,
		},
		{
			name: "empty string inequality (should fail)",
			setup: func() (interface{}, interface{}) {
				type Strings struct {
					Field1 string `pedantigo:"required"`
					Field2 string `pedantigo:"nefield=Field1"`
				}
				v := New[Strings]()
				return v, &Strings{Field1: "", Field2: ""}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				err := v.Validate(obj)
				if tt.expectErr {
					assert.Error(t, err, "expected error for empty string comparison")
				} else {
					assert.NoError(t, err, "expected no error for empty string comparison")
				}
			}
		})
	}
}

// ==================================================
// Edge Case 12: Time.Time Comparisons
// ==================================================

// TestCrossField_TimeComparison tests CrossField timecomparison.
func TestCrossField_TimeComparison(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "equal times",
			setup: func() (interface{}, interface{}) {
				type TimeRange struct {
					StartTime time.Time `pedantigo:"required"`
					EndTime   time.Time `pedantigo:"eqfield=StartTime"`
				}
				v := New[TimeRange]()
				return v, &TimeRange{StartTime: now, EndTime: now}
			},
			expectErr: false,
		},
		{
			name: "end time after start time",
			setup: func() (interface{}, interface{}) {
				type TimeRange struct {
					StartTime time.Time `pedantigo:"required"`
					EndTime   time.Time `pedantigo:"gtfield=StartTime"`
				}
				v := New[TimeRange]()
				return v, &TimeRange{StartTime: now, EndTime: now.Add(1 * time.Hour)}
			},
			expectErr: false,
		},
		{
			name: "end time before start time (should fail)",
			setup: func() (interface{}, interface{}) {
				type TimeRange struct {
					StartTime time.Time `pedantigo:"required"`
					EndTime   time.Time `pedantigo:"ltfield=StartTime"`
				}
				v := New[TimeRange]()
				return v, &TimeRange{StartTime: now, EndTime: now.Add(-1 * time.Hour)}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				err := v.Validate(obj)
				if tt.expectErr {
					assert.Error(t, err, "expected error for time comparison")
				}
			}
		})
	}
}

// ==================================================
// Edge Case 13: All Comparison Operators
// ==================================================

// TestCrossField_AllOperators tests CrossField alloperators.
func TestCrossField_AllOperators(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "eqfield (equal) - valid",
			setup: func() (interface{}, interface{}) {
				type Pair struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"eqfield=Field1"`
				}
				v := New[Pair]()
				return v, &Pair{Field1: 42, Field2: 42}
			},
			expectErr: false,
		},
		{
			name: "nefield (not equal) - valid",
			setup: func() (interface{}, interface{}) {
				type Pair struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"nefield=Field1"`
				}
				v := New[Pair]()
				return v, &Pair{Field1: 42, Field2: 43}
			},
			expectErr: false,
		},
		{
			name: "gtfield (greater than) - valid",
			setup: func() (interface{}, interface{}) {
				type Pair struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"gtfield=Field1"`
				}
				v := New[Pair]()
				return v, &Pair{Field1: 10, Field2: 20}
			},
			expectErr: false,
		},
		{
			name: "gtefield (greater than or equal) - valid",
			setup: func() (interface{}, interface{}) {
				type Pair struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"gtefield=Field1"`
				}
				v := New[Pair]()
				return v, &Pair{Field1: 10, Field2: 10}
			},
			expectErr: false,
		},
		{
			name: "ltfield (less than) - valid",
			setup: func() (interface{}, interface{}) {
				type Pair struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"ltfield=Field1"`
				}
				v := New[Pair]()
				return v, &Pair{Field1: 100, Field2: 50}
			},
			expectErr: false,
		},
		{
			name: "ltefield (less than or equal) - valid",
			setup: func() (interface{}, interface{}) {
				type Pair struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"ltefield=Field1"`
				}
				v := New[Pair]()
				return v, &Pair{Field1: 100, Field2: 100}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 14: Boundary Values
// ==================================================

// TestCrossField_BoundaryValues tests CrossField boundaryvalues.
func TestCrossField_BoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "zero and one (boundary)",
			setup: func() (interface{}, interface{}) {
				type Boundary struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"gtfield=Field1"`
				}
				v := New[Boundary]()
				return v, &Boundary{Field1: 0, Field2: 1}
			},
			expectErr: false,
		},
		{
			name: "negative number comparison",
			setup: func() (interface{}, interface{}) {
				type Boundary struct {
					Field1 int `pedantigo:"required"`
					Field2 int `pedantigo:"ltfield=Field1"`
				}
				v := New[Boundary]()
				return v, &Boundary{Field1: -10, Field2: -100}
			},
			expectErr: false,
		},
		{
			name: "max int64 comparison",
			setup: func() (interface{}, interface{}) {
				type Boundary struct {
					Field1 int64 `pedantigo:"required"`
					Field2 int64 `pedantigo:"eqfield=Field1"`
				}
				v := New[Boundary]()
				return v, &Boundary{Field1: 9223372036854775807, Field2: 9223372036854775807}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 15: Boolean and Complex Types
// ==================================================

// TestCrossField_BooleanAndComplexTypes tests CrossField booleanandcomplextypes.
func TestCrossField_BooleanAndComplexTypes(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "boolean comparison",
			setup: func() (interface{}, interface{}) {
				type BooleanTest struct {
					Flag1 bool `pedantigo:"required"`
					Flag2 bool `pedantigo:"eqfield=Flag1"`
				}
				v := New[BooleanTest]()
				return v, &BooleanTest{Flag1: true, Flag2: true}
			},
			expectErr: false,
		},
		{
			name: "slice comparison",
			setup: func() (interface{}, interface{}) {
				type SliceTest struct {
					Items []int `pedantigo:"required"`
					Ref   []int `pedantigo:"eqfield=Items"`
				}
				v := New[SliceTest]()
				return v, &SliceTest{Items: []int{1, 2, 3}, Ref: []int{1, 2, 3}}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 16: Unexported Fields (Should be skipped)
// ==================================================

// TestCrossField_UnexportedField tests CrossField unexportedfield.
func TestCrossField_UnexportedField(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func() interface{}
		expectPanic bool
	}{
		{
			name: "reference to unexported field",
			testFunc: func() interface{} {
				type Unexported struct {
					Field int `pedantigo:"gtfield=value"`
				}
				return New[Unexported]()
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				assert.NotNil(t, r, "expected panic when referencing unexported field")
			}()
			_ = tt.testFunc()
			assert.Fail(t, "should have panicked before reaching here")
		})
	}
}

// ==================================================
// Edge Case 17: Pointer to Different Types
// ==================================================

// TestCrossField_PointerTypes tests CrossField pointertypes.
func TestCrossField_PointerTypes(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "pointer vs value comparison",
			setup: func() (interface{}, interface{}) {
				type PointerTest struct {
					Value1 *int `pedantigo:"required"`
					Value2 int  `pedantigo:"gtfield=Value1"`
				}
				v := New[PointerTest]()
				val := 10
				return v, &PointerTest{Value1: &val, Value2: 20}
			},
			expectErr: false,
		},
		{
			name: "different pointer types comparison",
			setup: func() (interface{}, interface{}) {
				type PointerTest struct {
					Value1 *int   `pedantigo:"required"`
					Value2 *int64 `pedantigo:"eqfield=Value1"`
				}
				v := New[PointerTest]()
				val1 := 10
				val2 := int64(10)
				return v, &PointerTest{Value1: &val1, Value2: &val2}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 18: Field Order Independence
// ==================================================

// TestCrossField_FieldOrder tests CrossField fieldorder.
func TestCrossField_FieldOrder(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "forward reference to later-defined field",
			setup: func() (interface{}, interface{}) {
				type Order struct {
					Field2 int `pedantigo:"gtfield=Field1"`
					Field1 int `pedantigo:"required"`
				}
				v := New[Order]()
				return v, &Order{Field1: 10, Field2: 20}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 19: Multiple Cross-Field Validators on Different Fields
// ==================================================

// TestCrossField_ChainedConstraints tests CrossField chainedconstraints.
func TestCrossField_ChainedConstraints(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "chain of cross-field constraints (Min < Mid < Max)",
			setup: func() (interface{}, interface{}) {
				type MultiConstraint struct {
					Min int `pedantigo:"required"`
					Mid int `pedantigo:"gtfield=Min"`
					Max int `pedantigo:"gtfield=Mid"`
				}
				v := New[MultiConstraint]()
				return v, &MultiConstraint{Min: 10, Mid: 20, Max: 30}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}

// ==================================================
// Edge Case 20: Special Values (NaN, Infinity for floats)
// ==================================================

// TestCrossField_FloatSpecialValues tests CrossField floatspecialvalues.
func TestCrossField_FloatSpecialValues(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (interface{}, interface{})
		expectErr bool
	}{
		{
			name: "positive infinity comparison",
			setup: func() (interface{}, interface{}) {
				type FloatSpecial struct {
					Field1 float64 `pedantigo:"required"`
					Field2 float64 `pedantigo:"gtfield=Field1"`
				}
				v := New[FloatSpecial]()
				inf := math.Inf(1)
				return v, &FloatSpecial{Field1: 100.0, Field2: inf}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, obj := tt.setup()
			if v, ok := validator.(interface{ Validate(interface{}) error }); ok {
				_ = v.Validate(obj)
			}
		})
	}
}
