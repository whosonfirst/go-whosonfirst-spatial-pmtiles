package pip

import (
	"net/url"
	"strconv"

	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
)

type PointInPolygonRequest struct {
	Latitude            float64  `json:"latitude"`
	Longitude           float64  `json:"longitude"`
	Placetypes          []string `json:"placetypes,omitempty"`
	Geometries          string   `json:"geometries,omitempty"`
	AlternateGeometries []string `json:"alternate_geometries,omitempty"`
	IsCurrent           []int64  `json:"is_current,omitempty"`
	IsCeased            []int64  `json:"is_ceased,omitempty"`
	IsDeprecated        []int64  `json:"is_deprecated,omitempty"`
	IsSuperseded        []int64  `json:"is_superseded,omitempty"`
	IsSuperseding       []int64  `json:"is_superseding,omitempty"`
	InceptionDate       string   `json:"inception_date,omitempty"`
	CessationDate       string   `json:"cessation_date,omitempty"`
	Properties          []string `json:"properties,omitempty"`
	Sort                []string `json:"sort,omitempty"`
}

func NewSPRFilterFromPointInPolygonRequest(req *PointInPolygonRequest) (spatial.Filter, error) {

	q := url.Values{}
	q.Set("geometries", req.Geometries)

	q.Set("inception_date", req.InceptionDate)
	q.Set("cessation_date", req.CessationDate)

	for _, v := range req.AlternateGeometries {
		q.Add("alternate_geometry", v)
	}

	for _, v := range req.Placetypes {
		q.Add("placetype", v)
	}

	for _, v := range req.IsCurrent {
		q.Add("is_current", strconv.FormatInt(v, 10))
	}

	for _, v := range req.IsCeased {
		q.Add("is_ceased", strconv.FormatInt(v, 10))
	}

	for _, v := range req.IsDeprecated {
		q.Add("is_deprecated", strconv.FormatInt(v, 10))
	}

	for _, v := range req.IsSuperseded {
		q.Add("is_superseded", strconv.FormatInt(v, 10))
	}

	for _, v := range req.IsSuperseding {
		q.Add("is_superseding", strconv.FormatInt(v, 10))
	}

	return filter.NewSPRFilterFromQuery(q)
}
