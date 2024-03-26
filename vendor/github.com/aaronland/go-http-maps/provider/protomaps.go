package provider

import (
	_ "gocloud.dev/blob/fileblob"
)

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-maps/templates/javascript"
	aa_static "github.com/aaronland/go-http-static"
	"github.com/protomaps/go-pmtiles/pmtiles"
	"github.com/sfomuseum/go-http-protomaps"
	pmhttp "github.com/sfomuseum/go-sfomuseum-pmtiles/http"
	"github.com/sfomuseum/runtimevar"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

const PROTOMAPS_SCHEME string = "protomaps"

const pathRulesJavascript string = "/javascript/aaronland.protomaps.rules.js"

type ProtomapsProvider struct {
	Provider
	leafletOptions   *leaflet.LeafletOptions
	protomapsOptions *protomaps.ProtomapsOptions
	paintRules       string
	labelRules       string
	rulesTemplate    *template.Template
	logger           *log.Logger
	serve_tiles      bool
	cache_size       int
	bucket_uri       string
	path_tiles       string
	database         string
}

func init() {
	ctx := context.Background()
	RegisterProvider(ctx, PROTOMAPS_SCHEME, NewProtomapsProvider)
}

func ProtomapsOptionsFromURL(u *url.URL) (*protomaps.ProtomapsOptions, error) {

	opts := protomaps.DefaultProtomapsOptions()

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

	q_tile_url := q.Get(ProtomapsTileURLFlag)
	opts.TileURL = q_tile_url

	return opts, nil
}

func NewProtomapsProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	leaflet_opts, err := LeafletOptionsFromURL(u)

	if err != nil {
		return nil, fmt.Errorf("Failed to create leaflet options, %w", err)
	}

	protomaps_opts, err := ProtomapsOptionsFromURL(u)

	if err != nil {
		return nil, fmt.Errorf("Failed to create protomaps options, %w", err)
	}

	protomaps_opts.AppendLeafletResources = false
	protomaps_opts.AppendLeafletAssets = false

	t, err := javascript.LoadTemplates(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to load Javascript templates, %w", err)
	}

	rules_t := t.Lookup("rules")

	if t == nil {
		return nil, fmt.Errorf("Missing 'rules' Javascript template")
	}

	logger := log.New(io.Discard, "", 0)

	protomaps_opts.Logger = logger
	leaflet_opts.Logger = logger

	p := &ProtomapsProvider{
		leafletOptions:   leaflet_opts,
		protomapsOptions: protomaps_opts,
		logger:           logger,
		rulesTemplate:    rules_t,
	}

	custom_paint_uri := q.Get(ProtomapsPaintRulesURIFlag)
	custom_labels_uri := q.Get(ProtomapsLabelRulesURIFlag)

	if custom_paint_uri != "" {

		paint_rules, err := runtimevar.StringVar(ctx, custom_paint_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive custom paint rules from %s= query parameter, %w", ProtomapsPaintRulesURIFlag, err)
		}

		p.paintRules = paint_rules
	}

	if custom_labels_uri != "" {

		label_rules, err := runtimevar.StringVar(ctx, custom_labels_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive custom label rules from %s= query parameter, %w", ProtomapsLabelRulesURIFlag, err)
		}

		p.labelRules = label_rules
	}

	serve_tiles := false

	q_serve_tiles := q.Get(ProtomapsServeTilesFlag)

	if q_serve_tiles != "" {

		v, err := strconv.ParseBool(q_serve_tiles)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?protomaps-serve-tiles= parameter, %w")
		}

		serve_tiles = v
	}

	if serve_tiles {

		q_cache_size := q.Get(ProtomapsCacheSizeFlag)
		q_bucket_uri := q.Get(ProtomapsBucketURIFlag)
		q_database := q.Get(ProtomapsDatabaseFlag)

		sz, err := strconv.Atoi(q_cache_size)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?%s= parameter, %w", ProtomapsCacheSizeFlag, err)
		}

		q_tile_url := q.Get(ProtomapsTileURLFlag)

		p.cache_size = sz
		p.bucket_uri = q_bucket_uri
		p.database = q_database
		p.path_tiles = q_tile_url
		p.serve_tiles = true
	}

	return p, nil
}

func (p *ProtomapsProvider) Scheme() string {
	return PROTOMAPS_SCHEME
}

func (p *ProtomapsProvider) AppendResourcesHandler(handler http.Handler) http.Handler {

	handler = leaflet.AppendResourcesHandler(handler, p.leafletOptions)
	handler = protomaps.AppendResourcesHandler(handler, p.protomapsOptions)

	// Finally add the custom paint/label rules. These are served by the
	// rulesHandler() method below.

	static_opts := aa_static.DefaultResourcesOptions()
	static_opts.AppendJavaScriptAtEOF = p.protomapsOptions.AppendJavaScriptAtEOF
	static_opts.JS = []string{pathRulesJavascript}

	return aa_static.AppendResourcesHandlerWithPrefix(handler, static_opts, p.protomapsOptions.Prefix)
	return handler
}

func (p *ProtomapsProvider) AppendAssetHandlers(mux *http.ServeMux) error {

	err := leaflet.AppendAssetHandlers(mux, p.leafletOptions)

	if err != nil {
		return fmt.Errorf("Failed to append leaflet asset handler, %w", err)
	}

	err = protomaps.AppendAssetHandlers(mux, p.protomapsOptions)

	if err != nil {
		return fmt.Errorf("Failed to append protomaps asset handler, %w", err)
	}

	if p.serve_tiles {

		loop, err := pmtiles.NewServer(p.bucket_uri, "", p.logger, p.cache_size, "", "")

		if err != nil {
			return fmt.Errorf("Failed to create pmtiles.Loop, %w", err)
		}

		loop.Start()

		path_tiles := p.path_tiles

		if p.protomapsOptions.Prefix != "" {

			path_tiles, err = url.JoinPath(p.protomapsOptions.Prefix, path_tiles)

			if err != nil {
				return fmt.Errorf("Failed to join path with %s and %s", p.protomapsOptions.Prefix, path_tiles)
			}
		}

		pmtiles_handler := pmhttp.TileHandler(loop, p.logger)

		strip_path := strings.TrimRight(path_tiles, "/")
		pmtiles_handler = http.StripPrefix(strip_path, pmtiles_handler)

		mux.Handle(path_tiles, pmtiles_handler)

		// Because inevitably I will forget...
		protomaps_tiles_database := strings.Replace(p.database, ".pmtiles", "", 1)

		// Note: We are NOT using the local path_tiles because that will have the prefix
		// assigned by AppendResourcesHandlerWithPrefix

		pm_tile_url, err := url.JoinPath(p.path_tiles, protomaps_tiles_database)

		if err != nil {
			return fmt.Errorf("Failed to join path to derive Protomaps tile URL, %w", err)
		}

		pm_tile_url = fmt.Sprintf("%s/{z}/{x}/{y}.mvt", pm_tile_url)

		p.protomapsOptions.TileURL = pm_tile_url
	}

	err = p.appendRulesAssetHandlers(mux, p.protomapsOptions)

	if err != nil {
		return fmt.Errorf("Failed to assign rules asset handlers, %w", err)
	}

	return nil
}

func (p *ProtomapsProvider) SetLogger(logger *log.Logger) error {
	p.logger = logger
	p.protomapsOptions.Logger = logger
	p.leafletOptions.Logger = logger
	return nil
}

func (p *ProtomapsProvider) appendRulesAssetHandlers(mux *http.ServeMux, opts *protomaps.ProtomapsOptions) error {

	rules_handler, err := p.rulesHandler()

	if err != nil {
		return fmt.Errorf("Failed to create rules handler, %w", err)
	}

	// START OF middleware to minify rules

	if opts.RollupAssets {

		js_regexp, err := regexp.Compile("^(application|text)/(x-)?(java|ecma)script$")

		if err != nil {
			return fmt.Errorf("Failed to compile JS pattern for minifier, %w", err)
		}

		m := minify.New()
		m.AddFuncRegexp(js_regexp, js.Minify)

		rules_handler = m.Middleware(rules_handler)
	}

	// END OF middleware to minify rules

	path_rules := pathRulesJavascript

	if opts.Prefix != "" {

		path, err := url.JoinPath(opts.Prefix, path_rules)

		if err != nil {
			return fmt.Errorf("Failed to join path for paint rules, %w", err)
		}

		path_rules = path
	}

	mux.Handle(path_rules, rules_handler)
	return nil
}

func (p *ProtomapsProvider) rulesHandler() (http.Handler, error) {

	type ProtomapsRulesVars struct {
		PaintRules string
		LabelRules string
	}

	vars := ProtomapsRulesVars{
		PaintRules: p.paintRules,
		LabelRules: p.labelRules,
	}

	t := p.rulesTemplate

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		rsp.Header().Set("Content-type", "text/javascript")

		err := t.Execute(rsp, vars)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
