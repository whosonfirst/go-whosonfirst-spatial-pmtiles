package tables

import (
	"context"
	"database/sql"
	"fmt"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

const CONCORDANCES_TABLE_NAME string = "concordances"

type ConcordancesTable struct {
	database_sql.Table
	FeatureTable
	name string
}

type ConcordancesRow struct {
	Id           int64
	OtherID      string
	OtherSource  string
	LastModified int64
}

func NewConcordancesTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	t, err := NewConcordancesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, database_sql.InitializeTableError(t, err)
	}

	return t, nil
}

func NewConcordancesTable(ctx context.Context) (database_sql.Table, error) {

	t := ConcordancesTable{
		name: CONCORDANCES_TABLE_NAME,
	}

	return &t, nil
}

func (t *ConcordancesTable) Name() string {
	return t.name
}

func (t *ConcordancesTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, CONCORDANCES_TABLE_NAME)
}

func (t *ConcordancesTable) InitializeTable(ctx context.Context, db *sql.DB) error {
	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *ConcordancesTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *ConcordancesTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	db_driver := database_sql.Driver(db)

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

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

	concordances := properties.Concordances(f)
	lastmod := properties.LastModified(f)

	for other_source, other_id := range concordances {

		var insert_q string

		switch db_driver {
		case database_sql.POSTGRES_DRIVER:

			insert_q = fmt.Sprintf(`INSERT INTO %s (
				id, other_id, other_source, lastmodified
			) VALUES (
			  	 $1, $2, $3, $4
			) ON CONFLICT(id, other_source) DO UPDATE SET
				other_id = EXCLUDED.other_id,
				lastmodified = EXCLUDED.lastmodified`, t.Name())

		default:

			insert_q = fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, other_id, other_source, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())
		}

		stmt, err := tx.Prepare(insert_q)

		if err != nil {
			return database_sql.PrepareStatementError(t, err)
		}

		defer stmt.Close()

		_, err = stmt.Exec(id, other_id, other_source, lastmod)

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
