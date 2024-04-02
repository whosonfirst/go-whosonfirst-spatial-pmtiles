package server

import (
	"flag"
	"fmt"

	"github.com/aaronland/go-http-maps/provider"
	"github.com/sfomuseum/go-flags/multi"
	spatial_flags "github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

// Prepend this prefix to all assets (but not HTTP handlers). This is mostly for API Gateway integrations.
var path_prefix string

// The root URL for all API handlers
var path_api string

// The URL for the ping (health check) handler
var path_ping string

// The URL for the point in polygon web handler
var path_pip string

// The URL for data (GeoJSON) handler
var path_data string

// Enable the interactive /debug endpoint to query points and display results.
var enable_www bool

// Enable GeoJSON output for point-in-polygon API calls.
var enable_geojson bool

// Enable CORS headers for data-related and API handlers.
var enable_cors bool

// Allow HTTP credentials to be included in CORS requests.
var cors_allow_credentials bool

// One or more hosts to allow CORS requests from; may be a comma-separated list.
var cors_origins multi.MultiCSVString

// Enable gzip-encoding for data-related and API handlers.
var enable_gzip bool

// The initial latitude for map views to use.
var leaflet_initial_latitude float64

// The initial longitude for map views to use.
var leaflet_initial_longitude float64

// The initial zoom level for map views to use.
var leaflet_initial_zoom int

// An optional comma-separated bounding box ({MINX},{MINY},{MAXX},{MAXY}) to set the boundary for map views.
var leaflet_max_bounds string

// A valid aaronland/go-http-server URI.
var server_uri string

// A valid sfomuseum/go-http-auth URI.
var authenticator_uri string

// Emit timing metrics to the application's logger
var log_timings bool

func DefaultFlagSet() (*flag.FlagSet, error) {

	fs, err := spatial_flags.CommonFlags()

	if err != nil {
		return nil, fmt.Errorf("Failed to derive common spatial flags, %w", err)
	}

	err = spatial_flags.AppendIndexingFlags(fs)

	if err != nil {
		return nil, fmt.Errorf("Failed to append spatial indexing flags, %w", err)
	}

	err = AppendWWWFlags(fs)

	if err != nil {
		return nil, fmt.Errorf("Failed to append www flags, %w", err)
	}

	err = provider.AppendProviderFlags(fs)

	if err != nil {
		return nil, fmt.Errorf("Failed to append map provider flags, %w", err)
	}

	return fs, nil
}

func AppendWWWFlags(fs *flag.FlagSet) error {

	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI.")

	fs.StringVar(&authenticator_uri, "authenticator-uri", "null://", "A valid sfomuseum/go-http-auth URI.")

	fs.BoolVar(&enable_www, "enable-www", false, "Enable the interactive /debug endpoint to query points and display results.")

	fs.BoolVar(&enable_geojson, "enable-geojson", false, "Enable GeoJSON output for point-in-polygon API calls.")

	fs.BoolVar(&enable_cors, "enable-cors", false, "Enable CORS headers for data-related and API handlers.")
	fs.BoolVar(&cors_allow_credentials, "cors-allow-credentials", false, "Allow HTTP credentials to be included in CORS requests.")

	fs.Var(&cors_origins, "cors-origin", "One or more hosts to allow CORS requests from; may be a comma-separated list.")

	fs.BoolVar(&enable_gzip, "enable-gzip", false, "Enable gzip-encoding for data-related and API handlers.")

	fs.StringVar(&path_prefix, "path-prefix", "", "Prepend this prefix to all assets (but not HTTP handlers). This is mostly for API Gateway integrations.")

	fs.StringVar(&path_api, "path-api", "/api", "The root URL for all API handlers")
	fs.StringVar(&path_ping, "path-ping", "/health/ping", "The URL for the ping (health check) handler")
	fs.StringVar(&path_pip, "path-pip", "/point-in-polygon", "The URL for the point in polygon web handler")
	fs.StringVar(&path_data, "path-data", "/data", "The URL for data (GeoJSON) handler")

	fs.Float64Var(&leaflet_initial_latitude, "leaflet-initial-latitude", 37.616906, "The initial latitude for map views to use.")
	fs.Float64Var(&leaflet_initial_longitude, "leaflet-initial-longitude", -122.386665, "The initial longitude for map views to use.")
	fs.IntVar(&leaflet_initial_zoom, "leaflet-initial-zoom", 14, "The initial zoom level for map views to use.")
	fs.StringVar(&leaflet_max_bounds, "leaflet-max-bounds", "", "An optional comma-separated bounding box ({MINX},{MINY},{MAXX},{MAXY}) to set the boundary for map views.")

	fs.BoolVar(&log_timings, "log-timings", false, "Emit timing metrics to the application's logger")
	return nil
}
