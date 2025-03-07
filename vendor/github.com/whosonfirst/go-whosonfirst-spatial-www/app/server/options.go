package server

import (
	"context"
	"flag"
	"fmt"

	"github.com/sfomuseum/go-flags/flagset"
)

type RunOptions struct {
	AuthenticatorURI       string
	ServerURI              string
	SpatialDatabaseURI     string
	PropertiesReaderURI    string
	IteratorSources        map[string][]string
	EnableCustomPlacetypes bool
	CustomPlacetypes       string
	IsWhosOnFirst          bool
	EnableWWW              bool
	EnableGeoJSON          bool
	EnableGzip             bool
	LogTimings             bool
	EnableCORS             bool
	CORSOrigins            []string
	CORSAllowCredentials   bool
	PathPing               string
	PathAPI                string

	// A string label indicating the map provider to use. Valid options are: leaflet, protomaps.
	MapProvider string
	// A valid Leaflet tile layer URI.
	MapTileURI string
	// A comma-separated string indicating the map's initial view. Valid options are: 'LON,LAT', 'LON,LAT,ZOOM' or 'MINX,MINY,MAXX,MAXY'.
	InitialView string
	// A custom Leaflet style definition for geometries. This may either be a JSON-encoded string or a path on disk.
	LeafletStyle string
	// A custom Leaflet style definition for points. This may either be a JSON-encoded string or a path on disk.
	LeafletPointStyle string
	// A valid Protomaps theme label.
	ProtomapsTheme string
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WHOSONFIRST")

	if err != nil {
		return nil, fmt.Errorf("Failed to set flags from environment variables, %v", err)
	}

	iterator_sources := iterator_uris.AsMap()

	opts := &RunOptions{
		AuthenticatorURI: authenticator_uri,
		ServerURI:        server_uri,

		SpatialDatabaseURI:     spatial_database_uri,
		PropertiesReaderURI:    properties_reader_uri,
		IteratorSources:        iterator_sources,
		EnableCustomPlacetypes: enable_custom_placetypes,
		CustomPlacetypes:       custom_placetypes,

		EnableWWW:     enable_www,
		EnableGeoJSON: enable_geojson,
		EnableGzip:    enable_gzip,
		PathPing:      path_ping,
		PathAPI:       path_api,

		LogTimings:           log_timings,
		EnableCORS:           enable_cors,
		CORSOrigins:          cors_origins,
		CORSAllowCredentials: cors_allow_credentials,

		MapProvider:       map_provider,
		MapTileURI:        map_tile_uri,
		InitialView:       initial_view,
		LeafletStyle:      leaflet_style,
		LeafletPointStyle: leaflet_point_style,
		ProtomapsTheme:    protomaps_theme,
	}

	return opts, nil
}
