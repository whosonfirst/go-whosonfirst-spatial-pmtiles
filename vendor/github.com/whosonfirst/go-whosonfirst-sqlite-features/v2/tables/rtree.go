package tables

// https://www.sqlite.org/rtree.html

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
	_ "log"
)

type RTreeTableOptions struct {
	IndexAltFiles bool
}

func DefaultRTreeTableOptions() (*RTreeTableOptions, error) {

	opts := RTreeTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type RTreeTable struct {
	features.FeatureTable
	name    string
	options *RTreeTableOptions
}

func NewRTreeTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultRTreeTableOptions()

	if err != nil {
		return nil, err
	}

	return NewRTreeTableWithOptions(ctx, opts)
}

func NewRTreeTableWithOptions(ctx context.Context, opts *RTreeTableOptions) (sqlite.Table, error) {

	t := RTreeTable{
		name:    "rtree",
		options: opts,
	}

	return &t, nil
}

func NewRTreeTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultRTreeTableOptions()

	if err != nil {
		return nil, err
	}

	return NewRTreeTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewRTreeTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *RTreeTableOptions) (sqlite.Table, error) {

	t, err := NewRTreeTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *RTreeTable) Name() string {
	return t.name
}

func (t *RTreeTable) Schema() string {

	/*

		3.1.1. Column naming details

		In the argments to "rtree" in the CREATE VIRTUAL TABLE statement, the names of the columns are taken from the first token of each argument. All subsequent tokens within each argument are silently ignored. This means, for example, that if you try to give a column a type affinity or add a constraint such as UNIQUE or NOT NULL or DEFAULT to a column, those extra tokens are accepted as valid, but they do not change the behavior of the rtree. In an RTREE virtual table, the first column always has a type affinity of INTEGER and all other data columns have a type affinity of NUMERIC.

		Recommended practice is to omit any extra tokens in the rtree specification. Let each argument to "rtree" be a single ordinary label that is the name of the corresponding column, and omit all other tokens from the argument list.

		4.1. Auxiliary Columns

		Beginning with SQLite version 3.24.0 (2018-06-04), r-tree tables can have auxiliary columns that store arbitrary data. Auxiliary columns can be used in place of secondary tables such as "demo_data".

		Auxiliary columns are marked with a "+" symbol before the column name. Auxiliary columns must come after all of the coordinate boundary columns. There is a limit of no more than 100 auxiliary columns. The following example shows an r-tree table with auxiliary columns that is equivalent to the two tables "demo_index" and "demo_data" above:

		Note: Auxiliary columns must come at the end of a table definition
	*/

	sql := `CREATE VIRTUAL TABLE %s USING rtree (
		id,
		min_x,
		max_x,
		min_y,
		max_y,
		+wof_id INTEGER,
		+is_alt TINYINT,
		+alt_label TEXT,
		+geometry BLOB,
		+lastmodified INTEGER
	);`

	return fmt.Sprintf(sql, t.Name())
}

func (t *RTreeTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *RTreeTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *RTreeTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

	is_alt := alt.IsAlt(f) // this returns a boolean which is interpreted as a float by SQLite

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	geom_type, err := geometry.Type(f)

	if err != nil {
		return MissingPropertyError(t, "geometry type", err)
	}

	switch geom_type {
	case "Polygon", "MultiPolygon":
		// pass
	default:
		return nil
	}

	wof_id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	alt_label := ""

	if is_alt {

		label, err := properties.AltLabel(f)

		if err != nil {
			return MissingPropertyError(t, "alt label", err)
		}

		alt_label = label
	}

	lastmod := properties.LastModified(f)

	geojson_geom, err := geometry.Geometry(f)

	if err != nil {
		return MissingPropertyError(t, "geometry", err)
	}

	orb_geom := geojson_geom.Geometry()

	conn, err := db.Conn(ctx)

	if err != nil {
		return DatabaseConnectionError(t, err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, min_x, max_x, min_y, max_y, wof_id, is_alt, alt_label, geometry, lastmodified
	) VALUES (
		NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	var mp orb.MultiPolygon

	switch geom_type {
	case "MultiPolygon":
		mp = orb_geom.(orb.MultiPolygon)
	case "Polygon":
		mp = []orb.Polygon{orb_geom.(orb.Polygon)}
	default:
		// This should never happen (we check above) but just in case...
		return WrapError(t, fmt.Errorf("Invalid or unsupported geometry type, %s", geom_type))
	}

	for _, poly := range mp {

		// Store the geometry for each bounding box so we can use it to do
		// raycasting and filter points in any interior rings. For example in
		// whosonfirst/go-whosonfirst-spatial-sqlite

		bbox := poly.Bound()

		sw := bbox.Min
		ne := bbox.Max

		enc_geom := wkt.MarshalString(poly)

		_, err = stmt.Exec(sw.X(), ne.X(), sw.Y(), ne.Y(), wof_id, is_alt, alt_label, enc_geom, lastmod)

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
