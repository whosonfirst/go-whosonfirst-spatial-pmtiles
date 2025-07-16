package pmtiles

// Implement the whosonfirst/go-whosonfirst-spatial.SpatialIndex interface.

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/maptile/tilecover"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func (db *PMTilesSpatialDatabase) IndexFeature(context.Context, []byte) error {
	return spatial.ErrNotImplemented
}

func (db *PMTilesSpatialDatabase) RemoveFeature(context.Context, string) error {
	return spatial.ErrNotImplemented
}

func (db *PMTilesSpatialDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	defer func() {
		db.releaseSpatialDatabase(ctx, coord)
		go atomic.AddInt64(&db.count_pip, 1)
	}()

	return spatial_db.PointInPolygon(ctx, coord, filters...)
}

func (db *PMTilesSpatialDatabase) PointInPolygonWithIterator(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		spatial_db, err := db.spatialDatabaseFromCoord(ctx, coord)

		if err != nil {
			yield(nil, fmt.Errorf("Failed to create spatial database, %w", err))
			return
		}

		defer func() {
			db.releaseSpatialDatabase(ctx, coord)
			go atomic.AddInt64(&db.count_pip, 1)
		}()

		for r, err := range spatial_db.PointInPolygonWithIterator(ctx, coord, filters...) {
			yield(r, err)

			if err != nil {
				break
			}
		}
	}
}

func (db *PMTilesSpatialDatabase) Intersects(ctx context.Context, geom orb.Geometry, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	results := make([]spr.StandardPlacesResult, 0)

	for r, err := range db.IntersectsWithIterator(ctx, geom, filters...) {

		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	spr_results := &PMTilesResults{
		Places: results,
	}

	return spr_results, nil
}

func (db *PMTilesSpatialDatabase) IntersectsWithIterator(ctx context.Context, geom orb.Geometry, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		features, err := db.featuresFromTilesForGeom(ctx, geom)

		if err != nil {
			yield(nil, err)
			return
		}

		wg := new(sync.WaitGroup)

		for id, id_features := range features {

			logger := slog.Default()
			logger = logger.With("id", id)

			intersects := false

			for _, f := range id_features {

				f_geom := f.Geometry

				f_intersects, err := geo.Intersects(f_geom, geom)

				if err != nil {
					logger.Error("Failed to determine if feature intersects", "error", err)
					continue
				}

				intersects = f_intersects

				if intersects {
					logger.Debug("Feature does not intersect")
					break
				}
			}

			if intersects {

				var f *geojson.Feature

				switch len(id_features) {
				case 1:
					f = id_features[0]
				default:

					f = id_features[0]

					// Merge geometries from different tiles in to a single feature

					polys := make([]orb.Polygon, 0)

					for _, f2 := range id_features {

						switch f2.Geometry.GeoJSONType() {
						case "Polygon":
							polys = append(polys, f2.Geometry.(orb.Polygon))
						case "MultiPolygon":

							for _, p := range f2.Geometry.(orb.MultiPolygon) {
								polys = append(polys, p)
							}
						default:
							slog.Warn("Unsupported geometry type for merging", "type", f2.Geometry.GeoJSONType())
						}
					}

					f.Geometry = orb.MultiPolygon(polys)
				}

				enc_f, err := f.MarshalJSON()

				if err != nil {
					logger.Error("Failed to marshal feature", "error", err)
					yield(nil, err)
					return
				}

				s, err := spr.WhosOnFirstSPR(enc_f)

				if err != nil {
					logger.Error("Failed to derive SPR for feature", "error", err)
					yield(nil, err)
					return
				}

				if db.enable_feature_cache {

					wg.Add(1)

					go func(body []byte) {

						defer wg.Done()

						// TBD: Append/pass path to cache key here?

						_, err := db.cache_manager.CacheFeature(ctx, body)

						if err != nil {
							logger.Warn("Failed to create new feature cache", "id", s.Id(), "error", err)
						}

					}(enc_f)
				}

				yield(s, nil)
			}
		}

		wg.Wait()
	}
}

func (db *PMTilesSpatialDatabase) pruneSpatialDatabases(ctx context.Context) {

	logger := slog.Default()

	total := 0
	pruned := 0

	now := time.Now()

	defer func() {
		logger.Info("Time to prune databases", "total", total, "pruned", pruned, "time", time.Since(now))
	}()

	db.spatial_databases_cache_mutex.Lock()
	db.spatial_databases_releaser_mutex.Lock()

	defer func() {
		db.spatial_databases_cache_mutex.Unlock()
		db.spatial_databases_releaser_mutex.Unlock()
	}()

	for db_name, t_remove := range db.spatial_databases_releaser {

		total += 1

		spatial_db, db_exists := db.spatial_databases_cache[db_name]

		if db_exists {

			if now.Before(t_remove) {
				continue
			}

			// This is important. Without it memory is not freed up.
			spatial_db.Disconnect(ctx)
			delete(db.spatial_databases_cache, db_name)
			delete(db.spatial_databases_releaser, db_name)

			pruned += 1
		}
	}

	return
}

func (db *PMTilesSpatialDatabase) releaseSpatialDatabase(ctx context.Context, geom any) {

	var db_name string

	switch geom.(type) {
	case *orb.Point:
		db_name = db.spatialDatabaseNameFromCoord(ctx, geom.(*orb.Point))
	default:
		slog.Warn("Unsupport database release type", "type", fmt.Sprintf("%T", geom))
		return
	}

	count := db.spatial_databases_counter.Increment(db_name, -1)

	logger := slog.Default()
	logger = logger.With("db", db_name)
	logger = logger.With("count", count)

	if count > 0 {
		return
	}

	db.spatial_databases_releaser_mutex.Lock()
	defer db.spatial_databases_releaser_mutex.Unlock()

	_, exists := db.spatial_databases_releaser[db_name]

	if exists {
		return
	}

	i := int(float32(db.spatial_databases_ttl*1000) / 3.0)

	ttl_ms := rand.IntN(i)
	ttl_d := time.Duration(ttl_ms) * time.Millisecond

	now := time.Now()
	then := now.Add(ttl_d)

	db.spatial_databases_releaser[db_name] = then
}

func (db *PMTilesSpatialDatabase) Disconnect(ctx context.Context) error {

	db.spatial_databases_ticker_done <- true

	if db.cache_manager != nil {
		db.cache_manager.Close()
	}

	db.spatial_databases_cache_mutex.Lock()
	db.spatial_databases_releaser_mutex.Lock()

	for db_name, spatial_db := range db.spatial_databases_cache {
		spatial_db.Disconnect(ctx)
		delete(db.spatial_databases_cache, db_name)
		delete(db.spatial_databases_releaser, db_name)
	}

	db.spatial_databases_cache_mutex.Unlock()
	db.spatial_databases_releaser_mutex.Unlock()

	return nil
}

func (db *PMTilesSpatialDatabase) spatialDatabaseFromTile(ctx context.Context, coord *orb.Point) (database.SpatialDatabase, error) {

	path := db.tilePathFromCoord(ctx, coord)

	logger := slog.Default()
	logger = logger.With("path", path)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to create database", "time", time.Since(t1))
	}()

	t := db.mapTileFromCoord(ctx, coord)

	features, err := db.featuresForTile(ctx, t)

	if err != nil {
		logger.Error("Failed to derive features for tile", "error", err)
		return nil, fmt.Errorf("Failed to derive features for tile %s, %w", path, err)
	}

	logger = logger.With("spatial database uri", db.spatial_database_uri)
	logger = logger.With("count features", len(features))

	db_uri, err := url.Parse(db.spatial_database_uri)

	if err != nil {
		return nil, err
	}

	db_q := db_uri.Query()
	dsn := db_q.Get("dsn")

	if strings.Contains(dsn, "{dbname}") {

		dbname := fmt.Sprintf("%d-%d-%d", t.X, t.Y, t.Z)
		new_dsn := strings.Replace(dsn, "{dbname}", dbname, 1)

		db_q.Del("dsn")
		db_q.Set("dsn", new_dsn)

		db_uri.RawQuery = db_q.Encode()
	}

	spatial_db, err := database.NewSpatialDatabase(ctx, db_uri.String())

	if err != nil {
		logger.Error("Failed to instantiate spatial database", "error", err)
		return nil, fmt.Errorf("Failed to create spatial database for '%s', %w", db_uri.String(), err)
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

func (db *PMTilesSpatialDatabase) mapTileFromCoord(ctx context.Context, coord *orb.Point) maptile.Tile {

	zoom := uint32(db.zoom)
	z := maptile.Zoom(zoom)
	t := maptile.At(*coord, z)

	return t
}

func (db *PMTilesSpatialDatabase) tilePathFromCoord(ctx context.Context, coord *orb.Point) string {

	t := db.mapTileFromCoord(ctx, coord)

	return fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)
}

func (db *PMTilesSpatialDatabase) spatialDatabaseNameFromCoord(ctx context.Context, coord *orb.Point) string {

	t := db.mapTileFromCoord(ctx, coord)
	return fmt.Sprintf("%s-%d-%d-%d.db", db.database, t.Z, t.X, t.Y)
}

func (db *PMTilesSpatialDatabase) spatialDatabaseFromCoord(ctx context.Context, coord *orb.Point) (database.SpatialDatabase, error) {

	db_name := db.spatialDatabaseNameFromCoord(ctx, coord)

	db.spatial_databases_cache_mutex.Lock()
	defer db.spatial_databases_cache_mutex.Unlock()

	v, exists := db.spatial_databases_cache[db_name]

	if exists {
		db.spatial_databases_counter.Increment(db_name, 1)
		return v, nil
	}

	spatial_db, err := db.spatialDatabaseFromTile(ctx, coord)

	if err != nil {
		return nil, fmt.Errorf("Failed to create spatial database, %w", err)
	}

	db.spatial_databases_counter.Increment(db_name, 1)
	db.spatial_databases_cache[db_name] = spatial_db

	return spatial_db, nil
}

func (db *PMTilesSpatialDatabase) featuresFromTilesForGeom(ctx context.Context, geom orb.Geometry) (map[int64][]*geojson.Feature, error) {

	features_table := make(map[int64][]*geojson.Feature)

	zoom := maptile.Zoom(uint32(db.zoom))
	tiles, err := tilecover.Geometry(geom, zoom)

	if err != nil {
		slog.Error("Failed to derive tile cover", "error", err)
		return nil, fmt.Errorf("Failed to derive tile cover, %w", err)
	}

	mu := new(sync.RWMutex)

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for t, _ := range tiles {

		go func(t maptile.Tile) {

			defer func() {
				done_ch <- true
			}()

			features, err := db.featuresForTile(ctx, t)

			if err != nil {
				slog.Error("Failed to derive features for tile", "error", err)
				err_ch <- err
				return
			}

			seen := new(sync.Map)

			for _, f := range features {

				id := int64(f.ID.(float64))

				if id < 0 {
					slog.Warn("Unexpected WOF ID", "raw", f.ID)
					continue
				}

				// Skip if we've seen this ID in this tile

				_, exists := seen.LoadOrStore(id, true)

				if exists {
					continue
				}

				// Combine common features seen across (spanning) tiles

				mu.Lock()

				features_list, exists := features_table[id]

				if !exists {
					features_list = make([]*geojson.Feature, 0)
				}

				features_list = append(features_list, f)
				features_table[id] = features_list

				mu.Unlock()
			}
		}(t)

	}

	remaining := len(tiles)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, err
		}
	}

	return features_table, nil
}

func (db *PMTilesSpatialDatabase) featuresForTile(ctx context.Context, t maptile.Tile) ([]*geojson.Feature, error) {

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", db.database, t.Z, t.X, t.Y)

	// It's tempting to cache body (or the resultant FeatureCollection) here. Ancedotally
	// at zoom level 12 it's very easy to blow past the 400kb size limit for items in DynamoDB.
	// So, in an AWS context, we could write tile caches to a gocloud.dev/blob instance but
	// will that read really be faster than reading from the PMTiles database also in S3? Maybe?

	server_ctx, server_cancel := context.WithTimeout(ctx, 3*time.Second)
	defer server_cancel()

	status_code, _, body := db.server.Get(server_ctx, path)

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
