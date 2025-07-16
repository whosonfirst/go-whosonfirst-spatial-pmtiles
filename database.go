package pmtiles

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"sync"
	"time"

	_ "github.com/aaronland/gocloud-blob/s3"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"
	_ "modernc.org/sqlite"

	"github.com/protomaps/go-pmtiles/pmtiles"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-spatial-pmtiles/cache"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func init() {
	ctx := context.Background()
	database.RegisterSpatialDatabase(ctx, "pmtiles", NewPMTilesSpatialDatabase)
	reader.RegisterReader(ctx, "pmtiles", NewPMTilesSpatialDatabaseReader)
}

type PMTilesResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Places                    []spr.StandardPlacesResult `json:"places"`
}

func (r *PMTilesResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

type PMTilesSpatialDatabase struct {
	database.SpatialDatabase
	server                           *pmtiles.Server
	database                         string
	layer                            string
	enable_feature_cache             bool
	cache_manager                    cache.CacheManager
	zoom                             int
	spatial_database_uri             string
	spatial_databases_ttl            int
	spatial_databases_counter        *Counter
	spatial_databases_releaser       map[string]time.Time
	spatial_databases_cache          map[string]database.SpatialDatabase
	spatial_databases_cache_mutex    *sync.RWMutex
	spatial_databases_releaser_mutex *sync.RWMutex

	spatial_databases_ticker      *time.Ticker
	spatial_databases_ticker_done chan bool

	count_pip int64
}

func NewPMTilesSpatialDatabaseReader(ctx context.Context, uri string) (reader.Reader, error) {
	return NewPMTilesSpatialDatabase(ctx, uri)
}

func NewPMTilesSpatialDatabase(ctx context.Context, uri string) (database.SpatialDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	q_tile_path := q.Get("tiles")
	q_database := q.Get("database")
	q_layer := q.Get("layer")

	if q_layer == "" {
		q_layer = q_database
	}

	cache_size := 64
	zoom := 12

	q_cache_size := q.Get("pmtiles-cache-size")

	if q_cache_size != "" {

		sz, err := strconv.Atoi(q_cache_size)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?pmtiles-cache-size= parameter, %w", err)
		}

		cache_size = sz
	}

	q_zoom := q.Get("zoom")

	if q_zoom != "" {

		z, err := strconv.Atoi(q_zoom)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?zoom= parameter, %w", err)
		}

		zoom = z
	}

	logger := slog.Default()
	log_logger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)

	server, err := pmtiles.NewServer(q_tile_path, "", log_logger, cache_size, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create pmtiles.Loop, %w", err)
	}

	server.Start()

	spatial_databases_ttl := 30 // seconds

	if q.Has("database-ttl") {

		v, err := strconv.Atoi(q.Get("database-ttl"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?database-tll= parameter, %w", err)
		}

		spatial_databases_ttl = v
	}

	spatial_databases_counter := NewCounter()

	spatial_databases_releaser := make(map[string]time.Time)
	spatial_databases_releaser_mutex := new(sync.RWMutex)

	spatial_databases_cache := make(map[string]database.SpatialDatabase)
	spatial_databases_cache_mutex := new(sync.RWMutex)

	spatial_databases_ticker := time.NewTicker(time.Duration(spatial_databases_ttl) * time.Second)
	spatial_databases_ticker_done := make(chan bool)

	// To do: Check for query value

	// This triggers "distance errors" which I don't really understand yet
	// spatial_database_uri := "rtree://"

	// Note the {dbname}. This gets swapped out in spatialDatabaseFromTile.
	// That's important because it allows the creation of discrete databases
	// in memory which can be disconnected/deleted in order to free up memory.
	dsn := url.QueryEscape("file:{dbname}?mode=memory&cache=shared")

	spatial_database_uri := fmt.Sprintf("sqlite://sqlite?dsn=%s", dsn)

	db := &PMTilesSpatialDatabase{
		server:                           server,
		database:                         q_database,
		layer:                            q_layer,
		zoom:                             zoom,
		spatial_database_uri:             spatial_database_uri,
		spatial_databases_ttl:            spatial_databases_ttl,
		spatial_databases_counter:        spatial_databases_counter,
		spatial_databases_releaser:       spatial_databases_releaser,
		spatial_databases_releaser_mutex: spatial_databases_releaser_mutex,
		spatial_databases_cache:          spatial_databases_cache,
		spatial_databases_cache_mutex:    spatial_databases_cache_mutex,
		spatial_databases_ticker:         spatial_databases_ticker,
		spatial_databases_ticker_done:    spatial_databases_ticker_done,
		count_pip:                        int64(0),
	}

	go func() {

		for {
			select {
			case <-db.spatial_databases_ticker_done:
				return
			case <-spatial_databases_ticker.C:
				db.pruneSpatialDatabases(ctx)
			}
		}

	}()

	//

	enable_feature_cache := false

	q_enable_cache := q.Get("enable-cache")

	if q_enable_cache != "" {

		enabled, err := strconv.ParseBool(q_enable_cache)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?enable-cache= parameter, %w", err)
		}

		enable_feature_cache = enabled

	}

	if enable_feature_cache {

		// To do: Read from DB URI. Given that we weren't doing this for the original
		// docstore/mem stuff it seems okay just to swap out defaults. Note that we
		// are using https://pkg.go.dev/modernc.org/sqlite which is assumed to have
		// already been loaded (by go-whosonnfirst-spatial-sqlite)

		cache_manager_uri := "sql://sqlite?dsn={tmp}"
		cache_manager, err := cache.NewCacheManager(ctx, cache_manager_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create cache manager, %w", err)
		}

		db.cache_manager = cache_manager
		db.enable_feature_cache = enable_feature_cache
	}

	return db, nil
}
