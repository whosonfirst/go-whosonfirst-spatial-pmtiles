package pmtiles

import ()

import (
	_ "github.com/aaronland/gocloud-blob-s3"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"	
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"
)

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	aa_docstore "github.com/aaronland/gocloud-docstore"
	"github.com/jtacoma/uritemplates"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/protomaps/go-pmtiles/pmtiles"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"gocloud.dev/docstore"
)

func init() {
	ctx := context.Background()
	database.RegisterSpatialDatabase(ctx, "pmtiles", NewPMTilesSpatialDatabase)
	reader.RegisterReader(ctx, "pmtiles", NewPMTilesSpatialDatabaseReader)
}

type PMTilesSpatialDatabase struct {
	database.SpatialDatabase
	server                      *pmtiles.Server
	database                    string
	layer                       string
	enable_feature_cache        bool
	cache_manager               *CacheManager
	zoom                        int
	spatial_database_uri        string
	spatial_databases_counter   *sync.Map
	spatial_databases_increment int32
	spatial_databases_cache     *sync.Map
	spatial_databases_mutex     *sync.RWMutex
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
	log_logger := slog.NewLogLogger(logger.Handler(), slog.LevelInfo)

	server, err := pmtiles.NewServer(q_tile_path, "", log_logger, cache_size, "", "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create pmtiles.Loop, %w", err)
	}

	server.Start()

	spatial_databases_counter := new(sync.Map)
	spatial_databases_increment := int32(0)
	spatial_databases_cache := new(sync.Map)
	spatial_databases_mutex := new(sync.RWMutex)

	// To do: Check for query value

	// This triggers "distance errors" which I don't really understand yet
	// spatial_database_uri := "rtree://"
	spatial_database_uri := "sqlite://?dsn=modernc://mem"

	db := &PMTilesSpatialDatabase{
		server:                      server,
		database:                    q_database,
		layer:                       q_layer,
		zoom:                        zoom,
		spatial_database_uri:        spatial_database_uri,
		spatial_databases_counter:   spatial_databases_counter,
		spatial_databases_increment: spatial_databases_increment,
		spatial_databases_cache:     spatial_databases_cache,
		spatial_databases_mutex:     spatial_databases_mutex,
	}

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

		cache_ttl := 3600

		q_cache_ttl := q.Get("cache-ttl")

		if q_cache_ttl != "" {

			ttl, err := strconv.Atoi(q_cache_ttl)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse ?cache-ttl= parameter, %w", err)
			}

			if ttl < 0 {
				return nil, fmt.Errorf("Invalid cache-ttl value")
			}

			cache_ttl = ttl
		}

		feature_cache_uri_t := fmt.Sprintf("mem://%s/{key}", FEATURES_CACHE_TABLE)

		q_feature_cache_uri_t := q.Get("feature-cache-uri")

		if q_feature_cache_uri_t != "" {
			feature_cache_uri_t = q_feature_cache_uri_t
		}

		feature_cache_key := "Id"

		feature_cache_v := map[string]interface{}{
			"key": feature_cache_key,
		}

		feature_cache, err := openCollection(ctx, feature_cache_uri_t, feature_cache_v)

		if err != nil {
			return nil, fmt.Errorf("could not open feature cache collection: %w", err)
		}

		cache_manager_opts := &CacheManagerOptions{
			FeatureCollection: feature_cache,
			CacheTTL:          cache_ttl,
		}

		cache_manager := NewCacheManager(ctx, cache_manager_opts)
		db.cache_manager = cache_manager

		db.enable_feature_cache = enable_feature_cache
	}

	return db, nil
}

func (db *PMTilesSpatialDatabase) IndexFeature(context.Context, []byte) error {
	return fmt.Errorf("Not implemented.")
}

func (db *PMTilesSpatialDatabase) RemoveFeature(context.Context, string) error {
	return fmt.Errorf("Not implemented.")
}

func (db *PMTilesSpatialDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	/*

		$> ./bin/server -tile-path file:///usr/local/whosonfirst/go-whosonfirst-tippecanoe -enable-example -example-database wof
		2022/11/24 14:41:32 Listening for requests on http://localhost:8080
		2022/11/24 14:41:48 fetching wof 0-16384
		2022/11/24 14:41:48 fetched wof 0-0
		2022/11/24 14:41:48 fetching wof 39541-13802
		2022/11/24 14:41:48 fetched wof 39541-13802
		2022/11/24 14:41:48 [200] served /wof/8/41/98.mvt in 3.485603ms

		> go run cmd/query/main.go -spatial-database-uri 'pmtiles://?tiles=file:///usr/local/whosonfirst/go-whosonfirst-tippecanoe&database=wof'
		2022/11/25 18:33:32 fetching wof 0-16384
		2022/11/25 18:33:32 fetched wof 0-0
		2022/11/25 18:33:32 fetching wof 39541-13802
		2022/11/25 18:33:32 fetched wof 39541-13802
		map[wof:0xc0001005a0]

	*/

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	defer db.releaseSpatialDatabase(ctx, coord)

	return spatial_db.PointInPolygon(ctx, coord, filters...)
}

func (db *PMTilesSpatialDatabase) releaseSpatialDatabase(ctx context.Context, coord *orb.Point) {

	db.spatial_databases_mutex.Lock()
	defer db.spatial_databases_mutex.Unlock()

	db_name := db.spatialDatabaseNameFromCoord(ctx, coord)

	new_increment := atomic.AddInt32(&db.spatial_databases_increment, 1)
	db.spatial_databases_counter.Store(db_name, new_increment)

	go func(test_increment int32) {

		ttl := 10

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(ttl) * time.Second):

			db.spatial_databases_mutex.Lock()
			defer db.spatial_databases_mutex.Unlock()

			db_name := db.spatialDatabaseNameFromCoord(ctx, coord)
			current_increment := int32(0)

			v, exists := db.spatial_databases_counter.Load(db_name)

			if exists {
				current_increment = v.(int32)
			}

			if current_increment >= test_increment {
				return
			}

			db_v, exists := db.spatial_databases_cache.Load(db_name)

			if !exists {
				return
			}

			spatial_db := db_v.(database.SpatialDatabase)
			spatial_db.Disconnect(ctx)
			db.spatial_databases_cache.Delete(db_name)

			return
		}

	}(new_increment)

}

func (db *PMTilesSpatialDatabase) PointInPolygonCandidates(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*spatial.PointInPolygonCandidate, error) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	defer db.releaseSpatialDatabase(ctx, coord)
	
	return spatial_db.PointInPolygonCandidates(ctx, coord, filters...)
}

func (db *PMTilesSpatialDatabase) PointInPolygonWithChannels(ctx context.Context, spr_ch chan spr.StandardPlacesResult, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to create spatial database, %w", err)
		return
	}

	defer db.releaseSpatialDatabase(ctx, coord)

	spatial_db.PointInPolygonWithChannels(ctx, spr_ch, err_ch, done_ch, coord, filters...)
}

func (db *PMTilesSpatialDatabase) PointInPolygonCandidatesWithChannels(ctx context.Context, pip_ch chan *spatial.PointInPolygonCandidate, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to create spatial database, %w", err)
		return
	}

	defer db.releaseSpatialDatabase(ctx, coord)

	spatial_db.PointInPolygonCandidatesWithChannels(ctx, pip_ch, err_ch, done_ch, coord, filters...)
}

func (db *PMTilesSpatialDatabase) Disconnect(ctx context.Context) error {

	db.spatial_databases_cache.Range(func(k interface{}, v interface{}) bool {
		spatial_db := v.(database.SpatialDatabase)
		spatial_db.Disconnect(ctx)
		db.spatial_databases_cache.Delete(k.(string))
		return true
	})

	db.cache_manager.Close(ctx)
	return nil
}

// Read implements the whosonfirst/go-reader.Reader interface
func (db *PMTilesSpatialDatabase) Read(ctx context.Context, path string) (io.ReadSeekCloser, error) {

	if !db.enable_feature_cache {
		return nil, fmt.Errorf("Not found")
	}

	id, uri_args, err := uri.ParseURI(path)

	if err != nil {
		return nil, fmt.Errorf("Failed to path %s, %w", path, err)
	}

	fname, err := uri.Id2Fname(id, uri_args)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive filename from %s, %w", path, err)
	}

	fname = strings.Replace(fname, ".geojson", "", 1)

	fc, err := db.cache_manager.GetFeatureCache(ctx, fname)

	if err != nil {
		return nil, fmt.Errorf("Failed to read feature from cache for %s, %w", path, err)
	}

	r := strings.NewReader(fc.Body)

	rsc, err := ioutil.NewReadSeekCloser(r)

	if err != nil {
		return nil, fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", path, err)
	}

	return rsc, nil
}

func (db *PMTilesSpatialDatabase) ReaderURI(ctx context.Context, path string) string {
	return path
}

func (db *PMTilesSpatialDatabase) spatialDatabaseFromTile(ctx context.Context, t maptile.Tile) (database.SpatialDatabase, error) {

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)

	logger := slog.Default()
	logger = logger.With("path", path)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to create database", "time", time.Since(t1))
	}()

	logger.Debug("Get spatial database for tile")

	features, err := db.featuresForTile(ctx, t)

	if err != nil {
		logger.Error("Failed to derive features for tile", "error", err)
		return nil, fmt.Errorf("Failed to derive features for tile %s, %w", path, err)
	}

	logger = logger.With("spatial database uri", db.spatial_database_uri)
	logger = logger.With("count features", len(features))

	spatial_db, err := database.NewSpatialDatabase(ctx, db.spatial_database_uri)

	if err != nil {
		logger.Error("Failed to instantiate spatial database", "error", err)
		return nil, fmt.Errorf("Failed to create spatial database for '%s', %w", db.spatial_database_uri, err)
	}

	seen := make(map[string]bool)

	wg := new(sync.WaitGroup)

	for idx, f := range features {

		// START OF to remove once we've finished pruning layer data in featuresForTile

		str_id := fmt.Sprintf("%v", f.ID)

		if str_id != "" {

			_, ok := seen[str_id]

			if ok {
				continue
			}
		}

		seen[str_id] = true

		// END OF to remove once we've finished pruning layer data in featuresForTile

		body, err := f.MarshalJSON()

		id_rsp := gjson.GetBytes(body, "properties.wof:id")
		id := id_rsp.Int()

		if err != nil {
			logger.Error("Failed to marshal JSON for feature", "id", id, "index", idx, "error", err)
			return nil, fmt.Errorf("Failed to marshal JSON for feature %d at offset %d, %w", id, idx, err)
		}

		// START OF to remove once we've finished pruning layer data in featuresForTile

		str_id = id_rsp.String()

		_, ok := seen[str_id]

		if ok {
			continue
		}

		seen[str_id] = true

		// END OF to remove once we've finished pruning layer data in featuresForTile

		body, err = db.decodeMVT(ctx, body)

		if err != nil {
			logger.Error("Failed to unfurl MVT for feature", "id", id, "index", idx, "error", err)
			return nil, fmt.Errorf("Failed to unfurl MVT for feature %d at offset %d, %w", id, idx, err)
		}

		if db.enable_feature_cache {

			wg.Add(1)

			go func(body []byte) {

				defer wg.Done()

				// TBD: Append/pass path to cache key here?

				_, err := db.cache_manager.CacheFeature(ctx, body)

				if err != nil {
					logger.Warn("Failed to create new feature cache", "path", path, "error", err)
				}

			}(body)
		}

		err = spatial_db.IndexFeature(ctx, body)

		if err != nil {
			logger.Error("Failed to index feature", "id", id, "index", idx, "error", err)
			return nil, fmt.Errorf("Failed to index feature %d at offset %d, %w", id, idx, err)
		}
	}

	wg.Wait()

	return spatial_db, nil
}

func (db *PMTilesSpatialDatabase) spatialDatabaseNameFromCoord(ctx context.Context, coord *orb.Point) string {

	zoom := uint32(db.zoom)
	z := maptile.Zoom(zoom)
	t := maptile.At(*coord, z)

	return fmt.Sprintf("%s-%d-%d-%d.db", db.database, t.Z, t.X, t.Y)
}

func (db *PMTilesSpatialDatabase) spatialDatabaseFromCoord(ctx context.Context, coord *orb.Point) (database.SpatialDatabase, error) {

	db.spatial_databases_mutex.Lock()
	defer db.spatial_databases_mutex.Unlock()

	db_name := db.spatialDatabaseNameFromCoord(ctx, coord)

	v, exists := db.spatial_databases_cache.Load(db_name)

	if exists {
		return v.(database.SpatialDatabase), nil
	}

	zoom := uint32(db.zoom)
	z := maptile.Zoom(zoom)
	t := maptile.At(*coord, z)

	spatial_db, err := db.spatialDatabaseFromTile(ctx, t)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	db.spatial_databases_cache.Store(db_name, spatial_db)
	return spatial_db, nil
}

func (db *PMTilesSpatialDatabase) featuresForTile(ctx context.Context, t maptile.Tile) ([]*geojson.Feature, error) {

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)

	// It's tempting to cache body (or the resultant FeatureCollection) here. Ancedotally
	// at zoom level 12 it's very easy to blow past the 400kb size limit for items in DynamoDB.
	// So, in an AWS context, we could write tile caches to a gocloud.dev/blob instance but
	// will that read really be faster than reading from the PMTiles database also in S3? Maybe?

	status_code, _, body := db.server.Get(ctx, path)

	var features []*geojson.Feature

	switch status_code {

	case 200:

		layers, err := mvt.UnmarshalGzipped(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal tile, %w", err)
		}

		// Prune layers here

		layers.ProjectToWGS84(t)

		fc := layers.ToFeatureCollections()

		_, exists := fc[db.layer]

		if !exists {
			return nil, fmt.Errorf("Missing %s layer", db.layer)
		}

		features = fc[db.layer].Features

	case 204:

		// not sure what the semantics are here but 204 is not treated as an error in protomaps
		// https://github.com/protomaps/go-pmtiles/blob/0ac8f97530b3367142cfd250585d60936d0ce643/pmtiles/loop.go#L296

		features = make([]*geojson.Feature, 0)
	default:
		return nil, fmt.Errorf("Failed to get %s, unexpected status code %d", path, status_code)
	}

	return features, nil
}

// Expand WOF values that were stringified in the process of encoding them as MVT. Customs decoders are not yet supported.
// https://docs.mapbox.com/data/tilesets/guides/vector-tiles-standards/#how-to-encode-attributes-that-arent-strings-or-numbers
func (db *PMTilesSpatialDatabase) decodeMVT(ctx context.Context, body []byte) ([]byte, error) {

	props := gjson.GetBytes(body, "properties")

	if !props.Exists() {
		return body, nil
	}

	for k, v := range props.Map() {

		switch k {
		case "wof:superseded_by", "wof:supersedes", "wof:belongsto":

			var values []int64

			err := json.Unmarshal([]byte(v.String()), &values)

			if err != nil {
				return nil, fmt.Errorf("Failed to unmarshal %s value (%s), %w", k, v.String(), err)
			}

			path := fmt.Sprintf("properties.%s", k)
			body, err = sjson.SetBytes(body, path, values)

			if err != nil {
				return nil, fmt.Errorf("Failed to set %s, %w", path, err)
			}

		case "wof:hierarchy":

			var values []map[string]int64

			err := json.Unmarshal([]byte(v.String()), &values)

			if err != nil {
				return nil, fmt.Errorf("Failed to unmarshal %s value (%s), %w", k, v.String(), err)
			}

			path := fmt.Sprintf("properties.%s", k)
			body, err = sjson.SetBytes(body, path, values)

			if err != nil {
				return nil, fmt.Errorf("Failed to set %s, %w", path, err)
			}

		default:
			// pass
		}
	}

	return body, nil
}

func openCollection(ctx context.Context, uri_t string, values map[string]interface{}) (*docstore.Collection, error) {

	t, err := uritemplates.Parse(uri_t)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI template, %w", err)
	}

	col_uri, err := t.Expand(values)

	if err != nil {
		return nil, fmt.Errorf("Failed to expand URI template values, %w", err)
	}

	col, err := aa_docstore.OpenCollection(ctx, col_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to open collection, %w", err)
	}

	return col, nil
}
