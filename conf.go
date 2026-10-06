package confstore

import (
	"context"

	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
)

// LoadWithContext reads all bytes from provider, passing ctx to its Read
// method, and decodes them with codec into a newly allocated T.
//
// T is the configuration value type (not a pointer); codec.Unmarshal receives
// a *T. On success LoadWithContext returns a non-nil *T owned by the caller.
// If the provider's Read or the codec's Unmarshal fails, it returns a nil *T
// and that error unchanged. Cancellation behavior depends on the provider.
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

// Load is [LoadWithContext] with [context.Background]: it reads all bytes from
// provider and decodes them with codec into a newly allocated T.
//
// On success Load returns a non-nil *T. On failure it returns a nil *T and the
// provider or codec error unchanged.
//
//	cfg, err := confstore.Load[AppConf](file.New("config.json"), codec.JsonCodec())
func Load[T any](provider provider.Provider, codec codec.Codec) (*T, error) {
	return LoadWithContext[T](context.Background(), provider, codec)
}

// FillWithContext reads all bytes from provider, passing ctx to its Read
// method, and decodes them with codec into config.
//
// config is passed to codec.Unmarshal as-is, so it is normally a non-nil
// pointer such as *AppConf; JsonCodec rejects non-pointer targets. Fields
// already set in config and absent from the data keep their values, which
// makes Fill suitable for applying configuration over defaults (the exact
// merge semantics are those of the codec). If the provider fails, config is
// not touched. Errors from the provider or codec are returned unchanged; on a
// codec error config may be partially written, depending on the codec.
func FillWithContext[T any](ctx context.Context, provider provider.Provider, codec codec.Codec, config T) error {
	data, err := provider.Read(ctx)
	if err != nil {
		return err
	}
	return codec.Unmarshal(data, config)
}

// Fill is [FillWithContext] with [context.Background]: it reads all bytes from
// provider and decodes them with codec into config, which is normally a
// non-nil pointer.
//
//	cfg := AppConf{Addr: "127.0.0.1:8080"} // defaults
//	if err := confstore.Fill(p, codec.JsonCodec(), &cfg); err != nil {
//		return err
//	}
func Fill[T any](provider provider.Provider, codec codec.Codec, config T) error {
	return FillWithContext(context.Background(), provider, codec, config)
}
