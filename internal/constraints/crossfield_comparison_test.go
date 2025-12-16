package constraints_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SmrutAI/pedantigo"
)

// TestCrossFieldComparison tests all cross-field comparison constraints.
// This test covers eqfield, nefield, gtfield, gtefield, ltfield, ltefield.
func TestCrossFieldComparison_BasicOperators(t *testing.T) {
	tests := []struct {
		name      string
		testFunc  func(t *testing.T)
		expectErr bool
	}{
		// eqfield (field equals another field) tests
		{
			name: "eqfield: string equality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "eqfield: string equality invalid",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err)

				var ve *ValidationError
				require.ErrorAs(t, err, &ve)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Confirm" {
						foundError = true
					}
				}
				assert.True(t, foundError)
			},
			expectErr: true,
		},
		{
			name: "eqfield: email equality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "eqfield: int equality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "eqfield: float equality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "eqfield: zero values equal valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},

		// nefield (field not equal to another field) tests
		{
			name: "nefield: float inequality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "nefield: float inequality invalid",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err)
			},
			expectErr: true,
		},
		{
			name: "nefield: string inequality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "nefield: zero values inequality invalid",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err)
			},
			expectErr: true,
		},
		{
			name: "nefield: zero values inequality valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},

		// gtfield (field greater than another field) tests
		{
			name: "gtfield: int greater than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtfield: int greater than invalid equal",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err)
			},
			expectErr: true,
		},
		{
			name: "gtfield: float greater than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtfield: string greater than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtfield: negative numbers greater than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},

		// gtefield (field greater than or equal to another field) tests
		{
			name: "gtefield: int greater or equal valid greater",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtefield: int greater or equal valid equal",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtefield: int greater or equal invalid",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err)
			},
			expectErr: true,
		},
		{
			name: "gtefield: float greater or equal valid equal",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtefield: string greater or equal valid equal",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "gtefield: zero values greater or equal valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},

		// ltfield (field less than another field) tests
		{
			name: "ltfield: int less than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltfield: int less than invalid equal",
			testFunc: func(t *testing.T) {
				type AgeRangeWithLt struct {
					MinAge int `pedantigo:"gt=0,ltfield=MaxAge"`
					MaxAge int `pedantigo:"gt=0"`
				}

				validator := New[AgeRangeWithLt]()
				data := AgeRangeWithLt{
					MinAge: 30,
					MaxAge: 30,
				}

				err := validator.Validate(&data)
				require.Error(t, err)
			},
			expectErr: true,
		},
		{
			name: "ltfield: float less than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltfield: string less than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltfield: negative numbers less than valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},

		// ltefield (field less than or equal to another field) tests
		{
			name: "ltefield: int less or equal valid less",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltefield: int less or equal valid equal",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltefield: float less or equal valid equal",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltefield: string less or equal valid equal",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
		{
			name: "ltefield: zero values less or equal valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// ==================================================
// Cross-constraint combinations and edge cases
// ==================================================

// TestCrossFieldComparison tests CrossFieldComparison validation.
func TestCrossFieldComparison(t *testing.T) {
	tests := []struct {
		name      string
		testFunc  func(t *testing.T)
		expectErr bool
	}{
		{
			name: "multiple constraints valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err, "expected no validation errors for ActualAmount within bounds")
			},
			expectErr: false,
		},
		{
			name: "multiple constraints invalid",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err, "expected validation error for ActualAmount > MaxAmount")
			},
			expectErr: true,
		},
		{
			name: "three way dependency valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err, "expected no validation errors for valid order calculation")
			},
			expectErr: false,
		},
		{
			name: "time values if supported",
			testFunc: func(t *testing.T) {
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

				err := validator.Validate(&data)
				// Test just validates that no panic occurs
				_ = err
			},
			expectErr: false,
		},
		{
			name: "uint comparison valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err, "expected no validation errors for MinPort < MaxPort")
			},
			expectErr: false,
		},
		{
			name: "uint comparison invalid",
			testFunc: func(t *testing.T) {
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
				require.Error(t, err, "expected validation error for MinPort > MaxPort")
			},
			expectErr: true,
		},
		{
			name: "int32 comparison valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err, "expected no validation errors for Start < End")
			},
			expectErr: false,
		},
		{
			name: "float32 comparison valid",
			testFunc: func(t *testing.T) {
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
				assert.NoError(t, err, "expected no validation errors for MinVal < MaxVal")
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
