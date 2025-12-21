---
sidebar_position: 3
---

# Format Constraints

Format validation rules for emails, URLs, UUIDs, network addresses, cryptographic formats, and domain-specific identifiers.

## Email & URL Formats

### `email`

Validates that a string is a **valid email address** following RFC 5322 specifications.

```go
type Contact struct {
    // Email must be a valid email address
    Email string `json:"email" pedantigo:"required,email"`
    // Backup email is optional but must be valid if provided
    BackupEmail string `json:"backup_email,omitempty" pedantigo:"email"`
}
```

**Valid examples:** "john@example.com", "user.name+tag@company.co.uk", "support@example.org"
**Invalid examples:** "john@", "@example.com", "john.example.com", "john @example.com"

### `url`

Validates that a string is a **valid URL** with a scheme (http, https, ftp, etc.).

```go
type SocialProfile struct {
    // Website URL must be valid
    Website string `json:"website" pedantigo:"required,url"`
    // Optional portfolio URL
    Portfolio string `json:"portfolio,omitempty" pedantigo:"url"`
}
```

**Valid examples:** "https://example.com", "http://sub.example.com:8080/path"
**Invalid examples:** "example.com" (missing scheme), "ht!tp://example.com", "https://"

### `uri`

Validates that a string is a **valid URI** (Uniform Resource Identifier) with any scheme. Unlike `url`, this accepts database URIs, file URIs, and other non-HTTP schemes.

```go
type DatabaseConfig struct {
    // Database connection URI (any scheme allowed)
    ConnectionString string `json:"connection_string" pedantigo:"required,uri"`
}
```

**Valid examples:**
- `"postgres://user:pass@localhost:5432/mydb"` - PostgreSQL
- `"mysql://root@127.0.0.1:3306/app"` - MySQL
- `"redis://localhost:6379/0"` - Redis
- `"mongodb://localhost:27017/testdb"` - MongoDB
- `"s3://bucket-name/key"` - S3
- `"file:///etc/config.json"` - File URI

**Invalid examples:** "example.com" (missing scheme), "/path/to/file" (relative path)

## Identifier Formats

### `uuid`

Validates that a string is a **valid UUID** (Universally Unique Identifier) in standard format.

```go
type Document struct {
    // Unique document identifier must be a valid UUID
    ID string `json:"id" pedantigo:"required,uuid"`
    // Correlation ID for tracking, optional but must be UUID if provided
    CorrelationID string `json:"correlation_id,omitempty" pedantigo:"uuid"`
}
```

**Valid examples:** "550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
**Invalid examples:** "550e8400-e29b-41d4-a716", "not-a-uuid", "550e8400e29b41d4a716446655440000" (missing hyphens)

### `ulid`

Validates that a string is a **valid ULID** (Universally Unique Lexicographically Sortable Identifier).

```go
type Order struct {
    // ULID provides sortable unique ID with timestamp
    ID string `json:"id" pedantigo:"required,ulid"`
}
```

**Valid examples:** "01ARZ3NDEKTSV4RRFFQ69G5FAV", "01BX5ZZKBK4Q58B44HHY4NP0GF"
**Invalid examples:** "550e8400-e29b-41d4-a716-446655440000" (UUID, not ULID), "invalid-ulid"

## Network Address Formats

### `ip`

Validates that a string is a **valid IP address** (IPv4 or IPv6).

```go
type NetworkConfig struct {
    // Accept any valid IP version
    Address string `json:"address" pedantigo:"required,ip"`
}
```

**Valid examples:** "192.168.1.1", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "::1"
**Invalid examples:** "256.256.256.256", "192.168.1", "gggg::1"

### `ipv4`

Validates that a string is a **valid IPv4 address**.

```go
type ServerConfig struct {
    // IPv4 address for legacy systems
    IPv4Address string `json:"ipv4_address" pedantigo:"required,ipv4"`
}
```

**Valid examples:** "192.168.1.1", "10.0.0.1", "255.255.255.255"
**Invalid examples:** "192.168.1", "256.1.1.1", "192.168.1.1.1"

### `ipv6`

Validates that a string is a **valid IPv6 address**.

```go
type ModernNetworkConfig struct {
    // IPv6 address for modern networks
    Address string `json:"address" pedantigo:"required,ipv6"`
}
```

**Valid examples:** "2001:0db8:85a3::8a2e:0370:7334", "::1", "fe80::"
**Invalid examples:** "192.168.1.1" (IPv4), "gggg::1", "2001:0db8:85a3:::8a2e"

### `cidr`

Validates that a string is a **valid CIDR notation** for IPv4 or IPv6.

```go
type NetworkRoute struct {
    // CIDR block for routing, accepts both IPv4 and IPv6
    Network string `json:"network" pedantigo:"required,cidr"`
}
```

**Valid examples:** "192.168.1.0/24", "10.0.0.0/8", "2001:db8::/32"
**Invalid examples:** "192.168.1.1/33" (invalid prefix), "192.168.1.0" (missing prefix)

### `cidrv4`

Validates that a string is a **valid IPv4 CIDR notation**.

```go
type SubnetConfig struct {
    // IPv4 CIDR notation for subnet specification
    Subnet string `json:"subnet" pedantigo:"required,cidrv4"`
}
```

**Valid examples:** "192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12"
**Invalid examples:** "192.168.1/24" (incomplete), "256.1.0.0/8" (invalid IP), "2001:db8::/32" (IPv6)

### `cidrv6`

Validates that a string is a **valid IPv6 CIDR notation**.

```go
type IPv6Network struct {
    // IPv6 CIDR notation for modern networks
    Prefix string `json:"prefix" pedantigo:"required,cidrv6"`
}
```

**Valid examples:** "2001:db8::/32", "fe80::/10", "::1/128"
**Invalid examples:** "192.168.0.0/24" (IPv4), "2001:db8:::/32" (invalid), "2001:db8::/129" (prefix > 128)

### `mac`

Validates that a string is a **valid MAC address**.

```go
type NetworkInterface struct {
    // MAC address for device identification
    HardwareAddress string `json:"mac_address" pedantigo:"required,mac"`
}
```

**Valid examples:** "00:1a:2b:3c:4d:5e", "00-1A-2B-3C-4D-5E", "001a2b3c4d5e"
**Invalid examples:** "00:1a:2b:3c:4d" (incomplete), "gg:1a:2b:3c:4d:5e" (invalid hex), "00:1a:2b:3c:4d:5e:ff" (too long)

### `hostname`

Validates that a string is a **valid hostname** (flexible format).

```go
type ServerInfo struct {
    // Hostname for the server
    Hostname string `json:"hostname" pedantigo:"required,hostname"`
}
```

**Valid examples:** "localhost", "my-server", "server123", "example.local"
**Invalid examples:** "-invalid", "invalid-", "invalid..com" (consecutive dots)

### `hostname_rfc1123`

Validates that a string is a **valid RFC 1123 hostname** (stricter standard).

```go
type DNSRecord struct {
    // RFC 1123 compliant hostname
    Name string `json:"name" pedantigo:"required,hostname_rfc1123"`
}
```

**Valid examples:** "localhost", "my-server", "my-server.com", "sub.domain.example.com"
**Invalid examples:** "invalid_name" (underscores not allowed), "UPPERCASE" (see hostname rules), "-start" (starts with hyphen)

### `fqdn`

Validates that a string is a **fully qualified domain name**.

```go
type DomainConfig struct {
    // Must be a complete domain name
    Domain string `json:"domain" pedantigo:"required,fqdn"`
}
```

**Valid examples:** "example.com", "sub.example.org", "very.long.domain.name.example.co.uk"
**Invalid examples:** "localhost" (not qualified), ".example.com" (starts with dot), "example." (ends with dot)

### `port`

Validates that a string or integer is a **valid port number** (1-65535).

```go
type ServiceEndpoint struct {
    // Port number must be in valid range
    Port int `json:"port" pedantigo:"required,port"`
    // Also works with strings
    PortString string `json:"port_string" pedantigo:"required,port"`
}
```

**Valid examples:** "80", "443", "8080", "65535"
**Invalid examples:** "0" (port 0 invalid), "65536" (exceeds max), "-1" (negative), "99999" (out of range)

### `tcp_addr`

Validates that a string is a **valid TCP address** in `host:port` format.

```go
type TCPConnection struct {
    // TCP address for server connection
    Address string `json:"address" pedantigo:"required,tcp_addr"`
}
```

**Valid examples:** "localhost:8080", "192.168.1.1:443", "example.com:80"
**Invalid examples:** "localhost" (missing port), "example.com:99999" (invalid port), "example.com:" (empty port)

### `udp_addr`

Validates that a string is a **valid UDP address** in `host:port` format.

```go
type UDPConnection struct {
    // UDP address for datagram communication
    Address string `json:"address" pedantigo:"required,udp_addr"`
}
```

**Valid examples:** "localhost:5353", "192.168.1.1:53", "dns.example.com:53"
**Invalid examples:** "localhost" (missing port), "example.com:0" (port 0), ":::1:8080" (malformed)

### `tcp4_addr`

Validates that a string is a **valid IPv4 TCP address**.

```go
type IPv4TCPEndpoint struct {
    // IPv4 TCP endpoint
    Endpoint string `json:"endpoint" pedantigo:"required,tcp4_addr"`
}
```

**Valid examples:** "192.168.1.1:8080", "10.0.0.1:443"
**Invalid examples:** "2001:db8::1:8080" (IPv6), "localhost:8080" (hostname), "192.168.1.1" (missing port)

## Financial & Cryptographic Formats

### `credit_card`

Validates that a string is a **valid credit card number** using Luhn algorithm.

```go
type PaymentInfo struct {
    // Credit card number must pass Luhn validation
    CardNumber string `json:"card_number" pedantigo:"required,credit_card"`
}
```

**Valid examples:** "4532015112830366" (Visa-like), "5425233010103442" (Mastercard-like)
**Invalid examples:** "1234567890123456" (fails Luhn), "123456789" (too short)

**Note:** This validates format and Luhn checksum only. Use additional validation for PCI compliance and sensitive handling.

### `luhn_checksum`

Validates that a string **passes the Luhn algorithm** checksum.

```go
type AccountVerification struct {
    // Account number must pass Luhn validation
    AccountNumber string `json:"account_number" pedantigo:"required,luhn_checksum"`
}
```

**Valid examples:** "4532015112830366", "79927398713"
**Invalid examples:** "4532015112830367" (checksum digit wrong), "123456"

### `btc_addr`

Validates that a string is a **valid Bitcoin address** (P2PKH or P2SH format).

```go
type CryptoWallet struct {
    // Bitcoin address for receiving payments
    BTCAddress string `json:"btc_address" pedantigo:"required,btc_addr"`
}
```

**Valid examples:** "1A1z7agoat7SFfuzcVzANNVFTesXVbWgi5", "3J98t1WpEZ73CNmYviecrnyiWrnqRhWNLy"
**Invalid examples:** "0x742d35Cc6634C0532925a3b844Bc9e7595f1" (Ethereum), "invalid", "1A1z7agoat7SFfuzcVzANNVFTesXVbWg" (invalid checksum)

### `btc_addr_bech32`

Validates that a string is a **valid Bech32-encoded Bitcoin address** (SegWit format).

```go
type ModernWallet struct {
    // SegWit Bitcoin address
    Address string `json:"address" pedantigo:"required,btc_addr_bech32"`
}
```

**Valid examples:** "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4", "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"
**Invalid examples:** "1A1z7agoat7SFfuzcVzANNVFTesXVbWgi5" (P2PKH not Bech32), "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7" (invalid checksum)

### `eth_addr`

Validates that a string is a **valid Ethereum address** (0x prefix with 40 hex characters).

```go
type EthereumWallet struct {
    // Ethereum address for contract interactions
    Address string `json:"address" pedantigo:"required,eth_addr"`
}
```

**Valid examples:** "0x742d35Cc6634C0532925a3b844Bc9e7595f1A02", "0x0000000000000000000000000000000000000000"
**Invalid examples:** "742d35Cc6634C0532925a3b844Bc9e7595f1A02" (missing 0x), "0xZZZZ35Cc6634C0532925a3b844Bc9e7595f1A02" (invalid hex)

## Identity & Document Formats

### `isbn` / `isbn10` / `isbn13`

Validates that a string is a **valid ISBN** (International Standard Book Number).

```go
type Book struct {
    // Accept either ISBN-10 or ISBN-13
    ISBN string `json:"isbn" pedantigo:"required,isbn"`
    // Specifically ISBN-13
    ISBN13 string `json:"isbn_13" pedantigo:"required,isbn13"`
    // Specifically ISBN-10
    ISBN10 string `json:"isbn_10,omitempty" pedantigo:"isbn10"`
}
```

**Valid examples:**
- ISBN-10: "0-306-40615-2", "030640615X"
- ISBN-13: "978-0-306-40615-7", "9780306406157"

**Invalid examples:** "0-306-40615-1" (wrong check digit), "12345" (too short)

### `issn`

Validates that a string is a **valid ISSN** (International Standard Serial Number).

```go
type Magazine struct {
    // ISSN for identifying the journal
    ISSN string `json:"issn" pedantigo:"required,issn"`
}
```

**Valid examples:** "0028-0836" (Nature), "0000-0019" (valid format)
**Invalid examples:** "0028-083" (incomplete), "0028-0837" (invalid check digit)

### `ssn`

Validates that a string is a **valid US Social Security Number** format.

```go
type EmployeeInfo struct {
    // US Social Security Number
    SSN string `json:"ssn" pedantigo:"required,ssn"`
}
```

**Valid examples:** "123-45-6789" (format check only), "111223333"
**Invalid examples:** "000-00-0000" (invalid), "12-345-6789" (wrong format), "1234567890" (too long)

### `ein`

Validates that a string is a **valid US Employer Identification Number** format.

```go
type CompanyInfo struct {
    // US EIN for business identification
    EIN string `json:"ein" pedantigo:"required,ein"`
}
```

**Valid examples:** "12-3456789", "123456789"
**Invalid examples:** "00-0000000" (invalid), "123456" (too short)

### `e164`

Validates that a string is a **valid E.164 phone number format** (+country code and digits).

```go
type ContactInfo struct {
    // International phone number in E.164 format
    Phone string `json:"phone" pedantigo:"required,e164"`
}
```

**Valid examples:** "+1234567890", "+442071838750", "+33123456789"
**Invalid examples:** "1234567890" (missing +), "+1" (incomplete), "+12345 67890" (spaces)

## Encoding Formats

### `jwt`

Validates that a string is a **valid JWT** (JSON Web Token) structure.

```go
type AuthToken struct {
    // JWT bearer token
    Token string `json:"token" pedantigo:"required,jwt"`
}
```

**Valid examples:** "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
**Invalid examples:** "not.a.token", "eyJhbGciOiJIUzI1NiJ9.invalid", "invalid-token"

### `json`

Validates that a string is **valid JSON** format.

```go
type ConfigData struct {
    // String containing valid JSON
    MetadataJSON string `json:"metadata" pedantigo:"required,json"`
}
```

**Valid examples:** `{"name": "John"}`, `[1, 2, 3]`, `true`, `null`
**Invalid examples:** `{name: John}` (unquoted keys), `{invalid json`, `undefined`

### `base64` / `base64url` / `base64rawurl`

Validates that a string is **valid Base64 encoding**.

```go
type EncodedData struct {
    // Standard Base64 (may have padding)
    Data string `json:"data" pedantigo:"required,base64"`
    // URL-safe Base64 (may have padding)
    URLSafe string `json:"url_safe" pedantigo:"required,base64url"`
    // URL-safe Base64 without padding
    RawURL string `json:"raw_url" pedantigo:"required,base64rawurl"`
}
```

**Valid examples:**
- Base64: "SGVsbG8gV29ybGQ=", "SGVsbG8gV29ybGQ"
- Base64url: "SGVsbG8tV29ybGQ=", "SGVsbG8tV29ybGQ"
- Base64rawurl: "SGVsbG8tV29ybGQ"

**Invalid examples:** "SGVsbG8gV29ybGQ==" (incorrect padding), "!@#$%^&*()" (invalid characters)

## Hash Formats

### `md5` / `md4` / `sha256` / `sha384` / `sha512`

Validates that a string is a **valid hash** in the specified algorithm format.

```go
type FileIntegrity struct {
    // MD5 hash (128-bit, 32 hex chars)
    MD5 string `json:"md5" pedantigo:"required,md5"`
    // SHA256 hash (256-bit, 64 hex chars)
    SHA256 string `json:"sha256" pedantigo:"required,sha256"`
    // SHA512 hash (512-bit, 128 hex chars)
    SHA512 string `json:"sha512" pedantigo:"required,sha512"`
}
```

**Valid examples:**
- MD5: "5d41402abc4b2a76b9719d911017c592"
- SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
- SHA512: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"

**Invalid examples:** "5d41402abc4b2a76b97" (too short), "ZZZZ1402abc4b2a76b9719d911017c592" (invalid hex)

### `mongodb`

Validates that a string is a **valid MongoDB ObjectID** (24 hex characters).

```go
type MongoDocument struct {
    // MongoDB document identifier
    ID string `json:"id" pedantigo:"required,mongodb"`
}
```

**Valid examples:** "507f1f77bcf86cd799439011", "507f191e810c19729de860ea"
**Invalid examples:** "507f1f77bcf86cd799439" (too short), "507f1f77bcf86cd799439g11" (invalid hex)

## ISO Standard Formats

### `iso3166_alpha2` / `iso3166_alpha3`

Validates that a string is a **valid ISO 3166-1 country code**.

```go
type Address struct {
    // ISO 3166-1 alpha-2 code (US, GB, FR, etc.)
    CountryCode2 string `json:"country_code_2" pedantigo:"required,iso3166_alpha2"`
    // ISO 3166-1 alpha-3 code (USA, GBR, FRA, etc.)
    CountryCode3 string `json:"country_code_3" pedantigo:"required,iso3166_alpha3"`
}
```

**Valid examples:**
- Alpha-2: "US", "GB", "FR", "JP", "IN", "AU"
- Alpha-3: "USA", "GBR", "FRA", "JPN", "IND", "AUS"

**Invalid examples:** "XX", "USAA", "U.S.", "United States"

### `iso3166_numeric`

Validates that a string is a **valid ISO 3166-1 numeric country code**.

```go
type CountryNumeric struct {
    // ISO 3166-1 numeric code (840 for USA, 826 for GB, etc.)
    Code string `json:"code" pedantigo:"required,iso3166_numeric"`
}
```

**Valid examples:** "840" (USA), "826" (UK), "392" (Japan), "356" (India)
**Invalid examples:** "999" (non-existent), "99" (too short), "USAA" (not numeric)

### `iso3166_2`

Validates that a string is a **valid ISO 3166-2 subdivision code** (country-subdivision).

```go
type SubdivisionInfo struct {
    // ISO 3166-2 code like US-CA (California), GB-ENG (England)
    Code string `json:"subdivision" pedantigo:"required,iso3166_2"`
}
```

**Valid examples:** "US-CA", "GB-ENG", "FR-75", "AU-NSW", "IN-MH"
**Invalid examples:** "CA" (missing country prefix), "USA-CA" (wrong country format), "US-ZZ" (invalid subdivision)

### `iso3166_alpha2_eu` / `iso3166_alpha3_eu`

Validates that a string is a **valid ISO 3166-1 code for EU member states only**.

```go
type EUAddress struct {
    // Must be an EU country (alpha-2)
    CountryCode string `json:"country_code" pedantigo:"required,iso3166_alpha2_eu"`
}
```

**Valid examples:** "DE", "FR", "IT", "ES", "PL" (any EU member)
**Invalid examples:** "US" (not EU), "GB" (not EU), "CH" (not EU), "XX" (invalid)

**Note:** Constraints recognize current EU member states at the time of library release.

### `iso4217` / `iso4217_numeric`

Validates that a string is a **valid ISO 4217 currency code**.

```go
type PriceInfo struct {
    // ISO 4217 3-letter currency code
    Currency string `json:"currency" pedantigo:"required,iso4217"`
    // ISO 4217 numeric code
    CurrencyNumeric string `json:"currency_numeric" pedantigo:"required,iso4217_numeric"`
}
```

**Valid examples:**
- Alpha: "USD", "EUR", "GBP", "JPY", "INR", "AUD"
- Numeric: "840" (USD), "978" (EUR), "826" (GBP), "392" (JPY)

**Invalid examples:** "USDA", "US", "999" (invalid numeric), "ZZZ"

### `postcode` / `postcode_iso3166_alpha2`

Validates that a string is a **valid postal code** for a specific country. Both `postcode` and `postcode_iso3166_alpha2` work identically (the latter is an alias for better clarity that the parameter expects an ISO 3166-1 alpha-2 country code).

```go
type InternationalAddress struct {
    // US zipcode (5 or 9 digits)
    ZipUS string `json:"zip" pedantigo:"required,postcode=US"`
    // UK postcode format (using alias)
    PostcodeUK string `json:"postcode" pedantigo:"required,postcode_iso3166_alpha2=GB"`
    // Canadian postal code format
    PostcodeCA string `json:"postal_code" pedantigo:"required,postcode=CA"`
}
```

**Valid examples:**
- US: "12345", "12345-6789"
- GB: "SW1A 1AA", "B33 8TH"
- CA: "K1A 0B1", "V6B 4X8"

**Invalid examples:** "ABCDE" (US invalid), "invalid" (all invalid), "12345-67890" (US too long)

### `bcp47`

Validates that a string is a **valid BCP 47 language tag**.

```go
type LocalizationSettings struct {
    // BCP 47 language tag (en, fr-CA, zh-Hans-CN, etc.)
    Language string `json:"language" pedantigo:"required,bcp47"`
}
```

**Valid examples:** "en", "fr", "en-US", "zh-Hans-CN", "pt-BR", "de-CH-1996"
**Invalid examples:** "English" (not code), "en_US" (underscore not allowed), "en-USAA" (malformed)

## Geographic Formats

### `latitude` / `longitude`

Validates that a string or number is a **valid geographic coordinate**.

```go
type GeoLocation struct {
    // Latitude: -90 to 90 degrees
    Latitude float64 `json:"latitude" pedantigo:"required,latitude"`
    // Longitude: -180 to 180 degrees
    Longitude float64 `json:"longitude" pedantigo:"required,longitude"`
}
```

**Valid examples:**
- Latitude: "40.7128", "-33.8688", "0", "90"
- Longitude: "-74.0060", "151.2093", "-180", "180"

**Invalid examples:** "91" (latitude out of range), "181" (longitude out of range), "abc"

## Color Formats

### `hexcolor` / `rgb` / `rgba` / `hsl` / `hsla`

Validates that a string is a **valid color format**.

```go
type ThemeConfig struct {
    // Hexadecimal color
    PrimaryColor string `json:"primary" pedantigo:"required,hexcolor"`
    // RGB color
    SecondaryColor string `json:"secondary" pedantigo:"required,rgb"`
    // RGBA color with transparency
    TransparentColor string `json:"transparent" pedantigo:"rgba"`
}
```

**Valid examples:**
- Hex: "#FF5733", "#fff", "FF5733"
- RGB: "rgb(255, 87, 51)", "rgb(255,87,51)"
- RGBA: "rgba(255, 87, 51, 0.8)"
- HSL: "hsl(14, 100%, 60%)"
- HSLA: "hsla(14, 100%, 60%, 0.8)"

**Invalid examples:** "#GGGGGG" (invalid hex), "rgb(256, 0, 0)" (out of range), "rgb(255, 87)" (incomplete)

### `iscolor`

Validates that a string is a **valid color in any CSS format**. This is a convenience alias that accepts hexcolor, rgb, rgba, hsl, or hsla formats.

```go
type ThemeConfig struct {
    // Accept any valid CSS color format
    PrimaryColor string `json:"primary" pedantigo:"required,iscolor"`
    AccentColor  string `json:"accent" pedantigo:"iscolor"`
}
```

**Valid examples:**
- `"#FF5733"` - Hex color
- `"#fff"` - Short hex
- `"rgb(255, 87, 51)"` - RGB
- `"rgba(255, 87, 51, 0.8)"` - RGBA with alpha
- `"hsl(14, 100%, 60%)"` - HSL
- `"hsla(14, 100%, 60%, 0.8)"` - HSLA with alpha

**Invalid examples:** "red" (named colors not supported), "#GGGGGG" (invalid hex)

## Miscellaneous Formats

### `html`

Validates that a string **contains HTML markup**.

```go
type BlogPost struct {
    // Post content must contain HTML
    Content string `json:"content" pedantigo:"required,html"`
}
```

**Valid examples:** `<p>Hello</p>`, `<div>Content</div>`, `<img src="test.jpg" />`
**Invalid examples:** `Plain text`, `Hello <world`, `<>` (invalid tag)

### `cron`

Validates that a string is a **valid cron expression**.

```go
type ScheduledTask struct {
    // Cron schedule for task execution
    Schedule string `json:"schedule" pedantigo:"required,cron"`
}
```

**Valid examples:** "0 0 * * *" (daily), "0 */4 * * *" (every 4 hours), "0 0 1 1 *" (Jan 1st)
**Invalid examples:** "* * * * * *" (6 fields), "60 0 * * *" (invalid minute)

### `semver`

Validates that a string is a **valid Semantic Version** (major.minor.patch).

```go
type PackageInfo struct {
    // Semantic version for package
    Version string `json:"version" pedantigo:"required,semver"`
}
```

**Valid examples:** "1.0.0", "2.1.3", "0.0.1", "1.2.3-beta", "1.2.3+build.123"
**Invalid examples:** "1.0" (missing patch), "v1.0.0" (v prefix), "1.0.0.0" (too many parts)

### `datetime`

Validates that a string matches a **Go time layout format**.

```go
type Event struct {
    // ISO 8601 date format
    Date string `json:"date" pedantigo:"required,datetime=2006-01-02"`

    // Full datetime with seconds
    Timestamp string `json:"timestamp" pedantigo:"required,datetime=2006-01-02 15:04:05"`

    // RFC 3339 format
    CreatedAt string `json:"created_at" pedantigo:"datetime=2006-01-02T15:04:05Z07:00"`
}
```

Uses Go's [time package layouts](https://pkg.go.dev/time#pkg-constants). The reference time is `Mon Jan 2 15:04:05 MST 2006`.

**Common layouts:**
- `2006-01-02` — Date only (YYYY-MM-DD)
- `2006-01-02 15:04:05` — Datetime with space
- `2006-01-02T15:04:05Z07:00` — RFC 3339
- `15:04:05` — Time only
- `01/02/2006` — US date format

**Valid examples:** "2024-12-21" (for `2006-01-02`), "2024-12-21 14:30:00" (for `2006-01-02 15:04:05`)
**Invalid examples:** "21-12-2024" (wrong order), "2024/12/21" (wrong separator for layout)

## Filesystem Formats

### `filepath` / `dirpath`

Validates that a string is a **valid file or directory path**.

```go
type FileConfig struct {
    // Path to a file
    FilePath string `json:"file" pedantigo:"required,filepath"`
    // Path to a directory
    DirPath string `json:"directory" pedantigo:"required,dirpath"`
}
```

**Valid examples:**
- File: "/home/user/document.txt", "C:\\Users\\file.txt", "relative/path/file.go"
- Dir: "/home/user", "C:\\Users", "relative/path"

**Invalid examples:** "" (empty), contains null bytes

### `file` / `dir`

Validates that a path **exists** and is a file or directory respectively.

```go
type ExistingFileConfig struct {
    // File must exist on filesystem
    ExistingFile string `json:"existing_file" pedantigo:"required,file"`
    // Directory must exist on filesystem
    ExistingDir string `json:"existing_dir" pedantigo:"required,dir"`
}
```

**Valid examples:** "/etc/passwd" (if exists), "/home/user" (if exists)
**Invalid examples:** "/nonexistent/path", "/etc/passwd" (when validating as dir)

## Complete Format Example

Here's a comprehensive example using multiple format constraints:

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type UserAccount struct {
    // UUID identifier
    ID string `json:"id" pedantigo:"required,uuid"`

    // Email address
    Email string `json:"email" pedantigo:"required,email"`

    // Personal website
    Website string `json:"website,omitempty" pedantigo:"url"`

    // International phone
    Phone string `json:"phone" pedantigo:"required,e164"`

    // Country location
    Country string `json:"country" pedantigo:"required,iso3166_alpha2"`

    // Preferred language
    Language string `json:"language" pedantigo:"required,bcp47"`

    // Birth coordinates
    BirthLatitude float64 `json:"birth_latitude,omitempty" pedantigo:"latitude"`
    BirthLongitude float64 `json:"birth_longitude,omitempty" pedantigo:"longitude"`

    // Authentication
    JWTToken string `json:"jwt" pedantigo:"required,jwt"`
}

type CompanyInfo struct {
    // Company identifier
    ID string `json:"id" pedantigo:"required,uuid"`

    // Official website
    Website string `json:"website" pedantigo:"required,url"`

    // Tax identifier
    EIN string `json:"ein" pedantigo:"required,ein"`

    // Company country (EU only)
    Country string `json:"country" pedantigo:"required,iso3166_alpha2_eu"`

    // Operating currency
    Currency string `json:"currency" pedantigo:"required,iso4217"`

    // Server IP address
    ServerIP string `json:"server_ip" pedantigo:"required,ipv4"`

    // Network subnet
    Subnet string `json:"subnet" pedantigo:"required,cidrv4"`

    // Network interface MAC
    MacAddress string `json:"mac" pedantigo:"required,mac"`

    // DNS hostname
    Hostname string `json:"hostname" pedantigo:"required,hostname_rfc1123"`

    // Primary contact port
    ContactPort int `json:"contact_port" pedantigo:"required,port"`

    // Brand color scheme
    PrimaryColor string `json:"primary_color" pedantigo:"required,hexcolor"`

    // Release version
    SoftwareVersion string `json:"version" pedantigo:"required,semver"`
}

func main() {
    // Valid user account
    userJSON := []byte(`{
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "john@example.com",
        "website": "https://example.com",
        "phone": "+12025551234",
        "country": "US",
        "language": "en-US",
        "birth_latitude": 40.7128,
        "birth_longitude": -74.0060,
        "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
    }`)

    user, err := pedantigo.Unmarshal[UserAccount](userJSON)
    if err != nil {
        fmt.Printf("User validation failed: %v\n", err)
        return
    }

    fmt.Printf("Valid user: %+v\n", user)

    // Valid company info
    companyJSON := []byte(`{
        "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
        "website": "https://company.eu",
        "ein": "12-3456789",
        "country": "DE",
        "currency": "EUR",
        "server_ip": "192.168.1.1",
        "subnet": "192.168.0.0/24",
        "mac": "00:1a:2b:3c:4d:5e",
        "hostname": "api.company.eu",
        "contact_port": 443,
        "primary_color": "#FF5733",
        "version": "1.2.3"
    }`)

    company, err := pedantigo.Unmarshal[CompanyInfo](companyJSON)
    if err != nil {
        fmt.Printf("Company validation failed: %v\n", err)
        return
    }

    fmt.Printf("Valid company: %+v\n", company)
}
```

## Quick Reference Table

| Constraint | Example | Description |
|------------|---------|-------------|
| `email` | `email` | Valid email address |
| `url` | `url` | Valid URL with scheme |
| `uuid` | `uuid` | Valid UUID identifier |
| `ulid` | `ulid` | Valid ULID identifier |
| `ip` | `ip` | IPv4 or IPv6 address |
| `ipv4` | `ipv4` | IPv4 address |
| `ipv6` | `ipv6` | IPv6 address |
| `cidr` | `cidr` | CIDR notation (v4 or v6) |
| `cidrv4` | `cidrv4` | IPv4 CIDR notation |
| `cidrv6` | `cidrv6` | IPv6 CIDR notation |
| `mac` | `mac` | MAC address |
| `hostname` | `hostname` | Hostname (flexible) |
| `hostname_rfc1123` | `hostname_rfc1123` | RFC 1123 hostname |
| `fqdn` | `fqdn` | Fully qualified domain name |
| `port` | `port` | Port number (1-65535) |
| `tcp_addr` | `tcp_addr` | TCP address (host:port) |
| `udp_addr` | `udp_addr` | UDP address (host:port) |
| `tcp4_addr` | `tcp4_addr` | IPv4 TCP address |
| `credit_card` | `credit_card` | Credit card (Luhn check) |
| `luhn_checksum` | `luhn_checksum` | Luhn algorithm valid |
| `btc_addr` | `btc_addr` | Bitcoin address |
| `btc_addr_bech32` | `btc_addr_bech32` | SegWit Bitcoin address |
| `eth_addr` | `eth_addr` | Ethereum address |
| `isbn` | `isbn` | ISBN-10 or ISBN-13 |
| `isbn10` | `isbn10` | ISBN-10 format |
| `isbn13` | `isbn13` | ISBN-13 format |
| `issn` | `issn` | ISSN identifier |
| `ssn` | `ssn` | US Social Security Number |
| `ein` | `ein` | US Employer ID |
| `e164` | `e164` | E.164 phone format |
| `jwt` | `jwt` | JSON Web Token |
| `json` | `json` | Valid JSON string |
| `base64` | `base64` | Base64 encoding |
| `base64url` | `base64url` | URL-safe Base64 |
| `base64rawurl` | `base64rawurl` | URL-safe Base64 (no padding) |
| `md5` | `md5` | MD5 hash format |
| `md4` | `md4` | MD4 hash format |
| `sha256` | `sha256` | SHA256 hash format |
| `sha384` | `sha384` | SHA384 hash format |
| `sha512` | `sha512` | SHA512 hash format |
| `mongodb` | `mongodb` | MongoDB ObjectID |
| `iso3166_alpha2` | `iso3166_alpha2` | ISO country code (2-letter) |
| `iso3166_alpha3` | `iso3166_alpha3` | ISO country code (3-letter) |
| `iso3166_numeric` | `iso3166_numeric` | ISO country code (numeric) |
| `iso3166_2` | `iso3166_2` | ISO subdivision code |
| `iso3166_alpha2_eu` | `iso3166_alpha2_eu` | ISO code (EU countries only) |
| `iso3166_alpha3_eu` | `iso3166_alpha3_eu` | ISO code (EU, 3-letter) |
| `iso4217` | `iso4217` | ISO currency code |
| `iso4217_numeric` | `iso4217_numeric` | ISO currency code (numeric) |
| `postcode=CC` | `postcode=US` | Postal code for country |
| `postcode_iso3166_alpha2=CC` | `postcode_iso3166_alpha2=GB` | Postal code (alias) |
| `bcp47` | `bcp47` | BCP 47 language tag |
| `latitude` | `latitude` | Valid latitude (-90 to 90) |
| `longitude` | `longitude` | Valid longitude (-180 to 180) |
| `hexcolor` | `hexcolor` | Hexadecimal color |
| `rgb` | `rgb` | RGB color format |
| `rgba` | `rgba` | RGBA color format |
| `hsl` | `hsl` | HSL color format |
| `hsla` | `hsla` | HSLA color format |
| `html` | `html` | Contains HTML markup |
| `cron` | `cron` | Cron expression |
| `semver` | `semver` | Semantic version |
| `datetime=LAYOUT` | `datetime=2006-01-02` | Go time layout format |
| `filepath` | `filepath` | Valid file path |
| `dirpath` | `dirpath` | Valid directory path |
| `file` | `file` | File exists on filesystem |
| `dir` | `dir` | Directory exists on filesystem |

:::tip
Format constraints use standard patterns and are highly optimized for performance. Most validate format only; some like `file` and `dir` check filesystem existence.
:::
