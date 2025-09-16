package sanitize

import (
	go_http "net/http"
	"strconv"

	wof_sanitize "github.com/whosonfirst/go-sanitize"
)

func GetString(req *go_http.Request, param string) (string, error) {

	q := req.URL.Query()
	raw_value := q.Get(param)
	return wof_sanitize.SanitizeString(raw_value, sn_opts)
}

func GetInt64(req *go_http.Request, param string) (int64, error) {

	str_value, err := GetString(req, param)

	if err != nil {
		return -1, err
	}

	return strconv.ParseInt(str_value, 10, 64)
}

func GetBool(req *go_http.Request, param string) (bool, error) {

	str_value, err := GetString(req, param)

	if err != nil {
		return false, err
	}

	if str_value == "" {
		return false, nil
	}

	return strconv.ParseBool(str_value)
}
