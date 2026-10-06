package provider

import (
	"context"
	"errors"
)

var (
	// ErrNoValidProvider is joined into the error returned by Selector,
	// SelectorWithErrors, and Select.Read when no case produced a provider.
	// Test for it with errors.Is.
	ErrNoValidProvider = errors.New("no valid provider found")
	// ErrNotMatched is returned by a selector case whose condition does not
	// apply. Selector skips such cases without recording their error. Custom
	// case functions may return it (or wrap it) to mean "not mine".
	ErrNotMatched = errors.New("provider not matched")
	// ErrNilProvider indicates that a case matched but produced a nil
	// Provider. Selector records it and continues with the next case.
	ErrNilProvider = errors.New("provider is nil")
)

// Selector is equivalent to SelectorWithErrors: it returns the first non-nil
// Provider, and otherwise an error matching ErrNoValidProvider under errors.Is
// that also joins every case failure other than ErrNotMatched. Compare with
// errors.Is, not ==.
func Selector[T any](param T, cases ...func(T) (Provider, error)) (Provider, error) {
	return SelectorWithErrors(param, cases...)
}

// SelectorWithErrors calls each case function with param in order and returns
// the first non-nil Provider; later cases are not called. Cases returning
// [ErrNotMatched] are skipped silently. When no provider is selected it
// returns a nil Provider and an error joining every other case failure
// (including [ErrNilProvider]) with [ErrNoValidProvider]. Callers can use
// errors.Is to test for ErrNoValidProvider while still getting detailed
// context for debugging.
func SelectorWithErrors[T any](param T, cases ...func(T) (Provider, error)) (Provider, error) {
	var joined []error
	for _, c := range cases {
		provider, err := c(param)
		if err != nil {
			if errors.Is(err, ErrNotMatched) {
				continue
			}
			joined = append(joined, err)
			continue
		}
		if provider != nil {
			return provider, nil
		}
		joined = append(joined, ErrNilProvider)
	}
	joined = append(joined, ErrNoValidProvider)
	return nil, errors.Join(joined...)
}

// If returns a case function for [Selector], [SelectorWithErrors], or
// [NewSelect]. The case returns [ErrNotMatched] when cond reports false;
// otherwise it returns then(param), or [ErrNilProvider] if that is nil.
// then is called only when cond reports true.
func If[T any](cond func(T) bool, then func(T) Provider) func(T) (Provider, error) {
	return func(p T) (Provider, error) {
		if cond(p) {
			prov := then(p)
			if prov == nil {
				return nil, ErrNilProvider
			}
			return prov, nil
		}
		return nil, ErrNotMatched
	}
}

// IfE is like [If], but then may fail. When cond reports true and then
// returns an error, the case returns that error unchanged, and Selector records
// it in its joined error before trying the next case. A nil provider with a nil
// error becomes [ErrNilProvider].
func IfE[T any](cond func(T) bool, then func(T) (Provider, error)) func(T) (Provider, error) {
	return func(p T) (Provider, error) {
		if !cond(p) {
			return nil, ErrNotMatched
		}
		prov, err := then(p)
		if err != nil {
			return nil, err
		}
		if prov == nil {
			return nil, ErrNilProvider
		}
		return prov, nil
	}
}

// Select is a [Provider] that defers provider selection until Read. Create
// it with [NewSelect].
type Select[T any] struct {
	param T
	cases []func(T) (Provider, error)
}

// NewSelect returns a [Select] that will run [Selector] with param and cases
// on every Read. No case function is called by NewSelect itself.
func NewSelect[T any](param T, cases ...func(T) (Provider, error)) *Select[T] {
	return &Select[T]{param: param, cases: cases}
}

// Read implements the Provider interface for Select.
// It uses Selector to choose a Provider based on the parameter and cases,
// then calls Read on the selected Provider. Selection runs again on every
// call. If no provider is selected, the error matches ErrNoValidProvider and
// wraps the failures of the cases tried; errors from the selected provider's
// Read are returned unchanged.
func (s *Select[T]) Read(ctx context.Context) ([]byte, error) {
	provider, err := Selector(s.param, s.cases...)
	if err != nil {
		return nil, err
	}
	return provider.Read(ctx)
}
