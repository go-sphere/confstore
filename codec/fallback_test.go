package codec

import (
	"errors"
	"testing"
)

// testCodec is a simple Codec implementation for testing.
type testCodec struct {
	marshal   func(v any) ([]byte, error)
	unmarshal func(data []byte, v any) error
}

func (t testCodec) Marshal(v any) ([]byte, error)   { return t.marshal(v) }
func (t testCodec) Unmarshal(d []byte, v any) error { return t.unmarshal(d, v) }

func TestFallbackUnmarshalSuccess(t *testing.T) {
	c1 := testCodec{
		marshal:   func(v any) ([]byte, error) { return nil, errors.New("nope1") },
		unmarshal: func(data []byte, v any) error { return errors.New("nope1") },
	}
	c2 := testCodec{
		marshal:   func(v any) ([]byte, error) { return []byte("ok"), nil },
		unmarshal: func(data []byte, v any) error { return nil },
	}
	g := NewCodecGroup(c1, c2)
	var out any
	if err := g.Unmarshal([]byte("{}"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFallbackUnmarshalFailure(t *testing.T) {
	cErr := errors.New("bad")
	c1 := testCodec{
		marshal:   func(v any) ([]byte, error) { return nil, cErr },
		unmarshal: func(data []byte, v any) error { return cErr },
	}
	g := NewCodecGroup(c1)
	var out any
	if err := g.Unmarshal([]byte("{}"), &out); err == nil {
		t.Fatal("expected error, got nil")
	}
}
