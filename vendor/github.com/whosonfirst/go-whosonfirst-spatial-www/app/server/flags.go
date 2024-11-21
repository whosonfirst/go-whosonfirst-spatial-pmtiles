package server

import (
	"flag"
	"fmt"
	"sort"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
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

var spatial_database_uri string
var properties_reader_uri string
var enable_custom_placetypes bool
var custom_placetypes string

var iterator_uris spatial_flags.MultiCSVIteratorURIFlag

var map_provider_uri string

func DefaultFlagSet() (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("server")

	available_databases := database.Schemes()
	desc_databases := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: %s", available_databases)

	fs.StringVar(&spatial_database_uri, "spatial-database-uri", "rtree://", desc_databases)

	available_readers := reader.Schemes()
	desc_readers := fmt.Sprintf("A valid whosonfirst/go-reader.Reader URI. Available options are: %s", available_readers)

	fs.StringVar(&properties_reader_uri, "properties-reader-uri", "", fmt.Sprintf("%s. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.", desc_readers))

	fs.BoolVar(&enable_custom_placetypes, "enable-custom-placetypes", false, "Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.")

	fs.StringVar(&custom_placetypes, "custom-placetypes", "", "A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.")

	modes := emitter.Schemes()
	sort.Strings(modes)

	desc_iter := spatial_flags.IteratorURIFlagDescription()
	desc_iter = fmt.Sprintf("Zero or more URIs denoting data sources to use for indexing the spatial database at startup. %s", desc_iter)

	fs.Var(&iterator_uris, "iterator-uri", desc_iter)

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

	fs.StringVar(&map_provider_uri, "map-provider-uri", "leaflet://?leaflet-tile-url=https://tile.openstreetmap.org/{z}/{x}/{y}.png", "A valid aaronland/go-http-maps/provider URI.")

	return fs, nil
}
