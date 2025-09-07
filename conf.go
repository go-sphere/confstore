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

// LoadFromFunc creates a provider using the given factory function, then reads configuration from it
func LoadFromFunc[T any](factory func() (provider.Provider, error), codec codec.Codec) (*T, error) {
	p, err := factory()
	if err != nil {
		return nil, err
	}
	return Load[T](p, codec)
}

// Load reads configuration from the given provider and unmarshal it into the provided struct.
func Load[T any](provider provider.Provider, codec codec.Codec) (*T, error) {
	return LoadWithContext[T](context.Background(), provider, codec)
}

// FillWithContext reads configuration from the given provider and unmarshal it into the provided struct with context.
func FillWithContext(ctx context.Context, provider provider.Provider, codec codec.Codec, config any) error {
	data, err := provider.Read(ctx)
	if err != nil {
		return err
	}
	return codec.Unmarshal(data, config)
}

// FillFromFunc creates a provider using the given factory function, then reads configuration from it
func FillFromFunc(factory func() (provider.Provider, error), codec codec.Codec, config any) error {
	p, err := factory()
	if err != nil {
		return err
	}
	return Fill(p, codec, config)
}

// Fill reads configuration from the given provider and unmarshal it into the provided struct.
func Fill(provider provider.Provider, codec codec.Codec, config any) error {
	return FillWithContext(context.Background(), provider, codec, config)
}
