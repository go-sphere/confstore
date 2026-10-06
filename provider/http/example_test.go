package http_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	confhttp "github.com/go-sphere/confstore/provider/http"
)

func ExampleNew() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"addr":"0.0.0.0:8080","mode":"prod"}`))
	}))
	defer srv.Close()

	p := confhttp.New(srv.URL,
		confhttp.WithHeader("Authorization", "Bearer token"),
		confhttp.WithMaxBodySize(1<<20),
	)

	type AppConf struct {
		Addr string `json:"addr"`
		Mode string `json:"mode"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := confstore.LoadWithContext[AppConf](ctx, p, codec.JsonCodec())
	if err != nil {
		fmt.Println("load:", err)
		return
	}
	fmt.Printf("%+v\n", *cfg)
	// Output: {Addr:0.0.0.0:8080 Mode:prod}
}

func ExampleWithMaxBodySize() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer srv.Close()

	p := confhttp.New(srv.URL, confhttp.WithMaxBodySize(4))
	_, err := p.Read(context.Background())
	fmt.Println(errors.Is(err, confhttp.ErrBodyTooLarge))
	// Output: true
}

func ExampleIsRemoteURL() {
	for _, s := range []string{"https://example.com/app.json", "HTTP://example.com", "./config.json", "http:///no-host"} {
		fmt.Printf("%q %v\n", s, confhttp.IsRemoteURL(s))
	}
	// Output:
	// "https://example.com/app.json" true
	// "HTTP://example.com" true
	// "./config.json" false
	// "http:///no-host" false
}
