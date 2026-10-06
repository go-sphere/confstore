// Package reader provides configuration [github.com/go-sphere/confstore/provider.Provider]
// implementations backed by in-memory data: [NewReader] wraps an [io.Reader]
// and [NewBytes] serves a fixed byte slice.
//
// These are useful for embedded defaults, standard input, and tests.
//
// # Usage
//
//	import (
//		"github.com/go-sphere/confstore"
//		"github.com/go-sphere/confstore/codec"
//		"github.com/go-sphere/confstore/provider/reader"
//	)
//
//	p := reader.NewBytes([]byte(`{"addr":":8080"}`))
//	cfg, err := confstore.Load[AppConf](p, codec.JsonCodec())
package reader
