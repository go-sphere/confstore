# confstore

Lightweight configuration loader for Go. Combine a `Provider` (file, HTTP, etc.) with a `Codec` (JSON by default) to load typed configs.

## Install

```
go get github.com/go-sphere/confstore
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-sphere/confstore"
    "github.com/go-sphere/confstore/codec"
    "github.com/go-sphere/confstore/provider"
)

type AppConf struct {
    Addr string `json:"addr"`
    Mode string `json:"mode"`
}

func main() {
    // Load from a local JSON file
    p := provider.NewFile("./config.json", provider.WithTrimBOM())
    cfg, err := confstore.Load[AppConf](context.Background(), p, codec.JsonCodec())
    if err != nil { panic(err) }
    fmt.Printf("%+v\n", *cfg)
}
```

## Providers

- `provider.File` — load from filesystem or a custom `fs.FS`.
  - Options:
    - `provider.WithFS(fsys fs.FS)`
    - `provider.WithExpandEnv()` — expand env vars in the path
    - `provider.WithTrimBOM()` — trim UTF-8 BOM

- `provider.HTTP` — fetch from HTTP(S).
  - Options:
    - `provider.WithTimeout(d time.Duration)` — client-level timeout for the internal client
    - `provider.WithClient(c *http.Client)`
    - `provider.WithMethod(m string)`
    - `provider.WithHeader(key, value string)` / `provider.WithHeaders(h http.Header)`
    - `provider.WithMaxBodySize(n int64)` — limit response body size (bytes)

### HTTP example

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write([]byte(`{"addr":"0.0.0.0:8080","mode":"prod"}`))
}))
defer srv.Close()

p := provider.NewHTTP(srv.URL,
    provider.WithHeader("Accept", "application/json"),
    provider.WithMaxBodySize(1<<20), // 1MB
)
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
cfg, err := confstore.Load[AppConf](ctx, p, codec.JsonCodec())
```

## ExpandEnv Adapter

Wrap any provider to expand environment variables inside the raw bytes (text configs):

```go
wrapped := provider.NewExpandEnv(provider.NewFile("./config.json"))
```

## Codecs

- `codec.JsonCodec()` — JSON via stdlib
- `codec.FallbackCodecGroup` — try multiple codecs in order

```go
group := codec.NewCodecGroup(codec.JsonCodec() /*, yamlCodec, tomlCodec, ...*/)
```

## Notes

- Errors from the HTTP provider include method and URL. Non-2xx statuses report the full status string.
- When `WithMaxBodySize` is set, bodies exceeding the limit return `provider.ErrBodyTooLarge`.
- Prefer controlling request deadlines with `context.Context` (e.g., `context.WithTimeout`). By default the HTTP client has no timeout; if needed, `provider.WithTimeout` configures a client-level timeout.

## License

**confstore** is released under the MIT license. See [LICENSE](LICENSE) for details.
