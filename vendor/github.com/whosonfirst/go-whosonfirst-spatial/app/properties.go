package app

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

func NewPropertiesReaderWithFlagsSet(ctx context.Context, fs *flag.FlagSet) (reader.Reader, error) {

	reader_uri, _ := lookup.StringVar(fs, flags.PROPERTIES_READER_URI)

	if reader_uri == "" {
		return nil, nil
	}

	use_spatial_uri := fmt.Sprintf("{%s}", flags.SPATIAL_DATABASE_URI)

	if reader_uri == use_spatial_uri {

		spatial_database_uri, err := lookup.StringVar(fs, flags.SPATIAL_DATABASE_URI)

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve %s flag", flags.SPATIAL_DATABASE_URI)
		}

		reader_uri = spatial_database_uri
	}

	return reader.NewReader(ctx, reader_uri)
}
