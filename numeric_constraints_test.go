package pedantigo

import (
	"testing"
)

// ==================================================
// gt (greater than) constraint tests
// ==================================================

func TestGt_Int_Valid(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gt=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":5}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for stock=5 (gt 0), got %v", errs)
	}

	if product.Stock != 5 {
		t.Errorf("expected stock 5, got %d", product.Stock)
	}
}

func TestGt_Int_EqualToThreshold(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gt=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":0}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for stock=0 (not > 0)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Stock" && err.Message == "must be greater than 0" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be greater than 0' error, got %v", errs)
	}
}

func TestGt_Int_BelowThreshold(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gt=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":-5}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for stock=-5 (not > 0)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Stock" && err.Message == "must be greater than 0" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be greater than 0' error, got %v", errs)
	}
}

func TestGt_Float_Valid(t *testing.T) {
	type Product struct {
		Price float64 `json:"price" pedantigo:"gt=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"price":19.99}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for price=19.99 (gt 0), got %v", errs)
	}

	if product.Price != 19.99 {
		t.Errorf("expected price 19.99, got %f", product.Price)
	}
}

func TestGt_Float_BelowThreshold(t *testing.T) {
	type Product struct {
		Price float64 `json:"price" pedantigo:"gt=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"price":-1.5}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for price=-1.5 (not > 0)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Price" && err.Message == "must be greater than 0" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be greater than 0' error, got %v", errs)
	}
}

func TestGt_Uint_Valid(t *testing.T) {
	type Config struct {
		Port uint `json:"port" pedantigo:"gt=1024"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":8080}`)

	config, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for port=8080 (gt 1024), got %v", errs)
	}

	if config.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Port)
	}
}

func TestGt_WithPointer(t *testing.T) {
	type Product struct {
		Stock *int `json:"stock" pedantigo:"gt=0"`
	}

	validator := New[Product]()

	// Test invalid value
	jsonData := []byte(`{"stock":0}`)
	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for stock=0 (not > 0)")
	}

	// Test valid value
	jsonData = []byte(`{"stock":10}`)
	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for stock=10, got %v", errs)
	}

	if product.Stock == nil || *product.Stock != 10 {
		t.Errorf("expected stock=10, got %v", product.Stock)
	}
}

func TestGt_NilPointer(t *testing.T) {
	type Product struct {
		Stock *int `json:"stock" pedantigo:"gt=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":null}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", errs)
	}

	if product.Stock != nil {
		t.Errorf("expected nil Stock pointer, got %v", product.Stock)
	}
}

// ==================================================
// ge (greater or equal) constraint tests
// ==================================================

func TestGe_Int_Valid(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gte=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":5}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for stock=5 (ge 0), got %v", errs)
	}

	if product.Stock != 5 {
		t.Errorf("expected stock 5, got %d", product.Stock)
	}
}

func TestGe_Int_EqualToThreshold(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gte=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":0}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for stock=0 (ge 0), got %v", errs)
	}

	if product.Stock != 0 {
		t.Errorf("expected stock 0, got %d", product.Stock)
	}
}

func TestGe_Int_BelowThreshold(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gte=0"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"stock":-1}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for stock=-1 (not >= 0)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Stock" && err.Message == "must be at least 0" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at least 0' error, got %v", errs)
	}
}

// ==================================================
// lt (less than) constraint tests
// ==================================================

func TestLt_Int_Valid(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lt=100"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"discount":50}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for discount=50 (lt 100), got %v", errs)
	}

	if product.Discount != 50 {
		t.Errorf("expected discount 50, got %d", product.Discount)
	}
}

func TestLt_Int_EqualToThreshold(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lt=100"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"discount":100}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for discount=100 (not < 100)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Discount" && err.Message == "must be less than 100" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be less than 100' error, got %v", errs)
	}
}

func TestLt_Int_AboveThreshold(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lt=100"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"discount":150}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for discount=150 (not < 100)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Discount" && err.Message == "must be less than 100" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be less than 100' error, got %v", errs)
	}
}

// ==================================================
// le (less or equal) constraint tests
// ==================================================

func TestLe_Int_Valid(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lte=100"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"discount":50}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for discount=50 (le 100), got %v", errs)
	}

	if product.Discount != 50 {
		t.Errorf("expected discount 50, got %d", product.Discount)
	}
}

func TestLe_Int_EqualToThreshold(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lte=100"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"discount":100}`)

	product, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for discount=100 (le 100), got %v", errs)
	}

	if product.Discount != 100 {
		t.Errorf("expected discount 100, got %d", product.Discount)
	}
}

func TestLe_Int_AboveThreshold(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lte=100"`
	}

	validator := New[Product]()
	jsonData := []byte(`{"discount":150}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for discount=150 (not <= 100)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Discount" && err.Message == "must be at most 100" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at most 100' error, got %v", errs)
	}
}
