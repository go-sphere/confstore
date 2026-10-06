package codec_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-sphere/confstore/codec"
)

func ExampleNewCodecGroup() {
	// Accept either a JSON document or, failing that, raw text.
	type note struct {
		Text string `json:"text"`
	}
	fromText := codec.NewCodec(json.Marshal, func(data []byte, v any) error {
		n, ok := v.(*note)
		if !ok {
			return codec.ErrInvalidType
		}
		n.Text = string(data)
		return nil
	})
	g := codec.NewCodecGroup(codec.JsonCodec(), fromText)

	for _, in := range []string{`{"text":"from json"}`, `plain text`} {
		var n note
		if err := g.Unmarshal([]byte(in), &n); err != nil {
			fmt.Println("unmarshal:", err)
			return
		}
		fmt.Println(n.Text)
	}
	// Output:
	// from json
	// plain text
}

func ExampleStringCodec() {
	c := codec.StringCodec()
	var s string
	if err := c.Unmarshal([]byte("hello"), &s); err != nil {
		fmt.Println("unmarshal:", err)
		return
	}
	fmt.Println(s)

	var n int
	err := c.Unmarshal([]byte("42"), &n)
	fmt.Println(errors.Is(err, codec.ErrInvalidType))
	// Output:
	// hello
	// true
}
