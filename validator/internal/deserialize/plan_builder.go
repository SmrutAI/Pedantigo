package deserialize

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/SmrutAI/pedantigo/v2/validator/internal/tags"
)

// BuildTypePlan returns the plan for typ, registering it and every reachable
// struct type in index (shared/deduplicated per reflect.Type). Cyclic types get
// a back-edge to their in-progress node. Call with a fresh index at New[T]().
func BuildTypePlan(typ reflect.Type, tagName string, index map[reflect.Type]*TypePlan) *TypePlan {
	// Step 1: Deref pointers
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	// If not a struct, return a simple TypePlan with no Fields
	if typ.Kind() != reflect.Struct {
		return &TypePlan{
			Type:            typ,
			ExtraFieldIndex: -1,
			SelfDecoderKind: decoderKindOf(typ),
		}
	}

	// Otherwise use BuildNode to handle cycles
	return BuildNode(
		typ,
		index,
		func() *TypePlan {
			return &TypePlan{
				Type:            typ,
				ExtraFieldIndex: -1,
				SelfDecoderKind: decoderKindOf(typ),
			}
		},
		func(p *TypePlan) {
			// Step 3: Iterate through struct fields
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)

				// Step 4: Skip unexported fields
				if !field.IsExported() {
					continue
				}

				// Step 5: Handle extra_fields tag. Checked BEFORE the json:"-" skip below,
				// because the extra_fields map field is always tagged json:"-" (it is not
				// itself a real JSON key) - checking json:"-" first would always skip it
				// and ExtraFieldIndex would never be set. Mirrors DetectExtraField in
				// extras.go, which never consults the json tag for this check.
				pedantigoTag := field.Tag.Get(tagName)
				if pedantigoTag == tags.ExtraFieldsTag {
					p.ExtraFieldIndex = i
					continue
				}

				// Read json tag and skip if marked with "-"
				jsonTag := field.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}

				// Step 6: Resolve jsonName
				var jsonName string
				if jsonTag != "" {
					jsonName, _, _ = strings.Cut(jsonTag, ",")
				} else {
					jsonName = field.Name
				}

				// Step 7: Parse tags and build FieldPlan
				parsed := tags.ParseTagWithName(field.Tag, tagName)

				fp := FieldPlan{
					FieldIndex:      i,
					GoName:          field.Name,
					JSONName:        jsonName,
					Required:        parsed != nil && containsKey(parsed, "required"),
					StripWhitespace: parsed != nil && containsKey(parsed, "strip_whitespace"),
					ToLower:         parsed != nil && containsKey(parsed, "to_lower"),
					ToUpper:         parsed != nil && containsKey(parsed, "to_upper"),
				}

				// Set StaticDefault if present
				if parsed != nil {
					if defVal, ok := parsed["default"]; ok {
						fp.StaticDefault = &defVal
					}
					if method, ok := parsed["defaultUsingMethod"]; ok {
						fp.DefaultMethod = method
					}
				}

				// Step 8: Field type shape
				ft := field.Type
				fp.IsPointer = (ft.Kind() == reflect.Pointer)

				dft := ft
				if fp.IsPointer {
					dft = ft.Elem()
				}

				fp.DecoderKind = decoderKindOf(dft)

				// Determine Kind and nested plans. Switch on dft (deref'd type),
				// not ft, so a pointer-to-slice/pointer-to-map field (e.g. *[]T,
				// *map[K]V) is classified correctly instead of falling through
				// to KindScalar.
				switch {
				case dft.Kind() == reflect.Struct:
					fp.Kind = KindStruct
					fp.Nested = BuildTypePlan(dft, tagName, index)
				case dft.Kind() == reflect.Slice:
					fp.Kind = KindSlice
					et := dft.Elem()
					if et.Kind() == reflect.Pointer {
						et = et.Elem()
					}
					fp.ElemDecoderKind = decoderKindOf(et)
					if et.Kind() == reflect.Struct {
						fp.ElemNested = BuildTypePlan(et, tagName, index)
					}
				case dft.Kind() == reflect.Map:
					fp.Kind = KindMap
					vt := dft.Elem()
					if vt.Kind() == reflect.Pointer {
						vt = vt.Elem()
					}
					fp.ElemDecoderKind = decoderKindOf(vt)
					if vt.Kind() == reflect.Struct {
						fp.ElemNested = BuildTypePlan(vt, tagName, index)
					}
				default:
					fp.Kind = KindScalar
				}

				// Step 9: Append to Fields
				p.Fields = append(p.Fields, fp)
			}

			// Precompute the JSON-name set once for the ExtraAllow capture path
			// (interpreter.go), instead of rebuilding it on every DecodeStruct call.
			if p.ExtraFieldIndex >= 0 {
				p.JSONNameSet = make(map[string]struct{}, len(p.Fields))
				for _, fp := range p.Fields {
					p.JSONNameSet[fp.JSONName] = struct{}{}
				}
			}
		},
	)
}

// decoderKindOf determines whether a type implements WalkerDecoder or json.Unmarshaler.
func decoderKindOf(t reflect.Type) DecoderKind {
	pt := reflect.PointerTo(t)
	switch {
	case pt.Implements(walkerDecoderType):
		return DecoderWalker
	case pt.Implements(jsonUnmarshalerType):
		return DecoderJSONUnmarshaler
	default:
		return DecoderNone
	}
}

// containsKey is a helper to check if a map has a key (ignoring its value).
func containsKey(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

// ValidatePlanDefaults performs the fail-fast checks that used to run inside
// the old BuildFieldDeserializers: default=/defaultUsingMethod= tags are
// incompatible with StrictMissingFields=false, and defaultUsingMethod= must
// name a method with signature func(*T) (FieldType, error). Called once from
// New[T]() over the full plan index built by BuildTypePlan.
func ValidatePlanDefaults(index map[reflect.Type]*TypePlan, strictMissingFields bool) {
	for _, plan := range index {
		for _, fp := range plan.Fields {
			if fp.StaticDefault == nil && fp.DefaultMethod == "" {
				continue
			}
			field := plan.Type.Field(fp.FieldIndex)
			if !strictMissingFields {
				if fp.StaticDefault != nil {
					panic(fmt.Sprintf("field %s.%s has 'default=' tag but StrictMissingFields is false. Remove the tag or enable StrictMissingFields.",
						plan.Type.Name(), field.Name))
				}
				panic(fmt.Sprintf("field %s.%s has 'defaultUsingMethod=' tag but StrictMissingFields is false. Remove the tag or enable StrictMissingFields.",
					plan.Type.Name(), field.Name))
			}
			if fp.DefaultMethod != "" {
				if err := ValidateDefaultMethod(plan.Type, fp.DefaultMethod, field.Type); err != nil {
					panic(fmt.Sprintf("field %s: %v", field.Name, err))
				}
			}
		}
	}
}
