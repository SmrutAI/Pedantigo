package pedantigo

import (
	"testing"
)

// ==================================================
// min_length constraint tests
// ==================================================

func TestMinLength_Valid(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"min=3"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":"alice"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for username with 5 chars (min 3), got %v", errs)
	}

	if user.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", user.Username)
	}
}

func TestMinLength_ExactlyAtMinimum(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"min=3"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":"bob"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for username with exactly 3 chars, got %v", errs)
	}

	if user.Username != "bob" {
		t.Errorf("expected username 'bob', got %q", user.Username)
	}
}

func TestMinLength_BelowMinimum(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"min=3"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":"ab"}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for username with 2 chars (min 3)")
	}

	// Check error message
	foundError := false
	for _, err := range errs {
		if err.Field == "Username" && err.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at least 3 characters' error, got %v", errs)
	}
}

func TestMinLength_EmptyString(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"min=1"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":""}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for empty username (min 1)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Username" && err.Message == "must be at least 1 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at least 1 characters' error, got %v", errs)
	}
}

func TestMinLength_WithPointer(t *testing.T) {
	type User struct {
		Bio *string `json:"bio" pedantigo:"min=10"`
	}

	validator := New[User]()
	shortBio := "short"
	jsonData := []byte(`{"bio":"short"}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for bio with 5 chars (min 10)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Bio" && err.Message == "must be at least 10 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at least 10 characters' error, got %v", errs)
	}

	// Test with valid length
	jsonData = []byte(`{"bio":"this is a longer bio"}`)
	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for bio with 21 chars, got %v", errs)
	}

	if user.Bio == nil || *user.Bio != "this is a longer bio" {
		t.Errorf("expected bio 'this is a longer bio', got %v", user.Bio)
	}

	_ = shortBio // Suppress unused variable warning
}

func TestMinLength_NilPointer(t *testing.T) {
	type User struct {
		Bio *string `json:"bio" pedantigo:"min=10"`
	}

	validator := New[User]()
	jsonData := []byte(`{"bio":null}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", errs)
	}

	if user.Bio != nil {
		t.Errorf("expected nil Bio pointer, got %v", user.Bio)
	}
}

// ==================================================
// max_length constraint tests
// ==================================================

func TestMaxLength_Valid(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"max=10"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":"alice"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for username with 5 chars (max 10), got %v", errs)
	}

	if user.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", user.Username)
	}
}

func TestMaxLength_ExactlyAtMaximum(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"max=5"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":"alice"}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for username with exactly 5 chars, got %v", errs)
	}

	if user.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", user.Username)
	}
}

func TestMaxLength_AboveMaximum(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"max=5"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":"verylongusername"}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for username with 16 chars (max 5)")
	}

	// Check error message
	foundError := false
	for _, err := range errs {
		if err.Field == "Username" && err.Message == "must be at most 5 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at most 5 characters' error, got %v", errs)
	}
}

func TestMaxLength_EmptyString(t *testing.T) {
	type User struct {
		Username string `json:"username" pedantigo:"max=10"`
	}

	validator := New[User]()
	jsonData := []byte(`{"username":""}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty username (max 10), got %v", errs)
	}

	if user.Username != "" {
		t.Errorf("expected empty username, got %q", user.Username)
	}
}

func TestMaxLength_WithPointer(t *testing.T) {
	type User struct {
		Bio *string `json:"bio" pedantigo:"max=20"`
	}

	validator := New[User]()
	jsonData := []byte(`{"bio":"this is a very long biography that exceeds the maximum"}`)

	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for bio with 58 chars (max 20)")
	}

	foundError := false
	for _, err := range errs {
		if err.Field == "Bio" && err.Message == "must be at most 20 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be at most 20 characters' error, got %v", errs)
	}

	// Test with valid length
	jsonData = []byte(`{"bio":"short bio"}`)
	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for bio with 9 chars, got %v", errs)
	}

	if user.Bio == nil || *user.Bio != "short bio" {
		t.Errorf("expected bio 'short bio', got %v", user.Bio)
	}
}

func TestMaxLength_NilPointer(t *testing.T) {
	type User struct {
		Bio *string `json:"bio" pedantigo:"max=20"`
	}

	validator := New[User]()
	jsonData := []byte(`{"bio":null}`)

	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil pointer (validation skips nil), got %v", errs)
	}

	if user.Bio != nil {
		t.Errorf("expected nil Bio pointer, got %v", user.Bio)
	}
}

func TestMaxLength_Combined_MinMaxLength(t *testing.T) {
	type User struct {
		Password string `json:"password" pedantigo:"min=8,max=20"`
	}

	validator := New[User]()

	// Test too short
	jsonData := []byte(`{"password":"short"}`)
	_, errs := validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for password with 5 chars (min 8)")
	}

	// Test too long
	jsonData = []byte(`{"password":"thispasswordiswaytoolongforourvalidation"}`)
	_, errs = validator.Unmarshal(jsonData)
	if len(errs) == 0 {
		t.Error("expected validation error for password with 43 chars (max 20)")
	}

	// Test in range
	jsonData = []byte(`{"password":"goodpassword"}`)
	user, errs := validator.Unmarshal(jsonData)
	if len(errs) != 0 {
		t.Errorf("expected no errors for password with 12 chars (8-20), got %v", errs)
	}

	if user.Password != "goodpassword" {
		t.Errorf("expected password 'goodpassword', got %q", user.Password)
	}
}
