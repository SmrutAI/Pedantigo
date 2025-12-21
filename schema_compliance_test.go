package pedantigo

import (
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================================================
// JSON Schema Spec Compliance Tests
// These tests validate that generated schemas are valid JSON Schema
// according to the JSON Schema Draft 2020-12 specification.
// ==================================================

// compileSchema validates that a schema is a valid JSON Schema.
// Returns an error if the schema is invalid according to JSON Schema spec.
func compileSchema(t *testing.T, schemaBytes []byte) *jsonschema.Schema {
	t.Helper()

	c := jsonschema.NewCompiler()

	// Parse the schema
	var schemaMap any
	err := json.Unmarshal(schemaBytes, &schemaMap)
	require.NoError(t, err, "schema should be valid JSON")

	// Add the schema to the compiler and compile it
	err = c.AddResource("schema.json", schemaMap)
	require.NoError(t, err, "schema should be addable to compiler")

	compiled, err := c.Compile("schema.json")
	require.NoError(t, err, "schema should compile against JSON Schema meta-schema")

	return compiled
}

// TestSchemaCompliance_BasicTypes verifies basic type schemas are valid JSON Schema.
func TestSchemaCompliance_BasicTypes(t *testing.T) {
	tests := []struct {
		name  string
		setup func() ([]byte, error)
	}{
		{
			name: "string fields",
			setup: func() ([]byte, error) {
				type User struct {
					Name  string `json:"name" pedantigo:"required"`
					Email string `json:"email" pedantigo:"email"`
				}
				return New[User]().SchemaJSON()
			},
		},
		{
			name: "numeric fields",
			setup: func() ([]byte, error) {
				type Product struct {
					Price    float64 `json:"price" pedantigo:"gt=0"`
					Quantity int     `json:"quantity" pedantigo:"gte=0"`
					Discount float32 `json:"discount" pedantigo:"min=0,max=100"`
				}
				return New[Product]().SchemaJSON()
			},
		},
		{
			name: "boolean fields",
			setup: func() ([]byte, error) {
				type Config struct {
					Enabled bool `json:"enabled" pedantigo:"required"`
					Debug   bool `json:"debug"`
				}
				return New[Config]().SchemaJSON()
			},
		},
		{
			name: "array fields",
			setup: func() ([]byte, error) {
				type Tags struct {
					Items []string `json:"items" pedantigo:"min=1"`
					IDs   []int    `json:"ids" pedantigo:"required"`
				}
				return New[Tags]().SchemaJSON()
			},
		},
		{
			name: "map fields",
			setup: func() ([]byte, error) {
				type Settings struct {
					Data     map[string]string `json:"data"`
					Metadata map[string]int    `json:"metadata"`
				}
				return New[Settings]().SchemaJSON()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaBytes, err := tt.setup()
			require.NoError(t, err, "schema generation should succeed")

			compiled := compileSchema(t, schemaBytes)
			assert.NotNil(t, compiled, "schema should compile successfully")
		})
	}
}

// TestSchemaCompliance_Constraints verifies constraint schemas are valid JSON Schema.
func TestSchemaCompliance_Constraints(t *testing.T) {
	tests := []struct {
		name  string
		setup func() ([]byte, error)
	}{
		{
			name: "string length constraints (minLength, maxLength)",
			setup: func() ([]byte, error) {
				type User struct {
					Username string `json:"username" pedantigo:"min=3,max=20"`
					Bio      string `json:"bio" pedantigo:"max=500"`
					Code     string `json:"code" pedantigo:"len=6"`
				}
				return New[User]().SchemaJSON()
			},
		},
		{
			name: "numeric range constraints (minimum, maximum)",
			setup: func() ([]byte, error) {
				type Product struct {
					Price float64 `json:"price" pedantigo:"min=0,max=10000"`
					Stock int     `json:"stock" pedantigo:"gte=0,lte=1000"`
				}
				return New[Product]().SchemaJSON()
			},
		},
		{
			name: "exclusive range constraints (exclusiveMinimum, exclusiveMaximum)",
			setup: func() ([]byte, error) {
				type Range struct {
					Value float64 `json:"value" pedantigo:"gt=0,lt=100"`
				}
				return New[Range]().SchemaJSON()
			},
		},
		{
			name: "format constraints",
			setup: func() ([]byte, error) {
				type Contact struct {
					Email    string `json:"email" pedantigo:"email"`
					Website  string `json:"website" pedantigo:"url"`
					ID       string `json:"id" pedantigo:"uuid"`
					IPv4Addr string `json:"ipv4" pedantigo:"ipv4"`
					IPv6Addr string `json:"ipv6" pedantigo:"ipv6"`
				}
				return New[Contact]().SchemaJSON()
			},
		},
		{
			name: "pattern constraint",
			setup: func() ([]byte, error) {
				type Code struct {
					ZipCode string `json:"zipCode" pedantigo:"regexp=^[0-9]{5}$"`
					SKU     string `json:"sku" pedantigo:"regexp=^[A-Z]{3}-[0-9]{4}$"`
				}
				return New[Code]().SchemaJSON()
			},
		},
		{
			name: "enum constraint (oneof)",
			setup: func() ([]byte, error) {
				type Status struct {
					State string `json:"state" pedantigo:"oneof=pending active completed"`
					Role  string `json:"role" pedantigo:"oneof=admin user guest"`
				}
				return New[Status]().SchemaJSON()
			},
		},
		{
			name: "const constraint (eq)",
			setup: func() ([]byte, error) {
				type Fixed struct {
					Type    string `json:"type" pedantigo:"eq=user"`
					Version string `json:"version" pedantigo:"eq=1.0"`
				}
				return New[Fixed]().SchemaJSON()
			},
		},
		{
			name: "not constraint (ne)",
			setup: func() ([]byte, error) {
				type Exclusion struct {
					Status string `json:"status" pedantigo:"ne=deleted"`
				}
				return New[Exclusion]().SchemaJSON()
			},
		},
		{
			name: "multipleOf constraint",
			setup: func() ([]byte, error) {
				type Quantity struct {
					Count int `json:"count" pedantigo:"multiple_of=5"`
				}
				return New[Quantity]().SchemaJSON()
			},
		},
		{
			name: "default values",
			setup: func() ([]byte, error) {
				type Config struct {
					Port    int    `json:"port" pedantigo:"default=8080"`
					Host    string `json:"host" pedantigo:"default=localhost"`
					Enabled bool   `json:"enabled" pedantigo:"default=true"`
				}
				return New[Config]().SchemaJSON()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaBytes, err := tt.setup()
			require.NoError(t, err, "schema generation should succeed")

			compiled := compileSchema(t, schemaBytes)
			assert.NotNil(t, compiled, "schema should compile successfully")
		})
	}
}

// TestSchemaCompliance_NestedStructs verifies nested struct schemas are valid JSON Schema.
func TestSchemaCompliance_NestedStructs(t *testing.T) {
	tests := []struct {
		name  string
		setup func() ([]byte, error)
	}{
		{
			name: "simple nested struct",
			setup: func() ([]byte, error) {
				type Address struct {
					City string `json:"city" pedantigo:"required"`
					Zip  string `json:"zip" pedantigo:"min=5"`
				}
				type User struct {
					Name    string  `json:"name" pedantigo:"required"`
					Address Address `json:"address"`
				}
				return New[User]().SchemaJSON()
			},
		},
		{
			name: "deeply nested structs",
			setup: func() ([]byte, error) {
				type Level3 struct {
					Data string `json:"data" pedantigo:"required"`
				}
				type Level2 struct {
					Nested Level3 `json:"nested"`
				}
				type Level1 struct {
					Mid Level2 `json:"mid"`
				}
				return New[Level1]().SchemaJSON()
			},
		},
		{
			name: "pointer to struct",
			setup: func() ([]byte, error) {
				type Config struct {
					Key   string `json:"key" pedantigo:"required"`
					Value string `json:"value"`
				}
				type Service struct {
					Name   string  `json:"name"`
					Config *Config `json:"config"`
				}
				return New[Service]().SchemaJSON()
			},
		},
		{
			name: "slice of structs",
			setup: func() ([]byte, error) {
				type Item struct {
					ID   string `json:"id" pedantigo:"required"`
					Name string `json:"name"`
				}
				type Order struct {
					Items []Item `json:"items"`
				}
				return New[Order]().SchemaJSON()
			},
		},
		{
			name: "map of structs",
			setup: func() ([]byte, error) {
				type Metadata struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}
				type Resource struct {
					Meta map[string]Metadata `json:"meta"`
				}
				return New[Resource]().SchemaJSON()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaBytes, err := tt.setup()
			require.NoError(t, err, "schema generation should succeed")

			compiled := compileSchema(t, schemaBytes)
			assert.NotNil(t, compiled, "schema should compile successfully")
		})
	}
}

// TestSchemaCompliance_OpenAPI verifies OpenAPI schemas with $defs are valid JSON Schema.
func TestSchemaCompliance_OpenAPI(t *testing.T) {
	tests := []struct {
		name  string
		setup func() ([]byte, error)
	}{
		{
			name: "schema with $defs (nested struct)",
			setup: func() ([]byte, error) {
				type Address struct {
					City string `json:"city" pedantigo:"required"`
					Zip  string `json:"zip" pedantigo:"min=5"`
				}
				type User struct {
					Name    string  `json:"name" pedantigo:"required"`
					Address Address `json:"address"`
				}
				return New[User]().SchemaJSONOpenAPI()
			},
		},
		{
			name: "schema with multiple $defs",
			setup: func() ([]byte, error) {
				type Tag struct {
					Name string `json:"name" pedantigo:"required"`
				}
				type Author struct {
					Email string `json:"email" pedantigo:"email"`
				}
				type Article struct {
					Title  string `json:"title" pedantigo:"required"`
					Tags   []Tag  `json:"tags"`
					Author Author `json:"author"`
				}
				return New[Article]().SchemaJSONOpenAPI()
			},
		},
		{
			name: "deeply nested with $defs",
			setup: func() ([]byte, error) {
				type Permission struct {
					Name string `json:"name" pedantigo:"required"`
				}
				type Role struct {
					Title       string       `json:"title" pedantigo:"required"`
					Permissions []Permission `json:"permissions"`
				}
				type User struct {
					Username string `json:"username" pedantigo:"required"`
					Roles    []Role `json:"roles"`
				}
				return New[User]().SchemaJSONOpenAPI()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaBytes, err := tt.setup()
			require.NoError(t, err, "schema generation should succeed")

			compiled := compileSchema(t, schemaBytes)
			assert.NotNil(t, compiled, "schema should compile successfully")

			// Verify $defs exists in schema
			var schemaMap map[string]any
			err = json.Unmarshal(schemaBytes, &schemaMap)
			require.NoError(t, err)

			_, hasDefs := schemaMap["$defs"]
			assert.True(t, hasDefs, "OpenAPI schema should contain $defs")
		})
	}
}

// TestSchemaCompliance_Metadata verifies metadata schema fields are valid JSON Schema.
func TestSchemaCompliance_Metadata(t *testing.T) {
	tests := []struct {
		name  string
		setup func() ([]byte, error)
	}{
		{
			name: "title and description",
			setup: func() ([]byte, error) {
				type User struct {
					Name string `json:"name" pedantigo:"title=User Name,description=The user's full name"`
				}
				return New[User]().SchemaJSON()
			},
		},
		{
			name: "examples",
			setup: func() ([]byte, error) {
				type Email struct {
					Address string `json:"address" pedantigo:"email,examples=user@example.com|admin@test.org"`
				}
				return New[Email]().SchemaJSON()
			},
		},
		{
			name: "deprecated",
			setup: func() ([]byte, error) {
				type Legacy struct {
					OldField string `json:"old_field" pedantigo:"deprecated=Use newField instead"`
				}
				return New[Legacy]().SchemaJSON()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaBytes, err := tt.setup()
			require.NoError(t, err, "schema generation should succeed")

			compiled := compileSchema(t, schemaBytes)
			assert.NotNil(t, compiled, "schema should compile successfully")
		})
	}
}

// TestSchemaCompliance_ComplexTypes verifies complex type combinations are valid JSON Schema.
func TestSchemaCompliance_ComplexTypes(t *testing.T) {
	tests := []struct {
		name  string
		setup func() ([]byte, error)
	}{
		{
			name: "all constraint types combined",
			setup: func() ([]byte, error) {
				type ComplexUser struct {
					ID        string   `json:"id" pedantigo:"required,uuid"`
					Email     string   `json:"email" pedantigo:"required,email"`
					Username  string   `json:"username" pedantigo:"required,min=3,max=20,alpha"`
					Age       int      `json:"age" pedantigo:"gte=18,lte=120"`
					Score     float64  `json:"score" pedantigo:"gt=0,lt=100"`
					Status    string   `json:"status" pedantigo:"oneof=active inactive pending"`
					Tags      []string `json:"tags" pedantigo:"min=1"`
					CreatedAt string   `json:"created_at" pedantigo:"required"`
				}
				return New[ComplexUser]().SchemaJSON()
			},
		},
		{
			name: "nested with all features",
			setup: func() ([]byte, error) {
				type Address struct {
					Street  string `json:"street" pedantigo:"required,min=5"`
					City    string `json:"city" pedantigo:"required,min=2"`
					ZipCode string `json:"zip_code" pedantigo:"regexp=^[0-9]{5}$"`
					Country string `json:"country" pedantigo:"oneof=US CA UK"`
				}
				type Company struct {
					Name    string `json:"name" pedantigo:"required,min=1,max=100"`
					Website string `json:"website" pedantigo:"url"`
					Size    int    `json:"size" pedantigo:"gte=1"`
				}
				type Employee struct {
					ID      string   `json:"id" pedantigo:"required,uuid"`
					Name    string   `json:"name" pedantigo:"required,min=2"`
					Email   string   `json:"email" pedantigo:"required,email"`
					Address Address  `json:"address" pedantigo:"required"`
					Company Company  `json:"company"`
					Skills  []string `json:"skills" pedantigo:"min=1"`
				}
				return New[Employee]().SchemaJSON()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaBytes, err := tt.setup()
			require.NoError(t, err, "schema generation should succeed")

			compiled := compileSchema(t, schemaBytes)
			assert.NotNil(t, compiled, "schema should compile successfully")
		})
	}
}

// TestSchemaCompliance_ValidationWorks verifies the compiled schema can validate data.
func TestSchemaCompliance_ValidationWorks(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"required,min=3"`
		Email string `json:"email" pedantigo:"required,email"`
		Age   int    `json:"age" pedantigo:"gte=18"`
	}

	schemaBytes, err := New[User]().SchemaJSON()
	require.NoError(t, err)

	compiled := compileSchema(t, schemaBytes)
	require.NotNil(t, compiled)

	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{
			name:      "valid data",
			input:     `{"name": "John Doe", "email": "john@example.com", "age": 25}`,
			wantValid: true,
		},
		{
			name:      "missing required field",
			input:     `{"name": "John Doe", "age": 25}`,
			wantValid: false,
		},
		{
			name:      "name too short",
			input:     `{"name": "Jo", "email": "john@example.com", "age": 25}`,
			wantValid: false,
		},
		{
			name:      "age below minimum",
			input:     `{"name": "John Doe", "email": "john@example.com", "age": 15}`,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data any
			err := json.Unmarshal([]byte(tt.input), &data)
			require.NoError(t, err)

			err = compiled.Validate(data)
			if tt.wantValid {
				assert.NoError(t, err, "expected valid data")
			} else {
				assert.Error(t, err, "expected invalid data")
			}
		})
	}
}
