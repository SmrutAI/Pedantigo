package constraints

import (
	"reflect"
	"strings"
	"testing"
)

// Test helper types.
type SimpleStruct struct {
	Name  string
	Value int
}

type NestedOuter struct {
	Inner SimpleStruct
	Count int
}

type DeepNested struct {
	Level1 struct {
		Level2 struct {
			Level3 struct {
				DeepValue string
			}
		}
	}
}

type WithPointer struct {
	Inner *SimpleStruct
	Value int
}

type WithDoublePointer struct {
	Inner **SimpleStruct
	Value int
}

// unexported field struct (used via reflection in tests).
type withUnexported struct {
	Exported   string
	unexported string //nolint:unused // Accessed via reflection in TestParseFieldPath_UnexportedField_Panics
}

// TestParseFieldPath_SingleLevel tests parsing a simple field name without dots.
func TestParseFieldPath_SingleLevel(t *testing.T) {
	tests := []struct {
		name          string
		structType    reflect.Type
		path          string
		expectedParts []string
	}{
		{
			name:          "parse simple Name field",
			structType:    reflect.TypeOf(SimpleStruct{}),
			path:          "Name",
			expectedParts: []string{"Name"},
		},
		{
			name:          "parse simple Value field",
			structType:    reflect.TypeOf(SimpleStruct{}),
			path:          "Value",
			expectedParts: []string{"Value"},
		},
		{
			name:          "parse Count field from NestedOuter",
			structType:    reflect.TypeOf(NestedOuter{}),
			path:          "Count",
			expectedParts: []string{"Count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := ParseFieldPath(tt.structType, tt.path)

			if fp == nil {
				t.Fatal("ParseFieldPath returned nil")
			}

			if fp.Raw != tt.path {
				t.Errorf("Raw path mismatch: got %q, want %q", fp.Raw, tt.path)
			}

			if len(fp.Parts) != len(tt.expectedParts) {
				t.Errorf("Parts length mismatch: got %d, want %d", len(fp.Parts), len(tt.expectedParts))
			}

			for i, part := range tt.expectedParts {
				if i >= len(fp.Parts) {
					break
				}
				if fp.Parts[i] != part {
					t.Errorf("Part[%d] mismatch: got %q, want %q", i, fp.Parts[i], part)
				}
			}

			// TypeAtLevel and IndexAtLevel should have same length as Parts
			if len(fp.TypeAtLevel) != len(fp.Parts) {
				t.Errorf("TypeAtLevel length mismatch: got %d, want %d", len(fp.TypeAtLevel), len(fp.Parts))
			}

			if len(fp.IndexAtLevel) != len(fp.Parts) {
				t.Errorf("IndexAtLevel length mismatch: got %d, want %d", len(fp.IndexAtLevel), len(fp.Parts))
			}
		})
	}
}

// TestParseFieldPath_MultiLevel tests parsing nested dotted paths.
func TestParseFieldPath_MultiLevel(t *testing.T) {
	tests := []struct {
		name          string
		structType    reflect.Type
		path          string
		expectedParts []string
	}{
		{
			name:          "parse Inner.Name from NestedOuter",
			structType:    reflect.TypeOf(NestedOuter{}),
			path:          "Inner.Name",
			expectedParts: []string{"Inner", "Name"},
		},
		{
			name:          "parse Inner.Value from NestedOuter",
			structType:    reflect.TypeOf(NestedOuter{}),
			path:          "Inner.Value",
			expectedParts: []string{"Inner", "Value"},
		},
		{
			name:          "parse deep nested path",
			structType:    reflect.TypeOf(DeepNested{}),
			path:          "Level1.Level2.Level3.DeepValue",
			expectedParts: []string{"Level1", "Level2", "Level3", "DeepValue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := ParseFieldPath(tt.structType, tt.path)

			if fp == nil {
				t.Fatal("ParseFieldPath returned nil")
			}

			if fp.Raw != tt.path {
				t.Errorf("Raw path mismatch: got %q, want %q", fp.Raw, tt.path)
			}

			if len(fp.Parts) != len(tt.expectedParts) {
				t.Errorf("Parts length mismatch: got %d, want %d", len(fp.Parts), len(tt.expectedParts))
			}

			for i, part := range tt.expectedParts {
				if i >= len(fp.Parts) {
					break
				}
				if fp.Parts[i] != part {
					t.Errorf("Part[%d] mismatch: got %q, want %q", i, fp.Parts[i], part)
				}
			}

			// TypeAtLevel should be populated for each part
			if len(fp.TypeAtLevel) != len(fp.Parts) {
				t.Errorf("TypeAtLevel length mismatch: got %d, want %d", len(fp.TypeAtLevel), len(fp.Parts))
			}

			// Verify types are not nil
			for i, typ := range fp.TypeAtLevel {
				if typ == nil {
					t.Errorf("TypeAtLevel[%d] is nil", i)
				}
			}

			// IndexAtLevel should have correct field indices
			if len(fp.IndexAtLevel) != len(fp.Parts) {
				t.Errorf("IndexAtLevel length mismatch: got %d, want %d", len(fp.IndexAtLevel), len(fp.Parts))
			}
		})
	}
}

// TestParseFieldPath_WithPointer tests paths through pointer types.
func TestParseFieldPath_WithPointer(t *testing.T) {
	tests := []struct {
		name          string
		structType    reflect.Type
		path          string
		expectedParts []string
	}{
		{
			name:          "parse through pointer to struct",
			structType:    reflect.TypeOf(WithPointer{}),
			path:          "Inner.Name",
			expectedParts: []string{"Inner", "Name"},
		},
		{
			name:          "parse through pointer value field",
			structType:    reflect.TypeOf(WithPointer{}),
			path:          "Inner.Value",
			expectedParts: []string{"Inner", "Value"},
		},
		{
			name:          "parse through double pointer",
			structType:    reflect.TypeOf(WithDoublePointer{}),
			path:          "Inner.Name",
			expectedParts: []string{"Inner", "Name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := ParseFieldPath(tt.structType, tt.path)

			if fp == nil {
				t.Fatal("ParseFieldPath returned nil")
			}

			if len(fp.Parts) != len(tt.expectedParts) {
				t.Errorf("Parts length mismatch: got %d, want %d", len(fp.Parts), len(tt.expectedParts))
			}

			for i, part := range tt.expectedParts {
				if i >= len(fp.Parts) {
					break
				}
				if fp.Parts[i] != part {
					t.Errorf("Part[%d] mismatch: got %q, want %q", i, fp.Parts[i], part)
				}
			}
		})
	}
}

// TestParseFieldPath_InvalidField_Panics tests that invalid field names panic.
func TestParseFieldPath_InvalidField_Panics(t *testing.T) {
	tests := []struct {
		name       string
		structType reflect.Type
		path       string
	}{
		{
			name:       "non-existent field on SimpleStruct",
			structType: reflect.TypeOf(SimpleStruct{}),
			path:       "NonExistent",
		},
		{
			name:       "non-existent nested field",
			structType: reflect.TypeOf(NestedOuter{}),
			path:       "Inner.NonExistent",
		},
		{
			name:       "invalid first-level field",
			structType: reflect.TypeOf(NestedOuter{}),
			path:       "DoesNotExist.Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("ParseFieldPath did not panic for invalid field %q", tt.path)
				}
			}()

			ParseFieldPath(tt.structType, tt.path)
		})
	}
}

// TestParseFieldPath_UnexportedField_Panics tests that unexported fields panic.
func TestParseFieldPath_UnexportedField_Panics(t *testing.T) {
	tests := []struct {
		name       string
		structType reflect.Type
		path       string
	}{
		{
			name:       "unexported field access",
			structType: reflect.TypeOf(withUnexported{}),
			path:       "unexported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("ParseFieldPath did not panic for unexported field %q", tt.path)
				}
			}()

			ParseFieldPath(tt.structType, tt.path)
		})
	}
}

// TestFieldPath_ResolveValue_Success tests successful value resolution.
func TestFieldPath_ResolveValue_Success(t *testing.T) {
	tests := []struct {
		name          string
		structValue   interface{}
		path          string
		expectedValue interface{}
	}{
		{
			name:          "resolve simple Name field",
			structValue:   SimpleStruct{Name: "test", Value: 42},
			path:          "Name",
			expectedValue: "test",
		},
		{
			name:          "resolve simple Value field",
			structValue:   SimpleStruct{Name: "test", Value: 42},
			path:          "Value",
			expectedValue: 42,
		},
		{
			name: "resolve nested Inner.Value",
			structValue: NestedOuter{
				Inner: SimpleStruct{Name: "inner", Value: 42},
				Count: 10,
			},
			path:          "Inner.Value",
			expectedValue: 42,
		},
		{
			name: "resolve nested Inner.Name",
			structValue: NestedOuter{
				Inner: SimpleStruct{Name: "inner", Value: 42},
				Count: 10,
			},
			path:          "Inner.Name",
			expectedValue: "inner",
		},
		{
			name: "resolve deep nested value",
			structValue: func() DeepNested {
				d := DeepNested{}
				d.Level1.Level2.Level3.DeepValue = "deep"
				return d
			}(),
			path:          "Level1.Level2.Level3.DeepValue",
			expectedValue: "deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType := reflect.TypeOf(tt.structValue)
			fp := ParseFieldPath(structType, tt.path)

			value, err := fp.ResolveValue(reflect.ValueOf(tt.structValue))
			if err != nil {
				t.Fatalf("ResolveValue returned error: %v", err)
			}

			if value != tt.expectedValue {
				t.Errorf("ResolveValue mismatch: got %v, want %v", value, tt.expectedValue)
			}
		})
	}
}

// TestFieldPath_ResolveValue_WithPointer tests resolution through pointers.
func TestFieldPath_ResolveValue_WithPointer(t *testing.T) {
	tests := []struct {
		name          string
		structValue   interface{}
		path          string
		expectedValue interface{}
	}{
		{
			name: "resolve through pointer to struct",
			structValue: WithPointer{
				Inner: &SimpleStruct{Name: "ptr", Value: 100},
				Value: 5,
			},
			path:          "Inner.Name",
			expectedValue: "ptr",
		},
		{
			name: "resolve through pointer value field",
			structValue: WithPointer{
				Inner: &SimpleStruct{Name: "ptr", Value: 100},
				Value: 5,
			},
			path:          "Inner.Value",
			expectedValue: 100,
		},
		{
			name: "resolve through double pointer",
			structValue: func() WithDoublePointer {
				inner := &SimpleStruct{Name: "doubleptr", Value: 200}
				return WithDoublePointer{Inner: &inner, Value: 10}
			}(),
			path:          "Inner.Name",
			expectedValue: "doubleptr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType := reflect.TypeOf(tt.structValue)
			fp := ParseFieldPath(structType, tt.path)

			value, err := fp.ResolveValue(reflect.ValueOf(tt.structValue))
			if err != nil {
				t.Fatalf("ResolveValue returned error: %v", err)
			}

			if value != tt.expectedValue {
				t.Errorf("ResolveValue mismatch: got %v, want %v", value, tt.expectedValue)
			}
		})
	}
}

// TestFieldPath_ResolveValue_NilPointer tests error on nil pointer in path.
func TestFieldPath_ResolveValue_NilPointer(t *testing.T) {
	tests := []struct {
		name        string
		structValue interface{}
		path        string
	}{
		{
			name: "nil pointer in path",
			structValue: WithPointer{
				Inner: nil,
				Value: 5,
			},
			path: "Inner.Name",
		},
		{
			name: "nil double pointer in path",
			structValue: WithDoublePointer{
				Inner: nil,
				Value: 10,
			},
			path: "Inner.Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType := reflect.TypeOf(tt.structValue)
			fp := ParseFieldPath(structType, tt.path)

			value, err := fp.ResolveValue(reflect.ValueOf(tt.structValue))
			if err == nil {
				t.Errorf("ResolveValue should return error for nil pointer, got value: %v", value)
			}

			// Error message should indicate nil pointer
			if err != nil && !strings.Contains(err.Error(), "nil") {
				t.Errorf("Error message should mention 'nil', got: %v", err)
			}
		})
	}
}

// TestFieldPath_isNested tests the isNested helper.
func TestFieldPath_isNested(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected bool
	}{
		{
			name:     "single-level path",
			parts:    []string{"Name"},
			expected: false,
		},
		{
			name:     "two-level path",
			parts:    []string{"Inner", "Name"},
			expected: true,
		},
		{
			name:     "multi-level path",
			parts:    []string{"Level1", "Level2", "Level3", "DeepValue"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FieldPath{Parts: tt.parts}
			result := fp.isNested()

			if result != tt.expected {
				t.Errorf("isNested() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFieldPath_String tests the String method.
func TestFieldPath_String(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "simple path",
			raw:      "Name",
			expected: "Name",
		},
		{
			name:     "nested path",
			raw:      "Inner.Name",
			expected: "Inner.Name",
		},
		{
			name:     "deep nested path",
			raw:      "Level1.Level2.Level3.DeepValue",
			expected: "Level1.Level2.Level3.DeepValue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FieldPath{Raw: tt.raw}
			result := fp.String()

			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
