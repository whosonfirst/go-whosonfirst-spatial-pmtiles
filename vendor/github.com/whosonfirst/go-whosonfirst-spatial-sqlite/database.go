package sqlite

// https://www.sqlite.org/rtree.html

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/aaronland/go-sqlite-modernc"
	"github.com/aaronland/go-sqlite/v2"
	gocache "github.com/patrickmn/go-cache"
	"github.com/paulmach/orb"
	_ "github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/planar"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial-sqlite/wkttoorb"	
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2/tables"
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
	Logger        *log.Logger
	mu            *sync.RWMutex
	db            sqlite.Database
	rtree_table   sqlite.Table
	spr_table     sqlite.Table
	geojson_table sqlite.Table
	gocache       *gocache.Cache
	dsn           string
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

	if dsn == "" {
		return nil, fmt.Errorf("Missing 'dsn' parameter")
	}

	sqlite_db, err := sqlite.NewDatabase(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new SQLite database, %w", err)
	}

	return NewSQLiteSpatialDatabaseWithDatabase(ctx, uri, sqlite_db)
}

// NewSQLiteSpatialDatabaseWithDatabase returns a new `whosonfirst/go-whosonfirst-spatial/database.database.SpatialDatabase`
// instance for performing spatial operations derived from 'uri' and an existing `aaronland/go-sqlite/database.SQLiteDatabase`
// instance defined by 'sqlite_db'.
func NewSQLiteSpatialDatabaseWithDatabase(ctx context.Context, uri string, sqlite_db sqlite.Database) (database.SpatialDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	dsn := q.Get("dsn")

	rtree_table, err := tables.NewRTreeTableWithDatabase(ctx, sqlite_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create rtree table, %w", err)
	}

	spr_table, err := tables.NewSPRTableWithDatabase(ctx, sqlite_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spr table, %w", err)
	}

	// This is so we can satisfy the reader.Reader requirement
	// in the spatial.SpatialDatabase interface

	geojson_table, err := tables.NewGeoJSONTableWithDatabase(ctx, sqlite_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create geojson table, %w", err)
	}

	logger := log.Default()

	expires := 5 * time.Minute
	cleanup := 30 * time.Minute

	gc := gocache.New(expires, cleanup)

	mu := new(sync.RWMutex)

	spatial_db := &SQLiteSpatialDatabase{
		Logger:        logger,
		db:            sqlite_db,
		rtree_table:   rtree_table,
		spr_table:     spr_table,
		geojson_table: geojson_table,
		gocache:       gc,
		dsn:           dsn,
		mu:            mu,
	}

	return spatial_db, nil
}

// Disconnect will close the underlying database connection.
func (r *SQLiteSpatialDatabase) Disconnect(ctx context.Context) error {
	return r.db.Close(ctx)
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

	conn, err := r.db.Conn(ctx)

	if err != nil {
		return fmt.Errorf("Failed to establish database connection, %w", err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return fmt.Errorf("Failed to create transaction, %w", err)
	}

	// defer tx.Rollback()

	tables := []sqlite.Table{
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
func (r *SQLiteSpatialDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rsp_ch := make(chan spr.StandardPlacesResult)
	err_ch := make(chan error)
	done_ch := make(chan bool)

	results := make([]spr.StandardPlacesResult, 0)
	working := true

	go r.PointInPolygonWithChannels(ctx, rsp_ch, err_ch, done_ch, coord, filters...)

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-done_ch:
			working = false
		case rsp := <-rsp_ch:
			results = append(results, rsp)
		case err := <-err_ch:
			return nil, fmt.Errorf("Point in polygon request failed, %w", err)
		default:
			// pass
		}

		if !working {
			break
		}
	}

	spr_results := &SQLiteResults{
		Places: results,
	}

	return spr_results, nil
}

// PointInPolygonWithChannels will perform a point in polygon query against the database for records that contain 'coord' and
// that are inclusive of any filters defined by 'filters' emitting results to 'rsp_ch' (for matches), 'err_ch' (for errors) and 'done_ch'
// (when the query is completed).
func (r *SQLiteSpatialDatabase) PointInPolygonWithChannels(ctx context.Context, rsp_ch chan spr.StandardPlacesResult, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	defer func() {
		done_ch <- true
	}()

	rows, err := r.getIntersectsByCoord(ctx, coord, filters...)

	if err != nil {
		err_ch <- fmt.Errorf("Get intersects failed, %w", err)
		return
	}

	r.inflateResultsWithChannels(ctx, rsp_ch, err_ch, rows, coord, filters...)
	return
}

// PointAndPolygonCandidates will perform a point in polygon query against the database for records that contain 'coord' and
// that are inclusive of any filters defined by 'filters' returning the list of `spatial.PointInPolygonCandidate` candidate bounding
// boxes that match an initial RTree-based spatial query.
func (r *SQLiteSpatialDatabase) PointInPolygonCandidates(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*spatial.PointInPolygonCandidate, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rsp_ch := make(chan *spatial.PointInPolygonCandidate)
	err_ch := make(chan error)
	done_ch := make(chan bool)

	candidates := make([]*spatial.PointInPolygonCandidate, 0)
	working := true

	go r.PointInPolygonCandidatesWithChannels(ctx, rsp_ch, err_ch, done_ch, coord, filters...)

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-done_ch:
			working = false
		case rsp := <-rsp_ch:
			candidates = append(candidates, rsp)
		case err := <-err_ch:
			return nil, fmt.Errorf("Point in polygon (candidates) query failed, %w", err)
		default:
			// pass
		}

		if !working {
			break
		}
	}

	return candidates, nil
}

// PointAndPolygonCandidatesWithChannels will perform a point in polygon query against the database for records that contain 'coord' and
// that are inclusive of any filters defined by 'filters' returning the list of `spatial.PointInPolygonCandidate` candidate bounding
// boxes that match an initial RTree-based spatial query emitting results to 'rsp_ch' (for matches), 'err_ch' (for errors) and 'done_ch'
// (when the query is completed).
func (r *SQLiteSpatialDatabase) PointInPolygonCandidatesWithChannels(ctx context.Context, rsp_ch chan *spatial.PointInPolygonCandidate, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	defer func() {
		done_ch <- true
	}()

	intersects, err := r.getIntersectsByCoord(ctx, coord, filters...)

	if err != nil {
		err_ch <- err
		return
	}

	for _, sp := range intersects {

		bounds := sp.Bounds()

		c := &spatial.PointInPolygonCandidate{
			Id:        sp.Id,
			FeatureId: sp.FeatureId,
			AltLabel:  sp.AltLabel,
			Bounds:    bounds,
		}

		rsp_ch <- c
	}

	return
}

// getIntersectsByCoord will return the list of `RTreeSpatialIndex` instances for records that contain 'coord' and are inclusive of any filters
// defined in 'filters'. This method derives a very small bounding box from 'coord' and then invokes the `getIntersectsByRect` method.
func (r *SQLiteSpatialDatabase) getIntersectsByCoord(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*RTreeSpatialIndex, error) {

	// how small can this be?

	padding := 0.00001

	b := coord.Bound()
	rect := b.Pad(padding)

	return r.getIntersectsByRect(ctx, &rect, filters...)
}

// getIntersectsByCoord will return the list of `RTreeSpatialIndex` instances for records that intersect 'rect' and are inclusive of any filters
// defined in 'filters'.
func (r *SQLiteSpatialDatabase) getIntersectsByRect(ctx context.Context, rect *orb.Bound, filters ...spatial.Filter) ([]*RTreeSpatialIndex, error) {

	/*
	t1 := time.Now()
	defer func(){
		log.Printf("Time to rect, %v\n", time.Since(t1))
	}()
	*/
	
	conn, err := r.db.Conn(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to establish database connection, %w", err)
	}

	q := fmt.Sprintf("SELECT id, wof_id, is_alt, alt_label, geometry, min_x, min_y, max_x, max_y FROM %s  WHERE min_x <= ? AND max_x >= ?  AND min_y <= ? AND max_y >= ?", r.rtree_table.Name())

	// Left returns the left of the bound.
	// Right returns the right of the bound.

	minx := rect.Left()
	miny := rect.Bottom()
	maxx := rect.Right()
	maxy := rect.Top()

	rows, err := conn.QueryContext(ctx, q, minx, maxx, miny, maxy)

	if err != nil {
		return nil, fmt.Errorf("SQL query failed, %w", err)
	}

	defer rows.Close()

	intersects := make([]*RTreeSpatialIndex, 0)

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
			return nil, fmt.Errorf("Result row scan failed, %w", err)
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

		intersects = append(intersects, i)
	}

	return intersects, nil
}

// inflateResultsWithChannels creates `spr.StandardPlacesResult` instances for each record defined in 'possible' emitting results
// to 'rsp_ch' (on succcess) and 'err_ch' (if there was an error).
func (r *SQLiteSpatialDatabase) inflateResultsWithChannels(ctx context.Context, rsp_ch chan spr.StandardPlacesResult, err_ch chan error, possible []*RTreeSpatialIndex, c *orb.Point, filters ...spatial.Filter) {

	/*
	t1 := time.Now()

	defer func(){
		log.Printf("Time to inflate, %v\n", time.Since(t1))
	}()
	*/
	
	seen := new(sync.Map)

	wg := new(sync.WaitGroup)

	for _, sp := range possible {

		wg.Add(1)

		go func(sp *RTreeSpatialIndex) {
			defer wg.Done()
			r.inflateSpatialIndexWithChannels(ctx, rsp_ch, err_ch, seen, sp, c, filters...)
		}(sp)
	}

	wg.Wait()
}

// inflateSpatialIndexWithChannels creates `spr.StandardPlacesResult` instance for 'sp' applying any filters defined in 'filters'
// emitting results to 'rsp_ch' (on succcess) and 'err_ch' (if there was an error). If a given record is already found in 'seen' it
// will be skipped; if not it will be added (to 'seen') once the spatial index has been successfully inflated.
func (r *SQLiteSpatialDatabase) inflateSpatialIndexWithChannels(ctx context.Context, rsp_ch chan spr.StandardPlacesResult, err_ch chan error, seen *sync.Map, sp *RTreeSpatialIndex, c *orb.Point, filters ...spatial.Filter) {

	
	select {
	case <-ctx.Done():
		return
	default:
		// pass
	}

	sp_id := fmt.Sprintf("%s:%s", sp.Id, sp.AltLabel)
	feature_id := fmt.Sprintf("%s:%s", sp.FeatureId, sp.AltLabel)

	/*
	t1 := time.Now()
	
	defer func(){
		log.Printf("[%s] Time to inflate w/ channel, %v\n", sp_id, time.Since(t1))
	}()
	*/
	
	// have we already looked up the filters for this ID?
	// see notes below

	_, ok := seen.Load(feature_id)
	
	if ok {
		return
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

		/*

	2023/08/21 22:23:01 [102087463#2084:] orb 20.368308ms
2023/08/21 22:23:01 [102087463#2084:] not-orb 4.206974ms

*/
		
		// poly, err = wkt.UnmarshalPolygon(sp.geometry)

		o, err := wkttoorb.Scan(sp.geometry)

		if err != nil {
			return
		}

		poly = o.(orb.Polygon)
	}

	if err != nil {
		return
	}
	
	// END OF maybe move all this code in to whosonfirst/go-whosonfirst-sqlite-features/tables/rtree.go
	
	if !planar.PolygonContains(poly, *c) {
		return
	}

	// there is at least one ring that contains the coord
	// now we check the filters - whether or not they pass
	// we can skip every subsequent polygon with the same
	// ID

	_, ok = seen.LoadOrStore(feature_id, true)

	if ok {
		return
	}
	
	s, err := r.retrieveSPR(ctx, sp.Path())

	if err != nil {
		r.Logger.Printf("Failed to retrieve feature cache for %s, %v", sp_id, err)
		return
	}

	for _, f := range filters {

		err = filter.FilterSPR(f, s)

		if err != nil {
			// r.Logger.Printf("SKIP %s because filter error %s", sp_id, err)
			return
		}
	}

	rsp_ch <- s
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

// Read implements the whosonfirst/go-reader interface so that the database itself can be used as a
// reader.Reader instance (reading features from the `geojson` table.
func (r *SQLiteSpatialDatabase) Read(ctx context.Context, str_uri string) (io.ReadSeekCloser, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, err
	}

	conn, err := r.db.Conn(ctx)

	if err != nil {
		return nil, err
	}

	// TO DO : ALT STUFF HERE

	q := fmt.Sprintf("SELECT body FROM %s WHERE id = ?", r.geojson_table.Name())

	row := conn.QueryRowContext(ctx, q, id)

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
	return nil
}

// SetLogger implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
