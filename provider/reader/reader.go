package reader

import (
	"context"
	"io"
)

// Reader is a provider that reads all configuration bytes
// from an underlying io.Reader.
type Reader struct {
	reader io.Reader
}

// NewReader creates a new Reader that wraps the provided io.Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{reader: r}
}

// Read implements provider.Provider by returning all bytes
// from the underlying io.Reader. The context is accepted for
// interface compatibility and is not used for cancellation here.
func (r *Reader) Read(ctx context.Context) ([]byte, error) {
	return io.ReadAll(r.reader)
}

// Bytes is a provider that returns a fixed byte slice.
type Bytes struct {
	data []byte
}

// NewBytes creates a Bytes provider that always returns the
// provided byte slice.
func NewBytes(data []byte) *Bytes {
	return &Bytes{data: data}
}

func (b *Bytes) Read(ctx context.Context) ([]byte, error) {
	return b.data, nil
}
