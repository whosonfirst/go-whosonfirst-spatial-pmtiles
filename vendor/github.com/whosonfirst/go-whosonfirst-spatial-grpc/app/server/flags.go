package server

import (
	"flag"
	"fmt"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	spatial_flags "github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

var host string
var port int

var spatial_database_uri string
var properties_reader_uri string

var is_wof bool

var enable_custom_placetypes bool
var custom_placetypes string

var iterator_uris spatial_flags.MultiCSVIteratorURIFlag

func DefaultFlagSet() (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("server")

	fs.StringVar(&host, "host", "localhost", "The host to listen for requests on")
	fs.IntVar(&port, "port", 8082, "The port to listen for requests on")

	available_databases := database.Schemes()
	desc_databases := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: %s", available_databases)

	fs.StringVar(&spatial_database_uri, "spatial-database-uri", "rtree://", desc_databases)

	available_readers := reader.Schemes()
	desc_readers := fmt.Sprintf("A valid whosonfirst/go-reader.Reader URI. Available options are: %s", available_readers)

	fs.StringVar(&properties_reader_uri, "properties-reader-uri", "", fmt.Sprintf("%s. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.", desc_readers))

	fs.BoolVar(&is_wof, "is-wof", true, "Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents.")

	fs.BoolVar(&enable_custom_placetypes, "enable-custom-placetypes", false, "Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.")

	fs.StringVar(&custom_placetypes, "custom-placetypes", "", "A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.")

	// Indexing flags

	desc_iter := spatial_flags.IteratorURIFlagDescription()
	desc_iter = fmt.Sprintf("Zero or more URIs denoting data sources to use for indexing the spatial database at startup. %s", desc_iter)
	fs.Var(&iterator_uris, "iterator-uri", desc_iter)

	return fs, nil
}
