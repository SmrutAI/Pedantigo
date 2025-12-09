package constraints_test

import (
	"testing"

	. "github.com/SmrutAI/Pedantigo"
)

// ==================================================
// excluded_if constraint tests
// ==================================================
// Field must be absent (zero value) if another field equals specific value

func TestExcludedIf_ConditionMet_FieldAbsent(t *testing.T) {
	type Payment struct {
		Method     string `json:"method" pedantigo:"required"`
		CashAmount int    `json:"cash_amount" pedantigo:"excluded_if=Method card"`
	}

	validator := New[Payment]()
	payment := &Payment{
		Method:     "card",
		CashAmount: 0, // Absent/zero
	}

	err := validator.Validate(payment)
	if err != nil {
		t.Errorf("expected no errors when excluded_if condition met and field absent, got %v", err)
	}
}

func TestExcludedIf_ConditionMet_FieldPresent(t *testing.T) {
	type Payment struct {
		Method     string `json:"method" pedantigo:"required"`
		CashAmount int    `json:"cash_amount" pedantigo:"excluded_if=Method card"`
	}

	validator := New[Payment]()
	payment := &Payment{
		Method:     "card",
		CashAmount: 100, // Present/non-zero
	}

	err := validator.Validate(payment)
	if err == nil {
		t.Error("expected validation error when excluded_if condition met but field is present")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "CashAmount" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("expected error for CashAmount field, got %v", ve.Errors)
	}
}

func TestExcludedIf_ConditionNotMet_FieldPresent(t *testing.T) {
	type Payment struct {
		Method     string `json:"method" pedantigo:"required"`
		CashAmount int    `json:"cash_amount" pedantigo:"excluded_if=Method card"`
	}

	validator := New[Payment]()
	payment := &Payment{
		Method:     "cash",
		CashAmount: 100, // Present/non-zero (allowed because condition not met)
	}

	err := validator.Validate(payment)
	if err != nil {
		t.Errorf("expected no errors when excluded_if condition not met, got %v", err)
	}
}

func TestExcludedIf_ConditionNotMet_FieldAbsent(t *testing.T) {
	type Payment struct {
		Method     string `json:"method" pedantigo:"required"`
		CashAmount int    `json:"cash_amount" pedantigo:"excluded_if=Method card"`
	}

	validator := New[Payment]()
	payment := &Payment{
		Method:     "cash",
		CashAmount: 0, // Absent/zero
	}

	err := validator.Validate(payment)
	if err != nil {
		t.Errorf("expected no errors when field is absent regardless of condition, got %v", err)
	}
}

func TestExcludedIf_StringComparison(t *testing.T) {
	type Order struct {
		PaymentType string `json:"payment_type" pedantigo:"required"`
		CheckNumber string `json:"check_number" pedantigo:"excluded_if=PaymentType credit_card"`
	}

	validator := New[Order]()

	// Valid: payment_type is "credit_card" and check_number is empty
	validOrder := &Order{
		PaymentType: "credit_card",
		CheckNumber: "", // Zero value for string
	}

	err := validator.Validate(validOrder)
	if err != nil {
		t.Errorf("expected no errors for valid credit card order, got %v", err)
	}

	// Invalid: payment_type is "credit_card" but check_number is provided
	invalidOrder := &Order{
		PaymentType: "credit_card",
		CheckNumber: "CHK123456",
	}

	err = validator.Validate(invalidOrder)
	if err == nil {
		t.Error("expected error when check_number provided for credit card payment")
	}
}

func TestExcludedIf_BooleanCondition(t *testing.T) {
	type UserPreferences struct {
		OptIn       bool   `json:"opt_in" pedantigo:"required"`
		PhoneNumber string `json:"phone_number" pedantigo:"excluded_if=OptIn false"`
	}

	validator := New[UserPreferences]()

	// Valid: OptIn=false and PhoneNumber is absent
	validPrefs := &UserPreferences{
		OptIn:       false,
		PhoneNumber: "", // Zero value
	}

	err := validator.Validate(validPrefs)
	if err != nil {
		t.Errorf("expected no errors when OptIn false and phone absent, got %v", err)
	}

	// Invalid: OptIn=false but PhoneNumber is provided
	invalidPrefs := &UserPreferences{
		OptIn:       false,
		PhoneNumber: "+1234567890",
	}

	err = validator.Validate(invalidPrefs)
	if err == nil {
		t.Error("expected error when phone provided for OptIn=false")
	}

	// Valid: OptIn=true, PhoneNumber can be anything
	validPrefsWith := &UserPreferences{
		OptIn:       true,
		PhoneNumber: "+1234567890",
	}

	err = validator.Validate(validPrefsWith)
	if err != nil {
		t.Errorf("expected no errors when OptIn true, got %v", err)
	}
}

func TestExcludedIf_MultipleConditions(t *testing.T) {
	type Vehicle struct {
		Type         string `json:"type" pedantigo:"required"`
		LicensePlate string `json:"license_plate" pedantigo:"excluded_if=Type bicycle"`
		ParkingSpot  int    `json:"parking_spot" pedantigo:"excluded_if=Type bicycle"`
	}

	validator := New[Vehicle]()

	// Valid: Type=bicycle, both excluded fields absent
	validBike := &Vehicle{
		Type:         "bicycle",
		LicensePlate: "",
		ParkingSpot:  0,
	}

	err := validator.Validate(validBike)
	if err != nil {
		t.Errorf("expected no errors for bicycle without license plate, got %v", err)
	}

	// Invalid: Type=bicycle but has license plate
	invalidBike := &Vehicle{
		Type:         "bicycle",
		LicensePlate: "ABC123",
		ParkingSpot:  0,
	}

	err = validator.Validate(invalidBike)
	if err == nil {
		t.Error("expected error when bicycle has license plate")
	}
}

// ==================================================
// excluded_unless constraint tests
// ==================================================
// Field must be absent (zero value) unless another field equals specific value

func TestExcludedUnless_ConditionMet_FieldPresent(t *testing.T) {
	type Document struct {
		Status        string `json:"status" pedantigo:"required"`
		ApprovalNotes string `json:"approval_notes" pedantigo:"excluded_unless=Status approved"`
	}

	validator := New[Document]()

	// Valid: Status=approved and ApprovalNotes present
	validDoc := &Document{
		Status:        "approved",
		ApprovalNotes: "Looks good to me",
	}

	err := validator.Validate(validDoc)
	if err != nil {
		t.Errorf("expected no errors when excluded_unless condition met and field present, got %v", err)
	}
}

func TestExcludedUnless_ConditionMet_FieldAbsent(t *testing.T) {
	type Document struct {
		Status        string `json:"status" pedantigo:"required"`
		ApprovalNotes string `json:"approval_notes" pedantigo:"excluded_unless=Status approved"`
	}

	validator := New[Document]()

	// Valid: Status=approved and ApprovalNotes absent (field can be absent even when condition met)
	validDoc := &Document{
		Status:        "approved",
		ApprovalNotes: "", // Zero value
	}

	err := validator.Validate(validDoc)
	if err != nil {
		t.Errorf("expected no errors when excluded_unless condition met regardless of field value, got %v", err)
	}
}

func TestExcludedUnless_ConditionNotMet_FieldAbsent(t *testing.T) {
	type Document struct {
		Status        string `json:"status" pedantigo:"required"`
		ApprovalNotes string `json:"approval_notes" pedantigo:"excluded_unless=Status approved"`
	}

	validator := New[Document]()

	// Valid: Status!=approved and ApprovalNotes absent (required to be absent)
	validDoc := &Document{
		Status:        "pending",
		ApprovalNotes: "", // Zero value
	}

	err := validator.Validate(validDoc)
	if err != nil {
		t.Errorf("expected no errors when excluded_unless condition not met and field absent, got %v", err)
	}
}

func TestExcludedUnless_ConditionNotMet_FieldPresent(t *testing.T) {
	type Document struct {
		Status        string `json:"status" pedantigo:"required"`
		ApprovalNotes string `json:"approval_notes" pedantigo:"excluded_unless=Status approved"`
	}

	validator := New[Document]()

	// Invalid: Status!=approved but ApprovalNotes present
	invalidDoc := &Document{
		Status:        "pending",
		ApprovalNotes: "Some notes",
	}

	err := validator.Validate(invalidDoc)
	if err == nil {
		t.Error("expected validation error when excluded_unless condition not met but field is present")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "ApprovalNotes" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("expected error for ApprovalNotes field, got %v", ve.Errors)
	}
}

func TestExcludedUnless_MultipleValues(t *testing.T) {
	type Permission struct {
		Role            string `json:"role" pedantigo:"required"`
		SecretKeyAccess string `json:"secret_key_access" pedantigo:"excluded_unless=Role admin"`
	}

	validator := New[Permission]()

	// Valid: Role=admin, SecretKeyAccess can be present or absent
	validAdmin := &Permission{
		Role:            "admin",
		SecretKeyAccess: "secret123",
	}

	err := validator.Validate(validAdmin)
	if err != nil {
		t.Errorf("expected no errors for admin with secret access, got %v", err)
	}

	// Invalid: Role=user, SecretKeyAccess present (not allowed)
	invalidUser := &Permission{
		Role:            "user",
		SecretKeyAccess: "secret123",
	}

	err = validator.Validate(invalidUser)
	if err == nil {
		t.Error("expected error when non-admin user has secret access")
	}

	// Valid: Role=user, SecretKeyAccess absent
	validUser := &Permission{
		Role:            "user",
		SecretKeyAccess: "",
	}

	err = validator.Validate(validUser)
	if err != nil {
		t.Errorf("expected no errors for user without secret access, got %v", err)
	}
}

// ==================================================
// excluded_with constraint tests
// ==================================================
// Field must be absent (zero value) if another field is present (non-zero)

func TestExcludedWith_OtherFieldPresent_FieldAbsent(t *testing.T) {
	type User struct {
		HomePhone string `json:"home_phone" pedantigo:"required"`
		WorkPhone string `json:"work_phone" pedantigo:"excluded_with=HomePhone"`
	}

	validator := New[User]()

	// Valid: HomePhone present, WorkPhone absent
	validUser := &User{
		HomePhone: "+1234567890",
		WorkPhone: "", // Zero value
	}

	err := validator.Validate(validUser)
	if err != nil {
		t.Errorf("expected no errors when excluded_with condition met and field absent, got %v", err)
	}
}

func TestExcludedWith_OtherFieldPresent_FieldPresent(t *testing.T) {
	type User struct {
		HomePhone string `json:"home_phone" pedantigo:"required"`
		WorkPhone string `json:"work_phone" pedantigo:"excluded_with=HomePhone"`
	}

	validator := New[User]()

	// Invalid: Both HomePhone and WorkPhone present
	invalidUser := &User{
		HomePhone: "+1234567890",
		WorkPhone: "+0987654321",
	}

	err := validator.Validate(invalidUser)
	if err == nil {
		t.Error("expected validation error when excluded_with condition met but field is present")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "WorkPhone" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("expected error for WorkPhone field, got %v", ve.Errors)
	}
}

func TestExcludedWith_OtherFieldAbsent_FieldPresent(t *testing.T) {
	type User struct {
		HomePhone string `json:"home_phone" pedantigo:"required"`
		WorkPhone string `json:"work_phone" pedantigo:"excluded_with=HomePhone"`
	}

	validator := New[User]()

	// Valid: HomePhone absent, WorkPhone present (condition not met, so field can be present)
	validUser := &User{
		HomePhone: "", // Zero value
		WorkPhone: "+0987654321",
	}

	err := validator.Validate(validUser)
	if err != nil {
		t.Errorf("expected no errors when excluded_with condition not met, got %v", err)
	}
}

func TestExcludedWith_OtherFieldAbsent_FieldAbsent(t *testing.T) {
	type User struct {
		HomePhone string `json:"home_phone" pedantigo:"required"`
		WorkPhone string `json:"work_phone" pedantigo:"excluded_with=HomePhone"`
	}

	validator := New[User]()

	// Valid: Both absent
	validUser := &User{
		HomePhone: "",
		WorkPhone: "",
	}

	err := validator.Validate(validUser)
	if err != nil {
		t.Errorf("expected no errors when both fields absent, got %v", err)
	}
}

func TestExcludedWith_IntegerField(t *testing.T) {
	type Account struct {
		BankBalance    int `json:"bank_balance" pedantigo:"min=0"`
		CreditLineUsed int `json:"credit_line_used" pedantigo:"excluded_with=BankBalance"`
	}

	validator := New[Account]()

	// Valid: BankBalance present (non-zero), CreditLineUsed absent
	validAccount := &Account{
		BankBalance:    5000,
		CreditLineUsed: 0, // Zero value
	}

	err := validator.Validate(validAccount)
	if err != nil {
		t.Errorf("expected no errors when excluded_with condition met and field absent, got %v", err)
	}

	// Invalid: Both present (non-zero)
	invalidAccount := &Account{
		BankBalance:    5000,
		CreditLineUsed: 2000,
	}

	err = validator.Validate(invalidAccount)
	if err == nil {
		t.Error("expected error when both fields are present")
	}

	// Valid: BankBalance absent (zero), CreditLineUsed can be anything
	validAccountNoBal := &Account{
		BankBalance:    0, // Zero value
		CreditLineUsed: 2000,
	}

	err = validator.Validate(validAccountNoBal)
	if err != nil {
		t.Errorf("expected no errors when excluded_with condition not met, got %v", err)
	}
}

func TestExcludedWith_BooleanField(t *testing.T) {
	type Feature struct {
		EnabledGlobally bool   `json:"enabled_globally" pedantigo:"required"`
		OverrideReason  string `json:"override_reason" pedantigo:"excluded_with=EnabledGlobally"`
	}

	validator := New[Feature]()

	// Valid: EnabledGlobally=true, OverrideReason absent
	validFeature := &Feature{
		EnabledGlobally: true,
		OverrideReason:  "", // Zero value
	}

	err := validator.Validate(validFeature)
	if err != nil {
		t.Errorf("expected no errors when EnabledGlobally true and reason absent, got %v", err)
	}

	// Invalid: EnabledGlobally=true, OverrideReason provided
	invalidFeature := &Feature{
		EnabledGlobally: true,
		OverrideReason:  "Special case",
	}

	err = validator.Validate(invalidFeature)
	if err == nil {
		t.Error("expected error when both EnabledGlobally and OverrideReason present")
	}

	// Valid: EnabledGlobally=false, OverrideReason can be anything
	validFeatureOverride := &Feature{
		EnabledGlobally: false,
		OverrideReason:  "Special case",
	}

	err = validator.Validate(validFeatureOverride)
	if err != nil {
		t.Errorf("expected no errors when EnabledGlobally false, got %v", err)
	}
}

// ==================================================
// excluded_without constraint tests
// ==================================================
// Field must be absent (zero value) if another field is absent (zero)

func TestExcludedWithout_OtherFieldAbsent_FieldAbsent(t *testing.T) {
	type Address struct {
		Country string `json:"country" pedantigo:"required"`
		ZipCode string `json:"zip_code" pedantigo:"excluded_without=Country"`
	}

	validator := New[Address]()

	// Valid: Country absent, ZipCode absent
	validAddress := &Address{
		Country: "", // Zero value
		ZipCode: "", // Zero value
	}

	err := validator.Validate(validAddress)
	if err != nil {
		t.Errorf("expected no errors when excluded_without condition met and field absent, got %v", err)
	}
}

func TestExcludedWithout_OtherFieldAbsent_FieldPresent(t *testing.T) {
	type Address struct {
		Country string `json:"country" pedantigo:"required"`
		ZipCode string `json:"zip_code" pedantigo:"excluded_without=Country"`
	}

	validator := New[Address]()

	// Invalid: Country absent but ZipCode present
	invalidAddress := &Address{
		Country: "",      // Zero value
		ZipCode: "12345", // Present
	}

	err := validator.Validate(invalidAddress)
	if err == nil {
		t.Error("expected validation error when excluded_without condition met but field is present")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "ZipCode" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("expected error for ZipCode field, got %v", ve.Errors)
	}
}

func TestExcludedWithout_OtherFieldPresent_FieldPresent(t *testing.T) {
	type Address struct {
		Country string `json:"country" pedantigo:"required"`
		ZipCode string `json:"zip_code" pedantigo:"excluded_without=Country"`
	}

	validator := New[Address]()

	// Valid: Country present, ZipCode can be anything
	validAddress := &Address{
		Country: "USA",
		ZipCode: "12345",
	}

	err := validator.Validate(validAddress)
	if err != nil {
		t.Errorf("expected no errors when excluded_without condition not met, got %v", err)
	}
}

func TestExcludedWithout_OtherFieldPresent_FieldAbsent(t *testing.T) {
	type Address struct {
		Country string `json:"country" pedantigo:"required"`
		ZipCode string `json:"zip_code" pedantigo:"excluded_without=Country"`
	}

	validator := New[Address]()

	// Valid: Country present, ZipCode absent (field can be absent when condition not met)
	validAddress := &Address{
		Country: "USA",
		ZipCode: "", // Zero value
	}

	err := validator.Validate(validAddress)
	if err != nil {
		t.Errorf("expected no errors when other field present and this field absent, got %v", err)
	}
}

func TestExcludedWithout_IntegerField(t *testing.T) {
	type Shipping struct {
		Weight      int `json:"weight"` // No min constraint for optional field
		TrackingNum int `json:"tracking_num" pedantigo:"excluded_without=Weight"`
	}

	validator := New[Shipping]()

	// Valid: Weight absent (missing from JSON), TrackingNum absent (missing from JSON)
	validShip, err := validator.Unmarshal([]byte(`{}`))
	if err != nil {
		t.Errorf("expected no errors when both fields absent from JSON, got %v", err)
	}
	if validShip.Weight != 0 || validShip.TrackingNum != 0 {
		t.Error("expected zero values for missing fields")
	}

	// Invalid: Weight absent (missing from JSON) but TrackingNum present
	_, err = validator.Unmarshal([]byte(`{"tracking_num": 123456789}`))
	if err == nil {
		t.Error("expected error when excluded_without condition met but field is present")
	}

	// Valid: Weight present, TrackingNum can be anything
	validShipWith, err := validator.Unmarshal([]byte(`{"weight": 500, "tracking_num": 123456789}`))
	if err != nil {
		t.Errorf("expected no errors when excluded_without condition not met, got %v", err)
	}
	if validShipWith.Weight != 500 || validShipWith.TrackingNum != 123456789 {
		t.Error("unexpected values after unmarshal")
	}

	// Valid: Weight present, TrackingNum absent
	validShipWithoutNum, err := validator.Unmarshal([]byte(`{"weight": 500}`))
	if err != nil {
		t.Errorf("expected no errors when Weight present and TrackingNum absent, got %v", err)
	}
	if validShipWithoutNum.Weight != 500 || validShipWithoutNum.TrackingNum != 0 {
		t.Error("unexpected values after unmarshal")
	}
}

func TestExcludedWithout_BooleanField(t *testing.T) {
	type Notification struct {
		IsEnabled   bool   `json:"is_enabled" pedantigo:"required"`
		RetryPolicy string `json:"retry_policy" pedantigo:"excluded_without=IsEnabled"`
	}

	validator := New[Notification]()

	// Valid: IsEnabled=true, RetryPolicy can be anything
	validNotif := &Notification{
		IsEnabled:   true,
		RetryPolicy: "exponential",
	}

	err := validator.Validate(validNotif)
	if err != nil {
		t.Errorf("expected no errors when IsEnabled true, got %v", err)
	}

	// Invalid: IsEnabled=false, RetryPolicy present
	invalidNotif := &Notification{
		IsEnabled:   false,
		RetryPolicy: "exponential",
	}

	err = validator.Validate(invalidNotif)
	if err == nil {
		t.Error("expected error when IsEnabled false but RetryPolicy present")
	}

	// Valid: IsEnabled=false, RetryPolicy absent
	validNotifNoPolicy := &Notification{
		IsEnabled:   false,
		RetryPolicy: "", // Zero value
	}

	err = validator.Validate(validNotifNoPolicy)
	if err != nil {
		t.Errorf("expected no errors when IsEnabled false and policy absent, got %v", err)
	}
}

// ==================================================
// Integration tests combining multiple constraints
// ==================================================

func TestMultipleExclusionConstraints_Complex(t *testing.T) {
	type Subscription struct {
		Status             string `json:"status" pedantigo:"required"`
		CancellationReason string `json:"cancellation_reason" pedantigo:"excluded_unless=Status cancelled"`
		DowngradeReason    string `json:"downgrade_reason" pedantigo:"excluded_unless=Status downgraded"`
		SuspendedUntilDate string `json:"suspended_until_date" pedantigo:"excluded_without=Status"`
	}

	validator := New[Subscription]()

	// Valid: Status=active, reason fields absent, suspension date absent
	validActive := &Subscription{
		Status:             "active",
		CancellationReason: "",
		DowngradeReason:    "",
		SuspendedUntilDate: "",
	}

	err := validator.Validate(validActive)
	if err != nil {
		t.Errorf("expected no errors for active subscription, got %v", err)
	}

	// Valid: Status=cancelled, cancellation reason can be present
	validCancelled := &Subscription{
		Status:             "cancelled",
		CancellationReason: "Not needed",
		DowngradeReason:    "",
		SuspendedUntilDate: "",
	}

	err = validator.Validate(validCancelled)
	if err != nil {
		t.Errorf("expected no errors for cancelled subscription with reason, got %v", err)
	}

	// Invalid: Status=active but cancellation reason present
	invalidActive := &Subscription{
		Status:             "active",
		CancellationReason: "Not needed",
		DowngradeReason:    "",
		SuspendedUntilDate: "2025-01-01",
	}

	err = validator.Validate(invalidActive)
	if err == nil {
		t.Error("expected error for active subscription with cancellation reason")
	}
}

func TestConditionalExclusion_RealWorldPaymentExample(t *testing.T) {
	type PaymentMethod struct {
		Type           string `json:"type" pedantigo:"required"`
		CardNumber     string `json:"card_number" pedantigo:"excluded_unless=Type card"`
		BankAccount    string `json:"bank_account" pedantigo:"excluded_unless=Type bank_transfer"`
		CryptoCurrency string `json:"crypto_currency" pedantigo:"excluded_unless=Type crypto"`
		CardExpiryDate string `json:"card_expiry_date" pedantigo:"excluded_with=BankAccount,excluded_with=CryptoCurrency"`
		RoutingNumber  string `json:"routing_number" pedantigo:"excluded_without=Type"`
	}

	validator := New[PaymentMethod]()

	// Valid: Credit card payment with card details
	validCard := &PaymentMethod{
		Type:           "card",
		CardNumber:     "4111111111111111",
		CardExpiryDate: "12/25",
		BankAccount:    "",
		CryptoCurrency: "",
		RoutingNumber:  "",
	}

	err := validator.Validate(validCard)
	if err != nil {
		t.Errorf("expected no errors for card payment, got %v", err)
	}

	// Invalid: Card payment with bank account also present
	invalidCard := &PaymentMethod{
		Type:           "card",
		CardNumber:     "4111111111111111",
		CardExpiryDate: "12/25",
		BankAccount:    "123456789",
		CryptoCurrency: "",
		RoutingNumber:  "",
	}

	err = validator.Validate(invalidCard)
	if err == nil {
		t.Error("expected error for card payment with bank account")
	}

	// Valid: Bank transfer with account details
	validBank := &PaymentMethod{
		Type:           "bank_transfer",
		CardNumber:     "",
		CardExpiryDate: "",
		BankAccount:    "123456789",
		CryptoCurrency: "",
		RoutingNumber:  "021000021",
	}

	err = validator.Validate(validBank)
	if err != nil {
		t.Errorf("expected no errors for bank transfer, got %v", err)
	}

	// Valid: Crypto payment without card details
	validCrypto := &PaymentMethod{
		Type:           "crypto",
		CardNumber:     "",
		CardExpiryDate: "",
		BankAccount:    "",
		CryptoCurrency: "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		RoutingNumber:  "",
	}

	err = validator.Validate(validCrypto)
	if err != nil {
		t.Errorf("expected no errors for crypto payment, got %v", err)
	}
}
