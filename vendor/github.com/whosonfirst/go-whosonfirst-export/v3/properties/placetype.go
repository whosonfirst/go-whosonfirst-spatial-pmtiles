package properties

import (
	"context"

	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsurePlacetype(ctx context.Context, feature []byte) ([]byte, error) {

	_, err := wof_properties.Placetype(feature)

	if err != nil {
		return nil, err
	}

	return feature, nil
}
