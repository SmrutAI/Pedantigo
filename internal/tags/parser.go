package tags

import (
	"reflect"
	"strings"
)

// DefaultTagName is the default struct tag name used by Pedantigo.
const DefaultTagName = "pedantigo"

// ExtraFieldsTag is the tag value for fields that store extra/unknown JSON fields.
const ExtraFieldsTag = "extra_fields"

// ParseTag parses a struct tag using the default "pedantigo" tag name.
// Example: pedantigo:"required,email,min=18" -> map{"required": "", "email": "", "min": "18"}
// Special handling for oneof which has space-separated values: oneof=admin user guest.
func ParseTag(tag reflect.StructTag) map[string]string {
	return ParseTagWithName(tag, DefaultTagName)
}

// ParseTagWithName parses a struct tag using a custom tag name.
// This allows compatibility with other validation libraries like go-playground/validator.
// Example with tagName="validate": validate:"required,email" -> map{"required": "", "email": ""}.
func ParseTagWithName(tag reflect.StructTag, tagName string) map[string]string {
	validateTag := tag.Get(tagName)
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
		} else if idx := strings.IndexByte(part, ':'); idx != -1 {
			// Handle key:value syntax (e.g., exclude:response,log)
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

// ParseTagWithDive parses a struct tag using the default "pedantigo" tag name
// and returns a structured ParsedTag that separates collection-level, key-level,
// and element-level constraints.
//
// Syntax:
//   - pedantigo:"min=3"                    -> CollectionConstraints only
//   - pedantigo:"dive,email"               -> ElementConstraints only (dive present)
//   - pedantigo:"min=3,dive,min=5"         -> Both collection and element
//   - pedantigo:"dive,keys,min=2,endkeys,email" -> Map: key + value constraints
func ParseTagWithDive(tag reflect.StructTag) *ParsedTag {
	return ParseTagWithDiveAndName(tag, DefaultTagName)
}

// ParseTagWithDiveAndName parses a struct tag using a custom tag name
// and returns a structured ParsedTag that separates collection-level, key-level,
// and element-level constraints.
//
// This allows compatibility with other validation libraries like go-playground/validator.
// Example with tagName="validate": validate:"min=3,dive,email".
func ParseTagWithDiveAndName(tag reflect.StructTag, tagName string) *ParsedTag {
	validateTag := tag.Get(tagName)
	if validateTag == "" {
		return nil
	}

	parsed := &ParsedTag{
		CollectionConstraints: make(map[string]string),
		KeyConstraints:        make(map[string]string),
		ElementConstraints:    make(map[string]string),
	}

	parts := strings.Split(validateTag, ",")

	// State machine states
	const (
		stateCollection = iota
		stateDive
		stateKeysSection
		stateElementAfterKeys
		stateElement
	)

	state := stateCollection
	var keysFound bool
	var endkeysFound bool

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Handle special keywords
		if part == "dive" {
			if state == stateCollection {
				parsed.DivePresent = true
				state = stateDive
			}
			continue
		}

		if part == "keys" {
			if state != stateDive {
				panic("'keys' can only appear after 'dive'")
			}
			keysFound = true
			state = stateKeysSection
			continue
		}

		if part == "endkeys" {
			if !keysFound {
				panic("'endkeys' without preceding 'keys'")
			}
			endkeysFound = true
			state = stateElementAfterKeys
			continue
		}

		// Parse constraint (key=value or bare keyword)
		var constraintName, constraintValue string
		if idx := strings.IndexByte(part, '='); idx != -1 {
			constraintName = strings.TrimSpace(part[:idx])
			constraintValue = strings.TrimSpace(part[idx+1:])
		} else {
			constraintName = part
			constraintValue = ""
		}

		// Add to appropriate map based on current state
		switch state {
		case stateCollection:
			parsed.CollectionConstraints[constraintName] = constraintValue
		case stateDive:
			parsed.ElementConstraints[constraintName] = constraintValue
		case stateKeysSection:
			parsed.KeyConstraints[constraintName] = constraintValue
		case stateElementAfterKeys, stateElement:
			parsed.ElementConstraints[constraintName] = constraintValue
			state = stateElement
		}
	}

	// Validation: if keys was found, endkeys must also be found
	if keysFound && !endkeysFound {
		panic("'keys' without closing 'endkeys'")
	}

	return parsed
}
