package database

import (
	"context"

	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func init() {
	ctx := context.Background()
	RegisterSpatialDatabase(ctx, "null", NewNullSpatialDatabase)
}

type NullSpatialDatabase struct {
	SpatialDatabase
}

func NewNullSpatialDatabase(ctx context.Context, uri string) (SpatialDatabase, error) {
	db := &NullSpatialDatabase{}
	return db, nil
}

type NullResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
}

func (r *NullResults) Results() []spr.StandardPlacesResult {
	return make([]spr.StandardPlacesResult, 0)
}

func NewNullResults() spr.StandardPlacesResults {
	r := &NullResults{}
	return r
}
