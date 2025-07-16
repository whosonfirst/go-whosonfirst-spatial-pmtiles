package database

// Implement the whosonfirst/go-whosonfirst-spatial.SpatialIndex interface.

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func (r *RTreeSpatialDatabase) Disconnect(ctx context.Context) error {
	return nil
}

func (r *RTreeSpatialDatabase) IndexFeature(ctx context.Context, body []byte) error {

	is_alt := alt.IsAlt(body)
	alt_label, _ := properties.AltLabel(body)

	if is_alt && !r.index_alt_files {
		return nil
	}

	if is_alt && alt_label == "" {
		return fmt.Errorf("Invalid alt label")
	}

	err := r.setCache(ctx, body)

	if err != nil {
		return fmt.Errorf("Failed to cache feature, %w", err)
	}

	feature_id, err := properties.Id(body)

	if err != nil {
		return fmt.Errorf("Failed to derive ID, %w", err)
	}

	str_id := strconv.FormatInt(feature_id, 10)

	// START OF put me in go-whosonfirst-feature/geometry

	geojson_geom, err := geometry.Geometry(body)

	if err != nil {
		return fmt.Errorf("Failed to derive geometry, %w", err)
	}

	orb_geom := geojson_geom.Geometry()

	bounds := make([]orb.Bound, 0)

	switch orb_geom.GeoJSONType() {

	case "MultiPolygon":

		for _, poly := range orb_geom.(orb.MultiPolygon) {

			for _, ring := range poly {
				bounds = append(bounds, ring.Bound())
			}
		}

	case "Polygon":

		for _, ring := range orb_geom.(orb.Polygon) {
			bounds = append(bounds, ring.Bound())
		}
	default:
		bounds = append(bounds, orb_geom.Bound())
	}

	// END OF put me in go-whosonfirst-feature/geometry

	for i, bbox := range bounds {

		sp_id, err := spatial.SpatialIdWithFeature(body, i)

		if err != nil {
			return fmt.Errorf("Failed to derive spatial ID, %v", err)
		}

		min := bbox.Min
		max := bbox.Max

		min_x := min[0]
		min_y := min[1]

		max_x := max[0]
		max_y := max[1]

		llat := max_y - min_y
		llon := max_x - min_x

		pt := rtreego.Point{min_x, min_y}
		rect, err := rtreego.NewRect(pt, []float64{llon, llat})

		if err != nil {

			if r.strict {
				return fmt.Errorf("Failed to derive rtree bounds, %w", err)
			}

			slog.Error("Failed to index feature", "id", sp_id, "error", err)
			return nil
		}

		sp := &RTreeSpatialIndex{
			Rect:      &rect,
			Id:        sp_id,
			FeatureId: str_id,
			IsAlt:     is_alt,
			AltLabel:  alt_label,
		}

		r.mu.Lock()
		r.rtree.Insert(sp)

		r.mu.Unlock()
	}

	return nil
}

/*

TO DO: figure out suitable comparitor

/ DeleteWithComparator removes an object from the tree using a custom
// comparator for evaluating equalness. This is useful when you want to remove
// an object from a tree but don't have a pointer to the original object
// anymore.
func (tree *Rtree) DeleteWithComparator(obj Spatial, cmp Comparator) bool {
	n := tree.findLeaf(tree.root, obj, cmp)

// Comparator compares two spatials and returns whether they are equal.
type Comparator func(obj1, obj2 Spatial) (equal bool)

func defaultComparator(obj1, obj2 Spatial) bool {
	return obj1 == obj2
}

*/

func (r *RTreeSpatialDatabase) RemoveFeature(ctx context.Context, id string) error {

	obj := &RTreeSpatialIndex{
		Rect: nil,
		Id:   id,
	}

	comparator := func(obj1, obj2 rtreego.Spatial) bool {

		// 2021/10/12 11:17:11 COMPARE 1: '101737491#:0' 2: '101737491'
		// log.Printf("COMPARE 1: '%v' 2: '%v'\n", obj1.(*RTreeSpatialIndex).Id, obj2.(*RTreeSpatialIndex).Id)

		obj1_id := obj1.(*RTreeSpatialIndex).Id
		obj2_id := obj2.(*RTreeSpatialIndex).Id

		return strings.HasPrefix(obj1_id, obj2_id)
	}

	ok := r.rtree.DeleteWithComparator(obj, comparator)

	if !ok {
		return fmt.Errorf("Failed to remove %s from rtree", id)
	}

	return nil
}

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

func (r *RTreeSpatialDatabase) setCache(ctx context.Context, body []byte) error {

	s, err := spr.WhosOnFirstSPR(body)

	if err != nil {
		return err
	}

	geom, err := geometry.Geometry(body)

	if err != nil {
		return fmt.Errorf("Failed to derive geometry for feature, %w", err)
	}

	alt_label, err := properties.AltLabel(body)

	if err != nil {
		return fmt.Errorf("Failed to derive alt label, %w", err)
	}

	feature_id, err := properties.Id(body)

	if err != nil {
		return fmt.Errorf("Failed to derive feature ID, %w", err)
	}

	cache_key := fmt.Sprintf("%d:%s", feature_id, alt_label)

	cache_item := &RTreeCache{
		Geometry: geom,
		SPR:      s,
	}

	r.gocache.Set(cache_key, cache_item, -1)
	return nil
}

func (r *RTreeSpatialDatabase) retrieveCache(ctx context.Context, sp *RTreeSpatialIndex) (*RTreeCache, error) {

	cache_key := fmt.Sprintf("%s:%s", sp.FeatureId, sp.AltLabel)

	cache_item, ok := r.gocache.Get(cache_key)

	if !ok {
		return nil, fmt.Errorf("Invalid cache ID '%s'", cache_key)
	}

	return cache_item.(*RTreeCache), nil
}
