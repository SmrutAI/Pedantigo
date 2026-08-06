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
	Password string `json:"password" validate:"exclude:response|log"`
	Token    string `json:"token" validate:"exclude:log"`
	Port     int    `json:"port" validate:"omitzero"`
	Debug    bool   `json:"debug,omitempty"`
}

type TestConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port" validate:"omitzero"`
	APIKey   string `json:"api_key" validate:"exclude:response"`
	Internal string `json:"internal" validate:"exclude:log|response"`
	Enabled  *bool  `json:"enabled" validate:"omitzero"`
	Count    *int   `json:"count"`
}

type TestNested struct {
	Name    string      `json:"name"`
	Profile TestProfile `json:"profile"`
}

type TestProfile struct {
	Email    string `json:"email"`
	Password string `json:"password" validate:"exclude:response"`
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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "validate")

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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "validate")

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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestConfig{}), "validate")

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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "validate")

	// Port has omitzero tag
	portMeta := metadata["port"]
	assert.True(t, portMeta.OmitZero)

	// ID does not have omitzero tag
	idMeta := metadata["id"]
	assert.False(t, idMeta.OmitZero)
}

func TestBuildFieldMetadata_OmitEmpty(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "validate")

	// Debug has json:",omitempty" tag
	debugMeta := metadata["debug"]
	assert.True(t, debugMeta.OmitEmpty)

	// Name does not have omitempty tag
	nameMeta := metadata["name"]
	assert.False(t, nameMeta.OmitEmpty)
}

func TestBuildFieldMetadata_JSONDash(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestJSONDash{}), "validate")

	// Should have metadata for "name"
	assert.Contains(t, metadata, "name")

	// Should NOT have metadata for "internal" (json:"-")
	assert.NotContains(t, metadata, "internal")
	assert.NotContains(t, metadata, "-")
}

func TestBuildFieldMetadata_UnexportedFields(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestPrivateFields{}), "validate")

	// Should have metadata for exported field
	assert.Contains(t, metadata, "public")

	// Should NOT have metadata for unexported field
	assert.NotContains(t, metadata, "private")
}

func TestBuildFieldMetadata_PointerType(t *testing.T) {
	// Should handle pointer to struct
	metadata := BuildFieldMetadata(reflect.TypeOf(&TestUser{}), "validate")

	// Should still work and extract fields
	assert.Contains(t, metadata, "id")
	assert.Contains(t, metadata, "name")
	assert.Contains(t, metadata, "password")
}

func TestBuildFieldMetadata_NonStructType(t *testing.T) {
	// Should return empty map for non-struct types
	metadata := BuildFieldMetadata(reflect.TypeOf("string"), "validate")
	assert.Empty(t, metadata)

	metadata = BuildFieldMetadata(reflect.TypeOf(42), "validate")
	assert.Empty(t, metadata)

	metadata = BuildFieldMetadata(reflect.TypeOf([]int{1, 2, 3}), "validate")
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

	opts := Options{
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

	opts := Options{
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

	opts := Options{
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
		opts       Options
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
			opts: Options{
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
			opts: Options{
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
			opts: Options{
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

	opts := Options{
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

	opts := Options{
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

	opts := Options{
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

	opts := Options{
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
	opts1 := Options{
		Context:  "response",
		OmitZero: true,
	}
	fieldValue1 := reflect.ValueOf("value")
	assert.False(t, ShouldIncludeField(meta, fieldValue1, opts1, false), "should be excluded by context")

	// Test 2: Not excluded by context, but zero value (should exclude by omitzero)
	opts2 := Options{
		Context:  "api",
		OmitZero: true,
	}
	fieldValue2 := reflect.ValueOf("")
	assert.False(t, ShouldIncludeField(meta, fieldValue2, opts2, false), "should be excluded by omitzero")

	// Test 3: Not excluded by context, non-zero value (should include)
	opts3 := Options{
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

	opts := Options{
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

	opts := Options{
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

	opts := Options{
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

	opts := Options{
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
	ID       int    `json:"id" validate:"include:summary|public"`
	Email    string `json:"email" validate:"include:summary|contact"`
	Phone    string `json:"phone" validate:"include:contact"`
	Password string `json:"password"` // No include tags
}

func TestBuildFieldMetadata_IncludeContexts(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestIncludeUser{}), "validate")

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

	metadata := BuildFieldMetadata(reflect.TypeOf(user), "validate")

	// Test "summary" context - should only include ID and Email
	optsSummary := Options{
		Context:  "summary",
		OmitZero: false,
		TagName:  "validate",
	}
	resultSummary := ToFilteredMap(reflect.ValueOf(user), metadata, optsSummary)

	assert.Contains(t, resultSummary, "id")
	assert.Contains(t, resultSummary, "email")
	assert.NotContains(t, resultSummary, "phone")
	assert.NotContains(t, resultSummary, "password")

	// Test "contact" context - should only include Email and Phone
	optsContact := Options{
		Context:  "contact",
		OmitZero: false,
		TagName:  "validate",
	}
	resultContact := ToFilteredMap(reflect.ValueOf(user), metadata, optsContact)

	assert.NotContains(t, resultContact, "id")
	assert.Contains(t, resultContact, "email")
	assert.Contains(t, resultContact, "phone")
	assert.NotContains(t, resultContact, "password")

	// Test no context - should include all fields
	optsNone := Options{
		Context:  "",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(obj), "validate")
	opts := Options{
		Context:  "",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(user), "validate")
	opts := Options{
		Context:  "response",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(user), "validate")
	opts := Options{
		Context:  "",
		OmitZero: true, // OmitZero enabled
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(nested), "validate")
	opts := Options{
		Context:  "response",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(nested), "validate")
	opts := Options{
		Context:  "response",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}), "validate")
	opts := Options{
		Context:  "",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(*user), "validate")
	opts := Options{
		Context:  "",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(config), "validate")

	// Test "response" context
	optsResponse := Options{
		Context:  "response",
		OmitZero: false,
		TagName:  "validate",
	}
	resultResponse := ToFilteredMap(reflect.ValueOf(config), metadata, optsResponse)

	assert.Equal(t, "localhost", resultResponse["host"])
	assert.NotContains(t, resultResponse, "api_key")  // Excluded in "response"
	assert.NotContains(t, resultResponse, "internal") // Excluded in "response"

	// Test "log" context
	optsLog := Options{
		Context:  "log",
		OmitZero: false,
		TagName:  "validate",
	}
	resultLog := ToFilteredMap(reflect.ValueOf(config), metadata, optsLog)

	assert.Equal(t, "localhost", resultLog["host"])
	assert.Contains(t, resultLog, "api_key")     // NOT excluded in "log"
	assert.NotContains(t, resultLog, "internal") // Excluded in "log"

	// Test no context
	optsNone := Options{
		Context:  "",
		OmitZero: false,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(config), "validate")
	opts := Options{
		Context:  "",
		OmitZero: true,
		TagName:  "validate",
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

	metadata := BuildFieldMetadata(reflect.TypeOf(config), "validate")
	opts := Options{
		Context:  "",
		OmitZero: true,
		TagName:  "validate",
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
	metadata := BuildFieldMetadata(reflect.TypeOf(obj), "validate")
	opts := Options{
		Context:  "",
		OmitZero: false,
		TagName:  "validate",
	}

	result := ToFilteredMap(reflect.ValueOf(obj), metadata, opts)

	// Should return empty map
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

// ==================== isZeroValue Tests ====================

func TestIsZeroValue_Channel(t *testing.T) {
	// Nil channel
	var nilChan chan int
	assert.True(t, isZeroValue(reflect.ValueOf(nilChan)), "nil channel should be zero")

	// Non-nil channel
	nonNilChan := make(chan int)
	assert.False(t, isZeroValue(reflect.ValueOf(nonNilChan)), "non-nil channel should not be zero")
}

func TestIsZeroValue_Func(t *testing.T) {
	// Nil func
	var nilFunc func()
	assert.True(t, isZeroValue(reflect.ValueOf(nilFunc)), "nil func should be zero")

	// Non-nil func
	nonNilFunc := func() {}
	assert.False(t, isZeroValue(reflect.ValueOf(nonNilFunc)), "non-nil func should not be zero")
}

func TestIsZeroValue_Array(t *testing.T) {
	// Zero array (all elements are zero)
	var zeroArray [3]int
	assert.True(t, isZeroValue(reflect.ValueOf(zeroArray)), "zero array should be zero")

	// Non-zero array (has at least one non-zero element)
	nonZeroArray := [3]int{0, 1, 0}
	assert.False(t, isZeroValue(reflect.ValueOf(nonZeroArray)), "non-zero array should not be zero")

	// Empty array
	var emptyArray [0]int
	assert.True(t, isZeroValue(reflect.ValueOf(emptyArray)), "empty array should be zero")
}

func TestIsZeroValue_ArrayOfPointers(t *testing.T) {
	// Array of nil pointers
	var nilPtrArray [2]*int
	assert.True(t, isZeroValue(reflect.ValueOf(nilPtrArray)), "array of nil pointers should be zero")

	// Array with one non-nil pointer
	val := 42
	nonNilPtrArray := [2]*int{nil, &val}
	assert.False(t, isZeroValue(reflect.ValueOf(nonNilPtrArray)), "array with non-nil pointer should not be zero")
}

func TestIsZeroValue_Struct(t *testing.T) {
	type Inner struct {
		Value int
	}

	// Zero struct
	var zeroStruct Inner
	assert.True(t, isZeroValue(reflect.ValueOf(zeroStruct)), "zero struct should be zero")

	// Non-zero struct
	nonZeroStruct := Inner{Value: 42}
	assert.False(t, isZeroValue(reflect.ValueOf(nonZeroStruct)), "non-zero struct should not be zero")
}

func TestIsZeroValue_NestedStruct(t *testing.T) {
	type Nested struct {
		Inner struct {
			Value int
		}
	}

	// Zero nested struct
	var zeroNested Nested
	assert.True(t, isZeroValue(reflect.ValueOf(zeroNested)), "zero nested struct should be zero")

	// Non-zero nested struct
	nonZeroNested := Nested{}
	nonZeroNested.Inner.Value = 42
	assert.False(t, isZeroValue(reflect.ValueOf(nonZeroNested)), "non-zero nested struct should not be zero")
}

func TestIsZeroValue_Interface(t *testing.T) {
	// Nil interface
	var nilInterface interface{}
	assert.True(t, isZeroValue(reflect.ValueOf(&nilInterface).Elem()), "nil interface should be zero")

	// Non-nil interface
	var nonNilInterface interface{} = 42
	assert.False(t, isZeroValue(reflect.ValueOf(&nonNilInterface).Elem()), "non-nil interface should not be zero")
}

func TestIsZeroValue_Map(t *testing.T) {
	// Nil map
	var nilMap map[string]int
	assert.True(t, isZeroValue(reflect.ValueOf(nilMap)), "nil map should be zero")

	// Empty but non-nil map
	emptyMap := make(map[string]int)
	assert.False(t, isZeroValue(reflect.ValueOf(emptyMap)), "non-nil empty map should not be zero")
}

func TestIsZeroValue_Slice(t *testing.T) {
	// Nil slice
	var nilSlice []int
	assert.True(t, isZeroValue(reflect.ValueOf(nilSlice)), "nil slice should be zero")

	// Empty but non-nil slice
	emptySlice := make([]int, 0)
	assert.False(t, isZeroValue(reflect.ValueOf(emptySlice)), "non-nil empty slice should not be zero")
}

func TestIsZeroValue_Pointer(t *testing.T) {
	// Nil pointer
	var nilPtr *int
	assert.True(t, isZeroValue(reflect.ValueOf(nilPtr)), "nil pointer should be zero")

	// Non-nil pointer to zero value
	zeroVal := 0
	assert.False(t, isZeroValue(reflect.ValueOf(&zeroVal)), "non-nil pointer should not be zero")
}

func TestIsZeroValue_Primitives(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero float64", 0.0, true},
		{"non-zero float64", 3.14, false},
		{"zero string", "", true},
		{"non-zero string", "hello", false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"zero uint", uint(0), true},
		{"non-zero uint", uint(42), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroValue(reflect.ValueOf(tt.value))
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsZeroValue_StructWithPointerField(t *testing.T) {
	type WithPointer struct {
		Ptr *int
	}

	// Zero (nil pointer field)
	var zero WithPointer
	assert.True(t, isZeroValue(reflect.ValueOf(zero)), "struct with nil pointer should be zero")

	// Non-zero (non-nil pointer field)
	val := 42
	nonZero := WithPointer{Ptr: &val}
	assert.False(t, isZeroValue(reflect.ValueOf(nonZero)), "struct with non-nil pointer should not be zero")
}

func TestIsZeroValue_ArrayOfStructs(t *testing.T) {
	type Inner struct {
		Value int
	}

	// Zero array of structs
	var zeroArray [2]Inner
	assert.True(t, isZeroValue(reflect.ValueOf(zeroArray)), "array of zero structs should be zero")

	// Non-zero array of structs
	nonZeroArray := [2]Inner{{Value: 1}, {Value: 0}}
	assert.False(t, isZeroValue(reflect.ValueOf(nonZeroArray)), "array with non-zero struct should not be zero")
}
