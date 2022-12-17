package query

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

const ENABLE_GEOJSON string = "enable-geojson"
const SERVER_URI string = "server-uri"
const MODE string = "mode"

var mode string
var server_uri string
var enable_geojson bool

var sort_uris multi.MultiString

func DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs, err := flags.CommonFlags()

	if err != nil {
		return nil, err
	}

	err = flags.AppendQueryFlags(fs)

	if err != nil {
		return nil, err
	}

	err = flags.AppendIndexingFlags(fs)

	if err != nil {
		return nil, err
	}

	fs.StringVar(&mode, "mode", "cli", "Valid options are: cli")
	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI.")

	return fs, nil
}
