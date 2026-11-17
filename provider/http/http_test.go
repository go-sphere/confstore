package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPReadSuccessCases(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		headers    map[string]string
		wantBody   string
		wantHeader string
		wantMethod string
	}{
		{
			name:       "default-get",
			wantBody:   "ok",
			wantHeader: "",
			wantMethod: nethttp.MethodGet,
		},
		{
			name:       "custom-method-and-header",
			method:     nethttp.MethodPost,
			headers:    map[string]string{"X-Test": "yes"},
			wantBody:   "hello",
			wantHeader: "yes",
			wantMethod: nethttp.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
				if r.Method != tt.wantMethod {
					t.Fatalf("unexpected method: %s", r.Method)
				}
				if tt.wantHeader != "" && r.Header.Get("X-Test") != tt.wantHeader {
					t.Fatalf("unexpected header: %q", r.Header.Get("X-Test"))
				}
				_, _ = w.Write([]byte(tt.wantBody))
			}))
			defer srv.Close()

			opts := []Option{}
			if tt.method != "" {
				opts = append(opts, WithMethod(tt.method))
			}
			if len(tt.headers) > 0 {
				for k, v := range tt.headers {
					opts = append(opts, WithHeader(k, v))
				}
			}

			p := New(srv.URL, opts...)
			got, err := p.Read(context.Background())
			if err != nil {
				t.Fatalf("Read error: %v", err)
			}
			if string(got) != tt.wantBody {
				t.Fatalf("got %q, want %q", string(got), tt.wantBody)
			}
		})
	}
}

func TestHTTPStatusError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusInternalServerError)
		_, _ = w.Write([]byte("oops"))
	}))
	defer srv.Close()

	p := New(srv.URL, WithMethod(nethttp.MethodGet))
	_, err := p.Read(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unexpected status") || !strings.Contains(msg, nethttp.StatusText(nethttp.StatusInternalServerError)) {
		t.Fatalf("error lacks status context: %v", msg)
	}
	if !strings.Contains(msg, srv.URL) || !strings.Contains(msg, nethttp.MethodGet) {
		t.Fatalf("error lacks method/url context: %v", msg)
	}
}

func TestHTTPBodyTooLargeFastFail(t *testing.T) {
	big := bytes.Repeat([]byte("a"), 2000)
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Length", "2000")
		_, _ = w.Write(big)
	}))
	defer srv.Close()

	p := New(srv.URL, WithMaxBodySize(1024)) // 1KB
	_, err := p.Read(context.Background())
	if err == nil {
		t.Fatal("expected error for oversized body, got nil")
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestHTTPBodyTooLargeAfterRead(t *testing.T) {
	big := bytes.Repeat([]byte("b"), 2048)
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write(big)
	}))
	defer srv.Close()

	p := New(srv.URL, WithMaxBodySize(1024)) // 1KB
	_, err := p.Read(context.Background())
	if err == nil {
		t.Fatal("expected error for oversized body, got nil")
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestHTTPContextTimeout(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(50 * time.Millisecond):
			_, _ = io.WriteString(w, "late")
		}
	}))
	defer srv.Close()

	p := New(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err := p.Read(ctx)
	if err == nil {
		t.Fatal("expected context timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestHTTPWithClientPreservesTimeout(t *testing.T) {
	customClient := &nethttp.Client{
		Timeout: 42 * time.Second,
	}
	p := New("http://example.com", WithClient(customClient), WithTimeout(5*time.Second))
	if p.opts.client.Timeout != 42*time.Second {
		t.Fatalf("custom client Timeout was mutated: got %v, want 42s", p.opts.client.Timeout)
	}
}

func TestHTTPWithTimeoutDefaultClient(t *testing.T) {
	p := New("http://example.com", WithTimeout(15*time.Second))
	if p.opts.client.Timeout != 15*time.Second {
		t.Fatalf("default client Timeout was not set: got %v, want 15s", p.opts.client.Timeout)
	}
}

func newBlockingStreamServer(t *testing.T, statusCode int, contentLength string, chunk []byte) *httptest.Server {
	t.Helper()
	unblock := make(chan struct{})
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if contentLength != "" {
			w.Header().Set("Content-Length", contentLength)
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write(chunk)
		if flusher, ok := w.(nethttp.Flusher); ok {
			flusher.Flush()
		}
		select {
		case <-r.Context().Done():
		case <-unblock:
		}
	}))
	t.Cleanup(func() {
		close(unblock)
		srv.Close()
	})
	return srv
}

func TestHTTPBodyTooLargeStreaming(t *testing.T) {
	srv := newBlockingStreamServer(t, nethttp.StatusOK, "", bytes.Repeat([]byte("a"), 2048))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p := New(srv.URL, WithMaxBodySize(1024))
	_, err := p.Read(ctx)
	if ctx.Err() != nil {
		t.Fatalf("Read blocked until context expired: %v (err=%v)", ctx.Err(), err)
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestHTTPBodyTooLargeFastFailStreaming(t *testing.T) {
	srv := newBlockingStreamServer(t, nethttp.StatusOK, "4096", bytes.Repeat([]byte("a"), 2048))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p := New(srv.URL, WithMaxBodySize(1024))
	_, err := p.Read(ctx)
	if ctx.Err() != nil {
		t.Fatalf("Read blocked until context expired: %v (err=%v)", ctx.Err(), err)
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestHTTPNon2xxStreamingBoundedDrain(t *testing.T) {
	srv := newBlockingStreamServer(t, nethttp.StatusInternalServerError, "", bytes.Repeat([]byte("e"), 8192))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p := New(srv.URL)
	_, err := p.Read(ctx)
	if ctx.Err() != nil {
		t.Fatalf("Read blocked until context expired: %v (err=%v)", ctx.Err(), err)
	}
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("expected unexpected status error, got %v", err)
	}
}

func TestHTTPMaxBodySizeMaxInt64(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = io.WriteString(w, "hello")
	}))
	defer srv.Close()

	p := New(srv.URL, WithMaxBodySize(math.MaxInt64))
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q, want %q", string(got), "hello")
	}
}
