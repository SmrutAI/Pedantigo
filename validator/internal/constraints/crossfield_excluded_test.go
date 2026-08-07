package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

// checkValidationError verifies that a validation error contains the expected field error.
func checkValidationError(t *testing.T, err error, expectErr bool, errField string) {
	t.Helper()

	if expectErr {
		var ve *validator.ValidationError
		require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
		var found bool
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == errField {
				found = true
				break
			}
		}
		assert.True(t, found, "expected error for field %s, got %v", errField, ve.Errors)
	} else {
		require.NoError(t, err)
	}
}

// ==================================================
// excluded_if constraint tests
// ==================================================
// Field must be absent (zero value) if another field equals specific value

// TestExcludedIf tests ExcludedIf validation.
func TestExcludedIf(t *testing.T) {
	type Payment struct {
		Method     string `json:"method" validate:"required"`
		CashAmount int    `json:"cash_amount" validate:"excluded_if=Method card"`
	}

	type UserPreferences struct {
		OptIn       bool   `json:"opt_in" validate:"required"`
		PhoneNumber string `json:"phone_number" validate:"excluded_if=OptIn false"`
	}

	type Vehicle struct {
		Type         string `json:"type" validate:"required"`
		LicensePlate string `json:"license_plate" validate:"excluded_if=Type bicycle"`
		ParkingSpot  int    `json:"parking_spot" validate:"excluded_if=Type bicycle"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "condition met field absent - valid",
			validator: validator.New[Payment](),
			data: &Payment{
				Method:     "card",
				CashAmount: 0,
			},
			expectErr: false,
		},
		{
			name:      "condition met field present - invalid",
			validator: validator.New[Payment](),
			data: &Payment{
				Method:     "card",
				CashAmount: 100,
			},
			expectErr: true,
			errField:  "CashAmount",
		},
		{
			name:      "condition not met field present - valid",
			validator: validator.New[Payment](),
			data: &Payment{
				Method:     "cash",
				CashAmount: 100,
			},
			expectErr: false,
		},
		{
			name:      "condition not met field absent - valid",
			validator: validator.New[Payment](),
			data: &Payment{
				Method:     "cash",
				CashAmount: 0,
			},
			expectErr: false,
		},
		{
			name:      "boolean condition valid - false with absent",
			validator: validator.New[UserPreferences](),
			data: &UserPreferences{
				OptIn:       false,
				PhoneNumber: "",
			},
			expectErr: false,
		},
		{
			name:      "boolean condition invalid - false with present",
			validator: validator.New[UserPreferences](),
			data: &UserPreferences{
				OptIn:       false,
				PhoneNumber: "+1234567890",
			},
			expectErr: true,
			errField:  "PhoneNumber",
		},
		{
			name:      "boolean condition valid - true with present",
			validator: validator.New[UserPreferences](),
			data: &UserPreferences{
				OptIn:       true,
				PhoneNumber: "+1234567890",
			},
			expectErr: false,
		},
		{
			name:      "multiple conditions - bicycle without license plate",
			validator: validator.New[Vehicle](),
			data: &Vehicle{
				Type:         "bicycle",
				LicensePlate: "",
				ParkingSpot:  0,
			},
			expectErr: false,
		},
		{
			name:      "multiple conditions - bicycle with license plate invalid",
			validator: validator.New[Vehicle](),
			data: &Vehicle{
				Type:         "bicycle",
				LicensePlate: "ABC123",
				ParkingSpot:  0,
			},
			expectErr: true,
			errField:  "LicensePlate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *validator.Validator[Payment]:
				err := v.Validate(tt.data.(*Payment))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			case *validator.Validator[UserPreferences]:
				err := v.Validate(tt.data.(*UserPreferences))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			case *validator.Validator[Vehicle]:
				err := v.Validate(tt.data.(*Vehicle))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			}
		})
	}
}

// ==================================================
// excluded_unless constraint tests
// ==================================================
// Field must be absent (zero value) unless another field equals specific value

// TestExcludedUnless tests ExcludedUnless validation.
func TestExcludedUnless(t *testing.T) {
	type Document struct {
		Status        string `json:"status" validate:"required"`
		ApprovalNotes string `json:"approval_notes" validate:"excluded_unless=Status approved"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "condition met field present - valid",
			validator: validator.New[Document](),
			data: &Document{
				Status:        "approved",
				ApprovalNotes: "Looks good to me",
			},
			expectErr: false,
		},
		{
			name:      "condition met field absent - valid",
			validator: validator.New[Document](),
			data: &Document{
				Status:        "approved",
				ApprovalNotes: "",
			},
			expectErr: false,
		},
		{
			name:      "condition not met field absent - valid",
			validator: validator.New[Document](),
			data: &Document{
				Status:        "pending",
				ApprovalNotes: "",
			},
			expectErr: false,
		},
		{
			name:      "condition not met field present - invalid",
			validator: validator.New[Document](),
			data: &Document{
				Status:        "pending",
				ApprovalNotes: "Some notes",
			},
			expectErr: true,
			errField:  "ApprovalNotes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if v, ok := tt.validator.(*validator.Validator[Document]); ok {
				err := v.Validate(tt.data.(*Document))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			}
		})
	}
}

// ==================================================
// excluded_with constraint tests
// ==================================================
// Field must be absent (zero value) if another field is present (non-zero)

// TestExcludedWith tests ExcludedWith validation.
func TestExcludedWith(t *testing.T) {
	type User struct {
		HomePhone string `json:"home_phone" validate:"required"`
		WorkPhone string `json:"work_phone" validate:"excluded_with=HomePhone"`
	}

	type Account struct {
		BankBalance    int `json:"bank_balance" validate:"min=0"`
		CreditLineUsed int `json:"credit_line_used" validate:"excluded_with=BankBalance"`
	}

	type Feature struct {
		EnabledGlobally bool   `json:"enabled_globally" validate:"required"`
		OverrideReason  string `json:"override_reason" validate:"excluded_with=EnabledGlobally"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "other field present field absent - valid",
			validator: validator.New[User](),
			data: &User{
				HomePhone: "+1234567890",
				WorkPhone: "",
			},
			expectErr: false,
		},
		{
			name:      "other field present field present - invalid",
			validator: validator.New[User](),
			data: &User{
				HomePhone: "+1234567890",
				WorkPhone: "+0987654321",
			},
			expectErr: true,
			errField:  "WorkPhone",
		},
		{
			name:      "other field absent field present - valid",
			validator: validator.New[User](),
			data: &User{
				HomePhone: "",
				WorkPhone: "+0987654321",
			},
			expectErr: false,
		},
		{
			name:      "other field absent field absent - valid",
			validator: validator.New[User](),
			data: &User{
				HomePhone: "",
				WorkPhone: "",
			},
			expectErr: false,
		},
		{
			name:      "integer field present absent - valid",
			validator: validator.New[Account](),
			data: &Account{
				BankBalance:    5000,
				CreditLineUsed: 0,
			},
			expectErr: false,
		},
		{
			name:      "integer field both present - invalid",
			validator: validator.New[Account](),
			data: &Account{
				BankBalance:    5000,
				CreditLineUsed: 2000,
			},
			expectErr: true,
			errField:  "CreditLineUsed",
		},
		{
			name:      "integer field absent can be present - valid",
			validator: validator.New[Account](),
			data: &Account{
				BankBalance:    0,
				CreditLineUsed: 2000,
			},
			expectErr: false,
		},
		{
			name:      "boolean field true reason absent - valid",
			validator: validator.New[Feature](),
			data: &Feature{
				EnabledGlobally: true,
				OverrideReason:  "",
			},
			expectErr: false,
		},
		{
			name:      "boolean field true reason present - invalid",
			validator: validator.New[Feature](),
			data: &Feature{
				EnabledGlobally: true,
				OverrideReason:  "Special case",
			},
			expectErr: true,
			errField:  "OverrideReason",
		},
		{
			name:      "boolean field false reason present - valid",
			validator: validator.New[Feature](),
			data: &Feature{
				EnabledGlobally: false,
				OverrideReason:  "Special case",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *validator.Validator[User]:
				err := v.Validate(tt.data.(*User))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			case *validator.Validator[Account]:
				err := v.Validate(tt.data.(*Account))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			case *validator.Validator[Feature]:
				err := v.Validate(tt.data.(*Feature))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			}
		})
	}
}

// ==================================================
// excluded_without constraint tests
// ==================================================
// Field must be absent (zero value) if another field is absent (zero)

// TestExcludedWithout tests ExcludedWithout validation.
func TestExcludedWithout(t *testing.T) {
	type Address struct {
		Country string `json:"country" validate:"required"`
		ZipCode string `json:"zip_code" validate:"excluded_without=Country"`
	}

	type Notification struct {
		IsEnabled   bool   `json:"is_enabled" validate:"required"`
		RetryPolicy string `json:"retry_policy" validate:"excluded_without=IsEnabled"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "other field absent field absent - valid",
			validator: validator.New[Address](),
			data: &Address{
				Country: "",
				ZipCode: "",
			},
			expectErr: false,
		},
		{
			name:      "other field absent field present - invalid",
			validator: validator.New[Address](),
			data: &Address{
				Country: "",
				ZipCode: "12345",
			},
			expectErr: true,
			errField:  "ZipCode",
		},
		{
			name:      "other field present field present - valid",
			validator: validator.New[Address](),
			data: &Address{
				Country: "USA",
				ZipCode: "12345",
			},
			expectErr: false,
		},
		{
			name:      "other field present field absent - valid",
			validator: validator.New[Address](),
			data: &Address{
				Country: "USA",
				ZipCode: "",
			},
			expectErr: false,
		},
		{
			name:      "boolean field true policy present - valid",
			validator: validator.New[Notification](),
			data: &Notification{
				IsEnabled:   true,
				RetryPolicy: "exponential",
			},
			expectErr: false,
		},
		{
			name:      "boolean field false policy present - invalid",
			validator: validator.New[Notification](),
			data: &Notification{
				IsEnabled:   false,
				RetryPolicy: "exponential",
			},
			expectErr: true,
			errField:  "RetryPolicy",
		},
		{
			name:      "boolean field false policy absent - valid",
			validator: validator.New[Notification](),
			data: &Notification{
				IsEnabled:   false,
				RetryPolicy: "",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *validator.Validator[Address]:
				err := v.Validate(tt.data.(*Address))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			case *validator.Validator[Notification]:
				err := v.Validate(tt.data.(*Notification))
				checkValidationError(t, err, tt.expectErr, tt.errField)
			}
		})
	}
}

// ==================================================
// excluded_without unmarshal integration tests
// ==================================================

// TestExcludedWithoutUnmarshal tests ExcludedWithoutUnmarshal validation.
func TestExcludedWithoutUnmarshal(t *testing.T) {
	type Shipping struct {
		Weight      int `json:"weight"`
		TrackingNum int `json:"tracking_num" validate:"excluded_without=Weight"`
	}

	tests := []struct {
		name      string
		jsonData  string
		expectErr bool
		checkFn   func(*Shipping) bool
	}{
		{
			name:      "both fields absent from json - valid",
			jsonData:  `{}`,
			expectErr: false,
			checkFn: func(s *Shipping) bool {
				return s.Weight == 0 && s.TrackingNum == 0
			},
		},
		{
			name:      "weight absent tracking num present - invalid",
			jsonData:  `{"tracking_num": 123456789}`,
			expectErr: true,
		},
		{
			name:      "weight present tracking num present - valid",
			jsonData:  `{"weight": 500, "tracking_num": 123456789}`,
			expectErr: false,
			checkFn: func(s *Shipping) bool {
				return s.Weight == 500 && s.TrackingNum == 123456789
			},
		},
		{
			name:      "weight present tracking num absent - valid",
			jsonData:  `{"weight": 500}`,
			expectErr: false,
			checkFn: func(s *Shipping) bool {
				return s.Weight == 500 && s.TrackingNum == 0
			},
		},
	}

	vl := validator.New[Shipping]()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vl.Unmarshal([]byte(tt.jsonData))
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkFn != nil {
					assert.True(t, tt.checkFn(result), "data validation failed for %+v", result)
				}
			}
		})
	}
}

// ==================================================
// Integration tests combining multiple constraints
// ==================================================

// TestMultipleExclusionConstraints_Complex tests MultipleExclusionConstraints complex.
func TestMultipleExclusionConstraints_Complex(t *testing.T) {
	type Subscription struct {
		Status             string `json:"status" validate:"required"`
		CancellationReason string `json:"cancellation_reason" validate:"excluded_unless=Status cancelled"`
		DowngradeReason    string `json:"downgrade_reason" validate:"excluded_unless=Status downgraded"`
		SuspendedUntilDate string `json:"suspended_until_date" validate:"excluded_without=Status"`
	}

	vl := validator.New[Subscription]()

	tests := []struct {
		name      string
		data      *Subscription
		expectErr bool
	}{
		{
			name: "active subscription - valid",
			data: &Subscription{
				Status:             "active",
				CancellationReason: "",
				DowngradeReason:    "",
				SuspendedUntilDate: "",
			},
			expectErr: false,
		},
		{
			name: "cancelled subscription with reason - valid",
			data: &Subscription{
				Status:             "cancelled",
				CancellationReason: "Not needed",
				DowngradeReason:    "",
				SuspendedUntilDate: "",
			},
			expectErr: false,
		},
		{
			name: "active subscription with cancellation reason - invalid",
			data: &Subscription{
				Status:             "active",
				CancellationReason: "Not needed",
				DowngradeReason:    "",
				SuspendedUntilDate: "2025-01-01",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vl.Validate(tt.data)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConditionalExclusion_RealWorldPaymentExample(t *testing.T) {
	type PaymentMethod struct {
		Type           string `json:"type" validate:"required"`
		CardNumber     string `json:"card_number" validate:"excluded_unless=Type card"`
		BankAccount    string `json:"bank_account" validate:"excluded_unless=Type bank_transfer"`
		CryptoCurrency string `json:"crypto_currency" validate:"excluded_unless=Type crypto"`
		CardExpiryDate string `json:"card_expiry_date" validate:"excluded_with=BankAccount,excluded_with=CryptoCurrency"`
		RoutingNumber  string `json:"routing_number" validate:"excluded_without=Type"`
	}

	vl := validator.New[PaymentMethod]()

	tests := []struct {
		name      string
		data      *PaymentMethod
		expectErr bool
	}{
		{
			name: "credit card payment - valid",
			data: &PaymentMethod{
				Type:           "card",
				CardNumber:     "4111111111111111",
				CardExpiryDate: "12/25",
				BankAccount:    "",
				CryptoCurrency: "",
				RoutingNumber:  "",
			},
			expectErr: false,
		},
		{
			name: "card payment with bank account - invalid",
			data: &PaymentMethod{
				Type:           "card",
				CardNumber:     "4111111111111111",
				CardExpiryDate: "12/25",
				BankAccount:    "123456789",
				CryptoCurrency: "",
				RoutingNumber:  "",
			},
			expectErr: true,
		},
		{
			name: "bank transfer payment - valid",
			data: &PaymentMethod{
				Type:           "bank_transfer",
				CardNumber:     "",
				CardExpiryDate: "",
				BankAccount:    "123456789",
				CryptoCurrency: "",
				RoutingNumber:  "021000021",
			},
			expectErr: false,
		},
		{
			name: "crypto payment - valid",
			data: &PaymentMethod{
				Type:           "crypto",
				CardNumber:     "",
				CardExpiryDate: "",
				BankAccount:    "",
				CryptoCurrency: "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
				RoutingNumber:  "",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vl.Validate(tt.data)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
