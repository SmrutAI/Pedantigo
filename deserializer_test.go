package pedantigo

import (
	"testing"
	"time"
)

// Test type for defaultUsingMethod
type UserWithTimestamp struct {
	Email     string    `json:"email" pedantigo:"required"`
	Role      string    `json:"role" pedantigo:"default=user"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=SetCreationTime"`
}

// Method that provides dynamic default value
func (u *UserWithTimestamp) SetCreationTime() (time.Time, error) {
	// Return a fixed time for testing (not time.Now())
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), nil
}

// Test type with invalid method signature (should panic at New() time)
type InvalidMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=WrongSignature"`
}

// Wrong signature: returns only value, no error
func (i *InvalidMethodType) WrongSignature() time.Time {
	return time.Now()
}

// Test type with non-existent method
type NonExistentMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=DoesNotExist"`
}

// TestDeserializer_UnmarshalBehavior validates deserializer behavior across various scenarios:
// defaults, missing fields, explicit values, required fields, and validator options.
func TestDeserializer_UnmarshalBehavior(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"required"`
		Port    int    `json:"port" pedantigo:"default=8080"`
		Timeout int    `json:"timeout" pedantigo:"default=30"`
	}

	type Settings struct {
		Name   string `json:"name" pedantigo:"required"`
		Active bool   `json:"active" pedantigo:"required"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		validatorFn func() (any, error)
		wantErr     bool
		assertions  func(*testing.T, any)
	}{
		{
			name:     "missing fields with defaults",
			jsonData: []byte(`{"name":"myapp"}`),
			validatorFn: func() (any, error) {
				v := New[Config]()
				return v.Unmarshal([]byte(`{"name":"myapp"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				if config.Port != 8080 {
					t.Errorf("expected default port 8080, got %d", config.Port)
				}
				if config.Timeout != 30 {
					t.Errorf("expected default timeout 30, got %d", config.Timeout)
				}
				if config.Name != "myapp" {
					t.Errorf("expected name 'myapp', got %q", config.Name)
				}
			},
		},
		{
			name:     "explicit zero values not replaced with defaults",
			jsonData: []byte(`{"name":"myapp","port":0,"timeout":0}`),
			validatorFn: func() (any, error) {
				v := New[Config]()
				return v.Unmarshal([]byte(`{"name":"myapp","port":0,"timeout":0}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				// Explicit zeros should be kept, NOT replaced with defaults
				if config.Port != 0 {
					t.Errorf("expected port 0 (not default), got %d", config.Port)
				}
				if config.Timeout != 0 {
					t.Errorf("expected timeout 0 (not default), got %d", config.Timeout)
				}
			},
		},
		{
			name:     "explicit false value passes required validation",
			jsonData: []byte(`{"name":"test","active":false}`),
			validatorFn: func() (any, error) {
				v := New[Settings]()
				return v.Unmarshal([]byte(`{"name":"test","active":false}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				settings := result.(*Settings)
				if settings.Active != false {
					t.Errorf("expected active=false, got %v", settings.Active)
				}
			},
		},
		{
			name:     "missing required field fails validation",
			jsonData: []byte(`{"name":"test"}`),
			validatorFn: func() (any, error) {
				v := New[Settings]()
				return v.Unmarshal([]byte(`{"name":"test"}`))
			},
			wantErr: true,
			assertions: func(t *testing.T, result any) {
				// Error case - check error message through direct validation
				v := New[Settings]()
				_, err := v.Unmarshal([]byte(`{"name":"test"}`))
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "active" && fieldErr.Message == "is required" {
						foundError = true
					}
				}
				if !foundError {
					t.Errorf("expected 'is required' error for field 'active', got errors: %+v", ve.Errors)
				}
			},
		},
		{
			name:     "defaultUsingMethod called for missing fields",
			jsonData: []byte(`{"email":"test@example.com"}`),
			validatorFn: func() (any, error) {
				v := New[UserWithTimestamp]()
				return v.Unmarshal([]byte(`{"email":"test@example.com"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				user := result.(*UserWithTimestamp)
				expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				if !user.CreatedAt.Equal(expectedTime) {
					t.Errorf("expected created_at to be %v, got %v", expectedTime, user.CreatedAt)
				}
				if user.Role != "user" {
					t.Errorf("expected default role 'user', got %q", user.Role)
				}
			},
		},
		{
			name:     "defaultUsingMethod not called for explicit values",
			jsonData: []byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`),
			validatorFn: func() (any, error) {
				v := New[UserWithTimestamp]()
				return v.Unmarshal([]byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				user := result.(*UserWithTimestamp)
				explicitTime := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)
				if !user.CreatedAt.Equal(explicitTime) {
					t.Errorf("expected created_at to be %v (explicit), got %v", explicitTime, user.CreatedAt)
				}
			},
		},
		{
			name:     "strict mode applies defaults for missing fields",
			jsonData: []byte(`{"name":"myapp"}`),
			validatorFn: func() (any, error) {
				v := New[Config](ValidatorOptions{StrictMissingFields: true})
				return v.Unmarshal([]byte(`{"name":"myapp"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				if config.Port != 8080 {
					t.Errorf("expected port 8080 (default applied), got %d", config.Port)
				}
				if config.Timeout != 30 {
					t.Errorf("expected timeout 30 (default applied), got %d", config.Timeout)
				}
			},
		},
		{
			name:     "relaxed mode skips constraints on zero values",
			jsonData: []byte(`{"name":"myapp","age":0}`),
			validatorFn: func() (any, error) {
				type Profile struct {
					Name string `json:"name" pedantigo:"required"`
					Age  int    `json:"age" pedantigo:"min=1"`
				}
				v := New[Profile](ValidatorOptions{StrictMissingFields: false})
				return v.Unmarshal([]byte(`{"name":"myapp","age":0}`))
			},
			wantErr: true,
			assertions: func(t *testing.T, result any) {
				// Age=0 is explicit, so validation should still run and fail
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.validatorFn()

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil || !tt.wantErr {
				tt.assertions(t, result)
			}
		})
	}
}

// TestDeserializer_ValidatorSetup validates fail-fast validation during New().
// Invalid method signatures or non-existent methods should panic at validator creation time.
func TestDeserializer_ValidatorSetup(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectPanic bool
	}{
		{
			name: "invalid method signature panics",
			setup: func() {
				_ = New[InvalidMethodType]()
			},
			expectPanic: true,
		},
		{
			name: "non-existent method panics",
			setup: func() {
				_ = New[NonExistentMethodType]()
			},
			expectPanic: true,
		},
		{
			name: "valid method signature succeeds",
			setup: func() {
				_ = New[UserWithTimestamp]()
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic but none occurred")
					}
				}()
				tt.setup()
			} else {
				// Should not panic
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("unexpected panic: %v", r)
					}
				}()
				tt.setup()
			}
		})
	}
}
