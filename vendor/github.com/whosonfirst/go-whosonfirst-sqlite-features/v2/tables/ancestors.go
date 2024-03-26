package tables

import (
	"context"
	"fmt"
	"strings"

	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	sql_tables "github.com/whosonfirst/go-whosonfirst-sql/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
)

type AncestorsTable struct {
	features.FeatureTable
	name string
}

type AncestorsRow struct {
	Id                int64
	AncestorID        int64
	AncestorPlacetype string
	LastModified      int64
}

func NewAncestorsTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewAncestorsTable(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to create ancestors table, %w", err)
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, InitializeTableError(t, err)
	}

	return t, nil
}

func NewAncestorsTable(ctx context.Context) (sqlite.Table, error) {

	t := AncestorsTable{
		name: sql_tables.ANCESTORS_TABLE_NAME,
	}

	return &t, nil
}

func (t *AncestorsTable) Name() string {
	return t.name
}

func (t *AncestorsTable) Schema() string {
	schema, _ := sql_tables.LoadSchema("sqlite", sql_tables.ANCESTORS_TABLE_NAME)
	return schema
}

func (t *AncestorsTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *AncestorsTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *AncestorsTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	conn, err := db.Conn(ctx)

	if err != nil {
		return DatabaseConnectionError(t, err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return BeginTransactionError(t, err)
	}

	sql := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		return ExecuteStatementError(t, err)
	}

	hierarchies := properties.Hierarchies(f)
	lastmod := properties.LastModified(f)

	for _, h := range hierarchies {

		for pt_key, ancestor_id := range h {

			ancestor_placetype := strings.Replace(pt_key, "_id", "", -1)

			sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, ancestor_id, ancestor_placetype, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())

			stmt, err := tx.Prepare(sql)

			if err != nil {
				return PrepareStatementError(t, err)
			}

			defer stmt.Close()

			_, err = stmt.Exec(id, ancestor_id, ancestor_placetype, lastmod)

			if err != nil {
				return ExecuteStatementError(t, err)
			}

		}

	}

	err = tx.Commit()

	if err != nil {
		return CommitTransactionError(t, err)
	}

	return nil
}
