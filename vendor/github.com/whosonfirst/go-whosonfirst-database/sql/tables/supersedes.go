package tables

import (
	"context"
	"database/sql"
	"fmt"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

const SUPERSEDES_TABLE_NAME string = "supersedes"

type SupersedesTable struct {
	database_sql.Table
	FeatureTable
	name string
}

func NewSupersedesTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	t, err := NewSupersedesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSupersedesTable(ctx context.Context) (database_sql.Table, error) {

	t := SupersedesTable{
		name: SUPERSEDES_TABLE_NAME,
	}

	return &t, nil
}

func (t *SupersedesTable) Name() string {
	return t.name
}

func (t *SupersedesTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, SUPERSEDES_TABLE_NAME)
}

func (t *SupersedesTable) InitializeTable(ctx context.Context, db *sql.DB) error {
	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *SupersedesTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *SupersedesTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	lastmod := properties.LastModified(f)

	db_driver := database_sql.Driver(db)

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	var insert_q string

	switch db_driver {
	case database_sql.POSTGRES_DRIVER:

		insert_q = fmt.Sprintf(`INSERT INTO %s (
				id, superseded_id, superseded_by_id, lastmodified
			) VALUES (
			  	 $1, $2, $3, $4
			) ON CONFLICT(id) DO UPDATE SET
				superseded_id = EXCLUDED.superseded_id,
				superseded_by_id = EXCLUDED.superseded_by_id,
				lastmodified = EXCLUDED.lastmodified`, t.Name())

	default:

		insert_q = fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, superseded_id, superseded_by_id, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())
	}

	stmt, err := tx.Prepare(insert_q)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	superseded_by := properties.SupersededBy(f)

	for _, other_id := range superseded_by {

		_, err = stmt.Exec(id, id, other_id, lastmod)

		if err != nil {
			return database_sql.ExecuteStatementError(t, err)
		}

	}

	supersedes := properties.Supersedes(f)

	for _, other_id := range supersedes {

		_, err = stmt.Exec(id, other_id, id, lastmod)

		if err != nil {
			return database_sql.ExecuteStatementError(t, err)
		}

	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}
