package modernc

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/aaronland/go-sqlite/v2"
	"github.com/aaronland/go-sqlite/v2/database"
	_ "modernc.org/sqlite"
)

const SQLITE_SCHEME string = "modernc"
const SQLITE_DRIVER string = "sqlite"

// In principle this could also be done with a sync.OnceFunc call but that will
// require that everyone uses Go 1.21 (whose package import changes broke everything)
// which is literally days old as I write this. So maybe a few releases after 1.21.

var register_mu = new(sync.RWMutex)
var register_map = map[string]bool{}

func init() {

	ctx := context.Background()
	err := RegisterSQLiteSchemes(ctx)

	if err != nil {
		panic(err)
	}
}

// RegisterSQLiteSchemes will explicitly register all the schemes associated with the `client.Client` interface.
func RegisterSQLiteSchemes(ctx context.Context) error {

	roster := map[string]sqlite.DatabaseInitializationFunc{
		SQLITE_SCHEME: NewModerncDatabase,
	}

	register_mu.Lock()
	defer register_mu.Unlock()

	for scheme, fn := range roster {

		_, exists := register_map[scheme]

		if exists {
			continue
		}

		err := sqlite.RegisterDatabase(ctx, scheme, fn)

		if err != nil {
			return fmt.Errorf("Failed to register database for '%s', %w", scheme, err)
		}

		register_map[scheme] = true
	}

	return nil
}

func NewModerncDatabase(ctx context.Context, db_uri string) (sqlite.Database, error) {

	dsn, err := database.DSNFromURI(db_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	conn, err := sql.Open(SQLITE_DRIVER, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	return database.NewAnyDatabase(ctx, dsn, conn)
}
