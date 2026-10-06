package codec

import (
	"encoding/json"
	"errors"
)

type codec struct {
	encoder EncoderFunc
	decoder DecoderFunc
}

// NewCodec returns a [Codec] whose Marshal calls encoder and whose Unmarshal
// calls decoder. Results and errors are passed through unchanged. Both
// functions must be non-nil; a nil function panics when its method is called.
//
//	c := codec.NewCodec(json.Marshal, json.Unmarshal)
func NewCodec(encoder EncoderFunc, decoder DecoderFunc) Codec {
	return &codec{
		encoder: encoder,
		decoder: decoder,
	}
}

func (c *codec) Marshal(val any) ([]byte, error) {
	return c.encoder(val)
}

func (c *codec) Unmarshal(data []byte, val any) error {
	return c.decoder(data, val)
}

var (
	// ErrInvalidType is returned when a value's type is not supported by the
	// codec operation, such as a non-*string target for [StringCodec] or a
	// non-pointer target for [FallbackCodecGroup.Unmarshal].
	ErrInvalidType = errors.New("invalid type for codec operation")
	// ErrNilPointer is returned when a nil pointer (or nil interface) is passed
	// where a value is required, for both marshal and unmarshal operations.
	ErrNilPointer = errors.New("nil pointer cannot be marshaled")
)

// JsonCodec returns a [Codec] backed by [encoding/json.Marshal] and
// [encoding/json.Unmarshal]. Unmarshal follows encoding/json rules: the target
// must be a non-nil pointer, unknown fields are ignored, and fields absent from
// the input keep their existing values. Errors from encoding/json are returned
// unchanged.
func JsonCodec() Codec {
	return &codec{
		encoder: json.Marshal,
		decoder: json.Unmarshal,
	}
}

// StringCodec returns a [Codec] that copies raw bytes to and from strings
// without any parsing or transformation.
//
// Marshal accepts a string or a *string. Unmarshal requires a *string target
// and overwrites it with the data. A nil *string yields [ErrNilPointer]; any
// other type yields [ErrInvalidType]. Both are reported at run time because the
// Codec interface takes any.
func StringCodec() Codec {
	return &codec{
		encoder: func(val any) ([]byte, error) {
			if str, ok := val.(string); ok {
				return []byte(str), nil
			}
			if strPtr, ok := val.(*string); ok {
				if strPtr == nil {
					return nil, ErrNilPointer
				}
				return []byte(*strPtr), nil
			}
			return nil, ErrInvalidType
		},
		decoder: func(data []byte, val any) error {
			if strPtr, ok := val.(*string); ok {
				if strPtr == nil {
					return ErrNilPointer
				}
				*strPtr = string(data)
				return nil
			}
			return ErrInvalidType
		},
	}
}
