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
