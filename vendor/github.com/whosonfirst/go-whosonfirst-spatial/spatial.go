package spatial

import (
	"context"
	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type SpatialIndex interface {
	IndexFeature(context.Context, []byte) error
	RemoveFeature(context.Context, string) error
	PointInPolygon(context.Context, *orb.Point, ...Filter) (spr.StandardPlacesResults, error)
	PointInPolygonCandidates(context.Context, *orb.Point, ...Filter) ([]*PointInPolygonCandidate, error)
	PointInPolygonWithChannels(context.Context, chan spr.StandardPlacesResult, chan error, chan bool, *orb.Point, ...Filter)
	PointInPolygonCandidatesWithChannels(context.Context, chan *PointInPolygonCandidate, chan error, chan bool, *orb.Point, ...Filter)
	Disconnect(context.Context) error
}

type PointInPolygonCandidate struct {
	Id        string
	FeatureId string
	IsAlt     bool
	AltLabel  string
	Bounds    orb.Bound
}

type PropertiesResponse map[string]interface{}

type PropertiesResponseResults struct {
	Properties []*PropertiesResponse `json:"places"` // match spr response
}

type Filter interface {
	HasPlacetypes(flags.PlacetypeFlag) bool
	MatchesInception(flags.DateFlag) bool
	MatchesCessation(flags.DateFlag) bool
	IsCurrent(flags.ExistentialFlag) bool
	IsDeprecated(flags.ExistentialFlag) bool
	IsCeased(flags.ExistentialFlag) bool
	IsSuperseded(flags.ExistentialFlag) bool
	IsSuperseding(flags.ExistentialFlag) bool
	IsAlternateGeometry(flags.AlternateGeometryFlag) bool
	HasAlternateGeometry(flags.AlternateGeometryFlag) bool
}
