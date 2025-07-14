package properties

import (
	"github.com/tidwall/gjson"
)

func Repo(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, PATH_WOF_REPO)

	if !rsp.Exists() {
		return "", MissingProperty(PATH_WOF_REPO)
	}

	repo := rsp.String()
	return repo, nil
}
