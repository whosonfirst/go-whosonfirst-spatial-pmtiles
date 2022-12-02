package sqlite

import (
	"context"
	"fmt"
)

func LiveHardDieFast(ctx context.Context, db Database) error {

	conn, err := db.Conn(ctx)

	if err != nil {
		return fmt.Errorf("Failed to establish database connection, %w", err)
	}

	pragma := []string{
		"PRAGMA JOURNAL_MODE=OFF",
		"PRAGMA SYNCHRONOUS=OFF",
		"PRAGMA LOCKING_MODE=EXCLUSIVE",
		// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
		"PRAGMA PAGE_SIZE=4096",
		"PRAGMA CACHE_SIZE=1000000",
	}

	for _, p := range pragma {

		_, err = conn.Exec(p)

		if err != nil {
			return fmt.Errorf("Failed to set pragma '%s', %w", p, err)
		}
	}

	return nil
}
