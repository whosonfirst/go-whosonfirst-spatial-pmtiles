package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
	_ "log"
)

type GeometriesTableOptions struct {
	IndexAltFiles bool
}

func DefaultGeometriesTableOptions() (*GeometriesTableOptions, error) {

	opts := GeometriesTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type GeometriesTable struct {
	features.FeatureTable
	name    string
	options *GeometriesTableOptions
}

type GeometriesRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewGeometriesTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultGeometriesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeometriesTableWithOptions(ctx, opts)
}

func NewGeometriesTableWithOptions(ctx context.Context, opts *GeometriesTableOptions) (sqlite.Table, error) {

	t := GeometriesTable{
		name:    "geometries",
		options: opts,
	}

	return &t, nil
}

func NewGeometriesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultGeometriesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeometriesTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewGeometriesTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *GeometriesTableOptions) (sqlite.Table, error) {

	t, err := NewGeometriesTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *GeometriesTable) Name() string {
	return t.name
}

func (t *GeometriesTable) Schema() string {

	// really this should probably be the SPR table + geom but
	// let's just get this working first and then make it fancy
	// (20180109/thisisaaronland)

	// https://www.gaia-gis.it/spatialite-1.0a/SpatiaLite-tutorial.html
	// http://www.gaia-gis.it/gaia-sins/spatialite-sql-4.3.0.html

	// Note the InitSpatialMetaData() command because this:
	// https://stackoverflow.com/questions/17761089/cannot-create-column-with-spatialite-unexpected-metadata-layout

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		type TEXT,
		is_alt TINYINT,
		alt_label TEXT,
		lastmodified INTEGER
	);

	SELECT InitSpatialMetaData();
	SELECT AddGeometryColumn('%s', 'geom', 4326, 'GEOMETRY', 'XY');
	SELECT CreateSpatialIndex('%s', 'geom');

	CREATE UNIQUE INDEX by_id ON %s (id, alt_label);
	CREATE INDEX geometries_by_lastmod ON %s (lastmodified);`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *GeometriesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *GeometriesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *GeometriesTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

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

	geojson_geom, err := geometry.Geometry(f)

	if err != nil {
		return MissingPropertyError(t, "geometry", err)
	}

	orb_geom := geojson_geom.Geometry()

	str_wkt := wkt.MarshalString(orb_geom)

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, is_alt, alt_label, type, geom, lastmodified
	) VALUES (
		?, ?, ?, ?, GeomFromText('%s', 4326), ?
	)`, t.Name(), str_wkt)

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	geom_type := "common"

	_, err = stmt.Exec(id, is_alt, alt_label, geom_type, lastmod)

	if err != nil {
		return ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return CommitTransactionError(t, err)
	}

	return nil
}
