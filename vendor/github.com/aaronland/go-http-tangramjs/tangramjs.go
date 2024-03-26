package tangramjs

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aaronland/go-http-leaflet"
	aa_static "github.com/aaronland/go-http-static"
	"github.com/aaronland/go-http-tangramjs/static"
	"github.com/sfomuseum/go-http-rollup"
)

// NEXTZEN_MVT_ENDPOINT is the default endpoint for Nextzen vector tiles
const NEXTZEN_MVT_ENDPOINT string = "https://tile.nextzen.org/tilezen/vector/v1/512/all/{z}/{x}/{y}.mvt"

// NextzenOptions provides configuration variables for Nextzen map tiles.
type NextzenOptions struct {
	// A valid Nextzen developer API key
	APIKey string
	// The URL for a valid Tangram.js style.
	StyleURL string
	// The URL template to use for fetching Nextzen map tiles.
	TileURL string
}

// TangramJSOptions provides a list of JavaScript and CSS link to include with HTML output as well as options for Nextzen tiles and Leaflet.js.
type TangramJSOptions struct {
	// A list of Tangram.js Javascript files to append to HTML resources.
	JS []string
	// A list of Tangram.js CSS files to append to HTML resources.
	CSS []string
	// A NextzenOptions instance.
	NextzenOptions *NextzenOptions
	// A leaflet.LeafletOptions instance.
	LeafletOptions *leaflet.LeafletOptions
	// AppendJavaScriptAtEOF is a boolean flag to append JavaScript markup at the end of an HTML document
	// rather than in the <head> HTML element. Default is false
	AppendJavaScriptAtEOF bool
	RollupAssets          bool
	Prefix                string
	Logger                *log.Logger
	// By default the go-http-tangramjs package will also include and reference Leaflet.js resources using the aaronland/go-http-leaflet package. If you want or need to disable this behaviour set this variable to false.
	AppendLeafletResources bool
	// By default the go-http-tangramjs package will also include and reference Leaflet.js assets using the aaronland/go-http-leaflet package. If you want or need to disable this behaviour set this variable to false.
	AppendLeafletAssets bool
}

// Return a *NextzenOptions struct with default values.
func DefaultNextzenOptions() *NextzenOptions {

	opts := &NextzenOptions{
		APIKey:   "",
		StyleURL: "",
		TileURL:  NEXTZEN_MVT_ENDPOINT,
	}

	return opts
}

// Return a *TangramJSOptions struct with default values.
func DefaultTangramJSOptions() *TangramJSOptions {

	logger := log.New(io.Discard, "", 0)

	leaflet_opts := leaflet.DefaultLeafletOptions()
	nextzen_opts := DefaultNextzenOptions()

	opts := &TangramJSOptions{
		CSS: []string{},
		JS: []string{
			"/javascript/tangram.min.js",
		},
		LeafletOptions:         leaflet_opts,
		NextzenOptions:         nextzen_opts,
		Logger:                 logger,
		AppendLeafletResources: true,
		AppendLeafletAssets:    true,
	}

	return opts
}

// AppendResourcesHandler will rewrite any HTML produced by previous handler to include the necessary markup to load Tangram.js files and related assets.
func AppendResourcesHandler(next http.Handler, opts *TangramJSOptions) http.Handler {

	if opts.AppendLeafletResources {
		opts.LeafletOptions.AppendJavaScriptAtEOF = opts.AppendJavaScriptAtEOF
		opts.LeafletOptions.RollupAssets = opts.RollupAssets
		opts.LeafletOptions.Prefix = opts.Prefix
		opts.LeafletOptions.Logger = opts.Logger

		next = leaflet.AppendResourcesHandler(next, opts.LeafletOptions)
	}

	attrs := map[string]string{
		"nextzen-api-key":   opts.NextzenOptions.APIKey,
		"nextzen-style-url": opts.NextzenOptions.StyleURL,
		"nextzen-tile-url":  opts.NextzenOptions.TileURL,
	}

	static_opts := aa_static.DefaultResourcesOptions()
	static_opts.DataAttributes = attrs
	static_opts.AppendJavaScriptAtEOF = opts.AppendJavaScriptAtEOF

	js_uris := opts.JS
	css_uris := opts.CSS

	if opts.RollupAssets {

		if len(opts.JS) > 1 {
			js_uris = []string{
				"/javascript/tangramjs.rollup.js",
			}
		}

		if len(opts.CSS) > 1 {
			css_uris = []string{
				"/css/tangramjs.rollup.css",
			}
		}
	}

	static_opts.JS = js_uris
	static_opts.CSS = css_uris

	return aa_static.AppendResourcesHandlerWithPrefix(next, static_opts, opts.Prefix)
}

// Append all the files in the net/http FS instance containing the embedded Tangram.js assets to an *http.ServeMux instance.
func AppendAssetHandlers(mux *http.ServeMux, opts *TangramJSOptions) error {

	if opts.AppendLeafletAssets {

		opts.LeafletOptions.AppendJavaScriptAtEOF = opts.AppendJavaScriptAtEOF
		opts.LeafletOptions.RollupAssets = opts.RollupAssets
		opts.LeafletOptions.Prefix = opts.Prefix
		opts.LeafletOptions.Logger = opts.Logger

		err := leaflet.AppendAssetHandlers(mux, opts.LeafletOptions)

		if err != nil {
			return fmt.Errorf("Failed to append Leaflet assets, %w", err)
		}
	}

	if !opts.RollupAssets {
		return aa_static.AppendStaticAssetHandlersWithPrefix(mux, static.FS, opts.Prefix)
	}

	// START OF make sure we are still serving bundled Tangram styles

	err := serveSubDir(mux, opts, "tangram")

	if err != nil {
		return fmt.Errorf("Failed to append static asset handler for tangram FS, %w", err)
	}

	// END OF make sure we are still serving bundled Tangram styles

	// START OF this should eventually be made a generic function in go-http-rollup

	js_paths := make([]string, len(opts.JS))
	css_paths := make([]string, len(opts.CSS))

	for idx, path := range opts.JS {
		path = strings.TrimLeft(path, "/")
		js_paths[idx] = path
	}

	for idx, path := range opts.CSS {
		path = strings.TrimLeft(path, "/")
		css_paths[idx] = path
	}

	switch len(js_paths) {
	case 0:
		// pass
	case 1:
		err := serveSubDir(mux, opts, "javascript")

		if err != nil {
			return fmt.Errorf("Failed to append static asset handler for javascript FS, %w", err)
		}

	default:

		rollup_js_paths := map[string][]string{
			"tangramjs.rollup.js": js_paths,
		}

		rollup_js_opts := &rollup.RollupJSHandlerOptions{
			FS:     static.FS,
			Paths:  rollup_js_paths,
			Logger: opts.Logger,
		}

		rollup_js_handler, err := rollup.RollupJSHandler(rollup_js_opts)

		if err != nil {
			return fmt.Errorf("Failed to create rollup JS handler, %w", err)
		}

		rollup_js_uri := "/javascript/tangramjs.rollup.js"

		if opts.Prefix != "" {

			u, err := url.JoinPath(opts.Prefix, rollup_js_uri)

			if err != nil {
				return fmt.Errorf("Failed to append prefix to %s, %w", rollup_js_uri, err)
			}

			rollup_js_uri = u
		}

		mux.Handle(rollup_js_uri, rollup_js_handler)
	}

	// CSS

	switch len(css_paths) {
	case 0:
		// pass
	case 1:

		err := serveSubDir(mux, opts, "css")

		if err != nil {
			return fmt.Errorf("Failed to append static asset handler for css FS, %w", err)
		}

	default:

		rollup_css_paths := map[string][]string{
			"tangramjs.rollup.css": css_paths,
		}

		rollup_css_opts := &rollup.RollupCSSHandlerOptions{
			FS:     static.FS,
			Paths:  rollup_css_paths,
			Logger: opts.Logger,
		}

		rollup_css_handler, err := rollup.RollupCSSHandler(rollup_css_opts)

		if err != nil {
			return fmt.Errorf("Failed to create rollup CSS handler, %w", err)
		}

		rollup_css_uri := "/css/tangramjs.rollup.css"

		if opts.Prefix != "" {

			u, err := url.JoinPath(opts.Prefix, rollup_css_uri)

			if err != nil {
				return fmt.Errorf("Failed to append prefix to %s, %w", rollup_css_uri, err)
			}

			rollup_css_uri = u
		}

		mux.Handle(rollup_css_uri, rollup_css_handler)
	}

	// END OF this should eventually be made a generic function in go-http-rollup

	return nil
}

func serveSubDir(mux *http.ServeMux, opts *TangramJSOptions, dirname string) error {

	sub_fs, err := fs.Sub(static.FS, dirname)

	if err != nil {
		return fmt.Errorf("Failed to load %s FS, %w", dirname, err)
	}

	sub_prefix := dirname

	if opts.Prefix != "" {

		prefix, err := url.JoinPath(opts.Prefix, sub_prefix)

		if err != nil {
			return fmt.Errorf("Failed to append prefix to %s, %w", sub_prefix, err)
		}

		sub_prefix = prefix
	}

	err = aa_static.AppendStaticAssetHandlersWithPrefix(mux, sub_fs, sub_prefix)

	if err != nil {
		return fmt.Errorf("Failed to append static asset handler for %s FS, %w", dirname, err)
	}

	return nil
}
