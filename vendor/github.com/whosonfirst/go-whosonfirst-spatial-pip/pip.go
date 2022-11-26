package pip

import (
	"flag"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"net/url"
	"strconv"
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

func NewPointInPolygonRequestFromFlagSet(fs *flag.FlagSet) (*PointInPolygonRequest, error) {

	req := &PointInPolygonRequest{}

	latitude, err := lookup.Float64Var(fs, flags.LATITUDE)

	if err != nil {
		return nil, err
	}

	req.Latitude = latitude

	longitude, err := lookup.Float64Var(fs, flags.LONGITUDE)

	if err != nil {
		return nil, err
	}

	req.Longitude = longitude

	placetypes, err := lookup.MultiStringVar(fs, flags.PLACETYPES)

	if err != nil {
		return nil, err
	}

	req.Placetypes = placetypes

	inception_date, err := lookup.StringVar(fs, flags.INCEPTION_DATE)

	if err != nil {
		return nil, err
	}

	cessation_date, err := lookup.StringVar(fs, flags.CESSATION_DATE)

	if err != nil {
		return nil, err
	}

	req.InceptionDate = inception_date
	req.CessationDate = cessation_date

	geometries, err := lookup.StringVar(fs, flags.GEOMETRIES)

	if err != nil {
		return nil, err
	}

	req.Geometries = geometries

	alt_geoms, err := lookup.MultiStringVar(fs, flags.ALTERNATE_GEOMETRIES)

	if err != nil {
		return nil, err
	}

	req.AlternateGeometries = alt_geoms

	is_current, err := lookup.MultiInt64Var(fs, flags.IS_CURRENT)

	if err != nil {
		return nil, err
	}

	req.IsCurrent = is_current

	is_ceased, err := lookup.MultiInt64Var(fs, flags.IS_CEASED)

	if err != nil {
		return nil, err
	}

	req.IsCeased = is_ceased

	is_deprecated, err := lookup.MultiInt64Var(fs, flags.IS_DEPRECATED)

	if err != nil {
		return nil, err
	}

	req.IsDeprecated = is_deprecated

	is_superseded, err := lookup.MultiInt64Var(fs, flags.IS_SUPERSEDED)

	if err != nil {
		return nil, err
	}

	req.IsSuperseded = is_superseded

	is_superseding, err := lookup.MultiInt64Var(fs, flags.IS_SUPERSEDING)

	if err != nil {
		return nil, err
	}

	req.IsSuperseding = is_superseding

	sort_uris, err := lookup.MultiStringVar(fs, "sort-uri")

	if err != nil {
		return nil, err
	}

	req.Sort = sort_uris

	return req, nil
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
