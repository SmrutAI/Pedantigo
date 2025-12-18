package pedantigo

import (
	"fmt"
	"strconv"
	"sync"
)

// validateContext holds reusable buffers for a single Validate() call.
// Type-agnostic (no generics) so it can be pooled across all Validator[T] instances.
type validateContext struct {
	pathBuf []byte       // Reusable buffer for building field paths
	errs    []FieldError // Reusable error slice
}

// validateContextPool is the global pool for validation contexts.
// Shared across all Validator[T] instances since validateContext has no generic parameter.
var validateContextPool = sync.Pool{
	New: func() any {
		return &validateContext{
			pathBuf: make([]byte, 0, 128),
			errs:    make([]FieldError, 0, 8),
		}
	},
}

// appendPath appends a field name to the path buffer with "." separator.
// Returns the new path as a byte slice (shares backing array with buf).
func appendPath(buf, parent []byte, name string) []byte {
	if len(parent) > 0 {
		buf = append(buf, parent...)
		buf = append(buf, '.')
	}
	buf = append(buf, name...)
	return buf
}

// appendIndex appends an array index to the path buffer: "path[index]".
// Uses strconv.AppendInt to avoid fmt.Sprintf allocations.
func appendIndex(buf, path []byte, index int) []byte {
	buf = append(buf, path...)
	buf = append(buf, '[')
	buf = strconv.AppendInt(buf, int64(index), 10)
	buf = append(buf, ']')
	return buf
}

// appendMapKey appends a map key to the path buffer: "path[key]".
// Handles common key types without allocation; falls back to fmt.Sprint for complex types.
func appendMapKey(buf, path []byte, key any) []byte {
	buf = append(buf, path...)
	buf = append(buf, '[')
	switch k := key.(type) {
	case string:
		buf = append(buf, k...)
	case int:
		buf = strconv.AppendInt(buf, int64(k), 10)
	case int64:
		buf = strconv.AppendInt(buf, k, 10)
	case int32:
		buf = strconv.AppendInt(buf, int64(k), 10)
	case uint:
		buf = strconv.AppendUint(buf, uint64(k), 10)
	case uint64:
		buf = strconv.AppendUint(buf, k, 10)
	case uint32:
		buf = strconv.AppendUint(buf, uint64(k), 10)
	default:
		// Fallback for complex key types (e.g., custom types)
		buf = append(buf, fmt.Sprint(k)...)
	}
	buf = append(buf, ']')
	return buf
}
