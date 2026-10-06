// Package provider defines the [Provider] interface used by
// github.com/go-sphere/confstore to obtain raw configuration bytes, plus
// adapters that compose providers.
//
// Concrete sources live in subpackages: provider/file, provider/http, and
// provider/reader. This package adds:
//
//   - [ReaderFunc] to turn a function into a Provider.
//   - [NewExpandEnv] to expand $VAR and ${VAR} placeholders in the content.
//   - [Selector], [NewSelect], [If], and [IfE] to pick one provider at run
//     time from ordered cases, for example a remote URL versus a local path.
//
// # Usage
//
//	import (
//		"github.com/go-sphere/confstore/provider"
//		"github.com/go-sphere/confstore/provider/file"
//		confhttp "github.com/go-sphere/confstore/provider/http"
//	)
//
//	p := provider.NewExpandEnv(provider.NewSelect(path,
//		provider.If(confhttp.IsRemoteURL, func(s string) provider.Provider { return confhttp.New(s) }),
//		provider.If(file.IsLocalPath, func(s string) provider.Provider { return file.New(s) }),
//	))
//	data, err := p.Read(ctx)
//
// Selection failures match [ErrNoValidProvider] under errors.Is; compare with
// errors.Is rather than ==, because the returned error also joins the
// failures of the cases that were tried.
package provider
