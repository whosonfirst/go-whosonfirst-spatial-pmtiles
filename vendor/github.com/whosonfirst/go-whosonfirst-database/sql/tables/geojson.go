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

	str_body := string(f)

	var q string

	args := []any{
		id,
		str_body,
		source,
		is_alt,
		alt_label,
		lastmod,
	}

	switch database_sql.Driver(db) {
	case database_sql.POSTGRES_DRIVER:

		q = fmt.Sprintf(`INSERT INTO %s (
			id, body, source, is_alt, alt_label, lastmodified
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) ON CONFLICT (id, alt_label) DO UPDATE SET
			body = EXCLUDED.body,
			source = EXCLUDED.source,
			is_alt = EXCLUDED.is_alt,
			lastmodified = EXCLUDED.lastmodified`, t.Name())

	default:

		q = fmt.Sprintf(`INSERT OR REPLACE INTO %s (
			id, body, source, is_alt, alt_label, lastmodified
		) VALUES (
			?, ?, ?, ?, ?, ?
		)`, t.Name())

	}

	stmt, err := tx.Prepare(q)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}
