package codec

import (
	"errors"
	"fmt"
	"reflect"
)

// FallbackCodecGroup implements a fallback mechanism for multiple codecs.
// It tries each codec in order until one succeeds for both marshal and unmarshal operations.
type FallbackCodecGroup struct {
	codecs []Codec
}

// NewCodecGroup creates a new FallbackCodecGroup with the provided codecs.
// The codecs will be tried in the order they are provided.
func NewCodecGroup(codecs ...Codec) *FallbackCodecGroup {
	return &FallbackCodecGroup{codecs: codecs}
}

// Marshal attempts to marshal the value using each codec in order until one succeeds.
// Returns the marshaled data from the first successful codec, or an error if all codecs fail.
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

// Unmarshal attempts to unmarshal the data using each codec in order until one succeeds.
// Returns nil on the first successful unmarshal, or an error if all codecs fail.
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
