package server

import (
	"context"
	"flag"
	"fmt"

	"github.com/sfomuseum/go-flags/flagset"
)

type RunOptions struct {
	AuthenticatorURI        string
	ServerURI               string
	MapProviderURI          string
	SpatialDatabaseURI      string
	PropertiesReaderURI     string
	IteratorURI             string
	IteratorSources         []string
	EnableCustomPlacetypes  bool
	CustomPlacetypes        string
	IsWhosOnFirst           bool
	EnableWWW               bool
	EnableGeoJSON           bool
	EnableGzip              bool
	LogTimings              bool
	EnableCORS              bool
	CORSOrigins             []string
	CORSAllowCredentials    bool
	PathPing                string
	PathData                string
	PathAPI                 string
	PathPrefix              string
	PathPIP                 string
	LeafletInitialLatitude  float64
	LeafletInitialLongitude float64
	LeafletInitialZoom      int
	LeafletMaxBounds        string
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WHOSONFIRST")

	if err != nil {
		return nil, fmt.Errorf("Failed to set flags from environment variables, %v", err)
	}

	iterator_sources := fs.Args()

	opts := &RunOptions{
		AuthenticatorURI: authenticator_uri,
		ServerURI:        server_uri,

		SpatialDatabaseURI:     spatial_database_uri,
		PropertiesReaderURI:    properties_reader_uri,
		IteratorURI:            iterator_uri,
		IteratorSources:        iterator_sources,
		EnableCustomPlacetypes: enable_custom_placetypes,
		CustomPlacetypes:       custom_placetypes,
		IsWhosOnFirst:          is_wof,

		EnableWWW:     enable_www,
		EnableGeoJSON: enable_geojson,
		EnableGzip:    enable_gzip,
		PathPing:      path_ping,
		PathData:      path_data,
		PathAPI:       path_api,
		PathPrefix:    path_prefix,
		PathPIP:       path_pip,

		LogTimings:           log_timings,
		EnableCORS:           enable_cors,
		CORSOrigins:          cors_origins,
		CORSAllowCredentials: cors_allow_credentials,

		MapProviderURI:          map_provider_uri,
		LeafletInitialLatitude:  leaflet_initial_latitude,
		LeafletInitialLongitude: leaflet_initial_longitude,
		LeafletInitialZoom:      leaflet_initial_zoom,
		LeafletMaxBounds:        leaflet_max_bounds,
	}

	return opts, nil
}
