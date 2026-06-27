package query

import (
	"context"
	"fmt"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial/database"
	"github.com/whosonfirst/go-whosonfirst/v4/spr"
)

type PointInPolygonSpatialFunction struct {
	SpatialFunction
}

func init() {

	err := RegisterSpatialFunction(context.Background(), "pip", NewPointInPolygonSpatialFunction)

	if err != nil {
		panic(err)
	}
}

func NewPointInPolygonSpatialFunction(ctx context.Context, uri string) (SpatialFunction, error) {
	q := &PointInPolygonSpatialFunction{}
	return q, nil
}

func (q *PointInPolygonSpatialFunction) Execute(ctx context.Context, db database.SpatialDatabase, geom orb.Geometry, f ...spatial.Filter) (spr.StandardPlacesResults, error) {

	var pt orb.Point

	switch geom.GeoJSONType() {
	case "Point":
		pt = geom.(orb.Point)
	default:
		return nil, fmt.Errorf("Invalid geometry type")
	}

	return db.PointInPolygon(ctx, &pt, f...)
}
