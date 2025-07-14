package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

// https://github.com/whosonfirst/whosonfirst-properties/tree/main/properties/wof#parent_id
func ParentId(body []byte) (int64, error) {

	rsp := gjson.GetBytes(body, PATH_WOF_PARENTID)

	if !rsp.Exists() {
		return 0, MissingProperty(PATH_WOF_PARENTID)
	}

	id := rsp.Int()

	// https://github.com/whosonfirst/whosonfirst-properties/tree/main/properties/wof#parent_id

	if id < -4 {
		return 0, fmt.Errorf("Invalid or unrecognized parent ID value (%d)", id)
	}

	return id, nil
}
