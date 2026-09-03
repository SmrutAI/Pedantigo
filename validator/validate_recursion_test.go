package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recNode is a recursive type for testing deep recursion validation.
type recNode struct {
	Name     string     `json:"name" validate:"required,min=2"`
	Children []*recNode `json:"children" validate:"dive"`
}

// TestValidateB_DeepRecursiveInvalidCaught tests that invalid fields at depth > 1
// are caught and reported, not skipped due to cycle detection.
func TestValidateB_DeepRecursiveInvalidCaught(t *testing.T) {
	v := New[recNode]()

	// Build a 2-deep tree where the grandchild has an invalid Name (too short)
	tree := &recNode{
		Name: "root",
		Children: []*recNode{
			{
				Name: "child",
				Children: []*recNode{
					{
						Name: "x", // Invalid: min=2 requires at least 2 chars
					},
				},
			},
		},
	}

	err := v.Validate(tree)
	require.Error(t, err)

	// Extract the ValidationError and check that the deep field error is present
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	// Search for the grandchild's Name error. Field is a dotted Go-field path
	// (e.g. "Children[0].Children[0].Name"), not a bare field name.
	foundDeepError := false
	for _, fieldErr := range ve.Errors {
		if strings.Contains(fieldErr.Field, "Name") {
			foundDeepError = true
			break
		}
	}

	assert.True(t, foundDeepError, "expected validation error for deeply nested Name field")
}

// TestValidateB_CyclicInMemoryTerminates tests that a back-reference cycle with valid data
// terminates without hanging (visited-pointer guard prevents infinite loop).
func TestValidateB_CyclicInMemoryTerminates(t *testing.T) {
	v := New[recNode]()

	// Create a back-reference cycle: root -> child -> parent (back to root)
	root := &recNode{Name: "root"}
	child := &recNode{Name: "child", Children: []*recNode{}}
	root.Children = []*recNode{child}
	// Create cycle: child points back to root
	child.Children = []*recNode{root}

	// This should not hang; the visited-pointer guard prevents infinite recursion
	err := v.Validate(root)
	require.NoError(t, err, "cyclic reference with valid data should not cause error")
}

// TestValidateB_DiamondReValidated tests that a shared sub-object (diamond pattern)
// is validated, not skipped as a cycle.
func TestValidateB_DiamondReValidated(t *testing.T) {
	type innerNode struct {
		X int `json:"x" validate:"min=1"`
	}
	type diamondNode struct {
		A *innerNode `json:"a"`
		B *innerNode `json:"b"`
	}

	v := New[diamondNode]()

	// Both A and B point to the same invalid innerNode
	shared := &innerNode{X: 0} // Invalid: min=1 requires X >= 1
	node := &diamondNode{
		A: shared,
		B: shared,
	}

	err := v.Validate(node)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	// The shared inner node's constraint should be reported. Field is a dotted
	// Go-field path (e.g. "A.X"), not a bare field name.
	foundError := false
	for _, fieldErr := range ve.Errors {
		if strings.Contains(fieldErr.Field, "X") {
			foundError = true
			break
		}
	}

	assert.True(t, foundError, "expected validation error for shared inner node's X field")
}

// TestValidateB_DepthCapExceeded tests that exceeding MaxRecursionDepth produces an error.
//
// The root passed to Validate() counts as depth 1 (matching Unmarshal, where
// the outermost object is depth 1 too). For this 3-level tree (root -> child
// -> grandchild): root=1, child=2, grandchild=3, so MaxRecursionDepth: 2 is
// the cap that grandchild exceeds.
func TestValidateB_DepthCapExceeded(t *testing.T) {
	v := New[recNode](Options{
		StrictMissingFields: true,
		ExtraFields:         ExtraIgnore,
		MaxRecursionDepth:   2,
	})

	// Build a 3-level valid tree (root -> child -> grandchild), all with valid data
	tree := &recNode{
		Name: "root",
		Children: []*recNode{
			{
				Name: "child",
				Children: []*recNode{
					{Name: "grandchild"},
				},
			},
		},
	}

	err := v.Validate(tree)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

	// The error message should contain "max recursion depth 2 exceeded"
	foundDepthError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Message != "" && strings.Contains(fieldErr.Message, "max recursion depth 2 exceeded") {
			foundDepthError = true
			break
		}
	}

	assert.True(t, foundDepthError, "expected error message containing 'max recursion depth 2 exceeded', got errors: %+v", ve.Errors)
}

// TestValidateB_RaiseCapAllowsDeeper tests that raising MaxRecursionDepth allows deeper trees.
func TestValidateB_RaiseCapAllowsDeeper(t *testing.T) {
	v := New[recNode](Options{
		StrictMissingFields: true,
		ExtraFields:         ExtraIgnore,
		MaxRecursionDepth:   5,
	})

	// Build a 3-level valid tree (root -> child -> grandchild), all with valid data
	tree := &recNode{
		Name: "root",
		Children: []*recNode{
			{
				Name: "child",
				Children: []*recNode{
					{Name: "grandchild"},
				},
			},
		},
	}

	err := v.Validate(tree)
	require.NoError(t, err, "with MaxRecursionDepth=5, a 3-level tree should validate successfully")
}

// TestValidateB_DefaultCapKeepsExistingTreesGreen tests that the default MaxRecursionDepth (3)
// allows a 3-level valid tree to pass validation.
func TestValidateB_DefaultCapKeepsExistingTreesGreen(t *testing.T) {
	// Use default options (MaxRecursionDepth defaults to 3)
	v := New[recNode]()

	// Build a 3-level tree matching the pattern from TestCircularReference_SelfReferencing:
	// root -> [child1, child2 -> [grandchild]]
	tree := &recNode{
		Name: "root",
		Children: []*recNode{
			{Name: "child1"},
			{
				Name: "child2",
				Children: []*recNode{
					{Name: "grandchild"},
				},
			},
		},
	}

	err := v.Validate(tree)
	require.NoError(t, err, "default MaxRecursionDepth=3 should allow a 3-level tree")
}
