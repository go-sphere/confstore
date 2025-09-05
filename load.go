package confstore

import (
	"context"

	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
)

// Load reads configuration from the given provider and unmarshal it into the provided struct.
func Load[T any](ctx context.Context, provider provider.Provider, codec codec.Codec) (*T, error) {
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
