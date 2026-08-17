// Package pedantigogin installs Pedantigo into Gin's request binding pipeline.
//
// Gin splits request binding across two seams:
//   - JSON body decoding goes through codec/json.API.
//   - Query/form/header/URI validation goes through binding.Validator.
//
// Replacing only one of those seams would leave part of Gin's request pipeline
// outside Pedantigo. NewBinder wires both hooks at setup time.
package pedantigogin

import (
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin/binding"
	codecjson "github.com/gin-gonic/gin/codec/json"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

const defaultTagName = "binding"

// Option customizes binder setup.
type Option func(*config)

type config struct {
	tagName string
}

// WithTagName changes the struct tag namespace Pedantigo expects Gin-visible
// registered validators to use. The default is "binding".
func WithTagName(name string) Option {
	return func(cfg *config) {
		cfg.tagName = name
	}
}

// NewBinder installs Pedantigo into Gin's JSON decoder and validation hooks.
//
// It panics at setup time if Gin's global JSON decoder flags were already
// enabled, because those options cannot be forwarded through Pedantigo's
// runtime-type decoding path. It also panics if the requested tag name is
// incompatible with already-registered validators.
func NewBinder(opts ...Option) {
	cfg := config{tagName: defaultTagName}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.tagName == "" {
		cfg.tagName = defaultTagName
	}

	if binding.EnableDecoderUseNumber {
		panic("pedantigo gin: Gin's EnableDecoderUseNumber is incompatible with pedantigo's JSON decoder override")
	}
	if binding.EnableDecoderDisallowUnknownFields {
		panic("pedantigo gin: Gin's EnableDecoderDisallowUnknownFields is incompatible with pedantigo's JSON decoder override")
	}

	validator.RequireSingleRegisteredTagName(cfg.tagName)

	codecjson.API = &pedantigoCodec{fallback: codecjson.API}
	binding.Validator = &pedantigoValidator{}
}

// AsValidationError extracts a Pedantigo validation error from a Gin bind
// failure.
func AsValidationError(err error) (*validator.ValidationError, bool) {
	if err == nil {
		return nil, false
	}
	var ve *validator.ValidationError
	if !errors.As(err, &ve) {
		return nil, false
	}
	return ve, true
}

type pedantigoCodec struct {
	fallback codecjson.Core
}

func (c *pedantigoCodec) Marshal(v any) ([]byte, error) {
	return c.fallback.Marshal(v)
}

func (c *pedantigoCodec) Unmarshal(data []byte, v any) error {
	return c.fallback.Unmarshal(data, v)
}

func (c *pedantigoCodec) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return c.fallback.MarshalIndent(v, prefix, indent)
}

func (c *pedantigoCodec) NewEncoder(writer io.Writer) codecjson.Encoder {
	return c.fallback.NewEncoder(writer)
}

func (c *pedantigoCodec) NewDecoder(reader io.Reader) codecjson.Decoder {
	return &pedantigoDecoder{reader: reader}
}

type pedantigoDecoder struct {
	reader io.Reader
}

func (d *pedantigoDecoder) UseNumber() {}

func (d *pedantigoDecoder) DisallowUnknownFields() {}

func (d *pedantigoDecoder) Decode(v any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pedantigo gin: recovered panic during decode into %T: %v", v, r)
		}
	}()

	body, err := io.ReadAll(d.reader)
	if err != nil {
		return err
	}
	return validator.UnmarshalInto(body, v)
}

type pedantigoValidator struct{}

func (p *pedantigoValidator) ValidateStruct(obj any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pedantigo gin: recovered panic during validation of %T: %v", obj, r)
		}
	}()
	return validator.ValidateInto(obj)
}

func (p *pedantigoValidator) Engine() any {
	return nil
}
