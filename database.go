package pmtiles

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/docstore/memdocstore"
)

import (
	"context"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/protomaps/go-pmtiles/pmtiles"
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

	pm_cache_size := q.Get("pmtiles-cache-size")

	if pm_cache_size != "" {

		sz, err := strconv.Atoi(pm_cache_size)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?pmtiles-cache-size= parameter, %w", err)
		}

		cache_size = sz
	}

	loop, err := pmtiles.NewLoop(tile_path, logger, cache_size, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create pmtiles.Loop, %w", err)
	}

	loop.Start()

	feature_cache_uri := "mem://features/Id"
	tile_cache_uri := "mem://tiles/Path"

	feature_cache, err := docstore.OpenCollection(context.Background(), feature_cache_uri)

	if err != nil {
		return nil, fmt.Errorf("could not open collection: %w", err)
	}

	tile_cache, err := docstore.OpenCollection(context.Background(), tile_cache_uri)

	if err != nil {
		return nil, fmt.Errorf("could not open collection: %w", err)
	}

	cache_manager := NewCacheManager(feature_cache, tile_cache, logger)

	db := &PMTilesSpatialDatabase{
		loop:          loop,
		database:      database,
		logger:        logger,
		cache_manager: cache_manager,
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

		// props := f.Properties
		// fmt.Printf("Index %f %s\n", props.MustFloat64("wof:id"), props.MustString("wof:name"))

		body, err := f.MarshalJSON()

		if err != nil {
			return nil, fmt.Errorf("Failed to marshal JSON for feature at offset %d, %w", idx, err)
		}

		err = spatial_db.IndexFeature(ctx, body)

		if err != nil {
			return nil, fmt.Errorf("Failed to index feature at offset %d, %w", idx, err)
		}
	}

	// cache this?

	return spatial_db, nil
}

func (db *PMTilesSpatialDatabase) spatialDatabaseFromCoord(ctx context.Context, coord *orb.Point) (database.SpatialDatabase, error) {

	z := maptile.Zoom(uint32(9)) // fix me
	t := maptile.At(*coord, z)

	return db.spatialDatabaseFromTile(ctx, t)
}

func (db *PMTilesSpatialDatabase) featuresForTile(ctx context.Context, t maptile.Tile) ([]*geojson.Feature, error) {

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)

	if db.enable_feature_cache {

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

	go func() {

		_, err := db.cache_manager.CacheTile(ctx, path, fc[db.database].Features)

		if err != nil {
			db.logger.Printf("Failed to create new feature cache for %s, %v", path, err)
		}
	}()

	return fc[db.database].Features, nil
}
