package file_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"testing/fstest"

	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider/file"
)

func ExampleNew() {
	fsys := fstest.MapFS{
		// The leading bytes are a UTF-8 byte order mark.
		"conf/app.json": {Data: []byte("\xEF\xBB\xBF{\"addr\":\":8080\"}")},
	}
	p := file.New("conf/app.json", file.WithFS(fsys), file.WithTrimBOM())

	var cfg struct {
		Addr string `json:"addr"`
	}
	if err := confstore.Fill(p, codec.JsonCodec(), &cfg); err != nil {
		fmt.Println("load:", err)
		return
	}
	fmt.Println(cfg.Addr)
	// Output: :8080
}

func ExampleFile_Read_notExist() {
	p := file.New("missing.json", file.WithFS(fstest.MapFS{}))
	_, err := p.Read(context.Background())
	fmt.Println(errors.Is(err, fs.ErrNotExist))
	// Output: true
}

func ExampleIsLocalPath() {
	for _, p := range []string{"./config.json", "/etc/app.json", "file:///etc/app.json", "https://example.com/app.json", ""} {
		fmt.Printf("%q %v\n", p, file.IsLocalPath(p))
	}
	// Output:
	// "./config.json" true
	// "/etc/app.json" true
	// "file:///etc/app.json" true
	// "https://example.com/app.json" false
	// "" false
}
