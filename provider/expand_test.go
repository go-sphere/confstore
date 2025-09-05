package provider

import (
	"context"
	"testing"
)

type fixedProvider struct{ b []byte }

func (f fixedProvider) Read(ctx context.Context) ([]byte, error) { return f.b, nil }

func TestExpandEnv(t *testing.T) {
	t.Setenv("FOO", "BAR")
	raw := []byte("value=${FOO}")
	p := NewExpandEnv(fixedProvider{b: raw})
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "value=BAR" {
		t.Fatalf("got %q, want %q", string(got), "value=BAR")
	}
}

func TestExpandEnv_NoDollarFastPath(t *testing.T) {
	raw := []byte("no-vars")
	p := NewExpandEnv(fixedProvider{b: raw})
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "no-vars" {
		t.Fatalf("got %q, want %q", string(got), "no-vars")
	}
}
