package update

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	spatial_flags "github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

var source_iterator_uris spatial_flags.MultiIteratorURIFlag
var target_iterator_uris spatial_flags.MultiIteratorURIFlag

var exporter_uri string
var writer_uri string

var spatial_database_uri string
var mapshaper_server string

var is_current multi.MultiInt64
var is_ceased multi.MultiInt64
var is_deprecated multi.MultiInt64
var is_superseded multi.MultiInt64
var is_superseding multi.MultiInt64

func DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("pip")

	desc_iter := spatial_flags.IteratorURIFlagDescription()

	source_desc := fmt.Sprintf("Zero or more URIs denoting data sources to use for indexing the spatial database at startup. %s", desc_iter)
	target_desc := fmt.Sprintf("Zero or more URIs denoting target data sources whose hierarchies need updating. %s", desc_iter)

	fs.Var(&source_iterator_uris, "source-iterator-uri", source_desc)
	fs.Var(&target_iterator_uris, "target-iterator-uri", target_desc)

	fs.StringVar(&exporter_uri, "exporter-uri", "whosonfirst://", "A valid whosonfirst/go-whosonfirst-export URI.")
	fs.StringVar(&writer_uri, "writer-uri", "null://", "A valid whosonfirst/go-writer URI. This is where updated records will be written to.")

	fs.StringVar(&spatial_database_uri, "spatial-database-uri", "rtree://", "A valid whosonfirst/go-whosonfirst-spatial URI. This is the database of spatial records that will for PIP-ing.")

	// As in github:sfomuseum/go-sfomuseum-mapshaper and github:sfomuseum/docker-sfomuseum-mapshaper
	// One day the functionality exposed here will be ported to Go and this won't be necessary

	fs.StringVar(&mapshaper_server, "mapshaper-server", "", "A valid HTTP URI pointing to a sfomuseum/go-sfomuseum-mapshaper server endpoint.")

	fs.Var(&is_current, "is-current", "One or more existential flags (-1, 0, 1) to filter PIP results.")
	fs.Var(&is_ceased, "is-ceased", "One or more existential flags (-1, 0, 1) to filter PIP results.")
	fs.Var(&is_deprecated, "is-deprecated", "One or more existential flags (-1, 0, 1) to filter PIP results.")
	fs.Var(&is_superseded, "is-superseded", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	fs.Var(&is_superseding, "is-superseding", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n\n")
		fs.PrintDefaults()
	}

	return fs, nil
}
