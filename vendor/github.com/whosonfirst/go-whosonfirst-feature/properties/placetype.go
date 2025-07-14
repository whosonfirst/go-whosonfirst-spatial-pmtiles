package properties

import (
	"github.com/tidwall/gjson"
)

func Placetype(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, PATH_WOF_PLACETYPE)

	if !rsp.Exists() {
		return "", MissingProperty(PATH_WOF_PLACETYPE)
	}

	placetype := rsp.String()

	return placetype, nil
}
