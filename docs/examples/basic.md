---
sidebar_position: 1
---

# Common Patterns

Real-world validation examples using Pedantigo's Simple API with complete runnable code.

## User Registration

User registration forms often need email validation, password confirmation, and username constraints.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.your-org/pedantigo"
)

type User struct {
    Username            string `json:"username" pedantigo:"required,alphanum,min=3,max=20"`
    Email               string `json:"email" pedantigo:"required,email"`
    Password            string `json:"password" pedantigo:"required,min=8"`
    PasswordConfirm     string `json:"password_confirm" pedantigo:"required"`
}

func main() {
    // Valid registration data
    validJSON := `{
        "username": "john_doe",
        "email": "john@example.com",
        "password": "SecurePass123",
        "password_confirm": "SecurePass123"
    }`

    user, errs := pedantigo.Unmarshal[User]([]byte(validJSON))
    if errs != nil {
        for _, err := range errs {
            fmt.Printf("Validation error: %v\n", err)
        }
    } else {
        fmt.Printf("User created: %s (%s)\n", user.Username, user.Email)
    }

    // Invalid: Username too short
    invalidJSON := `{
        "username": "ab",
        "email": "jane@example.com",
        "password": "SecurePass123",
        "password_confirm": "SecurePass123"
    }`

    _, errs = pedantigo.Unmarshal[User]([]byte(invalidJSON))
    if errs != nil {
        fmt.Println("\nValidation errors found:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Invalid email format
    invalidEmail := `{
        "username": "jane_doe",
        "email": "not-an-email",
        "password": "SecurePass123",
        "password_confirm": "SecurePass123"
    }`

    _, errs = pedantigo.Unmarshal[User]([]byte(invalidEmail))
    if errs != nil {
        fmt.Println("\nEmail validation errors:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }
}
```

**Key Constraints:**
- `required` - Field must be present and non-empty
- `alphanum` - Only alphanumeric characters (letters and numbers)
- `min=3,max=20` - String length between 3 and 20 characters
- `email` - Valid email format (RFC 5322)

---

## E-commerce Product

Product validation requires price constraints, category enumeration, and pattern-based SKU validation.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.your-org/pedantigo"
)

type Product struct {
    Name     string  `json:"name" pedantigo:"required,max=200"`
    Price    float64 `json:"price" pedantigo:"required,positive"`
    Quantity int     `json:"quantity" pedantigo:"required,gte=0"`
    Category string  `json:"category" pedantigo:"required,oneof=electronics|clothing|books|home"`
    SKU      string  `json:"sku" pedantigo:"required,pattern=^[A-Z0-9]{3}-[A-Z0-9]{3}-[A-Z0-9]{4}$"`
}

func main() {
    // Valid product
    validJSON := `{
        "name": "Wireless Headphones",
        "price": 79.99,
        "quantity": 50,
        "category": "electronics",
        "sku": "WHP-BT5-2024"
    }`

    product, errs := pedantigo.Unmarshal[Product]([]byte(validJSON))
    if errs == nil {
        fmt.Printf("Product: %s - $%.2f (Stock: %d)\n", product.Name, product.Price, product.Quantity)
        fmt.Printf("SKU: %s | Category: %s\n", product.SKU, product.Category)
    }

    // Invalid: Negative price
    invalidPrice := `{
        "name": "Broken Item",
        "price": -10.00,
        "quantity": 5,
        "category": "electronics",
        "sku": "BRK-ITM-0001"
    }`

    _, errs = pedantigo.Unmarshal[Product]([]byte(invalidPrice))
    if errs != nil {
        fmt.Println("\nPrice validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Category not in allowed list
    invalidCategory := `{
        "name": "Mystery Item",
        "price": 15.00,
        "quantity": 10,
        "category": "toys",
        "sku": "MYS-TRY-0001"
    }`

    _, errs = pedantigo.Unmarshal[Product]([]byte(invalidCategory))
    if errs != nil {
        fmt.Println("\nCategory validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: SKU format mismatch
    invalidSKU := `{
        "name": "Invalid SKU Item",
        "price": 25.00,
        "quantity": 20,
        "category": "books",
        "sku": "invalid-sku-format"
    }`

    _, errs = pedantigo.Unmarshal[Product]([]byte(invalidSKU))
    if errs != nil {
        fmt.Println("\nSKU format validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }
}
```

**Key Constraints:**
- `required` - Field must be present
- `max=200` - Maximum 200 characters
- `positive` - Must be greater than 0
- `gte=0` - Greater than or equal to 0 (allows zero)
- `oneof=...` - Value must match one of the options
- `pattern=...` - Regular expression matching (3-4 alphanumeric segments separated by hyphens)

---

## Configuration File

Configuration validation ensures type safety, port ranges, and security for API keys and timeouts.

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"

    "github.your-org/pedantigo"
)

type Config struct {
    Server  ServerConfig  `json:"server" pedantigo:"required"`
    API     APIConfig     `json:"api" pedantigo:"required"`
    Timeout time.Duration `json:"timeout" pedantigo:"required,positive"`
}

type ServerConfig struct {
    Host string `json:"host" pedantigo:"required,hostname"`
    Port int    `json:"port" pedantigo:"required,port"`
}

type APIConfig struct {
    Key         string `json:"key" pedantigo:"required,min=32"`
    Environment string `json:"env" pedantigo:"required,oneof=development|staging|production"`
}

func main() {
    // Valid configuration
    validJSON := `{
        "server": {
            "host": "api.example.com",
            "port": 8080
        },
        "api": {
            "key": "sk_prod_aB1cD2eF3gH4iJ5kL6mN7oP8qR9sT0u",
            "env": "production"
        },
        "timeout": 30000000000
    }`

    config, errs := pedantigo.Unmarshal[Config]([]byte(validJSON))
    if errs == nil {
        fmt.Printf("Server: %s:%d\n", config.Server.Host, config.Server.Port)
        fmt.Printf("Environment: %s\n", config.API.Environment)
        fmt.Printf("Timeout: %v\n", config.Timeout)
    }

    // Invalid: Port out of range
    invalidPort := `{
        "server": {
            "host": "localhost",
            "port": 99999
        },
        "api": {
            "key": "sk_test_aB1cD2eF3gH4iJ5kL6mN7oP8qR9sT0u",
            "env": "development"
        },
        "timeout": 30000000000
    }`

    _, errs = pedantigo.Unmarshal[Config]([]byte(invalidPort))
    if errs != nil {
        fmt.Println("\nPort validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: API key too short
    invalidKey := `{
        "server": {
            "host": "localhost",
            "port": 3000
        },
        "api": {
            "key": "short_key",
            "env": "staging"
        },
        "timeout": 30000000000
    }`

    _, errs = pedantigo.Unmarshal[Config]([]byte(invalidKey))
    if errs != nil {
        fmt.Println("\nAPI key validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }
}
```

**Key Constraints:**
- `hostname` - Valid hostname format
- `port` - Valid port number (1-65535)
- `min=32` - Minimum 32 characters (suitable for API keys)
- `oneof=...` - Must be one of the specified environments

---

## Blog Post

Blog validation includes content length requirements, tag management with uniqueness, and email validation for authors.

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"

    "github.your-org/pedantigo"
)

type BlogPost struct {
    Title       string    `json:"title" pedantigo:"required,min=5,max=200"`
    Content     string    `json:"content" pedantigo:"required,min=100"`
    Author      string    `json:"author" pedantigo:"required,email"`
    Tags        []string  `json:"tags" pedantigo:"required,min=1,max=10,unique"`
    PublishedAt *time.Time `json:"published_at,omitempty" pedantigo:""`
}

func main() {
    // Valid blog post
    now := time.Now()
    validJSON := `{
        "title": "Getting Started with Go Validation",
        "content": "This comprehensive guide covers the fundamentals of validation in Go applications. We'll explore various techniques and best practices for ensuring data integrity throughout your application lifecycle. Proper validation is crucial for maintaining system security and reliability.",
        "author": "alice@example.com",
        "tags": ["go", "validation", "tutorial"],
        "published_at": "2024-12-18T10:30:00Z"
    }`

    post, errs := pedantigo.Unmarshal[BlogPost]([]byte(validJSON))
    if errs == nil {
        fmt.Printf("Title: %s\n", post.Title)
        fmt.Printf("Author: %s\n", post.Author)
        fmt.Printf("Tags: %v\n", post.Tags)
        if post.PublishedAt != nil {
            fmt.Printf("Published: %v\n", post.PublishedAt)
        }
    }

    // Invalid: Title too short
    invalidTitle := `{
        "title": "Blog",
        "content": "This is a longer piece of content that meets the minimum character requirement for blog posts. It's important to have sufficient detail.",
        "author": "bob@example.com",
        "tags": ["short"],
        "published_at": null
    }`

    _, errs = pedantigo.Unmarshal[BlogPost]([]byte(invalidTitle))
    if errs != nil {
        fmt.Println("\nTitle validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Duplicate tags
    duplicateTags := `{
        "title": "Duplicate Tags Example",
        "content": "This post demonstrates what happens when you use duplicate tags in the tags array. The unique constraint will catch this error.",
        "author": "carol@example.com",
        "tags": ["go", "validation", "go"],
        "published_at": null
    }`

    _, errs = pedantigo.Unmarshal[BlogPost]([]byte(duplicateTags))
    if errs != nil {
        fmt.Println("\nTag validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: No tags provided
    noTags := `{
        "title": "Post Without Tags",
        "content": "Every blog post should have at least one tag to categorize the content properly and improve discoverability.",
        "author": "dave@example.com",
        "tags": [],
        "published_at": null
    }`

    _, errs = pedantigo.Unmarshal[BlogPost]([]byte(noTags))
    if errs != nil {
        fmt.Println("\nTag count validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }
}
```

**Key Constraints:**
- `min=5,max=200` on strings - Length constraints
- `min=100` - Content must have at least 100 characters
- `email` - Author must be a valid email
- `min=1,max=10` on slices - Between 1 and 10 tags
- `unique` - All tags must be distinct (no duplicates)

---

## Address Validation

Address validation uses country codes, state abbreviations, and postal code patterns for different formats.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.your-org/pedantigo"
)

type Address struct {
    Street  string `json:"street" pedantigo:"required,min=5,max=100"`
    City    string `json:"city" pedantigo:"required,alpha,min=2"`
    State   string `json:"state" pedantigo:"required,len=2,alpha"`
    ZipCode string `json:"zip_code" pedantigo:"required,pattern=^[0-9]{5}(-[0-9]{4})?$"`
    Country string `json:"country" pedantigo:"required,iso3166_alpha2"`
}

func main() {
    // Valid US address
    validJSON := `{
        "street": "123 Main Street",
        "city": "Springfield",
        "state": "IL",
        "zip_code": "62701",
        "country": "US"
    }`

    address, errs := pedantigo.Unmarshal[Address]([]byte(validJSON))
    if errs == nil {
        fmt.Printf("Address: %s\n", address.Street)
        fmt.Printf("City, State Zip: %s, %s %s\n", address.City, address.State, address.ZipCode)
        fmt.Printf("Country: %s\n", address.Country)
    }

    // Valid: Extended ZIP+4 format
    extendedZip := `{
        "street": "456 Oak Avenue",
        "city": "Portland",
        "state": "OR",
        "zip_code": "97201-1234",
        "country": "US"
    }`

    addr, errs := pedantigo.Unmarshal[Address]([]byte(extendedZip))
    if errs == nil {
        fmt.Printf("\nExtended ZIP: %s-%s\n", addr.City, addr.ZipCode)
    }

    // Invalid: State code not exactly 2 characters
    invalidState := `{
        "street": "789 Pine Road",
        "city": "Denver",
        "state": "Colorado",
        "zip_code": "80202",
        "country": "US"
    }`

    _, errs = pedantigo.Unmarshal[Address]([]byte(invalidState))
    if errs != nil {
        fmt.Println("\nState validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: City contains numbers
    invalidCity := `{
        "street": "321 Elm Street",
        "city": "City123",
        "state": "CA",
        "zip_code": "90210",
        "country": "US"
    }`

    _, errs = pedantigo.Unmarshal[Address]([]byte(invalidCity))
    if errs != nil {
        fmt.Println("\nCity validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: ZIP code format
    invalidZip := `{
        "street": "654 Maple Lane",
        "city": "Seattle",
        "state": "WA",
        "zip_code": "9821",
        "country": "US"
    }`

    _, errs = pedantigo.Unmarshal[Address]([]byte(invalidZip))
    if errs != nil {
        fmt.Println("\nZIP code validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Invalid country code
    invalidCountry := `{
        "street": "987 Birch Drive",
        "city": "Toronto",
        "state": "ON",
        "zip_code": "M5A 1A1",
        "country": "INVALID"
    }`

    _, errs = pedantigo.Unmarshal[Address]([]byte(invalidCountry))
    if errs != nil {
        fmt.Println("\nCountry validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }
}
```

**Key Constraints:**
- `min=5,max=100` - Street address length
- `alpha` - Only alphabetic characters (no numbers or special characters)
- `len=2` - Exactly 2 characters for state code
- `pattern=...` - ZIP code format with optional ZIP+4 extension
- `iso3166_alpha2` - Valid ISO 3166-1 alpha-2 country code (e.g., US, GB, CA)

---

## Payment Processing

Payment validation requires card number validation, currency codes, and amount constraints for secure financial transactions.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.your-org/pedantigo"
)

type Payment struct {
    Amount      float64 `json:"amount" pedantigo:"required,positive"`
    Currency    string  `json:"currency" pedantigo:"required,iso4217"`
    CardNumber  string  `json:"card_number" pedantigo:"required,credit_card"`
    CVV         string  `json:"cvv" pedantigo:"required,oneof=len:3|len:4"`
    CardHolder  string  `json:"card_holder" pedantigo:"required,alpha"`
}

func main() {
    // Valid Visa payment
    validJSON := `{
        "amount": 99.99,
        "currency": "USD",
        "card_number": "4532015112830366",
        "cvv": "123",
        "card_holder": "John Smith"
    }`

    payment, errs := pedantigo.Unmarshal[Payment]([]byte(validJSON))
    if errs == nil {
        fmt.Printf("Amount: %.2f %s\n", payment.Amount, payment.Currency)
        fmt.Printf("Cardholder: %s\n", payment.CardHolder)
    }

    // Valid American Express (4-digit CVV)
    amexJSON := `{
        "amount": 150.00,
        "currency": "EUR",
        "card_number": "374245455400126",
        "cvv": "1234",
        "card_holder": "Jane Doe"
    }`

    amex, errs := pedantigo.Unmarshal[Payment]([]byte(amexJSON))
    if errs == nil {
        fmt.Printf("\nAmEx Payment: %.2f %s\n", amex.Amount, amex.Currency)
    }

    // Invalid: Zero amount
    zeroAmount := `{
        "amount": 0,
        "currency": "USD",
        "card_number": "4532015112830366",
        "cvv": "123",
        "card_holder": "Test User"
    }`

    _, errs = pedantigo.Unmarshal[Payment]([]byte(zeroAmount))
    if errs != nil {
        fmt.Println("\nAmount validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Invalid currency code
    invalidCurrency := `{
        "amount": 50.00,
        "currency": "XYZ",
        "card_number": "4532015112830366",
        "cvv": "123",
        "card_holder": "Test User"
    }`

    _, errs = pedantigo.Unmarshal[Payment]([]byte(invalidCurrency))
    if errs != nil {
        fmt.Println("\nCurrency validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Card number fails checksum
    invalidCard := `{
        "amount": 75.50,
        "currency": "USD",
        "card_number": "4532015112830367",
        "cvv": "123",
        "card_holder": "Test User"
    }`

    _, errs = pedantigo.Unmarshal[Payment]([]byte(invalidCard))
    if errs != nil {
        fmt.Println("\nCard number validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: CVV length incorrect
    invalidCVV := `{
        "amount": 100.00,
        "currency": "USD",
        "card_number": "4532015112830366",
        "cvv": "12",
        "card_holder": "Test User"
    }`

    _, errs = pedantigo.Unmarshal[Payment]([]byte(invalidCVV))
    if errs != nil {
        fmt.Println("\nCVV validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }

    // Invalid: Cardholder with numbers
    invalidHolder := `{
        "amount": 100.00,
        "currency": "USD",
        "card_number": "4532015112830366",
        "cvv": "123",
        "card_holder": "Test User 123"
    }`

    _, errs = pedantigo.Unmarshal[Payment]([]byte(invalidHolder))
    if errs != nil {
        fmt.Println("\nCardholder validation failed:")
        for _, err := range errs {
            fmt.Printf("  - %v\n", err)
        }
    }
}
```

**Key Constraints:**
- `required` - Field must be present
- `positive` - Amount must be greater than 0
- `iso4217` - Valid ISO 4217 currency code (e.g., USD, EUR, GBP)
- `credit_card` - Valid credit card number (Luhn algorithm validation)
- `oneof=len:3|len:4` - CVV must be exactly 3 or 4 digits
- `alpha` - Only alphabetic characters (cardholder name)

---

## Next Steps

Explore more advanced topics:
- **Constraints Reference** - Complete list of all available constraints
- **Custom Validators** - Implement custom validation logic
- **Error Handling** - Detailed error information and field paths
- **Performance** - Schema caching and validation benchmarks
