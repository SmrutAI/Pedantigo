---
sidebar_position: 2
---

# Custom Validators

Create your own validation constraints for domain-specific business rules. Pedantigo provides two complementary approaches: field-level validators for reusable validation logic, and struct-level validators for complex cross-struct validation.

## Field-Level Validators

Field-level validators allow you to define custom validation constraints that can be applied to any field via struct tags.

### Defining a Field Validator

A field validator is a function that receives a value and optional parameter, and returns an error if validation fails:

```go
type ValidationFunc func(value any, param string) error
```

**Simple Example:**

```go
// Phone number validator for US numbers
func ValidateUSPhone(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    if len(str) != 10 {
        return errors.New("must be exactly 10 digits")
    }

    for _, ch := range str {
        if ch < '0' || ch > '9' {
            return errors.New("must contain only digits")
        }
    }

    return nil
}
```

**Validator with Parameters:**

The `param` argument contains any value specified after `=` in the tag. This allows you to make validators configurable:

```go
// Custom range validator for strings based on length
func ValidateStringRange(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    parts := strings.Split(param, "-")
    if len(parts) != 2 {
        return errors.New("invalid parameter format, use min-max")
    }

    minLen, _ := strconv.Atoi(parts[0])
    maxLen, _ := strconv.Atoi(parts[1])

    if len(str) < minLen || len(str) > maxLen {
        return fmt.Errorf("must be between %d and %d characters", minLen, maxLen)
    }

    return nil
}
```

### Registering with RegisterValidation

Register your custom validator to make it available in struct tags:

```go
func RegisterValidation(name string, fn ValidationFunc) error
```

**Important constraints:**
- Cannot override built-in validators (required, email, min, max, etc.)
- Returns an error if the name is reserved
- Clears the validator cache after registration
- Thread-safe for concurrent use

**Registration Example:**

```go
package main

import (
    "errors"
    "strings"
    "github.com/smrutai/pedantigo"
)

func init() {
    // Register US phone validator
    pedantigo.RegisterValidation("us_phone", ValidateUSPhone)

    // Register credit card validator
    pedantigo.RegisterValidation("credit_card", ValidateCreditCard)
}

// Validates US phone numbers: 10 digits
func ValidateUSPhone(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    digits := strings.Map(func(r rune) rune {
        if r >= '0' && r <= '9' {
            return r
        }
        return -1
    }, str)

    if len(digits) != 10 {
        return errors.New("must contain exactly 10 digits")
    }

    return nil
}

// Validates credit card numbers using Luhn algorithm
func ValidateCreditCard(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    digits := strings.Map(func(r rune) rune {
        if r >= '0' && r <= '9' {
            return r
        }
        return -1
    }, str)

    if len(digits) < 13 || len(digits) > 19 {
        return errors.New("must be between 13 and 19 digits")
    }

    // Luhn algorithm implementation...
    return nil
}
```

### Using Custom Validators in Struct Tags

Once registered, use custom validators in the `pedantigo` struct tag:

```go
package main

import "github.com/smrutai/pedantigo"

type User struct {
    Email      string `json:"email" pedantigo:"required,email"`
    Phone      string `json:"phone" pedantigo:"us_phone"`
}

type Payment struct {
    CardNumber string `json:"card_number" pedantigo:"required,credit_card"`
}

func main() {
    // Use the validator like any built-in constraint
    data := []byte(`{
        "email": "user@example.com",
        "phone": "5551234567"
    }`)

    user, err := pedantigo.Unmarshal[User](data)
    if err != nil {
        // Handle validation errors
    }
}
```

### Custom Validators with Parameters

Use the `param` argument to pass configuration to your validator:

```go
// Generic length validator
func ValidateLength(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    expectedLen, err := strconv.Atoi(param)
    if err != nil {
        return errors.New("invalid parameter: expected number")
    }

    if len(str) != expectedLen {
        return fmt.Errorf("must be exactly %d characters", expectedLen)
    }

    return nil
}

// Register and use with parameter
pedantigo.RegisterValidation("length", ValidateLength)

type Product struct {
    SKU string `json:"sku" pedantigo:"length=12"` // Exactly 12 chars
    UPC string `json:"upc" pedantigo:"length=13"` // Exactly 13 chars
}
```

## Struct-Level Validators

For validation that involves the entire struct or multiple fields, use two approaches: the `Validatable` interface (recommended for simplicity) or `RegisterStructValidation` (for global registration).

### The Validatable Interface (Simpler Approach)

Implement the `Validatable` interface on your struct for automatic cross-field validation:

```go
type Validatable interface {
    Validate() error
}
```

**Example:**

```go
type PasswordChange struct {
    CurrentPassword string `json:"current_password" pedantigo:"required,minLength=8"`
    NewPassword     string `json:"new_password" pedantigo:"required,minLength=8"`
    Confirm         string `json:"confirm" pedantigo:"required,eqfield=NewPassword"`
}

// Implement Validatable for custom business logic
func (p PasswordChange) Validate() error {
    if p.CurrentPassword == p.NewPassword {
        return errors.New("new_password must differ from current_password")
    }
    return nil
}

// Usage
data := []byte(`{
    "current_password": "OldPass123",
    "new_password": "NewPass123",
    "confirm": "NewPass123"
}`)

change, err := pedantigo.Unmarshal[PasswordChange](data)
// Validate() is called automatically after field validation passes
```

The `Validatable` approach is recommended for most use cases because:
- Validation logic stays with your struct definition
- Self-contained and easy to read
- No global registration needed
- Works seamlessly with struct tag constraints

### RegisterStructValidation (Global Registration)

For scenarios where you want global validation registration or validation of types you don't own:

```go
func RegisterStructValidation[T any](fn StructLevelFunc[T]) error
```

Where `StructLevelFunc` has the signature:
```go
type StructLevelFunc[T any] func(obj *T) error
```

**Example:**

```go
type Order struct {
    Items      []OrderItem `json:"items" pedantigo:"required"`
    TotalPrice float64     `json:"total_price" pedantigo:"required,gt=0"`
    DiscountPercent int    `json:"discount_percent" pedantigo:"min=0,max=100"`
}

type OrderItem struct {
    Product string  `json:"product" pedantigo:"required"`
    Qty     int     `json:"qty" pedantigo:"required,min=1"`
    Price   float64 `json:"price" pedantigo:"required,gt=0"`
}

// Register struct-level validation
func init() {
    pedantigo.RegisterStructValidation[Order](ValidateOrder)
}

// Validates cross-struct relationships and business rules
func ValidateOrder(order *Order) error {
    // Calculate expected total
    var expectedTotal float64
    for _, item := range order.Items {
        expectedTotal += item.Price * float64(item.Qty)
    }

    // Apply discount
    expectedTotal = expectedTotal * (1 - float64(order.DiscountPercent) / 100)

    // Verify total price matches
    if math.Abs(expectedTotal - order.TotalPrice) > 0.01 {
        return fmt.Errorf("total_price does not match items total (expected %.2f, got %.2f)",
            expectedTotal, order.TotalPrice)
    }

    return nil
}

func main() {
    data := []byte(`{
        "items": [
            {"product": "Widget", "qty": 2, "price": 19.99},
            {"product": "Gadget", "qty": 1, "price": 49.99}
        ],
        "total_price": 89.97,
        "discount_percent": 0
    }`)

    order, err := pedantigo.Unmarshal[Order](data)
    // ValidateOrder is called automatically after field validation
}
```

**Key differences from Validatable:**
- Called after field-level validation for the struct
- Global registration means it applies to all instances
- Useful for validating types from external packages
- Receives a pointer to the struct

## Best Practices

### Return Clear Error Messages

Your validators should return descriptive error messages that help users understand what went wrong:

```go
// Bad
func ValidateEmail(value any, param string) error {
    if !isValidEmail(value.(string)) {
        return errors.New("invalid") // Too vague
    }
    return nil
}

// Good
func ValidateEmail(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    if !isValidEmail(str) {
        return errors.New("must be a valid email address (format: user@domain.com)")
    }

    return nil
}
```

### Handle Type Assertions Safely

Always check that the value is the expected type:

```go
func ValidatePositive(value any, param string) error {
    // Check type first
    num, ok := value.(float64)
    if !ok {
        // Try int
        if intVal, ok := value.(int); ok {
            num = float64(intVal)
        } else {
            return errors.New("must be a number")
        }
    }

    if num <= 0 {
        return errors.New("must be greater than zero")
    }

    return nil
}
```

### Make Validators Composable

Design validators to be independent and composable with other constraints:

```go
// These work together
type Username struct {
    Name string `json:"name" pedantigo:"required,alphanumeric,minLength=3,maxLength=20,username_available"`
}

// username_available checks if the username is not in the reserved list
func ValidateUsernameAvailable(value any, param string) error {
    username, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    reserved := map[string]bool{
        "admin": true,
        "root": true,
        "system": true,
    }

    if reserved[strings.ToLower(username)] {
        return errors.New("is a reserved username")
    }

    return nil
}
```

### Validate Input Format in Validators

Don't assume input is in the correct format; validate defensively:

```go
func ValidateRange(value any, param string) error {
    // Validate parameter format
    parts := strings.Split(param, "-")
    if len(parts) != 2 {
        return fmt.Errorf("invalid parameter format: expected 'min-max', got '%s'", param)
    }

    min, err := strconv.Atoi(parts[0])
    if err != nil {
        return fmt.Errorf("invalid minimum value: %w", err)
    }

    max, err := strconv.Atoi(parts[1])
    if err != nil {
        return fmt.Errorf("invalid maximum value: %w", err)
    }

    // Validate actual value
    num, ok := value.(int)
    if !ok {
        return errors.New("must be an integer")
    }

    if num < min || num > max {
        return fmt.Errorf("must be between %d and %d", min, max)
    }

    return nil
}
```

### Document Your Validators

Provide clear documentation for custom validators:

```go
// ValidateSlug validates URL-friendly slugs.
// Format: lowercase letters, numbers, and hyphens, 3-50 characters.
// Example: "my-awesome-blog-post"
func ValidateSlug(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    if len(str) < 3 || len(str) > 50 {
        return errors.New("must be between 3 and 50 characters")
    }

    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(str) {
        return errors.New("must contain only lowercase letters, numbers, and hyphens")
    }

    return nil
}
```

## Complete Example

Here's a comprehensive example combining field-level and struct-level validators:

```go
package main

import (
    "errors"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "github.com/smrutai/pedantigo"
)

// Custom field validators
func ValidateUSPhone(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    digits := strings.Map(func(r rune) rune {
        if r >= '0' && r <= '9' {
            return r
        }
        return -1
    }, str)

    if len(digits) != 10 {
        return errors.New("must contain exactly 10 digits")
    }

    return nil
}

func ValidateHexColor(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    if !regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`).MatchString(str) {
        return errors.New("must be valid hex color format (#RRGGBB)")
    }

    return nil
}

func ValidateSlug(value any, param string) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("must be a string")
    }

    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(str) {
        return errors.New("must contain only lowercase letters, numbers, and hyphens")
    }

    return nil
}

type ThemeConfig struct {
    Name          string `json:"name" pedantigo:"required,minLength=3,maxLength=50"`
    Slug          string `json:"slug" pedantigo:"required,slug"`
    PrimaryColor  string `json:"primary_color" pedantigo:"required,hex_color"`
    SecondaryColor string `json:"secondary_color" pedantigo:"hex_color"`
    AccentColor   string `json:"accent_color" pedantigo:"hex_color"`
}

// Validate custom cross-field rules
func (t ThemeConfig) Validate() error {
    if t.PrimaryColor == t.SecondaryColor {
        return errors.New("primary and secondary colors must be different")
    }
    return nil
}

type UserProfile struct {
    Email      string `json:"email" pedantigo:"required,email"`
    Phone      string `json:"phone" pedantigo:"us_phone"`
    Theme      ThemeConfig `json:"theme" pedantigo:"required"`
    Bio        string `json:"bio" pedantigo:"maxLength=500"`
}

func init() {
    // Register custom field validators
    pedantigo.RegisterValidation("us_phone", ValidateUSPhone)
    pedantigo.RegisterValidation("hex_color", ValidateHexColor)
    pedantigo.RegisterValidation("slug", ValidateSlug)
}

func main() {
    // Valid profile
    validData := []byte(`{
        "email": "user@example.com",
        "phone": "5551234567",
        "theme": {
            "name": "Ocean Blue",
            "slug": "ocean-blue",
            "primary_color": "#0066CC",
            "secondary_color": "#00CC66"
        },
        "bio": "Software engineer and open source enthusiast"
    }`)

    profile, err := pedantigo.Unmarshal[UserProfile](validData)
    if err == nil {
        fmt.Printf("Profile created: %s\n", profile.Email)
    }

    // Invalid profile - multiple validation errors
    invalidData := []byte(`{
        "email": "not-an-email",
        "phone": "555-123",
        "theme": {
            "name": "X",
            "slug": "Ocean Blue",
            "primary_color": "#0066CC",
            "secondary_color": "#0066CC"
        }
    }`)

    _, err = pedantigo.Unmarshal[UserProfile](invalidData)
    if err != nil {
        var ve *pedantigo.ValidationError
        if errors.As(err, &ve) {
            fmt.Println("Validation errors:")
            for _, fieldErr := range ve.Errors {
                fmt.Printf("  %s: %s\n", fieldErr.Field, fieldErr.Message)
            }
        }
    }
}
```

## Summary

Custom validators provide two powerful patterns:

| Feature | Use Case |
|---------|----------|
| **Field-Level Validators** | Reusable constraints applied via struct tags |
| **RegisterValidation** | Global registration for field validators |
| **Validatable Interface** | Custom struct validation with access to all fields |
| **RegisterStructValidation** | Global registration for struct-level validation |

Combine these approaches to handle validation requirements from simple format checks to complex business logic validation across multiple fields and nested structures.
