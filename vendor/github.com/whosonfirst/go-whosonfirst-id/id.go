package id

import (
	"context"
	"fmt"

	_ "github.com/aaronland/go-uid-proxy"
	_ "github.com/aaronland/go-uid-whosonfirst"
	
	"github.com/aaronland/go-uid"
)

func NewID() (int64, error) {

	ctx := context.Background()
	pr, err := NewProvider(ctx)

	if err != nil {
		return -1, err
	}

	return pr.NewID(ctx)
}

// type Provider is an interface for providing uniquer identifiers.
type Provider interface {
	// NewID returns a new unique 64-bit integers.
	NewID(context.Context) (int64, error)
}

// WOFProvider implements the Provider interface for generating unique Who's On First identifiers.
type WOFProvider struct {
	Provider
	uid_provider uid.Provider
}

// NweProvider returns a new `WOFProvider` instance configured with default
// settings.
func NewProvider(ctx context.Context) (Provider, error) {
	uri := "proxy://?provider=whosonfirst://"
	// uri := "proxy:///?provider=whosonfirst://&minimum=5&pool=memory%3A%2F%2F"
	return NewProviderWithURI(ctx, uri)
}

// NewProviderWithURI returns a new `WOFProvider` instance configured by
// 'uri' which is expected to be a valid `aaronland/go-uid-proxy` URI.
func NewProviderWithURI(ctx context.Context, uri string) (Provider, error) {

	uid_pr, err := uid.NewProvider(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new provider, %w", err)
	}

	wof_pr := &WOFProvider{
		uid_provider: uid_pr,
	}

	return wof_pr, nil
}

// NewID returns a new Who's On First identifier.
func (wof_pr *WOFProvider) NewID(ctx context.Context) (int64, error) {

	v, err := wof_pr.uid_provider.UID(ctx)

	if err != nil {
		return -1, fmt.Errorf("Failed to generate ID, %w", err)
	}

	id, ok := uid.AsInt64(v)

	if !ok {
		return -1, fmt.Errorf("Provider return invalid value")
	}

	return id, nil
}
