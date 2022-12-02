package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
)

type PropertiesTableOptions struct {
	IndexAltFiles bool
}

func DefaultPropertiesTableOptions() (*PropertiesTableOptions, error) {

	opts := PropertiesTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type PropertiesTable struct {
	features.FeatureTable
	name    string
	options *PropertiesTableOptions
}

type PropertiesRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewPropertiesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewPropertiesTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewPropertiesTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *PropertiesTableOptions) (sqlite.Table, error) {

	t, err := NewPropertiesTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, InitializeTableError(t, err)
	}

	return t, nil
}

func NewPropertiesTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewPropertiesTableWithOptions(ctx, opts)
}

func NewPropertiesTableWithOptions(ctx context.Context, opts *PropertiesTableOptions) (sqlite.Table, error) {

	t := PropertiesTable{
		name:    "properties",
		options: opts,
	}

	return &t, nil
}

func (t *PropertiesTable) Name() string {
	return t.name
}

func (t *PropertiesTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		body TEXT,
		is_alt BOOLEAN,
		alt_label TEXT,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX properties_by_id ON %s (id, alt_label);
	CREATE INDEX properties_by_alt ON %s (id, is_alt, alt_label);
	CREATE INDEX properties_by_lastmod ON %s (lastmodified);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *PropertiesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *PropertiesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *PropertiesTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

	is_alt := alt.IsAlt(f)

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	alt_label, err := properties.AltLabel(f)

	if err != nil {
		return MissingPropertyError(t, "alt label", err)
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
		id, body, is_alt, alt_label, lastmodified
	) VALUES (
		?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	rsp_props := gjson.GetBytes(f, "properties")
	str_props := rsp_props.String()

	_, err = stmt.Exec(id, str_props, is_alt, alt_label, lastmod)

	if err != nil {
		return ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return CommitTransactionError(t, err)
	}

	return nil
}
