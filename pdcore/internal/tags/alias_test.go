package tags

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandAlias_NilLookup(t *testing.T) {
	// Reset alias lookup to nil
	SetAliasLookup(nil)

	result, found := ExpandAlias("iscolor")
	assert.False(t, found)
	assert.Equal(t, "iscolor", result)
}

func TestExpandAlias_WithLookup(t *testing.T) {
	// Set up a test alias lookup
	SetAliasLookup(func(name string) (string, bool) {
		aliases := map[string]string{
			"iscolor": "hexcolor|rgb|rgba|hsl|hsla",
			"isuri":   "uri",
		}
		if expansion, ok := aliases[name]; ok {
			return expansion, true
		}
		return name, false
	})
	defer SetAliasLookup(nil) // Clean up

	tests := []struct {
		name      string
		input     string
		wantExp   string
		wantFound bool
	}{
		{"existing alias", "iscolor", "hexcolor|rgb|rgba|hsl|hsla", true},
		{"another alias", "isuri", "uri", true},
		{"non-existent alias", "unknown", "unknown", false},
		{"empty string", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := ExpandAlias(tt.input)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantExp, result)
		})
	}
}

func TestSetAliasLookup_Concurrent(t *testing.T) {
	// Test that SetAliasLookup and ExpandAlias are thread-safe
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writers
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			SetAliasLookup(func(name string) (string, bool) {
				return name + "_expanded", true
			})
		}(i)
	}

	// Concurrent readers
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = ExpandAlias("test")
		}()
	}

	wg.Wait()
	// If we reach here without race detector errors, the test passes
	SetAliasLookup(nil) // Clean up
}

func TestExpandAlias_ReturnsOriginalOnNoMatch(t *testing.T) {
	SetAliasLookup(func(name string) (string, bool) {
		if name == "known" {
			return "expanded", true
		}
		return name, false
	})
	defer SetAliasLookup(nil)

	// Test that unknown aliases return the original name
	result, found := ExpandAlias("unknown_constraint")
	assert.False(t, found)
	assert.Equal(t, "unknown_constraint", result)
}
