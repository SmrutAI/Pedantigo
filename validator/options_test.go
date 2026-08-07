package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Tests for resolveTagName
// ============================================================================

// TestResolveTagName_InstanceOverridesGlobal verifies instance TagName takes precedence.
func TestResolveTagName_InstanceOverridesGlobal(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Set global to "custom_validate"
	SetTagName("custom_validate")

	// Instance with TagName should override global
	opts := Options{TagName: "binding"}
	resolved := resolveTagName(opts)

	assert.Equal(t, "binding", resolved, "instance TagName should override global")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestResolveTagName_EmptyInstanceUsesGlobal verifies empty instance uses global.
func TestResolveTagName_EmptyInstanceUsesGlobal(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Set global to "custom_validate"
	SetTagName("custom_validate")

	// Instance without TagName should use global
	opts := Options{} // TagName is empty string
	resolved := resolveTagName(opts)

	assert.Equal(t, "custom_validate", resolved, "empty instance TagName should use global")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestResolveTagName_DefaultGlobal verifies default global is "validate".
func TestResolveTagName_DefaultGlobal(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Global should be default "validate"
	opts := Options{}
	resolved := resolveTagName(opts)

	assert.Equal(t, "validate", resolved, "default global should be 'validate'")
}

// TestDefaultOptions_TagNameEmpty verifies default options has empty TagName.
func TestDefaultOptions_TagNameEmpty(t *testing.T) {
	opts := DefaultOptions()
	assert.Empty(t, opts.TagName, "default options should have empty TagName")
}
