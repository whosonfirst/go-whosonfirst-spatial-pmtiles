package query

import (
	"context"
	"fmt"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type IntersectsSpatialFunction struct {
	SpatialFunction
}

func init() {

	err := RegisterSpatialFunction(context.Background(), "intersects", NewIntersectsSpatialFunction)

	if err != nil {
		panic(err)
	}
}

func NewIntersectsSpatialFunction(ctx context.Context, uri string) (SpatialFunction, error) {
	q := &IntersectsSpatialFunction{}
	return q, nil
}

func (q *IntersectsSpatialFunction) Execute(ctx context.Context, db database.SpatialDatabase, geom orb.Geometry, f ...spatial.Filter) (spr.StandardPlacesResults, error) {

	var poly orb.Geometry

	switch geom.GeoJSONType() {
	case "Polygon":
		poly = geom.(orb.Polygon)
	case "MultiPolygon":
		poly = geom.(orb.MultiPolygon)
	default:
		return nil, fmt.Errorf("Invalid geometry type")
	}

	return db.Intersects(ctx, poly, f...)
}
