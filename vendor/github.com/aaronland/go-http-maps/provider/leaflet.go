package provider

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aaronland/go-http-leaflet"
)

const LEAFLET_SCHEME string = "leaflet"

type LeafletProvider struct {
	Provider
	leafletOptions *leaflet.LeafletOptions
	logger         *log.Logger
}

func init() {

	ctx := context.Background()
	RegisterProvider(ctx, LEAFLET_SCHEME, NewLeafletProvider)
}

func LeafletOptionsFromURL(u *url.URL) (*leaflet.LeafletOptions, error) {

	opts := leaflet.DefaultLeafletOptions()

	q := u.Query()

	q_javascript_at_eof := q.Get(JavaScriptAtEOFFlag)

	if q_javascript_at_eof != "" {

		v, err := strconv.ParseBool(q_javascript_at_eof)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?%s= parameter, %w", JavaScriptAtEOFFlag, err)
		}

		if v == true {
			opts.AppendJavaScriptAtEOF = true
		}
	}

	q_rollup_assets := q.Get(RollupAssetsFlag)

	if q_rollup_assets != "" {

		v, err := strconv.ParseBool(q_rollup_assets)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?%s= parameter, %w", RollupAssetsFlag, err)
		}

		if v == true {
			opts.RollupAssets = true
		}
	}

	q_map_prefix := q.Get(MapPrefixFlag)

	if q_map_prefix != "" {
		opts.Prefix = q_map_prefix
	}

	q_enable_hash := q.Get("leaflet-enable-hash")

	if q_enable_hash != "" {

		v, err := strconv.ParseBool(q_enable_hash)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?leaflet-enable-hash= parameter, %w", err)
		}

		if v == true {
			opts.EnableHash()
		}
	}

	q_enable_fullscreen := q.Get("leaflet-enable-fullscreen")

	if q_enable_fullscreen != "" {

		v, err := strconv.ParseBool(q_enable_fullscreen)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?leaflet-enable-fullscreen= parameter, %w", err)
		}

		if v == true {
			opts.EnableFullscreen()
		}
	}

	q_enable_draw := q.Get("leaflet-enable-draw")

	if q_enable_draw != "" {

		v, err := strconv.ParseBool(q_enable_draw)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?leaflet-enable-draw= parameter, %w", err)
		}

		if v == true {
			opts.EnableDraw()
		}
	}

	q_tile_url := q.Get(LeafletTileURLFlag)

	opts.DataAttributes = map[string]string{
		"leaflet-tile-url": q_tile_url,
	}

	return opts, nil
}

func NewLeafletProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	leaflet_opts, err := LeafletOptionsFromURL(u)

	if err != nil {
		return nil, fmt.Errorf("Failed to create leaflet options, %w", err)
	}

	logger := log.New(io.Discard, "", 0)

	leaflet_opts.Logger = logger

	p := &LeafletProvider{
		leafletOptions: leaflet_opts,
		logger:         logger,
	}

	return p, nil
}

func (p *LeafletProvider) Scheme() string {
	return LEAFLET_SCHEME
}

func (p *LeafletProvider) AppendResourcesHandler(handler http.Handler) http.Handler {
	handler = leaflet.AppendResourcesHandler(handler, p.leafletOptions)
	return handler
}

func (p *LeafletProvider) AppendAssetHandlers(mux *http.ServeMux) error {

	err := leaflet.AppendAssetHandlers(mux, p.leafletOptions)

	if err != nil {
		return fmt.Errorf("Failed to append leaflet asset handler, %w", err)
	}

	return nil
}

func (p *LeafletProvider) SetLogger(logger *log.Logger) error {
	p.logger = logger
	p.leafletOptions.Logger = logger
	return nil
}
