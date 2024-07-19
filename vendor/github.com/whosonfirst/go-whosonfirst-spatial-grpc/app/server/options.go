package server

import (
	"context"
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

type RunOptions struct {
	Host                   string
	Port                   int
	SpatialDatabaseURI     string   `json:"spatial_database_uri"`
	PropertiesReaderURI    string   `json:"properties_reader_uri"`
	IteratorURI            string   `json:"iterator_uri"`
	IteratorSources        []string `json:"iterator_sources"`
	EnableCustomPlacetypes bool     `json:"enable_custom_placetypes"`
	CustomPlacetypes       string   `json:"custom_placetypes"`
	IsWhosOnFirst          bool     `json:"is_whosonfirst"`
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WHOSONFIRST")

	if err != nil {
		return nil, err
	}

	iterator_sources := fs.Args()

	opts := &RunOptions{
		Host:                   host,
		Port:                   port,
		SpatialDatabaseURI:     spatial_database_uri,
		PropertiesReaderURI:    properties_reader_uri,
		IteratorURI:            iterator_uri,
		IteratorSources:        iterator_sources,
		EnableCustomPlacetypes: enable_custom_placetypes,
		CustomPlacetypes:       custom_placetypes,
		IsWhosOnFirst:          is_wof,
	}

	return opts, nil
}
