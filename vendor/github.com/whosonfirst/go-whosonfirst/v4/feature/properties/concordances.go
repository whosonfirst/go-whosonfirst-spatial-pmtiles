package properties

import (
	"github.com/tidwall/gjson"
)

func Concordances(body []byte) map[string]any {

	concordances := make(map[string]any)

	rsp := gjson.GetBytes(body, "properties.wof:concordances")

	for k, v := range rsp.Map() {
		concordances[k] = v.Value()
	}

	return concordances
}
