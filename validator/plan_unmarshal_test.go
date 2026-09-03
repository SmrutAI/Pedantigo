package validator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Test #1: NestedRequired ===
// Tests a two-level struct where inner.name is required and missing.

type innerRequired struct {
	Name string `json:"name" validate:"required"`
}

type outerRequired struct {
	Inner innerRequired `json:"inner"`
}

func TestPlan_NestedRequired(t *testing.T) {
	// Valid case: both levels present
	out, err := New[outerRequired]().Unmarshal([]byte(`{"inner":{"name":"test"}}`))
	require.NoError(t, err)
	require.Equal(t, "test", out.Inner.Name)

	// Invalid case: nested.name missing
	// Field should be "Inner.Name" (dotted Go path, since it's nested not root-level)
	_, err = New[outerRequired]().Unmarshal([]byte(`{"inner":{}}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Equal(t, "Inner.Name", ve.Errors[0].Field)
}

// === Test #2: StaticDefaultNested ===
// Tests a nested field with static default: absent key uses default,
// explicit zero value is preserved (not overwritten).

type innerStaticDefault struct {
	Value string `json:"value" validate:"default=hello"`
}

type outerStaticDefault struct {
	Inner innerStaticDefault `json:"inner"`
}

func TestPlan_StaticDefaultNested(t *testing.T) {
	// Absent key: default applied
	out, err := New[outerStaticDefault]().Unmarshal([]byte(`{"inner":{}}`))
	require.NoError(t, err)
	require.Equal(t, "hello", out.Inner.Value)

	// Explicit empty string: NOT overwritten by default
	out, err = New[outerStaticDefault]().Unmarshal([]byte(`{"inner":{"value":""}}`))
	require.NoError(t, err)
	require.Empty(t, out.Inner.Value)

	// Explicit different value: used as-is
	out, err = New[outerStaticDefault]().Unmarshal([]byte(`{"inner":{"value":"world"}}`))
	require.NoError(t, err)
	require.Equal(t, "world", out.Inner.Value)
}

// === Test #3: DefaultUsingMethodNested ===
// Tests a nested field with defaultUsingMethod tag where the struct
// holding the field has the method.

type innerUsingMethod struct {
	Value string `json:"value" validate:"defaultUsingMethod=GetDefault"`
}

func (i *innerUsingMethod) GetDefault() (string, error) {
	return "method_default", nil
}

type outerUsingMethod struct {
	Inner innerUsingMethod `json:"inner"`
}

func TestPlan_DefaultUsingMethodNested(t *testing.T) {
	// Absent key: method default applied
	out, err := New[outerUsingMethod]().Unmarshal([]byte(`{"inner":{}}`))
	require.NoError(t, err)
	require.Equal(t, "method_default", out.Inner.Value)

	// Explicit value: used as-is
	out, err = New[outerUsingMethod]().Unmarshal([]byte(`{"inner":{"value":"explicit"}}`))
	require.NoError(t, err)
	require.Equal(t, "explicit", out.Inner.Value)
}

// === Test #4: SliceOfNestedRequired ===
// Tests a slice of structs where one element is missing a required field.
// Verifies error aggregation and dotted path naming (Items[1].Name).

type itemInSlice struct {
	Name string `json:"name" validate:"required"`
}

type containerSlice struct {
	Items []itemInSlice `json:"items" validate:"dive"`
}

func TestPlan_SliceOfNestedRequired(t *testing.T) {
	// Valid: both items have name
	out, err := New[containerSlice]().Unmarshal([]byte(`{"items":[{"name":"item1"},{"name":"item2"}]}`))
	require.NoError(t, err)
	require.Len(t, out.Items, 2)
	require.Equal(t, "item1", out.Items[0].Name)
	require.Equal(t, "item2", out.Items[1].Name)

	// Invalid: second item missing name
	_, err = New[containerSlice]().Unmarshal([]byte(`{"items":[{"name":"item1"},{}]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Equal(t, "Items[1].Name", ve.Errors[0].Field)

	// Invalid: both items missing name (aggregation test)
	_, err = New[containerSlice]().Unmarshal([]byte(`{"items":[{},{}]}`))
	require.Error(t, err)
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 2)
	require.Contains(t, ve.Errors[0].Field, "Items[0]")
	require.Contains(t, ve.Errors[0].Field, "Name")
	require.Contains(t, ve.Errors[1].Field, "Items[1]")
	require.Contains(t, ve.Errors[1].Field, "Name")
	require.NotEqual(t, ve.Errors[0].Field, ve.Errors[1].Field)
}

// === Test #5: MapOfNestedRequired ===
// Tests a map of structs where a value is missing a required field.

type containerMap struct {
	Items map[string]itemInSlice `json:"items" validate:"dive"`
}

func TestPlan_MapOfNestedRequired(t *testing.T) {
	// Valid: map values have required field
	out, err := New[containerMap]().Unmarshal([]byte(`{"items":{"key1":{"name":"value1"},"key2":{"name":"value2"}}}`))
	require.NoError(t, err)
	require.Len(t, out.Items, 2)
	require.Equal(t, "value1", out.Items["key1"].Name)
	require.Equal(t, "value2", out.Items["key2"].Name)

	// Invalid: one map value missing required field
	_, err = New[containerMap]().Unmarshal([]byte(`{"items":{"key1":{"name":"value1"},"key2":{}}}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	// Field should contain both Items and Name for the missing field in map value
	require.Contains(t, ve.Errors[0].Field, "Items")
	require.Contains(t, ve.Errors[0].Field, "Name")
}

// === Test #6: WalkerDecoderSliceElement ===
// Tests a custom WalkerDecoder type as a slice element, verifying that
// nested required fields are enforced and captured fields are correct.

type wdSliceItem struct {
	Kind  string `json:"kind" validate:"required"`
	Extra map[string]any
}

func (item *wdSliceItem) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	m, ok := decoded.(map[string]any)
	if !ok {
		return fmt.Errorf("wdSliceItem must be object, got %T", decoded)
	}
	type shadow struct {
		Kind string `json:"kind" validate:"required"`
	}
	var s shadow
	if err := recurse(&s, decoded); err != nil {
		return err
	}
	item.Kind = s.Kind
	item.Extra = map[string]any{}
	for k, v := range m {
		if k != "kind" {
			item.Extra[k] = v
		}
	}
	return nil
}

var _ WalkerDecoder = (*wdSliceItem)(nil)

type wdSliceContainer struct {
	Items []wdSliceItem `json:"items" validate:"required"`
}

func TestPlan_WalkerDecoderSliceElement(t *testing.T) {
	// Valid: items decode with custom DecodeWalk, extra fields captured
	out, err := New[wdSliceContainer]().Unmarshal([]byte(`{"items":[{"kind":"a","extra_key":"extra_val"}]}`))
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, "a", out.Items[0].Kind)
	require.Equal(t, "extra_val", out.Items[0].Extra["extra_key"])

	// Invalid: required field missing in slice element
	_, err = New[wdSliceContainer]().Unmarshal([]byte(`{"items":[{"extra_key":"extra_val"}]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Contains(t, ve.Errors[0].Field, "Items[0]")
	require.Contains(t, ve.Errors[0].Field, "Kind")
}

// === Test #7: JSONUnmarshalerLeafField ===
// Tests that json.Unmarshaler leaf fields (like SecretStr) are decoded
// correctly through the Unmarshal boundary.

type configWithSecret struct {
	APIKey SecretStr `json:"api_key" validate:"required"`
	Name   string    `json:"name"`
}

func TestPlan_JSONUnmarshalerLeafField(t *testing.T) {
	// Valid: SecretStr field is decoded correctly via json.Unmarshaler
	out, err := New[configWithSecret]().Unmarshal([]byte(`{"api_key":"s3cr3t","name":"test"}`))
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", out.APIKey.Value())
	require.Equal(t, "test", out.Name)

	// Invalid: required SecretStr field missing at top level
	// Field should be "api_key" (JSON tag name, since it's a direct top-level field)
	_, err = New[configWithSecret]().Unmarshal([]byte(`{"name":"test"}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Equal(t, "api_key", ve.Errors[0].Field)
}

// === Test #8: ExtraAllowTwoLevels ===
// Tests ExtraAllow mode with unknown fields at both top-level and nested levels.
// Both levels should have their Extras maps populated independently.

type nestedWithExtra struct {
	Value  string         `json:"value"`
	Extras map[string]any `json:"-" validate:"extra_fields"`
}

type topWithExtra struct {
	Name   string          `json:"name"`
	Nested nestedWithExtra `json:"nested"`
	Extras map[string]any  `json:"-" validate:"extra_fields"`
}

func TestPlan_ExtraAllowTwoLevels(t *testing.T) {
	vl := New[topWithExtra](Options{
		ExtraFields: ExtraAllow,
	})

	jsonData := []byte(`{
		"name": "test",
		"nested": {
			"value": "nested_val",
			"extra_nested": "nested_extra"
		},
		"extra_top": "top_extra"
	}`)

	obj, err := vl.Unmarshal(jsonData)
	require.NoError(t, err)

	// Check top-level extras captured
	require.NotNil(t, obj.Extras)
	assert.Equal(t, "top_extra", obj.Extras["extra_top"])

	// Check nested-level extras captured
	require.NotNil(t, obj.Nested.Extras)
	assert.Equal(t, "nested_extra", obj.Nested.Extras["extra_nested"])

	// Verify schema fields are intact
	assert.Equal(t, "test", obj.Name)
	assert.Equal(t, "nested_val", obj.Nested.Value)
}

// === Test #9: RecursiveTypeWithinDepthCap ===
// Tests a self-referential type nested 2 levels deep (within default cap of 3).
// Should succeed without exceeding recursion depth.

type nodeRecursive struct {
	Name  string         `json:"name" validate:"required"`
	Child *nodeRecursive `json:"child,omitempty"`
}

func TestPlan_RecursiveTypeWithinDepthCap(t *testing.T) {
	// Depth 2 (root -> child): within default cap of 3
	jsonData := []byte(`{
		"name": "root",
		"child": {
			"name": "child1"
		}
	}`)

	out, err := New[nodeRecursive]().Unmarshal(jsonData)
	require.NoError(t, err)
	require.Equal(t, "root", out.Name)
	require.NotNil(t, out.Child)
	require.Equal(t, "child1", out.Child.Name)
	require.Nil(t, out.Child.Child)
}

// === Test #10: ExactErrorTextForMissingNestedRequired ===
// Tests that error message and field path are exactly as expected
// for a missing nested required field.

type innerExact struct {
	Title string `json:"title" validate:"required"`
}

type outerExact struct {
	Inner innerExact `json:"inner"`
}

func TestPlan_ExactErrorTextForMissingNestedRequired(t *testing.T) {
	_, err := New[outerExact]().Unmarshal([]byte(`{"inner":{}}`))
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)

	// Check exact Field (dotted Go path for nested field)
	require.Equal(t, "Inner.Title", ve.Errors[0].Field)

	// Check exact Message (should be exactly "is required")
	require.Equal(t, "is required", ve.Errors[0].Message)
}

// === Test #11: PointerToSliceAndMapFields ===
// Tests that *[]T and *map[K]V fields (pointer-to-slice, pointer-to-map) are
// classified and decoded correctly, not misclassified as KindScalar.

type ptrCollections struct {
	Tags   *[]string               `json:"tags"`
	Scores *map[string]int         `json:"scores"`
	Items  *[]itemInSlice          `json:"items" validate:"dive"`
	Lookup *map[string]itemInSlice `json:"lookup" validate:"dive"`
}

func TestPlan_PointerToSliceAndMapFields(t *testing.T) {
	jsonData := []byte(`{
		"tags": ["a", "b", "c"],
		"scores": {"x": 1, "y": 2},
		"items": [{"name": "item1"}],
		"lookup": {"k": {"name": "item2"}}
	}`)

	out, err := New[ptrCollections]().Unmarshal(jsonData)
	require.NoError(t, err)

	require.NotNil(t, out.Tags)
	require.Equal(t, []string{"a", "b", "c"}, *out.Tags)

	require.NotNil(t, out.Scores)
	require.Equal(t, map[string]int{"x": 1, "y": 2}, *out.Scores)

	require.NotNil(t, out.Items)
	require.Len(t, *out.Items, 1)
	require.Equal(t, "item1", (*out.Items)[0].Name)

	require.NotNil(t, out.Lookup)
	require.Equal(t, "item2", (*out.Lookup)["k"].Name)

	// Absent keys leave the pointers nil, not a decode error.
	out2, err := New[ptrCollections]().Unmarshal([]byte(`{}`))
	require.NoError(t, err)
	require.Nil(t, out2.Tags)
	require.Nil(t, out2.Scores)
	require.Nil(t, out2.Items)
	require.Nil(t, out2.Lookup)
}

// === Test #12: WalkerDecoderSliceAggregatesRequiredErrors ===
// Tests that when a WalkerDecoder's recurse callback is used to decode a
// slice of structs, required-field errors from MULTIPLE elements are
// aggregated (via setValueForType's slice/map branches), not just the first.

type wdAggBlock struct {
	Name string `json:"name" validate:"required"`
}

type wdAggContent struct {
	Blocks []wdAggBlock `json:"-"`
}

func (c *wdAggContent) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	m, ok := decoded.(map[string]any)
	if !ok {
		return fmt.Errorf("wdAggContent must be object, got %T", decoded)
	}
	raw, ok := m["blocks"]
	if !ok {
		return nil
	}
	return recurse(&c.Blocks, raw)
}

var _ WalkerDecoder = (*wdAggContent)(nil)

type wdAggContentContainer struct {
	Content wdAggContent `json:"content"`
}

func TestPlan_WalkerDecoderSliceAggregatesRequiredErrors(t *testing.T) {
	_, err := New[wdAggContentContainer]().Unmarshal([]byte(`{
		"content": {
			"blocks": [{}, {}]
		}
	}`))
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)

	nameErrs := 0
	for _, fieldErr := range ve.Errors {
		if strings.Contains(fieldErr.Field, "Name") {
			nameErrs++
		}
	}
	assert.Equal(t, 2, nameErrs, "expected both invalid blocks' required errors to be aggregated, got: %+v", ve.Errors)
}

// === Test #13: WalkerDecoderRecurseTargets ===
// Tests setValueForType's dispatch branches (scalar leaf, map-of-struct with
// required-error aggregation and shape-mismatch, json.Unmarshaler leaf, and
// nil handling) via a WalkerDecoder whose recurse callback targets each kind
// directly - these are reachable only through WalkerDecoder.recurse, not
// through the normal struct-field decode path.

type wdTargetItem struct {
	Name string `json:"name" validate:"required"`
}

type wdTargetHolder struct {
	Scalar  int
	MapVals map[string]wdTargetItem
	Secret  SecretStr
}

func (h *wdTargetHolder) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	m, ok := decoded.(map[string]any)
	if !ok {
		return fmt.Errorf("wdTargetHolder must be object, got %T", decoded)
	}
	if v, present := m["scalar"]; present {
		if err := recurse(&h.Scalar, v); err != nil {
			return err
		}
	}
	if v, present := m["map_vals"]; present {
		if err := recurse(&h.MapVals, v); err != nil {
			return err
		}
	}
	if v, present := m["secret"]; present {
		if err := recurse(&h.Secret, v); err != nil {
			return err
		}
	}
	return nil
}

var _ WalkerDecoder = (*wdTargetHolder)(nil)

type wdTargetContainer struct {
	Holder wdTargetHolder `json:"holder"`
}

func TestPlan_WalkerDecoderRecurseTargets_Valid(t *testing.T) {
	out, err := New[wdTargetContainer]().Unmarshal([]byte(`{
		"holder": {
			"scalar": 42,
			"map_vals": {"a": {"name": "alpha"}},
			"secret": "s3cr3t"
		}
	}`))
	require.NoError(t, err)
	assert.Equal(t, 42, out.Holder.Scalar)
	require.Len(t, out.Holder.MapVals, 1)
	assert.Equal(t, "alpha", out.Holder.MapVals["a"].Name)
	assert.Equal(t, "s3cr3t", out.Holder.Secret.Value())
}

func TestPlan_WalkerDecoderRecurseTargets_MapAggregatesRequiredErrors(t *testing.T) {
	_, err := New[wdTargetContainer]().Unmarshal([]byte(`{
		"holder": {
			"map_vals": {"a": {}, "b": {}}
		}
	}`))
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)

	nameErrs := 0
	for _, fieldErr := range ve.Errors {
		if strings.Contains(fieldErr.Field, "Name") {
			nameErrs++
		}
	}
	assert.Equal(t, 2, nameErrs, "expected both invalid map values' required errors to be aggregated, got: %+v", ve.Errors)
}

func TestPlan_WalkerDecoderRecurseTargets_MapShapeMismatch(t *testing.T) {
	_, err := New[wdTargetContainer]().Unmarshal([]byte(`{
		"holder": {
			"map_vals": "not-an-object"
		}
	}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected object")
}
