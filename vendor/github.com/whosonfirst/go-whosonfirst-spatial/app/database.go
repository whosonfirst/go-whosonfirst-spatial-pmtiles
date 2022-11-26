package app

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

func NewSpatialDatabaseWithFlagSet(ctx context.Context, fl *flag.FlagSet) (database.SpatialDatabase, error) {

	spatial_uri, err := lookup.StringVar(fl, flags.SPATIAL_DATABASE_URI)

	if err != nil {
		return nil, err
	}

	return database.NewSpatialDatabase(ctx, spatial_uri)
}
