package sanitize

import (
	go_http "net/http"
)

func RequestString(req *go_http.Request, param string) (string, error) {

	switch req.Method {

	case "POST":
		return PostString(req, param)
	default:
		return GetString(req, param)
	}

}

func RequestInt64(req *go_http.Request, param string) (int64, error) {

	switch req.Method {

	case "POST":
		return PostInt64(req, param)
	default:
		return GetInt64(req, param)
	}

}
