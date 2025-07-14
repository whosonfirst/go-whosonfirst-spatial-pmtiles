package export

import (
	"context"

	"github.com/whosonfirst/go-whosonfirst-export/v3/properties"
)

// PrepareTimestamps ensure that 'feature' has all the necessary timestamp-related properties
// assigning default values where necessary.
func PrepareTimestamps(ctx context.Context, feature []byte) ([]byte, error) {
	return properties.EnsureTimestamps(ctx, feature)
}
