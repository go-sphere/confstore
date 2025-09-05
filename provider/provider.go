package provider

import "context"

// Provider represents a configuration provider. Providers can
// read configuration from a source (file, HTTP, etc.)
type Provider interface {
	// Read returns the entire configuration as raw []bytes to be parsed.
	// The provided context controls cancellation and deadlines.
	Read(ctx context.Context) ([]byte, error)
}
