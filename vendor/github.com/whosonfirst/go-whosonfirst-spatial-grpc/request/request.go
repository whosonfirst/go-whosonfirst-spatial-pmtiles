package request

import (
	"fmt"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/query"
)

func PIPRequestFromSpatialRequest(spatial_req *spatial.PointInPolygonRequest) *query.SpatialQuery {

	asInt64 := func(fl []spatial.ExistentialFlag) []int64 {

		j := make([]int64, len(fl))

		for idx, f := range fl {

			switch f {
			case spatial.ExistentialFlag_FALSE:
				j[idx] = int64(0)
			case spatial.ExistentialFlag_TRUE:
				j[idx] = int64(1)
			default:
				j[idx] = int64(-1)
			}
		}

		return j
	}

	lat := float64(spatial_req.Latitude)
	lon := float64(spatial_req.Longitude)

	pt := orb.Point([2]float64{lon, lat})
	geom := geojson.NewGeometry(pt)

	pip_q := &query.SpatialQuery{
		Geometry:            geom,
		Placetypes:          spatial_req.Placetypes,
		Geometries:          spatial_req.Geometries,
		AlternateGeometries: spatial_req.AlternateGeometries,
		InceptionDate:       spatial_req.InceptionDate,
		CessationDate:       spatial_req.CessationDate,
		Sort:                spatial_req.Sort,
		Properties:          spatial_req.Properties,
		IsCurrent:           asInt64(spatial_req.IsCurrent),
		IsCeased:            asInt64(spatial_req.IsCeased),
		IsDeprecated:        asInt64(spatial_req.IsDeprecated),
		IsSuperseded:        asInt64(spatial_req.IsSuperseded),
		IsSuperseding:       asInt64(spatial_req.IsSuperseding),
	}

	return pip_q
}

// https://github.com/whosonfirst/go-whosonfirst-spatial-pip/blob/main/pip.go

func NewPointInPolygonRequest(pip_q *query.SpatialQuery) (*spatial.PointInPolygonRequest, error) {

	geom := pip_q.Geometry

	if geom.Type != "Point" {
		return nil, fmt.Errorf("Invalid geometry type")
	}

	pt := geom.Coordinates.(orb.Point)

	lat32 := float32(pt.Lat())
	lon32 := float32(pt.Lon())

	is_current := existentialIntFlagsToProtobufExistentialFlags(pip_q.IsCurrent)
	is_ceased := existentialIntFlagsToProtobufExistentialFlags(pip_q.IsCeased)
	is_deprecated := existentialIntFlagsToProtobufExistentialFlags(pip_q.IsDeprecated)
	is_superseded := existentialIntFlagsToProtobufExistentialFlags(pip_q.IsSuperseded)
	is_superseding := existentialIntFlagsToProtobufExistentialFlags(pip_q.IsSuperseding)

	req := &spatial.PointInPolygonRequest{
		Latitude:            lat32,
		Longitude:           lon32,
		Placetypes:          pip_q.Placetypes,
		Geometries:          pip_q.Geometries,
		AlternateGeometries: pip_q.AlternateGeometries,
		InceptionDate:       pip_q.InceptionDate,
		CessationDate:       pip_q.CessationDate,
		IsCurrent:           is_current,
		IsCeased:            is_ceased,
		IsDeprecated:        is_deprecated,
		IsSuperseded:        is_superseded,
		IsSuperseding:       is_superseding,
		Sort:                pip_q.Sort,
		Properties:          pip_q.Properties,
	}

	return req, nil
}
