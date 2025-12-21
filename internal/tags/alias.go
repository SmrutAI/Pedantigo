package tags

import "sync"

// aliasLookup is set by the registry package to allow tag parsing
// to expand aliases. This avoids import cycles.
// The function is protected by aliasLookupMu for thread safety.
var (
	aliasLookup   func(name string) (string, bool)
	aliasLookupMu sync.RWMutex
)

// SetAliasLookup sets the function used to look up tag aliases.
// This should be called once by the registry package during initialization.
// Thread-safe: can be called concurrently with ExpandAlias.
func SetAliasLookup(fn func(name string) (string, bool)) {
	aliasLookupMu.Lock()
	defer aliasLookupMu.Unlock()
	aliasLookup = fn
}

// ExpandAlias expands an alias to its full tag definition.
// Returns the expansion and true if the alias exists,
// returns the original name and false otherwise.
// Thread-safe: can be called concurrently with SetAliasLookup.
func ExpandAlias(name string) (string, bool) {
	aliasLookupMu.RLock()
	defer aliasLookupMu.RUnlock()

	if aliasLookup == nil {
		return name, false
	}
	return aliasLookup(name)
}
