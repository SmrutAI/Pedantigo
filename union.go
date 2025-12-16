package pedantigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/SmrutAI/pedantigo/internal/constraints"
	"github.com/SmrutAI/pedantigo/schemagen"
)

// UnionVariant represents a variant type in a discriminated union.
// It maps a discriminator value to a specific Go struct type.
type UnionVariant struct {
	// DiscriminatorValue is the value of the discriminator field that selects this variant.
	// For example, if discriminator is "type" and value is "cat", this variant handles {"type": "cat", ...}
	DiscriminatorValue string

	// Type is the Go struct type for this variant.
	Type reflect.Type
}

// UnionOptions configures discriminated union behavior.
type UnionOptions struct {
	// DiscriminatorField is the JSON field name used to determine the variant type.
	// For example: "type", "kind", "pet_type"
	DiscriminatorField string

	// Variants maps discriminator values to their corresponding Go types.
	Variants []UnionVariant
}

// UnionValidator validates discriminated unions where a field determines the variant type.
// Stub: not yet implemented.
type UnionValidator[T any] struct {
	options  UnionOptions
	variants map[string]reflect.Type // discriminator value -> variant type
}

// NewUnion creates a UnionValidator for type T with discriminated union support.
// Stub: returns error indicating not implemented.
func NewUnion[T any](opts UnionOptions) (*UnionValidator[T], error) {
	if opts.DiscriminatorField == "" {
		return nil, errors.New("discriminator field is required")
	}
	if len(opts.Variants) == 0 {
		return nil, errors.New("at least one variant is required")
	}

	variants := make(map[string]reflect.Type)
	for _, v := range opts.Variants {
		if v.DiscriminatorValue == "" {
			return nil, errors.New("variant discriminator value cannot be empty")
		}
		if v.Type == nil {
			return nil, errors.New("variant type cannot be nil")
		}
		if _, exists := variants[v.DiscriminatorValue]; exists {
			return nil, errors.New("duplicate discriminator value: " + v.DiscriminatorValue)
		}
		variants[v.DiscriminatorValue] = v.Type
	}

	return &UnionValidator[T]{
		options:  opts,
		variants: variants,
	}, nil
}

// Unmarshal unmarshals JSON data into the appropriate union variant.
// Stub: returns error indicating not implemented.
func (v *UnionValidator[T]) Unmarshal(data []byte) (any, error) {
	// Step 1: Unmarshal to map[string]any to extract discriminator
	var jsonMap map[string]any
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Step 2: Check if discriminator field exists in the map
	discriminatorValue, exists := jsonMap[v.options.DiscriminatorField]
	if !exists || discriminatorValue == nil {
		return nil, fmt.Errorf(ErrMsgMissingDiscriminator, v.options.DiscriminatorField)
	}

	// Step 3: Convert discriminator value to string (handle both string and numeric JSON values)
	var discriminatorStr string
	switch val := discriminatorValue.(type) {
	case string:
		discriminatorStr = val
	case float64:
		// JSON numbers come through as float64
		discriminatorStr = fmt.Sprintf("%v", val)
	case int:
		discriminatorStr = fmt.Sprintf("%v", val)
	default:
		discriminatorStr = fmt.Sprintf("%v", val)
	}

	// Step 4: Look up variant type
	variantType, found := v.variants[discriminatorStr]
	if !found {
		return nil, fmt.Errorf(ErrMsgUnknownDiscriminator, discriminatorStr, v.options.DiscriminatorField)
	}

	// Step 5: Create a new instance of the variant type (pointer)
	variantPtr := reflect.New(variantType).Interface()

	// Step 6: Unmarshal the JSON data into the variant instance
	// Get the reflect.Type of the variant to create a generic validator
	if err := json.Unmarshal(data, variantPtr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into variant: %w", err)
	}

	// Step 7: Validate the variant using reflection-based validation
	variantPtrValue := reflect.ValueOf(variantPtr)
	variantValue := variantPtrValue.Elem()

	if err := v.validateVariant(variantValue, variantType); err != nil {
		return nil, err
	}

	// Step 8: Return dereferenced result (not pointer)
	return variantValue.Interface(), nil
}

// Validate validates a union value.
// Stub: returns error indicating not implemented.
func (v *UnionValidator[T]) Validate(obj any) error {
	// Step 1: Check if obj is nil
	if obj == nil {
		return errors.New("nil value is not a valid union variant")
	}

	// Step 2: Get reflect.Type of obj
	objType := reflect.TypeOf(obj)

	// Step 3: Check if the type is one of the union variants
	variantType := objType
	found := false
	for _, vType := range v.variants {
		if vType == variantType {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("type %T is not a valid union variant", obj)
	}

	// Step 4: Create a pointer to the object for validation
	// We need to pass a pointer to the validator
	objPtr := reflect.New(objType)
	objPtr.Elem().Set(reflect.ValueOf(obj))

	// Step 5: Validate using reflection-based validation
	if err := v.validateVariant(objPtr.Elem(), objType); err != nil {
		return err
	}

	return nil
}

// validateVariant validates a variant value using reflection-based validation.
// It checks all struct field constraints from tags without requiring explicit Validator creation.
func (v *UnionValidator[T]) validateVariant(variantValue reflect.Value, variantType reflect.Type) error {
	// Handle pointer types
	if variantType.Kind() == reflect.Ptr {
		variantType = variantType.Elem()
		if variantValue.Kind() == reflect.Ptr {
			variantValue = variantValue.Elem()
		}
	}

	// Only validate structs
	if variantType.Kind() != reflect.Struct {
		return nil
	}

	var fieldErrors []FieldError

	// Iterate through all fields and validate them
	for i := 0; i < variantType.NumField(); i++ {
		field := variantType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldValue := variantValue.Field(i)
		fieldPath := field.Name

		// Parse validation tags
		constraintsMap := make(map[string]string)
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			// Simple tag parsing: split by comma
			parts := splitTags(validateTag)
			for _, part := range parts {
				kv := splitKeyValue(part)
				if len(kv) == 1 {
					constraintsMap[kv[0]] = ""
				} else {
					constraintsMap[kv[0]] = kv[1]
				}
			}
		}

		// Skip fields without validation constraints
		if len(constraintsMap) == 0 {
			continue
		}

		// Check required constraint
		if _, hasRequired := constraintsMap["required"]; hasRequired {
			if fieldValue.IsZero() {
				fieldErrors = append(fieldErrors, FieldError{
					Field:   fieldPath,
					Message: "is required",
					Value:   fieldValue.Interface(),
				})
				continue
			}
		}

		// Build and apply other constraints
		constraintList := buildVariantConstraints(constraintsMap, field.Type)
		for _, constraint := range constraintList {
			if err := constraint.Validate(fieldValue.Interface()); err != nil {
				fieldErrors = append(fieldErrors, FieldError{
					Field:   fieldPath,
					Message: err.Error(),
					Value:   fieldValue.Interface(),
				})
			}
		}
	}

	if len(fieldErrors) > 0 {
		return &ValidationError{Errors: fieldErrors}
	}

	return nil
}

// buildVariantConstraints builds constraint validators for a field type.
// This is a simplified version that delegates to the constraints package.
func buildVariantConstraints(constraintsMap map[string]string, fieldType reflect.Type) []constraints.Constraint {
	// Import and use the internal constraints builder
	return constraints.BuildConstraints(constraintsMap, fieldType)
}

// splitTags splits a tag string by comma, handling quoted values.
func splitTags(tagStr string) []string {
	var tags []string
	var current strings.Builder
	var inQuotes bool

	for _, r := range tagStr {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case r == ',' && !inQuotes:
			if current.Len() > 0 {
				tags = append(tags, strings.TrimSpace(current.String()))
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tags = append(tags, strings.TrimSpace(current.String()))
	}

	return tags
}

// splitKeyValue splits a key=value pair.
func splitKeyValue(pair string) []string {
	parts := strings.SplitN(pair, "=", 2)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// Schema generates JSON Schema for the discriminated union using oneOf.
// Returns a schema with oneOf array containing all variant schemas,
// each with a const constraint on the discriminator field.
// Implementation.
func (v *UnionValidator[T]) Schema() *jsonschema.Schema {
	// Create a parseTagFunc that parses "validate" struct tags from variant structs
	// This function will be used by GenerateVariantSchema to apply validation constraints
	parseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
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

	// Generate union schema using the schemagen package
	return schemagen.GenerateUnionSchema(v.options.DiscriminatorField, v.variants, parseTagFunc)
}

// VariantFor is a helper to create UnionVariant from a type parameter.
// Usage: VariantFor[Cat]("cat").
func VariantFor[T any](discriminatorValue string) UnionVariant {
	var zero T
	return UnionVariant{
		DiscriminatorValue: discriminatorValue,
		Type:               reflect.TypeOf(zero),
	}
}
