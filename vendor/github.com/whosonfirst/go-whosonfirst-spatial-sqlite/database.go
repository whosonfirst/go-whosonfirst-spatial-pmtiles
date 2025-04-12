package sqlite

// https://www.sqlite.org/rtree.html

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-database/sql/tables"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial-sqlite/wkttoorb"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	sqlite_spr "github.com/whosonfirst/go-whosonfirst-sqlite-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

func init() {
	ctx := context.Background()
	database.RegisterSpatialDatabase(ctx, "sqlite", NewSQLiteSpatialDatabase)
	reader.RegisterReader(ctx, "sqlite", NewSQLiteSpatialDatabaseReader)
	writer.RegisterWriter(ctx, "sqlite", NewSQLiteSpatialDatabaseWriter)
}

// SQLiteSpatialDatabase is a struct that implements the `database.SpatialDatabase` for performing
// spatial queries on data stored in a SQLite databases from tables defined by the `whosonfirst/go-whosonfirst-sqlite-features/tables`
// package.
type SQLiteSpatialDatabase struct {
	database.SpatialDatabase
	mu            *sync.RWMutex
	db            *sql.DB
	rtree_table   database_sql.Table
	spr_table     database_sql.Table
	geojson_table database_sql.Table
	gocache       *gocache.Cache
	dsn           string
	is_tmp        bool
	tmp_path      string
}

// RTreeSpatialIndex is a struct representing an RTree based spatial index
type RTreeSpatialIndex struct {
	geometry  string
	bounds    orb.Bound
	Id        string
	FeatureId string
	// A boolean flag indicating whether the feature associated with the index is an alternate geometry.
	IsAlt bool
	// The label for the feature (associated with the index) if it is an alternate geometry.
	AltLabel string
}

func (sp RTreeSpatialIndex) Bounds() orb.Bound {
	return sp.bounds
}

func (sp RTreeSpatialIndex) Path() string {

	if sp.IsAlt {
		return fmt.Sprintf("%s-alt-%s", sp.FeatureId, sp.AltLabel)
	}

	return sp.FeatureId
}

// SQLiteResults is a struct that implements the `whosonfirst/go-whosonfirst-spr.StandardPlacesResults`
// interface for rows matching a spatial query.
type SQLiteResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	// Places is the list of `whosonfirst/go-whosonfirst-spr.StandardPlacesResult` instances returned for a spatial query.
	Places []spr.StandardPlacesResult `json:"places"`
}

// Results returns a `whosonfirst/go-whosonfirst-spr.StandardPlacesResults` instance for rows matching a spatial query.
func (r *SQLiteResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

func NewSQLiteSpatialDatabaseReader(ctx context.Context, uri string) (reader.Reader, error) {
	return NewSQLiteSpatialDatabase(ctx, uri)
}

func NewSQLiteSpatialDatabaseWriter(ctx context.Context, uri string) (writer.Writer, error) {
	return NewSQLiteSpatialDatabase(ctx, uri)
}

// NewSQLiteSpatialDatabase returns a new `whosonfirst/go-whosonfirst-spatial/database.database.SpatialDatabase`
// instance for performing spatial operations derived from 'uri'.
func NewSQLiteSpatialDatabase(ctx context.Context, uri string) (database.SpatialDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()
	dsn := q.Get("dsn")

	is_tmp := false
	tmp_path := ""

	if dsn == "{tmp}" {

		f, err := os.CreateTemp("", ".db")

		if err != nil {
			return nil, fmt.Errorf("Failed to create temp file, %w", err)
		}

		tmp_path = f.Name()
		is_tmp = true

		q.Del("dsn")
		q.Set("dsn", tmp_path)

		u.RawQuery = q.Encode()
		uri = u.String()
	}

	db, err := database_sql.OpenWithURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new database, %w", err)
	}

	spatial_db, err := NewSQLiteSpatialDatabaseWithDatabase(ctx, uri, db)

	if err != nil {
		return nil, err
	}

	if is_tmp {
		spatial_db.(*SQLiteSpatialDatabase).is_tmp = is_tmp
		spatial_db.(*SQLiteSpatialDatabase).tmp_path = tmp_path
	}

	return spatial_db, nil
}

// NewSQLiteSpatialDatabaseWithDatabase returns a new `whosonfirst/go-whosonfirst-spatial/database.database.SpatialDatabase`
// instance for performing spatial operations derived from 'uri' and an existing `aaronland/go-sqlite/database.SQLiteDatabase`
// instance defined by 'sqlite_db'.
func NewSQLiteSpatialDatabaseWithDatabase(ctx context.Context, uri string, db *sql.DB) (database.SpatialDatabase, error) {

	rtree_table, err := tables.NewRTreeTableWithDatabase(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create rtree table, %w", err)
	}

	spr_table, err := tables.NewSPRTableWithDatabase(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spr table, %w", err)
	}

	// This is so we can satisfy the reader.Reader requirement
	// in the spatial.SpatialDatabase interface

	geojson_table, err := tables.NewGeoJSONTableWithDatabase(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create geojson table, %w", err)
	}

	db_opts := database_sql.DefaultConfigureDatabaseOptions()

	db_opts.Tables = []database_sql.Table{
		rtree_table,
		spr_table,
		geojson_table,
	}

	db_opts.CreateTablesIfNecessary = true

	err = database_sql.ConfigureDatabase(ctx, db, db_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to configure database, %w", err)
	}

	expires := 5 * time.Minute
	cleanup := 30 * time.Minute

	gc := gocache.New(expires, cleanup)

	mu := new(sync.RWMutex)

	spatial_db := &SQLiteSpatialDatabase{
		db:            db,
		rtree_table:   rtree_table,
		spr_table:     spr_table,
		geojson_table: geojson_table,
		gocache:       gc,
		mu:            mu,
	}

	return spatial_db, nil
}

// Disconnect will close the underlying database connection.
func (r *SQLiteSpatialDatabase) Disconnect(ctx context.Context) error {

	if r.is_tmp {

		err := os.Remove(r.tmp_path)

		if err != nil {
			slog.Error("Failed to remove tmp db", "pth", r.tmp_path, "error", err)
		}
	}

	return r.db.Close()
}

// IndexFeature will index a Who's On First GeoJSON Feature record, defined in 'body', in the spatial database.
func (r *SQLiteSpatialDatabase) IndexFeature(ctx context.Context, body []byte) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.rtree_table.IndexRecord(ctx, r.db, body)

	if err != nil {
		return fmt.Errorf("Failed to index record in rtree table, %w", err)
	}

	err = r.spr_table.IndexRecord(ctx, r.db, body)

	if err != nil {
		return fmt.Errorf("Failed to index record in spr table, %w", err)
	}

	if r.geojson_table != nil {

		err = r.geojson_table.IndexRecord(ctx, r.db, body)

		if err != nil {
			return fmt.Errorf("Failed to index record in geojson table, %w", err)
		}
	}

	return nil
}

// RemoveFeature will remove the database record with ID 'id' from the database.
func (r *SQLiteSpatialDatabase) RemoveFeature(ctx context.Context, str_id string) error {

	id, err := strconv.ParseInt(str_id, 10, 64)

	if err != nil {
		return fmt.Errorf("Failed to parse string ID '%s', %w", str_id, err)
	}

	tx, err := r.db.Begin()

	if err != nil {
		return fmt.Errorf("Failed to create transaction, %w", err)
	}

	// defer tx.Rollback()

	tables := []database_sql.Table{
		r.rtree_table,
		r.spr_table,
	}

	if r.geojson_table != nil {
		tables = append(tables, r.geojson_table)
	}

	for _, t := range tables {

		var q string

		switch t.Name() {
		case "rtree":
			q = fmt.Sprintf("DELETE FROM %s WHERE wof_id = ?", t.Name())
		default:
			q = fmt.Sprintf("DELETE FROM %s WHERE id = ?", t.Name())
		}

		stmt, err := tx.Prepare(q)

		if err != nil {
			return fmt.Errorf("Failed to create query statement for %s, %w", t.Name(), err)
		}

		_, err = stmt.ExecContext(ctx, id)

		if err != nil {
			return fmt.Errorf("Failed execute query statement for %s, %w", t.Name(), err)
		}
	}

	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("Failed to commit transaction, %w", err)
	}

	return nil
}

// PointInPolygon will perform a point in polygon query against the database for records that contain 'coord' and
// that are inclusive of any filters defined by 'filters'.
func (db *SQLiteSpatialDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	results := make([]spr.StandardPlacesResult, 0)

	for r, err := range db.PointInPolygonWithIterator(ctx, coord, filters...) {

		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	spr_results := &SQLiteResults{
		Places: results,
	}

	return spr_results, nil
}

func (db *SQLiteSpatialDatabase) PointInPolygonWithIterator(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		rows, err := db.getIntersectsByCoord(ctx, coord, filters...)

		if err != nil {
			yield(nil, err)
			return
		}

		for r, err := range db.inflateResults(ctx, rows, coord, filters...) {

			if !yield(r, err) {
				return
			}
		}

		return
	}
}

func (db *SQLiteSpatialDatabase) Intersects(ctx context.Context, geom orb.Geometry, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	results := make([]spr.StandardPlacesResult, 0)

	for r, err := range db.IntersectsWithIterator(ctx, geom, filters...) {

		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	spr_results := &SQLiteResults{
		Places: results,
	}

	return spr_results, nil
}

func (db *SQLiteSpatialDatabase) IntersectsWithIterator(ctx context.Context, geom orb.Geometry, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		bound := geom.Bound()

		rows, err := db.getIntersectsByRect(ctx, &bound, filters...)

		if err != nil {
			yield(nil, err)
			return
		}

		// Do not return (yield) the same ID multiple times
		seen := new(sync.Map)

		for r, err := range db.inflateIntersectsResults(ctx, rows, geom, filters...) {

			if err != nil {
				yield(nil, err)
				break
			}

			_, exists := seen.Load(r.Id())

			if exists {
				continue
			}

			seen.Store(r.Id(), true)

			if !yield(r, nil) {
				break
			}
		}

		return
	}
}

// getIntersectsByCoord will return the list of `RTreeSpatialIndex` instances for records that contain 'coord' and are inclusive of any filters
// defined in 'filters'. This method derives a very small bounding box from 'coord' and then invokes the `getIntersectsByRect` method.
func (db *SQLiteSpatialDatabase) getIntersectsByCoord(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*RTreeSpatialIndex, error) {

	// how small can this be?

	padding := 0.00001

	b := coord.Bound()
	rect := b.Pad(padding)

	return db.getIntersectsByRect(ctx, &rect, filters...)
}

// getIntersectsByCoord will return the list of `RTreeSpatialIndex` instances for records that intersect 'rect' and are inclusive of any filters
// defined in 'filters'.
func (db *SQLiteSpatialDatabase) getIntersectsByRect(ctx context.Context, rect *orb.Bound, filters ...spatial.Filter) ([]*RTreeSpatialIndex, error) {

	logger := slog.Default()
	logger = logger.With("query", "intersects by rect")
	logger = logger.With("center", rect.Center())

	q := fmt.Sprintf("SELECT id, wof_id, is_alt, alt_label, geometry, min_x, min_y, max_x, max_y FROM %s  WHERE min_x <= ? OR max_x >= ?  OR min_y <= ? OR max_y >= ?", db.rtree_table.Name())

	// Left returns the left of the bound.
	// Right returns the right of the bound.

	minx := rect.Left()
	miny := rect.Bottom()
	maxx := rect.Right()
	maxy := rect.Top()

	rows, err := db.db.QueryContext(ctx, q, minx, maxx, miny, maxy)

	if err != nil {
		return nil, fmt.Errorf("SQL query failed, %w", err)
	}

	defer rows.Close()

	intersects := make([]*RTreeSpatialIndex, 0)

	for i, err := range db.rowsToSpatialIndices(ctx, rows, filters...) {

		if err != nil {
			return nil, err
		}

		intersects = append(intersects, i)
	}

	logger.Debug("Intersects by rect candidates", "r", rect, "count", len(intersects))
	return intersects, nil
}

// inflateResults creates `spr.StandardPlacesResult` instances for each record defined in 'possible'.
func (db *SQLiteSpatialDatabase) inflateResults(ctx context.Context, possible []*RTreeSpatialIndex, c *orb.Point, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		seen := new(sync.Map)

		for _, sp := range possible {

			// sp_id := fmt.Sprintf("%s:%s", sp.Id, sp.AltLabel)
			feature_id := fmt.Sprintf("%s:%s", sp.FeatureId, sp.AltLabel)

			logger := slog.Default()
			logger = logger.With("feature id", feature_id)
			logger = logger.With("latitude", c.Y())
			logger = logger.With("longitude", c.X())

			logger.Debug("Inflate spatial index")

			// have we already looked up the filters for this ID?
			// see notes below

			_, ok := seen.Load(feature_id)

			if ok {
				continue
			}

			// START OF maybe move all this code in to whosonfirst/go-whosonfirst-sqlite-features/tables/rtree.go

			var poly orb.Polygon
			var err error

			// This is to account for version of the whosonfirst/go-whosonfirst-sqlite-features
			// package < 0.10.0 that stored geometries as JSON-encoded strings. Subsequent versions
			// use WKT encoding.

			// This is the bottleneck. It appears to be this:
			// https://github.com/paulmach/orb/issues/132
			// maybe... https://github.com/Succo/wktToOrb/ ?

			if strings.HasPrefix(sp.geometry, "[[[") {
				// Investigate https://github.com/paulmach/orb/tree/master/geojson#performance
				err = json.Unmarshal([]byte(sp.geometry), &poly)
			} else {

				// poly, err = wkt.UnmarshalPolygon(sp.geometry)

				o, err := wkttoorb.Scan(sp.geometry)

				if err != nil {
					continue
				}

				poly = o.(orb.Polygon)
			}

			if err != nil {
				logger.Error("Failed to derive polygon", "error", err)
				continue
			}

			// END OF maybe move all this code in to whosonfirst/go-whosonfirst-sqlite-features/tables/rtree.go

			if !planar.PolygonContains(poly, *c) {
				logger.Debug("Coordinate not contained by feature polygon")
				continue
			}

			// there is at least one ring that contains the coord
			// now we check the filters - whether or not they pass
			// we can skip every subsequent polygon with the same
			// ID

			_, ok = seen.LoadOrStore(feature_id, true)

			if ok {
				continue
			}

			s, err := db.retrieveSPR(ctx, sp.Path())

			if err != nil {
				logger.Error("Failed to retrieve feature cache", "key", sp.Path(), "error", err)
				continue
			}

			matches := true

			for _, f := range filters {

				err = filter.FilterSPR(f, s)

				if err != nil {
					slog.Debug("Feature failed SPR filter", "feature_id", feature_id, "error", err)
					matches = false
					break
				}
			}

			if !matches {
				continue
			}

			logger.Debug("Return inflated SPR", "id", s.Id())
			yield(s, nil)
		}
	}
}

// inflateResults creates `spr.StandardPlacesResult` instances for each record defined in 'possible'.
func (db *SQLiteSpatialDatabase) inflateIntersectsResults(ctx context.Context, possible []*RTreeSpatialIndex, geom orb.Geometry, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		seen := new(sync.Map)

		for _, sp := range possible {

			// sp_id := fmt.Sprintf("%s:%s", sp.Id, sp.AltLabel)
			feature_id := fmt.Sprintf("%s:%s", sp.FeatureId, sp.AltLabel)

			logger := slog.Default()
			logger = logger.With("feature id", feature_id)
			logger = logger.With("geometry", geom.GeoJSONType())

			logger.Debug("Inflate spatial index")

			// have we already looked up the filters for this ID?
			// see notes below

			_, ok := seen.Load(feature_id)

			if ok {
				continue
			}

			// START OF maybe move all this code in to whosonfirst/go-whosonfirst-sqlite-features/tables/rtree.go

			var poly orb.Polygon
			var err error

			// This is to account for version of the whosonfirst/go-whosonfirst-sqlite-features
			// package < 0.10.0 that stored geometries as JSON-encoded strings. Subsequent versions
			// use WKT encoding.

			// This is the bottleneck. It appears to be this:
			// https://github.com/paulmach/orb/issues/132
			// maybe... https://github.com/Succo/wktToOrb/ ?

			if strings.HasPrefix(sp.geometry, "[[[") {
				// Investigate https://github.com/paulmach/orb/tree/master/geojson#performance
				err = json.Unmarshal([]byte(sp.geometry), &poly)
			} else {

				// poly, err = wkt.UnmarshalPolygon(sp.geometry)

				o, err := wkttoorb.Scan(sp.geometry)

				if err != nil {
					continue
				}

				poly = o.(orb.Polygon)
			}

			if err != nil {
				logger.Error("Failed to derive polygon", "error", err)
				continue
			}

			// END OF maybe move all this code in to whosonfirst/go-whosonfirst-sqlite-features/tables/rtree.go

			intersects := false

			ok, err = geo.Intersects(poly, geom)

			if err != nil {
				logger.Error("Failed to determine intersection", "error", err)
				continue
			}

			intersects = ok

			if !intersects {
				continue
			}

			// there is at least one ring that contains the coord
			// now we check the filters - whether or not they pass
			// we can skip every subsequent polygon with the same
			// ID

			_, ok = seen.LoadOrStore(feature_id, true)

			if ok {
				continue
			}

			s, err := db.retrieveSPR(ctx, sp.Path())

			if err != nil {
				logger.Error("Failed to retrieve feature cache", "key", sp.Path(), "error", err)
				continue
			}

			matches := true

			for _, f := range filters {

				err = filter.FilterSPR(f, s)

				if err != nil {
					slog.Debug("Feature failed SPR filter", "feature_id", feature_id, "error", err)
					matches = false
					break
				}
			}

			if !matches {
				continue
			}

			logger.Debug("Return inflated SPR", "id", s.Id())
			yield(s, nil)
		}
	}
}

// retrieveSPR retrieves a `spr.StandardPlacesResult` instance from the local database cache identified by 'uri_str'.
func (r *SQLiteSpatialDatabase) retrieveSPR(ctx context.Context, uri_str string) (spr.StandardPlacesResult, error) {

	c, ok := r.gocache.Get(uri_str)

	if ok {
		return c.(*sqlite_spr.SQLiteStandardPlacesResult), nil
	}

	id, uri_args, err := uri.ParseURI(uri_str)

	if err != nil {
		return nil, err
	}

	alt_label := ""

	if uri_args.IsAlternate {

		source, err := uri_args.AltGeom.String()

		if err != nil {
			return nil, err
		}

		alt_label = source
	}

	s, err := sqlite_spr.RetrieveSPR(ctx, r.db, r.spr_table, id, alt_label)

	if err != nil {
		return nil, err
	}

	r.gocache.Set(uri_str, s, -1)
	return s, nil
}

func (db *SQLiteSpatialDatabase) rowsToSpatialIndices(ctx context.Context, rows *sql.Rows, filters ...spatial.Filter) iter.Seq2[*RTreeSpatialIndex, error] {

	return func(yield func(*RTreeSpatialIndex, error) bool) {

		for rows.Next() {

			var id string
			var feature_id string
			var is_alt int32
			var alt_label string
			var geometry string
			var minx float64
			var miny float64
			var maxx float64
			var maxy float64

			err := rows.Scan(&id, &feature_id, &is_alt, &alt_label, &geometry, &minx, &miny, &maxx, &maxy)

			if err != nil {
				yield(nil, fmt.Errorf("Result row scan failed, %w", err))
				break
			}

			min := orb.Point{minx, miny}
			max := orb.Point{maxx, maxy}

			rect := orb.Bound{
				Min: min,
				Max: max,
			}

			i := &RTreeSpatialIndex{
				Id:        fmt.Sprintf("%s#%s", feature_id, id),
				FeatureId: feature_id,
				bounds:    rect,
				geometry:  geometry,
			}

			if is_alt == 1 {
				i.IsAlt = true
				i.AltLabel = alt_label
			}

			yield(i, nil)
		}

	}
}

// Read implements the whosonfirst/go-reader interface so that the database itself can be used as a
// reader.Reader instance (reading features from the `geojson` table.
func (r *SQLiteSpatialDatabase) Read(ctx context.Context, str_uri string) (io.ReadSeekCloser, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, err
	}

	// TO DO : ALT STUFF HERE

	q := fmt.Sprintf("SELECT body FROM %s WHERE id = ?", r.geojson_table.Name())

	row := r.db.QueryRowContext(ctx, q, id)

	var body string

	err = row.Scan(&body)

	if err != nil {
		return nil, err
	}

	sr := strings.NewReader(body)
	fh, err := ioutil.NewReadSeekCloser(sr)

	if err != nil {
		return nil, err
	}

	return fh, nil
}

// ReadURI implements the whosonfirst/go-reader interface so that the database itself can be used as a
// reader.Reader instance
func (r *SQLiteSpatialDatabase) ReaderURI(ctx context.Context, str_uri string) string {
	return str_uri
}

// Write implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance (by invoking the `IndexFeature` method).
func (r *SQLiteSpatialDatabase) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return 0, err
	}

	err = r.IndexFeature(ctx, body)

	if err != nil {
		return 0, err
	}

	return int64(len(body)), nil
}

// WriterURI implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance
func (r *SQLiteSpatialDatabase) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

// Flush implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) Flush(ctx context.Context) error {
	return nil
}

// Close implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) Close(ctx context.Context) error {
	return r.db.Close()
}

// SetLogger implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
