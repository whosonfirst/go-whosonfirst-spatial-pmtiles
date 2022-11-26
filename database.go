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
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"gocloud.dev/docstore"
	"log"
	"net/url"
	"strconv"
)

func init() {
	ctx := context.Background()
	database.RegisterSpatialDatabase(ctx, "pmtiles", NewPMTilesSpatialDatabase)
}

type PMTilesSpatialDatabase struct {
	database.SpatialDatabase
	loop                 *pmtiles.Loop
	logger               *log.Logger
	database             string
	enable_feature_cache bool
	feature_cache        *docstore.Collection
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

	db := &PMTilesSpatialDatabase{
		loop:     loop,
		database: database,
	}

	enable_fc := q.Get("enable-feature-cache")

	if enable_fc != "" {

		enable_feature_cache, err := strconv.ParseBool(enable_fc)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?enable-feature-cache= parameter, %w", err)
		}
		db.enable_feature_cache = enable_feature_cache

		if enable_feature_cache {

			feature_cache_uri := q.Get("feature-cache-uri")

			feature_cache, err := docstore.OpenCollection(context.Background(), feature_cache_uri)

			if err != nil {
				return nil, fmt.Errorf("could not open collection: %w", err)
			}

			db.feature_cache = feature_cache

			// To do: Start loop to prune feature cache
		}
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

	if db.feature_cache != nil {
		db.feature_cache.Close()
	}

	return nil
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

		tc := &TileFeaturesCache{
			Path: path,
		}

		err := db.feature_cache.Get(ctx, &tc)

		if err == nil {
			return tc.Features, nil
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

	if db.enable_feature_cache {

		go func() {

			tc := NewTileFeatureCache(path, fc[db.database].Features)
			err := db.feature_cache.Put(ctx, tc)

			if err != nil {
				db.logger.Printf("Failed to put feature cache for %s, %v", path, err)
			}
		}()
	}

	return fc[db.database].Features, nil
}
