package pedantigo

import (
	"encoding/json"
	"testing"
)

// ==================================================
// Schema() generation tests
// ==================================================

func TestSchema_BasicTypes(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	validator := New[User]()
	schema := validator.Schema()

	if schema == nil {
		t.Fatal("expected schema to be non-nil")
	}

	// Check schema type
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}

	// Check properties exist
	if schema.Properties == nil {
		t.Fatal("expected properties to be non-nil")
	}

	// Check name field
	nameProp, _ := schema.Properties.Get("name")
	if nameProp == nil {
		t.Fatal("expected 'name' property to exist")
	}
	if nameProp.Type != "string" {
		t.Errorf("expected name type 'string', got %s", nameProp.Type)
	}

	// Check age field
	ageProp, _ := schema.Properties.Get("age")
	if ageProp == nil {
		t.Fatal("expected 'age' property to exist")
	}
	if ageProp.Type != "integer" {
		t.Errorf("expected age type 'integer', got %s", ageProp.Type)
	}
}

func TestSchema_RequiredFields(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"required"`
		Email string `json:"email" pedantigo:"required"`
		Age   int    `json:"age"`
	}

	validator := New[User]()
	schema := validator.Schema()

	// Check required array
	if len(schema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(schema.Required))
	}

	requiredFields := make(map[string]bool)
	for _, field := range schema.Required {
		requiredFields[field] = true
	}

	if !requiredFields["name"] {
		t.Error("expected 'name' to be required")
	}
	if !requiredFields["email"] {
		t.Error("expected 'email' to be required")
	}
}

func TestSchema_NumericConstraints(t *testing.T) {
	type Product struct {
		Price    float64 `json:"price" pedantigo:"gt=0,lt=10000"`
		Stock    int     `json:"stock" pedantigo:"gte=0,lte=1000"`
		Discount int     `json:"discount" pedantigo:"min=0,max=100"`
	}

	validator := New[Product]()
	schema := validator.Schema()

	// Check price constraints (gt/lt should map to exclusiveMinimum/exclusiveMaximum)
	priceProp, _ := schema.Properties.Get("price")
	if priceProp == nil {
		t.Fatal("expected 'price' property to exist")
	}
	if string(priceProp.ExclusiveMinimum) != "0" {
		t.Errorf("expected price exclusiveMinimum 0, got %v", priceProp.ExclusiveMinimum)
	}
	if string(priceProp.ExclusiveMaximum) != "10000" {
		t.Errorf("expected price exclusiveMaximum 10000, got %v", priceProp.ExclusiveMaximum)
	}

	// Check stock constraints (ge/le should map to minimum/maximum)
	stockProp, _ := schema.Properties.Get("stock")
	if stockProp == nil {
		t.Fatal("expected 'stock' property to exist")
	}
	if string(stockProp.Minimum) != "0" {
		t.Errorf("expected stock minimum 0, got %v", stockProp.Minimum)
	}
	if string(stockProp.Maximum) != "1000" {
		t.Errorf("expected stock maximum 1000, got %v", stockProp.Maximum)
	}

	// Check discount constraints (min/max should also map to minimum/maximum)
	discountProp, _ := schema.Properties.Get("discount")
	if discountProp == nil {
		t.Fatal("expected 'discount' property to exist")
	}
	if string(discountProp.Minimum) != "0" {
		t.Errorf("expected discount minimum 0, got %v", discountProp.Minimum)
	}
	if string(discountProp.Maximum) != "100" {
		t.Errorf("expected discount maximum 100, got %v", discountProp.Maximum)
	}
}

func TestSchema_StringConstraints(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"min=3,max=20"`
		Bio      string `json:"bio" pedantigo:"max=500"`
	}

	validator := New[User]()
	schema := validator.Schema()

	// Check username constraints
	usernameProp, _ := schema.Properties.Get("username")
	if usernameProp == nil {
		t.Fatal("expected 'username' property to exist")
	}
	if usernameProp.MinLength == nil || *usernameProp.MinLength != 3 {
		t.Errorf("expected username minLength 3, got %v", usernameProp.MinLength)
	}
	if usernameProp.MaxLength == nil || *usernameProp.MaxLength != 20 {
		t.Errorf("expected username maxLength 20, got %v", usernameProp.MaxLength)
	}

	// Check bio constraints
	bioProp, _ := schema.Properties.Get("bio")
	if bioProp == nil {
		t.Fatal("expected 'bio' property to exist")
	}
	if bioProp.MaxLength == nil || *bioProp.MaxLength != 500 {
		t.Errorf("expected bio maxLength 500, got %v", bioProp.MaxLength)
	}
}

func TestSchema_FormatConstraints(t *testing.T) {
	type Contact struct {
		Email    string `json:"email" pedantigo:"email"`
		Website  string `json:"website" pedantigo:"url"`
		ID       string `json:"id" pedantigo:"uuid"`
		IPv4Addr string `json:"ipv4" pedantigo:"ipv4"`
		IPv6Addr string `json:"ipv6" pedantigo:"ipv6"`
	}

	validator := New[Contact]()
	schema := validator.Schema()

	// Check email format
	emailProp, _ := schema.Properties.Get("email")
	if emailProp == nil {
		t.Fatal("expected 'email' property to exist")
	}
	if emailProp.Format != "email" {
		t.Errorf("expected email format 'email', got %s", emailProp.Format)
	}

	// Check website format (url → uri)
	websiteProp, _ := schema.Properties.Get("website")
	if websiteProp == nil {
		t.Fatal("expected 'website' property to exist")
	}
	if websiteProp.Format != "uri" {
		t.Errorf("expected website format 'uri', got %s", websiteProp.Format)
	}

	// Check ID format
	idProp, _ := schema.Properties.Get("id")
	if idProp == nil {
		t.Fatal("expected 'id' property to exist")
	}
	if idProp.Format != "uuid" {
		t.Errorf("expected id format 'uuid', got %s", idProp.Format)
	}

	// Check IPv4 format
	ipv4Prop, _ := schema.Properties.Get("ipv4")
	if ipv4Prop == nil {
		t.Fatal("expected 'ipv4' property to exist")
	}
	if ipv4Prop.Format != "ipv4" {
		t.Errorf("expected ipv4 format 'ipv4', got %s", ipv4Prop.Format)
	}

	// Check IPv6 format
	ipv6Prop, _ := schema.Properties.Get("ipv6")
	if ipv6Prop == nil {
		t.Fatal("expected 'ipv6' property to exist")
	}
	if ipv6Prop.Format != "ipv6" {
		t.Errorf("expected ipv6 format 'ipv6', got %s", ipv6Prop.Format)
	}
}

func TestSchema_RegexConstraint(t *testing.T) {
	type Code struct {
		ZipCode string `json:"zipCode" pedantigo:"regexp=^[0-9]{5}$"`
	}

	validator := New[Code]()
	schema := validator.Schema()

	zipProp, _ := schema.Properties.Get("zipCode")
	if zipProp == nil {
		t.Fatal("expected 'zipCode' property to exist")
	}
	if zipProp.Pattern != "^[0-9]{5}$" {
		t.Errorf("expected zipCode pattern '^[0-9]{5}$', got %s", zipProp.Pattern)
	}
}

func TestSchema_NestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type User struct {
		Name    string  `json:"name" pedantigo:"required"`
		Address Address `json:"address"`
	}

	validator := New[User]()
	schema := validator.Schema()

	// Check that address property exists and is an object
	addressProp, _ := schema.Properties.Get("address")
	if addressProp == nil {
		t.Fatal("expected 'address' property to exist")
	}
	if addressProp.Type != "object" {
		t.Errorf("expected address type 'object', got %s", addressProp.Type)
	}

	// Check nested properties
	cityProp, _ := addressProp.Properties.Get("city")
	if cityProp == nil {
		t.Fatal("expected nested 'city' property to exist")
	}

	// Check nested required fields
	hasRequiredCity := false
	for _, req := range addressProp.Required {
		if req == "city" {
			hasRequiredCity = true
		}
	}
	if !hasRequiredCity {
		t.Error("expected 'city' to be required in nested Address")
	}
}

func TestSchema_SliceType(t *testing.T) {
	type Config struct {
		Tags   []string `json:"tags" pedantigo:"min=3"`
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	schema := validator.Schema()

	// Check tags array
	tagsProp, _ := schema.Properties.Get("tags")
	if tagsProp == nil {
		t.Fatal("expected 'tags' property to exist")
	}
	if tagsProp.Type != "array" {
		t.Errorf("expected tags type 'array', got %s", tagsProp.Type)
	}
	if tagsProp.Items == nil {
		t.Fatal("expected tags items to be defined")
	}
	if tagsProp.Items.Type != "string" {
		t.Errorf("expected tags items type 'string', got %s", tagsProp.Items.Type)
	}
	// Check that min_length constraint is applied to items
	if tagsProp.Items.MinLength == nil || *tagsProp.Items.MinLength != 3 {
		t.Errorf("expected tags items minLength 3, got %v", tagsProp.Items.MinLength)
	}

	// Check admins array with email format
	adminsProp, _ := schema.Properties.Get("admins")
	if adminsProp == nil {
		t.Fatal("expected 'admins' property to exist")
	}
	if adminsProp.Items == nil {
		t.Fatal("expected admins items to be defined")
	}
	if adminsProp.Items.Format != "email" {
		t.Errorf("expected admins items format 'email', got %s", adminsProp.Items.Format)
	}
}

func TestSchema_MapType(t *testing.T) {
	type Config struct {
		Settings map[string]string `json:"settings" pedantigo:"min=3"`
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	schema := validator.Schema()

	// Check settings map
	settingsProp, _ := schema.Properties.Get("settings")
	if settingsProp == nil {
		t.Fatal("expected 'settings' property to exist")
	}
	if settingsProp.Type != "object" {
		t.Errorf("expected settings type 'object', got %s", settingsProp.Type)
	}
	if settingsProp.AdditionalProperties == nil {
		t.Fatal("expected settings additionalProperties to be defined")
	}
	// Check that min_length constraint is applied to values
	if settingsProp.AdditionalProperties.MinLength == nil || *settingsProp.AdditionalProperties.MinLength != 3 {
		t.Errorf("expected settings values minLength 3, got %v", settingsProp.AdditionalProperties.MinLength)
	}

	// Check contacts map with email format
	contactsProp, _ := schema.Properties.Get("contacts")
	if contactsProp == nil {
		t.Fatal("expected 'contacts' property to exist")
	}
	if contactsProp.AdditionalProperties == nil {
		t.Fatal("expected contacts additionalProperties to be defined")
	}
	if contactsProp.AdditionalProperties.Format != "email" {
		t.Errorf("expected contacts values format 'email', got %s", contactsProp.AdditionalProperties.Format)
	}
}

func TestSchema_DefaultValues(t *testing.T) {
	type Config struct {
		Port    int    `json:"port" pedantigo:"default=8080"`
		Host    string `json:"host" pedantigo:"default=localhost"`
		Enabled bool   `json:"enabled" pedantigo:"default=true"`
	}

	validator := New[Config]()
	schema := validator.Schema()

	// Check port default
	portProp, _ := schema.Properties.Get("port")
	if portProp == nil {
		t.Fatal("expected 'port' property to exist")
	}
	portDefault, err := json.Marshal(portProp.Default)
	if err != nil || string(portDefault) != "8080" {
		t.Errorf("expected port default 8080, got %v", portProp.Default)
	}

	// Check host default
	hostProp, _ := schema.Properties.Get("host")
	if hostProp == nil {
		t.Fatal("expected 'host' property to exist")
	}
	if hostProp.Default != "localhost" {
		t.Errorf("expected host default 'localhost', got %v", hostProp.Default)
	}

	// Check enabled default
	enabledProp, _ := schema.Properties.Get("enabled")
	if enabledProp == nil {
		t.Fatal("expected 'enabled' property to exist")
	}
	enabledDefault, err := json.Marshal(enabledProp.Default)
	if err != nil || string(enabledDefault) != "true" {
		t.Errorf("expected enabled default true, got %v", enabledProp.Default)
	}
}
