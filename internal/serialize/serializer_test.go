package serialize

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test types for serialization.
type TestUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password" pedantigo:"exclude:response|log"`
	Token    string `json:"token" pedantigo:"exclude:log"`
	Port     int    `json:"port" pedantigo:"omitzero"`
	Debug    bool   `json:"debug,omitempty"`
}

type TestConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port" pedantigo:"omitzero"`
	APIKey   string `json:"api_key" pedantigo:"exclude:response"`
	Internal string `json:"internal" pedantigo:"exclude:log|response"`
	Enabled  *bool  `json:"enabled" pedantigo:"omitzero"`
	Count    *int   `json:"count"`
}

type TestNested struct {
	Name    string      `json:"name"`
	Profile TestProfile `json:"profile"`
}

type TestProfile struct {
	Email    string `json:"email"`
	Password string `json:"password" pedantigo:"exclude:response"`
}

type TestPrivateFields struct {
	Public  string `json:"public"`
	private string //nolint:unused // unexported for testing
}

type TestJSONDash struct {
	Name     string `json:"name"`
	Internal string `json:"-"`
}

// ==================== BuildFieldMetadata Tests ====================

func TestBuildFieldMetadata_Basic(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "pedantigo")

	// Should have metadata for all exported fields
	assert.Contains(t, metadata, "id")
	assert.Contains(t, metadata, "name")
	assert.Contains(t, metadata, "password")
	assert.Contains(t, metadata, "token")
	assert.Contains(t, metadata, "port")
	assert.Contains(t, metadata, "debug")

	// Verify basic field metadata
	assert.Equal(t, "id", metadata["id"].JSONName)
	assert.Equal(t, "name", metadata["name"].JSONName)
}

func TestBuildFieldMetadata_ExcludeContexts(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "pedantigo")

	// Password excluded from "response" and "log" contexts
	passwordMeta := metadata["password"]
	assert.True(t, passwordMeta.ExcludeContexts["response"])
	assert.True(t, passwordMeta.ExcludeContexts["log"])
	assert.False(t, passwordMeta.ExcludeContexts["api"])

	// Token excluded from "log" context only
	tokenMeta := metadata["token"]
	assert.True(t, tokenMeta.ExcludeContexts["log"])
	assert.False(t, tokenMeta.ExcludeContexts["response"])

	// ID has no exclusions
	idMeta := metadata["id"]
	assert.Empty(t, idMeta.ExcludeContexts)
}

func TestBuildFieldMetadata_MultipleExcludeContexts(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestConfig{}), "pedantigo")

	// Internal excluded from both "log" and "response"
	internalMeta := metadata["internal"]
	assert.True(t, internalMeta.ExcludeContexts["log"])
	assert.True(t, internalMeta.ExcludeContexts["response"])
	assert.Len(t, internalMeta.ExcludeContexts, 2)

	// APIKey excluded from "response" only
	apiKeyMeta := metadata["api_key"]
	assert.True(t, apiKeyMeta.ExcludeContexts["response"])
	assert.Len(t, apiKeyMeta.ExcludeContexts, 1)
}

func TestBuildFieldMetadata_OmitZero(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "pedantigo")

	// Port has omitzero tag
	portMeta := metadata["port"]
	assert.True(t, portMeta.OmitZero)

	// ID does not have omitzero tag
	idMeta := metadata["id"]
	assert.False(t, idMeta.OmitZero)
}

func TestBuildFieldMetadata_OmitEmpty(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "pedantigo")

	// Debug has json:",omitempty" tag
	debugMeta := metadata["debug"]
	assert.True(t, debugMeta.OmitEmpty)

	// Name does not have omitempty tag
	nameMeta := metadata["name"]
	assert.False(t, nameMeta.OmitEmpty)
}

func TestBuildFieldMetadata_JSONDash(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestJSONDash{}), "pedantigo")

	// Should have metadata for "name"
	assert.Contains(t, metadata, "name")

	// Should NOT have metadata for "internal" (json:"-")
	assert.NotContains(t, metadata, "internal")
	assert.NotContains(t, metadata, "-")
}

func TestBuildFieldMetadata_UnexportedFields(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestPrivateFields{}), "pedantigo")

	// Should have metadata for exported field
	assert.Contains(t, metadata, "public")

	// Should NOT have metadata for unexported field
	assert.NotContains(t, metadata, "private")
}

func TestBuildFieldMetadata_PointerType(t *testing.T) {
	// Should handle pointer to struct
	metadata := BuildFieldMetadata(reflect.TypeOf(&TestUser{}), "pedantigo")

	// Should still work and extract fields
	assert.Contains(t, metadata, "id")
	assert.Contains(t, metadata, "name")
	assert.Contains(t, metadata, "password")
}

func TestBuildFieldMetadata_NonStructType(t *testing.T) {
	// Should return empty map for non-struct types
	metadata := BuildFieldMetadata(reflect.TypeOf("string"), "pedantigo")
	assert.Empty(t, metadata)

	metadata = BuildFieldMetadata(reflect.TypeOf(42), "pedantigo")
	assert.Empty(t, metadata)

	metadata = BuildFieldMetadata(reflect.TypeOf([]int{1, 2, 3}), "pedantigo")
	assert.Empty(t, metadata)
}

// ==================== ShouldIncludeField Tests ====================

func TestShouldIncludeField_NoExclusion(t *testing.T) {
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "name",
		ExcludeContexts: make(map[string]bool),
		IncludeContexts: make(map[string]bool),
		OmitZero:        false,
		OmitEmpty:       false,
	}

	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
	}

	fieldValue := reflect.ValueOf("Alice")
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.True(t, result, "field with no exclusions should be included")
}

func TestShouldIncludeField_ExcludeContext_Matches(t *testing.T) {
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "password",
		ExcludeContexts: map[string]bool{"response": true, "log": true},
		IncludeContexts: make(map[string]bool),
		OmitZero:        false,
		OmitEmpty:       false,
	}

	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
	}

	fieldValue := reflect.ValueOf("secret123")
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.False(t, result, "field should be excluded in 'response' context")
}

func TestShouldIncludeField_ExcludeContext_NoMatch(t *testing.T) {
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "password",
		ExcludeContexts: map[string]bool{"log": true},
		IncludeContexts: make(map[string]bool),
		OmitZero:        false,
		OmitEmpty:       false,
	}

	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
	}

	fieldValue := reflect.ValueOf("secret123")
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.True(t, result, "field should be included when context doesn't match exclusion")
}

func TestShouldIncludeField_OmitZero_ZeroValue(t *testing.T) {
	tests := []struct {
		name       string
		meta       FieldMetadata
		opts       SerializeOptions
		fieldValue interface{}
		want       bool
	}{
		{
			name: "zero int with omitzero enabled",
			meta: FieldMetadata{
				JSONName:        "port",
				OmitZero:        true,
				IncludeContexts: make(map[string]bool),
			},
			opts: SerializeOptions{
				OmitZero: true,
			},
			fieldValue: 0,
			want:       false,
		},
		{
			name: "zero string with omitzero enabled",
			meta: FieldMetadata{
				JSONName:        "name",
				OmitZero:        true,
				IncludeContexts: make(map[string]bool),
			},
			opts: SerializeOptions{
				OmitZero: true,
			},
			fieldValue: "",
			want:       false,
		},
		{
			name: "false bool with omitzero enabled",
			meta: FieldMetadata{
				JSONName:        "enabled",
				OmitZero:        true,
				IncludeContexts: make(map[string]bool),
			},
			opts: SerializeOptions{
				OmitZero: true,
			},
			fieldValue: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldValue := reflect.ValueOf(tt.fieldValue)
			result := ShouldIncludeField(tt.meta, fieldValue, tt.opts, false)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestShouldIncludeField_OmitZero_NonZeroValue(t *testing.T) {
	meta := FieldMetadata{
		JSONName:        "port",
		OmitZero:        true,
		IncludeContexts: make(map[string]bool),
	}

	opts := SerializeOptions{
		OmitZero: true,
	}

	fieldValue := reflect.ValueOf(8080)
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.True(t, result, "non-zero value should be included even with omitzero")
}

func TestShouldIncludeField_OmitZero_NilPointer(t *testing.T) {
	meta := FieldMetadata{
		JSONName:        "count",
		OmitZero:        true,
		IncludeContexts: make(map[string]bool),
	}

	opts := SerializeOptions{
		OmitZero: true,
	}

	var ptr *int
	fieldValue := reflect.ValueOf(ptr)
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.False(t, result, "nil pointer should be omitted with omitzero")
}

func TestShouldIncludeField_OmitZero_NonNilPointer(t *testing.T) {
	meta := FieldMetadata{
		JSONName:        "count",
		OmitZero:        true,
		IncludeContexts: make(map[string]bool),
	}

	opts := SerializeOptions{
		OmitZero: true,
	}

	val := 42
	fieldValue := reflect.ValueOf(&val)
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.True(t, result, "non-nil pointer should be included")
}

func TestShouldIncludeField_OmitZero_Disabled(t *testing.T) {
	meta := FieldMetadata{
		JSONName:        "port",
		OmitZero:        true,
		IncludeContexts: make(map[string]bool),
	}

	opts := SerializeOptions{
		OmitZero: false, // OmitZero disabled in options
	}

	fieldValue := reflect.ValueOf(0)
	result := ShouldIncludeField(meta, fieldValue, opts, false)

	assert.True(t, result, "zero value should be included when OmitZero is disabled in options")
}

func TestShouldIncludeField_CombinedExcludeAndOmitZero(t *testing.T) {
	// Field with both exclude context and omitzero
	meta := FieldMetadata{
		JSONName:        "internal",
		ExcludeContexts: map[string]bool{"response": true},
		IncludeContexts: make(map[string]bool),
		OmitZero:        true,
	}

	// Test 1: Excluded by context (should exclude regardless of zero)
	opts1 := SerializeOptions{
		Context:  "response",
		OmitZero: true,
	}
	fieldValue1 := reflect.ValueOf("value")
	assert.False(t, ShouldIncludeField(meta, fieldValue1, opts1, false), "should be excluded by context")

	// Test 2: Not excluded by context, but zero value (should exclude by omitzero)
	opts2 := SerializeOptions{
		Context:  "api",
		OmitZero: true,
	}
	fieldValue2 := reflect.ValueOf("")
	assert.False(t, ShouldIncludeField(meta, fieldValue2, opts2, false), "should be excluded by omitzero")

	// Test 3: Not excluded by context, non-zero value (should include)
	opts3 := SerializeOptions{
		Context:  "api",
		OmitZero: true,
	}
	fieldValue3 := reflect.ValueOf("value")
	assert.True(t, ShouldIncludeField(meta, fieldValue3, opts3, false), "should be included")
}

// ==================== Include Context (Whitelist) Tests ====================

func TestShouldIncludeField_IncludeContext_HasWhitelist_FieldIncluded(t *testing.T) {
	// Field has include:summary tag
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "id",
		ExcludeContexts: make(map[string]bool),
		IncludeContexts: map[string]bool{"summary": true},
	}

	opts := SerializeOptions{
		Context: "summary",
	}

	fieldValue := reflect.ValueOf(123)
	hasWhitelist := true
	result := ShouldIncludeField(meta, fieldValue, opts, hasWhitelist)

	assert.True(t, result, "field with include:summary should be included in summary context")
}

func TestShouldIncludeField_IncludeContext_HasWhitelist_FieldNotIncluded(t *testing.T) {
	// Field does NOT have include:summary tag
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "password",
		ExcludeContexts: make(map[string]bool),
		IncludeContexts: make(map[string]bool), // No include tags
	}

	opts := SerializeOptions{
		Context: "summary",
	}

	fieldValue := reflect.ValueOf("secret")
	hasWhitelist := true
	result := ShouldIncludeField(meta, fieldValue, opts, hasWhitelist)

	assert.False(t, result, "field without include:summary should be excluded when whitelist is active")
}

func TestShouldIncludeField_IncludeContext_NoWhitelist(t *testing.T) {
	// No whitelist active - field should be included
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "password",
		ExcludeContexts: make(map[string]bool),
		IncludeContexts: make(map[string]bool),
	}

	opts := SerializeOptions{
		Context: "other",
	}

	fieldValue := reflect.ValueOf("secret")
	hasWhitelist := false
	result := ShouldIncludeField(meta, fieldValue, opts, hasWhitelist)

	assert.True(t, result, "field should be included when no whitelist is active")
}

func TestShouldIncludeField_ExcludeOverridesInclude(t *testing.T) {
	// Field has BOTH exclude:api and include:api (conflicting)
	// Exclude should win
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "conflicting",
		ExcludeContexts: map[string]bool{"api": true},
		IncludeContexts: map[string]bool{"api": true},
	}

	opts := SerializeOptions{
		Context: "api",
	}

	fieldValue := reflect.ValueOf("value")
	result := ShouldIncludeField(meta, fieldValue, opts, true)

	assert.False(t, result, "exclude should take precedence over include for same context")
}

// ==================== HasWhitelistContext Tests ====================

func TestHasWhitelistContext_ReturnsTrue(t *testing.T) {
	metadata := map[string]FieldMetadata{
		"id": {
			JSONName:        "id",
			IncludeContexts: map[string]bool{"summary": true},
		},
		"name": {
			JSONName:        "name",
			IncludeContexts: make(map[string]bool),
		},
	}

	result := HasWhitelistContext(metadata, "summary")
	assert.True(t, result, "should return true when any field has include:summary")
}

func TestHasWhitelistContext_ReturnsFalse(t *testing.T) {
	metadata := map[string]FieldMetadata{
		"id": {
			JSONName:        "id",
			IncludeContexts: make(map[string]bool),
		},
		"name": {
			JSONName:        "name",
			IncludeContexts: make(map[string]bool),
		},
	}

	result := HasWhitelistContext(metadata, "summary")
	assert.False(t, result, "should return false when no field has include:summary")
}

func TestHasWhitelistContext_EmptyContext(t *testing.T) {
	metadata := map[string]FieldMetadata{
		"id": {
			JSONName:        "id",
			IncludeContexts: map[string]bool{"summary": true},
		},
	}

	result := HasWhitelistContext(metadata, "")
	assert.False(t, result, "should return false for empty context")
}

// ==================== BuildFieldMetadata Include Tests ====================

type TestIncludeUser struct {
	ID       int    `json:"id" pedantigo:"include:summary|public"`
	Email    string `json:"email" pedantigo:"include:summary|contact"`
	Phone    string `json:"phone" pedantigo:"include:contact"`
	Password string `json:"password"` // No include tags
}

func TestBuildFieldMetadata_IncludeContexts(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestIncludeUser{}), "pedantigo")

	// ID has include:summary and include:public
	idMeta := metadata["id"]
	assert.True(t, idMeta.IncludeContexts["summary"])
	assert.True(t, idMeta.IncludeContexts["public"])
	assert.False(t, idMeta.IncludeContexts["contact"])

	// Email has include:summary and include:contact
	emailMeta := metadata["email"]
	assert.True(t, emailMeta.IncludeContexts["summary"])
	assert.True(t, emailMeta.IncludeContexts["contact"])

	// Phone has include:contact only
	phoneMeta := metadata["phone"]
	assert.True(t, phoneMeta.IncludeContexts["contact"])
	assert.False(t, phoneMeta.IncludeContexts["summary"])

	// Password has no include contexts
	passwordMeta := metadata["password"]
	assert.Empty(t, passwordMeta.IncludeContexts)
}

// ==================== ToFilteredMap Include Tests ====================

func TestToFilteredMap_IncludeWhitelist(t *testing.T) {
	user := TestIncludeUser{
		ID:       1,
		Email:    "alice@example.com",
		Phone:    "555-1234",
		Password: "secret",
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(user), "pedantigo")

	// Test "summary" context - should only include ID and Email
	optsSummary := SerializeOptions{
		Context:  "summary",
		OmitZero: false,
		TagName:  "pedantigo",
	}
	resultSummary := ToFilteredMap(reflect.ValueOf(user), metadata, optsSummary)

	assert.Contains(t, resultSummary, "id")
	assert.Contains(t, resultSummary, "email")
	assert.NotContains(t, resultSummary, "phone")
	assert.NotContains(t, resultSummary, "password")

	// Test "contact" context - should only include Email and Phone
	optsContact := SerializeOptions{
		Context:  "contact",
		OmitZero: false,
		TagName:  "pedantigo",
	}
	resultContact := ToFilteredMap(reflect.ValueOf(user), metadata, optsContact)

	assert.NotContains(t, resultContact, "id")
	assert.Contains(t, resultContact, "email")
	assert.Contains(t, resultContact, "phone")
	assert.NotContains(t, resultContact, "password")

	// Test no context - should include all fields
	optsNone := SerializeOptions{
		Context:  "",
		OmitZero: false,
		TagName:  "pedantigo",
	}
	resultNone := ToFilteredMap(reflect.ValueOf(user), metadata, optsNone)

	assert.Contains(t, resultNone, "id")
	assert.Contains(t, resultNone, "email")
	assert.Contains(t, resultNone, "phone")
	assert.Contains(t, resultNone, "password")
}

// ==================== ToFilteredMap Tests ====================

func TestToFilteredMap_Basic(t *testing.T) {
	type SimpleStruct struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	obj := SimpleStruct{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(obj), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(obj), metadata, opts)

	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "alice@example.com", result["email"])
	assert.Equal(t, 25, result["age"])
}

func TestToFilteredMap_ExcludesPassword(t *testing.T) {
	user := TestUser{
		ID:       1,
		Name:     "Alice",
		Password: "secret123",
		Token:    "token456",
		Port:     8080,
		Debug:    true,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(user), "pedantigo")
	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(user), metadata, opts)

	// Should include ID, Name
	assert.Equal(t, 1, result["id"])
	assert.Equal(t, "Alice", result["name"])

	// Should exclude Password (excluded in "response" context)
	assert.NotContains(t, result, "password")

	// Should include Token (not excluded in "response" context)
	assert.Equal(t, "token456", result["token"])
}

func TestToFilteredMap_OmitsZeroPort(t *testing.T) {
	user := TestUser{
		ID:       1,
		Name:     "Alice",
		Password: "secret123",
		Token:    "token456",
		Port:     0, // Zero value with omitzero tag
		Debug:    false,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(user), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: true, // OmitZero enabled
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(user), metadata, opts)

	// Should include ID, Name
	assert.Equal(t, 1, result["id"])
	assert.Equal(t, "Alice", result["name"])

	// Should NOT include Port (zero value with omitzero tag)
	assert.NotContains(t, result, "port")

	// Debug has omitempty in JSON tag, but omitempty is NOT handled by serialize package
	// (it's handled by json.Marshal), so it should be present
	assert.Contains(t, result, "debug")
}

func TestToFilteredMap_NestedStruct(t *testing.T) {
	nested := TestNested{
		Name: "Alice",
		Profile: TestProfile{
			Email:    "alice@example.com",
			Password: "secret123",
		},
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(nested), "pedantigo")
	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(nested), metadata, opts)

	// Should include name
	assert.Equal(t, "Alice", result["name"])

	// Should have nested profile
	require.Contains(t, result, "profile")
	profile, ok := result["profile"].(map[string]any)
	require.True(t, ok, "profile should be a map")

	// Profile should include email
	assert.Equal(t, "alice@example.com", profile["email"])

	// Profile should exclude password in "response" context
	assert.NotContains(t, profile, "password")
}

func TestToFilteredMap_NestedStructPointer(t *testing.T) {
	type NestedWithPointer struct {
		Name    string       `json:"name"`
		Profile *TestProfile `json:"profile"`
	}

	profile := &TestProfile{
		Email:    "alice@example.com",
		Password: "secret123",
	}

	nested := NestedWithPointer{
		Name:    "Alice",
		Profile: profile,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(nested), "pedantigo")
	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(nested), metadata, opts)

	// Should have nested profile
	require.Contains(t, result, "profile")
	profileMap, ok := result["profile"].(map[string]any)
	require.True(t, ok, "profile should be a map")

	// Profile should include email
	assert.Equal(t, "alice@example.com", profileMap["email"])

	// Profile should exclude password in "response" context
	assert.NotContains(t, profileMap, "password")
}

func TestToFilteredMap_NilPointer(t *testing.T) {
	var user *TestUser

	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(user), metadata, opts)

	// Should return nil for nil pointer
	assert.Nil(t, result)
}

func TestToFilteredMap_PointerToStruct(t *testing.T) {
	user := &TestUser{
		ID:       1,
		Name:     "Alice",
		Password: "secret123",
		Token:    "token456",
		Port:     8080,
		Debug:    true,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(*user), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(user), metadata, opts)

	// Should work with pointer to struct
	assert.Equal(t, 1, result["id"])
	assert.Equal(t, "Alice", result["name"])
}

func TestToFilteredMap_MultipleContexts(t *testing.T) {
	config := TestConfig{
		Host:     "localhost",
		Port:     8080,
		APIKey:   "secret-key",
		Internal: "internal-data",
		Count:    nil,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(config), "pedantigo")

	// Test "response" context
	optsResponse := SerializeOptions{
		Context:  "response",
		OmitZero: false,
		TagName:  "pedantigo",
	}
	resultResponse := ToFilteredMap(reflect.ValueOf(config), metadata, optsResponse)

	assert.Equal(t, "localhost", resultResponse["host"])
	assert.NotContains(t, resultResponse, "api_key")  // Excluded in "response"
	assert.NotContains(t, resultResponse, "internal") // Excluded in "response"

	// Test "log" context
	optsLog := SerializeOptions{
		Context:  "log",
		OmitZero: false,
		TagName:  "pedantigo",
	}
	resultLog := ToFilteredMap(reflect.ValueOf(config), metadata, optsLog)

	assert.Equal(t, "localhost", resultLog["host"])
	assert.Contains(t, resultLog, "api_key")     // NOT excluded in "log"
	assert.NotContains(t, resultLog, "internal") // Excluded in "log"

	// Test no context
	optsNone := SerializeOptions{
		Context:  "",
		OmitZero: false,
		TagName:  "pedantigo",
	}
	resultNone := ToFilteredMap(reflect.ValueOf(config), metadata, optsNone)

	assert.Equal(t, "localhost", resultNone["host"])
	assert.Contains(t, resultNone, "api_key")  // NOT excluded
	assert.Contains(t, resultNone, "internal") // NOT excluded
}

func TestToFilteredMap_PointerFields(t *testing.T) {
	enabled := true
	count := 42

	config := TestConfig{
		Host:     "localhost",
		Port:     0,
		APIKey:   "key",
		Internal: "data",
		Enabled:  &enabled,
		Count:    &count,
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(config), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: true,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(config), metadata, opts)

	// Port is zero with omitzero tag - should be omitted
	assert.NotContains(t, result, "port")

	// Enabled is non-nil pointer - should be included
	assert.Equal(t, true, result["enabled"])

	// Count is non-nil pointer - should be included
	assert.Equal(t, 42, result["count"])
}

func TestToFilteredMap_NilPointerFieldWithOmitZero(t *testing.T) {
	config := TestConfig{
		Host:     "localhost",
		Port:     8080,
		APIKey:   "key",
		Internal: "data",
		Enabled:  nil, // Nil pointer with omitzero tag
		Count:    nil, // Nil pointer without omitzero tag
	}

	metadata := BuildFieldMetadata(reflect.TypeOf(config), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: true,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(config), metadata, opts)

	// Enabled is nil pointer with omitzero tag - should be omitted
	assert.NotContains(t, result, "enabled")

	// Count is nil pointer without omitzero tag - should be included
	assert.Nil(t, result["count"])
}

func TestToFilteredMap_EmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	obj := EmptyStruct{}
	metadata := BuildFieldMetadata(reflect.TypeOf(obj), "pedantigo")
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
		TagName:  "pedantigo",
	}

	result := ToFilteredMap(reflect.ValueOf(obj), metadata, opts)

	// Should return empty map
	assert.NotNil(t, result)
	assert.Empty(t, result)
}
