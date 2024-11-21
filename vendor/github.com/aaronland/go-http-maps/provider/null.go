package provider

import (
	"context"
	"net/http"
)

const NULL_SCHEME string = "null"

type NullProvider struct {
	Provider
}

func init() {

	ctx := context.Background()
	RegisterProvider(ctx, NULL_SCHEME, NewNullProvider)
}

func NewNullProvider(ctx context.Context, uri string) (Provider, error) {

	p := &NullProvider{}
	return p, nil
}

func (p *NullProvider) Scheme() string {
	return NULL_SCHEME
}

func (p *NullProvider) AppendResourcesHandler(handler http.Handler) http.Handler {
	return handler
}

func (p *NullProvider) AppendAssetHandlers(mux *http.ServeMux) error {

	return nil
}
