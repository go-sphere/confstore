package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	nethttp "net/http"
	"strings"
	"testing"
	"time"
)

type rtFunc func(*nethttp.Request) (*nethttp.Response, error)

func (f rtFunc) RoundTrip(r *nethttp.Request) (*nethttp.Response, error) { return f(r) }

func TestHTTPReadSuccess(t *testing.T) {
	want := "hello"
	url := "http://example/ok"
	c := &nethttp.Client{Transport: rtFunc(func(r *nethttp.Request) (*nethttp.Response, error) {
		if r.URL.String() != url {
			t.Fatalf("unexpected url: %s", r.URL.String())
		}
		return &nethttp.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Body:          io.NopCloser(strings.NewReader(want)),
			ContentLength: int64(len(want)),
			Header:        make(nethttp.Header),
			Request:       r,
		}, nil
	})}

	p := New(url, WithClient(c))
	got, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if string(got) != want {
		t.Fatalf("got %q, want %q", string(got), want)
	}
}

func TestHTTPStatusError(t *testing.T) {
	url := "http://example/err"
	c := &nethttp.Client{Transport: rtFunc(func(r *nethttp.Request) (*nethttp.Response, error) {
		return &nethttp.Response{
			Status:        "500 Internal Server Error",
			StatusCode:    500,
			Body:          io.NopCloser(strings.NewReader("oops")),
			ContentLength: 4,
			Header:        make(nethttp.Header),
			Request:       r,
		}, nil
	})}

	p := New(url, WithMethod(nethttp.MethodGet), WithClient(c))
	_, err := p.Read(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unexpected status") || !strings.Contains(msg, nethttp.StatusText(nethttp.StatusInternalServerError)) {
		t.Fatalf("error lacks status context: %v", msg)
	}
	if !strings.Contains(msg, url) || !strings.Contains(msg, nethttp.MethodGet) {
		t.Fatalf("error lacks method/url context: %v", msg)
	}
}

func TestHTTPBodyTooLarge(t *testing.T) {
	big := bytes.Repeat([]byte("a"), 2000)
	url := "http://example/big"
	c := &nethttp.Client{Transport: rtFunc(func(r *nethttp.Request) (*nethttp.Response, error) {
		return &nethttp.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Body:          io.NopCloser(bytes.NewReader(big)),
			ContentLength: int64(len(big)),
			Header:        make(nethttp.Header),
			Request:       r,
		}, nil
	})}

	p := New(url, WithClient(c), WithMaxBodySize(1024)) // 1KB
	_, err := p.Read(context.Background())
	if err == nil {
		t.Fatal("expected error for oversized body, got nil")
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestHTTPContextTimeout(t *testing.T) {
	url := "http://example/slow"
	c := &nethttp.Client{Transport: rtFunc(func(r *nethttp.Request) (*nethttp.Response, error) {
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-time.After(50 * time.Millisecond):
			// too slow; should be canceled by context
			return &nethttp.Response{
				Status:        "200 OK",
				StatusCode:    200,
				Body:          io.NopCloser(strings.NewReader("late")),
				ContentLength: 4,
				Header:        make(nethttp.Header),
				Request:       r,
			}, nil
		}
	})}

	p := New(url, WithClient(c))
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
