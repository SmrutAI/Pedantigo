package serialize

import (
	"reflect"
	"strings"

	"github.com/SmrutAI/pedantigo/internal/tags"
)

// FieldMetadata holds pre-parsed serialization metadata for a field.
type FieldMetadata struct {
	FieldIndex      int
	JSONName        string
	ExcludeContexts map[string]bool // Set for O(1) lookup (blacklist)
	IncludeContexts map[string]bool // Set for O(1) lookup (whitelist)
	OmitZero        bool
	OmitEmpty       bool // From json:",omitempty"
}

// BuildFieldMetadata creates serialization metadata for each struct field.
// It uses the provided tagName to parse validation constraints (e.g., "pedantigo", "validate", "binding").
func BuildFieldMetadata(typ reflect.Type, tagName string) map[string]FieldMetadata {
	metadata := make(map[string]FieldMetadata)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return metadata
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		jsonName, omitEmpty := parseJSONTag(jsonTag, field.Name)

		constraintsMap := tags.ParseTagWithName(field.Tag, tagName)
		excludeContexts := make(map[string]bool)
		includeContexts := make(map[string]bool)
		var omitZero bool

		if constraintsMap != nil {
			// Parse exclude:context1|context2 (pipe-separated)
			if excludeVal, ok := constraintsMap["exclude"]; ok {
				for _, ctx := range strings.Split(excludeVal, "|") {
					ctx = strings.TrimSpace(ctx)
					if ctx != "" {
						excludeContexts[ctx] = true
					}
				}
			}
			// Parse include:context1|context2 (pipe-separated)
			if includeVal, ok := constraintsMap["include"]; ok {
				for _, ctx := range strings.Split(includeVal, "|") {
					ctx = strings.TrimSpace(ctx)
					if ctx != "" {
						includeContexts[ctx] = true
					}
				}
			}
			_, omitZero = constraintsMap["omitzero"]
		}

		metadata[jsonName] = FieldMetadata{
			FieldIndex:      i,
			JSONName:        jsonName,
			ExcludeContexts: excludeContexts,
			IncludeContexts: includeContexts,
			OmitZero:        omitZero,
			OmitEmpty:       omitEmpty,
		}
	}

	return metadata
}

func parseJSONTag(tag, defaultName string) (name string, omitEmpty bool) {
	if tag == "" {
		return defaultName, false
	}

	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = defaultName
	}

	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitEmpty = true
		}
	}

	return name, omitEmpty
}
