# Changelog

## Unreleased

## v0.0.6 (2026-11-19)

### Fixes

- `provider/http`: stop draining `resp.Body` when rejecting oversized responses (`WithMaxBodySize`), bound non-2xx connection-reuse draining to 4096 bytes with `io.CopyN`, and guard `WithMaxBodySize(math.MaxInt64)` against signed `int64` overflow in `io.LimitReader`.
- `codec`: clone the `codecs` slice in `NewCodecGroup` and guard against `nil` `Codec` elements in `FallbackCodecGroup.Marshal` and `FallbackCodecGroup.Unmarshal` so `nil` entries record an error and fall back instead of panicking.
- `provider`: fast-fail `(*Select).Read` when `ctx.Err() != nil` before running selector cases.
- `provider`: `Selector` now reports case errors joined with `ErrNoValidProvider` (via `SelectorWithErrors`).

### Documentation & Tooling

- Added package documentation and runnable examples across `confstore`, `codec`, `provider`, `provider/file`, `provider/http`, and `provider/reader`.
- Fixed `LoadWithContext[AppConf]` example and documented `provider/reader`, `provider.Select`, and `codec.StringCodec()` in `README.md`.
- Added `.gitattributes` and `CHANGELOG.md`.
