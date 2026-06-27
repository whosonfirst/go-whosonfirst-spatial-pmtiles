package database

// Implement the whosonfirst/go-whosonfirst/v4/spatial.SpatialAPI interface.

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"sync"

	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial/filter"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial/geo"
	"github.com/whosonfirst/go-whosonfirst/v4/spr"
)

func (db *RTreeSpatialDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	results := make([]spr.StandardPlacesResult, 0)

	for r, err := range db.PointInPolygonWithIterator(ctx, coord, filters...) {

		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	spr_results := &RTreeResults{
		Places: results,
	}

	return spr_results, nil
}

func (db *RTreeSpatialDatabase) PointInPolygonWithIterator(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		rows, err := db.getIntersectsByCoord(coord)

		if err != nil {
			yield(nil, err)
			return
		}

		for r, err := range db.inflateResults(ctx, rows, coord, filters...) {

			if !yield(r, err) {
				return
			}
		}

		return
	}
}

func (db *RTreeSpatialDatabase) Intersects(ctx context.Context, geom orb.Geometry, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	results := make([]spr.StandardPlacesResult, 0)

	for r, err := range db.IntersectsWithIterator(ctx, geom, filters...) {

		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	spr_results := &RTreeResults{
		Places: results,
	}

	return spr_results, nil
}

func (db *RTreeSpatialDatabase) IntersectsWithIterator(ctx context.Context, geom orb.Geometry, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		bound := geom.Bound()
		min := bound.Min
		max := bound.Max

		sw := rtreego.Point{min[0], min[1]}
		ne := rtreego.Point{max[0], max[1]}

		rect, err := rtreego.NewRectFromPoints(sw, ne)

		rows, err := db.getIntersectsByRect(&rect)

		if err != nil {
			yield(nil, err)
			return
		}

		// Do not return (yield) the same ID multiple times
		seen := new(sync.Map)

		for r, err := range db.inflateIntersectsResults(ctx, rows, geom, filters...) {

			if err != nil {
				if !yield(nil, err) {
					break
				}
			}

			_, exists := seen.Load(r.Id())

			if exists {
				continue
			}

			seen.Store(r.Id(), true)

			if !yield(r, nil) {
				break
			}
		}

		return
	}
}

func (r *RTreeSpatialDatabase) getIntersectsByCoord(coord *orb.Point) ([]rtreego.Spatial, error) {

	lat := coord.Y()
	lon := coord.X()

	pt := rtreego.Point{lon, lat}
	rect, err := rtreego.NewRect(pt, []float64{0.0001, 0.0001}) // how small can I make this?

	if err != nil {
		return nil, fmt.Errorf("Failed to derive rtree bounds, %w", err)
	}

	return r.getIntersectsByRect(&rect)
}

func (r *RTreeSpatialDatabase) getIntersectsByRect(rect *rtreego.Rect) ([]rtreego.Spatial, error) {

	results := r.rtree.SearchIntersect(*rect)
	return results, nil
}

func (r *RTreeSpatialDatabase) inflateResults(ctx context.Context, possible []rtreego.Spatial, c *orb.Point, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		seen := make(map[string]bool)
		mu := new(sync.RWMutex)

		done_ch := make(chan bool)
		spr_ch := make(chan spr.StandardPlacesResult)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for _, row := range possible {

			sp := row.(*RTreeSpatialIndex)

			go func(sp *RTreeSpatialIndex) {

				sp_id := sp.Id
				feature_id := sp.FeatureId

				defer func() {
					done_ch <- true
				}()

				select {
				case <-ctx.Done():
					return
				default:
					// pass
				}

				mu.RLock()
				_, ok := seen[feature_id]
				mu.RUnlock()

				if ok {
					return
				}

				mu.Lock()
				seen[feature_id] = true
				mu.Unlock()

				cache_item, err := r.retrieveCache(ctx, sp)

				if err != nil {
					slog.Error("Failed to retrieve cache item", "id", sp_id, "error", err)
					return
				}

				s := cache_item.SPR

				for _, f := range filters {

					err = filter.FilterSPR(f, s)

					if err != nil {
						return
					}
				}

				geom := cache_item.Geometry

				orb_geom := geom.Geometry()
				geom_type := orb_geom.GeoJSONType()

				contains := false

				switch geom_type {
				case "Polygon":
					contains = planar.PolygonContains(orb_geom.(orb.Polygon), *c)
				case "MultiPolygon":
					contains = planar.MultiPolygonContains(orb_geom.(orb.MultiPolygon), *c)
				default:
					slog.Debug("Geometry has unsupported geometry", "type", geom.Type)
				}

				if !contains {
					return
				}

				spr_ch <- s
			}(sp)
		}

		remaining := len(possible)

		for remaining > 0 {
			select {
			case <-done_ch:
				remaining -= 1
			case s := <-spr_ch:
				yield(s, nil)
			}
		}
	}
}

func (db *RTreeSpatialDatabase) inflateIntersectsResults(ctx context.Context, possible []rtreego.Spatial, geom orb.Geometry, filters ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {

	return func(yield func(spr.StandardPlacesResult, error) bool) {

		seen := make(map[string]bool)
		mu := new(sync.RWMutex)

		done_ch := make(chan bool)
		spr_ch := make(chan spr.StandardPlacesResult)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for _, row := range possible {

			sp := row.(*RTreeSpatialIndex)

			go func(sp *RTreeSpatialIndex) {

				defer func() {
					done_ch <- true
				}()

				sp_id := sp.Id
				feature_id := sp.FeatureId

				select {
				case <-ctx.Done():
					return
				default:
					// pass
				}

				mu.RLock()
				_, ok := seen[feature_id]
				mu.RUnlock()

				if ok {
					return
				}

				mu.Lock()
				seen[feature_id] = true
				mu.Unlock()

				cache_item, err := db.retrieveCache(ctx, sp)

				if err != nil {
					slog.Error("Failed to retrieve cache item", "id", sp_id, "error", err)
					return
				}

				s := cache_item.SPR

				for _, f := range filters {

					err = filter.FilterSPR(f, s)

					if err != nil {
						return
					}
				}

				item_geom := cache_item.Geometry

				item_orb_geom := item_geom.Geometry()
				item_geom_type := item_orb_geom.GeoJSONType()

				intersects := false

				switch item_geom_type {
				case "Polygon", "MultiPolygon":

					ok, err := geo.Intersects(item_orb_geom, geom)

					if err != nil {
						slog.Error("Failed to determine intersection", "error", err)
					}

					intersects = ok

				default:
					slog.Debug("Geometry has unsupported geometry", "type", item_geom_type)
				}

				if !intersects {
					return
				}

				spr_ch <- s
			}(sp)
		}

		remaining := len(possible)

		for remaining > 0 {
			select {
			case <-done_ch:
				remaining -= 1
			case s := <-spr_ch:
				yield(s, nil)
			}
		}
	}
}
