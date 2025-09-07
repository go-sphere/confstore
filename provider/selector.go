package provider

import "errors"

var (

	// ErrNoValidProvider indicates that no suitable provider could be found after attempting all available options.
	ErrNoValidProvider = errors.New("no valid provider found")
	// ErrNotMatched indicates that a specific condition for selecting a provider was not met.
	ErrNotMatched = errors.New("provider not matched")
	// ErrNilProvider indicates that a case returned a nil Provider without an error.
	ErrNilProvider = errors.New("provider is nil")
)

// Selector tries each case function in order with the given parameter.
// It returns the first Provider that does not return an error.
// If all cases return an error, it returns an error indicating no valid provider was found.
func Selector[T any](param T, cases ...func(T) (Provider, error)) (Provider, error) {
	for _, c := range cases {
		provider, err := c(param)
		if err != nil {
			// Ignore not-matched and continue trying others.
			if errors.Is(err, ErrNotMatched) {
				continue
			}
			// Any other error means this case failed; continue to next.
			continue
		}
		if provider != nil {
			return provider, nil
		}
		// Guard: a case that returns (nil, nil) should not succeed.
		// Treat as non-match and continue.
		continue
	}
	return nil, ErrNoValidProvider
}

// SelectorWithErrors behaves like Selector but aggregates non-matching errors
// (excluding ErrNotMatched) and returns them joined with ErrNoValidProvider
// when no provider is selected. Callers can use errors.Is to test for
// ErrNoValidProvider while still getting detailed context for debugging.
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

// If creates a case function for Selector.
// It takes a condition function and a then function.
// If the condition returns true for the given parameter, it calls the then function to get the Provider.
// If the condition returns false, it returns an error indicating no match.
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

// IfE is like If, but allows the then function to return an error.
// Useful when provider construction can fail and the error should be surfaced.
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
