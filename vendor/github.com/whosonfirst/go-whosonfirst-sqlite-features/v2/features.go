package features

import (
	"github.com/aaronland/go-sqlite/v2"
)

// FeatureTable is an interface that implements the `aaronland/go-sqlite.Table` interface
// for indexing Who's On First Feature records in SQLite databases. This interface is in
// turn implemented by code in the `tables` package.
type FeatureTable interface {
	sqlite.Table
	// IndexFeature will index a Who's On First Feature record, stored in a byte array, in a SQLite database.
	IndexFeature(sqlite.Database, []byte) error
}
