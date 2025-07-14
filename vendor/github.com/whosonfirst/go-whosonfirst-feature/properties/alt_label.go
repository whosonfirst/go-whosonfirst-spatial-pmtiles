package properties

import (
	"github.com/tidwall/gjson"
)

func AltLabel(body []byte) (string, error) {
	rsp := gjson.GetBytes(body, PATH_SRC_ALT_LABEL)
	return rsp.String(), nil
}
