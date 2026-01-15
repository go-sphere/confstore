package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestFileReadOptions(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		opts    []Option
		want    string
		wantErr bool
	}{
		{
			name: "with-fs",
			path: "config.json",
			opts: []Option{WithFS(fstest.MapFS{
				"config.json": {Data: []byte("value")},
			})},
			want: "value",
		},
		{
			name: "trim-bom",
			path: "config.json",
			opts: []Option{WithFS(fstest.MapFS{
				"config.json": {Data: append([]byte{0xEF, 0xBB, 0xBF}, []byte("text")...)},
			}), WithTrimBOM()},
			want: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.path, tt.opts...)
			got, err := p.Read(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestFileExpandEnvPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(filePath, []byte("env-data"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	t.Setenv("CONF_DIR", dir)

	p := New(filepath.Join("$CONF_DIR", "config.json"), WithExpandEnv())
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "env-data" {
		t.Fatalf("got %q, want %q", string(got), "env-data")
	}
}

func TestIsLocalPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "empty", path: "", want: false},
		{name: "http", path: "http://example.com/config", want: false},
		{name: "https", path: "https://example.com/config", want: false},
		{name: "file-scheme", path: "file:///tmp/config", want: true},
		{name: "absolute", path: filepath.Join(string(os.PathSeparator), "tmp", "config"), want: true},
		{name: "relative", path: "configs/app.json", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocalPath(tt.path); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileReadMissing(t *testing.T) {
	p := New("missing.json", WithFS(fstest.MapFS{}))
	if _, err := p.Read(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFileWithFSUsesProvidedPath(t *testing.T) {
	fsys := fstest.MapFS{
		"nested/config.json": {Data: []byte("value")},
	}
	p := New("nested/config.json", WithFS(fsys))
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "value" {
		t.Fatalf("got %q, want %q", string(got), "value")
	}
}
