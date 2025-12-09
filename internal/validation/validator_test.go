package validation

import (
	"reflect"
	"testing"
	"time"

	"github.com/SmrutAI/Pedantigo/internal/constraints"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseTestTag is a test helper that parses pedantigo:"..." tags
func parseTestTag(tag reflect.StructTag) map[string]string {
	tagStr := tag.Get("pedantigo")
	if tagStr == "" {
		return nil
	}

	result := make(map[string]string)
	pairs := splitTagPairs(tagStr)
	for _, pair := range pairs {
		key, value := parseTagPair(pair)
		if key != "" {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// buildTestConstraints wraps constraints.BuildConstraints to return ConstraintValidator
func buildTestConstraints(constraintsMap map[string]string, fieldType reflect.Type) []ConstraintValidator {
	builtConstraints := constraints.BuildConstraints(constraintsMap, fieldType)
	validators := make([]ConstraintValidator, len(builtConstraints))
	for i, c := range builtConstraints {
		validators[i] = c
	}
	return validators
}

// splitTagPairs splits comma-separated tag pairs, handling values with commas
func splitTagPairs(tag string) []string {
	var pairs []string
	var current string
	inValue := false

	for i := 0; i < len(tag); i++ {
		c := tag[i]
		if c == '=' {
			inValue = true
			current += string(c)
		} else if c == ',' && !inValue {
			if current != "" {
				pairs = append(pairs, current)
				current = ""
			}
		} else {
			if c == ',' && inValue {
				// Check if this comma is really the end of a value
				// Simple heuristic: if next char is a letter, it's a new pair
				if i+1 < len(tag) && isLetter(tag[i+1]) {
					pairs = append(pairs, current)
					current = ""
					inValue = false
				} else {
					current += string(c)
				}
			} else {
				current += string(c)
				if c == ',' {
					inValue = false
				}
			}
		}
	}
	if current != "" {
		pairs = append(pairs, current)
	}
	return pairs
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// parseTagPair splits a single tag pair (e.g., "min=18" or "required")
func parseTagPair(pair string) (string, string) {
	pair = trimSpace(pair)
	if pair == "" {
		return "", ""
	}

	for i := 0; i < len(pair); i++ {
		if pair[i] == '=' {
			return pair[:i], pair[i+1:]
		}
	}
	return pair, ""
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

// TestValidate_FlatStruct tests validation of simple structs with multiple constraints
func TestValidate_FlatStruct(t *testing.T) {
	type Person struct {
		Name  string `pedantigo:"required,min=2"`
		Email string `pedantigo:"required,email"`
		Age   int    `pedantigo:"min=18,max=100"`
	}

	tests := []struct {
		name     string
		input    Person
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name:     "all valid",
			input:    Person{Name: "John", Email: "john@example.com", Age: 25},
			wantErrs: 0,
		},
		{
			name:     "name too short",
			input:    Person{Name: "J", Email: "john@example.com", Age: 25},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Name", errs[0].Field)
			},
		},
		{
			name:     "age too low",
			input:    Person{Name: "John", Email: "john@example.com", Age: 10},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Age", errs[0].Field)
			},
		},
		{
			name:     "age too high",
			input:    Person{Name: "John", Email: "john@example.com", Age: 150},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Age", errs[0].Field)
			},
		},
		{
			name:     "invalid email",
			input:    Person{Name: "John", Email: "invalid-email", Age: 25},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Email", errs[0].Field)
			},
		},
		{
			name:     "name and age both invalid",
			input:    Person{Name: "J", Email: "john@example.com", Age: 10},
			wantErrs: 2,
		},
		{
			name:     "all fields invalid",
			input:    Person{Name: "J", Email: "bad-email", Age: 200},
			wantErrs: 3,
		},
		{
			name:     "empty name (required)",
			input:    Person{Name: "", Email: "john@example.com", Age: 25},
			wantErrs: 1, // Fails min=2 (empty string has length 0)
		},
		{
			name:     "empty email (required)",
			input:    Person{Name: "John", Email: "", Age: 25},
			wantErrs: 0, // 'required' only checked during Unmarshal, not Validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_NestedStruct tests recursive validation of nested structs
func TestValidate_NestedStruct(t *testing.T) {
	type Address struct {
		City    string `pedantigo:"required,min=2"`
		Country string `pedantigo:"required"`
	}

	type User struct {
		Name    string  `pedantigo:"required,min=2"`
		Address Address `pedantigo:"required"`
	}

	tests := []struct {
		name     string
		input    User
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name: "all valid",
			input: User{
				Name: "John",
				Address: Address{
					City:    "New York",
					Country: "USA",
				},
			},
			wantErrs: 0,
		},
		{
			name: "nested city too short",
			input: User{
				Name: "John",
				Address: Address{
					City:    "X",
					Country: "USA",
				},
			},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Address.City", errs[0].Field)
			},
		},
		{
			name: "nested country empty (required)",
			input: User{
				Name: "John",
				Address: Address{
					City:    "New York",
					Country: "",
				},
			},
			wantErrs: 0, // 'required' only checked during Unmarshal, not Validate
			errCheck: func(t *testing.T, errs []FieldError) {
				if len(errs) > 0 {
					assert.NotEqual(t, "Address.Country", errs[0].Field)
				}
			},
		},
		{
			name: "top-level name too short",
			input: User{
				Name: "J",
				Address: Address{
					City:    "New York",
					Country: "USA",
				},
			},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Name", errs[0].Field)
			},
		},
		{
			name: "both top-level and nested invalid",
			input: User{
				Name: "J",
				Address: Address{
					City:    "X",
					Country: "",
				},
			},
			wantErrs: 2, // Name + City (Country 'required' not checked in Validate)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateFunc func(val reflect.Value, path string) []FieldError
			validateFunc = func(val reflect.Value, path string) []FieldError {
				return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, validateFunc)
			}

			errs := validateFunc(reflect.ValueOf(tt.input), "")

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_Slices tests validation of slice elements
func TestValidate_Slices(t *testing.T) {
	type Team struct {
		Members []string `pedantigo:"min=1"`
	}

	tests := []struct {
		name     string
		input    Team
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name:     "all elements valid",
			input:    Team{Members: []string{"Alice", "Bob", "Charlie"}},
			wantErrs: 0,
		},
		{
			name:     "one element too short",
			input:    Team{Members: []string{"Alice", "B", "Charlie"}},
			wantErrs: 0, // "B" has length 1, satisfies min=1
			errCheck: func(t *testing.T, errs []FieldError) {
				if len(errs) > 0 {
					assert.Equal(t, "Members[1]", errs[0].Field)
				}
			},
		},
		{
			name:     "multiple elements too short",
			input:    Team{Members: []string{"Alice", "B", "C"}},
			wantErrs: 0, // "B" and "C" have length 1, satisfy min=1
		},
		{
			name:     "empty slice",
			input:    Team{Members: []string{}},
			wantErrs: 0,
		},
		{
			name:     "nil slice",
			input:    Team{Members: nil},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_SliceOfStructs tests validation of slice elements that are structs
func TestValidate_SliceOfStructs(t *testing.T) {
	type Item struct {
		Name string `pedantigo:"required,min=2"`
	}

	type List struct {
		Items []Item `pedantigo:""`
	}

	tests := []struct {
		name     string
		input    List
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name: "all items valid",
			input: List{Items: []Item{
				{Name: "Item1"},
				{Name: "Item2"},
			}},
			wantErrs: 0,
		},
		{
			name: "one item invalid",
			input: List{Items: []Item{
				{Name: "Item1"},
				{Name: "I"},
			}},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Items[1].Name", errs[0].Field)
			},
		},
		{
			name: "multiple items invalid",
			input: List{Items: []Item{
				{Name: "X"},
				{Name: "Item2"},
				{Name: ""},
			}},
			wantErrs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateFunc func(val reflect.Value, path string) []FieldError
			validateFunc = func(val reflect.Value, path string) []FieldError {
				return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, validateFunc)
			}

			errs := validateFunc(reflect.ValueOf(tt.input), "")

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_Maps tests validation of map values
func TestValidate_Maps(t *testing.T) {
	type Scores struct {
		Values map[string]int `pedantigo:"min=0,max=100"`
	}

	tests := []struct {
		name     string
		input    Scores
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name:     "all values valid",
			input:    Scores{Values: map[string]int{"Alice": 95, "Bob": 87}},
			wantErrs: 0,
		},
		{
			name:     "one value too low",
			input:    Scores{Values: map[string]int{"Alice": 95, "Bob": -5}},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Values[Bob]", errs[0].Field)
			},
		},
		{
			name:     "one value too high",
			input:    Scores{Values: map[string]int{"Alice": 150, "Bob": 87}},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Values[Alice]", errs[0].Field)
			},
		},
		{
			name:     "multiple values invalid",
			input:    Scores{Values: map[string]int{"Alice": 150, "Bob": -5, "Charlie": 90}},
			wantErrs: 2,
		},
		{
			name:     "empty map",
			input:    Scores{Values: map[string]int{}},
			wantErrs: 0,
		},
		{
			name:     "nil map",
			input:    Scores{Values: nil},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_MapOfStructs tests validation of map values that are structs
func TestValidate_MapOfStructs(t *testing.T) {
	type Entry struct {
		Value string `pedantigo:"required,min=2"`
	}

	type Registry struct {
		Entries map[string]Entry `pedantigo:""`
	}

	tests := []struct {
		name     string
		input    Registry
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name: "all entries valid",
			input: Registry{Entries: map[string]Entry{
				"key1": {Value: "Value1"},
				"key2": {Value: "Value2"},
			}},
			wantErrs: 0,
		},
		{
			name: "one entry invalid",
			input: Registry{Entries: map[string]Entry{
				"key1": {Value: "Value1"},
				"key2": {Value: "X"},
			}},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Entries[key2].Value", errs[0].Field)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateFunc func(val reflect.Value, path string) []FieldError
			validateFunc = func(val reflect.Value, path string) []FieldError {
				return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, validateFunc)
			}

			errs := validateFunc(reflect.ValueOf(tt.input), "")

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_ErrorAggregation tests that all validation errors are collected
func TestValidate_ErrorAggregation(t *testing.T) {
	type Person struct {
		FirstName string `pedantigo:"required,min=2"`
		LastName  string `pedantigo:"required,min=2"`
		Email     string `pedantigo:"required,email"`
		Age       int    `pedantigo:"min=18,max=100"`
	}

	tests := []struct {
		name      string
		input     Person
		wantCount int
		errCheck  func(t *testing.T, errs []FieldError)
	}{
		{
			name:      "all valid",
			input:     Person{FirstName: "John", LastName: "Doe", Email: "john@example.com", Age: 30},
			wantCount: 0,
		},
		{
			name:      "single error",
			input:     Person{FirstName: "John", LastName: "Doe", Email: "john@example.com", Age: 10},
			wantCount: 1,
		},
		{
			name:      "multiple errors - 2 fields",
			input:     Person{FirstName: "J", LastName: "Doe", Email: "john@example.com", Age: 10},
			wantCount: 2,
		},
		{
			name:      "multiple errors - 3 fields",
			input:     Person{FirstName: "J", LastName: "D", Email: "john@example.com", Age: 10},
			wantCount: 3,
		},
		{
			name:      "multiple errors - all fields",
			input:     Person{FirstName: "X", LastName: "Y", Email: "bad", Age: 150},
			wantCount: 4,
		},
		{
			name: "error count verification",
			input: Person{
				FirstName: "A",
				LastName:  "B",
				Email:     "invalid",
				Age:       200,
			},
			wantCount: 4,
			errCheck: func(t *testing.T, errs []FieldError) {
				assert.Len(t, errs, 4)
				// Verify we collected errors from all 4 fields
				fields := make(map[string]bool)
				for _, err := range errs {
					fields[err.Field] = true
				}
				expectedFields := map[string]bool{
					"FirstName": true,
					"LastName":  true,
					"Email":     true,
					"Age":       true,
				}
				for field := range expectedFields {
					assert.True(t, fields[field], "missing error for field %s", field)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantCount)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_CrossFieldValidation tests time range validation with constraints
func TestValidate_CrossFieldValidation(t *testing.T) {
	type DateRange struct {
		Start time.Time `pedantigo:"required"`
		End   time.Time `pedantigo:"required"`
	}

	tests := []struct {
		name         string
		input        DateRange
		wantFieldErr int
		errCheck     func(t *testing.T, errs []FieldError)
	}{
		{
			name: "valid date range with both fields set",
			input: DateRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			wantFieldErr: 0,
		},
		{
			name: "zero Start time (zero value for time.Time)",
			input: DateRange{
				Start: time.Time{},
				End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantFieldErr: 0,
			// Note: Zero times are not caught by standard field constraints
			// Cross-field validation would be handled at a different layer
		},
		{
			name: "both dates with valid constraints",
			input: DateRange{
				Start: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC),
			},
			wantFieldErr: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantFieldErr)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_NestedCollections tests deeply nested slices and maps
func TestValidate_NestedCollections(t *testing.T) {
	type Config struct {
		Name  string `pedantigo:"required,min=2"`
		Value string `pedantigo:"min=1"`
	}

	type Settings struct {
		Configs []Config `pedantigo:""`
	}

	tests := []struct {
		name     string
		input    Settings
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name: "all configs valid",
			input: Settings{
				Configs: []Config{
					{Name: "config1", Value: "value1"},
					{Name: "config2", Value: "value2"},
				},
			},
			wantErrs: 0,
		},
		{
			name: "first config invalid",
			input: Settings{
				Configs: []Config{
					{Name: "X", Value: "value1"},
					{Name: "config2", Value: "value2"},
				},
			},
			wantErrs: 1,
			errCheck: func(t *testing.T, errs []FieldError) {
				require.Greater(t, len(errs), 0)
				assert.Equal(t, "Configs[0].Name", errs[0].Field)
			},
		},
		{
			name: "multiple config errors",
			input: Settings{
				Configs: []Config{
					{Name: "X", Value: ""},
					{Name: "Y", Value: "value2"},
					{Name: "config3", Value: ""},
				},
			},
			wantErrs: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateFunc func(val reflect.Value, path string) []FieldError
			validateFunc = func(val reflect.Value, path string) []FieldError {
				return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, validateFunc)
			}

			errs := validateFunc(reflect.ValueOf(tt.input), "")

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_PointerFields tests validation of pointer fields
func TestValidate_PointerFields(t *testing.T) {
	type User struct {
		Name  string  `pedantigo:"required,min=2"`
		Email *string `pedantigo:"email"`
	}

	tests := []struct {
		name     string
		input    User
		wantErrs int
		errCheck func(t *testing.T, errs []FieldError)
	}{
		{
			name:     "nil pointer - no error",
			input:    User{Name: "John", Email: nil},
			wantErrs: 1, // email constraint rejects pointer types (requires string)
		},
		{
			name: "valid pointer value",
			input: func() User {
				email := "john@example.com"
				return User{Name: "John", Email: &email}
			}(),
			wantErrs: 1, // email constraint rejects pointer types (requires string)
		},
		{
			name: "invalid pointer value",
			input: func() User {
				email := "invalid"
				return User{Name: "John", Email: &email}
			}(),
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)

			if tt.errCheck != nil {
				tt.errCheck(t, errs)
			}
		})
	}
}

// TestValidate_ConstraintTypes tests various constraint types
func TestValidate_ConstraintTypes(t *testing.T) {
	type Record struct {
		Age    int    `pedantigo:"min=0,max=150"`
		Email  string `pedantigo:"email"`
		URL    string `pedantigo:"url"`
		UUID   string `pedantigo:"uuid"`
		IPv4   string `pedantigo:"ipv4"`
		Status string `pedantigo:"oneof=active inactive pending"`
	}

	tests := []struct {
		name     string
		input    Record
		wantErrs int
	}{
		{
			name: "all valid",
			input: Record{
				Age:    25,
				Email:  "test@example.com",
				URL:    "https://example.com",
				UUID:   "550e8400-e29b-41d4-a716-446655440000",
				IPv4:   "192.168.1.1",
				Status: "active",
			},
			wantErrs: 0,
		},
		{
			name: "multiple constraints fail",
			input: Record{
				Age:    200,
				Email:  "invalid",
				URL:    "not a url",
				UUID:   "not-a-uuid",
				IPv4:   "not-an-ip",
				Status: "unknown",
			},
			wantErrs: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)
		})
	}
}

// TestValidate_UnexportedFields tests that unexported fields are skipped
func TestValidate_UnexportedFields(t *testing.T) {
	type Data struct {
		Public  string `pedantigo:"required,min=2"`
		private string `pedantigo:"required,min=2"` // Should be ignored
	}

	tests := []struct {
		name     string
		input    Data
		wantErrs int
	}{
		{
			name:     "valid public, invalid private (ignored)",
			input:    Data{Public: "Valid", private: "X"},
			wantErrs: 0,
		},
		{
			name:     "invalid public, valid private (ignored)",
			input:    Data{Public: "X", private: "Valid"},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)
		})
	}
}

// TestValidate_NoConstraints tests fields without validation constraints
func TestValidate_NoConstraints(t *testing.T) {
	type Document struct {
		ID    string
		Title string `pedantigo:"required,min=2"`
		Body  string
	}

	tests := []struct {
		name     string
		input    Document
		wantErrs int
	}{
		{
			name:     "valid title, other fields unconstrained",
			input:    Document{ID: "", Title: "Valid", Body: ""},
			wantErrs: 0,
		},
		{
			name:     "invalid title",
			input:    Document{ID: "", Title: "X", Body: ""},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)
		})
	}
}

// TestValidateNestedElements tests the validateNestedElements helper function
func TestValidateNestedElements(t *testing.T) {
	type Item struct {
		Name string `pedantigo:"required,min=2"`
	}

	type Container struct {
		Items  []Item          `pedantigo:""`
		Lookup map[string]Item `pedantigo:""`
	}

	tests := []struct {
		name           string
		input          Container
		wantNestedErrs int
	}{
		{
			name: "valid nested items",
			input: Container{
				Items:  []Item{{Name: "Item1"}, {Name: "Item2"}},
				Lookup: map[string]Item{"a": {Name: "ItemA"}},
			},
			wantNestedErrs: 0,
		},
		{
			name: "invalid item in slice",
			input: Container{
				Items: []Item{{Name: "Item1"}, {Name: "X"}},
			},
			wantNestedErrs: 1,
		},
		{
			name: "invalid item in map",
			input: Container{
				Lookup: map[string]Item{"a": {Name: "X"}},
			},
			wantNestedErrs: 1,
		},
		{
			name: "multiple invalid items",
			input: Container{
				Items:  []Item{{Name: "X"}, {Name: "Y"}},
				Lookup: map[string]Item{"a": {Name: "Z"}},
			},
			wantNestedErrs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateFunc func(val reflect.Value, path string) []FieldError
			validateFunc = func(val reflect.Value, path string) []FieldError {
				return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, validateFunc)
			}

			errs := validateFunc(reflect.ValueOf(tt.input), "")

			assert.Len(t, errs, tt.wantNestedErrs)
		})
	}
}

// TestValidate_EmptyStruct tests validation of empty structs
func TestValidate_EmptyStruct(t *testing.T) {
	type Empty struct{}

	tests := []struct {
		name     string
		input    Empty
		wantErrs int
	}{
		{
			name:     "empty struct",
			input:    Empty{},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateValue(
				reflect.ValueOf(tt.input),
				"",
				false,
				parseTestTag,
				buildTestConstraints,
				func(val reflect.Value, path string) []FieldError {
					return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, nil)
				},
			)

			assert.Len(t, errs, tt.wantErrs)
		})
	}
}

// TestValidate_PathConstruction tests that field paths are constructed correctly
func TestValidate_PathConstruction(t *testing.T) {
	type Level3 struct {
		Value string `pedantigo:"required,min=2"`
	}

	type Level2 struct {
		Data Level3 `pedantigo:"required"`
	}

	type Level1 struct {
		Nested Level2 `pedantigo:"required"`
	}

	tests := []struct {
		name     string
		input    Level1
		wantPath string
	}{
		{
			name: "deeply nested path",
			input: Level1{
				Nested: Level2{
					Data: Level3{Value: "X"},
				},
			},
			wantPath: "Nested.Data.Value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateFunc func(val reflect.Value, path string) []FieldError
			validateFunc = func(val reflect.Value, path string) []FieldError {
				return ValidateValue(val, path, false, parseTestTag, buildTestConstraints, validateFunc)
			}

			errs := validateFunc(reflect.ValueOf(tt.input), "")

			require.Greater(t, len(errs), 0)
			assert.Equal(t, tt.wantPath, errs[0].Field)
		})
	}
}
