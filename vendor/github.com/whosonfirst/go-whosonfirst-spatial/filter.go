package spatial

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
)

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
