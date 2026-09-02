package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- string|array WalkerDecoder over a plain-struct element slice ---

type wdBlock struct {
	Type string `json:"type" validate:"required"`
	Text string `json:"text"`
}

type wdContent struct {
	Text   *string
	Blocks []wdBlock
}

func (c *wdContent) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	switch v := decoded.(type) {
	case string:
		c.Text = &v
		c.Blocks = nil
		return nil
	case []any:
		c.Text = nil
		return recurse(&c.Blocks, v)
	default:
		return fmt.Errorf("wdContent must be string or array, got %T", decoded)
	}
}

var _ WalkerDecoder = (*wdContent)(nil)

type wdMessage struct {
	Role    string    `json:"role" validate:"required"`
	Content wdContent `json:"content" validate:"required"`
}

func TestWalkerDecoder_StringVariant(t *testing.T) {
	out, err := New[wdMessage]().Unmarshal([]byte(`{"role":"user","content":"hello"}`))
	require.NoError(t, err)
	require.NotNil(t, out.Content.Text)
	require.Equal(t, "hello", *out.Content.Text)
	require.Nil(t, out.Content.Blocks)
}

func TestWalkerDecoder_ArrayVariant(t *testing.T) {
	out, err := New[wdMessage]().Unmarshal([]byte(`{"role":"user","content":[{"type":"text","text":"hi"}]}`))
	require.NoError(t, err)
	require.Nil(t, out.Content.Text)
	require.Len(t, out.Content.Blocks, 1)
	require.Equal(t, "text", out.Content.Blocks[0].Type)
	require.Equal(t, "hi", out.Content.Blocks[0].Text)
}

func TestWalkerDecoder_NestedRequiredEnforced(t *testing.T) {
	// Block missing its required "type" must fail Unmarshal.
	_, err := New[wdMessage]().Unmarshal([]byte(`{"role":"user","content":[{"text":"hi"}]}`))
	require.Error(t, err)
}

// --- WalkerDecoder as a SLICE ELEMENT (covers the setSliceField route) ---

type wdItem struct {
	Kind  string `json:"kind" validate:"required"`
	Extra map[string]any
}

func (it *wdItem) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	m, ok := decoded.(map[string]any)
	if !ok {
		return fmt.Errorf("wdItem must be object, got %T", decoded)
	}
	type shadow struct {
		Kind string `json:"kind" validate:"required"`
	}
	var s shadow
	if err := recurse(&s, decoded); err != nil {
		return err
	}
	it.Kind = s.Kind
	it.Extra = map[string]any{}
	for k, v := range m {
		if k != "kind" {
			it.Extra[k] = v
		}
	}
	return nil
}

var _ WalkerDecoder = (*wdItem)(nil)

type wdItemList struct {
	Items []wdItem `json:"items" validate:"required"`
}

func TestWalkerDecoder_SliceElementDecodeWalkFires(t *testing.T) {
	out, err := New[wdItemList]().Unmarshal([]byte(`{"items":[{"kind":"a","foo":1}]}`))
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, "a", out.Items[0].Kind)
	require.InDelta(t, float64(1), out.Items[0].Extra["foo"], 0)
}

func TestWalkerDecoder_SliceElementNestedRequiredEnforced(t *testing.T) {
	_, err := New[wdItemList]().Unmarshal([]byte(`{"items":[{"foo":1}]}`))
	require.Error(t, err)
}

func TestWalkerDecoder_SliceElementAggregatesAllErrors(t *testing.T) {
	_, err := New[wdItemList]().Unmarshal([]byte(`{"items":[{"foo":1},{"bar":2}]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 2)
	require.Contains(t, ve.Errors[0].Field, "Items[0]")
	require.Contains(t, ve.Errors[0].Field, "Kind")
	require.Contains(t, ve.Errors[1].Field, "Items[1]")
	require.Contains(t, ve.Errors[1].Field, "Kind")
	require.NotEqual(t, ve.Errors[0].Field, ve.Errors[1].Field)
}

func TestWalkerDecoder_SliceElementNullDoesNotPanic(t *testing.T) {
	_, err := New[wdItemList]().Unmarshal([]byte(`{"items":[null]}`))
	require.Error(t, err)
}

// --- Deep nesting: a WalkerDecoder struct (reached as a slice element) with a
// pointer field to another WalkerDecoder struct, mirroring the gateway's
// ContentBlock.Source *ImageSource shape. ---

type wdLeaf struct {
	Kind string
}

type wdLeafShadow struct {
	Kind string `json:"kind" validate:"required"`
}

func (l *wdLeaf) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	var s wdLeafShadow
	if err := recurse(&s, decoded); err != nil {
		return err
	}
	l.Kind = s.Kind
	return nil
}

var _ WalkerDecoder = (*wdLeaf)(nil)

type wdBranch struct {
	Name string
	Leaf *wdLeaf
}

type wdBranchShadow struct {
	Name string  `json:"name" validate:"required"`
	Leaf *wdLeaf `json:"leaf,omitempty"`
}

func (b *wdBranch) DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error {
	var s wdBranchShadow
	if err := recurse(&s, decoded); err != nil {
		return err
	}
	b.Name = s.Name
	b.Leaf = s.Leaf
	return nil
}

var _ WalkerDecoder = (*wdBranch)(nil)

type wdBranchList struct {
	Branches []wdBranch `json:"branches" validate:"required"`
}

func TestWalkerDecoder_DeepNestedPointerFieldFullyQualifiedPath(t *testing.T) {
	// Branches[0].Leaf is present but missing its required "kind" — the field
	// path must include every intermediate segment (Branches[0].Leaf.Kind),
	// not just the outermost slice index.
	_, err := New[wdBranchList]().Unmarshal([]byte(`{"branches":[{"name":"b1","leaf":{}}]}`))
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	require.Len(t, ve.Errors, 1)
	require.Equal(t, "Branches[0].Leaf.Kind", ve.Errors[0].Field)
}

func TestWalkerDecoder_DeepNestedPointerFieldValid(t *testing.T) {
	out, err := New[wdBranchList]().Unmarshal([]byte(`{"branches":[{"name":"b1","leaf":{"kind":"k1"}}]}`))
	require.NoError(t, err)
	require.Len(t, out.Branches, 1)
	require.Equal(t, "b1", out.Branches[0].Name)
	require.NotNil(t, out.Branches[0].Leaf)
	require.Equal(t, "k1", out.Branches[0].Leaf.Kind)
}

// --- json.Unmarshaler leaf fallback fixes SecretStr through the walker ---

func TestWalkerDecoder_JSONUnmarshalerLeafFallback(t *testing.T) {
	type cfg struct {
		APIKey SecretStr `json:"api_key" validate:"required"`
	}
	out, err := New[cfg]().Unmarshal([]byte(`{"api_key":"s3cr3t"}`))
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", out.APIKey.Value())
}
