package properties

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func Source(body []byte) (string, error) {

	var source string

	possible := []string{
		PATH_SRC_ALT_LABEL,
		PATH_SRC_GEOM,
	}

	for _, path := range possible {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {
			source = rsp.String()
			break
		}
	}

	if source == "" {
		return "", MissingProperty(fmt.Sprintf("%s or %s", PATH_SRC_ALT_LABEL, PATH_SRC_GEOM))
	}

	return source, nil
}
