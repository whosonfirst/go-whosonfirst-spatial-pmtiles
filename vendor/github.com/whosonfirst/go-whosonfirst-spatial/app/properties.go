package app

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

func NewPropertiesReaderWithFlagsSet(ctx context.Context, fs *flag.FlagSet) (reader.Reader, error) {

	reader_uri, _ := lookup.StringVar(fs, flags.PROPERTIES_READER_URI)

	if reader_uri == "" {
		return nil, nil
	}

	return reader.NewReader(ctx, reader_uri)
}
