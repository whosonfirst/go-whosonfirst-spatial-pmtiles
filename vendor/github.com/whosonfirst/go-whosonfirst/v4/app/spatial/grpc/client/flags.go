package client

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

var host string
var port int

var latitude float64
var longitude float64
var geometries string

var inception string
var cessation string

var props multi.MultiString
var placetypes multi.MultiString
var alt_geoms multi.MultiString

var is_current multi.MultiInt64
var is_ceased multi.MultiInt64
var is_deprecated multi.MultiInt64
var is_superseded multi.MultiInt64
var is_superseding multi.MultiInt64

var sort_uris multi.MultiString

var stdout bool
var null bool

func DefaultFlagSet() (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("client")

	fs.StringVar(&host, "host", "localhost", "The host of the gRPC server to connect to.")
	fs.IntVar(&port, "port", 8082, "The port of the gRPC server to connect to.")

	fs.BoolVar(&stdout, "stdout", true, "Emit results to STDOUT")
	fs.BoolVar(&null, "null", false, "Emit results to /dev/null")

	// query flags

	fs.Float64Var(&latitude, "latitude", 0.0, "A valid latitude.")
	fs.Float64Var(&longitude, "longitude", 0.0, "A valid longitude.")

	fs.StringVar(&geometries, "geometries", "all", "Valid options are: all, alt, default.")

	fs.StringVar(&inception, "inception", "", "A valid EDTF date string.")
	fs.StringVar(&cessation, "cessation", "", "A valid EDTF date string.")

	fs.Var(&props, "property", "One or more Who's On First properties to append to each result.")
	fs.Var(&placetypes, "placetype", "One or more place types to filter results by.")

	fs.Var(&alt_geoms, "alternate-geometry", "One or more alternate geometry labels (wof:alt_label) values to filter results by.")

	fs.Var(&is_current, "is-current", "One or more existential flags (-1, 0, 1) to filter results by.")
	fs.Var(&is_ceased, "is-ceased", "One or more existential flags (-1, 0, 1) to filter results by.")
	fs.Var(&is_deprecated, "is-deprecated", "One or more existential flags (-1, 0, 1) to filter results by.")
	fs.Var(&is_superseded, "is-superseded", "One or more existential flags (-1, 0, 1) to filter results by.")
	fs.Var(&is_superseding, "is-superseding", "One or more existential flags (-1, 0, 1) to filter results by.")

	fs.Var(&sort_uris, "sort-uri", "Zero or more whosonfirst/go-whosonfirst-spr/sort URIs.")

	return fs, nil
}
