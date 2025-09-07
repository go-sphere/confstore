package provider

import (
	"context"
	"errors"
	"testing"
)

type dummyProvider struct{ b []byte }

func (d dummyProvider) Read(ctx context.Context) ([]byte, error) { return d.b, nil }

type testProvider struct {
	readFunc func(context.Context) ([]byte, error)
}

func (tp *testProvider) Read(ctx context.Context) ([]byte, error) {
	return tp.readFunc(ctx)
}

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

func TestSelect_SuccessfulSelection(t *testing.T) {
	// Test NewSelect constructor and Read method when a provider is successfully selected.
	want := "config-data"
	s := NewSelect[string]("test",
		If(func(s string) bool { return s == "other" }, func(s string) Provider { return dummyProvider{b: []byte("wrong")} }),
		If(func(s string) bool { return s == "test" }, func(s string) Provider { return dummyProvider{b: []byte(want)} }),
	)

	got, err := s.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != want {
		t.Fatalf("got %q, want %q", string(got), want)
	}
}

func TestSelect_NoMatchingProvider(t *testing.T) {
	// Test Select.Read when no provider matches - should return ErrNoValidProvider.
	s := NewSelect[string]("nomatch",
		If(func(s string) bool { return s == "foo" }, func(s string) Provider { return dummyProvider{b: []byte("foo")} }),
		If(func(s string) bool { return s == "bar" }, func(s string) Provider { return dummyProvider{b: []byte("bar")} }),
	)

	data, err := s.Read(context.Background())
	if data != nil {
		t.Fatalf("expected nil data, got %v", data)
	}
	if !errors.Is(err, ErrNoValidProvider) {
		t.Fatalf("expected ErrNoValidProvider, got %v", err)
	}
}

func TestSelect_ProviderError(t *testing.T) {
	// Test Select.Read when the selected provider's Read method returns an error.
	providerErr := errors.New("provider read failed")
	readFunc := func(ctx context.Context) ([]byte, error) {
		return nil, providerErr
	}

	s := NewSelect[int](42,
		If(func(i int) bool { return i == 42 }, func(i int) Provider {
			return &testProvider{readFunc: readFunc}
		}),
	)

	data, err := s.Read(context.Background())
	if data != nil {
		t.Fatalf("expected nil data, got %v", data)
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error %v, got %v", providerErr, err)
	}
}

func TestSelect_ContextPropagation(t *testing.T) {
	// Test that context is properly passed to the selected provider's Read method.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to test context propagation

	s := NewSelect[bool](true,
		If(func(b bool) bool { return b }, func(b bool) Provider {
			return &testProvider{readFunc: func(ctx context.Context) ([]byte, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return []byte("should not reach here"), nil
				}
			}}
		}),
	)

	data, err := s.Read(ctx)
	if data != nil {
		t.Fatalf("expected nil data, got %v", data)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestSelect_WithComplexCases(t *testing.T) {
	// Test Select with a mix of If and IfE cases, where some may fail during provider creation.
	constructionErr := errors.New("construction failed")

	s := NewSelect[int](100,
		IfE(func(i int) bool { return i < 50 }, func(i int) (Provider, error) {
			return nil, constructionErr
		}),
		If(func(i int) bool { return i >= 100 }, func(i int) Provider {
			return dummyProvider{b: []byte("high-value")}
		}),
		If(func(i int) bool { return true }, func(i int) Provider {
			return dummyProvider{b: []byte("fallback")}
		}),
	)

	got, err := s.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "high-value" {
		t.Fatalf("got %q, want %q", string(got), "high-value")
	}
}
