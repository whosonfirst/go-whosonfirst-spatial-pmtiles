package database

import (
	"context"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func IndexDatabaseWithIterators(ctx context.Context, db SpatialDatabase, sources map[string][]string) error {

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		err := IndexDatabaseWithReader(ctx, db, r)

		if err != nil {
			return fmt.Errorf("Failed to index %s, %w", path, err)
		}

		return nil
	}

	defer debug.FreeOSMemory()

	for iter_uri, iter_sources := range sources {

		iter, err := iterator.NewIterator(ctx, iter_uri, iter_cb)

		if err != nil {
			return fmt.Errorf("Failed to create iterator for %s, %w", iter_uri, err)
		}

		err = iter.IterateURIs(ctx, iter_sources...)

		if err != nil {
			return fmt.Errorf("Failed to iterate sources for %s (%v), %w", iter_uri, iter_sources, err)
		}

		debug.FreeOSMemory()
	}

	return nil
}
