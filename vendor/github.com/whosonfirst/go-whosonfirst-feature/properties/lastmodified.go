package properties

import (
	"github.com/tidwall/gjson"
)

func LastModified(body []byte) int64 {

	rsp := gjson.GetBytes(body, PATH_WOF_LASTMODIFIED)

	if !rsp.Exists() {
		return -1
	}

	return rsp.Int()
}
