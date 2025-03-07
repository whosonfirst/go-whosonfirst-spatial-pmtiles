package server

import (
	"flag"
	"fmt"
	"sort"

	"github.com/aaronland/go-http-maps/v2"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	spatial_flags "github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

// The root URL for all API handlers
var path_api string

// The URL for the ping (health check) handler
var path_ping string

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

var initial_view string
var map_provider string
var map_tile_uri string
var protomaps_theme string
var leaflet_style string
var leaflet_point_style string

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

	fs.StringVar(&path_api, "path-api", "/api", "The root URL for all API handlers")
	fs.StringVar(&path_ping, "path-ping", "/health/ping", "The URL for the ping (health check) handler")

	fs.BoolVar(&log_timings, "log-timings", false, "Emit timing metrics to the application's logger")

	fs.StringVar(&map_provider, "map-provider", "leaflet", "Valid options are: leaflet, protomaps")
	fs.StringVar(&map_tile_uri, "map-tile-uri", maps.LEAFLET_OSM_TILE_URL, "A valid Leaflet tile layer URI. See documentation for special-case (interpolated tile) URIs.")
	fs.StringVar(&protomaps_theme, "protomaps-theme", "white", "A valid Protomaps theme label.")
	fs.StringVar(&leaflet_style, "leaflet_style", "", "A custom Leaflet style definition for geometries. This may either be a JSON-encoded string or a path on disk.")
	fs.StringVar(&leaflet_point_style, "leaflet_point_style", "", "A custom Leaflet style definition for points. This may either be a JSON-encoded string or a path on disk.")
	fs.StringVar(&initial_view, "initial-view", "", "A comma-separated string indicating the map's initial view. Valid options are: 'LON,LAT', 'LON,LAT,ZOOM' or 'MINX,MINY,MAXX,MAXY'.")

	return fs, nil
}
