// Package codec converts configuration values to and from bytes for
// github.com/go-sphere/confstore.
//
// [Codec] combines [Encoder] and [Decoder]. Ready-made codecs:
//
//   - [JsonCodec] uses encoding/json.
//   - [StringCodec] copies bytes to and from a string without parsing.
//   - [NewCodec] builds a Codec from any marshal/unmarshal function pair,
//     for example yaml.Marshal and yaml.Unmarshal from a YAML library.
//   - [NewCodecGroup] tries several codecs in order and uses the first that
//     succeeds.
//
// # Usage
//
//	import "github.com/go-sphere/confstore/codec"
//
//	var cfg struct {
//		Addr string `json:"addr"`
//	}
//	c := codec.NewCodecGroup(codec.JsonCodec())
//	if err := c.Unmarshal([]byte(`{"addr":":8080"}`), &cfg); err != nil {
//		return err
//	}
//
// Codecs returned by this package hold no mutable state and may be shared.
package codec
