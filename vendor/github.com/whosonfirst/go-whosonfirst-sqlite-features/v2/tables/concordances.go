package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
)

type ConcordancesTable struct {
	features.FeatureTable
	name string
}

type ConcordancesRow struct {
	Id           int64
	OtherID      string
	OtherSource  string
	LastModified int64
}

func NewConcordancesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewConcordancesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, InitializeTableError(t, err)
	}

	return t, nil
}

func NewConcordancesTable(ctx context.Context) (sqlite.Table, error) {

	t := ConcordancesTable{
		name: "concordances",
	}

	return &t, nil
}

func (t *ConcordancesTable) Name() string {
	return t.name
}

func (t *ConcordancesTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		other_id INTEGER NOT NULL,
		other_source TEXT,
		lastmodified INTEGER
	);

	CREATE INDEX concordances_by_id ON %s (id,lastmodified);
	CREATE INDEX concordances_by_other_id ON %s (other_source,other_id);	
	CREATE INDEX concordances_by_other_lastmod ON %s (other_source,other_id,lastmodified);
	CREATE INDEX concordances_by_lastmod ON %s (lastmodified);`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *ConcordancesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *ConcordancesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *ConcordancesTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

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

	concordances := properties.Concordances(f)
	lastmod := properties.LastModified(f)

	for other_source, other_id := range concordances {

		sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, other_id, other_source, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())

		stmt, err := tx.Prepare(sql)

		if err != nil {
			return PrepareStatementError(t, err)
		}

		defer stmt.Close()

		_, err = stmt.Exec(id, other_id, other_source, lastmod)

		if err != nil {
			return ExecuteStatementError(t, err)
		}
	}

	err = tx.Commit()

	if err != nil {
		return CommitTransactionError(t, err)
	}

	return nil
}
