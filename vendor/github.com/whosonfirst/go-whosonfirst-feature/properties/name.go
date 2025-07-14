package properties

import (
	"github.com/tidwall/gjson"
)

func Name(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, PATH_WOF_NAME)

	if !rsp.Exists() {
		return "", MissingProperty(PATH_WOF_NAME)
	}

	name := rsp.String()
	return name, nil
}
