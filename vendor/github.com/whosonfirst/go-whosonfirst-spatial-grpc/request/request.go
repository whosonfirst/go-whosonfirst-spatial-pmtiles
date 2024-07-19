package request

import (
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/pip"
)

func PIPRequestFromSpatialRequest(spatial_req *spatial.PointInPolygonRequest) *pip.PointInPolygonRequest {

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

	pip_req := &pip.PointInPolygonRequest{
		Latitude:            float64(spatial_req.Latitude),
		Longitude:           float64(spatial_req.Longitude),
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

	return pip_req
}

// https://github.com/whosonfirst/go-whosonfirst-spatial-pip/blob/main/pip.go

func NewPointInPolygonRequest(pip_req *pip.PointInPolygonRequest) (*spatial.PointInPolygonRequest, error) {

	lat32 := float32(pip_req.Latitude)
	lon32 := float32(pip_req.Longitude)

	is_current := existentialIntFlagsToProtobufExistentialFlags(pip_req.IsCurrent)
	is_ceased := existentialIntFlagsToProtobufExistentialFlags(pip_req.IsCeased)
	is_deprecated := existentialIntFlagsToProtobufExistentialFlags(pip_req.IsDeprecated)
	is_superseded := existentialIntFlagsToProtobufExistentialFlags(pip_req.IsSuperseded)
	is_superseding := existentialIntFlagsToProtobufExistentialFlags(pip_req.IsSuperseding)

	req := &spatial.PointInPolygonRequest{
		Latitude:            lat32,
		Longitude:           lon32,
		Placetypes:          pip_req.Placetypes,
		Geometries:          pip_req.Geometries,
		AlternateGeometries: pip_req.AlternateGeometries,
		InceptionDate:       pip_req.InceptionDate,
		CessationDate:       pip_req.CessationDate,
		IsCurrent:           is_current,
		IsCeased:            is_ceased,
		IsDeprecated:        is_deprecated,
		IsSuperseded:        is_superseded,
		IsSuperseding:       is_superseding,
		Sort:                pip_req.Sort,
		Properties:          pip_req.Properties,
	}

	return req, nil
}
