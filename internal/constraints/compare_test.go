package constraints

import (
	"reflect"
	"testing"
	"time"
)

func TestCompareValues_Numeric(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		left    any
		right   any
		wantRes bool
		wantErr bool
	}{
		// Int comparisons
		{name: "int eq - equal", op: "eq", left: 10, right: 10, wantRes: true},
		{name: "int eq - not equal", op: "eq", left: 10, right: 20, wantRes: false},
		{name: "int ne - not equal", op: "ne", left: 10, right: 20, wantRes: true},
		{name: "int ne - equal", op: "ne", left: 10, right: 10, wantRes: false},
		{name: "int gt - greater", op: "gt", left: 20, right: 10, wantRes: true},
		{name: "int gt - not greater", op: "gt", left: 10, right: 20, wantRes: false},
		{name: "int gt - equal", op: "gt", left: 10, right: 10, wantRes: false},
		{name: "int gte - greater", op: "gte", left: 20, right: 10, wantRes: true},
		{name: "int gte - equal", op: "gte", left: 10, right: 10, wantRes: true},
		{name: "int gte - less", op: "gte", left: 5, right: 10, wantRes: false},
		{name: "int lt - less", op: "lt", left: 10, right: 20, wantRes: true},
		{name: "int lt - not less", op: "lt", left: 20, right: 10, wantRes: false},
		{name: "int lt - equal", op: "lt", left: 10, right: 10, wantRes: false},
		{name: "int lte - less", op: "lte", left: 10, right: 20, wantRes: true},
		{name: "int lte - equal", op: "lte", left: 10, right: 10, wantRes: true},
		{name: "int lte - greater", op: "lte", left: 20, right: 10, wantRes: false},

		// Float comparisons
		{name: "float eq - equal", op: "eq", left: 3.14, right: 3.14, wantRes: true},
		{name: "float eq - not equal", op: "eq", left: 3.14, right: 2.71, wantRes: false},
		{name: "float gt - greater", op: "gt", left: 3.14, right: 2.71, wantRes: true},
		{name: "float lt - less", op: "lt", left: 2.71, right: 3.14, wantRes: true},
		{name: "float gte - equal", op: "gte", left: 3.14, right: 3.14, wantRes: true},
		{name: "float lte - equal", op: "lte", left: 3.14, right: 3.14, wantRes: true},

		// Mixed numeric types
		{name: "int vs float eq", op: "eq", left: 10, right: 10.0, wantRes: true},
		{name: "int vs float gt", op: "gt", left: 20, right: 10.0, wantRes: true},
		{name: "uint vs int eq", op: "eq", left: uint(10), right: 10, wantRes: true},
		{name: "uint vs int lt", op: "lt", left: uint(5), right: 10, wantRes: true},
		{name: "int8 vs int64", op: "eq", left: int8(10), right: int64(10), wantRes: true},
		{name: "uint16 vs float32", op: "gt", left: uint16(20), right: float32(10.5), wantRes: true},

		// Zero values
		{name: "int zero eq", op: "eq", left: 0, right: 0, wantRes: true},
		{name: "float zero eq", op: "eq", left: 0.0, right: 0.0, wantRes: true},
		{name: "negative numbers", op: "lt", left: -10, right: -5, wantRes: true},

		// Boundary values
		{name: "very large positive", op: "gt", left: 1000000, right: 0, wantRes: true},
		{name: "very large negative", op: "lt", left: -1000000, right: 0, wantRes: true},
	}

	// Common validation logic
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CompareValues(tt.op, tt.left, tt.right)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if res != tt.wantRes {
				t.Errorf("expected result=%v, got %v", tt.wantRes, res)
			}
		})
	}
}

func TestCompareValues_String(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		left    string
		right   string
		wantRes bool
	}{
		{name: "string eq - equal", op: "eq", left: "hello", right: "hello", wantRes: true},
		{name: "string eq - not equal", op: "eq", left: "hello", right: "world", wantRes: false},
		{name: "string ne - not equal", op: "ne", left: "hello", right: "world", wantRes: true},
		{name: "string ne - equal", op: "ne", left: "hello", right: "hello", wantRes: false},
		{name: "string gt - lexicographic", op: "gt", left: "world", right: "hello", wantRes: true},
		{name: "string gt - not greater", op: "gt", left: "hello", right: "world", wantRes: false},
		{name: "string gte - greater", op: "gte", left: "world", right: "hello", wantRes: true},
		{name: "string gte - equal", op: "gte", left: "hello", right: "hello", wantRes: true},
		{name: "string lt - lexicographic", op: "lt", left: "hello", right: "world", wantRes: true},
		{name: "string lte - less", op: "lte", left: "hello", right: "world", wantRes: true},
		{name: "string lte - equal", op: "lte", left: "hello", right: "hello", wantRes: true},
		{name: "empty strings eq", op: "eq", left: "", right: "", wantRes: true},
		{name: "empty vs non-empty lt", op: "lt", left: "", right: "a", wantRes: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CompareValues(tt.op, tt.left, tt.right)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if res != tt.wantRes {
				t.Errorf("expected result=%v, got %v", tt.wantRes, res)
			}
		})
	}
}

func TestCompareValues_Time(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	tests := []struct {
		name    string
		op      string
		left    time.Time
		right   time.Time
		wantRes bool
	}{
		{name: "time eq - equal", op: "eq", left: now, right: now, wantRes: true},
		{name: "time eq - not equal", op: "eq", left: earlier, right: later, wantRes: false},
		{name: "time ne - not equal", op: "ne", left: earlier, right: later, wantRes: true},
		{name: "time ne - equal", op: "ne", left: now, right: now, wantRes: false},
		{name: "time gt - after", op: "gt", left: later, right: earlier, wantRes: true},
		{name: "time gt - not after", op: "gt", left: earlier, right: later, wantRes: false},
		{name: "time gte - after", op: "gte", left: later, right: earlier, wantRes: true},
		{name: "time gte - equal", op: "gte", left: now, right: now, wantRes: true},
		{name: "time lt - before", op: "lt", left: earlier, right: later, wantRes: true},
		{name: "time lt - not before", op: "lt", left: later, right: earlier, wantRes: false},
		{name: "time lte - before", op: "lte", left: earlier, right: later, wantRes: true},
		{name: "time lte - equal", op: "lte", left: now, right: now, wantRes: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CompareValues(tt.op, tt.left, tt.right)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if res != tt.wantRes {
				t.Errorf("expected result=%v, got %v", tt.wantRes, res)
			}
		})
	}
}

func TestCompareValues_Nil(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		left    any
		right   any
		wantRes bool
		wantErr bool
	}{
		// Both nil
		{name: "nil eq nil", op: "eq", left: nil, right: nil, wantRes: true, wantErr: false},
		{name: "nil ne nil", op: "ne", left: nil, right: nil, wantRes: false, wantErr: false},
		{name: "nil gt nil", op: "gt", left: nil, right: nil, wantRes: false, wantErr: false},
		{name: "nil gte nil", op: "gte", left: nil, right: nil, wantRes: false, wantErr: false},
		{name: "nil lt nil", op: "lt", left: nil, right: nil, wantRes: false, wantErr: false},
		{name: "nil lte nil", op: "lte", left: nil, right: nil, wantRes: false, wantErr: false},

		// One nil (should error)
		{name: "nil vs int - error", op: "eq", left: nil, right: 10, wantRes: false, wantErr: true},
		{name: "int vs nil - error", op: "eq", left: 10, right: nil, wantRes: false, wantErr: true},
		{name: "nil vs string - error", op: "gt", left: nil, right: "hello", wantRes: false, wantErr: true},
		{name: "string vs nil - error", op: "lt", left: "hello", right: nil, wantRes: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CompareValues(tt.op, tt.left, tt.right)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if res != tt.wantRes {
				t.Errorf("expected result=%v, got %v", tt.wantRes, res)
			}
		})
	}
}

func TestCompareValues_TypeIncompatibility(t *testing.T) {
	tests := []struct {
		name  string
		op    string
		left  any
		right any
	}{
		{name: "string vs int", op: "eq", left: "hello", right: 10},
		{name: "int vs string", op: "gt", left: 10, right: "hello"},
		{name: "bool vs int", op: "eq", left: true, right: 10},
		{name: "slice vs int", op: "eq", left: []int{1, 2, 3}, right: 10},
		{name: "map vs string", op: "eq", left: map[string]int{"a": 1}, right: "hello"},
		{name: "time vs int", op: "gt", left: time.Now(), right: 10},
		{name: "int vs time", op: "lt", left: 10, right: time.Now()},
	}

	// All should return error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompareValues(tt.op, tt.left, tt.right)
			if err == nil {
				t.Error("expected incompatibility error, got nil")
			}
		})
	}
}

func TestCompareValues_InvalidOperator(t *testing.T) {
	tests := []struct {
		name  string
		op    string
		left  any
		right any
	}{
		{name: "invalid op - numeric", op: "invalid", left: 10, right: 20},
		{name: "invalid op - string", op: "unknown", left: "a", right: "b"},
		{name: "invalid op - time", op: "bad", left: time.Now(), right: time.Now()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompareValues(tt.op, tt.left, tt.right)
			if err == nil {
				t.Error("expected operator error, got nil")
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		wantNumeric bool
	}{
		{name: "int", value: 10, wantNumeric: true},
		{name: "int8", value: int8(10), wantNumeric: true},
		{name: "int16", value: int16(10), wantNumeric: true},
		{name: "int32", value: int32(10), wantNumeric: true},
		{name: "int64", value: int64(10), wantNumeric: true},
		{name: "uint", value: uint(10), wantNumeric: true},
		{name: "uint8", value: uint8(10), wantNumeric: true},
		{name: "uint16", value: uint16(10), wantNumeric: true},
		{name: "uint32", value: uint32(10), wantNumeric: true},
		{name: "uint64", value: uint64(10), wantNumeric: true},
		{name: "float32", value: float32(3.14), wantNumeric: true},
		{name: "float64", value: float64(3.14), wantNumeric: true},
		{name: "string", value: "hello", wantNumeric: false},
		{name: "bool", value: true, wantNumeric: false},
		{name: "slice", value: []int{1, 2, 3}, wantNumeric: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind := reflect.ValueOf(tt.value).Kind()
			result := isNumeric(kind)
			if result != tt.wantNumeric {
				t.Errorf("expected %v, got %v", tt.wantNumeric, result)
			}
		})
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		wantZero bool
	}{
		{name: "nil", value: nil, wantZero: true},
		{name: "int zero", value: 0, wantZero: true},
		{name: "int non-zero", value: 10, wantZero: false},
		{name: "string empty", value: "", wantZero: true},
		{name: "string non-empty", value: "hello", wantZero: false},
		{name: "bool false", value: false, wantZero: true},
		{name: "bool true", value: true, wantZero: false},
		{name: "float zero", value: 0.0, wantZero: true},
		{name: "float non-zero", value: 3.14, wantZero: false},
		{name: "slice nil", value: ([]int)(nil), wantZero: true},
		{name: "slice empty", value: []int{}, wantZero: false},
		{name: "slice non-empty", value: []int{1, 2}, wantZero: false},
		{name: "map nil", value: (map[string]int)(nil), wantZero: true},
		{name: "map empty", value: map[string]int{}, wantZero: false},
		{name: "time zero", value: time.Time{}, wantZero: true},
		{name: "time non-zero", value: time.Now(), wantZero: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsZeroValue(tt.value)
			if result != tt.wantZero {
				t.Errorf("expected %v, got %v", tt.wantZero, result)
			}
		})
	}
}

func TestIsNilValue(t *testing.T) {
	var nilPtr *int
	var nilSlice []int
	var nilMap map[string]int
	var nilChan chan int

	tests := []struct {
		name    string
		value   any
		wantNil bool
	}{
		{name: "nil", value: nil, wantNil: true},
		{name: "nil pointer", value: nilPtr, wantNil: true},
		{name: "nil slice", value: nilSlice, wantNil: true},
		{name: "nil map", value: nilMap, wantNil: true},
		{name: "nil channel", value: nilChan, wantNil: true},
		{name: "non-nil pointer", value: new(int), wantNil: false},
		{name: "empty slice", value: []int{}, wantNil: false},
		{name: "empty map", value: map[string]int{}, wantNil: false},
		{name: "int zero", value: 0, wantNil: false},
		{name: "string empty", value: "", wantNil: false},
		{name: "bool false", value: false, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNilValue(tt.value)
			if result != tt.wantNil {
				t.Errorf("expected %v, got %v", tt.wantNil, result)
			}
		})
	}
}

func TestIsTime(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		wantTime bool
	}{
		{name: "time.Time", value: time.Now(), wantTime: true},
		{name: "time.Time zero", value: time.Time{}, wantTime: true},
		{name: "int", value: 10, wantTime: false},
		{name: "string", value: "2024-01-01", wantTime: false},
		{name: "nil", value: nil, wantTime: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTime(tt.value)
			if result != tt.wantTime {
				t.Errorf("expected %v, got %v", tt.wantTime, result)
			}
		})
	}
}
