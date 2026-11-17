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

func TestFallbackCodecGroupNilCodecs(t *testing.T) {
	t.Run("only-nil", func(t *testing.T) {
		g := NewCodecGroup(nil)
		if _, err := g.Marshal("data"); err == nil {
			t.Fatal("expected Marshal error for nil codec, got nil")
		}
		var out string
		if err := g.Unmarshal([]byte("data"), &out); err == nil {
			t.Fatal("expected Unmarshal error for nil codec, got nil")
		}
	})

	t.Run("nil-after-failing", func(t *testing.T) {
		boom := errors.New("boom")
		failing := testCodec{
			marshal:   func(v any) ([]byte, error) { return nil, boom },
			unmarshal: func(data []byte, v any) error { return boom },
		}
		g := NewCodecGroup(failing, nil)
		if _, err := g.Marshal("data"); !errors.Is(err, boom) {
			t.Fatalf("expected error wrapping boom, got %v", err)
		}
		var out string
		if err := g.Unmarshal([]byte("data"), &out); !errors.Is(err, boom) {
			t.Fatalf("expected error wrapping boom, got %v", err)
		}
	})

	t.Run("nil-then-valid-fallback", func(t *testing.T) {
		g := NewCodecGroup(nil, JsonCodec())
		b, err := g.Marshal(map[string]string{"k": "v"})
		if err != nil {
			t.Fatalf("unexpected Marshal error: %v", err)
		}
		var out map[string]string
		if err := g.Unmarshal(b, &out); err != nil {
			t.Fatalf("unexpected Unmarshal error: %v", err)
		}
		if out["k"] != "v" {
			t.Fatalf("got %v, want k=v", out)
		}
	})
}

func TestFallbackCodecGroupClonesSlice(t *testing.T) {
	codecs := []Codec{JsonCodec()}
	g := NewCodecGroup(codecs...)
	codecs[0] = nil

	b, err := g.Marshal(map[string]int{"port": 8080})
	if err != nil {
		t.Fatalf("Marshal failed after mutating input slice: %v", err)
	}
	var out map[string]int
	if err := g.Unmarshal(b, &out); err != nil {
		t.Fatalf("Unmarshal failed after mutating input slice: %v", err)
	}
	if out["port"] != 8080 {
		t.Fatalf("got %v, want port=8080", out)
	}
}
