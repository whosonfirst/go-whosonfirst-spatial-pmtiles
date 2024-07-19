package client

import (
	"context"
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

type RunOptions struct {
	Host                string   `json:"host"`
	Port                int      `json:"port"`
	Stdout              bool     `json:"stdout"`
	Null                bool     `json:"null"`
	Latitude            float64  `json:"latitude"`
	Longitude           float64  `json:"longitude"`
	Placetypes          []string `json:"placetypes,omitempty"`
	Geometries          string   `json:"geometries,omitempty"`
	AlternateGeometries []string `json:"alternate_geometries,omitempty"`
	IsCurrent           []int64  `json:"is_current,omitempty"`
	IsCeased            []int64  `json:"is_ceased,omitempty"`
	IsDeprecated        []int64  `json:"is_deprecated,omitempty"`
	IsSuperseded        []int64  `json:"is_superseded,omitempty"`
	IsSuperseding       []int64  `json:"is_superseding,omitempty"`
	InceptionDate       string   `json:"inception_date,omitempty"`
	CessationDate       string   `json:"cessation_date,omitempty"`
	Properties          []string `json:"properties,omitempty"`
	Sort                []string `json:"sort,omitempty"`
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WHOSONFIRST")

	if err != nil {
		return nil, err
	}

	opts := &RunOptions{
		Host:                host,
		Port:                port,
		Stdout:              stdout,
		Null:                null,
		Latitude:            latitude,
		Longitude:           longitude,
		Placetypes:          placetypes,
		Geometries:          geometries,
		AlternateGeometries: alt_geoms,
		IsCurrent:           is_current,
		IsCeased:            is_ceased,
		IsDeprecated:        is_deprecated,
		IsSuperseded:        is_superseded,
		IsSuperseding:       is_superseding,
		InceptionDate:       inception,
		CessationDate:       cessation,
		Properties:          props,
		Sort:                sort_uris,
	}

	return opts, nil
}
