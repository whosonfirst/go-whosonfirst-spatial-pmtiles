package query

import (
	"context"
	"fmt"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-spr/v2/sort"
)

type SpatialQuery struct {
	Geometry            *geojson.Geometry `json:"geometry"`
	Placetypes          []string          `json:"placetypes,omitempty"`
	Geometries          string            `json:"geometries,omitempty"`
	AlternateGeometries []string          `json:"alternate_geometries,omitempty"`
	IsCurrent           []int64           `json:"is_current,omitempty"`
	IsCeased            []int64           `json:"is_ceased,omitempty"`
	IsDeprecated        []int64           `json:"is_deprecated,omitempty"`
	IsSuperseded        []int64           `json:"is_superseded,omitempty"`
	IsSuperseding       []int64           `json:"is_superseding,omitempty"`
	InceptionDate       string            `json:"inception_date,omitempty"`
	CessationDate       string            `json:"cessation_date,omitempty"`
	Properties          []string          `json:"properties,omitempty"`
	Sort                []string          `json:"sort,omitempty"`
}

func ExecuteQuery(ctx context.Context, db database.SpatialDatabase, fn SpatialFunction, req *SpatialQuery) (spr.StandardPlacesResults, error) {

	f, err := NewSPRFilterFromSpatialQuery(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to create point in polygon filter from request, %w", err)
	}

	var principal_sorter sort.Sorter
	var follow_on_sorters []sort.Sorter

	for idx, uri := range req.Sort {

		s, err := sort.NewSorter(ctx, uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create sorter for '%s', %w", uri, err)
		}

		if idx == 0 {
			principal_sorter = s
		} else {
			follow_on_sorters = append(follow_on_sorters, s)
		}
	}

	geojson_geom := req.Geometry
	orb_geom := geojson_geom.Geometry()

	rsp, err := fn.Execute(ctx, db, orb_geom, f)

	if err != nil {
		return nil, fmt.Errorf("Failed to perform point in polygon query, %w", err)
	}

	if principal_sorter != nil {

		sorted, err := principal_sorter.Sort(ctx, rsp, follow_on_sorters...)

		if err != nil {
			return nil, fmt.Errorf("Failed to sort results, %w", err)
		}

		rsp = sorted
	}

	return rsp, nil
}
