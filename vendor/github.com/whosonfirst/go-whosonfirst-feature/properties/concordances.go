package properties

import (
	"github.com/tidwall/gjson"
)

func Concordances(body []byte) map[string]interface{} {

	concordances := make(map[string]interface{})

	rsp := gjson.GetBytes(body, PATH_WOF_CONCORDANCES)

	for k, v := range rsp.Map() {
		concordances[k] = v.Value()
	}

	return concordances
}
