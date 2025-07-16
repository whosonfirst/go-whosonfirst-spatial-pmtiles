package database

// Implement the whosonfirst/go-whosonfirst-spatial.SpatialIndex interface.

import (
	"context"
	"iter"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func (db *NullSpatialDatabase) IndexFeature(context.Context, []byte) error {
	return nil
}

// RemoveFeature removes a Who's On First GeoJSON feature from the index.
func (db *NullSpatialDatabase) RemoveFeature(context.Context, string) error {
	return nil
}

// PointInPolygon performs a point-in-polygon operation to retrieve matching records from the index.
func (db *NullSpatialDatabase) PointInPolygon(context.Context, *orb.Point, ...spatial.Filter) (spr.StandardPlacesResults, error) {
	return NewNullResults(), nil
}

// PointInPolygon performs a point-in-polygon operation yielding matching records in an iterator.
func (db *NullSpatialDatabase) PointInPolygonWithIterator(context.Context, *orb.Point, ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {
	return func(yield func(spr.StandardPlacesResult, error) bool) {}
}

// Intersects performs a intersects operation (as in intersecting geometries) to retrieve matching records from the index.
func (db *NullSpatialDatabase) Intersects(context.Context, orb.Geometry, ...spatial.Filter) (spr.StandardPlacesResults, error) {
	return NewNullResults(), nil
}

// IntersectsWithIterator performs a intersects operation (as in intersecting geometries) yielding matching records in an iterator.
func (db *NullSpatialDatabase) IntersectsWithIterator(context.Context, orb.Geometry, ...spatial.Filter) iter.Seq2[spr.StandardPlacesResult, error] {
	return func(yield func(spr.StandardPlacesResult, error) bool) {}
}

// Disconnect closes any underlying connections used by the index.
func (db *NullSpatialDatabase) Disconnect(context.Context) error {
	return nil
}
