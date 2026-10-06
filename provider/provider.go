package provider

import (
	"context"
)

// Provider is a source of raw configuration bytes, such as a file or an HTTP
// endpoint. Decoding the bytes is left to a codec.
type Provider interface {
	// Read returns the entire configuration as raw bytes. Implementations
	// should honor ctx for cancellation and deadlines where the underlying
	// source allows it, and return a non-nil error when the bytes could not
	// be obtained. Each call reads the source again; nothing is cached.
	Read(ctx context.Context) ([]byte, error)
}

// ReaderFunc adapts an ordinary function to the [Provider] interface.
//
//	p := provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
//		return os.ReadFile("config.json")
//	})
type ReaderFunc func(ctx context.Context) ([]byte, error)

// Read calls f(ctx) and returns its result unchanged.
func (f ReaderFunc) Read(ctx context.Context) ([]byte, error) {
	return f(ctx)
}
