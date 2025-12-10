package pedantigo

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================================================
// Schema() generation tests - Table-driven
// ==================================================

// TestSchema_TypeMapping verifies correct JSON Schema type generation for Go types
func TestSchema_TypeMapping(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (interface{}, *jsonschema.Schema) // Returns validator and expected properties
		validate func(*testing.T, *jsonschema.Schema)
	}{
		{
			name: "basic types (string, int)",
			setup: func() (interface{}, *jsonschema.Schema) {
				type User struct {
					Name  string `json:"name"`
					Age   int    `json:"age"`
					Email string `json:"email"`
				}
				v := New[User]()
				return v, nil
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "object", schema.Type)
				require.NotNil(t, schema.Properties)

				nameProp, _ := schema.Properties.Get("name")
				require.NotNil(t, nameProp)
				assert.Equal(t, "string", nameProp.Type)

				ageProp, _ := schema.Properties.Get("age")
				require.NotNil(t, ageProp)
				assert.Equal(t, "integer", ageProp.Type)
			},
		},
		{
			name: "nested struct",
			setup: func() (interface{}, *jsonschema.Schema) {
				type Address struct {
					City string `json:"city" pedantigo:"required"`
					Zip  string `json:"zip" pedantigo:"min=5"`
				}
				type User struct {
					Name    string  `json:"name" pedantigo:"required"`
					Address Address `json:"address"`
				}
				v := New[User]()
				return v, nil
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				addressProp, _ := schema.Properties.Get("address")
				require.NotNil(t, addressProp)
				assert.Equal(t, "object", addressProp.Type)

				cityProp, _ := addressProp.Properties.Get("city")
				require.NotNil(t, cityProp)

				hasRequiredCity := false
				for _, req := range addressProp.Required {
					if req == "city" {
						hasRequiredCity = true
						break
					}
				}
				assert.True(t, hasRequiredCity, "expected 'city' to be required in nested struct")
			},
		},
		{
			name: "slice with item constraints",
			setup: func() (interface{}, *jsonschema.Schema) {
				type Config struct {
					Tags   []string `json:"tags" pedantigo:"min=3"`
					Admins []string `json:"admins" pedantigo:"email"`
				}
				v := New[Config]()
				return v, nil
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				tagsProp, _ := schema.Properties.Get("tags")
				require.NotNil(t, tagsProp)
				assert.Equal(t, "array", tagsProp.Type)
				require.NotNil(t, tagsProp.Items)
				assert.Equal(t, "string", tagsProp.Items.Type)
				require.NotNil(t, tagsProp.Items.MinLength)
				assert.Equal(t, uint64(3), *tagsProp.Items.MinLength)

				adminsProp, _ := schema.Properties.Get("admins")
				require.NotNil(t, adminsProp)
				require.NotNil(t, adminsProp.Items)
				assert.Equal(t, "email", adminsProp.Items.Format)
			},
		},
		{
			name: "map with value constraints",
			setup: func() (interface{}, *jsonschema.Schema) {
				type Config struct {
					Settings map[string]string `json:"settings" pedantigo:"min=3"`
					Contacts map[string]string `json:"contacts" pedantigo:"email"`
				}
				v := New[Config]()
				return v, nil
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				settingsProp, _ := schema.Properties.Get("settings")
				require.NotNil(t, settingsProp)
				assert.Equal(t, "object", settingsProp.Type)
				require.NotNil(t, settingsProp.AdditionalProperties)
				require.NotNil(t, settingsProp.AdditionalProperties.MinLength)
				assert.Equal(t, uint64(3), *settingsProp.AdditionalProperties.MinLength)

				contactsProp, _ := schema.Properties.Get("contacts")
				require.NotNil(t, contactsProp)
				require.NotNil(t, contactsProp.AdditionalProperties)
				assert.Equal(t, "email", contactsProp.AdditionalProperties.Format)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup, _ := tt.setup()
			// Type assertion to get validator and call Schema()
			switch v := setup.(type) {
			case interface{ Schema() *jsonschema.Schema }:
				schema := v.Schema()
				require.NotNil(t, schema)
				tt.validate(t, schema)
			default:
				require.Fail(t, "invalid validator type")
			}
		})
	}
}

// TestSchema_Constraints verifies constraint mapping to JSON Schema keywords
func TestSchema_Constraints(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{}
		validate func(*testing.T, *jsonschema.Schema)
	}{
		{
			name: "required fields",
			setup: func() interface{} {
				type User struct {
					Name  string `json:"name" pedantigo:"required"`
					Email string `json:"email" pedantigo:"required"`
					Age   int    `json:"age"`
				}
				return New[User]()
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Len(t, schema.Required, 2)
				requiredMap := make(map[string]bool)
				for _, field := range schema.Required {
					requiredMap[field] = true
				}
				assert.True(t, requiredMap["name"], "expected 'name' to be required")
				assert.True(t, requiredMap["email"], "expected 'email' to be required")
			},
		},
		{
			name: "numeric constraints (gt/lt/gte/lte/min/max)",
			setup: func() interface{} {
				type Product struct {
					Price    float64 `json:"price" pedantigo:"gt=0,lt=10000"`
					Stock    int     `json:"stock" pedantigo:"gte=0,lte=1000"`
					Discount int     `json:"discount" pedantigo:"min=0,max=100"`
				}
				return New[Product]()
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				priceProp, _ := schema.Properties.Get("price")
				assert.Equal(t, "0", string(priceProp.ExclusiveMinimum))
				assert.Equal(t, "10000", string(priceProp.ExclusiveMaximum))

				stockProp, _ := schema.Properties.Get("stock")
				assert.Equal(t, "0", string(stockProp.Minimum))
				assert.Equal(t, "1000", string(stockProp.Maximum))

				discountProp, _ := schema.Properties.Get("discount")
				assert.Equal(t, "0", string(discountProp.Minimum))
				assert.Equal(t, "100", string(discountProp.Maximum))
			},
		},
		{
			name: "string length constraints (min/max)",
			setup: func() interface{} {
				type User struct {
					Username string `json:"username" pedantigo:"min=3,max=20"`
					Bio      string `json:"bio" pedantigo:"max=500"`
				}
				return New[User]()
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				usernameProp, _ := schema.Properties.Get("username")
				require.NotNil(t, usernameProp.MinLength)
				assert.Equal(t, uint64(3), *usernameProp.MinLength)
				require.NotNil(t, usernameProp.MaxLength)
				assert.Equal(t, uint64(20), *usernameProp.MaxLength)

				bioProp, _ := schema.Properties.Get("bio")
				require.NotNil(t, bioProp.MaxLength)
				assert.Equal(t, uint64(500), *bioProp.MaxLength)
			},
		},
		{
			name: "format constraints (email, url, uuid, ipv4, ipv6)",
			setup: func() interface{} {
				type Contact struct {
					Email    string `json:"email" pedantigo:"email"`
					Website  string `json:"website" pedantigo:"url"`
					ID       string `json:"id" pedantigo:"uuid"`
					IPv4Addr string `json:"ipv4" pedantigo:"ipv4"`
					IPv6Addr string `json:"ipv6" pedantigo:"ipv6"`
				}
				return New[Contact]()
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				tests := []struct {
					field       string
					expectedFmt string
				}{
					{"email", "email"},
					{"website", "uri"}, // url constraint → uri format
					{"id", "uuid"},
					{"ipv4", "ipv4"},
					{"ipv6", "ipv6"},
				}
				for _, tt := range tests {
					prop, _ := schema.Properties.Get(tt.field)
					require.NotNil(t, prop, "expected field '%s' to exist", tt.field)
					assert.Equal(t, tt.expectedFmt, prop.Format, "field '%s'", tt.field)
				}
			},
		},
		{
			name: "regex pattern constraint",
			setup: func() interface{} {
				type Code struct {
					ZipCode string `json:"zipCode" pedantigo:"regexp=^[0-9]{5}$"`
				}
				return New[Code]()
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				zipProp, _ := schema.Properties.Get("zipCode")
				assert.Equal(t, "^[0-9]{5}$", zipProp.Pattern)
			},
		},
		{
			name: "default values (int, string, bool)",
			setup: func() interface{} {
				type Config struct {
					Port    int    `json:"port" pedantigo:"default=8080"`
					Host    string `json:"host" pedantigo:"default=localhost"`
					Enabled bool   `json:"enabled" pedantigo:"default=true"`
				}
				return New[Config]()
			},
			validate: func(t *testing.T, schema *jsonschema.Schema) {
				portProp, _ := schema.Properties.Get("port")
				portDefault, _ := json.Marshal(portProp.Default)
				assert.Equal(t, "8080", string(portDefault))

				hostProp, _ := schema.Properties.Get("host")
				assert.Equal(t, "localhost", hostProp.Default)

				enabledProp, _ := schema.Properties.Get("enabled")
				enabledDefault, _ := json.Marshal(enabledProp.Default)
				assert.Equal(t, "true", string(enabledDefault))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			switch validator := v.(type) {
			case interface{ Schema() *jsonschema.Schema }:
				schema := validator.Schema()
				require.NotNil(t, schema)
				tt.validate(t, schema)
			default:
				require.Fail(t, "invalid validator type")
			}
		})
	}
}

// ==================================================
// JSON Serialization tests (Schema/SchemaJSON/SchemaOpenAPI) - Table-driven
// ==================================================

// TestSchemaJSON_Serialization verifies JSON serialization methods and OpenAPI references
func TestSchemaJSON_Serialization(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{}
		validate func(*testing.T, interface{})
	}{
		{
			name: "SchemaJSON produces valid JSON",
			setup: func() interface{} {
				type User struct {
					Name  string `json:"name" pedantigo:"required,min=3"`
					Email string `json:"email" pedantigo:"required,email"`
					Age   int    `json:"age" pedantigo:"gte=18,lte=120"`
				}
				return New[User]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator, ok := v.(interface{ SchemaJSON() ([]byte, error) })
				require.True(t, ok, "validator missing SchemaJSON method")

				jsonBytes, err := validator.SchemaJSON()
				require.NoError(t, err)

				var schemaMap map[string]any
				err = json.Unmarshal(jsonBytes, &schemaMap)
				require.NoError(t, err)

				assert.Equal(t, "object", schemaMap["type"])

				properties, ok := schemaMap["properties"].(map[string]any)
				require.True(t, ok, "expected properties to be an object")

				// Check name field has minLength
				nameField, ok := properties["name"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, float64(3), nameField["minLength"])

				// Check email field has format
				emailField, ok := properties["email"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "email", emailField["format"])
			},
		},
		{
			name: "SchemaJSONOpenAPI with nested references and $defs",
			setup: func() interface{} {
				type Address struct {
					City string `json:"city" pedantigo:"required"`
					Zip  string `json:"zip" pedantigo:"min=5"`
				}
				type User struct {
					Name    string  `json:"name" pedantigo:"required,min=3"`
					Address Address `json:"address" pedantigo:"required"`
				}
				return New[User]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator, ok := v.(interface{ SchemaJSONOpenAPI() ([]byte, error) })
				require.True(t, ok, "validator missing SchemaJSONOpenAPI method")

				jsonBytes, err := validator.SchemaJSONOpenAPI()
				require.NoError(t, err)

				var schemaMap map[string]any
				err = json.Unmarshal(jsonBytes, &schemaMap)
				require.NoError(t, err)

				// Check that $defs exists
				defs, hasDefs := schemaMap["$defs"].(map[string]any)
				require.True(t, hasDefs, "expected $defs to exist in OpenAPI schema")

				// Check Address definition exists
				addressDef, ok := defs["Address"].(map[string]any)
				require.True(t, ok, "expected Address definition in $defs")

				// Check Address has required city
				addressRequired, ok := addressDef["required"].([]any)
				require.True(t, ok, "expected 'required' array in Address definition")
				hasCity := false
				for _, req := range addressRequired {
					if req == "city" {
						hasCity = true
						break
					}
				}
				assert.True(t, hasCity, "expected 'city' to be required in Address definition")

				// Check zip constraint in Address definition
				addressProps, ok := addressDef["properties"].(map[string]any)
				require.True(t, ok, "expected 'properties' in Address definition")
				zipProp, ok := addressProps["zip"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, float64(5), zipProp["minLength"])

				// Check root schema has $ref to Address
				properties, ok := schemaMap["properties"].(map[string]any)
				require.True(t, ok, "expected 'properties' in root schema")
				addressProp, ok := properties["address"].(map[string]any)
				require.True(t, ok, "expected 'address' property in root schema")
				ref, hasRef := addressProp["$ref"].(string)
				require.True(t, hasRef)
				assert.Equal(t, "#/$defs/Address", ref)
			},
		},
		{
			name: "SchemaOpenAPI returns nested definitions with constraints",
			setup: func() interface{} {
				type Contact struct {
					Email string `json:"email" pedantigo:"required,email"`
					Phone string `json:"phone" pedantigo:"min=10"`
				}
				type Company struct {
					Name    string  `json:"name" pedantigo:"required,min=3"`
					Contact Contact `json:"contact" pedantigo:"required"`
				}
				return New[Company]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator, ok := v.(interface{ SchemaOpenAPI() *jsonschema.Schema })
				require.True(t, ok, "validator missing SchemaOpenAPI method")

				schema := validator.SchemaOpenAPI()
				require.Greater(t, len(schema.Definitions), 0, "expected schema to have definitions")

				contactDef, ok := schema.Definitions["Contact"]
				require.True(t, ok, "expected Contact definition")

				// Check Contact has required email
				hasEmail := false
				for _, req := range contactDef.Required {
					if req == "email" {
						hasEmail = true
						break
					}
				}
				assert.True(t, hasEmail, "expected 'email' to be required in Contact definition")

				// Check constraints in Contact definition
				emailProp, _ := contactDef.Properties.Get("email")
				require.NotNil(t, emailProp)
				assert.Equal(t, "email", emailProp.Format)

				phoneProp, _ := contactDef.Properties.Get("phone")
				require.NotNil(t, phoneProp)
				require.NotNil(t, phoneProp.MinLength)
				assert.Equal(t, uint64(10), *phoneProp.MinLength)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			tt.validate(t, v)
		})
	}
}

// ==================================================
// Schema caching and concurrency tests - Table-driven
// ==================================================

// TestSchema_Caching verifies single-validator schema caching works correctly
func TestSchema_Caching(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{}
		validate func(*testing.T, interface{})
	}{
		{
			name: "Schema() caches pointer on repeated calls",
			setup: func() interface{} {
				type Product struct {
					Name  string  `json:"name" pedantigo:"required,min=3"`
					Price float64 `json:"price" pedantigo:"gt=0"`
				}
				return New[Product]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator := v.(interface{ Schema() *jsonschema.Schema })
				schema1 := validator.Schema()
				schema2 := validator.Schema()
				require.NotNil(t, schema1)
				require.NotNil(t, schema2)
				assert.Same(t, schema1, schema2)
			},
		},
		{
			name: "SchemaJSON() caches bytes on repeated calls",
			setup: func() interface{} {
				type Config struct {
					Host string `json:"host" pedantigo:"required,min=1"`
					Port int    `json:"port" pedantigo:"gt=0,lt=65536"`
				}
				return New[Config]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator := v.(interface{ SchemaJSON() ([]byte, error) })
				json1, err1 := validator.SchemaJSON()
				json2, err2 := validator.SchemaJSON()

				require.NoError(t, err1)
				require.NoError(t, err2)
				assert.Len(t, json1, len(json2))
				assert.True(t, bytesEqual(json1, json2), "expected SchemaJSON() to return identical cached bytes")
			},
		},
		{
			name: "SchemaOpenAPI() caches pointer on repeated calls",
			setup: func() interface{} {
				type Item struct {
					ID    string `json:"id" pedantigo:"required,uuid"`
					Title string `json:"title" pedantigo:"required,min=5"`
				}
				return New[Item]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator := v.(interface{ SchemaOpenAPI() *jsonschema.Schema })
				openapi1 := validator.SchemaOpenAPI()
				openapi2 := validator.SchemaOpenAPI()
				require.NotNil(t, openapi1)
				require.NotNil(t, openapi2)
				assert.Same(t, openapi1, openapi2)
			},
		},
		{
			name: "SchemaJSONOpenAPI() caches bytes on repeated calls",
			setup: func() interface{} {
				type Event struct {
					Name      string `json:"name" pedantigo:"required,min=1"`
					Timestamp int64  `json:"timestamp" pedantigo:"required,gte=0"`
				}
				return New[Event]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator := v.(interface{ SchemaJSONOpenAPI() ([]byte, error) })
				json1, err1 := validator.SchemaJSONOpenAPI()
				json2, err2 := validator.SchemaJSONOpenAPI()

				require.NoError(t, err1)
				require.NoError(t, err2)
				assert.Len(t, json1, len(json2))
				assert.True(t, bytesEqual(json1, json2), "expected SchemaJSONOpenAPI() to return identical cached bytes")
			},
		},
		{
			name: "independent validators have independent caches",
			setup: func() interface{} {
				type Cat struct {
					Name string `json:"name" pedantigo:"required"`
				}
				type Dog struct {
					Name string `json:"name" pedantigo:"required"`
				}
				// Return tuple of two validators
				return struct {
					cat interface{ Schema() *jsonschema.Schema }
					dog interface{ Schema() *jsonschema.Schema }
				}{New[Cat](), New[Dog]()}
			},
			validate: func(t *testing.T, v interface{}) {
				pair := v.(struct {
					cat interface{ Schema() *jsonschema.Schema }
					dog interface{ Schema() *jsonschema.Schema }
				})
				catSchema1 := pair.cat.Schema()
				dogSchema1 := pair.dog.Schema()
				catSchema2 := pair.cat.Schema()
				dogSchema2 := pair.dog.Schema()

				// Each validator caches its own
				assert.Same(t, catSchema1, catSchema2)
				assert.Same(t, dogSchema1, dogSchema2)
				// But different validators have different caches
				assert.NotSame(t, catSchema1, dogSchema1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			tt.validate(t, v)
		})
	}
}

// TestSchema_ConcurrencySafe verifies schema generation is thread-safe under concurrent access
func TestSchema_ConcurrencySafe(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{}
		validate func(*testing.T, interface{})
	}{
		{
			name: "Schema() is thread-safe with 100 concurrent goroutines",
			setup: func() interface{} {
				type User struct {
					Name  string `json:"name" pedantigo:"required"`
					Email string `json:"email" pedantigo:"required,email"`
				}
				return New[User]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator := v.(interface{ Schema() *jsonschema.Schema })
				numGoroutines := 100

				var wg sync.WaitGroup
				wg.Add(numGoroutines)
				schemaChan := make(chan *jsonschema.Schema, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					go func() {
						defer wg.Done()
						schemaChan <- validator.Schema()
					}()
				}

				wg.Wait()
				close(schemaChan)

				// Verify all concurrent calls returned same cached pointer
				pointers := make([]*jsonschema.Schema, 0, numGoroutines)
				for schema := range schemaChan {
					require.NotNil(t, schema)
					pointers = append(pointers, schema)
				}

				firstPtr := pointers[0]
				for i, ptr := range pointers {
					assert.Same(t, firstPtr, ptr, "goroutine %d", i)
				}
			},
		},
		{
			name: "SchemaJSON() is thread-safe with 100 concurrent goroutines",
			setup: func() interface{} {
				type Settings struct {
					Timeout int `json:"timeout" pedantigo:"gt=0,lt=60000"`
					Retries int `json:"retries" pedantigo:"gte=0,lte=10"`
				}
				return New[Settings]()
			},
			validate: func(t *testing.T, v interface{}) {
				validator := v.(interface{ SchemaJSON() ([]byte, error) })
				numGoroutines := 100

				var wg sync.WaitGroup
				wg.Add(numGoroutines)
				jsonChan := make(chan []byte, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					go func() {
						defer wg.Done()
						jsonBytes, err := validator.SchemaJSON()
						require.NoError(t, err)
						jsonChan <- jsonBytes
					}()
				}

				wg.Wait()
				close(jsonChan)

				// Verify all concurrent calls returned same cached bytes
				allBytes := make([][]byte, 0, numGoroutines)
				for jsonBytes := range jsonChan {
					require.NotNil(t, jsonBytes)
					allBytes = append(allBytes, jsonBytes)
				}

				firstBytes := allBytes[0]
				for i, jsonBytes := range allBytes {
					assert.True(t, bytesEqual(jsonBytes, firstBytes), "goroutine %d", i)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			tt.validate(t, v)
		})
	}
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ==================================================
// findTypeForDefinition, searchSliceType, searchMapType coverage tests
// ==================================================

// TestSchemaOpenAPI_SliceOfStructs tests schema generation with slices of structs
// This exercises searchSliceType() code path (currently 0% coverage)
// TestSchemaOpenAPI_SliceOfStructs tests SchemaOpenAPI sliceofstructs
func TestSchemaOpenAPI_SliceOfStructs(t *testing.T) {
	type Author struct {
		Name  string `json:"name" pedantigo:"required,min=2"`
		Email string `json:"email" pedantigo:"email"`
	}

	type Book struct {
		Title   string   `json:"title" pedantigo:"required,min=1"`
		Authors []Author `json:"authors" pedantigo:"required"`
	}

	validator := New[Book]()
	schema := validator.SchemaOpenAPI()

	// Should have Author in definitions
	authorDef, hasAuthor := schema.Definitions["Author"]
	require.True(t, hasAuthor, "expected Author definition in $defs")

	// Verify Author definition has constraints from pedantigo tags
	hasNameRequired := false
	for _, req := range authorDef.Required {
		if req == "name" {
			hasNameRequired = true
			break
		}
	}
	assert.True(t, hasNameRequired, "expected 'name' to be required in Author definition")

	nameProp, _ := authorDef.Properties.Get("name")
	require.NotNil(t, nameProp)
	require.NotNil(t, nameProp.MinLength)
	assert.Equal(t, uint64(2), *nameProp.MinLength)

	emailProp, _ := authorDef.Properties.Get("email")
	require.NotNil(t, emailProp)
	assert.Equal(t, "email", emailProp.Format)
}

// TestSchemaOpenAPI_PointerSliceOfStructs tests schema with pointer slices
// This exercises searchSliceType() with pointer unwrapping
// TestSchemaOpenAPI_PointerSliceOfStructs tests SchemaOpenAPI pointersliceofstructs
func TestSchemaOpenAPI_PointerSliceOfStructs(t *testing.T) {
	type Tag struct {
		Name  string `json:"name" pedantigo:"required,min=1"`
		Color string `json:"color" pedantigo:"regexp=^#[0-9a-fA-F]{6}$"`
	}

	type Article struct {
		Title string `json:"title" pedantigo:"required"`
		Tags  []*Tag `json:"tags"` // Pointer slice
	}

	validator := New[Article]()
	schema := validator.SchemaOpenAPI()

	// Should have Tag in definitions even though it's []*Tag
	tagDef, hasTag := schema.Definitions["Tag"]
	require.True(t, hasTag, "expected Tag definition in $defs (pointer slice)")

	// Verify Tag constraints are applied
	colorProp, _ := tagDef.Properties.Get("color")
	require.NotNil(t, colorProp)
	assert.Equal(t, "^#[0-9a-fA-F]{6}$", colorProp.Pattern)
}

// TestSchemaOpenAPI_MapOfStructs tests schema generation with maps of structs
// This exercises searchMapType() code path (currently 0% coverage)
// TestSchemaOpenAPI_MapOfStructs tests SchemaOpenAPI mapofstructs
func TestSchemaOpenAPI_MapOfStructs(t *testing.T) {
	type Contact struct {
		Email string `json:"email" pedantigo:"required,email"`
		Phone string `json:"phone" pedantigo:"min=10,max=15"`
	}

	type Company struct {
		Name     string             `json:"name" pedantigo:"required,min=1"`
		Contacts map[string]Contact `json:"contacts"`
	}

	validator := New[Company]()
	schema := validator.SchemaOpenAPI()

	// Should have Contact in definitions
	contactDef, hasContact := schema.Definitions["Contact"]
	require.True(t, hasContact, "expected Contact definition in $defs")

	// Verify Contact definition has constraints
	hasEmailRequired := false
	for _, req := range contactDef.Required {
		if req == "email" {
			hasEmailRequired = true
			break
		}
	}
	assert.True(t, hasEmailRequired, "expected 'email' to be required in Contact definition")

	emailProp, _ := contactDef.Properties.Get("email")
	require.NotNil(t, emailProp)
	assert.Equal(t, "email", emailProp.Format)

	phoneProp, _ := contactDef.Properties.Get("phone")
	require.NotNil(t, phoneProp)
	require.NotNil(t, phoneProp.MinLength)
	assert.Equal(t, uint64(10), *phoneProp.MinLength)
}

// TestSchemaOpenAPI_PointerMapOfStructs tests schema with pointer map values
// This exercises searchMapType() with pointer unwrapping
// TestSchemaOpenAPI_PointerMapOfStructs tests SchemaOpenAPI pointermapofstructs
func TestSchemaOpenAPI_PointerMapOfStructs(t *testing.T) {
	type Address struct {
		Street  string `json:"street" pedantigo:"required,min=1"`
		City    string `json:"city" pedantigo:"required,min=2"`
		ZipCode string `json:"zipCode" pedantigo:"regexp=^[0-9]{5}$"`
	}

	type Organization struct {
		Name      string              `json:"name" pedantigo:"required"`
		Locations map[string]*Address `json:"locations"` // Pointer map values
	}

	validator := New[Organization]()
	schema := validator.SchemaOpenAPI()

	// Should have Address in definitions even though it's map[string]*Address
	addressDef, hasAddress := schema.Definitions["Address"]
	require.True(t, hasAddress, "expected Address definition in $defs (pointer map values)")

	// Verify Address constraints are applied
	zipProp, _ := addressDef.Properties.Get("zipCode")
	require.NotNil(t, zipProp)
	assert.Equal(t, "^[0-9]{5}$", zipProp.Pattern)

	cityProp, _ := addressDef.Properties.Get("city")
	require.NotNil(t, cityProp)
	require.NotNil(t, cityProp.MinLength)
	assert.Equal(t, uint64(2), *cityProp.MinLength)
}

// TestSchemaOpenAPI_NestedStructInSlice tests deeply nested struct in slice
// This exercises recursive findTypeForDefinition through searchSliceType
// TestSchemaOpenAPI_NestedStructInSlice tests SchemaOpenAPI nestedstructinslice
func TestSchemaOpenAPI_NestedStructInSlice(t *testing.T) {
	type Permission struct {
		Name string `json:"name" pedantigo:"required,min=1"`
	}

	type Role struct {
		Title       string       `json:"title" pedantigo:"required,min=1"`
		Permissions []Permission `json:"permissions"`
	}

	type User struct {
		Username string `json:"username" pedantigo:"required,min=3"`
		Roles    []Role `json:"roles"`
	}

	validator := New[User]()
	schema := validator.SchemaOpenAPI()

	// Should have both Role and Permission in definitions
	roleDef, hasRole := schema.Definitions["Role"]
	require.True(t, hasRole, "expected Role definition in $defs")

	permDef, hasPerm := schema.Definitions["Permission"]
	require.True(t, hasPerm, "expected Permission definition in $defs (nested in slice)")

	// Verify Permission constraints applied
	nameProp, _ := permDef.Properties.Get("name")
	require.NotNil(t, nameProp)
	require.NotNil(t, nameProp.MinLength)
	assert.Equal(t, uint64(1), *nameProp.MinLength)

	// Verify Role constraints applied
	titleProp, _ := roleDef.Properties.Get("title")
	require.NotNil(t, titleProp)
	require.NotNil(t, titleProp.MinLength)
	assert.Equal(t, uint64(1), *titleProp.MinLength)
}

// TestSchemaOpenAPI_NestedStructInMap tests deeply nested struct in map
// This exercises recursive findTypeForDefinition through searchMapType
// TestSchemaOpenAPI_NestedStructInMap tests SchemaOpenAPI nestedstructinmap
func TestSchemaOpenAPI_NestedStructInMap(t *testing.T) {
	type Metadata struct {
		Key   string `json:"key" pedantigo:"required,min=1"`
		Value string `json:"value" pedantigo:"required"`
	}

	type Resource struct {
		Name     string              `json:"name" pedantigo:"required,min=1"`
		Metadata map[string]Metadata `json:"metadata"`
	}

	type Project struct {
		Title     string              `json:"title" pedantigo:"required,min=1"`
		Resources map[string]Resource `json:"resources"`
	}

	validator := New[Project]()
	schema := validator.SchemaOpenAPI()

	// Should have Resource and Metadata in definitions
	resourceDef, hasResource := schema.Definitions["Resource"]
	require.True(t, hasResource, "expected Resource definition in $defs")

	metadataDef, hasMetadata := schema.Definitions["Metadata"]
	require.True(t, hasMetadata, "expected Metadata definition in $defs (nested in map)")

	// Verify Metadata constraints applied
	keyProp, _ := metadataDef.Properties.Get("key")
	require.NotNil(t, keyProp)
	require.NotNil(t, keyProp.MinLength)
	assert.Equal(t, uint64(1), *keyProp.MinLength)

	// Verify Resource constraints applied
	nameProp, _ := resourceDef.Properties.Get("name")
	require.NotNil(t, nameProp)
	require.NotNil(t, nameProp.MinLength)
	assert.Equal(t, uint64(1), *nameProp.MinLength)
}

// TestSchemaOpenAPI_DirectTypeMatch tests findTypeForDefinition direct name matching
func TestSchemaOpenAPI_DirectTypeMatch(t *testing.T) {
	type Address struct {
		Street string `json:"street" pedantigo:"required"`
		City   string `json:"city" pedantigo:"required,min=2"`
	}

	type Person struct {
		Name    string  `json:"name" pedantigo:"required"`
		Address Address `json:"address" pedantigo:"required"`
	}

	validator := New[Person]()
	schema := validator.SchemaOpenAPI()

	// Should have Address in definitions
	addressDef, hasAddress := schema.Definitions["Address"]
	require.True(t, hasAddress, "expected Address definition in $defs")

	// Verify Address definition has constraints
	streetProp, _ := addressDef.Properties.Get("street")
	require.NotNil(t, streetProp)

	cityProp, _ := addressDef.Properties.Get("city")
	require.NotNil(t, cityProp)
	require.NotNil(t, cityProp.MinLength)
	assert.Equal(t, uint64(2), *cityProp.MinLength)
}

// TestSchemaOpenAPI_PointerFieldType tests findTypeForDefinition with pointer field types
func TestSchemaOpenAPI_PointerFieldType(t *testing.T) {
	type Config struct {
		Key   string `json:"key" pedantigo:"required"`
		Value string `json:"value" pedantigo:"min=1"`
	}

	type Service struct {
		Name   string  `json:"name" pedantigo:"required"`
		Config *Config `json:"config"` // Pointer to nested struct
	}

	validator := New[Service]()
	schema := validator.SchemaOpenAPI()

	// Should have Config in definitions (pointer should be unwrapped)
	configDef, hasConfig := schema.Definitions["Config"]
	require.True(t, hasConfig, "expected Config definition in $defs (pointer should be unwrapped)")

	// Verify Config definition has constraints
	keyProp, _ := configDef.Properties.Get("key")
	require.NotNil(t, keyProp)

	valueProp, _ := configDef.Properties.Get("value")
	require.NotNil(t, valueProp)
	require.NotNil(t, valueProp.MinLength)
	assert.Equal(t, uint64(1), *valueProp.MinLength)
}

// TestSchemaOpenAPI_DeeplyNestedStruct tests findTypeForDefinition recursive search
func TestSchemaOpenAPI_DeeplyNestedStruct(t *testing.T) {
	type Level3 struct {
		Data string `json:"data" pedantigo:"required,min=5"`
	}

	type Level2 struct {
		Info   string `json:"info" pedantigo:"required"`
		Nested Level3 `json:"nested"`
	}

	type Level1 struct {
		Title string `json:"title" pedantigo:"required"`
		Mid   Level2 `json:"mid"`
	}

	validator := New[Level1]()
	schema := validator.SchemaOpenAPI()

	// Should have all levels in definitions
	_, hasLevel2 := schema.Definitions["Level2"]
	assert.True(t, hasLevel2, "expected Level2 definition in $defs")

	level3Def, hasLevel3 := schema.Definitions["Level3"]
	require.True(t, hasLevel3, "expected Level3 definition in $defs (deeply nested)")

	// Verify Level3 definition has constraints
	dataProp, _ := level3Def.Properties.Get("data")
	require.NotNil(t, dataProp)
	require.NotNil(t, dataProp.MinLength)
	assert.Equal(t, uint64(5), *dataProp.MinLength)
}

// TestSchemaOpenAPI_MixedNestedTypes tests all search paths together
func TestSchemaOpenAPI_MixedNestedTypes(t *testing.T) {
	type Tag struct {
		Name string `json:"name" pedantigo:"required,min=1"`
	}

	type Metadata struct {
		Key string `json:"key" pedantigo:"required"`
	}

	type Comment struct {
		Text string `json:"text" pedantigo:"required,min=3"`
	}

	type Article struct {
		Title    string              `json:"title" pedantigo:"required"`
		Tags     []Tag               `json:"tags"`     // Slice of structs
		Meta     map[string]Metadata `json:"meta"`     // Map of structs
		Comments []Comment           `json:"comments"` // Another slice
		Author   *Tag                `json:"author"`   // Pointer to struct
	}

	validator := New[Article]()
	schema := validator.SchemaOpenAPI()

	// Should have all nested types in definitions
	tagDef, hasTag := schema.Definitions["Tag"]
	assert.True(t, hasTag, "expected Tag definition")
	assert.Greater(t, tagDef.Properties.Len(), 0, "expected Tag definition to have properties")

	metaDef, hasMeta := schema.Definitions["Metadata"]
	assert.True(t, hasMeta, "expected Metadata definition from map values")
	assert.Greater(t, metaDef.Properties.Len(), 0, "expected Metadata definition to have properties")

	commentDef, hasComment := schema.Definitions["Comment"]
	assert.True(t, hasComment, "expected Comment definition from slice")

	// Verify Comment definition has constraints
	textProp, _ := commentDef.Properties.Get("text")
	require.NotNil(t, textProp)
	require.NotNil(t, textProp.MinLength)
	assert.Equal(t, uint64(3), *textProp.MinLength)
}

// TestSchemaJSON_Caching tests all caching paths in SchemaJSON
func TestSchemaJSON_Caching(t *testing.T) {
	type Product struct {
		Name  string `json:"name" pedantigo:"required,min=1"`
		Price int    `json:"price" pedantigo:"gt=0"`
	}

	t.Run("first call generates and caches", func(t *testing.T) {
		validator := New[Product]()

		// First call should generate schema and JSON
		jsonBytes1, err := validator.SchemaJSON()
		require.NoError(t, err)
		assert.Greater(t, len(jsonBytes1), 0, "expected non-empty JSON bytes")

		// Verify it's valid JSON
		var schema1 map[string]any
		err = json.Unmarshal(jsonBytes1, &schema1)
		require.NoError(t, err)
	})

	t.Run("second call returns cached JSON", func(t *testing.T) {
		validator := New[Product]()

		// First call
		jsonBytes1, err1 := validator.SchemaJSON()
		require.NoError(t, err1)

		// Second call should return cached JSON (same pointer)
		jsonBytes2, err2 := validator.SchemaJSON()
		require.NoError(t, err2)

		// Should return exact same cached bytes
		assert.Equal(t, string(jsonBytes1), string(jsonBytes2))
	})

	t.Run("Schema called first then SchemaJSON uses cached schema", func(t *testing.T) {
		validator := New[Product]()

		// Call Schema() first to cache schema object
		schema1 := validator.Schema()
		require.NotNil(t, schema1)

		// Call SchemaJSON() - should use cached schema but generate JSON
		jsonBytes, err := validator.SchemaJSON()
		require.NoError(t, err)
		assert.Greater(t, len(jsonBytes), 0, "expected non-empty JSON bytes")

		// Verify constraints are in the JSON
		var schemaMap map[string]any
		err = json.Unmarshal(jsonBytes, &schemaMap)
		require.NoError(t, err)

		properties, ok := schemaMap["properties"].(map[string]any)
		require.True(t, ok, "expected properties object")

		nameProp, ok := properties["name"].(map[string]any)
		require.True(t, ok, "expected name property")

		// Check min length constraint
		minLen, ok := nameProp["minLength"].(float64)
		require.True(t, ok)
		assert.Equal(t, 1.0, minLen)
	})
}

// TestSchemaJSON_DefinitionUnwrapping tests definition unwrapping path
func TestSchemaJSON_DefinitionUnwrapping(t *testing.T) {
	// This tests the path where baseSchema.Properties is nil but has definitions
	// This happens with certain struct configurations
	// Config contains configuration settings
	type Config struct {
		Host string `json:"host" pedantigo:"required,url"`
		Port int    `json:"port" pedantigo:"gte=1,lte=65535"`
	}

	validator := New[Config]()
	jsonBytes, err := validator.SchemaJSON()
	require.NoError(t, err)

	var schemaMap map[string]any
	err = json.Unmarshal(jsonBytes, &schemaMap)
	require.NoError(t, err)

	// Should have properties (unwrapped from definition if needed)
	properties, ok := schemaMap["properties"].(map[string]any)
	require.True(t, ok, "expected properties object after unwrapping")

	// Verify constraints are applied
	hostProp, ok := properties["host"].(map[string]any)
	require.True(t, ok, "expected host property")

	format, ok := hostProp["format"].(string)
	require.True(t, ok)
	assert.Equal(t, "uri", format)
}
