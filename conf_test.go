package confstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
)

type appConf struct {
	Addr string `json:"addr"`
	Mode string `json:"mode"`
}

func TestLoadWithFileJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	content := []byte(`{"addr":"127.0.0.1:8080","mode":"dev"}`)
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	cfg, err := Load[appConf](provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
		return os.ReadFile(p)
	}), codec.JsonCodec())
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Addr != "127.0.0.1:8080" || cfg.Mode != "dev" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadWithContextCases(t *testing.T) {
	decodeErr := errors.New("decode failed")
	readErr := errors.New("read failed")

	tests := []struct {
		name    string
		reader  provider.Provider
		codec   codec.Codec
		want    appConf
		wantErr error
	}{
		{
			name: "success",
			reader: provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
				return []byte(`{"addr":"127.0.0.1:8080","mode":"dev"}`), nil
			}),
			codec: codec.JsonCodec(),
			want:  appConf{Addr: "127.0.0.1:8080", Mode: "dev"},
		},
		{
			name: "provider-error",
			reader: provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
				return nil, readErr
			}),
			codec:   codec.JsonCodec(),
			wantErr: readErr,
		},
		{
			name: "codec-error",
			reader: provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
				return []byte("bad"), nil
			}),
			codec: codec.NewCodec(func(val any) ([]byte, error) { return nil, nil }, func(data []byte, val any) error {
				return decodeErr
			}),
			wantErr: decodeErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadWithContext[appConf](context.Background(), tt.reader, tt.codec)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected config, got nil")
			}
			if *cfg != tt.want {
				t.Fatalf("got %+v, want %+v", *cfg, tt.want)
			}
		})
	}
}

func TestFillWithContextCases(t *testing.T) {
	decodeErr := errors.New("decode failed")

	tests := []struct {
		name    string
		reader  provider.Provider
		codec   codec.Codec
		cfg     *appConf
		want    appConf
		wantErr error
	}{
		{
			name: "success",
			reader: provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
				return []byte(`{"addr":"127.0.0.1:8080","mode":"prod"}`), nil
			}),
			codec: codec.JsonCodec(),
			cfg:   &appConf{},
			want:  appConf{Addr: "127.0.0.1:8080", Mode: "prod"},
		},
		{
			name: "codec-error",
			reader: provider.ReaderFunc(func(ctx context.Context) ([]byte, error) {
				return []byte("bad"), nil
			}),
			codec: codec.NewCodec(func(val any) ([]byte, error) { return nil, nil }, func(data []byte, val any) error {
				return decodeErr
			}),
			cfg:     &appConf{},
			wantErr: decodeErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FillWithContext(context.Background(), tt.reader, tt.codec, tt.cfg)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *tt.cfg != tt.want {
				t.Fatalf("got %+v, want %+v", *tt.cfg, tt.want)
			}
		})
	}
}
