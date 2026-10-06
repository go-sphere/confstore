package codec

import (
	"errors"
	"fmt"
	"reflect"
)

// FallbackCodecGroup is a [Codec] that tries a list of codecs in order and
// uses the result of the first one that succeeds. Use it when configuration
// may arrive in one of several formats.
//
// Create one with [NewCodecGroup]. The zero value has no codecs and every
// call returns an error. The codec list is fixed after construction, so a
// group is safe for concurrent use if its codecs are.
type FallbackCodecGroup struct {
	codecs []Codec
}

// NewCodecGroup returns a [FallbackCodecGroup] that tries codecs in the
// given order. The codecs slice is retained, not copied. Calling it with no
// codecs is allowed, but the resulting group fails every operation.
func NewCodecGroup(codecs ...Codec) *FallbackCodecGroup {
	return &FallbackCodecGroup{codecs: codecs}
}

// Marshal encodes value with each codec in order and returns the bytes from
// the first one that succeeds. If the group has no codecs it returns an error.
// If every codec fails it returns an error that joins each codec's error,
// prefixed with its index, so errors.Is matches any of them.
func (m *FallbackCodecGroup) Marshal(value any) ([]byte, error) {
	if len(m.codecs) == 0 {
		return nil, errors.New("fallback marshal: no codecs configured")
	}
	var joined error
	for i, c := range m.codecs {
		data, err := c.Marshal(value)
		if err == nil {
			return data, nil
		}
		joined = errors.Join(joined, fmt.Errorf("codec[%d]: %w", i, err))
	}
	return nil, fmt.Errorf("fallback marshal failed: %w", joined)
}

// Unmarshal decodes data into value with each codec in order and stops at the
// first one that succeeds.
//
// value must be a non-nil pointer: a nil value or nil pointer returns
// [ErrNilPointer], and a non-pointer returns [ErrInvalidType]. Each attempt
// decodes into a fresh zero value of the pointed-to type, and only a
// successful result is copied into value. As a result, value is left unchanged
// on failure, and on success it is fully replaced (existing field values are
// not preserved as defaults). If the group has no codecs it returns an error.
// If every codec fails it returns an error joining each codec's error, so
// errors.Is matches any of them.
func (m *FallbackCodecGroup) Unmarshal(data []byte, value any) error {
	if len(m.codecs) == 0 {
		return errors.New("fallback unmarshal: no codecs configured")
	}
	if value == nil {
		return ErrNilPointer
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer {
		return ErrInvalidType
	}
	if rv.IsNil() {
		return ErrNilPointer
	}
	var joined error
	for i, c := range m.codecs {
		// Decode into a temporary value to avoid partial writes.
		tmp := reflect.New(rv.Elem().Type())
		if err := c.Unmarshal(data, tmp.Interface()); err == nil {
			rv.Elem().Set(tmp.Elem())
			return nil
		} else {
			joined = errors.Join(joined, fmt.Errorf("codec[%d]: %w", i, err))
		}
	}
	return fmt.Errorf("fallback unmarshal failed: %w", joined)
}
