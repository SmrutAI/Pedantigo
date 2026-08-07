package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJwtConstraint tests jwtConstraint.Validate() for valid JWT format (3 base64url parts).
func TestJwtConstraint(t *testing.T) {
	runSimpleConstraintTests(t, jwtConstraint{}, []simpleTestCase{
		// Valid JWTs (3 base64url parts separated by dots)
		// Using obviously fake/test tokens to avoid gitleaks detection
		{"valid JWT 3 parts", "aGVhZGVy.cGF5bG9hZA.c2lnbmF0dXJl", false}, // header.payload.signature in base64
		{"valid JWT alphanumeric", "abc123.def456.ghi789", false},        // simple alphanumeric parts
		{"valid JWT with underscores", "abc_123.def_456.ghi_789", false}, // base64url allows underscores
		{"valid JWT with hyphens", "abc-123.def-456.ghi-789", false},     // base64url allows hyphens
		{"valid JWT longer parts", "abcdefghijklmnop.qrstuvwxyz0123456789.ABCDEFG", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid JWTs
		{"invalid not a jwt", "notajwt", true},
		{"invalid only 2 parts", "header.payload", true},
		{"invalid 4 parts", "header.payload.signature.extra", true},
		{"invalid 5 parts", "only.two.parts.here.extra", true},
		{"invalid empty parts", "...", true},
		{"invalid single dot", "a.b", true},
		{"invalid no dots", "nodots", true},
		{"invalid spaces", "header. payload.signature", true},
		{"invalid with newlines", "header\n.payload.signature", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestJsonConstraint tests jsonConstraint.Validate() for valid JSON strings.
func TestJsonConstraint(t *testing.T) {
	runSimpleConstraintTests(t, jsonConstraint{}, []simpleTestCase{
		// Valid JSON
		{"valid empty object", "{}", false},
		{"valid empty array", "[]", false},
		{"valid object with key", "{\"key\":\"value\"}", false},
		{"valid array with items", "[1,2,3]", false},
		{"valid nested object", "{\"outer\":{\"inner\":\"value\"}}", false},
		{"valid nested array", "[[1,2],[3,4]]", false},
		{"valid string", "\"hello\"", false},
		{"valid number", "123", false},
		{"valid boolean true", "true", false},
		{"valid boolean false", "false", false},
		{"valid null", "null", false},
		{"valid with whitespace", "{ \"key\" : \"value\" }", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid JSON
		{"invalid single brace", "{", true},
		{"invalid unquoted key", "{invalid}", true},
		{"invalid trailing comma", "{\"key\":\"value\",}", true},
		{"invalid plain text", "not json", true},
		{"invalid unclosed string", "{\"key\":\"value", true},
		{"invalid single quotes", "{'key':'value'}", true},
		{"invalid undefined", "undefined", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestBase64Constraint tests base64Constraint.Validate() for valid base64 encoding.
func TestBase64Constraint(t *testing.T) {
	runSimpleConstraintTests(t, base64Constraint{}, []simpleTestCase{
		// Valid base64
		{"valid hello world", "SGVsbG8gV29ybGQ=", false},
		{"valid abcd", "YWJjZA==", false},
		{"valid single char", "YQ==", false},
		{"valid two chars", "YWI=", false},
		{"valid three chars", "YWJj", false},
		{"valid long string", "VGhpcyBpcyBhIGxvbmdlciBzdHJpbmcgZm9yIHRlc3Rpbmc=", false},
		{"valid empty base64", "", false}, // empty string is valid/skipped
		{"valid with plus", "a+b+", false},
		{"valid with slash", "a/b/", false},
		// Invalid base64
		{"invalid special chars", "not valid base64!", true},
		{"invalid at symbol", "SGVsbG8@", true},
		{"invalid underscore (url encoding)", "SGVsbG8_V29ybGQ", true},
		{"invalid hyphen (url encoding)", "SGVsbG8-V29ybGQ", true},
		{"invalid wrong padding", "YWJjZA=", true},
		{"invalid padding in middle", "YW=JjZA==", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestBase64urlConstraint tests base64urlConstraint.Validate() for valid base64url encoding.
func TestBase64urlConstraint(t *testing.T) {
	runSimpleConstraintTests(t, base64urlConstraint{}, []simpleTestCase{
		// Valid base64url (uses - and _ instead of + and /, may have = padding)
		{"valid hello world", "SGVsbG8gV29ybGQ", false},
		{"valid abcd", "YWJjZA", false},
		{"valid with underscore", "a_b_", false},
		{"valid with hyphen", "a-b-", false},
		{"valid with padding", "YWJjZA==", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid base64url - contains + or /
		{"invalid has plus", "has+plus", true},
		{"invalid has slash", "has/slash", true},
		{"invalid has both", "has+and/both", true},
		{"invalid special chars", "invalid!chars", true},
		{"invalid at symbol", "invalid@char", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestBase64rawurlConstraint tests base64rawurlConstraint.Validate() for valid base64 raw URL encoding (no padding).
func TestBase64rawurlConstraint(t *testing.T) {
	runSimpleConstraintTests(t, base64rawurlConstraint{}, []simpleTestCase{
		// Valid base64rawurl (uses - and _, no = padding)
		{"valid hello world no padding", "SGVsbG8gV29ybGQ", false},
		{"valid abcd no padding", "YWJjZA", false},
		{"valid with underscore", "a_b_", false},
		{"valid with hyphen", "a-b-", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid base64rawurl - contains padding
		{"invalid has single padding", "YWJjZA=", true},
		{"invalid has double padding", "YWJjZA==", true},
		{"invalid has padding middle", "YWJj=ZA", true},
		// Invalid characters
		{"invalid has plus", "has+plus", true},
		{"invalid has slash", "has/slash", true},
		{"invalid special chars", "invalid!chars", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

func TestDataURIConstraint(t *testing.T) {
	c := datauriConstraint{}

	tests := []struct {
		name      string
		input     any
		wantErr   bool
		errorCode string
	}{
		// Valid Data URIs (RFC 2397)
		{"minimal data uri", "data:,hello", false, ""},
		{"text plain", "data:text/plain,hello%20world", false, ""},
		{"text plain base64", "data:text/plain;base64,SGVsbG8gV29ybGQ=", false, ""},
		{"image png base64", "data:image/png;base64,iVBORw0KGgo=", false, ""},
		{"with charset", "data:text/plain;charset=utf-8,hello", false, ""},
		{"html content", "data:text/html,<h1>Hello</h1>", false, ""},
		{"json content", "data:application/json,{}", false, ""},
		{"svg", "data:image/svg+xml,<svg></svg>", false, ""},

		// Edge cases
		{"empty string skipped", "", false, ""},
		{"nil skipped", nil, false, ""},

		// Invalid Data URIs
		{"missing data prefix", "text/plain,hello", true, CodeInvalidDataURI},
		{"missing comma", "data:text/plain", true, CodeInvalidDataURI},
		{"just data:", "data:", true, CodeInvalidDataURI},
		{"not a data uri", "https://example.com", true, CodeInvalidDataURI},
		{"random string", "hello world", true, CodeInvalidDataURI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Validate(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				var ce *ConstraintError
				require.ErrorAs(t, err, &ce)
				assert.Equal(t, tt.errorCode, ce.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBase32Constraint(t *testing.T) {
	c := base32Constraint{}

	// Test data: "Hello" in various encodings
	// Base32 of "Hello" = "JBSWY3DP" (without padding) or "JBSWY3DP======" (with padding)
	// Base32 uses A-Z and 2-7

	tests := []struct {
		name      string
		input     any
		wantErr   bool
		errorCode string
	}{
		// Valid Base32
		{"valid base32 with padding", "JBSWY3DPEHPK3PXP", false, ""},
		{"valid base32 hello", "JBSWY3DP", false, ""},
		{"valid with padding", "MFRGGZDFMY======", false, ""},
		{"valid uppercase", "GEZDGNBVGY3TQOJQ", false, ""},
		{"empty encoding", "", false, ""}, // Empty is valid (skip)

		// Edge cases
		{"nil skipped", nil, false, ""},

		// Invalid Base32
		{"lowercase invalid", "jbswy3dp", true, CodeInvalidBase32},
		{"contains 0", "JBSWY30P", true, CodeInvalidBase32},
		{"contains 1", "JBSWY31P", true, CodeInvalidBase32},
		{"contains 8", "JBSWY38P", true, CodeInvalidBase32},
		{"contains 9", "JBSWY39P", true, CodeInvalidBase32},
		{"invalid chars", "JBSWY3DP!!!", true, CodeInvalidBase32},
		{"wrong padding", "JBSWY3DP=", true, CodeInvalidBase32},
		{"spaces", "JBSW Y3DP", true, CodeInvalidBase32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Validate(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				var ce *ConstraintError
				require.ErrorAs(t, err, &ce)
				assert.Equal(t, tt.errorCode, ce.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataURIConstraint_PointerTypes(t *testing.T) {
	c := datauriConstraint{}

	// Valid pointer
	valid := "data:text/plain,hello"
	assert.NoError(t, c.Validate(&valid))

	// Nil pointer
	var nilPtr *string
	assert.NoError(t, c.Validate(nilPtr))
}

func TestBase32Constraint_PointerTypes(t *testing.T) {
	c := base32Constraint{}

	// Valid pointer
	valid := "JBSWY3DP"
	assert.NoError(t, c.Validate(&valid))

	// Nil pointer
	var nilPtr *string
	assert.NoError(t, c.Validate(nilPtr))
}
