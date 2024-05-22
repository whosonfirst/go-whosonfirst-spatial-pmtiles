package server

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	gohttp "net/http"
	"path/filepath"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/aaronland/go-http-bootstrap"
	"github.com/aaronland/go-http-maps"
	"github.com/aaronland/go-http-maps/provider"
	"github.com/aaronland/go-http-ping/v2"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/sfomuseum/go-http-auth"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/http"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/http/api"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/http/www"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/templates/html"
	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
)

func Run(ctx context.Context, logger *slog.Logger) error {

	fs, err := DefaultFlagSet()

	if err != nil {
		return fmt.Errorf("Failed to derive default flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *slog.Logger) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to derive options from flag set, %w", err)
	}

	return RunWithOptions(ctx, opts, logger)
}

func RunWithOptions(ctx context.Context, opts *RunOptions, logger *slog.Logger) error {

	spatial_opts := &app.SpatialApplicationOptions{
		SpatialDatabaseURI:     opts.SpatialDatabaseURI,
		PropertiesReaderURI:    opts.PropertiesReaderURI,
		IteratorURI:            opts.IteratorURI,
		EnableCustomPlacetypes: opts.EnableCustomPlacetypes,
		CustomPlacetypes:       opts.CustomPlacetypes,
		IsWhosOnFirst:          opts.IsWhosOnFirst,
	}

	spatial_app, err := app.NewSpatialApplication(ctx, spatial_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new spatial application, %w", err)
	}

	authenticator, err := auth.NewAuthenticator(ctx, opts.AuthenticatorURI)

	if err != nil {
		return fmt.Errorf("Failed to create authenticator, %w", err)
	}

	if len(opts.IteratorSources) > 0 {

		err = spatial_app.IndexPaths(ctx, opts.IteratorSources...)

		if err != nil {
			return fmt.Errorf("Failed to index paths, because %s", err)
		}
	}

	mux := gohttp.NewServeMux()

	ping_handler, err := ping.PingPongHandler()

	if err != nil {
		return fmt.Errorf("failed to create ping handler because %s", err)
	}

	mux.Handle(opts.PathPing, ping_handler)

	var cors_wrapper *cors.Cors

	if opts.EnableCORS {
		cors_wrapper = cors.New(cors.Options{
			AllowedOrigins:   opts.CORSOrigins,
			AllowCredentials: opts.CORSAllowCredentials,
		})
	}

	// data (geojson) handlers
	// SpatialDatabase implements reader.Reader

	data_handler, err := api.NewDataHandler(spatial_app.SpatialDatabase)

	if err != nil {
		return fmt.Errorf("Failed to create data handler, %v", err)
	}

	data_handler = http.CheckIndexingHandler(spatial_app, data_handler)

	data_handler = authenticator.WrapHandler(data_handler)

	if opts.EnableCORS {
		data_handler = cors_wrapper.Handler(data_handler)
	}

	if opts.EnableGzip {
		data_handler = gziphandler.GzipHandler(data_handler)
	}

	if !strings.HasSuffix(opts.PathData, "/") {
		opts.PathData = fmt.Sprintf("%s/", opts.PathData)
	}

	mux.Handle(opts.PathData, data_handler)

	// point-in-polygon handlers

	api_pip_opts := &api.PointInPolygonHandlerOptions{
		EnableGeoJSON: opts.EnableGeoJSON,
		LogTimings:    opts.LogTimings,
	}

	api_pip_handler, err := api.PointInPolygonHandler(spatial_app, api_pip_opts)

	if err != nil {
		return fmt.Errorf("failed to create point-in-polygon handler because %s", err)
	}

	api_pip_handler = authenticator.WrapHandler(api_pip_handler)

	if opts.EnableCORS {
		api_pip_handler = cors_wrapper.Handler(api_pip_handler)
	}

	if opts.EnableGzip {
		api_pip_handler = gziphandler.GzipHandler(api_pip_handler)
	}

	path_api_pip := filepath.Join(opts.PathAPI, "point-in-polygon")

	mux.Handle(path_api_pip, api_pip_handler)

	// www handlers

	if opts.EnableWWW {

		map_provider, err := provider.NewProvider(ctx, opts.MapProviderURI)

		if err != nil {
			return fmt.Errorf("Failed to create map provider, %w", err)
		}

		err = map_provider.AppendAssetHandlers(mux)

		if err != nil {
			return fmt.Errorf("Failed to append map provider asset handlers, %w", err)
		}

		t := template.New("spatial")

		t = t.Funcs(map[string]interface{}{

			"EnsureRoot": func(path string) string {

				path = strings.TrimLeft(path, "/")

				if opts.PathPrefix == "" {
					return "/" + path
				}

				path = filepath.Join(opts.PathPrefix, path)
				return path
			},

			"DataRoot": func() string {

				path := opts.PathData

				if opts.PathPrefix != "" {
					path = filepath.Join(opts.PathPrefix, path)
				}

				return path
			},

			"APIRoot": func() string {

				path := opts.PathAPI

				if opts.PathPrefix != "" {
					path = filepath.Join(opts.PathPrefix, path)
				}

				return path
			},
		})

		t, err = t.ParseFS(html.FS, "*.html")

		if err != nil {
			return fmt.Errorf("Unable to parse templates, %v", err)
		}

		bootstrap_opts := bootstrap.DefaultBootstrapOptions()

		err = bootstrap.AppendAssetHandlers(mux, bootstrap_opts)

		if err != nil {
			return fmt.Errorf("Failed to append bootstrap assets, %v", err)
		}

		err = www.AppendStaticAssetHandlers(mux)

		if err != nil {
			return fmt.Errorf("Failed to append static assets, %v", err)
		}

		// point-in-polygon page

		http_pip_opts := &www.PointInPolygonHandlerOptions{
			Templates:        t,
			InitialLatitude:  opts.LeafletInitialLatitude,
			InitialLongitude: opts.LeafletInitialLongitude,
			InitialZoom:      opts.LeafletInitialZoom,
			MaxBounds:        opts.LeafletMaxBounds,
			MapProvider:      map_provider.Scheme(),
		}

		http_pip_handler, err := www.PointInPolygonHandler(spatial_app, http_pip_opts)

		if err != nil {
			return fmt.Errorf("failed to create (bundled) www handler because %s", err)
		}

		maps_opts := maps.DefaultMapsOptions()

		err = maps.AppendAssetHandlers(mux, maps_opts)

		if err != nil {
			return fmt.Errorf("Failed to append map assets, %w", err)
		}

		http_pip_handler = bootstrap.AppendResourcesHandler(http_pip_handler, bootstrap_opts)
		http_pip_handler = maps.AppendResourcesHandlerWithProvider(http_pip_handler, map_provider, maps_opts)
		http_pip_handler = authenticator.WrapHandler(http_pip_handler)

		mux.Handle(opts.PathPIP, http_pip_handler)

		if !strings.HasSuffix(opts.PathPIP, "/") {
			path_pip_slash := fmt.Sprintf("%s/", opts.PathPIP)
			mux.Handle(path_pip_slash, http_pip_handler)
		}

		// index / splash page

		index_opts := &www.IndexHandlerOptions{
			Templates: t,
		}

		index_handler, err := www.IndexHandler(index_opts)

		if err != nil {
			return fmt.Errorf("Failed to create index handler, %v", err)
		}

		index_handler = bootstrap.AppendResourcesHandler(index_handler, bootstrap_opts)
		index_handler = authenticator.WrapHandler(index_handler)

		path_index := "/"

		mux.Handle(path_index, index_handler)
	}

	s, err := server.NewServer(ctx, opts.ServerURI)

	if err != nil {
		return fmt.Errorf("Failed to create new server for '%s', %v", server_uri, err)
	}

	logger.Info("Listening for requests", "address", s.Address())

	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		return fmt.Errorf("Failed to start server, %v", err)
	}

	return nil
}
