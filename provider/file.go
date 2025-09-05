package provider

import (
	"bytes"
	"context"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
)

// File provides configuration bytes loaded from a file on disk or any fs.FS.
// Required: a file path. Optional: supply a custom fs, expand env vars in path, trim UTF-8 BOM.
type File struct {
	path string
	opts *fileOptions
}

type fileOptions struct {
	fsys      fs.FS
	expandEnv bool
	trimBOM   bool
}

// FileOption configures optional behavior for the file provider.
type FileOption func(*fileOptions)

// WithFS sets a custom filesystem to read from. When provided, the path is
// interpreted relative to that filesystem and read via fs.ReadFile.
func WithFS(fsys fs.FS) FileOption { return func(o *fileOptions) { o.fsys = fsys } }

// WithExpandEnv enables environment-variable expansion in the provided path
// using os.ExpandEnv, e.g. "$HOME/app/config.json".
func WithExpandEnv() FileOption { return func(o *fileOptions) { o.expandEnv = true } }

// WithTrimBOM trims UTF-8 BOM if present at the beginning of the file.
func WithTrimBOM() FileOption { return func(o *fileOptions) { o.trimBOM = true } }

func newFileOptions(opts ...FileOption) *fileOptions {
	defaults := &fileOptions{}
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}

// NewFile creates a file-backed Provider.
// path: required file path. Options control reading behavior.
func NewFile(path string, opts ...FileOption) *File {
	return &File{path: path, opts: newFileOptions(opts...)}
}

// Read implements Provider by loading the file contents.
func (f *File) Read(_ context.Context) ([]byte, error) {
	path := f.path
	if f.opts.expandEnv {
		path = os.ExpandEnv(path)
	}

	var (
		data []byte
		err  error
	)
	if f.opts.fsys != nil {
		data, err = fs.ReadFile(f.opts.fsys, path)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}

	if f.opts.trimBOM && len(data) >= 3 {
		// Trim UTF-8 BOM if present
		if bytes.Equal(data[:3], []byte{0xEF, 0xBB, 0xBF}) {
			data = data[3:]
		}
	}
	return data, nil
}

// IsLocalPath reports whether the given path is a local filesystem path.
func IsLocalPath(path string) bool {
	if path == "" {
		return false
	}
	if filepath.IsAbs(path) {
		return true
	}
	if u, err := url.Parse(path); err == nil && u.Scheme != "" {
		return u.Scheme == "file"
	}
	return true
}
