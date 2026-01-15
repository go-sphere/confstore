package codec

import (
	"errors"
	"testing"
)

type simpleCodec struct {
	marshal   func(v any) ([]byte, error)
	unmarshal func(data []byte, v any) error
}

func (s simpleCodec) Marshal(v any) ([]byte, error)   { return s.marshal(v) }
func (s simpleCodec) Unmarshal(d []byte, v any) error { return s.unmarshal(d, v) }

func TestStringCodecCases(t *testing.T) {
	c := StringCodec()
	value := "hello"

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr error
	}{
		{name: "string", input: "plain", want: "plain"},
		{name: "string-pointer", input: &value, want: "hello"},
		{name: "invalid-type", input: 42, wantErr: ErrInvalidType},
		{name: "nil-pointer", input: (*string)(nil), wantErr: ErrNilPointer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Marshal(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestStringCodecUnmarshalCases(t *testing.T) {
	c := StringCodec()

	tests := []struct {
		name    string
		target  any
		want    string
		wantErr error
	}{
		{name: "success", target: new(string), want: "data"},
		{name: "invalid-type", target: new(int), wantErr: ErrInvalidType},
		{name: "nil-pointer", target: (*string)(nil), wantErr: ErrNilPointer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Unmarshal([]byte("data"), tt.target)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := *(tt.target.(*string)); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFallbackCodecGroupNoCodecs(t *testing.T) {
	g := NewCodecGroup()
	if _, err := g.Marshal("data"); err == nil {
		t.Fatal("expected error, got nil")
	}
	var out string
	if err := g.Unmarshal([]byte("data"), &out); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFallbackCodecGroupMarshalUsesSecond(t *testing.T) {
	firstErr := errors.New("nope")
	g := NewCodecGroup(
		simpleCodec{marshal: func(v any) ([]byte, error) { return nil, firstErr }, unmarshal: func(data []byte, v any) error { return firstErr }},
		simpleCodec{marshal: func(v any) ([]byte, error) { return []byte("ok"), nil }, unmarshal: func(data []byte, v any) error { return nil }},
	)
	got, err := g.Marshal("data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("got %q, want %q", string(got), "ok")
	}
}

func TestFallbackCodecGroupUnmarshalTypeChecks(t *testing.T) {
	g := NewCodecGroup(simpleCodec{marshal: func(v any) ([]byte, error) { return []byte("x"), nil }, unmarshal: func(data []byte, v any) error { return nil }})
	if err := g.Unmarshal([]byte("x"), "not-pointer"); err == nil {
		t.Fatal("expected error, got nil")
	}
	var sp *string
	if err := g.Unmarshal([]byte("x"), sp); err == nil {
		t.Fatal("expected error, got nil")
	}
}
