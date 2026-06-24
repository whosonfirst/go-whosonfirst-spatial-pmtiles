package spatial

import (
	"context"
	"iter"

	"github.com/paulmach/orb"
	_ "github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	_ "github.com/whosonfirst/go-writer/v3"
)

// SpatialAPI defines an interface for performing spatial operations on a collection of Who's On First GeoJSON features.
type SpatialAPI interface {
	// PointInPolygon performs a point-in-polygon operation to retrieve matching records from the index.
	PointInPolygon(context.Context, *orb.Point, ...Filter) (spr.StandardPlacesResults, error)
	// PointInPolygon performs a point-in-polygon operation yielding matching records in an iterator.
	PointInPolygonWithIterator(context.Context, *orb.Point, ...Filter) iter.Seq2[spr.StandardPlacesResult, error]
	// Intersects performs a intersects operation (as in intersecting geometries) to retrieve matching records from the index.
	Intersects(context.Context, orb.Geometry, ...Filter) (spr.StandardPlacesResults, error)
	// IntersectsWithIterator performs a intersects operation (as in intersecting geometries) yielding matching records in an iterator.
	IntersectsWithIterator(context.Context, orb.Geometry, ...Filter) iter.Seq2[spr.StandardPlacesResult, error]
}

// SpatialIndex defines an interface for performing spatial operations, including adding and removing features, on a collection of Who's On First GeoJSON features.
type SpatialIndex interface {
	SpatialAPI
	// IndexFeature adds a Who's On First GeoJSON feature to the index.
	IndexFeature(context.Context, []byte) error
	// RemoveFeature removes a Who's On First GeoJSON feature from the index.
	RemoveFeature(context.Context, string) error
	// Disconnect closes any underlying connections used by the index.
	Disconnect(context.Context) error
}
