package confstore

import (
	"context"
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
