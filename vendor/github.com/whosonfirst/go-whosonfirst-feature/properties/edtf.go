package properties

import (
	"github.com/sfomuseum/go-edtf"
	"github.com/tidwall/gjson"
)

func Inception(body []byte) string {

	rsp := gjson.GetBytes(body, PATH_EDTF_INCEPTION)

	if !rsp.Exists() {
		return edtf.UNKNOWN
	}

	return rsp.String()
}

func Cessation(body []byte) string {

	rsp := gjson.GetBytes(body, PATH_EDTF_CESSATION)

	if !rsp.Exists() {
		return edtf.UNKNOWN
	}

	return rsp.String()
}

func Deprecated(body []byte) string {

	rsp := gjson.GetBytes(body, PATH_EDTF_DEPRECATED)
	return rsp.String()
}
