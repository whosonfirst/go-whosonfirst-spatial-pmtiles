package properties

import (
	"context"

	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureRepo(ctx context.Context, feature []byte) ([]byte, error) {

	_, err := wof_properties.Repo(feature)

	if err != nil {
		return nil, err
	}

	// Validate placetype?

	return feature, nil
}
