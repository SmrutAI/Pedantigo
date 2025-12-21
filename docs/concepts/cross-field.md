---
sidebar_position: 3
---

# Cross-Field Validation

Validate relationships between multiple fields in a struct. Cross-field constraints allow you to express complex validation rules that depend on the values of multiple fields, such as password confirmation, date ranges, conditional requirements, and more.

## Field Comparison

Compare the values of two fields using comparison operators.

### Equality Constraints

Use `eqfield` to require a field to equal another field, and `nefield` to require inequality.

**Password Confirmation Example:**

```go
type RegisterRequest struct {
    Email            string `json:"email" pedantigo:"required,email"`
    Password         string `json:"password" pedantigo:"required,minLength=8"`
    PasswordConfirm  string `json:"password_confirm" pedantigo:"required,eqfield=Password"`
}

// Valid - passwords match
data := []byte(`{
    "email": "user@example.com",
    "password": "SecurePass123",
    "password_confirm": "SecurePass123"
}`)
user, err := pedantigo.Unmarshal[RegisterRequest](data)
if err != nil {
    fmt.Println(err) // No error
}

// Invalid - passwords don't match
badData := []byte(`{
    "email": "user@example.com",
    "password": "SecurePass123",
    "password_confirm": "DifferentPass456"
}`)
_, err = pedantigo.Unmarshal[RegisterRequest](badData)
if err != nil {
    // ValidationError: password_confirm must equal Password
}
```

**Not Equal Constraint:**

```go
type UpdateUsername struct {
    CurrentUsername string `json:"current_username" pedantigo:"required"`
    NewUsername     string `json:"new_username" pedantigo:"required,nefield=CurrentUsername"`
}

// Valid - new username differs from current
data := []byte(`{
    "current_username": "oldname",
    "new_username": "newname"
}`)
update, _ := pedantigo.Unmarshal[UpdateUsername](data)

// Invalid - new username is the same
badData := []byte(`{
    "current_username": "samename",
    "new_username": "samename"
}`)
_, err := pedantigo.Unmarshal[UpdateUsername](badData)
// ValidationError: new_username must not equal CurrentUsername
```

### Comparison Operators

Use `gtfield`, `gtefield`, `ltfield`, and `ltefield` to compare numeric or string values.

**Date Range Example:**

```go
type EventBooking struct {
    EventName  string    `json:"event_name" pedantigo:"required"`
    StartDate  time.Time `json:"start_date" pedantigo:"required"`
    EndDate    time.Time `json:"end_date" pedantigo:"required,gtfield=StartDate"`
}

// Valid - end date is after start date
data := []byte(`{
    "event_name": "Conference",
    "start_date": "2025-06-01T09:00:00Z",
    "end_date": "2025-06-03T17:00:00Z"
}`)
booking, _ := pedantigo.Unmarshal[EventBooking](data)

// Invalid - end date is before start date
badData := []byte(`{
    "event_name": "Conference",
    "start_date": "2025-06-03T17:00:00Z",
    "end_date": "2025-06-01T09:00:00Z"
}`)
_, err := pedantigo.Unmarshal[EventBooking](badData)
// ValidationError: end_date must be greater than StartDate
```

**Price Range Example:**

```go
type ProductListing struct {
    Name       string  `json:"name" pedantigo:"required"`
    MinPrice   float64 `json:"min_price" pedantigo:"required,gt=0"`
    MaxPrice   float64 `json:"max_price" pedantigo:"required,gtfield=MinPrice"`
    DiscountAt float64 `json:"discount_at" pedantigo:"gtefield=MinPrice,ltefield=MaxPrice"`
}

// Valid - prices are properly ordered
data := []byte(`{
    "name": "Widget",
    "min_price": 10.0,
    "max_price": 100.0,
    "discount_at": 50.0
}`)
product, _ := pedantigo.Unmarshal[ProductListing](data)

// Invalid - discount is outside the price range
badData := []byte(`{
    "name": "Widget",
    "min_price": 10.0,
    "max_price": 100.0,
    "discount_at": 150.0
}`)
_, err := pedantigo.Unmarshal[ProductListing](badData)
// ValidationError: discount_at must be less than or equal to MaxPrice
```

## Conditional Required

Make fields conditionally required based on the value of another field.

### required_if and required_unless

Use `required_if` to require a field when another field has a specific value. Use `required_unless` for the opposite condition.

**String Condition:**

```go
type ShippingForm struct {
    Country      string `json:"country" pedantigo:"required,oneof=US CA MX"`
    State        string `json:"state" pedantigo:"required_if=Country US"`
    Province     string `json:"province" pedantigo:"required_if=Country CA"`
    PostalCode   string `json:"postal_code" pedantigo:"required"`
}

// Valid - Country is US and State is provided
data := []byte(`{
    "country": "US",
    "state": "California",
    "postal_code": "90210"
}`)
form, _ := pedantigo.Unmarshal[ShippingForm](data)

// Valid - Country is CA, State is not required
data = []byte(`{
    "country": "CA",
    "province": "Ontario",
    "postal_code": "M5H 2N2"
}`)
form, _ = pedantigo.Unmarshal[ShippingForm](data)

// Invalid - Country is US but State is missing
badData := []byte(`{
    "country": "US",
    "postal_code": "90210"
}`)
_, err := pedantigo.Unmarshal[ShippingForm](badData)
// ValidationError: state is required when country equals US
```

**Boolean Condition:**

```go
type SubscriptionForm struct {
    HasBusiness      bool   `json:"has_business"`
    BusinessName     string `json:"business_name" pedantigo:"required_if=HasBusiness true"`
    BusinessLicense  string `json:"business_license" pedantigo:"required_if=HasBusiness true"`
    PersonalName     string `json:"personal_name" pedantigo:"required_unless=HasBusiness true"`
}

// Valid - business fields provided
data := []byte(`{
    "has_business": true,
    "business_name": "Acme Corp",
    "business_license": "ACME-2025-001"
}`)
form, _ := pedantigo.Unmarshal[SubscriptionForm](data)

// Valid - personal field provided when has_business is false
data = []byte(`{
    "has_business": false,
    "personal_name": "John Doe"
}`)
form, _ = pedantigo.Unmarshal[SubscriptionForm](data)

// Invalid - has_business is false but personal_name is missing
badData := []byte(`{
    "has_business": false
}`)
_, err := pedantigo.Unmarshal[SubscriptionForm](badData)
// ValidationError: personal_name is required unless has_business equals true
```

### required_with and required_without

Use `required_with` to require a field only if another field is present (non-zero). Use `required_without` for the opposite.

**Optional Dependencies:**

```go
type PaymentInfo struct {
    PaymentMethod string `json:"payment_method" pedantigo:"required,oneof=credit_card bank_transfer"`
    CardNumber    string `json:"card_number" pedantigo:"required_if=PaymentMethod credit_card"`
    CVV           string `json:"cvv" pedantigo:"required_with=CardNumber"`
    BankAccount   string `json:"bank_account" pedantigo:"required_if=PaymentMethod bank_transfer"`
    RoutingNumber string `json:"routing_number" pedantigo:"required_with=BankAccount"`
}

// Valid - credit card with CVV
data := []byte(`{
    "payment_method": "credit_card",
    "card_number": "4532-1234-5678-9010",
    "cvv": "123"
}`)
payment, _ := pedantigo.Unmarshal[PaymentInfo](data)

// Valid - bank transfer with routing number
data = []byte(`{
    "payment_method": "bank_transfer",
    "bank_account": "123456789",
    "routing_number": "021000021"
}`)
payment, _ = pedantigo.Unmarshal[PaymentInfo](data)

// Invalid - has card number but missing CVV
badData := []byte(`{
    "payment_method": "credit_card",
    "card_number": "4532-1234-5678-9010"
}`)
_, err := pedantigo.Unmarshal[PaymentInfo](badData)
// ValidationError: cvv is required when CardNumber is present
```

**Requires Absence:**

```go
type TwoFactorSettings struct {
    TwoFactorEnabled bool   `json:"two_factor_enabled"`
    BackupCode      string `json:"backup_code" pedantigo:"required_without=TwoFactorEnabled"`
}

// Valid - backup code provided when 2FA is disabled
data := []byte(`{
    "two_factor_enabled": false,
    "backup_code": "BACKUP-ABC-123"
}`)
settings, _ := pedantigo.Unmarshal[TwoFactorSettings](data)

// Valid - no backup code needed when 2FA is enabled
data = []byte(`{
    "two_factor_enabled": true
}`)
settings, _ = pedantigo.Unmarshal[TwoFactorSettings](data)

// Invalid - 2FA disabled but no backup code
badData := []byte(`{
    "two_factor_enabled": false
}`)
_, err := pedantigo.Unmarshal[TwoFactorSettings](badData)
// ValidationError: backup_code is required when TwoFactorEnabled is absent
```

## Conditional Excluded

Make fields conditionally forbidden based on the value of another field.

### excluded_if and excluded_unless

Use `excluded_if` to forbid a field when another field has a specific value. The field must be empty (zero value or empty string).

**Mutually Exclusive Fields:**

```go
type DiscountCode struct {
    AccountID   string `json:"account_id"`
    DiscountPercent int    `json:"discount_percent" pedantigo:"excluded_if=AccountID premium"`
    Notes       string `json:"notes" pedantigo:"excluded_unless=AccountID enterprise"`
}

// Valid - premium account (no discount percent allowed)
data := []byte(`{
    "account_id": "premium"
}`)
discount, _ := pedantigo.Unmarshal[DiscountCode](data)

// Valid - enterprise account (notes required)
data = []byte(`{
    "account_id": "enterprise",
    "notes": "VIP customer"
}`)
discount, _ = pedantigo.Unmarshal[DiscountCode](data)

// Invalid - premium account with discount percent
badData := []byte(`{
    "account_id": "premium",
    "discount_percent": 10
}`)
_, err := pedantigo.Unmarshal[DiscountCode](badData)
// ValidationError: discount_percent must be empty when account_id equals premium
```

### excluded_with and excluded_without

Use `excluded_with` to forbid a field when another field is present. Use `excluded_without` for the opposite.

**API Key vs Token:**

```go
type AuthRequest struct {
    APIKey string `json:"api_key" pedantigo:"minLength=20,maxLength=50"`
    Token  string `json:"token" pedantigo:"excluded_with=APIKey,minLength=20,maxLength=100"`
}

// Valid - uses API key only
data := []byte(`{
    "api_key": "sk-1234567890abcdefghij"
}`)
auth, _ := pedantigo.Unmarshal[AuthRequest](data)

// Valid - uses token only
data = []byte(`{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}`)
auth, _ = pedantigo.Unmarshal[AuthRequest](data)

// Invalid - both API key and token provided
badData := []byte(`{
    "api_key": "sk-1234567890abcdefghij",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}`)
_, err := pedantigo.Unmarshal[AuthRequest](badData)
// ValidationError: token must be empty when APIKey is present
```

## Custom Cross-Field Validation

For complex business rules that go beyond simple field comparisons, implement the `Validatable` interface on your struct.

### The Validatable Interface

Any struct can implement the `Validatable` interface to add custom cross-field validation logic:

```go
type Validatable interface {
    Validate() error
}
```

When a struct implements `Validate()`, Pedantigo automatically calls it after all tag-based validations pass. This allows you to express arbitrary business logic as validation rules.

:::warning Don't call pedantigo inside Validate()
Since Pedantigo calls your `Validate()` method automatically, calling `pedantigo.Validate()` or `pedantigo.Unmarshal()` on `self` inside `Validate()` causes infinite recursion. Your `Validate()` method should only contain custom business logic.
:::

**Complex Business Rules Example:**

```go
type FlightBooking struct {
    Origin      string    `json:"origin" pedantigo:"required"`
    Destination string    `json:"destination" pedantigo:"required,nefield=Origin"`
    DepartDate  time.Time `json:"depart_date" pedantigo:"required"`
    ReturnDate  time.Time `json:"return_date"`
    IsRoundTrip bool      `json:"is_round_trip"`
}

// Validate implements the Validatable interface for complex business rules
func (b FlightBooking) Validate() error {
    // Round trips require a return date
    if b.IsRoundTrip && b.ReturnDate.IsZero() {
        return fmt.Errorf("return_date is required for round trip bookings")
    }

    // Return date must be after departure date
    if !b.ReturnDate.IsZero() && !b.ReturnDate.After(b.DepartDate) {
        return fmt.Errorf("return_date must be after depart_date")
    }

    // Departure date must be in the future
    if b.DepartDate.Before(time.Now()) {
        return fmt.Errorf("depart_date must be in the future")
    }

    return nil
}

// Usage
data := []byte(`{
    "origin": "JFK",
    "destination": "LAX",
    "depart_date": "2025-12-25T10:00:00Z",
    "return_date": "2025-12-31T18:00:00Z",
    "is_round_trip": true
}`)
booking, err := pedantigo.Unmarshal[FlightBooking](data)
// Custom Validate() is called automatically after field validation
```

**Multiple Validation Errors:**

If your `Validate()` method returns an error, it will be collected in the `ValidationError` along with other field validation errors:

```go
type Account struct {
    Username string `json:"username" pedantigo:"required,minLength=3,maxLength=20"`
    Email    string `json:"email" pedantigo:"required,email"`
    Age      int    `json:"age" pedantigo:"min=18"`
}

func (a Account) Validate() error {
    // This error is collected alongside field validation errors
    if strings.Contains(a.Username, a.Email) {
        return fmt.Errorf("username cannot contain email address")
    }
    return nil
}

// Multiple errors are all returned together
data := []byte(`{
    "username": "ab",
    "email": "not-an-email",
    "age": 15
}`)
_, err := pedantigo.Unmarshal[Account](data)
// ValidationError with 4 errors:
// - username: must be at least 3 characters
// - email: must be a valid email
// - age: must be at least 18
// - Validate: username cannot contain email address
```

## Complete Example

Here's a comprehensive example combining all cross-field validation techniques:

```go
package main

import (
    "fmt"
    "time"

    pedantigo "github.com/SmrutAI/pedantigo"
)

type EventRegistration struct {
    // Basic fields
    Email           string    `json:"email" pedantigo:"required,email"`
    FullName        string    `json:"full_name" pedantigo:"required,minLength=2"`

    // Field comparisons
    Password        string    `json:"password" pedantigo:"required,minLength=12"`
    PasswordConfirm string    `json:"password_confirm" pedantigo:"required,eqfield=Password"`

    // Conditional requirements
    IsStudent       bool      `json:"is_student"`
    StudentID       string    `json:"student_id" pedantigo:"required_if=IsStudent true,len=10"`
    Company         string    `json:"company" pedantigo:"required_unless=IsStudent true"`

    // Date ranges
    StartDate       time.Time `json:"start_date" pedantigo:"required"`
    EndDate         time.Time `json:"end_date" pedantigo:"required,gtfield=StartDate"`

    // Mutually exclusive
    PhoneNumber     string    `json:"phone_number" pedantigo:"excluded_with=TelegramHandle,len=10"`
    TelegramHandle  string    `json:"telegram_handle" pedantigo:"excluded_with=PhoneNumber"`

    // Optional dependencies
    HasAccommodation bool      `json:"has_accommodation"`
    HotelName        string    `json:"hotel_name" pedantigo:"required_with=HasAccommodation"`
    CheckinDate      time.Time `json:"checkin_date" pedantigo:"required_with=HasAccommodation"`
}

// Validate implements custom cross-field validation logic
func (e EventRegistration) Validate() error {
    // Check-in date must be before event start
    if !e.CheckinDate.IsZero() && !e.CheckinDate.Before(e.StartDate) {
        return fmt.Errorf("check-in date must be before event start date")
    }

    // Event must be at least 1 day long
    if e.EndDate.Sub(e.StartDate) < 24*time.Hour {
        return fmt.Errorf("event must be at least 1 day long")
    }

    return nil
}

func main() {
    // Valid registration with all required fields
    validData := []byte(`{
        "email": "student@university.edu",
        "full_name": "Alice Johnson",
        "password": "SecurePass123456",
        "password_confirm": "SecurePass123456",
        "is_student": true,
        "student_id": "STUD123456",
        "start_date": "2025-06-15T09:00:00Z",
        "end_date": "2025-06-17T17:00:00Z",
        "has_accommodation": true,
        "hotel_name": "University Hotel",
        "checkin_date": "2025-06-14T15:00:00Z"
    }`)

    registration, err := pedantigo.Unmarshal[EventRegistration](validData)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("Registration successful for %s\n", registration.FullName)

    // Invalid registration with multiple errors
    invalidData := []byte(`{
        "email": "not-an-email",
        "full_name": "B",
        "password": "short",
        "password_confirm": "different",
        "is_student": false,
        "start_date": "2025-06-15T09:00:00Z",
        "end_date": "2025-06-14T17:00:00Z",
        "phone_number": "555-1234",
        "telegram_handle": "@alice"
    }`)

    _, err = pedantigo.Unmarshal[EventRegistration](invalidData)
    if err != nil {
        var ve *pedantigo.ValidationError
        if errors.As(err, &ve) {
            for _, fieldErr := range ve.Errors {
                fmt.Printf("- %s: %s\n", fieldErr.Field, fieldErr.Message)
            }
        }
    }
}
```

## Summary

Cross-field constraints provide powerful validation capabilities for complex data validation scenarios:

| Constraint | Use Case |
|-----------|----------|
| `eqfield=Field` | Password confirmation, matching values |
| `nefield=Field` | Ensure inequality, prevent duplicates |
| `gtfield=Field` | Numeric/date ranges, ordering |
| `gtefield=Field` | Greater than or equal comparisons |
| `ltfield=Field` | Upper bounds, before dates |
| `ltefield=Field` | Less than or equal comparisons |
| `required_if=Field Value` | Conditional requirements based on values |
| `required_unless=Field Value` | Required unless condition is true |
| `required_with=Field` | Required when another field is present |
| `required_without=Field` | Required when another field is absent |
| `excluded_if=Field Value` | Forbidden if condition is true |
| `excluded_unless=Field Value` | Forbidden unless condition is true |
| `excluded_with=Field` | Forbidden if another field is present |
| `excluded_without=Field` | Forbidden if another field is absent |
| `Validatable` interface | Custom business logic validation |

Combine struct tag constraints with the `Validatable` interface to handle even the most complex validation scenarios in a clear, declarative way.
