package database

// Implement the whosonfirst/go-whosonfirst/v4/spatial.SpatialAPI interface.

import (
	"context"
	"iter"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial"
	"github.com/whosonfirst/go-whosonfirst/v4/spr"
)

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
