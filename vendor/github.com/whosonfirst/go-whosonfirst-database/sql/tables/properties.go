package tables

import (
	"context"
	"database/sql"
	"fmt"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

const PROPERTIES_TABLE_NAME string = "properties"

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
	database_sql.Table
	FeatureTable
	name    string
	options *PropertiesTableOptions
}

type PropertiesRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewPropertiesTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewPropertiesTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewPropertiesTableWithDatabaseAndOptions(ctx context.Context, db *sql.DB, opts *PropertiesTableOptions) (database_sql.Table, error) {

	t, err := NewPropertiesTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, database_sql.InitializeTableError(t, err)
	}

	return t, nil
}

func NewPropertiesTable(ctx context.Context) (database_sql.Table, error) {

	opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewPropertiesTableWithOptions(ctx, opts)
}

func NewPropertiesTableWithOptions(ctx context.Context, opts *PropertiesTableOptions) (database_sql.Table, error) {

	t := PropertiesTable{
		name:    PROPERTIES_TABLE_NAME,
		options: opts,
	}

	return &t, nil
}

func (t *PropertiesTable) Name() string {
	return t.name
}

func (t *PropertiesTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, PROPERTIES_TABLE_NAME)
}

func (t *PropertiesTable) InitializeTable(ctx context.Context, db *sql.DB) error {

	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *PropertiesTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *PropertiesTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

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

	db_driver := database_sql.Driver(db)

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	var insert_q string

	switch db_driver {
	case database_sql.POSTGRES_DRIVER:

		insert_q = fmt.Sprintf(`INSERT INTO %s (
			id, body, is_alt, alt_label, lastmodified
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT(id, alt_label) DO UPDATE SET
			body = EXCLUDED.body,
			is_alt = EXCLUDED.is_alt,
			lastmodified = EXCLUDED.lastmodified`, t.Name())

	default:

		insert_q = fmt.Sprintf(`INSERT OR REPLACE INTO %s (
			id, body, is_alt, alt_label, lastmodified
		) VALUES (
			?, ?, ?, ?, ?
		)`, t.Name())
	}

	stmt, err := tx.Prepare(insert_q)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	rsp_props := gjson.GetBytes(f, "properties")
	str_props := rsp_props.String()

	_, err = stmt.Exec(id, str_props, is_alt, alt_label, lastmod)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}
