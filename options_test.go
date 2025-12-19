package pedantigo

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

	// Set global to "validate"
	SetTagName("validate")

	// Instance with TagName should override global
	opts := ValidatorOptions{TagName: "binding"}
	resolved := resolveTagName(opts)

	assert.Equal(t, "binding", resolved, "instance TagName should override global")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestResolveTagName_EmptyInstanceUsesGlobal verifies empty instance uses global.
func TestResolveTagName_EmptyInstanceUsesGlobal(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Set global to "validate"
	SetTagName("validate")

	// Instance without TagName should use global
	opts := ValidatorOptions{} // TagName is empty string
	resolved := resolveTagName(opts)

	assert.Equal(t, "validate", resolved, "empty instance TagName should use global")

	resetTagNameForTesting()
	resetValidatorCreatedForTesting()
}

// TestResolveTagName_DefaultGlobal verifies default global is "pedantigo".
func TestResolveTagName_DefaultGlobal(t *testing.T) {
	resetTagNameForTesting()
	resetValidatorCreatedForTesting()

	// Global should be default "pedantigo"
	opts := ValidatorOptions{}
	resolved := resolveTagName(opts)

	assert.Equal(t, "pedantigo", resolved, "default global should be 'pedantigo'")
}

// TestDefaultValidatorOptions_TagNameEmpty verifies default options has empty TagName.
func TestDefaultValidatorOptions_TagNameEmpty(t *testing.T) {
	opts := DefaultValidatorOptions()
	assert.Empty(t, opts.TagName, "default options should have empty TagName")
}
