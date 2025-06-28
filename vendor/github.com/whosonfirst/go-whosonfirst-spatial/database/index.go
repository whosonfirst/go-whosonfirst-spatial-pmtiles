package database

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

func IndexDatabaseWithIterators(ctx context.Context, db SpatialDatabase, sources map[string][]string) error {

	defer debug.FreeOSMemory()

	for iter_uri, iter_sources := range sources {

		iter, err := iterate.NewIterator(ctx, iter_uri)

		if err != nil {
			return fmt.Errorf("Failed to create iterator for %s, %w", iter_uri, err)
		}

		for rec, err := range iter.Iterate(ctx, iter_sources...) {

			if err != nil {
				return fmt.Errorf("Failed to iterate sources for %s (%v), %w", iter_uri, iter_sources, err)
			}

			defer rec.Body.Close()

			err := IndexDatabaseWithReader(ctx, db, rec.Body)

			if err != nil {
				return fmt.Errorf("Failed to index %s, %w", rec.Path, err)
			}

		}

		debug.FreeOSMemory()
	}

	return nil
}
