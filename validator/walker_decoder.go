package validator

import "github.com/SmrutAI/pedantigo/v2/validator/internal/deserialize"

// WalkerDecoder is implemented by a type that decodes itself from an
// already-decoded JSON value during Unmarshal, while pedantigo still enforces
// required/default/constraints on any nested structs the type delegates back to
// the walker via the recurse callback. See internal/deserialize.WalkerDecoder.
type WalkerDecoder = deserialize.WalkerDecoder
