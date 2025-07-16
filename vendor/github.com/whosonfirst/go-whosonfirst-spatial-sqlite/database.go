package sqlite

// https://www.sqlite.org/rtree.html

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/paulmach/orb"
	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-database/sql/tables"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
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
