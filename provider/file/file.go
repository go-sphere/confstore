package file

import (
	"bytes"
	"context"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// File is a provider.Provider that returns the contents of a single file
// from the OS filesystem or from an [io/fs.FS].
//
// Create it with [New]; the zero value is not usable. A File is immutable
// after construction and safe for concurrent use.
type File struct {
	path string
	opts *options
}

type options struct {
	fsys      fs.FS
	expandEnv bool
	trimBOM   bool
}

// Option configures optional behavior for [New]. Options are applied in
// order when New is called.
type Option func(*options)

// WithFS makes Read load the path from fsys via [io/fs.ReadFile] instead of
// the OS filesystem. The path must then follow io/fs path rules: slash
// separated and unrooted, such as "conf/app.json". A nil fsys keeps the OS
// filesystem.
func WithFS(fsys fs.FS) Option { return func(o *options) { o.fsys = fsys } }

// WithExpandEnv expands environment variables in the path with
// [os.ExpandEnv] on every Read, e.g. "$HOME/app/config.json". Undefined
// variables expand to an empty string. The file contents are not expanded;
// wrap the provider with provider.NewExpandEnv for that.
func WithExpandEnv() Option { return func(o *options) { o.expandEnv = true } }

// WithTrimBOM removes a leading UTF-8 byte order mark (EF BB BF) from the
// returned bytes when present. Other content is returned unchanged.
func WithTrimBOM() Option { return func(o *options) { o.trimBOM = true } }

func newOptions(opts ...Option) *options {
	defaults := &options{}
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}

// New returns a [File] that reads path. The path is not checked until Read;
// a relative path is resolved against the working directory at Read time (or
// against the [WithFS] filesystem). Without options the raw file bytes are
// returned unchanged.
func New(path string, opts ...Option) *File {
	return &File{path: path, opts: newOptions(opts...)}
}

// Read loads the file contents and returns the raw bytes, applying the
// configured options. Read errors from the filesystem are returned unchanged,
// so errors.Is(err, fs.ErrNotExist) detects a missing file.
//
// The context is only honored for fast-fail on cancellation; the underlying
// os.ReadFile / fs.ReadFile cannot be cancelled mid-flight. Pre-cancel a
// context to avoid the read entirely. If ctx is done before or after the
// read, Read returns ctx.Err() and no data.
func (f *File) Read(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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
	// The context may have expired while the read was in flight.
	if err := ctx.Err(); err != nil {
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

// IsLocalPath reports whether path should be treated as a local filesystem
// path. It returns false for an empty string and for URLs with a scheme other
// than "file" (such as "https://..."); absolute paths, relative paths, and
// "file:" URLs return true. IsLocalPath does not check that the file exists.
// Note that New does not interpret "file://" URLs; pass a plain path to it.
func IsLocalPath(path string) bool {
	if path == "" {
		return false
	}
	if filepath.IsAbs(path) {
		return true
	}
	u, err := url.Parse(path)
	if err == nil && u.Scheme != "" {
		// A URL with any scheme other than file:// is remote (e.g. http/https).
		return strings.EqualFold(u.Scheme, "file")
	}
	// Absolute, relative, or non-URL string => local.
	return true
}
