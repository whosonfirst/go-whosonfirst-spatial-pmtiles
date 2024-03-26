package database

import (
	"context"
	"fmt"
	"io/fs"
)

func IndexDatabaseWithFS(ctx context.Context, db SpatialDatabase, index_fs fs.FS) error {

	walk_func := func(path string, d fs.DirEntry, err error) error {

		if d.IsDir() {
			return nil
		}

		r, err := index_fs.Open(path)

		if err != nil {
			return fmt.Errorf("Failed to open %s for reading, %w", path, err)
		}

		defer r.Close()

		return IndexDatabaseWithReader(ctx, db, r)
	}

	return fs.WalkDir(index_fs, ".", walk_func)
}
