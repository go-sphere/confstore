package confstore

import (
	"context"

	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
)

// LoadWithContext reads configuration from the given provider and unmarshal it into the provided struct with context.
func LoadWithContext[T any](ctx context.Context, provider provider.Provider, codec codec.Codec) (*T, error) {
	data, err := provider.Read(ctx)
	if err != nil {
		return nil, err
	}
	var config T
	err = codec.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Load reads configuration from the given provider and unmarshal it into the provided struct.
func Load[T any](provider provider.Provider, codec codec.Codec) (*T, error) {
	return LoadWithContext[T](context.Background(), provider, codec)
}

// FillWithContext reads configuration from the given provider and unmarshal it into the provided struct with context.
func FillWithContext[T any](ctx context.Context, provider provider.Provider, codec codec.Codec, config *T) error {
	data, err := provider.Read(ctx)
	if err != nil {
		return err
	}
	return codec.Unmarshal(data, config)
}

// Fill reads configuration from the given provider and unmarshal it into the provided struct.
func Fill[T any](provider provider.Provider, codec codec.Codec, config *T) error {
	return FillWithContext[T](context.Background(), provider, codec, config)
}
