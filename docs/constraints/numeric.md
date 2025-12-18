---
sidebar_position: 2
---

# Numeric Constraints

Numeric validation rules in Pedantigo for validating integers, floating-point numbers, and other numeric types.

## Range Constraints

### `min` / `max`

Validates the **numeric value** is within the specified range (inclusive).

```go
type Product struct {
    // Price must be at least 0.01 and at most 999,999.99
    Price float64 `json:"price" pedantigo:"required,min=0.01,max=999999.99"`
    // Quantity must be between 1 and 1000
    Quantity int `json:"quantity" pedantigo:"required,min=1,max=1000"`
}
```

**Behavior:**
- Works with all numeric types (int, int8-64, uint, uint8-64, float32, float64)
- Both bounds are inclusive
- Empty/zero values skip validation (only `required` constraint enforces non-zero)
- Comparison uses numeric value, not string representation

### `gt` / `gte` / `lt` / `lte`

Validates numeric values with exclusive or inclusive comparison bounds.

- `gt=N` - Greater than (exclusive, value must be greater than N)
- `gte=N` - Greater than or equal (inclusive, value must be at least N)
- `lt=N` - Less than (exclusive, value must be less than N)
- `lte=N` - Less than or equal (inclusive, value must be at most N)

```go
type SurveyRating struct {
    // Rating must be strictly greater than 0
    Score int `json:"score" pedantigo:"required,gt=0"`
    // Confidence must be at least 0.5
    Confidence float64 `json:"confidence" pedantigo:"required,gte=0.5"`
    // Days until expiration must be less than 365
    DaysUntilExpiry int `json:"days_until_expiry" pedantigo:"required,lt=365"`
    // Discount percentage must be at most 50
    DiscountPercent int `json:"discount_percent" pedantigo:"required,lte=50"`
}
```

**Behavior:**
- `gt` and `lt` exclude the boundary value
- `gte` and `lte` include the boundary value
- Works with all numeric types
- Useful for enforcing strict domain rules

## Sign Constraints

### `positive`

Validates that a numeric value is **strictly greater than zero** (> 0).

```go
type BankAccount struct {
    // Balance can be positive (credits) or zero, but not negative
    Balance float64 `json:"balance" pedantigo:"required,positive"`
    // Withdrawal amount must be positive
    WithdrawalAmount float64 `json:"withdrawal_amount" pedantigo:"required,positive"`
}
```

**Valid examples:** 1, 0.01, 1000.50, 999999
**Invalid examples:** 0, -1, -0.01

**Note:** Use `gte=0` if you want to allow zero values.

### `negative`

Validates that a numeric value is **strictly less than zero** (< 0).

```go
type TemperatureReading struct {
    // Temperature below freezing point
    BelowFreezing float64 `json:"below_freezing" pedantigo:"negative"`
    // Temperature change (decrease)
    TemperatureChange float64 `json:"temperature_change" pedantigo:"negative"`
}
```

**Valid examples:** -1, -0.01, -273.15, -999
**Invalid examples:** 0, 1, 0.01

**Note:** Use `lte=0` if you want to allow zero values.

## Divisibility

### `multiple_of`

Validates that a numeric value is **evenly divisible** by the specified number.

```go
type TimeSlot struct {
    // Meeting duration in minutes must be divisible by 15
    DurationMinutes int `json:"duration_minutes" pedantigo:"required,multiple_of=15"`
}

type ShoppingCart struct {
    // Pack size: items must be divisible by 6 (half-dozen)
    Quantity int `json:"quantity" pedantigo:"required,min=6,multiple_of=6"`
}

type PercentageDiscount struct {
    // Discount must be in 5% increments
    DiscountPercent float64 `json:"discount_percent" pedantigo:"required,min=0,max=100,multiple_of=5"`
}
```

**Valid examples** (for `multiple_of=5`): 0, 5, 10, 15, 20, 100
**Invalid examples** (for `multiple_of=5`): 1, 3, 7, 22, 103

**Behavior:**
- Uses modulo operator (value % N == 0)
- Works with integers and floating-point numbers
- For floats, validates with high precision to handle floating-point arithmetic quirks

## Precision Constraints

### `max_digits`

Validates that a numeric value has **at most N total digits** (excluding the decimal point and sign).

```go
type ProductSKU struct {
    // Product code: maximum 8 digits
    Code int64 `json:"code" pedantigo:"required,max_digits=8"`
}

type BudgetAmount struct {
    // Budget: at most 10 total digits (e.g., 9,999,999,999.99)
    Amount float64 `json:"amount" pedantigo:"required,max_digits=10"`
}
```

**Examples** (with `max_digits=6`):
- Valid: 123456 (6 digits), 12345 (5 digits), 123.45 (5 digits)
- Invalid: 1234567 (7 digits), 123456.78 (8 digits)

**Behavior:**
- Counts only the numeric digits
- Ignores decimal point, sign, and exponent notation
- Useful for database column constraints or API limits

### `decimal_places`

Validates that a numeric value has **at most N decimal places** (digits after decimal point).

```go
type MoneyAmount struct {
    // Price in USD: maximum 2 decimal places
    Price float64 `json:"price" pedantigo:"required,decimal_places=2"`
}

type ExchangeRate struct {
    // Exchange rate: at most 6 decimal places
    Rate float64 `json:"rate" pedantigo:"required,min=0,decimal_places=6"`
}

type MeasurementValue struct {
    // Measurement: at most 4 decimal places
    Meters float64 `json:"meters" pedantigo:"required,decimal_places=4"`
}
```

**Examples** (with `decimal_places=2`):
- Valid: 10.50, 5.1, 100 (0 decimals), 0.05
- Invalid: 10.505 (3 decimals), 5.123 (3 decimals)

**Behavior:**
- Counts digits after the decimal point
- Integers (no decimal point) are considered to have 0 decimal places and always pass
- Useful for currency, measurements, and financial calculations
- Works with both `int` (implicitly 0 decimals) and `float64`

## Special Values

### `disallow_inf_nan`

Validates that a floating-point value is **neither infinity nor NaN** (Not a Number).

```go
type SensorData struct {
    // Sensor reading must be a valid number
    Temperature float64 `json:"temperature" pedantigo:"required,disallow_inf_nan"`
    // Altitude must be finite
    Altitude float64 `json:"altitude" pedantigo:"required,disallow_inf_nan"`
}

type CalculationResult struct {
    // Result of division must be a valid number
    Result float64 `json:"result" pedantigo:"required,disallow_inf_nan"`
}
```

**Valid examples:** 0.0, 1.5, -273.15, 999999.99, -0.00001
**Invalid examples:** math.Inf(1), math.Inf(-1), math.NaN()

**Behavior:**
- Only applies to float32 and float64 types
- Rejects positive infinity, negative infinity, and NaN
- Useful when calculations might produce invalid floating-point values
- Common after division operations or mathematical functions

## Combining Numeric Constraints

Numeric constraints can be combined to create powerful validation rules:

```go
type BankAccount struct {
    // Account balance: must be at least 0, at most 999 billion
    Balance float64 `json:"balance" pedantigo:"required,min=0,max=999000000000,decimal_places=2"`
}

type AgeVerification struct {
    // Age: must be between 18 and 150
    Age int `json:"age" pedantigo:"required,min=18,max=150"`
}

type PercentageScore struct {
    // Score: 0-100 with one decimal place
    Score float64 `json:"score" pedantigo:"required,min=0,max=100,decimal_places=1"`
}

type RatingScale struct {
    // 5-star rating: 1 to 5, in increments of 0.5
    Rating float64 `json:"rating" pedantigo:"required,min=1,max=5,multiple_of=0.5"`
}

type PaymentAmount struct {
    // Payment: at least 0.01, at most 999,999.99, exactly 2 decimal places
    Amount float64 `json:"amount" pedantigo:"required,min=0.01,max=999999.99,decimal_places=2"`
}

type ApiKeyID struct {
    // API key ID: exactly 32 digits, no sign
    KeyID int64 `json:"key_id" pedantigo:"required,max_digits=32"`
}
```

## Complete Example

Here's a comprehensive example showing multiple numeric constraints working together:

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type ECommerceProduct struct {
    // Product ID: positive integer, at most 10 digits
    ID int64 `json:"id" pedantigo:"required,positive,max_digits=10"`

    // Product name: required, 3-100 characters
    Name string `json:"name" pedantigo:"required,min=3,max=100"`

    // Price in USD: required, at least 0.01, exactly 2 decimal places
    Price float64 `json:"price" pedantigo:"required,min=0.01,decimal_places=2"`

    // Cost basis: optional, for internal tracking
    CostBasis float64 `json:"cost_basis,omitempty" pedantigo:"min=0,decimal_places=2"`

    // Stock quantity: non-negative, less than 1 million
    StockQuantity int `json:"stock_quantity" pedantigo:"required,min=0,max=999999"`

    // Rating: 0-5 stars, one decimal place
    Rating float64 `json:"rating,omitempty" pedantigo:"min=0,max=5,decimal_places=1"`

    // Review count: non-negative
    ReviewCount int `json:"review_count,omitempty" pedantigo:"min=0"`

    // Discount percentage: 0-100, whole numbers only
    DiscountPercent int `json:"discount_percent,omitempty" pedantigo:"min=0,max=100"`

    // Weight in kilograms: positive, 4 decimal places
    WeightKg float64 `json:"weight_kg,omitempty" pedantigo:"positive,decimal_places=4"`

    // Dimensions in centimeters: all positive
    WidthCm  float64 `json:"width_cm,omitempty" pedantigo:"positive,decimal_places=1"`
    HeightCm float64 `json:"height_cm,omitempty" pedantigo:"positive,decimal_places=1"`
    DepthCm  float64 `json:"depth_cm,omitempty" pedantigo:"positive,decimal_places=1"`

    // Warranty period in months: 0 to 36, multiple of 3
    WarrantyMonths int `json:"warranty_months,omitempty" pedantigo:"min=0,max=36,multiple_of=3"`
}

func main() {
    // Valid product
    productJSON := []byte(`{
        "id": 123456789,
        "name": "Professional Laptop Stand",
        "price": 49.99,
        "cost_basis": 20.00,
        "stock_quantity": 150,
        "rating": 4.5,
        "review_count": 237,
        "discount_percent": 10,
        "weight_kg": 2.5000,
        "width_cm": 45.5,
        "height_cm": 12.0,
        "depth_cm": 20.5,
        "warranty_months": 24
    }`)

    product, err := pedantigo.Unmarshal[ECommerceProduct](productJSON)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("Valid product: %+v\n", product)

    // Invalid product - multiple constraint violations
    invalidJSON := []byte(`{
        "id": 0,
        "name": "Laptop Stand",
        "price": 49.999,
        "stock_quantity": 1000000,
        "rating": 5.5,
        "review_count": -10,
        "discount_percent": 150,
        "weight_kg": -2.5,
        "width_cm": 45.123,
        "warranty_months": 25
    }`)

    _, err = pedantigo.Unmarshal[ECommerceProduct](invalidJSON)
    if err != nil {
        fmt.Printf("Validation errors:\n%v\n", err)
        // Output will show all numeric constraint violations:
        // - id: must be greater than 0
        // - price: can have at most 2 decimal places
        // - stock_quantity: must be at most 999999
        // - rating: must be at most 5
        // - review_count: must be at least 0
        // - discount_percent: must be at most 100
        // - weight_kg: must be greater than 0
        // - width_cm: can have at most 1 decimal place
        // - warranty_months: must be a multiple of 3
    }
}
```

## Validation Behavior Notes

- **Zero values:** Most numeric constraints skip validation for zero. Use `required` to enforce non-zero values, or `positive`/`negative` for sign constraints.
- **Type safety:** Constraints work with all numeric types (signed/unsigned integers, floating-point numbers).
- **Precision:** Floating-point comparisons use standard Go arithmetic; be cautious with `decimal_places` and extremely small values due to floating-point precision limits.
- **Nil values:** Nil pointers are skipped in validation; use `required` to enforce non-nil.
- **Infinity and NaN:** Only float32 and float64 can represent infinity or NaN. Use `disallow_inf_nan` to reject these values.

## Quick Reference Table

| Constraint | Example | Description | Type Support |
|------------|---------|-------------|--------------|
| `min=N` | `min=0` | Minimum value (inclusive) | All numeric |
| `max=N` | `max=100` | Maximum value (inclusive) | All numeric |
| `gt=N` | `gt=0` | Greater than (exclusive) | All numeric |
| `gte=N` | `gte=1` | Greater than or equal (inclusive) | All numeric |
| `lt=N` | `lt=100` | Less than (exclusive) | All numeric |
| `lte=N` | `lte=99` | Less than or equal (inclusive) | All numeric |
| `positive` | `positive` | Must be > 0 | All numeric |
| `negative` | `negative` | Must be < 0 | All numeric |
| `multiple_of=N` | `multiple_of=5` | Must be divisible by N | All numeric |
| `max_digits=N` | `max_digits=8` | Maximum total digits | All numeric |
| `decimal_places=N` | `decimal_places=2` | Maximum decimal places | float32, float64 |
| `disallow_inf_nan` | `disallow_inf_nan` | Reject infinity and NaN | float32, float64 |

## Comparison with Sign Constraints

| Constraint | Equivalent | Allows Zero | Example Use Case |
|------------|-----------|------------|------------------|
| `positive` | `gt=0` | No | Account balance must be positive |
| `gt=0` | `positive` | No | Same as positive |
| `gte=0` | N/A | Yes | Quantity can be zero or positive |
| `negative` | `lt=0` | No | Temperature below freezing |
| `lt=0` | `negative` | No | Same as negative |
| `lte=0` | N/A | Yes | Temperature can be zero or negative |

## Real-World Validation Examples

**E-commerce:**
```go
Price float64 `json:"price" pedantigo:"required,min=0.01,decimal_places=2"`
Quantity int `json:"quantity" pedantigo:"required,min=1,max=1000"`
DiscountPercent int `json:"discount" pedantigo:"min=0,max=100"`
```

**Banking:**
```go
Balance float64 `json:"balance" pedantigo:"required,decimal_places=2,max_digits=15"`
WithdrawalAmount float64 `json:"amount" pedantigo:"required,positive,decimal_places=2"`
InterestRate float64 `json:"rate" pedantigo:"required,min=0,decimal_places=4"`
```

**Scientific/Measurements:**
```go
Temperature float64 `json:"temp_celsius" pedantigo:"disallow_inf_nan"`
PressurePa float64 `json:"pressure" pedantigo:"required,positive"`
PH float64 `json:"ph" pedantigo:"required,min=0,max=14,decimal_places=2"`
```

**Rating Systems:**
```go
Stars float64 `json:"stars" pedantigo:"min=0,max=5,multiple_of=0.5"`
Percentage int `json:"percentage" pedantigo:"required,min=0,max=100"`
```
