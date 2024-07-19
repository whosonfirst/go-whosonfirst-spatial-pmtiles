package request

import (
	wof_spatial "github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	wof_filter "github.com/whosonfirst/go-whosonfirst-spatial/filter"
)

func SPRFilterFromPointInPolygonRequest(req *spatial.PointInPolygonRequest) (wof_spatial.Filter, error) {

	is_current := protobufExistentalFlagsToExistentialIntFlags(req.IsCurrent)
	is_ceased := protobufExistentalFlagsToExistentialIntFlags(req.IsCeased)
	is_deprecated := protobufExistentalFlagsToExistentialIntFlags(req.IsDeprecated)
	is_superseded := protobufExistentalFlagsToExistentialIntFlags(req.IsSuperseded)
	is_superseding := protobufExistentalFlagsToExistentialIntFlags(req.IsSuperseding)

	inputs := &wof_filter.SPRInputs{
		Placetypes:          req.Placetypes,
		IsCurrent:           is_current,
		IsCeased:            is_ceased,
		IsDeprecated:        is_deprecated,
		IsSuperseded:        is_superseded,
		IsSuperseding:       is_superseding,
		Geometries:          []string{req.Geometries},
		AlternateGeometries: req.AlternateGeometries,
		InceptionDate:       req.InceptionDate,
		CessationDate:       req.CessationDate,
	}

	return wof_filter.NewSPRFilterFromInputs(inputs)
}
