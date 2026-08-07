package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/invopop/jsonschema"

	"github.com/SmrutAI/pedantigo/v2/validator/internal/constraints"
	"github.com/SmrutAI/pedantigo/v2/validator/internal/deserialize"
	"github.com/SmrutAI/pedantigo/v2/validator/internal/serialize"
	"github.com/SmrutAI/pedantigo/v2/validator/internal/tags"
)

// Validator validates structs of type T.
type Validator[T any] struct {
	typ                reflect.Type
	options            Options
	tagName            string // Resolved tag name (instance override or global)
	fieldDeserializers map[string]deserialize.FieldDeserializer

	// Cached field constraints (built at creation time)
	fieldCache *constraints.FieldCache

	// Schema caching (lazy initialization with double-checked locking)
	schemaMu          sync.RWMutex
	cachedSchema      *jsonschema.Schema // Schema() result
	cachedSchemaJSON  []byte             // SchemaJSON() result
	cachedOpenAPI     *jsonschema.Schema // SchemaOpenAPI() result
	cachedOpenAPIJSON []byte             // SchemaJSONOpenAPI() result
	cachedSchemaLLM   *jsonschema.Schema // SchemaLLM() result (no $schema field)
	cachedLLMJSON     []byte             // SchemaJSONLLM() result

	// Extra fields support (nil when ExtraAllow disabled)
	extraFieldInfo *deserialize.ExtraFieldInfo

	// Re-entrancy guard for Validatable.Validate() calls (#15)
	// Tracks object pointers currently being validated to prevent infinite recursion
	// when a user's Validate() method calls back into pedantigo.
	validating sync.Map
}

// New creates a new Validator for type T with optional configuration.
func New[T any](opts ...Options) *Validator[T] {
	// Mark that a validator has been created (prevents late SetTagName calls)
	markValidatorCreated()

	var zero T
	typ := reflect.TypeOf(zero)

	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Resolve tag name (instance override or global)
	tagName := resolveTagName(options)

	vl := &Validator[T]{
		typ:                typ,
		options:            options,
		tagName:            tagName,
		fieldDeserializers: make(map[string]deserialize.FieldDeserializer),
	}

	// Build field deserializers at creation time (fail-fast)
	vl.fieldDeserializers = deserialize.BuildFieldDeserializers(
		typ,
		deserialize.BuilderOptions{
			StrictMissingFields: options.StrictMissingFields,
			TagName:             tagName,
		},
		vl.setFieldValue,
		vl.setDefaultValue,
	)

	// Validate dive/keys/endkeys tag usage at creation time (fail-fast)
	vl.validateDiveTags(typ, tagName)

	// Build field constraints at creation time (the key optimization)
	// Pass visited set to detect circular type references (#15)
	vl.fieldCache = vl.buildFieldConstraints(typ, tagName, make(map[reflect.Type]bool))

	// Detect extra_fields for ExtraAllow mode
	if options.ExtraFields == ExtraAllow {
		vl.extraFieldInfo = deserialize.DetectExtraField(typ, tagName)
		if vl.extraFieldInfo == nil {
			panic(ErrMsgExtraFieldRequired)
		}
	}

	return vl
}

// buildFieldConstraints builds and caches all field constraints at creation time.
// visited tracks types already being processed to detect circular references (#15).
func (v *Validator[T]) buildFieldConstraints(typ reflect.Type, tagName string, visited map[reflect.Type]bool) *constraints.FieldCache {
	// Handle pointer types
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	// Detect circular type references: if we're already building constraints
	// for this type, return nil to break the cycle. At validation time, nil
	// NestedCache means "don't recurse" which is correct — the parent type's
	// validator already covers its own fields.
	if visited[typ] {
		return nil
	}
	visited[typ] = true

	cache := constraints.NewFieldCache()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse tags once using the configured tag name
		parsedTag := tags.ParseTagWithDiveAndName(field.Tag, tagName)

		// Field type info
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		isCollection := fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map
		isMap := fieldType.Kind() == reflect.Map

		// Resolve field name using custom function or default (json tag or field name)
		jsonName := resolveFieldName(&field)

		cached := constraints.CachedField{
			Name:         field.Name,
			JSONName:     jsonName,
			FieldIndex:   i,
			IsCollection: isCollection,
			IsMap:        isMap,
		}

		if parsedTag != nil {
			cached.HasDive = parsedTag.DivePresent

			// Check for required tag
			if _, hasRequired := parsedTag.CollectionConstraints["required"]; hasRequired {
				cached.IsRequired = true
			}

			// Check for omitempty tag
			if _, hasOmitEmpty := parsedTag.CollectionConstraints["omitempty"]; hasOmitEmpty {
				cached.IsOmitEmpty = true
			}

			// Constraints before dive (or regular field constraints)
			if len(parsedTag.CollectionConstraints) > 0 {
				cached.Constraints = constraints.BuildConstraints(parsedTag.CollectionConstraints, field.Type)
				// Extract context-aware validators (called during ValidateCtx)
				cached.ContextConstraints = constraints.ExtractContextValidators(parsedTag.CollectionConstraints)
			}

			// Element constraints after dive
			if parsedTag.DivePresent && len(parsedTag.ElementConstraints) > 0 {
				cached.ElementConstraints = constraints.BuildConstraints(parsedTag.ElementConstraints, field.Type.Elem())
			}

			// Map key constraints
			if isMap && len(parsedTag.KeyConstraints) > 0 {
				cached.KeyConstraints = constraints.BuildConstraints(parsedTag.KeyConstraints, field.Type.Key())
			}

			// Cross-field constraints (eqfield, gtfield, etc.) and skip constraints
			cached.CrossFieldConstraints, cached.SkipConstraints = constraints.BuildCrossFieldConstraintsForField(
				parsedTag.CollectionConstraints, typ, i)
			cached.HasSkipConstraints = len(cached.SkipConstraints) > 0
		}

		// Recurse for nested structs (passing visited set for cycle detection)
		switch fieldType.Kind() {
		case reflect.Struct:
			cached.NestedCache = v.buildFieldConstraints(fieldType, tagName, visited)
		case reflect.Slice, reflect.Map:
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Pointer {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				cached.NestedCache = v.buildFieldConstraints(elemType, tagName, visited)
			}
		}

		cache.Fields = append(cache.Fields, cached)
	}

	return cache
}

// validateDiveTags validates that dive/keys/endkeys tags are used correctly.
// This is called at creation time to fail fast on invalid tag combinations.
func (v *Validator[T]) validateDiveTags(typ reflect.Type, tagName string) {
	// Handle pointer types
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse the tag with dive support using the configured tag name
		parsedTag := tags.ParseTagWithDiveAndName(field.Tag, tagName)
		if parsedTag == nil {
			continue
		}

		// Get the underlying field type (dereference pointers)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		isCollection := fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map
		isMap := fieldType.Kind() == reflect.Map

		// Panic: dive on non-collection field
		if parsedTag.DivePresent && !isCollection {
			panic(fmt.Sprintf("field %s.%s: 'dive' can only be used on slice or map types, got %s",
				typ.Name(), field.Name, fieldType.Kind()))
		}

		// Panic: keys on non-map field
		if len(parsedTag.KeyConstraints) > 0 && !isMap {
			panic(fmt.Sprintf("field %s.%s: 'keys' can only be used on map types, got %s",
				typ.Name(), field.Name, fieldType.Kind()))
		}

		// Panic: unique on non-collection field
		if _, hasUnique := parsedTag.CollectionConstraints["unique"]; hasUnique && !isCollection {
			panic(fmt.Sprintf("field %s.%s: 'unique' can only be used on slice or map types, got %s",
				typ.Name(), field.Name, fieldType.Kind()))
		}

		// Recursively validate nested structs
		switch fieldType.Kind() {
		case reflect.Struct:
			v.validateDiveTags(fieldType, tagName)
		case reflect.Slice:
			if fieldType.Elem().Kind() == reflect.Struct {
				v.validateDiveTags(fieldType.Elem(), tagName)
			}
		case reflect.Map:
			if fieldType.Elem().Kind() == reflect.Struct {
				v.validateDiveTags(fieldType.Elem(), tagName)
			}
		}
	}
}

// setFieldValue wraps the deserialize package SetFieldValue for use in validator.
// Passes options to enable required checking in nested structs (via dive).
func (v *Validator[T]) setFieldValue(fieldValue reflect.Value, inValue any, fieldType reflect.Type, goFieldName string) error {
	opts := deserialize.FieldOptions{
		StrictMissingFields: v.options.StrictMissingFields,
		TagName:             v.tagName,
		FieldName:           goFieldName,
	}
	// Create a recursive function that passes the options
	var recursiveSet func(reflect.Value, any, reflect.Type) error
	recursiveSet = func(fv reflect.Value, iv any, ft reflect.Type) error {
		return deserialize.SetFieldValueWithOptions(fv, iv, ft, recursiveSet, opts)
	}
	return deserialize.SetFieldValueWithOptions(fieldValue, inValue, fieldType, recursiveSet, opts)
}

// Validate validates a struct and returns any validation errors
// NOTE: 'required' is NOT checked here - it's only checked during Unmarshal
// Validate checks if the value satisfies the constraint.
func (v *Validator[T]) Validate(obj *T) error {
	if obj == nil {
		return &ValidationError{
			Errors: []FieldError{{Field: fieldNameRoot, Message: ErrMsgNilPointer}},
		}
	}

	// Get context from pool
	ctx := validateContextPool.Get().(*validateContext)

	// Reset buffers (keep capacity)
	ctx.pathBuf = ctx.pathBuf[:0]
	ctx.errs = ctx.errs[:0]

	// Validate all fields using struct tags (required is skipped via buildConstraints)
	v.validateWithCache(reflect.ValueOf(obj).Elem(), nil, ctx, v.fieldCache)

	// Check if struct implements Validatable for cross-field validation.
	// Use re-entrancy guard to prevent infinite recursion when user's
	// Validate() method calls back into pedantigo (#15).
	if validatable, ok := any(obj).(Validatable); ok {
		key := reflect.ValueOf(obj).Pointer()
		if _, alreadyValidating := v.validating.LoadOrStore(key, true); !alreadyValidating {
			defer v.validating.Delete(key)
			if err := validatable.Validate(); err != nil {
				// Check if it's a ValidationError with multiple errors
				var ve *ValidationError
				if errors.As(err, &ve) {
					ctx.errs = append(ctx.errs, ve.Errors...)
				} else {
					// Single error or custom error type
					ctx.errs = append(ctx.errs, FieldError{
						Field:   fieldNameRoot,
						Message: err.Error(),
					})
				}
			}
		}
	}

	// Extract errors before returning to pool
	var result error
	if len(ctx.errs) > 0 {
		result = &ValidationError{Errors: ctx.errs}
		ctx.errs = nil // Clear reference so pool doesn't hold onto errors
	}

	// Return to pool
	validateContextPool.Put(ctx)

	return result
}

// validateWithCache validates using pre-built cached constraints.
// Uses byte slice paths and appends errors to ctx.errs to minimize allocations.
func (v *Validator[T]) validateWithCache(val reflect.Value, path []byte, ctx *validateContext, cache *constraints.FieldCache) {
	if cache == nil {
		return
	}

	// Handle pointer indirection
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	for i := range cache.Fields {
		cached := &cache.Fields[i]
		fieldVal := val.Field(cached.FieldIndex)

		// Check skip constraints FIRST (e.g., skip_unless)
		// This is O(1) when HasSkipConstraints is false (no overhead)
		if cached.HasSkipConstraints {
			shouldSkip := false
			for _, sc := range cached.SkipConstraints {
				if sc.ShouldSkip(val) {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				continue // Skip ALL validation on this field
			}
		}

		// Build field path using buffer
		fieldPath := appendPath(ctx.pathBuf[:0], path, cached.Name)

		// NOTE: Required checking for nested fields (via dive) is now done during
		// deserialization in deserializeStructFields(), which can correctly distinguish
		// between "key missing from JSON" vs "key present with zero value".
		// The old IsZero() check here was incorrect for zero values like 0, 0.0, false, "".

		// omitempty: when set and field is zero, skip regular constraints and
		// collection/nested recursion, but ALWAYS run cross-field constraints
		// (required_with, required_if, eqfield, etc.) because they need to
		// evaluate the field even when it is at its zero value.
		isOmitEmptyZero := cached.IsOmitEmpty && fieldVal.IsZero()

		// Apply field constraints (skip for omitempty zero-value fields)
		if !isOmitEmptyZero {
			for _, c := range cached.Constraints {
				if err := c.Validate(fieldVal.Interface()); err != nil {
					ctx.errs = append(ctx.errs, v.newFieldError(string(fieldPath), err, fieldVal.Interface()))
				}
			}
		}

		// Apply cross-field constraints (always run, even for omitempty zero-value fields)
		for _, c := range cached.CrossFieldConstraints {
			if err := c.ValidateCrossField(fieldVal.Interface(), val, string(fieldPath)); err != nil {
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					ctx.errs = append(ctx.errs, valErr.Errors...)
				} else {
					ctx.errs = append(ctx.errs, FieldError{
						Field:   string(fieldPath),
						Message: err.Error(),
					})
				}
			}
		}

		// Handle collections with dive and nested struct recursion
		// (skip for omitempty zero-value fields)
		if !isOmitEmptyZero {
			if cached.IsCollection && cached.HasDive {
				if cached.IsMap {
					v.validateMapWithCache(fieldVal, fieldPath, ctx, cached)
				} else {
					v.validateSliceWithCache(fieldVal, fieldPath, ctx, cached)
				}
			} else if cached.NestedCache != nil && !cached.IsCollection {
				// Recurse for nested structs (but NOT collection elements without dive)
				v.validateWithCache(fieldVal, fieldPath, ctx, cached.NestedCache)
			}
		}
	}
}

// validateSliceWithCache validates slice elements using cached constraints.
// Uses appendIndex for zero-allocation index formatting.
func (v *Validator[T]) validateSliceWithCache(val reflect.Value, path []byte, ctx *validateContext, cached *constraints.CachedField) {
	for i := 0; i < val.Len(); i++ {
		elemVal := val.Index(i)
		// Build element path: "path[i]" using strconv.AppendInt (no allocation)
		elemPath := appendIndex(ctx.pathBuf[:0], path, i)

		// Apply element constraints
		for _, c := range cached.ElementConstraints {
			if err := c.Validate(elemVal.Interface()); err != nil {
				ctx.errs = append(ctx.errs, v.newFieldError(string(elemPath), err, elemVal.Interface()))
			}
		}

		// Recurse for nested structs
		if cached.NestedCache != nil {
			v.validateWithCache(elemVal, elemPath, ctx, cached.NestedCache)
		}
	}
}

// validateMapWithCache validates map entries using cached constraints.
// Uses appendMapKey for optimized key formatting.
func (v *Validator[T]) validateMapWithCache(val reflect.Value, path []byte, ctx *validateContext, cached *constraints.CachedField) {
	iter := val.MapRange()
	for iter.Next() {
		mapKey := iter.Key()
		mapVal := iter.Value()
		// Build element path: "path[key]" using type-optimized appending
		elemPath := appendMapKey(ctx.pathBuf[:0], path, mapKey.Interface())

		// Apply key constraints
		for _, c := range cached.KeyConstraints {
			if err := c.Validate(mapKey.Interface()); err != nil {
				ctx.errs = append(ctx.errs, v.newFieldError(string(elemPath), err, mapKey.Interface()))
			}
		}

		// Apply value constraints
		for _, c := range cached.ElementConstraints {
			if err := c.Validate(mapVal.Interface()); err != nil {
				ctx.errs = append(ctx.errs, v.newFieldError(string(elemPath), err, mapVal.Interface()))
			}
		}

		// Recurse for nested structs
		if cached.NestedCache != nil {
			v.validateWithCache(mapVal, elemPath, ctx, cached.NestedCache)
		}
	}
}

// newFieldError creates a FieldError, extracting Code from ConstraintError if available.
func (v *Validator[T]) newFieldError(field string, err error, value any) FieldError {
	fe := FieldError{
		Field:   field,
		Message: err.Error(),
		Value:   value,
	}

	var ce *constraints.ConstraintError
	if errors.As(err, &ce) {
		fe.Code = ce.Code
	}

	return fe
}

// Unmarshal unmarshals JSON data, applies defaults, and validates.
func (v *Validator[T]) Unmarshal(data []byte) (*T, error) {
	// Fast path: skip 2-step flow if StrictMissingFields is disabled
	// UNLESS ExtraAllow is set, in which case we need the 2-step flow
	if !v.options.StrictMissingFields && v.options.ExtraFields != ExtraAllow {
		var obj T

		// Use json.Decoder with DisallowUnknownFields for ExtraForbid
		if v.options.ExtraFields == ExtraForbid {
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&obj); err != nil {
				return &obj, &ValidationError{
					Errors: []FieldError{{
						Field:   fieldNameRoot,
						Message: "JSON decode error: " + ErrMsgUnknownField,
					}},
				}
			}
		} else {
			if err := json.Unmarshal(data, &obj); err != nil {
				return nil, &ValidationError{
					Errors: []FieldError{{
						Field:   fieldNameRoot,
						Message: fmt.Sprintf("JSON decode error: %v", err),
					}},
				}
			}
		}

		// Only run validators (skip required checks and defaults)
		if err := v.Validate(&obj); err != nil {
			return &obj, err
		}
		return &obj, nil
	}

	// Step 0.5: Pre-check for extra fields if ExtraForbid is set (handles nested structs)
	if v.options.ExtraFields == ExtraForbid {
		var obj T
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&obj); err != nil {
			return &obj, &ValidationError{
				Errors: []FieldError{{
					Field:   fieldNameRoot,
					Message: ErrMsgUnknownField,
				}},
			}
		}
	}

	// Step 1: Unmarshal to map[string]any to detect which fields exist
	var jsonMap map[string]any
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, &ValidationError{
			Errors: []FieldError{{
				Field:   fieldNameRoot,
				Message: fmt.Sprintf("JSON decode error: %v", err),
			}},
		}
	}

	// Step 2: Create new struct instance
	var obj T
	objValue := reflect.ValueOf(&obj).Elem()

	// Step 3: Apply field deserializers
	var fieldErrors []FieldError
	for fieldName, deserializer := range v.fieldDeserializers {
		var inValue any
		if val, exists := jsonMap[fieldName]; exists {
			inValue = val // Field present in JSON (might be nil for JSON null)
		} else {
			inValue = deserialize.FieldMissingSentinel // Field missing from JSON
		}

		if err := deserializer(&objValue, inValue); err != nil {
			// Check if it's a MultiRequiredFieldError with multiple paths
			var multiErr *deserialize.MultiRequiredFieldError
			if errors.As(err, &multiErr) {
				for _, reqErr := range multiErr.Errors {
					fieldErrors = append(fieldErrors, FieldError{
						Field:   reqErr.Field,
						Code:    constraints.CodeRequired,
						Message: reqErr.Error(),
					})
				}
			} else {
				// Check if it's a single RequiredFieldError
				var reqErr *deserialize.RequiredFieldError
				if errors.As(err, &reqErr) {
					fieldErrors = append(fieldErrors, FieldError{
						Field:   reqErr.Field,
						Code:    constraints.CodeRequired,
						Message: reqErr.Error(),
					})
				} else {
					fieldErrors = append(fieldErrors, FieldError{
						Field:   fieldName,
						Message: err.Error(),
					})
				}
			}
		}
	}

	// Step 3.5: Capture extra fields recursively (for this struct and all nested structs)
	if v.options.ExtraFields == ExtraAllow {
		v.captureExtrasRecursive(objValue, jsonMap, v.tagName)
	}

	// Step 4: Run validation constraints (min, max, email, etc.)
	// NOTE: 'required' is already skipped in Validate() via buildConstraints
	// Continue validation even if there were deserialization errors to collect all issues
	if err := v.Validate(&obj); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			fieldErrors = append(fieldErrors, ve.Errors...)
		}
	}

	// Return all collected errors (from both deserialization and validation)
	if len(fieldErrors) > 0 {
		return &obj, &ValidationError{Errors: fieldErrors}
	}

	return &obj, nil
}

// getJSONFieldName returns the JSON field name for a struct field, or empty string if ignored.
func getJSONFieldName(field *reflect.StructField, tagName string) string {
	// Skip unexported fields
	if !field.IsExported() {
		return ""
	}

	// Skip the extras field itself
	if field.Tag.Get(tagName) == tags.ExtraFieldsTag {
		return ""
	}

	// Skip fields with json:"-"
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return ""
	}

	// Return JSON name or field name
	if jsonTag != "" {
		if name, _, found := strings.Cut(jsonTag, ","); found {
			return name
		}
		return jsonTag
	}
	return field.Name
}

// captureExtrasRecursive recursively captures extra fields for structs with extras field.
func (v *Validator[T]) captureExtrasRecursive(structVal reflect.Value, jsonMap map[string]any, tagName string) {
	structType := structVal.Type()

	// Capture extras at this level if extra_fields field exists
	v.captureExtrasAtLevel(structType, structVal, jsonMap, tagName)

	// Recursively handle nested structs
	v.captureExtrasInNestedFields(structType, structVal, jsonMap, tagName)
}

// captureExtrasAtLevel captures extra fields for a single struct level.
func (v *Validator[T]) captureExtrasAtLevel(structType reflect.Type, structVal reflect.Value, jsonMap map[string]any, tagName string) {
	extraInfo := deserialize.DetectExtraField(structType, tagName)
	if extraInfo == nil {
		return
	}

	// Build set of known field names
	knownFields := make(map[string]bool)
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		if name := getJSONFieldName(&f, tagName); name != "" {
			knownFields[name] = true
		}
	}

	// Capture extra fields
	extras := make(map[string]any)
	for key, value := range jsonMap {
		if !knownFields[key] {
			extras[key] = value
		}
	}

	// Set extras on struct (always set, even if empty)
	structVal.Field(extraInfo.FieldIndex).Set(reflect.ValueOf(extras))
}

// captureExtrasInNestedFields recursively captures extras in nested struct fields.
func (v *Validator[T]) captureExtrasInNestedFields(structType reflect.Type, structVal reflect.Value, jsonMap map[string]any, tagName string) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := getJSONFieldName(&field, tagName)
		if fieldName == "" {
			continue
		}

		nestedJSON, exists := jsonMap[fieldName]
		if !exists {
			continue
		}

		fieldVal := structVal.Field(i)
		v.captureExtrasInField(fieldVal, field.Type, nestedJSON, tagName)
	}
}

// captureExtrasInField captures extras for a single field based on its type.
func (v *Validator[T]) captureExtrasInField(fieldVal reflect.Value, fieldType reflect.Type, nestedJSON any, tagName string) {
	switch fieldType.Kind() {
	case reflect.Struct:
		if nestedMap, ok := nestedJSON.(map[string]any); ok {
			v.captureExtrasRecursive(fieldVal, nestedMap, tagName)
		}
	case reflect.Pointer:
		if fieldType.Elem().Kind() == reflect.Struct && !fieldVal.IsNil() {
			if nestedMap, ok := nestedJSON.(map[string]any); ok {
				v.captureExtrasRecursive(fieldVal.Elem(), nestedMap, tagName)
			}
		}
	case reflect.Slice:
		v.captureExtrasInSlice(fieldVal, fieldType, nestedJSON, tagName)
	}
}

// captureExtrasInSlice captures extras in slice elements.
func (v *Validator[T]) captureExtrasInSlice(fieldVal reflect.Value, fieldType reflect.Type, nestedJSON any, tagName string) {
	elemType := fieldType.Elem()
	isStructSlice := elemType.Kind() == reflect.Struct
	isPtrStructSlice := elemType.Kind() == reflect.Pointer && elemType.Elem().Kind() == reflect.Struct

	if !isStructSlice && !isPtrStructSlice {
		return
	}

	sliceJSON, ok := nestedJSON.([]any)
	if !ok {
		return
	}

	for idx := 0; idx < fieldVal.Len() && idx < len(sliceJSON); idx++ {
		elemVal := fieldVal.Index(idx)
		if elemType.Kind() == reflect.Pointer {
			if elemVal.IsNil() {
				continue
			}
			elemVal = elemVal.Elem()
		}
		if nestedMap, ok := sliceJSON[idx].(map[string]any); ok {
			v.captureExtrasRecursive(elemVal, nestedMap, tagName)
		}
	}
}

// setDefaultValue wraps the deserialize package SetDefaultValue for use in validator.
func (v *Validator[T]) setDefaultValue(fieldValue reflect.Value, defaultValue string) {
	deserialize.SetDefaultValue(fieldValue, defaultValue, v.setDefaultValue)
}

// Marshal validates and marshals struct to JSON.
func (v *Validator[T]) Marshal(obj *T) ([]byte, error) {
	// Validate before marshaling
	if err := v.Validate(obj); err != nil {
		return nil, err
	}

	// If extras field exists, marshal with extras merged
	if v.extraFieldInfo != nil {
		return v.marshalWithExtras(obj)
	}

	// Standard marshal
	return json.Marshal(obj)
}

// marshalWithExtras marshals struct with extras merged into the output.
func (v *Validator[T]) marshalWithExtras(obj *T) ([]byte, error) {
	// Marshal to JSON first (normal struct fields)
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// Unmarshal to map for merging
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Get the struct value
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Pointer {
		objVal = objVal.Elem()
	}

	// Recursively merge extras from this struct and all nested structs
	v.mergeExtrasRecursive(objVal, result, v.tagName)

	// Marshal the merged result
	return json.Marshal(result)
}

// mergeExtrasRecursive recursively merges extras fields from nested structs into the result map.
func (v *Validator[T]) mergeExtrasRecursive(structVal reflect.Value, resultMap map[string]any, tagName string) {
	structType := structVal.Type()

	// Merge extras at this level
	v.mergeExtrasAtLevel(structType, structVal, resultMap)

	// Recursively handle nested structs
	v.mergeExtrasInNestedFields(structType, structVal, resultMap, tagName)
}

// mergeExtrasAtLevel merges extras field values into the result map at one level.
func (v *Validator[T]) mergeExtrasAtLevel(structType reflect.Type, structVal reflect.Value, resultMap map[string]any) {
	extraInfo := deserialize.DetectExtraField(structType, v.tagName)
	if extraInfo == nil {
		return
	}

	extrasField := structVal.Field(extraInfo.FieldIndex)
	if extrasField.IsNil() {
		return
	}

	extras := extrasField.Interface().(map[string]any)
	for k, val := range extras {
		// Don't override existing struct fields
		if _, exists := resultMap[k]; !exists {
			resultMap[k] = val
		}
	}
}

// mergeExtrasInNestedFields recursively merges extras in nested fields.
func (v *Validator[T]) mergeExtrasInNestedFields(structType reflect.Type, structVal reflect.Value, resultMap map[string]any, tagName string) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := getJSONFieldName(&field, tagName)
		if fieldName == "" {
			continue
		}

		fieldVal := structVal.Field(i)
		v.mergeExtrasInField(fieldVal, field.Type, resultMap, fieldName, tagName)
	}
}

// mergeExtrasInField merges extras for a single field based on its type.
func (v *Validator[T]) mergeExtrasInField(fieldVal reflect.Value, fieldType reflect.Type, resultMap map[string]any, fieldName, tagName string) {
	switch fieldType.Kind() {
	case reflect.Struct:
		if nestedMap, ok := resultMap[fieldName].(map[string]any); ok {
			v.mergeExtrasRecursive(fieldVal, nestedMap, tagName)
		}
	case reflect.Pointer:
		if fieldType.Elem().Kind() == reflect.Struct && !fieldVal.IsNil() {
			if nestedMap, ok := resultMap[fieldName].(map[string]any); ok {
				v.mergeExtrasRecursive(fieldVal.Elem(), nestedMap, tagName)
			}
		}
	case reflect.Slice:
		v.mergeExtrasInSlice(fieldVal, fieldType, resultMap, fieldName, tagName)
	}
}

// mergeExtrasInSlice merges extras in slice elements.
func (v *Validator[T]) mergeExtrasInSlice(fieldVal reflect.Value, fieldType reflect.Type, resultMap map[string]any, fieldName, tagName string) {
	elemType := fieldType.Elem()
	isStructSlice := elemType.Kind() == reflect.Struct
	isPtrStructSlice := elemType.Kind() == reflect.Pointer && elemType.Elem().Kind() == reflect.Struct

	if !isStructSlice && !isPtrStructSlice {
		return
	}

	sliceAny, ok := resultMap[fieldName].([]any)
	if !ok {
		return
	}

	for idx := 0; idx < fieldVal.Len() && idx < len(sliceAny); idx++ {
		elemVal := fieldVal.Index(idx)
		if elemType.Kind() == reflect.Pointer {
			if elemVal.IsNil() {
				continue
			}
			elemVal = elemVal.Elem()
		}
		if nestedMap, ok := sliceAny[idx].(map[string]any); ok {
			v.mergeExtrasRecursive(elemVal, nestedMap, tagName)
		}
	}
}

// MarshalWithOptions validates and marshals struct to JSON with options.
// Options allow context-based field exclusion and omitzero behavior.
func (v *Validator[T]) MarshalWithOptions(obj *T, opts MarshalOptions) ([]byte, error) {
	// Validate before marshaling
	if err := v.Validate(obj); err != nil {
		return nil, err
	}

	// Build field metadata for filtering
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return []byte("null"), nil
		}
		val = val.Elem()
	}

	metadata := serialize.BuildFieldMetadata(val.Type(), v.tagName)

	// Convert options
	serializeOpts := serialize.Options{
		Context:  opts.Context,
		OmitZero: opts.OmitZero,
		TagName:  v.tagName,
	}

	// Convert to filtered map
	filtered := serialize.ToFilteredMap(val, metadata, serializeOpts)

	// Marshal the filtered map
	return json.Marshal(filtered)
}

// Dict converts the object into a dict.
func (v *Validator[T]) Dict(obj *T) (map[string]interface{}, error) {
	// If extras field exists, merge extras into the dict
	if v.extraFieldInfo != nil {
		// Marshal to JSON (which includes extras via marshalWithExtras)
		data, err := v.Marshal(obj)
		if err != nil {
			return nil, err
		}
		var dict map[string]interface{}
		if err := json.Unmarshal(data, &dict); err != nil {
			return nil, err
		}
		return dict, nil
	}

	// Standard dict conversion
	data, _ := json.Marshal(obj)
	var dict map[string]interface{}
	if err := json.Unmarshal(data, &dict); err != nil {
		return nil, err
	}
	return dict, nil
}

// NewModel creates a validated instance of T from various input types.
// Accepts: []byte (JSON), T (struct), *T (pointer), or map[string]any (kwargs).
// This is the unified constructor that validates regardless of input source.
func (v *Validator[T]) NewModel(input any) (*T, error) {
	switch val := input.(type) {
	case []byte:
		return v.Unmarshal(val)
	case *T:
		if val == nil {
			return nil, &ValidationError{
				Errors: []FieldError{{Field: fieldNameRoot, Message: ErrMsgNilPointer}},
			}
		}
		if err := v.Validate(val); err != nil {
			return val, err
		}
		return val, nil
	case map[string]any:
		return v.unmarshalFromMap(val)
	case T:
		if err := v.Validate(&val); err != nil {
			return &val, err
		}
		return &val, nil
	default:
		var zero T
		return nil, &ValidationError{
			Errors: []FieldError{{
				Field:   fieldNameRoot,
				Message: fmt.Sprintf("unsupported input type: %T, expected []byte, %T, *%T, or map[string]any", input, zero, zero),
			}},
		}
	}
}

// unmarshalFromMap creates a validated struct from a map (kwargs pattern).
// Reuses the same deserialization logic as Unmarshal.
func (v *Validator[T]) unmarshalFromMap(jsonMap map[string]any) (*T, error) {
	// Create new struct instance
	var obj T
	objValue := reflect.ValueOf(&obj).Elem()

	// Apply field deserializers (same logic as Unmarshal)
	var fieldErrors []FieldError
	for fieldName, deserializer := range v.fieldDeserializers {
		var inValue any
		if val, exists := jsonMap[fieldName]; exists {
			inValue = val
		} else {
			inValue = deserialize.FieldMissingSentinel
		}

		if err := deserializer(&objValue, inValue); err != nil {
			// Check if it's a MultiRequiredFieldError with multiple paths
			var multiErr *deserialize.MultiRequiredFieldError
			if errors.As(err, &multiErr) {
				for _, reqErr := range multiErr.Errors {
					fieldErrors = append(fieldErrors, FieldError{
						Field:   reqErr.Field,
						Code:    constraints.CodeRequired,
						Message: reqErr.Error(),
					})
				}
			} else {
				// Check if it's a single RequiredFieldError
				var reqErr *deserialize.RequiredFieldError
				if errors.As(err, &reqErr) {
					fieldErrors = append(fieldErrors, FieldError{
						Field:   reqErr.Field,
						Code:    constraints.CodeRequired,
						Message: reqErr.Error(),
					})
				} else {
					fieldErrors = append(fieldErrors, FieldError{
						Field:   fieldName,
						Message: err.Error(),
					})
				}
			}
		}
	}

	// Capture extra fields recursively if ExtraAllow mode is enabled
	if v.options.ExtraFields == ExtraAllow {
		v.captureExtrasRecursive(objValue, jsonMap, v.tagName)
	}

	// Run validation constraints
	// Continue validation even if there were deserialization errors to collect all issues
	if err := v.Validate(&obj); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			fieldErrors = append(fieldErrors, ve.Errors...)
		}
	}

	// Return all collected errors (from both deserialization and validation)
	if len(fieldErrors) > 0 {
		return &obj, &ValidationError{Errors: fieldErrors}
	}

	return &obj, nil
}

// StructPartial validates only the specified fields of a struct.
// Fields not in the list are skipped entirely.
// Field names should match JSON field names (from json tags).
func (v *Validator[T]) StructPartial(obj *T, fields ...string) error {
	if obj == nil {
		return &ValidationError{
			Errors: []FieldError{{
				Field:   "",
				Code:    "NIL_POINTER",
				Message: ErrMsgNilPointer,
			}},
		}
	}

	// Create field inclusion set
	includeSet := make(map[string]bool)
	for _, f := range fields {
		includeSet[f] = true
	}

	// If no fields specified, nothing to validate
	if len(includeSet) == 0 {
		return nil
	}

	// Validate using field cache but filter by inclusion set
	structValue := reflect.ValueOf(obj).Elem()
	var errs []FieldError

	for i := range v.fieldCache.Fields {
		cached := &v.fieldCache.Fields[i]

		// Check if this field should be validated (by JSON name)
		if !includeSet[cached.JSONName] {
			continue
		}

		fieldValue := structValue.Field(cached.FieldIndex)

		// Run constraints for this field
		for _, c := range cached.Constraints {
			err := c.Validate(fieldValue.Interface())
			if err == nil {
				continue
			}
			code := codeValidationFailed
			message := err.Error()

			var constraintErr *constraints.ConstraintError
			if errors.As(err, &constraintErr) {
				code = constraintErr.Code
				message = constraintErr.Message
			}

			errs = append(errs, FieldError{
				Field:   cached.JSONName,
				Code:    code,
				Message: message,
				Value:   fieldValue.Interface(),
			})
		}

		// Apply cross-field constraints
		for _, c := range cached.CrossFieldConstraints {
			err := c.ValidateCrossField(fieldValue.Interface(), structValue, cached.JSONName)
			if err == nil {
				continue
			}
			var valErr *ValidationError
			if errors.As(err, &valErr) {
				errs = append(errs, valErr.Errors...)
			} else {
				errs = append(errs, FieldError{
					Field:   cached.JSONName,
					Message: err.Error(),
					Value:   fieldValue.Interface(),
				})
			}
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// StructExcept validates all fields except the specified ones.
// Excluded fields are skipped entirely.
// Field names should match JSON field names (from json tags).
func (v *Validator[T]) StructExcept(obj *T, excludeFields ...string) error {
	if obj == nil {
		return &ValidationError{
			Errors: []FieldError{{
				Field:   "",
				Code:    "NIL_POINTER",
				Message: ErrMsgNilPointer,
			}},
		}
	}

	// Create field exclusion set
	excludeSet := make(map[string]bool)
	for _, f := range excludeFields {
		excludeSet[f] = true
	}

	// Validate using field cache but filter by exclusion set
	structValue := reflect.ValueOf(obj).Elem()
	var errs []FieldError

	for i := range v.fieldCache.Fields {
		cached := &v.fieldCache.Fields[i]

		// Skip excluded fields (by JSON name)
		if excludeSet[cached.JSONName] {
			continue
		}

		fieldValue := structValue.Field(cached.FieldIndex)

		// Run constraints for this field
		for _, c := range cached.Constraints {
			err := c.Validate(fieldValue.Interface())
			if err == nil {
				continue
			}
			code := codeValidationFailed
			message := err.Error()

			var constraintErr *constraints.ConstraintError
			if errors.As(err, &constraintErr) {
				code = constraintErr.Code
				message = constraintErr.Message
			}

			errs = append(errs, FieldError{
				Field:   cached.JSONName,
				Code:    code,
				Message: message,
				Value:   fieldValue.Interface(),
			})
		}

		// Apply cross-field constraints
		for _, c := range cached.CrossFieldConstraints {
			err := c.ValidateCrossField(fieldValue.Interface(), structValue, cached.JSONName)
			if err == nil {
				continue
			}
			var valErr *ValidationError
			if errors.As(err, &valErr) {
				errs = append(errs, valErr.Errors...)
			} else {
				errs = append(errs, FieldError{
					Field:   cached.JSONName,
					Message: err.Error(),
					Value:   fieldValue.Interface(),
				})
			}
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// ValidateCtx validates with context support for context-aware validators.
// Context-aware validators registered with RegisterValidationCtx will receive
// the provided context, allowing them to access request-scoped values like
// database connections, authentication info, etc.
func (v *Validator[T]) ValidateCtx(ctx context.Context, obj *T) error {
	// First, run regular validation
	if err := v.Validate(obj); err != nil {
		return err
	}

	// Then run context-aware validators
	return v.validateContextOnly(ctx, obj)
}

// UnmarshalCtx unmarshals and validates with context.
// This allows context-aware validators to access the context during unmarshal.
func (v *Validator[T]) UnmarshalCtx(ctx context.Context, data []byte) (*T, error) {
	// First unmarshal with regular validation
	obj, err := v.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	// Then run context-aware validators (regular validation already passed)
	if err := v.validateContextOnly(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// validateContextOnly runs only context-aware validators (assumes regular validation passed).
func (v *Validator[T]) validateContextOnly(ctx context.Context, obj *T) error {
	structValue := reflect.ValueOf(obj).Elem()
	var errs []FieldError

	for i := range v.fieldCache.Fields {
		cached := &v.fieldCache.Fields[i]

		// Skip if no context validators
		if len(cached.ContextConstraints) == 0 {
			continue
		}

		fieldValue := structValue.Field(cached.FieldIndex)

		// Call each context-aware validator
		for _, cc := range cached.ContextConstraints {
			fn, ok := GetContextValidator(cc.Name)
			if !ok {
				continue
			}

			if err := fn(ctx, fieldValue.Interface(), cc.Param); err != nil {
				errs = append(errs, FieldError{
					Field:   cached.JSONName,
					Code:    constraints.CodeCustomValidation,
					Message: cc.Name + ": " + err.Error(),
					Value:   fieldValue.Interface(),
				})
			}
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// unmarshalInto implements the unmarshalable interface for type-erased unmarshal.
// It calls Unmarshal (which enforces required + defaults + constraints) and copies
// the result into the target pointer.
func (v *Validator[T]) unmarshalInto(data []byte, target any) error {
	result, err := v.Unmarshal(data)
	if err != nil {
		return err
	}
	*target.(*T) = *result
	return nil
}
