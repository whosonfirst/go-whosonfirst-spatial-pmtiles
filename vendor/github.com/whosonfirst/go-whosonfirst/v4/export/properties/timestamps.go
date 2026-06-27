package properties

import (
	"context"
	"fmt"
)

func EnsureTimestamps(ctx context.Context, feature []byte) ([]byte, error) {

	feature, err := EnsureCreated(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:created, %w", err)
	}

	feature, err = EnsureLastModified(ctx, feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:lastmodified, %w", err)
	}

	return feature, nil
}

func RemoveTimestamps(ctx context.Context, feature []byte) ([]byte, error) {

	feature, err := RemoveLastModified(ctx, feature)

	if err != nil {
		return nil, err
	}

	return feature, nil
}
