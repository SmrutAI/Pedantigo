package deserialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// PlanState holds the shared state during Unmarshal.
type PlanState struct {
	Index    map[reflect.Type]*TypePlan // the type-indexed plan graph, built at New[T]()
	TagName  string
	MaxDepth int               // Options.MaxRecursionDepth (already normalized to >=1)
	depth    map[*TypePlan]int // current-path re-entry count per plan node (self-referential depth)
}

// NewPlanState returns a fresh per-Unmarshal state (its own depth map).
func NewPlanState(index map[reflect.Type]*TypePlan, tagName string, maxDepth int) *PlanState {
	return &PlanState{Index: index, TagName: tagName, MaxDepth: maxDepth, depth: map[*TypePlan]int{}}
}

// DecodeStruct walks a precomputed TypePlan against a decoded map[string]any.
// It enforces self-referential depth limits, applies defaults, and collects required field errors.
func DecodeStruct(structVal reflect.Value, inputMap map[string]any, plan *TypePlan, st *PlanState, path string) error {
	// Depth guard (self-referential only): increment on entry, decrement on exit
	st.depth[plan]++
	if st.depth[plan] > st.MaxDepth {
		st.depth[plan]--
		return &MaxDepthExceededError{Path: path, Limit: st.MaxDepth}
	}
	defer func() { st.depth[plan]-- }()

	var reqErrs []*RequiredFieldError

	// isRoot is true only for the struct passed directly to Unmarshal/UnmarshalInto
	// (never for a nested struct or dive element reached via recursion - those
	// always pass a non-empty path). Two legacy behaviors, both preserved from the
	// old per-field-closure deserializer, key off this distinction:
	//   - required-field errors on ROOT fields use the JSON field name and carry
	//     no Code; on NESTED fields they use the dotted Go-field path and
	//     Code: CodeRequired (set by the caller in validator.go from IsRoot).
	//   - string transformations (strip_whitespace/to_lower/to_upper) were only
	//     ever applied by the old top-level deserializer closures; nested/dive
	//     struct fields were never mutated during deserialize.
	isRoot := path == ""

	// Iterate through each field in the plan
	for _, fp := range plan.Fields {
		// Build fully-qualified field path using the precomputed Go field name
		goFieldName := fp.GoName
		fieldPath := path
		if fieldPath == "" {
			fieldPath = goFieldName
		} else {
			fieldPath = fieldPath + "." + goFieldName
		}

		applyTransform := func(fv reflect.Value) {
			if isRoot && (fp.StripWhitespace || fp.ToLower || fp.ToUpper) {
				applyStringTransformations(fv, StringTransformations{
					StripWhitespace: fp.StripWhitespace,
					ToLower:         fp.ToLower,
					ToUpper:         fp.ToUpper,
				})
			}
		}

		// Get the field value from the input map
		val, ok := inputMap[fp.JSONName]

		if !ok {
			// Field is absent from input
			if fp.StaticDefault != nil {
				// Apply static default
				fieldValue := structVal.Field(fp.FieldIndex)
				var setDefault func(reflect.Value, string)
				setDefault = func(fv reflect.Value, dv string) {
					SetDefaultValue(fv, dv, setDefault)
				}
				setDefault(fieldValue, *fp.StaticDefault)
				applyTransform(fieldValue)
				continue
			}

			if fp.DefaultMethod != "" {
				// Call default method
				fieldValue := structVal.Field(fp.FieldIndex)
				if structVal.CanAddr() {
					ptrValue := structVal.Addr()
					method := ptrValue.MethodByName(fp.DefaultMethod)
					if method.IsValid() {
						results := method.Call(nil)
						if len(results) == 2 {
							if !results[1].IsNil() {
								return results[1].Interface().(error)
							}
							fieldValue.Set(results[0])
						}
					}
				}
				applyTransform(fieldValue)
				continue
			}

			// No default - check if required
			if fp.Required {
				if isRoot {
					reqErrs = append(reqErrs, &RequiredFieldError{Field: fp.JSONName, IsRoot: true})
				} else {
					reqErrs = append(reqErrs, &RequiredFieldError{Field: fieldPath, IsRoot: false})
				}
			}
			continue
		}

		// Field is present - set the value
		fieldValue := structVal.Field(fp.FieldIndex)
		if err := setValue(fieldValue, val, fp.Nested, fp.DecoderKind, fp.Kind, fp.IsPointer, fp.ElemNested, fp.ElemDecoderKind, st, fieldPath); err != nil {
			// Check if it's a multi-error from nested struct
			var multiErr *MultiRequiredFieldError
			if errors.As(err, &multiErr) {
				reqErrs = append(reqErrs, multiErr.Errors...)
			} else {
				// Wrap with the field that caused it (JSON name at root, dotted
				// Go path when nested) so the caller can attribute the error
				// instead of falling back to the struct root.
				attributedField := fieldPath
				if isRoot {
					attributedField = fp.JSONName
				}
				return &FieldDecodeError{Field: attributedField, Err: err}
			}
		}

		// Apply string transformations after successfully setting the field
		applyTransform(fieldValue)
	}

	// ExtraAllow capture: if plan has extra_fields field, collect unmapped keys
	if plan.ExtraFieldIndex >= 0 {
		extra := make(map[string]any)
		// Collect any extra keys not in the plan, using the precomputed JSON-name set
		for key, value := range inputMap {
			if _, expected := plan.JSONNameSet[key]; !expected {
				extra[key] = value
			}
		}
		// Set the extra_fields field
		extraField := structVal.Field(plan.ExtraFieldIndex)
		extraField.Set(reflect.ValueOf(extra))
	}

	// Return collected required errors
	if len(reqErrs) > 0 {
		return &MultiRequiredFieldError{Errors: reqErrs}
	}

	return nil
}

// setValue dispatches on field kind and decoder type to set the field value.
// It handles pointers, custom decoders, structs, slices, maps, and scalars.
func setValue(dst reflect.Value, decoded any, nested *TypePlan, dk DecoderKind, kind FieldKind, isPtr bool, elemNested *TypePlan, elemDK DecoderKind, st *PlanState, path string) error {
	// Handle pointer types: allocate and recurse
	if isPtr {
		if decoded == nil {
			// Explicit JSON null for a pointer
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		// Allocate new pointer and set its dereferenced element
		np := reflect.New(dst.Type().Elem())
		if err := setValue(np.Elem(), decoded, nested, dk, kind, false, elemNested, elemDK, st, path); err != nil {
			return err
		}
		dst.Set(np)
		return nil
	}

	// Handle nil values: set zero value. A non-pointer struct cannot legitimately
	// be JSON null (only a pointer can be absent/null) - reject it instead of
	// silently zero-filling, matching the old setSliceField null-struct guard.
	if decoded == nil {
		if dst.Kind() == reflect.Struct {
			return fmt.Errorf("%s: null is not a valid value for %s", path, dst.Type())
		}
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	// Handle custom decoders first (WalkerDecoder / json.Unmarshaler)
	if dk == DecoderWalker {
		if !dst.CanAddr() {
			return fmt.Errorf("%s: cannot address value for WalkerDecoder", path)
		}
		wd, ok := dst.Addr().Interface().(WalkerDecoder)
		if !ok {
			return fmt.Errorf("%s: type does not implement WalkerDecoder", path)
		}
		recurse := func(d2 any, dec2 any) error {
			rv := reflect.ValueOf(d2).Elem()
			return setValueForType(rv, dec2, st, path)
		}
		return wd.DecodeWalk(decoded, recurse)
	}

	if dk == DecoderJSONUnmarshaler {
		if !dst.CanAddr() {
			return fmt.Errorf("%s: cannot address value for json.Unmarshaler", path)
		}
		raw, err := json.Marshal(decoded)
		if err != nil {
			return err
		}
		u, ok := dst.Addr().Interface().(json.Unmarshaler)
		if !ok {
			return fmt.Errorf("%s: type does not implement json.Unmarshaler", path)
		}
		return u.UnmarshalJSON(raw)
	}

	// Dispatch on field kind for normal types
	switch kind {
	case KindStruct:
		m, ok := decoded.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object, got %T", path, decoded)
		}
		return DecodeStruct(dst, m, nested, st, path)

	case KindSlice:
		return decodeSlice(dst, decoded, elemNested, elemDK, st, path)

	case KindMap:
		return decodeMap(dst, decoded, elemNested, elemDK, st, path)

	case KindScalar:
		// Reuse SetFieldValueWithOptions for scalar leaves
		recurse := func(fv reflect.Value, iv any, ft reflect.Type, opts FieldOptions) error {
			return setValueForType(fv, iv, st, opts.Path)
		}
		return SetFieldValueWithOptions(dst, decoded, dst.Type(), recurse, FieldOptions{
			StrictMissingFields: true,
			TagName:             st.TagName,
			Path:                path,
		})

	default:
		return fmt.Errorf("%s: unknown field kind %v", path, kind)
	}
}

// decodeSlice handles deserialization of slice types.
// It dispatches each element via setValueForType (which fixes the gap where
// custom-decoder elements were not being dispatched).
func decodeSlice(dst reflect.Value, decoded any, elemNested *TypePlan, elemDK DecoderKind, st *PlanState, path string) error {
	arr, ok := decoded.([]any)
	if !ok {
		return fmt.Errorf("%s: expected array, got %T", path, decoded)
	}

	elemType := dst.Type().Elem()
	slice := reflect.MakeSlice(dst.Type(), len(arr), len(arr))
	var reqErrs []*RequiredFieldError

	for i, elemDecoded := range arr {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		elem := slice.Index(i)

		// Dispatch each element through setValueForType (handles custom decoders)
		if elemNested != nil || elemDK != DecoderNone {
			if err := setValueForType(elem, elemDecoded, st, elemPath); err != nil {
				var multiErr *MultiRequiredFieldError
				if errors.As(err, &multiErr) {
					reqErrs = append(reqErrs, multiErr.Errors...)
				} else {
					return err
				}
			}
		} else {
			// Scalar element - use SetFieldValueWithOptions
			recurse := func(fv reflect.Value, iv any, ft reflect.Type, opts FieldOptions) error {
				return setValueForType(fv, iv, st, opts.Path)
			}
			if err := SetFieldValueWithOptions(elem, elemDecoded, elemType, recurse, FieldOptions{
				StrictMissingFields: true,
				TagName:             st.TagName,
				Path:                elemPath,
			}); err != nil {
				return err
			}
		}
	}

	dst.Set(slice)

	if len(reqErrs) > 0 {
		return &MultiRequiredFieldError{Errors: reqErrs}
	}
	return nil
}

// decodeMap handles deserialization of map types.
// It dispatches each value via setValueForType (which fixes the gap where
// custom-decoder values were not being dispatched).
func decodeMap(dst reflect.Value, decoded any, elemNested *TypePlan, elemDK DecoderKind, st *PlanState, path string) error {
	m, ok := decoded.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: expected object, got %T", path, decoded)
	}

	keyType := dst.Type().Key()
	elemType := dst.Type().Elem()
	newMap := reflect.MakeMap(dst.Type())
	var reqErrs []*RequiredFieldError

	for k, v := range m {
		elemPath := fmt.Sprintf("%s[%s]", path, k)
		elem := reflect.New(elemType).Elem()

		// Dispatch each value through setValueForType (handles custom decoders)
		if elemNested != nil || elemDK != DecoderNone {
			if err := setValueForType(elem, v, st, elemPath); err != nil {
				var multiErr *MultiRequiredFieldError
				if errors.As(err, &multiErr) {
					reqErrs = append(reqErrs, multiErr.Errors...)
				} else {
					return err
				}
			}
		} else {
			// Scalar value - use SetFieldValueWithOptions
			recurse := func(fv reflect.Value, iv any, ft reflect.Type, opts FieldOptions) error {
				return setValueForType(fv, iv, st, opts.Path)
			}
			if err := SetFieldValueWithOptions(elem, v, elemType, recurse, FieldOptions{
				StrictMissingFields: true,
				TagName:             st.TagName,
				Path:                elemPath,
			}); err != nil {
				return err
			}
		}

		// Convert key if needed
		var convertedKey reflect.Value
		keyVal := reflect.ValueOf(k)
		switch {
		case keyVal.Type().AssignableTo(keyType):
			convertedKey = keyVal
		case keyVal.Type().ConvertibleTo(keyType):
			convertedKey = keyVal.Convert(keyType)
		default:
			return fmt.Errorf("%s: cannot convert map key %v to %v", elemPath, keyVal.Type(), keyType)
		}

		newMap.SetMapIndex(convertedKey, elem)
	}

	dst.Set(newMap)

	if len(reqErrs) > 0 {
		return &MultiRequiredFieldError{Errors: reqErrs}
	}
	return nil
}

// setValueForType is the generic recurse target used by WalkerDecoder's recurse callback
// and by decodeSlice/decodeMap for struct/custom-decoder elements.
// It handles pointer dereferencing, struct dispatch, slice/map iteration, and scalar leaves.
func setValueForType(dst reflect.Value, decoded any, st *PlanState, path string) error {
	// Handle pointer types: allocate if needed
	if dst.Kind() == reflect.Pointer {
		if decoded == nil {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		np := reflect.New(dst.Type().Elem())
		if err := setValueForType(np.Elem(), decoded, st, path); err != nil {
			return err
		}
		dst.Set(np)
		return nil
	}

	// Handle nil for non-pointer types. A non-pointer struct cannot legitimately
	// be JSON null - reject it instead of silently zero-filling, matching the
	// old setSliceField null-struct guard.
	if decoded == nil {
		if dst.Kind() == reflect.Struct {
			return fmt.Errorf("%s: null is not a valid value for %s", path, dst.Type())
		}
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	t := dst.Type()

	// Handle struct types: look up or build plan
	if t.Kind() == reflect.Struct {
		plan := st.Index[t]
		if plan == nil {
			// st.Index is the Validator[T]'s shared planIndex, reused across
			// concurrent Unmarshal calls - never mutate it here. Build into a
			// throwaway map instead of caching into the shared index.
			plan = BuildTypePlan(t, st.TagName, map[reflect.Type]*TypePlan{})
		}

		// Dispatch on self-decoder kind
		switch plan.SelfDecoderKind {
		case DecoderWalker:
			if !dst.CanAddr() {
				return fmt.Errorf("%s: cannot address value for WalkerDecoder", path)
			}
			wd, ok := dst.Addr().Interface().(WalkerDecoder)
			if !ok {
				return fmt.Errorf("%s: type does not implement WalkerDecoder", path)
			}
			recurse := func(d2 any, dec2 any) error {
				rv := reflect.ValueOf(d2).Elem()
				return setValueForType(rv, dec2, st, path)
			}
			return wd.DecodeWalk(decoded, recurse)

		case DecoderJSONUnmarshaler:
			if !dst.CanAddr() {
				return fmt.Errorf("%s: cannot address value for json.Unmarshaler", path)
			}
			raw, err := json.Marshal(decoded)
			if err != nil {
				return err
			}
			u, ok := dst.Addr().Interface().(json.Unmarshaler)
			if !ok {
				return fmt.Errorf("%s: type does not implement json.Unmarshaler", path)
			}
			return u.UnmarshalJSON(raw)

		case DecoderNone:
			m, ok := decoded.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected object, got %T", path, decoded)
			}
			return DecodeStruct(dst, m, plan, st, path)
		}
	}

	// Handle slice types
	if t.Kind() == reflect.Slice {
		arr, ok := decoded.([]any)
		if !ok {
			return fmt.Errorf("%s: expected array, got %T", path, decoded)
		}

		slice := reflect.MakeSlice(t, len(arr), len(arr))
		var reqErrs []*RequiredFieldError

		for i, elemDecoded := range arr {
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			elem := slice.Index(i)

			if err := setValueForType(elem, elemDecoded, st, elemPath); err != nil {
				var multiErr *MultiRequiredFieldError
				if errors.As(err, &multiErr) {
					reqErrs = append(reqErrs, multiErr.Errors...)
				} else {
					return err
				}
			}
		}

		dst.Set(slice)

		if len(reqErrs) > 0 {
			return &MultiRequiredFieldError{Errors: reqErrs}
		}
		return nil
	}

	// Handle map types
	if t.Kind() == reflect.Map {
		m, ok := decoded.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object, got %T", path, decoded)
		}

		keyType := t.Key()
		elemType := t.Elem()
		newMap := reflect.MakeMap(t)
		var reqErrs []*RequiredFieldError

		for k, v := range m {
			elemPath := fmt.Sprintf("%s[%s]", path, k)
			elem := reflect.New(elemType).Elem()

			if err := setValueForType(elem, v, st, elemPath); err != nil {
				var multiErr *MultiRequiredFieldError
				if errors.As(err, &multiErr) {
					reqErrs = append(reqErrs, multiErr.Errors...)
				} else {
					return err
				}
			}

			// Convert key
			var convertedKey reflect.Value
			keyVal := reflect.ValueOf(k)
			switch {
			case keyVal.Type().AssignableTo(keyType):
				convertedKey = keyVal
			case keyVal.Type().ConvertibleTo(keyType):
				convertedKey = keyVal.Convert(keyType)
			default:
				return fmt.Errorf("%s: cannot convert map key %v to %v", elemPath, keyVal.Type(), keyType)
			}

			newMap.SetMapIndex(convertedKey, elem)
		}

		dst.Set(newMap)

		if len(reqErrs) > 0 {
			return &MultiRequiredFieldError{Errors: reqErrs}
		}
		return nil
	}

	// Scalar leaf: use SetFieldValueWithOptions
	recurse := func(fv reflect.Value, iv any, ft reflect.Type, opts FieldOptions) error {
		return setValueForType(fv, iv, st, opts.Path)
	}
	return SetFieldValueWithOptions(dst, decoded, t, recurse, FieldOptions{
		StrictMissingFields: true,
		TagName:             st.TagName,
		Path:                path,
	})
}
