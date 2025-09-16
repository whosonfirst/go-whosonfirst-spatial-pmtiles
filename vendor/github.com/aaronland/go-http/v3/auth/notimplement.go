package auth

import (
	"net/http"
)

func notImplementedHandler() http.Handler {

	fn := func(rsp http.ResponseWriter, req *http.Request) {
		http.Error(rsp, "Not implemented", http.StatusNotImplemented)
	}

	return http.HandlerFunc(fn)
}
