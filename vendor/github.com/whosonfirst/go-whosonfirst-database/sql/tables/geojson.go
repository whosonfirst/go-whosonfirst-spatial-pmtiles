package tables

import (
	"context"
	"database/sql"
	"fmt"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

const GEOJSON_TABLE_NAME string = "geojson"

type GeoJSONTableOptions struct {
	IndexAltFiles          bool
	AllowMissingSourceGeom bool
}

func DefaultGeoJSONTableOptions() (*GeoJSONTableOptions, error) {

	opts := GeoJSONTableOptions{
		IndexAltFiles:          false,
		AllowMissingSourceGeom: true,
	}

	return &opts, nil
}

type GeoJSONTable struct {
	database_sql.Table
	FeatureTable
	name    string
	options *GeoJSONTableOptions
}

type GeoJSONRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewGeoJSONTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeoJSONTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewGeoJSONTableWithDatabaseAndOptions(ctx context.Context, db *sql.DB, opts *GeoJSONTableOptions) (database_sql.Table, error) {

	t, err := NewGeoJSONTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewGeoJSONTable(ctx context.Context) (database_sql.Table, error) {

	opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeoJSONTableWithOptions(ctx, opts)
}

func NewGeoJSONTableWithOptions(ctx context.Context, opts *GeoJSONTableOptions) (database_sql.Table, error) {

	t := GeoJSONTable{
		name:    GEOJSON_TABLE_NAME,
		options: opts,
	}

	return &t, nil
}

func (t *GeoJSONTable) Name() string {
	return t.name
}

func (t *GeoJSONTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, GEOJSON_TABLE_NAME)
}

func (t *GeoJSONTable) InitializeTable(ctx context.Context, db *sql.DB) error {

	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *GeoJSONTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *GeoJSONTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

	is_alt := alt.IsAlt(f)

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	source, err := properties.Source(f)

	if err != nil {

		if !t.options.AllowMissingSourceGeom {
			return MissingPropertyError(t, "source", err)
		}

		source = "unknown"
	}

	alt_label, err := properties.AltLabel(f)

	if err != nil {
		return MissingPropertyError(t, "alt label", err)
	}

	lastmod := properties.LastModified(f)

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, body, source, is_alt, alt_label, lastmodified
	) VALUES (
		?, ?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	str_body := string(f)

	_, err = stmt.Exec(id, str_body, source, is_alt, alt_label, lastmod)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}
