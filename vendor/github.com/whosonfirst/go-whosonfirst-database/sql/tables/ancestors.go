package tables

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

const ANCESTORS_TABLE_NAME string = "ancestors"

type AncestorsTable struct {
	database_sql.Table
	FeatureTable
	name string
}

type AncestorsRow struct {
	Id                int64
	AncestorID        int64
	AncestorPlacetype string
	LastModified      int64
}

func NewAncestorsTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	t, err := NewAncestorsTable(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to create '%s' table, %w", ANCESTORS_TABLE_NAME, err)
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, database_sql.InitializeTableError(t, err)
	}

	return t, nil
}

func NewAncestorsTable(ctx context.Context) (database_sql.Table, error) {

	t := AncestorsTable{
		name: ANCESTORS_TABLE_NAME,
	}

	return &t, nil
}

func (t *AncestorsTable) Name() string {
	return t.name
}

func (t *AncestorsTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, ANCESTORS_TABLE_NAME)

}

func (t *AncestorsTable) InitializeTable(ctx context.Context, db *sql.DB) error {
	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *AncestorsTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *AncestorsTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	db_driver := database_sql.Driver(db)

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	var delete_q string

	switch db_driver {
	case database_sql.POSTGRES_DRIVER:
		delete_q = fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, t.Name())
	default:
		delete_q = fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, t.Name())
	}

	stmt, err := tx.Prepare(delete_q)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	hierarchies := properties.Hierarchies(f)
	lastmod := properties.LastModified(f)

	for _, h := range hierarchies {

		for pt_key, ancestor_id := range h {

			ancestor_placetype := strings.Replace(pt_key, "_id", "", -1)

			var q string

			switch db_driver {
			case database_sql.POSTGRES_DRIVER:

				q = fmt.Sprintf(`INSERT INTO %s (
						id, ancestor_id, ancestor_placetype, lastmodified
					) VALUES (
			  	 		$1, $2, $3, $4
					) ON CONFLICT(id, ancestor_id) DO UPDATE SET
						ancestor_placetype = EXCLUDED.ancestor_placetype,
						lastmodified = EXCLUDED.lastmodified`, t.Name())

			default:

				q = fmt.Sprintf(`INSERT OR REPLACE INTO %s (
						id, ancestor_id, ancestor_placetype, lastmodified
					) VALUES (
			  	 		?, ?, ?, ?
					)`, t.Name())
			}

			stmt, err := tx.Prepare(q)

			if err != nil {
				return database_sql.PrepareStatementError(t, err)
			}

			defer stmt.Close()

			_, err = stmt.Exec(id, ancestor_id, ancestor_placetype, lastmod)

			if err != nil {
				return database_sql.ExecuteStatementError(t, err)
			}

		}

	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}
