package constraints_test

import (
	"testing"
	"time"

	. "github.com/SmrutAI/Pedantigo"
)

// ==================================================
// eqfield (field equals another field) constraint tests
// ==================================================

func TestEqField_IntegerEquality_Valid(t *testing.T) {
	type PasswordChange struct {
		Password string `pedantigo:"required"`
		Confirm  string `pedantigo:"required,eqfield=Password"`
	}

	validator := New[PasswordChange]()
	data := PasswordChange{
		Password: "secret123",
		Confirm:  "secret123",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal passwords, got %v", err)
	}
}

func TestEqField_IntegerEquality_Invalid(t *testing.T) {
	type PasswordChange struct {
		Password string `pedantigo:"required"`
		Confirm  string `pedantigo:"required,eqfield=Password"`
	}

	validator := New[PasswordChange]()
	data := PasswordChange{
		Password: "secret123",
		Confirm:  "different456",
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for non-equal passwords")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Confirm" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error on Confirm field, got %v", ve.Errors)
	}
}

func TestEqField_StringEquality_Valid(t *testing.T) {
	type SignUp struct {
		Email      string `pedantigo:"required,email"`
		EmailCheck string `pedantigo:"required,email,eqfield=Email"`
	}

	validator := New[SignUp]()
	data := SignUp{
		Email:      "user@example.com",
		EmailCheck: "user@example.com",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal emails, got %v", err)
	}
}

func TestEqField_StringEquality_Invalid(t *testing.T) {
	type SignUp struct {
		Email      string `pedantigo:"required,email"`
		EmailCheck string `pedantigo:"required,email,eqfield=Email"`
	}

	validator := New[SignUp]()
	data := SignUp{
		Email:      "user@example.com",
		EmailCheck: "other@example.com",
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for non-equal emails")
	}
}

func TestEqField_IntValueEquality_Valid(t *testing.T) {
	type TransactionVerify struct {
		Amount  int `pedantigo:"gt=0"`
		Confirm int `pedantigo:"gt=0,eqfield=Amount"`
	}

	validator := New[TransactionVerify]()
	data := TransactionVerify{
		Amount:  10000,
		Confirm: 10000,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal amounts, got %v", err)
	}
}

func TestEqField_IntValueEquality_Invalid(t *testing.T) {
	type TransactionVerify struct {
		Amount  int `pedantigo:"gt=0"`
		Confirm int `pedantigo:"gt=0,eqfield=Amount"`
	}

	validator := New[TransactionVerify]()
	data := TransactionVerify{
		Amount:  10000,
		Confirm: 5000,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for non-equal amounts")
	}
}

func TestEqField_FloatEquality_Valid(t *testing.T) {
	type PriceMatch struct {
		ExpectedPrice float64 `pedantigo:"gt=0"`
		ActualPrice   float64 `pedantigo:"gt=0,eqfield=ExpectedPrice"`
	}

	validator := New[PriceMatch]()
	data := PriceMatch{
		ExpectedPrice: 99.99,
		ActualPrice:   99.99,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal prices, got %v", err)
	}
}

func TestEqField_FloatEquality_Invalid(t *testing.T) {
	type PriceMatch struct {
		ExpectedPrice float64 `pedantigo:"gt=0"`
		ActualPrice   float64 `pedantigo:"gt=0,eqfield=ExpectedPrice"`
	}

	validator := New[PriceMatch]()
	data := PriceMatch{
		ExpectedPrice: 99.99,
		ActualPrice:   100.00,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for non-equal prices")
	}
}

func TestEqField_ZeroValues_Valid(t *testing.T) {
	type ZeroComparison struct {
		Field1 int `pedantigo:"eqfield=Field2"`
		Field2 int `pedantigo:"eqfield=Field1"`
	}

	validator := New[ZeroComparison]()
	data := ZeroComparison{
		Field1: 0,
		Field2: 0,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal zero values, got %v", err)
	}
}

// ==================================================
// nefield (field not equal to another field) constraint tests
// ==================================================

func TestNeField_IntegerInequality_Valid(t *testing.T) {
	type DiscountCode struct {
		OriginalPrice   float64 `pedantigo:"gt=0"`
		DiscountedPrice float64 `pedantigo:"gt=0,nefield=OriginalPrice"`
	}

	validator := New[DiscountCode]()
	data := DiscountCode{
		OriginalPrice:   100.00,
		DiscountedPrice: 80.00,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for unequal prices, got %v", err)
	}
}

func TestNeField_IntegerInequality_Invalid(t *testing.T) {
	type DiscountCode struct {
		OriginalPrice   float64 `pedantigo:"gt=0"`
		DiscountedPrice float64 `pedantigo:"gt=0,nefield=OriginalPrice"`
	}

	validator := New[DiscountCode]()
	data := DiscountCode{
		OriginalPrice:   100.00,
		DiscountedPrice: 100.00,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for equal prices (should not be equal)")
	}
}

func TestNeField_StringInequality_Valid(t *testing.T) {
	type UniqueLogins struct {
		Username string `pedantigo:"required,min=3"`
		Nickname string `pedantigo:"required,min=1,nefield=Username"`
	}

	validator := New[UniqueLogins]()
	data := UniqueLogins{
		Username: "john_doe",
		Nickname: "Johnny",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for unequal strings, got %v", err)
	}
}

func TestNeField_StringInequality_Invalid(t *testing.T) {
	type UniqueLogins struct {
		Username string `pedantigo:"required,min=3"`
		Nickname string `pedantigo:"required,min=1,nefield=Username"`
	}

	validator := New[UniqueLogins]()
	data := UniqueLogins{
		Username: "john_doe",
		Nickname: "john_doe",
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for equal strings (should not be equal)")
	}
}

func TestNeField_ZeroValues_Invalid(t *testing.T) {
	type UnequalCheck struct {
		Field1 int `pedantigo:"nefield=Field2"`
		Field2 int `pedantigo:"nefield=Field1"`
	}

	validator := New[UnequalCheck]()
	data := UnequalCheck{
		Field1: 0,
		Field2: 0,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for equal zero values (should not be equal)")
	}
}

func TestNeField_ZeroValues_Valid(t *testing.T) {
	type UnequalCheck struct {
		Field1 int `pedantigo:"nefield=Field2"`
		Field2 int `pedantigo:"nefield=Field1"`
	}

	validator := New[UnequalCheck]()
	data := UnequalCheck{
		Field1: 0,
		Field2: 5,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for unequal values, got %v", err)
	}
}

// ==================================================
// gtfield (field greater than another field) constraint tests
// ==================================================

func TestGtField_IntegerGreaterThan_Valid(t *testing.T) {
	type DateRange struct {
		StartYear int `pedantigo:"min=1900"`
		EndYear   int `pedantigo:"min=1900,gtfield=StartYear"`
	}

	validator := New[DateRange]()
	data := DateRange{
		StartYear: 2000,
		EndYear:   2024,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for EndYear > StartYear, got %v", err)
	}
}

func TestGtField_IntegerGreaterThan_Invalid_Equal(t *testing.T) {
	type DateRange struct {
		StartYear int `pedantigo:"min=1900"`
		EndYear   int `pedantigo:"min=1900,gtfield=StartYear"`
	}

	validator := New[DateRange]()
	data := DateRange{
		StartYear: 2000,
		EndYear:   2000,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for EndYear == StartYear (not greater)")
	}
}

func TestGtField_IntegerGreaterThan_Invalid_Less(t *testing.T) {
	type DateRange struct {
		StartYear int `pedantigo:"min=1900"`
		EndYear   int `pedantigo:"min=1900,gtfield=StartYear"`
	}

	validator := New[DateRange]()
	data := DateRange{
		StartYear: 2024,
		EndYear:   2000,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for EndYear < StartYear")
	}
}

func TestGtField_FloatGreaterThan_Valid(t *testing.T) {
	type PriceComparison struct {
		MinPrice float64 `pedantigo:"gt=0"`
		MaxPrice float64 `pedantigo:"gt=0,gtfield=MinPrice"`
	}

	validator := New[PriceComparison]()
	data := PriceComparison{
		MinPrice: 50.00,
		MaxPrice: 150.00,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MaxPrice > MinPrice, got %v", err)
	}
}

func TestGtField_FloatGreaterThan_Invalid(t *testing.T) {
	type PriceComparison struct {
		MinPrice float64 `pedantigo:"gt=0"`
		MaxPrice float64 `pedantigo:"gt=0,gtfield=MinPrice"`
	}

	validator := New[PriceComparison]()
	data := PriceComparison{
		MinPrice: 150.00,
		MaxPrice: 50.00,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for MaxPrice < MinPrice")
	}
}

func TestGtField_StringGreaterThan_Valid(t *testing.T) {
	type StringComparison struct {
		FirstName  string `pedantigo:"required"`
		SecondName string `pedantigo:"required,gtfield=FirstName"`
	}

	validator := New[StringComparison]()
	data := StringComparison{
		FirstName:  "Alice",
		SecondName: "Bob",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for 'Bob' > 'Alice' lexicographically, got %v", err)
	}
}

func TestGtField_StringGreaterThan_Invalid(t *testing.T) {
	type StringComparison struct {
		FirstName  string `pedantigo:"required"`
		SecondName string `pedantigo:"required,gtfield=FirstName"`
	}

	validator := New[StringComparison]()
	data := StringComparison{
		FirstName:  "Zebra",
		SecondName: "Apple",
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for 'Apple' < 'Zebra' lexicographically")
	}
}

func TestGtField_ZeroValues_Invalid(t *testing.T) {
	type ZeroGreaterCheck struct {
		Field1 int `pedantigo:"gtfield=Field2"`
		Field2 int
	}

	validator := New[ZeroGreaterCheck]()
	data := ZeroGreaterCheck{
		Field1: 0,
		Field2: 0,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for 0 not > 0")
	}
}

func TestGtField_NegativeNumbers_Valid(t *testing.T) {
	type NegativeComparison struct {
		Lower int `pedantigo:"gtfield=Upper"`
		Upper int
	}

	validator := New[NegativeComparison]()
	data := NegativeComparison{
		Lower: -5,
		Upper: -10,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for -5 > -10, got %v", err)
	}
}

// ==================================================
// gtefield (field greater than or equal to another field) constraint tests
// ==================================================

func TestGteField_IntegerGreaterOrEqual_Valid_Greater(t *testing.T) {
	type VersionComparison struct {
		MinVersion int `pedantigo:"gt=0"`
		MaxVersion int `pedantigo:"gt=0,gtefield=MinVersion"`
	}

	validator := New[VersionComparison]()
	data := VersionComparison{
		MinVersion: 1,
		MaxVersion: 2,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MaxVersion > MinVersion, got %v", err)
	}
}

func TestGteField_IntegerGreaterOrEqual_Valid_Equal(t *testing.T) {
	type VersionComparison struct {
		MinVersion int `pedantigo:"gt=0"`
		MaxVersion int `pedantigo:"gt=0,gtefield=MinVersion"`
	}

	validator := New[VersionComparison]()
	data := VersionComparison{
		MinVersion: 1,
		MaxVersion: 1,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MaxVersion == MinVersion, got %v", err)
	}
}

func TestGteField_IntegerGreaterOrEqual_Invalid(t *testing.T) {
	type VersionComparison struct {
		MinVersion int `pedantigo:"gt=0"`
		MaxVersion int `pedantigo:"gt=0,gtefield=MinVersion"`
	}

	validator := New[VersionComparison]()
	data := VersionComparison{
		MinVersion: 2,
		MaxVersion: 1,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for MaxVersion < MinVersion")
	}
}

func TestGteField_FloatGreaterOrEqual_Valid_Equal(t *testing.T) {
	type ScoreComparison struct {
		MinScore float64 `pedantigo:"gte=0"`
		MaxScore float64 `pedantigo:"gte=0,gtefield=MinScore"`
	}

	validator := New[ScoreComparison]()
	data := ScoreComparison{
		MinScore: 75.5,
		MaxScore: 75.5,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal scores, got %v", err)
	}
}

func TestGteField_StringGreaterOrEqual_Valid_Equal(t *testing.T) {
	type StringComparison struct {
		StartStr string `pedantigo:"required"`
		EndStr   string `pedantigo:"required,gtefield=StartStr"`
	}

	validator := New[StringComparison]()
	data := StringComparison{
		StartStr: "test",
		EndStr:   "test",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal strings, got %v", err)
	}
}

func TestGteField_ZeroValues_Valid(t *testing.T) {
	type ZeroGreaterOrEqualCheck struct {
		Field1 int `pedantigo:"gtefield=Field2"`
		Field2 int
	}

	validator := New[ZeroGreaterOrEqualCheck]()
	data := ZeroGreaterOrEqualCheck{
		Field1: 0,
		Field2: 0,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for 0 >= 0, got %v", err)
	}
}

// ==================================================
// ltfield (field less than another field) constraint tests
// ==================================================

func TestLtField_IntegerLessThan_Valid(t *testing.T) {
	type AgeRange struct {
		MinAge int `pedantigo:"gt=0,ltfield=MaxAge"`
		MaxAge int `pedantigo:"gt=0"`
	}

	validator := New[AgeRange]()
	data := AgeRange{
		MinAge: 18,
		MaxAge: 65,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MinAge < MaxAge, got %v", err)
	}
}

func TestLtField_IntegerLessThan_Invalid_Equal(t *testing.T) {
	// Create a custom struct with ltfield constraint
	type AgeRangeWithLt struct {
		MinAge int `pedantigo:"gt=0,ltfield=MaxAge"`
		MaxAge int `pedantigo:"gt=0"`
	}

	validatorLt := New[AgeRangeWithLt]()
	dataLt := AgeRangeWithLt{
		MinAge: 30,
		MaxAge: 30,
	}

	err := validatorLt.Validate(&dataLt)
	if err == nil {
		t.Fatal("expected validation error for MinAge == MaxAge (not less)")
	}
}

func TestLtField_IntegerLessThan_Invalid_Greater(t *testing.T) {
	type AgeRange struct {
		MinAge int `pedantigo:"gt=0,ltfield=MaxAge"`
		MaxAge int `pedantigo:"gt=0"`
	}

	validator := New[AgeRange]()
	data := AgeRange{
		MinAge: 65,
		MaxAge: 18,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for MinAge > MaxAge")
	}
}

func TestLtField_FloatLessThan_Valid(t *testing.T) {
	type TemperatureRange struct {
		MinTemp float64 `pedantigo:"ltfield=MaxTemp"`
		MaxTemp float64
	}

	validator := New[TemperatureRange]()
	data := TemperatureRange{
		MinTemp: -10.5,
		MaxTemp: 40.0,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MinTemp < MaxTemp, got %v", err)
	}
}

func TestLtField_FloatLessThan_Invalid(t *testing.T) {
	type TemperatureRange struct {
		MinTemp float64 `pedantigo:"ltfield=MaxTemp"`
		MaxTemp float64
	}

	validator := New[TemperatureRange]()
	data := TemperatureRange{
		MinTemp: 50.0,
		MaxTemp: 40.0,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for MinTemp > MaxTemp")
	}
}

func TestLtField_StringLessThan_Valid(t *testing.T) {
	type StringOrder struct {
		First  string `pedantigo:"required,ltfield=Second"`
		Second string `pedantigo:"required"`
	}

	validator := New[StringOrder]()
	data := StringOrder{
		First:  "Apple",
		Second: "Zebra",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for 'Apple' < 'Zebra' lexicographically, got %v", err)
	}
}

func TestLtField_StringLessThan_Invalid(t *testing.T) {
	type StringOrder struct {
		First  string `pedantigo:"required,ltfield=Second"`
		Second string `pedantigo:"required"`
	}

	validator := New[StringOrder]()
	data := StringOrder{
		First:  "Zebra",
		Second: "Apple",
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for 'Zebra' > 'Apple' lexicographically")
	}
}

func TestLtField_ZeroValues_Invalid(t *testing.T) {
	type ZeroLessCheck struct {
		Field1 int `pedantigo:"ltfield=Field2"`
		Field2 int
	}

	validator := New[ZeroLessCheck]()
	data := ZeroLessCheck{
		Field1: 0,
		Field2: 0,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for 0 not < 0")
	}
}

func TestLtField_NegativeNumbers_Valid(t *testing.T) {
	type NegativeComparison struct {
		Lower int `pedantigo:"ltfield=Upper"`
		Upper int
	}

	validator := New[NegativeComparison]()
	data := NegativeComparison{
		Lower: -10,
		Upper: -5,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for -10 < -5, got %v", err)
	}
}

// ==================================================
// ltefield (field less than or equal to another field) constraint tests
// ==================================================

func TestLteField_IntegerLessOrEqual_Valid_Less(t *testing.T) {
	type RankRange struct {
		StartRank int `pedantigo:"gt=0,ltefield=EndRank"`
		EndRank   int `pedantigo:"gt=0"`
	}

	validator := New[RankRange]()
	data := RankRange{
		StartRank: 10,
		EndRank:   20,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for StartRank < EndRank, got %v", err)
	}
}

func TestLteField_IntegerLessOrEqual_Valid_Equal(t *testing.T) {
	type RankRange struct {
		StartRank int `pedantigo:"gt=0,ltefield=EndRank"`
		EndRank   int `pedantigo:"gt=0"`
	}

	validator := New[RankRange]()
	data := RankRange{
		StartRank: 15,
		EndRank:   15,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for StartRank == EndRank, got %v", err)
	}
}

func TestLteField_IntegerLessOrEqual_Invalid(t *testing.T) {
	type RankRange struct {
		StartRank int `pedantigo:"gt=0,ltefield=EndRank"`
		EndRank   int `pedantigo:"gt=0"`
	}

	validator := New[RankRange]()
	data := RankRange{
		StartRank: 25,
		EndRank:   15,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for StartRank > EndRank")
	}
}

func TestLteField_FloatLessOrEqual_Valid_Equal(t *testing.T) {
	type PriceRange struct {
		MinPrice float64 `pedantigo:"gt=0,ltefield=MaxPrice"`
		MaxPrice float64 `pedantigo:"gt=0"`
	}

	validator := New[PriceRange]()
	data := PriceRange{
		MinPrice: 50.00,
		MaxPrice: 50.00,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal prices, got %v", err)
	}
}

func TestLteField_StringLessOrEqual_Valid_Equal(t *testing.T) {
	type StringBoundary struct {
		Start string `pedantigo:"required,ltefield=End"`
		End   string `pedantigo:"required"`
	}

	validator := New[StringBoundary]()
	data := StringBoundary{
		Start: "same",
		End:   "same",
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for equal strings, got %v", err)
	}
}

func TestLteField_ZeroValues_Valid(t *testing.T) {
	type ZeroLessOrEqualCheck struct {
		Field1 int `pedantigo:"ltefield=Field2"`
		Field2 int
	}

	validator := New[ZeroLessOrEqualCheck]()
	data := ZeroLessOrEqualCheck{
		Field1: 0,
		Field2: 0,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for 0 <= 0, got %v", err)
	}
}

// ==================================================
// Cross-constraint combinations and edge cases
// ==================================================

func TestCrossFieldComparison_MultipleConstraints(t *testing.T) {
	type TransactionBounds struct {
		MinAmount    float64 `pedantigo:"gt=0"`
		MaxAmount    float64 `pedantigo:"gt=0,gtfield=MinAmount"`
		ActualAmount float64 `pedantigo:"gt=0,gtefield=MinAmount,ltefield=MaxAmount"`
	}

	validator := New[TransactionBounds]()
	data := TransactionBounds{
		MinAmount:    10.0,
		MaxAmount:    100.0,
		ActualAmount: 50.0,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for ActualAmount within bounds, got %v", err)
	}
}

func TestCrossFieldComparison_MultipleConstraints_Invalid(t *testing.T) {
	type TransactionBounds struct {
		MinAmount    float64 `pedantigo:"gt=0"`
		MaxAmount    float64 `pedantigo:"gt=0,gtfield=MinAmount"`
		ActualAmount float64 `pedantigo:"gt=0,gtefield=MinAmount,ltefield=MaxAmount"`
	}

	validator := New[TransactionBounds]()
	data := TransactionBounds{
		MinAmount:    10.0,
		MaxAmount:    100.0,
		ActualAmount: 150.0,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for ActualAmount > MaxAmount")
	}
}

func TestCrossFieldComparison_ThreeWayDependency(t *testing.T) {
	type OrderValidation struct {
		OrderTotal     float64 `pedantigo:"gt=0"`
		DiscountAmount float64 `pedantigo:"gte=0,ltefield=OrderTotal"`
		FinalAmount    float64 `pedantigo:"gt=0,ltfield=OrderTotal,gtfield=DiscountAmount"`
	}

	validator := New[OrderValidation]()
	data := OrderValidation{
		OrderTotal:     100.0,
		DiscountAmount: 20.0,
		FinalAmount:    80.0,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for valid order calculation, got %v", err)
	}
}

func TestCrossFieldComparison_TimeValues_IfSupported(t *testing.T) {
	type EventSchedule struct {
		StartTime time.Time
		EndTime   time.Time
	}

	validator := New[EventSchedule]()
	now := time.Now()
	later := now.Add(2 * time.Hour)

	data := EventSchedule{
		StartTime: now,
		EndTime:   later,
	}

	// This test documents time.Time comparison capability
	// Expected behavior: validation should pass if EndTime > StartTime
	// Currently this may not work if cross-field constraints don't support time.Time
	err := validator.Validate(&data)
	_ = err // Just document the behavior
}

func TestCrossFieldComparison_UintComparison_Valid(t *testing.T) {
	type PortRange struct {
		MinPort uint `pedantigo:"gt=0,ltfield=MaxPort"`
		MaxPort uint `pedantigo:"gt=0"`
	}

	validator := New[PortRange]()
	data := PortRange{
		MinPort: 8000,
		MaxPort: 9000,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MinPort < MaxPort, got %v", err)
	}
}

func TestCrossFieldComparison_UintComparison_Invalid(t *testing.T) {
	type PortRange struct {
		MinPort uint `pedantigo:"gt=0,ltfield=MaxPort"`
		MaxPort uint `pedantigo:"gt=0"`
	}

	validator := New[PortRange]()
	data := PortRange{
		MinPort: 9000,
		MaxPort: 8000,
	}

	err := validator.Validate(&data)
	if err == nil {
		t.Fatal("expected validation error for MinPort > MaxPort")
	}
}

func TestCrossFieldComparison_Int32Comparison_Valid(t *testing.T) {
	type Int32Range struct {
		Start int32 `pedantigo:"ltfield=End"`
		End   int32
	}

	validator := New[Int32Range]()
	data := Int32Range{
		Start: 100,
		End:   200,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for Start < End, got %v", err)
	}
}

func TestCrossFieldComparison_Float32Comparison_Valid(t *testing.T) {
	type Float32Range struct {
		MinVal float32 `pedantigo:"ltfield=MaxVal"`
		MaxVal float32
	}

	validator := New[Float32Range]()
	data := Float32Range{
		MinVal: 1.5,
		MaxVal: 2.5,
	}

	err := validator.Validate(&data)
	if err != nil {
		t.Errorf("expected no validation errors for MinVal < MaxVal, got %v", err)
	}
}
