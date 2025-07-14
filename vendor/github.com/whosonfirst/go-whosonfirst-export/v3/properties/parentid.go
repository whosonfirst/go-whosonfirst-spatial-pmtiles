package properties

import (
	"context"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureParentId(ctx context.Context, feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_PARENTID)

	if rsp.Exists() {
		return feature, nil
	}

	feature, err := sjson.SetBytes(feature, wof_properties.PATH_WOF_PARENTID, -1)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_WOF_PARENTID, err)
	}

	return feature, nil
}
