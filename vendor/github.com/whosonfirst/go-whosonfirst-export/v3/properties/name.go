package properties

import (
	"context"
	"fmt"

	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureName(ctx context.Context, feature []byte) ([]byte, error) {

	name, err := wof_properties.Name(feature)

	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, fmt.Errorf("%s property is empty", wof_properties.PATH_WOF_NAME)
	}

	return feature, nil
}
