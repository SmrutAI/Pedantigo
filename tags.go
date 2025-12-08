package pedantigo

import (
	"reflect"
	"strings"
)

// parseTag parses a struct tag and returns constraints
// Example: pedantigo:"required,email,min=18" -> map{"required": "", "email": "", "min": "18"}
// Special handling for oneof which has space-separated values: oneof=admin user guest
func parseTag(tag reflect.StructTag) map[string]string {
	validateTag := tag.Get("pedantigo")
	if validateTag == "" {
		return nil
	}

	constraints := make(map[string]string)
	parts := strings.Split(validateTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a key=value constraint
		if idx := strings.IndexByte(part, '='); idx != -1 {
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])
			constraints[key] = value
		} else {
			// Simple constraint like "required" or "email"
			constraints[part] = ""
		}
	}

	return constraints
}
