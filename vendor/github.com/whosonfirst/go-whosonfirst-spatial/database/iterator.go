package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

// TBD - Should this be remove in favour of IndexDatabaseWithIterators or does it just supplement it?

// IndexDatabaseWithIterator is a general-purpose method for indexing a `database.Spatial.Database` instance with a
// whosonfirst/go-whosonfirst-iterate/v3 iterator. Only records whose geometry type are 'Polygon' or 'MultiPolygon'
// will be indexed.
func IndexDatabaseWithIterator(ctx context.Context, db SpatialDatabase, iterator_uri string, iterator_sources ...string) error {

	iter, err := iterate.NewIterator(ctx, iterator_uri)

	if err != nil {
		return fmt.Errorf("Failed to create iterator, %w", err)
	}

	for rec, err := range iter.Iterate(ctx, iterator_sources...) {

		if err != nil {
			fmt.Errorf("Failed to iterate URIs, %w", err)
		}

		defer rec.Body.Close()

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		slog.Debug("Index", "path", rec.Path)
		err = IndexDatabaseWithReader(ctx, db, rec.Body)

		if err != nil {
			return err
		}

	}

	return nil
}
