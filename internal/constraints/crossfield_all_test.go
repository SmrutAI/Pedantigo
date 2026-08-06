package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	pedantigo "github.com/SmrutAI/pedantigo"
)

// ============================================================================
// required_with_all Tests
// ============================================================================

func TestRequiredWithAll(t *testing.T) {
	type Form struct {
		A      string `json:"a"`
		B      string `json:"b"`
		C      string `json:"c"`
		Target string `json:"target" validate:"required_with_all=A B C"`
	}

	// Test 1: All present, target present - pass
	t.Run("all present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: "value"})
		require.NoError(t, err)
	})

	// Test 2: All present, target absent - fail
	t.Run("all present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: ""})
		require.Error(t, err)
	})

	// Test 3: One absent, target absent - pass (not all present)
	t.Run("one absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "3", Target: ""})
		require.NoError(t, err)
	})

	// Test 4: All absent, target absent - pass
	t.Run("all absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 5: Two fields version
	t.Run("two fields both present", func(t *testing.T) {
		type Form2 struct {
			A      string `json:"a"`
			B      string `json:"b"`
			Target string `json:"target" validate:"required_with_all=A B"`
		}
		validator := pedantigo.New[Form2]()
		err := validator.Validate(&Form2{A: "1", B: "2", Target: ""})
		require.Error(t, err)
	})

	// Test 6: Two of three present, target absent - pass
	t.Run("two of three present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 7: All present, target present with value - pass
	t.Run("all present target present with value", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "x", B: "y", C: "z", Target: "required value"})
		require.NoError(t, err)
	})

	// Test 8: One field absent (first), target absent - pass
	t.Run("first field absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "2", C: "3", Target: ""})
		require.NoError(t, err)
	})

	// Test 9: One field absent (last), target absent - pass
	t.Run("last field absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "", Target: ""})
		require.NoError(t, err)
	})
}

// ============================================================================
// required_without_all Tests
// ============================================================================

func TestRequiredWithoutAll(t *testing.T) {
	type Form struct {
		A      string `json:"a"`
		B      string `json:"b"`
		C      string `json:"c"`
		Target string `json:"target" validate:"required_without_all=A B C"`
	}

	// Test 1: All absent, target present - pass
	t.Run("all absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 2: All absent, target absent - fail
	t.Run("all absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: ""})
		require.Error(t, err)
	})

	// Test 3: One present, target absent - pass (not all absent)
	t.Run("one present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 4: All present, target absent - pass
	t.Run("all present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: ""})
		require.NoError(t, err)
	})

	// Test 5: Two fields version
	t.Run("two fields both absent", func(t *testing.T) {
		type Form2 struct {
			A      string `json:"a"`
			B      string `json:"b"`
			Target string `json:"target" validate:"required_without_all=A B"`
		}
		validator := pedantigo.New[Form2]()
		err := validator.Validate(&Form2{A: "", B: "", Target: ""})
		require.Error(t, err)
	})

	// Test 6: Two of three absent, target absent - pass
	t.Run("two of three absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 7: All absent, target present with value - pass
	t.Run("all absent target present with value", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: "required value"})
		require.NoError(t, err)
	})

	// Test 8: One field present (first), target absent - pass
	t.Run("first field present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 9: One field present (middle), target absent - pass
	t.Run("middle field present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "2", C: "", Target: ""})
		require.NoError(t, err)
	})
}

// ============================================================================
// excluded_with_all Tests
// ============================================================================

func TestExcludedWithAll(t *testing.T) {
	type Form struct {
		A      string `json:"a"`
		B      string `json:"b"`
		C      string `json:"c"`
		Target string `json:"target" validate:"excluded_with_all=A B C"`
	}

	// Test 1: All present, target present - fail
	t.Run("all present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: "value"})
		require.Error(t, err)
	})

	// Test 2: All present, target absent - pass
	t.Run("all present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: ""})
		require.NoError(t, err)
	})

	// Test 3: One absent, target present - pass (not all present)
	t.Run("one absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "3", Target: "value"})
		require.NoError(t, err)
	})

	// Test 4: All absent, target present - pass
	t.Run("all absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 5: Two fields version
	t.Run("two fields both present", func(t *testing.T) {
		type Form2 struct {
			A      string `json:"a"`
			B      string `json:"b"`
			Target string `json:"target" validate:"excluded_with_all=A B"`
		}
		validator := pedantigo.New[Form2]()
		err := validator.Validate(&Form2{A: "1", B: "2", Target: "value"})
		require.Error(t, err)
	})

	// Test 6: Two of three present, target present - pass
	t.Run("two of three present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 7: All present, target absent - pass
	t.Run("all present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "x", B: "y", C: "z", Target: ""})
		require.NoError(t, err)
	})

	// Test 8: One field absent (first), target present - pass
	t.Run("first field absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "2", C: "3", Target: "value"})
		require.NoError(t, err)
	})

	// Test 9: One field absent (last), target present - pass
	t.Run("last field absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 10: All absent, target absent - pass
	t.Run("all absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})
}

// ============================================================================
// excluded_without_all Tests
// ============================================================================

func TestExcludedWithoutAll(t *testing.T) {
	type Form struct {
		A      string `json:"a"`
		B      string `json:"b"`
		C      string `json:"c"`
		Target string `json:"target" validate:"excluded_without_all=A B C"`
	}

	// Test 1: All absent, target present - fail
	t.Run("all absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: "value"})
		require.Error(t, err)
	})

	// Test 2: All absent, target absent - pass
	t.Run("all absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 3: One present, target present - pass (not all absent)
	t.Run("one present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 4: All present, target present - pass
	t.Run("all present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: "value"})
		require.NoError(t, err)
	})

	// Test 5: Two fields version
	t.Run("two fields both absent", func(t *testing.T) {
		type Form2 struct {
			A      string `json:"a"`
			B      string `json:"b"`
			Target string `json:"target" validate:"excluded_without_all=A B"`
		}
		validator := pedantigo.New[Form2]()
		err := validator.Validate(&Form2{A: "", B: "", Target: "value"})
		require.Error(t, err)
	})

	// Test 6: Two of three absent, target present - pass
	t.Run("two of three absent target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 7: All absent, target absent - pass
	t.Run("all absent target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "", C: "", Target: ""})
		require.NoError(t, err)
	})

	// Test 8: One field present (first), target present - pass
	t.Run("first field present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 9: One field present (middle), target present - pass
	t.Run("middle field present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", B: "2", C: "", Target: "value"})
		require.NoError(t, err)
	})

	// Test 10: All present, target absent - pass
	t.Run("all present target absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "1", B: "2", C: "3", Target: ""})
		require.NoError(t, err)
	})
}

// ============================================================================
// Edge Cases and Integration
// ============================================================================

func TestAllConstraintsWithIntegers(t *testing.T) {
	type Form struct {
		Count1 int    `json:"count1"`
		Count2 int    `json:"count2"`
		Target string `json:"target" validate:"required_with_all=Count1 Count2"`
	}

	// Zero is considered absent for integers
	t.Run("int zero is absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Count1: 0, Count2: 5, Target: ""})
		require.NoError(t, err) // One is zero, so not all present
	})

	t.Run("int both nonzero", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Count1: 3, Count2: 5, Target: ""})
		require.Error(t, err) // Both present, target required
	})

	t.Run("int both zero", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Count1: 0, Count2: 0, Target: ""})
		require.NoError(t, err) // Both absent
	})

	t.Run("int negative values are present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Count1: -1, Count2: -5, Target: ""})
		require.Error(t, err) // Negative values are still "present"
	})
}

func TestAllConstraintsWithBooleans(t *testing.T) {
	type Form struct {
		Flag1  bool   `json:"flag1"`
		Flag2  bool   `json:"flag2"`
		Target string `json:"target" validate:"required_with_all=Flag1 Flag2"`
	}

	// false is considered absent for booleans
	t.Run("bool false is absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Flag1: false, Flag2: true, Target: ""})
		require.NoError(t, err)
	})

	t.Run("bool both true", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Flag1: true, Flag2: true, Target: ""})
		require.Error(t, err)
	})

	t.Run("bool both false", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Flag1: false, Flag2: false, Target: ""})
		require.NoError(t, err)
	})

	t.Run("bool mixed values", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Flag1: true, Flag2: false, Target: ""})
		require.NoError(t, err) // Not all present
	})
}

func TestAllConstraintsWithFloats(t *testing.T) {
	type Form struct {
		Price1 float64 `json:"price1"`
		Price2 float64 `json:"price2"`
		Target string  `json:"target" validate:"required_with_all=Price1 Price2"`
	}

	t.Run("float zero is absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Price1: 0.0, Price2: 99.99, Target: ""})
		require.NoError(t, err)
	})

	t.Run("float both nonzero", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Price1: 10.50, Price2: 20.75, Target: ""})
		require.Error(t, err)
	})

	t.Run("float negative values are present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{Price1: -5.5, Price2: 10.0, Target: ""})
		require.Error(t, err) // Both present (negative counts as present)
	})
}

func TestAllConstraintsMixedTypes(t *testing.T) {
	type Form struct {
		StringField string  `json:"string_field"`
		IntField    int     `json:"int_field"`
		BoolField   bool    `json:"bool_field"`
		FloatField  float64 `json:"float_field"`
		Target      string  `json:"target" validate:"required_with_all=StringField IntField BoolField FloatField"`
	}

	t.Run("mixed all present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			StringField: "value",
			IntField:    42,
			BoolField:   true,
			FloatField:  3.14,
			Target:      "",
		})
		require.Error(t, err) // All present, target required
	})

	t.Run("mixed one absent string", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			StringField: "",
			IntField:    42,
			BoolField:   true,
			FloatField:  3.14,
			Target:      "",
		})
		require.NoError(t, err) // String absent, not all present
	})

	t.Run("mixed one absent int", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			StringField: "value",
			IntField:    0,
			BoolField:   true,
			FloatField:  3.14,
			Target:      "",
		})
		require.NoError(t, err) // Int zero (absent), not all present
	})

	t.Run("mixed one absent bool", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			StringField: "value",
			IntField:    42,
			BoolField:   false,
			FloatField:  3.14,
			Target:      "",
		})
		require.NoError(t, err) // Bool false (absent), not all present
	})

	t.Run("mixed all absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			StringField: "",
			IntField:    0,
			BoolField:   false,
			FloatField:  0.0,
			Target:      "",
		})
		require.NoError(t, err) // All absent
	})
}

func TestAllConstraintsSingleField(t *testing.T) {
	type Form struct {
		A      string `json:"a"`
		Target string `json:"target" validate:"required_with_all=A"`
	}

	t.Run("single field present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "value", Target: ""})
		require.Error(t, err) // Single field present, target required
	})

	t.Run("single field absent", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "", Target: ""})
		require.NoError(t, err) // Field absent
	})

	t.Run("single field present target present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: "value", Target: "target_value"})
		require.NoError(t, err)
	})
}

func TestExcludedWithAllComplexScenario(t *testing.T) {
	type PaymentForm struct {
		CreditCard string `json:"credit_card"`
		CVV        string `json:"cvv"`
		ExpiryDate string `json:"expiry_date"`
		// CashAmount should not be present if all credit card fields are present
		CashAmount string `json:"cash_amount" validate:"excluded_with_all=CreditCard CVV ExpiryDate"`
	}

	t.Run("credit card payment - no cash", func(t *testing.T) {
		validator := pedantigo.New[PaymentForm]()
		err := validator.Validate(&PaymentForm{
			CreditCard: "4111111111111111",
			CVV:        "123",
			ExpiryDate: "12/25",
			CashAmount: "",
		})
		require.NoError(t, err)
	})

	t.Run("credit card payment - with cash conflict", func(t *testing.T) {
		validator := pedantigo.New[PaymentForm]()
		err := validator.Validate(&PaymentForm{
			CreditCard: "4111111111111111",
			CVV:        "123",
			ExpiryDate: "12/25",
			CashAmount: "100.00",
		})
		require.Error(t, err) // Cannot have both payment methods
	})

	t.Run("cash payment only", func(t *testing.T) {
		validator := pedantigo.New[PaymentForm]()
		err := validator.Validate(&PaymentForm{
			CreditCard: "",
			CVV:        "",
			ExpiryDate: "",
			CashAmount: "100.00",
		})
		require.NoError(t, err)
	})

	t.Run("partial credit card info with cash", func(t *testing.T) {
		validator := pedantigo.New[PaymentForm]()
		err := validator.Validate(&PaymentForm{
			CreditCard: "4111111111111111",
			CVV:        "",
			ExpiryDate: "12/25",
			CashAmount: "100.00",
		})
		require.NoError(t, err) // Not all credit card fields present, so no conflict
	})
}

func TestRequiredWithoutAllComplexScenario(t *testing.T) {
	type AuthForm struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		// If none of the identifiers are present, OTP is required
		OTP string `json:"otp" validate:"required_without_all=Username Email Phone"`
	}

	t.Run("no identifiers - otp required", func(t *testing.T) {
		validator := pedantigo.New[AuthForm]()
		err := validator.Validate(&AuthForm{
			Username: "",
			Email:    "",
			Phone:    "",
			OTP:      "",
		})
		require.Error(t, err)
	})

	t.Run("no identifiers - otp provided", func(t *testing.T) {
		validator := pedantigo.New[AuthForm]()
		err := validator.Validate(&AuthForm{
			Username: "",
			Email:    "",
			Phone:    "",
			OTP:      "123456",
		})
		require.NoError(t, err)
	})

	t.Run("username present - otp not required", func(t *testing.T) {
		validator := pedantigo.New[AuthForm]()
		err := validator.Validate(&AuthForm{
			Username: "john_doe",
			Email:    "",
			Phone:    "",
			OTP:      "",
		})
		require.NoError(t, err)
	})

	t.Run("all identifiers present - otp not required", func(t *testing.T) {
		validator := pedantigo.New[AuthForm]()
		err := validator.Validate(&AuthForm{
			Username: "john_doe",
			Email:    "john@example.com",
			Phone:    "555-1234",
			OTP:      "",
		})
		require.NoError(t, err)
	})
}

func TestAllConstraintsMultipleOnSameStruct(t *testing.T) {
	type ComplexForm struct {
		A       string `json:"a"`
		B       string `json:"b"`
		C       string `json:"c"`
		Target1 string `json:"target1" validate:"required_with_all=A B"`
		Target2 string `json:"target2" validate:"excluded_with_all=A B C"`
		Target3 string `json:"target3" validate:"required_without_all=A B C"`
		Target4 string `json:"target4" validate:"excluded_without_all=A B C"`
	}

	t.Run("all fields present", func(t *testing.T) {
		validator := pedantigo.New[ComplexForm]()
		err := validator.Validate(&ComplexForm{
			A:       "a",
			B:       "b",
			C:       "c",
			Target1: "t1", // Required because A and B present
			Target2: "",   // Must be absent because A, B, C all present
			Target3: "",   // Not required because not all absent
			Target4: "t4", // Can be present because not all absent
		})
		require.NoError(t, err)
	})

	t.Run("all fields absent", func(t *testing.T) {
		validator := pedantigo.New[ComplexForm]()
		err := validator.Validate(&ComplexForm{
			A:       "",
			B:       "",
			C:       "",
			Target1: "",   // Not required because A and B not both present
			Target2: "t2", // Can be present because not all fields present
			Target3: "t3", // Required because all A, B, C absent
			Target4: "",   // Must be absent because all A, B, C absent
		})
		require.NoError(t, err)
	})

	t.Run("partial fields - A and B only", func(t *testing.T) {
		validator := pedantigo.New[ComplexForm]()
		err := validator.Validate(&ComplexForm{
			A:       "a",
			B:       "b",
			C:       "",
			Target1: "t1", // Required because A and B present
			Target2: "t2", // Can be present because not all fields present
			Target3: "",   // Not required because not all absent
			Target4: "t4", // Can be present because not all absent
		})
		require.NoError(t, err)
	})
}

func TestAllConstraintsFieldOrdering(t *testing.T) {
	// Test that field order in constraint doesn't matter
	type Form1 struct {
		A      string `json:"a"`
		B      string `json:"b"`
		C      string `json:"c"`
		Target string `json:"target" validate:"required_with_all=A B C"`
	}

	type Form2 struct {
		A      string `json:"a"`
		B      string `json:"b"`
		C      string `json:"c"`
		Target string `json:"target" validate:"required_with_all=C B A"`
	}

	t.Run("forward order", func(t *testing.T) {
		validator := pedantigo.New[Form1]()
		err := validator.Validate(&Form1{A: "1", B: "2", C: "3", Target: ""})
		require.Error(t, err)
	})

	t.Run("reverse order", func(t *testing.T) {
		validator := pedantigo.New[Form2]()
		err := validator.Validate(&Form2{A: "1", B: "2", C: "3", Target: ""})
		require.Error(t, err)
	})
}

func TestAllConstraintsWithPointers(t *testing.T) {
	type Form struct {
		A      *string `json:"a"`
		B      *string `json:"b"`
		Target string  `json:"target" validate:"required_with_all=A B"`
	}

	a := "value_a"
	b := "value_b"

	t.Run("both pointers present", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: &a, B: &b, Target: ""})
		require.Error(t, err)
	})

	t.Run("one pointer nil", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: &a, B: nil, Target: ""})
		require.NoError(t, err)
	})

	t.Run("both pointers nil", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: nil, B: nil, Target: ""})
		require.NoError(t, err)
	})

	t.Run("both pointers present with target", func(t *testing.T) {
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{A: &a, B: &b, Target: "target_value"})
		require.NoError(t, err)
	})
}
