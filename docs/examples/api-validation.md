---
sidebar_position: 2
---

# API Validation

Integrate Pedantigo with HTTP APIs for request and response validation. This guide covers popular Go frameworks and best practices for handling validation errors in your API handlers.

## Quick Start

Pedantigo works with any Go HTTP framework. The pattern is simple:

1. **Read the request body**
2. **Pass to `pedantigo.Unmarshal[T]`** to parse and validate
3. **Return validation errors as JSON** if validation fails
4. **Process the validated data** if validation succeeds

```go
type CreateUserRequest struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Age      int    `json:"age" pedantigo:"min=18,max=120"`
}

// Works with any HTTP framework
user, err := pedantigo.Unmarshal[CreateUserRequest](body)
if err != nil {
    // Return validation error as JSON
    return
}

// user is validated - safe to use
```

## net/http Handler

Using Go's standard library `net/http`:

```go
package main

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "github.com/smrutai/pedantigo"
)

type CreateUserRequest struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
}

type ErrorResponse struct {
    Error  string                    `json:"error"`
    Errors []pedantigo.FieldError `json:"errors,omitempty"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // Read request body
    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }

    // Parse and validate
    req, err := pedantigo.Unmarshal[CreateUserRequest](body)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)

        // Check if validation error
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            json.NewEncoder(w).Encode(ErrorResponse{
                Error:  "validation_failed",
                Errors: validationErr.Errors,
            })
        } else {
            json.NewEncoder(w).Encode(ErrorResponse{
                Error: "invalid_request",
            })
        }
        return
    }

    // User is validated - create in database
    // user := db.CreateUser(req.Email, req.Username, req.Age)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "User created successfully",
        "email":   req.Email,
    })
}

func main() {
    http.HandleFunc("POST /users", CreateUserHandler)
    http.ListenAndServe(":8080", nil)
}
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid","username":"ab"}'
```

**Example Response (400 Bad Request):**
```json
{
  "error": "validation_failed",
  "errors": [
    {
      "field": "email",
      "code": "INVALID_EMAIL",
      "message": "must be a valid email address",
      "value": "invalid"
    },
    {
      "field": "username",
      "code": "MIN_LENGTH",
      "message": "must be at least 3 characters",
      "value": "ab"
    }
  ]
}
```

## Gin Framework

For Gin-based APIs:

```go
package main

import (
    "errors"
    "io"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/smrutai/pedantigo"
)

type CreateUserRequest struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
}

type ErrorResponse struct {
    Error  string                    `json:"error"`
    Errors []pedantigo.FieldError `json:"errors,omitempty"`
}

func CreateUserHandler(c *gin.Context) {
    // Read request body
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
        return
    }
    defer c.Request.Body.Close()

    // Parse and validate
    req, err := pedantigo.Unmarshal[CreateUserRequest](body)
    if err != nil {
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error:  "validation_failed",
                Errors: validationErr.Errors,
            })
        } else {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error: "invalid_request",
            })
        }
        return
    }

    // User is validated
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created successfully",
        "email":   req.Email,
    })
}

func main() {
    r := gin.Default()
    r.POST("/users", CreateUserHandler)
    r.Run(":8080")
}
```

**Usage:**
```go
// In your main router setup
router := gin.Default()
router.POST("/users", CreateUserHandler)
router.POST("/products", CreateProductHandler)
// ... other routes
```

## Echo Framework

For Echo-based APIs:

```go
package main

import (
    "errors"
    "io"
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/smrutai/pedantigo"
)

type CreateUserRequest struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
}

type ErrorResponse struct {
    Error  string                    `json:"error"`
    Errors []pedantigo.FieldError `json:"errors,omitempty"`
}

func CreateUserHandler(c echo.Context) error {
    // Read request body
    body, err := io.ReadAll(c.Request().Body)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "failed to read body",
        })
    }
    defer c.Request().Body.Close()

    // Parse and validate
    req, err := pedantigo.Unmarshal[CreateUserRequest](body)
    if err != nil {
        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Error:  "validation_failed",
                Errors: validationErr.Errors,
            })
        }
        return c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "invalid_request",
        })
    }

    // User is validated
    return c.JSON(http.StatusCreated, map[string]string{
        "message": "User created successfully",
        "email":   req.Email,
    })
}

func main() {
    e := echo.New()
    e.POST("/users", CreateUserHandler)
    e.Logger.Fatal(e.Start(":8080"))
}
```

## Error Response Formatting

### Organizing Errors by Field

For frontend applications, it's helpful to organize errors by field so each input can display its specific error:

```go
type FieldErrorResponse struct {
    Error       string              `json:"error"`
    FieldErrors map[string][]string `json:"field_errors"`
}

func FormatValidationErrors(validationErr *pedantigo.ValidationError) FieldErrorResponse {
    fieldErrors := make(map[string][]string)

    for _, fieldErr := range validationErr.Errors {
        fieldErrors[fieldErr.Field] = append(
            fieldErrors[fieldErr.Field],
            fieldErr.Message,
        )
    }

    return FieldErrorResponse{
        Error:       "validation_failed",
        FieldErrors: fieldErrors,
    }
}
```

**Example Response:**
```json
{
  "error": "validation_failed",
  "field_errors": {
    "email": ["must be a valid email address"],
    "username": ["must be at least 3 characters", "must be at most 20 characters"],
    "age": ["must be at least 18"]
  }
}
```

### Nested Field Errors

For nested structs, field paths include the full path:

```go
type Address struct {
    Street string `json:"street" pedantigo:"required,min=5"`
    City   string `json:"city" pedantigo:"required,min=2"`
    Zip    string `json:"zip" pedantigo:"required,pattern=^\\d{5}$"`
}

type CreateUserRequest struct {
    Email   string  `json:"email" pedantigo:"required,email"`
    Address Address `json:"address" pedantigo:"required"`
}

// Invalid request:
// {"email":"user@example.com","address":{"street":"Main","city":"NY","zip":"ABC"}}

// Error response:
// {
//   "field_errors": {
//     "address.street": ["must be at least 5 characters"],
//     "address.city": ["must be at least 2 characters"],
//     "address.zip": ["must match pattern ^\\d{5}$"]
//   }
// }
```

### Array Element Errors

For validation errors in array elements, the field path includes the index:

```go
type CreateOrderRequest struct {
    Items []OrderItem `json:"items" pedantigo:"required,minItems=1,maxItems=100"`
}

type OrderItem struct {
    ProductID string `json:"product_id" pedantigo:"required,uuid"`
    Quantity  int    `json:"quantity" pedantigo:"required,min=1,max=999"`
    Price     float64 `json:"price" pedantigo:"required,gt=0"`
}

// Error paths:
// "items[0].product_id": ["must be a valid UUID"]
// "items[1].quantity": ["must be at least 1"]
// "items[2].price": ["must be greater than 0"]
```

## Request + Response Validation

Validate both incoming requests and outgoing responses for complete API correctness:

```go
package main

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "github.com/smrutai/pedantigo"
)

// Request types
type CreateProductRequest struct {
    Name        string  `json:"name" pedantigo:"required,min=1,max=200"`
    Description string  `json:"description" pedantigo:"max=2000"`
    Price       float64 `json:"price" pedantigo:"required,gt=0"`
    SKU         string  `json:"sku" pedantigo:"required,pattern=^[A-Z0-9-]+$"`
}

// Response types
type Product struct {
    ID          string  `json:"id" pedantigo:"required,uuid"`
    Name        string  `json:"name" pedantigo:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" pedantigo:"required,gt=0"`
    SKU         string  `json:"sku" pedantigo:"required"`
    CreatedAt   string  `json:"created_at" pedantigo:"required,datetime"`
}

type CreateProductResponse struct {
    Success bool    `json:"success" pedantigo:"required"`
    Product Product `json:"product" pedantigo:"required"`
    Message string  `json:"message"`
}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
    // Validate request
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()

    req, err := pedantigo.Unmarshal[CreateProductRequest](body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]any{
            "error": "validation_failed",
        })
        return
    }

    // Create product (simulated)
    product := Product{
        ID:          "550e8400-e29b-41d4-a716-446655440000",
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        SKU:         req.SKU,
        CreatedAt:   "2024-12-18T10:30:00Z",
    }

    // Construct response
    response := CreateProductResponse{
        Success: true,
        Product: product,
        Message: "Product created successfully",
    }

    // Validate response before sending
    if err := pedantigo.Validate(&response); err != nil {
        // Log the error - response construction failed validation
        // This catches bugs in your response building logic
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "internal_error",
        })
        return
    }

    // Send validated response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}
```

**Benefits:**
- **Request validation**: Catch invalid input immediately
- **Response validation**: Ensure your API always returns well-formed data
- **Contract compliance**: Verify your API implementation matches its spec
- **Early detection**: Find bugs in response building before they reach clients

## OpenAPI/Swagger Integration

Generate OpenAPI-compatible schemas for your API endpoints:

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/smrutai/pedantigo"
)

type CreateUserRequest struct {
    Email    string `json:"email" pedantigo:"required,email"`
    Username string `json:"username" pedantigo:"required,min=3,max=20"`
    Age      int    `json:"age" pedantigo:"required,min=18,max=120"`
}

type User struct {
    ID        string `json:"id" pedantigo:"required,uuid"`
    Email     string `json:"email" pedantigo:"required,email"`
    Username  string `json:"username" pedantigo:"required"`
    Age       int    `json:"age" pedantigo:"required"`
    CreatedAt string `json:"created_at" pedantigo:"required,datetime"`
}

type APIResponse struct {
    Success bool   `json:"success" pedantigo:"required"`
    Data    *User  `json:"data"`
    Message string `json:"message"`
}

// Handler to serve OpenAPI schemas
func SchemaHandler(w http.ResponseWriter, r *http.Request) {
    // Get OpenAPI-compatible schemas for your types
    requestSchema := pedantigo.SchemaOpenAPI[CreateUserRequest]()
    responseSchema := pedantigo.SchemaOpenAPI[APIResponse]()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "request":  requestSchema,
        "response": responseSchema,
    })
}

// Or serve as raw JSON bytes
func SchemaJSONHandler(w http.ResponseWriter, r *http.Request) {
    schemaBytes, err := pedantigo.SchemaJSONOpenAPI[CreateUserRequest]()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(schemaBytes)
}
```

**OpenAPI Spec Integration:**

```yaml
openapi: 3.0.0
info:
  title: User API
  version: 1.0.0

paths:
  /users:
    post:
      summary: Create a user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/APIResponse'
        '400':
          description: Validation error

components:
  schemas:
    CreateUserRequest:
      # Generated by pedantigo.SchemaOpenAPI[CreateUserRequest]()
    APIResponse:
      # Generated by pedantigo.SchemaOpenAPI[APIResponse]()
```

## Reusable Error Handler

Create a helper function to reduce boilerplate across handlers:

```go
package middleware

import (
    "encoding/json"
    "errors"
    "net/http"
    "github.com/smrutai/pedantigo"
)

type ErrorResponse struct {
    Error       string                    `json:"error"`
    Errors      []pedantigo.FieldError `json:"errors,omitempty"`
    DebugInfo   string                 `json:"debug,omitempty"`
}

// ValidateRequest unmarshals and validates JSON request body
func ValidateRequest[T any](w http.ResponseWriter, body []byte) (*T, bool) {
    data, err := pedantigo.Unmarshal[T](body)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)

        var validationErr *pedantigo.ValidationError
        if errors.As(err, &validationErr) {
            json.NewEncoder(w).Encode(ErrorResponse{
                Error:  "validation_failed",
                Errors: validationErr.Errors,
            })
        } else {
            json.NewEncoder(w).Encode(ErrorResponse{
                Error: "invalid_request",
            })
        }
        return nil, false
    }
    return data, true
}

// SendError sends an error response
func SendError(w http.ResponseWriter, statusCode int, errorMsg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: errorMsg,
    })
}

// SendSuccess sends a successful JSON response
func SendSuccess[T any](w http.ResponseWriter, statusCode int, data T) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(data)
}
```

**Usage:**

```go
import (
    "io"
    "net/http"
    "yourapp/middleware"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()

    // Validate and get request
    req, ok := middleware.ValidateRequest[CreateUserRequest](w, body)
    if !ok {
        return // Error response already sent
    }

    // Process validated request
    user := db.CreateUser(req.Email, req.Username, req.Age)

    // Send success response
    middleware.SendSuccess(w, http.StatusCreated, user)
}
```

## Complete Example: Full CRUD API

Here's a complete example showing all validation patterns together:

```go
package main

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "uuid"
    "github.com/smrutai/pedantigo"
)

// Request DTOs
type CreateProductRequest struct {
    Name        string  `json:"name" pedantigo:"required,min=1,max=200"`
    Description string  `json:"description" pedantigo:"max=2000"`
    Price       float64 `json:"price" pedantigo:"required,gt=0"`
    StockQty    int     `json:"stock_qty" pedantigo:"required,min=0"`
}

type UpdateProductRequest struct {
    Name        *string  `json:"name" pedantigo:"min=1,max=200"`
    Description *string  `json:"description" pedantigo:"max=2000"`
    Price       *float64 `json:"price" pedantigo:"gt=0"`
    StockQty    *int     `json:"stock_qty" pedantigo:"min=0"`
}

// Response DTOs
type Product struct {
    ID          string  `json:"id" pedantigo:"required,uuid"`
    Name        string  `json:"name" pedantigo:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" pedantigo:"required,gt=0"`
    StockQty    int     `json:"stock_qty" pedantigo:"required,min=0"`
    CreatedAt   string  `json:"created_at" pedantigo:"required"`
}

type ListResponse struct {
    Success bool      `json:"success" pedantigo:"required"`
    Data    []Product `json:"data" pedantigo:"required"`
    Count   int       `json:"count" pedantigo:"required,min=0"`
}

type SingleResponse struct {
    Success bool    `json:"success" pedantigo:"required"`
    Data    Product `json:"data" pedantigo:"required"`
}

// Database (simulated)
var products = map[string]Product{}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()

    req, err := pedantigo.Unmarshal[CreateProductRequest](body)
    if err != nil {
        writeValidationError(w, err)
        return
    }

    product := Product{
        ID:          uuid.New().String(),
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        StockQty:    req.StockQty,
        CreatedAt:   "2024-12-18T10:30:00Z",
    }

    // Validate response
    response := SingleResponse{
        Success: true,
        Data:    product,
    }
    if err := pedantigo.Validate(&response); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    products[product.ID] = product

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

func GetAllProductsHandler(w http.ResponseWriter, r *http.Request) {
    productList := make([]Product, 0, len(products))
    for _, p := range products {
        productList = append(productList, p)
    }

    response := ListResponse{
        Success: true,
        Data:    productList,
        Count:   len(productList),
    }

    // Validate response
    if err := pedantigo.Validate(&response); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
    productID := r.PathValue("id")

    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()

    req, err := pedantigo.Unmarshal[UpdateProductRequest](body)
    if err != nil {
        writeValidationError(w, err)
        return
    }

    product, exists := products[productID]
    if !exists {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
        return
    }

    // Apply updates
    if req.Name != nil {
        product.Name = *req.Name
    }
    if req.Description != nil {
        product.Description = *req.Description
    }
    if req.Price != nil {
        product.Price = *req.Price
    }
    if req.StockQty != nil {
        product.StockQty = *req.StockQty
    }

    products[productID] = product

    response := SingleResponse{
        Success: true,
        Data:    product,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func writeValidationError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)

    var validationErr *pedantigo.ValidationError
    if errors.As(err, &validationErr) {
        json.NewEncoder(w).Encode(map[string]any{
            "error":  "validation_failed",
            "errors": validationErr.Errors,
        })
    } else {
        json.NewEncoder(w).Encode(map[string]string{
            "error": "invalid_request",
        })
    }
}

func main() {
    http.HandleFunc("POST /api/products", CreateProductHandler)
    http.HandleFunc("GET /api/products", GetAllProductsHandler)
    http.HandleFunc("PATCH /api/products/{id}", UpdateProductHandler)
    http.ListenAndServe(":8080", nil)
}
```

## Best Practices

1. **Always use `pedantigo` struct tag (not `validate`)** - This is required for the Simple API
2. **Handle validation errors explicitly** - Don't ignore errors; return them to the client
3. **Validate responses** - Catch bugs in response building before they reach clients
4. **Organize nested errors** - Group errors by field for better frontend integration
5. **Use OpenAPI schemas** - Generate schemas for documentation and client generation
6. **Reuse error handlers** - Create middleware to reduce boilerplate
7. **Type-safe DTOs** - Use separate request/response types to evolve your API safely
8. **Test edge cases** - Test validation with boundary values and invalid inputs

## See Also

- [Error Handling](/docs/api/errors) - Complete error handling guide
- [Simple API](/docs/api/simple-api) - Full Simple API reference
- [Constraints](/docs/constraints/string) - All available validation constraints
- [Cross-Field Validation](/docs/concepts/cross-field) - Implement complex validation rules
