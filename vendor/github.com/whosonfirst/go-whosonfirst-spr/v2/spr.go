package spr

import (
	"github.com/sfomuseum/go-edtf"
	"github.com/whosonfirst/go-whosonfirst-flags"
)

// StandardPlacesResult is an interface which defines the minimum set of methods that a system working with a collection of Who's On First (WOF) must implement for any given record. Not all records are the same so the SPR interface is meant to serve as a baseline for common data that describes every record.
type StandardPlacesResult interface {
	// The unique ID of the place result
	Id() string
	// The unique parent ID of the place result
	ParentId() string
	// The name of the place result
	Name() string
	// The Who's On First placetype of the place result
	Placetype() string
	// The two-letter country code of the place result
	Country() string
	// The (Git) repository name where the source record for the place result is stored.
	Repo() string
	// The relative path for the Who's On First record associated with the place result
	Path() string
	// The fully-qualified URI (URL) for the Who's On First record associated with the place result
	URI() string
	// The EDTF inception date of the place result
	Inception() *edtf.EDTFDate
	// The EDTF cessation date of the place result
	Cessation() *edtf.EDTFDate
	// The latitude for the principal centroid (typically "label") of the place result
	Latitude() float64
	// The longitude for the principal centroid (typically "label") of the place result
	Longitude() float64
	// The minimum latitude of the bounding box of the place result
	MinLatitude() float64
	// The minimum longitude of the bounding box of the place result
	MinLongitude() float64
	// The maximum latitude of the bounding box of the place result
	MaxLatitude() float64
	// The maximum longitude of the bounding box of the place result
	MaxLongitude() float64
	// The Who's On First "existential" flag denoting whether the place result is "current" or not
	IsCurrent() flags.ExistentialFlag
	// The Who's On First "existential" flag denoting whether the place result is "ceased" or not
	IsCeased() flags.ExistentialFlag
	// The Who's On First "existential" flag denoting whether the place result is superseded or not
	IsDeprecated() flags.ExistentialFlag
	// The Who's On First "existential" flag denoting whether the place result has been superseded
	IsSuperseded() flags.ExistentialFlag
	// The Who's On First "existential" flag denoting whether the place result supersedes other records
	IsSuperseding() flags.ExistentialFlag
	// The list of Who's On First IDs that supersede the place result
	SupersededBy() []int64
	// The list of Who's On First IDs that are superseded by the place result
	Supersedes() []int64
	// The list of Who's On First IDs that are ancestors of the place result
	BelongsTo() []int64
	// The Unix timestamp indicating when the place result was last modified
	LastModified() int64
}

// StandardPlacesResults provides an interface for returning a list of `StandardPlacesResult` results
type StandardPlacesResults interface {
	// Results is a list of `StandardPlacesResult` instances.
	Results() []StandardPlacesResult
}
