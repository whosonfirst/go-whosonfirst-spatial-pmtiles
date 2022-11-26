package pmtiles

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
)

import (
	"context"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/maptile"
	"github.com/protomaps/go-pmtiles/pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"log"
	"net/url"
)

type PMTilesDatabase struct {
	database.SpatialDatabase
	loop     *pmtiles.Loop
	logger   *log.Logger
	database string
}

func NewPMTilesDatabase(ctx context.Context, uri string) (*PMTilesDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	tile_path := q.Get("tiles")
	database := q.Get("database")

	logger := log.Default()

	cache_size := 64

	loop, err := pmtiles.NewLoop(tile_path, logger, cache_size, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create pmtiles.Loop, %w", err)
	}

	loop.Start()

	db := &PMTilesDatabase{
		loop:     loop,
		database: database,
	}

	return db, nil
}

func (db *PMTilesDatabase) IndexFeature(context.Context, []byte) error {
	return fmt.Errorf("Not implemented.")
}

func (db *PMTilesDatabase) RemoveFeature(context.Context, string) error {
	return fmt.Errorf("Not implemented.")
}

func (db *PMTilesDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

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

func (db *PMTilesDatabase) PointInPolygonCandidates(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*spatial.PointInPolygonCandidate, error) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	defer spatial_db.Disconnect(ctx)

	return spatial_db.PointInPolygonCandidates(ctx, coord, filters...)
}

func (db *PMTilesDatabase) PointInPolygonWithChannels(ctx context.Context, spr_ch chan spr.StandardPlacesResult, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to create spatial database, %w", err)
		return
	}

	defer spatial_db.Disconnect(ctx)

	spatial_db.PointInPolygonWithChannels(ctx, spr_ch, err_ch, done_ch, coord, filters...)
}

func (db *PMTilesDatabase) PointInPolygonCandidatesWithChannels(ctx context.Context, pip_ch chan *spatial.PointInPolygonCandidate, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to create spatial database, %w", err)
		return
	}

	defer spatial_db.Disconnect(ctx)

	spatial_db.PointInPolygonCandidatesWithChannels(ctx, pip_ch, err_ch, done_ch, coord, filters...)
}

func (db *PMTilesDatabase) Disconnect(ctx context.Context) error {
	return nil
}

func (db *PMTilesDatabase) spatialDatabaseFromTile(ctx context.Context, t maptile.Tile) (database.SpatialDatabase, error) {

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)

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

	spatial_db, err := database.NewSpatialDatabase(ctx, "sqlite://?dsn=:memory:")

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	for idx, f := range fc[db.database].Features {

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

func (db *PMTilesDatabase) spatialDatabaseFromCoord(ctx context.Context, coord *orb.Point) (database.SpatialDatabase, error) {

	z := maptile.Zoom(uint32(8)) // fix me
	t := maptile.At(*coord, z)

	return db.spatialDatabaseFromTile(ctx, t)
}
