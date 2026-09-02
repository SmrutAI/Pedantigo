package deserialize

// WalkerDecoder lets a type control how it is populated from an already-decoded
// JSON value during Unmarshal, while pedantigo's required/default/constraint
// enforcement still runs on any nested structs the type contains.
//
// decoded is the value produced by the walker's single top-level json decode:
// one of map[string]any, []any, string, float64, bool, or nil.
//
// For any nested struct or slice, the implementation MUST call
// recurse(dst, decoded) — passing a pointer to the destination field and the
// corresponding decoded sub-value — instead of json.Unmarshal, so that
// required/default and all constraints run at every deeper level.
type WalkerDecoder interface {
	DecodeWalk(decoded any, recurse func(dst any, decoded any) error) error
}
