package export

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whosonfirst/go-whosonfirst-export/v3/properties"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

// PrepareAltFeatureWithoutTimestamps ensures the presence of necessary field and/or default values for a Who's On First "alternate geometry"
// feature record absent the `wof:lastmodified` property. This is to enable checking whether a feature record being exported has been changed.
func PrepareAltFeatureWithoutTimestamps(ctx context.Context, feature []byte) ([]byte, error) {

	feature, err := properties.EnsureWOFIdAlt(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:id, %w", err)
	}

	_, err = wof_properties.Name(feature)

	if err != nil {
		slog.Warn("Failed to derive name for alternate geometry", "error", err)
	}

	feature, err = properties.EnsurePlacetype(ctx, feature)

	if err != nil {

		return nil, fmt.Errorf("Failed to ensure placetype, %w", err)
	}

	feature, err = properties.EnsureRepo(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure repo, %w", err)
	}

	feature, err = properties.EnsureSrcGeom(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure src:geom, %w", err)
	}

	feature, err = properties.EnsureGeomHash(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure geom:hash, %w", err)
	}

	feature, err = properties.EnsureSourceAltLabel(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure src:alt_label, %w", err)
	}

	feature, err = properties.RemoveTimestamps(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to remove timestamps, %w", err)
	}

	return feature, nil
}
