---
sidebar_position: 1
---

# Secret Types

Safely handle sensitive data with `SecretStr` and `SecretBytes` to prevent accidental exposure in logs, API responses, and error messages.

## The Problem

Sensitive data like passwords, API keys, encryption keys, and tokens are frequently exposed by accident:

- **In logs**: Developers print full structs with `fmt.Printf`, revealing secrets
- **In API responses**: Marshaling to JSON without masking sensitive fields
- **In error messages**: Stack traces or validation errors containing actual values
- **In middleware**: Logging frameworks capturing request/response bodies

Pedantigo's secret types automatically mask sensitive data while preserving the actual value for internal use.

---

## SecretStr

Use `SecretStr` for string-based secrets: passwords, API keys, tokens, and other sensitive strings.

### Overview

```go
type SecretStr struct {
    // value is unexported and private
}

// Create a new secret
func NewSecretStr(s string) SecretStr

// Get the actual secret value
func (s SecretStr) Value() string

// Returns "**********" (safe for logs)
func (s SecretStr) String() string

// Returns "**********" (safe for JSON output)
func (s SecretStr) MarshalJSON() ([]byte, error)

// Stores actual value from JSON input
func (s *SecretStr) UnmarshalJSON(data []byte) error
```

### Behavior

| Method | Output | Use Case |
|--------|--------|----------|
| `Value()` | Actual secret | Internal processing, passing to external services |
| `String()` | "**********" | fmt.Printf, string concatenation, logging |
| `MarshalJSON()` | "**********" | JSON serialization, API responses |
| `UnmarshalJSON()` | Stores actual value | JSON deserialization, form input |

### Example: In Struct Definitions

```go
type Config struct {
    // Database connection string (sensitive)
    DatabaseURL SecretStr `json:"database_url" pedantigo:"required"`

    // API authentication token
    APIToken SecretStr `json:"api_token" pedantigo:"required,min=20"`

    // Webhook secret for signature verification
    WebhookSecret SecretStr `json:"webhook_secret" pedantigo:"required"`
}
```

### Accessing the Value

```go
config, err := pedantigo.Unmarshal[Config](jsonData)
if err != nil {
    return err
}

// Get the actual secret for use
dbURL := config.DatabaseURL.Value()  // Returns: "postgres://user:pass@host/db"

// Connect to database
db, err := sql.Open("postgres", dbURL)
```

### Logging Safety

When logging structs, the secret is automatically masked:

```go
config := &Config{
    DatabaseURL: NewSecretStr("postgres://user:password123@localhost/mydb"),
    APIToken:    NewSecretStr("sk-1234567890abcdefghijklmn"),
}

// Safe - secrets are masked
fmt.Printf("Config: %v\n", config)
// Output: Config: {DatabaseURL:********** APIToken:**********}

// Also safe - String() method returns masked value
log.Infof("Config loaded with token %s", config.APIToken)
// Log: Config loaded with token **********
```

### API Response Safety

When marshaling to JSON, secrets are automatically masked:

```go
config := &Config{
    DatabaseURL: NewSecretStr("postgres://user:password123@localhost/mydb"),
    APIToken:    NewSecretStr("sk-1234567890abcdefghijklmn"),
}

// Marshal for API response
jsonData, err := pedantigo.Marshal(config)
if err != nil {
    return err
}

// The returned JSON will contain masked values
// {"database_url":"**********","api_token":"**********"}

// Safe to send in HTTP response without data leaks
http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)  // No secrets exposed
})
```

### Complete Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/smrutai/pedantigo"
)

type AppConfig struct {
    // Database credentials
    DatabaseURL SecretStr `json:"database_url" pedantigo:"required"`

    // External service API key
    ExternalAPIKey SecretStr `json:"external_api_key" pedantigo:"required,min=20"`

    // Application name (not sensitive)
    AppName string `json:"app_name" pedantigo:"required,min=1"`
}

func main() {
    // 1. Unmarshal from JSON (actual values stored internally)
    jsonInput := []byte(`{
        "database_url": "postgres://admin:secretpass@localhost:5432/myapp",
        "external_api_key": "sk-1234567890abcdefghijklmnopqrst",
        "app_name": "MyApp"
    }`)

    config, err := pedantigo.Unmarshal[AppConfig](jsonInput)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 2. Log the config - secrets are automatically masked
    fmt.Printf("Loaded config: %+v\n", config)
    // Output: Loaded config: {DatabaseURL:********** ExternalAPIKey:********** AppName:MyApp}

    // 3. Access actual secret value only when needed
    dbURL := config.DatabaseURL.Value()  // "postgres://admin:secretpass@localhost:5432/myapp"
    apiKey := config.ExternalAPIKey.Value()  // "sk-1234567890abcdefghijklmnopqrst"

    // Use secrets for actual operations
    setupDatabase(dbURL)
    setupExternalAPI(apiKey)

    // 4. Marshal back to JSON - secrets are masked
    response, _ := pedantigo.Marshal(config)
    fmt.Println(string(response))
    // Output: {"database_url":"**********","external_api_key":"**********","app_name":"MyApp"}

    // Safe to include in API responses without data leaks
}

func setupDatabase(url string) {
    // Use the actual secret
    fmt.Printf("Connecting to database at %s\n", url)
}

func setupExternalAPI(key string) {
    // Use the actual secret
    fmt.Printf("Setting up API with key %s\n", key)
}
```

---

## SecretBytes

Use `SecretBytes` for binary secrets: encryption keys, authentication certificates, and other binary data.

### Overview

```go
type SecretBytes struct {
    // value is unexported and private
}

// Create a new secret from bytes
func NewSecretBytes(b []byte) SecretBytes

// Get the actual secret bytes
func (s SecretBytes) Value() []byte

// Returns "**********" (safe for logs)
func (s SecretBytes) String() string

// Returns "**********" (safe for JSON output)
func (s SecretBytes) MarshalJSON() ([]byte, error)

// Expects base64-encoded string in JSON input
func (s *SecretBytes) UnmarshalJSON(data []byte) error
```

### Base64 Encoding for JSON

Since JSON cannot represent raw binary data, `SecretBytes` expects **base64-encoded strings** in JSON input:

```go
type EncryptionConfig struct {
    // Encryption key as base64-encoded bytes
    EncryptionKey SecretBytes `json:"encryption_key" pedantigo:"required"`
}

// JSON must contain base64-encoded value
jsonInput := []byte(`{
    "encryption_key": "aGVsbG8gd29ybGQ="  // base64 for "hello world"
}`)

config, _ := pedantigo.Unmarshal[EncryptionConfig](jsonInput)

// Access actual bytes
key := config.EncryptionKey.Value()  // []byte("hello world")

// Marshal back to JSON - masked
output, _ := pedantigo.Marshal(config)
// {"encryption_key":"**********"}
```

### Example: Encryption Key Management

```go
package main

import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/base64"
    "fmt"
    "github.com/smrutai/pedantigo"
)

type SecurityConfig struct {
    // 32-byte AES-256 encryption key (base64-encoded in JSON)
    EncryptionKey SecretBytes `json:"encryption_key" pedantigo:"required"`

    // HMAC key for signature verification
    HMACKey SecretBytes `json:"hmac_key" pedantigo:"required"`

    // Security level
    Level string `json:"level" pedantigo:"required,oneof=low medium high"`
}

func main() {
    // Generate keys (in real scenarios, load from secure storage)
    encKey := make([]byte, 32)  // 32 bytes for AES-256
    hmacKey := make([]byte, 32)

    // Encode to base64 for JSON
    encKeyB64 := base64.StdEncoding.EncodeToString(encKey)
    hmacKeyB64 := base64.StdEncoding.EncodeToString(hmacKey)

    // Create JSON input
    jsonInput := []byte(fmt.Sprintf(`{
        "encryption_key": "%s",
        "hmac_key": "%s",
        "level": "high"
    }`, encKeyB64, hmacKeyB64))

    // Load config
    config, err := pedantigo.Unmarshal[SecurityConfig](jsonInput)
    if err != nil {
        fmt.Printf("Failed to load security config: %v\n", err)
        return
    }

    // 1. Log the config - keys are masked
    fmt.Printf("Security config: %+v\n", config)
    // Output: Security config: {EncryptionKey:********** HMACKey:********** Level:high}

    // 2. Create cipher with actual key
    key := config.EncryptionKey.Value()  // Get actual 32 bytes
    block, _ := aes.NewCipher(key)
    cipher := cipher.NewGCM(block)

    // Use cipher for encryption/decryption
    nonce := make([]byte, cipher.NonceSize())
    ciphertext := cipher.Seal(nonce, nonce, []byte("secret data"), nil)
    fmt.Printf("Encrypted data: %s\n", base64.StdEncoding.EncodeToString(ciphertext))

    // 3. Marshal config - keys are masked
    output, _ := pedantigo.Marshal(config)
    // Safe to store or transmit: {"encryption_key":"**********","hmac_key":"**********","level":"high"}
}
```

---

## Complete Integration Example

This example demonstrates using both `SecretStr` and `SecretBytes` together in a real-world scenario:

```go
package main

import (
    "crypto/tls"
    "database/sql"
    "encoding/base64"
    "fmt"
    "log"
    "net/http"
    "github.com/smrutai/pedantigo"
)

type ServerConfig struct {
    // Server basics
    Host string `json:"host" pedantigo:"required"`
    Port int    `json:"port" pedantigo:"required,min=1,max=65535"`

    // Secrets - string-based
    JWTSecret  SecretStr `json:"jwt_secret" pedantigo:"required,min=32"`
    APIKey     SecretStr `json:"api_key" pedantigo:"required"`

    // Secrets - binary (base64-encoded)
    TLSCert SecretBytes `json:"tls_cert" pedantigo:"required"`
    TLSKey  SecretBytes `json:"tls_key" pedantigo:"required"`

    // Non-sensitive fields
    Environment string `json:"environment" pedantigo:"required,oneof=dev staging prod"`
    Timeout     int    `json:"timeout" pedantigo:"min=1,max=300"`
}

func main() {
    // Load config from JSON (from env file, config server, etc)
    configJSON := []byte(`{
        "host": "0.0.0.0",
        "port": 8443,
        "jwt_secret": "your-super-secret-jwt-key-must-be-at-least-32-chars",
        "api_key": "sk-1234567890abcdefghijklmnopqrstu",
        "tls_cert": "` + base64.StdEncoding.EncodeToString([]byte("cert-data")) + `",
        "tls_key": "` + base64.StdEncoding.EncodeToString([]byte("key-data")) + `",
        "environment": "prod",
        "timeout": 30
    }`)

    // Parse and validate
    config, err := pedantigo.Unmarshal[ServerConfig](configJSON)
    if err != nil {
        log.Fatalf("Invalid config: %v", err)
    }

    // Safe to log - secrets are automatically masked
    log.Infof("Starting server at %s:%d", config.Host, config.Port)
    log.Infof("Config: %+v", config)  // Secrets are "**********"

    // Set up HTTPS with actual secrets
    cert := config.TLSCert.Value()  // Get actual certificate bytes
    key := config.TLSKey.Value()    // Get actual key bytes

    tlsCert, err := tls.X509KeyPair(cert, key)
    if err != nil {
        log.Fatalf("Failed to load TLS credentials: %v", err)
    }

    // Set up server
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        // Use JWT secret for token validation
        jwtSecret := config.JWTSecret.Value()

        // Validate request with JWT
        token := r.Header.Get("Authorization")
        if !validateToken(token, jwtSecret) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        w.Write([]byte(`{"status":"ok"}`))
    })

    mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
        // Safe to return config - secrets are masked
        configJSON, _ := pedantigo.Marshal(config)
        w.Header().Set("Content-Type", "application/json")
        w.Write(configJSON)
        // Secrets are "**********" - no data leaks!
    })

    // Start HTTPS server
    server := &http.Server{
        Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
        Handler: mux,
        TLSConfig: &tls.Config{
            Certificates: []tls.Certificate{tlsCert},
        },
    }

    log.Fatal(server.ListenAndServeTLS("", ""))
}

func validateToken(token, secret string) bool {
    // Token validation using secret
    // This is a simplified example
    return token != "" && secret != ""
}
```

---

## Best Practices

### 1. Use the Correct Type

```go
// Good: Use SecretStr for strings
type Config struct {
    DatabasePassword SecretStr `json:"db_password"`
}

// Bad: Storing secrets in plain strings
type Config struct {
    DatabasePassword string `json:"db_password"`  // Will be visible in logs!
}
```

### 2. Always Get Value() When Needed

```go
// Good: Get actual secret only when you need to use it
password := config.DatabasePassword.Value()
db.Connect(password)

// Bad: Logging or printing the struct
log.Println(config)  // Even with SecretStr, don't do this carelessly
fmt.Println(config)  // Use specific fields only
```

### 3. Use Base64 for Binary Secrets

```go
// Good: Use base64-encoded strings in JSON
type Config struct {
    EncryptionKey SecretBytes `json:"encryption_key"`  // Expects base64 input
}

// In JSON:
// {"encryption_key": "aGVsbG8gd29ybGQ="}

// Bad: Trying to pass raw bytes (JSON doesn't support)
// {"encryption_key": [104, 101, 108, 108, 111]}  // Won't unmarshal properly
```

### 4. Combine with Other Constraints

```go
// Good: Add validation constraints
type Config struct {
    APIKey SecretStr `json:"api_key" pedantigo:"required,min=20,max=100"`
    Secret SecretStr `json:"secret" pedantigo:"required"`
}

// This ensures:
// - Fields are required
// - APIKey is 20-100 characters
// - Both values are preserved and masked
```

### 5. Never Store Masked Values

```go
// Good: Always work with actual secrets
actual := config.APIKey.Value()

// Bad: Trying to store the masked output
masked := config.APIKey.String()  // "**********"
useWithAPI(masked)  // This won't work!
```

---

## Error Handling

Both `SecretStr` and `SecretBytes` report validation errors without exposing secrets:

```go
type Config struct {
    APIKey SecretStr `json:"api_key" pedantigo:"required,min=32"`
}

// If validation fails, error message won't contain the actual secret
jsonInput := []byte(`{"api_key": "short"}`)
_, err := pedantigo.Unmarshal[Config](jsonInput)

// Error: validation error: field "api_key": constraint "min" failed
// The actual value "short" is NOT in the error message
log.Println(err)  // Safe to log - no secret exposed
```

---

## Thread Safety

Both `SecretStr` and `SecretBytes` are safe to use concurrently:

```go
var config *ServerConfig
var wg sync.WaitGroup

// Multiple goroutines can access the secrets safely
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()

        // All goroutines can call Value() concurrently
        secret := config.JWTSecret.Value()
        apiKey := config.APIKey.Value()
        encKey := config.EncryptionKey.Value()

        // Process with secrets
        processRequest(secret, apiKey, encKey)
    }()
}

wg.Wait()
```

---

## Summary

| Feature | SecretStr | SecretBytes |
|---------|-----------|------------|
| Use for | Passwords, API keys, tokens | Encryption keys, certificates |
| JSON Input | Plain string | Base64-encoded string |
| `Value()` | Returns actual string | Returns actual bytes |
| `String()` | Returns "**********" | Returns "**********" |
| `MarshalJSON()` | Outputs "**********" | Outputs "**********" |
| Safe for logs | Yes | Yes |
| Safe for JSON output | Yes | Yes |
| Safe for API responses | Yes | Yes |

Both types protect against accidental exposure while preserving full access to the actual secret values when needed.
