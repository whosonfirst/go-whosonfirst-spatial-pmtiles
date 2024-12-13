package tables

import (
	"context"
	"database/sql"

	database_sql "github.com/sfomuseum/go-database/sql"
)

type FeatureTable interface {
	database_sql.Table
	IndexFeature(context.Context, *sql.DB, []byte) error
}
