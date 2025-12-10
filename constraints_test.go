package pedantigo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Format Constraints ====================

// ==================================================
// url constraint tests
// ==================================================

// TestURL tests URL validation
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
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Website" && fieldErr.Message == "must be a valid URL (http or https)" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid URL (http or https)' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, config.Website)
					} else {
						require.NotNil(t, config.Website)
						assert.Equal(t, tt.expectVal, *config.Website)
					}
				}
			} else {
				type Config struct {
					Website string `json:"website" pedantigo:"url"`
				}
				validator := New[Config]()
				config, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Website" && fieldErr.Message == "must be a valid URL (http or https)" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid URL (http or https)' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal, config.Website)
				}
			}
		})
	}
}

// ==================================================
// uuid constraint tests
// ==================================================

// TestUUID tests UUID validation
func TestUUID(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		usePointer bool
		expectErr  bool
		expectVal  string
		expectNil  bool
	}{
		{"Valid V4", `{"id":"550e8400-e29b-41d4-a716-446655440000"}`, false, false, "550e8400-e29b-41d4-a716-446655440000", false},
		{"Valid V5", `{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}`, false, false, "886313e1-3b8a-5372-9b90-0c9aee199e5d", false},
		{"Invalid format", `{"id":"not-a-uuid"}`, false, true, "", false},
		{"Wrong dashes", `{"id":"550e8400e29b41d4a716446655440000"}`, false, true, "", false},
		{"Empty string", `{"id":""}`, false, false, "", false},
		{"Pointer invalid", `{"id":"not-a-uuid"}`, true, true, "", false},
		{"Pointer valid", `{"id":"550e8400-e29b-41d4-a716-446655440000"}`, true, false, "550e8400-e29b-41d4-a716-446655440000", false},
		{"Nil pointer", `{"id":null}`, true, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.usePointer {
				type Entity struct {
					ID *string `json:"id" pedantigo:"uuid"`
				}
				validator := New[Entity]()
				entity, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "ID" && fieldErr.Message == "must be a valid UUID" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid UUID' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, entity.ID)
					} else {
						require.NotNil(t, entity.ID)
						assert.Equal(t, tt.expectVal, *entity.ID)
					}
				}
			} else {
				type Entity struct {
					ID string `json:"id" pedantigo:"uuid"`
				}
				validator := New[Entity]()
				entity, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "ID" && fieldErr.Message == "must be a valid UUID" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid UUID' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal, entity.ID)
				}
			}
		})
	}
}

// ==================================================
// regex constraint tests
// ==================================================

// TestRegex_UppercasePattern tests Regex uppercasepattern
func TestRegex_UppercasePattern(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		usePointer bool
		expectErr  bool
		expectVal  string
		expectNil  bool
	}{
		{"Valid match", `{"value":"ABC"}`, false, false, "ABC", false},
		{"Invalid match", `{"value":"abc"}`, false, true, "", false},
		{"Wrong length", `{"value":"ABCD"}`, false, true, "", false},
		{"Empty string", `{"value":""}`, false, false, "", false},
		{"Pointer invalid", `{"value":"abc"}`, true, true, "", false},
		{"Pointer valid", `{"value":"ABC"}`, true, false, "ABC", false},
		{"Nil pointer", `{"value":null}`, true, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.usePointer {
				type Code struct {
					Value *string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
				}
				validator := New[Code]()
				code, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Value" && fieldErr.Message == "must match pattern '^[A-Z]{3}$'" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must match pattern' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, code.Value)
					} else {
						require.NotNil(t, code.Value)
						assert.Equal(t, tt.expectVal, *code.Value)
					}
				}
			} else {
				type Code struct {
					Value string `json:"value" pedantigo:"regexp=^[A-Z]{3}$"`
				}
				validator := New[Code]()
				code, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Value" && fieldErr.Message == "must match pattern '^[A-Z]{3}$'" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must match pattern' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal, code.Value)
				}
			}
		})
	}
}

func TestRegex_DigitsPattern(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		expectErr bool
		expectVal string
	}{
		{"Valid digits", `{"value":"1234"}`, false, "1234"},
		{"Invalid non-digits", `{"value":"abcd"}`, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Code struct {
				Value string `json:"value" pedantigo:"regexp=^\\d{4}$"`
			}
			validator := New[Code]()
			code, err := validator.Unmarshal([]byte(tt.json))

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectVal, code.Value)
			}
		})
	}
}

// ==================================================
// ipv4 constraint tests
// ==================================================

// TestIPv4 tests IPv4 validation
func TestIPv4(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		usePointer bool
		expectErr  bool
		expectVal  string
		expectNil  bool
	}{
		{"Valid localhost", `{"ip":"127.0.0.1"}`, false, false, "127.0.0.1", false},
		{"Valid private", `{"ip":"192.168.1.1"}`, false, false, "192.168.1.1", false},
		{"Invalid format", `{"ip":"not-an-ip"}`, false, true, "", false},
		{"Invalid IPv6", `{"ip":"2001:0db8:85a3::8a2e:0370:7334"}`, false, true, "", false},
		{"Empty string", `{"ip":""}`, false, false, "", false},
		{"Pointer invalid", `{"ip":"not-an-ip"}`, true, true, "", false},
		{"Pointer valid", `{"ip":"10.0.0.1"}`, true, false, "10.0.0.1", false},
		{"Nil pointer", `{"ip":null}`, true, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.usePointer {
				type Server struct {
					IP *string `json:"ip" pedantigo:"ipv4"`
				}
				validator := New[Server]()
				server, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv4 address" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid IPv4 address' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, server.IP)
					} else {
						require.NotNil(t, server.IP)
						assert.Equal(t, tt.expectVal, *server.IP)
					}
				}
			} else {
				type Server struct {
					IP string `json:"ip" pedantigo:"ipv4"`
				}
				validator := New[Server]()
				server, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv4 address" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid IPv4 address' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal, server.IP)
				}
			}
		})
	}
}

// ==================================================
// ipv6 constraint tests
// ==================================================

// TestIPv6 tests IPv6 validation
func TestIPv6(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		usePointer bool
		expectErr  bool
		expectVal  string
		expectNil  bool
	}{
		{"Valid localhost", `{"ip":"::1"}`, false, false, "::1", false},
		{"Valid full", `{"ip":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`, false, false, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Valid compressed", `{"ip":"2001:db8:85a3::8a2e:370:7334"}`, false, false, "2001:db8:85a3::8a2e:370:7334", false},
		{"Invalid format", `{"ip":"not-an-ip"}`, false, true, "", false},
		{"Invalid IPv4", `{"ip":"192.168.1.1"}`, false, true, "", false},
		{"Empty string", `{"ip":""}`, false, false, "", false},
		{"Pointer invalid", `{"ip":"not-an-ip"}`, true, true, "", false},
		{"Pointer valid", `{"ip":"fe80::1"}`, true, false, "fe80::1", false},
		{"Nil pointer", `{"ip":null}`, true, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.usePointer {
				type Server struct {
					IP *string `json:"ip" pedantigo:"ipv6"`
				}
				validator := New[Server]()
				server, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv6 address" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid IPv6 address' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, server.IP)
					} else {
						require.NotNil(t, server.IP)
						assert.Equal(t, tt.expectVal, *server.IP)
					}
				}
			} else {
				type Server struct {
					IP string `json:"ip" pedantigo:"ipv6"`
				}
				validator := New[Server]()
				server, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "IP" && fieldErr.Message == "must be a valid IPv6 address" {
							foundError = true
						}
					}
					assert.True(t, foundError, "expected 'must be a valid IPv6 address' error, got %v", ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal, server.IP)
				}
			}
		})
	}
}

// ==================== String Constraints ====================

// ==================================================
// min_length constraint tests
// ==================================================

// TestMinLength tests MinLength validation
func TestMinLength(t *testing.T) {
	tests := []struct {
		name      string
		minVal    int
		fieldName string
		json      string
		usePtr    bool
		expectErr bool
		expectVal string
		expectNil bool
	}{
		{
			name:      "Valid above min",
			minVal:    3,
			fieldName: "Username",
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
		},
		{
			name:      "Exactly at min",
			minVal:    3,
			fieldName: "Username",
			json:      `{"username":"bob"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "bob",
			expectNil: false,
		},
		{
			name:      "Below min",
			minVal:    3,
			fieldName: "Username",
			json:      `{"username":"ab"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Empty string",
			minVal:    1,
			fieldName: "Username",
			json:      `{"username":""}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Pointer below min",
			minVal:    10,
			fieldName: "Bio",
			json:      `{"bio":"short"}`,
			usePtr:    true,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Pointer valid",
			minVal:    10,
			fieldName: "Bio",
			json:      `{"bio":"this is a longer bio"}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "this is a longer bio",
			expectNil: false,
		},
		{
			name:      "Nil pointer",
			minVal:    10,
			fieldName: "Bio",
			json:      `{"bio":null}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.usePtr {
				// Non-pointer test case
				// User represents the data structure
				type User struct {
					Username string `json:"username" pedantigo:"min=3"`
				}

				// For empty string test, use min=1
				if tt.minVal == 1 {
					type UserMin1 struct {
						Username string `json:"username" pedantigo:"min=1"`
					}
					validator := New[UserMin1]()
					_, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr {
						require.Error(t, err)
						ve, ok := err.(*ValidationError)
						require.True(t, ok, "expected *ValidationError, got %T", err)
						expectedMsg := "must be at least 1 characters"
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
								foundError = true
								break
							}
						}
						assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
					} else {
						assert.NoError(t, err)
					}
				} else {
					validator := New[User]()
					user, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr {
						require.Error(t, err)
						ve, ok := err.(*ValidationError)
						require.True(t, ok, "expected *ValidationError, got %T", err)
						expectedMsg := "must be at least 3 characters"
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
								foundError = true
								break
							}
						}
						assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tt.expectVal, user.Username)
					}
				}
			} else {
				// Pointer test case
				// User represents the data structure
				type User struct {
					Bio *string `json:"bio" pedantigo:"min=10"`
				}

				validator := New[User]()
				user, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					expectedMsg := "must be at least 10 characters"
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Bio" && fieldErr.Message == expectedMsg {
							foundError = true
							break
						}
					}
					assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, user.Bio)
					} else {
						require.NotNil(t, user.Bio)
						assert.Equal(t, tt.expectVal, *user.Bio)
					}
				}
			}
		})
	}
}

// ==================================================
// max_length constraint tests
// ==================================================

// TestMaxLength tests MaxLength validation
func TestMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		maxVal    int
		fieldName string
		json      string
		usePtr    bool
		expectErr bool
		expectVal string
		expectNil bool
		minVal    int // For combined min/max tests
	}{
		{
			name:      "Valid below max",
			maxVal:    10,
			fieldName: "Username",
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Exactly at max",
			maxVal:    5,
			fieldName: "Username",
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Above max",
			maxVal:    5,
			fieldName: "Username",
			json:      `{"username":"verylongusername"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Empty string",
			maxVal:    10,
			fieldName: "Username",
			json:      `{"username":""}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Pointer above max",
			maxVal:    20,
			fieldName: "Bio",
			json:      `{"bio":"this is a very long biography that exceeds the maximum"}`,
			usePtr:    true,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Pointer valid",
			maxVal:    20,
			fieldName: "Bio",
			json:      `{"bio":"short bio"}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "short bio",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Nil pointer",
			maxVal:    20,
			fieldName: "Bio",
			json:      `{"bio":null}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "",
			expectNil: true,
			minVal:    0,
		},
		{
			name:      "Combined min/max valid",
			maxVal:    20,
			fieldName: "Password",
			json:      `{"password":"goodpassword"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "goodpassword",
			expectNil: false,
			minVal:    8,
		},
		{
			name:      "Combined below min",
			maxVal:    20,
			fieldName: "Password",
			json:      `{"password":"short"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    8,
		},
		{
			name:      "Combined above max",
			maxVal:    20,
			fieldName: "Password",
			json:      `{"password":"thispasswordiswaytoolongforourvalidation"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.usePtr {
				// Non-pointer test case
				if tt.minVal > 0 {
					// Combined min/max test - use Password field with both constraints
					// UserWithPassword represents the data structure
					type UserWithPassword struct {
						Password string `json:"password" pedantigo:"min=8,max=20"`
					}
					validator := New[UserWithPassword]()
					user, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr {
						require.Error(t, err)
						ve, ok := err.(*ValidationError)
						require.True(t, ok, "expected *ValidationError, got %T", err)
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == "Password" {
								foundError = true
								break
							}
						}
						assert.True(t, foundError, "expected error for Password field, got %v", ve.Errors)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tt.expectVal, user.Password)
					}
				} else {
					// Max-only tests - use Username field with only max constraint
					// UserWithUsername represents the data structure
					type UserWithUsername struct {
						Username string `json:"username" pedantigo:"max=10"`
					}

					// Handle different max values
					if tt.maxVal == 5 {
						type UserMax5 struct {
							Username string `json:"username" pedantigo:"max=5"`
						}
						validator := New[UserMax5]()
						user, err := validator.Unmarshal([]byte(tt.json))

						if tt.expectErr {
							require.Error(t, err)
							ve, ok := err.(*ValidationError)
							require.True(t, ok, "expected *ValidationError, got %T", err)
							expectedMsg := fmt.Sprintf("must be at most %d characters", tt.maxVal)
							foundError := false
							for _, fieldErr := range ve.Errors {
								if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
									foundError = true
									break
								}
							}
							assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
						} else {
							assert.NoError(t, err)
							assert.Equal(t, tt.expectVal, user.Username)
						}
					} else {
						validator := New[UserWithUsername]()
						user, err := validator.Unmarshal([]byte(tt.json))

						if tt.expectErr {
							require.Error(t, err)
							ve, ok := err.(*ValidationError)
							require.True(t, ok, "expected *ValidationError, got %T", err)
							expectedMsg := fmt.Sprintf("must be at most %d characters", tt.maxVal)
							foundError := false
							for _, fieldErr := range ve.Errors {
								if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
									foundError = true
									break
								}
							}
							assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
						} else {
							assert.NoError(t, err)
							assert.Equal(t, tt.expectVal, user.Username)
						}
					}
				}
			} else {
				// Pointer test case
				// User represents the data structure
				type User struct {
					Bio *string `json:"bio" pedantigo:"max=20"`
				}

				validator := New[User]()
				user, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr {
					require.Error(t, err)
					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)
					expectedMsg := "must be at most 20 characters"
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Bio" && fieldErr.Message == expectedMsg {
							foundError = true
							break
						}
					}
					assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
				} else {
					assert.NoError(t, err)
					if tt.expectNil {
						assert.Nil(t, user.Bio)
					} else {
						require.NotNil(t, user.Bio)
						assert.Equal(t, tt.expectVal, *user.Bio)
					}
				}
			}
		})
	}
}

// ==================== Numeric Constraints ====================

// ==================================================
// gt (greater than) constraint tests
// ==================================================

// TestGt tests Gt validation
func TestGt(t *testing.T) {
	tests := []struct {
		name         string
		valueType    string // "int", "float64", "uint", "intPtr"
		fieldName    string
		jsonValue    string
		expectErr    bool
		expectVal    any
		expectNil    bool
		expectErrMsg string
	}{
		// int tests
		{
			name:      "int valid above threshold",
			valueType: "int",
			fieldName: "Stock",
			jsonValue: "5",
			expectErr: false,
			expectVal: 5,
		},
		{
			name:         "int equal to threshold",
			valueType:    "int",
			fieldName:    "Stock",
			jsonValue:    "0",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		{
			name:         "int below threshold",
			valueType:    "int",
			fieldName:    "Stock",
			jsonValue:    "-5",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		// float64 tests
		{
			name:      "float64 valid above threshold",
			valueType: "float64",
			fieldName: "Price",
			jsonValue: "19.99",
			expectErr: false,
			expectVal: 19.99,
		},
		{
			name:         "float64 below threshold",
			valueType:    "float64",
			fieldName:    "Price",
			jsonValue:    "-1.5",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		// uint tests
		{
			name:      "uint valid above threshold",
			valueType: "uint",
			fieldName: "Port",
			jsonValue: "8080",
			expectErr: false,
			expectVal: uint(8080),
		},
		// pointer tests
		{
			name:         "intPtr with invalid value",
			valueType:    "intPtr",
			fieldName:    "Stock",
			jsonValue:    "0",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		{
			name:      "intPtr with valid value",
			valueType: "intPtr",
			fieldName: "Stock",
			jsonValue: "10",
			expectErr: false,
			expectVal: 10,
		},
		{
			name:      "intPtr with nil value",
			valueType: "intPtr",
			fieldName: "Stock",
			jsonValue: "null",
			expectErr: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.valueType {
			case "int":
				type Product struct {
					Stock int `json:"stock" pedantigo:"gt=0"`
				}

				validator := New[Product]()
				jsonData := []byte(`{"stock":` + tt.jsonValue + `}`)
				product, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					require.Error(t, err)

					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal.(int), product.Stock)
				}

			case "float64":
				type Product struct {
					Price float64 `json:"price" pedantigo:"gt=0"`
				}

				validator := New[Product]()
				jsonData := []byte(`{"price":` + tt.jsonValue + `}`)
				product, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					require.Error(t, err)

					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal.(float64), product.Price)
				}

			case "uint":
				type Config struct {
					Port uint `json:"port" pedantigo:"gt=1024"`
				}

				validator := New[Config]()
				jsonData := []byte(`{"port":` + tt.jsonValue + `}`)
				config, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					require.Error(t, err)

					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVal.(uint), config.Port)
				}

			case "intPtr":
				type Product struct {
					Stock *int `json:"stock" pedantigo:"gt=0"`
				}

				validator := New[Product]()
				jsonData := []byte(`{"stock":` + tt.jsonValue + `}`)
				product, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					require.Error(t, err)

					ve, ok := err.(*ValidationError)
					require.True(t, ok, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					assert.NoError(t, err)

					if tt.expectNil {
						assert.Nil(t, product.Stock)
					} else {
						require.NotNil(t, product.Stock)
						assert.Equal(t, tt.expectVal.(int), *product.Stock)
					}
				}
			}
		})
	}
}

// ==================================================
// ge (greater or equal) constraint tests
// ==================================================

// TestGe tests Ge validation
func TestGe(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gte=0"`
	}

	tests := []struct {
		name            string
		jsonData        []byte
		expectedValue   int
		expectError     bool
		expectedMessage string
	}{
		{
			name:          "int valid above threshold",
			jsonData:      []byte(`{"stock":5}`),
			expectedValue: 5,
			expectError:   false,
		},
		{
			name:          "int equal to threshold",
			jsonData:      []byte(`{"stock":0}`),
			expectedValue: 0,
			expectError:   false,
		},
		{
			name:            "int below threshold",
			jsonData:        []byte(`{"stock":-1}`),
			expectError:     true,
			expectedMessage: "must be at least 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Product]()
			product, err := validator.Unmarshal(tt.jsonData)

			if tt.expectError {
				require.Error(t, err)

				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Stock" && fieldErr.Message == tt.expectedMessage {
						foundError = true
						break
					}
				}

				assert.True(t, foundError, "expected error message %q, got %v", tt.expectedMessage, ve.Errors)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, product.Stock)
			}
		})
	}
}

// ==================================================
// lt (less than) constraint tests
// ==================================================

// TestLt tests Lt validation
func TestLt(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lt=100"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		expectedErr bool
		expectedVal int
	}{
		{
			name:        "int valid below threshold",
			jsonData:    []byte(`{"discount":50}`),
			expectedErr: false,
			expectedVal: 50,
		},
		{
			name:        "int equal to threshold",
			jsonData:    []byte(`{"discount":100}`),
			expectedErr: true,
			expectedVal: 0,
		},
		{
			name:        "int above threshold",
			jsonData:    []byte(`{"discount":150}`),
			expectedErr: true,
			expectedVal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Product]()
			product, err := validator.Unmarshal(tt.jsonData)

			if tt.expectedErr {
				require.Error(t, err)

				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Discount" && fieldErr.Message == "must be less than 100" {
						foundError = true
					}
				}

				assert.True(t, foundError, "expected 'must be less than 100' error, got %v", ve.Errors)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVal, product.Discount)
			}
		})
	}
}

// ==================================================
// le (less or equal) constraint tests
// ==================================================

// TestLe tests Le validation
func TestLe(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lte=100"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		expectedErr bool
		expectedVal int
	}{
		{
			name:        "int valid below threshold",
			jsonData:    []byte(`{"discount":50}`),
			expectedErr: false,
			expectedVal: 50,
		},
		{
			name:        "int equal to threshold",
			jsonData:    []byte(`{"discount":100}`),
			expectedErr: false,
			expectedVal: 100,
		},
		{
			name:        "int above threshold",
			jsonData:    []byte(`{"discount":150}`),
			expectedErr: true,
			expectedVal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Product]()
			product, err := validator.Unmarshal(tt.jsonData)

			if tt.expectedErr {
				require.Error(t, err)

				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Discount" && fieldErr.Message == "must be at most 100" {
						foundError = true
					}
				}

				assert.True(t, foundError, "expected 'must be at most 100' error, got %v", ve.Errors)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVal, product.Discount)
			}
		})
	}
}

// ==================== Enum Constraint ====================

// ==================================================
// enum constraint tests
// ==================================================

// TestEnum tests Enum validation
func TestEnum(t *testing.T) {
	tests := []struct {
		name       string
		testType   string // "string_valid", "string_invalid", "int_valid", "int_invalid", "slice", "map", "schema"
		json       string
		expectErr  bool
		errorField string
		errorMsg   string
	}{
		// String enum tests
		{
			name:      "valid string enum",
			testType:  "string_valid",
			json:      `{"role":"admin"}`,
			expectErr: false,
		},
		{
			name:       "invalid string enum",
			testType:   "string_invalid",
			json:       `{"role":"superadmin"}`,
			expectErr:  true,
			errorField: "Role",
			errorMsg:   "must be one of: admin, user, guest",
		},
		// Integer enum tests
		{
			name:      "valid integer enum",
			testType:  "int_valid",
			json:      `{"code":200}`,
			expectErr: false,
		},
		{
			name:       "invalid integer enum",
			testType:   "int_invalid",
			json:       `{"code":404}`,
			expectErr:  true,
			errorField: "Code",
			errorMsg:   "must be one of: 200, 201, 204",
		},
		// Collection tests
		{
			name:       "enum in slice",
			testType:   "slice",
			json:       `{"roles":["admin","user","superadmin"]}`,
			expectErr:  true,
			errorField: "Roles[2]",
			errorMsg:   "must be one of: admin, user, guest",
		},
		{
			name:       "enum in map",
			testType:   "map",
			json:       `{"permissions":{"file":"read","script":"delete"}}`,
			expectErr:  true,
			errorField: "Permissions[script]",
			errorMsg:   "must be one of: read, write, execute",
		},
		// Schema test
		{
			name:      "schema generation",
			testType:  "schema",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case "string_valid":
				type User struct {
					Role string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				validator := New[User]()
				user, err := validator.Unmarshal([]byte(tt.json))
				assert.NoError(t, err)
				assert.Equal(t, "admin", user.Role)

			case "string_invalid":
				type User struct {
					Role string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				validator := New[User]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, tt.errorMsg, ve.Errors)

			case "int_valid":
				type Status struct {
					Code int `json:"code" pedantigo:"oneof=200 201 204"`
				}
				validator := New[Status]()
				status, err := validator.Unmarshal([]byte(tt.json))
				assert.NoError(t, err)
				assert.Equal(t, 200, status.Code)

			case "int_invalid":
				type Status struct {
					Code int `json:"code" pedantigo:"oneof=200 201 204"`
				}
				validator := New[Status]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, tt.errorMsg, ve.Errors)

			case "slice":
				type Config struct {
					Roles []string `json:"roles" pedantigo:"oneof=admin user guest"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)
				assert.Len(t, ve.Errors, 1)
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error at field=%s, got %v", tt.errorField, ve.Errors)

			case "map":
				type Config struct {
					Permissions map[string]string `json:"permissions" pedantigo:"oneof=read write execute"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				ve, ok := err.(*ValidationError)
				require.True(t, ok, "expected *ValidationError, got %T", err)
				assert.Len(t, ve.Errors, 1)
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error at field=%s, got %v", tt.errorField, ve.Errors)

			case "schema":
				type User struct {
					Role string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				validator := New[User]()
				schema := validator.Schema()

				roleProp, ok := schema.Properties.Get("role")
				require.True(t, ok && roleProp != nil, "expected 'role' property to exist")

				assert.Len(t, roleProp.Enum, 3)

				expectedValues := map[string]bool{"admin": false, "user": false, "guest": false}
				for _, val := range roleProp.Enum {
					strVal, ok := val.(string)
					if !ok {
						t.Errorf("expected enum value to be string, got %T", val)
						continue
					}
					if _, exists := expectedValues[strVal]; exists {
						expectedValues[strVal] = true
					}
				}

				for val, found := range expectedValues {
					assert.True(t, found, "expected enum value '%s' not found", val)
				}
			}
		})
	}
}
