package database

import (
	"context"
	"fmt"
	"io"

	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

// IndexDatabaseWithIterator is a general-purpose method for indexing a `database.Spatial.Database` instance with a
// whosonfirst/go-whosonfirst-iterate/v2 iterator. Only records whose geometry type are 'Polygon' or 'MultiPolygon'
// will be indexed.
func IndexDatabaseWithIterator(ctx context.Context, db SpatialDatabase, iterator_uri string, iterator_sources ...string) error {

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		return IndexDatabaseWithReader(ctx, db, r)
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		return fmt.Errorf("Failed to create iterator, %w", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		fmt.Errorf("Failed to iterate URIs, %w", err)
	}

	return nil
}
