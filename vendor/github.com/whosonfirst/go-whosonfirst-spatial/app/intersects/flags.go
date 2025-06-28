package intersects

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-reader"
	iter_flags "github.com/whosonfirst/go-whosonfirst-iterate/v3/flags"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
)

var spatial_database_uri string
var properties_reader_uri string

var enable_custom_placetypes bool
var custom_placetypes string

var geom_source string
var geom_type string
var geom_value string

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

var mode string

var verbose bool

var sort_uris multi.MultiString

var iterator_uris iter_flags.MultiIteratorURIFlag

func DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("pip")

	available_databases := database.Schemes()
	desc_databases := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: %s", available_databases)

	fs.StringVar(&spatial_database_uri, "spatial-database-uri", "rtree://", desc_databases)

	available_readers := reader.Schemes()
	desc_readers := fmt.Sprintf("A valid whosonfirst/go-reader.Reader URI. Available options are: %s", available_readers)

	fs.StringVar(&properties_reader_uri, "properties-reader-uri", "", fmt.Sprintf("%s. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.", desc_readers))

	fs.BoolVar(&enable_custom_placetypes, "enable-custom-placetypes", false, "Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.")

	fs.StringVar(&custom_placetypes, "custom-placetypes", "", "A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.")

	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	// query flags

	fs.StringVar(&geom_source, "geometry-source", "flag", "Where to 'read' a geometry (to intersect) from. Valid options are: file, flag, stdin")
	fs.StringVar(&geom_type, "geometry-type", "", "The type of encoding used to perform an intersects operation. Valid options are: geojson, wkt, bbox")
	fs.StringVar(&geom_value, "geometry-value", "", "The value of geometry used to perform an intersects operation. This will vary depending on the value of the -geometry-source flag. For example if -geometry-source=flag then -geometry-value= will be that geometry passed as a string. If -geometry-source=file then -geometry-value= will be the path to a file on disk. If -geometry-source=stdin then -geometry-value will be left empty.")

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

	// Indexing flags

	desc_iter := iter_flags.IteratorURIFlagDescription()
	desc_iter = fmt.Sprintf("Zero or more URIs denoting data sources to use for indexing the spatial database at startup. %s", desc_iter)

	fs.Var(&iterator_uris, "iterator-uri", desc_iter)

	// Runtime / server flags

	fs.StringVar(&mode, "mode", "cli", "Valid options are: cli, lambda.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Perform an intersects operation (as in intersecting geometries) for an input geometry and on a set of Who's on First records stored in a spatial database.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n\n")
		fs.PrintDefaults()
	}

	return fs, nil
}
