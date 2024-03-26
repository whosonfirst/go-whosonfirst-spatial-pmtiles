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
	"github.com/aaronland/go-http-tangramjs"
	tilepack_http "github.com/tilezen/go-tilepacks/http"
	"github.com/tilezen/go-tilepacks/tilepack"
)

const TANGRAM_SCHEME string = "tangram"

type TangramProvider struct {
	Provider
	leafletOptions *leaflet.LeafletOptions
	tangramOptions *tangramjs.TangramJSOptions
	tilezenOptions *TilezenOptions
	logger         *log.Logger
}

func init() {

	ctx := context.Background()
	RegisterProvider(ctx, TANGRAM_SCHEME, NewTangramProvider)
}

func TangramJSOptionsFromURL(u *url.URL) (*tangramjs.TangramJSOptions, error) {

	opts := tangramjs.DefaultTangramJSOptions()

	q := u.Query()

	opts.NextzenOptions.APIKey = q.Get("nextzen-apikey")

	q_style_url := q.Get("nextzen-style-url")

	if q_style_url != "" {
		opts.NextzenOptions.StyleURL = q_style_url
	}

	q_tile_url := q.Get("nextzen-tile-url")

	if q_style_url != "" {
		opts.NextzenOptions.TileURL = q_tile_url
	}

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

	return opts, nil
}

func NewTangramProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	leaflet_opts, err := LeafletOptionsFromURL(u)

	if err != nil {
		return nil, fmt.Errorf("Failed to create leaflet options, %w", err)
	}

	tangram_opts, err := TangramJSOptionsFromURL(u)

	if err != nil {
		return nil, fmt.Errorf("Failed to create tilezen options, %w", err)
	}

	tangram_opts.AppendLeafletResources = false
	tangram_opts.AppendLeafletAssets = false

	tilezen_opts, err := TilezenOptionsFromURL(u)

	if err != nil {
		return nil, fmt.Errorf("Failed to create tilezen options, %w", err)
	}

	logger := log.New(io.Discard, "", 0)

	tangram_opts.Logger = logger
	leaflet_opts.Logger = logger

	p := &TangramProvider{
		leafletOptions: leaflet_opts,
		tangramOptions: tangram_opts,
		tilezenOptions: tilezen_opts,
		logger:         logger,
	}

	return p, nil
}

func (p *TangramProvider) Scheme() string {
	return TANGRAM_SCHEME
}

func (p *TangramProvider) AppendResourcesHandler(handler http.Handler) http.Handler {
	handler = leaflet.AppendResourcesHandler(handler, p.leafletOptions)
	handler = tangramjs.AppendResourcesHandler(handler, p.tangramOptions)
	return handler
}

func (p *TangramProvider) AppendAssetHandlers(mux *http.ServeMux) error {

	err := leaflet.AppendAssetHandlers(mux, p.leafletOptions)

	if err != nil {
		return fmt.Errorf("Failed to append leaflet asset handler, %w", err)
	}

	err = tangramjs.AppendAssetHandlers(mux, p.tangramOptions)

	if err != nil {
		return fmt.Errorf("Failed to append tangram asset handler, %w", err)
	}

	if p.tilezenOptions.EnableTilepack {

		tilepack_reader, err := tilepack.NewMbtilesReader(p.tilezenOptions.TilepackPath)

		if err != nil {
			return fmt.Errorf("Failed to create tilepack reader, %w", err)
		}

		tilepack_url := p.tilezenOptions.TilepackURL

		if p.tangramOptions.Prefix != "" {

			tilepack_url, err = url.JoinPath(p.tangramOptions.Prefix, tilepack_url)

			if err != nil {
				return fmt.Errorf("Failed to join path with %s and %s", p.tangramOptions.Prefix, tilepack_url)
			}
		}

		tilepack_handler := tilepack_http.MbtilesHandler(tilepack_reader)
		mux.Handle(tilepack_url, tilepack_handler)
	}

	return nil
}

func (p *TangramProvider) SetLogger(logger *log.Logger) error {
	p.logger = logger
	p.tangramOptions.Logger = logger
	p.leafletOptions.Logger = logger
	return nil
}
