package pmtiles

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/docstore/memdocstore"
)

import (
	"context"
	"fmt"
	"github.com/jtacoma/uritemplates"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/protomaps/go-pmtiles/pmtiles"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"gocloud.dev/docstore"
	"gocloud.dev/gcerrors"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

func init() {
	ctx := context.Background()
	database.RegisterSpatialDatabase(ctx, "pmtiles", NewPMTilesSpatialDatabase)
	reader.RegisterReader(ctx, "pmtiles", NewPMTilesSpatialDatabaseReader)
}

type PMTilesSpatialDatabase struct {
	database.SpatialDatabase
	loop                 *pmtiles.Loop
	logger               *log.Logger
	database             string
	enable_feature_cache bool
	enable_tile_cache    bool
	cache_manager        *CacheManager
	zoom                 int
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

	tile_path := q.Get("tiles")
	database := q.Get("database")

	logger := log.Default()

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

	loop, err := pmtiles.NewLoop(tile_path, logger, cache_size, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create pmtiles.Loop, %w", err)
	}

	loop.Start()

	db := &PMTilesSpatialDatabase{
		loop:     loop,
		database: database,
		logger:   logger,
		zoom:     zoom,
	}

	enable_cache := false

	q_enable_cache := q.Get("enable-cache")

	if q_enable_cache != "" {

		enabled, err := strconv.ParseBool(q_enable_cache)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?enable-cache= parameter, %w", err)
		}

		enable_cache = enabled
	}

	if enable_cache {

		feature_cache_uri_t := "mem://features/{key}"
		tile_cache_uri_t := "mem://tiles/{key}"
		cache_ttl := 300

		q_cache_ttl := q.Get("cache-ttl")
		q_feature_cache_uri_t := q.Get("feature-cache-uri")
		q_tile_cache_uri_t := q.Get("tile-cache-uri")

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

		if q_feature_cache_uri_t != "" {
			feature_cache_uri_t = q_feature_cache_uri_t
		}

		if q_tile_cache_uri_t != "" {
			tile_cache_uri_t = q_tile_cache_uri_t
		}

		feature_cache_key := "Id"
		tile_cache_key := "Path"

		feature_cache_v := map[string]interface{}{
			"key": feature_cache_key,
		}

		tile_cache_v := map[string]interface{}{
			"key": tile_cache_key,
		}

		feature_cache, err := openCollection(ctx, feature_cache_uri_t, feature_cache_v)

		if err != nil {
			return nil, fmt.Errorf("could not open feature cache collection: %w", err)
		}

		tile_cache, err := openCollection(ctx, tile_cache_uri_t, tile_cache_v)

		if err != nil {
			return nil, fmt.Errorf("could not open tile cache collection: %w", err)
		}

		cache_manager_opts := &CacheManagerOptions{
			FeatureCollection: feature_cache,
			TileCollection:    tile_cache,
			Logger:            logger,
			CacheTTL:          cache_ttl,
		}

		cache_manager := NewCacheManager(ctx, cache_manager_opts)

		db.cache_manager = cache_manager

		db.enable_feature_cache = true
		db.enable_tile_cache = true

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

	defer spatial_db.Disconnect(ctx)

	return spatial_db.PointInPolygon(ctx, coord, filters...)
}

func (db *PMTilesSpatialDatabase) PointInPolygonCandidates(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*spatial.PointInPolygonCandidate, error) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	defer spatial_db.Disconnect(ctx)

	return spatial_db.PointInPolygonCandidates(ctx, coord, filters...)
}

func (db *PMTilesSpatialDatabase) PointInPolygonWithChannels(ctx context.Context, spr_ch chan spr.StandardPlacesResult, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to create spatial database, %w", err)
		return
	}

	defer spatial_db.Disconnect(ctx)

	spatial_db.PointInPolygonWithChannels(ctx, spr_ch, err_ch, done_ch, coord, filters...)
}

func (db *PMTilesSpatialDatabase) PointInPolygonCandidatesWithChannels(ctx context.Context, pip_ch chan *spatial.PointInPolygonCandidate, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to create spatial database, %w", err)
		return
	}

	defer spatial_db.Disconnect(ctx)

	spatial_db.PointInPolygonCandidatesWithChannels(ctx, pip_ch, err_ch, done_ch, coord, filters...)
}

func (db *PMTilesSpatialDatabase) Disconnect(ctx context.Context) error {

	db.cache_manager.Close(ctx)
	return nil
}

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
		return nil, fmt.Errorf("Failed to get feature cache for %s, %w", path, err)
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

	features, err := db.featuresForTile(ctx, t)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive features for tile %s, %w", path, err)
	}

	spatial_db, err := database.NewSpatialDatabase(ctx, "sqlite://?dsn=:memory:")

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	for idx, f := range features {

		body, err := f.MarshalJSON()

		id_rsp := gjson.GetBytes(body, "properties.wof:id")
		id := id_rsp.Int()

		if err != nil {
			return nil, fmt.Errorf("Failed to marshal JSON for feature %d at offset %d, %w", id, idx, err)
		}

		err = spatial_db.IndexFeature(ctx, body)

		if err != nil {
			return nil, fmt.Errorf("Failed to index feature %d at offset %d, %w", id, idx, err)
		}
	}

	// cache this?

	return spatial_db, nil
}

func (db *PMTilesSpatialDatabase) spatialDatabaseFromCoord(ctx context.Context, coord *orb.Point) (database.SpatialDatabase, error) {

	zoom := uint32(db.zoom)

	z := maptile.Zoom(zoom)
	t := maptile.At(*coord, z)

	return db.spatialDatabaseFromTile(ctx, t)
}

func (db *PMTilesSpatialDatabase) featuresForTile(ctx context.Context, t maptile.Tile) ([]*geojson.Feature, error) {

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)

	if db.enable_tile_cache {

		tc, err := db.cache_manager.GetTileCache(ctx, path)

		if err != nil {

			if gcerrors.Code(err) != gcerrors.NotFound {
				db.logger.Printf("Failed to retrieve cache for %s, %v", path, err)
			}

		} else {

			features, err := db.cache_manager.UnmarshalTileCache(ctx, tc)

			if err != nil {
				db.logger.Printf("Failed to unmarshal features for %s, %v", path, err)
			} else {

				return features, nil
			}
		}
	}

	status_code, _, body := db.loop.Get(ctx, path)

	if status_code != 200 {
		return nil, fmt.Errorf("Failed to get %s, unexpected status code %d", path, status_code)
	}

	// not sure what the semantics are here but it's not treated as an error in protomaps
	// https://github.com/protomaps/go-pmtiles/blob/0ac8f97530b3367142cfd250585d60936d0ce643/pmtiles/loop.go#L296

	if status_code == 204 {
		features := make([]*geojson.Feature, 0)
		return features, nil
	}

	layers, err := mvt.UnmarshalGzipped(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal tile, %w", err)
	}

	layers.ProjectToWGS84(t)

	fc := layers.ToFeatureCollections()

	_, exists := fc[db.database]

	if !exists {
		return nil, fmt.Errorf("Missing %s layer", db.database)
	}

	wg := new(sync.WaitGroup)

	go func() {

		wg.Add(1)
		defer wg.Done()

		if db.enable_tile_cache {

			_, err := db.cache_manager.CacheTile(ctx, path, fc[db.database].Features)

			if err != nil {
				db.logger.Printf("Failed to create new feature cache for %s, %v", path, err)
			}

		} else if db.enable_feature_cache {

			_, err := db.cache_manager.CacheFeatures(ctx, fc[db.database].Features)

			if err != nil {
				db.logger.Printf("Failed to create new feature cache for %s, %v", path, err)
			}

		} else {
			// pass
		}

	}()

	wg.Wait()

	return fc[db.database].Features, nil
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

	col, err := docstore.OpenCollection(ctx, col_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to open collection, %w", err)
	}

	return col, nil
}
