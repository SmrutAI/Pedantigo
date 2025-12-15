package constraints

import "testing"

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
