package pedantigo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/invopop/jsonschema"

	"github.com/SmrutAI/pedantigo/internal/constraints"
	"github.com/SmrutAI/pedantigo/internal/deserialize"
	"github.com/SmrutAI/pedantigo/internal/serialize"
	"github.com/SmrutAI/pedantigo/internal/tags"
)

// Validator validates structs of type T.
type Validator[T any] struct {
	typ                reflect.Type
	options            ValidatorOptions
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

	// Extra fields support (nil when ExtraAllow disabled)
	extraFieldInfo *deserialize.ExtraFieldInfo
}

// New creates a new Validator for type T with optional configuration.
func New[T any](opts ...ValidatorOptions) *Validator[T] {
	// Mark that a validator has been created (prevents late SetTagName calls)
	markValidatorCreated()

	var zero T
	typ := reflect.TypeOf(zero)

	options := DefaultValidatorOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Resolve tag name (instance override or global)
	tagName := resolveTagName(options)

	validator := &Validator[T]{
		typ:                typ,
		options:            options,
		tagName:            tagName,
		fieldDeserializers: make(map[string]deserialize.FieldDeserializer),
	}

	// Build field deserializers at creation time (fail-fast)
	validator.fieldDeserializers = deserialize.BuildFieldDeserializers(
		typ,
		deserialize.BuilderOptions{
			StrictMissingFields: options.StrictMissingFields,
			TagName:             tagName,
		},
		validator.setFieldValue,
		validator.setDefaultValue,
	)

	// Validate dive/keys/endkeys tag usage at creation time (fail-fast)
	validator.validateDiveTags(typ, tagName)

	// Build field constraints at creation time (the key optimization)
	validator.fieldCache = validator.buildFieldConstraints(typ, tagName)

	// Detect extra_fields for ExtraAllow mode
	if options.ExtraFields == ExtraAllow {
		validator.extraFieldInfo = deserialize.DetectExtraField(typ, tagName)
		if validator.extraFieldInfo == nil {
			panic(ErrMsgExtraFieldRequired)
		}
	}

	return validator
}

// buildFieldConstraints builds and caches all field constraints at creation time.
func (v *Validator[T]) buildFieldConstraints(typ reflect.Type, tagName string) *constraints.FieldCache {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

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
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		isCollection := fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map
		isMap := fieldType.Kind() == reflect.Map

		cached := constraints.CachedField{
			Name:         field.Name,
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

			// Constraints before dive (or regular field constraints)
			if len(parsedTag.CollectionConstraints) > 0 {
				cached.Constraints = constraints.BuildConstraints(parsedTag.CollectionConstraints, field.Type)
			}

			// Element constraints after dive
			if parsedTag.DivePresent && len(parsedTag.ElementConstraints) > 0 {
				cached.ElementConstraints = constraints.BuildConstraints(parsedTag.ElementConstraints, field.Type.Elem())
			}

			// Map key constraints
			if isMap && len(parsedTag.KeyConstraints) > 0 {
				cached.KeyConstraints = constraints.BuildConstraints(parsedTag.KeyConstraints, field.Type.Key())
			}

			// Cross-field constraints (eqfield, gtfield, etc.)
			cached.CrossFieldConstraints = constraints.BuildCrossFieldConstraintsForField(
				parsedTag.CollectionConstraints, typ, i)
		}

		// Recurse for nested structs
		switch fieldType.Kind() {
		case reflect.Struct:
			cached.NestedCache = v.buildFieldConstraints(fieldType, tagName)
		case reflect.Slice, reflect.Map:
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				cached.NestedCache = v.buildFieldConstraints(elemType, tagName)
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
	if typ.Kind() == reflect.Ptr {
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
		if fieldType.Kind() == reflect.Ptr {
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
func (v *Validator[T]) setFieldValue(fieldValue reflect.Value, inValue any, fieldType reflect.Type) error {
	return deserialize.SetFieldValue(fieldValue, inValue, fieldType, v.setFieldValue)
}

// Validate validates a struct and returns any validation errors
// NOTE: 'required' is NOT checked here - it's only checked during Unmarshal
// Validate checks if the value satisfies the constraint.
func (v *Validator[T]) Validate(obj *T) error {
	if obj == nil {
		return &ValidationError{
			Errors: []FieldError{{Field: "root", Message: "cannot validate nil pointer"}},
		}
	}

	// Get context from pool
	ctx := validateContextPool.Get().(*validateContext)

	// Reset buffers (keep capacity)
	ctx.pathBuf = ctx.pathBuf[:0]
	ctx.errs = ctx.errs[:0]

	// Validate all fields using struct tags (required is skipped via buildConstraints)
	v.validateWithCache(reflect.ValueOf(obj).Elem(), nil, ctx, v.fieldCache)

	// Check if struct implements Validatable for cross-field validation
	if validatable, ok := any(obj).(Validatable); ok {
		if err := validatable.Validate(); err != nil {
			// Check if it's a ValidationError with multiple errors
			var ve *ValidationError
			if errors.As(err, &ve) {
				ctx.errs = append(ctx.errs, ve.Errors...)
			} else {
				// Single error or custom error type
				ctx.errs = append(ctx.errs, FieldError{
					Field:   "root",
					Message: err.Error(),
				})
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
	for val.Kind() == reflect.Ptr {
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

		// Build field path using buffer
		fieldPath := appendPath(ctx.pathBuf[:0], path, cached.Name)

		// Check required for nested struct fields (path != nil)
		if len(path) > 0 && v.options.StrictMissingFields && cached.IsRequired {
			if fieldVal.IsZero() {
				ctx.errs = append(ctx.errs, FieldError{
					Field:   string(fieldPath),
					Code:    constraints.CodeRequired,
					Message: "is required",
					Value:   fieldVal.Interface(),
				})
				continue // Skip further validation for this field
			}
		}

		// Apply field constraints
		for _, c := range cached.Constraints {
			if err := c.Validate(fieldVal.Interface()); err != nil {
				ctx.errs = append(ctx.errs, v.newFieldError(string(fieldPath), err, fieldVal.Interface()))
			}
		}

		// Apply cross-field constraints
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

		// Handle collections with dive (requires dive to recurse into elements, like playground)
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
						Field:   "root",
						Message: "JSON decode error: " + ErrMsgUnknownField,
					}},
				}
			}
		} else {
			if err := json.Unmarshal(data, &obj); err != nil {
				return nil, &ValidationError{
					Errors: []FieldError{{
						Field:   "root",
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
					Field:   "root",
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
				Field:   "root",
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
			fieldErrors = append(fieldErrors, FieldError{
				Field:   fieldName,
				Message: err.Error(),
			})
		}
	}

	// Return early if deserialization errors
	if len(fieldErrors) > 0 {
		return &obj, &ValidationError{Errors: fieldErrors}
	}

	// Step 3.5: Capture extra fields recursively (for this struct and all nested structs)
	if v.options.ExtraFields == ExtraAllow {
		v.captureExtrasRecursive(objValue, jsonMap, v.tagName)
	}

	// Step 4: Run validation constraints (min, max, email, etc.)
	// NOTE: 'required' is already skipped in Validate() via buildConstraints
	if err := v.Validate(&obj); err != nil {
		return &obj, err
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
	case reflect.Ptr:
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
	isPtrStructSlice := elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct

	if !isStructSlice && !isPtrStructSlice {
		return
	}

	sliceJSON, ok := nestedJSON.([]any)
	if !ok {
		return
	}

	for idx := 0; idx < fieldVal.Len() && idx < len(sliceJSON); idx++ {
		elemVal := fieldVal.Index(idx)
		if elemType.Kind() == reflect.Ptr {
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
	if objVal.Kind() == reflect.Ptr {
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
	case reflect.Ptr:
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
	isPtrStructSlice := elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct

	if !isStructSlice && !isPtrStructSlice {
		return
	}

	sliceAny, ok := resultMap[fieldName].([]any)
	if !ok {
		return
	}

	for idx := 0; idx < fieldVal.Len() && idx < len(sliceAny); idx++ {
		elemVal := fieldVal.Index(idx)
		if elemType.Kind() == reflect.Ptr {
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
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return []byte("null"), nil
		}
		val = val.Elem()
	}

	metadata := serialize.BuildFieldMetadata(val.Type(), v.tagName)

	// Convert options
	serializeOpts := serialize.SerializeOptions{
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
				Errors: []FieldError{{Field: "root", Message: "cannot validate nil pointer"}},
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
				Field:   "root",
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
			fieldErrors = append(fieldErrors, FieldError{
				Field:   fieldName,
				Message: err.Error(),
			})
		}
	}

	if len(fieldErrors) > 0 {
		return &obj, &ValidationError{Errors: fieldErrors}
	}

	// Capture extra fields recursively if ExtraAllow mode is enabled
	if v.options.ExtraFields == ExtraAllow {
		v.captureExtrasRecursive(objValue, jsonMap, v.tagName)
	}

	// Run validation constraints
	if err := v.Validate(&obj); err != nil {
		return &obj, err
	}

	return &obj, nil
}
