package maptile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	orb_maptile "github.com/paulmach/orb/maptile"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/query"
)

// PointInPolygonCandidateFeaturesFromTile will derive the bounds of map tile 't' and use that geometry
// to perform an "intersects" query against database 'db'. The result set will then be transformed in to
// GeoJSON FeatureCollection where each feature's geometry will be trim to extent of map tile 't'.
func PointInPolygonCandidateFeaturesFromTile(ctx context.Context, db database.SpatialDatabase, q *query.SpatialQuery, t *orb_maptile.Tile) (*geojson.FeatureCollection, error) {

	tile_bounds := t.Bound()
	tile_geom := tile_bounds.ToPolygon()

	q.Geometry = geojson.NewGeometry(tile_geom)

	intersects_fn, err := query.NewSpatialFunction(ctx, "intersects://")

	if err != nil {
		return nil, fmt.Errorf("Failed to construct spatial fuction (intersects://), %w", err)
	}

	intersects_rsp, err := query.ExecuteQuery(ctx, db, intersects_fn, q)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query, %w", err)
	}

	fc := geojson.NewFeatureCollection()

	for _, r := range intersects_rsp.Results() {

		id, err := strconv.ParseInt(r.Id(), 10, 64)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive WOF ID from SPR ID '%s', %w", r.Id(), err)
		}

		body, err := wof_reader.LoadBytes(ctx, db, id)

		if err != nil {

			if errors.Is(err, spatial.ErrNotFound) {
				slog.Warn("Failed to read data for WOF ID, not found", "id", id)
				continue
			}

			slog.Info("POO")
			return nil, fmt.Errorf("Failed to read data for WOF ID %d, %w", id, err)
		}

		f, err := geojson.UnmarshalFeature(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal feature for WOF ID %d, %w", id, err)
		}

		// Clipping happens below
		fc.Append(f)
	}

	col := make([]orb.Geometry, len(fc.Features))

	for idx, f := range fc.Features {
		col[idx] = f.Geometry
	}

	col = clip.Collection(tile_bounds, col)

	for idx, clipped_geom := range col {
		fc.Features[idx].Geometry = clipped_geom
	}

	return fc, nil
}
