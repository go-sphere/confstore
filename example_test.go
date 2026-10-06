package confstore_test

import (
	"fmt"

	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider/reader"
)

type AppConf struct {
	Addr string `json:"addr"`
	Mode string `json:"mode"`
}

func Example() {
	p := reader.NewBytes([]byte(`{"addr":"127.0.0.1:8080","mode":"dev"}`))
	cfg, err := confstore.Load[AppConf](p, codec.JsonCodec())
	if err != nil {
		fmt.Println("load:", err)
		return
	}
	fmt.Printf("%+v\n", *cfg)
	// Output: {Addr:127.0.0.1:8080 Mode:dev}
}

func ExampleFill() {
	// Values absent from the data keep their defaults.
	cfg := AppConf{Addr: "127.0.0.1:8080", Mode: "dev"}
	p := reader.NewBytes([]byte(`{"mode":"prod"}`))
	if err := confstore.Fill(p, codec.JsonCodec(), &cfg); err != nil {
		fmt.Println("fill:", err)
		return
	}
	fmt.Printf("%+v\n", cfg)
	// Output: {Addr:127.0.0.1:8080 Mode:prod}
}
