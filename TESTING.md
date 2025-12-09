# Testing Guide for Pedantigo Contributors

This guide outlines the testing standards and best practices for the Pedantigo project.

## Table of Contents

1. [Table-Driven Tests](#table-driven-tests)
2. [Test Structure](#test-structure)
3. [Test Organization](#test-organization)
4. [Naming Conventions](#naming-conventions)
5. [Coverage Requirements](#coverage-requirements)

---

## Table-Driven Tests

Table-driven tests are the standard pattern in Go. They reduce code duplication and make it easy to add new test cases.

### Basic Pattern

```go
func TestMin(t *testing.T) {
	type MinTest struct {
		Age int `pedantigo:"min=18"`
	}

	tests := []struct {
		name      string
		data      *MinTest
		expectErr bool
	}{
		{
			name:      "above minimum - valid",
			data:      &MinTest{Age: 25},
			expectErr: false,
		},
		{
			name:      "at minimum - valid",
			data:      &MinTest{Age: 18},
			expectErr: false,
		},
		{
			name:      "below minimum - invalid",
			data:      &MinTest{Age: 15},
			expectErr: true,
		},
	}

	validator := New[MinTest]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.expectErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
```

### Key Principles

1. **Define structs once** - Test structs are defined outside the test table
2. **Data in table** - Table entries contain test inputs and expected results
3. **Single loop** - One test execution loop handles all cases
4. **Use t.Run()** - Always use subtests with descriptive names

### Multiple Struct Types

When testing multiple types, use type switches:

```go
func TestCrossFieldConstraints(t *testing.T) {
	type StructA struct {
		Field1 string `pedantigo:"required"`
		Field2 string `pedantigo:"eqfield=Field1"`
	}

	type StructB struct {
		Age    int `pedantigo:"required"`
		MinAge int `pedantigo:"ltfield=Age"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
	}{
		{
			name:      "struct A - fields equal",
			validator: New[StructA](),
			data:      &StructA{Field1: "test", Field2: "test"},
			expectErr: false,
		},
		{
			name:      "struct B - age validation",
			validator: New[StructB](),
			data:      &StructB{Age: 25, MinAge: 18},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *Validator[StructA]:
				err := v.Validate(tt.data.(*StructA))
				// Check error...
			case *Validator[StructB]:
				err := v.Validate(tt.data.(*StructB))
				// Check error...
			}
		})
	}
}
```

---

## Test Structure

### Standard Test Layout

```go
func TestFeatureName(t *testing.T) {
	// 1. Define test structs
	type TestStruct struct {
		Field1 string `pedantigo:"required"`
		Field2 int    `pedantigo:"min=0"`
	}

	// 2. Create reusable validators (optional)
	validator := New[TestStruct]()

	// 3. Define test cases
	tests := []struct {
		name      string
		data      *TestStruct
		expectErr bool
		errField  string  // For validating specific field errors
	}{
		{
			name: "valid data",
			data: &TestStruct{
				Field1: "value",
				Field2: 10,
			},
			expectErr: false,
		},
		{
			name: "missing required field",
			data: &TestStruct{
				Field1: "",
				Field2: 10,
			},
			expectErr: true,
			errField:  "Field1",
		},
	}

	// 4. Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)

			// Check error existence
			if tt.expectErr && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			// Verify specific error field
			if tt.expectErr && err != nil {
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errField {
						foundError = true
						break
					}
				}
				if !foundError {
					t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
				}
			}
		})
	}
}
```

### Error Validation

Always validate error details, not just error existence:

```go
// ✅ Good: Validates specific error
if tt.expectErr {
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == tt.errField {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
	}
}

// ❌ Bad: Only checks if error exists
if err == nil {
	t.Error("expected error")
}
```

---

## Test Organization

### File Structure

```
internal/constraints/
├── constraints_test.go              # Basic constraints (min, max, email)
├── collections_test.go              # Collection constraints (unique, dive)
├── crossfield_comparison_test.go    # Comparison (eqfield, gtfield, ltfield)
├── crossfield_excluded_test.go      # Exclusion (excluded_if, excluded_unless)
├── crossfield_required_test.go      # Required (required_if, required_unless)
└── crossfield_edge_cases_test.go    # Edge cases for crossfield constraints
```

### When to Split Files

Consider splitting test files when:

- File exceeds ~1000 lines with logical groupings
- Testing distinct feature sets that don't interact
- Different constraint categories (basic vs crossfield vs collections)

Keep files together if:
- Tests share common struct definitions
- Features are tightly coupled
- File is manageable (<600 lines)

---

## Naming Conventions

### Test Functions

Format: `TestFeatureName_Scenario` (optional scenario suffix)

```go
// Good names
TestTagParser
TestValidation_RequiredField
TestUnmarshal_MissingField
TestCrossField_TypeIncompatibility

// Avoid
TestStuff
TestCase1
Test1
```

### Test Cases

Format: `"description - expected outcome"`

```go
tests := []struct {
	name string
	// ...
}{
	{name: "valid input - pass", /* ... */},
	{name: "missing required field - error", /* ... */},
	{name: "zero value - pass", /* ... */},
	{name: "nil pointer - error", /* ... */},
}
```

### Test Structs

Use clear, domain-relevant names:

```go
// Good
type Payment struct { ... }
type UserAccount struct { ... }
type Document struct { ... }

// Avoid
type TestStruct struct { ... }
type Foo struct { ... }
type Data struct { ... }
```

---

## Coverage Requirements

### Targets

- **Overall:** 85% minimum coverage
- **Critical paths:** 100% coverage (validation core, constraint checks)
- **Edge cases:** Must be covered (nil, zero values, type errors)

### Running Coverage

```bash
# Run tests with coverage report (generates coverage.html)
make test-coverage
```

### What to Test

**Essential coverage:**

1. **Happy path** - Valid inputs
2. **Validation errors** - Invalid inputs
3. **Edge cases:**
   - Zero values (0, false, "", nil)
   - Nil pointers
   - Empty collections
   - Boundary values
4. **Type errors** - Incompatible types
5. **Missing fields** - Required fields absent
6. **Cross-field behavior** - Field interactions

**Don't over-test:**

- Standard library behavior (json.Marshal, etc.)
- External dependencies (unless integration testing)
- Multiple test cases for the same code path

### Example: Good Coverage

```go
func TestValidation(t *testing.T) {
	tests := []struct{
		name string
		data interface{}
		expectErr bool
	}{
		{name: "valid input", data: valid, expectErr: false},              // Happy path
		{name: "missing required", data: missing, expectErr: true},        // Validation
		{name: "zero value", data: zero, expectErr: false},               // Edge case
		{name: "nil pointer", data: nil, expectErr: true},                // Edge case
		{name: "type mismatch", data: wrongType, expectErr: true},        // Type error
	}
	// ...
}
```

---

## Quick Reference

### Test Checklist

- [ ] Using table-driven test pattern
- [ ] Test structs defined outside table
- [ ] Using `t.Run()` for subtests
- [ ] Descriptive test names (function and cases)
- [ ] Testing happy path, errors, edge cases
- [ ] Validating specific error fields
- [ ] Coverage >85% (`make test-coverage`)
- [ ] All tests pass (`make test`)

### Commands

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run with coverage (generates coverage.html)
make test-coverage

# Run benchmarks
make bench

# Format code
make fmt

# Run all checks (fmt, vet, test)
make all
```

### Performance Tips

```go
// Cache validators outside loops
validator := New[User]()
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		err := validator.Validate(tt.data)
		// ...
	})
}

// Skip slow tests in short mode
func TestExpensive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	// Expensive operation...
}
```

---

## Examples from Codebase

### Basic Constraint

See `internal/constraints/constraints_test.go`:

```go
func TestEmail(t *testing.T) {
	type EmailTest struct {
		Email string `pedantigo:"email"`
	}

	tests := []struct {
		name      string
		email     string
		expectErr bool
	}{
		{name: "valid email", email: "user@example.com", expectErr: false},
		{name: "invalid email", email: "not-an-email", expectErr: true},
		{name: "empty string", email: "", expectErr: false},
	}

	validator := New[EmailTest]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&EmailTest{Email: tt.email})
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
```

### Crossfield Constraint

See `internal/constraints/crossfield_comparison_test.go`:

```go
func TestEqField(t *testing.T) {
	type PasswordConfirm struct {
		Password        string `pedantigo:"required"`
		PasswordConfirm string `pedantigo:"eqfield=Password"`
	}

	tests := []struct {
		name      string
		data      *PasswordConfirm
		expectErr bool
		errField  string
	}{
		{
			name: "passwords match",
			data: &PasswordConfirm{
				Password:        "secret",
				PasswordConfirm: "secret",
			},
			expectErr: false,
		},
		{
			name: "passwords differ",
			data: &PasswordConfirm{
				Password:        "secret",
				PasswordConfirm: "different",
			},
			expectErr: true,
			errField:  "PasswordConfirm",
		},
	}

	validator := New[PasswordConfirm]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.expectErr && err == nil {
				t.Error("expected error")
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.expectErr {
				ve := err.(*ValidationError)
				found := false
				for _, e := range ve.Errors {
					if e.Field == tt.errField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for %s", tt.errField)
				}
			}
		})
	}
}
```

---

## Getting Help

- Review existing tests in `internal/constraints/` for patterns
- Check `CLAUDE.md` for project-specific guidelines
- Run `make test-verbose` for detailed output
- Use `make test-coverage` to check coverage

**Remember:** Tests are documentation. Write tests that clearly demonstrate how features work.
