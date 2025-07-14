package properties

import (
	"context"
	"fmt"
	"time"

	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureLastModified(ctx context.Context, feature []byte) ([]byte, error) {

	now := int32(time.Now().Unix())

	feature, err := sjson.SetBytes(feature, wof_properties.PATH_WOF_LASTMODIFIED, now)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_WOF_LASTMODIFIED, err)
	}

	return feature, nil
}

func RemoveLastModified(ctx context.Context, feature []byte) ([]byte, error) {

	feature, err := sjson.DeleteBytes(feature, wof_properties.PATH_WOF_LASTMODIFIED)

	if err != nil {
		return nil, fmt.Errorf("Failed to unset lastmodified, %w", err)
	}

	return feature, nil
}
