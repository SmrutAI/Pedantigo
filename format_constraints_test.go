package pedantigo

import (
	"testing"
)

// ==================================================
// url constraint tests
// ==================================================

func TestURL(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		usePointer bool
		expectErr  bool
		expectVal  string
		expectNil  bool
	}{
		{"Valid HTTPS", `{"website":"https://example.com"}`, false, false, "https://example.com", false},
		{"Valid HTTP", `{"website":"http://example.com"}`, false, false, "http://example.com", false},
		{"Invalid format", `{"website":"not a url"}`, false, true, "", false},
		{"No scheme", `{"website":"example.com"}`, false, true, "", false},
		{"FTP scheme", `{"website":"ftp://example.com"}`, false, true, "", false},
		{"Empty string", `{"website":""}`, false, false, "", false},
		{"Pointer invalid", `{"website":"not a url"}`, true, true, "", false},
		{"Pointer valid", `{"website":"https://example.com"}`, true, false, "https://example.com", false},
		{"Nil pointer", `{"website":null}`, true, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.usePointer {
				type Config struct {
					Website *string `json:"website" pedantigo:"url"`
				}
				validator := New[Config]()
				config, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					if err == nil {
						t.Fatal("expected validation error")
					}
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Website" && fieldErr.Message == "must be a valid URL (http or https)" {
							foundError = true
						}
					}
					if !foundError {
						t.Errorf("expected 'must be a valid URL (http or https)' error, got %v", ve.Errors)
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if tt.expectNil {
						if config.Website != nil {
							t.Errorf("expected nil Website pointer, got %v", config.Website)
						}
					} else if config.Website == nil || *config.Website != tt.expectVal {
						t.Errorf("expected website %q, got %v", tt.expectVal, config.Website)
					}
				}
			} else {
				type Config struct {
					Website string `json:"website" pedantigo:"url"`
				}
				validator := New[Config]()
				config, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					if err == nil {
						t.Fatal("expected validation error")
					}
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Website" && fieldErr.Message == "must be a valid URL (http or https)" {
							foundError = true
						}
					}
					if !foundError {
						t.Errorf("expected 'must be a valid URL (http or https)' error, got %v", ve.Errors)
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if config.Website != tt.expectVal {
						t.Errorf("expected website %q, got %q", tt.expectVal, config.Website)
					}
				}
			}
		})
	}
}

// ==================================================
// uuid constraint tests
// ==================================================

func TestUUID_Valid_V4(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()
	jsonData := []byte(`{"id":"550e8400-e29b-41d4-a716-446655440000"}`)

	entity, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid UUID v4, got %v", err)
	}

	if entity.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected id '550e8400-e29b-41d4-a716-446655440000', got %q", entity.ID)
	}
}

func TestUUID_Valid_V5(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()
	jsonData := []byte(`{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}`)

	entity, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid UUID v5, got %v", err)
	}

	if entity.ID != "886313e1-3b8a-5372-9b90-0c9aee199e5d" {
		t.Errorf("expected id '886313e1-3b8a-5372-9b90-0c9aee199e5d', got %q", entity.ID)
	}
}

func TestUUID_InvalidFormat(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()
	jsonData := []byte(`{"id":"not-a-uuid"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid UUID format")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "ID" && fieldErr.Message == "must be a valid UUID" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be a valid UUID' error, got %v", ve.Errors)
	}
}

func TestUUID_InvalidFormat_WrongDashes(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()
	jsonData := []byte(`{"id":"550e8400e29b41d4a716446655440000"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for UUID without dashes")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "ID" && fieldErr.Message == "must be a valid UUID" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be a valid UUID' error, got %v", ve.Errors)
	}
}

func TestUUID_EmptyString(t *testing.T) {
	type Entity struct {
		ID string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()
	jsonData := []byte(`{"id":""}`)

	entity, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for empty UUID (validation skips empty), got %v", err)
	}

	if entity.ID != "" {
		t.Errorf("expected empty id, got %q", entity.ID)
	}
}

func TestUUID_WithPointer(t *testing.T) {
	type Entity struct {
		ID *string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()

	// Test invalid UUID
	jsonData := []byte(`{"id":"not-a-uuid"}`)
	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid UUID with pointer")
	}

	// Test valid UUID
	jsonData = []byte(`{"id":"550e8400-e29b-41d4-a716-446655440000"}`)
	entity, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid UUID with pointer, got %v", err)
	}

	if entity.ID == nil || *entity.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected id '550e8400-e29b-41d4-a716-446655440000', got %v", entity.ID)
	}
}

func TestUUID_NilPointer(t *testing.T) {
	type Entity struct {
		ID *string `json:"id" pedantigo:"uuid"`
	}

	validator := New[Entity]()
	jsonData := []byte(`{"id":null}`)

	entity, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", err)
	}

	if entity.ID != nil {
		t.Errorf("expected nil ID pointer, got %v", entity.ID)
	}
}

// ==================================================
// regex constraint tests
// ==================================================

func TestRegex_ValidMatch(t *testing.T) {
	type Code struct {
		Value string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
	}

	validator := New[Code]()
	jsonData := []byte(`{"value":"ABC"}`)

	code, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid regex match, got %v", err)
	}

	if code.Value != "ABC" {
		t.Errorf("expected value 'ABC', got %q", code.Value)
	}
}

func TestRegex_InvalidMatch(t *testing.T) {
	type Code struct {
		Value string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
	}

	validator := New[Code]()
	jsonData := []byte(`{"value":"abc"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid regex match")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Value" && fieldErr.Message == "must match pattern '^[A-Z]{3}$'" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must match pattern' error, got %v", ve.Errors)
	}
}

func TestRegex_WrongLength(t *testing.T) {
	type Code struct {
		Value string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
	}

	validator := New[Code]()
	jsonData := []byte(`{"value":"ABCD"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for wrong length")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Value" && fieldErr.Message == "must match pattern '^[A-Z]{3}$'" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must match pattern' error, got %v", ve.Errors)
	}
}

func TestRegex_DigitsPattern(t *testing.T) {
	type Code struct {
		Value string `json:"value" pedantigo:"regexp=^\\d{4}$"`
	}

	validator := New[Code]()

	// Test valid digits
	jsonData := []byte(`{"value":"1234"}`)
	code, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid digits, got %v", err)
	}

	if code.Value != "1234" {
		t.Errorf("expected value '1234', got %q", code.Value)
	}

	// Test invalid non-digits
	jsonData = []byte(`{"value":"abcd"}`)
	_, err = validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for non-digits")
	}
}

func TestRegex_EmptyString(t *testing.T) {
	type Code struct {
		Value string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
	}

	validator := New[Code]()
	jsonData := []byte(`{"value":""}`)

	code, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for empty string (validation skips empty), got %v", err)
	}

	if code.Value != "" {
		t.Errorf("expected empty value, got %q", code.Value)
	}
}

func TestRegex_WithPointer(t *testing.T) {
	type Code struct {
		Value *string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
	}

	validator := New[Code]()

	// Test invalid match
	jsonData := []byte(`{"value":"abc"}`)
	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid regex match with pointer")
	}

	// Test valid match
	jsonData = []byte(`{"value":"ABC"}`)
	code, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid regex match with pointer, got %v", err)
	}

	if code.Value == nil || *code.Value != "ABC" {
		t.Errorf("expected value 'ABC', got %v", code.Value)
	}
}

func TestRegex_NilPointer(t *testing.T) {
	type Code struct {
		Value *string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
	}

	validator := New[Code]()
	jsonData := []byte(`{"value":null}`)

	code, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", err)
	}

	if code.Value != nil {
		t.Errorf("expected nil Value pointer, got %v", code.Value)
	}
}

// ==================================================
// ipv4 constraint tests
// ==================================================

func TestIPv4_Valid_Localhost(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"127.0.0.1"}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid IPv4 localhost, got %v", err)
	}

	if server.IP != "127.0.0.1" {
		t.Errorf("expected ip '127.0.0.1', got %q", server.IP)
	}
}

func TestIPv4_Valid_PrivateNetwork(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"192.168.1.1"}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid private IPv4, got %v", err)
	}

	if server.IP != "192.168.1.1" {
		t.Errorf("expected ip '192.168.1.1', got %q", server.IP)
	}
}

func TestIPv4_InvalidFormat(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"not-an-ip"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid IPv4 format")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv4 address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be a valid IPv4 address' error, got %v", ve.Errors)
	}
}

func TestIPv4_InvalidIPv6(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"2001:0db8:85a3::8a2e:0370:7334"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for IPv6 (not IPv4)")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv4 address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be a valid IPv4 address' error, got %v", ve.Errors)
	}
}

func TestIPv4_EmptyString(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":""}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for empty IP (validation skips empty), got %v", err)
	}

	if server.IP != "" {
		t.Errorf("expected empty ip, got %q", server.IP)
	}
}

func TestIPv4_WithPointer(t *testing.T) {
	type Server struct {
		IP *string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()

	// Test invalid IP
	jsonData := []byte(`{"ip":"not-an-ip"}`)
	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid IPv4 with pointer")
	}

	// Test valid IP
	jsonData = []byte(`{"ip":"10.0.0.1"}`)
	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid IPv4 with pointer, got %v", err)
	}

	if server.IP == nil || *server.IP != "10.0.0.1" {
		t.Errorf("expected ip '10.0.0.1', got %v", server.IP)
	}
}

func TestIPv4_NilPointer(t *testing.T) {
	type Server struct {
		IP *string `json:"ip" pedantigo:"ipv4"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":null}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", err)
	}

	if server.IP != nil {
		t.Errorf("expected nil IP pointer, got %v", server.IP)
	}
}

// ==================================================
// ipv6 constraint tests
// ==================================================

func TestIPv6_Valid_Localhost(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"::1"}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid IPv6 localhost, got %v", err)
	}

	if server.IP != "::1" {
		t.Errorf("expected ip '::1', got %q", server.IP)
	}
}

func TestIPv6_Valid_FullFormat(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid IPv6 full format, got %v", err)
	}

	if server.IP != "2001:0db8:85a3:0000:0000:8a2e:0370:7334" {
		t.Errorf("expected ip '2001:0db8:85a3:0000:0000:8a2e:0370:7334', got %q", server.IP)
	}
}

func TestIPv6_Valid_Compressed(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"2001:db8:85a3::8a2e:370:7334"}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid IPv6 compressed format, got %v", err)
	}

	if server.IP != "2001:db8:85a3::8a2e:370:7334" {
		t.Errorf("expected ip '2001:db8:85a3::8a2e:370:7334', got %q", server.IP)
	}
}

func TestIPv6_InvalidFormat(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"not-an-ip"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid IPv6 format")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv6 address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be a valid IPv6 address' error, got %v", ve.Errors)
	}
}

func TestIPv6_InvalidIPv4(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":"192.168.1.1"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for IPv4 (not IPv6)")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv6 address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be a valid IPv6 address' error, got %v", ve.Errors)
	}
}

func TestIPv6_EmptyString(t *testing.T) {
	type Server struct {
		IP string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":""}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for empty IP (validation skips empty), got %v", err)
	}

	if server.IP != "" {
		t.Errorf("expected empty ip, got %q", server.IP)
	}
}

func TestIPv6_WithPointer(t *testing.T) {
	type Server struct {
		IP *string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()

	// Test invalid IP
	jsonData := []byte(`{"ip":"not-an-ip"}`)
	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid IPv6 with pointer")
	}

	// Test valid IP
	jsonData = []byte(`{"ip":"fe80::1"}`)
	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid IPv6 with pointer, got %v", err)
	}

	if server.IP == nil || *server.IP != "fe80::1" {
		t.Errorf("expected ip 'fe80::1', got %v", server.IP)
	}
}

func TestIPv6_NilPointer(t *testing.T) {
	type Server struct {
		IP *string `json:"ip" pedantigo:"ipv6"`
	}

	validator := New[Server]()
	jsonData := []byte(`{"ip":null}`)

	server, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", err)
	}

	if server.IP != nil {
		t.Errorf("expected nil IP pointer, got %v", server.IP)
	}
}
