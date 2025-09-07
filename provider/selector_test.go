package provider

import (
	"context"
	"errors"
	"testing"
)

type dummyProvider struct{ b []byte }

func (d dummyProvider) Read(ctx context.Context) ([]byte, error) { return d.b, nil }

func TestSelector_FirstMatchWins(t *testing.T) {
	// First case matches and returns a provider; second also matches but should not be used.
	p, err := Selector[int](42,
		If(func(i int) bool { return i == 42 }, func(i int) Provider { return dummyProvider{b: []byte("first")} }),
		If(func(i int) bool { return true }, func(i int) Provider { return dummyProvider{b: []byte("second")} }),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatalf("expected provider, got nil")
	}
}

func TestSelector_AllNotMatched(t *testing.T) {
	p, err := Selector[int](1,
		If(func(i int) bool { return false }, func(i int) Provider { return dummyProvider{b: []byte("x")} }),
		If(func(i int) bool { return false }, func(i int) Provider { return dummyProvider{b: []byte("y")} }),
	)
	if p != nil {
		t.Fatalf("expected nil provider, got %#v", p)
	}
	if !errors.Is(err, ErrNoValidProvider) {
		t.Fatalf("expected ErrNoValidProvider, got %v", err)
	}
}

func TestSelector_NilProviderIgnored(t *testing.T) {
	// A case returning (nil, nil) should not be selected; overall should fail.
	p, err := Selector[int](0,
		func(i int) (Provider, error) { return nil, nil },
		If(func(i int) bool { return false }, func(i int) Provider { return dummyProvider{b: []byte("x")} }),
	)
	if p != nil {
		t.Fatalf("expected nil provider, got %#v", p)
	}
	if !errors.Is(err, ErrNoValidProvider) {
		t.Fatalf("expected ErrNoValidProvider, got %v", err)
	}
}

func TestIfE_AndSelectorWithErrors(t *testing.T) {
	boom := errors.New("boom")
	// First matches but fails to build; second returns nil provider; expect aggregated error.
	p, err := SelectorWithErrors[int](7,
		IfE(func(i int) bool { return i == 7 }, func(i int) (Provider, error) { return nil, boom }),
		func(i int) (Provider, error) { return nil, nil },
	)
	if p != nil {
		t.Fatalf("expected nil provider, got %#v", p)
	}
	if !errors.Is(err, ErrNoValidProvider) {
		t.Fatalf("expected ErrNoValidProvider, got %v", err)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected joined boom error, got %v", err)
	}
	if !errors.Is(err, ErrNilProvider) {
		t.Fatalf("expected joined ErrNilProvider, got %v", err)
	}
}
