package pedantigo

import (
	"fmt"
	"testing"
)

// ==================================================
// min_length constraint tests
// ==================================================

func TestMinLength(t *testing.T) {
	tests := []struct {
		name      string
		minVal    int
		fieldName string
		json      string
		usePtr    bool
		expectErr bool
		expectVal string
		expectNil bool
	}{
		{
			name:      "Valid above min",
			minVal:    3,
			fieldName: "Username",
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
		},
		{
			name:      "Exactly at min",
			minVal:    3,
			fieldName: "Username",
			json:      `{"username":"bob"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "bob",
			expectNil: false,
		},
		{
			name:      "Below min",
			minVal:    3,
			fieldName: "Username",
			json:      `{"username":"ab"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Empty string",
			minVal:    1,
			fieldName: "Username",
			json:      `{"username":""}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Pointer below min",
			minVal:    10,
			fieldName: "Bio",
			json:      `{"bio":"short"}`,
			usePtr:    true,
			expectErr: true,
			expectVal: "",
			expectNil: false,
		},
		{
			name:      "Pointer valid",
			minVal:    10,
			fieldName: "Bio",
			json:      `{"bio":"this is a longer bio"}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "this is a longer bio",
			expectNil: false,
		},
		{
			name:      "Nil pointer",
			minVal:    10,
			fieldName: "Bio",
			json:      `{"bio":null}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.usePtr {
				// Non-pointer test case
				type User struct {
					Username string `json:"username" pedantigo:"min=3"`
				}

				// For empty string test, use min=1
				if tt.minVal == 1 {
					type UserMin1 struct {
						Username string `json:"username" pedantigo:"min=1"`
					}
					validator := New[UserMin1]()
					_, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr && err == nil {
						t.Fatal("expected validation error, got none")
					}
					if !tt.expectErr && err != nil {
						t.Errorf("expected no error, got %v", err)
					}

					if tt.expectErr {
						ve, ok := err.(*ValidationError)
						if !ok {
							t.Fatalf("expected *ValidationError, got %T", err)
						}
						expectedMsg := "must be at least 1 characters"
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
								foundError = true
								break
							}
						}
						if !foundError {
							t.Errorf("expected error message %q, got %v", expectedMsg, ve.Errors)
						}
					}
				} else {
					validator := New[User]()
					user, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr && err == nil {
						t.Fatal("expected validation error, got none")
					}
					if !tt.expectErr && err != nil {
						t.Errorf("expected no error, got %v", err)
					}

					if !tt.expectErr {
						if user.Username != tt.expectVal {
							t.Errorf("expected value %q, got %q", tt.expectVal, user.Username)
						}
					}

					if tt.expectErr {
						ve, ok := err.(*ValidationError)
						if !ok {
							t.Fatalf("expected *ValidationError, got %T", err)
						}
						expectedMsg := "must be at least 3 characters"
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
								foundError = true
								break
							}
						}
						if !foundError {
							t.Errorf("expected error message %q, got %v", expectedMsg, ve.Errors)
						}
					}
				}
			} else {
				// Pointer test case
				type User struct {
					Bio *string `json:"bio" pedantigo:"min=10"`
				}

				validator := New[User]()
				user, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr && err == nil {
					t.Fatal("expected validation error, got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
				}

				if !tt.expectErr {
					if tt.expectNil {
						if user.Bio != nil {
							t.Errorf("expected nil pointer, got %v", user.Bio)
						}
					} else {
						if user.Bio == nil || *user.Bio != tt.expectVal {
							t.Errorf("expected value %q, got %v", tt.expectVal, user.Bio)
						}
					}
				}

				if tt.expectErr {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					expectedMsg := "must be at least 10 characters"
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Bio" && fieldErr.Message == expectedMsg {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error message %q, got %v", expectedMsg, ve.Errors)
					}
				}
			}
		})
	}
}

// ==================================================
// max_length constraint tests
// ==================================================

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		maxVal    int
		fieldName string
		json      string
		usePtr    bool
		expectErr bool
		expectVal string
		expectNil bool
		minVal    int // For combined min/max tests
	}{
		{
			name:      "Valid below max",
			maxVal:    10,
			fieldName: "Username",
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Exactly at max",
			maxVal:    5,
			fieldName: "Username",
			json:      `{"username":"alice"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "alice",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Above max",
			maxVal:    5,
			fieldName: "Username",
			json:      `{"username":"verylongusername"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Empty string",
			maxVal:    10,
			fieldName: "Username",
			json:      `{"username":""}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Pointer above max",
			maxVal:    20,
			fieldName: "Bio",
			json:      `{"bio":"this is a very long biography that exceeds the maximum"}`,
			usePtr:    true,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Pointer valid",
			maxVal:    20,
			fieldName: "Bio",
			json:      `{"bio":"short bio"}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "short bio",
			expectNil: false,
			minVal:    0,
		},
		{
			name:      "Nil pointer",
			maxVal:    20,
			fieldName: "Bio",
			json:      `{"bio":null}`,
			usePtr:    true,
			expectErr: false,
			expectVal: "",
			expectNil: true,
			minVal:    0,
		},
		{
			name:      "Combined min/max valid",
			maxVal:    20,
			fieldName: "Password",
			json:      `{"password":"goodpassword"}`,
			usePtr:    false,
			expectErr: false,
			expectVal: "goodpassword",
			expectNil: false,
			minVal:    8,
		},
		{
			name:      "Combined below min",
			maxVal:    20,
			fieldName: "Password",
			json:      `{"password":"short"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    8,
		},
		{
			name:      "Combined above max",
			maxVal:    20,
			fieldName: "Password",
			json:      `{"password":"thispasswordiswaytoolongforourvalidation"}`,
			usePtr:    false,
			expectErr: true,
			expectVal: "",
			expectNil: false,
			minVal:    8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.usePtr {
				// Non-pointer test case
				if tt.minVal > 0 {
					// Combined min/max test - use Password field with both constraints
					type UserWithPassword struct {
						Password string `json:"password" pedantigo:"min=8,max=20"`
					}
					validator := New[UserWithPassword]()
					user, err := validator.Unmarshal([]byte(tt.json))

					if tt.expectErr && err == nil {
						t.Fatal("expected validation error, got none")
					}
					if !tt.expectErr && err != nil {
						t.Errorf("expected no error, got %v", err)
					}

					if !tt.expectErr {
						if user.Password != tt.expectVal {
							t.Errorf("expected value %q, got %q", tt.expectVal, user.Password)
						}
					}

					if tt.expectErr {
						ve, ok := err.(*ValidationError)
						if !ok {
							t.Fatalf("expected *ValidationError, got %T", err)
						}
						foundError := false
						for _, fieldErr := range ve.Errors {
							if fieldErr.Field == "Password" {
								foundError = true
								break
							}
						}
						if !foundError {
							t.Errorf("expected error for Password field, got %v", ve.Errors)
						}
					}
				} else {
					// Max-only tests - use Username field with only max constraint
					type UserWithUsername struct {
						Username string `json:"username" pedantigo:"max=10"`
					}

					// Handle different max values
					if tt.maxVal == 5 {
						type UserMax5 struct {
							Username string `json:"username" pedantigo:"max=5"`
						}
						validator := New[UserMax5]()
						user, err := validator.Unmarshal([]byte(tt.json))

						if tt.expectErr && err == nil {
							t.Fatal("expected validation error, got none")
						}
						if !tt.expectErr && err != nil {
							t.Errorf("expected no error, got %v", err)
						}

						if !tt.expectErr {
							if user.Username != tt.expectVal {
								t.Errorf("expected value %q, got %q", tt.expectVal, user.Username)
							}
						}

						if tt.expectErr {
							ve, ok := err.(*ValidationError)
							if !ok {
								t.Fatalf("expected *ValidationError, got %T", err)
							}
							expectedMsg := fmt.Sprintf("must be at most %d characters", tt.maxVal)
							foundError := false
							for _, fieldErr := range ve.Errors {
								if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
									foundError = true
									break
								}
							}
							if !foundError {
								t.Errorf("expected error message %q, got %v", expectedMsg, ve.Errors)
							}
						}
					} else {
						validator := New[UserWithUsername]()
						user, err := validator.Unmarshal([]byte(tt.json))

						if tt.expectErr && err == nil {
							t.Fatal("expected validation error, got none")
						}
						if !tt.expectErr && err != nil {
							t.Errorf("expected no error, got %v", err)
						}

						if !tt.expectErr {
							if user.Username != tt.expectVal {
								t.Errorf("expected value %q, got %q", tt.expectVal, user.Username)
							}
						}

						if tt.expectErr {
							ve, ok := err.(*ValidationError)
							if !ok {
								t.Fatalf("expected *ValidationError, got %T", err)
							}
							expectedMsg := fmt.Sprintf("must be at most %d characters", tt.maxVal)
							foundError := false
							for _, fieldErr := range ve.Errors {
								if fieldErr.Field == "Username" && fieldErr.Message == expectedMsg {
									foundError = true
									break
								}
							}
							if !foundError {
								t.Errorf("expected error message %q, got %v", expectedMsg, ve.Errors)
							}
						}
					}
				}
			} else {
				// Pointer test case
				type User struct {
					Bio *string `json:"bio" pedantigo:"max=20"`
				}

				validator := New[User]()
				user, err := validator.Unmarshal([]byte(tt.json))

				if tt.expectErr && err == nil {
					t.Fatal("expected validation error, got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
				}

				if !tt.expectErr {
					if tt.expectNil {
						if user.Bio != nil {
							t.Errorf("expected nil pointer, got %v", user.Bio)
						}
					} else {
						if user.Bio == nil || *user.Bio != tt.expectVal {
							t.Errorf("expected value %q, got %v", tt.expectVal, user.Bio)
						}
					}
				}

				if tt.expectErr {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					expectedMsg := "must be at most 20 characters"
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == "Bio" && fieldErr.Message == expectedMsg {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error message %q, got %v", expectedMsg, ve.Errors)
					}
				}
			}
		})
	}
}
