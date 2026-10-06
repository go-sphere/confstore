package provider_test

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-sphere/confstore/provider"
	"github.com/go-sphere/confstore/provider/file"
	confhttp "github.com/go-sphere/confstore/provider/http"
	"github.com/go-sphere/confstore/provider/reader"
)

func ExampleSelector() {
	pick := func(location string) (provider.Provider, error) {
		return provider.Selector(location,
			provider.If(confhttp.IsRemoteURL, func(s string) provider.Provider { return confhttp.New(s) }),
			provider.If(file.IsLocalPath, func(s string) provider.Provider { return file.New(s) }),
		)
	}

	for _, loc := range []string{"https://config.example.com/app.json", "./config.json", ""} {
		p, err := pick(loc)
		switch {
		case errors.Is(err, provider.ErrNoValidProvider):
			fmt.Printf("%q: no provider\n", loc)
		case err != nil:
			fmt.Printf("%q: %v\n", loc, err)
		default:
			fmt.Printf("%q: %T\n", loc, p)
		}
	}
	// Output:
	// "https://config.example.com/app.json": *http.HTTP
	// "./config.json": *file.File
	// "": no provider
}

func ExampleNewExpandEnv() {
	if err := os.Setenv("CONFSTORE_EXAMPLE_PORT", "8080"); err != nil {
		fmt.Println("setenv:", err)
		return
	}
	defer func() { _ = os.Unsetenv("CONFSTORE_EXAMPLE_PORT") }()

	p := provider.NewExpandEnv(reader.NewBytes([]byte(`{"addr":":${CONFSTORE_EXAMPLE_PORT}"}`)))
	data, err := p.Read(context.Background())
	if err != nil {
		fmt.Println("read:", err)
		return
	}
	fmt.Println(string(data))
	// Output: {"addr":":8080"}
}

func ExampleReaderFunc() {
	p := provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
		return []byte("generated"), nil
	})
	data, err := p.Read(context.Background())
	if err != nil {
		fmt.Println("read:", err)
		return
	}
	fmt.Println(string(data))
	// Output: generated
}
