package spatial

import (
	"context"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

// SpatialIndex defines an interface for performing spatial operations on a collection of Who's On First GeoJSON features.
type SpatialIndex interface {
	// IndexFeature adds a Who's On First GeoJSON feature to the index.
	IndexFeature(context.Context, []byte) error
	// RemoveFeature removes a Who's On First GeoJSON feature from the index.
	RemoveFeature(context.Context, string) error
	// PointInPolygon performs a point-in-polygon operation to retrieve matching records from the index.
	PointInPolygon(context.Context, *orb.Point, ...Filter) (spr.StandardPlacesResults, error)
	// PointInPolygon returns the initial candidates for a point-in-polygon operation.
	PointInPolygonCandidates(context.Context, *orb.Point, ...Filter) ([]*PointInPolygonCandidate, error)
	// PointInPolygon performs a point-in-polygon operation to retrieve matching records from the index returning each match (or errors) to user-defined channels.
	PointInPolygonWithChannels(context.Context, chan spr.StandardPlacesResult, chan error, chan bool, *orb.Point, ...Filter)
	// PointInPolygon returns the initial candidates for a point-in-polygon operation to a set of user-defined channels.
	PointInPolygonCandidatesWithChannels(context.Context, chan *PointInPolygonCandidate, chan error, chan bool, *orb.Point, ...Filter)
	// Disconnect closes any underlying connections used by the index.
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
