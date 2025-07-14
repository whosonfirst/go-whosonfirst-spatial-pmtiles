package properties

import (
	"github.com/tidwall/gjson"
)

func AltGeometries(body []byte) ([]string, error) {

	rsp := gjson.GetBytes(body, PATH_SRC_GEOM_ALT)
	possible := rsp.Array()

	count := len(possible)
	geoms := make([]string, count)

	for idx, r := range possible {
		geoms[idx] = r.String()
	}

	return geoms, nil
}
