package export

import (
	"context"
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-export/v3/properties"
)

// PrepareFeatureWithoutTimestamps ensures the presence of necessary field and/or default values for a Who's On First feature record
// absent the `wof:lastmodified` property. This is to enable checking whether a feature record being exported has been changed.
func PrepareFeatureWithoutTimestamps(ctx context.Context, feature []byte) ([]byte, error) {

	feature, err := properties.EnsureWOFId(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:id, %w", err)
	}

	feature, err = properties.EnsureName(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:name, %w", err)
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

	feature, err = properties.EnsureGeomCoords(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure geometry coordinates, %w", err)
	}

	feature, err = properties.EnsureEDTF(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure EDTF properties, %w", err)
	}

	feature, err = properties.EnsureParentId(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure parent ID, %w", err)
	}

	feature, err = properties.EnsureHierarchy(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure hierarchy, %w", err)
	}

	feature, err = properties.EnsureBelongsTo(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure belongs to, %w", err)
	}

	feature, err = properties.EnsureSupersedes(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure supersedes, %w", err)
	}

	feature, err = properties.EnsureSupersededBy(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure superseded by, %w", err)
	}

	feature, err = properties.RemoveTimestamps(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to remove timestamps, %w", err)
	}

	return feature, nil
}
