package request

import (
	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
)

func CoordsFromPointInPolygonRequest(req *spatial.PointInPolygonRequest) (orb.Point, error) {

	lat := float64(req.Latitude)
	lon := float64(req.Longitude)

	return orb.Point{lon, lat}, nil
}
