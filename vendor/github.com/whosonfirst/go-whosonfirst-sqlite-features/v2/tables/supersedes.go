package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
)

type SupersedesTable struct {
	features.FeatureTable
	name string
}

func NewSupersedesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

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

func NewSupersedesTable(ctx context.Context) (sqlite.Table, error) {

	t := SupersedesTable{
		name: "supersedes",
	}

	return &t, nil
}

func (t *SupersedesTable) Name() string {
	return t.name
}

func (t *SupersedesTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		superseded_id INTEGER NOT NULL,
		superseded_by_id INTEGER NOT NULL,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX supersedes_by ON %s (id,superseded_id, superseded_by_id);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name())
}

func (t *SupersedesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *SupersedesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *SupersedesTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	lastmod := properties.LastModified(f)

	conn, err := db.Conn(ctx)

	if err != nil {
		return DatabaseConnectionError(t, err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return BeginTransactionError(t, err)
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, superseded_id, superseded_by_id, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	superseded_by := properties.SupersededBy(f)

	for _, other_id := range superseded_by {

		_, err = stmt.Exec(id, id, other_id, lastmod)

		if err != nil {
			return ExecuteStatementError(t, err)
		}

	}

	supersedes := properties.Supersedes(f)

	for _, other_id := range supersedes {

		_, err = stmt.Exec(id, other_id, id, lastmod)

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
