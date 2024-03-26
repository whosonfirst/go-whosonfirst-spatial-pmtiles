package provider

import (
	"context"
	"io"
	"log"
	"net/http"
)

const NULL_SCHEME string = "null"

type NullProvider struct {
	Provider
	logger *log.Logger
}

func init() {

	ctx := context.Background()
	RegisterProvider(ctx, NULL_SCHEME, NewNullProvider)
}

func NewNullProvider(ctx context.Context, uri string) (Provider, error) {

	logger := log.New(io.Discard, "", 0)

	p := &NullProvider{
		logger: logger,
	}

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

func (p *NullProvider) SetLogger(logger *log.Logger) error {
	p.logger = logger
	return nil
}
