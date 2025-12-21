package pedantigo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test field name constants.
const testFieldUsername = "Username"

// assertValidationError checks if a ValidationError contains the expected field and message.
func assertValidationError(t *testing.T, err error, expectedField, expectedMessage string) {
	t.Helper()
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == expectedField && fieldErr.Message == expectedMessage {
			foundError = true
			break
		}
	}
	assert.True(t, foundError, "expected error field=%s msg=%s, got %v", expectedField, expectedMessage, ve.Errors)
}

// assertNoValidationError checks that no error occurred.
func assertNoValidationError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// assertNumericValidationError checks if a ValidationError contains the expected numeric field and message.
func assertNumericValidationError(t *testing.T, err error, expectedField, expectedMessage string) {
	t.Helper()
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == expectedField && fieldErr.Message == expectedMessage {
			foundError = true
			break
		}
	}
	assert.True(t, foundError, "expected error field=%s msg=%s, got %v", expectedField, expectedMessage, ve.Errors)
}

// ==================== Format Constraints ====================

// formatTestCase defines the structure for format constraint tests.
type formatTestCase struct {
	name       string
	json       string
	usePointer bool
	expectErr  bool
	expectVal  string
	expectNil  bool
}

// runFormatValidationTests is a generic helper for testing format constraints.
// It executes test logic for both pointer and non-pointer string field scenarios.
// The test functions passed as callbacks must handle struct creation and validation.
//
// testFunc is called with the test case and must perform validation assertions.
// The function signature is: func(t *testing.T, tt formatTestCase).
func runFormatValidationTests(
	t *testing.T,
	tests []formatTestCase,
	testFunc func(*testing.T, formatTestCase),
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFunc(t, tt)
		})
	}
}

// ==================================================
// Format Constraints (url, uuid, ipv4, ipv6)
// ==================================================

// TestFormatConstraints tests format validation constraints.
func TestFormatConstraints(t *testing.T) {
	formatTests := []formatTestConfig{
		{
			constraintType: "url",
			fieldName:      "Website",
			jsonFieldName:  "website",
			pedantigoTag:   "url",
			expectedErrMsg: "must be a valid URL (http or https)",
			testCases: []formatTestCase{
				{"Valid HTTPS", `{"website":"https://example.com"}`, false, false, "https://example.com", false},
				{"Valid HTTP", `{"website":"http://example.com"}`, false, false, "http://example.com", false},
				{"Invalid format", `{"website":"not a url"}`, false, true, "", false},
				{"No scheme", `{"website":"example.com"}`, false, true, "", false},
				{"FTP scheme", `{"website":"ftp://example.com"}`, false, true, "", false},
				{"Empty string", `{"website":""}`, false, false, "", false},
				{"Pointer invalid", `{"website":"not a url"}`, true, true, "", false},
				{"Pointer valid", `{"website":"https://example.com"}`, true, false, "https://example.com", false},
				{"Nil pointer", `{"website":null}`, true, false, "", true},
			},
		},
		{
			constraintType: "uuid",
			fieldName:      "ID",
			jsonFieldName:  "id",
			pedantigoTag:   "uuid",
			expectedErrMsg: "must be a valid UUID",
			testCases: []formatTestCase{
				{"Valid V4", `{"id":"550e8400-e29b-41d4-a716-446655440000"}`, false, false, "550e8400-e29b-41d4-a716-446655440000", false},
				{"Valid V5", `{"id":"886313e1-3b8a-5372-9b90-0c9aee199e5d"}`, false, false, "886313e1-3b8a-5372-9b90-0c9aee199e5d", false},
				{"Invalid format", `{"id":"not-a-uuid"}`, false, true, "", false},
				{"Wrong dashes", `{"id":"550e8400e29b41d4a716446655440000"}`, false, true, "", false},
				{"Empty string", `{"id":""}`, false, false, "", false},
				{"Pointer invalid", `{"id":"not-a-uuid"}`, true, true, "", false},
				{"Pointer valid", `{"id":"550e8400-e29b-41d4-a716-446655440000"}`, true, false, "550e8400-e29b-41d4-a716-446655440000", false},
				{"Nil pointer", `{"id":null}`, true, false, "", true},
			},
		},
		{
			constraintType: "ipv4",
			fieldName:      "IP",
			jsonFieldName:  "ip",
			pedantigoTag:   "ipv4",
			expectedErrMsg: "must be a valid IPv4 address",
			testCases: []formatTestCase{
				{"Valid localhost", `{"ip":"127.0.0.1"}`, false, false, "127.0.0.1", false},
				{"Valid private", `{"ip":"192.168.1.1"}`, false, false, "192.168.1.1", false},
				{"Invalid format", `{"ip":"not-an-ip"}`, false, true, "", false},
				{"Invalid IPv6", `{"ip":"2001:0db8:85a3::8a2e:0370:7334"}`, false, true, "", false},
				{"Empty string", `{"ip":""}`, false, false, "", false},
				{"Pointer invalid", `{"ip":"not-an-ip"}`, true, true, "", false},
				{"Pointer valid", `{"ip":"10.0.0.1"}`, true, false, "10.0.0.1", false},
				{"Nil pointer", `{"ip":null}`, true, false, "", true},
			},
		},
		{
			constraintType: "ipv6",
			fieldName:      "IP",
			jsonFieldName:  "ip",
			pedantigoTag:   "ipv6",
			expectedErrMsg: "must be a valid IPv6 address",
			testCases: []formatTestCase{
				{"Valid localhost", `{"ip":"::1"}`, false, false, "::1", false},
				{"Valid full", `{"ip":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`, false, false, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
				{"Valid compressed", `{"ip":"2001:db8:85a3::8a2e:370:7334"}`, false, false, "2001:db8:85a3::8a2e:370:7334", false},
				{"Invalid format", `{"ip":"not-an-ip"}`, false, true, "", false},
				{"Invalid IPv4", `{"ip":"192.168.1.1"}`, false, true, "", false},
				{"Empty string", `{"ip":""}`, false, false, "", false},
				{"Pointer invalid", `{"ip":"not-an-ip"}`, true, true, "", false},
				{"Pointer valid", `{"ip":"fe80::1"}`, true, false, "fe80::1", false},
				{"Nil pointer", `{"ip":null}`, true, false, "", true},
			},
		},
	}

	for _, ft := range formatTests {
		ft := ft // capture range variable
		t.Run(ft.constraintType, func(t *testing.T) {
			runFormatValidationTests(t, ft.testCases, func(t *testing.T, tt formatTestCase) {
				testFormatConstraintVariant(t, &ft, tt)
			})
		})
	}
}

// formatTestConfig holds configuration for format constraint testing.
type formatTestConfig struct {
	constraintType string
	fieldName      string
	jsonFieldName  string
	pedantigoTag   string
	expectedErrMsg string
	testCases      []formatTestCase
}

// testFormatConstraintVariant runs a single format constraint test variant.
func testFormatConstraintVariant(t *testing.T, ft *formatTestConfig, tt formatTestCase) {
	t.Helper()

	switch ft.constraintType {
	case "url":
		runFormatTest(t, tt, ft.fieldName, ft.expectedErrMsg, urlFormatValidator{})
	case "uuid":
		runFormatTest(t, tt, ft.fieldName, ft.expectedErrMsg, uuidFormatValidator{})
	case "ipv4":
		runFormatTest(t, tt, ft.fieldName, ft.expectedErrMsg, ipv4FormatValidator{})
	case "ipv6":
		runFormatTest(t, tt, ft.fieldName, ft.expectedErrMsg, ipv6FormatValidator{})
	}
}

// formatValidator interface enables unified format constraint testing.
type formatValidator interface {
	unmarshalPointer(json []byte) (val *string, err error)
	unmarshalValue(json []byte) (val string, err error)
}

// runFormatTest is a generic helper that validates format constraints using the provided validator.
func runFormatTest(t *testing.T, tt formatTestCase, fieldName, errMsg string, fv formatValidator) {
	t.Helper()

	if tt.usePointer {
		val, err := fv.unmarshalPointer([]byte(tt.json))
		if tt.expectErr {
			assertValidationError(t, err, fieldName, errMsg)
		} else {
			assertNoValidationError(t, err)
			if tt.expectNil {
				assert.Nil(t, val)
			} else {
				require.NotNil(t, val)
				assert.Equal(t, tt.expectVal, *val)
			}
		}
	} else {
		val, err := fv.unmarshalValue([]byte(tt.json))
		if tt.expectErr {
			assertValidationError(t, err, fieldName, errMsg)
		} else {
			assertNoValidationError(t, err)
			assert.Equal(t, tt.expectVal, val)
		}
	}
}

// URL format validator types and methods.
type urlFormatValidator struct{}
type urlPtrStruct struct {
	Website *string `json:"website" pedantigo:"url"`
}
type urlValStruct struct {
	Website string `json:"website" pedantigo:"url"`
}

func (urlFormatValidator) unmarshalPointer(json []byte) (*string, error) {
	v := New[urlPtrStruct]()
	r, err := v.Unmarshal(json)
	return r.Website, err
}
func (urlFormatValidator) unmarshalValue(json []byte) (string, error) {
	v := New[urlValStruct]()
	r, err := v.Unmarshal(json)
	return r.Website, err
}

// UUID format validator types and methods.
type uuidFormatValidator struct{}
type uuidPtrStruct struct {
	ID *string `json:"id" pedantigo:"uuid"`
}
type uuidValStruct struct {
	ID string `json:"id" pedantigo:"uuid"`
}

func (uuidFormatValidator) unmarshalPointer(json []byte) (*string, error) {
	v := New[uuidPtrStruct]()
	r, err := v.Unmarshal(json)
	return r.ID, err
}
func (uuidFormatValidator) unmarshalValue(json []byte) (string, error) {
	v := New[uuidValStruct]()
	r, err := v.Unmarshal(json)
	return r.ID, err
}

// IPv4 format validator types and methods.
type ipv4FormatValidator struct{}
type ipv4PtrStruct struct {
	IP *string `json:"ip" pedantigo:"ipv4"`
}
type ipv4ValStruct struct {
	IP string `json:"ip" pedantigo:"ipv4"`
}

func (ipv4FormatValidator) unmarshalPointer(json []byte) (*string, error) {
	v := New[ipv4PtrStruct]()
	r, err := v.Unmarshal(json)
	return r.IP, err
}
func (ipv4FormatValidator) unmarshalValue(json []byte) (string, error) {
	v := New[ipv4ValStruct]()
	r, err := v.Unmarshal(json)
	return r.IP, err
}

// IPv6 format validator types and methods.
type ipv6FormatValidator struct{}
type ipv6PtrStruct struct {
	IP *string `json:"ip" pedantigo:"ipv6"`
}
type ipv6ValStruct struct {
	IP string `json:"ip" pedantigo:"ipv6"`
}

func (ipv6FormatValidator) unmarshalPointer(json []byte) (*string, error) {
	v := New[ipv6PtrStruct]()
	r, err := v.Unmarshal(json)
	return r.IP, err
}
func (ipv6FormatValidator) unmarshalValue(json []byte) (string, error) {
	v := New[ipv6ValStruct]()
	r, err := v.Unmarshal(json)
	return r.IP, err
}

// ==================================================
// regex constraint tests
// ==================================================

// TestRegex_UppercasePattern tests Regex uppercasepattern.
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
					assertValidationError(t, err, "Value", "must match pattern '^[A-Z]{3}$'")
				} else {
					assertNoValidationError(t, err)
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
					assertValidationError(t, err, "Value", "must match pattern '^[A-Z]{3}$'")
				} else {
					assertNoValidationError(t, err)
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectVal, code.Value)
			}
		})
	}
}

// ==================== String Constraints ====================

// ==================================================
// min_length constraint tests
// ==================================================

// TestMinLength tests MinLength validation.
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
			fieldName: testFieldUsername,
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
		},
		{
			name:      "Exactly at min",
			minVal:    3,
			fieldName: testFieldUsername,
			json:      `{"username":"bob"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "bob",
			expectNil: false,
		},
		{
			name:      "Below min",
			minVal:    3,
			fieldName: testFieldUsername,
			json:      `{"username":"ab"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Empty string",
			minVal:    1,
			fieldName: testFieldUsername,
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
						var ve *ValidationError
						require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
						expectedMsg := "must be at least 1 characters"
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == testFieldUsername && fieldErr.Message == expectedMsg {
								foundError = true
								break
							}
						}
						assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
					} else {
						require.NoError(t, err)
					}
				} else {
					validator := New[User]()
					user, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr {
						require.Error(t, err)
						var ve *ValidationError
						require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
						expectedMsg := "must be at least 3 characters"
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == testFieldUsername && fieldErr.Message == expectedMsg {
								foundError = true
								break
							}
						}
						assert.True(t, foundError, "expected error message %q, got %v", expectedMsg, ve.Errors)
					} else {
						require.NoError(t, err)
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
					var ve *ValidationError
					require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
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

// max length test struct types (moved outside to reduce cognitive complexity).
type userMax10 struct {
	Username string `json:"username" pedantigo:"max=10"`
}
type userMax5 struct {
	Username string `json:"username" pedantigo:"max=5"`
}
type userMinMax struct {
	Password string `json:"password" pedantigo:"min=8,max=20"`
}
type userBioPtr struct {
	Bio *string `json:"bio" pedantigo:"max=20"`
}

// runMaxLengthStringTest handles non-pointer max length tests.
func runMaxLengthStringTest(t *testing.T, maxVal int, json string, expectErr bool, expectVal string) {
	t.Helper()
	var val string
	var err error
	if maxVal == 5 {
		v := New[userMax5]()
		r, e := v.Unmarshal([]byte(json))
		val, err = r.Username, e
	} else {
		v := New[userMax10]()
		r, e := v.Unmarshal([]byte(json))
		val, err = r.Username, e
	}
	if expectErr {
		assertValidationError(t, err, testFieldUsername, fmt.Sprintf("must be at most %d characters", maxVal))
	} else {
		require.NoError(t, err)
		assert.Equal(t, expectVal, val)
	}
}

// runMinMaxTest handles combined min/max tests.
func runMinMaxTest(t *testing.T, json string, expectErr bool, expectVal string) {
	t.Helper()
	v := New[userMinMax]()
	r, err := v.Unmarshal([]byte(json))
	if expectErr {
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		foundError := false
		for _, fieldErr := range ve.Errors {
			if fieldErr.Field == "Password" {
				foundError = true
				break
			}
		}
		assert.True(t, foundError, "expected error for Password field")
	} else {
		require.NoError(t, err)
		assert.Equal(t, expectVal, r.Password)
	}
}

// runBioPtrTest handles pointer max length tests.
func runBioPtrTest(t *testing.T, json string, expectErr bool, expectVal string, expectNil bool) {
	t.Helper()
	v := New[userBioPtr]()
	r, err := v.Unmarshal([]byte(json))
	if expectErr {
		assertValidationError(t, err, "Bio", "must be at most 20 characters")
	} else {
		require.NoError(t, err)
		if expectNil {
			assert.Nil(t, r.Bio)
		} else {
			require.NotNil(t, r.Bio)
			assert.Equal(t, expectVal, *r.Bio)
		}
	}
}

// TestMaxLength tests MaxLength validation.
func TestMaxLength(t *testing.T) {
	t.Run("string max=10 valid", func(t *testing.T) {
		runMaxLengthStringTest(t, 10, `{"username":"alice"}`, false, "alice")
	})
	t.Run("string max=5 at boundary", func(t *testing.T) {
		runMaxLengthStringTest(t, 5, `{"username":"alice"}`, false, "alice")
	})
	t.Run("string max=5 above max", func(t *testing.T) {
		runMaxLengthStringTest(t, 5, `{"username":"verylongusername"}`, true, "")
	})
	t.Run("string empty", func(t *testing.T) {
		runMaxLengthStringTest(t, 10, `{"username":""}`, false, "")
	})
	t.Run("pointer above max", func(t *testing.T) {
		runBioPtrTest(t, `{"bio":"this is a very long biography that exceeds the maximum"}`, true, "", false)
	})
	t.Run("pointer valid", func(t *testing.T) {
		runBioPtrTest(t, `{"bio":"short bio"}`, false, "short bio", false)
	})
	t.Run("pointer nil", func(t *testing.T) {
		runBioPtrTest(t, `{"bio":null}`, false, "", true)
	})
	t.Run("combined min/max valid", func(t *testing.T) {
		runMinMaxTest(t, `{"password":"goodpassword"}`, false, "goodpassword")
	})
	t.Run("combined below min", func(t *testing.T) {
		runMinMaxTest(t, `{"password":"short"}`, true, "")
	})
	t.Run("combined above max", func(t *testing.T) {
		runMinMaxTest(t, `{"password":"thispasswordiswaytoolongforourvalidation"}`, true, "")
	})
}

// ==================== Numeric Constraints ====================

// ==================================================
// gt (greater than) constraint tests
// ==================================================

// TestGt tests Gt validation.
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

					var ve *ValidationError
					require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					require.NoError(t, err)
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

					var ve *ValidationError
					require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					require.NoError(t, err)
					assert.InDelta(t, tt.expectVal.(float64), product.Price, 0.0001)
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

					var ve *ValidationError
					require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					assert.True(t, foundError, "expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
				} else {
					require.NoError(t, err)
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

					var ve *ValidationError
					require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

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

// TestGe tests Ge validation.
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

				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Stock" && fieldErr.Message == tt.expectedMessage {
						foundError = true
						break
					}
				}

				assert.True(t, foundError, "expected error message %q, got %v", tt.expectedMessage, ve.Errors)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedValue, product.Stock)
			}
		})
	}
}

// ==================================================
// lt and lte (less than / less or equal) constraint tests
// ==================================================

// ltLteProduct types for lt/lte testing.
type ltProduct struct {
	Discount int `json:"discount" pedantigo:"lt=100"`
}
type lteProduct struct {
	Discount int `json:"discount" pedantigo:"lte=100"`
}

// TestLtLte tests lt and lte constraints in a unified table-driven test.
func TestLtLte(t *testing.T) {
	tests := []struct {
		name           string
		constraint     string // "lt" or "lte"
		jsonData       []byte
		expectedErr    bool
		expectedVal    int
		expectedErrMsg string
	}{
		// lt (less than) tests
		{"lt: valid below threshold", "lt", []byte(`{"discount":50}`), false, 50, ""},
		{"lt: equal to threshold", "lt", []byte(`{"discount":100}`), true, 0, "must be less than 100"},
		{"lt: above threshold", "lt", []byte(`{"discount":150}`), true, 0, "must be less than 100"},
		// lte (less or equal) tests
		{"lte: valid below threshold", "lte", []byte(`{"discount":50}`), false, 50, ""},
		{"lte: equal to threshold", "lte", []byte(`{"discount":100}`), false, 100, ""},
		{"lte: above threshold", "lte", []byte(`{"discount":150}`), true, 0, "must be at most 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var discount int
			var err error

			switch tt.constraint {
			case "lt":
				v := New[ltProduct]()
				p, e := v.Unmarshal(tt.jsonData)
				discount, err = p.Discount, e
			case "lte":
				v := New[lteProduct]()
				p, e := v.Unmarshal(tt.jsonData)
				discount, err = p.Discount, e
			}

			if tt.expectedErr {
				assertNumericValidationError(t, err, "Discount", tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedVal, discount)
			}
		})
	}
}

// ==================== Enum Constraint ====================

// ==================================================
// enum constraint tests
// ==================================================

// TestEnum tests Enum validation.
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
				require.NoError(t, err)
				assert.Equal(t, "admin", user.Role)

			case "string_invalid":
				type User struct {
					Role string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				validator := New[User]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
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
				require.NoError(t, err)
				assert.Equal(t, 200, status.Code)

			case "int_invalid":
				type Status struct {
					Code int `json:"code" pedantigo:"oneof=200 201 204"`
				}
				validator := New[Status]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, tt.errorMsg, ve.Errors)

			case "slice":
				type Config struct {
					Roles []string `json:"roles" pedantigo:"dive,oneof=admin user guest"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
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
					Permissions map[string]string `json:"permissions" pedantigo:"dive,oneof=read write execute"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
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

// ==================== Eq Constraint ====================

// ==================================================
// eq constraint tests
// ==================================================

// ErrMsgEqMustEqual is the expected error message format for eq constraint violations.
const ErrMsgEqMustEqual = "must be equal to %s"

// TestEq tests the eq constraint for literal/constant value validation.
func TestEq(t *testing.T) {
	tests := []struct {
		name       string
		testType   string // "string_valid", "string_invalid", "int_valid", "int_invalid", "float_valid", "pointer_valid", "pointer_nil"
		json       string
		expectErr  bool
		errorField string
		eqValue    string
	}{
		// String eq tests
		{
			name:      "valid string eq - exact match",
			testType:  "string_valid",
			json:      `{"type":"user"}`,
			expectErr: false,
		},
		{
			name:       "invalid string eq - wrong value",
			testType:   "string_invalid",
			json:       `{"type":"admin"}`,
			expectErr:  true,
			errorField: "Type",
			eqValue:    "user",
		},
		{
			name:       "invalid string eq - empty value",
			testType:   "string_invalid",
			json:       `{"type":""}`,
			expectErr:  true,
			errorField: "Type",
			eqValue:    "user",
		},
		// Integer eq tests
		{
			name:      "valid int eq - exact match",
			testType:  "int_valid",
			json:      `{"version":1}`,
			expectErr: false,
		},
		{
			name:       "invalid int eq - wrong value",
			testType:   "int_invalid",
			json:       `{"version":2}`,
			expectErr:  true,
			errorField: "Version",
			eqValue:    "1",
		},
		// Float eq tests
		{
			name:      "valid float eq - exact match",
			testType:  "float_valid",
			json:      `{"rate":0.5}`,
			expectErr: false,
		},
		// Pointer tests
		{
			name:      "valid pointer eq - exact match",
			testType:  "pointer_valid",
			json:      `{"kind":"event"}`,
			expectErr: false,
		},
		{
			name:      "nil pointer skips validation",
			testType:  "pointer_nil",
			json:      `{"kind":null}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case "string_valid":
				type Message struct {
					Type string `json:"type" pedantigo:"eq=user"`
				}
				validator := New[Message]()
				msg, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.Equal(t, "user", msg.Type)

			case "string_invalid":
				type Message struct {
					Type string `json:"type" pedantigo:"eq=user"`
				}
				validator := New[Message]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				foundError := false
				expectedMsg := fmt.Sprintf(ErrMsgEqMustEqual, tt.eqValue)
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == expectedMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, expectedMsg, ve.Errors)

			case "int_valid":
				type Config struct {
					Version int `json:"version" pedantigo:"eq=1"`
				}
				validator := New[Config]()
				cfg, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.Equal(t, 1, cfg.Version)

			case "int_invalid":
				type Config struct {
					Version int `json:"version" pedantigo:"eq=1"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				foundError := false
				expectedMsg := fmt.Sprintf(ErrMsgEqMustEqual, tt.eqValue)
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == expectedMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, expectedMsg, ve.Errors)

			case "float_valid":
				type Rate struct {
					Rate float64 `json:"rate" pedantigo:"eq=0.5"`
				}
				validator := New[Rate]()
				r, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.InDelta(t, 0.5, r.Rate, 0.0001)

			case "pointer_valid":
				type Event struct {
					Kind *string `json:"kind" pedantigo:"eq=event"`
				}
				validator := New[Event]()
				ev, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				require.NotNil(t, ev.Kind)
				assert.Equal(t, "event", *ev.Kind)

			case "pointer_nil":
				type Event struct {
					Kind *string `json:"kind" pedantigo:"eq=event"`
				}
				validator := New[Event]()
				ev, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.Nil(t, ev.Kind)
			}
		})
	}
}

// TestEqWithCombinedConstraints tests eq with other constraints.
func TestEqWithCombinedConstraints(t *testing.T) {
	// eq combined with required
	t.Run("eq with required", func(t *testing.T) {
		type APIVersion struct {
			Version string `json:"version" pedantigo:"required,eq=v1"`
		}
		validator := New[APIVersion]()

		// Valid case
		result, err := validator.Unmarshal([]byte(`{"version":"v1"}`))
		require.NoError(t, err)
		assert.Equal(t, "v1", result.Version)

		// Missing field - should fail required
		_, err = validator.Unmarshal([]byte(`{}`))
		require.Error(t, err)
	})
}

// TestEqBoolValue tests eq constraint with boolean values.
func TestEqBoolValue(t *testing.T) {
	type Feature struct {
		Enabled bool `json:"enabled" pedantigo:"eq=true"`
	}
	validator := New[Feature]()

	// Valid - exactly true
	result, err := validator.Unmarshal([]byte(`{"enabled":true}`))
	require.NoError(t, err)
	assert.True(t, result.Enabled)

	// Invalid - false when eq=true
	_, err = validator.Unmarshal([]byte(`{"enabled":false}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	foundError := false
	expectedMsg := fmt.Sprintf(ErrMsgEqMustEqual, "true")
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Enabled" && fieldErr.Message == expectedMsg {
			foundError = true
		}
	}
	assert.True(t, foundError, "expected eq error, got %v", ve.Errors)
}

// ==================== Ne Constraint ====================

// ==================================================
// ne constraint tests
// ==================================================

// ErrMsgNeMustNotEqual is the expected error message format for ne constraint violations.
const ErrMsgNeMustNotEqual = "must not be equal to %s"

// TestNe tests the ne constraint for "not equal" validation.
func TestNe(t *testing.T) {
	tests := []struct {
		name       string
		testType   string
		json       string
		expectErr  bool
		errorField string
		neValue    string
	}{
		// String ne tests - pass when NOT equal
		{
			name:      "valid string ne - not equal passes",
			testType:  "string_valid",
			json:      `{"status":"active"}`,
			expectErr: false,
		},
		{
			name:       "invalid string ne - equal fails",
			testType:   "string_invalid",
			json:       `{"status":"banned"}`,
			expectErr:  true,
			errorField: "Status",
			neValue:    "banned",
		},
		// Integer ne tests
		{
			name:      "valid int ne - not equal passes",
			testType:  "int_valid",
			json:      `{"version":2}`,
			expectErr: false,
		},
		{
			name:       "invalid int ne - equal fails",
			testType:   "int_invalid",
			json:       `{"version":0}`,
			expectErr:  true,
			errorField: "Version",
			neValue:    "0",
		},
		// Pointer tests
		{
			name:      "valid pointer ne - not equal passes",
			testType:  "pointer_valid",
			json:      `{"kind":"active"}`,
			expectErr: false,
		},
		{
			name:      "nil pointer skips validation",
			testType:  "pointer_nil",
			json:      `{"kind":null}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case "string_valid":
				type Status struct {
					Status string `json:"status" pedantigo:"ne=banned"`
				}
				validator := New[Status]()
				s, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.Equal(t, "active", s.Status)

			case "string_invalid":
				type Status struct {
					Status string `json:"status" pedantigo:"ne=banned"`
				}
				validator := New[Status]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				foundError := false
				expectedMsg := fmt.Sprintf(ErrMsgNeMustNotEqual, tt.neValue)
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == expectedMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, expectedMsg, ve.Errors)

			case "int_valid":
				type Config struct {
					Version int `json:"version" pedantigo:"ne=0"`
				}
				validator := New[Config]()
				cfg, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.Equal(t, 2, cfg.Version)

			case "int_invalid":
				type Config struct {
					Version int `json:"version" pedantigo:"ne=0"`
				}
				validator := New[Config]()
				_, err := validator.Unmarshal([]byte(tt.json))
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve, "expected *ValidationError, got %T", err)
				foundError := false
				expectedMsg := fmt.Sprintf(ErrMsgNeMustNotEqual, tt.neValue)
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == expectedMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, expectedMsg, ve.Errors)

			case "pointer_valid":
				type Event struct {
					Kind *string `json:"kind" pedantigo:"ne=banned"`
				}
				validator := New[Event]()
				ev, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				require.NotNil(t, ev.Kind)
				assert.Equal(t, "active", *ev.Kind)

			case "pointer_nil":
				type Event struct {
					Kind *string `json:"kind" pedantigo:"ne=banned"`
				}
				validator := New[Event]()
				ev, err := validator.Unmarshal([]byte(tt.json))
				require.NoError(t, err)
				assert.Nil(t, ev.Kind)
			}
		})
	}
}

// TestNeWithCombinedConstraints tests ne with other constraints.
func TestNeWithCombinedConstraints(t *testing.T) {
	// ne combined with required
	t.Run("ne with required", func(t *testing.T) {
		type UserStatus struct {
			Status string `json:"status" pedantigo:"required,ne=deleted"`
		}
		validator := New[UserStatus]()

		// Valid case - not equal to "deleted"
		result, err := validator.Unmarshal([]byte(`{"status":"active"}`))
		require.NoError(t, err)
		assert.Equal(t, "active", result.Status)

		// Invalid case - equal to "deleted"
		_, err = validator.Unmarshal([]byte(`{"status":"deleted"}`))
		require.Error(t, err)

		// Missing field - should fail required
		_, err = validator.Unmarshal([]byte(`{}`))
		require.Error(t, err)
	})
}

// ==================== Oneofci Constraint ====================

// TestOneofci tests case-insensitive enum validation.
func TestOneofci(t *testing.T) {
	tests := []struct {
		name       string
		testType   string
		json       string
		expectErr  bool
		errorField string
		errorMsg   string
	}{
		// Case-insensitive matching
		{
			name:      "exact case match",
			testType:  "string_valid",
			json:      `{"role":"admin"}`,
			expectErr: false,
		},
		{
			name:      "uppercase match",
			testType:  "string_valid",
			json:      `{"role":"ADMIN"}`,
			expectErr: false,
		},
		{
			name:      "mixed case match",
			testType:  "string_valid",
			json:      `{"role":"Admin"}`,
			expectErr: false,
		},
		{
			name:      "random case match",
			testType:  "string_valid",
			json:      `{"role":"aDmIn"}`,
			expectErr: false,
		},
		{
			name:       "invalid value",
			testType:   "string_invalid",
			json:       `{"role":"superadmin"}`,
			expectErr:  true,
			errorField: "Role",
			errorMsg:   "must be one of (case-insensitive): admin, user, guest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type User struct {
				Role string `json:"role" pedantigo:"oneofci=admin user guest"`
			}
			validator := New[User]()
			user, err := validator.Unmarshal([]byte(tt.json))

			if tt.expectErr {
				require.Error(t, err)
				var ve *ValidationError
				require.ErrorAs(t, err, &ve)
				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == tt.errorField && fieldErr.Message == tt.errorMsg {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error field=%s msg=%s, got %v", tt.errorField, tt.errorMsg, ve.Errors)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, user.Role)
			}
		})
	}
}

// ==================== Iscolor Alias ====================

// TestIscolorAlias tests the built-in iscolor alias that expands to color format validators.
func TestIscolorAlias(t *testing.T) {
	type Config struct {
		Color string `json:"color" pedantigo:"iscolor"`
	}
	validator := New[Config]()

	tests := []struct {
		name      string
		json      string
		expectErr bool
	}{
		// Valid hexcolor
		{"valid hexcolor 3 digit", `{"color":"#fff"}`, false},
		{"valid hexcolor 6 digit", `{"color":"#ffffff"}`, false},
		{"valid hexcolor uppercase", `{"color":"#AABBCC"}`, false},
		// Valid rgb
		{"valid rgb", `{"color":"rgb(255,255,255)"}`, false},
		{"valid rgb with spaces", `{"color":"rgb(0, 128, 255)"}`, false},
		// Valid rgba
		{"valid rgba", `{"color":"rgba(255,255,255,1)"}`, false},
		{"valid rgba decimal alpha", `{"color":"rgba(0,0,0,0.5)"}`, false},
		// Valid hsl
		{"valid hsl", `{"color":"hsl(360,100%,50%)"}`, false},
		// Valid hsla
		{"valid hsla", `{"color":"hsla(180,50%,50%,0.8)"}`, false},
		// Empty string - should be skipped
		{"empty string", `{"color":""}`, false},
		// Invalid colors
		{"invalid - plain text", `{"color":"red"}`, true},
		{"invalid - missing hash", `{"color":"ffffff"}`, true},
		{"invalid - wrong format", `{"color":"color(255,0,0)"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Unmarshal([]byte(tt.json))
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ==================== URI Constraint ====================

// TestUriConstraint tests the uri constraint for database and other URI formats.
func TestUriConstraintIntegration(t *testing.T) {
	type Config struct {
		DatabaseURL string `json:"database_url" pedantigo:"uri"`
	}
	validator := New[Config]()

	tests := []struct {
		name      string
		json      string
		expectErr bool
	}{
		// Valid database URIs
		{"postgres URI", `{"database_url":"postgres://user:pass@localhost:5432/smrut"}`, false},
		{"postgresql URI", `{"database_url":"postgresql://user@localhost/db"}`, false},
		{"mysql URI", `{"database_url":"mysql://root:password@127.0.0.1:3306/mydb"}`, false},
		{"redis URI", `{"database_url":"redis://localhost:6379/0"}`, false},
		{"mongodb URI", `{"database_url":"mongodb://localhost:27017/testdb"}`, false},
		// Valid web URIs
		{"http URI", `{"database_url":"http://example.com"}`, false},
		{"https URI", `{"database_url":"https://example.com"}`, false},
		// Other valid schemes
		{"s3 URI", `{"database_url":"s3://bucket-name/key"}`, false},
		{"file URI", `{"database_url":"file:///etc/passwd"}`, false},
		// Empty string - should be skipped
		{"empty string", `{"database_url":""}`, false},
		// Invalid URIs
		{"no scheme", `{"database_url":"example.com"}`, true},
		{"relative path", `{"database_url":"/path/to/resource"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Unmarshal([]byte(tt.json))
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
