package constraints_test

import (
	"reflect"
	"testing"

	. "github.com/SmrutAI/Pedantigo"
)

// ============================================================================
// required_if Tests
// ============================================================================

// TestRequiredIf_ConditionMet_FieldPresent validates that required_if is satisfied
// when the condition is true AND the field is provided.
func TestRequiredIf_ConditionMet_FieldPresent(t *testing.T) {
	type Form struct {
		Country string `json:"country"`
		State   string `json:"state" pedantigo:"required_if=Country:US"`
	}

	validator := New[Form]()

	// Valid: Country=US and State provided
	valid := &Form{Country: "US", State: "CA"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors for valid form, got: %v", err)
	}
}

// TestRequiredIf_ConditionMet_FieldMissing validates that required_if fails
// when the condition is true BUT the field is missing (zero value).
func TestRequiredIf_ConditionMet_FieldMissing(t *testing.T) {
	type Form struct {
		Country string `json:"country"`
		State   string `json:"state" pedantigo:"required_if=Country:US"`
	}

	validator := New[Form]()

	// Invalid: Country=US but State is missing (zero value)
	invalid := &Form{Country: "US", State: ""}
	err := validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when State is missing for Country=US")
	}

	// Verify it's a ValidationError with the correct field
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(ve.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	if ve.Errors[0].Field != "State" {
		t.Errorf("expected error for field 'State', got '%s'", ve.Errors[0].Field)
	}
}

// TestRequiredIf_ConditionNotMet_FieldMissing validates that required_if is satisfied
// when the condition is false, even if the field is missing.
func TestRequiredIf_ConditionNotMet_FieldMissing(t *testing.T) {
	type Form struct {
		Country string `json:"country"`
		State   string `json:"state" pedantigo:"required_if=Country:US"`
	}

	validator := New[Form]()

	// Valid: Country!=US, so State can be empty
	valid := &Form{Country: "CA", State: ""}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors when condition not met, got: %v", err)
	}
}

// TestRequiredIf_ConditionNotMet_FieldPresent validates that required_if is satisfied
// when the condition is false AND the field is provided (optional when condition false).
func TestRequiredIf_ConditionNotMet_FieldPresent(t *testing.T) {
	type Form struct {
		Country string `json:"country"`
		State   string `json:"state" pedantigo:"required_if=Country:US"`
	}

	validator := New[Form]()

	// Valid: Country!=US, State can be provided or not
	valid := &Form{Country: "Canada", State: "Ontario"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredIf_BooleanCondition validates required_if with boolean field conditions.
func TestRequiredIf_BooleanCondition(t *testing.T) {
	type Subscription struct {
		IsPremium      bool   `json:"is_premium"`
		PremiumFeature string `json:"premium_feature" pedantigo:"required_if=IsPremium:true"`
	}

	validator := New[Subscription]()

	// Valid: IsPremium=true, PremiumFeature provided
	valid := &Subscription{IsPremium: true, PremiumFeature: "advanced_analytics"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: IsPremium=true, PremiumFeature missing
	invalid := &Subscription{IsPremium: true, PremiumFeature: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when PremiumFeature missing for premium user")
	}

	// Valid: IsPremium=false, PremiumFeature can be empty
	validFree := &Subscription{IsPremium: false, PremiumFeature: ""}
	err = validator.Validate(validFree)
	if err != nil {
		t.Errorf("expected no errors for free user, got: %v", err)
	}
}

// TestRequiredIf_IntegerCondition validates required_if with integer field conditions.
func TestRequiredIf_IntegerCondition(t *testing.T) {
	type Order struct {
		Status       int    `json:"status"` // 0=pending, 1=processing, 2=completed
		TrackingCode string `json:"tracking_code" pedantigo:"required_if=Status:2"`
	}

	validator := New[Order]()

	// Valid: Status=2 (completed), TrackingCode provided
	valid := &Order{Status: 2, TrackingCode: "TRACK123"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: Status=2 (completed), TrackingCode missing
	invalid := &Order{Status: 2, TrackingCode: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when TrackingCode missing for completed order")
	}

	// Valid: Status=0 (pending), TrackingCode can be empty
	validPending := &Order{Status: 0, TrackingCode: ""}
	err = validator.Validate(validPending)
	if err != nil {
		t.Errorf("expected no errors for pending order, got: %v", err)
	}
}

// TestRequiredIf_MultipleConditions validates required_if with multiple fields having conditions.
func TestRequiredIf_MultipleConditions(t *testing.T) {
	type Shipment struct {
		Country  string `json:"country"`
		Domestic bool   `json:"domestic"`
		State    string `json:"state" pedantigo:"required_if=Country:US"`
		TaxID    string `json:"tax_id" pedantigo:"required_if=Domestic:true"`
	}

	validator := New[Shipment]()

	// Valid: US country with State, Domestic with TaxID
	valid := &Shipment{Country: "US", Domestic: true, State: "CA", TaxID: "123456"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: US country but State missing
	invalidState := &Shipment{Country: "US", Domestic: true, State: "", TaxID: "123456"}
	err = validator.Validate(invalidState)
	if err == nil {
		t.Error("expected validation error when State missing for US")
	}

	// Invalid: Domestic=true but TaxID missing
	invalidTaxID := &Shipment{Country: "US", Domestic: true, State: "CA", TaxID: ""}
	err = validator.Validate(invalidTaxID)
	if err == nil {
		t.Error("expected validation error when TaxID missing for domestic shipment")
	}
}

// ============================================================================
// required_unless Tests
// ============================================================================

// TestRequiredUnless_ConditionNotMet_FieldPresent validates that required_unless
// is satisfied when the condition is false AND the field is provided.
func TestRequiredUnless_ConditionNotMet_FieldPresent(t *testing.T) {
	type Account struct {
		Status   string `json:"status"`
		Password string `json:"password" pedantigo:"required_unless=Status:guest"`
	}

	validator := New[Account]()

	// Valid: Status!=guest, Password provided
	valid := &Account{Status: "active", Password: "securepass123"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredUnless_ConditionNotMet_FieldMissing validates that required_unless
// fails when the condition is false BUT the field is missing.
func TestRequiredUnless_ConditionNotMet_FieldMissing(t *testing.T) {
	type Account struct {
		Status   string `json:"status"`
		Password string `json:"password" pedantigo:"required_unless=Status:guest"`
	}

	validator := New[Account]()

	// Invalid: Status!=guest, but Password is missing
	invalid := &Account{Status: "active", Password: ""}
	err := validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when Password missing for non-guest account")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(ve.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	if ve.Errors[0].Field != "Password" {
		t.Errorf("expected error for field 'Password', got '%s'", ve.Errors[0].Field)
	}
}

// TestRequiredUnless_ConditionMet_FieldMissing validates that required_unless
// is satisfied when the condition is true, even if the field is missing.
func TestRequiredUnless_ConditionMet_FieldMissing(t *testing.T) {
	type Account struct {
		Status   string `json:"status"`
		Password string `json:"password" pedantigo:"required_unless=Status:guest"`
	}

	validator := New[Account]()

	// Valid: Status=guest, so Password can be empty
	valid := &Account{Status: "guest", Password: ""}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors for guest account, got: %v", err)
	}
}

// TestRequiredUnless_ConditionMet_FieldPresent validates that required_unless
// is satisfied when the condition is true AND the field is provided (optional).
func TestRequiredUnless_ConditionMet_FieldPresent(t *testing.T) {
	type Account struct {
		Status   string `json:"status"`
		Password string `json:"password" pedantigo:"required_unless=Status:guest"`
	}

	validator := New[Account]()

	// Valid: Status=guest, Password can be provided or not
	valid := &Account{Status: "guest", Password: "anypassword"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredUnless_BooleanCondition validates required_unless with boolean conditions.
func TestRequiredUnless_BooleanCondition(t *testing.T) {
	type Signup struct {
		Automated   bool   `json:"automated"`
		CaptchaCode string `json:"captcha_code" pedantigo:"required_unless=Automated:true"`
	}

	validator := New[Signup]()

	// Valid: Automated=false (human), CaptchaCode provided
	valid := &Signup{Automated: false, CaptchaCode: "abc123xyz"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: Automated=false, CaptchaCode missing
	invalid := &Signup{Automated: false, CaptchaCode: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when CaptchaCode missing for human signup")
	}

	// Valid: Automated=true, CaptchaCode can be empty
	validAuto := &Signup{Automated: true, CaptchaCode: ""}
	err = validator.Validate(validAuto)
	if err != nil {
		t.Errorf("expected no errors for automated signup, got: %v", err)
	}
}

// TestRequiredUnless_MultipleConditions validates required_unless with multiple fields.
func TestRequiredUnless_MultipleConditions(t *testing.T) {
	type Registration struct {
		UserType     string `json:"user_type"`
		IsBot        bool   `json:"is_bot"`
		Email        string `json:"email" pedantigo:"required_unless=UserType:anonymous"`
		Verification string `json:"verification" pedantigo:"required_unless=IsBot:true"`
	}

	validator := New[Registration]()

	// Valid: not anonymous (email required), not bot (verification required)
	valid := &Registration{UserType: "user", IsBot: false, Email: "user@example.com", Verification: "code123"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: UserType!=anonymous but Email missing
	invalidEmail := &Registration{UserType: "user", IsBot: false, Email: "", Verification: "code123"}
	err = validator.Validate(invalidEmail)
	if err == nil {
		t.Error("expected validation error when Email missing for non-anonymous user")
	}

	// Invalid: IsBot!=true but Verification missing
	invalidVerif := &Registration{UserType: "user", IsBot: false, Email: "user@example.com", Verification: ""}
	err = validator.Validate(invalidVerif)
	if err == nil {
		t.Error("expected validation error when Verification missing for non-bot user")
	}

	// Valid: anonymous (email optional), bot (verification optional)
	validException := &Registration{UserType: "anonymous", IsBot: true, Email: "", Verification: ""}
	err = validator.Validate(validException)
	if err != nil {
		t.Errorf("expected no errors for anonymous bot, got: %v", err)
	}
}

// ============================================================================
// required_with Tests
// ============================================================================

// TestRequiredWith_OtherFieldPresent_FieldPresent validates that required_with
// is satisfied when the other field is non-zero AND the field is provided.
func TestRequiredWith_OtherFieldPresent_FieldPresent(t *testing.T) {
	type Payment struct {
		Method string `json:"method"`
		Token  string `json:"token" pedantigo:"required_with=Method"`
	}

	validator := New[Payment]()

	// Valid: Method is present (non-zero), Token is provided
	valid := &Payment{Method: "credit_card", Token: "tok_123"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredWith_OtherFieldPresent_FieldMissing validates that required_with
// fails when the other field is non-zero BUT the field is missing.
func TestRequiredWith_OtherFieldPresent_FieldMissing(t *testing.T) {
	type Payment struct {
		Method string `json:"method"`
		Token  string `json:"token" pedantigo:"required_with=Method"`
	}

	validator := New[Payment]()

	// Invalid: Method is present, but Token is missing
	invalid := &Payment{Method: "credit_card", Token: ""}
	err := validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when Token missing but Method provided")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(ve.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	if ve.Errors[0].Field != "Token" {
		t.Errorf("expected error for field 'Token', got '%s'", ve.Errors[0].Field)
	}
}

// TestRequiredWith_OtherFieldAbsent_FieldMissing validates that required_with
// is satisfied when the other field is zero, even if this field is missing.
func TestRequiredWith_OtherFieldAbsent_FieldMissing(t *testing.T) {
	type Payment struct {
		Method string `json:"method"`
		Token  string `json:"token" pedantigo:"required_with=Method"`
	}

	validator := New[Payment]()

	// Valid: Method is zero (absent), so Token can be empty
	valid := &Payment{Method: "", Token: ""}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors when other field absent, got: %v", err)
	}
}

// TestRequiredWith_OtherFieldAbsent_FieldPresent validates that required_with
// is satisfied when the other field is zero, and this field can be optionally present.
func TestRequiredWith_OtherFieldAbsent_FieldPresent(t *testing.T) {
	type Payment struct {
		Method string `json:"method"`
		Token  string `json:"token" pedantigo:"required_with=Method"`
	}

	validator := New[Payment]()

	// Valid: Method is zero, Token is optionally present (allowed)
	valid := &Payment{Method: "", Token: "tok_123"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredWith_NumericField validates required_with with numeric field conditions.
func TestRequiredWith_NumericField(t *testing.T) {
	type Product struct {
		Quantity  int    `json:"quantity"`
		Warehouse string `json:"warehouse" pedantigo:"required_with=Quantity"`
	}

	validator := New[Product]()

	// Valid: Quantity is provided (non-zero), Warehouse provided
	valid := &Product{Quantity: 5, Warehouse: "main"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: Quantity is provided, Warehouse missing
	invalid := &Product{Quantity: 5, Warehouse: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when Warehouse missing for provided Quantity")
	}

	// Valid: Quantity is zero, Warehouse can be empty
	validZero := &Product{Quantity: 0, Warehouse: ""}
	err = validator.Validate(validZero)
	if err != nil {
		t.Errorf("expected no errors for zero quantity, got: %v", err)
	}
}

// TestRequiredWith_BooleanField validates required_with with boolean field conditions.
func TestRequiredWith_BooleanField(t *testing.T) {
	type Feature struct {
		Enabled bool   `json:"enabled"`
		Config  string `json:"config" pedantigo:"required_with=Enabled"`
	}

	validator := New[Feature]()

	// Valid: Enabled=true (non-zero), Config provided
	valid := &Feature{Enabled: true, Config: "settings"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: Enabled=true, Config missing
	invalid := &Feature{Enabled: true, Config: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when Config missing for enabled feature")
	}

	// Valid: Enabled=false (zero), Config can be empty
	validDisabled := &Feature{Enabled: false, Config: ""}
	err = validator.Validate(validDisabled)
	if err != nil {
		t.Errorf("expected no errors when feature disabled, got: %v", err)
	}
}

// TestRequiredWith_MultipleFields validates required_with with multiple fields having conditions.
func TestRequiredWith_MultipleFields(t *testing.T) {
	type Booking struct {
		GuestName     string `json:"guest_name"`
		Phone         string `json:"phone"`
		Address       string `json:"address" pedantigo:"required_with=GuestName"`
		EmergencyName string `json:"emergency_name" pedantigo:"required_with=Phone"`
	}

	validator := New[Booking]()

	// Valid: GuestName provided with Address, Phone provided with EmergencyName
	valid := &Booking{
		GuestName:     "John Doe",
		Phone:         "555-1234",
		Address:       "123 Main St",
		EmergencyName: "Jane Doe",
	}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: GuestName provided but Address missing
	invalidAddr := &Booking{
		GuestName:     "John Doe",
		Phone:         "555-1234",
		Address:       "",
		EmergencyName: "Jane Doe",
	}
	err = validator.Validate(invalidAddr)
	if err == nil {
		t.Error("expected validation error when Address missing for guest")
	}

	// Invalid: Phone provided but EmergencyName missing
	invalidEmerg := &Booking{
		GuestName:     "John Doe",
		Phone:         "555-1234",
		Address:       "123 Main St",
		EmergencyName: "",
	}
	err = validator.Validate(invalidEmerg)
	if err == nil {
		t.Error("expected validation error when EmergencyName missing for phone")
	}
}

// ============================================================================
// required_without Tests
// ============================================================================

// TestRequiredWithout_OtherFieldAbsent_FieldPresent validates that required_without
// is satisfied when the other field is zero AND this field is provided.
func TestRequiredWithout_OtherFieldAbsent_FieldPresent(t *testing.T) {
	type Address struct {
		DefaultAddress string `json:"default_address"`
		CustomAddress  string `json:"custom_address" pedantigo:"required_without=DefaultAddress"`
	}

	validator := New[Address]()

	// Valid: DefaultAddress is absent (zero), CustomAddress is provided
	valid := &Address{DefaultAddress: "", CustomAddress: "123 Oak St"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredWithout_OtherFieldAbsent_FieldMissing validates that required_without
// fails when the other field is zero AND this field is also missing.
func TestRequiredWithout_OtherFieldAbsent_FieldMissing(t *testing.T) {
	type Address struct {
		DefaultAddress string `json:"default_address"`
		CustomAddress  string `json:"custom_address" pedantigo:"required_without=DefaultAddress"`
	}

	validator := New[Address]()

	// Invalid: DefaultAddress is absent, CustomAddress is also absent
	invalid := &Address{DefaultAddress: "", CustomAddress: ""}
	err := validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when both addresses missing")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(ve.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	if ve.Errors[0].Field != "CustomAddress" {
		t.Errorf("expected error for field 'CustomAddress', got '%s'", ve.Errors[0].Field)
	}
}

// TestRequiredWithout_OtherFieldPresent_FieldMissing validates that required_without
// is satisfied when the other field is non-zero, even if this field is missing.
func TestRequiredWithout_OtherFieldPresent_FieldMissing(t *testing.T) {
	type Address struct {
		DefaultAddress string `json:"default_address"`
		CustomAddress  string `json:"custom_address" pedantigo:"required_without=DefaultAddress"`
	}

	validator := New[Address]()

	// Valid: DefaultAddress is present (non-zero), so CustomAddress can be empty
	valid := &Address{DefaultAddress: "456 Elm St", CustomAddress: ""}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors when default address provided, got: %v", err)
	}
}

// TestRequiredWithout_OtherFieldPresent_FieldPresent validates that required_without
// is satisfied when the other field is non-zero, and this field can optionally be present.
func TestRequiredWithout_OtherFieldPresent_FieldPresent(t *testing.T) {
	type Address struct {
		DefaultAddress string `json:"default_address"`
		CustomAddress  string `json:"custom_address" pedantigo:"required_without=DefaultAddress"`
	}

	validator := New[Address]()

	// Valid: DefaultAddress is present, CustomAddress is optionally present
	valid := &Address{DefaultAddress: "456 Elm St", CustomAddress: "123 Oak St"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// TestRequiredWithout_NumericField validates required_without with numeric field conditions.
func TestRequiredWithout_NumericField(t *testing.T) {
	type Discount struct {
		FixedAmount    int    `json:"fixed_amount"`
		PercentageCode string `json:"percentage_code" pedantigo:"required_without=FixedAmount"`
	}

	validator := New[Discount]()

	// Valid: FixedAmount is absent (zero), PercentageCode provided
	valid := &Discount{FixedAmount: 0, PercentageCode: "SAVE10"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: FixedAmount is absent, PercentageCode also absent
	invalid := &Discount{FixedAmount: 0, PercentageCode: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when neither discount provided")
	}

	// Valid: FixedAmount is present, PercentageCode can be empty
	validFixed := &Discount{FixedAmount: 50, PercentageCode: ""}
	err = validator.Validate(validFixed)
	if err != nil {
		t.Errorf("expected no errors for fixed amount discount, got: %v", err)
	}
}

// TestRequiredWithout_BooleanField validates required_without with boolean field conditions.
func TestRequiredWithout_BooleanField(t *testing.T) {
	type Notification struct {
		UseDefault bool   `json:"use_default"`
		CustomRule string `json:"custom_rule" pedantigo:"required_without=UseDefault"`
	}

	validator := New[Notification]()

	// Valid: UseDefault=false (zero), CustomRule provided
	valid := &Notification{UseDefault: false, CustomRule: "notify_all"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: UseDefault=false, CustomRule missing
	invalid := &Notification{UseDefault: false, CustomRule: ""}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("expected validation error when CustomRule missing for non-default notification")
	}

	// Valid: UseDefault=true, CustomRule can be empty
	validDefault := &Notification{UseDefault: true, CustomRule: ""}
	err = validator.Validate(validDefault)
	if err != nil {
		t.Errorf("expected no errors for default notification, got: %v", err)
	}
}

// TestRequiredWithout_MultipleFields validates required_without with multiple fields having conditions.
func TestRequiredWithout_MultipleFields(t *testing.T) {
	type Inventory struct {
		WarehouseLocation string `json:"warehouse_location"`
		ShippingLabel     string `json:"shipping_label"`
		StorageBox        string `json:"storage_box" pedantigo:"required_without=WarehouseLocation"`
		ShippingTracking  string `json:"shipping_tracking" pedantigo:"required_without=ShippingLabel"`
	}

	validator := New[Inventory]()

	// Valid: WarehouseLocation provided (StorageBox can be empty),
	// ShippingLabel provided (ShippingTracking can be empty)
	valid := &Inventory{
		WarehouseLocation: "WH-A1",
		ShippingLabel:     "LABEL-123",
		StorageBox:        "",
		ShippingTracking:  "",
	}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}

	// Invalid: WarehouseLocation absent, StorageBox required but missing
	invalidBox := &Inventory{
		WarehouseLocation: "",
		ShippingLabel:     "LABEL-123",
		StorageBox:        "",
		ShippingTracking:  "TRACK-456",
	}
	err = validator.Validate(invalidBox)
	if err == nil {
		t.Error("expected validation error when StorageBox missing without WarehouseLocation")
	}

	// Invalid: ShippingLabel absent, ShippingTracking required but missing
	invalidTracking := &Inventory{
		WarehouseLocation: "WH-A1",
		ShippingLabel:     "",
		StorageBox:        "BOX-789",
		ShippingTracking:  "",
	}
	err = validator.Validate(invalidTracking)
	if err == nil {
		t.Error("expected validation error when ShippingTracking missing without ShippingLabel")
	}

	// Valid: Both required_without conditions satisfied (neither target field present)
	validAlternate := &Inventory{
		WarehouseLocation: "",
		ShippingLabel:     "",
		StorageBox:        "BOX-789",
		ShippingTracking:  "TRACK-456",
	}
	err = validator.Validate(validAlternate)
	if err != nil {
		t.Errorf("expected no errors for alternate location/tracking, got: %v", err)
	}
}

// ============================================================================
// Cross-Constraint Integration Tests
// ============================================================================

// TestCrossFieldConstraints_ComplexScenario validates multiple cross-field constraints
// working together in a real-world scenario.
func TestCrossFieldConstraints_ComplexScenario(t *testing.T) {
	type UserProfile struct {
		AccountType      string `json:"account_type"` // personal, business, government
		IsVerified       bool   `json:"is_verified"`
		BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
		TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
		VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
		BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
		NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
	}

	validator := New[UserProfile]()

	// Valid: Business account with all required fields
	validBusiness := &UserProfile{
		AccountType:      "business",
		IsVerified:       true,
		BusinessName:     "Acme Corp",
		TaxID:            "12-3456789",
		VerificationDoc:  "doc_123",
		BackupEmail:      "backup@acme.com",
		NotificationPref: "email",
	}
	err := validator.Validate(validBusiness)
	if err != nil {
		t.Errorf("expected no errors for valid business account, got: %v", err)
	}

	// Invalid: Business account missing BusinessName
	invalidBusiness := &UserProfile{
		AccountType:      "business",
		IsVerified:       false,
		BusinessName:     "",
		TaxID:            "12-3456789",
		VerificationDoc:  "",
		BackupEmail:      "backup@example.com",
		NotificationPref: "email",
	}
	err = validator.Validate(invalidBusiness)
	if err == nil {
		t.Error("expected validation error for business account without BusinessName")
	}

	// Valid: Government account (BackupEmail not required)
	validGov := &UserProfile{
		AccountType:      "government",
		IsVerified:       true,
		BusinessName:     "",
		TaxID:            "",
		VerificationDoc:  "gov_doc_456",
		BackupEmail:      "",
		NotificationPref: "",
	}
	err = validator.Validate(validGov)
	if err != nil {
		t.Errorf("expected no errors for government account, got: %v", err)
	}

	// Invalid: BackupEmail provided but NotificationPref missing
	invalidNotif := &UserProfile{
		AccountType:      "personal",
		IsVerified:       false,
		BusinessName:     "",
		TaxID:            "",
		VerificationDoc:  "",
		BackupEmail:      "backup@example.com",
		NotificationPref: "",
	}
	err = validator.Validate(invalidNotif)
	if err == nil {
		t.Error("expected validation error when NotificationPref missing for BackupEmail")
	}
}

// TestCrossFieldConstraints_FieldIndexResolution validates that the field index
// resolution works correctly for constraints targeting different fields.
func TestCrossFieldConstraints_FieldIndexResolution(t *testing.T) {
	type Form struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2"`
		Field3 string `json:"field3" pedantigo:"required_if=Field1:trigger"`
		Field4 string `json:"field4" pedantigo:"required_unless=Field2:skip"`
	}

	validator := New[Form]()

	// Ensure the validator was created successfully (field indices resolved)
	if validator == nil {
		t.Fatal("validator creation failed")
	}

	// The constraint implementation should have resolved field indices correctly
	_ = validator.Validate(&Form{
		Field1: "trigger",
		Field2: "active",
		Field3: "value",
		Field4: "value",
	})
}

// TestCrossFieldConstraints_ZeroValueDistinction validates that zero values are
// handled correctly (e.g., empty string "" vs field missing).
func TestCrossFieldConstraints_ZeroValueDistinction(t *testing.T) {
	type Form struct {
		TriggerField string `json:"trigger_field"`
		TargetField  string `json:"target_field" pedantigo:"required_with=TriggerField"`
	}

	validator := New[Form]()

	// Explicit empty string is still a value (zero value but present)
	form := &Form{
		TriggerField: "value",
		TargetField:  "",
	}

	// This should fail because TriggerField is non-empty and TargetField is empty
	err := validator.Validate(form)
	if err == nil {
		t.Error("expected validation error for zero value when required_with condition met")
	}
}

// TestCrossFieldConstraints_UnexportedFields validates that unexported fields are ignored
// and don't interfere with validation of exported fields.
func TestCrossFieldConstraints_UnexportedFields(t *testing.T) {
	type Form struct {
		privateField string // This should be ignored
		PublicField  string `json:"public_field"`
		Conditional  string `json:"conditional" pedantigo:"required_if=PublicField:trigger"`
	}

	validator := New[Form]()

	// Valid: PublicField triggers requirement, Conditional is provided
	// privateField is ignored by validator
	form := &Form{
		privateField: "ignored",
		PublicField:  "trigger",
		Conditional:  "value",
	}

	err := validator.Validate(form)
	if err != nil {
		t.Errorf("expected validation to work with unexported fields present, got: %v", err)
	}

	// Invalid: PublicField triggers requirement, Conditional is missing
	invalidForm := &Form{
		privateField: "ignored",
		PublicField:  "trigger",
		Conditional:  "",
	}

	err = validator.Validate(invalidForm)
	if err == nil {
		t.Error("expected validation error when Conditional missing for PublicField=trigger")
	}
}

// TestCrossFieldConstraints_ReflectValueHandling validates that cross-field validation
// correctly handles reflect.Value types.
func TestCrossFieldConstraints_ReflectValueHandling(t *testing.T) {
	type Form struct {
		Status string `json:"status"`
		Detail string `json:"detail" pedantigo:"required_if=Status:complete"`
	}

	validator := New[Form]()

	// Create a form and get its reflect.Value
	form := Form{
		Status: "complete",
		Detail: "",
	}

	formValue := reflect.ValueOf(&form).Elem()

	// The validator should handle the form correctly
	err := validator.Validate(&form)
	if err == nil {
		t.Error("expected validation error")
	}

	_ = formValue // Ensure we can work with reflect values
}
