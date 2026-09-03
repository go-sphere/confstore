package reader

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestReaderReadCases(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.Reader
		want    string
		wantErr bool
	}{
		{name: "success", reader: bytes.NewBufferString("data"), want: "data"},
		{name: "error", reader: errorReader{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewReader(tt.reader)
			got, err := p.Read(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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

func TestBytesProvider(t *testing.T) {
	p := NewBytes([]byte("fixed"))
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fixed" {
		t.Fatalf("got %q, want %q", string(got), "fixed")
	}
}

func TestBytesProviderCloning(t *testing.T) {
	src := []byte("original")
	p := NewBytes(src)

	got1, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate the returned slice
	got1[0] = 'X'

	got2, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got2) != "original" {
		t.Fatalf("got2 was mutated: %q, want %q", string(got2), "original")
	}
}
