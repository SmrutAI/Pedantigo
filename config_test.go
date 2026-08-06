package pedantigo

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Tests for Global Tag Name Configuration
// ============================================================================

// TestGetTagName_Default verifies the default tag name is "validate".
func TestGetTagName_Default(t *testing.T) {
	// Reset to default for this test
	resetTagNameForTesting()

	tagName := GetTagName()
	assert.Equal(t, "validate", tagName, "default tag name should be 'validate'")
}

// TestSetTagName_Custom verifies setting a custom tag name.
func TestSetTagName_Custom(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting() // Must reset flag for SetTagName to work

	SetTagName("custom_validate")
	assert.Equal(t, "custom_validate", GetTagName())

	resetValidatorCreatedForTesting() // Reset again before next SetTagName
	SetTagName("binding")
	assert.Equal(t, "binding", GetTagName())

	// Reset back to default
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestSetTagName_Empty_FallsBack verifies empty string falls back to default.
func TestSetTagName_Empty_FallsBack(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	SetTagName("")
	assert.Equal(t, "validate", GetTagName(), "empty tag name should fall back to 'validate'")

	resetValidatorCreatedForTesting()
}

// TestGetTagName_ThreadSafe verifies concurrent access is safe.
func TestGetTagName_ThreadSafe(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Set a known value
	SetTagName("custom_validate")

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tagName := GetTagName()
			assert.Equal(t, "custom_validate", tagName)
		}()
	}

	wg.Wait()

	// Reset back to default
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// ============================================================================
// Tests for Panic Protection
// ============================================================================

// TestSetTagName_PanicsAfterValidatorCreated verifies that calling SetTagName
// after any validator has been created will panic.
// NOTE: This test uses resetValidatorCreatedForTesting to simulate fresh state.
func TestSetTagName_PanicsAfterValidatorCreated(t *testing.T) {
	// Reset to fresh state
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Create a validator - this marks validatorCreated as true
	_ = New[struct{ Name string }]()

	// Now SetTagName should panic
	assert.Panics(t, func() {
		SetTagName("custom_validate")
	}, "SetTagName should panic when called after validator creation")

	// Reset for other tests
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestSetTagName_WorksBeforeValidatorCreated verifies that SetTagName works
// when called before any validator is created.
func TestSetTagName_WorksBeforeValidatorCreated(t *testing.T) {
	// Reset to fresh state where no validators exist
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// In a fresh state, SetTagName should NOT panic
	assert.NotPanics(t, func() {
		SetTagName("custom")
	})
	assert.Equal(t, "custom", GetTagName())

	// Reset for other tests
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestValidatorCreatedFlag_SetOnNew verifies that creating a validator sets the flag.
func TestValidatorCreatedFlag_SetOnNew(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Before creating any validator, the flag should be false
	assert.False(t, hasValidatorBeenCreated(), "flag should be false before any validator is created")

	// Create a validator
	_ = New[struct{ Age int }]()

	// After creating a validator, the flag should be true
	assert.True(t, hasValidatorBeenCreated(), "flag should be true after validator is created")

	// Reset for other tests
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}
