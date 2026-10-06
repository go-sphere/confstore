// Package file provides a configuration [github.com/go-sphere/confstore/provider.Provider]
// that reads a file from disk or from any [io/fs.FS].
//
// Create a provider with [New] and configure it with [WithFS],
// [WithExpandEnv], and [WithTrimBOM]. Use [IsLocalPath] together with
// provider.If to route local paths to this package when the location may
// also be a URL.
//
// # Usage
//
//	import (
//		"github.com/go-sphere/confstore"
//		"github.com/go-sphere/confstore/codec"
//		"github.com/go-sphere/confstore/provider/file"
//	)
//
//	p := file.New("$CONFIG_DIR/app.json", file.WithExpandEnv(), file.WithTrimBOM())
//	cfg, err := confstore.Load[AppConf](p, codec.JsonCodec())
//
// The file is read in full on every Read call; nothing is cached and there is
// no resource to close.
package file
