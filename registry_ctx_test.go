package pedantigo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// contextKey is a type for context keys to satisfy SA1029.
type contextKey string

const (
	ctxKeyDB          contextKey = "db"
	ctxKeyTenantID    contextKey = "tenant_id"
	ctxKeyPermissions contextKey = "permissions"
	ctxKeyMultiplier  contextKey = "multiplier"
	ctxKeyEcho        contextKey = "echo_key"
	ctxKeyTest        contextKey = "test_key"
)

func TestRegisterValidationCtx(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		fn        ValidationFuncCtx
		wantErr   bool
		errSubstr string
	}{
		{
			name: "register valid context validator",
			tag:  "db_unique_ctx",
			fn: func(ctx context.Context, value any, param string) error {
				db := ctx.Value(ctxKeyDB)
				if db == nil {
					return errors.New("no database in context")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "register with empty tag",
			tag:  "",
			fn: func(ctx context.Context, value any, param string) error {
				return nil
			},
			wantErr:   true,
			errSubstr: "empty",
		},
		{
			name:      "register with nil function",
			tag:       "nil_func",
			fn:        nil,
			wantErr:   true,
			errSubstr: "nil",
		},
		{
			name: "register duplicate tag (should overwrite or error)",
			tag:  "duplicate_tag_ctx",
			fn: func(ctx context.Context, value any, param string) error {
				return nil
			},
			wantErr: false, // First registration
		},
		{
			name: "validator with parameter parsing",
			tag:  "ctx_param_test",
			fn: func(ctx context.Context, value any, param string) error {
				if param == "" {
					return errors.New("parameter required")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterValidationCtx(tt.tag, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterValidationCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errSubstr != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
			}
		})
	}
}

func TestValidateCtx_BasicUsage(t *testing.T) {
	// Register a context-aware validator
	err := RegisterValidationCtx("ctx_required_db", func(ctx context.Context, value any, param string) error {
		db := ctx.Value(ctxKeyDB)
		if db == nil {
			return errors.New("database connection required in context")
		}
		// Simulate DB check
		if str, ok := value.(string); ok && str == "duplicate" {
			return errors.New("value already exists in database")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register validation: %v", err)
	}

	type User struct {
		Username string `json:"username" pedantigo:"required,ctx_required_db"`
	}

	tests := []struct {
		name    string
		ctx     context.Context
		user    *User
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid context with database",
			ctx:     context.WithValue(context.Background(), ctxKeyDB, "mock-db"),
			user:    &User{Username: "john"},
			wantErr: false,
		},
		{
			name:    "missing database in context",
			ctx:     context.Background(),
			user:    &User{Username: "john"},
			wantErr: true,
			errMsg:  "database",
		},
		{
			name:    "duplicate value in database",
			ctx:     context.WithValue(context.Background(), ctxKeyDB, "mock-db"),
			user:    &User{Username: "duplicate"},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCtx(tt.ctx, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestValidateCtx_WithTimeout(t *testing.T) {
	// Register a slow validator
	err := RegisterValidationCtx("slow_check", func(ctx context.Context, value any, param string) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err != nil {
		t.Fatalf("Failed to register validation: %v", err)
	}

	type SlowUser struct {
		ID string `json:"id" pedantigo:"slow_check"`
	}

	t.Run("context timeout triggers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		user := &SlowUser{ID: "test"}
		err := ValidateCtx(ctx, user)
		// Should return context deadline exceeded error
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	t.Run("context with sufficient timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		user := &SlowUser{ID: "test"}
		err := ValidateCtx(ctx, user)
		if err != nil {
			t.Errorf("ValidateCtx() should succeed with sufficient timeout, got error: %v", err)
		}
	})
}

func TestValidateCtx_WithCancellation(t *testing.T) {
	err := RegisterValidationCtx("cancellable", func(ctx context.Context, value any, param string) error {
		select {
		case <-time.After(50 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return errors.New("validation cancelled")
		}
	})
	if err != nil {
		t.Fatalf("Failed to register validation: %v", err)
	}

	type CancellableUser struct {
		Name string `json:"name" pedantigo:"cancellable"`
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	user := &CancellableUser{Name: "test"}
	err = ValidateCtx(ctx, user)
	if err == nil {
		t.Error("Expected cancellation error, got nil")
	}
}

func TestValidateCtx_MultipleFields(t *testing.T) {
	// Register multiple context-aware validators
	err := RegisterValidationCtx("check_tenant", func(ctx context.Context, value any, param string) error {
		tenant := ctx.Value(ctxKeyTenantID)
		if tenant == nil {
			return errors.New("tenant_id required")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register check_tenant: %v", err)
	}

	err = RegisterValidationCtx("check_permission", func(ctx context.Context, value any, param string) error {
		perms := ctx.Value(ctxKeyPermissions)
		if perms == nil {
			return errors.New("permissions required")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register check_permission: %v", err)
	}

	type SecureResource struct {
		TenantID string `json:"tenant_id" pedantigo:"required,check_tenant"`
		Action   string `json:"action" pedantigo:"required,check_permission"`
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKeyTenantID, "tenant-123")
	ctx = context.WithValue(ctx, ctxKeyPermissions, []string{"read", "write"})

	resource := &SecureResource{
		TenantID: "tenant-123",
		Action:   "read",
	}

	err = ValidateCtx(ctx, resource)
	if err != nil {
		t.Errorf("ValidateCtx() should pass with all context values, got error: %v", err)
	}
}

func TestValidateCtx_WithParameter(t *testing.T) {
	// Register validator that uses parameter
	err := RegisterValidationCtx("min_ctx", func(ctx context.Context, value any, param string) error {
		multiplier := ctx.Value(ctxKeyMultiplier)
		if multiplier == nil {
			return errors.New("multiplier required in context")
		}
		// Simple parameter validation
		if param == "" {
			return errors.New("parameter required")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register min_ctx: %v", err)
	}

	type ParamUser struct {
		Score int `json:"score" pedantigo:"min_ctx=10"`
	}

	ctx := context.WithValue(context.Background(), ctxKeyMultiplier, 2)
	user := &ParamUser{Score: 15}

	err = ValidateCtx(ctx, user)
	if err != nil {
		t.Errorf("ValidateCtx() should pass with parameter, got error: %v", err)
	}
}

func TestValidator_ValidateCtx(t *testing.T) {
	// Test Validator[T].ValidateCtx method
	err := RegisterValidationCtx("ctx_validator_test", func(ctx context.Context, value any, param string) error {
		if ctx.Value(ctxKeyTest) == nil {
			return errors.New("test_key missing")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register validation: %v", err)
	}

	type TestStruct struct {
		Field string `json:"field" pedantigo:"ctx_validator_test"`
	}

	v := New[TestStruct]()

	t.Run("with context value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ctxKeyTest, "test_value")
		data := &TestStruct{Field: "test"}
		err := v.ValidateCtx(ctx, data)
		if err != nil {
			t.Errorf("ValidateCtx() should pass with context value, got error: %v", err)
		}
	})

	t.Run("without context value", func(t *testing.T) {
		ctx := context.Background()
		data := &TestStruct{Field: "test"}
		err := v.ValidateCtx(ctx, data)
		if err == nil {
			t.Error("ValidateCtx() should fail without context value")
		}
	})
}

func TestValidateCtx_NilContext(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"required"`
	}

	user := &User{Name: "John"}

	// Passing TODO context for testing nil handling behavior
	err := ValidateCtx(context.TODO(), user)
	// This should work with an empty context
	_ = err
}

func TestValidateCtx_NilStruct(t *testing.T) {
	ctx := context.Background()
	var user *struct {
		Name string `json:"name" pedantigo:"required"`
	}

	err := ValidateCtx(ctx, user)
	// Should handle nil struct gracefully
	_ = err
}

func TestRegisterValidationCtx_ConcurrentRegistration(t *testing.T) {
	// Test concurrent registration (should be thread-safe)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			tag := "concurrent_test"
			err := RegisterValidationCtx(tag, func(ctx context.Context, value any, param string) error {
				return nil
			})
			// First goroutine should succeed, others might error or overwrite
			_ = err
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestValidateCtx_ContextValues(t *testing.T) {
	// Test that context values are properly passed through
	err := RegisterValidationCtx("echo_context", func(ctx context.Context, value any, param string) error {
		v := ctx.Value(ctxKeyEcho)
		if v != "echo_value" {
			return errors.New("context value mismatch")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register validation: %v", err)
	}

	type EchoStruct struct {
		Data string `json:"data" pedantigo:"echo_context"`
	}

	ctx := context.WithValue(context.Background(), ctxKeyEcho, "echo_value")
	data := &EchoStruct{Data: "test"}

	err = ValidateCtx(ctx, data)
	if err != nil {
		t.Errorf("ValidateCtx() failed to pass context values: %v", err)
	}
}
