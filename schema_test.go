package pedantigo

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/invopop/jsonschema"
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
				if schema.Type != "object" {
					t.Errorf("expected type 'object', got %s", schema.Type)
				}
				if schema.Properties == nil {
					t.Fatal("expected properties to be non-nil")
				}

				nameProp, _ := schema.Properties.Get("name")
				if nameProp == nil || nameProp.Type != "string" {
					t.Errorf("expected 'name' type 'string', got %v", nameProp)
				}

				ageProp, _ := schema.Properties.Get("age")
				if ageProp == nil || ageProp.Type != "integer" {
					t.Errorf("expected 'age' type 'integer', got %v", ageProp)
				}
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
				if addressProp == nil || addressProp.Type != "object" {
					t.Errorf("expected 'address' type 'object', got %v", addressProp)
				}

				cityProp, _ := addressProp.Properties.Get("city")
				if cityProp == nil {
					t.Fatal("expected nested 'city' property")
				}

				hasRequiredCity := false
				for _, req := range addressProp.Required {
					if req == "city" {
						hasRequiredCity = true
						break
					}
				}
				if !hasRequiredCity {
					t.Error("expected 'city' to be required in nested struct")
				}
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
				if tagsProp == nil || tagsProp.Type != "array" {
					t.Fatalf("expected 'tags' type 'array', got %v", tagsProp)
				}
				if tagsProp.Items == nil {
					t.Fatal("expected tags items to be defined")
				}
				if tagsProp.Items.Type != "string" {
					t.Errorf("expected tags items type 'string', got %s", tagsProp.Items.Type)
				}
				if tagsProp.Items.MinLength == nil || *tagsProp.Items.MinLength != 3 {
					t.Errorf("expected tags items minLength 3, got %v", tagsProp.Items.MinLength)
				}

				adminsProp, _ := schema.Properties.Get("admins")
				if adminsProp == nil || adminsProp.Items == nil {
					t.Fatal("expected 'admins' array with items")
				}
				if adminsProp.Items.Format != "email" {
					t.Errorf("expected admins items format 'email', got %s", adminsProp.Items.Format)
				}
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
				if settingsProp == nil || settingsProp.Type != "object" {
					t.Fatalf("expected 'settings' type 'object', got %v", settingsProp)
				}
				if settingsProp.AdditionalProperties == nil {
					t.Fatal("expected settings additionalProperties")
				}
				if settingsProp.AdditionalProperties.MinLength == nil || *settingsProp.AdditionalProperties.MinLength != 3 {
					t.Errorf("expected settings values minLength 3, got %v", settingsProp.AdditionalProperties.MinLength)
				}

				contactsProp, _ := schema.Properties.Get("contacts")
				if contactsProp == nil || contactsProp.AdditionalProperties == nil {
					t.Fatal("expected 'contacts' map with additionalProperties")
				}
				if contactsProp.AdditionalProperties.Format != "email" {
					t.Errorf("expected contacts values format 'email', got %s", contactsProp.AdditionalProperties.Format)
				}
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
				if schema == nil {
					t.Fatal("expected schema to be non-nil")
				}
				tt.validate(t, schema)
			default:
				t.Fatal("invalid validator type")
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
				if len(schema.Required) != 2 {
					t.Errorf("expected 2 required fields, got %d", len(schema.Required))
				}
				requiredMap := make(map[string]bool)
				for _, field := range schema.Required {
					requiredMap[field] = true
				}
				if !requiredMap["name"] || !requiredMap["email"] {
					t.Errorf("expected 'name' and 'email' to be required, got %v", schema.Required)
				}
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
				if string(priceProp.ExclusiveMinimum) != "0" || string(priceProp.ExclusiveMaximum) != "10000" {
					t.Errorf("price: expected exclusive min/max 0/10000, got %v/%v", priceProp.ExclusiveMinimum, priceProp.ExclusiveMaximum)
				}

				stockProp, _ := schema.Properties.Get("stock")
				if string(stockProp.Minimum) != "0" || string(stockProp.Maximum) != "1000" {
					t.Errorf("stock: expected min/max 0/1000, got %v/%v", stockProp.Minimum, stockProp.Maximum)
				}

				discountProp, _ := schema.Properties.Get("discount")
				if string(discountProp.Minimum) != "0" || string(discountProp.Maximum) != "100" {
					t.Errorf("discount: expected min/max 0/100, got %v/%v", discountProp.Minimum, discountProp.Maximum)
				}
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
				if usernameProp.MinLength == nil || *usernameProp.MinLength != 3 {
					t.Errorf("expected username minLength 3, got %v", usernameProp.MinLength)
				}
				if usernameProp.MaxLength == nil || *usernameProp.MaxLength != 20 {
					t.Errorf("expected username maxLength 20, got %v", usernameProp.MaxLength)
				}

				bioProp, _ := schema.Properties.Get("bio")
				if bioProp.MaxLength == nil || *bioProp.MaxLength != 500 {
					t.Errorf("expected bio maxLength 500, got %v", bioProp.MaxLength)
				}
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
					if prop == nil {
						t.Fatalf("expected field '%s' to exist", tt.field)
					}
					if prop.Format != tt.expectedFmt {
						t.Errorf("field '%s': expected format '%s', got '%s'", tt.field, tt.expectedFmt, prop.Format)
					}
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
				if zipProp.Pattern != "^[0-9]{5}$" {
					t.Errorf("expected pattern '^[0-9]{5}$', got '%s'", zipProp.Pattern)
				}
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
				if string(portDefault) != "8080" {
					t.Errorf("expected port default 8080, got %v", portProp.Default)
				}

				hostProp, _ := schema.Properties.Get("host")
				if hostProp.Default != "localhost" {
					t.Errorf("expected host default 'localhost', got %v", hostProp.Default)
				}

				enabledProp, _ := schema.Properties.Get("enabled")
				enabledDefault, _ := json.Marshal(enabledProp.Default)
				if string(enabledDefault) != "true" {
					t.Errorf("expected enabled default true, got %v", enabledProp.Default)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			switch validator := v.(type) {
			case interface{ Schema() *jsonschema.Schema }:
				schema := validator.Schema()
				if schema == nil {
					t.Fatal("expected schema to be non-nil")
				}
				tt.validate(t, schema)
			default:
				t.Fatal("invalid validator type")
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
				if !ok {
					t.Fatal("validator missing SchemaJSON method")
				}

				jsonBytes, err := validator.SchemaJSON()
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				var schemaMap map[string]any
				if err := json.Unmarshal(jsonBytes, &schemaMap); err != nil {
					t.Fatalf("expected valid JSON, got error: %v", err)
				}

				if schemaMap["type"] != "object" {
					t.Errorf("expected type 'object', got %v", schemaMap["type"])
				}

				properties, ok := schemaMap["properties"].(map[string]any)
				if !ok {
					t.Fatal("expected properties to be an object")
				}

				// Check name field has minLength
				nameField, ok := properties["name"].(map[string]any)
				if !ok || nameField["minLength"] != float64(3) {
					t.Errorf("expected name minLength 3, got %v", nameField)
				}

				// Check email field has format
				emailField, ok := properties["email"].(map[string]any)
				if !ok || emailField["format"] != "email" {
					t.Errorf("expected email format 'email', got %v", emailField)
				}
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
				if !ok {
					t.Fatal("validator missing SchemaJSONOpenAPI method")
				}

				jsonBytes, err := validator.SchemaJSONOpenAPI()
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				var schemaMap map[string]any
				if err := json.Unmarshal(jsonBytes, &schemaMap); err != nil {
					t.Fatalf("expected valid JSON, got error: %v", err)
				}

				// Check that $defs exists
				defs, hasDefs := schemaMap["$defs"].(map[string]any)
				if !hasDefs {
					t.Fatal("expected $defs to exist in OpenAPI schema")
				}

				// Check Address definition exists
				addressDef, ok := defs["Address"].(map[string]any)
				if !ok {
					t.Fatal("expected Address definition in $defs")
				}

				// Check Address has required city
				addressRequired, ok := addressDef["required"].([]any)
				if !ok {
					t.Fatal("expected 'required' array in Address definition")
				}
				hasCity := false
				for _, req := range addressRequired {
					if req == "city" {
						hasCity = true
						break
					}
				}
				if !hasCity {
					t.Error("expected 'city' to be required in Address definition")
				}

				// Check zip constraint in Address definition
				addressProps, ok := addressDef["properties"].(map[string]any)
				if !ok {
					t.Fatal("expected 'properties' in Address definition")
				}
				zipProp, ok := addressProps["zip"].(map[string]any)
				if !ok || zipProp["minLength"] != float64(5) {
					t.Errorf("expected zip minLength 5 in Address, got %v", zipProp)
				}

				// Check root schema has $ref to Address
				properties, ok := schemaMap["properties"].(map[string]any)
				if !ok {
					t.Fatal("expected 'properties' in root schema")
				}
				addressProp, ok := properties["address"].(map[string]any)
				if !ok {
					t.Fatal("expected 'address' property in root schema")
				}
				ref, hasRef := addressProp["$ref"].(string)
				if !hasRef || ref != "#/$defs/Address" {
					t.Errorf("expected $ref '#/$defs/Address', got %s", ref)
				}
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
				if !ok {
					t.Fatal("validator missing SchemaOpenAPI method")
				}

				schema := validator.SchemaOpenAPI()
				if len(schema.Definitions) == 0 {
					t.Fatal("expected schema to have definitions")
				}

				contactDef, ok := schema.Definitions["Contact"]
				if !ok {
					t.Fatal("expected Contact definition")
				}

				// Check Contact has required email
				hasEmail := false
				for _, req := range contactDef.Required {
					if req == "email" {
						hasEmail = true
						break
					}
				}
				if !hasEmail {
					t.Error("expected 'email' to be required in Contact definition")
				}

				// Check constraints in Contact definition
				emailProp, _ := contactDef.Properties.Get("email")
				if emailProp == nil || emailProp.Format != "email" {
					t.Errorf("expected email format in Contact definition, got %v", emailProp)
				}

				phoneProp, _ := contactDef.Properties.Get("phone")
				if phoneProp == nil || phoneProp.MinLength == nil || *phoneProp.MinLength != 10 {
					t.Errorf("expected phone minLength 10 in Contact definition, got %v", phoneProp)
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
				if schema1 == nil || schema2 == nil {
					t.Fatal("expected non-nil schemas")
				}
				if schema1 != schema2 {
					t.Error("expected Schema() to return same cached pointer")
				}
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

				if err1 != nil || err2 != nil {
					t.Fatalf("unexpected errors: %v, %v", err1, err2)
				}
				if len(json1) != len(json2) {
					t.Errorf("expected same cached length, got %d vs %d", len(json1), len(json2))
				}
				if !bytesEqual(json1, json2) {
					t.Error("expected SchemaJSON() to return identical cached bytes")
				}
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
				if openapi1 == nil || openapi2 == nil {
					t.Fatal("expected non-nil schemas")
				}
				if openapi1 != openapi2 {
					t.Error("expected SchemaOpenAPI() to return same cached pointer")
				}
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

				if err1 != nil || err2 != nil {
					t.Fatalf("unexpected errors: %v, %v", err1, err2)
				}
				if len(json1) != len(json2) {
					t.Errorf("expected same cached length, got %d vs %d", len(json1), len(json2))
				}
				if !bytesEqual(json1, json2) {
					t.Error("expected SchemaJSONOpenAPI() to return identical cached bytes")
				}
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
				if catSchema1 != catSchema2 {
					t.Error("expected Cat validator to cache same pointer")
				}
				if dogSchema1 != dogSchema2 {
					t.Error("expected Dog validator to cache same pointer")
				}
				// But different validators have different caches
				if catSchema1 == dogSchema1 {
					t.Error("expected different validators to have independent caches")
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
					if schema == nil {
						t.Fatal("expected non-nil schema from concurrent call")
					}
					pointers = append(pointers, schema)
				}

				firstPtr := pointers[0]
				for i, ptr := range pointers {
					if ptr != firstPtr {
						t.Errorf("goroutine %d got different pointer than first call", i)
					}
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
						if err != nil {
							panic(fmt.Sprintf("concurrent error: %v", err))
						}
						jsonChan <- jsonBytes
					}()
				}

				wg.Wait()
				close(jsonChan)

				// Verify all concurrent calls returned same cached bytes
				allBytes := make([][]byte, 0, numGoroutines)
				for jsonBytes := range jsonChan {
					if jsonBytes == nil {
						t.Fatal("expected non-nil bytes from concurrent call")
					}
					allBytes = append(allBytes, jsonBytes)
				}

				firstBytes := allBytes[0]
				for i, jsonBytes := range allBytes {
					if !bytesEqual(jsonBytes, firstBytes) {
						t.Errorf("goroutine %d got different cached bytes than first call", i)
					}
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
	if !hasAuthor {
		t.Fatal("expected Author definition in $defs")
	}

	// Verify Author definition has constraints from pedantigo tags
	hasNameRequired := false
	for _, req := range authorDef.Required {
		if req == "name" {
			hasNameRequired = true
			break
		}
	}
	if !hasNameRequired {
		t.Error("expected 'name' to be required in Author definition")
	}

	nameProp, _ := authorDef.Properties.Get("name")
	if nameProp == nil || nameProp.MinLength == nil || *nameProp.MinLength != 2 {
		t.Errorf("expected name minLength 2 in Author, got %v", nameProp)
	}

	emailProp, _ := authorDef.Properties.Get("email")
	if emailProp == nil || emailProp.Format != "email" {
		t.Errorf("expected email format 'email' in Author, got %v", emailProp)
	}
}

// TestSchemaOpenAPI_PointerSliceOfStructs tests schema with pointer slices
// This exercises searchSliceType() with pointer unwrapping
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
	if !hasTag {
		t.Fatal("expected Tag definition in $defs (pointer slice)")
	}

	// Verify Tag constraints are applied
	colorProp, _ := tagDef.Properties.Get("color")
	if colorProp == nil || colorProp.Pattern != "^#[0-9a-fA-F]{6}$" {
		t.Errorf("expected color pattern in Tag definition, got %v", colorProp)
	}
}

// TestSchemaOpenAPI_MapOfStructs tests schema generation with maps of structs
// This exercises searchMapType() code path (currently 0% coverage)
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
	if !hasContact {
		t.Fatal("expected Contact definition in $defs")
	}

	// Verify Contact definition has constraints
	hasEmailRequired := false
	for _, req := range contactDef.Required {
		if req == "email" {
			hasEmailRequired = true
			break
		}
	}
	if !hasEmailRequired {
		t.Error("expected 'email' to be required in Contact definition")
	}

	emailProp, _ := contactDef.Properties.Get("email")
	if emailProp == nil || emailProp.Format != "email" {
		t.Errorf("expected email format 'email' in Contact, got %v", emailProp)
	}

	phoneProp, _ := contactDef.Properties.Get("phone")
	if phoneProp == nil || phoneProp.MinLength == nil || *phoneProp.MinLength != 10 {
		t.Errorf("expected phone minLength 10 in Contact, got %v", phoneProp)
	}
}

// TestSchemaOpenAPI_PointerMapOfStructs tests schema with pointer map values
// This exercises searchMapType() with pointer unwrapping
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
	if !hasAddress {
		t.Fatal("expected Address definition in $defs (pointer map values)")
	}

	// Verify Address constraints are applied
	zipProp, _ := addressDef.Properties.Get("zipCode")
	if zipProp == nil || zipProp.Pattern != "^[0-9]{5}$" {
		t.Errorf("expected zipCode pattern in Address definition, got %v", zipProp)
	}

	cityProp, _ := addressDef.Properties.Get("city")
	if cityProp == nil || cityProp.MinLength == nil || *cityProp.MinLength != 2 {
		t.Errorf("expected city minLength 2 in Address, got %v", cityProp)
	}
}

// TestSchemaOpenAPI_NestedStructInSlice tests deeply nested struct in slice
// This exercises recursive findTypeForDefinition through searchSliceType
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
	if !hasRole {
		t.Fatal("expected Role definition in $defs")
	}

	permDef, hasPerm := schema.Definitions["Permission"]
	if !hasPerm {
		t.Fatal("expected Permission definition in $defs (nested in slice)")
	}

	// Verify Permission constraints applied
	nameProp, _ := permDef.Properties.Get("name")
	if nameProp == nil || nameProp.MinLength == nil || *nameProp.MinLength != 1 {
		t.Errorf("expected name minLength 1 in Permission, got %v", nameProp)
	}

	// Verify Role constraints applied
	titleProp, _ := roleDef.Properties.Get("title")
	if titleProp == nil || titleProp.MinLength == nil || *titleProp.MinLength != 1 {
		t.Errorf("expected title minLength 1 in Role, got %v", titleProp)
	}
}

// TestSchemaOpenAPI_NestedStructInMap tests deeply nested struct in map
// This exercises recursive findTypeForDefinition through searchMapType
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
	if !hasResource {
		t.Fatal("expected Resource definition in $defs")
	}

	metadataDef, hasMetadata := schema.Definitions["Metadata"]
	if !hasMetadata {
		t.Fatal("expected Metadata definition in $defs (nested in map)")
	}

	// Verify Metadata constraints applied
	keyProp, _ := metadataDef.Properties.Get("key")
	if keyProp == nil || keyProp.MinLength == nil || *keyProp.MinLength != 1 {
		t.Errorf("expected key minLength 1 in Metadata, got %v", keyProp)
	}

	// Verify Resource constraints applied
	nameProp, _ := resourceDef.Properties.Get("name")
	if nameProp == nil || nameProp.MinLength == nil || *nameProp.MinLength != 1 {
		t.Errorf("expected name minLength 1 in Resource, got %v", nameProp)
	}
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
	if !hasAddress {
		t.Fatal("expected Address definition in $defs")
	}

	// Verify Address definition has constraints
	streetProp, _ := addressDef.Properties.Get("street")
	if streetProp == nil {
		t.Error("expected 'street' property in Address definition")
	}

	cityProp, _ := addressDef.Properties.Get("city")
	if cityProp == nil || cityProp.MinLength == nil || *cityProp.MinLength != 2 {
		t.Errorf("expected city minLength 2 in Address, got %v", cityProp)
	}
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
	if !hasConfig {
		t.Fatal("expected Config definition in $defs (pointer should be unwrapped)")
	}

	// Verify Config definition has constraints
	keyProp, _ := configDef.Properties.Get("key")
	if keyProp == nil {
		t.Error("expected 'key' property in Config definition")
	}

	valueProp, _ := configDef.Properties.Get("value")
	if valueProp == nil || valueProp.MinLength == nil || *valueProp.MinLength != 1 {
		t.Errorf("expected value minLength 1 in Config, got %v", valueProp)
	}
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
	if !hasLevel2 {
		t.Error("expected Level2 definition in $defs")
	}

	level3Def, hasLevel3 := schema.Definitions["Level3"]
	if !hasLevel3 {
		t.Fatal("expected Level3 definition in $defs (deeply nested)")
	}

	// Verify Level3 definition has constraints
	dataProp, _ := level3Def.Properties.Get("data")
	if dataProp == nil || dataProp.MinLength == nil || *dataProp.MinLength != 5 {
		t.Errorf("expected data minLength 5 in Level3, got %v", dataProp)
	}
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
	if !hasTag {
		t.Error("expected Tag definition")
	}
	if tagDef.Properties.Len() == 0 {
		t.Error("expected Tag definition to have properties")
	}

	metaDef, hasMeta := schema.Definitions["Metadata"]
	if !hasMeta {
		t.Error("expected Metadata definition from map values")
	}
	if metaDef.Properties.Len() == 0 {
		t.Error("expected Metadata definition to have properties")
	}

	commentDef, hasComment := schema.Definitions["Comment"]
	if !hasComment {
		t.Error("expected Comment definition from slice")
	}

	// Verify Comment definition has constraints
	textProp, _ := commentDef.Properties.Get("text")
	if textProp == nil || textProp.MinLength == nil || *textProp.MinLength != 3 {
		t.Errorf("expected text minLength 3 in Comment, got %v", textProp)
	}
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
		if err != nil {
			t.Fatalf("expected no error on first call, got %v", err)
		}

		if len(jsonBytes1) == 0 {
			t.Error("expected non-empty JSON bytes")
		}

		// Verify it's valid JSON
		var schema1 map[string]any
		if err := json.Unmarshal(jsonBytes1, &schema1); err != nil {
			t.Fatalf("expected valid JSON, got error: %v", err)
		}
	})

	t.Run("second call returns cached JSON", func(t *testing.T) {
		validator := New[Product]()

		// First call
		jsonBytes1, err1 := validator.SchemaJSON()
		if err1 != nil {
			t.Fatalf("first call error: %v", err1)
		}

		// Second call should return cached JSON (same pointer)
		jsonBytes2, err2 := validator.SchemaJSON()
		if err2 != nil {
			t.Fatalf("second call error: %v", err2)
		}

		// Should return exact same cached bytes
		if string(jsonBytes1) != string(jsonBytes2) {
			t.Error("expected cached JSON to match")
		}
	})

	t.Run("Schema called first then SchemaJSON uses cached schema", func(t *testing.T) {
		validator := New[Product]()

		// Call Schema() first to cache schema object
		schema1 := validator.Schema()
		if schema1 == nil {
			t.Fatal("expected schema to be generated")
		}

		// Call SchemaJSON() - should use cached schema but generate JSON
		jsonBytes, err := validator.SchemaJSON()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(jsonBytes) == 0 {
			t.Error("expected non-empty JSON bytes")
		}

		// Verify constraints are in the JSON
		var schemaMap map[string]any
		if err := json.Unmarshal(jsonBytes, &schemaMap); err != nil {
			t.Fatalf("expected valid JSON, got error: %v", err)
		}

		properties, ok := schemaMap["properties"].(map[string]any)
		if !ok {
			t.Fatal("expected properties object")
		}

		nameProp, ok := properties["name"].(map[string]any)
		if !ok {
			t.Fatal("expected name property")
		}

		// Check min length constraint
		if minLen, ok := nameProp["minLength"].(float64); !ok || minLen != 1 {
			t.Errorf("expected name minLength 1, got %v", nameProp["minLength"])
		}
	})
}

// TestSchemaJSON_DefinitionUnwrapping tests definition unwrapping path
func TestSchemaJSON_DefinitionUnwrapping(t *testing.T) {
	// This tests the path where baseSchema.Properties is nil but has definitions
	// This happens with certain struct configurations
	type Config struct {
		Host string `json:"host" pedantigo:"required,url"`
		Port int    `json:"port" pedantigo:"gte=1,lte=65535"`
	}

	validator := New[Config]()
	jsonBytes, err := validator.SchemaJSON()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(jsonBytes, &schemaMap); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// Should have properties (unwrapped from definition if needed)
	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties object after unwrapping")
	}

	// Verify constraints are applied
	hostProp, ok := properties["host"].(map[string]any)
	if !ok {
		t.Fatal("expected host property")
	}

	if format, ok := hostProp["format"].(string); !ok || format != "uri" {
		t.Errorf("expected host format 'uri', got %v", hostProp["format"])
	}
}
