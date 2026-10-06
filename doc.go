// Package confstore loads typed configuration by combining a byte source
// with a decoder.
//
// A [github.com/go-sphere/confstore/provider.Provider] returns the raw
// configuration bytes (from a file, an HTTP endpoint, an io.Reader, or a fixed
// slice) and a [github.com/go-sphere/confstore/codec.Codec] decodes them. Use
// [Load] to allocate and return a new value of type T, or [Fill] to decode into
// a value the caller already owns, for example one pre-populated with
// defaults. [LoadWithContext] and [FillWithContext] pass a context to the
// provider for cancellation and deadlines.
//
// # Usage
//
//	import (
//		"github.com/go-sphere/confstore"
//		"github.com/go-sphere/confstore/codec"
//		"github.com/go-sphere/confstore/provider/file"
//	)
//
//	type AppConf struct {
//		Addr string `json:"addr"`
//		Mode string `json:"mode"`
//	}
//
//	p := file.New("./config.json", file.WithTrimBOM())
//	cfg, err := confstore.Load[AppConf](p, codec.JsonCodec())
//	if err != nil {
//		return err
//	}
//
// # Building blocks
//
//   - Package provider/file reads from disk or any io/fs.FS.
//   - Package provider/http fetches from an HTTP(S) endpoint.
//   - Package provider/reader adapts an io.Reader or a fixed byte slice.
//   - Package provider wraps providers: NewExpandEnv expands environment
//     variables in the content, and Selector/NewSelect choose a provider at
//     runtime (for example local path versus remote URL).
//   - Package codec supplies JsonCodec, StringCodec, NewCodec for custom
//     formats, and NewCodecGroup to try several codecs in order.
//
// The functions in this package hold no state and do not wrap errors: a
// provider or codec error is returned unchanged.
package confstore
