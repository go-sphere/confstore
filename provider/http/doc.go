// Package http provides a configuration [github.com/go-sphere/confstore/provider.Provider]
// that fetches bytes from an HTTP(S) endpoint.
//
// Create a provider with [New] and configure the request with [WithMethod],
// [WithHeader], and [WithHeaders]; configure the transport with [WithClient]
// or [WithTimeout]; and bound the response with [WithMaxBodySize]. Use
// [IsRemoteURL] together with provider.If to route URLs to this package when
// the location may also be a local path.
//
// # Usage
//
// The package name collides with net/http, so import it under an alias:
//
//	import (
//		"github.com/go-sphere/confstore"
//		"github.com/go-sphere/confstore/codec"
//		confhttp "github.com/go-sphere/confstore/provider/http"
//	)
//
//	p := confhttp.New("https://config.example.com/app.json",
//		confhttp.WithHeader("Accept", "application/json"),
//		confhttp.WithMaxBodySize(1<<20),
//	)
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	cfg, err := confstore.LoadWithContext[AppConf](ctx, p, codec.JsonCodec())
//
// Every Read performs a new request; responses are not cached. Only 2xx
// responses are accepted.
package http
