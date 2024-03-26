package http

import (
	gohttp "net/http"
)

// NullHandler provides a `http.Handler` that returns an empty response for all requests.
func NullHandler() gohttp.Handler {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {
		return
	}

	return gohttp.HandlerFunc(fn)
}
