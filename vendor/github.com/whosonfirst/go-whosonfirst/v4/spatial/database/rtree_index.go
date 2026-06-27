package database

// Implement the whosonfirst/go-whosonfirst/v4/spatial.SpatialIndex interface.

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst/v4/feature/alt"
	"github.com/whosonfirst/go-whosonfirst/v4/feature/geometry"
	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial"
	"github.com/whosonfirst/go-whosonfirst/v4/spr"
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
