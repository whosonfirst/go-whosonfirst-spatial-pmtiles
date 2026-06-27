package hierarchy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
	wof_reader "github.com/whosonfirst/go-whosonfirst/v4/reader"
	"github.com/whosonfirst/go-whosonfirst/v4/spr"
)

// PointInPolygonHierarchyResolverUpdateCallback is a function definition for a custom callback to convert 'spr' in to a dictionary of properties
// containining hierarchy information. Records in 'spr' are expected to be able to be read from 'r'.
type PointInPolygonHierarchyResolverUpdateCallback func(context.Context, reader.Reader, spr.StandardPlacesResult) (map[string]any, error)

// DefaultPointInPolygonHierarchyResolverUpdateCallback returns a `PointInPolygonHierarchyResolverUpdateCallback` function that will return a dictionary
// containing the following properties: wof:parent_id, wof:country, wof:hierarchy
func DefaultPointInPolygonHierarchyResolverUpdateCallback() PointInPolygonHierarchyResolverUpdateCallback {

	fn := func(ctx context.Context, r reader.Reader, parent_spr spr.StandardPlacesResult) (map[string]any, error) {

		to_update := make(map[string]any)

		if parent_spr == nil {

			to_update = map[string]any{
				"properties.wof:parent_id": -1,
			}

		} else {

			parent_id, err := strconv.ParseInt(parent_spr.Id(), 10, 64)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse ID (%s), %w", parent_spr.Id(), err)
			}

			parent_f, err := wof_reader.LoadBytes(ctx, r, parent_id)

			if err != nil {
				return nil, fmt.Errorf("Failed to load body for %d, %w", parent_id, err)
			}

			parent_hierarchy := properties.Hierarchies(parent_f)
			parent_country := properties.Country(parent_f)

			to_update = map[string]any{
				"properties.wof:parent_id": parent_id,
				"properties.wof:country":   parent_country,
				"properties.wof:hierarchy": parent_hierarchy,
			}
		}

		return to_update, nil
	}

	return fn
}
