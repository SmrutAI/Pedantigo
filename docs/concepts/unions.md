---
sidebar_position: 4
---

# Discriminated Unions

Type-safe handling of JSON with multiple possible shapes using a discriminator field.

:::tip
For JSON Schema generation of unions, see [Schema Generation](/docs/concepts/schema). For streaming union responses from LLMs, see [Streaming Validation](/docs/concepts/streaming) and [LLM Examples](/docs/examples/llm-streaming).
:::

## What Are Discriminated Unions?

Discriminated unions allow you to validate JSON payloads that can be one of several different types, where a specific field (the discriminator) determines which variant is present. This is common in APIs that send polymorphic data.

For example, a payment system might send different data structures depending on the payment method:

```json
{ "type": "credit_card", "cardNumber": "4111111111111111", "cvc": "123" }
{ "type": "bank_transfer", "accountNumber": "12345", "routingNumber": "987654" }
{ "type": "digital_wallet", "walletId": "wallet_123", "provider": "apple_pay" }
```

A discriminated union automatically routes each JSON payload to the correct variant type and validates it accordingly.

## Defining Variant Types

Each variant in your union should be a separate struct with appropriate validation constraints:

```go
type CreditCard struct {
    Type       string `json:"type" validate:"required"`
    CardNumber string `json:"cardNumber" validate:"required,pattern=^[0-9]{16}$"`
    CVC        string `json:"cvc" validate:"required,pattern=^[0-9]{3}$"`
    ExpiryDate string `json:"expiryDate" validate:"required,pattern=^[0-9]{2}/[0-9]{2}$"`
}

type BankTransfer struct {
    Type           string `json:"type" validate:"required"`
    AccountNumber  string `json:"accountNumber" validate:"required,pattern=^[0-9]{10,12}$"`
    RoutingNumber  string `json:"routingNumber" validate:"required,pattern=^[0-9]{9}$"`
    AccountHolderName string `json:"accountHolderName" validate:"required,min=2"`
}

type DigitalWallet struct {
    Type     string `json:"type" validate:"required"`
    WalletID string `json:"walletId" validate:"required,min=1"`
    Provider string `json:"provider" validate:"required,enum=apple_pay|google_pay|paypal"`
}
```

Each variant struct:
- **Must have a discriminator field** (typically a `Type` field) that contains the variant identifier
- **Must have constraints** - validation rules are applied per variant
- **Can have its own validation logic** - implement `Validatable` interface for cross-field checks

## Creating a UnionValidator

Unlike the Simple API, discriminated unions require explicit creation with `validator.NewUnion()`:

```go
unionValidator, err := validator.NewUnion[any](validator.UnionOptions{
    DiscriminatorField: "type",
    Variants: []validator.UnionVariant{
        validator.VariantFor[CreditCard]("credit_card"),
        validator.VariantFor[BankTransfer]("bank_transfer"),
        validator.VariantFor[DigitalWallet]("digital_wallet"),
    },
})

if err != nil {
    // Handle creation errors (invalid discriminator, duplicate variants, etc.)
    log.Fatal(err)
}
```

### UnionOptions

The `UnionOptions` struct configures union behavior:

```go
type UnionOptions struct {
    // DiscriminatorField is the JSON field name that determines which variant to use
    DiscriminatorField string

    // Variants is a slice of UnionVariant defining all possible types
    Variants []UnionVariant
}
```

### UnionVariant

Each variant is created with `VariantFor[T]()`:

```go
validator.VariantFor[CreditCard]("credit_card")
```

The generic type parameter is the Go struct, and the string argument is the discriminator value to match in the JSON.

## Unmarshaling Union Data

When you unmarshal JSON with a union validator, it automatically:
1. Inspects the discriminator field value
2. Selects the matching variant
3. Validates the JSON against that variant's constraints
4. Returns the validated variant as an `any` type

```go
jsonData := []byte(`{
    "type": "credit_card",
    "cardNumber": "4111111111111111",
    "cvc": "123",
    "expiryDate": "12/25"
}`)

result, err := unionValidator.Unmarshal(jsonData)
if err != nil {
    // Handle validation errors
    var ve *validator.ValidationError
    if errors.As(err, &ve) {
        for _, fieldErr := range ve.Errors {
            fmt.Printf("Field %s: %s\n", fieldErr.Field, fieldErr.Message)
        }
    }
    return
}

// result is interface{}, need type assertion
payment := result.(CreditCard)
fmt.Printf("Processing card: %s\n", payment.CardNumber)
```

### Type Assertion Pattern

After unmarshaling, use type assertion to access the specific variant:

```go
switch payment := result.(type) {
case CreditCard:
    fmt.Printf("Credit card ending in: %s\n", payment.CardNumber[len(payment.CardNumber)-4:])
case BankTransfer:
    fmt.Printf("Bank transfer to account: %s\n", payment.AccountNumber)
case DigitalWallet:
    fmt.Printf("Digital wallet: %s\n", payment.Provider)
}
```

### Error Handling

Discriminated union errors include field path information:

```go
jsonData := []byte(`{
    "type": "credit_card",
    "cardNumber": "invalid",
    "cvc": "12"
}`)

_, err := unionValidator.Unmarshal(jsonData)
if err != nil {
    var ve *validator.ValidationError
    if errors.As(err, &ve) {
        for _, fieldErr := range ve.Errors {
            // Example: Field "cardNumber" error
            fmt.Printf("Field %s: %s\n", fieldErr.Field, fieldErr.Message)
        }
    }
}
```

Common errors:
- **Missing discriminator field** - Union can't determine which variant to use
- **Unknown discriminator value** - No variant matches the provided value
- **Validation errors** - The selected variant fails its constraints

## Validating Existing Variant Values

You can also validate already-instantiated variant structs:

```go
card := CreditCard{
    Type:       "credit_card",
    CardNumber: "4111111111111111",
    CVC:        "123",
    ExpiryDate: "12/25",
}

// Validate the existing value
err := unionValidator.Validate(card)
if err != nil {
    // Handle validation errors
}
```

## JSON Schema for Unions

Discriminated unions generate OpenAPI-compatible JSON Schema using `oneOf` with a discriminator:

```go
schema := unionValidator.Schema()

// This produces a JSON Schema like:
// {
//   "oneOf": [
//     { "$ref": "#/definitions/CreditCard" },
//     { "$ref": "#/definitions/BankTransfer" },
//     { "$ref": "#/definitions/DigitalWallet" }
//   ],
//   "discriminator": {
//     "propertyName": "type",
//     "mapping": {
//       "credit_card": "#/definitions/CreditCard",
//       "bank_transfer": "#/definitions/BankTransfer",
//       "digital_wallet": "#/definitions/DigitalWallet"
//     }
//   }
// }
```

This schema can be published in API documentation (OpenAPI/Swagger) to inform clients about the possible union variants.

## Cross-Field Validation in Variants

Each variant can implement the `Validatable` interface for cross-field checks:

```go
type CreditCard struct {
    Type       string `json:"type" validate:"required"`
    CardNumber string `json:"cardNumber" validate:"required"`
    CVC        string `json:"cvc" validate:"required"`
    ExpiryDate string `json:"expiryDate" validate:"required"`
}

func (c CreditCard) Validate() error {
    // Parse expiry date and check it hasn't passed
    parts := strings.Split(c.ExpiryDate, "/")
    if len(parts) != 2 {
        return errors.New("expiryDate must be in MM/YY format")
    }

    month, err := strconv.Atoi(parts[0])
    if err != nil || month < 1 || month > 12 {
        return errors.New("expiryDate month must be 01-12")
    }

    year, err := strconv.Atoi(parts[1])
    if err != nil {
        return errors.New("expiryDate year is invalid")
    }

    currentYear := time.Now().Year() % 100
    currentMonth := int(time.Now().Month())

    if year < currentYear || (year == currentYear && month < currentMonth) {
        return errors.New("card has expired")
    }

    return nil
}
```

## Complete Example: Payment Processing

Here's a complete payment processing example:

```go
package main

import (
    "errors"
    "fmt"
    "github.com/SmrutAI/pedantigo/v2/validator"
)

// Define payment method variants
type CreditCard struct {
    Type       string `json:"type" validate:"required"`
    CardNumber string `json:"cardNumber" validate:"required,pattern=^[0-9]{16}$"`
    CVC        string `json:"cvc" validate:"required,pattern=^[0-9]{3}$"`
    ExpiryDate string `json:"expiryDate" validate:"required"`
}

func (c CreditCard) Validate() error {
    // Validate expiry date format
    if len(c.ExpiryDate) != 5 || c.ExpiryDate[2] != '/' {
        return errors.New("expiryDate must be in MM/YY format")
    }
    return nil
}

type BankTransfer struct {
    Type           string `json:"type" validate:"required"`
    AccountNumber  string `json:"accountNumber" validate:"required,pattern=^[0-9]{10,12}$"`
    RoutingNumber  string `json:"routingNumber" validate:"required,pattern=^[0-9]{9}$"`
}

type DigitalWallet struct {
    Type     string `json:"type" validate:"required"`
    WalletID string `json:"walletId" validate:"required,min=1"`
    Provider string `json:"provider" validate:"required,enum=apple_pay|google_pay|paypal"`
}

func main() {
    // Create union validator once
    unionValidator, err := validator.NewUnion[any](validator.UnionOptions{
        DiscriminatorField: "type",
        Variants: []validator.UnionVariant{
            validator.VariantFor[CreditCard]("credit_card"),
            validator.VariantFor[BankTransfer]("bank_transfer"),
            validator.VariantFor[DigitalWallet]("digital_wallet"),
        },
    })

    if err != nil {
        panic(err)
    }

    // Example 1: Valid credit card
    creditCardJSON := []byte(`{
        "type": "credit_card",
        "cardNumber": "4111111111111111",
        "cvc": "123",
        "expiryDate": "12/25"
    }`)

    result, err := unionValidator.Unmarshal(creditCardJSON)
    if err != nil {
        fmt.Printf("Credit card validation failed: %v\n", err)
        return
    }

    if card, ok := result.(CreditCard); ok {
        fmt.Printf("Processing credit card: %s\n", card.CardNumber)
    }

    // Example 2: Valid bank transfer
    bankJSON := []byte(`{
        "type": "bank_transfer",
        "accountNumber": "12345678901",
        "routingNumber": "987654321"
    }`)

    result, err = unionValidator.Unmarshal(bankJSON)
    if err != nil {
        fmt.Printf("Bank transfer validation failed: %v\n", err)
        return
    }

    if bank, ok := result.(BankTransfer); ok {
        fmt.Printf("Processing bank transfer to account: %s\n", bank.AccountNumber)
    }

    // Example 3: Valid digital wallet
    walletJSON := []byte(`{
        "type": "digital_wallet",
        "walletId": "wallet_abc123",
        "provider": "apple_pay"
    }`)

    result, err = unionValidator.Unmarshal(walletJSON)
    if err != nil {
        fmt.Printf("Digital wallet validation failed: %v\n", err)
        return
    }

    if wallet, ok := result.(DigitalWallet); ok {
        fmt.Printf("Processing digital wallet: %s (%s)\n", wallet.WalletID, wallet.Provider)
    }

    // Example 4: Invalid credit card (bad card number)
    invalidJSON := []byte(`{
        "type": "credit_card",
        "cardNumber": "invalid",
        "cvc": "123",
        "expiryDate": "12/25"
    }`)

    _, err = unionValidator.Unmarshal(invalidJSON)
    if err != nil {
        var ve *validator.ValidationError
        if errors.As(err, &ve) {
            fmt.Println("Validation errors:")
            for _, fieldErr := range ve.Errors {
                fmt.Printf("  Field %s: %s\n", fieldErr.Field, fieldErr.Message)
            }
        }
    }
}
```

## Streaming Discriminated Unions

For LLM outputs or streaming APIs, you can use `StreamParser` with unions:

```go
// Create stream parser for union types
parser := validator.NewStreamUnionParser[any](validator.UnionOptions{
    DiscriminatorField: "type",
    Variants: []validator.UnionVariant{
        validator.VariantFor[CreditCard]("credit_card"),
        validator.VariantFor[BankTransfer]("bank_transfer"),
        validator.VariantFor[DigitalWallet]("digital_wallet"),
    },
})

// Feed streaming data
parser.Feed(`{"type": "credit_card"`)
parser.Feed(`, "cardNumber": "411111`)
parser.Feed(`1111111111"`)
parser.Feed(`, "cvc": "123"`)
parser.Feed(`, "expiryDate": "12/25"}`)

// Get validated result
result, err := parser.Complete()
```

## Best Practices

1. **Always set discriminator first in JSON** - Some streaming scenarios require the discriminator field early
2. **Use consistent discriminator values** - Document the exact values expected (e.g., "credit_card" vs "creditCard")
3. **Implement Validatable for complex variants** - Cross-field validation catches logic errors
4. **Test all variants** - Ensure each variant path is validated properly
5. **Document variants in API docs** - Include the discriminator values and variant schemas
6. **Use type assertions carefully** - Always check the type after unmarshaling, or use switch statements

## Common Pitfalls

:::warning Watch Out For These

1. **Missing discriminator field** - If JSON lacks the discriminator field, you get a clear error: `discriminator field "type" is missing`

2. **Unknown discriminator value** - If the value doesn't match any variant: `unknown discriminator value "fish" for field "type"`

3. **Forgetting type assertions** - `Unmarshal()` returns `any`. You must cast: `card := result.(CreditCard)`

4. **Variant validation still applies** - After type resolution, all constraints on that variant are checked. Invalid fields = validation error.

5. **Discriminator position in streaming** - For streaming JSON, ensure the discriminator field appears early in the JSON structure.

:::

## Lightweight Alternative with skip_unless

For scenarios where you don't need streaming support or JSON Schema `oneOf` generation, `skip_unless` provides an alternative to `UnionValidator` that keeps all variants in a single struct.

### The Pattern

```go
type TV struct {
    Channel int `json:"channel" validate:"required,min=1,max=999"`
}

type Fan struct {
    Speed int `json:"speed" validate:"required,min=1,max=5"`
}

type Suite struct {
    SuiteType string `json:"suite_type" validate:"required,oneof=tv fan"`
    TV        TV     `json:"tv" validate:"skip_unless=SuiteType tv"`
    Fan       Fan    `json:"fan" validate:"skip_unless=SuiteType fan"`
}
```

### How It Works

- `skip_unless=FieldName value` skips **ALL validation** on a field when the condition is NOT met
- When `SuiteType` is `"tv"`: TV is validated, Fan is completely skipped
- When `SuiteType` is `"fan"`: Fan is validated, TV is completely skipped
- Invalid `SuiteType` values are caught by the `oneof` constraint

### Example Usage

```go
suiteValidator := validator.New[Suite]()

// TV mode - TV is validated, Fan is skipped
tvData := Suite{
    SuiteType: "tv",
    TV:        TV{Channel: 42},
    Fan:       Fan{Speed: 0}, // Would fail min=1, but is skipped
}
err := suiteValidator.Validate(&tvData) // ✓ Valid

// Fan mode - Fan is validated, TV is skipped
fanData := Suite{
    SuiteType: "fan",
    TV:        TV{Channel: 0}, // Would fail min=1, but is skipped
    Fan:       Fan{Speed: 3},
}
err = suiteValidator.Validate(&fanData) // ✓ Valid

// TV mode with invalid TV - fails validation
invalidData := Suite{
    SuiteType: "tv",
    TV:        TV{Channel: 0}, // Fails: min=1
    Fan:       Fan{Speed: 0},
}
err = suiteValidator.Validate(&invalidData) // ✗ Error: tv.channel must be at least 1
```

### When to Use skip_unless vs UnionValidator

| Feature | `skip_unless` | `UnionValidator` |
|---------|--------------|------------------|
| Single struct | ✓ | ✗ (separate variant types) |
| Setup complexity | Simple | Moderate |
| JSON Schema `oneOf` | ✗ | ✓ |
| Streaming support | ✗ | ✓ |
| Type assertion needed | ✗ | ✓ |
| Multiple discriminator values | ✗ | ✓ |

**Use `skip_unless` when:**
- You want a simple, single-struct solution
- You don't need JSON Schema `oneOf` generation
- You don't need streaming validation
- Each discriminator value maps to specific fields in one struct

**Use `UnionValidator` when:**
- You need distinct variant types with their own methods
- You need JSON Schema `oneOf` with discriminator mapping
- You're using streaming JSON (LLM outputs)
- You need the full type-safe union pattern

## Key Differences from the Simple API

Discriminated unions cannot use the Simple API because:
- They return `any` type (needs type assertion)
- They require explicit variant registration
- They need detailed configuration (discriminator field, variant mapping)

This is why `validator.NewUnion()` is required instead of `validator.Unmarshal[T]()`.

## See Also

- [Validation Basics](/docs/concepts/validation) - Core validation concepts
- [Streaming Validation](/docs/concepts/streaming) - For LLM outputs and real-time data
- [Cross-Field Validation](/docs/concepts/cross-field) - Validating relationships between fields
