package pedantigo

import (
	"testing"
)

// ==================================================
// gt (greater than) constraint tests
// ==================================================

func TestGt(t *testing.T) {
	tests := []struct {
		name         string
		valueType    string // "int", "float64", "uint", "intPtr"
		fieldName    string
		jsonValue    string
		expectErr    bool
		expectVal    any
		expectNil    bool
		expectErrMsg string
	}{
		// int tests
		{
			name:      "int valid above threshold",
			valueType: "int",
			fieldName: "Stock",
			jsonValue: "5",
			expectErr: false,
			expectVal: 5,
		},
		{
			name:         "int equal to threshold",
			valueType:    "int",
			fieldName:    "Stock",
			jsonValue:    "0",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		{
			name:         "int below threshold",
			valueType:    "int",
			fieldName:    "Stock",
			jsonValue:    "-5",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		// float64 tests
		{
			name:      "float64 valid above threshold",
			valueType: "float64",
			fieldName: "Price",
			jsonValue: "19.99",
			expectErr: false,
			expectVal: 19.99,
		},
		{
			name:         "float64 below threshold",
			valueType:    "float64",
			fieldName:    "Price",
			jsonValue:    "-1.5",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		// uint tests
		{
			name:      "uint valid above threshold",
			valueType: "uint",
			fieldName: "Port",
			jsonValue: "8080",
			expectErr: false,
			expectVal: uint(8080),
		},
		// pointer tests
		{
			name:         "intPtr with invalid value",
			valueType:    "intPtr",
			fieldName:    "Stock",
			jsonValue:    "0",
			expectErr:    true,
			expectErrMsg: "must be greater than 0",
		},
		{
			name:      "intPtr with valid value",
			valueType: "intPtr",
			fieldName: "Stock",
			jsonValue: "10",
			expectErr: false,
			expectVal: 10,
		},
		{
			name:      "intPtr with nil value",
			valueType: "intPtr",
			fieldName: "Stock",
			jsonValue: "null",
			expectErr: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.valueType {
			case "int":
				type Product struct {
					Stock int `json:"stock" pedantigo:"gt=0"`
				}

				validator := New[Product]()
				jsonData := []byte(`{"stock":` + tt.jsonValue + `}`)
				product, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					if err == nil {
						t.Fatalf("expected validation error, got nil")
					}

					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					if !foundError {
						t.Errorf("expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
					}
				} else {
					if err != nil {
						t.Errorf("expected no errors, got %v", err)
					}

					if product.Stock != tt.expectVal.(int) {
						t.Errorf("expected %v, got %v", tt.expectVal, product.Stock)
					}
				}

			case "float64":
				type Product struct {
					Price float64 `json:"price" pedantigo:"gt=0"`
				}

				validator := New[Product]()
				jsonData := []byte(`{"price":` + tt.jsonValue + `}`)
				product, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					if err == nil {
						t.Fatalf("expected validation error, got nil")
					}

					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					if !foundError {
						t.Errorf("expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
					}
				} else {
					if err != nil {
						t.Errorf("expected no errors, got %v", err)
					}

					if product.Price != tt.expectVal.(float64) {
						t.Errorf("expected %v, got %v", tt.expectVal, product.Price)
					}
				}

			case "uint":
				type Config struct {
					Port uint `json:"port" pedantigo:"gt=1024"`
				}

				validator := New[Config]()
				jsonData := []byte(`{"port":` + tt.jsonValue + `}`)
				config, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					if err == nil {
						t.Fatalf("expected validation error, got nil")
					}

					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					if !foundError {
						t.Errorf("expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
					}
				} else {
					if err != nil {
						t.Errorf("expected no errors, got %v", err)
					}

					if config.Port != tt.expectVal.(uint) {
						t.Errorf("expected %v, got %v", tt.expectVal, config.Port)
					}
				}

			case "intPtr":
				type Product struct {
					Stock *int `json:"stock" pedantigo:"gt=0"`
				}

				validator := New[Product]()
				jsonData := []byte(`{"stock":` + tt.jsonValue + `}`)
				product, err := validator.Unmarshal(jsonData)

				if tt.expectErr {
					if err == nil {
						t.Fatalf("expected validation error, got nil")
					}

					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}

					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.fieldName && fieldErr.Message == tt.expectErrMsg {
							foundError = true
							break
						}
					}

					if !foundError {
						t.Errorf("expected error message %q, got %v", tt.expectErrMsg, ve.Errors)
					}
				} else {
					if err != nil {
						t.Errorf("expected no errors, got %v", err)
					}

					if tt.expectNil {
						if product.Stock != nil {
							t.Errorf("expected nil pointer, got %v", product.Stock)
						}
					} else {
						if product.Stock == nil {
							t.Errorf("expected non-nil pointer, got nil")
						} else if *product.Stock != tt.expectVal.(int) {
							t.Errorf("expected %v, got %v", tt.expectVal, *product.Stock)
						}
					}
				}
			}
		})
	}
}

// ==================================================
// ge (greater or equal) constraint tests
// ==================================================

func TestGe(t *testing.T) {
	type Product struct {
		Stock int `json:"stock" pedantigo:"gte=0"`
	}

	tests := []struct {
		name            string
		jsonData        []byte
		expectedValue   int
		expectError     bool
		expectedMessage string
	}{
		{
			name:          "int valid above threshold",
			jsonData:      []byte(`{"stock":5}`),
			expectedValue: 5,
			expectError:   false,
		},
		{
			name:          "int equal to threshold",
			jsonData:      []byte(`{"stock":0}`),
			expectedValue: 0,
			expectError:   false,
		},
		{
			name:            "int below threshold",
			jsonData:        []byte(`{"stock":-1}`),
			expectError:     true,
			expectedMessage: "must be at least 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Product]()
			product, err := validator.Unmarshal(tt.jsonData)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected validation error, got none")
				}

				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Stock" && fieldErr.Message == tt.expectedMessage {
						foundError = true
						break
					}
				}

				if !foundError {
					t.Errorf("expected error message %q, got %v", tt.expectedMessage, ve.Errors)
				}
			} else {
				if err != nil {
					t.Errorf("expected no errors, got %v", err)
				}

				if product.Stock != tt.expectedValue {
					t.Errorf("expected stock %d, got %d", tt.expectedValue, product.Stock)
				}
			}
		})
	}
}

// ==================================================
// lt (less than) constraint tests
// ==================================================

func TestLt(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lt=100"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		expectedErr bool
		expectedVal int
	}{
		{
			name:        "int valid below threshold",
			jsonData:    []byte(`{"discount":50}`),
			expectedErr: false,
			expectedVal: 50,
		},
		{
			name:        "int equal to threshold",
			jsonData:    []byte(`{"discount":100}`),
			expectedErr: true,
			expectedVal: 0,
		},
		{
			name:        "int above threshold",
			jsonData:    []byte(`{"discount":150}`),
			expectedErr: true,
			expectedVal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Product]()
			product, err := validator.Unmarshal(tt.jsonData)

			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}

				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Discount" && fieldErr.Message == "must be less than 100" {
						foundError = true
					}
				}

				if !foundError {
					t.Errorf("expected 'must be less than 100' error, got %v", ve.Errors)
				}
			} else {
				if err != nil {
					t.Errorf("expected no errors, got %v", err)
				}

				if product.Discount != tt.expectedVal {
					t.Errorf("expected discount %d, got %d", tt.expectedVal, product.Discount)
				}
			}
		})
	}
}

// ==================================================
// le (less or equal) constraint tests
// ==================================================

func TestLe(t *testing.T) {
	type Product struct {
		Discount int `json:"discount" pedantigo:"lte=100"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		expectedErr bool
		expectedVal int
	}{
		{
			name:        "int valid below threshold",
			jsonData:    []byte(`{"discount":50}`),
			expectedErr: false,
			expectedVal: 50,
		},
		{
			name:        "int equal to threshold",
			jsonData:    []byte(`{"discount":100}`),
			expectedErr: false,
			expectedVal: 100,
		},
		{
			name:        "int above threshold",
			jsonData:    []byte(`{"discount":150}`),
			expectedErr: true,
			expectedVal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Product]()
			product, err := validator.Unmarshal(tt.jsonData)

			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}

				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "Discount" && fieldErr.Message == "must be at most 100" {
						foundError = true
					}
				}

				if !foundError {
					t.Errorf("expected 'must be at most 100' error, got %v", ve.Errors)
				}
			} else {
				if err != nil {
					t.Errorf("expected no errors, got %v", err)
				}

				if product.Discount != tt.expectedVal {
					t.Errorf("expected discount %d, got %d", tt.expectedVal, product.Discount)
				}
			}
		})
	}
}
