package reader

import (
	"context"
	"io"
)

// Reader is a provider.Provider that reads all configuration bytes
// from an underlying [io.Reader].
//
// The io.Reader is consumed by the first Read; a later Read returns only
// what the io.Reader yields afterwards (typically nothing). Use [Bytes] when
// the same content must be returned repeatedly. Reader is not safe for
// concurrent use unless the wrapped io.Reader is. Create it with [NewReader].
type Reader struct {
	reader io.Reader
}

// NewReader returns a [Reader] that wraps r, which must be non-nil. The
// caller keeps ownership of r and is responsible for closing it, if needed,
// after reading.
func NewReader(r io.Reader) *Reader {
	return &Reader{reader: r}
}

// Read implements provider.Provider by returning all remaining bytes
// from the underlying io.Reader, as by [io.ReadAll]. Errors from the
// io.Reader are returned unchanged (io.EOF is not an error). The context is
// accepted for interface compatibility and is not used for cancellation here.
func (r *Reader) Read(ctx context.Context) ([]byte, error) {
	return io.ReadAll(r.reader)
}

// Bytes is a provider.Provider that returns the same fixed bytes on every
// Read. It is immutable after construction and safe for concurrent use.
// Create it with [NewBytes]; the zero value returns empty data.
type Bytes struct {
	data []byte
}

// NewBytes creates a Bytes provider that always returns the
// provided byte slice. The input is cloned at construction so the
// provider is decoupled from the caller's backing slice.
func NewBytes(data []byte) *Bytes {
	return &Bytes{data: append([]byte(nil), data...)}
}

// Read returns a fresh copy of the configured bytes and a nil error, so
// callers may modify the result freely. The context is not used.
func (b *Bytes) Read(ctx context.Context) ([]byte, error) {
	return append([]byte(nil), b.data...), nil
}
