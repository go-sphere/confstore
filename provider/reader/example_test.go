package reader_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-sphere/confstore/provider/reader"
)

func ExampleNewReader() {
	p := reader.NewReader(strings.NewReader("addr=:8080"))
	ctx := context.Background()

	first, err := p.Read(ctx)
	if err != nil {
		fmt.Println("read:", err)
		return
	}
	// The io.Reader is drained by the first Read.
	second, err := p.Read(ctx)
	if err != nil {
		fmt.Println("read:", err)
		return
	}
	fmt.Printf("%q %q\n", first, second)
	// Output: "addr=:8080" ""
}

func ExampleNewBytes() {
	p := reader.NewBytes([]byte("addr=:8080"))
	data, err := p.Read(context.Background())
	if err != nil {
		fmt.Println("read:", err)
		return
	}
	data[0] = 'X' // Read returns a copy; the provider is unaffected.
	again, err := p.Read(context.Background())
	if err != nil {
		fmt.Println("read:", err)
		return
	}
	fmt.Println(string(again))
	// Output: addr=:8080
}
