package properties

import (
	"github.com/tidwall/gjson"
)

func BelongsTo(body []byte) []int64 {

	by := make([]int64, 0)

	rsp := gjson.GetBytes(body, PATH_WOF_BELONGSTO)

	for _, r := range rsp.Array() {
		by = append(by, r.Int())
	}

	return by
}
