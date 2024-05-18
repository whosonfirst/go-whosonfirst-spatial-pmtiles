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
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-http-auth"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/http"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/http/api"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/http/www"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/templates/html"
	app "github.com/whosonfirst/go-whosonfirst-spatial/app/spatial"
	spatial_flags "github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

type RunOptions struct {
	FlagSet       *flag.FlagSet
	EnvFlagPrefix string
}

func Run(ctx context.Context, logger *slog.Logger) error {

	fs, err := DefaultFlagSet()

	if err != nil {
		return fmt.Errorf("Failed to derive default flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *slog.Logger) error {

	opts := &RunOptions{
		FlagSet:       fs,
		EnvFlagPrefix: "WHOSONFIRST",
	}

	return RunWithOptions(ctx, opts, logger)
}

func RunWithOptions(ctx context.Context, opts *RunOptions, logger *slog.Logger) error {

	fs := opts.FlagSet

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, opts.EnvFlagPrefix, true)

	if err != nil {
		return fmt.Errorf("Failed to set flags from environment variables, %v", err)
	}

	err = spatial_flags.ValidateCommonFlags(fs)

	if err != nil {
		return fmt.Errorf("Failed to validate common flags, %v", err)
	}

	err = spatial_flags.ValidateIndexingFlags(fs)

	if err != nil {
		return fmt.Errorf("Failed to validate indexing flags, %v", err)
	}

	spatial_app, err := app.NewSpatialApplicationWithFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to create new spatial application, because: %w", err)
	}

	authenticator, err := auth.NewAuthenticator(ctx, authenticator_uri)

	if err != nil {
		return fmt.Errorf("Failed to create authenticator, %w", err)
	}

	paths := fs.Args()

	err = spatial_app.IndexPaths(ctx, paths...)

	if err != nil {
		return fmt.Errorf("Failed to index paths, because %s", err)
	}

	mux := gohttp.NewServeMux()

	ping_handler, err := ping.PingPongHandler()

	if err != nil {
		return fmt.Errorf("failed to create ping handler because %s", err)
	}

	mux.Handle(path_ping, ping_handler)

	var cors_wrapper *cors.Cors

	if enable_cors {
		cors_wrapper = cors.New(cors.Options{
			AllowedOrigins:   cors_origins,
			AllowCredentials: cors_allow_credentials,
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

	if enable_cors {
		data_handler = cors_wrapper.Handler(data_handler)
	}

	if enable_gzip {
		data_handler = gziphandler.GzipHandler(data_handler)
	}

	if !strings.HasSuffix(path_data, "/") {
		path_data = fmt.Sprintf("%s/", path_data)
	}

	mux.Handle(path_data, data_handler)

	// point-in-polygon handlers

	api_pip_opts := &api.PointInPolygonHandlerOptions{
		EnableGeoJSON: enable_geojson,
		LogTimings:    log_timings,
	}

	api_pip_handler, err := api.PointInPolygonHandler(spatial_app, api_pip_opts)

	if err != nil {
		return fmt.Errorf("failed to create point-in-polygon handler because %s", err)
	}

	api_pip_handler = authenticator.WrapHandler(api_pip_handler)

	if enable_cors {
		api_pip_handler = cors_wrapper.Handler(api_pip_handler)
	}

	if enable_gzip {
		api_pip_handler = gziphandler.GzipHandler(api_pip_handler)
	}

	path_api_pip := filepath.Join(path_api, "point-in-polygon")

	mux.Handle(path_api_pip, api_pip_handler)

	// www handlers

	if enable_www {

		provider_uri, err := provider.ProviderURIFromFlagSet(fs)

		if err != nil {
			return fmt.Errorf("Failed to derive map provider URI, %w", err)
		}

		map_provider, err := provider.NewProvider(ctx, provider_uri)

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

				if path_prefix == "" {
					return "/" + path
				}

				path = filepath.Join(path_prefix, path)
				return path
			},

			"DataRoot": func() string {

				path := path_data

				if path_prefix != "" {
					path = filepath.Join(path_prefix, path)
				}

				return path
			},

			"APIRoot": func() string {

				path := path_api

				if path_prefix != "" {
					path = filepath.Join(path_prefix, path)
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
			InitialLatitude:  leaflet_initial_latitude,
			InitialLongitude: leaflet_initial_longitude,
			InitialZoom:      leaflet_initial_zoom,
			MaxBounds:        leaflet_max_bounds,
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

		mux.Handle(path_pip, http_pip_handler)

		if !strings.HasSuffix(path_pip, "/") {
			path_pip_slash := fmt.Sprintf("%s/", path_pip)
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

	s, err := server.NewServer(ctx, server_uri)

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
