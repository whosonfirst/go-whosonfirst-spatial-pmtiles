package database

// Implement the whosonfirst/go-whosonfirst-spatial.SpatialIndex interface.

import (
	"context"
)

func (db *NullSpatialDatabase) IndexFeature(context.Context, []byte) error {
	return nil
}

// RemoveFeature removes a Who's On First GeoJSON feature from the index.
func (db *NullSpatialDatabase) RemoveFeature(context.Context, string) error {
	return nil
}

// Disconnect closes any underlying connections used by the index.
func (db *NullSpatialDatabase) Disconnect(context.Context) error {
	return nil
}
