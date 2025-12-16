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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}))

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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}))

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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestConfig{}))

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
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}))

	// Port has omitzero tag
	portMeta := metadata["port"]
	assert.True(t, portMeta.OmitZero)

	// ID does not have omitzero tag
	idMeta := metadata["id"]
	assert.False(t, idMeta.OmitZero)
}

func TestBuildFieldMetadata_OmitEmpty(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}))

	// Debug has json:",omitempty" tag
	debugMeta := metadata["debug"]
	assert.True(t, debugMeta.OmitEmpty)

	// Name does not have omitempty tag
	nameMeta := metadata["name"]
	assert.False(t, nameMeta.OmitEmpty)
}

func TestBuildFieldMetadata_JSONDash(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestJSONDash{}))

	// Should have metadata for "name"
	assert.Contains(t, metadata, "name")

	// Should NOT have metadata for "internal" (json:"-")
	assert.NotContains(t, metadata, "internal")
	assert.NotContains(t, metadata, "-")
}

func TestBuildFieldMetadata_UnexportedFields(t *testing.T) {
	metadata := BuildFieldMetadata(reflect.TypeOf(TestPrivateFields{}))

	// Should have metadata for exported field
	assert.Contains(t, metadata, "public")

	// Should NOT have metadata for unexported field
	assert.NotContains(t, metadata, "private")
}

func TestBuildFieldMetadata_PointerType(t *testing.T) {
	// Should handle pointer to struct
	metadata := BuildFieldMetadata(reflect.TypeOf(&TestUser{}))

	// Should still work and extract fields
	assert.Contains(t, metadata, "id")
	assert.Contains(t, metadata, "name")
	assert.Contains(t, metadata, "password")
}

func TestBuildFieldMetadata_NonStructType(t *testing.T) {
	// Should return empty map for non-struct types
	metadata := BuildFieldMetadata(reflect.TypeOf("string"))
	assert.Empty(t, metadata)

	metadata = BuildFieldMetadata(reflect.TypeOf(42))
	assert.Empty(t, metadata)

	metadata = BuildFieldMetadata(reflect.TypeOf([]int{1, 2, 3}))
	assert.Empty(t, metadata)
}

// ==================== ShouldIncludeField Tests ====================

func TestShouldIncludeField_NoExclusion(t *testing.T) {
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "name",
		ExcludeContexts: make(map[string]bool),
		OmitZero:        false,
		OmitEmpty:       false,
	}

	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
	}

	fieldValue := reflect.ValueOf("Alice")
	result := ShouldIncludeField(meta, fieldValue, opts)

	assert.True(t, result, "field with no exclusions should be included")
}

func TestShouldIncludeField_ExcludeContext_Matches(t *testing.T) {
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "password",
		ExcludeContexts: map[string]bool{"response": true, "log": true},
		OmitZero:        false,
		OmitEmpty:       false,
	}

	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
	}

	fieldValue := reflect.ValueOf("secret123")
	result := ShouldIncludeField(meta, fieldValue, opts)

	assert.False(t, result, "field should be excluded in 'response' context")
}

func TestShouldIncludeField_ExcludeContext_NoMatch(t *testing.T) {
	meta := FieldMetadata{
		FieldIndex:      0,
		JSONName:        "password",
		ExcludeContexts: map[string]bool{"log": true},
		OmitZero:        false,
		OmitEmpty:       false,
	}

	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
	}

	fieldValue := reflect.ValueOf("secret123")
	result := ShouldIncludeField(meta, fieldValue, opts)

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
				JSONName: "port",
				OmitZero: true,
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
				JSONName: "name",
				OmitZero: true,
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
				JSONName: "enabled",
				OmitZero: true,
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
			result := ShouldIncludeField(tt.meta, fieldValue, tt.opts)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestShouldIncludeField_OmitZero_NonZeroValue(t *testing.T) {
	meta := FieldMetadata{
		JSONName: "port",
		OmitZero: true,
	}

	opts := SerializeOptions{
		OmitZero: true,
	}

	fieldValue := reflect.ValueOf(8080)
	result := ShouldIncludeField(meta, fieldValue, opts)

	assert.True(t, result, "non-zero value should be included even with omitzero")
}

func TestShouldIncludeField_OmitZero_NilPointer(t *testing.T) {
	meta := FieldMetadata{
		JSONName: "count",
		OmitZero: true,
	}

	opts := SerializeOptions{
		OmitZero: true,
	}

	var ptr *int
	fieldValue := reflect.ValueOf(ptr)
	result := ShouldIncludeField(meta, fieldValue, opts)

	assert.False(t, result, "nil pointer should be omitted with omitzero")
}

func TestShouldIncludeField_OmitZero_NonNilPointer(t *testing.T) {
	meta := FieldMetadata{
		JSONName: "count",
		OmitZero: true,
	}

	opts := SerializeOptions{
		OmitZero: true,
	}

	val := 42
	fieldValue := reflect.ValueOf(&val)
	result := ShouldIncludeField(meta, fieldValue, opts)

	assert.True(t, result, "non-nil pointer should be included")
}

func TestShouldIncludeField_OmitZero_Disabled(t *testing.T) {
	meta := FieldMetadata{
		JSONName: "port",
		OmitZero: true,
	}

	opts := SerializeOptions{
		OmitZero: false, // OmitZero disabled in options
	}

	fieldValue := reflect.ValueOf(0)
	result := ShouldIncludeField(meta, fieldValue, opts)

	assert.True(t, result, "zero value should be included when OmitZero is disabled in options")
}

func TestShouldIncludeField_CombinedExcludeAndOmitZero(t *testing.T) {
	// Field with both exclude context and omitzero
	meta := FieldMetadata{
		JSONName:        "internal",
		ExcludeContexts: map[string]bool{"response": true},
		OmitZero:        true,
	}

	// Test 1: Excluded by context (should exclude regardless of zero)
	opts1 := SerializeOptions{
		Context:  "response",
		OmitZero: true,
	}
	fieldValue1 := reflect.ValueOf("value")
	assert.False(t, ShouldIncludeField(meta, fieldValue1, opts1), "should be excluded by context")

	// Test 2: Not excluded by context, but zero value (should exclude by omitzero)
	opts2 := SerializeOptions{
		Context:  "api",
		OmitZero: true,
	}
	fieldValue2 := reflect.ValueOf("")
	assert.False(t, ShouldIncludeField(meta, fieldValue2, opts2), "should be excluded by omitzero")

	// Test 3: Not excluded by context, non-zero value (should include)
	opts3 := SerializeOptions{
		Context:  "api",
		OmitZero: true,
	}
	fieldValue3 := reflect.ValueOf("value")
	assert.True(t, ShouldIncludeField(meta, fieldValue3, opts3), "should be included")
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

	metadata := BuildFieldMetadata(reflect.TypeOf(obj))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(user))
	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(user))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: true, // OmitZero enabled
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

	metadata := BuildFieldMetadata(reflect.TypeOf(nested))
	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(nested))
	opts := SerializeOptions{
		Context:  "response",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(TestUser{}))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(*user))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(config))

	// Test "response" context
	optsResponse := SerializeOptions{
		Context:  "response",
		OmitZero: false,
	}
	resultResponse := ToFilteredMap(reflect.ValueOf(config), metadata, optsResponse)

	assert.Equal(t, "localhost", resultResponse["host"])
	assert.NotContains(t, resultResponse, "api_key")  // Excluded in "response"
	assert.NotContains(t, resultResponse, "internal") // Excluded in "response"

	// Test "log" context
	optsLog := SerializeOptions{
		Context:  "log",
		OmitZero: false,
	}
	resultLog := ToFilteredMap(reflect.ValueOf(config), metadata, optsLog)

	assert.Equal(t, "localhost", resultLog["host"])
	assert.Contains(t, resultLog, "api_key")     // NOT excluded in "log"
	assert.NotContains(t, resultLog, "internal") // Excluded in "log"

	// Test no context
	optsNone := SerializeOptions{
		Context:  "",
		OmitZero: false,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(config))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: true,
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

	metadata := BuildFieldMetadata(reflect.TypeOf(config))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: true,
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
	metadata := BuildFieldMetadata(reflect.TypeOf(obj))
	opts := SerializeOptions{
		Context:  "",
		OmitZero: false,
	}

	result := ToFilteredMap(reflect.ValueOf(obj), metadata, opts)

	// Should return empty map
	assert.NotNil(t, result)
	assert.Empty(t, result)
}
